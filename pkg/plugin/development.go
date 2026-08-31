package plugin

import (
	"os"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// developmentModeEnvVar is the environment variable that switches the plugin backend into development mode.
// It is set by the development docker-compose (docker-compose.yaml), which is used when working directly from the sources.
const developmentModeEnvVar = "THRUK_DATASOURCE_DEVELOPMENT"

// isDevelopmentMode reports whether the plugin runs in development mode.
// The backend plugin process is started by Grafana and inherits Grafana's environment
// This returns true with the source docker-compose.yml which sets THRUK_DATASOURCE_DEVELOPMENT=true .
func isDevelopmentMode() bool {
	return strings.EqualFold(os.Getenv(developmentModeEnvVar), "true")
}

// buildDevelopmentFieldConfig returns the field configuration used in development mode
// each table cell is colored by the detected field type so that type detection is visible at a glance.
func buildDevelopmentFieldConfig(fieldType data.FieldType, specifiedType string) *data.FieldConfig {
	//nolint:exhaustive // thruk only returns some of these types
	switch fieldType {
	case data.FieldTypeInt64:
		//nolint: exhaustruct_v5
		return &data.FieldConfig{
			Description: "int64",
			Color:       map[string]any{"mode": "fixed", "fixedColor": "blue"},
			Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
		}
	case data.FieldTypeFloat64:
		//nolint: exhaustruct_v5
		return &data.FieldConfig{
			Description: "float64",
			Color:       map[string]any{"mode": "fixed", "fixedColor": "silver"},
			Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
		}
	case data.FieldTypeTime:
		//nolint: exhaustruct_v5
		return &data.FieldConfig{
			Description: "time",
			Color:       map[string]any{"mode": "fixed", "fixedColor": "green"},
			Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
		}
	case data.FieldTypeBool:
		// Bool fields have built in coloring, light green and light red
		//nolint: exhaustruct_v5
		return &data.FieldConfig{
			Description: "bool",
			Custom:      map[string]any{"cellOptions": map[string]any{"mode": "thresholds", "type": "color-background"}},
		}
	case data.FieldTypeString:
		// array of strings gets a different color, fuchsia
		if specifiedType == "array_of_strings" {
			//nolint: exhaustruct_v5
			return &data.FieldConfig{
				Description: "string",
				Color:       map[string]any{"mode": "fixed", "fixedColor": "fuchsia"},
				Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
			}
		}

		// normal string that is to be displayed as a string
		//nolint: exhaustruct_v5
		return &data.FieldConfig{
			Description: "string",
			Color:       map[string]any{"mode": "fixed", "fixedColor": "purple"},
			Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
		}
	// unknown ones are selected as strings
	default:
		//nolint: exhaustruct_v5
		return &data.FieldConfig{
			Description: "string",
			Color:       map[string]any{"mode": "fixed", "fixedColor": "black"},
			Custom:      map[string]any{"cellOptions": map[string]any{"type": "color-background"}},
		}
	}
}
