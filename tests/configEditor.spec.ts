import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, page }) => {
  await createDataSourceConfigPage({ type: 'consolmonitoring-thruk-datasource' });
  await expect(page.getByLabel('URL')).toBeVisible();
  await expect(page.getByPlaceholder('Enter a numeric log level')).toBeVisible();
  await expect(page.getByPlaceholder('Enter a log path')).toBeVisible();
});

test('"Save & test" should be successful when the provisioned Thruk instance is reachable', async ({
  gotoDataSourceConfigPage,
}) => {
  const configPage = await gotoDataSourceConfigPage('demo-thruk-org');
  await expect(configPage.saveAndTest()).toBeOK();
});

test('"Save & test" should fail when no URL is configured', async ({ createDataSourceConfigPage }) => {
  const configPage = await createDataSourceConfigPage({ type: 'consolmonitoring-thruk-datasource' });
  const resp = await configPage.saveAndTest();
  const body = await resp.json();
  expect(body.status).not.toBe('OK');
});
