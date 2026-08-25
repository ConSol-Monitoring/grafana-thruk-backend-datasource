package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/useragent"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"go.uber.org/zap"
)

// What a query coming in from Grafana will contain
// Defined in types.ts as ThrukQuery in frontend part.
type QueryModel struct {
	Table     string   `json:"table"`
	Columns   []string `json:"columns"`
	Condition string   `json:"condition"`
	Limit     int      `json:"limit"`
	// can be a string
	// can be a object {"label": "Timeseries","value": "graph"}
	Type any `json:"type"`

	// metadata injected by the frontend for backend logging/auditing
	//nolint:tagliatelle // Thruk API names it like that, does not use camelCase
	DashboardUID   string `json:"dashboardUID,omitempty"`
	DashboardTitle string `json:"dashboardTitle,omitempty"`
	PanelId        int64  `json:"panelId,omitempty"`
	PanelName      string `json:"panelName,omitempty"`
	PanelPluginId  string `json:"panelPluginId,omitempty"`
	App            string `json:"app,omitempty"`
	RequestUrl     string `json:"requestUrl,omitempty"`
}

// ============================================== Query Metadata ======================================================

// QueryMetadata captures per-request metadata from Grafana SDK
// Plus the frontend-injected dashboard/panel context from the query JSON.
type QueryMetadata struct {
	// OrgID is deprecated on the SDK, not defining it
	Namespace      string
	PluginID       string
	PluginVersion  string
	DatasourceUID  string
	GrafanaVersion string
	User           *backend.User

	DashboardUID   string
	DashboardTitle string
	PanelId        int64
	PanelName      string
	PanelPluginId  string
	App            string
	RequestUrl     string

	// authHeaders are the forwarded headers relevant for authenticating the request against Thruk.
	// Used for caching.
	authHeaders map[string][]string
}

// buildQueryMetadataFromContext extracts user/org/header metadata from the query context and the request.
func buildQueryMetadataFromContext(ctx context.Context, req *backend.QueryDataRequest) QueryMetadata {
	pluginContext := backend.PluginConfigFromContext(ctx)
	meta := QueryMetadata{
		Namespace:     pluginContext.Namespace,
		PluginID:      pluginContext.PluginID,
		PluginVersion: pluginContext.PluginVersion,
		User:          backend.UserFromContext(ctx),
	}
	if pluginContext.DataSourceInstanceSettings != nil {
		meta.DatasourceUID = pluginContext.DataSourceInstanceSettings.UID
	}
	if ua := useragent.FromContext(ctx); ua != nil {
		meta.GrafanaVersion = ua.GrafanaVersion()
	}
	if req != nil {
		meta.authHeaders = buildAuthHeaders(req.GetHTTPHeaders())
	}
	return meta
}

// hasCookie reports whether the forwarded Cookie header contains a cookie with the given name.
func (qm *QueryMetadata) hasCookie(name string) bool {
	for _, value := range qm.authHeaders["Cookie"] {
		for _, part := range strings.Split(value, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, name+"=") {
				return true
			}
		}
	}
	return false
}

// String returns a compact single-line representation of the request metadata for logging/auditing.
func (qm *QueryMetadata) String() string {
	user := "none"
	if qm.User != nil {
		user = fmt.Sprintf("%s (%s, %s, role=%s)", qm.User.Login, qm.User.Name, qm.User.Email, qm.User.Role)
	}

	org := qm.Namespace
	// orgID is depreacted, not using it as an alternative

	return fmt.Sprintf(
		"user=%s org=%s grafanaVersion=%s dsUid=%s dashboardUid=%s dashboardTitle=%q panelId=%d panelName=%q app=%s url=%q thruk_auth=%t",
		user, org, qm.GrafanaVersion, qm.DatasourceUID,
		qm.DashboardUID, qm.DashboardTitle, qm.PanelId, qm.PanelName, qm.App, qm.RequestUrl,
		qm.hasCookie("thruk_auth"),
	)
}

// ====================================================================================================================

