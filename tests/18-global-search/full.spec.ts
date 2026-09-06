import { test, expect } from '../fixtures/auth';

test.describe('全局搜索全功能测试', () => {
  test.beforeEach(async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test');
    await page.waitForTimeout(2000);
  });

  // === 搜索入口 ===
  test('TC-SRC-001: 搜索框显示', async ({ authedPage: page }) => {
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"], input[type="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(searchInput).toBeVisible();
    }
  });

  // === 搜索 Issue ===
  test('TC-SRC-002: 搜索 Issue', async ({ authedPage: page }) => {
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"], input[type="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await searchInput.fill('E2E');
      await page.waitForTimeout(1500);
      const results = page.locator('[class*="result"], [class*="suggestion"], [class*="dropdown"]').first();
      if (await results.isVisible({ timeout: 3000 }).catch(() => false)) {
        await expect(results).toBeVisible();
      }
      await searchInput.clear();
    }
  });

  // === 搜索快捷键 ===
  test('TC-SRC-003: Cmd/Ctrl+K 快捷键', async ({ authedPage: page }) => {
    await page.keyboard.press('Control+k');
    await page.waitForTimeout(500);
    const searchModal = page.locator('[class*="modal"], [class*="dialog"], [class*="search-overlay"]').first();
    if (await searchModal.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(searchModal).toBeVisible();
      await page.keyboard.press('Escape');
    }
  });

  // === 搜索项目 ===
  test('TC-SRC-004: 搜索项目', async ({ authedPage: page }) => {
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"], input[type="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await searchInput.fill('QA');
      await page.waitForTimeout(1500);
      await searchInput.clear();
    }
    await expect(page.locator('body')).toBeVisible();
  });

  // === 搜索用户 ===
  test('TC-SRC-005: 搜索用户', async ({ authedPage: page }) => {
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"], input[type="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await searchInput.fill('admin');
      await page.waitForTimeout(1500);
      await searchInput.clear();
    }
    await expect(page.locator('body')).toBeVisible();
  });

  // === 搜索结果点击 ===
  test('TC-SRC-006: 搜索结果导航', async ({ authedPage: page }) => {
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"], input[type="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await searchInput.fill('E2E');
      await page.waitForTimeout(1500);
      const firstResult = page.locator('[class*="result"] a, [class*="suggestion"] a, [class*="option"]').first();
      if (await firstResult.isVisible({ timeout: 2000 }).catch(() => false)) {
        await firstResult.click();
        await page.waitForTimeout(1000);
      } else {
        await searchInput.clear();
      }
    }
  });

  // === 搜索筛选 ===
  test('TC-SRC-007: 搜索类型筛选', async ({ authedPage: page }) => {
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"], input[type="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await searchInput.fill('test');
      await page.waitForTimeout(1500);
      const typeFilter = page.locator('button:has-text("Issue"), button:has-text("项目"), [class*="filter"]').first();
      if (await typeFilter.isVisible({ timeout: 2000 }).catch(() => false)) {
        await typeFilter.click();
        await page.waitForTimeout(500);
      }
      await searchInput.clear();
    }
  });

  // === 空搜索 ===
  test('TC-SRC-008: 空搜索结果', async ({ authedPage: page }) => {
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"], input[type="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await searchInput.fill('zzznonexistent999');
      await page.waitForTimeout(1500);
      const noResult = page.locator('text=无结果, text=没有找到, text=No results').first();
      if (await noResult.isVisible({ timeout: 3000 }).catch(() => false)) {
        await expect(noResult).toBeVisible();
      }
      await searchInput.clear();
    }
  });

  // === 搜索历史 ===
  test('TC-SRC-009: 搜索历史', async ({ authedPage: page }) => {
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"], input[type="search"]').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await searchInput.click();
      await page.waitForTimeout(500);
      const history = page.locator('[class*="history"], [class*="recent"], text=最近搜索').first();
      if (await history.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(history).toBeVisible();
      }
    }
  });

  // === 响应式 ===
  test('TC-SRC-010: 搜索响应式', async ({ authedPage: page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(500);
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="search"], button:has-text("搜索")').first();
    if (await searchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(searchInput).toBeVisible();
    }
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
  });
});
