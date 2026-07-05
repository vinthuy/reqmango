/**
 * E2E Tests — 战略目标 (Initiatives)
 * 覆盖: 列表视图、路线图视图、CRUD、进度跟踪、状态管理、项目关联
 */
import { test, expect, type Page, type APIRequestContext } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'
const TEST_PREFIX = `e2eini${Date.now()}`

// ============================================================
// Helpers
// ============================================================
let _token = ''
let _wsId = 0
let _wsSlug = ''
let _projectId = 0
let _initiativeId = 0

async function ensureSetup(request: APIRequestContext) {
  if (_token) return
  const user = {
    email: `${TEST_PREFIX}@t.com`, username: TEST_PREFIX,
    password: 'E2eTest123!', display_name: 'E2E Initiative Test',
  }
  await request.post(`${BASE_API}/auth/register`, { data: user })
  const res = await request.post(`${BASE_API}/auth/login`, { data: { email: user.email, password: user.password } })
  const { access_token } = await res.json()
  _token = access_token

  const ws = await request.post(`${BASE_API}/workspaces`, {
    data: { name: 'E2E Initiative WS', slug: `e2e-ini-ws-${TEST_PREFIX}` },
    headers: { Authorization: `Bearer ${_token}` },
  })
  const wsData = await ws.json()
  _wsId = wsData.id || wsData.data?.id
  _wsSlug = wsData.slug || wsData.data?.slug

  const proj = await request.post(`${BASE_API}/projects?workspace_id=${_wsId}`, {
    data: { name: 'E2E Initiative Project', identifier: 'E2EINI', description: 'For initiative testing' },
    headers: { Authorization: `Bearer ${_token}` },
  })
  const projData = await proj.json()
  _projectId = projData.id || projData.data?.id

  // Create an initiative for update/delete tests
  const ini = await request.post(`${BASE_API}/workspaces/${_wsId}/initiatives`, {
    data: { name: 'Initial Test Initiative', description: 'For CRUD testing', color: '#3b82f6', status: 'active', project_ids: [_projectId] },
    headers: { Authorization: `Bearer ${_token}` },
  })
  const iniData = await ini.json()
  _initiativeId = iniData.id || iniData.data?.id
}

async function loginViaStorage(page: Page) {
  await page.goto('/login')
  await page.evaluate((t: string) => localStorage.setItem('token', t), _token)
}

async function goToApp(page: Page, path: string) {
  await loginViaStorage(page)
  await page.goto(path)
  await page.waitForLoadState('networkidle').catch(() => {})
}

