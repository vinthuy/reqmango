<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { agentTemplateApi, type AgentTemplateResponse, type AgentTemplateCreate, type AgentTemplateUpdate } from '@/api/agent-template'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const workspaceId = ref(0)
const templates = ref<AgentTemplateResponse[]>([])
const loading = ref(true)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const saving = ref(false)

const newTemplate = ref<AgentTemplateCreate>({
  name: '',
  description: '',
  icon: '',
  system_prompt: '',
  available_skills: [],
  available_tools: [],
  default_config: {},
  version: '1.0.0'
})

const editingTemplate = ref<AgentTemplateResponse | null>(null)
const editForm = ref<AgentTemplateUpdate>({})

const filteredTemplates = computed(() => {
  return templates.value.sort((a, b) => {
    if (a.is_preset && !b.is_preset) return -1
    if (!a.is_preset && b.is_preset) return 1
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  })
})

async function loadTemplates() {
  loading.value = true
  try {
    const wsId = parseInt(route.params.wsParam as string, 10)
    workspaceId.value = wsId
    templates.value = await agentTemplateApi.list(wsId) || []
  } catch (err) {
    console.error('Failed to load templates:', err)
    templates.value = []
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  newTemplate.value = {
    name: '',
    description: '',
    icon: '',
    system_prompt: '',
    available_skills: [],
    available_tools: [],
    default_config: {},
    version: '1.0.0'
  }
  showCreateModal.value = true
}

function openEditModal(template: AgentTemplateResponse) {
  editingTemplate.value = template
  editForm.value = {
    name: template.name,
    description: template.description,
    icon: template.icon,
    system_prompt: template.system_prompt,
    available_skills: template.available_skills,
    available_tools: template.available_tools,
    default_config: template.default_config,
    version: template.version,
    status: template.status
  }
  showEditModal.value = true
}

async function handleCreate() {
  if (!newTemplate.value.name || !newTemplate.value.system_prompt) {
    alert(t('agentTemplate.nameAndPromptRequired') || 'Name and system prompt are required')
    return
  }
  saving.value = true
  try {
    await agentTemplateApi.create(workspaceId.value, newTemplate.value)
    showCreateModal.value = false
    await loadTemplates()
  } catch (err: any) {
    console.error('Failed to create template:', err)
    alert(err.response?.data?.message || t('agentTemplate.createFailed') || 'Failed to create template')
  } finally {
    saving.value = false
  }
}

async function handleUpdate() {
  if (!editingTemplate.value) return
  if (!editForm.value.name || !editForm.value.system_prompt) {
    alert(t('agentTemplate.nameAndPromptRequired') || 'Name and system prompt are required')
    return
  }
  saving.value = true
  try {
    await agentTemplateApi.update(workspaceId.value, editingTemplate.value.id, editForm.value)
    showEditModal.value = false
    await loadTemplates()
  } catch (err: any) {
    console.error('Failed to update template:', err)
    alert(err.response?.data?.message || t('agentTemplate.updateFailed') || 'Failed to update template')
  } finally {
    saving.value = false
  }
}

async function handleDelete(templateId: number) {
  if (!confirm(t('agentTemplate.deleteConfirm') || 'Are you sure you want to delete this template?')) {
    return
  }
  try {
    await agentTemplateApi.delete(workspaceId.value, templateId)
    await loadTemplates()
  } catch (err: any) {
    console.error('Failed to delete template:', err)
    alert(err.response?.data?.message || t('agentTemplate.deleteFailed') || 'Failed to delete template')
  }
}

function goBack() {
  router.push(`/workspaces/${workspaceId.value}/agents`)
}

function parseCommaSeparated(value: any): string[] {
  if (typeof value === 'string') {
    return value.split(',').map((s: string) => s.trim()).filter((s: string) => s)
  }
  return value || []
}

function getStatusColor(status: string) {
  switch (status) {
    case 'active': return 'bg-green-100 text-green-700'
    case 'inactive': return 'bg-gray-100 text-gray-500'
    default: return 'bg-blue-100 text-blue-700'
  }
}

function getStatusText(status: string) {
  switch (status) {
    case 'active': return t('agentTemplate.active') || 'Active'
    case 'inactive': return t('agentTemplate.inactive') || 'Inactive'
    default: return status
  }
}

onMounted(loadTemplates)
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <button @click="goBack" class="text-gray-500 hover:text-gray-700">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h1 class="text-xl font-semibold text-gray-800">{{ t('agentTemplate.title') || 'Agent Templates' }}</h1>
            <p class="text-sm text-gray-500">{{ t('agentTemplate.description') || 'Manage agent role templates for AI collaboration' }}</p>
          </div>
        </div>
        <button
          @click="openCreateModal"
          class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors shadow-sm"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          {{ t('agentTemplate.create') || 'Create Template' }}
        </button>
      </div>
    </header>

    <!-- Main Content -->
    <main class="p-6">
      <div class="max-w-5xl mx-auto">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-20">
          <div class="animate-spin rounded-full h-8 w-8 border-2 border-indigo-600 border-t-transparent"></div>
        </div>

        <!-- Empty State -->
        <div v-else-if="templates.length === 0" class="text-center py-20">
          <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
          <h3 class="text-base font-medium text-gray-900 dark:text-gray-100 mb-1">{{ t('agentTemplate.noTemplates') || 'No Templates' }}</h3>
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{{ t('agentTemplate.noTemplatesHint') || 'Create your first agent template to get started' }}</p>
          <button
            @click="openCreateModal"
            class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
          >
            {{ t('agentTemplate.createFirst') || 'Create First Template' }}
          </button>
        </div>

        <!-- Template Grid -->
        <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div
            v-for="template in filteredTemplates"
            :key="template.id"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start justify-between mb-3">
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center text-white text-xl">
                  {{ template.icon || '🤖' }}
                </div>
                <div class="min-w-0">
                  <div class="flex items-center space-x-2">
                    <h3 class="font-semibold text-gray-900 truncate">{{ template.name }}</h3>
                    <span v-if="template.is_preset" class="px-2 py-0.5 bg-blue-100 text-blue-600 rounded text-xs font-medium">
                      {{ t('agentTemplate.preset') || 'Preset' }}
                    </span>
                  </div>
                  <div class="flex items-center space-x-2 mt-0.5">
                    <span class="text-xs text-gray-400">v{{ template.version }}</span>
                    <span :class="['px-2 py-0.5 rounded text-xs font-medium', getStatusColor(template.status)]">
                      {{ getStatusText(template.status) }}
                    </span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-1">
                <button
                  v-if="!template.is_preset"
                  @click="openEditModal(template)"
                  class="p-2 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                  :title="t('common.edit') || 'Edit'"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </button>
                <button
                  v-if="!template.is_preset"
                  @click="handleDelete(template.id)"
                  class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  :title="t('common.delete') || 'Delete'"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
            <p v-if="template.description" class="text-sm text-gray-500 line-clamp-2 mb-3">{{ template.description }}</p>
            <div class="flex flex-wrap gap-1">
              <span v-for="skill in (template.available_skills || []).slice(0, 3)" :key="skill" class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-700">
                {{ skill }}
              </span>
              <span v-if="(template.available_skills || []).length > 3" class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-500">
                +{{ (template.available_skills || []).length - 3 }}
              </span>
              <span v-if="!template.available_skills || template.available_skills.length === 0" class="text-xs text-gray-400">
                {{ t('agentTemplate.noSkills') || 'No skills' }}
              </span>
            </div>
            <div class="mt-3 pt-3 border-t border-gray-100 text-xs text-gray-400">
              {{ t('agentTemplate.createdAt') || 'Created' }}: {{ new Date(template.created_at).toLocaleDateString() }}
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showCreateModal = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold text-gray-900">{{ t('agentTemplate.create') || 'Create Agent Template' }}</h3>
          <button @click="showCreateModal = false" class="text-gray-400 hover:text-gray-600">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.name') || 'Name' }}</label>
            <input
              v-model="newTemplate.name"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              :placeholder="t('agentTemplate.namePlaceholder') || 'Enter template name'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.description') || 'Description' }}</label>
            <textarea
              v-model="newTemplate.description"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              rows="2"
              :placeholder="t('agentTemplate.descriptionPlaceholder') || 'Enter description'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentTemplate.icon') || 'Icon' }}</label>
            <input
              v-model="newTemplate.icon"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              :placeholder="t('agentTemplate.iconPlaceholder') || 'Emoji or icon URL'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentTemplate.systemPrompt') || 'System Prompt' }}</label>
            <textarea
              v-model="newTemplate.system_prompt"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              rows="6"
              :placeholder="t('agentTemplate.systemPromptPlaceholder') || 'Enter system prompt...'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentTemplate.availableSkills') || 'Available Skills (comma separated)' }}</label>
            <input
              v-model="newTemplate.available_skills"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              :placeholder="t('agentTemplate.skillsPlaceholder') || 'skill1, skill2, skill3'"
              @blur="newTemplate.available_skills = parseCommaSeparated(newTemplate.available_skills)"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentTemplate.availableTools') || 'Available Tools (comma separated)' }}</label>
            <input
              v-model="newTemplate.available_tools"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              :placeholder="t('agentTemplate.toolsPlaceholder') || 'tool1, tool2, tool3'"
              @blur="newTemplate.available_tools = parseCommaSeparated(newTemplate.available_tools)"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentTemplate.version') || 'Version' }}</label>
            <input
              v-model="newTemplate.version"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              :placeholder="t('agentTemplate.versionPlaceholder') || '1.0.0'"
            />
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showCreateModal = false"
            class="px-4 py-2.5 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition text-sm font-medium"
          >
            {{ t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="handleCreate"
            :disabled="saving"
            class="px-4 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition text-sm font-medium"
          >
            {{ saving ? (t('common.saving') || 'Saving...') : (t('common.create') || 'Create') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showEditModal = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold text-gray-900">{{ t('agentTemplate.edit') || 'Edit Agent Template' }}</h3>
          <button @click="showEditModal = false" class="text-gray-400 hover:text-gray-600">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.name') || 'Name' }}</label>
            <input
              v-model="editForm.name"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.description') || 'Description' }}</label>
            <textarea
              v-model="editForm.description"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              rows="2"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentTemplate.icon') || 'Icon' }}</label>
            <input
              v-model="editForm.icon"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentTemplate.systemPrompt') || 'System Prompt' }}</label>
            <textarea
              v-model="editForm.system_prompt"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              rows="6"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentTemplate.availableSkills') || 'Available Skills (comma separated)' }}</label>
            <input
              v-model="editForm.available_skills"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              @blur="editForm.available_skills = parseCommaSeparated(editForm.available_skills)"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentTemplate.availableTools') || 'Available Tools (comma separated)' }}</label>
            <input
              v-model="editForm.available_tools"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              @blur="editForm.available_tools = parseCommaSeparated(editForm.available_tools)"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentTemplate.version') || 'Version' }}</label>
            <input
              v-model="editForm.version"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.status') || 'Status' }}</label>
            <select
              v-model="editForm.status"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            >
              <option value="active">{{ t('agentTemplate.active') || 'Active' }}</option>
              <option value="inactive">{{ t('agentTemplate.inactive') || 'Inactive' }}</option>
            </select>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showEditModal = false"
            class="px-4 py-2.5 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition text-sm font-medium"
          >
            {{ t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="handleUpdate"
            :disabled="saving"
            class="px-4 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition text-sm font-medium"
          >
            {{ saving ? (t('common.saving') || 'Saving...') : (t('common.save') || 'Save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>