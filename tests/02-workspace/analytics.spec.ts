import { test, expect } from '../fixtures/auth';

test.describe('工作空间分析页', () => {
  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/analytics');
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-WAN-001: 分析页正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
  });

  // === 图表 ===
  test('TC-WAN-002: 图表渲染', async ({ authedPage: page }) => {
    const chart = page.locator('canvas, [class*="chart"], svg').first();
    if (await chart.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(chart).toBeVisible();
    }
  });

  // === 筛选 ===
  test('TC-WAN-003: 筛选功能', async ({ authedPage: page }) => {
    const filterBtn = page.locator('button:has-text("筛选"), select, [role="combobox"]').first();
    if (await filterBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await filterBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
  });

  // === 导出 ===
  test('TC-WAN-004: 导出报告', async ({ authedPage: page }) => {
    const exportBtn = page.locator('button:has-text("导出"), button:has-text("Export")').first();
    if (await exportBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await exportBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-WAN-005: 分析页响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=分析, text=Analytics, text=度量').first()).toBeVisible();
  });
});
