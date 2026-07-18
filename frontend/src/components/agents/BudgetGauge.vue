<template>
  <div class="space-y-2">
    <div v-if="maxTokens > 0" class="flex items-center gap-2 text-xs">
      <span class="text-gray-500 w-16">Tokens</span>
      <div class="flex-1 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
        <div
          class="h-full rounded-full transition-all"
          :class="usedTokens / maxTokens > 0.9 ? 'bg-red-500' : 'bg-indigo-500'"
          :style="{ width: Math.min(100, (usedTokens / maxTokens) * 100) + '%' }"
        ></div>
      </div>
      <span class="text-gray-500 w-24 text-right">{{ usedTokens.toLocaleString() }} / {{ maxTokens.toLocaleString() }}</span>
    </div>
    <div v-if="maxIterations > 0" class="flex items-center gap-2 text-xs">
      <span class="text-gray-500 w-16">Iterations</span>
      <div class="flex-1 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
        <div
          class="h-full rounded-full transition-all bg-purple-500"
          :style="{ width: Math.min(100, (currentIteration / maxIterations) * 100) + '%' }"
        ></div>
      </div>
      <span class="text-gray-500 w-24 text-right">{{ currentIteration }} / {{ maxIterations }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  maxTokens: number
  usedTokens: number
  maxIterations: number
  currentIteration: number
}>()
</script>
