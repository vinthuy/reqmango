import { test as base, type Page } from '@playwright/test';

type AuthFixtures = {
  authedPage: Page;
};

export const test = base.extend<AuthFixtures>({
  authedPage: async ({ page }, use) => {
    const email = process.env.TEST_EMAIL || 'qa_tester@reqmango.com';
    const password = process.env.TEST_PASSWORD || 'Test@12345';

    // Navigate to login page
    await page.goto('/login');
    await page.waitForTimeout(500);

    // Set email via native input setter to trigger Vue reactivity
    await page.evaluate(({ email }) => {
      const input = document.querySelector('input[type=email]') as HTMLInputElement;
      if (input) {
        const nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')?.set;
        if (nativeInputValueSetter) {
          nativeInputValueSetter.call(input, email);
        } else {
          input.value = email;
        }
        input.dispatchEvent(new Event('input', { bubbles: true }));
        input.dispatchEvent(new Event('change', { bubbles: true }));
      }
    }, { email });

    // Fill password
    await page.fill('input[type=password]', password);
    await page.waitForTimeout(200);

    // Click login
    await page.click('button:has-text("登录")');

    // Wait for navigation to home page
    await page.waitForURL('**/', { timeout: 10000 });
    await page.waitForTimeout(1000);

    await use(page);
  },
});

export { expect } from '@playwright/test';
