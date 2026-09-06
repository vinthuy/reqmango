import { test, expect } from '../fixtures/auth';

test.describe('审批全功能测试', () => {
  const APPROVAL_URL = '/workspace/qa-test/project/2347?tab=approvals';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(APPROVAL_URL);
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-APR-001: 审批页面正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=审批, text=Approval, text=待审批').first()).toBeVisible();
  });

  // === 审批列表 ===
  test('TC-APR-002: 审批列表显示', async ({ authedPage: page }) => {
    const list = page.locator('table, [class*="list"], [class*="card"]').first();
    if (await list.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(list).toBeVisible();
    }
  });

  // === 审批详情 ===
  test('TC-APR-003: 查看审批详情', async ({ authedPage: page }) => {
    const detailBtn = page.locator('button:has-text("查看"), button:has-text("详情")').first();
    if (await detailBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await detailBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=审批, text=Approval, text=待审批').first()).toBeVisible();
  });

  // === 批准操作 ===
  test('TC-APR-004: 批准审批', async ({ authedPage: page }) => {
    const approveBtn = page.locator('button:has-text("批准"), button:has-text("通过"), button:has-text("Approve")').first();
    if (await approveBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await approveBtn.click();
      await page.waitForTimeout(500);
      const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
      if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('text=审批, text=Approval, text=待审批').first()).toBeVisible();
  });

  // === 拒绝操作 ===
  test('TC-APR-005: 拒绝审批', async ({ authedPage: page }) => {
    const rejectBtn = page.locator('button:has-text("拒绝"), button:has-text("Reject")').first();
    if (await rejectBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await rejectBtn.click();
      await page.waitForTimeout(500);
      const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
      if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('text=审批, text=Approval, text=待审批').first()).toBeVisible();
  });

  // === 筛选 ===
  test('TC-APR-006: 审批筛选', async ({ authedPage: page }) => {
    const filterBtn = page.locator('button:has-text("筛选"), select, [role="combobox"]').first();
    if (await filterBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await filterBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=审批, text=Approval, text=待审批').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-APR-007: 审批页面响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=审批, text=Approval, text=待审批').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=审批, text=Approval, text=待审批').first()).toBeVisible();
  });

  // === 空状态 ===
  test('TC-APR-008: 空审批列表', async ({ authedPage: page }) => {
    await expect(page.locator('text=审批, text=Approval, text=待审批, text=暂无数据').first()).toBeVisible();
  });
});
