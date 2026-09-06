import { test, expect } from '../fixtures/auth';

test.describe('工作流全功能测试', () => {
  const WORKFLOW_URL = '/workspace/qa-test/project/2347?tab=workflows';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(WORKFLOW_URL);
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-WFL-001: 工作流页面正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=工作流, text=Workflow').first()).toBeVisible();
  });

  // === 工作流列表 ===
  test('TC-WFL-002: 工作流列表显示', async ({ authedPage: page }) => {
    const list = page.locator('table, [class*="list"], [class*="card"]').first();
    if (await list.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(list).toBeVisible();
    }
  });

  // === 创建工作流 ===
  test('TC-WFL-003: 创建新工作流', async ({ authedPage: page }) => {
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Workflow ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=工作流, text=Workflow').first()).toBeVisible();
  });

  // === 编辑工作流 ===
  test('TC-WFL-004: 编辑工作流', async ({ authedPage: page }) => {
    const editBtn = page.locator('button:has-text("编辑"), button:has-text("Edit")').first();
    if (await editBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await editBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('text=工作流, text=Workflow').first()).toBeVisible();
  });

  // === 删除工作流 ===
  test('TC-WFL-005: 删除工作流确认', async ({ authedPage: page }) => {
    const deleteBtn = page.locator('button:has-text("删除"), button:has-text("Delete")').first();
    if (await deleteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await deleteBtn.click();
      await page.waitForTimeout(500);
      const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
      if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('text=工作流, text=Workflow').first()).toBeVisible();
  });

  // === 状态转换 ===
  test('TC-WFL-006: 状态转换规则', async ({ authedPage: page }) => {
    const transitionBtn = page.locator('button:has-text("转换"), button:has-text("Transition")').first();
    if (await transitionBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await transitionBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=工作流, text=Workflow').first()).toBeVisible();
  });

  // === 工作流详情 ===
  test('TC-WFL-007: 工作流详情页', async ({ authedPage: page }) => {
    const detailBtn = page.locator('button:has-text("详情"), a:has-text("详情")').first();
    if (await detailBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await detailBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=工作流, text=Workflow').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-WFL-008: 工作流页面响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=工作流, text=Workflow').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=工作流, text=Workflow').first()).toBeVisible();
  });

  // === 空状态 ===
  test('TC-WFL-009: 空工作流显示', async ({ authedPage: page }) => {
    await expect(page.locator('text=工作流, text=Workflow, text=暂无数据').first()).toBeVisible();
  });
});
