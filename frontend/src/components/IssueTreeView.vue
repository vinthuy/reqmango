<template>
  <div class="issue-tree-view bg-white rounded-lg border border-gray-200 relative">
    <!-- Success toast -->
    <div v-if="toastMessage" class="absolute top-2 right-2 z-10 bg-green-50 border border-green-200 text-green-700 text-sm px-3 py-2 rounded-md shadow-sm transition-opacity">
      {{ toastMessage }}
    </div>

    <!-- Toolbar -->
    <div class="px-4 py-2.5 border-b border-gray-100">
      <div class="flex items-center gap-3">
        <!-- Search -->
        <div class="relative flex-1 max-w-sm">
          <svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索工作项（跨层级）..."
            class="w-full pl-9 pr-8 py-1.5 border border-gray-200 rounded-md text-sm bg-gray-50 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-colors"
            @keydown.enter="doSearch"
            @input="onSearchInput"
          />
          <button v-if="searchQuery" @click="clearSearch" class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Search mode badge -->
        <span v-if="isSearchMode" class="text-xs bg-amber-100 text-amber-700 px-2 py-0.5 rounded-full font-medium">
          搜索结果模式
        </span>

        <div class="flex-1" />

        <!-- Expand / Collapse All -->
        <button @click="expandAll" class="text-xs text-gray-500 hover:text-indigo-600 transition-colors">
          展开全部
        </button>
        <button @click="collapseAll" class="text-xs text-gray-500 hover:text-indigo-600 transition-colors">
          收起全部
        </button>

        <!-- Create -->
        <button @click="$emit('create')" class="flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 text-white text-xs rounded-md hover:bg-indigo-700 transition-colors font-medium">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          <span>新建</span>
        </button>
      </div>
    </div>

    <!-- 批量操作工具栏 -->
    <div v-if="selectedIds.size > 0" class="sticky top-0 z-20 bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-200 dark:border-indigo-800 rounded-lg mx-4 mt-3 px-4 py-2 flex items-center gap-3 flex-wrap">
      <span class="text-sm text-indigo-700 dark:text-indigo-300 font-medium">已选 {{ selectedIds.size }} 项</span>

      <div class="relative">
        <button @click="showBatchState = !showBatchState" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          更改状态
        </button>
        <div v-if="showBatchState" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 py-1 w-32">
          <button v-for="s in states" :key="s.id" @click="batchChangeState(s.id)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-200">{{ s.name }}</button>
        </div>
      </div>

      <div class="relative">
        <button @click="showBatchPriority = !showBatchPriority" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          更改优先级
        </button>
        <div v-if="showBatchPriority" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 py-1 w-32">
          <button v-for="p in priorityOptions" :key="p.value" @click="batchChangePriority(p.value)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-200">{{ p.label }}</button>
        </div>
      </div>

      <div class="relative">
        <button @click="showBatchAssign = !showBatchAssign" class="px-2.5 py-1 text-xs border border-indigo-300 dark:border-indigo-700 rounded-md bg-white dark:bg-gray-700 dark:text-gray-200 hover:bg-indigo-50 dark:hover:bg-gray-600 transition-colors">
          批量分配
        </button>
        <div v-if="showBatchAssign" class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 p-2 w-48">
          <UserSelect v-model="batchAssigneeId" :users="memberOptions" placeholder="选择负责人" @update:model-value="batchAssign" />
        </div>
      </div>

      <button @click="execBatchDelete" class="px-2.5 py-1 text-xs border border-red-300 dark:border-red-700 rounded-md bg-white dark:bg-gray-700 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors">
        批量删除
      </button>

      <button @click="clearSelection" class="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">取消选择</button>
    </div>

    <!-- Column header -->
    <div class="flex items-center px-4 py-2 bg-gray-50 dark:bg-gray-700/50 border-b border-gray-100 dark:border-gray-700 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
      <div class="w-8 shrink-0"></div>
      <div class="flex-1 min-w-0">标题</div>
      <div class="w-16 text-center shrink-0">优先级</div>
      <div class="w-20 text-center shrink-0">状态</div>
      <div class="w-16 text-center shrink-0">子项</div>
    </div>

    <!-- Tree content -->
    <div v-if="loading" class="text-center py-16">
      <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
      </svg>
      <p class="mt-2 text-gray-500 text-sm">加载中...</p>
    </div>

    <div v-else-if="searchResults.length > 0" class="divide-y divide-gray-100">
      <!-- Search results mode -->
      <div v-for="result in searchResults" :key="'sr-' + result.root_issue.id">
        <TreeNodeItem
          :node="result.root_issue"
          :depth="0"
          :expanded-nodes="expandedNodes"
          :children-map="childrenMap"
          :loading-children="loadingChildren"
          :search-matched-path="getSearchMatchedPath(result)"
          :search-matched-id="result.matched_issue.id"
          :project-identifier="projectIdentifier"
          :selected-ids="selectedIds"
          @toggle="toggleNode"
          @select="$emit('select', $event)"
          @toggle-select="toggleSelect"
          @create-child="handleCreateChild"
        />
      </div>
    </div>

    <div v-else-if="rootNodes.length === 0 && !loading" class="text-center py-16">
      <svg class="h-12 w-12 text-gray-300 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
      </svg>
      <p class="mt-2 text-gray-500">暂无工作项</p>
      <p class="mt-1 text-sm text-gray-400">点击"新建"按钮创建第一个工作项</p>
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
    <div v-if="!isSearchMode && totalPages > 1" class="px-4 py-3 border-t border-gray-200 flex items-center justify-between bg-gray-50/50">
      <span class="text-sm text-gray-500">共 {{ totalCount }} 项</span>
      <div class="flex items-center gap-1">
        <button @click="page--" :disabled="page <= 1" class="px-3 py-1 border rounded text-sm disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-100 transition-colors">上一页</button>
        <template v-for="p in visiblePages" :key="p">
          <button v-if="p === '...'" disabled class="px-2 py-1 text-sm text-gray-400">...</button>
          <button v-else @click="page = Number(p)" class="px-3 py-1 border rounded text-sm transition-colors"
            :class="page === Number(p) ? 'bg-indigo-600 text-white border-indigo-600' : 'hover:bg-gray-100'">{{ p }}</button>
        </template>
        <button @click="page++" :disabled="page >= totalPages" class="px-3 py-1 border rounded text-sm disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-100 transition-colors">下一页</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import issueApi from '@/api/issue'
