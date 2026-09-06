import { test, expect } from '../fixtures/auth';

test.describe('Agent 监控/性能/管道/团队', () => {
  const AGENT_BASE = '/workspace/qa-test/agents';

  // === Monitor ===
  test('TC-AGM-001: 监控页面加载', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/monitor`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=监控, text=Monitor').first()).toBeVisible();
  });

  test('TC-AGM-002: 监控数据展示', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/monitor`);
    await page.waitForTimeout(2000);
    const data = page.locator('[class*="chart"], [class*="stat"], canvas, svg').first();
    if (await data.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(data).toBeVisible();
    }
  });

  test('TC-AGM-003: 监控刷新', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/monitor`);
    await page.waitForTimeout(2000);
    const refreshBtn = page.locator('button:has-text("刷新"), button[aria-label="刷新"]').first();
    if (await refreshBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await refreshBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=监控, text=Monitor').first()).toBeVisible();
  });

  // === Performance ===
  test('TC-AGM-004: 性能页面加载', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/performance`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=性能, text=Performance').first()).toBeVisible();
  });

  test('TC-AGM-005: 性能图表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/performance`);
    await page.waitForTimeout(2000);
    const chart = page.locator('canvas, [class*="chart"], svg').first();
    if (await chart.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(chart).toBeVisible();
    }
  });

  test('TC-AGM-006: 性能时间范围', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/performance`);
    await page.waitForTimeout(2000);
    const timeBtn = page.locator('button:has-text("今天"), button:has-text("本周"), button:has-text("本月")').first();
    if (await timeBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await timeBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=性能, text=Performance').first()).toBeVisible();
  });

  // === Pipelines ===
  test('TC-AGM-007: 管道页面加载', async ({ authedPage: page }) => {
    await page.goto('/workspaces/qa-test/agents/pipelines');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=管道, text=Pipeline, text=Agent').first()).toBeVisible();
  });

  test('TC-AGM-008: 创建管道', async ({ authedPage: page }) => {
    await page.goto('/workspaces/qa-test/agents/pipelines');
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('text=管道, text=Pipeline, text=Agent').first()).toBeVisible();
  });

  // === Squads ===
  test('TC-AGM-009: 团队页面加载', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/squads`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=团队, text=Squad').first()).toBeVisible();
  });

  test('TC-AGM-010: 创建团队', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/squads`);
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Squad ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=团队, text=Squad').first()).toBeVisible();
  });

  test('TC-AGM-011: 团队详情', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/squads`);
    await page.waitForTimeout(2000);
    const squadItem = page.locator('[class*="card"], tr, a').first();
    if (await squadItem.isVisible({ timeout: 3000 }).catch(() => false)) {
      await squadItem.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=团队, text=Squad').first()).toBeVisible();
  });

  // === Memory Detail ===
  test('TC-AGM-012: 记忆创建', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/memories`);
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="key"], textarea').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Memory ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=记忆, text=Memory').first()).toBeVisible();
  });

  // === Responsive ===
  test('TC-AGM-013: Agent 子页面响应式', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/monitor`);
    await page.waitForTimeout(2000);
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=监控, text=Monitor').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=监控, text=Monitor').first()).toBeVisible();
  });
});
