<template>
  <div class="sub-issues-panel">
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center space-x-2">
        <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
        </svg>
        <span class="text-sm font-semibold text-gray-700">{{ t('issue.subIssues') }}</span>
        <span v-if="subIssues.length > 0" class="text-xs bg-gray-100 text-gray-500 px-1.5 py-0.5 rounded-full">
          {{ completedCount }}/{{ subIssues.length }}
        </span>
      </div>
      <button
        @click="$emit('add-sub-issue')"
        class="text-xs font-medium text-indigo-600 hover:text-indigo-700 hover:bg-indigo-50 px-2 py-1 rounded transition-colors"
      >
        + {{ t('issue.addSubIssue') }}
      </button>
    </div>

    <div v-if="subIssues.length === 0" class="bg-white border border-gray-200 rounded-lg px-4 py-8 text-center text-sm text-gray-400">
      {{ t('issue.noSubIssues') }}
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="(issue, index) in subIssues"
        :key="issue.id"
        class="bg-white border border-gray-200 rounded-lg overflow-hidden hover:border-gray-300 transition-colors"
      >
        <div
          class="flex items-center gap-3 px-4 py-3 cursor-pointer select-none"
          @click="$emit('navigate', issue.id)"
        >
          <button
            class="p-1 text-gray-400 hover:text-gray-600"
            @click.stop="toggleExpand(issue.id)"
          >
            <svg
              class="w-4 h-4 transition-transform duration-200"
              :class="{ 'rotate-90': expandedIds.includes(issue.id) }"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>

          <div class="w-1.5 h-1.5 rounded-full flex-shrink-0" :style="{ backgroundColor: getStateColor(issue.state_group || '') }"></div>

          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="font-mono text-xs text-gray-400">#{{ issue.sequence_id }}</span>
              <span
                v-if="issue.issue_type"
                class="px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap"
                :style="{ backgroundColor: (issue.issue_type.color || '#e5e7eb') + '20', color: issue.issue_type.color || '#6b7280' }"
              >{{ issue.issue_type.name }}</span>
              <span class="text-sm font-medium text-gray-800 truncate">{{ issue.name }}</span>
            </div>
            <div class="flex items-center gap-3 mt-1">
              <span v-if="issue.priority && issue.priority !== 'none'" class="flex items-center gap-1">
                <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: getPriorityColor(issue.priority) }"></span>
                <span class="text-xs text-gray-500">{{ getPriorityName(issue.priority) }}</span>
              </span>
              <span v-if="issue.state_name" class="text-xs text-gray-500">{{ issue.state_name }}</span>
              <span v-if="issue.assignees && issue.assignees.length > 0" class="text-xs text-gray-500">
                {{ issue.assignees[0].display_name }}
              </span>
            </div>
          </div>

          <div class="flex items-center gap-1">
            <button
              v-if="index > 0"
              class="p-1.5 text-gray-400 hover:text-gray-600"
              @click.stop="moveUp(index)"
              :title="t('issue.moveUp')"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
              </svg>
            </button>
            <button
              v-if="index < subIssues.length - 1"
              class="p-1.5 text-gray-400 hover:text-gray-600"
              @click.stop="moveDown(index)"
              :title="t('issue.moveDown')"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <button
              class="p-1.5 text-gray-400 hover:text-red-500"
              @click.stop="$emit('remove-sub-issue', issue.id)"
              :title="t('issue.remove')"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
              </svg>
            </button>
          </div>
        </div>

        <div
          v-if="expandedIds.includes(issue.id)"
          class="border-t border-gray-100 px-4 py-3 bg-gray-50/50"
        >
          <div v-if="issue.description_html" class="text-xs text-gray-600 mb-3" v-html="issue.description_html"></div>
          <div class="flex items-center gap-4 text-xs text-gray-500">
            <div v-if="issue.target_date" class="flex items-center gap-1">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              {{ formatDate(issue.target_date) }}
            </div>
            <div v-if="issue.start_date" class="flex items-center gap-1">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              {{ formatDate(issue.start_date) }}
            </div>
            <div v-if="issue.sub_issues_count" class="flex items-center gap-1">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
              </svg>
              {{ issue.sub_issues_count }} {{ t('issue.subItems') }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { IssueResponse } from '@/types/issue'
import { getPriorityName, getPriorityColor } from '@/types/issue'

const { t } = useI18n()

const props = defineProps<{
  subIssues: IssueResponse[]
}>()

defineEmits<{
  (e: 'navigate', issueId: number): void
  (e: 'add-sub-issue'): void
  (e: 'remove-sub-issue', issueId: number): void
  (e: 'reorder', issueIds: number[]): void
}>()

const emit = defineEmits<{
  (e: 'navigate', issueId: number): void
  (e: 'add-sub-issue'): void
  (e: 'remove-sub-issue', issueId: number): void
  (e: 'reorder', issueIds: number[]): void
}>()

const expandedIds = ref<number[]>([])

const completedCount = computed(() => {
  return props.subIssues.filter(issue => issue.state_group === 'done' || issue.state_group === 'cancelled').length
})

function toggleExpand(issueId: number) {
  const index = expandedIds.value.indexOf(issueId)
  if (index > -1) {
    expandedIds.value.splice(index, 1)
  } else {
    expandedIds.value.push(issueId)
  }
}

function getStateColor(stateGroup: string): string {
  const colors: Record<string, string> = {
    backlog: '#9ca3af',
    todo: '#6b7280',
    in_progress: '#3b82f6',
    done: '#10b981',
    cancelled: '#ef4444'
  }
  return colors[stateGroup] || '#9ca3af'
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

function moveUp(index: number) {
  if (index <= 0) return
  const newList = [...props.subIssues]
  const temp = newList[index]
  newList[index] = newList[index - 1]
  newList[index - 1] = temp
  emit('reorder', newList.map(i => i.id))
}

function moveDown(index: number) {
  if (index >= props.subIssues.length - 1) return
  const newList = [...props.subIssues]
  const temp = newList[index]
  newList[index] = newList[index + 1]
  newList[index + 1] = temp
  emit('reorder', newList.map(i => i.id))
}
</script>

<style scoped>
.sub-issues-panel {
  @apply space-y-4;
}
</style>