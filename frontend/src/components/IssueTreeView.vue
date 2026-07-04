<template>
  <div class="issue-tree-view bg-white rounded-lg border border-gray-200 relative">
    <!-- Success toast -->
    <div v-if="toastMessage" class="absolute top-2 right-2 z-10 bg-green-50 border border-green-200 text-green-700 text-sm px-3 py-2 rounded-md shadow-sm transition-opacity">
      {{ toastMessage }}
    </div>

    <!-- Toolbar -->
    <div class="px-4 py-2.5 border-b border-gray-100">
      <div class="flex items-center gap-3">
        <div class="flex-1" />

        <!-- Expand / Collapse All -->
        <button @click="expandAll" class="text-xs text-gray-500 hover:text-indigo-600 transition-colors">
          {{ t('issueTreeView.expandAll') }}
        </button>
        <button @click="collapseAll" class="text-xs text-gray-500 hover:text-indigo-600 transition-colors">
          {{ t('issueTreeView.collapseAll') }}
        </button>

        <!-- Create -->
        <button @click="$emit('create')" class="flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 text-white text-xs rounded-md hover:bg-indigo-700 transition-colors font-medium">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          <span>{{ t('issueTreeView.create') }}</span>
        </button>
      </div>
    </div>

    <!-- 批量操作工具栏 -->
    <div v-if="selectedIds.size > 0" class="sticky top-0 z-20 bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-200 dark:border-indigo-800 rounded-lg mx-4 mt-3 px-4 py-2 flex items-center gap-3 flex-wrap">
      <span class="text-sm text-indigo-700 dark:text-indigo-300 font-medium">{{ t('issueTreeView.selected') }} {{ selectedIds.size }} {{ t('issueTreeView.items') }}</span>

      <div class="relative">
        <button @click="showBatchState = !showBatchState" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          {{ t('issueTreeView.changeState') }}
        </button>
        <div v-if="showBatchState" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 py-1 w-32">
          <button v-for="s in states" :key="s.id" @click="batchChangeState(s.id)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-200">{{ s.name }}</button>
        </div>
      </div>

      <div class="relative">
        <button @click="showBatchPriority = !showBatchPriority" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          {{ t('issueTreeView.changePriority') }}
        </button>
        <div v-if="showBatchPriority" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 py-1 w-32">
          <button v-for="p in priorityOptions" :key="p.value" @click="batchChangePriority(p.value)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-200">{{ p.label }}</button>
        </div>
      </div>

      <div class="relative">
        <button @click="showBatchAssign = !showBatchAssign" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          {{ t('issueTreeView.batchAssign') }}
        </button>
        <div v-if="showBatchAssign" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 p-2 w-48">
          <UserSelect v-model="batchAssigneeId" :users="memberOptions" :placeholder="t('issueTreeView.selectAssignee')" @update:model-value="batchAssign" />
        </div>
      </div>

      <button @click="execBatchDelete" class="px-2.5 py-1 text-xs border border-red-300 dark:border-red-700 rounded-md bg-white dark:bg-gray-700 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors">
        {{ t('issueTreeView.batchDelete') }}
      </button>

      <button @click="clearSelection" class="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">{{ t('issueTreeView.clearSelection') }}</button>
    </div>

    <!-- Column header -->
    <div class="flex items-center px-4 py-2 bg-gray-50 dark:bg-gray-700/50 border-b border-gray-100 dark:border-gray-700 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
      <div class="w-8 shrink-0"></div>
      <div class="flex-1 min-w-0">{{ t('issueTreeView.columnTitle') }}</div>
      <div class="w-16 text-center shrink-0">{{ t('issueTreeView.columnPriority') }}</div>
      <div class="w-20 text-center shrink-0">{{ t('issueTreeView.columnState') }}</div>
      <div class="w-16 text-center shrink-0">{{ t('issueTreeView.columnChildren') }}</div>
    </div>

    <!-- Tree content -->
    <div v-if="loading" class="text-center py-16">
      <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
      </svg>
      <p class="mt-2 text-gray-500 text-sm">{{ t('issueTreeView.loading') }}</p>
    </div>

    <div v-else-if="rootNodes.length === 0 && !loading" class="text-center py-16">
      <svg class="h-12 w-12 text-gray-300 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
      </svg>
      <p class="mt-2 text-gray-500">{{ t('issueTreeView.noIssues') }}</p>
      <p class="mt-1 text-sm text-gray-400">{{ t('issueTreeView.noIssuesHint') }}</p>
    </div>

    <div v-else class="divide-y divide-gray-100">
      <TreeNodeItem
        v-for="node in rootNodes"
        :key="node.id"
        :node="node"
        :depth="0"
        :expanded-nodes="expandedNodes"
        :children-map="childrenMap"
        :loading-children="loadingChildren"
        :project-identifier="projectIdentifier"
        :selected-ids="selectedIds"
        @toggle="toggleNode"
        @select="$emit('select', $event)"
        @toggle-select="toggleSelect"
        @create-child="handleCreateChild"
      />
    </div>

    <!-- Pagination for root level -->
    <div v-if="totalPages > 1" class="px-4 py-3 border-t border-gray-200 flex items-center justify-between bg-gray-50/50">
      <span class="text-sm text-gray-500">{{ t('issueTreeView.total') }} {{ t('issueTreeView.totalItems', { count: totalCount }) }}</span>
      <div class="flex items-center gap-1">
        <button @click="page--" :disabled="page <= 1" class="px-3 py-1 border rounded text-sm disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-100 transition-colors">{{ t('issueTreeView.prevPage') }}</button>
        <template v-for="p in visiblePages" :key="p">
          <button v-if="p === '...'" disabled class="px-2 py-1 text-sm text-gray-400">...</button>
          <button v-else @click="page = Number(p)" class="px-3 py-1 border rounded text-sm transition-colors"
            :class="page === Number(p) ? 'bg-indigo-600 text-white border-indigo-600' : 'hover:bg-gray-100'">{{ p }}</button>
        </template>
        <button @click="page++" :disabled="page >= totalPages" class="px-3 py-1 border rounded text-sm disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-100 transition-colors">{{ t('issueTreeView.nextPage') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import issueApi from '@/api/issue'
import projectApi from '@/api/project'
import api from '@/api'
import type { TreeIssueResponse } from '@/types/issue'
import TreeNodeItem from './TreeNodeItem.vue'
import UserSelect from '@/components/UserSelect.vue'
import { useConfirm } from '@/composables/useConfirm'
import { useI18n } from '@/composables/useI18n'

const props = defineProps<{ projectId: number; workspaceId: number; rql?: string; filterSortBy?: string; filterSortDir?: string; filterGroupBy?: string }>()

const emit = defineEmits<{
  (e: 'select', issue: any): void
  (e: 'create'): void
}>()

const { t } = useI18n()

// ── State ──
const rootNodes = ref<TreeIssueResponse[]>([])
const loading = ref(false)
const page = ref(1)
const limit = ref(20)
const totalCount = ref(0)
const totalPages = ref(1)
const projectIdentifier = ref('')

// Tree state
const expandedNodes = ref<Set<number>>(new Set())
const childrenMap = ref<Map<number, TreeIssueResponse[]>>(new Map())
const loadingChildren = ref<Set<number>>(new Set())

// ── Batch selection ──
const selectedIds = ref(new Set<number>())
const showBatchState = ref(false)
const showBatchPriority = ref(false)
const showBatchAssign = ref(false)
const batchAssigneeId = ref<number | undefined>(undefined)
const states = ref<any[]>([])
const members = ref<any[]>([])

const memberOptions = computed(() => members.value.map((m: any) => ({
  id: m.user_id,
  display_name: m.user?.display_name || m.user?.username,
  email: m.user?.email
})))

const priorityOptions = [
  { value: 'urgent', label: t('issueTreeView.urgent') },
  { value: 'high', label: t('issueTreeView.high') },
  { value: 'medium', label: t('issueTreeView.medium') },
  { value: 'low', label: t('issueTreeView.low') },
  { value: 'none', label: t('issueTreeView.none') },
]

const { confirm } = useConfirm()

const toastMessage = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

function showToast(msg: string) {
  toastMessage.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMessage.value = '' }, 2500)
}

