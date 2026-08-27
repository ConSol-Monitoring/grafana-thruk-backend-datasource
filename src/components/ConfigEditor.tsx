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
import { applyThrukDefaults } from './configDefaults';

interface Props extends DataSourcePluginOptionsEditorProps<ThrukDataSourceOptions> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;

  // Persist the plugin defaults on every change so the saved datasource
  // configuration always carries keepCookies=['thruk_auth'].
  const onChangeOptions = (config: typeof options) => {
    onOptionsChange(applyThrukDefaults(config));
  };

  const optionsDefaulted = applyThrukDefaults(options);

  const onLogLevelChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChangeOptions({
      ...options,
      jsonData: {
        ...options.jsonData,
        logLevel: Number(event.target.value),
      },
    });
  };

  const onLogPathChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChangeOptions({
      ...options,
      jsonData: {
        ...options.jsonData,
        logPath: event.target.value,
      },
    });
  };

  return (
    <>
      <ConnectionSettings config={optionsDefaulted} onChange={onChangeOptions} />

      <Auth
        {...convertLegacyAuthProps({
          config: optionsDefaulted,
          onChange: onChangeOptions,
        })}
      />

      <ConfigSection title="Advanced settings" isCollapsible isInitiallyOpen={true}>
        <AdvancedHttpSettings config={optionsDefaulted} onChange={onChangeOptions} />

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
          tooltip={'Optional log file path. Leave empty to log only to Grafana. Can specify $HOME or use relative paths.'}
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
