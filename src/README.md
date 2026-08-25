# Thruk Datasource with Backend

A Grafana backend datasource plugin that queries [Thruk's REST API](https://www.thruk.org/documentation/rest.html) to display monitoring data from OMD / Thruk instances in Grafana.

This is the backend-enabled successor of the frontend-only Thruk datasource. It runs the queries against Thruk in a Go backend process, which allows caching, per-user authentication through forwarded cookies and better performance.

## Requirements

- Grafana 12.4 or newer
- A Thruk instance with the REST API enabled (`/r/v1/`)

## Installation

Search for `thruk` in the Grafana plugins directory and pick the plugin with backend, or use the grafana-cli:

```
grafana-cli plugins install consolmonitoring-thruk-datasource
```

There are two versions that could come up when searching for Thruk, be sure to install the one that has backend components with the id consolmonitoring-thruk-datasource and not sni-thruk-datasource.

## Create Datasource

Add a new datasource and select `consolmonitoring-thruk-datasource`.

- Set the URL to your Thruk instance, ex. for OMD Thruk `https://<host>/<sitename>/thruk`
- Configure authentication (basic auth or a Thruk API key) as needed
- For Grafana OMD usage, cookie authentication through existing thruk_auth works as well.
- Log Level and Log Path can be adjusted in the advanced settings

## Query Types

### Table Queries

Using a table panel, you can display most data from the REST API. Only text, numbers and timestamps can be displayed in a sane way, support for nested data structures is limited.

Select the REST path from where you want to display data, then choose all columns. Aggregation functions can be added as well and always affect the column following afterwards.

### Variable Queries

Thruk's REST API can be used to fill Grafana variables. For example to get all hosts of a certain hostgroup, use this query:

```
SELECT name FROM hosts WHERE groups >= 'linux'
```

### Annotation Queries

Annotation queries can be used to add logfile entries into your graphs. Please note that annotations are shared across all graphs in a dashboard. It is important to use at least a time filter.

### Timeseries based panels

Although Thruk is not a time series database and usually only returns table data, some queries can be converted to fake time series if the panel cannot handle table data. This is done according to the preferred visualization type that the panel recommends.

You can either use queries which have 2 columns (name, value) or queries which only return a single result row with numeric values only.

## Using Variables

Dashboard variables can be used in almost all queries. For example if you define a dashboard variable named `host` you can then use `$host` in your queries.

There is a special syntax for a time filter: `field = $time` which will be replaced by `(field >= starttime AND field <= endtime)`. This can be used to reduce results to the dashboard's timeframe:

```
SELECT time, message FROM /alerts WHERE host_name = "$host" AND time = $time
```

## Documentation

More information about backend plugins is available at the [Grafana plugin tools](https://grafana.com/developers/plugin-tools/key-concepts/backend-plugins/).

## Development

The source and development instructions live in the [GitHub repository](https://github.com/ConSol-Monitoring/grafana-thruk-backend-datasource).

## Changelog

See [CHANGELOG.md](https://github.com/ConSol-Monitoring/grafana-thruk-backend-datasource/blob/master/CHANGELOG.md).
