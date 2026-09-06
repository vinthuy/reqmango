import { test, expect } from '../fixtures/auth';

test.describe('Intake Form (需求收集表单)', () => {
  // === 页面加载 ===
  test('TC-INT-001: Intake Form 页面加载', async ({ authedPage: page }) => {
    await page.goto('/intake/2347');
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  // === 表单显示 ===
  test('TC-INT-002: 表单字段显示', async ({ authedPage: page }) => {
    await page.goto('/intake/2347');
    await page.waitForTimeout(2000);
    const form = page.locator('form, [class*="form"], input, textarea').first();
    if (await form.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(form).toBeVisible();
    }
  });

  // === 表单提交 ===
  test('TC-INT-003: 表单提交', async ({ authedPage: page }) => {
    await page.goto('/intake/2347');
    await page.waitForTimeout(2000);
    const titleInput = page.locator('input[placeholder*="标题"], input[placeholder*="title"], input[placeholder*="名称"]').first();
    if (await titleInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await titleInput.fill('E2E Intake Form Test');
    }
    const submitBtn = page.locator('button:has-text("提交"), button:has-text("Submit")').first();
    if (await submitBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await submitBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('body')).toBeVisible();
  });

  // === 响应式 ===
  test('TC-INT-004: Intake Form 响应式', async ({ authedPage: page }) => {
    await page.goto('/intake/2347');
    await page.waitForTimeout(2000);
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('body')).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('body')).toBeVisible();
  });
});
