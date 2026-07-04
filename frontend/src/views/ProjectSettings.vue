<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { workspaceApi } from '@/api/workspace'
import { projectApi } from '@/api/project'
import * as workflowApi from '@/api/workflow'
import api from '@/api'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import type { Workspace } from '@/types'
import type { ProjectResponse, ProjectUpdate, ProjectSubscriber } from '@/types/project'
import { useConfirm } from '@/composables/useConfirm'
import ProjectMemberList from '@/components/ProjectMemberList.vue'
import ModuleList from '@/components/ModuleList.vue'
import CycleList from '@/components/CycleList.vue'
import CustomFieldList from '@/components/CustomFieldList.vue'
import CustomFieldForm from '@/components/CustomFieldForm.vue'
import EstimatePointManager from '@/components/EstimatePointManager.vue'
import ProjectIssueTypeManager from '@/components/ProjectIssueTypeManager.vue'
import WorkItemTemplateManager from '@/components/WorkItemTemplateManager.vue'
import ReleaseList from '@/components/ReleaseList.vue'
import WebhookManager from '@/components/WebhookManager.vue'

const { confirm } = useConfirm()
const { t } = useI18n()
const toast = useToast()

const route = useRoute()
const router = useRouter()
const slug = computed(() => (route.params as any).slug as string)
const projectId = computed(() => parseInt((route.params as any).id as string, 10))

const loading = ref(false)
const workspaceId = ref(0)
const workspace = ref<Workspace | null>(null)
const project = ref<ProjectResponse | null>(null)
const activeSection = ref('overview')

// ===== Settings sub-page state =====
const settingsLoading = ref(false)
const settingsError = ref('')
const deleteLoading = ref(false)
const showEditModal = ref(false)
const showDeleteConfirm = ref(false)
const editForm = reactive({ name: '', identifier: '', description: '', color: '#6366f1' })
const stats = ref({ issuesCount: 0, membersCount: 0 })

// ===== Data =====
const states = ref<any[]>([])
const labels = ref<any[]>([])
const workflows = ref<any[]>([])
const automations = ref<any[]>([])
const members = ref<any[]>([])

// ===== Custom Field Form =====
const showFieldForm = ref(false)
const editingField = ref<any>(null)

function handleCreateField() {
  editingField.value = null
  showFieldForm.value = true
}

function handleEditField(field: any) {
  editingField.value = field
  showFieldForm.value = true
}

function handleFieldSaved() {
  showFieldForm.value = false
  editingField.value = null
}

// ===== Default Assignee / Project Lead / Subscribers =====
const defaultAssigneeId = ref<number | null>(null)
const projectLeadId = ref<number | null>(null)
const subscribers = ref<ProjectSubscriber[]>([])
const updatingAssignee = ref(false)
const updatingLead = ref(false)
const showSubscriberPicker = ref(false)
const subscriberPickerUserId = ref<number | null>(null)

function getMemberDisplay(member: any): string {
  if (!member) return ''
  return member.user?.display_name || member.display_name || member.username || member.user?.username || `User #${member.user_id}`
}

function getMemberById(userId: number | null | undefined): any | undefined {
  if (!userId) return undefined
  return members.value.find((m: any) => m.user_id === userId)
}

function getAutomationTriggerLabel(triggerType: string): string {
  try {
    const parsed = JSON.parse(triggerType)
    if (parsed && typeof parsed.type === 'string') {
      const key = parsed.type.replace(/\./g, '_')
      return (t as any)(`settings.triggerTypes.${key}`) || parsed.type
    }
  } catch {}
  const key = triggerType.replace(/\./g, '_')
  return (t as any)(`settings.triggerTypes.${key}`) || triggerType
}

// ===== Menu items =====
const menuItems = computed(() => [
  { id: 'overview', label: t('settings.overview'), icon: '📊' },
  { id: 'members', label: t('settings.members'), icon: '👥' },
  { id: 'states', label: t('settings.states'), icon: '🔄' },
  { id: 'labels', label: t('settings.labels'), icon: '🏷️' },
  { id: 'issue-types', label: t('settings.issueTypes'), icon: '📐' },
  { id: 'templates', label: t('settings.templates'), icon: '📋' },
  { id: 'modules', label: t('settings.modules'), icon: '📦' },
  { id: 'cycles', label: t('settings.cycles'), icon: '🔄' },
  { id: 'releases', label: t('settings.releases'), icon: '🚀' },
  { id: 'webhooks', label: t('settings.webhooks'), icon: '🔌' },
  { id: 'custom-fields', label: t('settings.customFields'), icon: '🔧' },
  { id: 'estimate-points', label: t('settings.estimatePoints'), icon: '📏' },
  { id: 'workflows', label: t('settings.workflows'), icon: '⚙️' },
  { id: 'automations', label: t('settings.automations'), icon: '🤖' },
  { id: 'delete', label: t('settings.deleteProject'), icon: '🗑️' },
])

