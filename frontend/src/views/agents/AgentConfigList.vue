<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { agentConfigApi, type AgentConfigResponse, type AgentConfigCreate, type AgentConfigUpdate } from '@/api/agent-config'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const workspaceId = ref(0)
const configs = ref<AgentConfigResponse[]>([])
const loading = ref(true)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const saving = ref(false)

const newConfig = ref<AgentConfigCreate>({
  name: '',
  description: '',
  provider: 'openai',
  model: '',
  api_key: '',
  api_endpoint: '',
  inference_level: 'standard',
  service_level: 'standard',
  max_tokens: 4096,
  temperature: 0.7,
  top_p: 0.9,
  is_default: false
})

const editingConfig = ref<AgentConfigResponse | null>(null)
const editForm = ref<AgentConfigUpdate>({})

const providers = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'google', label: 'Google' },
  { value: 'azure', label: 'Azure OpenAI' },
  { value: 'local', label: 'Local' }
]

const inferenceLevels = [
  { value: 'standard', label: 'Standard' },
  { value: 'turbo', label: 'Turbo' },
  { value: 'premium', label: 'Premium' }
]

const serviceLevels = [
  { value: 'standard', label: 'Standard' },
  { value: 'priority', label: 'Priority' },
  { value: 'dedicated', label: 'Dedicated' }
]

const filteredConfigs = computed(() => {
  return configs.value.sort((a, b) => {
    if (a.is_default && !b.is_default) return -1
    if (!a.is_default && b.is_default) return 1
    if (a.is_active && !b.is_active) return -1
    if (!a.is_active && b.is_active) return 1
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  })
})

