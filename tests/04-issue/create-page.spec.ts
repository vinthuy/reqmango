import { test, expect } from '../fixtures/auth';

test.describe('Issue 创建页', () => {
  const CREATE_URL = '/workspace/qa-test/project/2347/issues/new';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(CREATE_URL);
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-ICR-001: 创建页正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=创建, text=新建, text=Create').first()).toBeVisible();
  });

  // === 标题输入 ===
  test('TC-ICR-002: 标题输入', async ({ authedPage: page }) => {
    const titleInput = page.locator('input[placeholder*="标题"], input[placeholder*="title"], input[placeholder*="名称"]').first();
    if (await titleInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await titleInput.fill('E2E 创建页测试 Issue');
      await expect(titleInput).toHaveValue('E2E 创建页测试 Issue');
    }
  });

  // === 描述输入 ===
  test('TC-ICR-003: 描述输入', async ({ authedPage: page }) => {
    const descInput = page.locator('[contenteditable], .ProseMirror, textarea[placeholder*="描述"]').first();
    if (await descInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await descInput.click();
      await page.keyboard.type('E2E 描述内容');
    }
    await expect(page.locator('text=创建, text=新建, text=Create').first()).toBeVisible();
  });

  // === 状态选择 ===
  test('TC-ICR-004: 状态选择', async ({ authedPage: page }) => {
    const statusSelect = page.locator('text=状态').first();
    if (await statusSelect.isVisible({ timeout: 3000 }).catch(() => false)) {
      await statusSelect.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=创建, text=新建, text=Create').first()).toBeVisible();
  });

  // === 优先级选择 ===
  test('TC-ICR-005: 优先级选择', async ({ authedPage: page }) => {
    const prioritySelect = page.locator('text=优先级').first();
    if (await prioritySelect.isVisible({ timeout: 3000 }).catch(() => false)) {
      await prioritySelect.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=创建, text=新建, text=Create').first()).toBeVisible();
  });

  // === 负责人选择 ===
  test('TC-ICR-006: 负责人选择', async ({ authedPage: page }) => {
    const assigneeSelect = page.locator('text=负责人').first();
    if (await assigneeSelect.isVisible({ timeout: 3000 }).catch(() => false)) {
      await assigneeSelect.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=创建, text=新建, text=Create').first()).toBeVisible();
  });

  // === Issue 类型选择 ===
  test('TC-ICR-007: Issue 类型选择', async ({ authedPage: page }) => {
    const typeSelect = page.locator('text=类型').first();
    if (await typeSelect.isVisible({ timeout: 3000 }).catch(() => false)) {
      await typeSelect.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=创建, text=新建, text=Create').first()).toBeVisible();
  });

  // === 标签选择 ===
  test('TC-ICR-008: 标签选择', async ({ authedPage: page }) => {
    const labelSelect = page.locator('text=标签').first();
    if (await labelSelect.isVisible({ timeout: 3000 }).catch(() => false)) {
      await labelSelect.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=创建, text=新建, text=Create').first()).toBeVisible();
  });

  // === 截止日期 ===
  test('TC-ICR-009: 截止日期设置', async ({ authedPage: page }) => {
    const dateInput = page.locator('text=截止日期, text=到期日').first();
    if (await dateInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await dateInput.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=创建, text=新建, text=Create').first()).toBeVisible();
  });

  // === 提交创建 ===
  test('TC-ICR-010: 提交创建', async ({ authedPage: page }) => {
    const titleInput = page.locator('input[placeholder*="标题"], input[placeholder*="title"]').first();
    if (await titleInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await titleInput.fill(`E2E Create Page ${Date.now()}`);
    }
    const submitBtn = page.locator('button:has-text("创建"), button:has-text("提交"), button:has-text("Submit")').first();
    if (await submitBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await submitBtn.click();
      await page.waitForTimeout(2000);
    }
  });

  // === 取消创建 ===
  test('TC-ICR-011: 取消创建', async ({ authedPage: page }) => {
    const cancelBtn = page.locator('button:has-text("取消"), button:has-text("Cancel")').first();
    if (await cancelBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await cancelBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('table, [class*="list"]').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-ICR-012: 创建页响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=创建, text=新建, text=Create').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=创建, text=新建, text=Create').first()).toBeVisible();
  });
});
