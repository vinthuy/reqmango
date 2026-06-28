<template>
  <div class="issue-kanban">
    <!-- 搜索栏 -->
    <div class="bg-white rounded-lg border border-gray-200 mb-4">
      <div class="px-4 py-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3 flex-1">
            <button
              @click="showRQL = !showRQL"
              class="px-3 py-1.5 text-sm border rounded-md transition-colors"
              :class="showRQL ? 'bg-indigo-100 border-indigo-300 text-indigo-700' : 'border-gray-300 text-gray-600 hover:bg-gray-50'"
            >
              RQL
            </button>
            <div class="relative flex-1 max-w-md" v-if="!showRQL">
              <input
                v-model="filters.search"
                type="text"
                :placeholder="t('issueKanban.searchPlaceholder')"
                class="w-full pl-8 pr-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                @keydown.enter="reload"
              />
              <svg class="w-4 h-4 text-gray-400 absolute left-2.5 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <div class="flex-1 max-w-md" v-else>
              <RQLInput
                v-model="rqlQuery"
                :placeholder="t('issueKanban.rqlPlaceholder')"
                :show-history="true"
                :show-hints="true"
                :error="rqlError"
                @search="onRQLSearch"
              />
            </div>
            <select v-if="!showRQL" v-model="filters.state_id" @change="reload" class="px-3 py-1.5 border border-gray-300 rounded-md text-sm">
              <option value="0">{{ t('issueList.allStates') }}</option>
              <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
            <select v-model="filters.priority" @change="reload" class="px-3 py-1.5 border border-gray-300 rounded-md text-sm">
              <option value="">{{ t('issueList.allPriorities') }}</option>
              <option value="urgent">{{ t('issue.priorityUrgent') }}</option><option value="high">{{ t('issue.priorityHigh') }}</option><option value="medium">{{ t('issue.priorityMedium') }}</option><option value="low">{{ t('issue.priorityLow') }}</option><option value="none">{{ t('issue.priorityNone') }}</option>
            </select>
            <select v-model="groupBy" class="px-3 py-1.5 border border-gray-300 rounded-md text-sm text-gray-600">
              <option value="state">{{ t('issueKanban.groupByState') }}</option>
              <option value="assignee">{{ t('issueKanban.groupByAssignee') }}</option>
              <option value="priority">{{ t('issueKanban.groupByPriority') }}</option>
              <option value="labels">{{ t('issueKanban.groupByLabels') }}</option>
            </select>
            <select v-model="swimlaneBy" class="px-3 py-1.5 border border-gray-300 rounded-md text-sm text-gray-600">
              <option value="">{{ t('issueKanban.noSwimlane') }}</option>
              <option value="assignee">{{ t('issueKanban.swimlaneAssignee') }}</option>
              <option value="priority">{{ t('issueKanban.swimlanePriority') }}</option>
              <option value="type">{{ t('issueKanban.swimlaneType') }}</option>
            </select>
          </div>
          <div class="flex items-center space-x-2 ml-3">
            <button @click="showAdvanced = !showAdvanced" class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50">
              {{ t('issueKanban.advancedSearch') }}
              <svg class="w-3 h-3 inline ml-1" :class="{ 'rotate-180': showAdvanced }" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
            </button>
            <button @click="showImportModal = true" class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50">{{ t('common.import') }}</button>
          </div>
        </div>
      </div>
      <!-- 高级搜索 -->
      <div v-if="showAdvanced" class="px-4 pb-3 border-t border-gray-100 pt-3 bg-gray-50">
        <div class="grid grid-cols-2 gap-3">
          <div><label class="block text-xs text-gray-500 mb-1">{{ t('issue.cycle') }}</label>
            <select v-model="filters.cycle_id" @change="reload" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm">
              <option value="0">{{ t('common.all') }}</option>
              <option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select></div>
          <div><label class="block text-xs text-gray-500 mb-1">{{ t('issue.assignee') }}</label>
            <UserSelect
              v-model="filtersAssignee"
              :users="memberOptions"
              :placeholder="t('common.all')"
              :clearable="true"
              @update:model-value="reload"
            /></div>
          <div><label class="block text-xs text-gray-500 mb-1">{{ t('issue.startDate') }}</label>
            <input v-model="filters.filter_start_date" type="date" @change="reload" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm" /></div>
          <div><label class="block text-xs text-gray-500 mb-1">{{ t('issue.targetDate') }}</label>
            <input v-model="filters.filter_target_date" type="date" @change="reload" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm" /></div>
        </div>
          <div v-if="customFields.length > 0" class="mt-3 pt-3 border-t border-gray-200">
            <div class="flex items-center justify-between mb-2">
              <span class="text-xs text-gray-500">{{ t('issueKanban.customFieldFilter') }}</span>
              <button @click="addCFCondition" class="text-xs text-indigo-600 hover:text-indigo-800">{{ t('issueKanban.addCondition') }}</button>
            </div>
            <div v-for="(cond, idx) in cfConditions" :key="idx" class="flex gap-2 mb-2">
              <select v-model="cond.field_id" @change="reload" class="flex-1 px-2 py-1.5 border border-gray-300 rounded text-sm">
                <option :value="0">{{ t('issueKanban.selectField') }}</option>
                <option v-for="cf in customFields" :key="cf.id" :value="cf.id">{{ cf.name }}</option>
              </select>
              <input type="text" v-model="cond.value" @input="reload" :placeholder="t('issueKanban.value')" class="flex-1 px-2 py-1.5 border border-gray-300 rounded text-sm" />
              <button @click="removeCFCondition(idx)" class="px-2 py-1.5 text-xs text-red-500 border border-red-200 rounded hover:bg-red-50">×</button>
            </div>
          </div>
        <div class="mt-2 flex justify-end"><button @click="resetFilters" class="text-sm text-gray-500 hover:text-indigo-600">{{ t('issueKanban.resetFilters') }}</button></div>
      </div>
    </div>

    <!-- 批量操作工具栏 -->
    <div v-if="selectedIds.size > 0" class="sticky top-0 z-20 bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-200 dark:border-indigo-800 rounded-lg px-4 py-2 flex items-center gap-3 mb-3 flex-wrap">
      <span class="text-sm text-indigo-700 dark:text-indigo-300 font-medium">{{ t('issueList.selected') }} {{ selectedIds.size }} {{ t('common.items') }}</span>

      <div class="relative">
        <button @click="showBatchState = !showBatchState" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          {{ t('issueList.changeState') }}
        </button>
        <div v-if="showBatchState" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 py-1 w-32">
          <button v-for="s in states" :key="s.id" @click="batchChangeState(s.id)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-200">{{ s.name }}</button>
        </div>
      </div>

      <div class="relative">
        <button @click="showBatchPriority = !showBatchPriority" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          {{ t('issueList.changePriority') }}
        </button>
        <div v-if="showBatchPriority" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 py-1 w-32">
          <button v-for="p in priorityOptions" :key="p.value" @click="batchChangePriority(p.value)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-200">{{ p.label }}</button>
        </div>
      </div>

      <div class="relative">
        <button @click="showBatchAssign = !showBatchAssign" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          {{ t('issueList.batchAssign') }}
        </button>
        <div v-if="showBatchAssign" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 p-2 w-48">
          <UserSelect v-model="batchAssigneeId" :users="memberOptions" :placeholder="t('issueList.selectAssignee')" @update:model-value="batchAssign" />
        </div>
      </div>

      <button @click="execBatchDelete" class="px-2.5 py-1 text-xs border border-red-300 dark:border-red-700 rounded-md bg-white dark:bg-gray-700 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors">
        {{ t('issueList.batchDelete') }}
      </button>

      <button @click="clearSelection" class="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">{{ t('issueList.clearSelection') }}</button>
    </div>

    <!-- Success toast -->
    <div v-if="toastMessage" class="fixed top-4 right-4 z-50 bg-green-50 dark:bg-green-900/50 border border-green-200 dark:border-green-800 text-green-700 dark:text-green-300 text-sm px-4 py-2 rounded-md shadow-lg transition-opacity">
      {{ toastMessage }}
    </div>

    <!-- 看板列 -->
    <div v-if="loading" class="text-center py-12 text-gray-400 text-sm">{{ t('common.loading') }}</div>

    <!-- 无泳道模式 -->
    <div v-else-if="!swimlaneBy" class="grid gap-4" :style="gridStyle">
      <div
        v-for="column in kanbanColumns"
        :key="column.id"
        class="bg-gray-100 dark:bg-gray-800 rounded-lg p-3 min-h-[200px]"
      >
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full" :style="{ backgroundColor: column.color }"></span>
            <h3 class="text-sm font-medium text-gray-700">{{ column.label }}</h3>
          </div>
          <div class="flex items-center space-x-1">
            <button v-if="groupBy === 'state'" @click="openQuickCreate(column.key as number)" class="w-5 h-5 flex items-center justify-center text-gray-500 hover:text-indigo-600 hover:bg-gray-200 rounded text-sm" :title="t('issueKanban.quickCreate')">+</button>
            <span class="text-xs bg-gray-300 text-gray-600 px-1.5 py-0.5 rounded-full">{{ (groupedIssues[column.key] || []).length }}</span>
          </div>
        </div>
        <div v-if="groupBy === 'state' && quickCreateStateId === column.key" class="mb-3">
          <QuickCreateInput :project-id="projectId" :workspace-id="workspaceId" :issue-types="issueTypes" :default-state-id="(column.key as number)" :show-priority="false" inline show-cancel @created="onQuickCreated" @cancel="closeQuickCreate" />
        </div>
        <div class="space-y-2">
          <div v-for="issue in groupedIssues[column.key] || []" :key="issue.id" @click="$emit('select', issue)" class="bg-white dark:bg-gray-700 rounded-md border border-gray-200 dark:border-gray-600 p-2.5 cursor-pointer hover:shadow-md hover:border-indigo-300 dark:hover:border-indigo-600 transition-shadow relative group" draggable="true" @dragstart="onDragStart($event, issue)">
            <div class="absolute top-1.5 right-1.5 opacity-0 group-hover:opacity-100 transition-opacity" @click.stop>
              <input type="checkbox" :checked="selectedIds.has(issue.id)" @change="toggleSelect(issue.id)" class="rounded border-gray-300 dark:border-gray-500 w-3.5 h-3.5" />
            </div>
            <div class="flex items-start justify-between">
              <span class="text-xs text-gray-400 font-mono">DEMO-{{ issue.sequence_id }}</span>
              <span :class="priorityDotClass(issue.priority)" class="w-1.5 h-1.5 rounded-full inline-block"></span>
            </div>
            <p class="text-sm text-gray-800 dark:text-gray-100 mt-1 leading-snug line-clamp-2">{{ issue.name }}</p>
            <div class="mt-2 flex items-center justify-between">
              <div class="flex -space-x-1">
                <div v-for="(a, idx) in (issue.assignees || []).slice(0, 2)" :key="a.id" class="w-5 h-5 rounded-full border border-white flex items-center justify-center text-[10px] font-medium text-white" :style="{ backgroundColor: assigneeColor(idx) }" :title="a.display_name || a.username">{{ getInitials(a.display_name || a.username) }}</div>
              </div>
              <span v-if="issue.cycle" class="text-[10px] text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">{{ issue.cycle.name }}</span>
              <button @click.stop="$emit('select', issue)" class="text-[10px] text-indigo-500 hover:text-indigo-700 font-medium">{{ t('issueKanban.details') }}</button>
            </div>
          </div>
          <div v-if="!(groupedIssues[column.key] || []).length" class="text-center text-xs text-gray-400 py-6">{{ t('issueKanban.dragHere') }}</div>
        </div>
      </div>
    </div>

    <!-- 泳道模式 -->
    <div v-else class="swimlane-board space-y-6">
      <div
        v-for="swimlane in swimlaneKeys"
        :key="swimlane.key"
        class="swimlane-row"
      >
        <!-- 泳道标签 -->
        <div class="flex items-center space-x-2 mb-2 px-1">
          <span class="w-3 h-3 rounded-full flex-shrink-0" :style="{ backgroundColor: swimlane.color }"></span>
          <span class="text-sm font-semibold text-gray-700 dark:text-gray-200">{{ swimlane.label }}</span>
          <span class="text-xs text-gray-400">{{ countSwimlaneIssues(swimlane.key) }} {{ t('common.items') }}</span>
        </div>
        <div class="grid gap-3" :style="gridStyle">
          <div
            v-for="column in kanbanColumns"
            :key="column.id"
            class="bg-gray-100 dark:bg-gray-800 rounded-lg p-3 min-h-[120px]"
          >
            <div class="flex items-center justify-between mb-2">
              <div class="flex items-center space-x-1.5">
                <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: column.color }"></span>
                <h4 class="text-xs font-medium text-gray-500">{{ column.label }}</h4>
              </div>
              <span class="text-[10px] bg-gray-300 text-gray-600 px-1.5 py-0.5 rounded-full">
                {{ (swimlaneGroupedIssues?.[swimlane.key]?.[column.key] || []).length }}
              </span>
            </div>
            <div class="space-y-1.5">
              <div
                v-for="issue in swimlaneGroupedIssues?.[swimlane.key]?.[column.key] || []"
                :key="issue.id"
                @click="$emit('select', issue)"
                class="bg-white dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600 p-2 cursor-pointer hover:shadow-md hover:border-indigo-300 dark:hover:border-indigo-600 transition-shadow relative group"
                draggable="true"
                @dragstart="onDragStart($event, issue)"
              >
                <div class="flex items-start justify-between">
                  <span class="text-[10px] text-gray-400 font-mono">{{ issue.sequence_id }}</span>
                  <span :class="priorityDotClass(issue.priority)" class="w-1.5 h-1.5 rounded-full inline-block"></span>
                </div>
                <p class="text-xs text-gray-800 dark:text-gray-100 mt-0.5 leading-snug line-clamp-2">{{ issue.name }}</p>
              </div>
              <div v-if="!(swimlaneGroupedIssues?.[swimlane.key]?.[column.key] || []).length" class="text-center text-[10px] text-gray-300 py-4">-</div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <ImportIssuesModal
      :visible="showImportModal"
      :project-id="projectId"
      :workspace-id="workspaceId"
      @close="showImportModal = false"
      @success="onImportSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import issueApi from '@/api/issue'
