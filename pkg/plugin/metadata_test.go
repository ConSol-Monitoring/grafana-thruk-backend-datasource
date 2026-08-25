package plugin

import (
	"context"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/useragent"
)

func TestBuildQueryMeta(t *testing.T) {
	userAgent, err := useragent.New("10.4.0", "linux", "amd64")
	if err != nil {
		t.Fatalf("failed to create user agent: %v", err)
	}

	pc := backend.PluginContext{
		OrgID:         1,
		Namespace:     "1",
		PluginID:      "consolmonitoring-thruk-datasource",
		PluginVersion: "1.0.0",
		User: &backend.User{
			Login: "alice",
			Name:  "Alice",
			Email: "alice@example.com",
			Role:  "Viewer",
		},
		DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{UID: "ds-1"},
		UserAgent:                  userAgent,
	}

	ctx := backend.WithPluginContext(context.Background(), pc)
	ctx = backend.WithUser(ctx, pc.User)
	ctx = useragent.WithUserAgent(ctx, userAgent)

	req := &backend.QueryDataRequest{
		PluginContext: pc,
		Headers: map[string]string{
			"http_Cookie": "thruk_auth=secret; other=value",
		},
	}

	meta := buildQueryMetadataFromContext(ctx, req)
	if meta.User == nil || meta.User.Login != "alice" {
		t.Fatalf("expected user alice, got %+v", meta.User)
	}
	if meta.DatasourceUID != "ds-1" {
		t.Fatalf("expected dsUid ds-1, got %s", meta.DatasourceUID)
	}
	if meta.GrafanaVersion != "10.4.0" {
		t.Fatalf("expected grafana version 10.4.0, got %s", meta.GrafanaVersion)
	}

	if !meta.hasCookie("thruk_auth") {
		t.Fatal("expected thruk_auth cookie to be present")
	}
	if meta.hasCookie("nonexistent") {
		t.Fatal("did not expect nonexistent cookie")
	}
	if got := meta.String(); !strings.Contains(got, "user=alice (Alice, alice@example.com, role=Viewer)") ||
		!strings.Contains(got, "thruk_auth=true") {
		t.Fatalf("unexpected meta string: %s", got)
	}
}

func TestBuildQueryMetaWithoutUser(t *testing.T) {
	ctx := context.Background()
	req := &backend.QueryDataRequest{}

	meta := buildQueryMetadataFromContext(ctx, req)
	if meta.User != nil {
		t.Fatalf("expected nil user, got %+v", meta.User)
	}
	if meta.authHeaders != nil {
		t.Fatalf("expected nil auth headers, got %v", meta.authHeaders)
	}
	if got := meta.String(); !strings.Contains(got, "user=none") {
		t.Fatalf("unexpected meta string: %s", got)
	}
}

func TestBuildAuthHeaders(t *testing.T) {
	headers := map[string][]string{
		"Cookie":                       {"thruk_auth=abc"},
		"Authorization":                {"Bearer xyz"},
		"X-Thruk-Output-Metadata-Only": {"true"},
	}

	auth := buildAuthHeaders(headers)
	if len(auth) != 2 {
		t.Fatalf("expected 2 auth headers, got %d: %v", len(auth), auth)
	}
	if _, ok := auth["Cookie"]; !ok {
		t.Fatal("expected Cookie header")
	}
	if _, ok := auth["Authorization"]; !ok {
		t.Fatal("expected Authorization header")
	}
	if _, ok := auth["X-Thruk-Output-Metadata-Only"]; ok {
		t.Fatal("did not expect metadata-only header in auth headers")
	}
}

func TestCacheHeaderScoping(t *testing.T) {
	cachedResultsMutex.Lock()
	cachedResults = cachedResults[:0]
	cachedResultsMutex.Unlock()

	thrukURL := "https://thruk.example.com/r/v1/"
	hdrA := &map[string][]string{"Cookie": {"thruk_auth=AAA"}}
	hdrB := &map[string][]string{"Cookie": {"thruk_auth=BBB"}}

	if err := writeCachedResult(&QueryModel{Table: "/index"}, "ds-1", thrukURL, hdrA, &backend.DataResponse{}); err != nil {
		t.Fatalf("failed to write cached result: %v", err)
	}

	if got := findCachedResult("ds-1", thrukURL, hdrA); got == nil {
		t.Fatal("expected cache hit for identical auth context")
	}

	if got := findCachedResult("ds-1", thrukURL, hdrB); got != nil {
		t.Fatal("expected cache miss for different auth context")
	}
}