func query(ctx context.Context, datasource *Datasource, query backend.DataQuery, backendReq *backend.QueryDataRequest) backend.DataResponse {
	var queryModel QueryModel
	if err := json.Unmarshal(query.JSON, &queryModel); err != nil {
		datasource.logger.Debugf("refId=%s unmarshal error: %v", query.RefID, err)
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	// merge the frontend-injected dashboard/panel context into a per-query copy
	queryMetadata := buildQueryMetadataFromContext(ctx, backendReq)
	queryMetadata.DashboardUID = queryModel.DashboardUID
	queryMetadata.DashboardTitle = queryModel.DashboardTitle
	queryMetadata.PanelId = queryModel.PanelId
	queryMetadata.PanelName = queryModel.PanelName
	queryMetadata.PanelPluginId = queryModel.PanelPluginId
	queryMetadata.App = queryModel.App
	queryMetadata.RequestUrl = queryModel.RequestUrl

	datasource.logger.Debugf("%s refId=%s table=%s columns=%v condition=%q limit=%d type=%v",
		queryMetadata.String(), query.RefID, queryModel.Table, queryModel.Columns, queryModel.Condition, queryModel.Limit, queryModel.Type)

	rewriteAliasedEndpointsChanged := rewriteAliasedEndpoints(&queryModel)

	if rewriteAliasedEndpointsChanged {
		datasource.logger.Debugf("rewritten refId=%s table=%s columns=%v condition=%q limit=%d type=%v",
			query.RefID, queryModel.Table, queryModel.Columns, queryModel.Condition, queryModel.Limit, queryModel.Type)
	}

	thrukURL := buildQueryURL(datasource, queryModel)

	cachedResult, err := getCachedResult(&queryModel, datasource.uid, thrukURL, &queryMetadata.authHeaders)
	if err != nil {
		datasource.logger.Debugf("refId=%s error when getting cached result: %s", query.RefID, err.Error())
	}
	if cachedResult != nil {
		datasource.logger.Debugf("refId=%s found and using cached result for query %s", query.RefID, thrukURL)
		return *cachedResult.result
	}

	thrukReq, err := http.NewRequestWithContext(ctx, http.MethodGet, thrukURL, nil)
	if err != nil {
		datasource.logger.Debugf("refId=%s failed to create request: %v", query.RefID, err)
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("failed to create request: %v", err))
	}
	thrukReq.Header.Set("X-Thruk-Outputformat", "wrapped_json")

	// httpClient is set to forward cookies, and datasource is configured to add 'thruk_auth' to fowarded cookies list by default
	// set the Cookie explicitly again to be safe
	if cookies := queryMetadata.authHeaders["Cookie"]; len(cookies) > 0 {
		thrukReq.Header.Set("Cookie", strings.Join(cookies, "; "))
	}

	datasource.logger.Debugf("refId=%s HTTP GET %s", query.RefID, thrukURL)

	start := time.Now()
	resp, err := datasource.httpClient.Do(thrukReq)
	elapsed := time.Since(start)

	if err != nil {
		datasource.logger.Debugf("refId=%s request failed after %v: %v", query.RefID, elapsed, err)
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("request failed: %v", err))
	}
	defer func() (*backend.CheckHealthResult, error) {
		err := resp.Body.Close()
		return nil, fmt.Errorf("error when closing response reader: %w", err)
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		datasource.logger.Debugf("refId=%s failed to read response: %v", query.RefID, err)
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("failed to read response: %v", err))
	}

	datasource.logger.Debugf("refId=%s response code: %d %s, elapsed: %v, bytes: %d", query.RefID, resp.StatusCode, resp.Status, elapsed, len(body))

	if resp.StatusCode != http.StatusOK {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("thruk returned status: %d , body: %s", resp.StatusCode, string(body)))
	}

	parseStart := time.Now()
	result := parseThrukResponse(body, queryModel, query.TimeRange, datasource.logger)
	datasource.logger.Debugf("refId=%s parsed in %v", query.RefID, time.Since(parseStart))

	err = writeCachedResult(&queryModel, datasource.uid, thrukURL, &queryMetadata.authHeaders, &result)
	if err != nil {
		datasource.logger.Debugf("refId=%s error when writing cached result: %s", query.RefID, err.Error())
	}

	return result
}

func buildQueryURL(datasource *Datasource, queryModel QueryModel) string {
	rewriteAliasedEndpoints(&queryModel)

	path := strings.TrimPrefix(queryModel.Table, "/")
	queryUrl := fmt.Sprintf("%s/r/v1/%s", datasource.url, path)

	limit := queryModel.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	queryUrl += "?limit=" + strconv.Itoa(limit)

	if len(queryModel.Columns) > 0 && !(len(queryModel.Columns) == 1 && queryModel.Columns[0] == "*") {
		queryUrl += "&columns=" + url.QueryEscape(strings.Join(queryModel.Columns, ","))
	}
	if queryModel.Condition != "" {
		queryUrl += "&q=" + url.QueryEscape(queryModel.Condition)
	}

	return queryUrl
}

