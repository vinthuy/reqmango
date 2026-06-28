<template>
  <div class="filter-bar space-y-2">
    <!-- ── Row 1: Filters button + chips + actions ── -->
    <div class="flex items-center gap-2 flex-wrap">
      <!-- Filters button with dropdown -->
      <div class="relative" ref="fieldDropdownRef">
        <button
          @click="showFieldDropdown = !showFieldDropdown"
          class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md border transition-colors"
          :class="activeFilterCount > 0
            ? 'bg-indigo-50 border-indigo-200 text-indigo-700 hover:bg-indigo-100 dark:bg-indigo-900/30 dark:border-indigo-800 dark:text-indigo-300'
            : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50 dark:bg-gray-800 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-750'"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
          </svg>
          <span>{{ t('filter.filterButton') }}</span>
          <span v-if="activeFilterCount > 0" class="inline-flex items-center justify-center w-4 h-4 rounded-full bg-indigo-600 text-white text-[10px] font-bold">{{ activeFilterCount }}</span>
        </button>

        <!-- Field dropdown -->
        <div v-if="showFieldDropdown" class="absolute top-full left-0 mt-1 w-52 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-40 py-1 max-h-64 overflow-y-auto" @click.stop>
          <template v-for="field in FILTER_FIELDS" :key="field.key">
            <button
              @click="startEditing(field); showFieldDropdown = false"
              class="w-full text-left px-3 py-1.5 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-750 transition flex items-center gap-2"
            >
              <span class="text-gray-400 w-3.5 text-center shrink-0">{{ getFieldIcon(field.key) }}</span>
              <span>{{ t(field.labelKey) }}</span>
            </button>
          </template>
        </div>
      </div>

      <!-- ── Filter chips ── -->
      <template v-for="(f, idx) in filterState.filters" :key="idx">
        <!-- Inline editor if this chip is being edited -->
        <div v-if="editingIndex === idx" class="flex items-center gap-1 animate-fadeIn" ref="editorRef">
          <span class="text-xs font-medium text-gray-500">{{ t(FILTER_FIELDS.find(fi => fi.key === f.field)?.labelKey || '') }}</span>
          <!-- Operator selector -->
          <select
            v-model="editOp"
            class="text-xs border border-gray-200 dark:border-gray-600 rounded px-1.5 py-1 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300"
          >
            <option v-for="op in currentField?.operators || []" :key="op" :value="op">{{ t('filter.op' + op.replace(/\s/g, '').replace(/^./, s => s.toUpperCase()).replace(/[^a-zA-Z]/g, '')) }}</option>
          </select>
          <!-- Value input -->
          <template v-if="!NO_VALUE_OPERATORS.includes(editOp)">
            <!-- date_range type -->
            <template v-if="DATE_RANGE_OPERATORS.includes(editOp)">
              <input type="date" v-model="editValStart" class="text-xs border border-gray-200 rounded px-1.5 py-1 bg-white dark:bg-gray-800" />
              <span class="text-xs text-gray-400">to</span>
              <input type="date" v-model="editValEnd" class="text-xs border border-gray-200 rounded px-1.5 py-1 bg-white dark:bg-gray-800" />
            </template>
            <!-- date type -->
            <template v-else-if="currentField?.type === 'date'">
              <input type="date" v-model="editVal" class="text-xs border border-gray-200 rounded px-1.5 py-1 bg-white dark:bg-gray-800" />
            </template>
            <!-- select / multi type -->
            <template v-else-if="currentField?.type === 'select' || currentField?.type === 'multi'">
              <select
                v-if="!MULTI_VALUE_OPERATORS.includes(editOp)"
                v-model="editVal"
                class="text-xs border border-gray-200 rounded px-1.5 py-1 bg-white dark:bg-gray-800 max-w-36"
              >
                <option value="">{{ t('filter.selectValue') }}</option>
                <option v-for="opt in currentOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
              <!-- multi select placeholder - simplified to single for now -->
              <select
                v-else
                v-model="editVal"
                class="text-xs border border-gray-200 rounded px-1.5 py-1 bg-white dark:bg-gray-800 max-w-36"
              >
                <option value="">{{ t('filter.selectValue') }}</option>
                <option v-for="opt in currentOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </template>
            <!-- text type -->
            <template v-else>
              <input
                v-model="editVal"
                type="text"
                :placeholder="t('filter.enterValue')"
                class="text-xs border border-gray-200 rounded px-1.5 py-1 bg-white dark:bg-gray-800 w-32"
              />
            </template>
          </template>
          <button @click="applyEdit(idx)" class="p-0.5 text-green-600 hover:bg-green-50 rounded" title="Apply">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>
          </button>
          <button @click="cancelEdit()" class="p-0.5 text-gray-400 hover:bg-gray-100 rounded" title="Cancel">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
          </button>
        </div>

        <!-- Chip display -->
        <span
          v-else
          @click="startEditChip(idx)"
          class="inline-flex items-center gap-1 px-2 py-1 text-xs rounded-md cursor-pointer border transition-colors bg-indigo-50 border-indigo-200 text-indigo-700 hover:bg-indigo-100 dark:bg-indigo-900/30 dark:border-indigo-800 dark:text-indigo-300 group"
        >
          <span class="font-medium">{{ t(FILTER_FIELDS.find(fi => fi.key === f.field)?.labelKey || f.field) }}</span>
          <span class="text-indigo-400 dark:text-indigo-500">{{ getOperatorLabel(f.operator) }}</span>
          <span v-if="f.displayValue" class="text-indigo-600 dark:text-indigo-400 truncate max-w-[120px]">{{ f.displayValue }}</span>
          <button @click.stop="filterInstance.removeFilter(idx)" class="ml-0.5 p-0.5 rounded hover:bg-indigo-200 dark:hover:bg-indigo-800 opacity-0 group-hover:opacity-100 transition-opacity">
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
          </button>
        </span>
      </template>

      <!-- Add filter button (shown when not editing) -->
      <button
        v-if="editingIndex === null && activeFilterCount > 0"
        @click="openFieldDropdown"
        class="px-2 py-1 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition"
        :title="t('filter.addFilter')"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
      </button>

      <!-- Clear all -->
      <button
        v-if="activeFilterCount > 0"
        @click="clearAllFilters"
        class="px-2 py-1 text-xs text-gray-400 hover:text-red-500 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition ml-1"
      >
        {{ t('filter.clearAll') }}
      </button>
    </div>

    <!-- ── Row 2: RQL collapsible area ── -->
    <div class="relative">
      <button
        @click="showRQL = !showRQL"
        class="flex items-center gap-1 text-[11px] text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition"
      >
        <svg class="w-3 h-3" :class="{ 'rotate-90': showRQL }" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
        <span>{{ t('filter.rqlToggle') }}</span>
      </button>
      <div v-if="showRQL" class="mt-1.5 bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700 rounded-md p-2">
        <div class="flex items-start gap-2">
          <textarea
            :value="filterInstance.rql.value"
            readonly
            rows="2"
            class="flex-1 text-xs font-mono bg-transparent text-gray-600 dark:text-gray-400 resize-none outline-none"
            :placeholder="t('filter.rqlPlaceholder')"
          ></textarea>
          <div class="flex flex-col gap-1 shrink-0">
            <button
              @click="copyRQL"
              class="px-2 py-0.5 text-[10px] text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition"
              :title="t('filter.rqlCopy')"
            >
              {{ t('filter.rqlCopy') }}
            </button>
          </div>
        </div>
        <p class="text-[10px] text-gray-400 mt-1">{{ t('filter.rqlHint') }}</p>
      </div>
    </div>

    <!-- ── Row 3: View bar (view toggle + save view) ── -->
    <div class="flex items-center justify-between border-t border-gray-100 dark:border-gray-800 pt-2">
      <!-- View toggle -->
      <div class="flex items-center gap-0.5">
        <button
          v-for="vm in viewModes" :key="vm.value"
          @click="switchView(vm.value)"
          class="px-2.5 py-1 text-xs font-medium rounded transition-colors"
          :class="viewMode === vm.value
            ? 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
            : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800/50'"
          :title="vm.label"
        >
          {{ vm.label }}
        </button>
      </div>

      <!-- Save view -->
      <div class="flex items-center gap-2">
        <button
          v-if="activeFilterCount > 0 || groupBy || sortBy !== 'created_at_desc'"
          @click="showSaveDialog = true"
          class="flex items-center gap-1 px-2.5 py-1 text-xs font-medium bg-indigo-600 text-white rounded-md hover:bg-indigo-700 transition shadow-sm"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>
          <span>{{ activeViewId ? t('filter.updateView') : t('filter.saveView') }}</span>
        </button>

        <!-- Existing SavedViewSelector -->
        <slot name="viewSelector" />
      </div>
    </div>

    <!-- ── Group By + Order By row ── -->
    <div class="flex items-center gap-4 border-t border-gray-100 dark:border-gray-800 pt-2">
      <!-- Group By -->
      <div class="flex items-center gap-1.5">
        <span class="text-[11px] text-gray-400 uppercase tracking-wide">{{ t('filter.groupBy') }}</span>
        <select v-model="groupBy" @change="onGroupByChange" class="text-xs border border-gray-200 dark:border-gray-600 rounded px-1.5 py-0.5 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300">
          <option value="">{{ t('filter.groupByNone') }}</option>
          <option value="state_id">{{ t('filter.groupByState') }}</option>
          <option value="priority">{{ t('filter.groupByPriority') }}</option>
          <option value="assignee_id">{{ t('filter.groupByAssignee') }}</option>
          <option value="label">{{ t('filter.groupByLabel') }}</option>
          <option value="cycle_id">{{ t('filter.groupByCycle') }}</option>
          <option value="module_id">{{ t('filter.groupByModule') }}</option>
          <option value="type_id">{{ t('filter.groupByType') }}</option>
        </select>
      </div>

      <!-- Order By -->
      <div class="flex items-center gap-1.5">
        <span class="text-[11px] text-gray-400 uppercase tracking-wide">{{ t('filter.orderBy') }}</span>
        <select v-model="sortBy" @change="onSortChange" class="text-xs border border-gray-200 dark:border-gray-600 rounded px-1.5 py-0.5 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300">
          <option value="created_at_desc">{{ t('filter.orderLastCreated') }}</option>
          <option value="updated_at_desc">{{ t('filter.orderLastUpdated') }}</option>
          <option value="priority_desc">{{ t('filter.orderPriority') }}</option>
          <option value="start_date_asc">{{ t('filter.orderStartDate') }}</option>
          <option value="target_date_asc">{{ t('filter.orderDueDate') }}</option>
        </select>
      </div>
    </div>

    <!-- ── Save view dialog ── -->
    <teleport to="body">
      <div v-if="showSaveDialog" class="fixed inset-0 bg-black/20 z-50 flex items-center justify-center" @click.self="showSaveDialog = false">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-5 w-full max-w-sm mx-4">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">{{ activeViewId ? t('filter.updateView') : t('filter.saveView') }}</h3>
          <div class="space-y-3">
            <div>
              <label class="text-xs text-gray-500 mb-1 block">Name</label>
              <input
                v-model="saveViewName"
                type="text"
                class="w-full border border-gray-200 dark:border-gray-600 rounded-lg px-3 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
                placeholder="e.g. Active bugs"
                @keydown.enter="saveView"
              />
            </div>
          </div>
          <div class="flex justify-end gap-2 mt-4">
            <button @click="showSaveDialog = false" class="px-3 py-1.5 text-xs text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md transition">{{ t('common.cancel') }}</button>
            <button @click="saveView" class="px-3 py-1.5 text-xs font-medium bg-indigo-600 text-white rounded-md hover:bg-indigo-700 transition" :disabled="!saveViewName.trim()">{{ t('common.save') }}</button>
          </div>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { FILTER_FIELDS, NO_VALUE_OPERATORS, MULTI_VALUE_OPERATORS, DATE_RANGE_OPERATORS, type FilterCondition } from '@/types/filters'
