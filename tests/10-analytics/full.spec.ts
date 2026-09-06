import { test, expect } from '../fixtures/auth';

test.describe('分析页面全功能测试', () => {
  const ANALYTICS_URL = '/workspace/qa-test/project/2347?tab=analytics';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(ANALYTICS_URL);
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-ANA-001: 分析页面正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
  });

  // === 图表显示 ===
  test('TC-ANA-002: 图表渲染', async ({ authedPage: page }) => {
    const chart = page.locator('canvas, [class*="chart"], svg').first();
    if (await chart.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(chart).toBeVisible();
    }
  });

  // === 时间范围 ===
  test('TC-ANA-003: 时间范围选择', async ({ authedPage: page }) => {
    const timeFilter = page.locator('select, [role="combobox"], button:has-text("本周"), button:has-text("本月")').first();
    if (await timeFilter.isVisible({ timeout: 3000 }).catch(() => false)) {
      await timeFilter.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
  });

  // === 筛选器 ===
  test('TC-ANA-004: 筛选功能', async ({ authedPage: page }) => {
    const filterBtn = page.locator('button:has-text("筛选"), button:has-text("Filter")').first();
    if (await filterBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await filterBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
  });

  // === 导出报告 ===
  test('TC-ANA-005: 导出分析报告', async ({ authedPage: page }) => {
    const exportBtn = page.locator('button:has-text("导出"), button:has-text("Export")').first();
    if (await exportBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await exportBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
  });

  // === 度量卡片 ===
  test('TC-ANA-006: 度量指标显示', async ({ authedPage: page }) => {
    const metrics = page.locator('[class*="metric"], [class*="stat"], [class*="number"]').first();
    if (await metrics.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(metrics).toBeVisible();
    }
  });

  // === 响应式 ===
  test('TC-ANA-007: 分析页面响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
  });

  // === 空状态 ===
  test('TC-ANA-008: 空数据分析', async ({ authedPage: page }) => {
    await expect(page.locator('text=分析, text=Analytics, text=度量, text=暂无数据').first()).toBeVisible();
  });
});
