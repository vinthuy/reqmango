import { test, expect } from '@playwright/test'

test.describe('登录页面测试', () => {
  test('页面加载正常', async ({ page }) => {
    await page.goto('/login')
    
    await expect(page).toHaveTitle('Reqman AI - Project Management')
    await expect(page.locator('h1')).toContainText('Reqman AI')
    await expect(page.locator('p:text("登录您的账户")')).toBeVisible()
  })

  test('空表单提交验证', async ({ page }) => {
    await page.goto('/login')
    
    await page.locator('button[type="submit"]').click()
    
    await expect(page.locator('div.bg-red-50')).toContainText('请填写所有字段')
  })

  test('错误密码登录', async ({ page }) => {
    await page.goto('/login')
    
    await page.locator('input[type="email"]').fill('test@example.com')
    await page.locator('input[type="password"]').fill('wrongpassword')
    await page.locator('button[type="submit"]').click()
    
    await page.waitForTimeout(1000)
    const errorDiv = page.locator('div.bg-red-50')
    await expect(errorDiv).toBeVisible()
    await expect(errorDiv).toContainText('登录失败，请检查邮箱和密码')
  })

  test('成功登录后跳转到首页', async ({ page }) => {
    await page.goto('/login')
    
    await page.locator('input[type="email"]').fill('test@example.com')
    await page.locator('input[type="password"]').fill('test123456')
    
    const responsePromise = page.waitForResponse(response => 
      response.url().includes('/api/v1/auth/login') && response.status() === 200
    )
    
    await page.locator('button[type="submit"]').click()
    
    await responsePromise
    await page.waitForLoadState('networkidle')
    
    await expect(page).toHaveURL('/')
  })

  test('跳转到注册页面', async ({ page }) => {
    await page.goto('/login')
    
    await page.locator('text=注册').click()
    
    await page.waitForNavigation()
    await expect(page).toHaveURL('/register')
  })
})