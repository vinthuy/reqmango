<template>
  <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-gray-900">{{ progress?.total_issues ?? 0 }}</div>
      <div class="text-xs text-gray-500 mt-1">总工作项</div>
    </div>
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-green-600">{{ progress?.completed_issues ?? 0 }}</div>
      <div class="text-xs text-gray-500 mt-1">已完成</div>
    </div>
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-indigo-600">{{ Math.round(progress?.progress ?? 0) }}%</div>
      <div class="text-xs text-gray-500 mt-1">进度</div>
    </div>
    <div class="bg-white rounded-lg border border-gray-200 p-4 text-center">
      <div class="text-2xl font-bold text-orange-600">{{ (progress?.total_issues ?? 0) - (progress?.completed_issues ?? 0) }}</div>
      <div class="text-xs text-gray-500 mt-1">待完成</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CycleProgress } from '@/types/cycle'

defineProps<{
  progress: CycleProgress | null
}>()
</script>
