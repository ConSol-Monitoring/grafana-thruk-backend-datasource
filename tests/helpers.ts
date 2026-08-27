import type { Page } from '@playwright/test';

// Sets the "FROM" table combobox in the plugin's query editor to the given table.
// The combobox offers the typed value as a selectable option (either an existing
// table or a "Use custom value" entry), which is clicked to commit the selection.
export async function setFromTable(page: Page, table: string): Promise<void> {
  const fromInput = page.getByTestId('query-editor-from');
  await fromInput.waitFor({ state: 'visible' });
  await fromInput.click();
  await fromInput.fill(table);

  const option = page.getByRole('option').filter({ hasText: table }).first();
  await option.waitFor({ state: 'visible' });
  await option.click();
}