// intended to parse thruk reponses in wrapped_json format
// The "data" field of the json can either be an array of JSON objects or simply a single  JSON object
func parseThrukResponse(body []byte, qm QueryModel, timeRange backend.TimeRange, logger *zap.SugaredLogger) backend.DataResponse {
	var thrukResp ThrukWrappedJsonResponse

	// Try wrapped_json format first: { "data": <array|object> , "meta": {...} }
	var wrappedJsonFormat struct {
		Data json.RawMessage               `json:"data"`
		Meta *ThrukWrappedJsonResponseMeta `json:"meta"`
	}

	if err := json.Unmarshal(body, &wrappedJsonFormat); err == nil && wrappedJsonFormat.Data != nil {
		thrukResp.Meta = wrappedJsonFormat.Meta
		// Try data as array first, then as single object
		if err := json.Unmarshal(wrappedJsonFormat.Data, &thrukResp.Data); err != nil {
			// if unmarshalling did not work, it is probably a single json object
			var dataObj map[string]any
			if err2 := json.Unmarshal(wrappedJsonFormat.Data, &dataObj); err2 == nil {
				thrukResp.Data = []map[string]any{dataObj}
			}
		}
	} else {
		// Not wrapped_json, try as an array of jsonObjects
		var arrayOfJsonObjects []map[string]any
		if err := json.Unmarshal(body, &arrayOfJsonObjects); err == nil {
			thrukResp.Data = arrayOfJsonObjects
		} else {
			// Try as single JsonObject
			var singleJsonObject map[string]any
			if err := json.Unmarshal(body, &singleJsonObject); err != nil {
				logger.Debugf("failed to parse response: %v", err)
				return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("failed to parse response: %v", err))
			}
			thrukResp.Data = []map[string]any{singleJsonObject}
		}
	}

	if len(thrukResp.Data) == 0 {
		logger.Debugf("empty response, 0 rows returned")
		return backend.DataResponse{Frames: data.Frames{data.NewFrame("response")}}
	}

	visType := parseVisualizationType(qm.Type)

	if visType == "graph" {
		return buildTimeseriesFrames(&thrukResp, timeRange, qm, logger)
	}

	return buildTableFrame(&qm, &thrukResp, visType, logger)
}

// This function assumes that thrukResponse.Data is of type []map[string]any
// Even when the response was a single object, it is converted in parseThrukResponse into []map[string]any
func buildTableFrame(qm *QueryModel, thrukResp *ThrukWrappedJsonResponse, visType string, logger *zap.SugaredLogger) backend.DataResponse {
	// add known query types from query model and columns
	overrideKnownGrafanaDataTypes(qm, thrukResp.Meta)

	columnMetadatas := buildColumnMetadataMap(thrukResp)
	columns := determineColumnsFromThrukResponse(thrukResp)

	frame := data.NewFrame("response")
	for _, col := range columns {
		processUnitType(col, columnMetadatas)
		unknownFieldType := false

		fieldType, metadataWritenType := inferFieldType(col, columnMetadatas)
		if fieldType == data.FieldTypeUnknown {
			unknownFieldType = true
			fieldType = data.FieldTypeString
		}

		field := data.NewFieldFromFieldType(fieldType, 0)
		field.Name = col

		logger.Debugf("building column: %s, fieldType is unknown: %t, final fieldType: %s", col, unknownFieldType, FieldTypeToString(fieldType))

		for _, row := range thrukResp.Data {
			val := row[col]
			// d.logger.Printf("[Column: %s] val: %v", col, val)

			// unknown field types default to strings with white background
			if unknownFieldType {
				field.Append(anyToString(val))
				field.Config = &data.FieldConfig{
					Description: "string",
					Color:       map[string]any{"mode": "fixed", "fixedColor": "white"},
					Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
				}
				continue
			}

			//nolint:exhaustive // thruk only returns some of these types
			switch fieldType {
			case data.FieldTypeInt64:
				field.Append(anyToInt64(val))
				field.Config = &data.FieldConfig{
					Description: "int64",
					Color:       map[string]any{"mode": "fixed", "fixedColor": "blue"},
					Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
				}
			case data.FieldTypeFloat64:
				field.Append(anyToFloat64(val))
				field.Config = &data.FieldConfig{
					Description: "float64",
					Color:       map[string]any{"mode": "fixed", "fixedColor": "silver"},
					Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
				}
			case data.FieldTypeTime:
				field.Append(anyToTime(val))
				field.Config = &data.FieldConfig{
					Description: "time",
					Color:       map[string]any{"mode": "fixed", "fixedColor": "green"},
					Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
				}
			case data.FieldTypeBool:
				field.Append(anyToBool(val))
				// Bool fields have built in coloring , light green and light red
				field.Config = &data.FieldConfig{
					Description: "bool",
					Custom:      map[string]any{"cellOptions": map[string]any{"mode": "thresholds", "type": "color-background"}},
				}
			case data.FieldTypeString:
				switch metadataWritenType {
				// array of strings
				// gets a different color, fuchsia
				case "array_of_strings":
					val2 := []string{}

					if valAsAnyArray, ok := val.([]any); ok {
						for _, elem := range valAsAnyArray {
							val2 = append(val2, anyToString(elem))
						}
					}
					field.Append(anyToString(val2))
					field.Config = &data.FieldConfig{
						Description: "string",
						Color:       map[string]any{"mode": "fixed", "fixedColor": "fuchsia"},
						Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
					}
				// normal string that is to be displayed as a string
				default:
					field.Append(anyToString(val))
					field.Config = &data.FieldConfig{
						Description: "string",
						Color:       map[string]any{"mode": "fixed", "fixedColor": "purple"},
						Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
					}
				}

			default:
				field.Append(anyToString(val))
				field.Config = &data.FieldConfig{
					Description: "string",
					Color:       map[string]any{"mode": "fixed", "fixedColor": "black"},
					Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
				}
			}
		}

		frame.Fields = append(frame.Fields, field)
	}

	logger.Debugf("table: %d rows, %d columns", len(thrukResp.Data), len(columns))
	frame.Meta = &data.FrameMeta{PreferredVisualization: data.VisType(visType)}
	return backend.DataResponse{Frames: data.Frames{frame}}
}

