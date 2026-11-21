<template>
  <div>
    <TimeTrackPanel :issue-id="issueId" />
    <RecurrenceConfig :issue-id="issueId" />
  </div>
</template>

<script setup lang="ts">
import TimeTrackPanel from './TimeTrackPanel.vue'
import RecurrenceConfig from './RecurrenceConfig.vue'

defineProps<{ issueId: number }>()
</script>
