<template>
  <div class="issue-filter-bar bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700">
    <!-- Row 1: Mode + Search/RQL + Quick filters + Actions + View toggle -->
    <div class="flex items-center gap-2 px-3 py-2">
      <!-- Basic / RQL toggle -->
      <div class="inline-flex bg-gray-100 dark:bg-gray-800 rounded-lg p-0.5 shrink-0">
        <button @click="switchMode('basic')" :class="['px-2.5 py-1 text-xs rounded-md font-medium transition-colors', mode === 'basic' ? 'bg-white dark:bg-gray-700 shadow-sm text-gray-900 dark:text-gray-100' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700']">{{ t('issueList.filter') }}</button>
        <button @click="switchMode('rql')" :class="['px-2.5 py-1 text-xs rounded-md font-medium transition-colors', mode === 'rql' ? 'bg-white dark:bg-gray-700 shadow-sm text-gray-900 dark:text-gray-100' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700']">RQL</button>
      </div>

      <!-- BASIC mode -->
      <template v-if="mode === 'basic'">
        <div class="relative flex-1 max-w-md">
          <svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
          <input v-model="searchText" type="text" :placeholder="t('issueList.searchPlaceholder')"
            class="w-full pl-9 pr-3 py-1.5 border border-gray-200 dark:border-gray-700 rounded-md text-sm bg-gray-50 dark:bg-gray-800 focus:bg-white focus:ring-2 focus:ring-indigo-500" />
        </div>
        <button v-for="qf in quickFilters" :key="qf.key" @click="toggleQuick(qf)"
          :class="['px-2.5 py-1 rounded-full text-xs font-medium whitespace-nowrap transition-colors', quickActive.has(qf.key) ? 'bg-indigo-600 text-white' : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200']">
          {{ qf.label }}
        </button>
        <!-- + Add filter -->
        <div class="relative">
          <button @click="showAdd = !showAdd" class="flex items-center gap-1 px-2.5 py-1.5 text-xs text-gray-600 dark:text-gray-400 border border-gray-200 dark:border-gray-700 rounded-md hover:bg-gray-50 dark:hover:bg-gray-800" :class="{ 'bg-gray-100': showAdd }">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
            <span v-if="conditions.length > 0" class="w-4 h-4 rounded-full bg-indigo-600 text-white text-[10px] flex items-center justify-center">{{ conditions.length }}</span>
          </button>
          <div v-if="showAdd" class="absolute left-0 top-full mt-1 w-48 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-30 py-1" @click.stop>
            <button v-for="opt in addFilterOptions" :key="opt.key" @click="addCondition(opt); showAdd = false" class="w-full text-left px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-700">{{ opt.label }}</button>
          </div>
        </div>
        <!-- Search button (only shown when conditions exist) -->
        <button v-if="conditions.length > 0" @click="emitSearch" class="px-3 py-1.5 bg-indigo-600 text-white text-xs rounded-md hover:bg-indigo-700 font-medium shrink-0">{{ t('common.search') }}</button>
      </template>

      <!-- RQL mode -->
      <template v-else>
        <div class="relative flex-1">
          <svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
          <input v-model="rqlText" type="text" :placeholder="t('issueList.rqlPlaceholder')"
            class="w-full pl-9 pr-16 py-1.5 border border-gray-200 dark:border-gray-700 rounded-md text-sm font-mono bg-gray-50 dark:bg-gray-800 focus:bg-white focus:ring-2 focus:ring-indigo-500"
            @keydown.enter="emitRQL" />
          <button @click="emitRQL" class="absolute right-1.5 top-1/2 -translate-y-1/2 px-2.5 py-0.5 bg-indigo-600 text-white text-xs rounded hover:bg-indigo-700 font-medium">{{ t('common.search') }}</button>
        </div>
      </template>

      <div class="flex-1" />
      <slot name="actions" />
      <!-- View toggle -->
      <div class="inline-flex bg-gray-100 dark:bg-gray-800 rounded-lg p-0.5 shrink-0">
        <button v-for="v in viewOptions" :key="v.id" @click="$emit('update:viewMode', v.id)"
          :class="['px-2.5 py-1 text-xs rounded-md transition-colors', viewMode === v.id ? 'bg-white dark:bg-gray-700 shadow-sm text-gray-900 dark:text-gray-100 font-medium' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700']"
          :title="v.label">{{ v.icon }}</button>
      </div>
    </div>

    <!-- Row 2: Filter conditions (chips) + Inline editor -->
    <div v-if="conditions.length > 0 || editingCondition" class="flex items-center gap-1.5 px-3 pb-2 flex-wrap">
      <!-- Existing condition chips -->
      <span v-for="(c, i) in conditions" :key="i"
        class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-indigo-50 dark:bg-indigo-900/20 text-indigo-700 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800 cursor-pointer hover:bg-indigo-100 dark:hover:bg-indigo-900/40"
        @click="editCondition(i)">
        <span class="text-gray-400">{{ c.label }}</span>
        <span class="text-gray-500">{{ c.operator }}</span>
        <span class="font-medium">{{ c.displayValue }}</span>
        <button @click.stop="removeCondition(i)" class="ml-0.5 text-gray-400 hover:text-red-500">×</button>
      </span>
      <button v-if="conditions.length > 0" @click="clearAll" class="text-[11px] text-gray-400 hover:text-indigo-600 ml-1">{{ t('issueList.clearFilters') }}</button>

      <!-- Inline editor for new/editing condition -->
      <template v-if="editingCondition">
        <span class="text-xs text-gray-300 dark:text-gray-600">|</span>
        <span class="text-xs text-gray-500">{{ editingCondition.label }}</span>
        <!-- Operator picker -->
        <select v-model="editingCondition.operator" class="px-1.5 py-0.5 border border-gray-200 dark:border-gray-600 rounded text-xs bg-white dark:bg-gray-800 dark:text-gray-200">
          <option v-for="op in getOperators(editingCondition.key)" :key="op.value" :value="op.value">{{ op.label }}</option>
        </select>
        <!-- Value picker by field type -->
        <select v-if="editingCondition.key === 'state_id'" v-model="editingCondition.value" class="px-2 py-0.5 border border-gray-200 dark:border-gray-600 rounded text-xs bg-white dark:bg-gray-800 dark:text-gray-200">
          <option value="">{{ t('common.select') }}</option>
          <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
        <select v-else-if="editingCondition.key === 'priority'" v-model="editingCondition.value" class="px-2 py-0.5 border border-gray-200 dark:border-gray-600 rounded text-xs bg-white dark:bg-gray-800 dark:text-gray-200">
          <option value="">{{ t('common.select') }}</option>
          <option value="urgent">{{ t('issue.priorityUrgent') }}</option><option value="high">{{ t('issue.priorityHigh') }}</option><option value="medium">{{ t('issue.priorityMedium') }}</option><option value="low">{{ t('issue.priorityLow') }}</option><option value="none">{{ t('issue.priorityNone') }}</option>
        </select>
        <select v-else-if="editingCondition.key === 'cycle_id'" v-model="editingCondition.value" class="px-2 py-0.5 border border-gray-200 dark:border-gray-600 rounded text-xs bg-white dark:bg-gray-800 dark:text-gray-200">
          <option value="">{{ t('common.select') }}</option>
          <option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <input v-else-if="editingCondition.key === 'start_date' || editingCondition.key === 'target_date'" type="date" v-model="editingCondition.value" class="px-2 py-0.5 border border-gray-200 dark:border-gray-600 rounded text-xs bg-white dark:bg-gray-800 dark:text-gray-200" />
        <input v-else type="text" v-model="editingCondition.value" :placeholder="t('issueList.inputValue')" class="px-2 py-0.5 border border-gray-200 dark:border-gray-600 rounded text-xs bg-white dark:bg-gray-800 dark:text-gray-200 w-32"
          @keydown.enter="confirmEdit" />
        <button @click="confirmEdit" class="px-2 py-0.5 bg-indigo-600 text-white text-xs rounded hover:bg-indigo-700">{{ t('common.confirm') }}</button>
        <button @click="editingCondition = null" class="text-xs text-gray-400 hover:text-gray-600">{{ t('common.cancel') }}</button>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useI18n } from '@/composables/useI18n'