// ── Computed ──
const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const tp = totalPages.value; const p = page.value
  if (tp <= 7) { for (let i = 1; i <= tp; i++) pages.push(i); return pages }
  pages.push(1); if (p > 3) pages.push('...')
  for (let i = Math.max(2, p - 1); i <= Math.min(tp - 1, p + 1); i++) pages.push(i)
  if (p < tp - 2) pages.push('...'); pages.push(tp)
  return pages
})

// ═══ Data loading — filters ONLY via props.rql (unified FilterBar) ═══
async function loadRootNodes() {
  loading.value = true
  try {
    const params: any = { limit: limit.value, offset: (page.value - 1) * limit.value }
    if (props.rql) params.rql = props.rql
    const result = await issueApi.listTreeIssues(props.projectId, params)
    rootNodes.value = result.items
    totalCount.value = result.total
    totalPages.value = Math.max(1, Math.ceil(result.total / limit.value))
  } catch (e) {
    console.error('Failed to load tree:', e)
  } finally {
    loading.value = false
  }
}

async function toggleNode(nodeId: number) {
  if (expandedNodes.value.has(nodeId)) {
    expandedNodes.value.delete(nodeId)
    expandedNodes.value = new Set(expandedNodes.value)
    return
  }

  expandedNodes.value.add(nodeId)
  expandedNodes.value = new Set(expandedNodes.value)

  if (!childrenMap.value.has(nodeId)) {
    loadingChildren.value.add(nodeId)
    loadingChildren.value = new Set(loadingChildren.value)
    try {
      const children = await issueApi.getIssueChildren(nodeId)
      childrenMap.value.set(nodeId, children)
    } catch (e) {
      console.error('Failed to load children:', e)
    } finally {
      loadingChildren.value.delete(nodeId)
      loadingChildren.value = new Set(loadingChildren.value)
    }
  }
}

