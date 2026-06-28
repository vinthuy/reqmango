<template>
  <div class="fb">
    <!-- ═══ Row 1: Filters trigger + chips + add + clear ═══ -->
    <div class="fb-r1">
      <!-- Filters dropdown trigger -->
      <div class="fb-drop-ctn" ref="dropRef">
        <button
          class="fb-trigger"
          :class="{ open: showDropdown }"
          @click="showDropdown = !showDropdown"
        >
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24" class="fb-icon"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"/></svg>
          {{ t('filter.filterButton') }}
          <svg class="fb-chev" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
        </button>
        <div v-if="showDropdown" class="fb-dropdown" @click.stop>
          <button
            v-for="field in FILTER_FIELDS"
            :key="field.key"
            @click="addFilterField(field.key); showDropdown = false"
          >{{ t(field.labelKey) }}</button>
        </div>
      </div>

      <!-- Chip row -->
      <div class="fb-chip-row">
        <span
          v-for="(f, idx) in filters"
          :key="idx"
          class="fb-chip"
          @click="editFilter(idx)"
        >
          <span class="cf">{{ fieldLabel(f.field) }}</span>
          <span class="co">{{ opLabel(f.operator) }}</span>
          <span v-if="!noValueOps.includes(f.operator)" class="cv">{{ f.displayValue || f.value }}</span>
          <button class="cx" @click.stop="removeFilter(idx)">&times;</button>
        </span>
        <button class="fb-add-btn" @click="showDropdown = true">
          <svg class="fb-icon-sm" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          {{ t('filter.addFilter') }}
        </button>
      </div>

      <button
        v-if="filters.length > 0"
        class="fb-clear-btn"
        @click="clearAllFilters"
      >{{ t('filter.clearAll') }}</button>
      <button v-else class="fb-clear-btn" style="visibility:hidden">{{ t('filter.clearAll') }}</button>
    </div>

    <!-- ═══ Row 2: Inline editor (no confirm — instant update) ═══ -->
    <div v-if="editing" class="fb-editor">
      <span class="fb-ed-label">{{ fieldLabel(editField) }}</span>
      <select
        v-model="editOp"
        @change="applyEditInstant"
        class="fb-ed-sel"
      >
        <option v-for="op in currentOperators" :key="op" :value="op">{{ opLabel(op) }}</option>
      </select>

      <template v-if="!noValueOps.includes(editOp)">
        <!-- Between / Not between → dual date -->
        <template v-if="dateRangeOps.includes(editOp)">
          <input type="date" v-model="editValStart" @change="applyEditInstant" class="fb-ed-inp" />
          <span class="fb-ed-sep">&ndash;</span>
          <input type="date" v-model="editValEnd" @change="applyEditInstant" class="fb-ed-inp" />
        </template>
        <!-- Single date -->
        <template v-else-if="currentFieldDef?.type === 'date'">
          <input type="date" v-model="editVal" @change="applyEditInstant" class="fb-ed-inp" />
        </template>
        <!-- Select / Multi -->
        <template v-else-if="currentFieldDef?.type === 'select' || currentFieldDef?.type === 'multi'">
          <select v-if="!multiOps.includes(editOp)" v-model="editVal" @change="applyEditInstant" class="fb-ed-sel">
            <option value="">{{ t('filter.selectValue') }}</option>
            <option v-for="opt in currentOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>
          <select v-else v-model="editVal" @change="applyEditInstant" class="fb-ed-sel">
            <option value="">{{ t('filter.selectValue') }}</option>
            <option v-for="opt in currentOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>
        </template>
        <!-- Text -->
        <template v-else>
          <input
            v-model="editVal"
            @keydown.enter="applyEditInstant"
            @blur="applyEditInstant"
            type="text"
            :placeholder="t('filter.enterValue')"
            class="fb-ed-inp"
          />
        </template>
      </template>
      <span class="fb-ed-hint">&larr; {{ t('filter.instantHint') }}</span>
    </div>

    <!-- ═══ Row 3: RQL area (collapsible) ═══ -->
    <div class="fb-rql">
      <button class="fb-rql-tog" @click="showRQL = !showRQL">
        <svg class="fb-icon-sm" fill="none" stroke="currentColor" viewBox="0 0 24 24" :style="{ transform: showRQL ? 'rotate(90deg)' : '' }"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
        &lt;/&gt; RQL
      </button>
      <input
        v-if="showRQL"
        class="fb-rql-input"
        :value="rqlStr"
        readonly
        :placeholder="t('filter.rqlPlaceholder')"
      />
      <template v-if="showRQL">
        <button class="fb-rql-btn" @click="copyRQL">{{ t('filter.rqlCopy') }}</button>
        <button class="fb-rql-btn rql-primary" @click="applyRQL">{{ t('filter.rqlApply') }}</button>
      </template>
    </div>

    <!-- ═══ Row 4: Bottom bar — view toggle + save view ═══ -->
    <div class="fb-bot">
      <div class="fb-vt">
        <button
          v-for="vm in viewModes"
          :key="vm.value"
          :class="{ active: viewMode === vm.value }"
          @click="switchView(vm.value)"
        >{{ vm.label }}</button>
      </div>
      <div class="fb-sp"></div>
      <button class="fb-sv" @click="onSaveView">
        <svg class="fb-icon-sm" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"/></svg>
        {{ t('filter.saveView') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import {
  FILTER_FIELDS,
  NO_VALUE_OPERATORS,
  MULTI_VALUE_OPERATORS,
  DATE_RANGE_OPERATORS,
  buildRQL,
  type FilterCondition,
} from '@/types/filters'

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
  (e: 'saveView', data: { name: string; filters: FilterCondition[]; viewType: string }): void
}>()

const { t } = useI18n()

// ── Constants ──
const noValueOps = NO_VALUE_OPERATORS
const multiOps = MULTI_VALUE_OPERATORS
const dateRangeOps = DATE_RANGE_OPERATORS

// ── State ──
const filters = ref<FilterCondition[]>([])
const showDropdown = ref(false)
const showRQL = ref(false)
const editing = ref(false)
const editField = ref('')
const editOp = ref('is')
const editVal = ref<any>('')
const editValStart = ref('')
const editValEnd = ref('')
const editIdx = ref(-1) // -1 = new filter
const activeViewId = ref<number | null>(null)

const dropRef = ref<HTMLElement | null>(null)

// ── Computed ──
const rqlStr = computed(() => buildRQL(filters.value))

const currentFieldDef = computed(() => FILTER_FIELDS.find((f: any) => f.key === editField.value))

const currentOperators = computed(() => currentFieldDef.value?.operators || ['is'])

const currentOptions = computed(() => {
  const key = editField.value
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
  if (key === 'cycle_id') return props.cycles.map((c: any) => ({ value: String(c.id), label: c.name }))
  if (key === 'label') return props.labels.map((l: any) => ({ value: l.name || l.label, label: l.name || l.label }))
  return []
})

const viewModes = computed(() => [
  { value: 'list', label: t('project.view.list') },
  { value: 'kanban', label: t('project.view.kanban') },
  { value: 'tree', label: t('project.view.tree') },
  { value: 'calendar', label: t('project.view.calendar') },
  { value: 'gantt', label: t('project.view.gantt') },
])

// ── Labels ──
function fieldLabel(key: string) {
  const f = FILTER_FIELDS.find((fi: any) => fi.key === key)
  return f ? t(f.labelKey) : key
}
function opLabel(op: string) {
  const key = 'op' + op.replace(/\s/g, '').replace(/^./, (s: string) => s.toUpperCase()).replace(/[^a-zA-Z]/g, '')
  return t(('filter.' + key) as any) || op
}

// ── Add filter ──
function addFilterField(fieldKey: string) {
  const field = FILTER_FIELDS.find((f: any) => f.key === fieldKey)
  if (!field) return
  editIdx.value = -1
  editField.value = fieldKey
  editOp.value = field.operators[0] || 'is'
  editVal.value = ''
  editValStart.value = ''
  editValEnd.value = ''
  editing.value = true
}

// ── Edit existing filter ──
function editFilter(idx: number) {
  const f = filters.value[idx]
  editIdx.value = idx
  editField.value = f.field
  editOp.value = f.operator
  if (dateRangeOps.includes(f.operator) && Array.isArray(f.value)) {
    editValStart.value = f.value[0] || ''
    editValEnd.value = f.value[1] || ''
    editVal.value = ''
  } else {
    editVal.value = f.value ?? ''
    editValStart.value = ''
    editValEnd.value = ''
  }
  editing.value = true
}

// ── Apply edit (instant — no confirm) ──
function applyEditInstant() {
  const field = editField.value
  if (!field) return

  // Build display value
  let displayValue = ''
  let actualValue: any = editVal.value

  if (dateRangeOps.includes(editOp.value)) {
    if (!editValStart.value || !editValEnd.value) return // wait for both dates
    actualValue = [editValStart.value, editValEnd.value]
    displayValue = `${editValStart.value} – ${editValEnd.value}`
  } else if (!noValueOps.includes(editOp.value)) {
    const oval = editVal.value ?? ''
    if (multiOps.includes(editOp.value) && !Array.isArray(editVal.value)) {
      // For multi ops, still use string
    }
    if (oval === '' && currentFieldDef.value?.type !== 'text') return // require value
    const opt = currentOptions.value.find((o: any) => o.value === String(editVal.value))
    displayValue = opt?.label || String(editVal.value !== undefined && editVal.value !== null ? editVal.value : '')
    if (!oval && currentFieldDef.value?.type === 'text') {
      displayValue = ''
    }
  }

  const condition: FilterCondition = {
    field,
    operator: editOp.value,
    value: actualValue,
    displayValue,
  }

  if (editIdx.value === -1) {
    filters.value.push(condition)
  } else {
    filters.value[editIdx.value] = condition
  }

  editing.value = false
  emitFiltersChanged()
}

function removeFilter(idx: number) {
  filters.value.splice(idx, 1)
  editing.value = false
  emitFiltersChanged()
}

function clearAllFilters() {
  filters.value = []
  editing.value = false
  emitFiltersChanged()
}

function emitFiltersChanged() {
  emit('filtersChanged', rqlStr.value, [...filters.value])
}

// ── RQL ──
function copyRQL() {
  navigator.clipboard?.writeText(rqlStr.value)
}

function applyRQL() {
  // RQL → filters reverse sync (read-only display for now)
  emit('filtersChanged', rqlStr.value, [...filters.value])
}

// ── View ──
function switchView(mode: string) {
  emit('update:viewMode', mode)
}

// ── Save view ──
function onSaveView() {
  const name = activeViewId.value ? `Updated view ${new Date().toLocaleTimeString()}` : `View ${new Date().toLocaleTimeString()}`
  emit('saveView', {
    name,
    filters: [...filters.value],
    viewType: props.viewMode,
  })
}

// ── Restore from SavedView ──
function restoreFromView(view: any) {
  activeViewId.value = view.id || null
  if (view.filters && Array.isArray(view.filters) && view.filters.length > 0) {
    filters.value = [...view.filters]
  }
  if (view.view_type) {
    switchView(view.view_type)
  }
  emitFiltersChanged()
}

// ── Click outside ──
function onClickOutside(e: MouseEvent) {
  if (showDropdown.value && dropRef.value && !dropRef.value.contains(e.target as Node)) {
    showDropdown.value = false
  }
}

onMounted(() => document.addEventListener('click', onClickOutside))
onUnmounted(() => document.removeEventListener('click', onClickOutside))

defineExpose({ restoreFromView, filters, rqlStr })

// ── Watch rql changes ──
watch(rqlStr, () => {
  // Auto-emit on change for save view consistency
})
</script>

<style scoped>
/* ═══ Plane-Aligned FilterBar ═══ */
.fb {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  font-size: 12px;
}

/* Row 1 */
.fb-r1 {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  flex-wrap: wrap;
}

.fb-drop-ctn { position: relative; }

.fb-trigger {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 10px;
  border-radius: 6px;
  border: 1px solid #d1d5db;
  background: #fff;
  color: #374151;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all .12s;
  white-space: nowrap;
}
.fb-trigger:hover { background: #f9fafb; border-color: #9ca3af; }
.fb-trigger.open { background: #f3f4f6; border-color: #6b7280; }

.fb-icon, .fb-icon-sm { width: 14px; height: 14px; }
.fb-icon-sm { width: 11px; height: 11px; }
.fb-chev { width: 10px; height: 10px; }

.fb-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,.1);
  z-index: 60;
  width: 200px;
  overflow: hidden;
  padding: 4px;
}
.fb-dropdown button {
  display: block;
  width: 100%;
  padding: 6px 10px;
  text-align: left;
  font-size: 12px;
  color: #374151;
  background: none;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}
.fb-dropdown button:hover { background: #f3f4f6; }

/* Chip */
.fb-chip-row {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
  flex: 1;
  min-width: 0;
}
.fb-chip {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 2px 7px;
  background: #eef2ff;
  color: #4338ca;
  border: 1px solid #c7d2fe;
  border-radius: 12px;
  font-size: 11px;
  cursor: pointer;
  transition: all .12s;
  white-space: nowrap;
}
.fb-chip:hover { background: #e0e7ff; border-color: #a5b4fc; }
.fb-chip .cf { color: #6b7280; }
.fb-chip .co { color: #9ca3af; }
.fb-chip .cv { font-weight: 500; }
.fb-chip .cx {
  background: none;
  border: none;
  color: #9ca3af;
  font-size: 12px;
  cursor: pointer;
  padding: 0;
  line-height: 1;
  margin-left: 1px;
  display: flex;
  align-items: center;
}
.fb-chip .cx:hover { color: #ef4444; }

.fb-add-btn {
  padding: 4px 8px;
  border-radius: 5px;
  font-size: 11px;
  border: 1px dashed #cbd5e1;
  background: transparent;
  color: #6b7280;
  cursor: pointer;
  transition: all .12s;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 3px;
}
.fb-add-btn:hover { border-color: #3b82f6; color: #3b82f6; background: #f8faff; }

.fb-clear-btn {
  font-size: 11px;
  color: #9ca3af;
  background: none;
  border: none;
  cursor: pointer;
  padding: 3px 6px;
  white-space: nowrap;
}
.fb-clear-btn:hover { color: #ef4444; }

/* Inline editor */
.fb-editor {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  background: #f9fafb;
  border-bottom: 1px solid #eff6ff;
  flex-wrap: wrap;
}
.fb-ed-label { font-size: 11px; color: #6b7280; font-weight: 500; white-space: nowrap; }
.fb-ed-sel, .fb-ed-inp {
  padding: 3px 6px;
  border: 1px solid #d1d5db;
  border-radius: 5px;
  font-size: 11px;
  background: #fff;
  outline: none;
}
.fb-ed-sel:focus, .fb-ed-inp:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px rgba(59,130,246,.1); }
.fb-ed-sep { color: #9ca3af; font-size: 11px; padding: 0 2px; }
.fb-ed-hint { font-size: 10px; color: #6b7280; margin-left: 6px; }

/* RQL */
.fb-rql {
  border-top: 1px solid #f3f4f6;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  background: #fafbfc;
}
.fb-rql-tog {
  padding: 2px 7px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
  background: #f3f4f6;
  color: #6b7280;
  border: 1px solid #e5e7eb;
  cursor: pointer;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 3px;
}
.fb-rql-input {
  flex: 1;
  padding: 4px 6px;
  border: 1px solid #e5e7eb;
  border-radius: 5px;
  font-size: 11px;
  font-family: monospace;
  background: #f9fafb;
}
.fb-rql-btn {
  padding: 2px 7px;
  font-size: 10px;
  border-radius: 4px;
  border: 1px solid #d1d5db;
  background: #fff;
  color: #6b7280;
  cursor: pointer;
  white-space: nowrap;
}
.fb-rql-btn:hover { background: #f3f4f6; }
.fb-rql-btn.rql-primary { background: #111827; color: #fff; border-color: #111827; }
.fb-rql-btn.rql-primary:hover { background: #1f2937; }

/* Bottom bar */
.fb-bot {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-top: 1px solid #f3f4f6;
  background: #fff;
}
.fb-sp { flex: 1; }
.fb-vt {
  display: inline-flex;
  background: #f3f4f6;
  border-radius: 5px;
  padding: 1px;
}
.fb-vt button {
  padding: 3px 8px;
  font-size: 11px;
  border-radius: 4px;
  border: none;
  background: transparent;
  color: #6b7280;
  cursor: pointer;
}
.fb-vt button.active { background: #fff; color: #111827; font-weight: 500; box-shadow: 0 1px 2px rgba(0,0,0,.05); }
.fb-sv {
  padding: 3px 8px;
  font-size: 11px;
  border-radius: 5px;
  border: 1px solid #d1d5db;
  background: #fff;
  color: #6b7280;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 3px;
}
.fb-sv:hover { border-color: #3b82f6; color: #3b82f6; }
</style>