interface FilterCondition {
  key: string; label: string; operator: string; value: any; displayValue: string
}

const props = defineProps<{
  viewMode: string
  states: any[]; cycles: any[]; labels: any[]
}>()

const emit = defineEmits<{
  (e: 'update:viewMode', mode: string): void
  (e: 'search', query: string, conditions: FilterCondition[]): void
  (e: 'rql', query: string): void
}>()

const { t } = useI18n()

const mode = ref<'basic' | 'rql'>('basic')
const searchText = ref('')
const rqlText = ref('')
const showAdd = ref(false)
const conditions = ref<FilterCondition[]>([])
const editingCondition = ref<FilterCondition | null>(null)
const quickActive = ref(new Set<string>())

// --- Quick filters ---
const quickFilters = [
  { key: 'mine', label: t('issueList.quickFilterMine'), params: { assignee_id: 'me' } },
  { key: 'unassigned', label: t('issueList.quickFilterUnassigned'), params: { assignee_id: null } },
  { key: 'high_priority', label: t('issueList.quickFilterHighPriority'), params: { priority: 'high' } },
  { key: 'today', label: t('issueList.quickFilterToday'), params: { created_date: 'today' } },
  { key: 'due_soon', label: t('issueList.quickFilterDueSoon'), params: { target_date: 'week' } },
]

function toggleQuick(qf: any) {
  if (quickActive.value.has(qf.key)) { quickActive.value.delete(qf.key) } else { quickActive.value.add(qf.key) }
  quickActive.value = new Set(quickActive.value)
  // Quick filters apply immediately
  emitSearch()
}

