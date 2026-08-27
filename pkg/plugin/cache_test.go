package plugin

import (
	"fmt"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// queryModelFor builds a minimal query model for cache-policy lookups.
// Only the table participates in the cache policy decision.
func queryModelFor(table string) *QueryModel {
	//nolint:exhaustruct_v5 // only the table participates in cache policy lookups
	return &QueryModel{Table: table}
}

// resetCache clears the global cache. Tests that assert on global cache state (size, eviction
// order) use this and therefore do not run in parallel with each other.
func resetCache() {
	cachedResultsMutex.Lock()
	cachedResults = map[cacheKey]*CachedResult{}
	cacheOrder = nil
	cachedResultsMutex.Unlock()
}

// cacheSize returns the number of entries currently held by the global cache.
func cacheSize() int {
	cachedResultsMutex.RLock()
	defer cachedResultsMutex.RUnlock()

	return len(cachedResults)
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

//nolint:paralleltest // mutates the global cache and asserts on its state
func TestCacheReplacement(t *testing.T) {
	resetCache()

	thrukURL := "https://thruk.example.com/r/v1/replacement"
	uid := "ds-replacement"

	//nolint: exhaustruct_v5
	first := &backend.DataResponse{Status: backend.StatusOK}
	//nolint: exhaustruct_v5
	second := &backend.DataResponse{Status: backend.StatusBadGateway}

	err := writeCachedResult(queryModelFor("/index"), uid, thrukURL, nil, first)
	if err != nil {
		t.Fatalf("write first: %v", err)
	}

	err = writeCachedResult(queryModelFor("/index"), uid, thrukURL, nil, second)
	if err != nil {
		t.Fatalf("write second: %v", err)
	}

	entries := cacheSize()
	if entries != 1 {
		t.Fatalf("expected a single cache entry after writing the same key twice, got %d", entries)
	}

	got, err := getCachedResult(queryModelFor("/index"), uid, thrukURL, nil)
	if err != nil {
		t.Fatalf("expected cache hit: %v", err)
	}

	if got.result.Status != backend.StatusBadGateway {
		t.Fatalf("expected the latest result to win, got status %v", got.result.Status)
	}
}

//nolint:paralleltest // mutates the global cache and asserts on its size
func TestCacheCapacityEviction(t *testing.T) {
	resetCache()

	uid := "ds-capacity"

	for i := range maxCachedResults + 1 {
		thrukURL := fmt.Sprintf("https://thruk.example.com/r/v1/%d", i)
		//nolint: exhaustruct_v5
		err := writeCachedResult(queryModelFor("/index"), uid, thrukURL, nil, &backend.DataResponse{})
		if err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}

	entries := cacheSize()
	if entries != maxCachedResults {
		t.Fatalf("expected cache size %d after overflow, got %d", maxCachedResults, entries)
	}

	_, err := getCachedResult(queryModelFor("/index"), uid, "https://thruk.example.com/r/v1/0", nil)
	if err == nil {
		t.Fatal("expected the oldest entry to have been evicted")
	}

	newest := fmt.Sprintf("https://thruk.example.com/r/v1/%d", maxCachedResults)

	_, err = getCachedResult(queryModelFor("/index"), uid, newest, nil)
	if err != nil {
		t.Fatal("expected the newest entry to still be cached")
	}
}

//nolint:paralleltest // mutates the global cache and asserts on its state
func TestCacheExpiryCleanup(t *testing.T) {
	resetCache()

	thrukURL := "https://thruk.example.com/r/v1/expiry"
	uid := "ds-expiry"

	key := cacheKey{datasourceUID: uid, thrukURL: thrukURL, authScope: ""}

	cachedResultsMutex.Lock()
	cachedResults[key] = &CachedResult{
		//nolint: exhaustruct_v5
		result:         &backend.DataResponse{},
		expirationTime: time.Now().Add(-time.Minute),
	}
	cacheOrder = append(cacheOrder, key)
	cachedResultsMutex.Unlock()

	cleanupExpiredResults()

	_, err := getCachedResult(queryModelFor("/index"), uid, thrukURL, nil)
	if err == nil {
		t.Fatal("expected expired entry to be removed by cleanup")
	}

	entries := cacheSize()
	if entries != 0 {
		t.Fatalf("expected cache to be empty after cleanup, got %d entries", entries)
	}
}

//nolint:paralleltest // mutates the global cache and asserts on its state
func TestEvictDatasourceResults(t *testing.T) {
	resetCache()

	uidA := "ds-evict-a"
	uidB := "ds-evict-b"

	//nolint: exhaustruct_v5
	err := writeCachedResult(queryModelFor("/index"), uidA, "https://thruk.example.com/r/v1/a", nil, &backend.DataResponse{})
	if err != nil {
		t.Fatalf("write A: %v", err)
	}

	//nolint: exhaustruct_v5
	err = writeCachedResult(queryModelFor("/index"), uidB, "https://thruk.example.com/r/v1/b", nil, &backend.DataResponse{})
	if err != nil {
		t.Fatalf("write B: %v", err)
	}

	evictDatasourceResults(uidA)

	_, err = getCachedResult(queryModelFor("/index"), uidA, "https://thruk.example.com/r/v1/a", nil)
	if err == nil {
		t.Fatal("expected datasource A entries to be evicted")
	}

	_, err = getCachedResult(queryModelFor("/index"), uidB, "https://thruk.example.com/r/v1/b", nil)
	if err != nil {
		t.Fatal("expected datasource B entry to remain")
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