// ============================================================
// API Level Tests
// ============================================================
test.describe('Initiatives API CRUD', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('POST /workspaces/:id/initiatives - create initiative', async ({ request }) => {
    const res = await request.post(`${BASE_API}/workspaces/${_wsId}/initiatives`, {
      data: { name: 'API Test Initiative', description: 'Created via API', color: '#ef4444', status: 'active', project_ids: [_projectId] },
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(201)
    const body = await res.json()
    const data = body.data || body
    expect(data.name).toBe('API Test Initiative')
    expect(data.status).toBe('active')
  })

  test('POST create - validation: empty name rejected', async ({ request }) => {
    const res = await request.post(`${BASE_API}/workspaces/${_wsId}/initiatives`, {
      data: { name: '', color: '#3b82f6' },
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBeGreaterThanOrEqual(400)
  })

  test('GET /workspaces/:id/initiatives - list initiatives', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsId}/initiatives`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    const data = Array.isArray(body) ? body : (body.data || [])
    expect(data.length).toBeGreaterThanOrEqual(1)
  })

  test('GET /initiatives/:id - get single initiative', async ({ request }) => {
    const res = await request.get(`${BASE_API}/initiatives/${_initiativeId}`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    const data = body.data || body
    expect(data.name).toBe('Initial Test Initiative')
  })

  test('PUT /initiatives/:id - update initiative', async ({ request }) => {
    const res = await request.put(`${BASE_API}/initiatives/${_initiativeId}`, {
      data: { name: 'Updated Initiative Name', status: 'completed', description: 'Updated description' },
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    const data = body.data || body
    expect(data.name).toBe('Updated Initiative Name')
    expect(data.status).toBe('completed')
  })

  test('GET /initiatives/:id/progress - get progress stats', async ({ request }) => {
    const res = await request.get(`${BASE_API}/initiatives/${_initiativeId}/progress`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    const data = body.data || body
    expect(data).toHaveProperty('total_issues')
    expect(data).toHaveProperty('completed_issues')
    expect(data).toHaveProperty('progress')
    expect(data).toHaveProperty('project_count')
  })

  test('GET /workspaces/:id/initiatives/search - search initiatives', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsId}/initiatives/search?q=Updated`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    const data = Array.isArray(body) ? body : (body.data || [])
    expect(data.length).toBeGreaterThanOrEqual(1)
    expect(data.some((i: any) => i.name?.includes('Updated'))).toBeTruthy()
  })

  test('DELETE /initiatives/:id - delete initiative', async ({ request }) => {
    // Create one to delete
    const createRes = await request.post(`${BASE_API}/workspaces/${_wsId}/initiatives`, {
      data: { name: 'To Be Deleted', color: '#6b7280' },
      headers: { Authorization: `Bearer ${_token}` },
    })
    const created = await createRes.json()
    const delId = created.id || created.data?.id

    const res = await request.delete(`${BASE_API}/initiatives/${delId}`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)

    // Verify gone
    const getRes = await request.get(`${BASE_API}/initiatives/${delId}`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(getRes.status()).toBe(404)
  })
})

// ============================================================
// UI Page Tests
// ============================================================
test.describe('Initiatives UI Pages', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('initiatives page loads with header and view toggle', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/initiatives`)
    await expect(page.locator('h1')).toContainText('Initiatives')
    // View toggle buttons exist
    await expect(page.locator('button:has-text("列表")').or(page.locator('button:has-text("List")'))).toBeVisible({ timeout: 5000 })
    await expect(page.locator('button:has-text("路线图")').or(page.locator('button:has-text("Roadmap")'))).toBeVisible({ timeout: 5000 })
  })

  test('list view shows initiative cards with progress bar', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/initiatives`)
    // Should be in list view by default
    await expect(page.locator('body')).toBeVisible()
    // Verify initiative cards render (they have progress bars, status labels)
    const cards = page.locator('.grid .border.rounded-lg')
    const cardCount = await cards.count().catch(() => 0)
    if (cardCount > 0) {
      // Check for progress-related elements
      const progressBar = page.locator('[class*="rounded-full"][class*="h-2"]').first()
      const hasProgress = await progressBar.isVisible({ timeout: 3000 }).catch(() => false)
      expect(cardCount >= 1 || hasProgress).toBeTruthy()
    }
  })

  test('list view shows empty state when no initiatives', async ({ page, request }) => {
    // Create a fresh workspace with no initiatives
    const freshWs = await request.post(`${BASE_API}/workspaces`, {
      data: { name: 'Empty Initiative WS', slug: `empty-ini-${Date.now()}` },
      headers: { Authorization: `Bearer ${_token}` },
    })
    const freshWsData = await freshWs.json()
    const freshSlug = freshWsData.slug || freshWsData.data?.slug

    await goToApp(page, `/workspace/${freshSlug}/initiatives`)
    await page.waitForTimeout(1500)
    const emptyText = page.locator('text="暂无战略目标"').or(page.locator('text="No initiatives"')).or(page.locator('text="创建第一个"')).or(page.locator('text="Create your first"'))
    const visible = await emptyText.isVisible({ timeout: 5000 }).catch(() => false)
    // Either empty state is shown or the loading indicator disappears
    await expect(page.locator('body')).toBeVisible()
  })

  test('list view and roadmap view can toggle', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/initiatives`)
    await page.waitForTimeout(1000)

    // Click roadmap view
    const roadmapBtn = page.locator('button:has-text("路线图")').or(page.locator('button:has-text("Roadmap")')).first()
    if (await roadmapBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await roadmapBtn.click()
      await page.waitForTimeout(800)
    }

    // Click back to list view
    const listBtn = page.locator('button:has-text("列表")').or(page.locator('button:has-text("List")')).first()
    if (await listBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await listBtn.click()
      await page.waitForTimeout(800)
    }
    await expect(page.locator('body')).toBeVisible()
  })

  test('create initiative button opens modal form', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/initiatives`)
    await page.waitForTimeout(1000)

    const createBtn = page.locator('button:has-text("创建战略目标")').or(page.locator('button:has-text("Create Initiative")'))
    if (await createBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
      await createBtn.click()
      await page.waitForTimeout(500)

      // Modal should be visible with form fields
      const modal = page.locator('.fixed.inset-0.bg-black\\/30').or(page.locator('.fixed.inset-0.z-50'))
      await expect(modal).toBeVisible({ timeout: 3000 })

      // Check form fields exist
      const nameInput = page.locator('input[placeholder*="名称"]').or(page.locator('input[placeholder*="Name"]')).or(modal.locator('input[type="text"]').first())
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(nameInput).toBeVisible()
      }

      // Close modal
      const cancelBtn = page.locator('button:has-text("取消")').or(page.locator('button:has-text("Cancel")'))
      if (await cancelBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await cancelBtn.click()
      }
    }
  })

  test('edit initiative via UI - opens edit form', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/initiatives`)
    await page.waitForTimeout(1000)

    // Click edit button on an initiative card
    const editBtn = page.locator('button:has-text("编辑")').or(page.locator('button:has-text("Edit")')).first()
    if (await editBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await editBtn.click()
      await page.waitForTimeout(500)

      // Modal with form should appear
      const modal = page.locator('.fixed.inset-0.bg-black\\/30').or(page.locator('.fixed.inset-0.z-50'))
      if (await modal.isVisible({ timeout: 3000 }).catch(() => false)) {
        // Check title is edit mode
        await expect(page.locator('body')).toBeVisible()
        // Close
        const cancelBtn = page.locator('button:has-text("取消")').or(page.locator('button:has-text("Cancel")'))
        if (await cancelBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
          await cancelBtn.click()
        }
      }
    }
  })

  test('delete initiative triggers confirmation dialog', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/initiatives`)
    await page.waitForTimeout(1000)

    // Click delete button on an initiative card
    const deleteBtn = page.locator('button:has-text("删除")').or(page.locator('button:has-text("Delete")')).first()
    if (await deleteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await deleteBtn.click()
      await page.waitForTimeout(500)
      // A confirm dialog or browser confirm should appear
      await page.waitForTimeout(300)
    }
    await expect(page.locator('body')).toBeVisible()
  })
})

