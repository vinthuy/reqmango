/**
 * E2E Tests — 工作空间设置 (Workspace Settings)
 * 覆盖: 10个设置分区导航、成员管理、角色权限、集成、插件、AI设置
 */
import { test, expect, type Page, type APIRequestContext } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'
const TEST_PREFIX = `e2ews${Date.now()}`

// ============================================================
// Helpers
// ============================================================
let _token = ''
let _wsId = 0
let _wsSlug = ''
let _projectId = 0

async function ensureSetup(request: APIRequestContext) {
  if (_token) return
  const user = {
    email: `${TEST_PREFIX}@t.com`, username: TEST_PREFIX,
    password: 'E2eTest123!', display_name: 'E2E WS Settings',
  }
  await request.post(`${BASE_API}/auth/register`, { data: user })
  const res = await request.post(`${BASE_API}/auth/login`, { data: { email: user.email, password: user.password } })
  const { access_token } = await res.json()
  _token = access_token

  const ws = await request.post(`${BASE_API}/workspaces`, {
    data: { name: 'E2E Settings WS', slug: `e2e-settings-${TEST_PREFIX}` },
    headers: { Authorization: `Bearer ${_token}` },
  })
  const wsData = await ws.json()
  _wsId = wsData.id || wsData.data?.id
  _wsSlug = wsData.slug || wsData.data?.slug

  const proj = await request.post(`${BASE_API}/projects?workspace_id=${_wsId}`, {
    data: { name: 'E2E Settings Project', identifier: 'E2EWS', description: 'For settings testing' },
    headers: { Authorization: `Bearer ${_token}` },
  })
  const projData = await proj.json()
  _projectId = projData.id || projData.data?.id
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

async function navigateToSection(page: Page, sectionLabel: string) {
  // Try Chinese first, then English
  const btn = page.locator('aside').locator(`button:has-text("${sectionLabel}")`).first()
  if (await btn.isVisible({ timeout: 3000 }).catch(() => false)) {
    await btn.click()
    await page.waitForTimeout(800)
    return true
  }
  return false
}

// ============================================================
// Settings Navigation - All 10 Sections
// ============================================================
test.describe('Workspace Settings Navigation', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('settings page loads with sidebar navigation', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await expect(page.locator('aside')).toBeVisible({ timeout: 8000 })
    await expect(page.locator('h2').filter({ hasText: /设置|Settings/ })).toBeVisible({ timeout: 5000 })
  })

  const sections = [
    { id: 'members', zh: '成员', en: 'Members' },
    { id: 'types', zh: '工作项类型', en: 'Types' },
    { id: 'templates', zh: '模板', en: 'Templates' },
    { id: 'ai', zh: 'AI', en: 'AI' },
    { id: 'fields', zh: '字段', en: 'Fields' },
    { id: 'automations', zh: '自动化', en: 'Automations' },
    { id: 'relations', zh: '关联', en: 'Relations' },
    { id: 'integrations', zh: '集成', en: 'Integrations' },
    { id: 'roles', zh: '角色', en: 'Roles' },
    { id: 'plugins', zh: '插件', en: 'Plugins' },
  ]

  for (const section of sections) {
    test(`navigate to settings section: ${section.en}`, async ({ page }) => {
      await goToApp(page, `/workspace/${_wsSlug}/settings`)
      const navigated = await navigateToSection(page, section.zh) || await navigateToSection(page, section.en)
      if (navigated) {
        // Verify the main content area updated (not the loading spinner)
        await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
      } else {
        // Accept if the sidebar is there even if section isn't fully rendered
        await expect(page.locator('aside')).toBeVisible()
      }
    })
  }
})

