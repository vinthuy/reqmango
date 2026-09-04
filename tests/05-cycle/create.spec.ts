import { test, expect } from '../fixtures/auth';

test.describe('Cycle 创建', () => {
  test('TC-CYC-001: 周期列表页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347?tab=cycles');
    await page.waitForTimeout(2000);
    await expect(page.locator('button:has-text("创建新周期"), button:has-text("创建第一个周期")')).toBeVisible();
  });

  test('TC-CYC-002: 创建新周期向导', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347?tab=cycles');
    await page.waitForTimeout(2000);
    await page.click('button:has-text("创建新周期"), button:has-text("创建第一个周期")');
    await page.waitForTimeout(2000);
    await expect(page.locator('h1:has-text("创建新周期")')).toBeVisible();
    await expect(page.locator('text=基本信息')).toBeVisible();
    await expect(page.locator('text=选择工作项')).toBeVisible();
    await expect(page.locator('text=确认')).toBeVisible();
  });

  test('TC-CYC-003: 周期状态筛选', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347?tab=cycles');
    await page.waitForTimeout(2000);
    await expect(page.locator('select, [role=combobox]').first()).toBeVisible();
  });
});
