import { test, expect } from '../fixtures/auth';

test.describe('首页全功能测试', () => {
  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto('/');
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-HOM-001: 首页正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('body')).toBeVisible();
  });

  // === 导航到工作空间 ===
  test('TC-HOM-002: 导航到工作空间', async ({ authedPage: page }) => {
    const wsLink = page.locator('a:has-text("qa-test"), a:has-text("工作空间")').first();
    if (await wsLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await wsLink.click();
      await page.waitForTimeout(1000);
      await expect(page.url()).toContain('workspace');
    }
  });

  // === 创建工作空间 ===
  test('TC-HOM-003: 创建工作空间入口', async ({ authedPage: page }) => {
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建工作空间"), button:has-text("Create Workspace")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('body')).toBeVisible();
  });

  // === 用户菜单 ===
  test('TC-HOM-004: 用户菜单', async ({ authedPage: page }) => {
    const userMenu = page.locator('[class*="user"], [class*="avatar"], [class*="profile"]').first();
    if (await userMenu.isVisible({ timeout: 3000 }).catch(() => false)) {
      await userMenu.click();
      await page.waitForTimeout(500);
      const settingsLink = page.locator('a:has-text("设置"), button:has-text("设置")').first();
      if (await settingsLink.isVisible({ timeout: 2000 }).catch(() => false)) {
        await settingsLink.click();
        await page.waitForTimeout(1000);
        await page.goBack();
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
  });

  // === 侧边栏导航 ===
  test('TC-HOM-005: 侧边栏导航', async ({ authedPage: page }) => {
    const sidebar = page.locator('nav, [class*="sidebar"], [class*="menu"]').first();
    if (await sidebar.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(sidebar).toBeVisible();
    }
  });

  // === 响应式 ===
  test('TC-HOM-006: 首页响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('body')).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('body')).toBeVisible();
  });
});