import customFieldApi from '@/api/custom-field'
import api from '@/api'
import UserSelect from '@/components/UserSelect.vue'
import { RQLInput } from '@/components/RQL'
import { useRQL } from '@/composables/useRQL'
import { useConfirm } from '@/composables/useConfirm'
import QuickCreateInput from '@/components/QuickCreateInput.vue'
import ImportIssuesModal from '@/components/ImportIssuesModal.vue'
import * as issueTypeApi from '@/api/issue-type'

const props = defineProps<{ projectId: number; workspaceId: number }>()
defineEmits<{ (e: 'select', issue: any): void }>()
const { t } = useI18n()

// RQL 搜索相关
const {
  rql: rqlQuery,
  error: rqlError,
  search: doRQLSearch,
  results: rqlResults
} = useRQL()

const showRQL = ref(false)

const onRQLSearch = async (_query: string) => {
  await doRQLSearch(props.projectId, 'issue')
  if (rqlResults.value.length > 0) {
    issues.value = rqlResults.value
  }
}

const issues = ref<any[]>([])
const states = ref<any[]>([])
const cycles = ref<any[]>([])
const members = ref<any[]>([])
const loading = ref(false)
const showAdvanced = ref(false)
const groupBy = ref<'state' | 'assignee' | 'priority' | 'labels'>('state')
const swimlaneBy = ref<'assignee' | 'priority' | 'type' | ''>('')

