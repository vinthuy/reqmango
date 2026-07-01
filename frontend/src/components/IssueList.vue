<template>
  <div class="issue-list bg-white rounded-xl border border-gray-100">
    <!-- Toolbar: 操作按钮 -->
    <div class="px-4 py-2.5 border-b border-gray-100">
      <div class="flex items-center gap-3">
        <!-- Action buttons (right-aligned) -->
        <div class="flex items-center gap-1.5 ml-auto">
          <!-- 快速创建切换 -->
          <button
            @click="showQuickCreate = !showQuickCreate"
            class="flex items-center gap-1.5 px-2.5 py-1.5 text-xs text-gray-600 border border-gray-200 rounded-md hover:bg-gray-50 transition-colors"
            :class="{ 'bg-gray-100 border-gray-300': showQuickCreate }"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>{{ t('issueKanban.quickCreate') }}</span>
          </button>

          <!-- 列配置 -->
          <div class="relative">
            <button
              @click="showColumns = !showColumns"
              class="flex items-center gap-1.5 px-2.5 py-1.5 text-xs text-gray-600 border border-gray-200 rounded-md hover:bg-gray-50 transition-colors"
              :class="{ 'bg-gray-100 border-gray-300': showColumns }"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
              </svg>
              <span>{{ t('issueList.columnConfig') }}</span>
            </button>
            <div v-if="showColumns" class="absolute left-0 top-full mt-1 w-44 bg-white border border-gray-200 rounded-lg shadow-lg z-20 py-1">
              <div class="px-3 py-1.5 text-[11px] text-gray-400 font-medium uppercase tracking-wider border-b border-gray-100">{{ t('issueList.displayColumns') }}</div>
              <label v-for="col in effectiveColumns" :key="col.key" class="flex items-center px-3 py-1.5 hover:bg-gray-50 cursor-pointer text-sm">
                <input type="checkbox" :checked="visibleColumnKeys.has(col.key)" @change="toggleColumn(col.key)" class="rounded border-gray-300 mr-2" />
                {{ t(col.labelKey || "") || col.label }}
              </label>
            </div>
          </div>

          <!-- 导入 -->
          <button @click="showImportModal = true" class="flex items-center gap-1.5 px-2.5 py-1.5 text-xs text-gray-600 border border-gray-200 rounded-md hover:bg-gray-50 transition-colors">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
            </svg>
            <span>{{ t('common.import') }}</span>
          </button>

          <!-- 新建 -->
          <button @click="goToCreate" class="flex items-center gap-1.5 px-3 py-1.5 bg-neutral-900 text-white text-xs rounded-md hover:bg-neutral-800 transition-colors font-medium">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>{{ t('project.create') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 快速创建 -->
    <QuickCreateInput
      v-if="showQuickCreate"
      :project-id="projectId"
      :workspace-id="workspaceId"
      :issue-types="issueTypes"
      @created="onQuickCreated"
    />

    <!-- 批量操作工具栏 -->
    <div v-if="selectedIds.size > 0" class="sticky top-0 z-20 bg-gray-50 border border-gray-200 rounded-lg px-4 py-2 flex items-center gap-3 mb-3 flex-wrap">
      <span class="text-sm text-gray-700 font-medium">{{ t('issueList.selected') }} {{ selectedIds.size }} {{ t('common.items') }}</span>

      <div class="relative">
        <button @click="showBatchState = !showBatchState" class="px-2.5 py-1 text-xs border border-gray-300 rounded-md bg-white hover:bg-gray-50 transition-colors">
          {{ t('issueList.changeState') }}
        </button>
        <div v-if="showBatchState" class="absolute left-0 top-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg z-30 py-1 w-32">
          <button v-for="s in states" :key="s.id" @click="batchChangeState(s.id)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50">{{ s.name }}</button>
        </div>
      </div>

      <div class="relative">
        <button @click="showBatchPriority = !showBatchPriority" class="px-2.5 py-1 text-xs border border-gray-300 rounded-md bg-white hover:bg-gray-50 transition-colors">
          {{ t('issueList.changePriority') }}
        </button>
        <div v-if="showBatchPriority" class="absolute left-0 top-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg z-30 py-1 w-32">
          <button v-for="p in priorityOptions" :key="p.value" @click="batchChangePriority(p.value)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50">{{ p.label }}</button>
        </div>
      </div>

      <div class="relative">
        <button @click="showBatchAssign = !showBatchAssign" class="px-2.5 py-1 text-xs border border-gray-300 rounded-md bg-white hover:bg-gray-50 transition-colors">
          {{ t('issueList.batchAssign') }}
        </button>
        <div v-if="showBatchAssign" class="absolute left-0 top-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg z-30 p-2 w-48">
          <UserSelect v-model="batchAssigneeId" :users="memberOptions" :placeholder="t('issueList.selectAssignee')" @update:model-value="batchAssign" />
        </div>
      </div>

      <button @click="execBatchDelete" class="px-2.5 py-1 text-xs border border-red-200 rounded-md bg-white text-red-500 hover:bg-red-50 transition-colors">
        {{ t('issueList.batchDelete') }}
      </button>

      <button @click="clearSelection" class="text-sm text-gray-500 hover:text-gray-700">{{ t('issueList.clearSelection') }}</button>
    </div>

    <!-- Success toast -->
    <div v-if="toastMessage" class="fixed top-4 right-4 z-50 bg-green-50 border border-green-200 text-green-700 text-sm px-4 py-2 rounded-md shadow-lg transition-opacity">
      {{ toastMessage }}
    </div>

    <!-- 列表内容 -->
    <div class="overflow-visible">
      <div v-if="loading" class="text-center py-16">
        <svg class="animate-spin h-8 w-8 text-gray-400 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
        <p class="mt-2 text-gray-500 text-sm">{{ t('common.loading') }}</p>
      </div>
      <div v-else-if="issues.length === 0" class="text-center py-16">
        <svg class="h-12 w-12 text-gray-300 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
        <p class="mt-2 text-gray-500">{{ t('cycle.noIssues') }}</p>
        <p class="mt-1 text-sm text-gray-400">{{ t('issueList.noIssuesHint') }}</p>
      </div>
      <table v-else class="w-full">
        <thead class="border-b border-gray-100 sticky top-0 bg-white z-10 overflow-visible">
          <tr>
            <th class="w-10 px-3 py-2 text-left">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" class="rounded border-gray-300" />
            </th>
            <th v-for="col in visibleColumns" :key="col.key"
              class="relative px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider group"
              :class="col.width"
            >
              <div class="flex items-center gap-1 select-none">
                <button
                  @click="toggleSort(col.key)"
                  class="flex items-center gap-0.5 hover:text-gray-900 transition-colors"
                  :class="{ 'text-indigo-600': sortBy === col.key }"
                >
                  <span>{{ t(col.labelKey || "") || col.label }}</span>
                  <span class="flex flex-col leading-none -space-y-0.5">
                    <svg class="w-2.5 h-2.5" :class="sortBy === col.key && sortDir === 'asc' ? 'text-indigo-600' : 'text-gray-300 group-hover:text-gray-400'" viewBox="0 0 10 6"><path d="M5 0l5 6H0z" fill="currentColor"/></svg>
                    <svg class="w-2.5 h-2.5" :class="sortBy === col.key && sortDir === 'desc' ? 'text-indigo-600' : 'text-gray-300 group-hover:text-gray-400'" viewBox="0 0 10 6"><path d="M5 6l5-6H0z" fill="currentColor"/></svg>
                  </span>
                </button>
              </div>
            </th>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-20">{{ t('issueList.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="group in groupedIssues" :key="group.key">
            <tr v-if="group.label" class="bg-gray-50">
              <td colspan="100%" class="px-3 py-2">
                <div class="flex items-center gap-2">
                  <span class="w-2 h-2 rounded-full bg-indigo-500"></span>
                  <span class="text-sm font-medium text-gray-700">{{ group.label }}</span>
                  <span class="text-xs text-gray-400">({{ group.issues.length }})</span>
                </div>
              </td>
            </tr>
            <tr v-for="issue in group.issues" :key="issue.id"
              class="border-b border-gray-100 hover:bg-gray-50 cursor-pointer transition-colors"
              :class="{ 'bg-blue-50/50': selectedIds.has(issue.id) }">
              <td class="px-3 py-2.5" @click.stop>
                <input type="checkbox" :checked="selectedIds.has(issue.id)" @change="toggleSelect(issue.id)" class="rounded border-gray-300" />
              </td>
              <td v-for="col in visibleColumns" :key="col.key" class="px-3 py-2.5" @click="$emit('select', issue)">
                <!-- 编号 -->
                <span v-if="col.key === 'sequence_id'" class="text-xs text-gray-400 font-mono">{{ projectIdentifier }}-{{ issue.sequence_id }}</span>
                <!-- 标题 -->
                <span v-else-if="col.key === 'name'" class="text-sm text-gray-800 font-medium line-clamp-2 hover:text-gray-900 transition-colors">{{ issue.name }}</span>
                <!-- 优先级 -->
                <span v-else-if="col.key === 'priority'" :class="priorityClass(issue.priority)" class="text-xs px-1.5 py-0.5 rounded whitespace-nowrap">{{ priorityLabel(issue.priority) }}</span>
                <!-- 类型 -->
                <span v-else-if="col.key === 'issue_type'" class="text-xs whitespace-nowrap">
                  <span v-if="issue.issue_type" class="inline-flex items-center gap-1">
                    <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: issue.issue_type.color }"></span>
                    <span class="text-gray-600">{{ issue.issue_type.name }}</span>
                  </span>
                  <span v-else class="text-gray-400">-</span>
                </span>
                <!-- 状态 -->
                <span v-else-if="col.key === 'state'" class="text-xs text-gray-600 whitespace-nowrap">{{ getStateName(issue.state_id) }}</span>
                <!-- 负责人 -->
                <div v-else-if="col.key === 'assignees'" class="flex -space-x-1">
                  <div v-for="(a, idx) in (issue.assignees || []).slice(0, 3)" :key="a.id"
                    class="w-5 h-5 rounded-full border border-white flex items-center justify-center text-[10px] font-medium text-white ring-2 ring-white"
                    :style="{ backgroundColor: assigneeColor(idx) }" :title="a.display_name || a.username">{{ getInitials(a.display_name || a.username) }}</div>
                  <span v-if="!issue.assignees?.length" class="text-xs text-gray-400">-</span>
                </div>
                <!-- 周期 -->
                <span v-else-if="col.key === 'cycle'" class="text-xs text-gray-500 whitespace-nowrap">{{ getCycleName(issue) }}</span>
                <!-- 日期 -->
                <span v-else-if="col.key === 'start_date'" class="text-xs text-gray-500">{{ formatDate(issue.start_date) }}</span>
                <span v-else-if="col.key === 'target_date'" class="text-xs text-gray-500">{{ formatDate(issue.target_date) }}</span>
                <span v-else-if="col.key === 'created_at'" class="text-xs text-gray-400">{{ formatDate(issue.created_at) }}</span>
                <!-- 自定义字段 -->
                <span v-else-if="col.key.startsWith('cf_')" class="text-xs text-gray-600 whitespace-nowrap">{{ getCFValue(issue.id, col.key) }}</span>
              </td>
              <td class="px-3 py-2.5" @click.stop>
                <button @click="$emit('select', issue)" class="text-xs text-blue-500 hover:text-blue-700 font-medium">{{ t('issueList.view') }}</button>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div v-if="totalPages > 1" class="px-4 py-3 border-t border-gray-200 flex items-center justify-between bg-gray-50/50">
      <span class="text-sm text-gray-500">{{ t('issueList.total') }} {{ totalCount }} {{ t('common.items') }}</span>
      <div class="flex items-center gap-1">
        <button @click="page--" :disabled="page <= 1" class="px-3 py-1 border rounded text-sm disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-100 transition-colors">{{ t('issueList.prevPage') }}</button>
        <template v-for="p in visiblePages" :key="p">
          <button v-if="p === '...'" disabled class="px-2 py-1 text-sm text-gray-400">...</button>
          <button v-else @click="page = Number(p)" class="px-3 py-1 border rounded text-sm transition-colors"
            :class="page === Number(p) ? 'bg-neutral-900 text-white border-neutral-900' : 'hover:bg-gray-100'">{{ p }}</button>
        </template>
        <button @click="page++" :disabled="page >= totalPages" class="px-3 py-1 border rounded text-sm disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-100 transition-colors">{{ t('issueList.nextPage') }}</button>
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
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import issueApi from '@/api/issue'
import customFieldApi from '@/api/custom-field'
import projectApi from '@/api/project'
import api from '@/api'
import UserSelect from '@/components/UserSelect.vue'
import { useConfirm } from '@/composables/useConfirm'
import QuickCreateInput from '@/components/QuickCreateInput.vue'
import ImportIssuesModal from '@/components/ImportIssuesModal.vue'
import * as issueTypeApi from '@/api/issue-type'

const props = defineProps<{ projectId: number; workspaceId: number; rql?: string; filterSortBy?: string; filterSortDir?: string; filterGroupBy?: string }>()
const router = useRouter()
const { t, locale } = useI18n()

const emit = defineEmits<{
  (e: 'select', issue: any): void
  (e: 'delete', issue: any): void
}>()

// ── Project Identifier ──
const projectIdentifier = ref('')

// ── Column config ──
interface ColumnDef { key: string; label?: string; labelKey?: string; width: string; defaultVisible: boolean; sortable: boolean; filterable: boolean }
const staticColumns: ColumnDef[] = [
  { key: 'sequence_id', labelKey: 'issueList.columnSequenceId', width: 'w-20', defaultVisible: true, sortable: true, filterable: false },
  { key: 'name',         labelKey: 'issueList.columnName', width: '',       defaultVisible: true, sortable: true, filterable: true },
  { key: 'priority',     labelKey: 'issue.priority', width: 'w-20', defaultVisible: true, sortable: true, filterable: true },
  { key: 'issue_type',   labelKey: 'issue.type', width: 'w-20', defaultVisible: true, sortable: true, filterable: true },
  { key: 'state',        labelKey: 'issue.state',   width: 'w-28', defaultVisible: true, sortable: true, filterable: true },
  { key: 'assignees',    labelKey: 'issue.assignee', width: 'w-28', defaultVisible: true, sortable: false, filterable: true },
  { key: 'cycle',        labelKey: 'issue.cycle',   width: 'w-24', defaultVisible: true, sortable: false, filterable: true },
  { key: 'start_date',   labelKey: 'issue.startDate', width: 'w-28', defaultVisible: false, sortable: true, filterable: false },
  { key: 'target_date',  labelKey: 'issue.targetDate', width: 'w-28', defaultVisible: false, sortable: true, filterable: false },
  { key: 'created_at',   labelKey: 'issueList.columnCreatedAt', width: 'w-36', defaultVisible: false, sortable: true, filterable: false },
]

const STORAGE_KEY = 'issuelist_columns'

const customFields = ref<any[]>([])

const effectiveColumns = computed(() => {
  const cols = [...staticColumns]
  for (const cf of customFields.value) {
    cols.push({ key: 'cf_' + cf.id, label: cf.name, width: 'w-28', defaultVisible: false, sortable: false, filterable: false } as ColumnDef)
  }
  return cols
})

function loadColumnPrefs(): Set<string> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return new Set(JSON.parse(raw))
  } catch { /* */ }
  return new Set(effectiveColumns.value.filter(c => c.defaultVisible).map(c => c.key))
}

