import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/__tests__/setup.ts'],
    include: ['src/**/*.test.ts', 'src/**/*.spec.ts'],
    // The suite runs 70+ files in parallel, and several tests mount full Antd
    // pages or dynamically import large modules. On slower machines or loaded
    // CI runners those legitimately need more than Vitest's 5s default, which
    // turned into timeouts unrelated to the behaviour under test — the
    // assertions inside each test still fail fast when something is wrong.
    testTimeout: 20000,
    hookTimeout: 20000,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['src/components/**/*.vue', 'src/composables/**/*.ts', 'src/stores/**/*.ts'],
    },
  },
})
