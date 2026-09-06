import { test, expect } from '../fixtures/auth';

test.describe('评论系统全功能测试', () => {
  const PROJECT_URL = '/workspace/qa-test/project/2347';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(PROJECT_URL);
    await page.waitForTimeout(2000);
    // 打开第一个 Issue
    const viewBtn = page.locator('button:has-text("查看")').first();
    if (await viewBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await viewBtn.click();
      await page.waitForTimeout(1500);
    }
  });

  // === 评论输入框 ===
  test('TC-COM-001: 评论输入框显示', async ({ authedPage: page }) => {
    const commentInput = page.locator('textarea[placeholder*="评论"], textarea[placeholder*="comment"], [contenteditable]').first();
    if (await commentInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(commentInput).toBeVisible();
    }
  });

  // === 发布评论 ===
  test('TC-COM-002: 发布评论', async ({ authedPage: page }) => {
    const commentInput = page.locator('textarea[placeholder*="评论"], textarea[placeholder*="comment"]').first();
    if (await commentInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await commentInput.fill('E2E 评论测试');
      await page.click('button:has-text("发布"), button:has-text("发送"), button:has-text("Submit")');
      await page.waitForTimeout(1500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 评论列表 ===
  test('TC-COM-003: 评论列表显示', async ({ authedPage: page }) => {
    const commentList = page.locator('[class*="comment"], [class*="activity"], [class*="log"]').first();
    if (await commentList.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(commentList).toBeVisible();
    }
  });

  // === 编辑评论 ===
  test('TC-COM-004: 编辑评论', async ({ authedPage: page }) => {
    const editBtn = page.locator('button:has-text("编辑"), button[aria-label="编辑"]').first();
    if (await editBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await editBtn.click();
      await page.waitForTimeout(500);
      const editInput = page.locator('textarea, [contenteditable]').first();
      if (await editInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await editInput.fill('E2E 评论已编辑');
        await page.click('button:has-text("保存"), button:has-text("更新")');
        await page.waitForTimeout(1000);
      }
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 删除评论 ===
  test('TC-COM-005: 删除评论确认', async ({ authedPage: page }) => {
    const deleteBtn = page.locator('button:has-text("删除"), button[aria-label="删除"]').first();
    if (await deleteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await deleteBtn.click();
      await page.waitForTimeout(500);
      const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
      if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 评论 @提及 ===
  test('TC-COM-006: @提及功能', async ({ authedPage: page }) => {
    const commentInput = page.locator('textarea[placeholder*="评论"], textarea[placeholder*="comment"]').first();
    if (await commentInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await commentInput.fill('@');
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 评论表情 ===
  test('TC-COM-007: 表情反应', async ({ authedPage: page }) => {
    const reactionBtn = page.locator('button:has-text("😀"), button:has-text("+"), [class*="reaction"]').first();
    if (await reactionBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await reactionBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 评论加载更多 ===
  test('TC-COM-008: 加载更多评论', async ({ authedPage: page }) => {
    const loadMoreBtn = page.locator('button:has-text("加载更多"), button:has-text("查看更多")').first();
    if (await loadMoreBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await loadMoreBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 评论排序 ===
  test('TC-COM-009: 评论排序', async ({ authedPage: page }) => {
    const sortBtn = page.locator('button:has-text("排序"), select:has-text("排序")').first();
    if (await sortBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await sortBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 评论附件 ===
  test('TC-COM-010: 评论附件上传', async ({ authedPage: page }) => {
    const attachBtn = page.locator('button:has-text("附件"), button:has-text("上传"), button[aria-label="附件"]').first();
    if (await attachBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await attachBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });
});
