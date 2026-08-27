import { DataSourceSettings } from '@grafana/data';
import { ThrukDataSourceOptions } from '../types';

export const defaultLogLevel = 0;
export const defaultKeepCookies = ['thruk_auth'];

// applyThrukDefaults merges the plugin's defaults into jsonData so that a saved
// datasource configuration always carries keepCookies=['thruk_auth'] and a log
// level, even when the user never touches those fields. A log path is intentionally
// not defaulted: logging goes to Grafana's own pipeline unless the user configures
// an explicit file path.
export function applyThrukDefaults(
  config: DataSourceSettings<ThrukDataSourceOptions>
): DataSourceSettings<ThrukDataSourceOptions> {
  return {
    ...config,
    jsonData: {
      ...config.jsonData,
      logLevel: config.jsonData.logLevel ?? defaultLogLevel,
      keepCookies: config.jsonData.keepCookies ?? defaultKeepCookies,
    },
  };
}
