import { ref, computed } from 'vue'
import zhCN from '@/locales/zh-CN.json'
import enUS from '@/locales/en-US.json'

type DeepRecord = { [key: string]: string | DeepRecord }
type Locale = 'zh-CN' | 'en-US'

const messages: Record<Locale, DeepRecord> = { 'zh-CN': zhCN as DeepRecord, 'en-US': enUS as DeepRecord }

// Module-scoped reactive locale — shared across all components
const currentLocale = ref<Locale>((localStorage.getItem('locale') as Locale) || 'zh-CN')

// Map locale to html lang
const langMap: Record<Locale, string> = { 'zh-CN': 'zh-CN', 'en-US': 'en' }

export function useI18n() {
  function t(key: string, options?: Record<string, string | number> | string, fallback?: string): string {
    // Handle overload: if options is a string, treat as fallback
    if (typeof options === 'string') {
      fallback = options
      options = undefined
    }
    const keys = key.split('.')
    let result: any = messages[currentLocale.value]
    for (const k of keys) {
      if (result && typeof result === 'object' && k in result) {
        result = result[k]
      } else {
        return fallback ?? key
      }
    }
    let text = typeof result === 'string' ? result : (fallback ?? key)
    // Replace {key} placeholders with values from options
    if (options) {
      for (const [k, v] of Object.entries(options)) {
        text = text.replace(new RegExp(`\\{${k}\\}`, 'g'), String(v))
      }
    }
    return text
  }

  function setLocale(locale: Locale) {
    currentLocale.value = locale
    localStorage.setItem('locale', locale)
    document.documentElement.lang = langMap[locale]
  }

  const locale = computed(() => currentLocale.value)
  const isZh = computed(() => currentLocale.value === 'zh-CN')

  // Initialize html lang on first use
  if (!document.documentElement.lang || document.documentElement.lang === 'zh-CN') {
    document.documentElement.lang = langMap[currentLocale.value]
  }

  return { t, locale, isZh, setLocale }
}
