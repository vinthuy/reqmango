import { randomUUID } from 'node:crypto'
import { test as base, expect, type Page } from '@playwright/test'

const api = '/api/v1'
type IntakeProject = {
  id: number
  workspaceId: number
  slug: string
  token: string
  backlogId: number
}

const test = base.extend<{ project: IntakeProject }>({
  project: async ({ request }, use) => {
    const run = randomUUID()
    const user = { email: `intake-${run}@test.com`, username: `intake-${run}`, password: 'IntakeTest123!' }
    const registered = await request.post(`${api}/auth/register`, { data: user })
    expect(registered.status()).toBe(201)
    const login = await request.post(`${api}/auth/login`, { data: user })
    expect(login.ok()).toBeTruthy()
    const { access_token: token } = await login.json()
    expect(token).toBeTruthy()
    const headers = { Authorization: `Bearer ${token}` }
    const workspaceRes = await request.post(`${api}/workspaces`, {
      headers, data: { name: 'Intake regression', slug: `intake-${run}` },
    })
    expect(workspaceRes.status()).toBe(201)
    const workspace = await workspaceRes.json()
    try {
      const created = await request.post(`${api}/projects?workspace_id=${workspace.id}`, {
        headers, data: { name: 'Intake regression', identifier: 'INTAKE' },
      })
      expect(created.status()).toBe(201)
      const { id } = await created.json()
      const settings = await request.put(`${api}/projects/${id}/intake/settings`, {
        headers, data: { form_enabled: true },
      })
      expect(settings.ok()).toBeTruthy()
      const statesRes = await request.get(`${api}/projects/${id}/settings/states`, { headers })
      expect(statesRes.ok()).toBeTruthy()
      const { data: states } = await statesRes.json()
      const backlog = states.find((state: { group: string }) => state.group === 'backlog')
      expect(backlog).toBeDefined()
      await use({ id, workspaceId: workspace.id, slug: workspace.slug, token, backlogId: backlog.id })
    } finally {
      const deleted = await request.delete(`${api}/workspaces/${workspace.id}`, { headers })
      expect(deleted.ok()).toBeTruthy()
    }
  },
})

async function submit(page: Page, project: IntakeProject, name: string) {
  await page.goto(`/intake/${project.id}`)
  await expect(page.getByTestId('intake-form-project')).toHaveText('Intake regression')
  expect(await page.evaluate(() => localStorage.getItem('token'))).toBeNull()
  await page.locator('input[name="name"]').fill(name)
  await page.locator('textarea[name="description"]').fill('Submitted anonymously for triage')
  await page.locator('input[name="submitter"]').fill('Requester')
  await page.locator('input[name="email"]').fill('requester@test.com')
  await page.locator('select[name="priority"]').selectOption('high')
  const response = page.waitForResponse(res => res.url().endsWith(`/intake/${project.id}`) && res.request().method() === 'POST')
  await page.getByTestId('intake-form-submit').click()
  const result = await response
  expect(result.status()).toBe(201)
  const item = await result.json()
  expect(item.status).toBe('pending')
  expect(item.id).toBeGreaterThan(0)
  await expect(page.getByTestId('intake-form-done')).toBeVisible()
  await expect(page.getByTestId('intake-form-done')).toContainText(String(item.sequence_id))
  return item as { id: number; sequence_id: number }
}

async function openQueue(page: Page, project: IntakeProject) {
  await page.addInitScript(token => {
    localStorage.setItem('token', token)
    localStorage.setItem('locale', 'en-US')
  }, project.token)
  await page.goto(`/workspace/${project.slug}/project/${project.id}/intake`)
  await expect(page.getByTestId('intake-page')).toBeVisible()
}

