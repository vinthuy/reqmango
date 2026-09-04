import { test, expect } from '../fixtures/auth';

test.describe('Agent Runtimes', () => {
  test('TC-AGT-018: Runtimes 列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/runtimes');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
