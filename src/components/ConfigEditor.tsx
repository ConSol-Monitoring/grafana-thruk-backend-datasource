import React from 'react';
import { InlineField, Input } from '@grafana/ui';
import {
  ConnectionSettings,
  ConfigSection,
  Auth,
  AdvancedHttpSettings,
  convertLegacyAuthProps,
} from '@grafana/plugin-ui';
import { DataSourcePluginOptionsEditorProps, LogLevel } from '@grafana/data';
import { ThrukDataSourceOptions } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<ThrukDataSourceOptions> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;

  const optionsDefaulted = {
    ...options,
    jsonData: {
      ...options.jsonData,
      logLevel: options.jsonData.logLevel || 0,
      logPath: options.jsonData.logPath || '${OMD_ROOT}/var/log/grafana/consolmonitoring-thruk-datasource.log',
    },
  };

  const onLogLevelChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...options.jsonData,
        logLevel: Number(event.target.value),
      },
    });
  };

  const onLogPathChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...options.jsonData,
        logPath: event.target.value,
      },
    });
  };

  return (
    <>
      <ConnectionSettings config={optionsDefaulted} onChange={props.onOptionsChange} />

      <Auth
        {...convertLegacyAuthProps({
          config: optionsDefaulted,
          onChange: props.onOptionsChange,
        })}
      />

      <ConfigSection title="Advanced settings" isCollapsible isInitiallyOpen={true}>
        <AdvancedHttpSettings config={optionsDefaulted} onChange={props.onOptionsChange} />

        <InlineField
          label="Log Level"
          labelWidth={14}
          interactive
          tooltip={'LogLevel to use for the plugin. Uses syslog(3) style levels, valid levels are [0-7]'}
        >
          <Input
            id="config-editor-log-level"
            onChange={onLogLevelChange}
            value={optionsDefaulted.jsonData.logLevel}
            placeholder="Enter a numeric log level"
            width={40}
          />
        </InlineField>

        <InlineField
          label="Log Path"
          labelWidth={14}
          interactive
          tooltip={'Log Path to use for the plugin. Can specify $HOME or use relative paths.'}
        >
          <Input
            id="config-editor-log-path"
            onChange={onLogPathChange}
            value={optionsDefaulted.jsonData.logPath}
            placeholder="Enter a log path"
            width={40}
          />
        </InlineField>
      </ConfigSection>
    </>
  );
}
