import { test, expect } from '../fixtures/auth';

test.describe('Agent Loops', () => {
  test('TC-AGT-016: Loops 列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/loops');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