const currentMenuLabel = computed(() => {
  const item = menuItems.value.find((i: { id: string }) => i.id === activeSection.value)
  return item?.label || ''
})

// ===== State groups =====
const STATE_GROUP_KEYS = ['backlog', 'unstarted', 'started', 'completed', 'cancelled'] as const

const stateGroups = computed(() => {
  const groupNames: Record<string, string> = {
    backlog: t('settings.stateGroupBacklogName'),
    unstarted: t('settings.stateGroupUnstartedName'),
    started: t('settings.stateGroupStartedName'),
    completed: t('settings.stateGroupCompletedName'),
    cancelled: t('settings.stateGroupCancelledName'),
  }
  const groupDescriptions: Record<string, string> = {
    backlog: t('settings.stateGroupBacklog'),
    unstarted: t('settings.stateGroupUnstarted'),
    started: t('settings.stateGroupStarted'),
    completed: t('settings.stateGroupCompleted'),
    cancelled: t('settings.stateGroupCancelled'),
  }
  return STATE_GROUP_KEYS.map(id => ({
    id,
    name: groupNames[id],
    description: groupDescriptions[id],
    states: states.value.filter((s: any) => s.group === id)
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
  } catch (e: any) {
    console.error('Failed to load data:', e)
    toast.error(e?.response?.data?.message || 'Failed to load settings data')
  } finally { loading.value = false }
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
  } catch (e: any) { console.error('Failed to save state:', e); toast.error(e?.response?.data?.message || 'Failed to save state') }
}
async function handleDeleteState(_groupId: string, state: any) {
  if (!projectId.value) return
  if (!(await confirm({ title: t('settings.deleteState'), message: t('settings.confirmDeleteState', { 0: state.name }), danger: true, confirmText: t('common.delete') }))) return
  try { await api.delete(`/projects/${projectId.value}/settings/states/${state.id}`); await loadData() }
  catch (e: any) { console.error('Failed to delete state:', e); toast.error(e?.response?.data?.message || 'Failed to delete state') }
}
async function handleCreateDefaultStates() {
  if (!projectId.value) return
  try {
    await api.post(`/projects/${projectId.value}/settings/states/default?workspace_id=${workspaceId.value}`)
    await loadData()
  } catch (e: any) { console.error('Failed to create default states:', e); toast.error(e?.response?.data?.message || 'Failed to create default states') }
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
  } catch (e: any) { console.error('Failed to save label:', e); toast.error(e?.response?.data?.message || 'Failed to save label') }
}
async function handleDeleteLabel(label: any) {
  if (!projectId.value) return
  if (!(await confirm({ title: t('settings.deleteLabel'), message: t('settings.confirmDeleteLabel', { 0: label.name }), danger: true, confirmText: t('common.delete') }))) return
  try { await api.delete(`/projects/${projectId.value}/settings/labels/${label.id}`); await loadData() }
  catch (e: any) { console.error('Failed to delete label:', e); toast.error(e?.response?.data?.message || 'Failed to delete label') }
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
  } catch (e: any) { console.error('Failed to create workflow:', e); toast.error(e?.response?.data?.message || 'Failed to create workflow') }
}
async function handleDeleteWorkflow(workflow: any) {
  if (!projectId.value) return
  if (!(await confirm({ title: t('settings.deleteWorkflow'), message: t('settings.confirmDeleteWorkflow', { 0: workflow.name }), danger: true, confirmText: t('common.delete') }))) return
  try { await workflowApi.deleteWorkflow(projectId.value, workflow.id); await loadData() }
  catch (e: any) { console.error('Failed to delete workflow:', e); toast.error(e?.response?.data?.message || 'Failed to delete workflow') }
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
    const triggerMap: Record<string, string> = {
      issue_created: 'issue.created', issue_updated: 'issue.updated',
      state_changed: 'issue.state_changed', assignee_changed: 'issue.assigned',
      comment_added: 'comment.added',
    }
    const triggerType = JSON.stringify({ type: triggerMap[newAutomationForm.value.trigger] || newAutomationForm.value.trigger })
    await workflowApi.createAutomation(projectId.value, {
      name: newAutomationForm.value.name,
      description: newAutomationForm.value.description,
      trigger_type: triggerType,
      conditions: conds,
      actions: acts
    })
    showAutomationModal.value = false
    await loadData()
  } catch (e: any) { console.error('Failed to create automation:', e); toast.error(e?.response?.data?.message || 'Failed to create automation') }
}
async function handleDeleteAutomation(automation: any) {
  if (!projectId.value) return
  if (!(await confirm({ title: t('settings.deleteAutomation'), message: t('settings.confirmDeleteAutomation', { 0: automation.name }), danger: true, confirmText: t('common.delete') }))) return
  try { await workflowApi.deleteAutomation(projectId.value, automation.id); await loadData() }
  catch (e: any) { console.error('Failed to delete automation:', e); toast.error(e?.response?.data?.message || 'Failed to delete automation') }
}