const visibleColumnKeys = ref(loadColumnPrefs())
const showColumns = ref(false)

const visibleColumns = computed(() => effectiveColumns.value.filter(c => visibleColumnKeys.value.has(c.key)))

function toggleColumn(key: string) {
  const s = new Set(visibleColumnKeys.value)
  s.has(key) ? s.delete(key) : s.add(key)
  visibleColumnKeys.value = s
  saveColumnPrefs()
}

function saveColumnPrefs() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify([...visibleColumnKeys.value]))
}

// ── Column sort state ──
const sortBy = ref<string>('')
const sortDir = ref<'asc' | 'desc'>('desc')

function toggleSort(colKey: string) {
  if (sortBy.value === colKey) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = colKey
    sortDir.value = 'desc'
  }
  page.value = 1
  loadIssues()
}

// ── Data state ──
const issues = ref<any[]>([])
const states = ref<any[]>([])
const cycles = ref<any[]>([])
const members = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const limit = ref(20)
const totalCount = ref(0)
const totalPages = ref(1)
const selectedIds = ref(new Set<number>())
const showImportModal = ref(false)
const showQuickCreate = ref(false)

// ── Batch Operations ──
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

// ── Computed ──
const isAllSelected = computed(() => issues.value.length > 0 && issues.value.every(i => selectedIds.value.has(i.id)))