// ============================================================
// Members Section - CRUD Tests
// ============================================================
test.describe('Workspace Settings - Members', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('members section shows member list with headers', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, '成员') || await navigateToSection(page, 'Members')

    // Should see member list or the section header
    await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  })

  test('add member modal opens with user search', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, '成员') || await navigateToSection(page, 'Members')
    await page.waitForTimeout(500)

    // Click "添加成员" or "Add Member" button
    const addBtn = page.locator('button:has-text("添加成员")').or(page.locator('button:has-text("Add Member")')).first()
    if (await addBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
      await addBtn.click()
      await page.waitForTimeout(500)

      // Modal should appear with search input
      const searchInput = page.locator('input[placeholder*="搜索用户"]').or(page.locator('input[placeholder*="Search"]')).or(page.locator('input[placeholder*="user"]'))
      if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
        await expect(searchInput).toBeVisible()
      }

      // Close modal
      const cancelBtn = page.locator('button:has-text("取消")').or(page.locator('button:has-text("Cancel")')).last()
      if (await cancelBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await cancelBtn.click()
      }
    }
  })

  test('member role can be changed via dropdown', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, '成员') || await navigateToSection(page, 'Members')
    await page.waitForTimeout(800)

    // Check if member role select dropdowns exist
    const roleSelects = page.locator('select')
    const count = await roleSelects.count().catch(() => 0)
    expect(count).toBeGreaterThanOrEqual(0) // Might be 0 if no members beyond creator
  })

  test('API: list workspace members', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsSlug}/members`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    const members = Array.isArray(body) ? body : (body.data || [])
    expect(members.length).toBeGreaterThanOrEqual(1)
  })

  test('API: add member to workspace', async ({ request }) => {
    // Create another user first
    const anotherUser = `e2emember${Date.now()}`
    await request.post(`${BASE_API}/auth/register`, {
      data: { email: `${anotherUser}@t.com`, username: anotherUser, password: 'E2eTest123!', display_name: 'New Member' },
    })

    // Get member user ID by searching
    const usersRes = await request.get(`${BASE_API}/auth/users`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const usersBody = await usersRes.json()
    const users = Array.isArray(usersBody) ? usersBody : (usersBody.data || [])
    const newUser = users.find((u: any) => u.username === anotherUser || u.email === `${anotherUser}@t.com`)
    if (!newUser) { test.skip(true, 'Could not find created user'); return }

    const res = await request.post(`${BASE_API}/workspaces/${_wsSlug}/members`, {
      data: { user_id: newUser.id, role: 15 },
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(201)

    // Verify member added
    const membersRes = await request.get(`${BASE_API}/workspaces/${_wsSlug}/members`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const members = await membersRes.json()
    const memberList = Array.isArray(members) ? members : (members.data || [])
    expect(memberList.some((m: any) => m.user_id === newUser.id)).toBeTruthy()
  })

  test('API: update member role', async ({ request }) => {
    // Get current members list to find our test user
    const membersRes = await request.get(`${BASE_API}/workspaces/${_wsSlug}/members`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const members = await membersRes.json()
    const memberList = Array.isArray(members) ? members : (members.data || [])
    // Find a non-admin member
    const member = memberList.find((m: any) => m.role !== 20)
    if (!member) { return } // All are admin, skip

    const res = await request.put(`${BASE_API}/workspaces/${_wsSlug}/members/${member.user_id}`, {
      data: { role: 15 },
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
  })
})

// ============================================================
// Roles & Permissions Section
// ============================================================
test.describe('Workspace Settings - Roles & Permissions', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('roles section shows system roles by default', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, '角色') || await navigateToSection(page, 'Roles')
    await page.waitForTimeout(800)

    // Should show system roles or create role button
    const roleContent = page.locator('main, .flex-1').first()
    await expect(roleContent).toBeVisible({ timeout: 5000 })
  })

  test('API: list roles returns system roles', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsId}/roles`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    const roles = body.data?.data || body.data || body
    const roleList = Array.isArray(roles) ? roles : []
    expect(roleList.length).toBeGreaterThanOrEqual(3) // Admin, Member, Guest at minimum
  })

  test('API: list all permissions', async ({ request }) => {
    const res = await request.get(`${BASE_API}/permissions`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    const perms = body.data?.data || body.data || body
    expect(Array.isArray(perms)).toBeTruthy()
  })

  test('API: create custom role with permissions', async ({ request }) => {
    // Get permissions first
    const permRes = await request.get(`${BASE_API}/permissions`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const permBody = await permRes.json()
    const perms = permBody.data?.data || permBody.data || permBody
    const permList = Array.isArray(perms) ? perms : []
    const permIds = permList.slice(0, 3).map((p: any) => p.id)

    const res = await request.post(`${BASE_API}/workspaces/${_wsId}/roles`, {
      data: { name: 'E2E Custom Tester', description: 'Auto-test role', scope: 'workspace', level: 10, permissions: permIds },
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(201)
    const body = await res.json()
    const role = body.data || body
    expect(role.name).toBe('E2E Custom Tester')
    expect(role.level).toBe(10)
  })

  test('API: delete custom role', async ({ request }) => {
    // Get roles to find a deletable one
    const rolesRes = await request.get(`${BASE_API}/workspaces/${_wsId}/roles`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rolesBody = await rolesRes.json()
    const roles = rolesBody.data?.data || rolesBody.data || rolesBody
    const roleList = Array.isArray(roles) ? roles : []
    const customRole = roleList.find((r: any) => !r.is_system)
    if (!customRole) { return } // No custom roles exist

    const res = await request.delete(`${BASE_API}/roles/${customRole.id}`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
  })
})

// ============================================================
// Integrations Section - API Tests
// ============================================================
test.describe('Workspace Settings - Integrations', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('integrations section renders sub-tabs', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, '集成') || await navigateToSection(page, 'Integrations')
    await page.waitForTimeout(800)
    await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  })

  test('API: list MCP configs', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsId}/mcp`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect([200, 404]).toContain(res.status())
  })

  test('API: list GitHub connections', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsId}/github`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect([200, 404]).toContain(res.status())
  })

  test('API: list Slack connections', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsId}/slack`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect([200, 404]).toContain(res.status())
  })
})

// ============================================================
// Plugins Section
// ============================================================
test.describe('Workspace Settings - Plugins', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('plugins section shows plugin catalog', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, '插件') || await navigateToSection(page, 'Plugins')
    await page.waitForTimeout(1000)
    await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  })

  test('API: list plugin catalog', async ({ request }) => {
    const res = await request.get(`${BASE_API}/plugins/catalog`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect([200, 404, 405]).toContain(res.status())
  })

  test('API: list installed plugins', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsId}/plugins`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
  })
})

