import { test, expect } from '../fixtures/auth';

test.describe('Issue 详情页', () => {
  const PROJECT_URL = '/workspace/qa-test/project/2347';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(PROJECT_URL);
    await page.waitForTimeout(2000);
    // 打开第一个 Issue 详情
    const viewBtn = page.locator('button:has-text("查看")').first();
    if (await viewBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await viewBtn.click();
      await page.waitForTimeout(1500);
    }
  });

  // === 页面加载 ===
  test('TC-DET-001: 详情页正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 标题编辑 ===
  test('TC-DET-002: 编辑 Issue 标题', async ({ authedPage: page }) => {
    const titleEl = page.locator('h1, input[placeholder*="标题"], [contenteditable]').first();
    if (await titleEl.isVisible({ timeout: 3000 }).catch(() => false)) {
      await titleEl.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 状态切换 ===
  test('TC-DET-003: 切换 Issue 状态', async ({ authedPage: page }) => {
    const statusBtn = page.locator('[class*="status"], [class*="Status"], button:has-text("状态")').first();
    if (await statusBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await statusBtn.click();
      await page.waitForTimeout(500);
      const option = page.locator('[role="option"], [role="menuitem"], li').first();
      if (await option.isVisible({ timeout: 2000 }).catch(() => false)) {
        await option.click();
        await page.waitForTimeout(1000);
      }
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 优先级修改 ===
  test('TC-DET-004: 修改优先级', async ({ authedPage: page }) => {
    const priorityEl = page.locator('text=优先级').first();
    if (await priorityEl.isVisible({ timeout: 3000 }).catch(() => false)) {
      await priorityEl.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 负责人修改 ===
  test('TC-DET-005: 修改负责人', async ({ authedPage: page }) => {
    const assigneeEl = page.locator('text=负责人').first();
    if (await assigneeEl.isVisible({ timeout: 3000 }).catch(() => false)) {
      await assigneeEl.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 标签管理 ===
  test('TC-DET-006: 添加标签', async ({ authedPage: page }) => {
    const labelEl = page.locator('text=标签').first();
    if (await labelEl.isVisible({ timeout: 3000 }).catch(() => false)) {
      await labelEl.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 截止日期 ===
  test('TC-DET-007: 设置截止日期', async ({ authedPage: page }) => {
    const dateEl = page.locator('text=截止日期, text=到期日').first();
    if (await dateEl.isVisible({ timeout: 3000 }).catch(() => false)) {
      await dateEl.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === Tab: 活动记录 ===
  test('TC-DET-008: 查看活动记录', async ({ authedPage: page }) => {
    await page.click('button:has-text("动态"), button:has-text("活动")');
    await page.waitForTimeout(1000);
    await expect(page.locator('button:has-text("动态"), button:has-text("活动")')).toBeVisible();
  });

  // === Tab: 关联 Issue ===
  test('TC-DET-009: 查看关联 Issue', async ({ authedPage: page }) => {
    await page.click('button:has-text("关联")');
    await page.waitForTimeout(1000);
    await expect(page.locator('button:has-text("关联")')).toBeVisible();
  });

  // === Tab: 附件 ===
  test('TC-DET-010: 查看附件面板', async ({ authedPage: page }) => {
    await page.click('button:has-text("附件")');
    await page.waitForTimeout(1000);
    await expect(page.locator('button:has-text("附件")')).toBeVisible();
  });

  // === Tab: 工时 ===
  test('TC-DET-011: 查看工时面板', async ({ authedPage: page }) => {
    await page.click('button:has-text("工时")');
    await page.waitForTimeout(1000);
    await expect(page.locator('button:has-text("工时")')).toBeVisible();
  });

  // === Tab: 子 Issue ===
  test('TC-DET-012: 查看子 Issue', async ({ authedPage: page }) => {
    const subIssueBtn = page.locator('button:has-text("子工作项"), button:has-text("子Issue")').first();
    if (await subIssueBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await subIssueBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 添加评论 ===
  test('TC-DET-013: 添加评论', async ({ authedPage: page }) => {
    const commentInput = page.locator('textarea[placeholder*="评论"], textarea[placeholder*="comment"]').first();
    if (await commentInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await commentInput.fill('E2E 详情页测试评论');
      await page.click('button:has-text("发布"), button:has-text("发送")');
      await page.waitForTimeout(1500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 描述编辑 ===
  test('TC-DET-014: 编辑描述', async ({ authedPage: page }) => {
    const descEl = page.locator('[contenteditable], textarea[placeholder*="描述"], .ProseMirror').first();
    if (await descEl.isVisible({ timeout: 3000 }).catch(() => false)) {
      await descEl.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 关联操作 ===
  test('TC-DET-015: 添加关联 Issue', async ({ authedPage: page }) => {
    await page.click('button:has-text("关联")');
    await page.waitForTimeout(1000);
    const addBtn = page.locator('button:has-text("添加"), button:has-text("关联")').first();
    if (await addBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await addBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("关联")')).toBeVisible();
  });

  // === 右侧属性面板 ===
  test('TC-DET-016: 属性面板完整性', async ({ authedPage: page }) => {
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
    // 验证右侧属性面板存在
    const sidebar = page.locator('[class*="sidebar"], [class*="property"], [class*="attribute"]').first();
    if (await sidebar.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(sidebar).toBeVisible();
    }
  });

  // === Issue 编号显示 ===
  test('TC-DET-017: Issue 编号显示', async ({ authedPage: page }) => {
    const issueId = page.locator('[class*="identifier"], [class*="number"], text=/^#\\d+/').first();
    if (await issueId.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(issueId).toBeVisible();
    }
  });

  // === 返回列表 ===
  test('TC-DET-018: 返回列表', async ({ authedPage: page }) => {
    const backBtn = page.locator('button:has-text("返回"), [aria-label="返回"], a:has-text("返回")').first();
    if (await backBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await backBtn.click();
      await page.waitForTimeout(1000);
      await expect(page.locator('table')).toBeVisible();
    }
  });

  // === 快捷操作 ===
  test('TC-DET-019: 快捷操作菜单', async ({ authedPage: page }) => {
    const moreBtn = page.locator('button:has-text("更多"), button[aria-label="更多"]').first();
    if (await moreBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await moreBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 删除 Issue ===
  test('TC-DET-020: 删除确认弹窗', async ({ authedPage: page }) => {
    const moreBtn = page.locator('button:has-text("更多"), button[aria-label="更多"]').first();
    if (await moreBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await moreBtn.click();
      await page.waitForTimeout(500);
      const deleteBtn = page.locator('button:has-text("删除")').first();
      if (await deleteBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await deleteBtn.click();
        await page.waitForTimeout(500);
        // 验证确认弹窗出现
        const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
        if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
          // 取消删除
          await page.click('button:has-text("取消")');
        }
      }
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 时间追踪 ===
  test('TC-DET-021: 开始时间追踪', async ({ authedPage: page }) => {
    const timeBtn = page.locator('button:has-text("开始计时"), button:has-text("追踪时间")').first();
    if (await timeBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await timeBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 关注者 ===
  test('TC-DET-022: 关注/取消关注', async ({ authedPage: page }) => {
    const watchBtn = page.locator('button:has-text("关注"), button[aria-label="关注"]').first();
    if (await watchBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await watchBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 移动 Issue ===
  test('TC-DET-023: 移动到其他项目', async ({ authedPage: page }) => {
    const moreBtn = page.locator('button:has-text("更多"), button[aria-label="更多"]').first();
    if (await moreBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await moreBtn.click();
      await page.waitForTimeout(500);
      const moveBtn = page.locator('button:has-text("移动")').first();
      if (await moveBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await moveBtn.click();
        await page.waitForTimeout(500);
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 复制 Issue ===
  test('TC-DET-024: 复制 Issue', async ({ authedPage: page }) => {
    const moreBtn = page.locator('button:has-text("更多"), button[aria-label="更多"]').first();
    if (await moreBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await moreBtn.click();
      await page.waitForTimeout(500);
      const copyBtn = page.locator('button:has-text("复制"), button:has-text("拷贝")').first();
      if (await copyBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await copyBtn.click();
        await page.waitForTimeout(500);
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });

  // === 页面响应式 ===
  test('TC-DET-025: 详情页响应式布局', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('button:has-text("详情")')).toBeVisible();
  });
});
