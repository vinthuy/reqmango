<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-semibold text-gray-900">Work Item Templates</h1>
        <p class="text-sm text-gray-500 mt-1">Create reusable templates to quickly pre-fill work item forms</p>
      </div>
      <button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ Create Template</button>
    </div>

    <div v-if="templates.length === 0" class="bg-white rounded-xl border border-gray-200 p-12 text-center">
      <div class="text-5xl mb-4">📋</div>
      <h3 class="text-lg font-medium text-gray-900 mb-2">No templates yet</h3>
      <p class="text-sm text-gray-500 mb-4">Create your first work item template to speed up issue creation</p>
      <button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ Create your first template</button>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="tmpl in templates" :key="tmpl.id" class="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition">
        <div class="flex items-center space-x-3 mb-3">
          <div class="w-10 h-10 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-xl flex items-center justify-center text-white text-lg">📝</div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center space-x-2">
              <h3 class="font-medium text-gray-900 truncate">{{ tmpl.name }}</h3>
              <span v-if="tmpl.is_default" class="shrink-0 px-2 py-0.5 bg-indigo-100 text-indigo-600 rounded text-xs font-medium">Default</span>
            </div>
            <div class="flex items-center space-x-2 mt-1">
              <span v-if="tmpl.issue_type" class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium" :style="{ backgroundColor: tmpl.issue_type.color + '20', color: tmpl.issue_type.color }">
                {{ tmpl.issue_type.name }}
              </span>
            </div>
          </div>
        </div>
        <p v-if="tmpl.description" class="text-sm text-gray-500 mb-3 line-clamp-2">{{ tmpl.description }}</p>

        <div class="space-y-1.5 mb-3 text-xs text-gray-600">
          <div v-if="tmpl.defaults.name_prefix" class="flex items-center">
            <span class="text-gray-400 w-20">Name prefix:</span>
            <span class="font-mono bg-gray-100 px-1.5 py-0.5 rounded">{{ tmpl.defaults.name_prefix }}</span>
          </div>
          <div v-if="tmpl.defaults.priority" class="flex items-center">
            <span class="text-gray-400 w-20">Priority:</span>
            <span>{{ getPriorityLabel(tmpl.defaults.priority) }}</span>
          </div>
          <div v-if="tmpl.defaults.state_id" class="flex items-center">
            <span class="text-gray-400 w-20">State:</span>
            <span>{{ getStateName(tmpl.defaults.state_id) }}</span>
          </div>
          <div v-if="tmpl.defaults.assignee_ids && tmpl.defaults.assignee_ids.length > 0" class="flex items-center">
            <span class="text-gray-400 w-20">Assignees:</span>
            <span>{{ tmpl.defaults.assignee_ids.length }} member(s)</span>
          </div>
          <div v-if="tmpl.defaults.label_ids && tmpl.defaults.label_ids.length > 0" class="flex items-center">
            <span class="text-gray-400 w-20">Labels:</span>
            <span>{{ tmpl.defaults.label_ids.length }} label(s)</span>
          </div>
        </div>

        <div class="pt-3 border-t border-gray-100 flex space-x-2">
          <button @click="openEdit(tmpl)" class="text-xs text-indigo-600 hover:text-indigo-800">Edit</button>
          <button @click="confirmDelete(tmpl)" class="text-xs text-red-500 hover:text-red-700">Delete</button>
        </div>
      </div>
    </div>

    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="closeModal">
      <div class="bg-white rounded-xl p-6 w-full max-w-lg max-h-[85vh] overflow-y-auto">
        <h3 class="text-lg font-semibold mb-4">{{ editingTemplate ? 'Edit Template' : 'Create Template' }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1">Name *</label>
            <input v-model="form.name" class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Description</label>
            <input v-model="form.description" class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Issue Type</label>
            <select v-model.number="form.issue_type_id" class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500">
              <option :value="null">None</option>
              <option v-for="type in issueTypes" :key="type.id" :value="type.id">{{ type.name }}</option>
            </select>
          </div>
          <div class="flex items-center">
            <input v-model="form.is_default" type="checkbox" id="is_default" class="w-4 h-4 text-indigo-600 rounded focus:ring-indigo-500" />
            <label for="is_default" class="ml-2 text-sm font-medium text-gray-700">Set as default template</label>
          </div>

          <div class="border-t pt-4 mt-4">
            <h4 class="text-sm font-semibold text-gray-700 mb-3">Defaults</h4>
            <div class="space-y-3">
              <div>
                <label class="block text-sm font-medium mb-1">Name Prefix</label>
                <input v-model="form.defaults.name_prefix" placeholder="e.g., [BUG] " class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium mb-1">Priority</label>
                <select v-model="form.defaults.priority" class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-sm">
                  <option value="">None</option>
                  <option value="urgent">Urgent</option>
                  <option value="high">High</option>
                  <option value="medium">Medium</option>
                  <option value="low">Low</option>
                  <option value="none">None</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium mb-1">State</label>
                <select v-model.number="form.defaults.state_id" class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-sm">
                  <option :value="null">None</option>
                  <option v-for="state in states" :key="state.id" :value="state.id">{{ state.name }}</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium mb-1">Assignees</label>
                <div class="border rounded-lg p-2 max-h-32 overflow-y-auto space-y-1">
                  <label v-for="member in members" :key="member.user_id" class="flex items-center space-x-2 text-sm cursor-pointer hover:bg-gray-50 p-1 rounded">
                    <input type="checkbox" :value="member.user_id" v-model="form.defaults.assignee_ids" class="rounded text-indigo-600 focus:ring-indigo-500" />
                    <span>{{ member.user?.display_name || member.display_name || `User #${member.user_id}` }}</span>
                  </label>
                  <div v-if="members.length === 0" class="text-xs text-gray-400 text-center py-2">No members found</div>
                </div>
              </div>
              <div>
                <label class="block text-sm font-medium mb-1">Labels</label>
                <div class="border rounded-lg p-2 max-h-32 overflow-y-auto space-y-1">
                  <label v-for="label in labels" :key="label.id" class="flex items-center space-x-2 text-sm cursor-pointer hover:bg-gray-50 p-1 rounded">
                    <input type="checkbox" :value="label.id" v-model="form.defaults.label_ids" class="rounded text-indigo-600 focus:ring-indigo-500" />
                    <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: label.color }"></span>
                    <span>{{ label.name }}</span>
                  </label>
                  <div v-if="labels.length === 0" class="text-xs text-gray-400 text-center py-2">No labels found</div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="closeModal" class="px-4 py-2 border rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="saveTemplate" :disabled="!form.name.trim()" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ editingTemplate ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import workItemTemplateApi from '@/api/work-item-template'
