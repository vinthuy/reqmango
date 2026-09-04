import { test, expect } from '../fixtures/auth';

test.describe('角色与权限', () => {
  test('TC-SET-003: 角色列表', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("角色与权限")');
    await page.waitForTimeout(1000);
    await expect(page.locator('main').first()).toBeVisible();
  });
});
