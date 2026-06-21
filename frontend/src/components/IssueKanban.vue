<template>
  <div class="issue-kanban">
    <!-- 搜索栏 -->
    <div class="bg-white rounded-lg border border-gray-200 mb-4">
      <div class="px-4 py-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3 flex-1">
            <div class="relative flex-1 max-w-md">
              <input
                v-model="filters.search"
                type="text"
                placeholder="搜索工作项...（名称 / 编号）"
                class="w-full pl-8 pr-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                @keydown.enter="reload"
              />
              <svg class="w-4 h-4 text-gray-400 absolute left-2.5 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <select v-model="filters.state_id" @change="reload" class="px-3 py-1.5 border border-gray-300 rounded-md text-sm">
              <option value="0">所有状态</option>
              <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
            <select v-model="filters.priority" @change="reload" class="px-3 py-1.5 border border-gray-300 rounded-md text-sm">
              <option value="">所有优先级</option>
              <option value="urgent">紧急</option><option value="high">高</option><option value="medium">中</option><option value="low">低</option><option value="none">无</option>
            </select>
          </div>
          <div class="flex items-center space-x-2 ml-3">
            <button @click="showAdvanced = !showAdvanced" class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50">
              高级搜索
              <svg class="w-3 h-3 inline ml-1" :class="{ 'rotate-180': showAdvanced }" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
            </button>
          </div>
        </div>
      </div>
      <!-- 高级搜索 -->
      <div v-if="showAdvanced" class="px-4 pb-3 border-t border-gray-100 pt-3 bg-gray-50">
        <div class="grid grid-cols-2 gap-3">
          <div><label class="block text-xs text-gray-500 mb-1">周期</label>
            <select v-model="filters.cycle_id" @change="reload" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm">
              <option value="0">全部</option>
              <option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select></div>
          <div><label class="block text-xs text-gray-500 mb-1">负责人</label>
            <UserSelect
              v-model="filtersAssignee"
              :users="memberOptions"
              placeholder="全部"
              :clearable="true"
              @update:model-value="reload"
            /></div>
          <div><label class="block text-xs text-gray-500 mb-1">开始日期</label>
            <input v-model="filters.filter_start_date" type="date" @change="reload" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm" /></div>
          <div><label class="block text-xs text-gray-500 mb-1">截止日期</label>
            <input v-model="filters.filter_target_date" type="date" @change="reload" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm" /></div>
        </div>
          <div v-if="customFields.length > 0" class="grid grid-cols-2 gap-3 mt-3 pt-3 border-t border-gray-200">
            <div>
              <label class="block text-xs text-gray-500 mb-1">自定义字段</label>
              <select v-model="filters.cf_field_id" @change="reload" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm">
                <option :value="0">全部</option>
                <option v-for="cf in customFields" :key="cf.id" :value="cf.id">{{ cf.name }}</option>
              </select>
            </div>
            <div v-if="filters.cf_field_id > 0">
              <label class="block text-xs text-gray-500 mb-1">字段值</label>
              <input type="text" v-model="filters.cf_value" @input="reload" placeholder="输入值模糊搜索..." class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm" />
            </div>
          </div>
        <div class="mt-2 flex justify-end"><button @click="resetFilters" class="text-sm text-gray-500 hover:text-indigo-600">重置筛选</button></div>
      </div>
    </div>

    <!-- 看板列 -->
    <div v-if="loading" class="text-center py-12 text-gray-400 text-sm">加载中...</div>
    <div v-else class="grid gap-4" :style="gridStyle">
      <div
        v-for="state in states"
        :key="state.id"
        class="bg-gray-100 rounded-lg p-3 min-h-[200px]"
      >
        <!-- 列头 -->
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full" :style="{ backgroundColor: state.color || '#6366f1' }"></span>
            <h3 class="text-sm font-medium text-gray-700">{{ state.name }}</h3>
          </div>
          <span class="text-xs bg-gray-300 text-gray-600 px-1.5 py-0.5 rounded-full">{{ groupedIssues[state.id]?.length || 0 }}</span>
        </div>

        <!-- 卡片列表 -->
        <div class="space-y-2">
          <div
            v-for="issue in groupedIssues[state.id] || []"
            :key="issue.id"
            @click="$emit('select', issue)"
            class="bg-white rounded-md border border-gray-200 p-2.5 cursor-pointer hover:shadow-md hover:border-indigo-300 transition-shadow"
            draggable="true"
            @dragstart="onDragStart($event, issue)"
          >
            <div class="flex items-start justify-between">
              <span class="text-xs text-gray-400 font-mono">DEMO-{{ issue.sequence_id }}</span>
              <span :class="priorityDotClass(issue.priority)" class="w-1.5 h-1.5 rounded-full inline-block" :title="priorityLabel(issue.priority)"></span>
            </div>
            <p class="text-sm text-gray-800 mt-1 leading-snug line-clamp-2">{{ issue.name }}</p>
            <div class="mt-2 flex items-center justify-between">
              <div class="flex -space-x-1">
                <div
                  v-for="(a, idx) in (issue.assignees || []).slice(0, 2)"
                  :key="a.id"
                  class="w-5 h-5 rounded-full border border-white flex items-center justify-center text-[10px] font-medium text-white"
                  :style="{ backgroundColor: assigneeColor(idx) }"
                  :title="a.display_name || a.username"
                >{{ getInitials(a.display_name || a.username) }}</div>
              </div>
              <span v-if="issue.cycle" class="text-[10px] text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">{{ issue.cycle.name }}</span>
              <button @click.stop="$emit('select', issue)" class="text-[10px] text-indigo-500 hover:text-indigo-700 font-medium">详情</button>
            </div>
          </div>

          <!-- 空列 placeholder -->
          <div v-if="!groupedIssues[state.id]?.length" class="text-center text-xs text-gray-400 py-6">
            拖放工作项到此处
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import issueApi from '@/api/issue'
import customFieldApi from '@/api/custom-field'
import api from '@/api'
import UserSelect from '@/components/UserSelect.vue'

