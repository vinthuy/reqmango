/**
 * useI18n Composable 单元测试
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

// Mock the locale JSON files
vi.mock('@/locales/zh-CN.json', () => ({
  default: {
    common: { save: '保存', cancel: '取消', search: '搜索' },
    filter: { title: '筛选', clear: '清除', hasFilters: '已选 {count} 项' },
    nested: { deep: { key: '深层嵌套值' } },
  },
}))

vi.mock('@/locales/en-US.json', () => ({
  default: {
    common: { save: 'Save', cancel: 'Cancel', search: 'Search' },
    filter: { title: 'Filter', clear: 'Clear', hasFilters: '{count} selected' },
    nested: { deep: { key: 'Deep nested value' } },
  },
}))

import { useI18n } from './useI18n'

describe('useI18n', () => {
  beforeEach(() => {
    localStorage.clear()
    // Resetting locale requires module re-import, so we use setLocale
  })

  it('should translate simple key in zh-CN (default)', () => {
    const { t, locale } = useI18n()
    expect(locale.value).toBe('zh-CN')
    expect(t('common.save')).toBe('保存')
    expect(t('common.cancel')).toBe('取消')
    expect(t('common.search')).toBe('搜索')
  })

  it('should translate nested keys', () => {
    const { t } = useI18n()
    expect(t('nested.deep.key')).toBe('深层嵌套值')
  })

  it('should interpolate placeholders', () => {
    const { t } = useI18n()
    expect(t('filter.hasFilters', { count: 5 })).toBe('已选 5 项')
  })

  it('should return fallback for missing keys', () => {
    const { t } = useI18n()
    expect(t('nonexistent.key', 'default text')).toBe('default text')
  })

  it('should return key itself if no fallback provided', () => {
    const { t } = useI18n()
    expect(t('nonexistent.key')).toBe('nonexistent.key')
  })

  it('should switch to English via setLocale', () => {
    const { t, setLocale, locale } = useI18n()
    setLocale('en-US')

    expect(locale.value).toBe('en-US')
    expect(t('common.save')).toBe('Save')
    expect(t('common.cancel')).toBe('Cancel')
    expect(t('nested.deep.key')).toBe('Deep nested value')
    expect(t('filter.hasFilters', { count: 3 })).toBe('3 selected')
  })

  it('should persist locale to localStorage', () => {
    const { setLocale } = useI18n()
    setLocale('en-US')
    expect(localStorage.getItem('locale')).toBe('en-US')

    setLocale('zh-CN')
    expect(localStorage.getItem('locale')).toBe('zh-CN')
  })

  it('isZh should reflect zh-CN status', () => {
    const { isZh, setLocale } = useI18n()
    expect(isZh.value).toBe(true)

    setLocale('en-US')
    expect(isZh.value).toBe(false)
  })

  it('should handle options as string (overload for fallback)', () => {
    const { t } = useI18n()
    // Options param can be string fallback
    expect(t('missing.key', 'fallback text')).toBe('fallback text')
  })

  it('should handle partial key paths (returns key)', () => {
    const { t } = useI18n()
    // 'common' is an object, not a string
    expect(t('common')).toBe('common')
  })
})
