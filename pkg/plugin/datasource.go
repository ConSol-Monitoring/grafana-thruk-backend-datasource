package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"go.uber.org/zap"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

const defaultLimit = 1000

// This struct contains our own definition of the Datasource and the components it needs
// It should implement CheckHealth() , Query() , Dispose() , CallResource() etc.
type Datasource struct {
	url        string
	httpClient *http.Client
	uid        string
	logger     *zap.SugaredLogger
}

// There are more fields in the settings.JSONData of type json.RawMessage , but not all of them are needed.
// Only the necessary ones are defined in this struct to be unmarshalled.
//
// DatasourceSettingsJSONDataPartial is a partial type to parse backend.DataSourceInstanceSettings.JSONData with
// jsonData is assembled by the Grafana datasource config UI (src/components/ConfigEditor.tsx).
// It mixes the plugin's own options ThrukDataSourceOptions, fields written by @grafana/plugin-ui components ConnectionSettings, Auth, AdvancedHttpSettings, and Grafana core:
type DatasourceSettingsJSONDataPartial struct {
	// the plugin's own options
	// from interface ThrukDataSourceOptions in src/types.ts
	// ======================
	// 'thruk_auth' is always added when parsing props in ConfigEditor.tsx
	KeepCookies []string `json:"keepCookies"`
	// Has its own <Input> field in ConfigEditor.tsx
	LogLevel int64 `json:"logLevel"`
	// Has its own <Input> field in ConfigEditor.tsx
	LogPath string `json:"logPath"`
	// ======================

	// from Auth part part of the ConfigEditor.tsx , no need to parse here
	// Tls configuration is parsed in grafana-plugin-sdk-go/backend/http_settings.go:parseHTTPSettings
	// TlsAuth       *bool   `json:"tlsAuth,omitempty"`
	// TlsSkipVerify *bool   `json:"tlsSkipVerify,omitempty"`
	// ServerName      *string  `json:"serverName,omitempty"`

	// from Auth part part of the ConfigEditor.tsx , no need to parse here
	// Headers are parsed in grafana-plugin-sdk-go/backend/http_settings.go:parseHTTPSettings
	// HTTPHeaderName1 *string  `json:"httpHeaderName1,omitempty"`
}

// This function is to be implemented accoring to the SDK interface
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	u := strings.TrimRight(settings.URL, "/")

	var jsonDataPartial DatasourceSettingsJSONDataPartial
	if settings.JSONData != nil {
		if err := json.Unmarshal(settings.JSONData, &jsonDataPartial); err != nil {
			return nil, fmt.Errorf("failed to parse jsonData: %w", err)
		}
	}

	lg, err := createLoggerFromDatasourceSettings(&jsonDataPartial)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	lg = lg.With("datasource", settings.UID)
	lg.Debugf("settings:\n%s", DataSourceInstanceSettingsToString(&settings))

	// SDK provides a way of building http client options directly from context. This sets up a lot of things e.g:
	// Headers to forward, TLS configuration, Basic HTTP Authentication, Proxy, Timeouts, SigV4
	httpOpts, err := settings.HTTPClientOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get http client options from context: %w", err)
	}
	HTTPClientOptionsSetDefaults(&httpOpts)

	lg.Debugf("http client options: %s", HTTPClientOptionsToString(httpOpts))

	provider := httpclient.NewProvider()
	client, err := provider.New(httpOpts)
	if err != nil {
		return nil, fmt.Errorf("could not create http client using provider: %w", err)
	}

	return &Datasource{
		url:        u,
		httpClient: client,
		uid:        settings.UID,
		logger:     lg,
	}, nil
}

// This function is to be implemented accoring to the SDK interface
func (d *Datasource) Dispose() {

}

// This function is to be implemented accoring to the SDK interface
func (d *Datasource) CheckHealth(ctx context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	d.logger.Debugf("checking connection to Thruk")

	thrukURL := d.url + "/r/v1/thruk?columns=thruk_version"

	req, err := http.NewRequestWithContext(ctx, "GET", thrukURL, nil)
	if err != nil {
		d.logger.Debugf("failed to create request: %v", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to create request: %v", err),
		}, nil
	}
	d.logger.Debugf("request cookies: %v", cookieNames(req.Header.Values("Cookie")))

	d.logger.Debugf("HTTP GET %s", thrukURL)
	start := time.Now()
	resp, err := d.httpClient.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		d.logger.Debugf("connection failed after %v: %v", elapsed, err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	// /r/v1/thruk?columns=thruk_version only returns thruk_version
	var CheckHealthResponseType struct {
		ThrukVersion string `json:"thruk_version"`
	}

	d.logger.Debugf("response code: %d, elapsed: %v", resp.StatusCode, elapsed)
	d.logger.Debugf("response body: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Unexpected status %d", resp.StatusCode),
		}, nil
	}

	if err := json.Unmarshal(body, &CheckHealthResponseType); err != nil {
		d.logger.Debugf("failed to parse response: %v", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to parse response: %v", err),
		}, nil
	}

	if CheckHealthResponseType.ThrukVersion == "" {
		d.logger.Debugf("no thruk_version in response")
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Invalid URL, did not find Thruk version in response",
		}, nil
	}

	d.logger.Debugf("connected to Thruk v%s", CheckHealthResponseType.ThrukVersion)
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Successfully connected to Thruk v" + CheckHealthResponseType.ThrukVersion,
	}, nil
}

