# Thruk Datasource with Backend

![Thruk Datasource query editor](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/src/img/screenshot.png)

Thruk Datasource with Backend connects Grafana to the [Thruk](https://www.thruk.org/) REST API. Use it to query and visualize monitoring data from OMD and Thruk instances in tables, stat and pie-chart panels, dashboard variables, and annotations. The plugin includes a backend component and supports Grafana alerting.

## Requirements

- Grafana 12.4.0 or later.
- A reachable Thruk or OMD instance with its REST API enabled.
- Credentials that are permitted to query the target Thruk instance. The datasource supports basic authentication, `thruk_auth` based cookie authentication and API key authentication using the `X-Thruk-Auth-Key` header.

## Getting Started

1. In Grafana, open **Connections** > **Add new connection**, search for `thruk`, and select **Thruk Datasource with Backend**.
2. Enter the base URL of your Thruk instance, for example `https://example.com/sitename/thruk` for an OMD site.
3. Configure the authentication required by that instance, then select **Save & test**.
4. Create a panel and select the datasource. Choose a REST path and the columns to display, or enter a Thruk query.

To install from the command line, run:

```sh
grafana cli plugins install consolmonitoring-thruk-datasource
```

Restart Grafana after installing the plugin. The plugin ships separately from OMD currently, so it is not included in [OMD-Labs](https://labs.consol.de/omd/) Grafana installations.

## Querying Thruk

### Table queries

Use a table panel to display data returned by the REST API. Text, numbers, and timestamps are supported. Support for nested data structures is limited.

Choose the REST path to query, then select the columns to display. Aggregation functions can be added and affect the following column.

### Variable queries

Use Thruk REST API queries to populate Grafana dashboard variables. For example, to return the hosts in a host group:

```sql
SELECT name FROM hosts WHERE groups >= 'linux'
```

### Annotation queries

Use annotation queries to add log-file entries to graphs. Annotations are shared by all panels in a dashboard. Include a time filter in every annotation query.

![Annotations editor](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/img/annotations.png)

### Stat and pie-chart queries

Use REST endpoints that return aggregated values, or aggregation functions such as `avg`, `sum`, `min`, `max`, and `count`, with stat panels.

For pie charts, use either a two-column result (`name`, `value`) or a single row containing numeric values. For example:

```sql
SELECT count() state, state FROM /hosts
```

![Pie chart](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/img/piechart.png)

### Dashboard variables and time filters

Dashboard variables can be used in queries. For a variable named `host`, reference it as `$host`:

```sql
SELECT time, message FROM /hosts/$host/alerts WHERE time = $time
```

The special form `field = $time` is expanded to `(field >= starttime AND field <= endtime)`, restricting results to the dashboard time range. The preceding query is equivalent to:

```sql
SELECT time, message FROM /alerts WHERE host_name = "$host" AND time = $time
```

![Variables editor](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/img/variables.png)

## Documentation

- [Thruk project documentation](https://www.thruk.org/documentation.html)
- [Grafana data source documentation](https://grafana.com/docs/grafana/latest/datasources/)
- [Plugin repository](https://github.com/ConSol-Monitoring/grafana-thruk-backend-datasource)

## Development

For a local source installation, clone the repository and build the plugin:

```sh
git clone https://github.com/ConSol-Monitoring/grafana-thruk-backend-datasource.git
cd grafana-thruk-backend-datasource
make build
make dev
```

Development requires Docker, Node.js 24 for the containerized frontend build, and Go 1.27 or later for the backend build. The development Grafana instance is available at [http://localhost:3000](http://localhost:3000) after running `make dev`. Run `make build` once before starting the development container, then add the datasource manually.

The development environment provisions a datasource connected to `demo.thruk.org` and a dashboard using it.

![Provisioned demo datasource](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/img/provisioned-demo-thruk-org.png)

![Provisioned demo dashboard](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/img/provisioned-demo-thruk-dashboard.png)

For testing, the a public Thruk demo instance is available at [https://demo.thruk.org/demo/thruk/](https://demo.thruk.org/demo/thruk/) with basic-auth credentials `test` / `test`. The `provisioning/datasources/datasources.yml` file also contains commented examples for basic authentication and the `X-Thruk-Auth-Key` header to start off as well.

### Maintainer release process

Update the version in `package.json` and `CHANGELOG.md`, commit and tag the release, then build and sign the archive:

```sh
export RELVERSION=0.7.2
export GRAFANA_ACCESS_POLICY_TOKEN=...
git commit -am "Release v${RELVERSION}"
git tag -a "v${RELVERSION}" -m "Create release tag v${RELVERSION}"
make GRAFANA_ACCESS_POLICY_TOKEN="${GRAFANA_ACCESS_POLICY_TOKEN}" releasebuild
```

Create the GitHub release and submit the plugin update through the [Grafana Plugins Admin page](https://grafana.com/orgs/consolmonitoring/plugins).

## Contributing and support

Contributions and bug reports are welcome. Please open an issue or pull request in the [GitHub repository](https://github.com/ConSol-Monitoring/grafana-thruk-backend-datasource). Include your Grafana version, Thruk version, datasource configuration with secrets removed, and steps to reproduce any issue.

## License

This project is licensed under the [MIT License](https://github.com/ConSol-Monitoring/grafana-thruk-backend-datasource/blob/master/LICENSE).

## Changelog

See [CHANGELOG.md](https://github.com/ConSol-Monitoring/grafana-thruk-backend-datasource/blob/master/CHANGELOG.md).
