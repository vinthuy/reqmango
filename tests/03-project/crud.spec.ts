import { test, expect } from '../fixtures/auth';
import { randomProject } from '../helpers/test-data';

test.describe('项目管理', () => {
  test('TC-PRJ-001: 项目列表显示', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=QA验收测试项目')).toBeVisible();
  });

  test('TC-PRJ-002: 进入项目页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test');
    await page.click('h3:has-text("QA验收测试项目")');
    await page.waitForTimeout(2000);
    await expect(page.locator('h1')).toContainText('QA验收测试项目');
    await expect(page.locator('button:has-text("工作项")')).toBeVisible();
  });

  test('TC-PRJ-003: 项目标签页导航', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
    // Check all tabs exist
    for (const tab of ['工作项', '周期', '模块', '更新', '度量']) {
      await expect(page.locator(`button:has-text("${tab}")`).first()).toBeVisible();
    }
  });
});
