<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { workspaceApi } from '@/api/workspace'
import { projectApi } from '@/api/project'
import * as workflowApi from '@/api/workflow'
import api from '@/api'
import type { Workspace } from '@/types'
import type { ProjectResponse } from '@/types/project'
import { useConfirm } from '@/composables/useConfirm'

const { confirm } = useConfirm()

const route = useRoute()
const router = useRouter()
const slug = computed(() => (route.params as any).slug as string)
const projectId = computed(() => parseInt((route.params as any).id as string, 10))

const loading = ref(false)
const workspaceId = ref(0)
const workspace = ref<Workspace | null>(null)
const project = ref<ProjectResponse | null>(null)
const activeSection = ref('states')

// ===== Data =====
const states = ref<any[]>([])
const labels = ref<any[]>([])
const workflows = ref<any[]>([])
const automations = ref<any[]>([])
const members = ref<any[]>([])

// ===== State groups =====
const STATE_GROUPS = [
  { id: 'backlog', name: 'Backlog', description: '待规划的工作项' },
  { id: 'unstarted', name: 'Unstarted', description: '已规划但未开始' },
  { id: 'started', name: 'Started', description: '正在执行中' },
  { id: 'completed', name: 'Completed', description: '已完成' },
  { id: 'cancelled', name: 'Cancelled', description: '已取消' },
] as const

const stateGroups = computed(() => {
  return STATE_GROUPS.map(group => ({
    ...group,
    states: states.value.filter((s: any) => s.group === group.id)
  }))
})

const totalStates = computed(() => states.value.length)

// ===== Load data =====
async function loadData() {
  loading.value = true
  try {
    const pid = projectId.value
    const results = await Promise.allSettled([
      api.get(`/projects/${pid}/settings/states`).then(r => r.data),
      api.get(`/projects/${pid}/settings/labels`).then(r => r.data),
      workflowApi.listWorkflows(pid),
      workflowApi.listAutomations(pid),
      api.get(`/projects/${pid}/members`).then(r => r.data),
    ])
    states.value = results[0].status === 'fulfilled' ? (Array.isArray(results[0].value) ? results[0].value : []) : []
    labels.value = results[1].status === 'fulfilled' ? (Array.isArray(results[1].value) ? results[1].value : []) : []
    workflows.value = results[2].status === 'fulfilled' ? (Array.isArray(results[2].value) ? results[2].value : []) : []
    automations.value = results[3].status === 'fulfilled' ? (Array.isArray(results[3].value) ? results[3].value : []) : []
    members.value = results[4].status === 'fulfilled' ? (Array.isArray(results[4].value) ? results[4].value : (results[4].value?.data || [])) : []
  } catch (e) { console.error('Failed to load data:', e) }
  finally { loading.value = false }
}

// ===== State handlers =====
const showStateModal = ref(false)
const editingState = ref<{ groupId: string; state: any } | null>(null)
const newStateForm = ref({ name: '', color: '#3B82F6', groupId: '' })

function handleAddState(groupId: string) {
  newStateForm.value = { name: '', color: '#3B82F6', groupId }
  editingState.value = null
  showStateModal.value = true
}
function handleEditState(groupId: string, state: any) {
  editingState.value = { groupId, state }
  newStateForm.value = { name: state.name, color: state.color, groupId }
  showStateModal.value = true
}
async function handleSaveState() {
  if (!newStateForm.value.name || !projectId.value) return
  try {
    if (editingState.value?.state) {
      await api.put(`/projects/${projectId.value}/settings/states/${editingState.value.state.id}`, {
        name: newStateForm.value.name, color: newStateForm.value.color
      })
    } else {
      await api.post(`/projects/${projectId.value}/settings/states?workspace_id=${workspaceId.value}`, {
        name: newStateForm.value.name, color: newStateForm.value.color, group: newStateForm.value.groupId
      })
    }
    showStateModal.value = false
    await loadData()
  } catch (e) { console.error('Failed to save state:', e) }
}
async function handleDeleteState(_groupId: string, state: any) {
  if (!projectId.value) return
  if (!(await confirm({ title: '删除状态', message: `确定要删除状态 "${state.name}" 吗？此操作不可恢复。`, danger: true, confirmText: '删除' }))) return
  try { await api.delete(`/projects/${projectId.value}/settings/states/${state.id}`); await loadData() }
  catch (e) { console.error('Failed to delete state:', e) }
}
async function handleCreateDefaultStates() {
  if (!projectId.value) return
  try {
    await api.post(`/projects/${projectId.value}/settings/states/default?workspace_id=${workspaceId.value}`)
    await loadData()
  } catch (e) { console.error('Failed to create default states:', e) }
}

