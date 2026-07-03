<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { FilterCondition, FilterField, SortOption, GroupOption } from '../types/filters'
import { FILTER_FIELDS, SORT_OPTIONS, GROUP_OPTIONS, parseRQL } from '../types/filters'
import { useFilters } from '../composables/useFilters'
import SavedViewSelector from '@/components/SavedViewSelector.vue'
import type { SavedView } from '@/types/saved-view'
import SearchTemplateSelector from '@/components/SearchTemplateSelector.vue'
import type { SearchTemplate } from '@/types/search-template'
import { listCustomFields } from '@/api/custom-field'
import { suggestIssues } from '@/api/issue'
import type { CustomField } from '@/types/custom-field'
import type { IssueSearchResult } from '@/types/issue'
import api from '@/api'

const { t } = useI18n()

const props = defineProps<{
  projectId: number
  workspaceId: number
  currentView: 'list' | 'kanban' | 'tree' | 'calendar' | 'gantt'
  projectIdentifier: string
}>()

const emit = defineEmits<{
  (e: 'viewChange', view: 'list' | 'kanban' | 'tree' | 'calendar' | 'gantt'): void
  (e: 'filtersChanged', rql: string, sortBy: SortOption | null, groupBy: GroupOption | null): void
}>()

const { state, rql, isEmpty, removeFilter, clearAll, setSortBy, setGroupBy, setQuickSearch, addToHistory } = useFilters()

const showFieldDropdown = ref(false)
const editingIndex = ref<number | null>(null)
const showRQL = ref(false)
const rqlText = ref('')
const isEditingRQL = ref(false)
const showSortDropdown = ref(false)
const showGroupDropdown = ref(false)

const searchSuggestions = ref<IssueSearchResult[]>([])
const showSuggestions = ref(false)
let suggestDebounce: ReturnType<typeof setTimeout> | null = null

const states = ref<any[]>([])
const cycles = ref<any[]>([])
const members = ref<any[]>([])
const modules = ref<any[]>([])
const issueTypes = ref<any[]>([])
const labels = ref<any[]>([])
const customFields = ref<CustomField[]>([])

// Merge system fields with custom fields
const allFilterFields = computed<FilterField[]>(() => {
  const custom: FilterField[] = customFields.value.map(cf => {
    const key = `cf_${cf.id}`
    const isMulti = cf.field_type === 'dropdown' && cf.is_multi_select
    return {
      key,
      dbKey: key,
      labelKey: `__custom__${cf.name}`,
      type: cf.field_type === 'dropdown' ? (isMulti ? 'multi' : 'select')
        : cf.field_type === 'date' ? 'date'
        : cf.field_type === 'boolean' ? 'select'
        : cf.field_type === 'member' ? 'multi'
        : 'text',
      valueType: cf.field_type === 'number' ? 'number' : 'string',
      operators: buildCustomFieldOperators(cf),
    } as FilterField
  })
  return [...FILTER_FIELDS, ...custom]
})

function buildCustomFieldOperators(cf: CustomField): string[] {
  switch (cf.field_type) {
    case 'text':
      return ['contains', 'does not contain', 'is empty', 'is not empty']
    case 'number':
      return ['is', 'is not', 'is empty', 'is not empty']
    case 'dropdown':
      return ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty']
    case 'boolean':
      return ['is']
    case 'date':
      return ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty', 'is not empty']
    case 'member':
      return ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty']
    case 'url':
      return ['contains', 'does not contain', 'is empty', 'is not empty']
    default:
      return ['is', 'is not', 'contains', 'does not contain', 'is empty', 'is not empty']
  }
}

function getFieldLabel(fieldKey: string): string {
  const custom = customFields.value.find(cf => `cf_${cf.id}` === fieldKey)
  if (custom) return custom.name
  const sys = FILTER_FIELDS.find(f => f.key === fieldKey)
  if (sys) return t(sys.labelKey)
  return fieldKey
}

const priorityOptions = [
  { value: 'none', label: t('issue.priorityNone') },
  { value: 'low', label: t('issue.priorityLow') },
  { value: 'medium', label: t('issue.priorityMedium') },
  { value: 'high', label: t('issue.priorityHigh') },
  { value: 'urgent', label: t('issue.priorityUrgent') },
]

const activeFilterChips = computed(() => state.filters)

watch(() => rql.value, (newRQL: string) => {
  emit('filtersChanged', newRQL, state.sortBy, state.groupBy)
  if (!isEditingRQL.value) {
    rqlText.value = newRQL
  }
})

