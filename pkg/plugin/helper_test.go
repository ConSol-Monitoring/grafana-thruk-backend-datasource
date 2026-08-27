package plugin

import (
	"errors"
	"testing"
)

func TestValidateAPIPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{name: "root", path: "/"},
		{name: "relative endpoint", path: "hosts"},
		{name: "nested endpoint", path: "/hosts/example"},
		{name: "escaped space", path: "/hosts/example%20host"},
		{name: "empty", path: "", wantErr: ErrInvalidAPIPath},
		{name: "query string", path: "/hosts?limit=1", wantErr: ErrInvalidAPIPath},
		{name: "raw traversal", path: "../config", wantErr: ErrAPIPathEscapes},
		{name: "nested raw traversal", path: "/hosts/../../config", wantErr: ErrAPIPathEscapes},
		{name: "encoded traversal", path: "/hosts/%2e%2e/config", wantErr: ErrAPIPathEscapes},
		{name: "encoded slash traversal", path: "/hosts%2f..%2fconfig", wantErr: ErrAPIPathEscapes},
		{name: "double encoded traversal", path: "/hosts/%252e%252e/config", wantErr: ErrAPIPathEscapes},
		{name: "backslash traversal", path: `hosts\..\config`, wantErr: ErrInvalidAPIPath},
		{name: "encoded backslash traversal", path: "hosts%5c..%5cconfig", wantErr: ErrInvalidAPIPath},
		{name: "invalid escape", path: "/hosts/%zz", wantErr: ErrInvalidAPIPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateAPIPath(tt.path)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("validateAPIPath(%q) returned unexpected error: %v", tt.path, err)
			}

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateAPIPath(%q) error = %v, want %v", tt.path, err, tt.wantErr)
			}
		})
	}
}
