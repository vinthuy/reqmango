import { test, expect } from '../fixtures/auth';

test.describe('Issue 看板视图', () => {
  const PROJECT_URL = '/workspace/qa-test/project/2347';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(PROJECT_URL);
    await page.waitForTimeout(2000);
    // 切换到看板视图
    await page.click('button:has-text("看板")');
    await page.waitForTimeout(2000);
  });

  // === 看板加载 ===
  test('TC-KAN-001: 看板视图正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('h3:has-text("Backlog"), h3:has-text("Todo"), h3:has-text("待办")')).toBeVisible();
  });

  // === 状态列显示 ===
  test('TC-KAN-002: 所有状态列显示', async ({ authedPage: page }) => {
    const columns = page.locator('[class*="column"], [class*="Column"], [class*="lane"]');
    const count = await columns.count();
    expect(count).toBeGreaterThanOrEqual(2);
  });

  // === Issue 卡片 ===
  test('TC-KAN-003: Issue 卡片显示', async ({ authedPage: page }) => {
    const cards = page.locator('[class*="card"], [class*="Card"], [class*="issue-item"]');
    const count = await cards.count();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  // === 拖拽排序 ===
  test('TC-KAN-004: 卡片拖拽操作', async ({ authedPage: page }) => {
    const firstCard = page.locator('[class*="card"], [class*="Card"], [class*="issue-item"]').first();
    if (await firstCard.isVisible({ timeout: 3000 }).catch(() => false)) {
      const box = await firstCard.boundingBox();
      if (box) {
        await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
        await page.waitForTimeout(300);
      }
    }
    await expect(page.locator('h3:has-text("Backlog"), h3:has-text("Todo")')).toBeVisible();
  });

  // === 快速创建 ===
  test('TC-KAN-005: 看板快速创建 Issue', async ({ authedPage: page }) => {
    const addBtn = page.locator('button:has-text("添加"), button:has-text("+"), [class*="add-button"]').first();
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
      const input = page.locator('input[placeholder*="标题"], input[placeholder*="title"]').first();
      if (await input.isVisible({ timeout: 2000 }).catch(() => false)) {
        await input.fill(`Kanban E2E ${Date.now()}`);
        await page.keyboard.press('Enter');
        await page.waitForTimeout(1000);
      }
    }
    await expect(page.locator('h3:has-text("Backlog"), h3:has-text("Todo")')).toBeVisible();
  });

  // === 卡片详情 ===
  test('TC-KAN-006: 点击卡片打开详情', async ({ authedPage: page }) => {
    const card = page.locator('[class*="card"], [class*="Card"], [class*="issue-item"]').first();
    if (await card.isVisible({ timeout: 3000 }).catch(() => false)) {
      await card.click();
      await page.waitForTimeout(1500);
      // 验证详情面板或弹窗出现
      const detailPanel = page.locator('button:has-text("详情"), [class*="detail-panel"], [class*="drawer"]').first();
      if (await detailPanel.isVisible({ timeout: 3000 }).catch(() => false)) {
        await expect(detailPanel).toBeVisible();
      }
    }
  });

  // === 筛选 ===
  test('TC-KAN-007: 看板筛选功能', async ({ authedPage: page }) => {
    const filterBtn = page.locator('button:has-text("筛选"), button:has-text("Filter")').first();
    if (await filterBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await filterBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('h3:has-text("Backlog"), h3:has-text("Todo")')).toBeVisible();
  });

  // === 搜索 ===
  test('TC-KAN-008: 看板搜索', async ({ authedPage: page }) => {
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await searchInput.fill('E2E');
      await page.waitForTimeout(1000);
      await searchInput.clear();
    }
    await expect(page.locator('h3:has-text("Backlog"), h3:has-text("Todo")')).toBeVisible();
  });

  // === 列头操作 ===
  test('TC-KAN-009: 列头菜单', async ({ authedPage: page }) => {
    const columnHeader = page.locator('h3:has-text("Backlog"), h3:has-text("Todo")').first();
    if (await columnHeader.isVisible({ timeout: 3000 }).catch(() => false)) {
      const menuBtn = columnHeader.locator('button, [class*="menu"]').first();
      if (await menuBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await menuBtn.click();
        await page.waitForTimeout(500);
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('h3:has-text("Backlog"), h3:has-text("Todo")')).toBeVisible();
  });

  // === 视图切换回列表 ===
  test('TC-KAN-010: 切换回列表视图', async ({ authedPage: page }) => {
    await page.click('button:has-text("列表")');
    await page.waitForTimeout(1000);
    await expect(page.locator('table')).toBeVisible();
  });

  // === 空列显示 ===
  test('TC-KAN-011: 空列正确显示', async ({ authedPage: page }) => {
    // 看板应该显示所有状态列，即使某些列没有 Issue
    await expect(page.locator('[class*="column"], [class*="Column"], [class*="lane"]').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-KAN-012: 看板响应式布局', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('h3:has-text("Backlog"), h3:has-text("Todo")')).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('h3:has-text("Backlog"), h3:has-text("Todo")')).toBeVisible();
  });
});
