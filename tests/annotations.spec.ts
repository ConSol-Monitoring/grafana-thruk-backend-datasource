import { test, expect } from '@grafana/plugin-e2e';
import { setFromTable } from './helpers';

test('annotation query against /log returns events', async ({ annotationEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await annotationEditPage.datasource.set(ds.name);

  // Point the query at the Thruk logfile endpoint, which returns time and message columns.
  await setFromTable(page, '/log');

  const response = await annotationEditPage.runQuery({ timeout: 30000 });
  expect(response.ok()).toBeTruthy();
});
