<template>
  <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-gray-900">{{ progress?.total_issues ?? 0 }}</div>
      <div class="text-xs text-gray-500 mt-1">{{ t('cycleProgress.totalIssues') }}</div>
    </div>
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-green-600">{{ progress?.completed_issues ?? 0 }}</div>
      <div class="text-xs text-gray-500 mt-1">{{ t('cycleProgress.completed') }}</div>
    </div>
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-indigo-600">{{ Math.round(progress?.progress ?? 0) }}%</div>
      <div class="text-xs text-gray-500 mt-1">{{ t('cycleProgress.progress') }}</div>
    </div>
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-orange-600">{{ (progress?.total_issues ?? 0) - (progress?.completed_issues ?? 0) }}</div>
      <div class="text-xs text-gray-500 mt-1">{{ t('cycleProgress.remaining') }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import type { CycleProgress } from '@/types/cycle'

const { t } = useI18n()

defineProps<{
  progress: CycleProgress | null
}>()
</script>
