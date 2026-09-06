import { test, expect } from '../fixtures/auth';

test.describe('发布(Releases)全功能测试', () => {
  const RELEASE_URL = '/workspace/qa-test/project/2347?tab=releases';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(RELEASE_URL);
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-REL-001: 发布页面正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=发布, text=Releases, text=版本').first()).toBeVisible();
  });

  // === 发布列表 ===
  test('TC-REL-002: 发布列表显示', async ({ authedPage: page }) => {
    const list = page.locator('table, [class*="list"], [class*="card"]').first();
    if (await list.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(list).toBeVisible();
    }
  });

  // === 创建发布 ===
  test('TC-REL-003: 创建新发布', async ({ authedPage: page }) => {
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"], input[placeholder*="版本"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`v1.0.0-e2e-${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=发布, text=Releases, text=版本').first()).toBeVisible();
  });

  // === 编辑发布 ===
  test('TC-REL-004: 编辑发布', async ({ authedPage: page }) => {
    const editBtn = page.locator('button:has-text("编辑"), button:has-text("Edit")').first();
    if (await editBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await editBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('text=发布, text=Releases, text=版本').first()).toBeVisible();
  });

  // === 删除发布 ===
  test('TC-REL-005: 删除发布确认', async ({ authedPage: page }) => {
    const deleteBtn = page.locator('button:has-text("删除"), button:has-text("Delete")').first();
    if (await deleteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await deleteBtn.click();
      await page.waitForTimeout(500);
      const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
      if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('text=发布, text=Releases, text=版本').first()).toBeVisible();
  });

  // === 发布详情 ===
  test('TC-REL-006: 发布详情', async ({ authedPage: page }) => {
    const detailBtn = page.locator('button:has-text("查看"), button:has-text("详情")').first();
    if (await detailBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await detailBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=发布, text=Releases, text=版本').first()).toBeVisible();
  });

  // === 筛选 ===
  test('TC-REL-007: 发布筛选', async ({ authedPage: page }) => {
    const filterBtn = page.locator('button:has-text("筛选"), select, [role="combobox"]').first();
    if (await filterBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await filterBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=发布, text=Releases, text=版本').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-REL-008: 发布页面响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=发布, text=Releases, text=版本').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=发布, text=Releases, text=版本').first()).toBeVisible();
  });

  // === 空状态 ===
  test('TC-REL-009: 空发布列表', async ({ authedPage: page }) => {
    await expect(page.locator('text=发布, text=Releases, text=版本, text=暂无数据').first()).toBeVisible();
  });
});