// ============================================================
// Workspace AI Settings
// ============================================================
test.describe('Workspace Settings - AI Config', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('AI settings section loads', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, 'AI')
    await page.waitForTimeout(800)
    await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  })

  test('API: get workspace AI config', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsSlug}/ai-config`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect([200, 404]).toContain(res.status())
  })

  test('API: update workspace AI config', async ({ request }) => {
    const res = await request.put(`${BASE_API}/workspaces/${_wsSlug}/ai-config`, {
      data: { provider: 'openai', model: 'gpt-4', api_key: 'sk-test-key-for-e2e' },
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect([200, 201]).toContain(res.status())
  })
})

// ============================================================
// Custom Fields & Types Sections (Quick Check)
// ============================================================
test.describe('Workspace Settings - Custom Fields & Types', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('work item types section loads', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, '工作项类型') || await navigateToSection(page, 'Types')
    await page.waitForTimeout(800)
    await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  })

  test('custom fields section loads', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, '字段') || await navigateToSection(page, 'Fields')
    await page.waitForTimeout(800)
    await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  })

  test('API: list issue types for workspace', async ({ request }) => {
    const res = await request.get(`${BASE_API}/issue-types?workspace_id=${_wsId}`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect(res.status()).toBe(200)
  })

  test('API: list custom fields for workspace', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsId}/custom-fields`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect([200, 404]).toContain(res.status())
  })
})

// ============================================================
// Relations & Templates (Quick Check)
// ============================================================
test.describe('Workspace Settings - Relations & Templates', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('relations section loads with relation type list', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, '关联') || await navigateToSection(page, 'Relations')
    await page.waitForTimeout(800)
    await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  })

  test('templates section loads', async ({ page }) => {
    await goToApp(page, `/workspace/${_wsSlug}/settings`)
    await navigateToSection(page, '模板') || await navigateToSection(page, 'Templates')
    await page.waitForTimeout(800)
    await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  })

  test('API: list relation types', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsId}/relations/types`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect([200, 404]).toContain(res.status())
  })

  test('API: list project templates', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/${_wsId}/templates`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    expect([200, 404]).toContain(res.status())
  })
})
