### Changelog

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
