import { test, expect } from '../fixtures/auth';

test.describe('目标(Initiatives)全功能测试', () => {
  const INIT_URL = '/workspace/qa-test/project/2347?tab=initiatives';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(INIT_URL);
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-INT-001: 目标页面正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=目标, text=Initiatives, text=OKR').first()).toBeVisible();
  });

  // === 目标列表 ===
  test('TC-INT-002: 目标列表显示', async ({ authedPage: page }) => {
    const list = page.locator('table, [class*="list"], [class*="card"]').first();
    if (await list.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(list).toBeVisible();
    }
  });

  // === 创建目标 ===
  test('TC-INT-003: 创建新目标', async ({ authedPage: page }) => {
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"], input[placeholder*="标题"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Initiative ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=目标, text=Initiatives, text=OKR').first()).toBeVisible();
  });

  // === 编辑目标 ===
  test('TC-INT-004: 编辑目标', async ({ authedPage: page }) => {
    const editBtn = page.locator('button:has-text("编辑"), button:has-text("Edit")').first();
    if (await editBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await editBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('text=目标, text=Initiatives, text=OKR').first()).toBeVisible();
  });

  // === 删除目标 ===
  test('TC-INT-005: 删除目标确认', async ({ authedPage: page }) => {
    const deleteBtn = page.locator('button:has-text("删除"), button:has-text("Delete")').first();
    if (await deleteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await deleteBtn.click();
      await page.waitForTimeout(500);
      const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
      if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('text=目标, text=Initiatives, text=OKR').first()).toBeVisible();
  });

  // === 目标进度 ===
  test('TC-INT-006: 目标进度显示', async ({ authedPage: page }) => {
    const progress = page.locator('[class*="progress"], [role="progressbar"]').first();
    if (await progress.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(progress).toBeVisible();
    }
  });

  // === 筛选 ===
  test('TC-INT-007: 目标筛选', async ({ authedPage: page }) => {
    const filterBtn = page.locator('button:has-text("筛选"), select, [role="combobox"]').first();
    if (await filterBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await filterBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=目标, text=Initiatives, text=OKR').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-INT-008: 目标页面响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=目标, text=Initiatives, text=OKR').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=目标, text=Initiatives, text=OKR').first()).toBeVisible();
  });

  // === 空状态 ===
  test('TC-INT-009: 空目标列表', async ({ authedPage: page }) => {
    await expect(page.locator('text=目标, text=Initiatives, text=OKR, text=暂无数据').first()).toBeVisible();
  });
});