watch(showRQL, (show) => {
  if (show) {
    rqlText.value = rql.value
  }
})

watch(() => state.sortBy, () => {
  emit('filtersChanged', rql.value, state.sortBy, state.groupBy)
})

watch(() => state.groupBy, () => {
  emit('filtersChanged', rql.value, state.sortBy, state.groupBy)
})

async function loadStates() {
  try {
    const r = await api.get(`/projects/${props.projectId}/settings/states`)
    states.value = r.data
  } catch (e) { /* */ }
}

async function loadCycles() {
  try {
    const r = await api.get(`/projects/${props.projectId}/cycles`)
    cycles.value = r.data
  } catch (e) { /* */ }
}

async function loadMembers() {
  try {
    const r = await api.get(`/workspaces/${props.workspaceId}/members`)
    members.value = r.data
  } catch (e) { /* */ }
}

async function loadIssueTypes() {
  try {
    const r = await api.get(`/projects/${props.projectId}/settings/issue-types`)
    issueTypes.value = r.data
  } catch (e) { /* */ }
}

async function loadLabels() {
  try {
    const r = await api.get(`/projects/${props.projectId}/settings/labels`)
    labels.value = r.data
  } catch (e) { /* */ }
}

async function loadModules() {
  try {
    const r = await api.get(`/modules?project_id=${props.projectId}`)
    modules.value = r.data
  } catch (e) { /* */ }
}

async function loadCustomFields() {
  try {
    customFields.value = await listCustomFields(props.workspaceId, props.projectId)
  } catch (e) { /* */ }
}

async function fetchSuggestions(query: string) {
  if (query.length < 2) {
    searchSuggestions.value = []
    showSuggestions.value = false
    return
  }
  
  if (suggestDebounce) clearTimeout(suggestDebounce)
  
  suggestDebounce = setTimeout(async () => {
    try {
      searchSuggestions.value = await suggestIssues(props.projectId, query, 8)
      showSuggestions.value = searchSuggestions.value.length > 0
    } catch {
      searchSuggestions.value = []
      showSuggestions.value = false
    }
  }, 200)
}

function handleQuickSearchChange(value: string) {
  setQuickSearch(value)
  fetchSuggestions(value)
}

function selectSuggestion(suggestion: IssueSearchResult) {
  const query = `${suggestion.project_identifier}-${suggestion.sequence_id}`
  setQuickSearch(query)
  addToHistory(query)
  showSuggestions.value = false
}

function onSearchFocus() {
  if (state.quickSearch.length >= 2 && searchSuggestions.value.length > 0) {
    showSuggestions.value = true
  }
}

function onSearchBlur() {
  window.setTimeout(() => {
    showSuggestions.value = false
  }, 200)
}

function applyHistory(query: string) {
  setQuickSearch(query)
  addToHistory(query)
  showSuggestions.value = false
}

function toggleFieldDropdown(e: Event) {
  e.stopPropagation()
  showFieldDropdown.value = !showFieldDropdown.value
  showSortDropdown.value = false
  showGroupDropdown.value = false
  editingIndex.value = null
}

function toggleSortDropdown(e: Event) {
  e.stopPropagation()
  showSortDropdown.value = !showSortDropdown.value
  showFieldDropdown.value = false
  showGroupDropdown.value = false
}

function toggleGroupDropdown(e: Event) {
  e.stopPropagation()
  showGroupDropdown.value = !showGroupDropdown.value
  showFieldDropdown.value = false
  showSortDropdown.value = false
}

function selectField(field: FilterField) {
  showFieldDropdown.value = false
  const newCondition: FilterCondition = {
    field: field.key,
    operator: field.operators[0],
    value: '',
    displayValue: ''
  }
  state.filters.push(newCondition)
  editingIndex.value = state.filters.length - 1
}

function handleFieldChange(index: number, newFieldKey: string) {
  state.filters[index].field = newFieldKey
  const fieldDef = allFilterFields.value.find(f => f.key === newFieldKey)
  state.filters[index].operator = fieldDef?.operators[0] || 'is'
  state.filters[index].value = ''
  state.filters[index].displayValue = ''
}

function handleOperatorChange(index: number, operator: string) {
  state.filters[index].operator = operator
  state.filters[index].value = ''
  state.filters[index].displayValue = ''
}

function handleValueChange(index: number, value: any, displayValue: string) {
  state.filters[index].value = value
  state.filters[index].displayValue = displayValue
  if (value !== '' && !Array.isArray(value) && value !== null && value !== undefined) {
    editingIndex.value = null
  }
}

