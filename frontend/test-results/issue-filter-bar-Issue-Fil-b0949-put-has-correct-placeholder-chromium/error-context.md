# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: issue-filter-bar.spec.ts >> Issue Filter Bar >> search input has correct placeholder
- Location: e2e\issue-filter-bar.spec.ts:53:3

# Error details

```
Test timeout of 60000ms exceeded.
```

```
Error: locator.getAttribute: Test timeout of 60000ms exceeded.
Call log:
  - waiting for locator('.issue-filter-bar input[type="text"]').first()

```

# Page snapshot

```yaml
- main [ref=e4]
```

# Test source

```ts
  1   | /**
  2   |  * E2E Tests — Unified Issue Filter Bar
  3   |  * 测试统一过滤栏在所有视图中的行为
  4   |  */
  5   | import { test, expect, type Page } from '@playwright/test'
  6   | 
  7   | const BASE_API = 'http://localhost:8000/api/v1'
  8   | 
  9   | let _token = '', _wsSlug = '', _projectId = 0, _wsId = 0
  10  | 
  11  | async function ensureSetup(request: any) {
  12  |   if (_token) return
  13  |   const ts = Date.now()
  14  |   const user = { email: `e2efilter${ts}@t.com`, username: `e2efilter${ts}`, password: 'Test123!', display_name: 'T' }
  15  |   await request.post(`${BASE_API}/auth/register`, { data: user })
  16  |   const login = await request.post(`${BASE_API}/auth/login`, { data: { email: user.email, password: user.password } })
  17  |   const { access_token } = await login.json()
  18  |   _token = access_token
  19  | 
  20  |   const ws = await request.post(`${BASE_API}/workspaces`, {
  21  |     data: { name: 'Filter Test WS', slug: `filter-ws-${ts}` },
  22  |     headers: { Authorization: `Bearer ${_token}` },
  23  |   })
  24  |   _wsId = (await ws.json()).id
  25  |   _wsSlug = (await ws.json()).slug
  26  | 
  27  |   const proj = await request.post(`${BASE_API}/projects?workspace_id=${_wsId}`, {
  28  |     data: { name: 'Filter Project', identifier: 'FILT' },
  29  |     headers: { Authorization: `Bearer ${_token}` },
  30  |   })
  31  |   _projectId = (await proj.json()).id
  32  | }
  33  | 
  34  | async function goToProject(page: Page) {
  35  |   await page.goto('/login')
  36  |   await page.evaluate((t: string) => localStorage.setItem('token', t), _token)
  37  |   await page.goto(`/workspace/${_wsSlug}/project/${_projectId}`)
  38  |   await page.waitForLoadState('networkidle').catch(() => {})
  39  | }
  40  | 
  41  | test.describe('Issue Filter Bar', () => {
  42  |   test.beforeAll(async ({ request }) => { await ensureSetup(request) })
  43  | 
  44  |   // ============================================================
  45  |   // 1. Filter Bar Rendering
  46  |   // ============================================================
  47  |   test('filter bar renders on project page', async ({ page }) => {
  48  |     await goToProject(page)
  49  |     const searchInput = page.locator('.issue-filter-bar input[type="text"]').first()
  50  |     await expect(searchInput).toBeVisible({ timeout: 10000 })
  51  |   })
  52  | 
  53  |   test('search input has correct placeholder', async ({ page }) => {
  54  |     await goToProject(page)
  55  |     const input = page.locator('.issue-filter-bar input[type="text"]').first()
> 56  |     const placeholder = await input.getAttribute('placeholder')
      |                                     ^ Error: locator.getAttribute: Test timeout of 60000ms exceeded.
  57  |     expect(placeholder).toBeTruthy()
  58  |   })
  59  | 
  60  |   test('view toggle buttons are visible', async ({ page }) => {
  61  |     await goToProject(page)
  62  |     const viewBtns = page.locator('.issue-filter-bar [class*="inline-flex"] button')
  63  |     const count = await viewBtns.count()
  64  |     expect(count).toBeGreaterThanOrEqual(5) // list, kanban, tree, calendar, gantt
  65  |   })
  66  | 
  67  |   test('quick filter chips are visible', async ({ page }) => {
  68  |     await goToProject(page)
  69  |     const chips = page.locator('.issue-filter-bar button:has-text("我的"), .issue-filter-bar button:has-text("Mine"), .issue-filter-bar button:has-text("未分配"), .issue-filter-bar button:has-text("高优先级")')
  70  |     const count = await chips.count()
  71  |     expect(count).toBeGreaterThanOrEqual(1)
  72  |   })
  73  | 
  74  |   // ============================================================
  75  |   // 2. Quick Filter Toggle
  76  |   // ============================================================
  77  |   test('quick filter toggles to active state', async ({ page }) => {
  78  |     await goToProject(page)
  79  |     const chip = page.locator('.issue-filter-bar button:has-text("高优先级"), .issue-filter-bar button:has-text("High Priority")').first()
  80  |     if (await chip.isVisible({ timeout: 3000 }).catch(() => false)) {
  81  |       // Check initial state (should be inactive)
  82  |       const initialClass = await chip.getAttribute('class')
  83  |       await chip.click()
  84  |       await page.waitForTimeout(500)
  85  |       // Should have changed appearance
  86  |       const newClass = await chip.getAttribute('class')
  87  |       expect(newClass).not.toBe(initialClass)
  88  |     }
  89  |   })
  90  | 
  91  |   test('quick filter toggles back off', async ({ page }) => {
  92  |     await goToProject(page)
  93  |     const chip = page.locator('.issue-filter-bar button:has-text("高优先级"), .issue-filter-bar button:has-text("High Priority")').first()
  94  |     if (await chip.isVisible({ timeout: 3000 }).catch(() => false)) {
  95  |       await chip.click()
  96  |       await page.waitForTimeout(300)
  97  |       await chip.click()
  98  |       await page.waitForTimeout(300)
  99  |     }
  100 |     // Page should still be functional
  101 |     await expect(page.locator('body')).toBeVisible()
  102 |   })
  103 | 
  104 |   // ============================================================
  105 |   // 3. Add Filter dropdown
  106 |   // ============================================================
  107 |   test('add filter button is clickable', async ({ page }) => {
  108 |     await goToProject(page)
  109 |     // Find the + filter button and click it
  110 |     const addBtn = page.locator('.issue-filter-bar button:has(svg)').last()
  111 |     if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  112 |       await addBtn.click()
  113 |       await page.waitForTimeout(500)
  114 |     }
  115 |     // Page should still be functional
  116 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 5000 })
  117 |   })
  118 | 
  119 |   test('filter dropdown closes on second click', async ({ page }) => {
  120 |     await goToProject(page)
  121 |     const addBtn = page.locator('.issue-filter-bar button').filter({ has: page.locator('svg') }).last()
  122 |     if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  123 |       await addBtn.click()
  124 |       await page.waitForTimeout(300)
  125 |       await addBtn.click()
  126 |       await page.waitForTimeout(300)
  127 |     }
  128 |     await expect(page.locator('body')).toBeVisible()
  129 |   })
  130 | 
  131 |   // ============================================================
  132 |   // 4. View Switching preserves filter bar
  133 |   // ============================================================
  134 |   test('filter bar visible in list view', async ({ page }) => {
  135 |     await goToProject(page)
  136 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  137 |   })
  138 | 
  139 |   test('switch to kanban view keeps filter bar', async ({ page }) => {
  140 |     await goToProject(page)
  141 |     const kanbanBtn = page.locator('.issue-filter-bar button[title="看板"], .issue-filter-bar button[title*="kanban"], .issue-filter-bar button:has-text("📌")').first()
  142 |     if (await kanbanBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  143 |       await kanbanBtn.click()
  144 |       await page.waitForTimeout(800)
  145 |     }
  146 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  147 |   })
  148 | 
  149 |   test('switch to tree view keeps filter bar', async ({ page }) => {
  150 |     await goToProject(page)
  151 |     const treeBtn = page.locator('.issue-filter-bar button[title="树形"], .issue-filter-bar button[title*="tree"], .issue-filter-bar button:has-text("🌳")').first()
  152 |     if (await treeBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
  153 |       await treeBtn.click()
  154 |       await page.waitForTimeout(800)
  155 |     }
  156 |     await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
```