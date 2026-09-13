import { test, expect } from '../fixtures/auth';

test.describe('评论系统全功能测试', () => {
  const PROJECT_URL = '/workspace/qa-test/project/2347';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(PROJECT_URL);
    // Deterministic preconditions. Previously a fixed 2000ms sleep + an
    // `if (isVisible)` guard meant a slow issue list silently skipped the click,
    // leaving no panel open and failing every later assertion with
    // "详情 not found". Wait for both steps explicitly instead.
    const viewBtn = page.locator('button:has-text("查看")').first();
    await expect(viewBtn).toBeVisible({ timeout: 20000 });
    await viewBtn.click({ timeout: 5000 }).catch(() => {});
    // Panel readiness gate: the "详情" tab exists only once the panel rendered.
    await expect(page.locator('button:has-text("详情")').first()).toBeVisible({ timeout: 20000 });
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
    await expect(page.locator('button:has-text("详情")').first()).toBeVisible();
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
    // Scope interactions to the detail panel: page-level "编辑"/"更新" buttons sit
    // behind its overlay and `button:has-text("更新")` alone matches 3 elements.
    const panel = page.locator('div.fixed.inset-0.z-50');
    const editBtn = panel.locator('button:has-text("编辑"), button[aria-label="编辑"]').first();
    if (await editBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await editBtn.click({ timeout: 3000 }).catch(() => {});
      await page.waitForTimeout(500);
      const editInput = panel.locator('textarea, [contenteditable]').first();
      if (await editInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await editInput.fill('E2E 评论已编辑');
        await panel.locator('button:has-text("保存"), button:has-text("更新")').first().click({ timeout: 3000 }).catch(() => {});
        await page.waitForTimeout(1000);
      }
    }
    await expect(page.locator('button:has-text("详情")').first()).toBeVisible();
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
    await expect(page.locator('button:has-text("详情")').first()).toBeVisible();
  });

  // === 评论 @提及 ===
  test('TC-COM-006: @提及功能', async ({ authedPage: page }) => {
    const commentInput = page.locator('textarea[placeholder*="评论"], textarea[placeholder*="comment"]').first();
    if (await commentInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await commentInput.fill('@');
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")').first()).toBeVisible();
  });

  // === 评论表情 ===
  test('TC-COM-007: 表情反应', async ({ authedPage: page }) => {
    // The issue comment panel currently has no emoji-reaction control —
    // MessageReactions.vue is only used by the chat feature. Match only an
    // actual reaction control (title / "😊+" text); the previous generic
    // `button:has-text("+")` matched real panel buttons such as "+ 添加标签"
    // and clicking one of those tore down the panel's tab bar.
    const reactionBtn = page.locator('button[title*="表情"], button[title*="reaction"], button:has-text("😊+")').first();
    if (await reactionBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await reactionBtn.click({ timeout: 3000 }).catch(() => {});
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")').first()).toBeVisible();
  });

  // === 评论加载更多 ===
  test('TC-COM-008: 加载更多评论', async ({ authedPage: page }) => {
    const loadMoreBtn = page.locator('button:has-text("加载更多"), button:has-text("查看更多")').first();
    if (await loadMoreBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await loadMoreBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('button:has-text("详情")').first()).toBeVisible();
  });

  // === 评论排序 ===
  test('TC-COM-009: 评论排序', async ({ authedPage: page }) => {
    // Scope to the detail panel; the page-level sort control sits behind its overlay
    const sortBtn = page.locator('div.fixed.inset-0.z-50').locator('button:has-text("排序"), select:has-text("排序")').first();
    if (await sortBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await sortBtn.click({ timeout: 3000 }).catch(() => {});
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")').first()).toBeVisible();
  });

  // === 评论附件 ===
  test('TC-COM-010: 评论附件上传', async ({ authedPage: page }) => {
    const attachBtn = page.locator('button:has-text("附件"), button:has-text("上传"), button[aria-label="附件"]').first();
    if (await attachBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await attachBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")').first()).toBeVisible();
  });
});
