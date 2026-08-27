package plugin

import (
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// queryModelFor builds a minimal query model for cache-policy lookups.
// Only the table participates in the cache policy decision.
func queryModelFor(table string) *QueryModel {
	//nolint:exhaustruct_v5 // only the table participates in cache policy lookups
	return &QueryModel{Table: table}
}

func TestAuthScopeNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		headers map[string][]string
		want    string
	}{
		{name: "nil", headers: nil, want: ""},
		{name: "empty", headers: map[string][]string{}, want: ""},
		{name: "single cookie", headers: map[string][]string{"Cookie": {"thruk_auth=AAA"}}, want: "cookie=thruk_auth=AAA;"},
		{name: "values are sorted", headers: map[string][]string{"Cookie": {"b", "a"}}, want: "cookie=a,b;"},
		{name: "keys are canonicalized", headers: map[string][]string{"cookie": {"x"}}, want: "cookie=x;"},
		{name: "multiple keys sorted", headers: map[string][]string{"Authorization": {"Bearer z"}, "Cookie": {"thruk_auth=AAA"}}, want: "authorization=Bearer z;cookie=thruk_auth=AAA;"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := authScope(tc.headers); got != tc.want {
				t.Fatalf("authScope() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCacheAuthIsolation(t *testing.T) {
	t.Parallel()

	thrukURL := "https://thruk.example.com/r/v1/auth-isolation"
	uid := "ds-auth-isolation"

	hdrAAA := map[string][]string{"Cookie": {"thruk_auth=AAA"}}
	hdrBBB := map[string][]string{"Cookie": {"thruk_auth=BBB"}}
	hdrBearer := map[string][]string{"Authorization": {"Bearer token"}}

	//nolint: exhaustruct_v5
	err := writeCachedResult(queryModelFor("/index"), uid, thrukURL, hdrAAA, &backend.DataResponse{})
	if err != nil {
		t.Fatalf("failed to write cached result: %v", err)
	}

	_, err = getCachedResult(queryModelFor("/index"), uid, thrukURL, hdrAAA)
	if err != nil {
		t.Fatalf("expected cache hit for identical auth context: %v", err)
	}

	_, err = getCachedResult(queryModelFor("/index"), uid, thrukURL, hdrBBB)
	if err == nil {
		t.Fatal("expected cache miss for different cookie")
	}

	_, err = getCachedResult(queryModelFor("/index"), uid, thrukURL, hdrBearer)
	if err == nil {
		t.Fatal("expected cache miss for different bearer token")
	}

	_, err = getCachedResult(queryModelFor("/index"), uid, thrukURL, nil)
	if err == nil {
		t.Fatal("expected cache miss for nil auth headers against an authenticated entry")
	}

	_, err = getCachedResult(queryModelFor("/index"), uid, thrukURL, map[string][]string{})
	if err == nil {
		t.Fatal("expected cache miss for empty auth headers against an authenticated entry")
	}
}

func TestCacheEmptyAuthOnlyIndex(t *testing.T) {
	t.Parallel()

	thrukURL := "https://thruk.example.com/r/v1/empty-auth"
	uid := "ds-empty-auth"

	// /users is auth-dependent and must not be cacheable without an auth context
	//nolint: exhaustruct_v5
	err := writeCachedResult(queryModelFor("/users"), uid, thrukURL, nil, &backend.DataResponse{})
	if err == nil {
		t.Fatal("expected error writing cache entry for /users with empty auth")
	}

	// /index is user-independent and may be cached without an auth context
	//nolint: exhaustruct_v5
	err = writeCachedResult(queryModelFor("/index"), uid, thrukURL, nil, &backend.DataResponse{})
	if err != nil {
		t.Fatalf("expected success writing cache entry for /index with empty auth: %v", err)
	}

	_, err = getCachedResult(queryModelFor("/index"), uid, thrukURL, nil)
	if err != nil {
		t.Fatalf("expected cache hit for empty auth /index: %v", err)
	}

	hdr := map[string][]string{"Cookie": {"thruk_auth=AAA"}}

	_, err = getCachedResult(queryModelFor("/index"), uid, thrukURL, hdr)
	if err == nil {
		t.Fatal("expected cache miss for authenticated request against an empty-auth entry")
	}
}

func TestRewriteAliasedEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		table       string
		wantTable   string
		wantChanged bool
	}{
		{name: "index", table: "/index", wantTable: "/", wantChanged: true},
		{name: "stats", table: "/thruk/stats", wantTable: "/thruk/metrics", wantChanged: true},
		{name: "node control", table: "/thruk/node-control/nodes", wantTable: "/thruk/nc/nodes", wantChanged: true},
		{name: "not an alias", table: "/hosts", wantTable: "/hosts", wantChanged: false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			//nolint:exhaustruct_v5 // only the table participates in alias rewriting
			queryModel := QueryModel{Table: testCase.table}

			changed := rewriteAliasedEndpoints(&queryModel)
			if changed != testCase.wantChanged {
				t.Fatalf("rewriteAliasedEndpoints() changed = %t, want %t", changed, testCase.wantChanged)
			}

			if queryModel.Table != testCase.wantTable {
				t.Fatalf("rewriteAliasedEndpoints() table = %q, want %q", queryModel.Table, testCase.wantTable)
			}
		})
	}
}