// ---- Batch selection ----
const selectedIds = ref(new Set<number>())
const showBatchState = ref(false)
const showBatchPriority = ref(false)
const showBatchAssign = ref(false)
const batchAssigneeId = ref<number | undefined>(undefined)

const priorityOptions = computed(() => [
  { value: 'urgent', label: t('issue.priorityUrgent') },
  { value: 'high', label: t('issue.priorityHigh') },
  { value: 'medium', label: t('issue.priorityMedium') },
  { value: 'low', label: t('issue.priorityLow') },
  { value: 'none', label: t('issue.priorityNone') },
])

const toastMessage = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

function showToast(msg: string) {
  toastMessage.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMessage.value = '' }, 2500)
}

const { confirm } = useConfirm()

const customFields = ref<any[]>([])
const cfConditions = ref<Array<{ field_id: number; value: string }>>([])
function addCFCondition() { cfConditions.value.push({ field_id: 0, value: '' }) }
function removeCFCondition(idx: number) { cfConditions.value.splice(idx, 1); reload() }
const filters = ref({ search: '', state_id: 0, priority: '', cycle_id: 0, assignee_id: 0, filter_start_date: '', filter_target_date: '', cf_field_id: 0, cf_value: '' })

const memberOptions = computed(() => members.value.map(m => ({
  id: m.user_id,
  display_name: m.user?.display_name || m.user?.username,
  email: m.user?.email
})))

