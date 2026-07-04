<template>
  <div class="saved-view-selector flex items-center gap-2">
    <!-- Current view dropdown -->
    <div class="relative" v-click-outside="closeDropdown">
      <button
        @click="open = !open"
        class="flex items-center gap-2 px-3 py-1.5 text-sm border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition"
      >
        <span>📋</span>
        <span class="max-w-[120px] truncate">{{ currentViewName }}</span>
        <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      <!-- Dropdown menu -->
      <div v-if="open" class="absolute left-0 top-full mt-1 w-64 bg-white rounded-lg shadow-lg border border-gray-200 z-50 py-1">
        <div class="px-3 py-2 text-xs font-semibold text-gray-500 uppercase">{{ t('filter.savedViews') }}</div>
        <div v-if="views.length === 0" class="px-3 py-2 text-sm text-gray-400">{{ t('filter.noSavedViews') }}</div>
        <button
          v-for="view in views"
          :key="view.id"
          @click="selectView(view)"
          class="w-full flex items-center justify-between px-3 py-2 text-sm hover:bg-gray-50 transition group"
          :class="{ 'bg-indigo-50 text-indigo-700': selectedViewId === view.id }"
        >
          <span class="flex items-center gap-2">
            <span>{{ view.view_type === 'kanban' ? '📋' : '📃' }}</span>
            <span class="truncate max-w-[140px]">{{ view.name }}</span>
            <span v-if="view.is_default" class="text-xs bg-indigo-100 text-indigo-600 px-1.5 rounded">{{ t('filter.defaultViewBadge') }}</span>
          </span>
          <span class="hidden group-hover:flex items-center gap-1">
            <button @click.stop="editView(view)" :title="t('filter.editView')" class="text-gray-400 hover:text-indigo-600">✏️</button>
            <button @click.stop="setDefault(view)" :title="t('filter.setAsDefault')" class="text-gray-400 hover:text-indigo-600">⭐</button>
            <button @click.stop="duplicateView(view)" :title="t('filter.duplicateView')" class="text-gray-400 hover:text-indigo-600">📋</button>
            <button @click.stop="promptDelete(view)" :title="t('filter.delete')" class="text-gray-400 hover:text-red-500">🗑️</button>
          </span>
        </button>
        <div class="border-t border-gray-100 mt-1 pt-1">
          <button @click="saveCurrent" class="w-full flex items-center gap-2 px-3 py-2 text-sm text-indigo-600 hover:bg-indigo-50 transition">
            <span>💾</span> {{ t('filter.saveCurrentView') }}
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Save / Edit View Modal -->
  <div v-if="showSaveModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showSaveModal = false">
    <div class="bg-white rounded-xl p-6 w-full max-w-md">
      <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ editingView ? t('filter.editView') : t('filter.saveView') }}</h3>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('filter.viewNameLabel') }}</label>
          <input v-model="saveForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" :placeholder="t('filter.viewNamePlaceholder')" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('filter.viewTypeLabel') }}</label>
          <select v-model="saveForm.view_type" class="w-full px-4 py-2 border border-gray-300 rounded-lg">
            <option value="list">{{ t('project.view.list') }}</option>
            <option value="kanban">{{ t('project.view.kanban') }}</option>
            <option value="tree">{{ t('project.view.tree') }}</option>
            <option value="gantt">{{ t('project.view.gantt') }}</option>
            <option value="calendar">{{ t('project.view.calendar') }}</option>
          </select>
        </div>
        <div class="flex items-center gap-2">
          <input v-model="saveForm.is_shared" type="checkbox" id="shared" class="rounded" />
          <label for="shared" class="text-sm text-gray-700">{{ t('filter.shareWithMembers') }}</label>
        </div>
      </div>
      <div class="flex justify-end space-x-3 mt-6">
        <button @click="closeSaveModal" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('filter.cancel') }}</button>
        <button @click="doSave" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ editingView ? t('filter.update') : t('filter.save') }}</button>
      </div>
    </div>
  </div>

  <!-- Delete Confirm -->
  <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showDeleteConfirm = false">
    <div class="bg-white rounded-xl p-6 w-full max-w-sm">
      <h3 class="text-lg font-semibold text-gray-900 mb-2">{{ t('filter.deleteViewTitle') }}</h3>
      <p class="text-sm text-gray-500 mb-4">{{ t('filter.deleteViewConfirm', { name: deletingView?.name || '' }) }}</p>
      <div class="flex justify-end space-x-3">
        <button @click="showDeleteConfirm = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('filter.cancel') }}</button>
        <button @click="doDelete" class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700">{{ t('filter.delete') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import * as savedViewApi from '@/api/saved-view'
import type { SavedView, SavedViewCreate, SavedViewUpdate } from '@/types/saved-view'

const { t } = useI18n()

const props = defineProps<{
  projectId: number
  currentFilters?: Record<string, any>
  currentRQL?: string
  currentColumns?: string[]
  currentGroupBy?: string
  currentSubGroupBy?: string
  currentSortConfig?: { field: string; dir: 'asc' | 'desc' }[]
  viewType?: 'list' | 'kanban' | 'tree' | 'gantt' | 'calendar'
}>()

