import { test, expect } from '../fixtures/auth';

test.describe('模板管理', () => {
  test('TC-SET-006: 模板列表', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("模板")');
    await page.waitForTimeout(1000);
    await expect(page.locator('main').first()).toBeVisible();
  });
});
