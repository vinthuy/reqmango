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

    <div v-else class="space-y-2">
      <div v-for="(t, index) in types" :key="t.id" class="flex items-center justify-between bg-white rounded-lg border border-gray-200 p-3 hover:border-gray-300 transition">
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
        <div class="flex items-center gap-1">
          <button @click="moveUp(index)" :disabled="index === 0" class="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-30">↑</button>
          <button @click="moveDown(index)" :disabled="index === types.length - 1" class="p-1 text-gray-400 hover:text-gray-600 disabled:opacity-30">↓</button>
          <button @click="saveReorder" v-if="reorderDirty" class="ml-2 text-xs text-indigo-600 hover:text-indigo-700 font-medium">Save Order</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import api from '@/api'

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
}

const types = ref<IssueType[]>([])
const loading = ref(false)
const copying = ref(false)
const reorderDirty = ref(false)

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
</script>