function expandAll() {
  const newExpanded = new Set(expandedNodes.value)
  for (const node of rootNodes.value) {
    if (node.has_children) newExpanded.add(node.id)
  }
  childrenMap.value.forEach((children) => {
    for (const child of children) {
      if (child.has_children) newExpanded.add(child.id)
    }
  })
  expandedNodes.value = newExpanded

  for (const id of expandedNodes.value) {
    if (!childrenMap.value.has(id)) {
      loadChildrenSilent(id)
    }
  }
}

function collapseAll() {
  expandedNodes.value = new Set()
}

async function handleCreateChild(payload: { parentId: number; name: string; priority: string }) {
  try {
    // Use first available state as default, fallback to 1
    const defaultStateId = states.value.length > 0 ? states.value[0]?.id : 1
    await issueApi.createIssue(props.projectId, props.workspaceId, {
      name: payload.name,
      priority: payload.priority as any,
      state_id: defaultStateId,
      parent_id: payload.parentId
    })

    const updateNodeInList = (list: TreeIssueResponse[]) => {
      for (const node of list) {
        if (node.id === payload.parentId) {
          node.sub_issues_count = (node.sub_issues_count || 0) + 1
          node.has_children = true
          return
        }
      }
    }
    updateNodeInList(rootNodes.value)

    if (expandedNodes.value.has(payload.parentId)) {
      const children = await issueApi.getIssueChildren(payload.parentId)
      childrenMap.value.set(payload.parentId, children)
    } else {
      expandedNodes.value.add(payload.parentId)
      expandedNodes.value = new Set(expandedNodes.value)
      const children = await issueApi.getIssueChildren(payload.parentId)
      childrenMap.value.set(payload.parentId, children)
    }

    showToast(t('issueTreeView.childCreated'))
  } catch (e) {
    console.error('Failed to create child issue:', e)
    showToast(t('issueTreeView.createFailed'))
  }
}

async function loadChildrenSilent(nodeId: number) {
  try {
    const children = await issueApi.getIssueChildren(nodeId)
    childrenMap.value.set(nodeId, children)
  } catch { /* */ }
}

async function loadProjectInfo() {
  try {
    const project = await projectApi.getProject(props.projectId)
    projectIdentifier.value = project.identifier || 'PROJ'
  } catch { /* */ }
}

// ── Lifecycle ──
onMounted(() => {
  Promise.all([loadRootNodes(), loadProjectInfo(), loadStates(), loadMembers()])
})

watch(page, () => {
  loadRootNodes()
})
watch(() => props.rql, () => {
  page.value = 1
  loadRootNodes()
})

// ── Batch actions ──
function toggleSelect(id: number) {
  const s = new Set(selectedIds.value); s.has(id) ? s.delete(id) : s.add(id); selectedIds.value = s
}

async function batchChangeState(stateId: number) {
  showBatchState.value = false
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { state_id: stateId })
    clearSelection()
    showToast(t('issueTreeView.stateUpdated'))
    reloadTree()
  } catch (e) { console.error('Batch state failed:', e) }
}

async function batchChangePriority(priority: string) {
  showBatchPriority.value = false
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { priority: priority as any })
    clearSelection()
    showToast(t('issueTreeView.priorityUpdated'))
    reloadTree()
  } catch (e) { console.error('Batch priority failed:', e) }
}

async function batchAssign(userId: string | number | undefined) {
  showBatchAssign.value = false
  if (!userId) return
  const uid = typeof userId === 'string' ? Number(userId) : userId
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { assignee_ids: [uid] })
    clearSelection()
    showToast(t('issueTreeView.assigned'))
    reloadTree()
  } catch (e) { console.error('Batch assign failed:', e) }
}

function clearSelection() {
  selectedIds.value = new Set()
  showBatchState.value = false
  showBatchPriority.value = false
  showBatchAssign.value = false
}

async function execBatchDelete() {
  if (!(await confirm(t('issueTreeView.confirmDelete', { count: selectedIds.value.size })))) return
  try {
    await issueApi.bulkDeleteIssues([...selectedIds.value])
    clearSelection()
    showToast(t('issueTreeView.deleted'))
    reloadTree()
  } catch (e) { console.error('Batch delete failed:', e) }
}

function reloadTree() {
  childrenMap.value = new Map()
  expandedNodes.value = new Set()
  loadRootNodes()
}

async function loadStates() {
  try {
    const r = await api.get(`/projects/${props.projectId}/settings/states`)
    const raw = r.data
    states.value = Array.isArray(raw) ? raw : (raw?.data ?? [])
  } catch (e) { /* */ }
}
async function loadMembers() {
  try { const r = await api.get(`/workspaces/${props.workspaceId}/members`); members.value = r.data } catch (e) { /* */ }
}
</script>

<style scoped>
.issue-tree-view {
  min-height: 400px;
}
</style>
