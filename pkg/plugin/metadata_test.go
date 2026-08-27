package plugin

import (
	"context"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/useragent"
)

func TestBuildQueryMeta(t *testing.T) {
	t.Parallel()

	userAgent, err := useragent.New("10.4.0", "linux", "amd64")
	if err != nil {
		t.Fatalf("failed to create user agent: %v", err)
	}

	//nolint: exhaustruct_v5
	pluginContext := backend.PluginContext{
		Namespace:     "1",
		PluginID:      "consolmonitoring-thruk-datasource",
		PluginVersion: "1.0.0",
		User: &backend.User{
			Login: "alice",
			Name:  "Alice",
			Email: "alice@example.com",
			Role:  "Viewer",
		},
		//nolint: exhaustruct_v5
		DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{UID: "ds-1"},
		UserAgent:                  userAgent,
	}

	ctx := backend.WithPluginContext(context.Background(), pluginContext)
	ctx = backend.WithUser(ctx, pluginContext.User)
	ctx = useragent.WithUserAgent(ctx, userAgent)

	//nolint:exhaustruct_v5
	req := &backend.QueryDataRequest{
		PluginContext: pluginContext,
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
	t.Parallel()

	ctx := context.Background()
	//nolint: exhaustruct_v5
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
	t.Parallel()

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
	t.Parallel()

	thrukURL := "https://thruk.example.com/r/v1/header-scoping"
	uid := "ds-header-scoping"

	hdrA := map[string][]string{"Cookie": {"thruk_auth=AAA"}}
	hdrB := map[string][]string{"Cookie": {"thruk_auth=BBB"}}

	//nolint: exhaustruct_v5
	err := writeCachedResult(queryModelFor("/index"), uid, thrukURL, hdrA, &backend.DataResponse{})
	if err != nil {
		t.Fatalf("failed to write cached result: %v", err)
	}

	_, err = getCachedResult(queryModelFor("/index"), uid, thrukURL, hdrA)
	if err != nil {
		t.Fatal("expected cache hit for identical auth context")
	}

	_, err = getCachedResult(queryModelFor("/index"), uid, thrukURL, hdrB)
	if err == nil {
		t.Fatal("expected cache miss for different auth context")
	}
}
