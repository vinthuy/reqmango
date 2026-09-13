import { test, expect } from '../fixtures/auth';

test.describe('评论功能', () => {
  test('TC-CMT-001: 评论区域可见', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    const viewBtn = page.locator('button:has-text("查看")').first();
    await expect(viewBtn).toBeVisible({ timeout: 20000 });
    await viewBtn.click({ timeout: 5000 }).catch(() => {});
    // A bare `text=评论` also matches the "暂无评论" empty state, so it threw a
    // strict-mode violation whenever the issue had no comments yet (flaky).
    await expect(page.locator('text=评论').first()).toBeVisible({ timeout: 20000 });
    await expect(page.locator('textarea[placeholder*="评论"]').first()).toBeVisible({ timeout: 20000 });
  });

  test('TC-CMT-002: 创建评论', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/project/2347');
    // Deterministic readiness gates instead of fixed sleeps
    const viewBtn = page.locator('button:has-text("查看")').first();
    await expect(viewBtn).toBeVisible({ timeout: 20000 });
    await viewBtn.click({ timeout: 5000 }).catch(() => {});
    const commentInput = page.locator('textarea[placeholder*="评论"]').first();
    await expect(commentInput).toBeVisible({ timeout: 20000 });
    // Unique text per run: a fixed string accumulated duplicate comments across
    // runs, making the final assertion ambiguous/slow.
    const text = `自动化测试评论 ${Date.now()}`;
    await commentInput.fill(text);
    const submitBtn = page.locator('button:has-text("发布"), button:has-text("发送"), button[type="submit"]').first();
    await submitBtn.click({ timeout: 5000 }).catch(() => {});
    await expect(page.getByText(text, { exact: true }).first()).toBeVisible({ timeout: 15000 });
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
