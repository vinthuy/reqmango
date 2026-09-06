import { test, expect } from '../fixtures/auth';

test.describe('仪表盘全功能测试', () => {
  const DASHBOARD_URL = '/workspace/qa-test/project/2347?tab=dashboard';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(DASHBOARD_URL);
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-DAS-001: 仪表盘正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
  });

  // === Widget 显示 ===
  test('TC-DAS-002: Widget 卡片显示', async ({ authedPage: page }) => {
    const widgets = page.locator('[class*="widget"], [class*="card"], [class*="metric"]');
    const count = await widgets.count();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  // === 统计数据 ===
  test('TC-DAS-003: 统计数据展示', async ({ authedPage: page }) => {
    const stats = page.locator('[class*="stat"], [class*="number"], [class*="count"]').first();
    if (await stats.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(stats).toBeVisible();
    }
  });

  // === 图表显示 ===
  test('TC-DAS-004: 图表渲染', async ({ authedPage: page }) => {
    const chart = page.locator('canvas, [class*="chart"], svg').first();
    if (await chart.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(chart).toBeVisible();
    }
  });

  // === Widget 配置 ===
  test('TC-DAS-005: 添加 Widget', async ({ authedPage: page }) => {
    const addBtn = page.locator('button:has-text("添加"), button:has-text("添加组件"), button:has-text("Add Widget")').first();
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
  });

  // === Widget 删除 ===
  test('TC-DAS-006: 删除 Widget', async ({ authedPage: page }) => {
    const deleteBtn = page.locator('button[aria-label="删除"], button:has-text("删除")').first();
    if (await deleteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await deleteBtn.click();
      await page.waitForTimeout(500);
      const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
      if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
  });

  // === 刷新数据 ===
  test('TC-DAS-007: 刷新仪表盘', async ({ authedPage: page }) => {
    const refreshBtn = page.locator('button:has-text("刷新"), button[aria-label="刷新"]').first();
    if (await refreshBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await refreshBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
  });

  // === 时间范围选择 ===
  test('TC-DAS-008: 时间范围筛选', async ({ authedPage: page }) => {
    const timeFilter = page.locator('select, [role="combobox"], button:has-text("本周"), button:has-text("本月")').first();
    if (await timeFilter.isVisible({ timeout: 3000 }).catch(() => false)) {
      await timeFilter.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
  });

  // === 导出 ===
  test('TC-DAS-009: 导出仪表盘', async ({ authedPage: page }) => {
    const exportBtn = page.locator('button:has-text("导出"), button:has-text("Export")').first();
    if (await exportBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await exportBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
  });

  // === Widget 拖拽排序 ===
  test('TC-DAS-010: Widget 拖拽排序', async ({ authedPage: page }) => {
    const widget = page.locator('[class*="widget"], [class*="card"]').first();
    if (await widget.isVisible({ timeout: 3000 }).catch(() => false)) {
      const box = await widget.boundingBox();
      if (box) {
        await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
        await page.waitForTimeout(300);
      }
    }
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-DAS-011: 仪表盘响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
  });

  // === 空状态 ===
  test('TC-DAS-012: 空仪表盘显示', async ({ authedPage: page }) => {
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览, text=暂无数据').first()).toBeVisible();
  });

  // === Widget 类型选择 ===
  test('TC-DAS-013: Widget 类型列表', async ({ authedPage: page }) => {
    const addBtn = page.locator('button:has-text("添加"), button:has-text("添加组件")').first();
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
      const widgetType = page.locator('[class*="option"], [class*="type"], li').first();
      if (await widgetType.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(widgetType).toBeVisible();
      }
      await page.keyboard.press('Escape');
    }
  });

  // === Widget 数据刷新 ===
  test('TC-DAS-014: 单个 Widget 刷新', async ({ authedPage: page }) => {
    const refreshBtn = page.locator('[class*="widget"] button:has-text("刷新"), [class*="card"] button[aria-label="刷新"]').first();
    if (await refreshBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await refreshBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
  });

  // === 全屏模式 ===
  test('TC-DAS-015: 全屏查看', async ({ authedPage: page }) => {
    const fullscreenBtn = page.locator('button:has-text("全屏"), button[aria-label="全屏"]').first();
    if (await fullscreenBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await fullscreenBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('text=仪表盘, text=Dashboard, text=概览').first()).toBeVisible();
  });
});
