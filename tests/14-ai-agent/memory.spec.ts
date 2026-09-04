import { test, expect } from '../fixtures/auth';

test.describe('Agent Memory', () => {
  test('TC-AGT-020: Memory 列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/memories');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