import * as issueTypeApi from '@/api/issue-type'
import * as stateApi from '@/api/project-settings'
import projectApi from '@/api/project'
import { useConfirm } from '@/composables/useConfirm'
import type { WorkItemTemplate, WorkItemTemplateCreate, WorkItemTemplateUpdate } from '@/types/work-item-template'
import type { IssuePriority } from '@/types/issue'

const props = defineProps<{
  projectId: number
  workspaceId: number
}>()

const { confirm } = useConfirm()

const templates = ref<WorkItemTemplate[]>([])
const issueTypes = ref<any[]>([])
const states = ref<any[]>([])
const labels = ref<any[]>([])
const members = ref<any[]>([])
const showModal = ref(false)
const editingTemplate = ref<WorkItemTemplate | null>(null)

const form = ref<{
  name: string
  description: string
  issue_type_id: number | null
  is_default: boolean
  defaults: {
    name_prefix: string
    priority: string
    state_id: number | null
    assignee_ids: number[]
    label_ids: number[]
  }
}>({
  name: '',
  description: '',
  issue_type_id: null,
  is_default: false,
  defaults: {
    name_prefix: '',
    priority: '',
    state_id: null,
    assignee_ids: [],
    label_ids: []
  }
})

function getPriorityLabel(priority: string): string {
  const labels: Record<string, string> = {
    urgent: 'Urgent',
    high: 'High',
    medium: 'Medium',
    low: 'Low',
    none: 'None'
  }
  return labels[priority] || priority
}

function getStateName(stateId: number): string {
  const state = states.value.find((s: any) => s.id === stateId)
  return state?.name || `State #${stateId}`
}

