import { test, expect } from '@grafana/plugin-e2e';

test('provisioned dashboard renders Thruk data', async ({ gotoDashboardPage, readProvisionedDashboard }) => {
  const dashboard = await readProvisionedDashboard({ fileName: 'dashboard.json' });
  const page = await gotoDashboardPage({ uid: dashboard.uid });

  await page.waitForPanelsQueriesToComplete();

  const hostsPanel = page.getPanelByTitle('/hosts');
  await hostsPanel.scrollIntoView();
  await expect(
    hostsPanel.locator.getByRole('columnheader', { name: 'address', exact: true })
  ).toBeVisible({ timeout: 30000 });
});
