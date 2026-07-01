<template>
  <button
    @click="toggle"
    class="px-2 py-1 text-[11px] font-medium rounded-md text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
    :title="isZh ? t('language.switchToZh') : t('language.switchToEn')"
  >
    {{ isZh ? '中' : 'EN' }}
  </button>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
const { t, isZh, setLocale } = useI18n()
function toggle() { setLocale(isZh.value ? 'en-US' : 'zh-CN') }
</script>
