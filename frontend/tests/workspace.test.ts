import { test, expect } from '@playwright/test'

test.describe('工作空间功能测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.locator('input[type="email"]').fill('demo@example.com')
    await page.locator('input[type="password"]').fill('demo1234')
    await page.locator('button[type="submit"]').click()
    await page.waitForURL('/')
  })

  test('查看工作空间列表', async ({ page }) => {
    await expect(page.locator('h2:text("工作空间")')).toBeVisible()
    
    const workspaceCards = page.locator('.bg-white.border.border-gray-200.rounded-lg')
    const count = await workspaceCards.count()
    expect(count).toBeGreaterThanOrEqual(1)
    
    await expect(workspaceCards.first().locator('h3')).toContainText('Demo Workspace')
  })

  test('点击工作空间进入详情', async ({ page }) => {
    const workspaceCards = page.locator('.bg-white.border.border-gray-200.rounded-lg')
    const testWorkspace = workspaceCards.filter({ hasText: 'Demo Workspace' })
    await testWorkspace.click()
    
    await page.waitForURL('/workspace/demo')
    await expect(page.locator('h1')).toContainText('Demo Workspace')
  })

  test('创建新工作空间', async ({ page }) => {
    const workspaceName = '新工作空间' + Date.now()
    
    await page.locator('button:text("创建工作空间")').click()
    
    await page.locator('input[placeholder="工作空间名称"]').fill(workspaceName)
    await page.locator('input[placeholder="url-slug"]').fill('new-workspace-' + Date.now())
    await page.locator('div.bg-white.rounded-xl button:text("创建")').click()
    
    await page.waitForTimeout(500)
    const workspaceCards = page.locator('.bg-white.border.border-gray-200.rounded-lg')
    await expect(workspaceCards.filter({ hasText: workspaceName })).toBeVisible()
  })
})
