# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: workspace-settings-e2e.spec.ts >> Workspace Settings Navigation >> navigate to settings section: Automations
- Location: e2e\workspace-settings-e2e.spec.ts:93:5

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator:  locator('main, .flex-1').first()
Expected: visible
Received: hidden
Timeout:  5000ms

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('main, .flex-1').first()
    14 × locator resolved to <div class="flex-1"></div>
       - unexpected value "hidden"

```

```yaml
- banner:
  - button "R ReqMango"
  - navigation:
    - link "项目":
      - /url: /workspace/e2e-settings-e2ews1783254433346
    - link "战略目标":
      - /url: /workspace/e2e-settings-e2ews1783254433346/initiatives
    - link "设置":
      - /url: /workspace/e2e-settings-e2ews1783254433346/settings
  - button "E E2E Settings WS":
    - text: E E2E Settings WS
    - img
  - button "中"
  - button "🌙"
  - button:
    - img
  - button "E"
- main:
  - complementary:
    - heading "工作空间设置" [level=2]
    - paragraph: Configure workspace-wide settings
    - navigation:
      - button "👥 成员 1"
      - button "📋 工作项类型 0"
      - button "📦 模板 0"
      - button "🤖 AI 0"
      - button "📝 自定义字段 0"
      - button "🤖 自动化 0"
      - button "🔗 关联关系 0"
      - button "🔌 集成 0"
      - button "🔑 角色与权限 0"
      - button "🧩 插件 0"
  - main:
    - heading "自动化" [level=1]
    - paragraph: 工作空间级自动化规则
    - text: 🤖
    - heading "暂无自动化规则" [level=3]
    - paragraph: 创建自动化规则来简化您的工作流程
    - button "+ 创建自动化"