import { useFilters, provideFilters } from '@/composables/useFilters'

const props = defineProps<{
  projectId: number
  viewMode: string
  states: any[]
  cycles: any[]
  labels: any[]
}>()

const emit = defineEmits<{
  (e: 'update:viewMode', mode: string): void
  (e: 'filtersChanged', rql: string, filters: FilterCondition[]): void
  (e: 'groupByChange', value: string): void
  (e: 'sortChange', value: string): void
  (e: 'saveView', data: { name: string; filters: FilterCondition[]; groupBy: string; sortBy: string; viewType: string }): void
}>()

const { t } = useI18n()

// ── Filter state ──
const filterInstance = useFilters(props.projectId)
provideFilters(filterInstance)

const filterState = filterInstance.state
const activeFilterCount = filterInstance.activeFilterCount

// ── UI state ──
const showFieldDropdown = ref(false)
const showRQL = ref(false)
const editingIndex = ref<number | null>(null)
const showSaveDialog = ref(false)
const saveViewName = ref('')
const activeViewId = ref<number | null>(null)
const fieldDropdownRef = ref<HTMLElement | null>(null)
const editorRef = ref<HTMLElement | null>(null)

// ── View modes ──
const viewModes = computed(() => [
  { value: 'list', label: t('project.view.list') },
  { value: 'kanban', label: t('project.view.kanban') },
  { value: 'tree', label: t('project.view.tree') },
  { value: 'calendar', label: t('project.view.calendar') },
  { value: 'gantt', label: t('project.view.gantt') },
])

