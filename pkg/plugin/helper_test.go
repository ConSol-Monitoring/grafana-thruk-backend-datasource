package plugin

import (
	"errors"
	"net/http"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
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

func TestForwardAuthHeadersAllowlist(t *testing.T) {
	t.Parallel()

	inHeaders := http.Header{
		"Cookie":         {"thruk_auth=secret"},
		"Authorization":  {"Bearer token"},
		"X-Id-Token":     {"id-token"},
		"X-Grafana-User": {"alice"},
		"X-Trace-Id":     {"trace-id"},
	}

	outReq, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://thruk.example.com/r/v1/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	forwardAuthHeaders(outReq, inHeaders)

	if got := outReq.Header.Get("Cookie"); got != "thruk_auth=secret" {
		t.Fatalf("Cookie = %q, want %q", got, "thruk_auth=secret")
	}

	if got := outReq.Header.Get("Authorization"); got != "Bearer token" {
		t.Fatalf("Authorization = %q, want %q", got, "Bearer token")
	}

	if got := outReq.Header.Get("X-Id-Token"); got != "id-token" {
		t.Fatalf("X-Id-Token = %q, want %q", got, "id-token")
	}

	if got := outReq.Header.Get("X-Grafana-User"); got != "" {
		t.Fatalf("X-Grafana-User must not be forwarded, got %q", got)
	}

	if got := outReq.Header.Get("X-Trace-Id"); got != "" {
		t.Fatalf("X-Trace-Id must not be forwarded, got %q", got)
	}
}

func TestHTTPClientOptionsSetDefaultsDisablesHeaderForwarding(t *testing.T) {
	t.Parallel()

	var opts httpclient.Options

	HTTPClientOptionsSetDefaults(&opts)

	if opts.ForwardHTTPHeaders {
		t.Fatal("ForwardHTTPHeaders must be disabled; the plugin forwards an explicit allow-list")
	}
}
