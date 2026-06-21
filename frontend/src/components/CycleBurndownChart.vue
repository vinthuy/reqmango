<template>
  <div class="bg-white rounded-lg border border-gray-200 p-4">
    <h4 class="text-sm font-medium text-gray-700 mb-3">燃尽图</h4>

    <div v-if="!data" class="text-center py-8 text-gray-400 text-sm">
      暂无燃尽图数据（需要设置起止日期）
    </div>

    <div v-else>
      <div class="relative h-48">
        <svg viewBox="0 0 400 200" class="w-full h-full" preserveAspectRatio="none">
          <line v-for="i in 5" :key="'h'+i" x1="0" :y1="i*40" x2="400" :y2="i*40" stroke="#f3f4f6" stroke-width="1" />
          <polyline :points="idealLinePoints" fill="none" stroke="#9CA3AF" stroke-width="2" stroke-dasharray="6,3" />
          <polyline :points="actualLinePoints" fill="none" stroke="#3B82F6" stroke-width="2" />
        </svg>
      </div>

      <div class="flex items-center justify-center space-x-6 mt-3 text-xs text-gray-500">
        <span class="flex items-center"><span class="w-4 border-t-2 border-dashed border-gray-400 mr-1"></span>理想线</span>
        <span class="flex items-center"><span class="w-4 border-t-2 border-blue-500 mr-1"></span>实际完成</span>
        <span :class="data.is_on_track ? 'text-green-600 font-medium' : 'text-red-600 font-medium'">
          {{ data.is_on_track ? '进度正常' : '进度落后' }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { BurndownData } from '@/types/cycle'

const props = defineProps<{
  data: BurndownData | null
}>()

const idealLinePoints = computed(() => {
  if (!props.data) return ''
  const { total_issues, total_days } = props.data
  const xStep = 400 / Math.max(total_days, 1)
  const yScale = 200 / Math.max(total_issues, 1)
  let points = ''
  for (let d = 0; d <= total_days; d++) {
    const x = d * xStep
    const y = Math.max(0, (total_issues - (total_issues / total_days) * d)) * yScale
    points += `${x},${y} `
  }
  return points.trim()
})

const actualLinePoints = computed(() => {
  if (!props.data) return ''
  const { total_issues, total_days, days_elapsed, actual_completed } = props.data
  const xStep = 400 / Math.max(total_days, 1)
  const yScale = 200 / Math.max(total_issues, 1)
  const startX = 0
  const startY = total_issues * yScale
  const endX = Math.min(days_elapsed, total_days) * xStep
  const endY = Math.max(0, (total_issues - actual_completed)) * yScale
  return `${startX},${startY} ${endX},${endY}`
})
</script>
