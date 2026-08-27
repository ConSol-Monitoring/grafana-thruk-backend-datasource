import { test, expect } from '@grafana/plugin-e2e';

// Creates a real Grafana alert rule against the provisioned Thruk datasource and
// waits for the server-side scheduler to evaluate it without errors. This runs
// without browser cookies, exercising the backend alerting query path directly.
test('alert rule evaluates against the Thruk datasource', async ({ request, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });

  const folderTitle = `thruk-e2e-alerting-${Date.now()}`;
  const ruleTitle = 'thruk e2e alert';

  const folderResponse = await request.post('/api/folders', { data: { title: folderTitle } });
  expect(folderResponse.ok()).toBeTruthy();
  const folder = await folderResponse.json();

  const ruleResponse = await request.post('/api/v1/provisioning/alert-rules', {
    data: {
      title: ruleTitle,
      condition: 'C',
      data: [
        {
          refId: 'A',
          queryType: '',
          relativeTimeRange: { from: 300, to: 0 },
          datasourceUid: ds.uid,
          model: {
            table: '/hosts',
            columns: ['count()'],
            condition: '',
            limit: 1000,
            type: 'graph',
            refId: 'A',
          },
        },
        {
          refId: 'B',
          queryType: '',
          datasourceUid: '__expr__',
          model: {
            refId: 'B',
            type: 'reduce',
            expression: 'A',
            reducer: 'last',
            settings: { mode: 'dropNN' },
          },
        },
        {
          refId: 'C',
          queryType: '',
          datasourceUid: '__expr__',
          model: {
            refId: 'C',
            type: 'threshold',
            expression: 'B',
            conditions: [
              {
                evaluator: { params: [100], type: 'gt' },
                query: { params: ['B'] },
                reducer: { params: [], type: 'last' },
                type: 'query',
              },
            ],
          },
        },
      ],
      noDataState: 'NoData',
      execErrState: 'Error',
      for: '1m',
      orgID: 1,
      folderUID: folder.uid,
      ruleGroup: 'thruk-e2e',
      isPaused: false,
    },
  });
  expect(ruleResponse.ok()).toBeTruthy();
  const rule = await ruleResponse.json();

  try {
    let lastError: string | null = 'rule did not reach a terminal state';

    for (let attempt = 0; attempt < 30; attempt++) {
      // The scheduler evaluates rule groups roughly once a minute.
      await new Promise((resolve) => setTimeout(resolve, 5000));

      const rulesResponse = await request.get('/api/prometheus/grafana/api/v1/rules');
      expect(rulesResponse.ok()).toBeTruthy();
      const payload = await rulesResponse.json();

      const evaluatedRule = payload.data.groups
        .flatMap((group: { rules: Array<{ name: string; health?: string; lastError?: string | null }> }) => group.rules)
        .find((candidate: { name: string }) => candidate.name === ruleTitle);

      if (evaluatedRule && evaluatedRule.health) {
        lastError = evaluatedRule.lastError ?? null;
        expect(evaluatedRule.health).toBe('ok');
        break;
      }
    }

    expect(lastError).toBeNull();
  } finally {
    await request.delete(`/api/v1/provisioning/alert-rules/${rule.uid}`);
    await request.delete(`/api/folders/${folder.uid}`);
  }
});
