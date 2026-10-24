import { describe, it, expect } from 'vitest'
import {
  DEFAULT_LOCALE,
  browserLanguages,
  matchLocale,
  resolveInitialLocale,
} from './locale'

describe('matchLocale', () => {
  it('maps the primary subtag, not the exact tag', () => {
    expect(matchLocale('zh')).toBe('zh-CN')
    expect(matchLocale('zh-CN')).toBe('zh-CN')
    expect(matchLocale('zh-Hans-CN')).toBe('zh-CN')
    expect(matchLocale('en')).toBe('en-US')
    expect(matchLocale('en-GB')).toBe('en-US')
  })

  it('keeps regional Chinese on Chinese instead of falling back to English', () => {
    expect(matchLocale('zh-TW')).toBe('zh-CN')
    expect(matchLocale('zh_HK')).toBe('zh-CN')
  })

  it('ignores case and surrounding whitespace', () => {
    expect(matchLocale('  EN-us ')).toBe('en-US')
    expect(matchLocale('ZH-cn')).toBe('zh-CN')
  })

  it('returns null for unsupported or empty input', () => {
    expect(matchLocale('fr-FR')).toBeNull()
    expect(matchLocale('')).toBeNull()
    expect(matchLocale(null)).toBeNull()
    expect(matchLocale(undefined)).toBeNull()
  })
})

describe('resolveInitialLocale', () => {
  it('prefers an explicit stored choice over the browser', () => {
    expect(resolveInitialLocale('en-US', ['zh-CN'])).toBe('en-US')
    expect(resolveInitialLocale('zh-CN', ['en-US'])).toBe('zh-CN')
  })

  it('ignores a stored value that is no longer supported', () => {
    expect(resolveInitialLocale('fr-FR', ['en-US'])).toBe('en-US')
  })

  it('follows the browser when nothing is stored', () => {
    expect(resolveInitialLocale(null, ['en-US'])).toBe('en-US')
    expect(resolveInitialLocale(null, ['zh-CN'])).toBe('zh-CN')
  })

  it('takes the first supported language from the ordered list', () => {
    expect(resolveInitialLocale(null, ['fr-FR', 'de-DE', 'en-GB'])).toBe('en-US')
    expect(resolveInitialLocale(null, ['fr-FR', 'zh-TW'])).toBe('zh-CN')
  })

  it('falls back to the default when nothing matches', () => {
    expect(resolveInitialLocale(null, [])).toBe(DEFAULT_LOCALE)
    expect(resolveInitialLocale(null, ['fr-FR', 'de-DE'])).toBe(DEFAULT_LOCALE)
    expect(resolveInitialLocale(undefined, [])).toBe(DEFAULT_LOCALE)
  })
})

describe('browserLanguages', () => {
  it('reads the ordered navigator.languages list', () => {
    const original = navigator.languages
    Object.defineProperty(navigator, 'languages', { value: ['fr-FR', 'en-US'], configurable: true })
    try {
      expect(browserLanguages()).toEqual(['fr-FR', 'en-US'])
    } finally {
      Object.defineProperty(navigator, 'languages', { value: original, configurable: true })
    }
  })

  it('falls back to navigator.language when the list is empty', () => {
    const originalList = navigator.languages
    const originalLang = navigator.language
    Object.defineProperty(navigator, 'languages', { value: [], configurable: true })
    Object.defineProperty(navigator, 'language', { value: 'en-GB', configurable: true })
    try {
      expect(browserLanguages()).toEqual(['en-GB'])
    } finally {
      Object.defineProperty(navigator, 'languages', { value: originalList, configurable: true })
      Object.defineProperty(navigator, 'language', { value: originalLang, configurable: true })
    }
  })
})