function switchView(mode: string) {
  emit('update:viewMode', mode)
}

// ── Group By / Sort By ──
const groupBy = ref('')
const sortBy = ref('created_at_desc')

function onGroupByChange() {
  emit('groupByChange', groupBy.value)
}
function onSortChange() {
  emit('sortChange', sortBy.value)
}

// ── Edit state ──
const editField = ref('')
const editOp = ref('is')
const editVal = ref<any>('')
const editValStart = ref('')
const editValEnd = ref('')

const currentField = computed(() => FILTER_FIELDS.find(f => f.key === editField.value))

const currentOptions = computed(() => {
  const cf = currentField.value
  if (!cf) return []
  const key = cf.key
  if (key === 'state_id') return props.states.map((s: any) => ({ value: String(s.id), label: s.name || s.label }))
  if (key === 'state_group') return [
    { value: 'backlog', label: t('filter.stateGroupBacklog') },
    { value: 'unstarted', label: t('filter.stateGroupUnstarted') },
    { value: 'started', label: t('filter.stateGroupStarted') },
    { value: 'completed', label: t('filter.stateGroupCompleted') },
    { value: 'cancelled', label: t('filter.stateGroupCancelled') },
  ]
  if (key === 'priority') return [
    { value: 'urgent', label: t('issue.priorityUrgent') },
    { value: 'high', label: t('issue.priorityHigh') },
    { value: 'medium', label: t('issue.priorityMedium') },
    { value: 'low', label: t('issue.priorityLow') },
    { value: 'none', label: t('issue.priorityNone') },
  ]
  if (key === 'assignee_id') return [] // Loaded elsewhere
  if (key === 'cycle_id') return props.cycles.map((c: any) => ({ value: String(c.id), label: c.name }))
  if (key === 'label') return props.labels.map((l: any) => ({ value: l.name || l.label, label: l.name || l.label }))
  if (key === 'module_id') return [] // Loaded elsewhere
  if (key === 'type_id') return [] // Loaded elsewhere
  return []
})