const filtersAssignee = computed({
  get: () => filters.value.assignee_id > 0 ? filters.value.assignee_id : undefined,
  set: (v: number | undefined) => { filters.value.assignee_id = v || 0 }
})

const groupedIssues = computed(() => {
  const map: Record<string, any[]> = {}

  if (groupBy.value === 'state') {
    states.value.forEach(s => { map[s.id] = [] })
    issues.value.forEach(i => {
      if (map[i.state_id]) map[i.state_id].push(i)
    })
  } else if (groupBy.value === 'assignee') {
    // Build columns from unique assignees first
    const seen = new Set<number>()
    issues.value.forEach(i => {
      if (i.assignees && i.assignees.length > 0) {
        i.assignees.forEach((a: any) => {
          if (!seen.has(a.id)) { seen.add(a.id); map[a.id] = [] }
        })
      }
    })
    map['__unassigned__'] = []
    issues.value.forEach(i => {
      if (i.assignees && i.assignees.length > 0) {
        i.assignees.forEach((a: any) => {
          if (map[a.id]) map[a.id].push(i)
        })
      } else {
        map['__unassigned__'].push(i)
      }
    })
  } else if (groupBy.value === 'priority') {
    const keys = ['urgent', 'high', 'medium', 'low', 'none']
    keys.forEach(k => { map[k] = [] })
    issues.value.forEach(i => {
      const key = i.priority || 'none'
      if (map[key]) map[key].push(i)
    })
  } else if (groupBy.value === 'labels') {
    const seen = new Set<number>()
    issues.value.forEach(i => {
      if (i.labels && i.labels.length > 0) {
        i.labels.forEach((l: any) => {
          if (!seen.has(l.id)) { seen.add(l.id); map[l.id] = [] }
        })
      }
    })
    map['__nolabel__'] = []
    issues.value.forEach(i => {
      if (i.labels && i.labels.length > 0) {
        i.labels.forEach((l: any) => {
          if (map[l.id]) map[l.id].push(i)
        })
      } else {
        map['__nolabel__'].push(i)
      }
    })
  }

  return map
})

