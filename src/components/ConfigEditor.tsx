import React from 'react';
import {
  ConnectionSettings,
  ConfigSection,
  Auth,
  AdvancedHttpSettings,
  convertLegacyAuthProps,
} from '@grafana/plugin-ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
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
      </ConfigSection>
    </>
  );
}
