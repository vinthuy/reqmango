import { test, expect, type APIRequestContext } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'
const TEST_PREFIX = 'e2ereport' + Date.now()

let _token = ''
let _projectId = 0

async function ensureSetup(request: APIRequestContext) {
  if (_token) return
  const user = { email: TEST_PREFIX + '@t.com', username: TEST_PREFIX, password: 'E2eTest123!', display_name: 'E2E Report' }
  await request.post(BASE_API + '/auth/register', { data: user })
  const res = await request.post(BASE_API + '/auth/login', { data: { email: user.email, password: user.password } })
  const { access_token } = await res.json()
  _token = access_token
  const ws = await request.post(BASE_API + '/workspaces', { data: { name: 'E2E Report WS', slug: 'e2e-rpt-' + TEST_PREFIX }, headers: { Authorization: 'Bearer ' + _token } })
  const wsData = await ws.json()
  const wsId = wsData.id || wsData.data?.id
  const proj = await request.post(BASE_API + '/projects?workspace_id=' + wsId, { data: { name: 'E2E Report Proj', identifier: 'E2ERPT', description: 'Report testing' }, headers: { Authorization: 'Bearer ' + _token } })
  const projData = await proj.json()
  _projectId = projData.id || projData.data?.id
}

function auth() { return { Authorization: 'Bearer ' + _token } }

test.describe('Report - Generate API', () => {
  test.beforeEach(async ({ request }) => await ensureSetup(request))

  test('API: distribution by state', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
    const d = await res.json()
    expect(d).toHaveProperty('labels')
    expect(d).toHaveProperty('values')
    expect(d).toHaveProperty('total')
  })

  test('API: distribution by priority', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'priority', chart: 'pie' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
    const d = await res.json()
    expect(d.group_by).toBe('priority')
  })

  test('API: distribution by assignee', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'assignee', chart: 'bar' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: distribution by type', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'type', chart: 'doughnut' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: created_trend', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'created_trend', group_by: 'state', chart: 'line', interval: 'day' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
    const d = await res.json()
    expect(d.type).toBe('created_trend')
  })

  test('API: created_vs_resolved', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'created_vs_resolved', chart: 'bar', interval: 'week' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
    const d = await res.json()
    expect(d.type).toBe('created_vs_resolved')
    expect(d).toHaveProperty('labels')
    expect(d).toHaveProperty('values')
  })

  test('API: avg_age', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'avg_age', group_by: 'state', chart: 'bar' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
    const d = await res.json()
    expect(d.type).toBe('avg_age')
    expect(d).toHaveProperty('labels')
    expect(d).toHaveProperty('values')
  })

  test('API: current_age', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'current_age', group_by: 'priority', chart: 'bar' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })
})

test.describe('Report - RQL Operators', () => {
  test.beforeEach(async ({ request }) => await ensureSetup(request))

  test('API: RQL = operator', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'state = "Todo"' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL != operator', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'state != "Done"' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL IN operator', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'priority', chart: 'pie', rql: 'priority IN ("urgent", "high")' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL NOT IN operator', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'state NOT IN ("Done")' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL LIKE operator', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'name LIKE "test"' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL NOT LIKE operator', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'name NOT LIKE "debug"' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL >= operator', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'start_date >= "2024-01-01"' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL <= operator', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'target_date <= "2024-12-31"' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL IS NULL', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'assignee IS NULL' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL IS NOT NULL', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'assignee IS NOT NULL' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL combined AND', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'priority IN ("urgent") AND state != "Done"' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })

  test('API: RQL date range', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart: 'bar', rql: 'created_at >= "2024-01-01" AND created_at <= "2024-12-31"' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
  })
})

test.describe('Report - Saved Reports CRUD', () => {
  test.beforeEach(async ({ request }) => await ensureSetup(request))

  test('API: create saved report', async ({ request }) => {
    const res = await request.post(BASE_API + '/projects/' + _projectId + '/saved-reports', {
      data: { name: 'Test Report', report_type: 'distribution', group_by: 'state', chart_type: 'bar', rql: 'priority = "high"', date_from: '2024-01-01', date_to: '2024-12-31' },
      headers: auth(),
    })
    expect(res.ok()).toBeTruthy()
    const d = await res.json()
    expect(d.name).toBe('Test Report')
    expect(d.report_type).toBe('distribution')
    expect(d.rql).toBe('priority = "high"')
  })

  test('API: list saved reports', async ({ request }) => {
    const res = await request.get(BASE_API + '/projects/' + _projectId + '/saved-reports', { headers: auth() })
    expect(res.ok()).toBeTruthy()
    const d = await res.json()
    expect(Array.isArray(d)).toBeTruthy()
  })

  test('API: update saved report', async ({ request }) => {
    const c = await request.post(BASE_API + '/projects/' + _projectId + '/saved-reports', { data: { name: 'To Update', report_type: 'distribution', group_by: 'priority', chart_type: 'pie' }, headers: auth() })
    const created = await c.json()
    const res = await request.patch(BASE_API + '/projects/' + _projectId + '/saved-reports/' + created.id, { data: { name: 'Updated', chart_type: 'doughnut' }, headers: auth() })
    expect(res.ok()).toBeTruthy()
    const d = await res.json()
    expect(d.name).toBe('Updated')
    expect(d.chart_type).toBe('doughnut')
  })

  test('API: delete saved report', async ({ request }) => {
    const c = await request.post(BASE_API + '/projects/' + _projectId + '/saved-reports', { data: { name: 'To Delete', report_type: 'distribution', group_by: 'state', chart_type: 'bar' }, headers: auth() })
    const created = await c.json()
    const res = await request.delete(BASE_API + '/projects/' + _projectId + '/saved-reports/' + created.id, { headers: auth() })
    expect(res.ok()).toBeTruthy()
  })
})

test.describe('Report - Chart Types', () => {
  test.beforeEach(async ({ request }) => await ensureSetup(request))

  for (const chart of ['bar', 'pie', 'doughnut', 'line', 'table']) {
    test('API: chart ' + chart, async ({ request }) => {
      const res = await request.post(BASE_API + '/projects/' + _projectId + '/reports', { data: { report_type: 'distribution', group_by: 'state', chart }, headers: auth() })
      expect(res.ok()).toBeTruthy()
      const d = await res.json()
      expect(d).toHaveProperty('labels')
      expect(d).toHaveProperty('values')
    })
  }
})