const dateRangeValues = ref<Record<number, { from: string; to: string }>>({})
const showDateShortcuts = ref<number | null>(null)

const dateShortcuts = [
  { label: 'filter.dateToday', fn: getTodayRange },
  { label: 'filter.dateThisWeek', fn: getThisWeekRange },
  { label: 'filter.dateThisMonth', fn: getThisMonthRange },
  { label: 'filter.dateLastWeek', fn: getLastWeekRange },
  { label: 'filter.dateLastMonth', fn: getLastMonthRange },
  { label: 'filter.dateThisYear', fn: getThisYearRange },
]

function formatDate(date: Date): string {
  return date.toISOString().split('T')[0]
}

function getTodayRange(): { from: string; to: string } {
  const today = new Date()
  return { from: formatDate(today), to: formatDate(today) }
}

function getThisWeekRange(): { from: string; to: string } {
  const now = new Date()
  const day = now.getDay() || 7
  const monday = new Date(now)
  monday.setDate(now.getDate() - day + 1)
  monday.setHours(0, 0, 0, 0)
  const sunday = new Date(monday)
  sunday.setDate(monday.getDate() + 6)
  sunday.setHours(23, 59, 59, 999)
  return { from: formatDate(monday), to: formatDate(sunday) }
}

function getThisMonthRange(): { from: string; to: string } {
  const now = new Date()
  const firstDay = new Date(now.getFullYear(), now.getMonth(), 1)
  const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0)
  return { from: formatDate(firstDay), to: formatDate(lastDay) }
}

function getLastWeekRange(): { from: string; to: string } {
  const now = new Date()
  const day = now.getDay() || 7
  const lastMonday = new Date(now)
  lastMonday.setDate(now.getDate() - day - 6)
  lastMonday.setHours(0, 0, 0, 0)
  const lastSunday = new Date(lastMonday)
  lastSunday.setDate(lastMonday.getDate() + 6)
  lastSunday.setHours(23, 59, 59, 999)
  return { from: formatDate(lastMonday), to: formatDate(lastSunday) }
}

function getLastMonthRange(): { from: string; to: string } {
  const now = new Date()
  const firstDay = new Date(now.getFullYear(), now.getMonth() - 1, 1)
  const lastDay = new Date(now.getFullYear(), now.getMonth(), 0)
  return { from: formatDate(firstDay), to: formatDate(lastDay) }
}

function getThisYearRange(): { from: string; to: string } {
  const now = new Date()
  const firstDay = new Date(now.getFullYear(), 0, 1)
  const lastDay = new Date(now.getFullYear(), 11, 31)
  return { from: formatDate(firstDay), to: formatDate(lastDay) }
}

function applyDateShortcut(index: number, shortcut: typeof dateShortcuts[0]) {
  const range = shortcut.fn()
  dateRangeValues.value[index] = range
  state.filters[index].value = [range.from, range.to]
  state.filters[index].displayValue = t(shortcut.label)
  showDateShortcuts.value = null
  editingIndex.value = null
}

function handleDateRangeFromChange(index: number, value: string) {
  if (!dateRangeValues.value[index]) {
    dateRangeValues.value[index] = { from: '', to: '' }
  }
  dateRangeValues.value[index].from = value
  checkDateRangeComplete(index)
}

function handleDateRangeToChange(index: number, value: string) {
  if (!dateRangeValues.value[index]) {
    dateRangeValues.value[index] = { from: '', to: '' }
  }
  dateRangeValues.value[index].to = value
  checkDateRangeComplete(index)
}

function checkDateRangeComplete(index: number) {
  const range = dateRangeValues.value[index]
  if (range && range.from && range.to) {
    state.filters[index].value = [range.from, range.to]
    state.filters[index].displayValue = `${range.from} - ${range.to}`
    editingIndex.value = null
  }
}

function handleMultiSelectChange(index: number, values: any[], displayValues: string[]) {
  if (values.length > 0) {
    state.filters[index].value = values
    state.filters[index].displayValue = displayValues.join(', ')
    editingIndex.value = null
  }
}

function removeCondition(index: number) {
  removeFilter(index)
  if (editingIndex.value === index) {
    editingIndex.value = null
  }
}

function handleClearAll() {
  clearAll()
  editingIndex.value = null
}

function toggleRQL() {
  showRQL.value = !showRQL.value
}

function copyRQL() {
  navigator.clipboard.writeText(rqlText.value)
}