```

# Test source

```ts
  1   | /**
  2   |  * E2E Tests — 工作空间设置 (Workspace Settings)
  3   |  * 覆盖: 10个设置分区导航、成员管理、角色权限、集成、插件、AI设置
  4   |  */
  5   | import { test, expect, type Page, type APIRequestContext } from '@playwright/test'
  6   | 
  7   | const BASE_API = 'http://localhost:8000/api/v1'
  8   | const TEST_PREFIX = `e2ews${Date.now()}`
  9   | 
  10  | // ============================================================
  11  | // Helpers
  12  | // ============================================================
  13  | let _token = ''
  14  | let _wsId = 0
  15  | let _wsSlug = ''
  16  | let _projectId = 0
  17  | 
  18  | async function ensureSetup(request: APIRequestContext) {
  19  |   if (_token) return
  20  |   const user = {
  21  |     email: `${TEST_PREFIX}@t.com`, username: TEST_PREFIX,
  22  |     password: 'E2eTest123!', display_name: 'E2E WS Settings',
  23  |   }
  24  |   await request.post(`${BASE_API}/auth/register`, { data: user })
  25  |   const res = await request.post(`${BASE_API}/auth/login`, { data: { email: user.email, password: user.password } })
  26  |   const { access_token } = await res.json()
  27  |   _token = access_token
  28  | 
  29  |   const ws = await request.post(`${BASE_API}/workspaces`, {
  30  |     data: { name: 'E2E Settings WS', slug: `e2e-settings-${TEST_PREFIX}` },
  31  |     headers: { Authorization: `Bearer ${_token}` },
  32  |   })
  33  |   const wsData = await ws.json()
  34  |   _wsId = wsData.id || wsData.data?.id
  35  |   _wsSlug = wsData.slug || wsData.data?.slug
  36  | 
  37  |   const proj = await request.post(`${BASE_API}/projects?workspace_id=${_wsId}`, {
  38  |     data: { name: 'E2E Settings Project', identifier: 'E2EWS', description: 'For settings testing' },
  39  |     headers: { Authorization: `Bearer ${_token}` },
  40  |   })
  41  |   const projData = await proj.json()
  42  |   _projectId = projData.id || projData.data?.id
  43  | }
  44  | 
  45  | async function loginViaStorage(page: Page) {
  46  |   await page.goto('/login')
  47  |   await page.evaluate((t: string) => localStorage.setItem('token', t), _token)
  48  | }
  49  | 
  50  | async function goToApp(page: Page, path: string) {
  51  |   await loginViaStorage(page)
  52  |   await page.goto(path)
  53  |   await page.waitForLoadState('networkidle').catch(() => {})
  54  | }
  55  | 
  56  | async function navigateToSection(page: Page, sectionLabel: string) {
  57  |   // Try Chinese first, then English
  58  |   const btn = page.locator('aside').locator(`button:has-text("${sectionLabel}")`).first()
  59  |   if (await btn.isVisible({ timeout: 3000 }).catch(() => false)) {
  60  |     await btn.click()
  61  |     await page.waitForTimeout(800)
  62  |     return true
  63  |   }
  64  |   return false
  65  | }
  66  | 
  67  | // ============================================================
  68  | // Settings Navigation - All 10 Sections
  69  | // ============================================================
  70  | test.describe('Workspace Settings Navigation', () => {
  71  |   test.beforeAll(async ({ request }) => { await ensureSetup(request) })
  72  | 
  73  |   test('settings page loads with sidebar navigation', async ({ page }) => {
  74  |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  75  |     await expect(page.locator('aside')).toBeVisible({ timeout: 8000 })
  76  |     await expect(page.locator('h2').filter({ hasText: /设置|Settings/ })).toBeVisible({ timeout: 5000 })
  77  |   })
  78  | 
  79  |   const sections = [
  80  |     { id: 'members', zh: '成员', en: 'Members' },
  81  |     { id: 'types', zh: '工作项类型', en: 'Types' },
  82  |     { id: 'templates', zh: '模板', en: 'Templates' },
  83  |     { id: 'ai', zh: 'AI', en: 'AI' },
  84  |     { id: 'fields', zh: '字段', en: 'Fields' },
  85  |     { id: 'automations', zh: '自动化', en: 'Automations' },
  86  |     { id: 'relations', zh: '关联', en: 'Relations' },
  87  |     { id: 'integrations', zh: '集成', en: 'Integrations' },
  88  |     { id: 'roles', zh: '角色', en: 'Roles' },
  89  |     { id: 'plugins', zh: '插件', en: 'Plugins' },
  90  |   ]
  91  | 
  92  |   for (const section of sections) {
  93  |     test(`navigate to settings section: ${section.en}`, async ({ page }) => {
  94  |       await goToApp(page, `/workspace/${_wsSlug}/settings`)
  95  |       const navigated = await navigateToSection(page, section.zh) || await navigateToSection(page, section.en)
  96  |       if (navigated) {
  97  |         // Verify the main content area updated (not the loading spinner)
> 98  |         await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
      |                                                             ^ Error: expect(locator).toBeVisible() failed
  99  |       } else {
  100 |         // Accept if the sidebar is there even if section isn't fully rendered
  101 |         await expect(page.locator('aside')).toBeVisible()
  102 |       }
  103 |     })
  104 |   }
  105 | })
  106 | 
  107 | // ============================================================
  108 | // Members Section - CRUD Tests
  109 | // ============================================================
  110 | test.describe('Workspace Settings - Members', () => {
  111 |   test.beforeAll(async ({ request }) => { await ensureSetup(request) })
  112 | 
  113 |   test('members section shows member list with headers', async ({ page }) => {
  114 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  115 |     await navigateToSection(page, '成员') || await navigateToSection(page, 'Members')
  116 | 
  117 |     // Should see member list or the section header
  118 |     await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  119 |   })
  120 | 
  121 |   test('add member modal opens with user search', async ({ page }) => {
  122 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  123 |     await navigateToSection(page, '成员') || await navigateToSection(page, 'Members')
  124 |     await page.waitForTimeout(500)
  125 | 
  126 |     // Click "添加成员" or "Add Member" button
  127 |     const addBtn = page.locator('button:has-text("添加成员")').or(page.locator('button:has-text("Add Member")')).first()
  128 |     if (await addBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
  129 |       await addBtn.click()
  130 |       await page.waitForTimeout(500)
  131 | 
  132 |       // Modal should appear with search input
  133 |       const searchInput = page.locator('input[placeholder*="搜索用户"]').or(page.locator('input[placeholder*="Search"]')).or(page.locator('input[placeholder*="user"]'))
  134 |       if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
  135 |         await expect(searchInput).toBeVisible()
  136 |       }
  137 | 
  138 |       // Close modal
  139 |       const cancelBtn = page.locator('button:has-text("取消")').or(page.locator('button:has-text("Cancel")')).last()
  140 |       if (await cancelBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
  141 |         await cancelBtn.click()
  142 |       }
  143 |     }
  144 |   })
  145 | 
  146 |   test('member role can be changed via dropdown', async ({ page }) => {
  147 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  148 |     await navigateToSection(page, '成员') || await navigateToSection(page, 'Members')
  149 |     await page.waitForTimeout(800)
  150 | 
  151 |     // Check if member role select dropdowns exist
  152 |     const roleSelects = page.locator('select')
  153 |     const count = await roleSelects.count().catch(() => 0)
  154 |     expect(count).toBeGreaterThanOrEqual(0) // Might be 0 if no members beyond creator
  155 |   })
  156 | 
  157 |   test('API: list workspace members', async ({ request }) => {
  158 |     const res = await request.get(`${BASE_API}/workspaces/${_wsSlug}/members`, {
  159 |       headers: { Authorization: `Bearer ${_token}` },
  160 |     })
  161 |     expect(res.status()).toBe(200)
  162 |     const body = await res.json()
  163 |     const members = Array.isArray(body) ? body : (body.data || [])
  164 |     expect(members.length).toBeGreaterThanOrEqual(1)
  165 |   })
  166 | 
  167 |   test('API: add member to workspace', async ({ request }) => {
  168 |     // Create another user first
  169 |     const anotherUser = `e2emember${Date.now()}`
  170 |     await request.post(`${BASE_API}/auth/register`, {
  171 |       data: { email: `${anotherUser}@t.com`, username: anotherUser, password: 'E2eTest123!', display_name: 'New Member' },
  172 |     })
  173 | 
  174 |     // Get member user ID by searching
  175 |     const usersRes = await request.get(`${BASE_API}/auth/users`, {
  176 |       headers: { Authorization: `Bearer ${_token}` },
  177 |     })
  178 |     const usersBody = await usersRes.json()
  179 |     const users = Array.isArray(usersBody) ? usersBody : (usersBody.data || [])
  180 |     const newUser = users.find((u: any) => u.username === anotherUser || u.email === `${anotherUser}@t.com`)
  181 |     if (!newUser) { test.skip(true, 'Could not find created user'); return }
  182 | 
  183 |     const res = await request.post(`${BASE_API}/workspaces/${_wsSlug}/members`, {
  184 |       data: { user_id: newUser.id, role: 15 },
  185 |       headers: { Authorization: `Bearer ${_token}` },
  186 |     })
  187 |     expect(res.status()).toBe(201)
  188 | 
  189 |     // Verify member added
  190 |     const membersRes = await request.get(`${BASE_API}/workspaces/${_wsSlug}/members`, {
  191 |       headers: { Authorization: `Bearer ${_token}` },
  192 |     })
  193 |     const members = await membersRes.json()
  194 |     const memberList = Array.isArray(members) ? members : (members.data || [])
  195 |     expect(memberList.some((m: any) => m.user_id === newUser.id)).toBeTruthy()
  196 |   })
  197 | 
  198 |   test('API: update member role', async ({ request }) => {
```