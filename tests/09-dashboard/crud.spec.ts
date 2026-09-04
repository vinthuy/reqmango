import { test, expect } from '../fixtures/auth';

test.describe('仪表盘', () => {
  test('TC-DB-001: 仪表盘页面加载', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347/dashboards');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
