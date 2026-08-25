package plugin

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type CachedResult struct {
	datasourceUID  string
	thrukUrl       string
	headers        *map[string][]string
	result         *backend.DataResponse
	expirationTime time.Time
}

//nolint:gochecknoglobals // need cache to be globally defined
var (
	cachedResults      = []*CachedResult{}
	cachedResultsMutex sync.RWMutex
)

func findCachedResult(datasourceUID string, thrukUrl string, headers *map[string][]string) *CachedResult {
	for _, result := range cachedResults {
		if result.datasourceUID == datasourceUID && result.thrukUrl == thrukUrl {
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
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cleanupExpiredResults()
		}
	}()
}

type CachePolicy struct {
	// policy will only work on these tables, if defined
	filterToTables *[]string
	// policy will only work when these headers are present
	filterToHeaders *[]string
	cacheDuration   time.Duration
}

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

		if policy.filterToHeaders != nil &&
			headers != nil &&
			len(*headers) > 0 &&
			!slices.ContainsFunc(*policy.filterToHeaders, func(e string) bool { _, ok := (*headers)[e]; return ok }) {
			continue
		}
		return &policy
	}
	return nil
}

func getCachedResult(qm *QueryModel, datasourceUID string, thrukUrl string, headers *map[string][]string) (*CachedResult, error) {
	cachedResultsMutex.RLock()
	defer cachedResultsMutex.RUnlock()

	cachePolicy := findCachePolicy(qm, headers)
	if cachePolicy == nil {
		return nil, fmt.Errorf("Could not find cache policy for query model with table %s", qm.Table)
	}

	cachedResult := findCachedResult(datasourceUID, thrukUrl, headers)
	if cachedResult == nil {
		return nil, fmt.Errorf("There are no cached results for this datasourceUID: %s and thrukUrl: %s", datasourceUID, thrukUrl)
	}

	now := time.Now()
	if cachedResult.expirationTime.Before(now) {
		return nil, fmt.Errorf("Cached result expiration time : %s is expired, current time: %s", cachedResult.expirationTime.Format(time.RFC3339), now.Format(time.RFC3339))
	}

	return cachedResult, nil
}

func writeCachedResult(qm *QueryModel, datasourceUID string, thrukUrl string, headers *map[string][]string, result *backend.DataResponse) error {
	if result == nil {
		return fmt.Errorf("There is no result to write into the cache")
	}

	cachePolicy := findCachePolicy(qm, headers)
	if cachePolicy == nil {
		return fmt.Errorf("Could not find cache policy for query model with table %s", qm.Table)
	}

	cachedResultsMutex.Lock()
	defer cachedResultsMutex.Unlock()

	cachedResults = append(cachedResults, &CachedResult{
		datasourceUID:  datasourceUID,
		thrukUrl:       thrukUrl,
		headers:        headers,
		result:         result,
		expirationTime: time.Now().Add(cachePolicy.cacheDuration),
	})

	return nil
}

func rewriteAliasedEndpoints(qm *QueryModel) (changed bool) {
	// Aliases come from Thruk Docs
	// https://www.thruk.org/documentation/rest.html
	changed = false
	// Convert to the endpoint with lower lexicographical value
	switch qm.Table {
	case "/index":
		qm.Table = "/"
		changed = true
	case "/thruk/stats":
		qm.Table = "/thruk/metrics"
		changed = true
	case "/thruk/node-control/nodes":
		qm.Table = "/thruk/nc/odes"
		changed = true
	}

	return changed
}