// buildTimeseriesFrames converts tabular Thruk data into Grafana time series frames.
// This is the Go equivalent of the older frontend only plugin _fakeTimeseries() method.
// Each data row becomes its own frame. Columns with aggregation functions (e.g. "count()")
// or numeric values become the value column; remaining columns form the series alias.
// The value is spread across 10 evenly-spaced time points covering the query's time range.
func buildTimeseriesFrames(thrukResp *ThrukWrappedJsonResponse, timeRange backend.TimeRange, qm QueryModel, logger *zap.SugaredLogger) backend.DataResponse {
	const steps = 10
	from := timeRange.From.Unix()
	to := timeRange.To.Unix()
	step := (to - from) / steps

	if step <= 0 {
		step = 1
	}

	metaColumns := buildColumnMetadataMap(thrukResp)
	columns := determineColumnsFromThrukResponse(thrukResp)

	// Use user-specified columns if provided, otherwise use all response columns
	orderedColumns := columns
	if len(qm.Columns) > 0 && !(len(qm.Columns) == 1 && qm.Columns[0] == "*") {
		orderedColumns = qm.Columns
	}

	dataRows := thrukResp.Data

	// Convert single-row with many columns into key-value pairs, same as frontend
	if len(dataRows) == 1 && len(orderedColumns) > 2 {
		converted := make([]map[string]any, 0, len(orderedColumns))
		for _, key := range orderedColumns {
			converted = append(converted, map[string]any{
				"__key":   key,
				"__value": dataRows[0][key],
			})
		}
		dataRows = converted
		orderedColumns = []string{"__key", "__value"}
		// Override meta types for the converted columns
		metaColumns["__key"] = ThrukWrappedJsonResponseMetaColumn{Name: "__key"}
		metaColumns["__value"] = ThrukWrappedJsonResponseMetaColumn{Name: "__value",
			GrafanaDataType: data.FieldTypeFloat64}
	}

	// Find value column: first aggregation column, or first numeric, or first available
	valueCol := findValueColumn(orderedColumns, metaColumns, dataRows)

	// Name columns are all remaining columns not used as value
	var nameCols []string

	for _, col := range orderedColumns {
		if col != valueCol {
			nameCols = append(nameCols, col)
		}
	}

	logger.Debugf("timeseries: %d rows, valueCol=%s, nameCols=%v", len(dataRows), valueCol, nameCols)

	var frames data.Frames
	for _, row := range dataRows {
		val := row[valueCol]
		alias := valueCol
		if len(nameCols) > 0 {
			parts := make([]string, 0, len(nameCols))
			for _, nc := range nameCols {
				parts = append(parts, fmt.Sprintf("%v", row[nc]))
			}
			alias = strings.Join(parts, ";")
		}

		frame := data.NewFrame(alias)
		frame.Fields = append(frame.Fields,
			data.NewField("time", nil, make([]time.Time, steps)),
			data.NewField(alias, nil, make([]float64, steps)),
		)

		for i := range steps {
			frame.Set(0, i, time.Unix(from+step*int64(i), 0).UTC())
			frame.Set(1, i, anyToFloat64(val))
		}

		frame.Meta = &data.FrameMeta{
			PreferredVisualization: data.VisTypeGraph,
		}
		frames = append(frames, frame)
	}

	return backend.DataResponse{Frames: frames}
}