// This function is to be implemented accoring to the SDK interface
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	d.logger.Debugf("received %d queries", len(req.Queries))

	results := make([]backend.DataResponse, len(req.Queries))
	var wg sync.WaitGroup
	for i, q := range req.Queries {
		wg.Add(1)
		go func(i int, q backend.DataQuery) {
			defer wg.Done()
			results[i] = query(ctx, d, q, req)
		}(i, q)
	}
	wg.Wait()

	response := backend.NewQueryDataResponse()
	for i, q := range req.Queries {
		response.Responses[q.RefID] = results[i]
	}

	//responseJSON, _ := response.DeepCopy().MarshalJSON()
	//d.logger.Debugf("[QueryData] response:\n%v", string(responseJSON))

	return response, nil
}

// This function is to be implemented accoring to the SDK interface
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	d.logger.Debugf("path: %s url: %s", req.Path, req.URL)

	var thrukPath string
	var extraHeaders map[string]string

	switch req.Path {
	case "tables":
		thrukPath = "/r/v1/index?columns=url&protocol=get"
	case "columns":
		parsedURL, err := url.Parse(req.URL)
		if err != nil {
			d.logger.Debugf("failed to parse request url")
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusBadRequest,
				Body:   fmt.Appendf([]byte{}, "failed to parse request url: %s", req.URL),
			})
		}

		table := parsedURL.Query().Get("table")

		if table == "" {
			d.logger.Debugf("missing table parameter")
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusBadRequest,
				Body:   []byte("missing 'table' query parameter"),
			})
		}

		table = strings.TrimPrefix(table, "/")

		thrukPath = "/r/v1/" + table

		extraHeaders = map[string]string{"X-Thruk-Output-Metadata-Only": "true"}
	case "variable-query":
		parsedURL, err := url.Parse(req.URL)
		if err != nil {
			d.logger.Debugf("failed to parse request url")
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusBadRequest,
				Body:   fmt.Appendf([]byte{}, "failed to parse request url: %s", req.URL),
			})
		}

		table := parsedURL.Query().Get("table")
		q := parsedURL.Query().Get("q")
		columns := parsedURL.Query().Get("columns")
		limit := parsedURL.Query().Get("limit")

		if table == "" {
			d.logger.Debugf("variable-query missing table parameter")
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusBadRequest,
				Body:   []byte("missing 'table' query parameter"),
			})
		}

		table = strings.TrimPrefix(table, "/")

		thrukPath = "/r/v1/" + table + "?columns=" + url.QueryEscape(columns) +
			"&q=" + url.QueryEscape(q) +
			"&limit=" + url.QueryEscape(limit)

		extraHeaders = map[string]string{"X-Thruk-Output-Metadata-Only": "true"}
	default:
		thrukPath = "/r/v1/" + strings.TrimPrefix(req.Path, "/")
	}

	thrukURL := d.url + thrukPath
	d.logger.Debugf("GET thrukURL: %s", thrukURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", thrukURL, nil)
	if err != nil {
		d.logger.Debugf("failed to create request: %v", err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   fmt.Appendf([]byte{}, "failed to create request: %v", err),
		})
	}

	for k, v := range extraHeaders {
		httpReq.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := d.httpClient.Do(httpReq)
	elapsed := time.Since(start)
	if err != nil {
		d.logger.Debugf("request failed after %v: %v", elapsed, err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   fmt.Appendf([]byte{}, "request failed: %v", err),
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.logger.Debugf("failed to read response: %v", err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   fmt.Appendf([]byte{}, "failed to read response: %v", err),
		})
	}

	d.logger.Debugf("response %d (%v, %d bytes)", resp.StatusCode, elapsed, len(body))

	return sender.Send(&backend.CallResourceResponse{
		Status: resp.StatusCode,
		Body:   body,
	})
}
