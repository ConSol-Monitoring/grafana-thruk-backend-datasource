import { DataQueryRequest, DataSourceInstanceSettings } from '@grafana/data';
import { DataSource } from 'datasource';
import { ThrukDataSourceOptions, ThrukQuery } from './types';

const mockSettings: DataSourceInstanceSettings<ThrukDataSourceOptions> = {
  id: 1,
  uid: 'test-uid',
  type: 'consolmonitoring-thruk-datasource',
  name: 'thruk',
  url: 'https://thruk.example.com',
  access: 'proxy',
  jsonData: {},
  readOnly: false,
  meta: {} as any,
};

test('parse variables query', async () => {
  const ds = new DataSource(mockSettings);
  const result = await ds.parseVariableQuery('SELECT name from /hosts WHERE name like ^abc LIMIT 137');

  expect(result.table).toEqual('/hosts');
  expect(result.columns).toEqual(['name']);
  expect(result.condition).toEqual('name like ^abc');
  expect(result.limit).toEqual(137);
});

test('injectQueryMetadata copies dashboard/panel context into target', () => {
  const ds = new DataSource(mockSettings);
  const target: ThrukQuery = {
    refId: 'A',
    table: '/hosts',
    columns: ['*'],
    condition: '',
    limit: 1000,
    type: 'table',
  };
  const request = {
    app: 'dashboard',
    dashboardUID: 'abc123',
    dashboardTitle: 'My Dashboard',
    panelId: 42,
    panelName: 'Hosts Panel',
    panelPluginId: 'table',
  } as DataQueryRequest<ThrukQuery>;

  ds.injectQueryMetadata(target, request);

  expect(target.app).toEqual('dashboard');
  expect(target.dashboardUID).toEqual('abc123');
  expect(target.dashboardTitle).toEqual('My Dashboard');
  expect(target.panelId).toEqual(42);
  expect(target.panelName).toEqual('Hosts Panel');
  expect(target.panelPluginId).toEqual('table');
  expect(target.requestUrl).toBeDefined();
});

test('injectQueryMetadata does not overwrite existing requestUrl', () => {
  const ds = new DataSource(mockSettings);
  const target: ThrukQuery = {
    refId: 'A',
    table: '/hosts',
    columns: ['*'],
    condition: '',
    limit: 1000,
    type: 'table',
    requestUrl: '/d/known/url',
  };
  const request = { app: 'explore' } as DataQueryRequest<ThrukQuery>;

  ds.injectQueryMetadata(target, request);

  expect(target.app).toEqual('explore');
  expect(target.requestUrl).toEqual('/d/known/url');
});
