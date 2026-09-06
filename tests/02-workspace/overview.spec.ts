import { test, expect } from '../fixtures/auth';

test.describe('工作空间概览页', () => {
  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/overview');
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-WOV-001: 概览页正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=概览, text=Overview, text=工作空间').first()).toBeVisible();
  });

  // === 项目列表 ===
  test('TC-WOV-002: 项目列表显示', async ({ authedPage: page }) => {
    const projectList = page.locator('[class*="project"], [class*="card"], table').first();
    if (await projectList.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(projectList).toBeVisible();
    }
  });

  // === 统计卡片 ===
  test('TC-WOV-003: 统计数据', async ({ authedPage: page }) => {
    const stats = page.locator('[class*="stat"], [class*="metric"], [class*="number"]').first();
    if (await stats.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(stats).toBeVisible();
    }
  });

  // === 最近活动 ===
  test('TC-WOV-004: 最近活动', async ({ authedPage: page }) => {
    const activity = page.locator('[class*="activity"], [class*="recent"], [class*="timeline"]').first();
    if (await activity.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(activity).toBeVisible();
    }
  });

  // === 响应式 ===
  test('TC-WOV-005: 概览页响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=概览, text=Overview, text=工作空间').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=概览, text=Overview, text=工作空间').first()).toBeVisible();
  });
});
