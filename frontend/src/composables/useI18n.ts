import { ref, computed } from 'vue'
import zhCN from '@/locales/zh-CN.json'
import enUS from '@/locales/en-US.json'

type Messages = typeof zhCN
type Locale = 'zh-CN' | 'en-US'

const messages: Record<Locale, Messages> = { 'zh-CN': zhCN, 'en-US': enUS }
const currentLocale = ref<Locale>((localStorage.getItem('locale') as Locale) || 'zh-CN')

export function useI18n() {
  function t(path: string): string {
    const keys = path.split('.')
    let result: any = messages[currentLocale.value]
    for (const k of keys) {
      if (result && typeof result === 'object') result = result[k]
      else return path
    }
    return typeof result === 'string' ? result : path
  }

  function setLocale(locale: Locale) {
    currentLocale.value = locale
    localStorage.setItem('locale', locale)
    document.documentElement.lang = locale
  }

  const locale = computed(() => currentLocale.value)
  const isZh = computed(() => currentLocale.value === 'zh-CN')

  return { t, locale, isZh, setLocale }
}