// ===== Overview / Edit handlers =====
async function handleEditProject() {
  editForm.name = project.value?.name || ''
  editForm.identifier = project.value?.identifier || ''
  editForm.description = project.value?.description || ''
  editForm.color = project.value?.color || '#6366f1'
  showEditModal.value = true
}
async function handleSaveProject() {
  if (!editForm.name) { settingsError.value = 'Please enter project name'; return }
  settingsLoading.value = true; settingsError.value = ''
  try {
    const data: ProjectUpdate = { name: editForm.name, identifier: editForm.identifier || undefined, description: editForm.description || undefined, color: editForm.color }
    const updated = await projectApi.updateProject(projectId.value, data)
    project.value = { ...project.value, ...updated } as ProjectResponse
    showEditModal.value = false
  } catch (e: any) { settingsError.value = e.response?.data?.message || 'Save failed' }
  finally { settingsLoading.value = false }
}
async function handleDeleteProject() {
  deleteLoading.value = true
  try { await projectApi.deleteProject(projectId.value); router.push(`/workspace/${slug.value}`) }
  catch (e: any) { toast.error(e.response?.data?.message || 'Delete failed'); showDeleteConfirm.value = false }
  finally { deleteLoading.value = false }
}

// ===== Default Assignee handlers =====
async function handleUpdateDefaultAssignee() {
  if (!projectId.value) return
  updatingAssignee.value = true
  try {
    const updated = await projectApi.updateDefaultAssignee(projectId.value, defaultAssigneeId.value)
    project.value = { ...project.value, ...updated } as ProjectResponse
  } catch (e: any) { console.error('Failed to update default assignee:', e); toast.error(e?.response?.data?.message || 'Failed to update default assignee') }
  finally { updatingAssignee.value = false }
}

// ===== Project Lead handlers =====
async function handleUpdateProjectLead() {
  if (!projectId.value) return
  updatingLead.value = true
  try {
    const updated = await projectApi.updateProjectLead(projectId.value, projectLeadId.value)
    project.value = { ...project.value, ...updated } as ProjectResponse
  } catch (e: any) { console.error('Failed to update project lead:', e); toast.error(e?.response?.data?.message || 'Failed to update project lead') }
  finally { updatingLead.value = false }
}

// ===== Subscriber handlers =====
async function loadSubscribers() {
  if (!projectId.value) return
  try {
    const data = await projectApi.listProjectSubscribers(projectId.value)
    subscribers.value = Array.isArray(data) ? data : []
  } catch (e) { console.error('Failed to load subscribers:', e) }
  // Non-critical; silently ignore
}

function openSubscriberPicker() {
  subscriberPickerUserId.value = null
  showSubscriberPicker.value = true
}

async function handleAddSubscriber() {
  if (!projectId.value || !subscriberPickerUserId.value) return
  try {
    await projectApi.addProjectSubscriber(projectId.value, subscriberPickerUserId.value)
    showSubscriberPicker.value = false
    await loadSubscribers()
  } catch (e: any) { console.error('Failed to add subscriber:', e); toast.error(e?.response?.data?.message || 'Failed to add subscriber') }
}

