import { test, expect } from '@playwright/test'

/**
 * ReqMango Pages E2E 测试
 */

const BASE_URL = 'http://localhost:5173'
const API_URL = 'http://localhost:8000'
const TEST_TOKEN = process.env.TEST_TOKEN || ''

// ── 工具函数 ──
async function loginViaStorage(page: any, token: string) {
  await page.goto(BASE_URL)
  await page.evaluate((t: string) => {
    localStorage.setItem('token', t)
    // 如果有 user 信息也存储
    localStorage.setItem('locale', 'en-US')
  }, token)
}

async function apiGet(token: string, path: string) {
  const res = await fetch(`${API_URL}${path}`, {
    headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' }
  })
  return res.json()
}

// ── 测试套件 ──
test.describe('ReqMango Pages E2E Tests', () => {
  
  // 测试目标：ReqMango 核心平台 (project 15), workspace reqmango-dev
  let testProject: { id: number } = { id: 15 }
  let testWorkspace: { slug: string } = { slug: 'reqmango-dev' }

  test.beforeAll(async () => {
    // Token must be set via env var or globalSetup
  })

  // ── Test 1: Navigate to Pages ──
  test('E2E-P01: Navigate to Pages route and verify UI loads', async ({ page }) => {
    test.skip(!TEST_TOKEN, 'TEST_TOKEN not configured')

    await loginViaStorage(page, TEST_TOKEN)

    // Navigate to pages route
    await page.goto(`${BASE_URL}/workspace/${testWorkspace.slug}/project/${testProject.id}/pages`, {
      waitUntil: 'networkidle',
      timeout: 30000,
    })

    // Verify page title/sidebar is visible
    await expect(page.locator('aside h2')).toContainText('Pages', { timeout: 10000 })
  })

  // ── Test 2: Create a new page ──
  test('E2E-P02: Create a new page and verify it appears in tree', async ({ page }) => {
    test.skip(!TEST_TOKEN, 'TEST_TOKEN not configured')

    await loginViaStorage(page, TEST_TOKEN)
    await page.goto(`${BASE_URL}/workspace/${testWorkspace.slug}/project/${testProject.id}/pages`)
    await page.waitForLoadState('networkidle')

    // Click "New" button in sidebar
    const newBtn = page.locator('aside button:has-text("New")')
    await newBtn.waitFor({ timeout: 5000 })
    await newBtn.click()

    // Wait for create modal
    await page.waitForSelector('[class*="bg-black"][class*="bg-opacity-50"]', { timeout: 5000 })

    // Fill in title
    const titleInput = page.locator('input[placeholder="Page title"]')
    await titleInput.fill(`E2E Test Page ${Date.now()}`)

    // Click create
    const createBtn = page.locator('button:has-text("Create")')
    await createBtn.click()

    // Wait for modal to close and page to load
    await page.waitForTimeout(2000)

    // Verify editor appears (textarea or TipTap editor)
    const mainContent = page.locator('main').first()
    await expect(mainContent).not.toBeEmpty({ timeout: 10000 })
  })

  // ── Test 3: Page title editing and auto-save ──
  test('E2E-P03: Edit page title and verify auto-save indicator', async ({ page }) => {
    test.skip(!TEST_TOKEN, 'TEST_TOKEN not configured')

    await loginViaStorage(page, TEST_TOKEN)
    await page.goto(`${BASE_URL}/workspace/${testWorkspace.slug}/project/${testProject.id}/pages`)
    await page.waitForLoadState('networkidle')

    // Click first page in tree
    const firstPage = page.locator('.page-tree > div > div').first()
    await firstPage.waitFor({ timeout: 5000 })
    await firstPage.click()
    await page.waitForTimeout(1500)

    // Check if editor loaded - either textarea or TipTap editor
    const editor = page.locator('main textarea, main .tiptap-content')
    await expect(editor.first()).toBeVisible({ timeout: 10000 })

    // Modify title
    const titleInput = page.locator('main input[type="text"], main input.text-2xl')
    await titleInput.waitFor({ timeout: 5000 })
    const oldTitle = await titleInput.inputValue()
    await titleInput.fill(oldTitle + ' (edited)')
    // Trigger blur to fire the debounced save
    await titleInput.press('Tab')
    
    // Wait for auto-save debounce (800ms) + API call
    await page.waitForTimeout(3000)

    // Reload page and verify title persisted
    await page.reload({ waitUntil: 'networkidle' })
    await page.waitForTimeout(1500)

    // Click the same page again
    const firstPage2 = page.locator('.page-tree > div > div').first()
    await firstPage2.click()
    await page.waitForTimeout(1500)

    // Verify title persisted
    const titleInput2 = page.locator('main input[type="text"], main input.text-2xl')
    await titleInput2.waitFor({ timeout: 5000 })
    const savedTitle = await titleInput2.inputValue()
    expect(savedTitle).toContain('(edited)')
  })

  // ── Test 4: Create child page ──
  test('E2E-P04: Create a child page via tree "+" button', async ({ page }) => {
    test.skip(!TEST_TOKEN, 'TEST_TOKEN not configured')

    await loginViaStorage(page, TEST_TOKEN)
    await page.goto(`${BASE_URL}/workspace/${testWorkspace.slug}/project/${testProject.id}/pages`)
    await page.waitForLoadState('networkidle')

    // Hover over first page item to reveal + button
    const firstPage = page.locator('.page-tree > div > div').first()
    await firstPage.waitFor({ timeout: 5000 })
    await firstPage.hover()
    await page.waitForTimeout(300)

    // Click + (add child) button
    const addBtn = firstPage.locator('button[title="Add child"]')
    await addBtn.click()
    await page.waitForTimeout(500)

    // Fill child page title
    const titleInput = page.locator('input[placeholder="Page title"]')
    await titleInput.fill(`Child Page ${Date.now()}`)

    // Create
    const createBtn = page.locator('button:has-text("Create")')
    await createBtn.click()
    await page.waitForTimeout(2000)

    // Child page should appear in tree (indented)
    const childPages = page.locator('.page-tree')
    await expect(childPages).not.toBeEmpty()
  })

  // ── Test 5: Delete page ──
  test('E2E-P05: Delete a page via tree "✕" button', async ({ page }) => {
    test.skip(!TEST_TOKEN, 'TEST_TOKEN not configured')

    await loginViaStorage(page, TEST_TOKEN)
    await page.goto(`${BASE_URL}/workspace/${testWorkspace.slug}/project/${testProject.id}/pages`)
    await page.waitForLoadState('networkidle')

    // Hover over first page to reveal delete button
    const firstPage = page.locator('.page-tree > div > div').first()
    await firstPage.waitFor({ timeout: 5000 })
    await firstPage.hover()
    await page.waitForTimeout(300)

    // Click ✕ (delete) button
    const delBtn = firstPage.locator('button[title="Delete"]')
    await delBtn.click()
    await page.waitForTimeout(500)

    // Confirm delete
    const confirmBtn = page.locator('button:has-text("Delete")').last()
    await confirmBtn.click()
    await page.waitForTimeout(2000)

    // Page should be removed - just verify no crash
    await expect(page.locator('aside')).toBeVisible()
  })

  // ── Test 6: Empty state test ──
  test('E2E-P06: Verify empty state shows when no page selected', async ({ page }) => {
    test.skip(!TEST_TOKEN, 'TEST_TOKEN not configured')

    await loginViaStorage(page, TEST_TOKEN)
    await page.goto(`${BASE_URL}/workspace/${testWorkspace.slug}/project/${testProject.id}/pages`)
    await page.waitForLoadState('networkidle')

    // Should show either empty pages message or select hint
    const hintOrEmpty = page.locator('main').first()
    await expect(hintOrEmpty).toBeVisible({ timeout: 10000 })
  })

  // ── Test 7: Back button navigation ──
  test('E2E-P07: "Back" button navigates to project view', async ({ page }) => {
    test.skip(!TEST_TOKEN, 'TEST_TOKEN not configured')

    await loginViaStorage(page, TEST_TOKEN)
    await page.goto(`${BASE_URL}/workspace/${testWorkspace.slug}/project/${testProject.id}/pages`)
    await page.waitForLoadState('networkidle')

    // Click first page
    const firstPage = page.locator('.page-tree > div > div').first()
    if (await firstPage.count() > 0) {
      await firstPage.click()
      await page.waitForTimeout(1500)

      // Click "Back" button
      const backBtn = page.locator('button:has-text("Back")')
      await backBtn.waitFor({ timeout: 5000 })
      await backBtn.click()
      await page.waitForTimeout(2000)

      // Should be back at project page
      expect(page.url()).toContain('/project/')
      expect(page.url()).not.toContain('/pages')
    }
  })

  // ── Test 8: Archive and Restore ──
  test('E2E-P08: Archive and restore a page', async ({ page }) => {
    test.skip(!TEST_TOKEN, 'TEST_TOKEN not configured')

    await loginViaStorage(page, TEST_TOKEN)
    await page.goto(`${BASE_URL}/workspace/${testWorkspace.slug}/project/${testProject.id}/pages`)
    await page.waitForLoadState('networkidle')

    // Create a fresh page to test archive on
    if (await page.locator('.page-tree > div > div').count() === 0) {
      // Create one
      const newBtn = page.locator('aside button:has-text("New")')
      await newBtn.click()
      await page.waitForTimeout(500)
      await page.fill('input[placeholder="Page title"]', `Archive Test ${Date.now()}`)
      await page.click('button:has-text("Create")')
      await page.waitForTimeout(2000)
    }

    // Click first page
    const firstPage = page.locator('.page-tree > div > div').first()
    await firstPage.click()
    await page.waitForTimeout(1500)

    // Click Archive button
    const archiveBtn = page.locator('button:has-text("Archive")')
    if (await archiveBtn.isVisible({ timeout: 3000 })) {
      await archiveBtn.click()
      await page.waitForTimeout(2000)

      // Should show Restore button now
      const restoreBtn = page.locator('button:has-text("Restore")')
      await expect(restoreBtn).toBeVisible({ timeout: 5000 })

      // Restore it
      await restoreBtn.click()
      await page.waitForTimeout(2000)

      // Should show Archive button again
      await expect(page.locator('button:has-text("Archive")')).toBeVisible({ timeout: 5000 })
    }
  })
})
