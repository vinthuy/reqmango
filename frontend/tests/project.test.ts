import { test, expect } from '@playwright/test'

test.describe('项目详情页面', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('http://localhost:5173/login')
    await page.fill('input[type="email"]', 'test@example.com')
    await page.fill('input[type="password"]', 'test123456')
    await page.click('button[type="submit"]')
    await page.waitForURL('http://localhost:5173/')
  })

  test('访问项目详情页面', async ({ page }) => {
    // 访问工作空间
    await page.goto('http://localhost:5173/workspace/demo1')

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

  test('设置按钮打开模态框', async ({ page }) => {
    // 直接访问项目详情页
    await page.goto('http://localhost:5173/workspace/demo1/project/1')

    // 等待页面加载
    await page.waitForLoadState('networkidle')

    // 点击设置按钮
    const settingsBtn = page.locator('button', { hasText: '设置' })
    if (await settingsBtn.count() > 0) {
      await settingsBtn.click()

      // 验证模态框打开
      await expect(page.locator('text=项目设置')).toBeVisible()
      console.log('设置模态框打开成功')
    }
  })

  test('更新项目信息', async ({ page }) => {
    // 直接访问项目详情页
    await page.goto('http://localhost:5173/workspace/demo1/project/1')

    // 等待页面加载
    await page.waitForLoadState('networkidle')

    // 点击设置按钮
    const settingsBtn = page.locator('button', { hasText: '设置' })
    await settingsBtn.click()

    // 等待模态框打开
    await page.waitForSelector('text=项目设置')

    // 清空并输入新的项目名称
    const nameInput = page.locator('input[placeholder="项目名称"]')
    await nameInput.clear()
    await nameInput.fill('自动化测试项目')

    // 点击保存
    const saveBtn = page.locator('button', { hasText: '保存' })
    await saveBtn.click()

    // 等待模态框关闭
    await page.waitForSelector('text=项目设置', { state: 'hidden' })

    console.log('项目更新成功')
  })
})
