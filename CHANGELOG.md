### Changelog

0.7.1    27-08-2026
    - fix formatting using prettier
    - fix version definitions in plugin.json and package.json

0.7.0    27-08-2026
    - scope the response cache to the authenticated request, preventing cross-user cache hits and fixing a nil-pointer panic on headerless lookups
    - forward only an allow-list of authentication headers (Cookie, Authorization, X-Id-Token) to Thruk instead of all forwarded headers
    - bound the response cache to 500 entries with FIFO eviction and evict a datasource's entries when it is disposed
    - log through Grafana's own backend logging pipeline by default, the logfile is an optional secondary log-to file target
    - write a rotated log file only when a log path is explicitly configured
    - fix log-file descriptor leaks and add flushing/closing when datasources are reconfigured or removed
    - report an unwritable log path with an actionable error naming the resolved path
    - persist the thruk_auth allowed-cookie default so cookie authentication works without manual configuration
    - add annotation support in the frontend, plugin.json was declaring it but datasource constructor did not set this.annotations = {}
    - add E2E coverage for queries, variables, annotations, alerting, and authentication
    - declare directly imported frontend dependencies (@hello-pangea/dnd, lodash, rxjs)
    - reject path traversal in Thruk API paths, previously it was using filepath to evalute which could add ../../
    - correct the Thruk node-control endpoint alias to /thruk/nc/nodes
    - return resource-sender errors correctly
    - declare the plugin MIT licensed, the plugin.json had GLP2 mistakenly
    - set the minimum supported Grafana version to 12.4
    - update screenshots, the provisioned demo dashboard, README, and plugin metadata
    - add provisioned alerting rules

0.6.0    26-08-2026
    - infer numeric table-column types from response data when Thruk metadata does not specify a type
    - fix header-filtered cache policies when required headers are absent
    - fix cache-policy headers during cache lookups
    - validate Thruk API paths for queries and resource requests, rejecting malformed paths and paths that escape the API root
    - limit Thruk API response bodies to 100 MB
    - cancel pending debounced queries when the query editor is replaced or unmounted
    - assign unique IDs to the log level and log path configuration inputs

0.5.0     26-08-2026
    - pre release version, code taken from github.com/inqrphl/grafana-thruk-datasource/tree/add-backend-support with backend support added.
    - clean up the code more, and adjust references to the new plugin ID
    - add golangci
