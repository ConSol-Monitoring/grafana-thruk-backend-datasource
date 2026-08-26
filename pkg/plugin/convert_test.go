package plugin

import (
	"errors"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func TestInferUnspecifiedFieldType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		response  *ThrukWrappedJSONResponse
		want      data.FieldType
		wantError error
	}{
		{
			name: "float",
			response: &ThrukWrappedJSONResponse{
				Data: []map[string]any{{"value": float64(42)}},
				Meta: nil,
			},
			want:      data.FieldTypeFloat64,
			wantError: nil,
		},
		{
			name: "string",
			response: &ThrukWrappedJSONResponse{
				Data: []map[string]any{{"value": "42"}},
				Meta: nil,
			},
			want:      data.FieldTypeString,
			wantError: nil,
		},
		{
			name:      "empty response",
			response:  &ThrukWrappedJSONResponse{Data: nil, Meta: nil},
			want:      data.FieldTypeString,
			wantError: ErrNoRowsInResponseData,
		},
		{
			name: "missing column",
			response: &ThrukWrappedJSONResponse{
				Data: []map[string]any{{"other": "value"}},
				Meta: nil,
			},
			want:      data.FieldTypeString,
			wantError: ErrResponseDataDoesNotHaveColumn,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := inferUnspecifiedFieldType(testCase.response, "value")
			if got != testCase.want {
				t.Fatalf("inferUnspecifiedFieldType() type = %v, want %v", got, testCase.want)
			}

			if !errors.Is(err, testCase.wantError) {
				t.Fatalf("inferUnspecifiedFieldType() error = %v, want %v", err, testCase.wantError)
			}
		})
	}
}