const kanbanColumns = computed(() => {
  if (groupBy.value === 'state') {
    return states.value.map(s => ({
      id: 'state_' + s.id,
      label: s.name,
      color: s.color || '#6366f1',
      key: s.id,
    }))
  } else if (groupBy.value === 'assignee') {
    const cols: Array<{ id: string; label: string; color: string; key: string | number }> = []
    const seen = new Set<number>()
    issues.value.forEach(i => {
      if (i.assignees && i.assignees.length > 0) {
        i.assignees.forEach((a: any) => {
          if (!seen.has(a.id)) {
            seen.add(a.id)
            cols.push({
              id: 'user_' + a.id,
              label: a.display_name || a.username,
              color: assigneeColor(cols.length),
              key: a.id,
            })
          }
        })
      }
    })
    cols.push({ id: 'unassigned', label: t('issueKanban.unassigned'), color: '#9ca3af', key: '__unassigned__' })
    return cols
  } else if (groupBy.value === 'priority') {
    return [
      { id: 'priority_urgent', label: t('issue.priorityUrgent'), color: '#ef4444', key: 'urgent' },
      { id: 'priority_high', label: t('issue.priorityHigh'), color: '#f97316', key: 'high' },
      { id: 'priority_medium', label: t('issue.priorityMedium'), color: '#eab308', key: 'medium' },
      { id: 'priority_low', label: t('issue.priorityLow'), color: '#22c55e', key: 'low' },
      { id: 'priority_none', label: t('issue.priorityNone'), color: '#9ca3af', key: 'none' },
    ]
  } else {
    // labels
    const cols: Array<{ id: string; label: string; color: string; key: string | number }> = []
    const seen = new Set<number>()
    issues.value.forEach(i => {
      if (i.labels && i.labels.length > 0) {
        i.labels.forEach((l: any) => {
          if (!seen.has(l.id)) {
            seen.add(l.id)
            cols.push({
              id: 'label_' + l.id,
              label: l.name,
              color: l.color || '#6366f1',
              key: l.id,
            })
          }
        })
      }
    })
    cols.push({ id: 'nolabel', label: t('issueKanban.unassigned'), color: '#9ca3af', key: '__nolabel__' })
    return cols
  }
})

