# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: workspace-settings-e2e.spec.ts >> Workspace Settings - Roles & Permissions >> roles section shows system roles by default
- Location: e2e\workspace-settings-e2e.spec.ts:223:3

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
    13 × locator resolved to <div class="flex-1"></div>
       - unexpected value "hidden"

```

```yaml
- banner:
  - button "R ReqMango"
  - navigation:
    - link "项目":
      - /url: /workspace/e2e-settings-e2ews1783254493398
    - link "战略目标":
      - /url: /workspace/e2e-settings-e2ews1783254493398/initiatives
    - link "设置":
      - /url: /workspace/e2e-settings-e2ews1783254493398/settings
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
    - heading "角色与权限管理" [level=2]
    - paragraph: 管理工作空间的自定义角色和权限分配
    - button "新建角色"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: E2E Custom Tester
    - paragraph: Auto-test role
    - text: 3 个权限
    - button "编辑"
    - button "删除"
    - button "展开权限"
    - text: Admin 系统角色
    - paragraph: Full access to all resources
    - text: 55 个权限
    - button "展开权限"
    - text: Member 系统角色
    - paragraph: Create and edit content
    - text: 33 个权限
    - button "展开权限"
    - text: Guest 系统角色
    - paragraph: Read-only access
    - text: 15 个权限
    - button "展开权限"
    - heading "所有可用权限" [level=3]
    - text: agent:manageManage AI Agents ai:useUse AI Features attachment:createCreate Attachments attachment:deleteDelete Attachments attachment:viewView Attachments automation:manageManage Automations automation:viewView Automations comment:createCreate Comments comment:deleteDelete Comments comment:editEdit Comments comment:viewView Comments cycle:createCreate Cycles cycle:deleteDelete Cycles cycle:editEdit Cycles cycle:viewView Cycles issue:createCreate Issues issue:deleteDelete Issues issue:editEdit Issues issue:exportExport Issues issue:importImport Issues issue:viewView Issues member:manage_projectManage Project Members member:view_projectView Project Members module:createCreate Modules module:deleteDelete Modules module:editEdit Modules module:viewView Modules page:createCreate Pages page:deleteDelete Pages page:editEdit Pages page:viewView Pages project:deleteDelete Project project:manageManage Project project:viewView Project release:manageManage Releases release:viewView Releases report:createCreate Reports report:viewView Reports settings:manageManage Settings settings:viewView Settings time_track:createLog Time time_track:deleteDelete Time Logs time_track:viewView Time Tracking workflow:manageManage Workflows workflow:viewView Workflows initiative:manageManage Initiatives initiative:viewView Initiatives member:manageManage Members member:viewView Members project:createCreate Project project:view_allView All Projects role:manageManage Roles workspace:deleteDelete Workspace workspace:manageManage Workspace workspace:viewView Workspace