async function loadConfigs() {
  loading.value = true
  try {
    const wsId = parseInt(route.params.wsParam as string, 10)
    workspaceId.value = wsId
    configs.value = await agentConfigApi.list(wsId) || []
  } catch (err) {
    console.error('Failed to load configs:', err)
    configs.value = []
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  newConfig.value = {
    name: '',
    description: '',
    provider: 'openai',
    model: '',
    api_key: '',
    api_endpoint: '',
    inference_level: 'standard',
    service_level: 'standard',
    max_tokens: 4096,
    temperature: 0.7,
    top_p: 0.9,
    is_default: false
  }
  showCreateModal.value = true
}

function openEditModal(config: AgentConfigResponse) {
  editingConfig.value = config
  editForm.value = {
    name: config.name,
    description: config.description,
    provider: config.provider,
    model: config.model,
    api_endpoint: config.api_endpoint,
    inference_level: config.inference_level,
    service_level: config.service_level,
    max_tokens: config.max_tokens,
    temperature: config.temperature,
    top_p: config.top_p,
    is_default: config.is_default,
    is_active: config.is_active
  }
  showEditModal.value = true
}

async function handleCreate() {
  if (!newConfig.value.name || !newConfig.value.model || !newConfig.value.api_key) {
    alert(t('agentConfig.nameModelKeyRequired') || 'Name, model and API key are required')
    return
  }
  saving.value = true
  try {
    await agentConfigApi.create(workspaceId.value, newConfig.value)
    showCreateModal.value = false
    await loadConfigs()
  } catch (err: any) {
    console.error('Failed to create config:', err)
    alert(err.response?.data?.message || t('agentConfig.createFailed') || 'Failed to create config')
  } finally {
    saving.value = false
  }
}

async function handleUpdate() {
  if (!editingConfig.value) return
  if (!editForm.value.name || !editForm.value.model) {
    alert(t('agentConfig.nameModelRequired') || 'Name and model are required')
    return
  }
  saving.value = true
  try {
    await agentConfigApi.update(workspaceId.value, editingConfig.value.id, editForm.value)
    showEditModal.value = false
    await loadConfigs()
  } catch (err: any) {
    console.error('Failed to update config:', err)
    alert(err.response?.data?.message || t('agentConfig.updateFailed') || 'Failed to update config')
  } finally {
    saving.value = false
  }
}

async function handleDelete(configId: number) {
  if (!confirm(t('agentConfig.deleteConfirm') || 'Are you sure you want to delete this config?')) {
    return
  }
  try {
    await agentConfigApi.delete(workspaceId.value, configId)
    await loadConfigs()
  } catch (err: any) {
    console.error('Failed to delete config:', err)
    alert(err.response?.data?.message || t('agentConfig.deleteFailed') || 'Failed to delete config')
  }
}

function goBack() {
  router.push(`/workspaces/${workspaceId.value}/agents`)
}

function getStatusColor(status: boolean) {
  return status ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
}

function getStatusText(status: boolean) {
  return status ? (t('agentConfig.active') || 'Active') : (t('agentConfig.inactive') || 'Inactive')
}

function getProviderLabel(provider: string) {
  return providers.find(p => p.value === provider)?.label || provider
}

function getInferenceLabel(level: string) {
  return inferenceLevels.find(l => l.value === level)?.label || level
}

function getServiceLabel(level: string) {
  return serviceLevels.find(l => l.value === level)?.label || level
}

onMounted(loadConfigs)
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
            <h1 class="text-xl font-semibold text-gray-800">{{ t('agentConfig.title') || 'Agent Configs' }}</h1>
            <p class="text-sm text-gray-500">{{ t('agentConfig.description') || 'Manage AI model configurations for agents' }}</p>
          </div>
        </div>
        <button
          @click="openCreateModal"
          class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors shadow-sm"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          {{ t('agentConfig.create') || 'Create Config' }}
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
        <div v-else-if="configs.length === 0" class="text-center py-20">
          <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
            </svg>
          </div>
          <h3 class="text-base font-medium text-gray-900 dark:text-gray-100 mb-1">{{ t('agentConfig.noConfigs') || 'No Configurations' }}</h3>
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{{ t('agentConfig.noConfigsHint') || 'Create your first model configuration to get started' }}</p>
          <button
            @click="openCreateModal"
            class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
          >
            {{ t('agentConfig.createFirst') || 'Create First Config' }}
          </button>
        </div>

        <!-- Config List -->
        <div v-else class="space-y-4">
          <div
            v-for="config in filteredConfigs"
            :key="config.id"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center gap-4">
                <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center text-white text-lg font-semibold">
                  {{ config.provider.charAt(0).toUpperCase() }}
                </div>
                <div>
                  <div class="flex items-center space-x-2">
                    <h3 class="font-semibold text-gray-900">{{ config.name }}</h3>
                    <span v-if="config.is_default" class="px-2 py-0.5 bg-amber-100 text-amber-700 rounded text-xs font-medium">
                      {{ t('agentConfig.default') || 'Default' }}
                    </span>
                  </div>
                  <div class="flex items-center space-x-3 mt-1">
                    <span class="text-sm text-gray-600">{{ getProviderLabel(config.provider) }}</span>
                    <span class="text-gray-400">|</span>
                    <span class="text-sm text-gray-600 font-mono">{{ config.model }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <span :class="['px-3 py-1 rounded-full text-xs font-medium', getStatusColor(config.is_active)]">
                  {{ getStatusText(config.is_active) }}
                </span>
                <button
                  @click="openEditModal(config)"
                  class="p-2 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                  :title="t('common.edit') || 'Edit'"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </button>
                <button
                  @click="handleDelete(config.id)"
                  class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  :title="t('common.delete') || 'Delete'"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>

            <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div>
                <div class="text-xs text-gray-400 mb-1">{{ t('agentConfig.inferenceLevel') || 'Inference' }}</div>
                <div class="text-gray-700 font-medium">{{ getInferenceLabel(config.inference_level) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-400 mb-1">{{ t('agentConfig.serviceLevel') || 'Service' }}</div>
                <div class="text-gray-700 font-medium">{{ getServiceLabel(config.service_level) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-400 mb-1">{{ t('agentConfig.maxTokens') || 'Max Tokens' }}</div>
                <div class="text-gray-700 font-medium">{{ config.max_tokens.toLocaleString() }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-400 mb-1">{{ t('agentConfig.temperature') || 'Temperature' }}</div>
                <div class="text-gray-700 font-medium">{{ config.temperature }}</div>
              </div>
            </div>

            <div v-if="config.description" class="mt-4 pt-4 border-t border-gray-100">
              <p class="text-sm text-gray-500">{{ config.description }}</p>
            </div>
            <div class="mt-4 pt-4 border-t border-gray-100 text-xs text-gray-400">
              {{ t('agentConfig.createdAt') || 'Created' }}: {{ new Date(config.created_at).toLocaleString() }}
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showCreateModal = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold text-gray-900">{{ t('agentConfig.create') || 'Create Config' }}</h3>
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
              v-model="newConfig.name"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              :placeholder="t('agentConfig.namePlaceholder') || 'Enter config name'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.description') || 'Description' }}</label>
            <textarea
              v-model="newConfig.description"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              rows="2"
              :placeholder="t('agentConfig.descriptionPlaceholder') || 'Enter description'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.provider') || 'Provider' }}</label>
            <select
              v-model="newConfig.provider"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            >
              <option v-for="p in providers" :key="p.value" :value="p.value">{{ p.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.model') || 'Model' }}</label>
            <input
              v-model="newConfig.model"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              :placeholder="t('agentConfig.modelPlaceholder') || 'gpt-4o'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.apiKey') || 'API Key' }}</label>
            <input
              v-model="newConfig.api_key"
              type="password"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              :placeholder="t('agentConfig.apiKeyPlaceholder') || 'sk-...'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.apiEndpoint') || 'API Endpoint (optional)' }}</label>
            <input
              v-model="newConfig.api_endpoint"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              :placeholder="t('agentConfig.apiEndpointPlaceholder') || 'https://api.openai.com/v1'"
            />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.inferenceLevel') || 'Inference Level' }}</label>
              <select
                v-model="newConfig.inference_level"
                class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              >
                <option v-for="l in inferenceLevels" :key="l.value" :value="l.value">{{ l.label }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.serviceLevel') || 'Service Level' }}</label>
              <select
                v-model="newConfig.service_level"
                class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              >
                <option v-for="l in serviceLevels" :key="l.value" :value="l.value">{{ l.label }}</option>
              </select>
            </div>
          </div>
          <div class="grid grid-cols-3 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.maxTokens') || 'Max Tokens' }}</label>
              <input
                v-model.number="newConfig.max_tokens"
                type="number"
                class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.temperature') || 'Temperature' }}</label>
              <input
                v-model.number="newConfig.temperature"
                type="number"
                step="0.1"
                min="0"
                max="2"
                class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.topP') || 'Top P' }}</label>
              <input
                v-model.number="newConfig.top_p"
                type="number"
                step="0.1"
                min="0"
                max="1"
                class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              />
            </div>
          </div>
          <div class="flex items-center">
            <input
              v-model="newConfig.is_default"
              type="checkbox"
              class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
            />
            <label class="ml-2 text-sm text-gray-700">{{ t('agentConfig.setAsDefault') || 'Set as default config' }}</label>
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
          <h3 class="text-lg font-semibold text-gray-900">{{ t('agentConfig.edit') || 'Edit Config' }}</h3>
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
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.provider') || 'Provider' }}</label>
            <select
              v-model="editForm.provider"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            >
              <option v-for="p in providers" :key="p.value" :value="p.value">{{ p.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.model') || 'Model' }}</label>
            <input
              v-model="editForm.model"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.apiKey') || 'API Key (leave empty to keep current)' }}</label>
            <input
              v-model="editForm.api_key"
              type="password"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              :placeholder="t('agentConfig.apiKeyPlaceholder') || 'sk-...'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.apiEndpoint') || 'API Endpoint' }}</label>
            <input
              v-model="editForm.api_endpoint"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
            />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.inferenceLevel') || 'Inference Level' }}</label>
              <select
                v-model="editForm.inference_level"
                class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              >
                <option v-for="l in inferenceLevels" :key="l.value" :value="l.value">{{ l.label }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.serviceLevel') || 'Service Level' }}</label>
              <select
                v-model="editForm.service_level"
                class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              >
                <option v-for="l in serviceLevels" :key="l.value" :value="l.value">{{ l.label }}</option>
              </select>
            </div>
          </div>
          <div class="grid grid-cols-3 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.maxTokens') || 'Max Tokens' }}</label>
              <input
                v-model.number="editForm.max_tokens"
                type="number"
                class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.temperature') || 'Temperature' }}</label>
              <input
                v-model.number="editForm.temperature"
                type="number"
                step="0.1"
                min="0"
                max="2"
                class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('agentConfig.topP') || 'Top P' }}</label>
              <input
                v-model.number="editForm.top_p"
                type="number"
                step="0.1"
                min="0"
                max="1"
                class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              />
            </div>
          </div>
          <div class="flex items-center space-x-4">
            <label class="flex items-center">
              <input
                v-model="editForm.is_default"
                type="checkbox"
                class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <span class="ml-2 text-sm text-gray-700">{{ t('agentConfig.setAsDefault') || 'Set as default' }}</span>
            </label>
            <label class="flex items-center">
              <input
                v-model="editForm.is_active"
                type="checkbox"
                class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <span class="ml-2 text-sm text-gray-700">{{ t('agentConfig.isActive') || 'Active' }}</span>
            </label>
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