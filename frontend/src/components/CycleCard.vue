<template>
  <div
    class="cycle-card bg-white border border-gray-200 rounded-lg p-4 hover:border-indigo-300 hover:shadow-sm cursor-pointer transition-all"
    @click="$emit('click')"
  >
    <!-- 头部：标题和状态 -->
    <div class="flex items-start justify-between mb-3">
      <div>
        <h3 class="text-base font-semibold text-gray-900">{{ cycle.name }}</h3>
        <p v-if="cycle.description" class="text-sm text-gray-500 mt-0.5 line-clamp-2">
          {{ cycle.description }}
        </p>
      </div>
      <span :class="getStatusClass(cycle.status)">
        {{ getStatusText(cycle.status) }}
      </span>
    </div>

    <!-- 进度条 -->
    <div class="mb-3">
      <div class="flex items-center justify-between text-xs text-gray-500 mb-1">
        <span>{{ t('cycleCard.progress') }}</span>
        <span>{{ cycle.progress }}%</span>
      </div>
      <div class="w-full bg-gray-200 rounded-full h-2">
        <div
          class="h-2 rounded-full transition-all duration-300"
          :class="getProgressClass(cycle.progress)"
          :style="{ width: cycle.progress + '%' }"
        ></div>
      </div>
    </div>

    <!-- 统计信息 -->
    <div class="grid grid-cols-2 gap-3 mb-3">
      <div class="text-center p-2 bg-gray-50 rounded">
        <div class="text-lg font-semibold text-gray-900">{{ cycle.total_issues }}</div>
        <div class="text-xs text-gray-500">{{ t('cycleCard.totalIssues') }}</div>
      </div>
      <div class="text-center p-2 bg-gray-50 rounded">
        <div class="text-lg font-semibold text-green-600">{{ cycle.completed_issues }}</div>
        <div class="text-xs text-gray-500">{{ t('cycleCard.completed') }}</div>
      </div>
    </div>

    <!-- 日期信息 -->
    <div class="text-xs text-gray-500 space-y-1 mb-3">
      <div v-if="cycle.start_date" class="flex items-center">
        <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        {{ t('cycleCard.startDate') }}: {{ formatDate(cycle.start_date) }}
      </div>
      <div v-if="cycle.end_date" class="flex items-center" :class="{ 'text-red-500': isOverdue }">
        <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        {{ t('cycleCard.endDate') }}: {{ formatDate(cycle.end_date) }}
        <span v-if="isOverdue" class="ml-1 text-red-500">{{ t('cycleCard.expired') }}</span>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="flex items-center justify-between pt-3 border-t border-gray-100">
      <div class="flex items-center space-x-2">
        <button
          v-if="cycle.status === 'upcoming'"
          @click.stop="$emit('start', cycle)"
          class="px-3 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700"
        >
          {{ t('cycleCard.start') }}
        </button>
        <button
          v-if="cycle.status === 'active'"
          @click.stop="$emit('end', cycle)"
          class="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          {{ t('cycleCard.end') }}
        </button>
        <button
          v-if="cycle.status !== 'completed' && cycle.status !== 'cancelled'"
          @click.stop="$emit('cancel', cycle)"
          class="px-3 py-1 text-xs text-gray-600 border border-gray-300 rounded hover:bg-gray-50"
        >
          {{ t('cycleCard.cancel') }}
        </button>
      </div>

      <!-- 更多操作 -->
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
            {{ t('cycleCard.viewDetails') }}
          </button>
          <button
            @click="$emit('delete', cycle); showMenu = false"
            class="w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50"
          >
            {{ t('cycleCard.delete') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { CycleResponse, CycleStatus } from '@/types/cycle'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

// Props
const props = defineProps<{
  cycle: CycleResponse
}>()

// Emits
defineEmits<{
  (e: 'click'): void
  (e: 'start', cycle: CycleResponse): void
  (e: 'end', cycle: CycleResponse): void
  (e: 'cancel', cycle: CycleResponse): void
  (e: 'delete', cycle: CycleResponse): void
}>()

// State
const showMenu = ref(false)

// Status classes
function getStatusClass(status: CycleStatus): string {
  const classes: Record<CycleStatus, string> = {
    upcoming: 'px-2 py-0.5 text-xs rounded bg-blue-100 text-blue-700',
    active: 'px-2 py-0.5 text-xs rounded bg-green-100 text-green-700',
    completed: 'px-2 py-0.5 text-xs rounded bg-gray-100 text-gray-600',
    cancelled: 'px-2 py-0.5 text-xs rounded bg-red-100 text-red-700'
  }
  return classes[status] || classes.upcoming
}

// Status text
function getStatusText(status: CycleStatus): string {
  const texts: Record<CycleStatus, string> = {
    upcoming: t('cycleCard.upcoming'),
    active: t('cycleCard.active'),
    completed: t('cycleCard.completed_status'),
    cancelled: t('cycleCard.cancelled')
  }
  return texts[status] || t('cycleCard.unknown')
}

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
  return `${date.getMonth() + 1}/${date.getDate()}`
}

// Check if overdue
const isOverdue = computed(() => {
  if (!props.cycle?.end_date) return false
  return new Date(props.cycle.end_date) < new Date()
})
</script>

<style scoped>
.cycle-card:hover {
  transform: translateY(-2px);
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>