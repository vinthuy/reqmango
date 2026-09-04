import { test, expect } from '../fixtures/auth';

test.describe('Agent Autopilot', () => {
  test('TC-AGT-021: Autopilot 列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/autopilot');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