import projectApi from '@/api/project'
import api from '@/api'
import type { TreeIssueResponse, TreeSearchResult } from '@/types/issue'
import TreeNodeItem from './TreeNodeItem.vue'
import UserSelect from '@/components/UserSelect.vue'
import { useConfirm } from '@/composables/useConfirm'

const props = defineProps<{ projectId: number; workspaceId: number }>()

const emit = defineEmits<{
  (e: 'select', issue: any): void
  (e: 'create'): void
}>()

// ---- State ----
const rootNodes = ref<TreeIssueResponse[]>([])
const searchResults = ref<TreeSearchResult[]>([])
const searchQuery = ref('')
const isSearchMode = ref(false)
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

// ---- Batch selection ----
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
  { value: 'urgent', label: '紧急' },
  { value: 'high', label: '高' },
  { value: 'medium', label: '中' },
  { value: 'low', label: '低' },
  { value: 'none', label: '无' },
]

const { confirm } = useConfirm()

// Debounce search timer
let searchTimer: ReturnType<typeof setTimeout> | null = null

// Toast
const toastMessage = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

function showToast(msg: string) {
  toastMessage.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMessage.value = '' }, 2500)
}

// ---- Computed ----
const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const tp = totalPages.value; const p = page.value
  if (tp <= 7) { for (let i = 1; i <= tp; i++) pages.push(i); return pages }
  pages.push(1); if (p > 3) pages.push('...')
  for (let i = Math.max(2, p - 1); i <= Math.min(tp - 1, p + 1); i++) pages.push(i)
  if (p < tp - 2) pages.push('...'); pages.push(tp)
  return pages
})

