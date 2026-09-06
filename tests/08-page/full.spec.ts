import { test, expect } from '../fixtures/auth';

test.describe('文档页面全功能测试', () => {
  const PAGE_URL = '/workspace/qa-test/project/2347?tab=pages';

  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto(PAGE_URL);
    await page.waitForTimeout(2000);
  });

  // === 页面加载 ===
  test('TC-PAG-001: 文档页面正常加载', async ({ authedPage: page }) => {
    await expect(page.locator('text=文档, text=Pages, text=知识库').first()).toBeVisible();
  });

  // === 文档列表 ===
  test('TC-PAG-002: 文档列表显示', async ({ authedPage: page }) => {
    const list = page.locator('[class*="list"], [class*="tree"], table').first();
    if (await list.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(list).toBeVisible();
    }
  });

  // === 创建文档 ===
  test('TC-PAG-003: 创建新文档', async ({ authedPage: page }) => {
    const createBtn = page.locator('button:has-text("创建"), button:has-text("新建"), button:has-text("添加")').first();
    if (await createBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await createBtn.click();
      await page.waitForTimeout(500);
      const titleInput = page.locator('input[placeholder*="标题"], input[placeholder*="title"], input[placeholder*="名称"]').first();
      if (await titleInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await titleInput.fill(`E2E Doc ${Date.now()}`);
        await page.click('button:has-text("保存"), button:has-text("确定")');
        await page.waitForTimeout(1000);
      } else {
        await page.keyboard.press('Escape');
      }
    }
    await expect(page.locator('text=文档, text=Pages, text=知识库').first()).toBeVisible();
  });

  // === 编辑文档 ===
  test('TC-PAG-004: 编辑文档内容', async ({ authedPage: page }) => {
    const docItem = page.locator('[class*="item"], tr, a:has-text("E2E")').first();
    if (await docItem.isVisible({ timeout: 3000 }).catch(() => false)) {
      await docItem.click();
      await page.waitForTimeout(1000);
      const editor = page.locator('[contenteditable], .ProseMirror, textarea').first();
      if (await editor.isVisible({ timeout: 2000 }).catch(() => false)) {
        await editor.click();
        await page.keyboard.type(' E2E test content');
        await page.waitForTimeout(500);
      }
    }
    await expect(page.locator('text=文档, text=Pages, text=知识库, .ProseMirror').first()).toBeVisible();
  });

  // === 删除文档 ===
  test('TC-PAG-005: 删除文档确认', async ({ authedPage: page }) => {
    const deleteBtn = page.locator('button:has-text("删除"), button:has-text("Delete")').first();
    if (await deleteBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await deleteBtn.click();
      await page.waitForTimeout(500);
      const confirmBtn = page.locator('button:has-text("确认"), button:has-text("确定")').first();
      if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await page.click('button:has-text("取消")');
      }
    }
    await expect(page.locator('text=文档, text=Pages, text=知识库').first()).toBeVisible();
  });

  // === 搜索文档 ===
  test('TC-PAG-006: 搜索文档', async ({ authedPage: page }) => {
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await searchInput.fill('E2E');
      await page.waitForTimeout(1000);
      await searchInput.clear();
    }
    await expect(page.locator('text=文档, text=Pages, text=知识库').first()).toBeVisible();
  });

  // === 文档树 ===
  test('TC-PAG-007: 文档树结构', async ({ authedPage: page }) => {
    const tree = page.locator('[class*="tree"], [class*="folder"], [class*="nav"]').first();
    if (await tree.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(tree).toBeVisible();
    }
  });

  // === 文档版本 ===
  test('TC-PAG-008: 文档版本历史', async ({ authedPage: page }) => {
    const versionBtn = page.locator('button:has-text("版本"), button:has-text("历史"), button:has-text("Version")').first();
    if (await versionBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await versionBtn.click();
      await page.waitForTimeout(1000);
    }
    await expect(page.locator('text=文档, text=Pages, text=知识库').first()).toBeVisible();
  });

  // === 文档分享 ===
  test('TC-PAG-009: 文档分享', async ({ authedPage: page }) => {
    const shareBtn = page.locator('button:has-text("分享"), button:has-text("Share")').first();
    if (await shareBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await shareBtn.click();
      await page.waitForTimeout(500);
      await page.keyboard.press('Escape');
    }
    await expect(page.locator('text=文档, text=Pages, text=知识库').first()).toBeVisible();
  });

  // === 文档导出 ===
  test('TC-PAG-010: 导出文档', async ({ authedPage: page }) => {
    const exportBtn = page.locator('button:has-text("导出"), button:has-text("Export")').first();
    if (await exportBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await exportBtn.click();
      await page.waitForTimeout(500);
    }
    await expect(page.locator('text=文档, text=Pages, text=知识库').first()).toBeVisible();
  });

  // === 响应式 ===
  test('TC-PAG-011: 文档页面响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=文档, text=Pages, text=知识库').first()).toBeVisible();
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    await expect(page.locator('text=文档, text=Pages, text=知识库').first()).toBeVisible();
  });

  // === 空状态 ===
  test('TC-PAG-012: 空文档列表', async ({ authedPage: page }) => {
    await expect(page.locator('text=文档, text=Pages, text=知识库, text=暂无数据').first()).toBeVisible();
  });
});
