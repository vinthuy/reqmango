import { test, expect, type Page } from '@playwright/test'

const BASE = 'http://localhost:5173'
const API = 'http://localhost:8000/api/v1'

async function login(page: Page) {
  await page.goto(`${BASE}/login`)
  await page.fill('input[type="email"]', 'admin@reqmango.com')
  await page.fill('input[type="password"]', 'demo1234')
  await page.click('button[type="submit"]')
  await page.waitForURL('**/workspace/**', { timeout: 10000 })
}

// Helper: find any issue in the first project and return its URL
async function openFirstIssue(page: Page): Promise<string> {
  // Navigate to the first project's issues list; click the first issue row.
  // This selector is intentionally generic — adjust if the list UI changes.
  await page.goto(`${BASE}/`)
  // Wait for the sidebar / project list to render, then click the first issue link
  await page.waitForTimeout(1000)
  const issueLink = page.locator('a[href*="/issues/"]').first()
  await issueLink.click()
  await page.waitForSelector('[data-test="tab-btn"]', { timeout: 10000 })
  return page.url()
}

test.describe('Chat feature', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('Chat tab renders and accepts a message', async ({ page }) => {
    await openFirstIssue(page)
    // Click the Chat tab (last tab button)
    const chatTab = page.locator('[data-test="tab-btn"]', { hasText: /聊天|Chat/ }).last()
    await chatTab.click()
    // Type and send a message
    const textarea = page.locator('textarea').first()
    await textarea.fill('E2E test message ' + Date.now())
    await textarea.press('Enter')
    // The message should appear in the list
    await expect(page.locator('text=E2E test message')).toBeVisible({ timeout: 5000 })
  })

  test('Reactions toggle on click', async ({ page }) => {
    await openFirstIssue(page)
    await page.locator('[data-test="tab-btn"]', { hasText: /聊天|Chat/ }).last().click()
    // Send a message first to react to
    const ta = page.locator('textarea').first()
    await ta.fill('Reaction target ' + Date.now())
    await ta.press('Enter')
    await expect(page.locator('text=Reaction target')).toBeVisible({ timeout: 5000 })
    // Open the emoji picker and click 👍
    await page.locator('text=😊+').first().click()
    await page.locator('button:has-text("👍")').first().click()
    await expect(page.locator('button:has-text("👍")').first()).toBeVisible({ timeout: 5000 })
  })

  test('Edit message within 30-min window', async ({ page }) => {
    await openFirstIssue(page)
    await page.locator('[data-test="tab-btn"]', { hasText: /聊天|Chat/ }).last().click()
    const ta = page.locator('textarea').first()
    const marker = 'EditTarget-' + Date.now()
    await ta.fill(marker)
    await ta.press('Enter')
    await expect(page.locator(`text=${marker}`)).toBeVisible({ timeout: 5000 })
    // Hover the message to reveal the edit button
    const msg = page.locator(`text=${marker}`).first()
    await msg.hover()
    await page.locator('[title="编辑"], [title="Edit"]').first().click()
    // A window.prompt appears — Playwright handles it via dialog handler
    await page.evaluate(() => {
      window.prompt = () => 'Edited content'
    })
    // Re-trigger edit (prompt was overridden after first click)
    await msg.hover()
    await page.locator('[title="编辑"], [title="Edit"]').first().click()
    await expect(page.locator('text=Edited content')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('text=(已编辑)|\\(edited\\)')).toBeVisible({ timeout: 5000 })
  })

  test('Multi-tab sync: message in tab1 appears in tab2', async ({ browser }) => {
    const ctx = await browser.newContext()
    const p1 = await ctx.newPage()
    const p2 = await ctx.newPage()
    await login(p1)
    await login(p2)
    await openFirstIssue(p1)
    await openFirstIssue(p2)
    await p1.locator('[data-test="tab-btn"]', { hasText: /聊天|Chat/ }).last().click()
    await p2.locator('[data-test="tab-btn"]', { hasText: /聊天|Chat/ }).last().click()
    const marker = 'MultiTab-' + Date.now()
    await p1.locator('textarea').first().fill(marker)
    await p1.locator('textarea').first().press('Enter')
    // tab2 should receive the message via SSE within 3s
    await expect(p2.locator(`text=${marker}`)).toBeVisible({ timeout: 5000 })
  })
})
