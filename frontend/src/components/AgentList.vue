<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { agentApi } from '@/api/agent'
import type { Agent, AgentCreateRequest, AgentUpdateRequest } from '@/types/agent'

const props = defineProps<{ workspaceId: number }>()

const agents = ref<Agent[]>([])
const loading = ref(true)
const showModal = ref(false)
const editingAgent = ref<Agent | null>(null)
const saving = ref(false)

const form = ref<AgentCreateRequest>({
  name: '',
  avatar: '🤖',
  agent_type: 'builtin',
  capabilities: [],
  status: 'active',
  system_prompt: '',
})

const availableCapabilities = [
  { key: 'search', label: 'Search' },
  { key: 'create', label: 'Create Issues' },
  { key: 'update', label: 'Update Issues' },
  { key: 'analyze', label: 'Analyze' },
  { key: 'comment', label: 'Comment' },
  { key: 'list', label: 'List Resources' },
  { key: 'summarize', label: 'Summarize' },
]

async function fetchAgents() {
  loading.value = true
  try {
    agents.value = await agentApi.list(props.workspaceId)
  } catch (e) {
    console.error('Failed to fetch agents', e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingAgent.value = null
  form.value = { name: '', avatar: '🤖', agent_type: 'builtin', capabilities: [], status: 'active', system_prompt: '' }
  showModal.value = true
}

function openEdit(agent: Agent) {
  editingAgent.value = agent
  form.value = {
    name: agent.name,
    avatar: agent.avatar,
    agent_type: agent.agent_type,
    capabilities: agent.capabilities || [],
    status: agent.status,
    system_prompt: agent.system_prompt || '',
    model_override: agent.model_override,
  }
  showModal.value = true
}

async function save() {
  if (!form.value.name.trim()) return
  saving.value = true
  try {
    if (editingAgent.value) {
      const updateReq: AgentUpdateRequest = {
        name: form.value.name,
        avatar: form.value.avatar,
        agent_type: form.value.agent_type,
        capabilities: form.value.capabilities,
        status: form.value.status,
        system_prompt: form.value.system_prompt || undefined,
        model_override: form.value.model_override,
      }
      await agentApi.update(props.workspaceId, editingAgent.value.id, updateReq)
    } else {
      await agentApi.create(props.workspaceId, form.value)
    }
    showModal.value = false
    await fetchAgents()
  } catch (e) {
    console.error('Failed to save agent', e)
  } finally {
    saving.value = false
  }
}

async function remove(agent: Agent) {
  if (!confirm(`Delete agent "${agent.name}"?`)) return
  try {
    await agentApi.delete(props.workspaceId, agent.id)
    await fetchAgents()
  } catch (e) {
    console.error('Failed to delete agent', e)
  }
}

function toggleCapability(key: string) {
  const idx = form.value.capabilities!.indexOf(key)
  if (idx >= 0) {
    form.value.capabilities!.splice(idx, 1)
  } else {
    form.value.capabilities!.push(key)
  }
}

onMounted(fetchAgents)
</script>

<template>
  <div class="agent-list">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-200">AI Agents</h3>
      <button
        @click="openCreate"
        class="px-3 py-1.5 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
      >
        + New Agent
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center py-8">
      <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-indigo-600"></div>
    </div>

    <!-- Empty -->
    <div v-else-if="agents.length === 0" class="text-center py-8 text-gray-500">
      <p class="text-4xl mb-2">🤖</p>
      <p>No AI agents configured yet.</p>
      <p class="text-sm">Create one to enable automatic triage, assignment, and more.</p>
    </div>

    <!-- Agent Cards -->
    <div v-else class="grid gap-3">
      <div
        v-for="agent in agents"
        :key="agent.id"
        class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 hover:shadow-sm transition-shadow"
      >
        <div class="flex items-start justify-between">
          <div class="flex items-center gap-3">
            <span class="text-2xl">{{ agent.avatar }}</span>
            <div>
              <div class="flex items-center gap-2">
                <h4 class="font-medium text-gray-800 dark:text-gray-100">{{ agent.name }}</h4>
                <span
                  :class="agent.status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300' : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'"
                  class="text-xs px-1.5 py-0.5 rounded-full"
                >
                  {{ agent.status }}
                </span>
                <span class="text-xs px-1.5 py-0.5 rounded-full bg-indigo-100 text-indigo-700 dark:bg-indigo-900 dark:text-indigo-300">
                  {{ agent.agent_type }}
                </span>
              </div>
              <div class="flex flex-wrap gap-1 mt-1">
                <span
                  v-for="cap in (agent.capabilities || [])"
                  :key="cap"
                  class="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400"
                >
                  {{ cap }}
                </span>
                <span
                  v-if="!agent.capabilities || agent.capabilities.length === 0"
                  class="text-xs text-gray-400 italic"
                >
                  all capabilities
                </span>
              </div>
            </div>
          </div>
          <div class="flex gap-1">
            <button
              @click="openEdit(agent)"
              class="p-1 text-gray-400 hover:text-indigo-600 transition-colors"
              title="Edit agent"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
              </svg>
            </button>
            <button
              @click="remove(agent)"
              class="p-1 text-gray-400 hover:text-red-600 transition-colors"
              title="Delete agent"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div
        v-if="showModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @click.self="showModal = false"
      >
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md p-6 mx-4">
          <h3 class="text-lg font-semibold mb-4 text-gray-800 dark:text-gray-100">
            {{ editingAgent ? 'Edit Agent' : 'New Agent' }}
          </h3>

          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name</label>
              <input
                v-model="form.name"
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="e.g. Triage Agent"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Avatar (emoji)</label>
              <input
                v-model="form.avatar"
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100"
                maxlength="10"
              />
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Type</label>
                <select
                  v-model="form.agent_type"
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100"
                >
                  <option value="builtin">Built-in</option>
                  <option value="custom">Custom</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Status</label>
                <select
                  v-model="form.status"
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100"
                >
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Capabilities</label>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="cap in availableCapabilities"
                  :key="cap.key"
                  @click="toggleCapability(cap.key)"
                  :class="[
                    'px-2.5 py-1 text-xs rounded-full border transition-colors',
                    form.capabilities?.includes(cap.key)
                      ? 'bg-indigo-100 border-indigo-300 text-indigo-700 dark:bg-indigo-900 dark:border-indigo-700 dark:text-indigo-300'
                      : 'bg-gray-50 border-gray-200 text-gray-600 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-400 hover:border-indigo-300'
                  ]"
                >
                  {{ cap.label }}
                </button>
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                System Prompt <span class="text-gray-400 font-normal">(optional)</span>
              </label>
              <textarea
                v-model="form.system_prompt"
                rows="3"
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 text-sm"
                placeholder="Custom instructions for this agent..."
              ></textarea>
            </div>
          </div>

          <div class="flex justify-end gap-3 mt-6">
            <button
              @click="showModal = false"
              class="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200"
            >
              Cancel
            </button>
            <button
              @click="save"
              :disabled="saving || !form.name.trim()"
              class="px-4 py-2 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition-colors"
            >
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
