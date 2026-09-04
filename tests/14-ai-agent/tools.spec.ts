import { test, expect } from '../fixtures/auth';

test.describe('Agent Tools', () => {
  test('TC-AGT-017: Tools 列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/tools');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