// ── Field icons ──
function getFieldIcon(field: string): string {
  const icons: Record<string, string> = {
    title: 'Aa', state_id: '⚑', state_group: '⬡', priority: '!',
    assignee_id: '👤', label: '#', cycle_id: '↻', module_id: '□',
    type_id: 'T', start_date: '▶', target_date: '■', created_at: '◉',
  }
  return icons[field] || '·'
}

// ── Operator label ──
function getOperatorLabel(op: string): string {
  const key = 'op' + op.replace(/\s/g, '').replace(/^./, s => s.toUpperCase()).replace(/[^a-zA-Z]/g, '')
  return t('filter.' + key as any) || op
}

// ── Start editing (new filter) ──
function startEditing(field: typeof FILTER_FIELDS[0]) {
  editingIndex.value = -1 // -1 = new filter
  editField.value = field.key
  editOp.value = field.operators[0] || 'is'
  editVal.value = ''
  editValStart.value = ''
  editValEnd.value = ''
}

function startEditChip(idx: number) {
  const f = filterState.filters[idx]
  editingIndex.value = idx
  editField.value = f.field
  editOp.value = f.operator
  if (DATE_RANGE_OPERATORS.includes(f.operator) && Array.isArray(f.value)) {
    editValStart.value = f.value[0] || ''
    editValEnd.value = f.value[1] || ''
    editVal.value = ''
  } else {
    editVal.value = f.value ?? ''
    editValStart.value = ''
    editValEnd.value = ''
  }
}

