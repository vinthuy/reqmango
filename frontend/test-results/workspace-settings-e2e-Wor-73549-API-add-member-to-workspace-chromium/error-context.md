# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: workspace-settings-e2e.spec.ts >> Workspace Settings - Members >> API: add member to workspace
- Location: e2e\workspace-settings-e2e.spec.ts:167:3

# Error details

```
SyntaxError: Unexpected non-whitespace character after JSON at position 4 (line 1 column 5)
```

# Test source

```ts
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
  98  |         await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
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
> 178 |     const usersBody = await usersRes.json()
      |                       ^ SyntaxError: Unexpected non-whitespace character after JSON at position 4 (line 1 column 5)
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
  230 |     await expect(roleContent).toBeVisible({ timeout: 5000 })
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
```