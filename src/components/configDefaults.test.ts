import { DataSourceSettings } from '@grafana/data';
import { ThrukDataSourceOptions } from '../types';
import { applyThrukDefaults, defaultKeepCookies, defaultLogLevel, defaultLogPath } from './configDefaults';

function baseConfig(): DataSourceSettings<ThrukDataSourceOptions> {
  return {
    id: 1,
    uid: 'test-uid',
    orgId: 1,
    name: 'thruk',
    typeLogoUrl: '',
    type: 'consolmonitoring-thruk-datasource',
    typeName: 'Thruk',
    access: 'proxy',
    url: 'https://thruk.example.com',
    user: '',
    database: '',
    basicAuth: false,
    basicAuthUser: '',
    isDefault: false,
    jsonData: {},
    secureJsonFields: {},
    readOnly: false,
    withCredentials: false,
  };
}

test('applyThrukDefaults fills keepCookies, logLevel and logPath', () => {
  const result = applyThrukDefaults(baseConfig());

  expect(result.jsonData.keepCookies).toEqual(defaultKeepCookies);
  expect(result.jsonData.logLevel).toBe(defaultLogLevel);
  expect(result.jsonData.logPath).toBe(defaultLogPath);
});

test('applyThrukDefaults preserves explicit values including empty keepCookies', () => {
  const result = applyThrukDefaults({
    ...baseConfig(),
    jsonData: { keepCookies: [], logLevel: 7, logPath: '/custom.log' },
  });

  expect(result.jsonData.keepCookies).toEqual([]);
  expect(result.jsonData.logLevel).toBe(7);
  expect(result.jsonData.logPath).toBe('/custom.log');
});

test('applyThrukDefaults preserves a zero log level', () => {
  const result = applyThrukDefaults({
    ...baseConfig(),
    jsonData: { logLevel: 0 },
  });

  expect(result.jsonData.logLevel).toBe(0);
});
