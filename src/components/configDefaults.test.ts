import { DataSourceSettings } from '@grafana/data';
import { ThrukDataSourceOptions } from '../types';
import { applyThrukDefaults, defaultKeepCookies } from './configDefaults';

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

test('applyThrukDefaults fills keepCookies', () => {
  const result = applyThrukDefaults(baseConfig());

  expect(result.jsonData.keepCookies).toEqual(defaultKeepCookies);
});

test('applyThrukDefaults preserves explicit values including empty keepCookies', () => {
  const result = applyThrukDefaults({
    ...baseConfig(),
    jsonData: { keepCookies: [] },
  });

  expect(result.jsonData.keepCookies).toEqual([]);
});