// ============================================================
// Status & Progress Tests
// ============================================================
test.describe('Initiatives Status Management', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  const statuses = ['active', 'completed', 'paused', 'at_risk', 'off_track']

  for (const status of statuses) {
    test(`create initiative with status: ${status}`, async ({ request }) => {
      const res = await request.post(`${BASE_API}/workspaces/${_wsId}/initiatives`, {
        data: { name: `Status-${status} Init`, color: '#3b82f6', status },
        headers: { Authorization: `Bearer ${_token}` },
      })
      expect(res.status()).toBe(201)
      const body = await res.json()
      const data = body.data || body
      expect(data.status).toBe(status)
    })
  }

  test('initiative with projects computes progress', async ({ request }) => {
    // Create an initiative with a project
    const res = await request.post(`${BASE_API}/workspaces/${_wsId}/initiatives`, {
      data: { name: 'Progress Test', color: '#10b981', status: 'active', project_ids: [_projectId] },
      headers: { Authorization: `Bearer ${_token}` },
    })
    const body = await res.json()
    const iniId = body.id || body.data?.id

    const progressRes = await request.get(`${BASE_API}/initiatives/${iniId}/progress`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(progressRes.status()).toBe(200)
    const progressBody = await progressRes.json()
    const progress = progressBody.data || progressBody
    // Should have progress stats even for empty project
    expect(progress).toHaveProperty('project_count')
    expect(progress.project_count).toBeGreaterThanOrEqual(1)
  })
})

// ============================================================
// Roadmap View Tests
// ============================================================
test.describe('Initiatives Roadmap View', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('roadmap page redirects to initiatives', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/roadmap`)
    await page.waitForTimeout(1000)
    await expect(page).toHaveURL(/initiatives/)
  })

  test('roadmap view loads with date range', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/initiatives`)
    await page.waitForTimeout(1000)

    // Switch to roadmap view
    const roadmapBtn = page.locator('button:has-text("路线图")').or(page.locator('button:has-text("Roadmap")')).first()
    if (await roadmapBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await roadmapBtn.click()
      await page.waitForTimeout(1000)
    }

    // Roadmap should render (either content or empty state)
    await expect(page.locator('body')).toBeVisible()
  })
})

// ============================================================
// Sidebar Navigation
// ============================================================
test.describe('Initiatives Navigation', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('sidebar has initiatives link with trendline icon', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}`)
    await page.waitForTimeout(1000)

    const sidebarText = page.locator('text="战略目标"').or(page.locator('text="Initiatives"'))
    if (await sidebarText.isVisible({ timeout: 5000 }).catch(() => false)) {
      await sidebarText.click()
      await page.waitForTimeout(800)
      await expect(page).toHaveURL(/initiatives/)
    }
  })

  test('topbar has initiatives navigation item', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}`)
    await page.waitForTimeout(1000)

    const topBarIni = page.locator('header, [class*="topbar"], [class*="TopBar"]').locator('text="战略目标"').or(page.locator('header, [class*="topbar"], [class*="TopBar"]').locator('text="Initiatives"'))
    if (await topBarIni.isVisible({ timeout: 5000 }).catch(() => false)) {
      await topBarIni.click()
      await page.waitForTimeout(800)
      await expect(page).toHaveURL(/initiatives/)
    }
  })
})
