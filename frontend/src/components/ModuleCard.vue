<template>
  <div
    class="module-card bg-white border border-gray-200 rounded-lg p-4 hover:border-indigo-300 hover:shadow-sm cursor-pointer transition-all"
    @click="$emit('click')"
  >
    <!-- 头部：标题 -->
    <div class="flex items-start justify-between mb-3">
      <div class="flex items-center space-x-2">
        <svg class="w-5 h-5 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <h3 class="text-base font-semibold text-gray-900">{{ module.name }}</h3>
      </div>
    </div>

    <!-- 描述 -->
    <p v-if="module.description" class="text-sm text-gray-500 mb-3 line-clamp-2">
      {{ module.description }}
    </p>

    <!-- 进度条 -->
    <div class="mb-3">
      <div class="flex items-center justify-between text-xs text-gray-500 mb-1">
        <span>进度</span>
        <span>{{ module.progress }}%</span>
      </div>
      <div class="w-full bg-gray-200 rounded-full h-2">
        <div
          class="h-2 rounded-full transition-all duration-300"
          :class="getProgressClass(module.progress)"
          :style="{ width: module.progress + '%' }"
        ></div>
      </div>
    </div>

    <!-- 统计信息 -->
    <div class="grid grid-cols-2 gap-3 mb-3">
      <div class="text-center p-2 bg-gray-50 rounded">
        <div class="text-lg font-semibold text-gray-900">{{ module.total_issues }}</div>
        <div class="text-xs text-gray-500">总工作项</div>
      </div>
      <div class="text-center p-2 bg-gray-50 rounded">
        <div class="text-lg font-semibold text-green-600">{{ module.completed_issues }}</div>
        <div class="text-xs text-gray-500">已完成</div>
      </div>
    </div>

    <!-- 目标日期 -->
    <div v-if="module.target_date" class="text-xs text-gray-500 mb-3">
      <div class="flex items-center" :class="{ 'text-red-500': isOverdue }">
        <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        目标: {{ formatDate(module.target_date) }}
        <span v-if="isOverdue" class="ml-1 text-red-500">(已过期)</span>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="flex items-center justify-end pt-3 border-t border-gray-100">
      <div class="relative" @click.stop>
        <button
          @click="showMenu = !showMenu"
          class="p-1 text-gray-400 hover:text-gray-600 rounded"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
          </svg>
        </button>

        <div
          v-if="showMenu"
          class="absolute right-0 mt-1 w-28 bg-white border border-gray-200 rounded-md shadow-lg z-10"
        >
          <button
            @click="$emit('click'); showMenu = false"
            class="w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-50"
          >
            查看详情
          </button>
          <button
            @click="$emit('delete', module); showMenu = false"
            class="w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50"
          >
            删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { ModuleResponse } from '@/types/module'

// Props
const props = defineProps<{
  module: ModuleResponse
}>()

// Emits
defineEmits<{
  (e: 'click'): void
  (e: 'delete', module: ModuleResponse): void
}>()

// State
const showMenu = ref(false)

// Progress bar class
function getProgressClass(progress: number): string {
  if (progress >= 100) return 'bg-green-500'
  if (progress >= 75) return 'bg-blue-500'
  if (progress >= 50) return 'bg-yellow-500'
  if (progress >= 25) return 'bg-orange-500'
  return 'bg-red-500'
}

// Format date
function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()}`
}

// Check if overdue
const isOverdue = computed(() => {
  if (!props.module?.target_date) return false
  return new Date(props.module.target_date) < new Date()
})
</script>

<style scoped>
.module-card:hover {
  transform: translateY(-2px);
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>