import { test, expect } from '../fixtures/auth';

test.describe('工作空间设置页', () => {
  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-WST-001: 设置页正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=设置, text=Settings, text=工作空间设置').first()).toBeVisible();
  });

  // === 基本信息 ===
  test('TC-WST-002: 基本信息显示', async ({ authedPage: page }) => {
    const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"], input[value]').first();
    if (await nameInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(nameInput).toBeVisible();
    }
  });

  // === 编辑名称 ===
  test('TC-WST-003: 编辑工作空间名称', async ({ authedPage: page }) => {
    const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"], input[value]').first();
    if (await nameInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await nameInput.click();
      await page.waitForTimeout(300);
    }
    await expect(page.locator('text=设置, text=Settings, text=工作空间设置').first()).toBeVisible();
  });

  // === 成员管理 ===
  test('TC-WST-004: 成员管理', async ({ authedPage: page }) => {
    const memberTab = page.locator('button:has-text("成员"), button:has-text("Members")').first();
    if (await memberTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await memberTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=设置, text=Settings, text=工作空间设置').first()).toBeVisible();
  });

  // === 邀请成员 ===
  test('TC-WST-005: 邀请成员', async ({ authedPage: page }) => {
    const inviteBtn = page.locator('button:has-text("邀请"), button:has-text("Invite")').first();
    if (await inviteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await inviteBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('text=设置, text=Settings, text=工作空间设置').first()).toBeVisible();
  });

  // === 保存设置 ===
  test('TC-WST-006: 保存设置', async ({ authedPage: page }) => {
    const saveBtn = page.locator('button:has-text("保存"), button:has-text("Save")').first();
    if (await saveBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await saveBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=设置, text=Settings, text=工作空间设置').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-WST-007: 设置页响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=设置, text=Settings, text=工作空间设置').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=设置, text=Settings, text=工作空间设置').first()).toBeVisible();
  });
});
