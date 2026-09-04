import { test, expect } from '@playwright/test';

test.describe('用户登录', () => {
  test('TC-AUTH-004: 使用有效凭据登录', async ({ page }) => {
    await page.goto('/login');
    await page.evaluate(() => {
      const input = document.querySelector('input[type=email]') as HTMLInputElement;
      if (input) {
        const nativeSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')?.set;
        if (nativeSetter) nativeSetter.call(input, 'qa_tester@reqmango.com');
        else input.value = 'qa_tester@reqmango.com';
        input.dispatchEvent(new Event('input', { bubbles: true }));
        input.dispatchEvent(new Event('change', { bubbles: true }));
      }
    });
    await page.fill('input[type=password]', 'Test@12345');
    await page.click('button:has-text("登录")');
    await page.waitForTimeout(2000);
    await page.waitForURL('**/', { timeout: 15000 }).catch(() => {});
    await expect(page.locator('body')).toContainText('工作空间', { timeout: 10000 });
  });

  test('TC-AUTH-005: 使用无效密码登录失败', async ({ page }) => {
    await page.goto('/login');
    await page.evaluate(() => {
      const input = document.querySelector('input[type=email]') as HTMLInputElement;
      if (input) {
        const nativeSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')?.set;
        if (nativeSetter) nativeSetter.call(input, 'qa_tester@reqmango.com');
        else input.value = 'qa_tester@reqmango.com';
        input.dispatchEvent(new Event('input', { bubbles: true }));
        input.dispatchEvent(new Event('change', { bubbles: true }));
      }
    });
    await page.fill('input[type=password]', 'WrongPassword');
    await page.click('button:has-text("登录")');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/login/);
  });

  test('TC-AUTH-006: 未登录时跳转到登录页', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/login/);
  });
});
