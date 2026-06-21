import { test, expect } from '@playwright/test'

test.describe('工作空间功能测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.locator('input[type="email"]').fill('test@example.com')
    await page.locator('input[type="password"]').fill('test123456')
    await page.locator('button[type="submit"]').click()
    await page.waitForURL('/')
  })

  test('查看工作空间列表', async ({ page }) => {
    await expect(page.locator('h2:text("工作空间")')).toBeVisible()
    
    const workspaceCards = page.locator('.bg-white.border.border-gray-200.rounded-lg')
    await expect(workspaceCards).toHaveCountGreaterThanOrEqual(1)
    
    await expect(workspaceCards.locator('h3')).toContainText('测试工作空间')
  })

  test('点击工作空间进入详情', async ({ page }) => {
    const workspaceCards = page.locator('.bg-white.border.border-gray-200.rounded-lg')
    const testWorkspace = workspaceCards.filter({ hasText: '测试工作空间' })
    await testWorkspace.click()
    
    await page.waitForURL('/workspace/test-workspace')
    await expect(page.locator('h1')).toContainText('测试工作空间')
  })

  test('创建新工作空间', async ({ page }) => {
    const initialCount = await page.locator('.bg-white.border.border-gray-200.rounded-lg').count()
    
    await page.locator('button:text("创建工作空间")').click()
    
    await page.locator('input[placeholder="工作空间名称"]').fill('新工作空间')
    await page.locator('input[placeholder="url-slug"]').fill('new-workspace-' + Date.now())
    await page.locator('div.bg-white.rounded-xl button:text("创建")').click()
    
    await page.waitForTimeout(500)
    const workspaceCards = page.locator('.bg-white.border.border-gray-200.rounded-lg')
    await expect(workspaceCards).toHaveCount(initialCount + 1)
  })
})