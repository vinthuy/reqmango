<template>
  <div class="space-y-4">
    <!-- Title (only when no external header provides it) -->
    <div v-if="showTitle" class="card">
      <input
        :value="issue?.name"
        class="w-full text-xl font-semibold border-0 outline-none focus:ring-0"
        :placeholder="t('issue.titlePlaceholder')"
        @input="$emit('update:title', ($event.target as HTMLInputElement).value)"
      />
    </div>

    <!-- Description -->
    <div class="card">
      <div class="text-sm font-medium text-gray-500 mb-2">{{ t('issue.description') }}</div>
      <RichTextEditor
        :model-value="issue?.description_html || ''"
        :placeholder="t('issue.descriptionPlaceholder')"
        @update:model-value="$emit('update:description', $event)"
      />
    </div>

    <!-- Comments -->
    <div class="card">
      <div class="text-sm font-medium text-gray-500 mb-2">{{ t('issue.comments') }}</div>
      <CommentList :issue-id="issueId" :project-id="projectId" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import RichTextEditor from './RichTextEditor.vue'
import CommentList from './CommentList.vue'

const { t } = useI18n()

defineProps<{
  issueId: number
  issue: any
  workspaceId: number
  projectId: number
  issueTypeId: number
  members: any[]
  showTitle?: boolean
}>()

const emit = defineEmits<{
  'update:title': [value: string]
  'update:description': [value: string]
  'navigate': [issueId: number]
}>()

</script>
