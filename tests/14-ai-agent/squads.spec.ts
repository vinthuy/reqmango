import { test, expect } from '../fixtures/auth';

test.describe('Agent Squads', () => {
  test('TC-AGT-026: Squads 列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/squads');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
