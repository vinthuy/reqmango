import { test, expect } from '../fixtures/auth';

test.describe('CI/CD Manager', () => {
  test('TC-AGT-024: CI/CD 页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/cicd');
    await page.waitForTimeout(2000);
    await expect(page.locator('main').first()).toBeVisible();
  });
});