// ---- Methods ----
async function loadRootNodes() {
  loading.value = true
  try {
    const params: any = { limit: limit.value, offset: (page.value - 1) * limit.value }
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

async function doSearch() {
  const q = searchQuery.value.trim()
  if (!q) {
    clearSearch()
    return
  }

  isSearchMode.value = true
  loading.value = true
  try {
    const result = await issueApi.listTreeIssues(props.projectId, {
      search: q,
      limit: 50,
      offset: 0
    })
    // When search is provided, backend returns TreeSearchResult[]
    searchResults.value = (result.items as any) as TreeSearchResult[]
    totalCount.value = result.total

    // Auto-expand ancestor chains
    const toExpand = new Set<number>()
    for (const sr of searchResults.value) {
      if (sr.ancestor_chain && sr.ancestor_chain.length > 0) {
        toExpand.add(sr.root_issue.id)
        for (const a of sr.ancestor_chain) {
          toExpand.add(a.id)
        }
      }
    }
    // Also expand root if matched is not root
    for (const sr of searchResults.value) {
      if (sr.root_issue.id !== sr.matched_issue.id) {
        toExpand.add(sr.root_issue.id)
      }
    }

    // Load children for ancestor chain nodes
    const allAncestorIds = new Set<number>()
    for (const sr of searchResults.value) {
      allAncestorIds.add(sr.root_issue.id)
      for (const a of sr.ancestor_chain) {
        allAncestorIds.add(a.id)
      }
    }

    for (const id of allAncestorIds) {
      if (!childrenMap.value.has(id)) {
        try {
          const children = await issueApi.getIssueChildren(id)
          childrenMap.value.set(id, children)
        } catch { /* */ }
      }
    }

    expandedNodes.value = toExpand
  } catch (e) {
    console.error('Search failed:', e)
  } finally {
    loading.value = false
  }
}

function getSearchMatchedPath(result: TreeSearchResult): number[] {
  const path: number[] = [result.root_issue.id]
  for (const a of result.ancestor_chain) {
    path.push(a.id)
  }
  return path
}

function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  if (!searchQuery.value.trim()) {
    clearSearch()
    return
  }
  searchTimer = setTimeout(() => doSearch(), 400)
}

function clearSearch() {
  searchQuery.value = ''
  isSearchMode.value = false
  searchResults.value = []
  expandedNodes.value = new Set()
  childrenMap.value = new Map()
  page.value = 1
  loadRootNodes()
}

async function toggleNode(nodeId: number) {
  if (expandedNodes.value.has(nodeId)) {
    expandedNodes.value.delete(nodeId)
    expandedNodes.value = new Set(expandedNodes.value)
    return
  }

  // Expand
  expandedNodes.value.add(nodeId)
  expandedNodes.value = new Set(expandedNodes.value)

  // Lazy load children if not already loaded
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
  // Expand all root nodes
  for (const node of rootNodes.value) {
    if (node.has_children) newExpanded.add(node.id)
  }
  // Also expand children
  childrenMap.value.forEach((children) => {
    for (const child of children) {
      if (child.has_children) newExpanded.add(child.id)
    }
  })
  expandedNodes.value = newExpanded

  // Load children for all expanded nodes
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
    await issueApi.createIssue(props.projectId, props.workspaceId, {
      name: payload.name,
      priority: payload.priority as any,
      state_id: 1,
      parent_id: payload.parentId
    })

    // Update parent's sub_issues_count locally
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

    // If parent is expanded, refresh its children
    if (expandedNodes.value.has(payload.parentId)) {
      const children = await issueApi.getIssueChildren(payload.parentId)
      childrenMap.value.set(payload.parentId, children)
    } else {
      // Auto-expand parent to show the new child
      expandedNodes.value.add(payload.parentId)
      expandedNodes.value = new Set(expandedNodes.value)
      const children = await issueApi.getIssueChildren(payload.parentId)
      childrenMap.value.set(payload.parentId, children)
    }

    showToast('子工作项创建成功')
  } catch (e) {
    console.error('Failed to create child issue:', e)
    showToast('创建失败，请重试')
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

// ---- Lifecycle ----
onMounted(() => {
  Promise.all([loadRootNodes(), loadProjectInfo(), loadStates(), loadMembers()])
})

watch(page, () => {
  if (!isSearchMode.value) loadRootNodes()
})

// ---- Batch actions ----
function toggleSelect(id: number) {
  const s = new Set(selectedIds.value); s.has(id) ? s.delete(id) : s.add(id); selectedIds.value = s
}

async function batchChangeState(stateId: number) {
  showBatchState.value = false
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { state_id: stateId })
    clearSelection()
    showToast('状态已更新')
    reloadTree()
  } catch (e) { console.error('Batch state failed:', e) }
}

async function batchChangePriority(priority: string) {
  showBatchPriority.value = false
  try {
    await issueApi.bulkUpdateIssues(props.projectId, [...selectedIds.value], { priority: priority as any })
    clearSelection()
    showToast('优先级已更新')
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
    showToast('负责人已分配')
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
  if (!(await confirm(`确定要删除选中的 ${selectedIds.value.size} 个工作项吗？此操作不可撤销。`))) return
  try {
    await issueApi.bulkDeleteIssues([...selectedIds.value])
    clearSelection()
    showToast('已删除选中工作项')
    reloadTree()
  } catch (e) { console.error('Batch delete failed:', e) }
}

function reloadTree() {
  childrenMap.value = new Map()
  expandedNodes.value = new Set()
  if (isSearchMode.value) { doSearch() } else { loadRootNodes() }
}

async function loadStates() {
  try { const r = await api.get(`/projects/${props.projectId}/settings/states`); states.value = r.data } catch (e) { /* */ }
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
