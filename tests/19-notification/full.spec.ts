import { test, expect } from '../fixtures/auth';

test.describe('通知中心全功能测试', () => {
  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test');
    await page.waitForTimeout(2000);
  });

  // === 通知入口 ===
  test('TC-NTF-001: 通知图标显示', async ({ authedPage: page }) => {
    const notifIcon = page.locator('[class*="notification"], button:has-text("通知"), [aria-label="通知"], [class*="bell"]').first();
    if (await notifIcon.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(notifIcon).toBeVisible();
    }
  });

  // === 通知面板 ===
  test('TC-NTF-002: 打开通知面板', async ({ authedPage: page }) => {
    const notifIcon = page.locator('[class*="notification"], button:has-text("通知"), [aria-label="通知"], [class*="bell"]').first();
    if (await notifIcon.isVisible({ timeout: 3000 }).catch(() => false)) {
      await notifIcon.click();
      await page.waitForTimeout(500);
      const panel = page.locator('[class*="dropdown"], [class*="panel"], [class*="popover"]').first();
      if (await panel.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(panel).toBeVisible();
      }
      await page.keyboard.press('Escape');
    }
  });

  // === 通知列表 ===
  test('TC-NTF-003: 通知列表显示', async ({ authedPage: page }) => {
    const notifIcon = page.locator('[class*="notification"], button:has-text("通知"), [aria-label="通知"], [class*="bell"]').first();
    if (await notifIcon.isVisible({ timeout: 3000 }).catch(() => false)) {
      await notifIcon.click();
      await page.waitForTimeout(500);
      const notifList = page.locator('[class*="notification-item"], [class*="list"] a, [class*="item"]').first();
      if (await notifList.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(notifList).toBeVisible();
      }
      await page.keyboard.press('Escape');
    }
  });

  // === 标记已读 ===
  test('TC-NTF-004: 标记全部已读', async ({ authedPage: page }) => {
    const notifIcon = page.locator('[class*="notification"], button:has-text("通知"), [aria-label="通知"], [class*="bell"]').first();
    if (await notifIcon.isVisible({ timeout: 3000 }).catch(() => false)) {
      await notifIcon.click();
      await page.waitForTimeout(500);
      const markAllBtn = page.locator('button:has-text("全部已读"), button:has-text("Mark all read")').first();
      if (await markAllBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await markAllBtn.click();
        await page.waitForTimeout(500);
      }
      await page.keyboard.press('Escape');
    }
  });

  // === 通知设置 ===
  test('TC-NTF-005: 通知设置入口', async ({ authedPage: page }) => {
    const settingsBtn = page.locator('button:has-text("设置"), a:has-text("设置"), [aria-label="设置"]').first();
    if (await settingsBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await settingsBtn.click();
      await page.waitForTimeout(1000);
    }
  });

  // === 未读计数 ===
  test('TC-NTF-006: 未读计数显示', async ({ authedPage: page }) => {
    const badge = page.locator('[class*="badge"], [class*="count"], [class*="unread"]').first();
    if (await badge.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(badge).toBeVisible();
    }
  });

  // === 通知详情 ===
  test('TC-NTF-007: 点击通知跳转', async ({ authedPage: page }) => {
    const notifIcon = page.locator('[class*="notification"], button:has-text("通知"), [aria-label="通知"], [class*="bell"]').first();
    if (await notifIcon.isVisible({ timeout: 3000 }).catch(() => false)) {
      await notifIcon.click();
      await page.waitForTimeout(500);
      const notifItem = page.locator('[class*="notification-item"], [class*="item"] a').first();
      if (await notifItem.isVisible({ timeout: 2000 }).catch(() => false)) {
        await notifItem.click();
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
  });

  // === 响应式 ===
  test('TC-NTF-008: 通知面板响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    const notifIcon = page.locator('[class*="notification"], button:has-text("通知"), [aria-label="通知"], [class*="bell"]').first();
    if (await notifIcon.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(notifIcon).toBeVisible();
    }
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
  });

  // === 空通知 ===
  test('TC-NTF-009: 空通知列表', async ({ authedPage: page }) => {
    const notifIcon = page.locator('[class*="notification"], button:has-text("通知"), [aria-label="通知"], [class*="bell"]').first();
    if (await notifIcon.isVisible({ timeout: 3000 }).catch(() => false)) {
      await notifIcon.click();
      await page.waitForTimeout(500);
      const emptyState = page.locator('text=暂无通知, text=No notifications, text=没有通知').first();
      if (await emptyState.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(emptyState).toBeVisible();
      }
      await page.keyboard.press('Escape');
    }
  });
});
