<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">Work Item Types</h2>
        <p class="text-sm text-gray-500 mt-1">Configure which work item types are available in this project</p>
      </div>
      <button @click="copyFromWorkspace" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium" :disabled="copying">
        {{ copying ? 'Copying...' : 'Copy from Workspace' }}
      </button>
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <div v-else-if="types.length === 0" class="text-center py-12 bg-gray-50 rounded-lg border border-gray-200">
      <p class="text-gray-500 mb-4">No issue types configured for this project yet.</p>
      <button @click="copyFromWorkspace" class="text-indigo-600 hover:text-indigo-700 font-medium">Copy from Workspace →</button>
    </div>

    <div v-else class="space-y-4">
      <div v-for="(t, index) in types" :key="t.id" class="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <div class="flex items-center justify-between p-3 hover:bg-gray-50 transition">
          <div class="flex items-center gap-3">
            <span class="text-gray-400 text-sm w-6 text-center">{{ index + 1 }}</span>
            <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: t.color || '#6366F1' }"></div>
            <div>
              <span class="text-sm font-medium text-gray-800">{{ t.name }}</span>
              <span v-if="t.description" class="text-xs text-gray-400 ml-2">{{ t.description }}</span>
            </div>
            <span v-if="t.is_default" class="px-2 py-0.5 bg-indigo-100 text-indigo-600 rounded text-xs font-medium">Default</span>
            <span v-if="t.project_id" class="px-2 py-0.5 bg-green-100 text-green-600 rounded text-xs">Project</span>
            <span v-else class="px-2 py-0.5 bg-gray-100 text-gray-500 rounded text-xs">Workspace</span>
          </div>
          <div class="flex items-center gap-2">
            <button @click="loadTypeFields(t)" class="px-2 py-1 text-xs text-gray-500 hover:text-indigo-600 hover:bg-indigo-50 rounded border border-gray-200">
              {{ t.fields?.length || 0 }} 字段
            </button>
            <button @click="moveUp(index)" :disabled="index === 0" class="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-30">↑</button>
            <button @click="moveDown(index)" :disabled="index === types.length - 1" class="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-30">↓</button>
            <button @click="saveReorder" v-if="reorderDirty" class="ml-2 text-xs text-indigo-600 hover:text-indigo-700 font-medium">Save Order</button>
          </div>
        </div>

        <div v-if="expandedType === t.id" class="border-t border-gray-200 p-4 bg-gray-50">
          <div class="flex items-center justify-between mb-3">
            <h4 class="text-sm font-medium text-gray-700">关联字段</h4>
            <button @click="openFieldBindModal" class="text-xs text-indigo-600 hover:text-indigo-700 font-medium">+ 添加字段</button>
          </div>

          <div v-if="!typeFieldsLoading && (typeFields.length === 0)" class="text-center py-4 text-gray-400 text-sm">
            暂无关联字段，点击上方按钮添加
          </div>

          <div v-else-if="!typeFieldsLoading" class="space-y-2">
            <div v-for="(field, fIndex) in typeFields" :key="field.field_id" class="flex items-center justify-between p-2 bg-white rounded border border-gray-200">
              <div class="flex items-center gap-2">
                <span class="text-xs font-mono bg-gray-100 px-1.5 py-0.5 rounded">{{ getFieldTypeLabel(field.field_type) }}</span>
                <span class="text-sm text-gray-700">{{ field.name }}</span>
                <span v-if="field.is_required" class="text-xs text-red-600 font-medium">*</span>
              </div>
              <button @click="removeField(t.id, field.field_id, fIndex)" class="text-xs text-red-500 hover:text-red-700">移除</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showFieldBindModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showFieldBindModal = false">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-lg">
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-semibold">添加字段到 {{ selectedType?.name }}</h3>
          <button @click="showFieldBindModal = false" class="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
        </div>
        <div class="p-6 max-h-80 overflow-y-auto">
          <p class="text-sm text-gray-500 mb-4">从项目自定义字段中选择要关联的字段</p>
          <div v-if="availableFields.length === 0" class="text-center py-8 text-gray-400">暂无可用字段，请先创建自定义字段</div>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import api from '@/api'
import customFieldApi from '@/api/custom-field'
import issueTypeApi from '@/api/issue-type'

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
  level: number
  is_default: boolean
  sequence: number
  is_active: boolean
  project_id?: number
  workspace_id: number
  fields?: any[]
}

const types = ref<IssueType[]>([])
const loading = ref(false)
const copying = ref(false)
const reorderDirty = ref(false)

const expandedType = ref<number | null>(null)
const selectedType = ref<IssueType | null>(null)
const typeFields = ref<any[]>([])
const typeFieldsLoading = ref(false)
const showFieldBindModal = ref(false)
const availableFields = ref<any[]>([])

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

async function copyFromWorkspace() {
  copying.value = true
  try {
    await api.post(`/projects/${props.projectId}/issue-types/copy-from-workspace?workspace_id=${props.workspaceId}`)
    await loadTypes()
  } catch (e: any) {
    const msg = e.response?.data?.message || 'Failed to copy'
    alert(msg)
  }
  finally { copying.value = false }
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
    typeFields.value = await issueTypeApi.getIssueTypeFields(type.id)
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
    alert(e.response?.data?.message || '添加字段失败')
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
    text: '文本',
    number: '数字',
    select: '下拉',
    multi_select: '多选',
    date: '日期',
    boolean: '布尔',
    url: '链接',
    member: '成员',
  }
  return labels[type] || type
}
</script>
