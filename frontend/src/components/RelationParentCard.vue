<template>
  <div class="border border-blue-200 rounded-lg overflow-hidden bg-blue-50/30">
    <!-- Card header -->
    <div class="flex items-center justify-between px-3 py-2 bg-blue-50 border-b border-blue-100">
      <span class="text-xs font-semibold text-blue-700">PARENT</span>
      <button
        v-if="parent"
        data-test="change-parent"
        class="text-[10px] font-medium text-blue-600 hover:text-blue-800 hover:underline"
        @click="$emit('change')"
      >{{ t('issueKanban.change') }}</button>
    </div>

    <!-- Parent exists -->
    <div v-if="parent" class="flex items-center gap-2 px-3 py-2">
      <!-- Type badge -->
      <span
        class="px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap shrink-0"
        :style="{ backgroundColor: getIssueType(parent)?.color + '20', color: getIssueType(parent)?.color }"
      >{{ getIssueType(parent)?.name || '—' }}</span>
      <!-- ID -->
      <span class="text-[10px] text-gray-400 shrink-0 font-mono">#{{ parent.sequence_id }}</span>
      <!-- Title (clickable) -->
      <span
        class="relation-parent-title text-xs font-medium text-gray-800 flex-1 min-w-0 truncate cursor-pointer hover:text-indigo-600"
        data-test="navigate-parent"
        @click="$emit('navigate', parent.id)"
      >{{ parent.name }}</span>
      <!-- State -->
      <span class="flex items-center gap-1 text-[10px] shrink-0">
        <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: stateColor(parent.state_group) }"></span>
        {{ parent.state_name }}
      </span>
      <!-- Priority -->
      <span class="flex items-center gap-1 text-[10px] shrink-0">
        <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: priorityColor(parent.priority) }"></span>
        {{ t(`issue.priority${(parent.priority || 'none').charAt(0).toUpperCase() + (parent.priority || 'none').slice(1)}`) }}
      </span>
      <!-- Assignee -->
      <span class="text-[10px] text-gray-500 shrink-0 w-12 truncate">
        {{ parent.assignees?.[0]?.display_name || parent.assignees?.[0]?.username || '—' }}
      </span>
      <!-- Due Date -->
      <span class="text-[10px] text-gray-400 shrink-0 w-12 text-right">
        {{ parent.target_date ? formatDate(parent.target_date) : '—' }}
      </span>
      <!-- Remove -->
      <button
        data-test="remove-parent"
        class="text-gray-300 hover:text-red-500 shrink-0 text-sm leading-none"
        @click="$emit('remove')"
      >&times;</button>
    </div>

    <!-- Empty state -->
    <div v-else class="px-4 py-6 text-center text-xs text-gray-400">
      {{ t('issue.parentIssue') }}
      <button
        data-test="set-parent"
        class="block mx-auto mt-2 text-blue-500 hover:text-blue-700"
        @click="$emit('change')"
      >+ {{ t('issue.setParent') }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'
import { stateColor, priorityColor, formatDate } from '@/composables/useRelationHelpers'

const { t } = useI18n()

interface IssueType {
  id: number
  name: string
  color: string
}

interface ParentIssue {
  id: number
  sequence_id: number
  name: string
  state_name: string
  state_group: string
  priority: string
  assignees?: Array<{ id: number; display_name?: string; username?: string }>
  target_date?: string | null
  issue_type?: IssueType
}

const props = defineProps<{
  parent: ParentIssue | null
  issueTypes?: IssueType[]
}>()

defineEmits<{
  change: []
  remove: []
  navigate: [issueId: number]
}>()

function getIssueType(issue: ParentIssue) {
  if (!props.issueTypes) return null
  return props.issueTypes.find((t) => t.id === issue?.issue_type?.id) || issue?.issue_type || null
}

</script>
