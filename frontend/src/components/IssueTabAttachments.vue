<template>
  <AttachmentManager :issue-id="issueId" :project-id="projectId" />
</template>

<script setup lang="ts">
import AttachmentManager from './AttachmentManager.vue'

defineProps<{
  issueId: number
  projectId: number
}>()
</script>