async function handleRemoveSubscriber(userId: number) {
  if (!projectId.value) return
  if (!(await confirm({ title: t('settings.removeSubscriber'), message: t('settings.confirmRemoveSubscriber'), danger: true, confirmText: t('settings.remove') }))) return
  try {
    await projectApi.removeProjectSubscriber(projectId.value, userId)
    await loadSubscribers()
  } catch (e: any) { console.error('Failed to remove subscriber:', e); toast.error(e?.response?.data?.message || 'Failed to remove subscriber') }
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
    defaultAssigneeId.value = project.value.default_assignee_id ?? null
    projectLeadId.value = project.value.project_lead_id ?? null
    try {
      const membersData = await projectApi.listProjectMembers(projectId.value)
      stats.value.membersCount = Array.isArray(membersData) ? membersData.length : 0
    } catch (_) { /* ignore */ }
    await loadData()
    await loadSubscribers()
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
      <nav class="flex-1 p-2 overflow-y-auto">
        <button
          v-for="item in menuItems"
          :key="item.id"
          @click="activeSection = item.id"
          :class="['w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors', activeSection === item.id ? 'bg-indigo-50 text-indigo-600' : 'text-gray-600 hover:bg-gray-50']"
        >
          <span>{{ item.icon }}</span><span>{{ item.label }}</span>
        </button>
      </nav>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 overflow-auto">
      <header class="bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
        <div>
          <h1 class="text-xl font-semibold text-gray-800">{{ currentMenuLabel }}</h1>
          <p class="text-sm text-gray-500 mt-1">{{ t('settings.configureProject') }}</p>
        </div>
        <button @click="goBack" class="px-4 py-2 text-sm text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 transition">{{ t('settings.backToProject') }}</button>
      </header>

      <div v-if="loading" class="flex items-center justify-center h-64"><div class="animate-spin h-8 w-8 border-4 border-indigo-500 border-t-transparent rounded-full"></div></div>

      <div class="p-6">
        <!-- Overview -->
        <div v-if="!loading && activeSection === 'overview'" class="bg-white rounded-lg border border-gray-200">
          <div class="p-6">
            <h2 class="text-lg font-semibold text-gray-800 mb-4">{{ t('settings.projectOverview') }}</h2>
            <div class="grid grid-cols-3 gap-4">
              <div class="p-4 bg-gray-50 rounded-lg">
                <p class="text-sm text-gray-500">{{ t('settings.totalStates') }}</p>
                <p class="text-2xl font-bold text-gray-800">{{ totalStates }}</p>
              </div>
              <div class="p-4 bg-gray-50 rounded-lg">
                <p class="text-sm text-gray-500">{{ t('settings.members') }}</p>
                <p class="text-2xl font-bold text-gray-800">{{ stats.membersCount }}</p>
              </div>
              <div class="p-4 bg-gray-50 rounded-lg">
                <p class="text-sm text-gray-500">{{ t('settings.projectLead') }}</p>
                <p class="text-lg font-bold text-gray-800">{{ project?.project_lead?.display_name || getMemberDisplay(getMemberById(project?.project_lead_id)) || '—' }}</p>
              </div>
            </div>
            <div class="pt-4 mt-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('settings.description') }}</label>
              <p class="text-gray-600">{{ project?.description || t('settings.noDescription') }}</p>
            </div>
          </div>
          <div class="px-6 py-4 border-t border-gray-200">
            <button @click="handleEditProject" class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 transition">{{ t('settings.editProject') }}</button>
          </div>
        </div>

        <!-- Members -->
        <div v-if="!loading && activeSection === 'members'" class="space-y-6">
          <div class="bg-white rounded-lg border border-gray-200">
            <ProjectMemberList :project-id="projectId" :workspace-id="workspaceId" />
          </div>

          <!-- Default Assignee -->
          <div class="bg-white rounded-lg border border-gray-200 p-6">
            <h3 class="text-lg font-semibold text-gray-800 mb-4">{{ t('settings.defaultAssignee') }}</h3>
            <p class="text-sm text-gray-500 mb-4">{{ t('settings.newIssuesAssigned') }}</p>
            <div class="flex items-center gap-3">
              <select
                v-model.number="defaultAssigneeId"
                class="flex-1 max-w-md px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              >
                <option :value="null">{{ t('settings.none') }}</option>
                <option v-for="m in members" :key="m.user_id" :value="m.user_id">
                  {{ getMemberDisplay(m) }}
                </option>
              </select>
              <button
                @click="handleUpdateDefaultAssignee"
                :disabled="updatingAssignee"
                class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition"
              >
                {{ updatingAssignee ? t('settings.saving') : t('settings.save') }}
              </button>
            </div>
            <div v-if="getMemberById(defaultAssigneeId)" class="mt-3 text-sm text-gray-600">
              {{ t('settings.current') }}: <span class="font-medium">{{ getMemberDisplay(getMemberById(defaultAssigneeId)) }}</span>
            </div>
          </div>

          <!-- Project Lead -->
          <div class="bg-white rounded-lg border border-gray-200 p-6">
            <h3 class="text-lg font-semibold text-gray-800 mb-4">{{ t('settings.projectLead') }}</h3>
            <p class="text-sm text-gray-500 mb-4">{{ t('settings.designateProjectLead') }}</p>
            <div class="flex items-center gap-3">
              <select
                v-model.number="projectLeadId"
                class="flex-1 max-w-md px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              >
                <option :value="null">{{ t('settings.none') }}</option>
                <option v-for="m in members" :key="m.user_id" :value="m.user_id">
                  {{ getMemberDisplay(m) }}
                </option>
              </select>
              <button
                @click="handleUpdateProjectLead"
                :disabled="updatingLead"
                class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition"
              >
                {{ updatingLead ? t('settings.saving') : t('settings.save') }}
              </button>
            </div>
            <div v-if="getMemberById(projectLeadId)" class="mt-3 text-sm text-gray-600">
              {{ t('settings.current') }}: <span class="font-medium">{{ getMemberDisplay(getMemberById(projectLeadId)) }}</span>
            </div>
          </div>

          <!-- Project Subscribers -->
          <div class="bg-white rounded-lg border border-gray-200 p-6">
            <h3 class="text-lg font-semibold text-gray-800 mb-4">{{ t('settings.subscribers') }}</h3>
            <p class="text-sm text-gray-500 mb-4">{{ t('settings.subscribersReceive') }}</p>
            <div v-if="subscribers.length > 0" class="divide-y divide-gray-100 border border-gray-200 rounded-lg mb-4">
              <div v-for="sub in subscribers" :key="sub.id" class="px-4 py-3 flex items-center justify-between">
                <div>
                  <span class="text-sm font-medium text-gray-800">
                    {{ sub.user?.display_name || `User #${sub.user_id}` }}
                  </span>
                  <span v-if="sub.user?.username" class="text-sm text-gray-400 ml-2">@{{ sub.user.username }}</span>
                </div>
                <button @click="handleRemoveSubscriber(sub.user_id)" class="text-gray-400 hover:text-red-500 text-sm font-medium">{{ t('settings.remove') }}</button>
              </div>
            </div>
            <div v-else class="text-sm text-gray-400 py-4">{{ t('settings.noSubscribers') }}</div>
            <button @click="openSubscriberPicker" class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 transition">
              + {{ t('settings.addSubscriber') }}
            </button>
          </div>
        </div>

        <!-- States -->
        <div v-if="!loading && activeSection === 'states'" class="space-y-6">
          <div class="flex items-center justify-between">
            <div><h2 class="text-lg font-semibold text-gray-900">{{ t('settings.workItemStates') }}</h2><p class="text-sm text-gray-500 mt-1">{{ t('settings.stateGroupsDesc') }}</p></div>
            <div class="flex gap-2">
              <button v-if="states.length === 0" @click="handleCreateDefaultStates" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">{{ t('settings.createDefaultStates') }}</button>
            </div>
          </div>
          <div v-for="group in stateGroups" :key="group.id" class="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div class="px-4 py-3 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
              <div class="flex items-center space-x-3">
                <span class="text-sm font-semibold text-gray-700 uppercase tracking-wide">{{ group.name }}</span>
                <span class="text-xs text-gray-400">{{ group.description }}</span>
                <span class="text-xs bg-gray-200 px-2 py-0.5 rounded-full">{{ group.states.length }}</span>
              </div>
              <button @click="handleAddState(group.id)" class="text-indigo-600 hover:text-indigo-700 text-sm font-medium">+ {{ t('settings.addState') }}</button>
            </div>
            <div class="divide-y divide-gray-100">
              <div v-for="state in group.states" :key="state.id" class="px-4 py-3 flex items-center justify-between hover:bg-gray-50 transition-colors cursor-pointer" @click="handleEditState(group.id, state)">
                <div class="flex items-center space-x-3">
                  <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: state.color }"></div>
                  <span class="text-sm text-gray-800">{{ state.name }}</span>
                  <span v-if="state.is_default" class="px-2 py-0.5 bg-indigo-100 text-indigo-600 rounded text-xs font-medium">{{ t('settings.default') }}</span>
                </div>
                <div class="flex items-center space-x-2">
                  <button @click.stop="handleEditState(group.id, state)" class="p-1 text-gray-400 hover:text-gray-600">✏️</button>
                  <button @click.stop="handleDeleteState(group.id, state)" class="p-1 text-gray-400 hover:text-red-500">🗑️</button>
                </div>
              </div>
              <div v-if="group.states.length === 0" class="px-4 py-6 text-center text-gray-400 text-sm">{{ t('settings.noStates') }}</div>
            </div>
          </div>
        </div>

        <!-- Labels -->
        <div v-if="!loading && activeSection === 'labels'" class="space-y-6">
          <div class="flex items-center justify-between">
            <div><h2 class="text-lg font-semibold text-gray-900">{{ t('settings.labels') }}</h2><p class="text-sm text-gray-500 mt-1">{{ t('settings.labelsDesc') }}</p></div>
            <button @click="handleAddLabel" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">+ {{ t('settings.addLabel') }}</button>
          </div>
          <div class="bg-white rounded-xl border border-gray-200 p-6">
            <div class="flex flex-wrap gap-3">
              <div v-for="label in labels" :key="label.id" @click="handleEditLabel(label)" class="inline-flex items-center px-3 py-1.5 rounded-full cursor-pointer hover:opacity-80 transition-opacity" :style="{ backgroundColor: label.color + '20', borderColor: label.color }">
                <div class="w-2 h-2 rounded-full mr-2" :style="{ backgroundColor: label.color }"></div>
                <span class="text-sm font-medium" :style="{ color: label.color }">{{ label.name }}</span>
                <button @click.stop="handleDeleteLabel(label)" class="ml-2 text-gray-400 hover:text-red-500">✕</button>
              </div>
              <div v-if="labels.length === 0" class="w-full text-center text-gray-400 py-8">{{ t('settings.noLabels') }}</div>
            </div>
          </div>
        </div>

        <!-- Issue Types -->
        <div v-if="!loading && activeSection === 'issue-types'" class="bg-white rounded-lg border border-gray-200">
          <ProjectIssueTypeManager :project-id="projectId" :workspace-id="workspaceId" />
        </div>

        <!-- Templates -->
        <div v-if="!loading && activeSection === 'templates'" class="bg-white rounded-lg border border-gray-200">
          <WorkItemTemplateManager :project-id="projectId" :workspace-id="workspaceId" />
        </div>

        <!-- Modules -->
        <div v-if="!loading && activeSection === 'modules'" class="bg-white rounded-lg border border-gray-200">
          <ModuleList :project-id="projectId" :workspace-id="workspaceId" />
        </div>

        <!-- Cycles -->
        <div v-if="!loading && activeSection === 'cycles'" class="bg-white rounded-lg border border-gray-200">
          <CycleList :project-id="projectId" :workspace-id="workspaceId" />
        </div>

        <!-- Releases -->
        <div v-if="!loading && activeSection === 'releases'" class="bg-white rounded-lg border border-gray-200">
          <ReleaseList :project-id="projectId" />
        </div>
        <div v-if="!loading && activeSection === 'webhooks'" class="bg-white rounded-lg border border-gray-200">
          <WebhookManager :project-id="projectId" :workspace-id="workspaceId" />
        </div>

        <!-- Custom Fields -->
        <div v-if="!loading && activeSection === 'custom-fields'" class="bg-white rounded-lg border border-gray-200">
          <CustomFieldList :project-id="projectId" @create="handleCreateField" @edit="handleEditField" />
        </div>

        <!-- Estimate Points -->
        <div v-if="!loading && activeSection === 'estimate-points'" class="bg-white rounded-lg border border-gray-200">
          <EstimatePointManager :project-id="projectId" />
        </div>

        <!-- Workflows -->
        <div v-if="!loading && activeSection === 'workflows'" class="space-y-6">
          <div class="flex items-center justify-between">
            <div><h2 class="text-lg font-semibold text-gray-900">{{ t('settings.workflows') }}</h2><p class="text-sm text-gray-500 mt-1">{{ t('settings.workflowsDesc') }}</p></div>
            <button @click="handleAddWorkflow" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">+ {{ t('settings.createWorkflow') }}</button>
          </div>
          <div class="grid gap-4">
            <div v-for="workflow in workflows" :key="workflow.id" class="bg-white rounded-xl border border-gray-200 p-4 hover:border-gray-300 hover:shadow-md transition-all">
              <div class="flex items-center justify-between">
                <div @click="handleViewWorkflow(workflow.id)" class="flex-1 cursor-pointer">
                  <div class="flex items-center justify-between">
                    <div>
                      <h3 class="font-medium text-gray-900">{{ workflow.name }}</h3>
                      <p class="mt-1 text-sm text-gray-500">{{ (workflow.transitions || []).length }} {{ t('settings.transitions') }}</p>
                    </div>
                    <span class="text-gray-400">→</span>
                  </div>
                </div>
                <button @click.stop="handleDeleteWorkflow(workflow)" class="ml-3 text-gray-400 hover:text-red-500 p-1">✕</button>
              </div>
            </div>
            <div v-if="workflows.length === 0" class="text-center text-gray-400 py-12 bg-white rounded-xl border border-gray-200">{{ t('settings.noWorkflows') }}</div>
          </div>
        </div>

        <!-- Automations -->
        <div v-if="!loading && activeSection === 'automations'" class="space-y-6">
          <div class="flex items-center justify-between">
            <div><h2 class="text-lg font-semibold text-gray-900">{{ t('settings.automations') }}</h2><p class="text-sm text-gray-500 mt-1">{{ t('settings.automationsDesc') }}</p></div>
            <button @click="handleAddAutomation" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">+ {{ t('settings.createAutomation') }}</button>
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
                    <p class="text-sm text-gray-500">{{ automation.description || t('settings.noDescription') }}</p>
                  </div>
                </div>
                <div class="flex items-center space-x-3">
                  <span :class="['px-3 py-1 rounded-full text-xs font-medium', automation.is_enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500']">
                    {{ automation.is_enabled ? t('settings.enabled') : t('settings.disabled') }}
                  </span>
                  <button @click="handleDeleteAutomation(automation)" class="text-gray-400 hover:text-red-500">🗑️</button>
                </div>
              </div>
              <div class="mt-4 pt-4 border-t border-gray-100">
                <div class="flex items-center space-x-2 text-sm">
                  <span class="px-2 py-1 bg-indigo-100 text-indigo-700 rounded font-medium">{{ t('settings.trigger') }}: {{ getAutomationTriggerLabel(automation.trigger_type) }}</span>
                </div>
              </div>
            </div>
            <div v-if="automations.length === 0" class="text-center text-gray-400 py-12 bg-white rounded-xl border border-gray-200">{{ t('settings.noAutomations') }}</div>
          </div>
        </div>

        <!-- Delete Project -->
        <div v-if="!loading && activeSection === 'delete'" class="bg-white rounded-lg border border-gray-200 p-6">
          <h2 class="text-lg font-semibold text-gray-800 mb-4">{{ t('settings.deleteProject') }}</h2>
          <div class="bg-red-50 border border-red-200 rounded-lg p-4">
            <div class="flex items-start gap-3">
              <span class="text-2xl">⚠️</span>
              <div>
                <h3 class="font-medium text-red-800">{{ t('settings.permanentlyDelete') }}</h3>
                <p class="text-sm text-red-700 mt-1">{{ t('settings.deleteProjectDesc') }}</p>
              </div>
            </div>
            <button @click="showDeleteConfirm = true" class="mt-4 w-full px-4 py-2 bg-red-600 text-white text-sm rounded-lg hover:bg-red-700 transition">
              {{ t('settings.deleteProject') }}
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Edit Project Modal -->
    <div v-if="showEditModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showEditModal = false">
      <div class="bg-white rounded-xl shadow-lg p-6 w-full max-w-md mx-4">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-800">{{ t('settings.editProject') }}</h3>
          <button @click="showEditModal = false" class="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
        </div>
        <div v-if="settingsError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">{{ settingsError }}</div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.projectName') }}</label>
            <input v-model="editForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" :placeholder="t('settings.namePlaceholder')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.projectIdentifier') }}</label>
            <input v-model="editForm.identifier" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 uppercase" :placeholder="t('settings.identifierPlaceholder')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.projectDescription') }}</label>
            <textarea v-model="editForm.description" rows="3" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" :placeholder="t('settings.descriptionPlaceholder')"></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.projectColor') }}</label>
            <div class="flex items-center gap-2">
              <input v-model="editForm.color" type="color" class="w-10 h-10 rounded border border-gray-300 cursor-pointer" />
              <input v-model="editForm.color" type="text" class="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" placeholder="#6366f1" />
            </div>
          </div>
        </div>
        <div class="flex gap-3 mt-6">
          <button @click="showEditModal = false" class="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition">{{ t('settings.cancel') }}</button>
          <button @click="handleSaveProject" :disabled="settingsLoading" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition">
            {{ settingsLoading ? t('settings.saving') : t('settings.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirm Modal -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showDeleteConfirm = false">
      <div class="bg-white rounded-xl shadow-lg p-6 w-full max-w-md mx-4">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center text-xl">⚠️</div>
          <div>
            <h3 class="text-lg font-semibold text-gray-800">{{ t('settings.confirmDelete') }}</h3>
            <p class="text-sm text-gray-500">{{ t('settings.cannotUndo') }}</p>
          </div>
        </div>
        <p class="text-gray-600 mb-6" v-html="t('settings.deleteProjectConfirm', { name: `<strong>${project?.name}</strong>` })"></p>
        <div class="flex gap-3">
          <button @click="showDeleteConfirm = false" class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition">{{ t('settings.cancel') }}</button>
          <button @click="handleDeleteProject" :disabled="deleteLoading" class="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50 transition">
            {{ deleteLoading ? t('settings.deleting') : t('settings.deleteProject') }}
          </button>
        </div>
      </div>
    </div>

    <!-- State Modal -->
    <div v-if="showStateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showStateModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ editingState ? t('settings.editState') : t('settings.addState') }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.name') }}</label><input v-model="newStateForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.color') }}</label><input v-model="newStateForm.color" type="color" class="w-full h-12 border border-gray-300 rounded-lg cursor-pointer" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.stateGroup') }}</label>
            <select v-model="newStateForm.groupId" class="w-full px-4 py-2 border border-gray-300 rounded-lg">
              <option v-for="g in stateGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showStateModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('settings.cancel') }}</button>
          <button @click="handleSaveState" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ editingState ? t('settings.update') : t('settings.create') }}</button>
        </div>
      </div>
    </div>

    <!-- Label Modal -->
    <div v-if="showLabelModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showLabelModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ editingLabel ? t('settings.editLabel') : t('settings.addLabel') }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.name') }}</label><input v-model="newLabelForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.color') }}</label><input v-model="newLabelForm.color" type="color" class="w-full h-12 border border-gray-300 rounded-lg cursor-pointer" /></div>
          <div class="mt-2"><div class="inline-flex items-center px-4 py-2 rounded-full" :style="{ backgroundColor: newLabelForm.color + '20' }"><div class="w-2 h-2 rounded-full mr-2" :style="{ backgroundColor: newLabelForm.color }"></div><span class="text-sm font-medium" :style="{ color: newLabelForm.color }">{{ newLabelForm.name || t('settings.preview') }}</span></div></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showLabelModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('settings.cancel') }}</button>
          <button @click="handleSaveLabel" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ editingLabel ? t('settings.update') : t('settings.create') }}</button>
        </div>
      </div>
    </div>

    <!-- Workflow Modal -->
    <div v-if="showWorkflowModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showWorkflowModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('settings.createWorkflow') }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.name') }}</label><input v-model="newWorkflowForm.name" type="text" :placeholder="t('settings.workflowNamePlaceholder')" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.descriptionOptional') }}</label><input v-model="newWorkflowForm.description" type="text" :placeholder="t('settings.briefDescription')" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showWorkflowModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('settings.cancel') }}</button>
          <button @click="handleSaveWorkflow" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ t('settings.create') }}</button>
        </div>
      </div>
    </div>

    <!-- Automation Modal -->
    <div v-if="showAutomationModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showAutomationModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-lg">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('settings.createAutomation') }}</h3>
        <div class="space-y-4">
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.name') }}</label><input v-model="newAutomationForm.name" type="text" :placeholder="t('settings.automationNamePlaceholder')" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div><label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.descriptionOptional') }}</label><input v-model="newAutomationForm.description" type="text" :placeholder="t('settings.automationDescriptionPlaceholder')" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" /></div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">{{ t('settings.trigger') }}</div>
            <select v-model="newAutomationForm.trigger" class="w-full px-4 py-2 border border-gray-300 rounded-lg">
              <option value="issue_created">{{ t('settings.triggerTypes.issue_created') }}</option>
              <option value="issue_updated">{{ t('settings.triggerTypes.issue_updated') }}</option>
              <option value="state_changed">{{ t('settings.triggerTypes.state_changed') }}</option>
              <option value="assignee_changed">{{ t('settings.triggerTypes.assignee_changed') }}</option>
              <option value="comment_added">{{ t('settings.triggerTypes.comment_added') }}</option>
            </select>
          </div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">{{ t('settings.conditions') }} (JSON)</div>
            <textarea v-model="newAutomationForm.conditions" rows="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-xs font-mono" :placeholder="t('settings.conditionsPlaceholder')"></textarea>
          </div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <div class="text-sm font-medium text-gray-700 mb-2">{{ t('settings.actions') }} (JSON)</div>
            <textarea v-model="newAutomationForm.actions" rows="3" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-xs font-mono" :placeholder="t('settings.actionsPlaceholder')"></textarea>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showAutomationModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('settings.cancel') }}</button>
          <button @click="handleSaveAutomation" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ t('settings.create') }}</button>
        </div>
      </div>
    </div>

    <!-- Subscriber Picker Modal -->
    <div v-if="showSubscriberPicker" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showSubscriberPicker = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('settings.addSubscriber') }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('settings.selectMember') }}</label>
            <select v-model.number="subscriberPickerUserId" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
              <option :value="null" disabled>{{ t('settings.selectMemberPlaceholder') }}</option>
              <option v-for="m in members" :key="m.user_id" :value="m.user_id">
                {{ getMemberDisplay(m) }}
              </option>
            </select>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showSubscriberPicker = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('settings.cancel') }}</button>
          <button @click="handleAddSubscriber" :disabled="!subscriberPickerUserId" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition">{{ t('settings.add') }}</button>
        </div>
      </div>
    </div>

    <!-- Custom Field Form Modal -->
    <div v-if="showFieldForm" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showFieldForm = false">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[80vh] overflow-y-auto">
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h2 class="text-lg font-semibold text-gray-800">{{ editingField ? t('settings.editField') : t('settings.createField') }}</h2>
          <button @click="showFieldForm = false" class="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
        </div>
        <div class="p-6">
          <CustomFieldForm
            :field="editingField"
            :project-id="projectId"
            :workspace-id="workspaceId"
            @submit="handleFieldSaved"
            @cancel="showFieldForm = false"
          />
        </div>
      </div>
    </div>
  </div>
</template>
