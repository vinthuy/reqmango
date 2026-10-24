/**
 * Vitest global setup
 * Mocks browser APIs for consistent component testing
 */
import { vi } from 'vitest'

// jsdom already defines window.confirm/window.alert as no-ops.
// Replace them with vi spies so assertions work after vi.clearAllMocks().
// Use vi.stubGlobal which creates proper spy tracking.
vi.stubGlobal('confirm', vi.fn((_msg?: string) => true))
vi.stubGlobal('alert', vi.fn())

// Stub matchMedia for responsive tests
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Pin the browser language. `useI18n` negotiates the locale once, at import
// time, from localStorage and `navigator.languages`; jsdom reports en-US, which
// would flip every component assertion that expects the Chinese default. The
// negotiation rules themselves are covered by `src/utils/locale.test.ts`, which
// stubs these fields per case.
Object.defineProperty(window.navigator, 'language', { value: 'zh-CN', configurable: true })
Object.defineProperty(window.navigator, 'languages', { value: ['zh-CN'], configurable: true })

export {}  
