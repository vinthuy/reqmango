import { config } from '@vue/test-utils'
import { createPinia } from 'pinia'

// Global test setup for Vue component tests
config.global.plugins = [createPinia()]

// Mock browser APIs not available in jsdom
if (typeof global.IntersectionObserver === 'undefined') {
  global.IntersectionObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as unknown as typeof IntersectionObserver
}

if (typeof global.ResizeObserver === 'undefined') {
  global.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as unknown as typeof ResizeObserver
}