const memberOptions = computed(() => members.value.map((m: any) => ({
  id: m.user_id,
  display_name: m.user?.display_name || m.user?.username,
  email: m.user?.email
})))

interface GroupedIssue {
  key: string
  label: string
  issues: any[]
}

const groupedIssues = computed((): GroupedIssue[] => {
  const groupBy = props.filterGroupBy
  if (!groupBy || groupBy === 'none') {
    return [{ key: 'all', label: '', issues: issues.value }]
  }

  const groups: Record<string, GroupedIssue> = {}

  for (const issue of issues.value) {
    let key: string
    let label: string

    switch (groupBy) {
      case 'state_id':
        key = String(issue.state_id)
        label = getStateName(issue.state_id)
        break
      case 'priority':
        key = issue.priority || 'none'
        label = priorityLabel(issue.priority || 'none')
        break
      case 'assignee_id':
        const assignee = (issue.assignees && issue.assignees.length > 0) ? issue.assignees[0] : null
        key = assignee ? String(assignee.id) : 'unassigned'
        label = assignee ? (assignee.display_name || assignee.username) : t('issueList.unassigned')
        break
      case 'type_id':
        key = String(issue.issue_type?.id || 0)
        label = issue.issue_type?.name || '-'
        break
      case 'cycle_id':
        key = String(issue.cycle_link?.id || 0)
        label = getCycleName(issue)
        break
      case 'module_id':
        key = String(issue.module_id || 0)
        label = issue.module?.name || '-'
        break
      case 'label':
        const labels = issue.labels || []
        key = labels.length > 0 ? String(labels[0].id) : 'no_label'
        label = labels.length > 0 ? labels[0].name : t('issueList.noLabel')
        break
      default:
        key = 'all'
        label = ''
    }

    if (!groups[key]) {
      groups[key] = { key, label, issues: [] }
    }
    groups[key].issues.push(issue)
  }

  return Object.values(groups).sort((a, b) => a.label.localeCompare(b.label))
})

