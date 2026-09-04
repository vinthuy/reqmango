import { test, expect } from '../fixtures/auth';

test.describe('Tester Agent', () => {
  test('TC-AGT-023: Tester Agent 页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/tester');
    await page.waitForTimeout(2000);
    await expect(page.locator('main').first()).toBeVisible();
  });
});
