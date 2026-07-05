<template>
  <div class="border border-gray-200 rounded-lg overflow-hidden">
    <!-- Card header -->
    <div
      data-test="card-header"
      class="flex items-center justify-between px-3 py-2 cursor-pointer select-none"
      :style="{ backgroundColor: color + '10', borderBottom: isExpanded ? `1px solid ${color}30` : 'none' }"
      @click="isExpanded = !isExpanded"
    >
      <div class="flex items-center gap-2">
        <svg
          class="w-3 h-3 transition-transform"
          :class="{ 'rotate-90': isExpanded }"
          :style="{ color }"
          fill="none" stroke="currentColor" viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <span class="text-xs font-semibold" :style="{ color }">{{ typeName.toUpperCase() }}</span>
        <span
          v-if="items.length > 0"
          class="px-1.5 py-0.5 rounded-full text-[10px] font-medium"
          :style="{ backgroundColor: color + '30', color }"
        >{{ items.length }}</span>
      </div>
      <button
        v-if="items.length > 0"
        data-test="add-link-header"
        class="text-[10px] font-medium hover:underline"
        :style="{ color }"
        @click.stop="$emit('add')"
      >+ Link</button>
    </div>

    <!-- Card body -->
    <div v-if="isExpanded">
      <!-- Empty state -->
      <div v-if="items.length === 0" class="px-4 py-6 text-center text-xs text-gray-400">
        {{ t('issueKanban.noRelations') }}
        <button
          data-test="add-link-empty"
          class="block mx-auto mt-2 text-blue-500 hover:text-blue-700"
          @click="$emit('add')"
        >+ {{ t('issue.addRelation') }}</button>
      </div>

      <!-- Relation rows -->
      <div
        v-for="item in items"
        :key="item.id"
        class="relation-row flex items-center gap-2 px-3 py-2 border-b border-gray-50 last:border-b-0 hover:bg-gray-50 transition-colors"
      >
        <!-- Type badge -->
        <span
          class="px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap shrink-0"
          :style="{ backgroundColor: getIssueType(item.related_issue)?.color + '20', color: getIssueType(item.related_issue)?.color }"
        >{{ getIssueType(item.related_issue)?.name || '—' }}</span>
        <!-- ID -->
        <span class="text-[10px] text-gray-400 shrink-0 font-mono">#{{ item.related_issue.sequence_id }}</span>
        <!-- Title (clickable) -->
        <span
          class="relation-row-clickable text-xs font-medium text-gray-800 flex-1 min-w-0 truncate cursor-pointer hover:text-indigo-600"
          @click="$emit('navigate', item.related_issue_id)"
        >{{ item.related_issue.name }}</span>
        <!-- State -->
        <span class="flex items-center gap-1 text-[10px] shrink-0">
          <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: stateColor(item.related_issue.state_group) }"></span>
          {{ item.related_issue.state_name }}
        </span>
        <!-- Priority -->
        <span class="flex items-center gap-1 text-[10px] shrink-0">
          <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: priorityColor(item.related_issue.priority) }"></span>
          {{ t(`issue.priority${(item.related_issue.priority || 'none').charAt(0).toUpperCase() + (item.related_issue.priority || 'none').slice(1)}`) }}
        </span>
        <!-- Assignee -->
        <span class="text-[10px] text-gray-500 shrink-0 w-12 truncate">
          {{ item.related_issue.assignees?.[0]?.display_name || item.related_issue.assignees?.[0]?.username || '—' }}
        </span>
        <!-- Due Date -->
        <span class="text-[10px] text-gray-400 shrink-0 w-12 text-right">
          {{ item.related_issue.target_date ? formatDate(item.related_issue.target_date) : '—' }}
        </span>
        <!-- Remove -->
        <button
          data-test="remove-relation"
          class="text-gray-300 hover:text-red-500 shrink-0 text-sm leading-none"
          @click="$emit('remove', item.id)"
        >&times;</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { stateColor, priorityColor, formatDate } from '@/composables/useRelationHelpers'

const { t } = useI18n()

interface IssueType {
  id: number
  name: string
  color: string
}

interface RelationItem {
  id: number
  related_issue: {
    id: number
    sequence_id: number
    name: string
    state_name: string
    state_group: string
    priority: string
    assignees?: Array<{ id: number; display_name?: string; username?: string }>
    start_date?: string | null
    target_date?: string | null
    issue_type?: IssueType
  }
  related_issue_id: number
}

const props = defineProps<{
  typeName: string
  items: RelationItem[]
  color: string
  issueTypes: IssueType[]
}>()

defineEmits<{
  add: []
  remove: [relationId: number]
  navigate: [issueId: number]
}>()

const isExpanded = ref(true)

function getIssueType(issue: RelationItem['related_issue']) {
  if (!props.issueTypes) return null
  return props.issueTypes.find((t) => t.id === issue?.issue_type?.id) || issue?.issue_type || null
}

</script>
