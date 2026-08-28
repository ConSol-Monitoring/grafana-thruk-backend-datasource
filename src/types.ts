import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface ThrukQuery extends DataQuery {
  table: string;
  columns: string[];
  condition: string;
  limit: number;
  type: 'table' | 'graph' | 'logs' | 'timeseries';

  // metadata injected by the frontend for backend logging/auditing
  dashboardUID?: string;
  dashboardTitle?: string;
  panelId?: number;
  panelName?: string;
  panelPluginId?: string;
  app?: string;
  requestUrl?: string;
}

export const defaultQuery: Partial<ThrukQuery> = {
  table: '/',
  columns: ['*'],
  condition: '',
  type: 'table',
};

export interface ThrukDataSourceOptions extends DataSourceJsonData {
  keepCookies?: string[];
}
