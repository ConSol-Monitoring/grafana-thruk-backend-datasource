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
		{name: "root", path: "/", wantErr: nil},
		{name: "relative endpoint", path: "hosts", wantErr: nil},
		{name: "nested endpoint", path: "/hosts/example", wantErr: nil},
		{name: "escaped space", path: "/hosts/example%20host", wantErr: nil},
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validateAPIPath(testCase.path)
			if testCase.wantErr == nil && err != nil {
				t.Fatalf("validateAPIPath(%q) returned unexpected error: %v", testCase.path, err)
			}

			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				t.Fatalf("validateAPIPath(%q) error = %v, want %v", testCase.path, err, testCase.wantErr)
			}
		})
	}
}
