<template>
  <div class="issue-list bg-white rounded-lg border border-gray-200">
    <!-- 工具栏 -->
    <div class="border-b border-gray-200">
      <div class="px-4 py-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3 flex-1">
            <div class="relative flex-1 max-w-md">
              <input
                v-model="filters.search" type="text"
                placeholder="搜索工作项...（名称 / 编号）"
                class="w-full pl-8 pr-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                @keydown.enter="search"
              />
              <svg class="w-4 h-4 text-gray-400 absolute left-2.5 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <select v-model="filters.state_id" @change="search" class="px-3 py-1.5 border border-gray-300 rounded-md text-sm">
              <option value="0">所有状态</option>
              <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
            <select v-model="filters.priority" @change="search" class="px-3 py-1.5 border border-gray-300 rounded-md text-sm">
              <option value="">所有优先级</option>
              <option value="urgent">紧急</option><option value="high">高</option><option value="medium">中</option><option value="low">低</option><option value="none">无</option>
            </select>
          </div>
          <div class="flex items-center space-x-2 ml-3">
            <!-- 列配置 -->
            <div class="relative">
              <button @click="showColumns = !showColumns"
                class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
                :class="{ 'bg-gray-100': showColumns }">
                列配置
                <svg class="w-3 h-3 inline ml-1" :class="{ 'rotate-180': showColumns }" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
              </button>
              <div v-if="showColumns" class="absolute right-0 top-full mt-1 w-44 bg-white border border-gray-200 rounded-lg shadow-lg z-20 py-1">
                <div class="px-3 py-1.5 text-xs text-gray-400 font-medium border-b border-gray-100">显示列</div>
                <label v-for="col in effectiveColumns" :key="col.key"
                  class="flex items-center px-3 py-1.5 hover:bg-gray-50 cursor-pointer text-sm">
                  <input type="checkbox" :checked="visibleColumnKeys.has(col.key)" @change="toggleColumn(col.key)" class="rounded border-gray-300 mr-2" />
                  {{ col.label }}
                </label>
              </div>
            </div>
            <button @click="showAdvanced = !showAdvanced" class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50">
              高级搜索
              <svg class="w-3 h-3 inline ml-1" :class="{ 'rotate-180': showAdvanced }" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
            </button>
            <button @click="goToCreate" class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700">新建</button>
          </div>
        </div>
      </div>
      <!-- 高级搜索 -->
      <div v-if="showAdvanced" class="px-4 pb-3 border-t border-gray-100 pt-3 bg-gray-50">
        <div class="grid grid-cols-4 gap-3">
          <div><label class="block text-xs text-gray-500 mb-1">周期</label>
            <select v-model="filters.cycle_id" @change="search" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm">
              <option value="0">全部</option><option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select></div>
          <div><label class="block text-xs text-gray-500 mb-1">负责人</label>
            <UserSelect
              v-model="filtersAssignee"
              :users="memberOptions"
              placeholder="全部"
              :clearable="true"
              @update:model-value="search"
            /></div>
          <div><label class="block text-xs text-gray-500 mb-1">起始日期</label>
            <input type="date" v-model="filters.start_date" @change="search" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm" /></div>
          <div><label class="block text-xs text-gray-500 mb-1">截止日期</label>
            <input type="date" v-model="filters.end_date" @change="search" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm" /></div>
        </div>
        <div v-if="customFields.length > 0" class="grid grid-cols-4 gap-3 mt-3 pt-3 border-t border-gray-200">
            <div>
              <label class="block text-xs text-gray-500 mb-1">自定义字段</label>
              <select v-model="filters.cf_field_id" @change="search" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm">
                <option :value="0">全部</option>
                <option v-for="cf in customFields" :key="cf.id" :value="cf.id">{{ cf.name }}</option>
              </select>
            </div>
            <div v-if="filters.cf_field_id > 0">
              <label class="block text-xs text-gray-500 mb-1">字段值</label>
              <input type="text" v-model="filters.cf_value" @input="search" placeholder="输入值模糊搜索..." class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm" />
            </div>
          </div>
          <div class="mt-2 flex justify-end"><button @click="resetFilters" class="text-sm text-gray-500 hover:text-indigo-600">重置筛选</button></div>
      </div>
    </div>

    <!-- 批量操作工具栏 -->
    <div v-if="selectedIds.size > 0" class="px-4 py-2 bg-indigo-50 border-b border-indigo-100 flex items-center space-x-4">
      <span class="text-sm text-indigo-700 font-medium">已选 {{ selectedIds.size }} 项</span>
      <select v-model="batchAction" class="px-2 py-1 border border-indigo-300 rounded text-sm">
        <option value="">批量操作...</option>
        <option value="state">更改状态</option><option value="priority">更改优先级</option><option value="cycle">更改周期</option><option value="delete">删除</option>
      </select>
      <template v-if="batchAction === 'state'">
        <select v-model="batchStateId" class="px-2 py-1 border border-indigo-300 rounded text-sm">
          <option :value="0">选择状态</option><option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
        <button @click="execBatch" class="px-3 py-1 bg-indigo-600 text-white text-sm rounded hover:bg-indigo-700">应用</button>
      </template>
      <template v-else-if="batchAction === 'priority'">
        <select v-model="batchPriority" class="px-2 py-1 border border-indigo-300 rounded text-sm">
          <option value="">选择优先级</option><option value="urgent">紧急</option><option value="high">高</option><option value="medium">中</option><option value="low">低</option><option value="none">无</option>
        </select>
        <button @click="execBatch" class="px-3 py-1 bg-indigo-600 text-white text-sm rounded hover:bg-indigo-700">应用</button>
      </template>
      <template v-else-if="batchAction === 'cycle'">
        <select v-model="batchCycleId" class="px-2 py-1 border border-indigo-300 rounded text-sm">
          <option :value="0">选择周期</option><option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <button @click="execBatch" class="px-3 py-1 bg-indigo-600 text-white text-sm rounded hover:bg-indigo-700">应用</button>
      </template>
      <template v-else-if="batchAction === 'delete'">
        <button @click="execBatchDelete" class="px-3 py-1 bg-red-600 text-white text-sm rounded hover:bg-red-700">确认删除</button>
      </template>
      <button @click="batchAction = ''; selectedIds.clear()" class="text-sm text-gray-500 hover:text-gray-700">取消</button>
    </div>

    <!-- 列表内容 -->
    <div class="overflow-x-auto">
      <div v-if="loading" class="text-center py-12">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
        <p class="mt-2 text-gray-500 text-sm">加载中...</p>
      </div>
      <div v-else-if="issues.length === 0" class="text-center py-16">
        <svg class="h-12 w-12 text-gray-300 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
        <p class="mt-2 text-gray-500">暂无工作项</p>
      </div>
      <table v-else class="w-full">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="w-10 px-3 py-2 text-left">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" class="rounded border-gray-300" />
            </th>
            <th v-for="col in visibleColumns" :key="col.key"
              class="px-3 py-2 text-left text-xs font-medium text-gray-500"
              :class="col.width"
            >{{ col.label }}</th>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 w-20">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="issue in issues" :key="issue.id"
            class="border-b border-gray-100 hover:bg-gray-50 cursor-pointer"
            :class="{ 'bg-indigo-50': selectedIds.has(issue.id) }">
            <td class="px-3 py-2" @click.stop>
              <input type="checkbox" :checked="selectedIds.has(issue.id)" @change="toggleSelect(issue.id)" class="rounded border-gray-300" />
            </td>
            <td v-for="col in visibleColumns" :key="col.key" class="px-3 py-2" @click="col.key !== 'actions' && $emit('select', issue)">
              <!-- 编号 -->
              <span v-if="col.key === 'sequence_id'" class="text-xs text-gray-500 font-mono">DEMO-{{ issue.sequence_id }}</span>
              <!-- 标题 -->
              <span v-else-if="col.key === 'name'" class="text-sm text-gray-800 font-medium line-clamp-2">{{ issue.name }}</span>
              <!-- 优先级 -->
              <span v-else-if="col.key === 'priority'" :class="priorityClass(issue.priority)" class="text-xs px-1.5 py-0.5 rounded whitespace-nowrap">{{ priorityLabel(issue.priority) }}</span>
              <!-- 状态 -->
              <!-- 类型 -->
              <span v-else-if="col.key === 'issue_type'" class="text-xs whitespace-nowrap">
                <span v-if="issue.issue_type" class="inline-flex items-center space-x-0.5">
                  <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: issue.issue_type.color }"></span>
                  <span>{{ issue.issue_type.name }}</span>
                </span>
                <span v-else class="text-gray-400">-</span>
              </span>
              <span v-else-if="col.key === 'state'" class="text-xs text-gray-600 whitespace-nowrap">{{ getStateName(issue.state_id) }}</span>
              <!-- 负责人 -->
              <div v-else-if="col.key === 'assignees'" class="flex -space-x-1">
                <div v-for="(a, idx) in (issue.assignees || []).slice(0, 3)" :key="a.id"
                  class="w-5 h-5 rounded-full border border-white flex items-center justify-center text-[10px] font-medium text-white"
                  :style="{ backgroundColor: assigneeColor(idx) }" :title="a.display_name || a.username">{{ getInitials(a.display_name || a.username) }}</div>
                <span v-if="!issue.assignees?.length" class="text-xs text-gray-400">-</span>
              </div>
              <!-- 周期 -->
              <span v-else-if="col.key === 'cycle'" class="text-xs text-gray-500 whitespace-nowrap">{{ getCycleName(issue) }}</span>
              <!-- 开始日期 -->
              <span v-else-if="col.key === 'start_date'" class="text-xs text-gray-500">{{ formatDate(issue.start_date) }}</span>
              <!-- 截止日期 -->
              <span v-else-if="col.key === 'target_date'" class="text-xs text-gray-500">{{ formatDate(issue.target_date) }}</span>
              <!-- 创建时间 -->
              <span v-else-if="col.key === 'created_at'" class="text-xs text-gray-400">{{ formatDate(issue.created_at) }}</span>
              <!-- 自定义字段 -->
              <span v-else-if="col.key.startsWith('cf_')" class="text-xs text-gray-600 whitespace-nowrap">{{ getCFValue(issue.id, col.key) }}</span>
            </td>
            <td class="px-3 py-2" @click.stop>
              <button @click="$emit('select', issue)" class="text-xs text-indigo-600 hover:text-indigo-800 mr-2">查看</button>
              <button @click="$emit('delete', issue)" class="text-xs text-red-500 hover:text-red-700">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div v-if="totalPages > 1" class="px-4 py-3 border-t border-gray-200 flex items-center justify-between">
      <span class="text-sm text-gray-500">共 {{ totalCount }} 项</span>
      <div class="flex items-center space-x-1">
        <button @click="page--" :disabled="page <= 1" class="px-3 py-1 border rounded text-sm disabled:opacity-50">上一页</button>
        <template v-for="p in visiblePages" :key="p">
          <button v-if="p === '...'" disabled class="px-2 py-1 text-sm text-gray-400">...</button>
          <button v-else @click="page = Number(p)" class="px-3 py-1 border rounded text-sm"
            :class="page === Number(p) ? 'bg-indigo-600 text-white border-indigo-600' : 'hover:bg-gray-50'">{{ p }}</button>
        </template>
        <button @click="page++" :disabled="page >= totalPages" class="px-3 py-1 border rounded text-sm disabled:opacity-50">下一页</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import issueApi from '@/api/issue'
