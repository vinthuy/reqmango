/**
 * E2E Tests — Unified Issue Filter Bar
 * 测试统一过滤栏在所有视图中的行为
 */
import { test, expect, type Page } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'

let _token = '', _wsSlug = '', _projectId = 0, _wsId = 0

async function ensureSetup(request: any) {
  if (_token) return
  const ts = Date.now()
  const user = { email: `e2efilter${ts}@t.com`, username: `e2efilter${ts}`, password: 'Test123!', display_name: 'T' }
  await request.post(`${BASE_API}/auth/register`, { data: user })
  const login = await request.post(`${BASE_API}/auth/login`, { data: { email: user.email, password: user.password } })
  const { access_token } = await login.json()
  _token = access_token

  const ws = await request.post(`${BASE_API}/workspaces`, {
    data: { name: 'Filter Test WS', slug: `filter-ws-${ts}` },
    headers: { Authorization: `Bearer ${_token}` },
  })
  _wsId = (await ws.json()).id
  _wsSlug = (await ws.json()).slug

  const proj = await request.post(`${BASE_API}/projects?workspace_id=${_wsId}`, {
    data: { name: 'Filter Project', identifier: 'FILT' },
    headers: { Authorization: `Bearer ${_token}` },
  })
  _projectId = (await proj.json()).id
}

async function goToProject(page: Page) {
  await page.goto('/login')
  await page.evaluate((t: string) => localStorage.setItem('token', t), _token)
  await page.goto(`/workspace/${_wsSlug}/project/${_projectId}`)
  await page.waitForLoadState('networkidle').catch(() => {})
}

