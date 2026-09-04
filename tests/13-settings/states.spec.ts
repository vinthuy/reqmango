import { test, expect } from '../fixtures/auth';

test.describe('状态管理', () => {
  test('TC-SET-004: 状态列表', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("状态")');
    await page.waitForTimeout(1000);
    await expect(page.locator('main').first()).toBeVisible();
  });
});
