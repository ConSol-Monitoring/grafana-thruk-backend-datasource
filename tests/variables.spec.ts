import { test, expect } from '@grafana/plugin-e2e';

test('variable query returns host names', async ({ variableEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });

  // A new variable defaults to "Query" type.
  await variableEditPage.datasource.set(ds.name);

  const queryInput = variableEditPage.getByGrafanaSelector(
    variableEditPage.ctx.selectors.pages.Dashboard.Settings.Variables.Edit.QueryVariable.queryOptionsQueryInput
  );
  await queryInput.waitFor({ state: 'visible' });

  const responsePromise = page.waitForResponse((resp) => resp.url().includes('/resources/variable-query'), {
    timeout: 60000,
  });
  await queryInput.click();
  await queryInput.pressSequentially('SELECT name FROM hosts LIMIT 10');
  await variableEditPage.runQuery();
  const response = await responsePromise;

  expect(response.ok()).toBeTruthy();

  const body = await response.json();
  expect(Array.isArray(body)).toBeTruthy();
  expect(body.length).toBeGreaterThan(0);
});
