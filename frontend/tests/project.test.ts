import { test, expect } from '@playwright/test'

test.describe('项目详情页面', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login')
    await page.fill('input[type="email"]', 'demo@example.com')
    await page.fill('input[type="password"]', 'demo1234')
    await page.click('button[type="submit"]')
    await page.waitForURL('/')
  })

  test('访问项目详情页面', async ({ page }) => {
    // 访问工作空间
    await page.goto('/workspace/demo')

    // 等待工作空间加载
    await page.waitForSelector('text=项目', { timeout: 10000 })

    // 找到并点击项目卡片
    const projectCard = page.locator('.bg-white.rounded-lg').filter({ hasText: '项目' }).first()
    if (await projectCard.count() > 0) {
      await projectCard.click()
      await page.waitForURL(/\/workspace\/.*\/project\/\d+/)

      // 验证页面加载
      await expect(page.locator('text=工作项')).toBeVisible()
      console.log('项目详情页面加载成功')
    }
  })

  test('设置按钮跳转到设置页面', async ({ page }) => {
    // 直接访问项目详情页
    await page.goto('/workspace/demo/project/1')

    // 等待页面加载
    await page.waitForLoadState('networkidle')

    // 点击设置按钮
    const settingsBtn = page.locator('button', { hasText: '设置' })
    if (await settingsBtn.count() > 0) {
      await settingsBtn.click()

      // 验证跳转到设置页面
      await page.waitForURL('/workspace/demo/settings')
      console.log('跳转到设置页面成功')
    }
  })

  test('查看工作项列表', async ({ page }) => {
    // 直接访问项目详情页
    await page.goto('/workspace/demo/project/1')

    // 等待页面加载
    await page.waitForLoadState('networkidle')

    // 验证工作项管理标签页显示
    await expect(page.locator('text=工作项管理')).toBeVisible()
    console.log('工作项列表页面加载成功')
  })
})