const emit = defineEmits<{
  (e: 'select', view: SavedView): void
  (e: 'save-request', filters: Record<string, any>, rql?: string): void
}>()

// State
const views = ref<SavedView[]>([])
const selectedViewId = ref<number | null>(null)
const open = ref(false)
const showSaveModal = ref(false)
const showDeleteConfirm = ref(false)
const deletingView = ref<SavedView | null>(null)
const editingView = ref<SavedView | null>(null)

const saveForm = ref({
  name: '',
  view_type: 'list' as 'list' | 'kanban' | 'tree' | 'gantt' | 'calendar',
  is_shared: false,
})

const currentViewName = computed(() => {
  if (selectedViewId.value) {
    const v = views.value.find(x => x.id === selectedViewId.value)
    if (v) return v.name
  }
  const def = views.value.find(v => v.is_default)
  if (def) return def.name
  return t('filter.savedViews')
})

// Load views on mount and when projectId changes
onMounted(() => loadViews())
watch(() => props.projectId, () => { loadViews(); selectedViewId.value = null })

async function loadViews() {
  try {
    views.value = await savedViewApi.listSavedViews(props.projectId)
    // Auto-select and apply default view
    const def = views.value.find(v => v.is_default)
    if (def && !selectedViewId.value) {
      selectedViewId.value = def.id
      emit('select', def)
    }
  } catch (e) { console.error('Failed to load saved views:', e) }
}

function selectView(view: SavedView) {
  selectedViewId.value = view.id
  open.value = false
  emit('select', view)
}

function saveCurrent() {
  editingView.value = null
  saveForm.value = {
    name: '',
    view_type: props.viewType || 'list',
    is_shared: false,
  }
  open.value = false
  showSaveModal.value = true
}

function editView(view: SavedView) {
  editingView.value = view
  saveForm.value = {
    name: view.name,
    view_type: view.view_type as any,
    is_shared: view.is_shared,
  }
  open.value = false
  showSaveModal.value = true
}

function closeSaveModal() {
  showSaveModal.value = false
  editingView.value = null
}

async function doSave() {
  if (!saveForm.value.name.trim()) return
  try {
    if (editingView.value) {
      // Update existing view
      const data: SavedViewUpdate = {
        name: saveForm.value.name.trim(),
        view_type: saveForm.value.view_type,
        is_shared: saveForm.value.is_shared,
      }
      const updated = await savedViewApi.updateSavedView(props.projectId, editingView.value.id, data)
      const idx = views.value.findIndex(v => v.id === editingView.value!.id)
      if (idx !== -1) {
        views.value[idx] = { ...views.value[idx], ...updated }
      }
      editingView.value = null
    } else {
      // Create new view
      const data: SavedViewCreate = {
        name: saveForm.value.name.trim(),
        view_type: saveForm.value.view_type,
        filters: props.currentFilters || {},
        rql: props.currentRQL || '',
        columns: props.currentColumns || [],
        sort_config: props.currentSortConfig || [],
        group_by: props.currentGroupBy,
        sub_group_by: props.currentSubGroupBy,
        is_shared: saveForm.value.is_shared,
      }
      const created = await savedViewApi.createSavedView(props.projectId, data)
      views.value.push(created)
      selectedViewId.value = created.id
      emit('select', created)
    }
    showSaveModal.value = false
    emit('save-request', props.currentFilters || {}, props.currentRQL || '')
  } catch (e) { console.error('Failed to save view:', e) }
}

async function setDefault(view: SavedView) {
  try {
    await savedViewApi.setDefaultView(props.projectId, view.id)
    views.value.forEach(v => { v.is_default = v.id === view.id })
  } catch (e) { console.error('Failed to set default view:', e) }
}

async function duplicateView(view: SavedView) {
  try {
    const dup = await savedViewApi.duplicateSavedView(props.projectId, view.id)
    views.value.push(dup)
  } catch (e) { console.error('Failed to duplicate view:', e) }
}

function promptDelete(view: SavedView) {
  deletingView.value = view
  showDeleteConfirm.value = true
}

async function doDelete() {
  if (!deletingView.value) return
  try {
    await savedViewApi.deleteSavedView(props.projectId, deletingView.value.id)
    if (selectedViewId.value === deletingView.value.id) selectedViewId.value = null
    views.value = views.value.filter(v => v.id !== deletingView.value!.id)
    showDeleteConfirm.value = false
    deletingView.value = null
  } catch (e) { console.error('Failed to delete view:', e) }
}

function closeDropdown() {
  open.value = false
}

// v-click-outside directive (simple implementation)
const vClickOutside = {
  mounted(el: any, binding: any) {
    el.__clickOutside__ = (e: MouseEvent) => {
      if (!el.contains(e.target)) binding.value()
    }
    document.addEventListener('click', el.__clickOutside__)
  },
  unmounted(el: any) {
    document.removeEventListener('click', el.__clickOutside__)
  },
}
</script>
