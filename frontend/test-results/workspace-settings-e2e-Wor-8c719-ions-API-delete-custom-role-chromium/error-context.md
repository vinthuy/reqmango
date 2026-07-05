# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: workspace-settings-e2e.spec.ts >> Workspace Settings - Roles & Permissions >> API: delete custom role
- Location: e2e\workspace-settings-e2e.spec.ts:275:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 200
Received: 404
```

# Test source

```ts
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
> 289 |     expect(res.status()).toBe(200)
      |                          ^ Error: expect(received).toBe(expected) // Object.is equality
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
  331 | test.describe('Workspace Settings - Plugins', () => {
  332 |   test.beforeAll(async ({ request }) => { await ensureSetup(request) })
  333 | 
  334 |   test('plugins section shows plugin catalog', async ({ page }) => {
  335 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  336 |     await navigateToSection(page, '插件') || await navigateToSection(page, 'Plugins')
  337 |     await page.waitForTimeout(1000)
  338 |     await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  339 |   })
  340 | 
  341 |   test('API: list plugin catalog', async ({ request }) => {
  342 |     const res = await request.get(`${BASE_API}/plugins/catalog`, {
  343 |       headers: { Authorization: `Bearer ${_token}` },
  344 |     })
  345 |     expect([200, 404, 405]).toContain(res.status())
  346 |   })
  347 | 
  348 |   test('API: list installed plugins', async ({ request }) => {
  349 |     const res = await request.get(`${BASE_API}/workspaces/${_wsId}/plugins`, {
  350 |       headers: { Authorization: `Bearer ${_token}` },
  351 |     })
  352 |     expect(res.status()).toBe(200)
  353 |   })
  354 | })
  355 | 
  356 | // ============================================================
  357 | // Workspace AI Settings
  358 | // ============================================================
  359 | test.describe('Workspace Settings - AI Config', () => {
  360 |   test.beforeAll(async ({ request }) => { await ensureSetup(request) })
  361 | 
  362 |   test('AI settings section loads', async ({ page }) => {
  363 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  364 |     await navigateToSection(page, 'AI')
  365 |     await page.waitForTimeout(800)
  366 |     await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  367 |   })
  368 | 
  369 |   test('API: get workspace AI config', async ({ request }) => {
  370 |     const res = await request.get(`${BASE_API}/workspaces/${_wsSlug}/ai-config`, {
  371 |       headers: { Authorization: `Bearer ${_token}` },
  372 |     })
  373 |     expect([200, 404]).toContain(res.status())
  374 |   })
  375 | 
  376 |   test('API: update workspace AI config', async ({ request }) => {
  377 |     const res = await request.put(`${BASE_API}/workspaces/${_wsSlug}/ai-config`, {
  378 |       data: { provider: 'openai', model: 'gpt-4', api_key: 'sk-test-key-for-e2e' },
  379 |       headers: { Authorization: `Bearer ${_token}` },
  380 |     })
  381 |     expect([200, 201]).toContain(res.status())
  382 |   })
  383 | })
  384 | 
  385 | // ============================================================
  386 | // Custom Fields & Types Sections (Quick Check)
  387 | // ============================================================
  388 | test.describe('Workspace Settings - Custom Fields & Types', () => {
  389 |   test.beforeAll(async ({ request }) => { await ensureSetup(request) })
```