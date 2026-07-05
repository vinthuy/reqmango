# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: workspace-settings-e2e.spec.ts >> Workspace Settings - AI Config >> API: get workspace AI config
- Location: e2e\workspace-settings-e2e.spec.ts:369:3

# Error details

```
Error: expect(received).toContain(expected) // indexOf

Expected value: 400
Received array: [200, 404]
```

# Test source

```ts
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
> 373 |     expect([200, 404]).toContain(res.status())
      |                        ^ Error: expect(received).toContain(expected) // indexOf
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
  390 | 
  391 |   test('work item types section loads', async ({ page }) => {
  392 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  393 |     await navigateToSection(page, '工作项类型') || await navigateToSection(page, 'Types')
  394 |     await page.waitForTimeout(800)
  395 |     await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  396 |   })
  397 | 
  398 |   test('custom fields section loads', async ({ page }) => {
  399 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  400 |     await navigateToSection(page, '字段') || await navigateToSection(page, 'Fields')
  401 |     await page.waitForTimeout(800)
  402 |     await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  403 |   })
  404 | 
  405 |   test('API: list issue types for workspace', async ({ request }) => {
  406 |     const res = await request.get(`${BASE_API}/issue-types?workspace_id=${_wsId}`, {
  407 |       headers: { Authorization: `Bearer ${_token}` },
  408 |     })
  409 |     expect(res.status()).toBe(200)
  410 |   })
  411 | 
  412 |   test('API: list custom fields for workspace', async ({ request }) => {
  413 |     const res = await request.get(`${BASE_API}/workspaces/${_wsId}/custom-fields`, {
  414 |       headers: { Authorization: `Bearer ${_token}` },
  415 |     })
  416 |     expect([200, 404]).toContain(res.status())
  417 |   })
  418 | })
  419 | 
  420 | // ============================================================
  421 | // Relations & Templates (Quick Check)
  422 | // ============================================================
  423 | test.describe('Workspace Settings - Relations & Templates', () => {
  424 |   test.beforeAll(async ({ request }) => { await ensureSetup(request) })
  425 | 
  426 |   test('relations section loads with relation type list', async ({ page }) => {
  427 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  428 |     await navigateToSection(page, '关联') || await navigateToSection(page, 'Relations')
  429 |     await page.waitForTimeout(800)
  430 |     await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  431 |   })
  432 | 
  433 |   test('templates section loads', async ({ page }) => {
  434 |     await goToApp(page, `/workspace/${_wsSlug}/settings`)
  435 |     await navigateToSection(page, '模板') || await navigateToSection(page, 'Templates')
  436 |     await page.waitForTimeout(800)
  437 |     await expect(page.locator('main, .flex-1').first()).toBeVisible({ timeout: 5000 })
  438 |   })
  439 | 
  440 |   test('API: list relation types', async ({ request }) => {
  441 |     const res = await request.get(`${BASE_API}/workspaces/${_wsId}/relations/types`, {
  442 |       headers: { Authorization: `Bearer ${_token}` },
  443 |     })
  444 |     expect([200, 404]).toContain(res.status())
  445 |   })
  446 | 
  447 |   test('API: list project templates', async ({ request }) => {
  448 |     const res = await request.get(`${BASE_API}/workspaces/${_wsId}/templates`, {
  449 |       headers: { Authorization: `Bearer ${_token}` },
  450 |     })
  451 |     expect([200, 404]).toContain(res.status())
  452 |   })
  453 | })
  454 | 
```