function applyRQL() {
  if (!rqlText.value.trim()) {
    state.filters = []
    state.sortBy = null
    state.groupBy = null
    state.quickSearch = ''
    emit('filtersChanged', rql.value, state.sortBy, state.groupBy)
    return
  }

  let rqlStr = rqlText.value
  let extractedQuickSearch = ''

  // 提取 (name LIKE "%keyword%" OR description LIKE "%keyword%")（关键词快速搜索）
  // 注意：sequence_id = N 不再提取到快速搜索，而是作为筛选条件保留（与筛选区的"编号"字段一致）
  const likeMatch = rqlStr.match(/\(name\s+LIKE\s+"%(.+?)%"\s+OR\s+description\s+LIKE\s+"%(.+?)%"\)/i)
  if (likeMatch) {
    extractedQuickSearch = likeMatch[1]
    rqlStr = rqlStr.replace(/\(name\s+LIKE\s+"%.+?%"\s+OR\s+description\s+LIKE\s+"%.+?%"\)/i, '')
    // 清理残留的 AND
    rqlStr = rqlStr.replace(/\s*AND\s*AND\s*/gi, ' AND ').replace(/^\s*AND\s*/i, '').replace(/\s*AND\s*$/i, '').trim()
  }

  state.quickSearch = extractedQuickSearch

  if (rqlStr) {
    const parsed = parseRQL(rqlStr)
    state.filters = parsed.filters
    state.sortBy = parsed.sortBy || null
  } else {
    state.filters = []
    state.sortBy = null
  }
  state.groupBy = null

  emit('filtersChanged', rql.value, state.sortBy, state.groupBy)
}

function handleClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.filter-bar')) {
    showFieldDropdown.value = false
    showSortDropdown.value = false
    showGroupDropdown.value = false
  }
}

function getOperatorsForField(fieldKey: string): string[] {
  const field = allFilterFields.value.find(f => f.key === fieldKey)
  return field?.operators || []
}

function getFieldType(fieldKey: string): FilterField['type'] {
  const field = allFilterFields.value.find(f => f.key === fieldKey)
  return field?.type || 'text'
}

function getOperatorTranslation(operator: string): string {
  const camelCase = operator
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join('')
  return t(`filter.op${camelCase}`)
}

function getOptionsForField(fieldKey: string): { value: any; label: string }[] {
  // System fields
  switch (fieldKey) {
    case 'state_id':
      return states.value.map(s => ({ value: s.id, label: s.name }))
    case 'priority':
      return priorityOptions
    case 'assignee_id':
      return members.value.map(m => ({ value: m.user?.id || m.id, label: m.user?.display_name || m.display_name || m.email || 'Unknown' }))
    case 'cycle_id':
      return cycles.value.map(c => ({ value: c.id, label: c.name }))
    case 'module_id':
      return modules.value.map(m => ({ value: m.id, label: m.name }))
    case 'type_id':
      return issueTypes.value.map(it => ({ value: it.id, label: it.name }))
    case 'state_group':
      return [
        { value: 'backlog', label: t('filter.stateGroupBacklog') },
        { value: 'unstarted', label: t('filter.stateGroupUnstarted') },
        { value: 'started', label: t('filter.stateGroupStarted') },
        { value: 'completed', label: t('filter.stateGroupCompleted') },
        { value: 'cancelled', label: t('filter.stateGroupCancelled') },
      ]
    case 'label':
      return labels.value.map(l => ({ value: l.id, label: l.name }))
  }
  // Custom fields: dropdown, boolean, member
  const cfKey = fieldKey.match(/^cf_(\d+)$/)
  if (cfKey) {
    const cf = customFields.value.find(c => c.id === parseInt(cfKey[1]))
    if (cf) {
      if (cf.field_type === 'boolean') {
        return [
          { value: 'true', label: t('common.yes') || 'Yes' },
          { value: 'false', label: t('common.no') || 'No' },
        ]
      }
      if (cf.field_type === 'dropdown' && cf.options) {
        return cf.options.filter(o => o.is_active).map(o => ({ value: o.id, label: o.value }))
      }
      if (cf.field_type === 'member') {
        return members.value.map(m => ({ value: m.user?.id || m.id, label: m.user?.display_name || m.display_name || m.email || 'Unknown' }))
      }
    }
  }
  return []
}

function selectSortOption(option: SortOption) {
  showSortDropdown.value = false
  if (state.sortBy?.key === option.key) {
    state.sortBy.direction = state.sortBy.direction === 'asc' ? 'desc' : 'asc'
  } else {
    setSortBy({ ...option })
  }
}

