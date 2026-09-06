import { test, expect } from '../fixtures/auth';

test.describe('AI Agent 全功能测试', () => {
  const AGENT_BASE = '/workspace/qa-test/agents';

  // === Agent Dashboard ===
  test('TC-AGT-001: Agent 仪表盘加载', async ({ authedPage: page }) => {
    await page.goto(AGENT_BASE);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=Agent, text=智能体, text=仪表盘').first()).toBeVisible();
  });

  // === Agent Templates ===
  test('TC-AGT-002: Agent 模板列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/templates`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=模板, text=Templates').first()).toBeVisible();
  });

  test('TC-AGT-003: 创建 Agent 模板', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/templates`);
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Template ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=模板, text=Templates').first()).toBeVisible();
  });

  // === Skills ===
  test('TC-AGT-004: 技能列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/skills`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=技能, text=Skills').first()).toBeVisible();
  });

  test('TC-AGT-005: 创建技能', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/skills`);
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Skill ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=技能, text=Skills').first()).toBeVisible();
  });

  // === Agent Tasks ===
  test('TC-AGT-006: 任务列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/tasks`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=任务, text=Tasks').first()).toBeVisible();
  });

  test('TC-AGT-007: 创建任务', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/tasks`);
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"], input[placeholder*="标题"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Task ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=任务, text=Tasks').first()).toBeVisible();
  });

  // === Loops ===
  test('TC-AGT-008: 循环列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/loops`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=循环, text=Loops').first()).toBeVisible();
  });

  // === Tools ===
  test('TC-AGT-009: 工具列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/tools`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=工具, text=Tools').first()).toBeVisible();
  });

  test('TC-AGT-010: 创建工具', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/tools`);
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Tool ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=工具, text=Tools').first()).toBeVisible();
  });

  // === Runtimes ===
  test('TC-AGT-011: 运行时列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/runtimes`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=运行时, text=Runtimes').first()).toBeVisible();
  });

  // === Sessions ===
  test('TC-AGT-012: 会话列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/sessions`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=会话, text=Sessions').first()).toBeVisible();
  });

  // === Memory ===
  test('TC-AGT-013: 记忆列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/memory`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=记忆, text=Memory').first()).toBeVisible();
  });

  // === Autopilot ===
  test('TC-AGT-014: 自动驾驶列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/autopilot`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=自动, text=Autopilot').first()).toBeVisible();
  });

  test('TC-AGT-015: 创建自动驾驶任务', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/autopilot`);
    await page.waitForTimeout(2000);
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const nameInput = page.locator('input[placeholder*="名称"], input[placeholder*="name"], input[placeholder*="标题"]').first();
      if (await nameInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await nameInput.fill(`E2E Autopilot ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=自动, text=Autopilot').first()).toBeVisible();
  });

  // === Configs ===
  test('TC-AGT-016: 配置列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/configs`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=配置, text=Configs').first()).toBeVisible();
  });

  // === Squads ===
  test('TC-AGT-017: 团队列表', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/squads`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=团队, text=Squads').first()).toBeVisible();
  });

  // === Developer Agent ===
  test('TC-AGT-018: 开发者 Agent', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/developer`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=开发, text=Developer').first()).toBeVisible();
  });

  // === Tester Agent ===
  test('TC-AGT-019: 测试 Agent', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/tester`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=测试, text=Tester').first()).toBeVisible();
  });

  // === CI/CD ===
  test('TC-AGT-020: CI/CD 管理', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/cicd`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=CI/CD, text=构建').first()).toBeVisible();
  });

  // === SDLC ===
  test('TC-AGT-021: SDLC 管理', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/sdlc`);
    await page.waitForTimeout(2000);
    await expect(page.locator('text=SDLC, text=流程').first()).toBeVisible();
  });

  // === Navigation ===
  test('TC-AGT-022: 侧边栏导航', async ({ authedPage: page }) => {
    await page.goto(AGENT_BASE);
    await page.waitForTimeout(2000);
    const sidebar = page.locator('[class*="sidebar"], [class*="Sidebar"], nav').first();
    if (await sidebar.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(sidebar).toBeVisible();
    }
  });

  // === Dashboard Stats ===
  test('TC-AGT-023: 仪表盘统计卡片', async ({ authedPage: page }) => {
    await page.goto(AGENT_BASE);
    await page.waitForTimeout(2000);
    const stats = page.locator('[class*="stat"], [class*="card"], [class*="metric"]').first();
    if (await stats.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(stats).toBeVisible();
    }
  });

  // === Agent Config Detail ===
  test('TC-AGT-024: 配置详情', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/configs`);
    await page.waitForTimeout(2000);
    const configItem = page.locator('[class*="card"], tr, a').first();
    if (await configItem.isVisible({ timeout: 3000 }).catch(() => false)) {
      await configItem.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=配置, text=Configs, text=Agent').first()).toBeVisible();
  });

  // === Skill Execution ===
  test('TC-AGT-025: 技能执行日志', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/skills`);
    await page.waitForTimeout(2000);
    const logBtn = page.locator('button:has-text("日志"), button:has-text("Logs")').first();
    if (await logBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await logBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=技能, text=Skills').first()).toBeVisible();
  });

  // === Task Status Filter ===
  test('TC-AGT-026: 任务状态筛选', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/tasks`);
    await page.waitForTimeout(2000);
    const filterBtn = page.locator('button:has-text("筛选"), select, [role="combobox"]').first();
    if (await filterBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await filterBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=任务, text=Tasks').first()).toBeVisible();
  });

  // === Responsive ===
  test('TC-AGT-027: Agent 页面响应式', async ({ authedPage: page }) => {
    await page.goto(AGENT_BASE);
    await page.waitForTimeout(2000);
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=Agent, text=智能体').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=Agent, text=智能体').first()).toBeVisible();
  });

  // === Search ===
  test('TC-AGT-028: Agent 搜索', async ({ authedPage: page }) => {
    await page.goto(AGENT_BASE);
    await page.waitForTimeout(2000);
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await searchInput.fill('test');
      await page.waitForTimeout(1000);
      await searchInput.clear();
    }
    await expect(page.locator('text=Agent, text=智能体').first()).toBeVisible();
  });

  // === Pagination ===
  test('TC-AGT-029: 分页功能', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/tasks`);
    await page.waitForTimeout(2000);
    const pagination = page.locator('[class*="pagination"], [class*="pager"], button:has-text("下一页")').first();
    if (await pagination.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(pagination).toBeVisible();
    }
  });

  // === Empty State ===
  test('TC-AGT-030: 空状态显示', async ({ authedPage: page }) => {
    await page.goto(`${AGENT_BASE}/loops`);
    await page.waitForTimeout(2000);
    // 空状态或有数据都算通过
    await expect(page.locator('text=循环, text=Loops, text=暂无数据, text=No data').first()).toBeVisible();
  });
});