async function loadTemplates() {
  try {
    templates.value = await workItemTemplateApi.listTemplates(props.projectId)
  } catch (e) {
    console.error('Failed to load templates:', e)
  }
}

async function loadIssueTypes() {
  try {
    issueTypes.value = await issueTypeApi.getIssueTypes(props.workspaceId, props.projectId)
  } catch (e) {
    console.error('Failed to load issue types:', e)
  }
}

async function loadStates() {
  try {
    states.value = await stateApi.listStates(props.projectId)
  } catch (e) {
    console.error('Failed to load states:', e)
  }
}

async function loadLabels() {
  try {
    labels.value = await stateApi.listLabels(props.projectId)
  } catch (e) {
    console.error('Failed to load labels:', e)
  }
}

async function loadMembers() {
  try {
    const data = await projectApi.listProjectMembers(props.projectId)
    members.value = Array.isArray(data) ? data : []
  } catch (e) {
    console.error('Failed to load members:', e)
  }
}

function openCreate() {
  editingTemplate.value = null
  form.value = {
    name: '',
    description: '',
    issue_type_id: null,
    is_default: false,
    defaults: {
      name_prefix: '',
      priority: '',
      state_id: null,
      assignee_ids: [],
      label_ids: []
    }
  }
  showModal.value = true
}

function openEdit(tmpl: WorkItemTemplate) {
  editingTemplate.value = tmpl
  form.value = {
    name: tmpl.name,
    description: tmpl.description || '',
    issue_type_id: tmpl.issue_type_id || null,
    is_default: tmpl.is_default,
    defaults: {
      name_prefix: tmpl.defaults.name_prefix || '',
      priority: tmpl.defaults.priority || '',
      state_id: tmpl.defaults.state_id || null,
      assignee_ids: tmpl.defaults.assignee_ids ? [...tmpl.defaults.assignee_ids] : [],
      label_ids: tmpl.defaults.label_ids ? [...tmpl.defaults.label_ids] : []
    }
  }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingTemplate.value = null
}

async function saveTemplate() {
  if (!form.value.name.trim()) return

  const defaults: any = {}
  if (form.value.defaults.name_prefix) defaults.name_prefix = form.value.defaults.name_prefix
  if (form.value.defaults.priority) defaults.priority = form.value.defaults.priority as IssuePriority
  if (form.value.defaults.state_id) defaults.state_id = form.value.defaults.state_id
  if (form.value.defaults.assignee_ids.length > 0) defaults.assignee_ids = form.value.defaults.assignee_ids
  if (form.value.defaults.label_ids.length > 0) defaults.label_ids = form.value.defaults.label_ids

  try {
    if (editingTemplate.value) {
      const data: WorkItemTemplateUpdate = {
        name: form.value.name,
        description: form.value.description || undefined,
        issue_type_id: form.value.issue_type_id || undefined,
        is_default: form.value.is_default,
        defaults
      }
      await workItemTemplateApi.updateTemplate(props.projectId, editingTemplate.value.id, data)
    } else {
      const data: WorkItemTemplateCreate = {
        name: form.value.name,
        description: form.value.description || undefined,
        issue_type_id: form.value.issue_type_id || undefined,
        is_default: form.value.is_default,
        defaults
      }
      await workItemTemplateApi.createTemplate(props.projectId, data)
    }
    closeModal()
    await loadTemplates()
  } catch (e) {
    console.error('Failed to save template:', e)
  }
}

async function confirmDelete(tmpl: WorkItemTemplate) {
  if (await confirm({
    title: 'Delete Template',
    message: `Are you sure you want to delete "${tmpl.name}"? This action cannot be undone.`,
    confirmText: 'Delete',
    danger: true,
  })) {
    try {
      await workItemTemplateApi.deleteTemplate(props.projectId, tmpl.id)
      await loadTemplates()
    } catch (e) {
      console.error('Failed to delete template:', e)
    }
  }
}

onMounted(async () => {
  await Promise.all([
    loadTemplates(),
    loadIssueTypes(),
    loadStates(),
    loadLabels(),
    loadMembers()
  ])
})
</script>
