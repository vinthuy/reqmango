import { test, expect } from '../fixtures/auth';

test.describe('SDLC Orchestrator', () => {
  test('TC-AGT-025: SDLC 页面', async ({ authedPage: page }) => {
    await page.goto('/workspace/qa-test/agents/sdlc');
    await page.waitForTimeout(2000);
    await expect(page.locator('main').first()).toBeVisible();
  });
});