```

# Test source

```ts
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
  199 |     // Get current members list to find our test user
  200 |     const membersRes = await request.get(`${BASE_API}/workspaces/${_wsSlug}/members`, {
  201 |       headers: { Authorization: `Bearer ${_token}` },
  202 |     })
  203 |     const members = await membersRes.json()
  204 |     const memberList = Array.isArray(members) ? members : (members.data || [])
  205 |     // Find a non-admin member
  206 |     const member = memberList.find((m: any) => m.role !== 20)
  207 |     if (!member) { return } // All are admin, skip
  208 | 
  209 |     const res = await request.put(`${BASE_API}/workspaces/${_wsSlug}/members/${member.user_id}`, {
  210 |       data: { role: 15 },
  211 |       headers: { Authorization: `Bearer ${_token}` },
  212 |     })
  213 |     expect(res.status()).toBe(200)
  214 |   })
  215 | })
  216 | 
  217 | // ============================================================
  218 | // Roles & Permissions Section
  219 | // ============================================================
  220 | test.describe('Workspace Settings - Roles & Permissions', () => {
  221 |   test.beforeAll(async ({ request }) => { await ensureSetup(request) })
  222 | 
  223 |   test('roles section shows system roles by default', async ({ page }) => {
  224 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  225 |     await navigateToSection(page, '角色') || await navigateToSection(page, 'Roles')
  226 |     await page.waitForTimeout(800)
  227 | 
  228 |     // Should show system roles or create role button
  229 |     const roleContent = page.locator('main, .flex-1').first()
> 230 |     await expect(roleContent).toBeVisible({ timeout: 5000 })
      |                               ^ Error: expect(locator).toBeVisible() failed
  231 |   })
  232 | 
  233 |   test('API: list roles returns system roles', async ({ request }) => {
  234 |     const res = await request.get(`${BASE_API}/workspaces/${_wsId}/roles`, {
  235 |       headers: { Authorization: `Bearer ${_token}` },
  236 |     })
  237 |     expect(res.status()).toBe(200)
  238 |     const body = await res.json()
  239 |     const roles = body.data?.data || body.data || body
  240 |     const roleList = Array.isArray(roles) ? roles : []
  241 |     expect(roleList.length).toBeGreaterThanOrEqual(3) // Admin, Member, Guest at minimum
  242 |   })
  243 | 
  244 |   test('API: list all permissions', async ({ request }) => {
  245 |     const res = await request.get(`${BASE_API}/permissions`, {
  246 |       headers: { Authorization: `Bearer ${_token}` },
  247 |     })
  248 |     expect(res.status()).toBe(200)
  249 |     const body = await res.json()
  250 |     const perms = body.data?.data || body.data || body
  251 |     expect(Array.isArray(perms)).toBeTruthy()
  252 |   })
  253 | 
  254 |   test('API: create custom role with permissions', async ({ request }) => {
  255 |     // Get permissions first
  256 |     const permRes = await request.get(`${BASE_API}/permissions`, {
  257 |       headers: { Authorization: `Bearer ${_token}` },
  258 |     })
  259 |     const permBody = await permRes.json()
  260 |     const perms = permBody.data?.data || permBody.data || permBody
  261 |     const permList = Array.isArray(perms) ? perms : []
  262 |     const permIds = permList.slice(0, 3).map((p: any) => p.id)
  263 | 
  264 |     const res = await request.post(`${BASE_API}/workspaces/${_wsId}/roles`, {
  265 |       data: { name: 'E2E Custom Tester', description: 'Auto-test role', scope: 'workspace', level: 10, permissions: permIds },
  266 |       headers: { Authorization: `Bearer ${_token}` },
  267 |     })
  268 |     expect(res.status()).toBe(201)
  269 |     const body = await res.json()
  270 |     const role = body.data || body
  271 |     expect(role.name).toBe('E2E Custom Tester')
  272 |     expect(role.level).toBe(10)
  273 |   })
  274 | 
  275 |   test('API: delete custom role', async ({ request }) => {
  276 |     // Get roles to find a deletable one
  277 |     const rolesRes = await request.get(`${BASE_API}/workspaces/${_wsId}/roles`, {
  278 |       headers: { Authorization: `Bearer ${_token}` },
  279 |     })
  280 |     const rolesBody = await rolesRes.json()
  281 |     const roles = rolesBody.data?.data || rolesBody.data || rolesBody
  282 |     const roleList = Array.isArray(roles) ? roles : []
  283 |     const customRole = roleList.find((r: any) => !r.is_system)
  284 |     if (!customRole) { return } // No custom roles exist
  285 | 
  286 |     const res = await request.delete(`${BASE_API}/roles/${customRole.id}`, {
  287 |       headers: { Authorization: `Bearer ${_token}` },
  288 |     })
  289 |     expect(res.status()).toBe(200)
  290 |   })
  291 | })
  292 | 
  293 | // ============================================================
  294 | // Integrations Section - API Tests
  295 | // ============================================================
  296 | test.describe('Workspace Settings - Integrations', () => {
  297 |   test.beforeAll(async ({ request }) => { await ensureSetup(request) })
  298 | 
  299 |   test('integrations section renders sub-tabs', async ({ page }) => {
  300 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  301 |     await navigateToSection(page, '集成') || await navigateToSection(page, 'Integrations')
  302 |     await page.waitForTimeout(800)
  303 |     await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  304 |   })
  305 | 
  306 |   test('API: list MCP configs', async ({ request }) => {
  307 |     const res = await request.get(`${BASE_API}/workspaces/${_wsId}/mcp`, {
  308 |       headers: { Authorization: `Bearer ${_token}` },
  309 |     })
  310 |     expect([200, 404]).toContain(res.status())
  311 |   })
  312 | 
  313 |   test('API: list GitHub connections', async ({ request }) => {
  314 |     const res = await request.get(`${BASE_API}/workspaces/${_wsId}/github`, {
  315 |       headers: { Authorization: `Bearer ${_token}` },
  316 |     })
  317 |     expect([200, 404]).toContain(res.status())
  318 |   })
  319 | 
  320 |   test('API: list Slack connections', async ({ request }) => {
  321 |     const res = await request.get(`${BASE_API}/workspaces/${_wsId}/slack`, {
  322 |       headers: { Authorization: `Bearer ${_token}` },
  323 |     })
  324 |     expect([200, 404]).toContain(res.status())
  325 |   })
  326 | })
  327 | 
  328 | // ============================================================
  329 | // Plugins Section
  330 | // ============================================================
```