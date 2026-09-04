import { test, expect } from '../fixtures/auth';

test.describe('Developer Agent', () => {
  test('TC-AGT-022: Developer Agent 页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/developer');
    await page.waitForTimeout(2000);
    await expect(page.locator('main').first()).toBeVisible();
  });
});
