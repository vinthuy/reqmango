<template>
  <span :class="badgeClass" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium">
    {{ label }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ status: string }>()

const badgeClass = computed(() => {
  switch (props.status) {
    case 'running': return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
    case 'completed': return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
    case 'failed': return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
    case 'escalated': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
    case 'stopped': return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
    default: return 'bg-gray-100 text-gray-800'
  }
})

const label = computed(() => {
  switch (props.status) {
    case 'running': return 'Running'
    case 'completed': return 'Completed'
    case 'failed': return 'Failed'
    case 'escalated': return 'Escalated'
    case 'stopped': return 'Stopped'
    default: return props.status
  }
})
</script>
