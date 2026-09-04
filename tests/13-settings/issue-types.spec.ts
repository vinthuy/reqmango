import { test, expect } from '../fixtures/auth';

test.describe('Issue 类型管理', () => {
  test('TC-SET-001: Issue 类型列表', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("工作项类型")');
    await page.waitForTimeout(1000);
    await expect(page.locator('h2:has-text("工作项类型")')).toBeVisible();
  });

  test('TC-SET-002: 新建 Issue 类型', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/settings');
    await page.click('button:has-text("工作项类型")');
    await page.waitForTimeout(1000);
    await page.click('button:has-text("新建类型")');
    await page.waitForTimeout(1000);
    await expect(page.locator('input, textarea').first()).toBeVisible();
  });
});
