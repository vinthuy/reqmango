import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: '.',
  timeout: 60000,
  expect: { timeout: 10000 },
  fullyParallel: false,
  workers: 1,
  retries: 1,
  reporter: [
    ['html', { open: 'never' }],
    ['list']
  ],
  use: {
    baseURL: 'http://localhost:5173',
    screenshot: 'only-on-failure',
    trace: 'on-first-retry',
    locale: 'zh-CN',
    // Without an explicit actionTimeout, Playwright falls back to the whole
    // test timeout (60s) for any action whose target never becomes actionable.
    // Several specs click broad selectors that can resolve to elements behind
    // an open modal/overlay, which then burned the full 60s per attempt and
    // twice that with retries. These bounds keep a non-actionable element a
    // fast, clearly-reported failure instead of a multi-minute stall, and they
    // do not weaken any assertion.
    actionTimeout: 10000,
    navigationTimeout: 30000,
  },
  projects: [
    { name: 'chrome', use: { channel: 'chrome' } },
  ],
});
