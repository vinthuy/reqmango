import { test, expect } from '../fixtures/auth';

test.describe('Agent Templates', () => {
  test('TC-AGT-011: 模板列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/templates');
    await page.waitForTimeout(2000);
    await expect(page.locator('main').first()).toBeVisible();
  });

  test('TC-AGT-012: 模板 CRUD 入口', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/templates');
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("初始化")').first();
    await expect(createBtn).toBeVisible();
  });
});
