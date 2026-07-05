<template>
  <div class="border border-green-200 rounded-lg overflow-hidden bg-green-50/30">
    <!-- Card header -->
    <div class="flex items-center justify-between px-3 py-2 bg-green-50 border-b border-green-100">
      <div class="flex items-center gap-2">
        <span class="text-xs font-semibold text-green-700">{{ t('subIssue.title') }}</span>
        <span
          v-if="subIssues.length > 0"
          class="px-1.5 py-0.5 rounded-full text-[10px] font-medium bg-green-100 text-green-700"
        >{{ completedCount }}/{{ subIssues.length }}</span>
      </div>
      <button
        data-test="add-subissue"
        class="text-[10px] font-medium text-green-600 hover:text-green-800 hover:underline"
        @click="$emit('add')"
      >+ {{ t('subIssue.add') }}</button>
    </div>

    <!-- Empty state -->
    <div v-if="subIssues.length === 0" class="px-4 py-6 text-center text-xs text-gray-400">
      {{ t('subIssue.noSubIssues') }}
    </div>

    <!-- Table of sub-issues -->
    <div v-else class="overflow-x-auto">
      <table class="w-full text-[10px]">
        <thead>
          <tr class="border-b border-green-100 text-gray-500">
            <th class="px-3 py-1.5 text-left w-8"></th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.type') }}</th>
            <th class="px-2 py-1.5 text-left">ID</th>
            <th class="px-2 py-1.5 text-left">{{ t('subIssue.title') }}</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.status') }}</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.priority') }}</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.assignee') }}</th>
            <th class="px-2 py-1.5 text-right">{{ t('subIssue.targetDate') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="issue in subIssues"
            :key="issue.id"
            class="border-b border-green-50 last:border-b-0 hover:bg-green-50/50 transition-colors"
          >
            <!-- Checkbox -->
            <td class="px-3 py-2">
              <input
                type="checkbox"
                :checked="issue.state_group === 'done'"
                @change="$emit('toggle', issue.id)"
              />
            </td>
            <!-- Type badge -->
            <td class="px-2 py-2">
              <span
                class="px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap"
                :style="{ backgroundColor: issue.issue_type?.color + '20', color: issue.issue_type?.color }"
              >{{ issue.issue_type?.name || '—' }}</span>
            </td>
            <!-- ID -->
            <td class="px-2 py-2 font-mono text-gray-400">#{{ issue.sequence_id }}</td>
            <!-- Title (clickable) -->
            <td class="px-2 py-2">
              <span
                class="subissue-clickable text-xs font-medium text-gray-800 cursor-pointer hover:text-indigo-600"
                @click="$emit('navigate', issue.id)"
              >{{ issue.name }}</span>
            </td>
            <!-- State -->
            <td class="px-2 py-2">
              <span class="flex items-center gap-1">
                <span
                  class="w-1.5 h-1.5 rounded-full"
                  :style="{ backgroundColor: stateColor(issue.state_group) }"
                ></span>
                {{ issue.state_name }}
              </span>
            </td>
            <!-- Priority -->
            <td class="px-2 py-2">
              <span class="flex items-center gap-1">
                <span
                  class="w-1.5 h-1.5 rounded-full"
                  :style="{ backgroundColor: priorityColor(issue.priority) }"
                ></span>
                {{ t(`issue.priority${(issue.priority || 'none').charAt(0).toUpperCase() + (issue.priority || 'none').slice(1)}`) }}
              </span>
            </td>
            <!-- Assignee -->
            <td class="px-2 py-2 text-gray-500">
              {{ issue.assignees?.[0]?.display_name || issue.assignees?.[0]?.username || '—' }}
            </td>
            <!-- Due date -->
            <td class="px-2 py-2 text-gray-400 text-right">
              {{ issue.target_date ? formatDate(issue.target_date) : '—' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { stateColor, priorityColor, formatDate } from '@/composables/useRelationHelpers'

const { t } = useI18n()

interface IssueType {
  id: number
  name: string
  color: string
}

interface Assignee {
  id: number
  display_name?: string
  username?: string
}

interface SubIssue {
  id: number
  sequence_id: number
  name: string
  state_name: string
  state_group: string
  priority: string
  assignees?: Assignee[]
  target_date?: string | null
  issue_type?: IssueType
}

const props = defineProps<{
  subIssues: SubIssue[]
}>()

defineEmits<{
  add: []
  toggle: [issueId: number]
  navigate: [issueId: number]
}>()

const completedCount = computed(() => props.subIssues.filter((s) => s.state_group === 'done').length)
</script>
