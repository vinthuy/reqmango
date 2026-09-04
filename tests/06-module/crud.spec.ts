import { test, expect } from '../fixtures/auth';

test.describe('模块管理', () => {
  test('TC-MOD-001: 模块列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347?tab=modules');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });

  test('TC-MOD-002: 模块创建入口', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347?tab=modules');
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建")').first();
    await expect(createBtn).toBeVisible();
  });
});