import customFieldApi from '@/api/custom-field'
import api from '@/api'
import UserSelect from '@/components/UserSelect.vue'

const props = defineProps<{ projectId: number; workspaceId: number }>()
const router = useRouter()

defineEmits<{
  (e: 'select', issue: any): void
  (e: 'delete', issue: any): void
}>()

// ---- Column config ----
interface ColumnDef { key: string; label: string; width: string; defaultVisible: boolean }
const staticColumns: ColumnDef[] = [
  { key: 'sequence_id', label: '编号', width: 'w-20', defaultVisible: true },
  { key: 'name',         label: '标题', width: '',       defaultVisible: true },
  { key: 'priority',     label: '优先级', width: 'w-20', defaultVisible: true },
  { key: 'issue_type',   label: '类型', width: 'w-20', defaultVisible: true },
  { key: 'state',        label: '状态',   width: 'w-28', defaultVisible: true },
  { key: 'assignees',    label: '负责人', width: 'w-28', defaultVisible: true },
  { key: 'cycle',        label: '周期',   width: 'w-24', defaultVisible: true },
  { key: 'start_date',   label: '开始日期', width: 'w-28', defaultVisible: false },
  { key: 'target_date',  label: '截止日期', width: 'w-28', defaultVisible: false },
  { key: 'created_at',   label: '创建时间', width: 'w-36', defaultVisible: false },
]