function openFieldDropdown() {
  showFieldDropdown.value = true
}

function applyEdit(idx: number) {
  const field = editField.value
  if (!field) { cancelEdit(); return }

  // Build display label
  let displayValue = ''
  let actualValue = editVal.value
  if (DATE_RANGE_OPERATORS.includes(editOp.value)) {
    actualValue = [editValStart.value, editValEnd.value]
    displayValue = `${editValStart.value} .. ${editValEnd.value}`
  } else if (!NO_VALUE_OPERATORS.includes(editOp.value)) {
    const opt = currentOptions.value.find((o: any) => o.value === String(editVal.value))
    displayValue = opt?.label || String(editVal.value)
    if (actualValue === '') return // Require a value for non-empty operators
  }

  const condition: FilterCondition = {
    field,
    operator: editOp.value,
    value: actualValue,
    displayValue,
  }

  if (idx === -1) {
    filterInstance.addFilter(condition)
  } else {
    filterInstance.updateFilter(idx, condition)
  }
  cancelEdit()
  emitFiltersChanged()
}

function cancelEdit() {
  editingIndex.value = null
  editField.value = ''
  editOp.value = 'is'
  editVal.value = ''
  editValStart.value = ''
  editValEnd.value = ''
}

function clearAllFilters() {
  filterInstance.clearAll()
  emitFiltersChanged()
}

function emitFiltersChanged() {
  emit('filtersChanged', filterInstance.rql.value, filterInstance.toJSON())
}

// ── RQL ──
function copyRQL() {
  navigator.clipboard?.writeText(filterInstance.rql.value)
}

// ── Save view ──
async function saveView() {
  const name = saveViewName.value.trim()
  if (!name) return
  emit('saveView', {
    name,
    filters: filterInstance.toJSON(),
    groupBy: groupBy.value,
    sortBy: sortBy.value,
    viewType: props.viewMode,
  })
  saveViewName.value = ''
  showSaveDialog.value = false
}

// ── Restore from SavedView ──
function restoreFromView(view: any) {
  activeViewId.value = view.id || null
  if (view.filters && Array.isArray(view.filters) && view.filters.length > 0) {
    filterInstance.setFilters(view.filters)
  }
  if (view.group_by) {
    groupBy.value = view.group_by
  }
  if (view.sort_config && view.sort_config.length > 0) {
    const sc = view.sort_config[0]
    sortBy.value = `${sc.field}_${sc.dir}`
  }
  if (view.view_type) {
    switchView(view.view_type)
  }
  emitFiltersChanged()
}

// ── Click outside ──
function handleClickOutside(e: MouseEvent) {
  if (showFieldDropdown.value && fieldDropdownRef.value && !fieldDropdownRef.value.contains(e.target as Node)) {
    showFieldDropdown.value = false
  }
  if (editingIndex.value !== null && editorRef.value && !editorRef.value.contains(e.target as Node)) {
    // Don't close inline editor on outside click — user must confirm/cancel
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))

// ── Watch rql changes → emit ──
watch(() => filterState.filters.length, () => {
  emitFiltersChanged()
}, { deep: true })

// Expose for parent
defineExpose({ restoreFromView, filterInstance, groupBy, sortBy })
</script>

<style scoped>
.filter-bar {
  @apply px-4 py-2;
}

@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

.animate-fadeIn {
  animation: fadeIn 0.15s ease-out;
}
</style>
