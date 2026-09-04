import { test, expect } from '../fixtures/auth';

test.describe('审批流程', () => {
  test('TC-APR-001: 审批列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/approvals');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
