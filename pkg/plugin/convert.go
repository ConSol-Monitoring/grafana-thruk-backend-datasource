package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func anyToTime(value any) time.Time {
	switch val := value.(type) {
	case float64:
		return time.Unix(int64(val), 0)
	case json.Number:
		i, err := val.Int64()
		if err == nil {
			return time.Unix(i, 0)
		}
	}

	return time.Time{}
}

func anyToFloat64(value any) float64 {
	switch val := value.(type) {
	case float64:
		return val
	case json.Number:
		f, err := val.Float64()
		if err == nil {
			return f
		}
	}

	return 0
}

func anyToInt64(value any) int64 {
	switch val := value.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	case json.Number:
		f, err := val.Int64()
		if err == nil {
			return f
		}
	}

	return 0
}

func anyToBool(value any) bool {
	switch val := value.(type) {
	case int64:
		if vInt64, ok := value.(int64); ok && vInt64 == 1 {
			return true
		}

		return false
	case float64:
		if vFloat64, ok := value.(float64); ok && vFloat64 == 1 {
			return true
		}

		return false
	case bool:
		return val
	case string:
		b, err := strconv.ParseBool(val)
		if err == nil {
			return b
		}
	}

	return false
}

func anyToString(value any) string {
	switch val := value.(type) {
	case string:
		return val
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return fmt.Sprintf("%f", val)
	case bool:
		return strconv.FormatBool(val)
	case []string:
		return strings.Join(val, ",")
	}

	return ""
}

// To get all possible types from Thruk, run this command in Thruk sources /src/thruk/lib/Thruk/Controller/Rest/V1
// grep -nrw '"type":' docs.pm > types.txt .
// grep -nrw '"type":' livestatus_docs.pm > types.txt .
//
//nolint:nonamedreturns // naming helps to understand what it is
func inferFieldType(columnName string, columnMetadatas map[string]ThrukWrappedJSONResponseMetaColumn) (fieldType data.FieldType, specified bool, specifiedType string) {
	if columnMetadata, ok := columnMetadatas[columnName]; ok {
		// if the columnMetadata has a saved type, use it
		if columnMetadata.GrafanaDataType != data.FieldTypeUnknown {
			return columnMetadata.GrafanaDataType, false, columnMetadata.Type
		}

		switch columnMetadata.Type {
		case "number":
			return data.FieldTypeFloat64, true, columnMetadata.Type
		case "time":
			return data.FieldTypeTime, true, columnMetadata.Type
		case "bool", "boolean":
			return data.FieldTypeBool, true, columnMetadata.Type
		case "string":
			return data.FieldTypeString, true, columnMetadata.Type
		case "array_of_strings":
			return data.FieldTypeString, true, columnMetadata.Type
		case "":
			return data.FieldTypeUnknown, false, ""
		default:
			return data.FieldTypeUnknown, true, columnMetadata.Type
		}
	}

	// there is no column metadata for that column
	if strings.HasPrefix(columnName, "last_") || strings.HasPrefix(columnName, "next_") ||
		strings.HasPrefix(columnName, "start_") || strings.HasPrefix(columnName, "end_") ||
		strings.HasPrefix(columnName, "time") {
		return data.FieldTypeTime, false, ""
	}

	if strings.HasPrefix(columnName, "time_") {
		return data.FieldTypeFloat64, false, ""
	}

	return data.FieldTypeUnknown, false, ""
}

// ErrNoRowsInResponseData error type.
var ErrNoRowsInResponseData = errors.New("there is no rows in response to infer from")

// ErrResponseDataDoesNotHaveColumn error type.
var ErrResponseDataDoesNotHaveColumn = errors.New("response data does not have column")

// inferUnspecifiedFieldType function determines the type to use if metadata does not contain anything regarding the type
// this is done by checking if the first value of that column can be parsed to float64, and if not it falls back to string.
func inferUnspecifiedFieldType(thrukResp *ThrukWrappedJSONResponse, columnName string) (data.FieldType, error) {
	if thrukResp == nil {
		return data.FieldTypeString, fmt.Errorf("%w, argument: thrukResp", ErrArgumentNil)
	}

	if len(thrukResp.Data) == 0 {
		return data.FieldTypeString, ErrNoRowsInResponseData
	}

	val, ok := thrukResp.Data[0][columnName]

	if !ok {
		return data.FieldTypeString, fmt.Errorf("%w , column: %s", ErrResponseDataDoesNotHaveColumn, columnName)
	}

	_, toFloat64Ok := val.(float64)

	if toFloat64Ok {
		return data.FieldTypeFloat64, nil
	}

	return data.FieldTypeString, nil
}

// Parses the optional units added in Thruk function _get_columns_meta_for_path on API calls.
func processUnitType(columnName string, columnMetadatas map[string]ThrukWrappedJSONResponseMetaColumn) {
	if mc, ok := columnMetadatas[columnName]; ok {
		if mc.Config != nil {
			if configStructConverted, convOk := mc.Config.(struct{ Unit string }); convOk {
				switch configStructConverted.Unit {
				case "%":
					// there is nothing to do with known unit types right now
					return
				case "s":
					// there is nothing to do with known unit types right now
					return
				}
			}
		}
	}
}