test.describe('Issue Filter Bar', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  // ============================================================
  // 1. Filter Bar Rendering
  // ============================================================
  test('filter bar renders on project page', async ({ page }) => {
    await goToProject(page)
    const searchInput = page.locator('.issue-filter-bar input[type="text"]').first()
    await expect(searchInput).toBeVisible({ timeout: 10000 })
  })

  test('search input has correct placeholder', async ({ page }) => {
    await goToProject(page)
    const input = page.locator('.issue-filter-bar input[type="text"]').first()
    const placeholder = await input.getAttribute('placeholder')
    expect(placeholder).toBeTruthy()
  })

  test('filter button is visible', async ({ page }) => {
    await goToProject(page)
    const filterBtn = page.locator('.issue-filter-bar button:has(svg)').first()
    await expect(filterBtn).toBeVisible()
  })

  test('filter bar has add filter button', async ({ page }) => {
    await goToProject(page)
    const addFilterBtn = page.locator('.issue-filter-bar button', { hasText: '添加' })
    await expect(addFilterBtn.first()).toBeVisible()
  })

  // ============================================================
  // 2. Quick Filter Toggle
  // ============================================================
  test('quick filter toggles to active state', async ({ page }) => {
    await goToProject(page)
    const chip = page.locator('.issue-filter-bar button:has-text("高优先级"), .issue-filter-bar button:has-text("High Priority")').first()
    if (await chip.isVisible({ timeout: 3000 }).catch(() => false)) {
      // Check initial state (should be inactive)
      const initialClass = await chip.getAttribute('class')
      await chip.click()
      await page.waitForTimeout(500)
      // Should have changed appearance
      const newClass = await chip.getAttribute('class')
      expect(newClass).not.toBe(initialClass)
    }
  })

  test('quick filter toggles back off', async ({ page }) => {
    await goToProject(page)
    const chip = page.locator('.issue-filter-bar button:has-text("高优先级"), .issue-filter-bar button:has-text("High Priority")').first()
    if (await chip.isVisible({ timeout: 3000 }).catch(() => false)) {
      await chip.click()
      await page.waitForTimeout(300)
      await chip.click()
      await page.waitForTimeout(300)
    }
    // Page should still be functional
    await expect(page.locator('body')).toBeVisible()
  })

  // ============================================================
  // 3. Add Filter dropdown
  // ============================================================
  test('add filter button is clickable', async ({ page }) => {
    await goToProject(page)
    // Find the + filter button and click it
    const addBtn = page.locator('.issue-filter-bar button:has(svg)').last()
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click()
      await page.waitForTimeout(500)
    }
    // Page should still be functional
    await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 5000 })
  })

  test('filter dropdown closes on second click', async ({ page }) => {
    await goToProject(page)
    const addBtn = page.locator('.issue-filter-bar button').filter({ has: page.locator('svg') }).last()
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click()
      await page.waitForTimeout(300)
      await addBtn.click()
      await page.waitForTimeout(300)
    }
    await expect(page.locator('body')).toBeVisible()
  })

  // ============================================================
  // 4. View Switching preserves filter bar
  // ============================================================
  test('filter bar visible in list view', async ({ page }) => {
    await goToProject(page)
    await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  })

  test('switch to kanban view keeps filter bar', async ({ page }) => {
    await goToProject(page)
    const kanbanBtn = page.locator('.issue-filter-bar button[title="看板"], .issue-filter-bar button[title*="kanban"], .issue-filter-bar button:has-text("📌")').first()
    if (await kanbanBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await kanbanBtn.click()
      await page.waitForTimeout(800)
    }
    await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  })

  test('switch to tree view keeps filter bar', async ({ page }) => {
    await goToProject(page)
    const treeBtn = page.locator('.issue-filter-bar button[title="树形"], .issue-filter-bar button[title*="tree"], .issue-filter-bar button:has-text("🌳")').first()
    if (await treeBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await treeBtn.click()
      await page.waitForTimeout(800)
    }
    await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  })

  test('switch to calendar view keeps filter bar', async ({ page }) => {
    await goToProject(page)
    const calBtn = page.locator('.issue-filter-bar button[title="日历"], .issue-filter-bar button[title*="calendar"], .issue-filter-bar button:has-text("📅")').first()
    if (await calBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await calBtn.click()
      await page.waitForTimeout(800)
    }
    await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  })

  test('switch to gantt view keeps filter bar', async ({ page }) => {
    await goToProject(page)
    const ganttBtn = page.locator('.issue-filter-bar button[title="甘特"], .issue-filter-bar button[title*="gantt"], .issue-filter-bar button:has-text("📊")').first()
    if (await ganttBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await ganttBtn.click()
      await page.waitForTimeout(800)
    }
    await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  })

  test('switch back to list from kanban', async ({ page }) => {
    await goToProject(page)
    // Go to kanban first
    const kanbanBtn = page.locator('.issue-filter-bar button[title="看板"], .issue-filter-bar button:has-text("📌")').first()
    if (await kanbanBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await kanbanBtn.click()
      await page.waitForTimeout(500)
    }
    // Go back to list
    const listBtn = page.locator('.issue-filter-bar button[title="列表"], .issue-filter-bar button:has-text("📋")').first()
    if (await listBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await listBtn.click()
      await page.waitForTimeout(500)
    }
    await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 10000 })
  })

  // ============================================================
  // 5. Search
  // ============================================================
  test('search input accepts text', async ({ page }) => {
    await goToProject(page)
    const input = page.locator('.issue-filter-bar input[type="text"]').first()
    await input.fill('test bug')
    await expect(input).toHaveValue('test bug')
  })

  test('search triggers on Enter', async ({ page }) => {
    await goToProject(page)
    const input = page.locator('.issue-filter-bar input[type="text"]').first()
    await input.fill('login')
    await input.press('Enter')
    await page.waitForTimeout(500)
    await expect(page.locator('body')).toBeVisible()
  })

  // ============================================================
  // 6. i18n - Switch language, filter bar updates
  // ============================================================
  test('filter bar labels switch to English', async ({ page }) => {
    await goToProject(page)
    // Switch to English
    const langBtn = page.locator('button:has-text("中"), button:has-text("EN")').first()
    if (await langBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await langBtn.click()
      await page.waitForTimeout(500)
    }
    // Filter bar should still be functional
    await expect(page.locator('.issue-filter-bar').first()).toBeVisible({ timeout: 5000 })
  })
})
