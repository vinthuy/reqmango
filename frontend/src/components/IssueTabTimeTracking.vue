<template>
  <TimeTrackPanel :issue-id="issueId" />
</template>

<script setup lang="ts">
import TimeTrackPanel from './TimeTrackPanel.vue'

defineProps<{
  issueId: number
}>()
</script>