// ===== Label handlers =====
const showLabelModal = ref(false)
const editingLabel = ref<any>(null)
const newLabelForm = ref({ name: '', color: '#3B82F6' })

function handleAddLabel() {
  editingLabel.value = null
  newLabelForm.value = { name: '', color: '#3B82F6' }
  showLabelModal.value = true
}
function handleEditLabel(label: any) {
  editingLabel.value = { ...label }
  newLabelForm.value = { name: label.name, color: label.color }
  showLabelModal.value = true
}
async function handleSaveLabel() {
  if (!newLabelForm.value.name || !projectId.value) return
  try {
    if (editingLabel.value) {
      await api.put(`/projects/${projectId.value}/settings/labels/${editingLabel.value.id}`, {
        name: newLabelForm.value.name, color: newLabelForm.value.color
      })
    } else {
      await api.post(`/projects/${projectId.value}/settings/labels?workspace_id=${workspaceId.value}`, {
        name: newLabelForm.value.name, color: newLabelForm.value.color
      })
    }
    showLabelModal.value = false
    await loadData()
  } catch (e) { console.error('Failed to save label:', e) }
}
async function handleDeleteLabel(label: any) {
  if (!projectId.value) return
  if (!(await confirm({ title: '删除标签', message: `确定要删除标签 "${label.name}" 吗？此操作不可恢复。`, danger: true, confirmText: '删除' }))) return
  try { await api.delete(`/projects/${projectId.value}/settings/labels/${label.id}`); await loadData() }
  catch (e) { console.error('Failed to delete label:', e) }
}

// ===== Workflow handlers =====
const showWorkflowModal = ref(false)
const newWorkflowForm = ref({ name: '', description: '' })

function handleViewWorkflow(workflowId: number) {
  router.push(`/workspace/${slug.value}/project/${projectId.value}/settings/workflows/${workflowId}`)
}
function handleAddWorkflow() {
  newWorkflowForm.value = { name: '', description: '' }
  showWorkflowModal.value = true
}
async function handleSaveWorkflow() {
  if (!newWorkflowForm.value.name || !projectId.value) return
  try {
    await workflowApi.createWorkflow(projectId.value, {
      name: newWorkflowForm.value.name,
      description: newWorkflowForm.value.description
    })
    showWorkflowModal.value = false
    await loadData()
  } catch (e) { console.error('Failed to create workflow:', e) }
}
async function handleDeleteWorkflow(workflow: any) {
  if (!projectId.value) return
  if (!(await confirm({ title: '删除工作流', message: `确定要删除工作流 "${workflow.name}" 吗？此操作不可恢复。`, danger: true, confirmText: '删除' }))) return
  try { await workflowApi.deleteWorkflow(projectId.value, workflow.id); await loadData() }
  catch (e) { console.error('Failed to delete workflow:', e) }
}

// ===== Automation handlers =====
const showAutomationModal = ref(false)
const newAutomationForm = ref({ name: '', description: '', trigger: 'issue_created', conditions: '[]', actions: '[]' })

function handleAddAutomation() {
  newAutomationForm.value = { name: '', description: '', trigger: 'issue_created', conditions: '[]', actions: '[]' }
  showAutomationModal.value = true
}
async function handleSaveAutomation() {
  if (!newAutomationForm.value.name || !projectId.value) return
  try {
    let conds = '[]', acts = '[]'
    try { conds = JSON.stringify(JSON.parse(newAutomationForm.value.conditions || '[]')); } catch { conds = '[]' }
    try { acts = JSON.stringify(JSON.parse(newAutomationForm.value.actions || '[]')); } catch { acts = '[]' }
    await workflowApi.createAutomation(projectId.value, {
      name: newAutomationForm.value.name,
      description: newAutomationForm.value.description,
      trigger_type: newAutomationForm.value.trigger,
      conditions: conds,
      actions: acts
    })
    showAutomationModal.value = false
    await loadData()
  } catch (e) { console.error('Failed to create automation:', e) }
}
async function handleDeleteAutomation(automation: any) {
  if (!projectId.value) return
  if (!(await confirm({ title: '删除自动化', message: `确定要删除自动化 "${automation.name}" 吗？此操作不可恢复。`, danger: true, confirmText: '删除' }))) return
  try { await workflowApi.deleteAutomation(projectId.value, automation.id); await loadData() }
  catch (e) { console.error('Failed to delete automation:', e) }
}

