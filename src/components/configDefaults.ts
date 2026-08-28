import { DataSourceSettings } from '@grafana/data';
import { ThrukDataSourceOptions } from '../types';

export const defaultKeepCookies = ['thruk_auth'];

// applyThrukDefaults merges the plugin's defaults into jsonData so that a saved
// datasource configuration always carries keepCookies=['thruk_auth'], even when
// the user never touches that field.
export function applyThrukDefaults(
  config: DataSourceSettings<ThrukDataSourceOptions>
): DataSourceSettings<ThrukDataSourceOptions> {
  return {
    ...config,
    jsonData: {
      ...config.jsonData,
      keepCookies: config.jsonData.keepCookies ?? defaultKeepCookies,
    },
  };
}
