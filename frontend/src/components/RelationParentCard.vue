<template>
  <div class="border border-blue-200 rounded-lg overflow-hidden bg-blue-50/30">
    <!-- Card header with actions -->
    <div class="flex items-center justify-between px-3 py-2 bg-blue-50 border-b border-blue-100">
      <div class="flex items-center gap-2">
        <span class="text-xs font-semibold text-blue-700">PARENT</span>
        <span
          v-if="parent"
          class="px-1.5 py-0.5 rounded-full text-[10px] font-medium bg-blue-100 text-blue-700"
        >1</span>
      </div>
      <div class="flex items-center gap-1">
        <button
          data-test="change-parent"
          class="text-[10px] font-medium text-blue-600 hover:text-blue-800 hover:underline"
          @click="$emit('change')"
        >{{ parent ? t('issueKanban.change') : '+ ' + t('issue.setParent') }}</button>
        <button
          v-if="parent"
          data-test="remove-parent"
          class="text-gray-300 hover:text-red-500 shrink-0 text-sm leading-none ml-1"
          @click="$emit('remove')"
        >&times;</button>
      </div>
    </div>

    <!-- Parent list table -->
    <div v-if="parent" class="overflow-x-auto">
      <table class="w-full text-[10px]">
        <thead>
          <tr class="border-b border-blue-100 text-gray-500">
            <th class="px-3 py-1.5 text-left">{{ t('issue.type') }}</th>
            <th class="px-2 py-1.5 text-left">ID</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.title') }}</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.status') }}</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.priority') }}</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.assignee') }}</th>
            <th class="px-2 py-1.5 text-right">{{ t('issue.targetDate') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr class="hover:bg-blue-50/50 transition-colors">
            <!-- Type badge -->
            <td class="px-3 py-2">
              <span
                class="px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap"
                :style="{ backgroundColor: getIssueType(parent)?.color + '20', color: getIssueType(parent)?.color }"
              >{{ getIssueType(parent)?.name || '—' }}</span>
            </td>
            <!-- ID -->
            <td class="px-2 py-2 font-mono text-gray-400">#{{ parent.sequence_id }}</td>
            <!-- Title (clickable) -->
            <td class="px-2 py-2">
              <span
                class="relation-parent-title text-xs font-medium text-gray-800 cursor-pointer hover:text-indigo-600"
                @click="$emit('navigate', parent.id)"
              >{{ parent.name }}</span>
            </td>
            <!-- State -->
            <td class="px-2 py-2">
              <span class="flex items-center gap-1">
                <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: stateColor(parent.state_group) }"></span>
                {{ parent.state_name }}
              </span>
            </td>
            <!-- Priority -->
            <td class="px-2 py-2">
              <span class="flex items-center gap-1">
                <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: priorityColor(parent.priority) }"></span>
                {{ t(`issue.priority${(parent.priority || 'none').charAt(0).toUpperCase() + (parent.priority || 'none').slice(1)}`) }}
              </span>
            </td>
            <!-- Assignee -->
            <td class="px-2 py-2 text-gray-500">
              {{ parent.assignees?.[0]?.display_name || parent.assignees?.[0]?.username || '—' }}
            </td>
            <!-- Due date -->
            <td class="px-2 py-2 text-gray-400 text-right">
              {{ parent.target_date ? formatDate(parent.target_date) : '—' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Empty state -->
    <div v-else class="px-4 py-6 text-center text-xs text-gray-400">
      {{ t('issue.parentIssue') }}
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
