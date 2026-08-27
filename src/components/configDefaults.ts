import { DataSourceSettings } from '@grafana/data';
import { ThrukDataSourceOptions } from '../types';

export const defaultLogPath = '${OMD_ROOT}/var/log/grafana/consolmonitoring-thruk-datasource.log';
export const defaultLogLevel = 0;
export const defaultKeepCookies = ['thruk_auth'];

// applyThrukDefaults merges the plugin's defaults into jsonData so that a saved
// datasource configuration always carries keepCookies=['thruk_auth'], a log
// level and a log path, even when the user never touches those fields.
export function applyThrukDefaults(
  config: DataSourceSettings<ThrukDataSourceOptions>
): DataSourceSettings<ThrukDataSourceOptions> {
  return {
    ...config,
    jsonData: {
      ...config.jsonData,
      logLevel: config.jsonData.logLevel ?? defaultLogLevel,
      logPath: config.jsonData.logPath ?? defaultLogPath,
      keepCookies: config.jsonData.keepCookies ?? defaultKeepCookies,
    },
  };
}