const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const tp = totalPages.value; const p = page.value
  if (tp <= 7) { for (let i = 1; i <= tp; i++) pages.push(i); return pages }
  pages.push(1); if (p > 3) pages.push('...')
  for (let i = Math.max(2, p - 1); i <= Math.min(tp - 1, p + 1); i++) pages.push(i)
  if (p < tp - 2) pages.push('...'); pages.push(tp)
  return pages
})

// ── Helpers ──
function priorityClass(p: string) {
  const m: Record<string, string> = { urgent: 'bg-red-100 text-red-700', high: 'bg-orange-100 text-orange-700', medium: 'bg-yellow-100 text-yellow-700', low: 'bg-green-100 text-green-700', none: 'bg-gray-100 text-gray-500' }
  return m[p] || m.none
}
function priorityLabel(p: string) { const m: Record<string, string> = { urgent: t('issue.priorityUrgent'), high: t('issue.priorityHigh'), medium: t('issue.priorityMedium'), low: t('issue.priorityLow'), none: t('issue.priorityNone') }; return m[p] || p }
function getStateName(id: number) { return states.value.find((s: any) => s.id === id)?.name || '-' }
function getCycleName(i: any) { return i.cycle?.name || i.cycle_link?.name || '-' }
function getInitials(n: string) { return (n || '?')[0]?.toUpperCase() || '?' }
function assigneeColor(i: number) { return ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'][i % 6] }
function formatDate(d: string | null | undefined) { if (!d) return '-'; return new Date(d).toLocaleDateString(locale.value) }

// ── Actions ──
function goToCreate() { router.push(`/workspaces/${props.workspaceId}/projects/${props.projectId}/issues/new`) }
function onQuickCreated() { page.value = 1; loadIssues() }
function onImportSuccess() { page.value = 1; loadIssues() }

function toggleSelectAll() {
  if (isAllSelected.value) selectedIds.value.clear()
  else issues.value.forEach(i => selectedIds.value.add(i.id))
  selectedIds.value = new Set(selectedIds.value)
}
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

// ═══ Data loading ═══
async function loadIssues() {
  loading.value = true
  try {
    const params: any = { limit: limit.value, offset: (page.value - 1) * limit.value }
    // RQL filter from FilterBar
    if (props.rql) params.rql = props.rql
    // Sort: use FilterBar sort if provided, otherwise fall back to column sort
    if (props.filterSortBy) {
      params.sort_by = props.filterSortBy
      params.sort_dir = props.filterSortDir || 'desc'
    } else if (sortBy.value) {
      params.sort_by = sortBy.value
      params.sort_dir = sortDir.value
    }

    const result = await issueApi.listIssues(props.projectId, props.workspaceId, params)
    issues.value = result.items; totalCount.value = result.total; totalPages.value = Math.max(1, Math.ceil(result.total / limit.value))
  } catch (e) { console.error('Failed to load issues:', e) }
  finally { loading.value = false }
}

async function loadStates() { try { const r = await api.get(`/projects/${props.projectId}/settings/states`); states.value = r.data } catch (e) { /* */ } }
async function loadCycles() { try { const r = await api.get(`/projects/${props.projectId}/cycles`); cycles.value = r.data } catch (e) { /* */ } }
async function loadMembers() { try { const r = await api.get(`/workspaces/${props.workspaceId}/members`); members.value = r.data } catch (e) { /* */ } }
async function loadCustomFields() {
  try { customFields.value = await customFieldApi.listCustomFields(props.workspaceId, props.projectId) } catch (e) { /* */ }
  // Re-initialize column prefs now that custom fields are loaded
  const allKeys = effectiveColumns.value.map(c => c.key)
  const newKeys = new Set(visibleColumnKeys.value)
  for (const key of allKeys) {
    if (!newKeys.has(key) && effectiveColumns.value.find(c => c.key === key)?.defaultVisible) {
      newKeys.add(key)
    }
  }
  visibleColumnKeys.value = newKeys
}
const issueTypes = ref<any[]>([])
async function loadIssueTypes() {
  try { issueTypes.value = await issueTypeApi.getIssueTypes(props.workspaceId, props.projectId) } catch (e) { /* */ }
}

// Custom field value cache
const cfValueCache = ref<Record<number, Record<number, string>>>({})

async function loadCFValues(issueIds: number[]) {
  if (!issueIds.length) return
  try {
    const results = await Promise.all(issueIds.map(id =>
      customFieldApi.listIssueCustomFieldValues(id).catch(() => [] as any[])
    ))
    const cache: Record<number, Record<number, string>> = {}
    issueIds.forEach((id, idx) => {
      cache[id] = {}
      for (const v of (results[idx] || [])) {
        cache[id][v.field_id] = v.value || ''
      }
    })
    cfValueCache.value = cache
  } catch { /* */ }
}

function getCFValue(issueId: number, colKey: string): string {
  const fieldId = parseInt(colKey.replace('cf_', ''))
  return cfValueCache.value[issueId]?.[fieldId] || '-'
}

async function loadProjectInfo() {
  try {
    const project = await projectApi.getProject(props.projectId)
    projectIdentifier.value = project.identifier || 'PROJ'
  } catch { /* */ }
}

// ── Lifecycle ──
onMounted(() => {
  if (props.workspaceId > 0) Promise.all([loadIssues(), loadStates(), loadCycles(), loadMembers(), loadCustomFields(), loadIssueTypes(), loadProjectInfo()])
})

onUnmounted(() => {
})

watch(page, () => loadIssues())
watch(() => props.workspaceId, (id) => {
  if (id > 0) Promise.all([loadIssues(), loadStates(), loadCycles(), loadMembers(), loadCustomFields(), loadIssueTypes(), loadProjectInfo()])
})
watch(() => props.rql, () => {
  page.value = 1
  loadIssues()
})

watch([() => props.filterSortBy, () => props.filterSortDir], () => {
  page.value = 1
  loadIssues()
})

watch(() => props.filterGroupBy, () => {
  page.value = 1
  loadIssues()
})

watch(issues, (newIssues) => {
  if (newIssues.length) loadCFValues(newIssues.map(i => i.id))
})
</script>