const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${kanbanColumns.value.length}, minmax(260px, 1fr))`
}))

// ---- Swimlane support ----

const swimlaneKeys = computed(() => {
  if (!swimlaneBy.value) return []

  if (swimlaneBy.value === 'assignee') {
    const seen = new Set<string>()
    const keys: Array<{ key: string; label: string; color: string }> = []
    issues.value.forEach(i => {
      if (i.assignees && i.assignees.length > 0) {
        i.assignees.forEach((a: any) => {
          const k = String(a.id)
          if (!seen.has(k)) {
            seen.add(k)
            keys.push({ key: k, label: a.display_name || a.username, color: assigneeColor(keys.length) })
          }
        })
      }
    })
    keys.push({ key: '__none__', label: t('issueKanban.unassigned'), color: '#9ca3af' })
    return keys
  }

  if (swimlaneBy.value === 'priority') {
    return [
      { key: 'urgent', label: t('issue.priorityUrgent'), color: '#ef4444' },
      { key: 'high', label: t('issue.priorityHigh'), color: '#f97316' },
      { key: 'medium', label: t('issue.priorityMedium'), color: '#eab308' },
      { key: 'low', label: t('issue.priorityLow'), color: '#22c55e' },
      { key: 'none', label: t('issue.priorityNone'), color: '#9ca3af' },
    ]
  }

  if (swimlaneBy.value === 'type') {
    const seen = new Set<number>()
    const keys: Array<{ key: string; label: string; color: string }> = []
    issues.value.forEach(i => {
      const tid = i.issue_type_id
      if (tid && !seen.has(tid)) {
        seen.add(tid)
        const it = issueTypes.value.find((type: any) => type.id === tid)
        keys.push({ key: String(tid), label: it?.name || t('issueKanban.unassigned'), color: it?.color || '#6366f1' })
      }
    })
    keys.push({ key: '__none__', label: t('issueKanban.unassigned'), color: '#9ca3af' })
    return keys
  }

  return []
})

function getSwimlaneKeyForIssue(issue: any): string {
  if (swimlaneBy.value === 'assignee') {
    if (issue.assignees && issue.assignees.length > 0) return String(issue.assignees[0].id)
    return '__none__'
  }
  if (swimlaneBy.value === 'priority') {
    return issue.priority || 'none'
  }
  if (swimlaneBy.value === 'type') {
    return issue.issue_type_id ? String(issue.issue_type_id) : '__none__'
  }
  return '__none__'
}

const swimlaneGroupedIssues = computed(() => {
  if (!swimlaneBy.value) return null

  const result: Record<string, Record<string | number, any[]>> = {}
  swimlaneKeys.value.forEach(s => { result[s.key] = {} })
  if (groupBy.value === 'state') {
    states.value.forEach(s => { swimlaneKeys.value.forEach(sk => { result[sk.key][s.id] = [] }) })
  }

  issues.value.forEach(i => {
    const sk = getSwimlaneKeyForIssue(i)
    if (!result[sk]) result[sk] = {}

    if (groupBy.value === 'state') {
      if (!result[sk][i.state_id]) result[sk][i.state_id] = []
      result[sk][i.state_id].push(i)
    } else if (groupBy.value === 'assignee') {
      if (i.assignees && i.assignees.length > 0) {
        i.assignees.forEach((a: any) => {
          if (!result[sk][a.id]) result[sk][a.id] = []
          result[sk][a.id].push(i)
        })
      } else {
        if (!result[sk]['__unassigned__']) result[sk]['__unassigned__'] = []
        result[sk]['__unassigned__'].push(i)
      }
    } else if (groupBy.value === 'priority') {
      const k = i.priority || 'none'
      if (!result[sk][k]) result[sk][k] = []
      result[sk][k].push(i)
    } else if (groupBy.value === 'labels') {
      if (i.labels && i.labels.length > 0) {
        i.labels.forEach((l: any) => {
          if (!result[sk][l.id]) result[sk][l.id] = []
          result[sk][l.id].push(i)
        })
      } else {
        if (!result[sk]['__nolabel__']) result[sk]['__nolabel__'] = []
        result[sk]['__nolabel__'].push(i)
      }
    }
  })
  return result
})

function countSwimlaneIssues(swimlaneKey: string): number {
  if (!swimlaneGroupedIssues.value) return 0
  const cols = swimlaneGroupedIssues.value[swimlaneKey]
  if (!cols) return 0
  let count = 0
  for (const k of Object.keys(cols)) {
    count += (cols[k] || []).length
  }
  return count
}

function priorityDotClass(p: string) {
  const m: Record<string, string> = { urgent: 'bg-red-500', high: 'bg-orange-500', medium: 'bg-yellow-500', low: 'bg-green-500', none: 'bg-gray-400' }
  return m[p] || m.none
}
function getInitials(n: string) { return (n || '?')[0]?.toUpperCase() || '?' }
function assigneeColor(i: number) { return ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'][i % 5] }

function onDragStart(e: DragEvent, issue: any) {
  e.dataTransfer?.setData('text/plain', String(issue.id))
}

function reload() { loadIssues() }
function resetFilters() {
  filters.value = { search: '', state_id: 0, priority: '', cycle_id: 0, assignee_id: 0, filter_start_date: '', filter_target_date: '', cf_field_id: 0, cf_value: '' }
  cfConditions.value = []
  reload()
}

onMounted(() => Promise.all([loadIssues(), loadStates(), loadCycles(), loadMembers(), loadCustomFields(), loadIssueTypes()]))
async function loadCustomFields() {
  try { customFields.value = await customFieldApi.listCustomFields(props.workspaceId, props.projectId) } catch { /* */ }
}
const issueTypes = ref<any[]>([])
async function loadIssueTypes() {
  try { issueTypes.value = await issueTypeApi.getIssueTypes(props.workspaceId, props.projectId) } catch { /* */ }
}
const quickCreateStateId = ref<number | null>(null)
function openQuickCreate(stateId: number) { quickCreateStateId.value = stateId }
function closeQuickCreate() { quickCreateStateId.value = null }
function onQuickCreated() { quickCreateStateId.value = null; loadIssues() }

const showImportModal = ref(false)
function onImportSuccess() { loadIssues() }

async function loadIssues() {
  loading.value = true
  try {
    const params: any = { limit: 200 }
    if (filters.value.state_id && filters.value.state_id > 0) params.state_id = filters.value.state_id
    if (filters.value.priority) params.priority = filters.value.priority
    if (filters.value.cycle_id && filters.value.cycle_id > 0) params.cycle_id = filters.value.cycle_id
    if (filters.value.assignee_id && filters.value.assignee_id > 0) params.assignee_id = filters.value.assignee_id
    if (filters.value.search) params.search = filters.value.search
    if (filters.value.filter_start_date) params.start_date = filters.value.filter_start_date
    if (filters.value.filter_target_date) params.target_date = filters.value.filter_target_date
    const activeCF = cfConditions.value.filter(c => c.field_id > 0)
    if (activeCF.length > 0) {
      params.cf_and = JSON.stringify(activeCF)
    }
    const result = await issueApi.listIssues(props.projectId, props.workspaceId, params)
    issues.value = result.items
  } catch (e) { console.error('Failed to load issues:', e) }
  finally { loading.value = false }
}

async function loadStates() {
  try { const r = await api.get(`/projects/${props.projectId}/settings/states`); states.value = r.data } catch (e) { /* */ }
}
async function loadCycles() {
  try { const r = await api.get(`/projects/${props.projectId}/cycles`); cycles.value = r.data } catch (e) { /* */ }
}
async function loadMembers() {
  try { const r = await api.get(`/workspaces/${props.workspaceId}/members`); members.value = r.data } catch (e) { /* */ }
}

// ---- Batch actions ----
function toggleSelect(id: number) {
  const s = new Set(selectedIds.value); s.has(id) ? s.delete(id) : s.add(id); selectedIds.value = s
}

async function batchChangeState(stateId: number) {
  showBatchState.value = false
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { state_id: stateId })
    clearSelection()
    showToast(t('issueList.toastStateUpdated'))
    loadIssues()
  } catch (e) { console.error('Batch state failed:', e) }
}

async function batchChangePriority(priority: string) {
  showBatchPriority.value = false
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { priority: priority as any })
    clearSelection()
    showToast(t('issueList.toastPriorityUpdated'))
    loadIssues()
  } catch (e) { console.error('Batch priority failed:', e) }
}

async function batchAssign(userId: string | number | undefined) {
  showBatchAssign.value = false
  if (!userId) return
  const uid = typeof userId === 'string' ? Number(userId) : userId
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { assignee_ids: [uid] })
    clearSelection()
    showToast(t('issueList.toastAssigned'))
    loadIssues()
  } catch (e) { console.error('Batch assign failed:', e) }
}

function clearSelection() {
  selectedIds.value = new Set()
  showBatchState.value = false
  showBatchPriority.value = false
  showBatchAssign.value = false
}

async function execBatchDelete() {
  if (!(await confirm(t('issueList.confirmDelete').replace('{0}', String(selectedIds.value.size))))) return
  try {
    await issueApi.bulkDeleteIssues([...selectedIds.value])
    clearSelection()
    showToast(t('issueList.toastDeleted'))
    loadIssues()
  } catch (e) { console.error('Batch delete failed:', e) }
}
</script>
```

