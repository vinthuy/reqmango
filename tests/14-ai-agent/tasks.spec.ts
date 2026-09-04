import { test, expect } from '../fixtures/auth';

test.describe('Agent Tasks', () => {
  test('TC-AGT-015: Tasks 列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/tasks');
    await page.waitForTimeout(2000);
    await expect(page.locator('main').first()).toBeVisible();
  });
});