function selectGroupOption(option: GroupOption) {
  showGroupDropdown.value = false
  if (option.key === 'none') {
    setGroupBy(null)
  } else {
    setGroupBy(option)
  }
}

function handleSavedViewSelect(view: SavedView) {
  // Apply view type
  if (view.view_type && ['list', 'kanban', 'tree', 'gantt', 'calendar'].includes(view.view_type)) {
    emit('viewChange', view.view_type)
  }
  // Apply RQL from saved view (new field takes precedence)
  if (view.rql) {
    const parsed = parseRQL(view.rql)
    state.filters = parsed.filters
    state.sortBy = parsed.sortBy || null
  } else if (view.filters) {
    const f = view.filters as any
    if (f.rql) {
      const parsed = parseRQL(f.rql)
      state.filters = parsed.filters
      state.sortBy = parsed.sortBy || null
    } else if (Array.isArray(f)) {
      state.filters = f
    }
  }
  // Apply sort config
  if (view.sort_config && Array.isArray(view.sort_config) && view.sort_config.length > 0) {
    const sc = view.sort_config[0]
    setSortBy({ key: sc.field, direction: sc.dir, labelKey: '' })
  } else {
    state.sortBy = null
  }
  // Apply group by
  if (view.group_by) {
    const groupOpt = GROUP_OPTIONS.find(o => o.key === view.group_by)
    if (groupOpt) {
      setGroupBy(groupOpt)
    }
  } else {
    setGroupBy(null)
  }
}

