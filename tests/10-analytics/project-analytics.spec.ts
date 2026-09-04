import { test, expect } from '../fixtures/auth';

test.describe('项目分析', () => {
  test('TC-ANA-001: 度量页面加载', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
    await page.click('button:has-text("度量")');
    await page.waitForTimeout(2000);
    await expect(page.locator('main')).toBeVisible();
  });
});