// --- Add filter options ---
const addFilterOptions = [
  { key: 'state_id', label: t('issue.state') },
  { key: 'priority', label: t('issue.priority') },
  { key: 'assignee_id', label: t('issue.assignee') },
  { key: 'cycle_id', label: t('issue.cycle') },
  { key: 'start_date', label: t('issue.startDate') },
  { key: 'target_date', label: t('issue.targetDate') },
]

// --- Operators by field type (Jira-style) ---
function getOperators(key: string): { value: string; label: string }[] {
  if (key === 'state_id' || key === 'cycle_id' || key === 'assignee_id') {
    return [
      { value: '=', label: '=' },
      { value: '!=', label: '!=' },
      { value: 'in', label: 'in' },
      { value: 'is_empty', label: t('issueList.opEmpty') },
    ]
  }
  if (key === 'priority') {
    return [
      { value: '=', label: '=' },
      { value: '!=', label: '!=' },
      { value: 'in', label: 'in' },
    ]
  }
  if (key === 'start_date' || key === 'target_date') {
    return [
      { value: '=', label: '=' },
      { value: '!=', label: '!=' },
      { value: '>=', label: '>=' },
      { value: '<=', label: '<=' },
      { value: '>', label: '>' },
      { value: '<', label: '<' },
      { value: 'is_empty', label: t('issueList.opEmpty') },
    ]
  }
  // Text/keyword fields
  return [
    { value: '~', label: '~ (' + t('issueList.opContains') + ')' },
    { value: '!~', label: '!~ (' + t('issueList.opNotContains') + ')' },
    { value: '=', label: '=' },
    { value: '!=', label: '!=' },
    { value: 'is_empty', label: t('issueList.opEmpty') },
    { value: 'is_not_empty', label: t('issueList.opNotEmpty') },
  ]
}

// --- Condition management ---
function addCondition(opt: any) {
  const op = getOperators(opt.key)[0]
  editingCondition.value = reactive({
    key: opt.key, label: opt.label,
    operator: op.value, value: '', displayValue: '',
  })
}

function editCondition(index: number) {
  const c = conditions.value[index]
  conditions.value.splice(index, 1)
  editingCondition.value = reactive({ ...c })
}

function confirmEdit() {
  if (!editingCondition.value) return
  const c = editingCondition.value
  if (!c.value && c.operator !== 'is_empty' && c.operator !== 'is_not_empty') {
    editingCondition.value = null; return
  }
  // Build display value
  let displayValue = c.value
  if (c.key === 'state_id') { const s = props.states.find((x: any) => x.id === Number(c.value)); displayValue = s?.name || c.value }
  else if (c.key === 'cycle_id') { const cy = props.cycles.find((x: any) => x.id === Number(c.value)); displayValue = cy?.name || c.value }
  else if (c.operator === 'is_empty') displayValue = t('issueList.opEmpty')
  else if (c.operator === 'is_not_empty') displayValue = t('issueList.opNotEmpty')

  conditions.value.push({ ...c, displayValue: String(displayValue) })
  editingCondition.value = null
  syncRQL()
}

function removeCondition(index: number) {
  conditions.value.splice(index, 1)
  syncRQL()
}

function clearAll() {
  conditions.value = []; searchText.value = ''; rqlText.value = ''; quickActive.value = new Set()
  emitSearch()
}

// --- Search ---
function emitSearch() {
  syncRQL()
  emit('search', searchText.value, [...conditions.value])
}

// --- RQL ↔ Filter sync ---
function syncRQL() {
  const parts: string[] = []
  for (const c of conditions.value) {
    if (c.operator === 'is_empty') { parts.push(`${c.key} IS EMPTY`); continue }
    if (c.operator === 'is_not_empty') { parts.push(`${c.key} IS NOT EMPTY`); continue }
    const val = isNaN(Number(c.value)) ? `"${c.value}"` : c.value
    parts.push(`${c.key} ${c.operator} ${val}`)
  }
  if (searchText.value) parts.push(`name ~ "${searchText.value}"`)
  rqlText.value = parts.join(' AND ')
}

function emitRQL() {
  if (!rqlText.value.trim()) return
  emit('rql', rqlText.value)
  mode.value = 'basic'
}

function switchMode(newMode: 'basic' | 'rql') {
  if (newMode === 'rql') { syncRQL() }
  else { /* parseRQL() — would parse rqlText back to conditions */ }
  mode.value = newMode
}

// View options
const viewOptions = [
  { id: 'list' as const, icon: '📋', label: '列表' },
  { id: 'kanban' as const, icon: '📌', label: '看板' },
  { id: 'tree' as const, icon: '🌳', label: '树形' },
  { id: 'calendar' as const, icon: '📅', label: '日历' },
  { id: 'gantt' as const, icon: '📊', label: '甘特' },
]
</script>
