<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">{{ $t('issueType.workItemTypes') }}</h2>
        <p class="text-sm text-gray-500 mt-1">{{ $t('issueType.configureDesc') }}</p>
        <p class="text-xs text-gray-400 mt-1">{{ $t('issueType.importHint') }}</p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="openImportModal" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium" :disabled="importing">
          {{ importing ? $t('issueType.importing') : $t('issueType.importFromWorkspace') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-400">{{ $t('common.loading') }}</div>

    <div v-else-if="types.length === 0" class="text-center py-12 bg-gray-50 rounded-lg border border-gray-200">
      <p class="text-gray-500 mb-4">{{ $t('issueType.noTypes') }}</p>
      <div class="flex items-center justify-center gap-4">
        <button @click="openImportModal" class="text-indigo-600 hover:text-indigo-700 font-medium">{{ $t('issueType.importFromWorkspace') }} →</button>
      </div>
    </div>

    <div v-else class="space-y-4">
      <div v-for="(t, index) in types" :key="t.id" class="bg-white rounded-lg border border-gray-200 overflow-hidden" :class="{ 'opacity-60': t.is_imported }">
        <div class="flex items-center justify-between p-3 hover:bg-gray-50 transition">
          <div class="flex items-center gap-3">
            <span class="text-gray-400 text-sm w-6 text-center">{{ index + 1 }}</span>
            <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: t.color || '#6366F1' }"></div>
            <div>
              <span class="text-sm font-medium text-gray-800">{{ t.name }}</span>
              <span v-if="t.description" class="text-xs text-gray-400 ml-2">{{ t.description }}</span>
            </div>
            <span v-if="t.is_default" class="px-2 py-0.5 bg-indigo-100 text-indigo-600 rounded text-xs font-medium">Default</span>
            <span v-if="t.is_imported" class="px-2 py-0.5 bg-emerald-100 text-emerald-700 rounded text-xs font-medium">📦 {{ $t('issueType.imported') }}</span>
          </div>
          <div class="flex items-center gap-2">
            <button @click="loadTypeFields(t)" class="px-2 py-1 text-xs text-gray-500 hover:text-indigo-600 hover:bg-indigo-50 rounded border border-gray-200">
              {{ t.fields?.length || 0 }} {{ $t('issueTypeManager.fields') }}
            </button>
            <button v-if="t.is_imported" @click="handleUnimportType(t)" :disabled="unimportingId === t.id" class="px-2 py-1 text-xs text-amber-600 hover:text-amber-700 hover:bg-amber-50 rounded border border-amber-200" :title="$t('issueType.unimport')">
              {{ unimportingId === t.id ? '...' : $t('issueType.unimport') }}
            </button>
            <button v-if="!t.is_inherited" @click="moveUp(index)" :disabled="index === 0" class="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-30">↑</button>
            <button v-if="!t.is_inherited" @click="moveDown(index)" :disabled="index === types.length - 1" class="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-30">↓</button>
            <button v-if="!t.is_inherited && reorderDirty" @click="saveReorder" class="ml-2 text-xs text-indigo-600 hover:text-indigo-700 font-medium">{{ $t('issueType.saveOrder') }}</button>
          </div>
        </div>

        <div v-if="expandedType === t.id" class="border-t border-gray-200 p-4 bg-gray-50">
          <div class="flex items-center justify-between mb-3">
            <h4 class="text-sm font-medium text-gray-700">{{ $t('issueTypeManager.relatedFields') }}</h4>
            <button v-if="!t.is_inherited" @click="openFieldBindModal" class="text-xs text-indigo-600 hover:text-indigo-700 font-medium">+ {{ $t('issueTypeManager.addField') }}</button>
            <span v-else class="text-xs text-gray-400">{{ $t('issueType.imported') }} · {{ $t('issueTypeManager.relatedFields') }}</span>
          </div>

          <div v-if="!typeFieldsLoading && (typeFields.length === 0)" class="text-center py-4 text-gray-400 text-sm">
            {{ $t('issueTypeManager.noFields') }}
          </div>

          <div v-else-if="!typeFieldsLoading" class="space-y-2">
            <div v-for="(field, fIndex) in typeFields" :key="field.field_id" class="flex items-center justify-between p-2 bg-white rounded border border-gray-200">
              <div class="flex items-center gap-2">
                <span class="text-xs font-mono bg-gray-100 px-1.5 py-0.5 rounded">{{ getFieldTypeLabel(field.field_type) }}</span>
                <span class="text-sm text-gray-700">{{ field.name }}</span>
                <span v-if="field.is_required" class="text-xs text-red-600 font-medium">*</span>
              </div>
              <button v-if="!t.is_inherited" @click="removeField(t.id, field.field_id, fIndex)" class="text-xs text-red-500 hover:text-red-700">{{ $t('issueTypeManager.remove') }}</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Field bind modal -->
    <div v-if="showFieldBindModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showFieldBindModal = false">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-lg">
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-semibold">{{ $t('issueTypeManager.addFieldTo') }} {{ selectedType?.name }}</h3>
          <button @click="showFieldBindModal = false" class="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
        </div>
        <div class="p-6 max-h-80 overflow-y-auto">
          <p class="text-sm text-gray-500 mb-4">{{ $t('issueTypeManager.selectField') }}</p>
          <div v-if="availableFields.length === 0" class="text-center py-8 text-gray-400">{{ $t('issueTypeManager.noAvailableFields') }}</div>
          <div v-else class="space-y-2">
            <div v-for="field in availableFields" :key="field.id" @click="handleAddField(field)" class="flex items-center justify-between p-3 border border-gray-200 rounded-lg hover:border-indigo-300 hover:bg-indigo-50 cursor-pointer transition">
              <div class="flex items-center gap-3">
                <span class="text-xs font-mono bg-gray-100 px-1.5 py-0.5 rounded">{{ getFieldTypeLabel(field.field_type) }}</span>
                <span class="text-sm font-medium text-gray-800">{{ field.name }}</span>
              </div>
              <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Import modal (Plane v3-style: project references workspace type by link) -->
    <div v-if="showImportModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showImportModal = false">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-lg">
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-semibold">{{ $t('issueType.importTitle') }}</h3>
          <button @click="showImportModal = false" class="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
        </div>
        <div class="p-6 max-h-96 overflow-y-auto">
          <p class="text-sm text-gray-500 mb-4">{{ $t('issueType.importDesc') }}</p>
          <div v-if="importableLoading" class="text-center py-8 text-gray-400">{{ $t('common.loading') }}</div>
          <div v-else-if="importableTypes.length === 0" class="text-center py-8 text-gray-400">{{ $t('issueType.noImportable') }}</div>
          <div v-else class="space-y-2">
            <div v-for="it in importableTypes" :key="it.id" class="flex items-center justify-between p-3 border border-gray-200 rounded-lg hover:border-indigo-300 hover:bg-indigo-50 transition">
              <div class="flex items-center gap-3">
                <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: it.color || '#6366F1' }"></div>
                <div>
                  <span class="text-sm font-medium text-gray-800">{{ it.name }}</span>
                  <span v-if="it.description" class="text-xs text-gray-400 ml-2">{{ it.description }}</span>
                </div>
              </div>
              <button @click="handleImportType(it)" :disabled="importingTypeId === it.id" class="px-3 py-1 text-xs text-white bg-indigo-600 hover:bg-indigo-700 rounded font-medium disabled:opacity-50">
                {{ importingTypeId === it.id ? $t('issueType.importing') : $t('issueType.importFromWorkspace') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import api from '@/api'
import customFieldApi from '@/api/custom-field'
import issueTypeApi from '@/api/issue-type'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'

const { t: $t } = useI18n()
const toast = useToast()

const props = defineProps<{
  projectId: number
  workspaceId: number
}>()

interface IssueType {
  id: number
  name: string
  color: string
  icon: string
  description?: string
  level?: number
  is_default: boolean
  sequence: number
  is_active: boolean
  project_id?: number
  workspace_id: number
  is_imported?: boolean
  is_inherited?: boolean
  fields?: any[]
}

const types = ref<IssueType[]>([])
const loading = ref(false)
const reorderDirty = ref(false)

const expandedType = ref<number | null>(null)
const selectedType = ref<IssueType | null>(null)
const typeFields = ref<any[]>([])
const typeFieldsLoading = ref(false)
const showFieldBindModal = ref(false)
const availableFields = ref<any[]>([])

// Plane v3-style Import model state
const showImportModal = ref(false)
const importableTypes = ref<IssueType[]>([])
const importableLoading = ref(false)
const importing = ref(false)
const importingTypeId = ref<number | null>(null)
const unimportingId = ref<number | null>(null)

onMounted(() => loadTypes())
watch(() => props.projectId, () => loadTypes())

async function loadTypes() {
  loading.value = true
  try {
    const res = await api.get(`/projects/${props.projectId}/issue-types?workspace_id=${props.workspaceId}`)
    types.value = Array.isArray(res.data) ? res.data : []
    reorderDirty.value = false
  } catch (e) { console.error('Failed to load issue types:', e) }
  finally { loading.value = false }
}

// ==================== Plane v3-style Import Model ====================

async function openImportModal() {
  showImportModal.value = true
  importableLoading.value = true
  try {
    importableTypes.value = await issueTypeApi.listImportableTypes(props.workspaceId, props.projectId)
  } catch (e) {
    console.error('Failed to load importable types:', e)
    importableTypes.value = []
  } finally {
    importableLoading.value = false
  }
}

async function handleImportType(t: IssueType) {
  importingTypeId.value = t.id
  importing.value = true
  try {
    await issueTypeApi.importIssueType(props.projectId, t.id)
    toast.success($t('issueType.importSuccess'))
    // Refresh both the importable list and the main list
    await Promise.all([loadTypes(), openImportModal()])
  } catch (e: any) {
    const msg = e.response?.data?.message || $t('issueType.importFailed')
    toast.error(msg)
  } finally {
    importingTypeId.value = null
    importing.value = false
  }
}

async function handleUnimportType(t: IssueType) {
  unimportingId.value = t.id
  try {
    await issueTypeApi.unimportIssueType(props.projectId, t.id)
    toast.success($t('issueType.unimportSuccess'))
    await loadTypes()
  } catch (e: any) {
    const msg = e.response?.data?.message || $t('issueType.unimportFailed')
    toast.error(msg)
  } finally {
    unimportingId.value = null
  }
}

function moveUp(index: number) {
  if (index <= 0) return
  const arr = [...types.value]
  ;[arr[index - 1], arr[index]] = [arr[index], arr[index - 1]]
  types.value = arr
  reorderDirty.value = true
}

function moveDown(index: number) {
  if (index >= types.value.length - 1) return
  const arr = [...types.value]
  ;[arr[index], arr[index + 1]] = [arr[index + 1], arr[index]]
  types.value = arr
  reorderDirty.value = true
}

async function saveReorder() {
  try {
    await api.patch(`/projects/${props.projectId}/issue-types/reorder`, {
      type_ids: types.value.map(t => t.id),
    })
    reorderDirty.value = false
  } catch (e) { console.error('Failed to reorder:', e) }
}

async function loadTypeFields(type: IssueType) {
  if (expandedType.value === type.id) {
    expandedType.value = null
    return
  }
  expandedType.value = type.id
  selectedType.value = type
  typeFieldsLoading.value = true
  try {
    // Pass projectId so the backend applies the Import-aware logic
    // (imported types expose all attached fields without enrollment).
    typeFields.value = await issueTypeApi.getIssueTypeFields(type.id, props.projectId)
  } catch (e) { console.error('Failed to load type fields:', e); typeFields.value = [] }
  finally { typeFieldsLoading.value = false }
}

async function openFieldBindModal() {
  try {
    availableFields.value = await customFieldApi.listCustomFields(props.projectId)
  } catch (e) { console.error('Failed to load fields:', e); availableFields.value = [] }
  showFieldBindModal.value = true
}

async function handleAddField(field: any) {
  if (!selectedType.value) return
  try {
    await issueTypeApi.addFieldToIssueType(selectedType.value.id, {
      field_id: field.id,
      is_required: field.is_required || false,
    })
    showFieldBindModal.value = false
    await loadTypeFields(selectedType.value)
  } catch (e: any) {
    toast.error(e.response?.data?.message || $t('issueTypeManager.addFieldFailed'))
  }
}

async function removeField(typeId: number, fieldId: number, index: number) {
  try {
    await issueTypeApi.removeFieldFromIssueType(typeId, fieldId)
    typeFields.value.splice(index, 1)
  } catch (e) { console.error('Failed to remove field:', e) }
}

function getFieldTypeLabel(type: string): string {
  const labels: Record<string, string> = {
    text: $t('issueTypeManager.fieldTypes.text'),
    number: $t('issueTypeManager.fieldTypes.number'),
    select: $t('issueTypeManager.fieldTypes.select'),
    multi_select: $t('issueTypeManager.fieldTypes.multi_select'),
    date: $t('issueTypeManager.fieldTypes.date'),
    boolean: $t('issueTypeManager.fieldTypes.boolean'),
    url: $t('issueTypeManager.fieldTypes.url'),
    member: $t('issueTypeManager.fieldTypes.member'),
  }
  return labels[type] || type
}
</script>
