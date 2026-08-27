package plugin

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// CachedResult represents a cached result, has its own expiration time.
type CachedResult struct {
	datasourceUID  string
	thrukURL       string
	authScope      string
	result         *backend.DataResponse
	expirationTime time.Time
}

//nolint:gochecknoglobals // need cache to be globally defined
var (
	cachedResults      = []*CachedResult{}
	cachedResultsMutex sync.RWMutex
)

// authScope returns a deterministic string that identifies the authentication context of a request.
// Nil and empty header maps are equivalent and both yield an empty scope.
func authScope(headers map[string][]string) string {
	if len(headers) == 0 {
		return ""
	}

	normalized := make(map[string][]string, len(headers))
	for name, values := range headers {
		canonical := http.CanonicalHeaderKey(name)
		normalized[canonical] = append(normalized[canonical], values...)
	}

	names := make([]string, 0, len(normalized))
	for name := range normalized {
		names = append(names, name)
	}

	sort.Strings(names)

	var scope strings.Builder

	for _, name := range names {
		values := append([]string(nil), normalized[name]...)
		sort.Strings(values)

		fmt.Fprintf(&scope, "%s=%s;", strings.ToLower(name), strings.Join(values, ","))
	}

	return scope.String()
}

func findCachedResult(datasourceUID string, thrukURL string, scope string) *CachedResult {
	for _, result := range cachedResults {
		if result.datasourceUID == datasourceUID && result.thrukURL == thrukURL && result.authScope == scope {
			return result
		}
	}

	return nil
}

func cleanupExpiredResults() {
	cachedResultsMutex.Lock()
	defer cachedResultsMutex.Unlock()

	newCachedresults := make([]*CachedResult, 0)

	now := time.Now()
	for _, cachedResult := range cachedResults {
		if cachedResult.expirationTime.Before(now) {
			continue
		}

		newCachedresults = append(newCachedresults, cachedResult)
	}

	cachedResults = newCachedresults
}

//nolint:gochecknoinits // need a global ticker
func init() {
	go func() {
		const cleanupMinutePeriod = 5

		ticker := time.NewTicker(cleanupMinutePeriod * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			cleanupExpiredResults()
		}
	}()
}

// CachePolicy collections are searched and applied when trying to get, validate and save a cached result.
type CachePolicy struct {
	// policy will only work on these tables, if defined
	filterToTables *[]string
	// policy will only work when these headers are present
	filterToHeaders *[]string
	// requiresAuth marks the policy as unsafe to use without an authentication context.
	// A request with an empty auth scope is only eligible for policies that do not require auth.
	requiresAuth  bool
	cacheDuration time.Duration
}

//nolint:gochecknoglobals // need cache policies to be globally defined
var (
	cachePolicies = []CachePolicy{
		{
			filterToTables:  &[]string{"/", "/index", "/thruk"},
			filterToHeaders: nil,
			requiresAuth:    false,
			cacheDuration:   1 * time.Hour,
		},
		{
			filterToTables:  &[]string{"/users"},
			filterToHeaders: nil,
			requiresAuth:    true,
			cacheDuration:   1 * time.Minute,
		},
		{
			filterToTables:  &[]string{"/sites"},
			filterToHeaders: nil,
			requiresAuth:    true,
			cacheDuration:   1 * time.Minute,
		},
		{
			filterToTables:  nil,
			filterToHeaders: &[]string{"X-Thruk-Output-Metadata-Only"},
			requiresAuth:    true,
			cacheDuration:   1 * time.Hour,
		},
	}
)

func findCachePolicy(qm *QueryModel, authHeaders map[string][]string, scope string) *CachePolicy {
	for _, policy := range cachePolicies {
		if policy.filterToTables != nil &&
			(qm == nil || !slices.Contains(*policy.filterToTables, qm.Table)) {
			continue
		}

		// never serve a policy that requires authentication to a request without an auth context
		if policy.requiresAuth && scope == "" {
			continue
		}

		if policy.filterToHeaders != nil {
			hasRequiredHeader := len(authHeaders) > 0 &&
				slices.ContainsFunc(*policy.filterToHeaders,
					func(e string) bool {
						_, ok := authHeaders[e]

						return ok
					})
			if !hasRequiredHeader {
				continue
			}
		}

		return &policy
	}

	return nil
}

// ErrCouldNotFindCachePolicy error type.
var ErrCouldNotFindCachePolicy = errors.New("could not find cache policy")

// ErrNoCachedResults error type.
var ErrNoCachedResults = errors.New("there are no cached results")

// ErrCachedResultExpired error type.
var ErrCachedResultExpired = errors.New("there is a cached result, but it is expired")

func getCachedResult(queryModel *QueryModel, datasourceUID string, thrukURL string, authHeaders map[string][]string) (*CachedResult, error) {
	cachedResultsMutex.RLock()
	defer cachedResultsMutex.RUnlock()

	scope := authScope(authHeaders)

	cachePolicy := findCachePolicy(queryModel, authHeaders, scope)
	if cachePolicy == nil {
		return nil, fmt.Errorf("%w , table: %s", ErrCouldNotFindCachePolicy, queryModel.Table)
	}

	cachedResult := findCachedResult(datasourceUID, thrukURL, scope)
	if cachedResult == nil {
		return nil, fmt.Errorf("%w , datasourceUID: %s , thrukUrl: %s", ErrNoCachedResults, datasourceUID, thrukURL)
	}

	now := time.Now()
	if cachedResult.expirationTime.Before(now) {
		return nil, fmt.Errorf("%w , cache time : %s, current time: %s", ErrCachedResultExpired, cachedResult.expirationTime.Format(time.RFC3339), now.Format(time.RFC3339))
	}

	return cachedResult, nil
}

func writeCachedResult(queryModel *QueryModel, datasourceUID string, thrukURL string, authHeaders map[string][]string, result *backend.DataResponse) error {
	if result == nil {
		return fmt.Errorf("%w , argument: result", ErrArgumentNil)
	}

	scope := authScope(authHeaders)

	cachePolicy := findCachePolicy(queryModel, authHeaders, scope)
	if cachePolicy == nil {
		return fmt.Errorf("%w , table: %s", ErrCouldNotFindCachePolicy, queryModel.Table)
	}

	cachedResultsMutex.Lock()
	defer cachedResultsMutex.Unlock()

	cachedResults = append(cachedResults, &CachedResult{
		datasourceUID:  datasourceUID,
		thrukURL:       thrukURL,
		authScope:      scope,
		result:         result,
		expirationTime: time.Now().Add(cachePolicy.cacheDuration),
	})

	return nil
}

func rewriteAliasedEndpoints(queryModel *QueryModel) bool {
	// Aliases come from Thruk Docs
	// https://www.thruk.org/documentation/rest.html
	changed := false
	// Convert to the endpoint with lower lexicographical value
	switch queryModel.Table {
	case "/index":
		queryModel.Table = "/"
		changed = true
	case "/thruk/stats":
		queryModel.Table = "/thruk/metrics"
		changed = true
	case "/thruk/node-control/nodes":
		queryModel.Table = "/thruk/nc/nodes"
		changed = true
	}

	return changed
}
