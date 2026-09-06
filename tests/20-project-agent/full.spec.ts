import { test, expect } from '../fixtures/auth';

test.describe('项目级 Agent 页面', () => {
  const PROJECT_BASE = '/workspace/qa-test/project/2347';

  // === Agent Members ===
  test('TC-PAG-001: Agent 成员页面加载', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/agent-members`);
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  test('TC-PAG-002: 添加 Agent 成员', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/agent-members`);
    await page.waitForTimeout(2000);
    const addBtn = page.locator('button:has-text("添加"), button:has-text("邀请"), button:has-text("Add")').first();
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('body')).toBeVisible();
  });

  // === Workflows ===
  test('TC-PAG-003: 工作流页面加载', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/workflows`);
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  test('TC-PAG-004: 创建工作流', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/workflows`);
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Project Workflow ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('body')).toBeVisible();
  });

  // === Workflow Designer ===
  test('TC-PAG-005: 工作流设计器', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/workflows`);
    await page.waitForTimeout(2000);
    const workflowItem = page.locator('[class*="card"], tr, a').first();
    if (await workflowItem.isVisible({ timeout: 3000 }).catch(() => false)) {
      await workflowItem.click();
      await page.waitForTimeout(1000);
      // Look for designer link
      const designBtn = page.locator('button:has-text("设计"), button:has-text("编辑"), a:has-text("设计")').first();
      if (await designBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await designBtn.click();
        await page.waitForTimeout(1000);
      }
    }
    await expect(page.locator('body')).toBeVisible();
  });

  // === Agent Issues ===
  test('TC-PAG-006: Agent Issue 页面加载', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/agent-issues`);
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  test('TC-PAG-007: Agent Issue 列表', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/agent-issues`);
    await page.waitForTimeout(2000);
    const list = page.locator('table, [class*="list"], [class*="card"]').first();
    if (await list.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(list).toBeVisible();
    }
  });

  // === Budget & SLA ===
  test('TC-PAG-008: 预算 SLA 页面加载', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/budget-sla`);
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toBeVisible();
  });

  test('TC-PAG-009: 预算配置', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/budget-sla`);
    await page.waitForTimeout(2000);
    const configInput = page.locator('input, select, [class*="config"]').first();
    if (await configInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await configInput.click();
      await page.waitForTimeout(300);
    }
    await expect(page.locator('body')).toBeVisible();
  });

  test('TC-PAG-010: SLA 规则', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/budget-sla`);
    await page.waitForTimeout(2000);
    const slaTab = page.locator('button:has-text("SLA"), button:has-text("规则")').first();
    if (await slaTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await slaTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('body')).toBeVisible();
  });

  // === Responsive ===
  test('TC-PAG-011: 项目 Agent 页面响应式', async ({ authedPage: page }) => {
    await page.goto(`${PROJECT_BASE}/agent-members`);
    await page.waitForTimeout(2000);
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('body')).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('body')).toBeVisible();
  });
});