const props = defineProps<{ projectId: number; workspaceId: number }>()
defineEmits<{ (e: 'select', issue: any): void }>()

const issues = ref<any[]>([])
const states = ref<any[]>([])
const cycles = ref<any[]>([])
const members = ref<any[]>([])
const loading = ref(false)
const showAdvanced = ref(false)

const customFields = ref<any[]>([])
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
  const map: Record<number, any[]> = {}
  states.value.forEach(s => { map[s.id] = [] })
  issues.value.forEach(i => {
    if (map[i.state_id]) map[i.state_id].push(i)
  })
  return map
})

const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${states.value.length}, minmax(260px, 1fr))`
}))

function priorityDotClass(p: string) {
  const m: Record<string, string> = { urgent: 'bg-red-500', high: 'bg-orange-500', medium: 'bg-yellow-500', low: 'bg-green-500', none: 'bg-gray-400' }
  return m[p] || m.none
}
function priorityLabel(p: string) {
  const m: Record<string, string> = { urgent: '紧急', high: '高', medium: '中', low: '低', none: '无' }
  return m[p] || p
}
function getInitials(n: string) { return (n || '?')[0]?.toUpperCase() || '?' }
function assigneeColor(i: number) { return ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'][i % 5] }

function onDragStart(e: DragEvent, issue: any) {
  e.dataTransfer?.setData('text/plain', String(issue.id))
}

function reload() { loadIssues() }
function resetFilters() {
  filters.value = { search: '', state_id: 0, priority: '', cycle_id: 0, assignee_id: 0, filter_start_date: '', filter_target_date: '', cf_field_id: 0, cf_value: '' }
  reload()
}

onMounted(() => Promise.all([loadIssues(), loadStates(), loadCycles(), loadMembers(), loadCustomFields()]))
async function loadCustomFields() {
  try { customFields.value = await customFieldApi.listCustomFields(props.workspaceId, props.projectId) } catch { /* */ }
}

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
    if (filters.value.cf_field_id && filters.value.cf_field_id > 0) {
      params.cf_field_id = filters.value.cf_field_id
      if (filters.value.cf_value) params.cf_value = filters.value.cf_value
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
</script>
```

