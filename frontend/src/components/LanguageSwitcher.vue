<template>
  <button
    @click="toggle"
    class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 transition"
    :title="isZh ? 'Switch to English' : '切换到中文'"
  >
    {{ isZh ? '中' : 'EN' }}
  </button>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
const { isZh, setLocale } = useI18n()
function toggle() { setLocale(isZh.value ? 'en-US' : 'zh-CN') }
</script>