function goBack() {
  router.push(`/workspace/${slug.value}/project/${projectId.value}`)
}

onMounted(async () => {
  if (!slug.value) return
  try {
    workspace.value = await workspaceApi.getBySlug(slug.value)
    workspaceId.value = workspace.value.id
    project.value = await projectApi.getProject(projectId.value)
    await loadData()
  } catch (e) { console.error('Failed to load:', e) }
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex">
    <!-- Left Sidebar -->
    <aside class="w-64 bg-white border-r border-gray-200 flex flex-col shrink-0">
      <div class="p-4 border-b border-gray-200">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-lg flex items-center justify-center text-white text-lg font-bold"
            :style="{ backgroundColor: project?.color || '#6366f1' }">
            {{ project?.name?.charAt(0)?.toUpperCase() || 'P' }}
          </div>
          <div>
            <h2 class="font-semibold text-gray-800 text-sm">{{ project?.name }}</h2>
            <p class="text-xs text-gray-500">Project Settings</p>
          </div>
        </div>
      </div>
      <nav class="flex-1 p-4 space-y-1">
        <button
          v-for="item in [
            { id: 'states', label: 'States', icon: '🔄', count: totalStates },
            { id: 'labels', label: 'Labels', icon: '🏷️', count: labels.length },
            { id: 'workflows', label: 'Workflows', icon: '⚙️', count: workflows.length },
            { id: 'automations', label: 'Automations', icon: '🤖', count: automations.length },
          ]"
          :key="item.id"
          @click="activeSection = item.id"
          :class="['w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm font-medium transition-colors', activeSection === item.id ? 'bg-indigo-50 text-indigo-700' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-800']"
        >
          <span class="flex items-center space-x-3">
            <span>{{ item.icon }}</span><span>{{ item.label }}</span>
          </span>
          <span class="text-xs bg-gray-100 px-2 py-0.5 rounded-full">{{ item.count }}</span>
        </button>
      </nav>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 overflow-auto">
      <header class="bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
        <div>
          <h1 class="text-xl font-semibold text-gray-800">{{ activeSection === 'states' ? 'States' : activeSection === 'labels' ? 'Labels' : activeSection === 'workflows' ? 'Workflows' : activeSection === 'automations' ? 'Automations' : '' }}</h1>
          <p class="text-sm text-gray-500 mt-1">Configure project-level settings</p>
        </div>
        <button @click="goBack" class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50">← Back to Project</button>
      </header>

      <div v-if="loading" class="flex items-center justify-center h-64"><div class="animate-spin h-8 w-8 border-4 border-indigo-500 border-t-transparent rounded-full"></div></div>

      <!-- States Section -->
      <div v-if="!loading && activeSection === 'states'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div><h2 class="text-lg font-semibold text-gray-900">Work Item States</h2><p class="text-sm text-gray-500 mt-1">5 fixed state groups. Manage states within each group.</p></div>
          <div class="flex gap-2">
            <button v-if="states.length === 0" @click="handleCreateDefaultStates" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">Create Default States</button>
          </div>
        </div>
        <div class="space-y-6">
          <div v-for="group in stateGroups" :key="group.id" class="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div class="px-4 py-3 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
              <div class="flex items-center space-x-3">
                <span class="text-sm font-semibold text-gray-700 uppercase tracking-wide">{{ group.name }}</span>
                <span class="text-xs text-gray-400">{{ group.description }}</span>
                <span class="text-xs bg-gray-200 px-2 py-0.5 rounded-full">{{ group.states.length }}</span>
              </div>
              <button @click="handleAddState(group.id)" class="text-indigo-600 hover:text-indigo-700 text-sm font-medium">+ Add State</button>
            </div>
            <div class="divide-y divide-gray-100">
              <div v-for="state in group.states" :key="state.id" class="px-4 py-3 flex items-center justify-between hover:bg-gray-50 transition-colors cursor-pointer" @click="handleEditState(group.id, state)">
                <div class="flex items-center space-x-3">
                  <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: state.color }"></div>
                  <span class="text-sm text-gray-800">{{ state.name }}</span>
                  <span v-if="state.is_default" class="px-2 py-0.5 bg-indigo-100 text-indigo-600 rounded text-xs font-medium">Default</span>
                </div>
                <div class="flex items-center space-x-2">
                  <button @click.stop="handleEditState(group.id, state)" class="p-1 text-gray-400 hover:text-gray-600">✏️</button>
                  <button @click.stop="handleDeleteState(group.id, state)" class="p-1 text-gray-400 hover:text-red-500">🗑️</button>
                </div>
              </div>
              <div v-if="group.states.length === 0" class="px-4 py-6 text-center text-gray-400 text-sm">No states in this group. Click "+ Add State" to add one.</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Labels Section -->
      <div v-if="!loading && activeSection === 'labels'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div><h2 class="text-lg font-semibold text-gray-900">Labels</h2><p class="text-sm text-gray-500 mt-1">Categorize work items with color-coded labels</p></div>
          <button @click="handleAddLabel" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">+ Add Label</button>
        </div>
        <div class="flex flex-wrap gap-3">
          <div v-for="label in labels" :key="label.id" @click="handleEditLabel(label)" class="inline-flex items-center px-3 py-1.5 rounded-full cursor-pointer hover:opacity-80 transition-opacity" :style="{ backgroundColor: label.color + '20', borderColor: label.color }">
            <div class="w-2 h-2 rounded-full mr-2" :style="{ backgroundColor: label.color }"></div>
            <span class="text-sm font-medium" :style="{ color: label.color }">{{ label.name }}</span>
            <button @click.stop="handleDeleteLabel(label)" class="ml-2 text-gray-400 hover:text-red-500">✕</button>
          </div>
          <div v-if="labels.length === 0" class="w-full text-center text-gray-400 py-8">No labels. Click "+ Add Label" to create.</div>
        </div>
      </div>

      <!-- Workflows Section -->
      <div v-if="!loading && activeSection === 'workflows'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div><h2 class="text-lg font-semibold text-gray-900">Workflows</h2><p class="text-sm text-gray-500 mt-1">Manage state transition rules for this project</p></div>
          <button @click="handleAddWorkflow" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">+ Create Workflow</button>
        </div>
        <div class="grid gap-4">
          <div v-for="workflow in workflows" :key="workflow.id" class="bg-white rounded-xl border border-gray-200 p-4 hover:border-gray-300 hover:shadow-md transition-all">
            <div class="flex items-center justify-between">
              <div @click="handleViewWorkflow(workflow.id)" class="flex-1 cursor-pointer">
                <div class="flex items-center justify-between">
                  <div>
                    <h3 class="font-medium text-gray-900">{{ workflow.name }}</h3>
                    <p class="mt-1 text-sm text-gray-500">{{ (workflow.transitions || []).length }} transitions</p>
                  </div>
                  <span class="text-gray-400">→</span>
                </div>
              </div>
              <button @click.stop="handleDeleteWorkflow(workflow)" class="ml-3 text-gray-400 hover:text-red-500 p-1">✕</button>
            </div>
          </div>
          <div v-if="workflows.length === 0" class="text-center text-gray-400 py-12 bg-white rounded-xl border border-gray-200">No workflows configured. Click "+ Create Workflow" to add.</div>
        </div>
      </div>

      <!-- Automations Section -->
      <div v-if="!loading && activeSection === 'automations'" class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div><h2 class="text-lg font-semibold text-gray-900">Automations</h2><p class="text-sm text-gray-500 mt-1">Project-level automation rules</p></div>
          <button @click="handleAddAutomation" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">+ Create Automation</button>
        </div>
        <div class="space-y-4">
          <div v-for="automation in automations" :key="automation.id" class="bg-white rounded-xl border border-gray-200 p-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-3">
                <div :class="['w-10 h-10 rounded-lg flex items-center justify-center', automation.is_enabled ? 'bg-green-100' : 'bg-gray-100']">
                  <span class="text-lg">🤖</span>
                </div>
                <div>
                  <h3 class="font-medium text-gray-900">{{ automation.name }}</h3>
                  <p class="text-sm text-gray-500">{{ automation.description || 'No description' }}</p>
                </div>
              </div>
              <div class="flex items-center space-x-3">
                <span :class="['px-3 py-1 rounded-full text-xs font-medium', automation.is_enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500']">
                  {{ automation.is_enabled ? 'Enabled' : 'Disabled' }}
                </span>
                <button @click="handleDeleteAutomation(automation)" class="text-gray-400 hover:text-red-500">🗑️</button>
              </div>
            </div>
            <div class="mt-4 pt-4 border-t border-gray-100">
              <div class="flex items-center space-x-2 text-sm">
                <span class="px-2 py-1 bg-indigo-100 text-indigo-700 rounded font-medium">Trigger: {{ automation.trigger_type }}</span>
              </div>
            </div>
          </div>
          <div v-if="automations.length === 0" class="text-center text-gray-400 py-12 bg-white rounded-xl border border-gray-200">No automations. Click "+ Create Automation" to add.</div>
        </div>
      </div>
    </main>

    <!-- State Modal -->
    <div v-if="showStateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showStateModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ editingState ? 'Edit State' : 'Add State' }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Name</label><input v-model="newStateForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Color</label><input v-model="newStateForm.color" type="color" class="w-full h-12 border border-gray-300 rounded-lg cursor-pointer" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Group</label>
            <select v-model="newStateForm.groupId" class="w-full px-4 py-2 border border-gray-300 rounded-lg">
              <option v-for="g in STATE_GROUPS" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showStateModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="handleSaveState" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ editingState ? 'Update' : 'Create' }}</button>
        </div>
      </div>
    </div>

    <!-- Label Modal -->
    <div v-if="showLabelModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showLabelModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ editingLabel ? 'Edit Label' : 'Add Label' }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Name</label><input v-model="newLabelForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Color</label><input v-model="newLabelForm.color" type="color" class="w-full h-12 border border-gray-300 rounded-lg cursor-pointer" /></div>
          <div class="mt-2"><div class="inline-flex items-center px-4 py-2 rounded-full" :style="{ backgroundColor: newLabelForm.color + '20' }"><div class="w-2 h-2 rounded-full mr-2" :style="{ backgroundColor: newLabelForm.color }"></div><span class="text-sm font-medium" :style="{ color: newLabelForm.color }">{{ newLabelForm.name || 'Preview' }}</span></div></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showLabelModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="handleSaveLabel" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ editingLabel ? 'Update' : 'Create' }}</button>
        </div>
      </div>
    </div>

    <!-- Workflow Modal -->
    <div v-if="showWorkflowModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showWorkflowModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">Create Workflow</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Name</label><input v-model="newWorkflowForm.name" type="text" placeholder="e.g., Standard Workflow" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Description (optional)</label><input v-model="newWorkflowForm.description" type="text" placeholder="Brief description" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showWorkflowModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="handleSaveWorkflow" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">Create</button>
        </div>
      </div>
    </div>

    <!-- Automation Modal -->
    <div v-if="showAutomationModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showAutomationModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-lg">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">Create Automation</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Name</label><input v-model="newAutomationForm.name" type="text" placeholder="e.g., Auto-assign bugs" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">Description (optional)</label><input v-model="newAutomationForm.description" type="text" placeholder="What does this automation do?" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">Trigger</div>
            <select v-model="newAutomationForm.trigger" class="w-full px-4 py-2 border border-gray-300 rounded-lg">
              <option value="issue_created">Issue Created</option>
              <option value="issue_updated">Issue Updated</option>
              <option value="state_changed">State Changed</option>
              <option value="assignee_changed">Assignee Changed</option>
              <option value="comment_added">Comment Added</option>
            </select>
          </div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">Conditions (JSON)</div>
            <textarea v-model="newAutomationForm.conditions" rows="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-xs font-mono" placeholder='[{"field":"priority","operator":"equals","value":"urgent"}]'></textarea>
          </div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">Actions (JSON)</div>
            <textarea v-model="newAutomationForm.actions" rows="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-xs font-mono" placeholder='[{"type":"assign","field":"assignee","value":"1"}]'></textarea>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showAutomationModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="handleSaveAutomation" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">Create</button>
        </div>
      </div>
    </div>
  </div>
</template>
