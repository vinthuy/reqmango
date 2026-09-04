import { test, expect } from '../fixtures/auth';

test.describe('标签管理', () => {
  test('TC-SET-005: 自定义字段页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("自定义字段")');
    await page.waitForTimeout(1000);
    await expect(page.locator('main').first()).toBeVisible();
  });
});
