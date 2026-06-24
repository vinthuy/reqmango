import { test, expect } from '@playwright/test'

test.describe('工作项列表视图测试', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login')
    await page.fill('input[type="email"]', 'demo@example.com')
    await page.fill('input[type="password"]', 'demo1234')
    await page.click('button[type="submit"]')
    await page.waitForURL('/')
    
    // 进入项目页面
    await page.goto('/workspace/demo/project/1')
    await page.waitForLoadState('networkidle')
  })

  test('列表视图页面加载', async ({ page }) => {
    // 等待工作项列表加载
    await page.waitForSelector('table', { timeout: 10000 })
    
    // 验证列表包含表头
    await expect(page.locator('th:has-text("编号")')).toBeVisible()
    await expect(page.locator('th:has-text("标题")')).toBeVisible()
    await expect(page.locator('th:has-text("优先级")')).toBeVisible()
    
    console.log('工作项列表页面加载成功')
  })

  test('列表视图显示工作项', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 检查是否有工作项数据或空状态
    const tableRows = page.locator('tbody tr')
    const rowCount = await tableRows.count()
    
    // 应该有工作项数据或显示"暂无工作项"
    if (rowCount > 0) {
      console.log(`列表显示 ${rowCount} 个工作项`)
    } else {
      await expect(page.locator('text=暂无工作项')).toBeVisible()
    }
  })

  test('搜索功能', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 输入搜索关键词
    const searchInput = page.locator('input[placeholder*="搜索工作项"]')
    await searchInput.fill('用户')
    await searchInput.press('Enter')
    
    await page.waitForTimeout(500)
    console.log('搜索功能测试完成')
  })

  test('状态筛选', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 选择一个状态进行筛选
    const stateSelect = page.locator('select').filter({ hasText: '所有状态' }).first()
    await stateSelect.selectOption({ index: 1 })
    
    await page.waitForTimeout(500)
    console.log('状态筛选功能测试完成')
  })

  test('优先级筛选', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 选择一个优先级
    const prioritySelect = page.locator('select').filter({ hasText: '所有优先级' }).first()
    await prioritySelect.selectOption('high')
    
    await page.waitForTimeout(500)
    console.log('优先级筛选功能测试完成')
  })

  test('高级搜索展开', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 点击高级搜索按钮
    await page.locator('button:has-text("高级搜索")').click()
    
    // 验证高级搜索区域展开 - 使用更精确的选择器
    await expect(page.locator('.bg-gray-50 label:has-text("周期")')).toBeVisible()
    await expect(page.locator('.bg-gray-50 label:has-text("负责人")')).toBeVisible()
    
    console.log('高级搜索展开测试完成')
  })

  test('重置筛选', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 先打开高级搜索
    await page.locator('button:has-text("高级搜索")').click()
    await page.waitForSelector('text=周期')
    
    // 点击重置
    await page.locator('button:has-text("重置筛选")').click()
    
    await page.waitForTimeout(300)
    console.log('重置筛选测试完成')
  })

  test('新建按钮跳转', async ({ page }) => {
    await page.waitForLoadState('networkidle')

    // 点击新建按钮 - 使用更精确的选择器
    const createBtn = page.locator('button:text-is("新建工作项")')
    await createBtn.click()

    // 验证跳转到新建页面
    await page.waitForURL(/\/issues\/new/)
    console.log('新建按钮跳转测试完成')
  })

  test('查看按钮打开详情面板', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 检查是否有工作项
    const viewBtn = page.locator('button:has-text("查看")').first()
    if (await viewBtn.count() > 0) {
      await viewBtn.click()
      
      // 等待详情面板打开
      await page.waitForTimeout(500)
      console.log('查看按钮测试完成')
    } else {
      console.log('没有工作项可测试')
    }
  })
})
