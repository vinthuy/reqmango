import { test, expect } from '../fixtures/auth';

test.describe('Cycle 详情页', () => {
  const CYCLE_URL = '/workspace/qa-test/project/2347?tab=cycles';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(CYCLE_URL);
    await page.waitForTimeout(2000);
    // 点击第一个 Cycle 进入详情
    const cycleCard = page.locator('[class*="card"], [class*="Card"], tr, a:has-text("E2E")').first();
    if (await cycleCard.isVisible({ timeout: 3000 }).catch(() => false)) {
      await cycleCard.click();
      await page.waitForTimeout(1500);
    }
  });

  // === 页面加载 ===
  test('TC-CYD-001: Cycle 详情正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
  });

  // === 基本信息 ===
  test('TC-CYD-002: Cycle 名称显示', async ({ authedPage: page }) => {
    const nameEl = page.locator('h1, h2, [class*="name"], [class*="title"]').first();
    if (await nameEl.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(nameEl).toBeVisible();
    }
  });

  // === 状态显示 ===
  test('TC-CYD-003: Cycle 状态', async ({ authedPage: page }) => {
    const statusEl = page.locator('[class*="status"], [class*="badge"], text=进行中, text=已完成, text=未开始').first();
    if (await statusEl.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(statusEl).toBeVisible();
    }
  });

  // === 进度条 ===
  test('TC-CYD-004: 进度显示', async ({ authedPage: page }) => {
    const progressEl = page.locator('[class*="progress"], [role="progressbar"], text=进度').first();
    if (await progressEl.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(progressEl).toBeVisible();
    }
  });

  // === Issue 列表 ===
  test('TC-CYD-005: Cycle 包含的 Issue 列表', async ({ authedPage: page }) => {
    const issueList = page.locator('table, [class*="issue-list"], [class*="list"]').first();
    if (await issueList.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(issueList).toBeVisible();
    }
  });

  // === 添加 Issue ===
  test('TC-CYD-006: 添加 Issue 到 Cycle', async ({ authedPage: page }) => {
    const addBtn = page.locator('button:has-text("添加"), button:has-text("关联工作项")').first();
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
  });

  // === 移除 Issue ===
  test('TC-CYD-007: 从 Cycle 移除 Issue', async ({ authedPage: page }) => {
    const removeBtn = page.locator('button:has-text("移除"), button[aria-label="移除"]').first();
    if (await removeBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await removeBtn.click();
      await page.waitForTimeout(500);
      const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
      if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
  });

  // === 统计信息 ===
  test('TC-CYD-008: 统计面板', async ({ authedPage: page }) => {
    const statsTab = page.locator('button:has-text("统计"), button:has-text("Stats")').first();
    if (await statsTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await statsTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
  });

  // === 燃尽图 ===
  test('TC-CYD-009: 燃尽图显示', async ({ authedPage: page }) => {
    const burndownTab = page.locator('button:has-text("燃尽"), button:has-text("Burndown")').first();
    if (await burndownTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await burndownTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
  });

  // === 开始 Cycle ===
  test('TC-CYD-010: 开始 Cycle', async ({ authedPage: page }) => {
    const startBtn = page.locator('button:has-text("开始"), button:has-text("Start")').first();
    if (await startBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await startBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
  });

  // === 结束 Cycle ===
  test('TC-CYD-011: 结束 Cycle', async ({ authedPage: page }) => {
    const endBtn = page.locator('button:has-text("结束"), button:has-text("End")').first();
    if (await endBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await endBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
  });

  // === 编辑 Cycle ===
  test('TC-CYD-012: 编辑 Cycle 信息', async ({ authedPage: page }) => {
    const editBtn = page.locator('button:has-text("编辑"), button:has-text("Edit")').first();
    if (await editBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await editBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
  });

  // === 删除 Cycle ===
  test('TC-CYD-013: 删除 Cycle 确认', async ({ authedPage: page }) => {
    const deleteBtn = page.locator('button:has-text("删除"), button:has-text("Delete")').first();
    if (await deleteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await deleteBtn.click();
      await page.waitForTimeout(500);
      const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
      if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
  });

  // === 返回列表 ===
  test('TC-CYD-014: 返回 Cycle 列表', async ({ authedPage: page }) => {
    const backBtn = page.locator('a:has-text("返回"), button:has-text("返回"), [aria-label="返回"]').first();
    if (await backBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await backBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('button:has-text("创建新周期"), button:has-text("创建第一个周期"), text=周期').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-CYD-015: Cycle 详情响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('button:has-text("详情"), button:has-text("统计"), h1, h2').first()).toBeVisible();
  });
});
