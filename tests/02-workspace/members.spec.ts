import { test, expect } from '../fixtures/auth';

test.describe('工作空间成员', () => {
  test('TC-WS-004: 查看成员列表', async ({ authedPage: page }) => {
    // 先进入一个工作空间
    const wsCard = page.locator('h3').first();
    if (await wsCard.isVisible()) {
      await wsCard.click();
      await page.waitForTimeout(2000);
    }
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("成员")');
    await page.waitForTimeout(1000);
    await expect(page.getByText('qa_tester', { exact: true })).toBeVisible();
  });
});
