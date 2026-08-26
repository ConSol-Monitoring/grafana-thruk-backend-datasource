package plugin

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// CachedResult represents a cached result, has its own expiration time.
type CachedResult struct {
	datasourceUID  string
	thrukURL       string
	headers        *map[string][]string
	result         *backend.DataResponse
	expirationTime time.Time
}

//nolint:gochecknoglobals // need cache to be globally defined
var (
	cachedResults      = []*CachedResult{}
	cachedResultsMutex sync.RWMutex
)

//nolint:nestif
func findCachedResult(datasourceUID string, thrukURL string, headers *map[string][]string) *CachedResult {
	for _, result := range cachedResults {
		if result.datasourceUID == datasourceUID && result.thrukURL == thrukURL {
			headersMatch := true

			if headers != nil {
				if len(*headers) != len(*result.headers) {
					continue
				}

				for header, value := range *headers {
					resultValue, ok := (*result.headers)[header]

					if !ok {
						headersMatch = false

						break
					}

					if !slices.Equal(value, resultValue) {
						headersMatch = false

						break
					}
				}
			}

			if headersMatch {
				return result
			}
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
	cacheDuration   time.Duration
}

//nolint:gochecknoglobals // need cache policies to be globally defined
var (
	cachePolicies = []CachePolicy{
		{
			&[]string{"/", "/index", "/thruk"},
			nil,
			1 * time.Hour,
		},
		{
			&[]string{"/users"},
			nil,
			1 * time.Minute,
		},
		{
			&[]string{"/sites"},
			nil,
			1 * time.Minute,
		},
		{
			nil,
			&[]string{"X-Thruk-Output-Metadata-Only"},
			1 * time.Hour,
		},
	}
)

func findCachePolicy(qm *QueryModel, headers *map[string][]string) *CachePolicy {
	for _, policy := range cachePolicies {
		if policy.filterToTables != nil &&
			(qm == nil || !slices.Contains(*policy.filterToTables, qm.Table)) {
			continue
		}

		if policy.filterToHeaders != nil {
			hasRequiredHeader := headers != nil && len(*headers) > 0 &&
				slices.ContainsFunc(*policy.filterToHeaders,
					func(e string) bool {
						_, ok := (*headers)[e]

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

func getCachedResult(queryModel *QueryModel, datasourceUID string, thrukURL string, headers *map[string][]string) (*CachedResult, error) {
	cachedResultsMutex.RLock()
	defer cachedResultsMutex.RUnlock()

	cachePolicy := findCachePolicy(queryModel, headers)
	if cachePolicy == nil {
		return nil, fmt.Errorf("%w , table: %s", ErrCouldNotFindCachePolicy, queryModel.Table)
	}

	cachedResult := findCachedResult(datasourceUID, thrukURL, headers)
	if cachedResult == nil {
		return nil, fmt.Errorf("%w , datasourceUID: %s , thrukUrl: %s", ErrNoCachedResults, datasourceUID, thrukURL)
	}

	now := time.Now()
	if cachedResult.expirationTime.Before(now) {
		return nil, fmt.Errorf("%w , cache time : %s, current time: %s", ErrCachedResultExpired, cachedResult.expirationTime.Format(time.RFC3339), now.Format(time.RFC3339))
	}

	return cachedResult, nil
}

func writeCachedResult(queryModel *QueryModel, datasourceUID string, thrukURL string, headers *map[string][]string, result *backend.DataResponse) error {
	if result == nil {
		return fmt.Errorf("%w , argument: result", ErrArgumentNil)
	}

	cachePolicy := findCachePolicy(queryModel, headers)
	if cachePolicy == nil {
		return fmt.Errorf("%w , table: %s", ErrCouldNotFindCachePolicy, queryModel.Table)
	}

	cachedResultsMutex.Lock()
	defer cachedResultsMutex.Unlock()

	cachedResults = append(cachedResults, &CachedResult{
		datasourceUID:  datasourceUID,
		thrukURL:       thrukURL,
		headers:        headers,
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
		queryModel.Table = "/thruk/nc/odes"
		changed = true
	}

	return changed
}
