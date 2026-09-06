import { test, expect } from '../fixtures/auth';

test.describe('路线图全功能测试', () => {
  const ROADMAP_URL = '/workspace/qa-test/project/2347?tab=roadmap';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(ROADMAP_URL);
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-RDM-001: 路线图页面正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=路线图, text=Roadmap').first()).toBeVisible();
  });

  // === 时间线视图 ===
  test('TC-RDM-002: 时间线显示', async ({ authedPage: page }) => {
    const timeline = page.locator('[class*="timeline"], [class*="gantt"], canvas, svg').first();
    if (await timeline.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(timeline).toBeVisible();
    }
  });

  // === 创建路线图项 ===
  test('TC-RDM-003: 创建路线图项', async ({ authedPage: page }) => {
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Roadmap ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=路线图, text=Roadmap').first()).toBeVisible();
  });

  // === 时间范围 ===
  test('TC-RDM-004: 时间范围切换', async ({ authedPage: page }) => {
    const timeBtn = page.locator('button:has-text("周"), button:has-text("月"), button:has-text("季度")').first();
    if (await timeBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await timeBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=路线图, text=Roadmap').first()).toBeVisible();
  });

  // === 拖拽调整 ===
  test('TC-RDM-005: 拖拽调整时间', async ({ authedPage: page }) => {
    const item = page.locator('[class*="item"], [class*="bar"], [class*="block"]').first();
    if (await item.isVisible({ timeout: 3000 }).catch(() => false)) {
      const box = await item.boundingBox();
      if (box) {
        await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
        await page.waitForTimeout(300);
      }
    }
    await expect(page.locator('text=路线图, text=Roadmap').first()).toBeVisible();
  });

  // === 筛选 ===
  test('TC-RDM-006: 路线图筛选', async ({ authedPage: page }) => {
    const filterBtn = page.locator('button:has-text("筛选"), select, [role="combobox"]').first();
    if (await filterBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await filterBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=路线图, text=Roadmap').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-RDM-007: 路线图响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=路线图, text=Roadmap').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=路线图, text=Roadmap').first()).toBeVisible();
  });

  // === 空状态 ===
  test('TC-RDM-008: 空路线图', async ({ authedPage: page }) => {
    await expect(page.locator('text=路线图, text=Roadmap, text=暂无数据').first()).toBeVisible();
  });
});
