import { test, expect } from '@grafana/plugin-e2e';
import { setFromTable } from './helpers';

// The provisioned demo-thruk-org datasource authenticates against demo.thruk.org
// with basic auth (test/test). This test proves the end-to-end authentication path
// by executing a real query through the backend.
//
// Cookie-only OMD authentication and the X-Thruk-Auth-Key header cannot be exercised
// against the public demo and are documented as local-OMD manual tests.
test('basic-auth datasource executes a query without errors', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);

  // Point the query at the /hosts endpoint. Setting the table triggers a real
  // backend query through the provisioned basic-auth datasource.
  const responsePromise = panelEditPage.waitForQueryDataResponse();
  await setFromTable(page, '/hosts');

  const response = await responsePromise;
  expect(response.ok()).toBeTruthy();
  await expect(panelEditPage.panel.getErrorIcon()).not.toBeVisible();
});
