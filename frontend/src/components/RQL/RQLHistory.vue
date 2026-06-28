<!-- frontend/src/components/RQL/RQLHistory.vue -->

<template>
  <div class="absolute left-0 top-full mt-1 w-80 bg-white border border-gray-200 rounded-lg shadow-lg z-50">
    <div class="px-3 py-2 border-b border-gray-100 flex items-center justify-between">
      <span class="text-sm font-medium text-gray-700">查询历史</span>
      <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
    <div class="max-h-64 overflow-y-auto">
      <div v-if="history.length === 0" class="px-3 py-4 text-sm text-gray-500 text-center">
        暂无历史记录
      </div>
      <div
        v-for="item in history"
        :key="item.id"
        @click="$emit('select', item)"
        class="px-3 py-2 hover:bg-gray-50 cursor-pointer border-b border-gray-50 last:border-b-0"
      >
        <div class="text-sm text-gray-800 truncate">{{ item.rql }}</div>
        <div class="text-xs text-gray-400 mt-0.5">{{ formatTime(item.timestamp) }}</div>
      </div>
    </div>
    <div v-if="history.length > 0" class="px-3 py-2 border-t border-gray-100">
      <button @click="clearHistory" class="text-xs text-red-500 hover:text-red-700">
        清除历史记录
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { HISTORY_KEY } from '@/composables/useRQL'
const history = ref<any[]>([])

onMounted(() => {
  try {
    const data = localStorage.getItem(HISTORY_KEY)
    history.value = data ? JSON.parse(data) : []
  } catch {
    history.value = []
  }
})

const clearHistory = () => {
  localStorage.removeItem(HISTORY_KEY)
  history.value = []
}

const formatTime = (timestamp: number): string => {
  const diff = Date.now() - timestamp
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return `${Math.floor(diff / 86400000)} 天前`
}
</script>