// finds the column with numerical values to use in timeseries visualization
func findValueColumn(columns []string, metaColumns map[string]ThrukWrappedJsonResponseMetaColumn, dataRows []map[string]any) string {
	if len(columns) == 0 {
		return ""
	}

	// First preference: column using aggregation function e.g. "count()" , "max()"
	for _, col := range columns {
		if strings.Contains(col, "(") && strings.Contains(col, ")") {
			return col
		}
	}

	// Second preference: first numeric column
	if len(dataRows) > 0 {
		for _, col := range columns {
			if metaColumn, ok := metaColumns[col]; ok {
				if metaColumn.GrafanaDataType != data.FieldTypeUnknown {
					if metaColumn.GrafanaDataType == data.FieldTypeFloat64 || metaColumn.GrafanaDataType == data.FieldTypeInt64 {
						return col
					}
				}
				if metaColumn.Type == "number" {
					return col
				}
			}
			// Fallback: check the actual value
			if _, isNum := dataRows[0][col].(float64); isNum {
				return col
			}
			if _, isNum := dataRows[0][col].(json.Number); isNum {
				return col
			}
		}
	}

	// Third preference: first available column
	return columns[0]
}

// if we know the table used in query model, we can iterate through the columns and add their backend types by hand
// this is a band-aid fix, only use it if thruk does not report column type metadata incorrectly.
func overrideKnownGrafanaDataTypes(queryModel *QueryModel, meta *ThrukWrappedJsonResponseMeta) {
	findAndChangeType := func(meta *ThrukWrappedJsonResponseMeta, name string, t data.FieldType) {
		for i := range meta.Columns {
			if meta.Columns[i].Name == name {
				meta.Columns[i].GrafanaDataType = t
			}
		}
	}

	//nolint:gocritic // this is an example table, its normal that it has a single case
	switch queryModel.Table {
	// example table that does not exist on Thruk API
	case "example-non-existent-table":
		// force set it to int64
		findAndChangeType(meta, "example-field", data.FieldTypeInt64)
	}
}

func parseVisualizationType(typeVal any) string {
	if s, ok := typeVal.(string); ok {
		if s == "timeseries" {
			return "graph"
		}
		return s
	}
	if obj, ok := typeVal.(map[string]any); ok {
		if v, ok := obj["value"].(string); ok {
			if v == "timeseries" {
				return "graph"
			}
			return v
		}
	}
	return "table"
}

// works on wrapped_json calls where metadata is present, in such calls it looks for resp.Meta.Columns
// or normal json calls where everything is on the same object, in such calls it looks for first row
func determineColumnsFromThrukResponse(resp *ThrukWrappedJsonResponse) []string {
	if resp.Meta != nil && len(resp.Meta.Columns) > 0 {
		cols := make([]string, 0, len(resp.Meta.Columns))
		for _, c := range resp.Meta.Columns {
			cols = append(cols, c.Name)
		}
		return cols
	}
	if len(resp.Data) > 0 {
		cols := make([]string, 0, len(resp.Data[0]))
		for key := range resp.Data[0] {
			cols = append(cols, key)
		}
		return cols
	}
	return nil
}

// buildAuthHeaders picks the auth-relevant headers out of the headers forwarded by Grafana. Returns nil when there is no auth context.
func buildAuthHeaders(headers http.Header) map[string][]string {
	// authHeadersToScopeCache are the forwarded headers relevant for identifying the authenticated user/request.
	// They are used both for logging and to scope the response cache so that one user's Thruk data is never served to another.
	var authHeadersToScopeCache = []string{"Cookie", "Authorization", "X-Id-Token", "X-Grafana-User"}

	var authHeaders map[string][]string
	for _, name := range authHeadersToScopeCache {
		if values, ok := headers[name]; ok {
			if authHeaders == nil {
				authHeaders = make(map[string][]string)
			}
			authHeaders[name] = values
		}
	}
	return authHeaders
}

// builds a map from columnMetadata.Name -> columnMetadata
// useful for fast map lookups directly using column name
func buildColumnMetadataMap(resp *ThrukWrappedJsonResponse) map[string]ThrukWrappedJsonResponseMetaColumn {
	metadataMap := make(map[string]ThrukWrappedJsonResponseMetaColumn)

	if resp.Meta != nil {
		for _, c := range resp.Meta.Columns {
			// column.Name becomes the key
			metadataMap[c.Name] = c
		}
	}

	return metadataMap
}
