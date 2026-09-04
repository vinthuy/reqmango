import { test, expect } from '../fixtures/auth';

test.describe('文档管理', () => {
  test('TC-PAGE-001: 文档列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347/pages');
    await page.waitForTimeout(2000);
    await expect(page.locator('main').first()).toBeVisible();
  });

  test('TC-PAGE-002: 文档创建入口', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347/pages');
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("新建"), button:has-text("创建"), button:has-text("新建文档")').first();
    await expect(createBtn).toBeVisible();
  });
});