test('anonymous submission stays out of the backlog until accepted', async ({ page, request, project }) => {
  const name = `Accept ${randomUUID()}`
  const item = await submit(page, project, name)
  const headers = { Authorization: `Bearer ${project.token}` }
  const issuesUrl = `${api}/issues?project_id=${project.id}&workspace_id=${project.workspaceId}`
  const pendingBacklog = await request.get(issuesUrl, { headers })
  expect(pendingBacklog.ok()).toBeTruthy()
  expect(JSON.stringify(await pendingBacklog.json())).not.toContain(name)

  await openQueue(page, project)
  const row = page.getByTestId(`in-item-${item.id}`)
  await expect(row).toContainText(name)
  await row.click()
  await expect(page.getByTestId('in-desc')).toHaveText('Submitted anonymously for triage')
  await page.getByTestId('in-accept-state').selectOption(String(project.backlogId))
  const response = page.waitForResponse(res => res.url().endsWith(`/intake/${item.id}/triage`) && res.request().method() === 'POST')
  await page.getByTestId('in-accept').click()
  const triaged = await response
  expect(triaged.ok()).toBeTruthy()
  expect(await triaged.json()).toMatchObject({ issue_id: item.id, status: 'accepted', state_id: project.backlogId })
  await expect(row).toHaveCount(0)
  const accepted = await request.get(`${api}/issues/${item.id}`, { headers })
  expect(accepted.ok()).toBeTruthy()
  expect(await accepted.json()).toMatchObject({ name, state_id: project.backlogId, priority: 'high' })
  await page.goto(`/workspace/${project.slug}/project/${project.id}`)
  await expect(page.getByText(name, { exact: true }).first()).toBeVisible()
})

test('rejected submissions retain the reason and stay out of the backlog', async ({ page, request, project }) => {
  const name = `Reject ${randomUUID()}`
  const item = await submit(page, project, name)
  await openQueue(page, project)
  await page.getByTestId(`in-item-${item.id}`).click()
  await page.getByTestId('in-reject-reason').fill('Outside project scope')
  const response = page.waitForResponse(res => res.url().endsWith(`/intake/${item.id}/triage`) && res.request().method() === 'POST')
  await page.getByTestId('in-reject').click()
  expect((await response).ok()).toBeTruthy()
  await expect(page.getByTestId(`in-item-${item.id}`)).toHaveCount(0)
  const headers = { Authorization: `Bearer ${project.token}` }
  const rejected = await request.get(`${api}/projects/${project.id}/intake?status=rejected`, { headers })
  expect(rejected.ok()).toBeTruthy()
  expect((await rejected.json()).items).toEqual(expect.arrayContaining([
    expect.objectContaining({ id: item.id, status: 'rejected', note: 'Outside project scope' }),
  ]))
  const backlog = await request.get(`${api}/issues?project_id=${project.id}&workspace_id=${project.workspaceId}`, { headers })
  expect(backlog.ok()).toBeTruthy()
  expect(JSON.stringify(await backlog.json())).not.toContain(name)
})

test('empty titles are rejected and a disabled form cannot accept submissions', async ({ page, request, project }) => {
  await page.goto(`/intake/${project.id}`)
  await expect(page.getByTestId('intake-form-project')).toHaveText('Intake regression')
  await page.getByTestId('intake-form-submit').click()
  await expect(page.getByTestId('intake-form-error')).toBeVisible()
  await expect(page.getByTestId('intake-form-done')).toHaveCount(0)
  const headers = { Authorization: `Bearer ${project.token}` }
  const disabled = await request.put(`${api}/projects/${project.id}/intake/settings`, { headers, data: { form_enabled: false } })
  expect(disabled.ok()).toBeTruthy()
  await page.reload()
  await expect(page.getByTestId('intake-form-disabled')).toBeVisible()
  await expect(page.getByTestId('intake-form-submit')).toHaveCount(0)
  const submitted = await request.post(`${api}/intake/${project.id}`, { data: { name: 'Must not be accepted' } })
  expect(submitted.status()).toBe(403)
  const queue = await request.get(`${api}/projects/${project.id}/intake`, { headers })
  expect(queue.ok()).toBeTruthy()
  expect((await queue.json()).total).toBe(0)
})
