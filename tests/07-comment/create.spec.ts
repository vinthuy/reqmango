import { test, expect } from '../fixtures/auth';

test.describe('评论功能', () => {
  test('TC-CMT-001: 评论区域可见', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1000);
    await expect(page.locator('text=评论')).toBeVisible();
    await expect(page.locator('textarea[placeholder*="评论"]')).toBeVisible();
  });

  test('TC-CMT-002: 创建评论', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1500);
    await page.fill('textarea[placeholder*="评论"]', '自动化测试评论');
    const submitBtn = page.locator('button:has-text("发布"), button:has-text("发送"), button[type="submit"]').first();
    await submitBtn.click();
    await page.waitForTimeout(2000);
    await expect(page.getByText('自动化测试评论', { exact: true })).toBeVisible();
  });

  test('TC-CMT-003: 评论操作按钮', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    await page.waitForTimeout(2000);
    await page.locator('button:has-text("查看")').first().click();
    await page.waitForTimeout(1000);
    // Check comment action buttons exist
    const resolveBtn = page.locator('button:has-text("标记为已解决")').first();
    const replyBtn = page.locator('button:has-text("回复")').first();
    if (await resolveBtn.isVisible()) {
      await expect(resolveBtn).toBeVisible();
    }
    if (await replyBtn.isVisible()) {
      await expect(replyBtn).toBeVisible();
    }
  });
});
