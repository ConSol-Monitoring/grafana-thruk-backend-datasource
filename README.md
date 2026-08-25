# Thruk Grafana Datasource - a Grafana backend datasource using Thruks REST API

![Thruk Grafana Backend Datasource](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/src/img/screenshot.png "Thruk Grafana Datasource")

This is an updated version of Thruk Grafana Datasource with Backend Components.

## Installation

Search for `thruk` in the Grafana plugins directory, and be sure to pick the plugin with Backend

Alternatively use the grafana-cli command. The ID of the plugin is consolmonitoring-thruk-datasource

    %> grafana-cli plugins install consolmonitoring-thruk-datasource

Also [OMD-Labs](https://labs.consol.de/omd/) will soon come with this datasource included, so when you use OMD-Labs, the plugin will be already included.

Otherwise follow these steps:

    %> cd var/grafana/plugins
    %> git clone -b release-<tag> https://github.com/ConSol-Monitoring/grafana-thruk-backend-datasource
    %> make build
    %> restart grafana

## Create Datasource

Add a new datasource and select: consolmonitoring-thruk-datasource

Use the Grafana proxy if needed.

- Type 'Thruk'
- Url to Thruk, ex.: <https://localhost/sitename/thruk>

## Table Queries

Using the table panel, you can display most data from the rest api. However only text, numbers and timestamps can be displayed in a sane way.
Support for nested data structures is limited.

Select the rest path from where you want to display data. Then choose all columns. Aggregation functions can be added as well and always affect the column following afterwards.

## Variable Queries

Thruks Rest api can be used to fill grafana variables. For example to get all hosts of a certain hostgroup, use this example query:

    SELECT name FROM hosts WHERE groups >= 'linux'

This is used when picking variables in Dashboard setups automatically.

## Annotation Queries

Annotation queries can be used to add logfile entries into your graphs.
Please note that annotations are shared across all graphs in a dashboard.

It is important to use at least a time filter.

![Annotations](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/img/annotations.png "Annotations Editor")

## Single Stat Queries

Single stats are best used with REST endpoints which return aggregated values
already or use aggregation functions like, `avg`, `sum`, `min`, `max` or `count`.

## Timeseries based panels

Althouth Thruk isn't a timeseries databases und usually only returns table data, some queries can be converted to fake timeseries if the panel cannot handle table data.

This is done according to the preferred visualization type that the panel recommends.

You can either use queries which have 2 columns (name, value) or queries which only return a single result row with numeric values only.

### Statistic Data Pie Chart

For example the pie chart plugin can be used with stats queries like this:

    SELECT count() state, state FROM /hosts

The query is expected to fetch 2 columns. The first is the value, the second is the name.

### Single Host Pie Chart

Ex.: Use statistics data for a single host to put it into a pie chart:

    SELECT num_services_ok, num_services_warn, num_services_crit, num_services_unknown FROM /hosts WHERE name = '$name' LIMIT 1

![Pie Chart](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/img/piechart.png "Pie Chart")

## Using Variables

Dashboard variables can be used in almost all queries. For example if you
define a dashboard variable named `host` you can then use `$host` in your
queries.

There is a special syntax for time filter: `field = $time` which will be
replaced by `(field >= starttime AND field <= endtime)`. This can be used to
reduce results to the dashboards timeframe.

    SELECT time, message FROM /hosts/$host/alerts WHERE time = $time

which is the same as

    SELECT time, message FROM /alerts WHERE host_name = "$host" AND time = $time

![Variables](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/img/variables.png "Variables Editor")

## Development

To test and improve the plugin you can run Grafana instance in Docker using following command (in the source directory of this plugin)

This is a backend-enabled plugin, which means that appropiate go version according to go.mod has to be installed.

  %> make build
  %> make dev

This will start a grafana container and a build watcher which updates the plugin is the dist/ folder.

The dev instance can be accessed at <http://localhost:3000>

Note: You need to add the datasource manually and you need to run "make build" once before starting the dev container, otherwise Grafana won't find the datasource.

After starting the container, you should see one datasource already up, which should be connected to demo.thruk.org:

![Provisioned demo.thruk.org](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/img/provisioned-demo-thruk-org.png)

There is already a provisioned dashboard using this datasource

![Provisioned demo.thruk.org dashboard](![alt text](https://raw.githubusercontent.com/ConSol-Monitoring/grafana-thruk-backend-datasource/master/img/provisioned-demo-thruk-dashboard.png))

The grafana widget documentation is available here: <https://developers.grafana.com/ui/latest/>

More information about the backend plugins are available here: <https://grafana.com/developers/plugin-tools/key-concepts/backend-plugins/>

### Testing

For testing you can use the provisioned demo Thruk instance at:

- URL: <https://demo.thruk.org/demo/thruk/>
- Basic Auth: test / test

In the provisioning/datasources/datasources.yml , there are two commented out example datasources:

First one uses basic Auth to connect to a OMD Thruk instance

Second one uses Thruk API key using X-Thruk-Auth-Key header.

### Create Release

How to create a new release:

    %> export RELVERSION=1.0.7
    %> export GRAFANA_ACCESS_POLICY_TOKEN=...
    %> vi package.json # replace version
    %> vi CHANGELOG.md # add changelog entry
    %> git commit -am "Release v${RELVERSION}"
    %> git tag -a v${RELVERSION} -m "Create release tag v${RELVERSION}"
    %> make GRAFANA_ACCESS_POLICY_TOKEN=${GRAFANA_ACCESS_POLICY_TOKEN} releasebuild
    # create release here https://github.com/ConSol-Monitoring/grafana-thruk-backend-datasource/releases/new
    # submit plugin update here https://grafana.com/orgs/consolmonitoring/plugins

## Changelog

see CHANGELOG.md
