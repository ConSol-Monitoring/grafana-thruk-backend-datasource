package plugin

import (
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func TestIsDevelopmentMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{name: "unset or empty", envValue: "", want: false},
		{name: "true", envValue: "true", want: true},
		{name: "TRUE", envValue: "TRUE", want: true},
		{name: "false", envValue: "false", want: false},
		{name: "1", envValue: "1", want: false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(developmentModeEnvVar, testCase.envValue)

			if got := isDevelopmentMode(); got != testCase.want {
				t.Fatalf("isDevelopmentMode() = %t, want %t", got, testCase.want)
			}
		})
	}
}

func TestBuildDevelopmentFieldConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fieldType     data.FieldType
		specifiedType string
		wantColor     string
	}{
		{name: "int64", fieldType: data.FieldTypeInt64, specifiedType: "", wantColor: "blue"},
		{name: "float64", fieldType: data.FieldTypeFloat64, specifiedType: "", wantColor: "silver"},
		{name: "time", fieldType: data.FieldTypeTime, specifiedType: "", wantColor: "green"},
		{name: "string", fieldType: data.FieldTypeString, specifiedType: "", wantColor: "purple"},
		{name: "array of strings", fieldType: data.FieldTypeString, specifiedType: "array_of_strings", wantColor: "fuchsia"},
		{name: "unknown", fieldType: data.FieldTypeUnknown, specifiedType: "", wantColor: "black"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			config := buildDevelopmentFieldConfig(testCase.fieldType, testCase.specifiedType)
			if config == nil {
				t.Fatal("buildDevelopmentFieldConfig() returned nil config")
			}

			color, ok := config.Color["fixedColor"].(string)
			if !ok {
				t.Fatalf("buildDevelopmentFieldConfig() color = %v, want fixedColor %q", config.Color, testCase.wantColor)
			}

			if color != testCase.wantColor {
				t.Fatalf("buildDevelopmentFieldConfig() fixedColor = %q, want %q", color, testCase.wantColor)
			}
		})
	}
}

func TestBuildTableFrameColorsFieldsOnlyInDevelopmentMode(t *testing.T) {
	thrukResp := &ThrukWrappedJSONResponse{
		Data: []map[string]any{
			{
				"host_name":             "host1",
				"state":                 float64(0),
				"last_check":            float64(1780000000),
				"active_checks_enabled": float64(1),
			},
		},
		Meta: &ThrukWrappedJSONResponseMeta{
			Columns: []ThrukWrappedJSONResponseMetaColumn{
				{Name: "host_name", Type: "string", GrafanaDataType: data.FieldTypeUnknown, Config: nil},
				{Name: "state", Type: "number", GrafanaDataType: data.FieldTypeUnknown, Config: nil},
				{Name: "last_check", Type: "time", GrafanaDataType: data.FieldTypeUnknown, Config: nil},
				{Name: "active_checks_enabled", Type: "bool", GrafanaDataType: data.FieldTypeUnknown, Config: nil},
			},
			RequestDuration: 0,
			ParseDuration:   0,
		},
	}

	loggers := &Loggers{sdk: log.NewNullLogger()}

	checkFieldConfigs := func(t *testing.T, resp backend.DataResponse, wantSet bool) {
		t.Helper()

		if len(resp.Frames) != 1 {
			t.Fatalf("expected exactly 1 frame, got %d", len(resp.Frames))
		}

		for _, field := range resp.Frames[0].Fields {
			if wantSet && field.Config == nil {
				t.Fatalf("field %q: expected dev coloring config, got nil", field.Name)
			}

			if !wantSet && field.Config != nil {
				t.Fatalf("field %q: expected no coloring config, got %+v", field.Name, field.Config)
			}
		}
	}

	t.Run("development mode colors the fields", func(t *testing.T) {
		t.Setenv(developmentModeEnvVar, "true")

		//nolint: exhaustruct_v5
		resp := buildTableFrame(&QueryModel{}, thrukResp, "table", loggers)

		checkFieldConfigs(t, resp, true)
	})

	t.Run("production mode leaves the fields uncolored", func(t *testing.T) {
		t.Setenv(developmentModeEnvVar, "false")

		//nolint: exhaustruct_v5
		resp := buildTableFrame(&QueryModel{}, thrukResp, "table", loggers)

		checkFieldConfigs(t, resp, false)
	})
}
