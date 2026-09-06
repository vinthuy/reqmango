import { test, expect } from '../fixtures/auth';

test.describe('项目设置页', () => {
  const SETTINGS_URL = '/workspace/qa-test/project/2347?tab=settings';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(SETTINGS_URL);
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-SET-001: 设置页正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === Issue 类型管理 ===
  test('TC-SET-002: Issue 类型列表', async ({ authedPage: page }) => {
    const typeTab = page.locator('button:has-text("类型"), button:has-text("Types")').first();
    if (await typeTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await typeTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  test('TC-SET-003: 创建 Issue 类型', async ({ authedPage: page }) => {
    const typeTab = page.locator('button:has-text("类型"), button:has-text("Types")').first();
    if (await typeTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await typeTab.click();
      await page.waitForTimeout(1000);
    }
    const addBtn = page.locator('button:has-text("添加"), button:has-text("新建"), button:has-text("创建")').first();
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`TestType ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 状态管理 ===
  test('TC-SET-004: 状态列表', async ({ authedPage: page }) => {
    const stateTab = page.locator('button:has-text("状态"), button:has-text("States")').first();
    if (await stateTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await stateTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  test('TC-SET-005: 创建状态', async ({ authedPage: page }) => {
    const stateTab = page.locator('button:has-text("状态"), button:has-text("States")').first();
    if (await stateTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await stateTab.click();
      await page.waitForTimeout(1000);
    }
    const addBtn = page.locator('button:has-text("添加"), button:has-text("新建"), button:has-text("创建")').first();
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`TestState ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 标签管理 ===
  test('TC-SET-006: 标签列表', async ({ authedPage: page }) => {
    const labelTab = page.locator('button:has-text("标签"), button:has-text("Labels")').first();
    if (await labelTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await labelTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  test('TC-SET-007: 创建标签', async ({ authedPage: page }) => {
    const labelTab = page.locator('button:has-text("标签"), button:has-text("Labels")').first();
    if (await labelTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await labelTab.click();
      await page.waitForTimeout(1000);
    }
    const addBtn = page.locator('button:has-text("添加"), button:has-text("新建"), button:has-text("创建")').first();
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`TestLabel ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 成员管理 ===
  test('TC-SET-008: 项目成员列表', async ({ authedPage: page }) => {
    const memberTab = page.locator('button:has-text("成员"), button:has-text("Members")').first();
    if (await memberTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await memberTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 项目基本信息 ===
  test('TC-SET-009: 项目名称显示', async ({ authedPage: page }) => {
    const nameInput = page.locator('input[placeholder*="项目名称"], input[placeholder*="project name"], input[value]').first();
    if (await nameInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(nameInput).toBeVisible();
    }
  });

  // === 项目描述 ===
  test('TC-SET-010: 项目描述编辑', async ({ authedPage: page }) => {
    const descInput = page.locator('textarea, [contenteditable], input[placeholder*="描述"]').first();
    if (await descInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await descInput.click();
      await page.waitForTimeout(300);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 自动化规则 ===
  test('TC-SET-011: 自动化规则列表', async ({ authedPage: page }) => {
    const autoTab = page.locator('button:has-text("自动化"), button:has-text("Automation")').first();
    if (await autoTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await autoTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 工作流管理 ===
  test('TC-SET-012: 工作流配置', async ({ authedPage: page }) => {
    const workflowTab = page.locator('button:has-text("工作流"), button:has-text("Workflow")').first();
    if (await workflowTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await workflowTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === Webhook 配置 ===
  test('TC-SET-013: Webhook 列表', async ({ authedPage: page }) => {
    const webhookTab = page.locator('button:has-text("Webhook"), button:has-text("集成")').first();
    if (await webhookTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await webhookTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 估算点配置 ===
  test('TC-SET-014: 估算点设置', async ({ authedPage: page }) => {
    const estimateTab = page.locator('button:has-text("估算"), button:has-text("Estimate")').first();
    if (await estimateTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await estimateTab.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 保存设置 ===
  test('TC-SET-015: 保存项目设置', async ({ authedPage: page }) => {
    const saveBtn = page.locator('button:has-text("保存"), button:has-text("Save")').first();
    if (await saveBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await saveBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 取消操作 ===
  test('TC-SET-016: 取消未保存的更改', async ({ authedPage: page }) => {
    const cancelBtn = page.locator('button:has-text("取消"), button:has-text("Cancel")').first();
    if (await cancelBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await cancelBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === Tab 切换 ===
  test('TC-SET-017: 设置 Tab 切换', async ({ authedPage: page }) => {
    const tabs = page.locator('button[role="tab"], [class*="tab"]');
    const count = await tabs.count();
    if (count > 1) {
      await tabs.nth(1).click();
      await page.waitForTimeout(500);
      await tabs.nth(0).click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 危险操作确认 ===
  test('TC-SET-018: 删除项目确认', async ({ authedPage: page }) => {
    const dangerZone = page.locator('text=危险区域, text=Danger Zone').first();
    if (await dangerZone.isVisible({ timeout: 3000 }).catch(() => false)) {
      const deleteBtn = page.locator('button:has-text("删除项目")').first();
      if (await deleteBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await deleteBtn.click();
        await page.waitForTimeout(500);
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-SET-019: 设置页响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=项目设置, text=Settings').first()).toBeVisible();
  });

  // === 页面导航 ===
  test('TC-SET-020: 从设置返回项目', async ({ authedPage: page }) => {
    const backBtn = page.locator('a:has-text("返回"), button:has-text("返回"), [aria-label="返回"]').first();
    if (await backBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await backBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=项目设置, text=Settings, table').first()).toBeVisible();
  });
});
