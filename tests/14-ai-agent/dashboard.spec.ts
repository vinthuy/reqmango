import { test, expect } from '../fixtures/auth';

test.describe('AI Agent 仪表盘', () => {
  test('TC-AGT-001: Agent 仪表盘加载', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await expect(page.locator('h1:has-text("Agent Dashboard")')).toBeVisible();
  });

  test('TC-AGT-002: Agent 统计信息', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await expect(page.locator('body')).toContainText('Agent');
  });

  test('TC-AGT-003: 导航到 Agent Templates', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="templates"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/templates/);
  });

  test('TC-AGT-004: 导航到 Skills', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="skills"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/skills/);
  });

  test('TC-AGT-005: 导航到 Tools', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="tools"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/tools/);
  });

  test('TC-AGT-006: 导航到 Configs', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="configs"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/configs/);
  });

  test('TC-AGT-007: 导航到 Tasks', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="tasks"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/tasks/);
  });

  test('TC-AGT-008: 导航到 Loops', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="loops"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/loops/);
  });

  test('TC-AGT-009: 导航到 Monitor', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="monitor"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/monitor/);
  });

  test('TC-AGT-010: 导航到 Performance', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents');
    await page.waitForTimeout(2000);
    await page.click('a[href*="performance"]');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/agents\/performance/);
  });
});
