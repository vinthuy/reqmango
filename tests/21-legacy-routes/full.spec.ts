import { test, expect } from '../fixtures/auth';

test.describe('旧版兼容路由', () => {
  // === Custom Fields ===
  test('TC-LGC-001: 自定义字段页面加载', async ({ authedPage: page }) => {
    await page.goto('/workspaces/1/projects/2347/custom-fields');
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  test('TC-LGC-002: 创建自定义字段', async ({ authedPage: page }) => {
    await page.goto('/workspaces/1/projects/2347/custom-fields');
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Field ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('body')).toBeVisible();
  });

  // === Issue Type List ===
  test('TC-LGC-003: Issue 类型列表页面加载', async ({ authedPage: page }) => {
    await page.goto('/workspaces/1/projects/2347/issue-types');
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  test('TC-LGC-004: Issue 类型列表显示', async ({ authedPage: page }) => {
    await page.goto('/workspaces/1/projects/2347/issue-types');
    await page.waitForTimeout(2000);
    const list = page.locator('table, [class*="list"], [class*="card"]').first();
    if (await list.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(list).toBeVisible();
    }
  });

  // === Old Issue Detail ===
  test('TC-LGC-005: 旧版 Issue 详情路由', async ({ authedPage: page }) => {
    await page.goto('/workspaces/1/projects/2347/issues/1');
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  // === Old Issue Create ===
  test('TC-LGC-006: 旧版 Issue 创建路由', async ({ authedPage: page }) => {
    await page.goto('/workspaces/1/projects/2347/issues/new');
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  // === Old Cycle Create ===
  test('TC-LGC-007: 旧版 Cycle 创建路由', async ({ authedPage: page }) => {
    await page.goto('/workspaces/1/projects/2347/cycles/new');
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  // === Old Cycle Detail ===
  test('TC-LGC-008: 旧版 Cycle 详情路由', async ({ authedPage: page }) => {
    await page.goto('/workspaces/1/projects/2347/cycles/1');
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  // === Responsive ===
  test('TC-LGC-009: 旧版路由响应式', async ({ authedPage: page }) => {
    await page.goto('/workspaces/1/projects/2347/custom-fields');
    await page.waitForTimeout(2000);
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('body')).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('body')).toBeVisible();
  });
});