function handleSearchTemplateApply(template: SearchTemplate) {
  if (template.view_type && ['list', 'kanban', 'tree', 'gantt', 'calendar'].includes(template.view_type)) {
    emit('viewChange', template.view_type)
  }
  if (template.rql_template) {
    const parsed = parseRQL(template.rql_template)
    state.filters = parsed.filters
    state.sortBy = parsed.sortBy || null
  }
  if (template.sort_config && Array.isArray(template.sort_config) && template.sort_config.length > 0) {
    const sc = template.sort_config[0]
    setSortBy({ key: sc.field, direction: sc.dir, labelKey: '' })
  }
  if (template.group_by) {
    const groupOpt = GROUP_OPTIONS.find(o => o.key === template.group_by)
    if (groupOpt) {
      setGroupBy(groupOpt)
    }
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  if (props.projectId > 0) {
    Promise.all([loadStates(), loadCycles(), loadMembers(), loadModules(), loadIssueTypes(), loadLabels(), loadCustomFields()])
  }
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="filter-bar border-b border-gray-200 bg-white">
    <!-- 快速搜索（独立区域） -->
    <div class="px-4 py-2 border-b border-gray-100">
      <div class="relative max-w-xl">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          :value="state.quickSearch"
          @input="handleQuickSearchChange(($event.target as HTMLInputElement).value)"
          @focus="onSearchFocus"
          @blur="onSearchBlur"
          type="text"
          :placeholder="t('filter.quickSearchPlaceholder')"
          class="w-full pl-9 pr-9 py-1.5 text-sm bg-gray-50 border border-gray-200 rounded-md outline-none focus:border-indigo-400 focus:ring-1 focus:ring-indigo-400 focus:bg-white transition-all"
        />
        <button
          v-if="state.quickSearch"
          @click="setQuickSearch('')"
          class="absolute right-3 top-1/2 -translate-y-1/2 p-0.5 hover:bg-gray-200 rounded transition-colors"
        >
          <svg class="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
        
        <div v-if="showSuggestions || state.searchHistory.length > 0" class="absolute top-full left-0 right-0 mt-1 bg-white border border-gray-200 rounded-md shadow-lg z-50 max-h-60 overflow-y-auto">
          <template v-if="searchSuggestions.length > 0">
            <div class="px-3 py-1.5 text-xs font-medium text-gray-500 border-b border-gray-100">
              {{ t('filter.suggestions') }}
            </div>
            <div
              v-for="suggestion in searchSuggestions"
              :key="suggestion.id"
              @click="selectSuggestion(suggestion)"
              class="px-3 py-2 hover:bg-indigo-50 cursor-pointer flex items-center gap-2"
            >
              <span class="text-sm font-medium text-indigo-600">{{ suggestion.project_identifier }}-{{ suggestion.sequence_id }}</span>
              <span class="text-sm text-gray-700 truncate">{{ suggestion.name }}</span>
            </div>
          </template>
          <template v-if="state.searchHistory.length > 0">
            <div v-if="searchSuggestions.length > 0" class="px-3 py-1.5 text-xs font-medium text-gray-500 border-b border-gray-100">
              {{ t('filter.searchHistory') }}
            </div>
            <div v-else class="px-3 py-1.5 text-xs font-medium text-gray-500 border-b border-gray-100">
              {{ t('filter.searchHistory') }}
            </div>
            <div
              v-for="(query, index) in state.searchHistory"
              :key="'history-' + index"
              @click="applyHistory(query)"
              class="px-3 py-2 hover:bg-indigo-50 cursor-pointer flex items-center justify-between"
            >
              <span class="text-sm text-gray-700">{{ query }}</span>
              <svg class="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- 筛选条件区域 -->
    <div class="px-4 py-2.5">
      <div class="flex items-center gap-3 flex-wrap">
        <div class="relative">
          <button
            @click="toggleFieldDropdown"
            class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
            </svg>
            <span>{{ t('filter.filterButton') }}</span>
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>

          <div v-if="showFieldDropdown" class="absolute left-0 top-full mt-1 w-64 bg-white border border-gray-200 rounded-lg shadow-lg z-50 py-1 max-h-80 overflow-y-auto">
            <div class="px-3 py-1.5 text-xs font-medium text-gray-500 border-b border-gray-100">
              {{ t('filter.selectField') }}
            </div>
            <button
              v-for="field in allFilterFields"
              :key="field.key"
              @click="selectField(field)"
              class="w-full px-3 py-1.5 text-left text-sm text-gray-700 hover:bg-gray-50"
            >
              {{ getFieldLabel(field.key) }}
            </button>
          </div>
        </div>

        <div v-for="(filter, index) in activeFilterChips" :key="index" class="flex items-center">
          <div v-if="editingIndex === index" class="flex items-center gap-1.5 bg-indigo-50 border border-indigo-200 rounded-full px-2 py-1">
            <select
              :value="filter.field"
              @change="(e) => handleFieldChange(index, (e.target as HTMLSelectElement).value)"
              class="text-sm bg-transparent border-none outline-none text-indigo-700"
            >
              <option v-for="field in allFilterFields" :key="field.key" :value="field.key">
                {{ getFieldLabel(field.key) }}
              </option>
            </select>
            <select
              :value="filter.operator"
              @change="(e) => handleOperatorChange(index, (e.target as HTMLSelectElement).value)"
              class="text-sm bg-transparent border-none outline-none text-indigo-700"
            >
              <option v-for="op in getOperatorsForField(filter.field)" :key="op" :value="op">
                {{ getOperatorTranslation(op) }}
              </option>
            </select>
            <template v-if="getFieldType(filter.field) === 'text'">
              <input
                v-model="filter.value"
                @input="handleValueChange(index, filter.value, filter.value)"
                type="text"
                :placeholder="t('filter.enterValue')"
                class="text-sm bg-white border border-indigo-300 rounded px-2 py-0.5 outline-none w-32"
              />
            </template>
            <template v-else-if="getFieldType(filter.field) === 'number'">
              <input
                v-model="filter.value"
                @input="handleValueChange(index, filter.value, filter.value)"
                type="number"
                min="1"
                step="1"
                :placeholder="t('filter.enterValue')"
                class="text-sm bg-white border border-indigo-300 rounded px-2 py-0.5 outline-none w-24"
              />
            </template>
            <template v-else-if="getFieldType(filter.field) === 'select'">
              <select
                :value="filter.value"
                @change="(e) => handleValueChange(index, (e.target as HTMLSelectElement).value, (e.target as HTMLSelectElement).options[(e.target as HTMLSelectElement).selectedIndex].text)"
                class="text-sm bg-white border border-indigo-300 rounded px-2 py-0.5 outline-none w-32"
              >
                <option value="">{{ t('filter.selectValue') }}</option>
                <option v-for="opt in getOptionsForField(filter.field)" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </template>
            <template v-else-if="getFieldType(filter.field) === 'multi'">
              <div class="relative">
                <select
                  :value="filter.value || ''"
                  multiple
                  @change="(e) => {
                    const target = e.target as HTMLSelectElement
                    const values = Array.from(target.selectedOptions).map(o => o.value)
                    const displayValues = Array.from(target.selectedOptions).map(o => o.text)
                    handleMultiSelectChange(index, values, displayValues)
                  }"
                  class="text-sm bg-white border border-indigo-300 rounded px-2 py-0.5 outline-none w-40"
                  size="4"
                >
                  <option v-for="opt in getOptionsForField(filter.field)" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
              </div>
            </template>
            <template v-else-if="getFieldType(filter.field) === 'date'">
              <template v-if="filter.operator === 'between' || filter.operator === 'not between'">
                <input
                  type="date"
                  :value="dateRangeValues[index]?.from || ''"
                  @change="(e) => handleDateRangeFromChange(index, (e.target as HTMLInputElement).value)"
                  class="text-sm bg-white border border-indigo-300 rounded px-2 py-0.5 outline-none"
                />
                <span class="text-indigo-400">-</span>
                <input
                  type="date"
                  :value="dateRangeValues[index]?.to || ''"
                  @change="(e) => handleDateRangeToChange(index, (e.target as HTMLInputElement).value)"
                  class="text-sm bg-white border border-indigo-300 rounded px-2 py-0.5 outline-none"
                />
                <div class="relative">
                  <button
                    @click.stop="showDateShortcuts = showDateShortcuts === index ? null : index"
                    class="text-xs text-indigo-500 hover:text-indigo-700 px-1"
                  >
                    {{ t('filter.dateShortcuts') }}
                  </button>
                  <div
                    v-if="showDateShortcuts === index"
                    class="absolute top-full left-0 mt-1 w-36 bg-white border border-gray-200 rounded-lg shadow-lg z-50 py-1"
                  >
                    <button
                      v-for="shortcut in dateShortcuts"
                      :key="shortcut.label"
                      @click="applyDateShortcut(index, shortcut)"
                      class="w-full px-3 py-1.5 text-left text-sm text-gray-700 hover:bg-gray-50"
                    >
                      {{ t(shortcut.label) }}
                    </button>
                  </div>
                </div>
              </template>
              <template v-else>
                <input
                  v-model="filter.value"
                  @input="handleValueChange(index, filter.value, filter.value)"
                  type="date"
                  class="text-sm bg-white border border-indigo-300 rounded px-2 py-0.5 outline-none"
                />
              </template>
            </template>
          </div>
          <div v-else class="flex items-center gap-1 bg-indigo-50 text-indigo-700 rounded-full px-3 py-1">
            <span class="text-sm">{{ getFieldLabel(filter.field) }}</span>
            <span class="text-sm">{{ getOperatorTranslation(filter.operator) }}</span>
            <span class="text-sm font-medium">{{ filter.displayValue || filter.value }}</span>
            <button
              @click="removeCondition(index)"
              class="ml-1 p-0.5 hover:bg-indigo-100 rounded-full transition-colors"
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <button
          v-if="!isEmpty"
          @click="handleClearAll"
          class="text-sm text-gray-500 hover:text-gray-700 transition-colors"
        >
          {{ t('filter.clearAll') }}
        </button>

        <button
          v-if="isEmpty || editingIndex !== null"
          @click="toggleFieldDropdown"
          class="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 transition-colors"
        >
          <span>+ {{ t('filter.addFilter') }}</span>
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        <div class="ml-auto flex items-center gap-2">
          <div v-if="['list', 'tree'].includes(props.currentView)" class="relative">
            <button
              v-if="!state.sortBy"
              @click="toggleSortDropdown"
              class="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 transition-colors"
            >
              <span>{{ t('filter.orderBy') }}</span>
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <button
              v-else
              @click="toggleSortDropdown"
              class="flex items-center gap-1 text-sm text-indigo-600 bg-indigo-50 px-2 py-1 rounded-md hover:bg-indigo-100 transition-colors"
            >
              <span>{{ t(state.sortBy.labelKey) }}</span>
              <svg class="w-3 h-3" :class="{ 'rotate-180': state.sortBy.direction === 'asc' }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <div v-if="showSortDropdown" class="absolute right-0 top-full mt-1 w-56 bg-white border border-gray-200 rounded-lg shadow-lg z-50 py-1">
              <button
                v-for="option in SORT_OPTIONS"
                :key="option.key"
                @click="selectSortOption(option)"
                class="w-full px-3 py-1.5 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center justify-between"
              >
                <span>{{ t(option.labelKey) }}</span>
                <span v-if="state.sortBy?.key === option.key" class="text-indigo-600">
                  {{ state.sortBy.direction === 'asc' ? '▲' : '▼' }}
                </span>
              </button>
            </div>
          </div>

          <div v-if="['list', 'tree'].includes(props.currentView)" class="relative">
            <button
              v-if="!state.groupBy"
              @click="toggleGroupDropdown"
              class="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 transition-colors"
            >
              <span>{{ t('filter.groupBy') }}</span>
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <button
              v-else
              @click="toggleGroupDropdown"
              class="flex items-center gap-1 text-sm text-indigo-600 bg-indigo-50 px-2 py-1 rounded-md hover:bg-indigo-100 transition-colors"
            >
              <span>{{ t(state.groupBy.labelKey) }}</span>
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <div v-if="showGroupDropdown" class="absolute right-0 top-full mt-1 w-56 bg-white border border-gray-200 rounded-lg shadow-lg z-50 py-1">
              <button
                v-for="option in GROUP_OPTIONS"
                :key="option.key"
                @click="selectGroupOption(option)"
                class="w-full px-3 py-1.5 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center justify-between"
              >
                <span>{{ t(option.labelKey) }}</span>
                <svg v-if="state.groupBy?.key === option.key" class="w-3 h-3 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
              </button>
            </div>
          </div>

          <button
            @click="toggleRQL"
            class="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 transition-colors"
          >
            <span>RQL</span>
            <svg class="w-3 h-3 transition-transform" :class="{ 'rotate-180': showRQL }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </div>
      </div>

      <div v-if="showRQL" class="mt-3 pt-3 border-t border-gray-100">
        <div class="flex items-center gap-2">
          <input
            v-model="rqlText"
            type="text"
            :placeholder="t('filter.rqlPlaceholder')"
            class="flex-1 text-sm bg-gray-50 border border-gray-200 rounded-md px-3 py-2 outline-none focus:border-indigo-400 focus:ring-1 focus:ring-indigo-400"
            @focus="isEditingRQL = true"
            @blur="isEditingRQL = false"
          />
          <button @click="copyRQL" class="px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-md transition-colors">
            {{ t('filter.rqlCopy') }}
          </button>
          <button @click="applyRQL" class="px-3 py-1.5 text-sm text-indigo-600 hover:bg-indigo-50 rounded-md transition-colors">
            {{ t('filter.rqlApply') }}
          </button>
        </div>
        <p class="mt-1 text-xs text-gray-400">{{ t('filter.rqlHint') }}</p>
      </div>
    </div>

    <div class="px-4 py-2 border-t border-gray-100 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <SavedViewSelector
          :project-id="projectId"
          :current-filters="{ rql: rql, filters: state.filters }"
          :current-rql="rql"
          :current-sort-config="state.sortBy ? [{ field: state.sortBy.key, dir: state.sortBy.direction }] : []"
          :current-group-by="state.groupBy?.key"
          :view-type="currentView"
          @select="handleSavedViewSelect"
        />
        <SearchTemplateSelector
          :project-id="projectId"
          :current-rql="rql"
          :current-view="currentView"
          @apply="handleSearchTemplateApply"
        />
      </div>
      <div class="flex items-center gap-1">
        <button
          @click="emit('viewChange', 'list')"
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md transition-colors"
          :class="currentView === 'list' ? 'bg-gray-100 text-gray-800' : 'text-gray-500 hover:bg-gray-50'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
          </svg>
          <span>{{ t('project.view.list') }}</span>
        </button>
        <button
          @click="emit('viewChange', 'kanban')"
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md transition-colors"
          :class="currentView === 'kanban' ? 'bg-gray-100 text-gray-800' : 'text-gray-500 hover:bg-gray-50'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
          </svg>
          <span>{{ t('project.view.kanban') }}</span>
        </button>
        <button
          @click="emit('viewChange', 'tree')"
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md transition-colors"
          :class="currentView === 'tree' ? 'bg-gray-100 text-gray-800' : 'text-gray-500 hover:bg-gray-50'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <span>{{ t('project.view.tree') }}</span>
        </button>
        <button
          @click="emit('viewChange', 'calendar')"
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md transition-colors"
          :class="currentView === 'calendar' ? 'bg-gray-100 text-gray-800' : 'text-gray-500 hover:bg-gray-50'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <span>{{ t('project.view.calendar') }}</span>
        </button>
        <button
          @click="emit('viewChange', 'gantt')"
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md transition-colors"
          :class="currentView === 'gantt' ? 'bg-gray-100 text-gray-800' : 'text-gray-500 hover:bg-gray-50'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          <span>{{ t('project.view.gantt') }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.filter-bar {
  position: relative;
}
.filter-bar select {
  min-width: 80px;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%234b5563' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 4px center;
  padding-right: 20px;
}
.filter-bar select:focus {
  outline: 2px solid #818cf8;
  outline-offset: -1px;
}
</style>
