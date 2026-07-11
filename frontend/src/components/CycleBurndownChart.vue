<template>
  <div class="bg-white rounded-lg border border-gray-200 p-4">
    <h4 class="text-sm font-medium text-gray-700 mb-3">{{ t('cycleBurndown.title') }}</h4>

    <div v-if="!data" class="text-center py-8 text-gray-400 text-sm">
      {{ t('cycleBurndown.noData') }}
    </div>

    <div v-else>
      <div class="relative h-48">
        <svg viewBox="0 0 400 200" class="w-full h-full" preserveAspectRatio="none">
          <line v-for="i in 5" :key="'h'+i" x1="0" :y1="i*40" x2="400" :y2="i*40" stroke="#f3f4f6" stroke-width="1" />
          <polyline :points="idealLinePoints" fill="none" stroke="#9CA3AF" stroke-width="2" stroke-dasharray="6,3" />
          <polyline :points="actualLinePoints" fill="none" stroke="#3B82F6" stroke-width="2" />
          <circle
            v-for="(point, index) in actualPoints"
            :key="'actual-'+index"
            :cx="point.x"
            :cy="point.y"
            r="3"
            fill="#3B82F6"
            v-show="index <= data.days_elapsed"
          />
        </svg>
      </div>

      <div class="flex items-center justify-center space-x-6 mt-3 text-xs text-gray-500">
        <span class="flex items-center"><span class="w-4 border-t-2 border-dashed border-gray-400 mr-1"></span>{{ t('cycleBurndown.ideal') }}</span>
        <span class="flex items-center"><span class="w-4 border-t-2 border-blue-500 mr-1"></span>{{ t('cycleBurndown.actual') }}</span>
        <span :class="data.is_on_track ? 'text-green-600 font-medium' : 'text-red-600 font-medium'">
          {{ data.is_on_track ? t('cycleBurndown.onTrack') : t('cycleBurndown.behind') }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { BurndownData } from '@/types/cycle'

const { t } = useI18n()

const props = defineProps<{
  data: BurndownData | null
}>()

const idealLinePoints = computed(() => {
  if (!props.data) return ''
  const { total_issues, total_days } = props.data
  const xStep = 400 / Math.max(total_days, 1)
  const yScale = 200 / Math.max(total_issues, 1)

  if (props.data.daily_points && props.data.daily_points.length > 0) {
    return props.data.daily_points
      .map((p) => {
        const x = p.day_index * xStep
        const y = p.ideal_remaining * yScale
        return `${x},${y}`
      })
      .join(' ')
  }

  let points = ''
  for (let d = 0; d <= total_days; d++) {
    const x = d * xStep
    const y = Math.max(0, (total_issues - (total_issues / total_days) * d)) * yScale
    points += `${x},${y} `
  }
  return points.trim()
})

const actualPoints = computed(() => {
  if (!props.data) return []
  const { total_issues, total_days } = props.data
  const xStep = 400 / Math.max(total_days, 1)
  const yScale = 200 / Math.max(total_issues, 1)

  if (props.data.daily_points && props.data.daily_points.length > 0) {
    return props.data.daily_points.map((p) => ({
      x: p.day_index * xStep,
      y: p.actual_remaining * yScale
    }))
  }

  const startX = 0
  const startY = total_issues * yScale
  const endX = Math.min(props.data.days_elapsed, total_days) * xStep
  const endY = Math.max(0, (total_issues - props.data.actual_completed)) * yScale
  return [{ x: startX, y: startY }, { x: endX, y: endY }]
})

const actualLinePoints = computed(() => {
  if (!props.data) return ''
  const { total_issues, total_days } = props.data
  const xStep = 400 / Math.max(total_days, 1)
  const yScale = 200 / Math.max(total_issues, 1)

  if (props.data.daily_points && props.data.daily_points.length > 0) {
    const elapsed = props.data.days_elapsed
    return props.data.daily_points
      .filter((_, index) => index <= elapsed)
      .map((p) => {
        const x = p.day_index * xStep
        const y = p.actual_remaining * yScale
        return `${x},${y}`
      })
      .join(' ')
  }

  const startX = 0
  const startY = total_issues * yScale
  const endX = Math.min(props.data.days_elapsed, total_days) * xStep
  const endY = Math.max(0, (total_issues - props.data.actual_completed)) * yScale
  return `${startX},${startY} ${endX},${endY}`
})
</script>
