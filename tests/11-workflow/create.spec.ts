import { test, expect } from '../fixtures/auth';

test.describe('工作流管理', () => {
  test('TC-WF-001: 工作流列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347/workflows');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
