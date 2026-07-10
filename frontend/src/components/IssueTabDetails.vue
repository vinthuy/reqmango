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

    <!-- Sub-issues Panel -->
    <div class="card">
      <SubIssuesPanel
        :sub-issues="subIssues || []"
        @navigate="handleNavigate"
        @add-sub-issue="handleAddSubIssue"
        @remove-sub-issue="handleRemoveSubIssue"
        @reorder="handleReorder"
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
import SubIssuesPanel from './SubIssuesPanel.vue'
import type { IssueResponse } from '@/types/issue'

const { t } = useI18n()

defineProps<{
  issueId: number
  issue: any
  workspaceId: number
  projectId: number
  issueTypeId: number
  members: any[]
  showTitle?: boolean
  subIssues?: IssueResponse[]
}>()

const emit = defineEmits<{
  'update:title': [value: string]
  'update:description': [value: string]
  'navigate': [issueId: number]
  'add-sub-issue': []
  'remove-sub-issue': [issueId: number]
  'reorder-sub-issues': [issueIds: number[]]
}>()

function handleNavigate(issueId: number) {
  emit('navigate', issueId)
}

function handleAddSubIssue() {
  emit('add-sub-issue')
}

function handleRemoveSubIssue(issueId: number) {
  emit('remove-sub-issue', issueId)
}

function handleReorder(issueIds: number[]) {
  emit('reorder-sub-issues', issueIds)
}
</script>
