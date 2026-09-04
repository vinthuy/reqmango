import { test, expect } from '../fixtures/auth';

test.describe('Agent Skills', () => {
  test('TC-AGT-013: Skills 列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/skills');
    await page.waitForTimeout(2000);
    await expect(page.locator('main').first()).toBeVisible();
  });

  test('TC-AGT-014: Skills CRUD 入口', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/skills');
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建")').first();
    await expect(createBtn).toBeVisible();
  });
});
