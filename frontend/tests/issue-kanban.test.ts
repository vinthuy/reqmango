import { test, expect } from '@playwright/test'

test.describe('工作项看板视图测试', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login')
    await page.fill('input[type="email"]', 'demo@example.com')
    await page.fill('input[type="password"]', 'demo1234')
    await page.click('button[type="submit"]')
    await page.waitForURL('/')
    
    // 进入项目页面（看板视图）
    await page.goto('/workspace/demo/project/1?view=kanban')
    await page.waitForLoadState('networkidle')
  })

  test('看板视图页面加载', async ({ page }) => {
    // 等待看板加载
    await page.waitForSelector('.issue-kanban', { timeout: 10000 })
    
    // 验证看板视图可见
    await expect(page.locator('.issue-kanban')).toBeVisible()
    
    console.log('工作项看板页面加载成功')
  })

  test('看板列显示', async ({ page }) => {
    // 等待状态列加载
    await page.waitForSelector('text=待处理', { timeout: 10000 }).catch(() => {})
    
    // 检查是否有状态列
    const columns = page.locator('.bg-gray-100.rounded-lg.p-3')
    const columnCount = await columns.count()
    
    console.log(`看板显示 ${columnCount} 个状态列`)
    expect(columnCount).toBeGreaterThan(0)
  })

  test('搜索功能', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 输入搜索关键词
    const searchInput = page.locator('input[placeholder*="搜索工作项"]')
    await searchInput.fill('登录')
    await searchInput.press('Enter')
    
    await page.waitForTimeout(500)
    console.log('看板搜索功能测试完成')
  })

  test('状态筛选', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 选择一个状态进行筛选
    const stateSelect = page.locator('select').filter({ hasText: '所有状态' }).first()
    const optionCount = await stateSelect.locator('option').count()
    if (optionCount > 1) {
      await stateSelect.selectOption({ index: 1 })
    }
    
    await page.waitForTimeout(500)
    console.log('看板状态筛选测试完成')
  })

  test('优先级筛选', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 选择一个优先级
    const prioritySelect = page.locator('select').filter({ hasText: '所有优先级' }).first()
    await prioritySelect.selectOption('high')
    
    await page.waitForTimeout(500)
    console.log('看板优先级筛选测试完成')
  })

  test('高级搜索展开', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 点击高级搜索按钮
    await page.locator('button:has-text("高级搜索")').click()
    
    // 验证高级搜索区域展开 - 使用更精确的选择器
    await expect(page.locator('.bg-gray-50 label:has-text("周期")')).toBeVisible()
    await expect(page.locator('.bg-gray-50 label:has-text("负责人")')).toBeVisible()
    
    console.log('看板高级搜索展开测试完成')
  })

  test('重置筛选', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 先打开高级搜索
    await page.locator('button:has-text("高级搜索")').click()
    await page.waitForSelector('text=周期')
    
    // 点击重置
    await page.locator('button:has-text("重置筛选")').click()
    
    await page.waitForTimeout(300)
    console.log('看板重置筛选测试完成')
  })

  test('点击工作项卡片打开详情', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 查找第一个工作项卡片
    const card = page.locator('.bg-white.rounded-md.border').first()
    
    if (await card.count() > 0) {
      await card.click()
      await page.waitForTimeout(500)
      console.log('点击工作项卡片测试完成')
    } else {
      console.log('没有工作项卡片可测试')
    }
  })

  test('工作项卡片显示信息', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 检查卡片内容
    const cards = page.locator('.bg-white.rounded-md.border')
    const cardCount = await cards.count()
    
    if (cardCount > 0) {
      // 检查卡片是否包含编号
      const firstCard = cards.first()
      const hasSequenceId = await firstCard.locator('text=DEMO-').count() > 0
      expect(hasSequenceId).toBeTruthy()
      
      console.log(`工作项卡片显示测试完成，共 ${cardCount} 个卡片`)
    } else {
      // 检查空状态提示
      await expect(page.locator('text=拖放工作项到此处').first()).toBeVisible()
      console.log('看板为空状态')
    }
  })

  test('详情按钮', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    
    // 查找详情按钮
    const detailBtn = page.locator('button:has-text("详情")').first()
    
    if (await detailBtn.count() > 0) {
      await detailBtn.click()
      await page.waitForTimeout(500)
      console.log('看板详情按钮测试完成')
    } else {
      console.log('没有详情按钮可测试')
    }
  })
})