const STORAGE_KEY = 'issuelist_columns'

const customFields = ref<any[]>([])

// Dynamic columns: static + custom fields
const effectiveColumns = computed(() => {
  const cols = [...staticColumns]
  for (const cf of customFields.value) {
    cols.push({
      key: 'cf_' + cf.id,
      label: cf.name,
      width: 'w-28',
      defaultVisible: false
    } as ColumnDef)
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

// ---- State ----
const issues = ref<any[]>([])
const states = ref<any[]>([])
const cycles = ref<any[]>([])
const members = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const limit = ref(20)
const totalCount = ref(0)
const totalPages = ref(1)
const showAdvanced = ref(false)
const selectedIds = ref(new Set<number>())

const filters = ref({
  search: '', state_id: 0, priority: '', cycle_id: 0, assignee_id: 0,
  start_date: '', end_date: '', cf_field_id: 0, cf_value: ''
})

const batchAction = ref('')
const batchStateId = ref(0)
const batchPriority = ref('')
const batchCycleId = ref(0)

// ---- Computed ----
const isAllSelected = computed(() => issues.value.length > 0 && issues.value.every(i => selectedIds.value.has(i.id)))

const memberOptions = computed(() => members.value.map(m => ({
  id: m.user_id,
  display_name: m.user?.display_name || m.user?.username,
  email: m.user?.email
})))

const filtersAssignee = computed({
  get: () => filters.value.assignee_id > 0 ? filters.value.assignee_id : undefined,
  set: (v: number | undefined) => { filters.value.assignee_id = v || 0 }
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

// ---- Helpers ----
function priorityClass(p: string) {
  const m: Record<string, string> = { urgent: 'bg-red-100 text-red-700', high: 'bg-orange-100 text-orange-700', medium: 'bg-yellow-100 text-yellow-700', low: 'bg-green-100 text-green-700', none: 'bg-gray-100 text-gray-500' }
  return m[p] || m.none
}
function priorityLabel(p: string) { const m: Record<string, string> = { urgent: '紧急', high: '高', medium: '中', low: '低', none: '无' }; return m[p] || p }
function getStateName(id: number) { return states.value.find((s: any) => s.id === id)?.name || '-' }
function getCycleName(i: any) { return i.cycle?.name || i.cycle_link?.name || '-' }
function getInitials(n: string) { return (n || '?')[0]?.toUpperCase() || '?' }
function assigneeColor(i: number) { return ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'][i % 6] }
function formatDate(d: string | null | undefined) { if (!d) return '-'; return new Date(d).toLocaleDateString('zh-CN') }

// ---- Actions ----
function goToCreate() { router.push(`/workspaces/${props.workspaceId}/projects/${props.projectId}/issues/new`) }
function search() { page.value = 1; loadIssues() }
function resetFilters() {
  filters.value = { search: '', state_id: 0, priority: '', cycle_id: 0, assignee_id: 0, start_date: '', end_date: '', cf_field_id: 0, cf_value: '' }
  search()
}
function toggleSelectAll() {
  if (isAllSelected.value) selectedIds.value.clear()
  else issues.value.forEach(i => selectedIds.value.add(i.id))
  selectedIds.value = new Set(selectedIds.value)
}
function toggleSelect(id: number) {
  const s = new Set(selectedIds.value); s.has(id) ? s.delete(id) : s.add(id); selectedIds.value = s
}
async function execBatch() {
  const ids = [...selectedIds.value]; if (!ids.length) return
  try {
    if (batchAction.value === 'state' && batchStateId.value) await Promise.all(ids.map(id => issueApi.updateIssue(id, { state_id: batchStateId.value } as any)))
    else if (batchAction.value === 'priority' && batchPriority.value) await Promise.all(ids.map(id => issueApi.updateIssue(id, { priority: batchPriority.value } as any)))
    else if (batchAction.value === 'cycle' && batchCycleId.value) await Promise.all(ids.map(id => issueApi.setIssueCycle(id, batchCycleId.value)))
    selectedIds.value = new Set(); batchAction.value = ''; loadIssues()
  } catch (e) { console.error('Batch failed:', e) }
}
async function execBatchDelete() {
  try { await issueApi.bulkDeleteIssues([...selectedIds.value]); selectedIds.value = new Set(); batchAction.value = ''; loadIssues() }
  catch (e) { console.error('Batch delete failed:', e) }
}

// ---- Data loading ----
onMounted(() => { Promise.all([loadIssues(), loadStates(), loadCycles(), loadMembers(), loadCustomFields()]) })
watch(page, () => loadIssues())

async function loadIssues() {
  loading.value = true
  try {
    const params: any = { limit: limit.value, offset: (page.value - 1) * limit.value }
    if (filters.value.state_id && filters.value.state_id > 0) params.state_id = filters.value.state_id
    if (filters.value.priority) params.priority = filters.value.priority
    if (filters.value.cycle_id && filters.value.cycle_id > 0) params.cycle_id = filters.value.cycle_id
    if (filters.value.assignee_id && filters.value.assignee_id > 0) params.assignee_id = filters.value.assignee_id
    if (filters.value.search) params.search = filters.value.search
    if (filters.value.cf_field_id && filters.value.cf_field_id > 0) {
      params.cf_field_id = filters.value.cf_field_id
      if (filters.value.cf_value) params.cf_value = filters.value.cf_value
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
}

// Custom field value cache: issueId → fieldId → display string
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

// Load CF values after issues load
watch(issues, (newIssues) => {
  if (newIssues.length) {
    loadCFValues(newIssues.map(i => i.id))
  }
})
</script>

