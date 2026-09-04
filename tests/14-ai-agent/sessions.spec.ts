import { test, expect } from '../fixtures/auth';

test.describe('Agent Sessions', () => {
  test('TC-AGT-019: Sessions 列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/sessions');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
