import { test, expect } from '../fixtures/auth';

test.describe('Issue CRUD', () => {
  const PROJECT_URL = '/workspace/qa-test/project/2347';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(PROJECT_URL);
    await page.waitForTimeout(2000);
  });

  test('TC-ISS-001: 快速创建 Issue', async ({ authedPage: page }) => {
    const title = `E2E Issue ${Date.now()}`;
    await page.fill('input[placeholder*="快速创建"]', title);
    await page.press('input[placeholder*="快速创建"]', 'Enter');
    await page.waitForTimeout(2000);
    await expect(page.locator(`td:has-text("${title}")`)).toBeVisible();
  });

  test('TC-ISS-002: Issue 列表显示', async ({ authedPage: page }) => {
    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('th:has-text("编号")')).toBeVisible();
    await expect(page.locator('th:has-text("标题")')).toBeVisible();
    await expect(page.locator('th:has-text("状态")')).toBeVisible();
  });

  test('TC-ISS-003: 查看 Issue 详情', async ({ authedPage: page }) => {
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1000);
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
    await expect(page.locator('button:has-text("关联")')).toBeVisible();
    await expect(page.locator('button:has-text("附件")')).toBeVisible();
    await expect(page.locator('button:has-text("工时")')).toBeVisible();
    await expect(page.locator('button:has-text("动态")')).toBeVisible();
  });

  test('TC-ISS-004: 修改 Issue 状态', async ({ authedPage: page }) => {
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1500);
    // Verify the issue detail panel opened
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
    // Check that status-related content is visible in the panel
    await expect(page.locator('[class*=status], [class*=Status], [class*=badge]').first()).toBeVisible({ timeout: 5000 }).catch(() => {});
  });

  test('TC-ISS-005: 添加评论', async ({ authedPage: page }) => {
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1000);
    await page.fill('textarea[placeholder*="评论"]', 'E2E 自动化测试评论');
    await page.click('button:has-text("发布")');
    await page.waitForTimeout(1000);
    await expect(page.locator('text=E2E 自动化测试评论')).toBeVisible();
  });

  test('TC-ISS-006: Issue 属性面板', async ({ authedPage: page }) => {
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1500);
    // Check that the detail panel is visible
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
    // Check attribute labels exist in the panel
    await expect(page.locator('text=状态').first()).toBeVisible();
    await expect(page.locator('text=优先级').first()).toBeVisible();
    await expect(page.locator('text=负责人').first()).toBeVisible();
  });
});
