<template>
  <div class="p-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <button @click="router.back()" class="text-gray-400 hover:text-gray-600">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/></svg>
        </button>
        <div>
          <h1 class="text-2xl font-bold text-gray-900">{{ squad?.name || t('ai.squad.detail.loading') }}</h1>
          <p class="text-gray-500 mt-1">{{ squad?.description }}</p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <span :class="getStatusClass(squad?.status || '')" class="px-3 py-1 text-xs font-semibold rounded-full">
          {{ getStatusText(squad?.status || '') }}
        </span>
      </div>
    </div>

    <!-- Tabs -->
    <div class="border-b border-gray-200 mb-6">
      <nav class="flex space-x-8">
        <button v-for="tab in tabs" :key="tab.key" @click="activeTab = tab.key"
          :class="activeTab === tab.key ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
          class="whitespace-nowrap py-3 px-1 border-b-2 font-medium text-sm transition-colors">
          {{ tab.label }}
        </button>
      </nav>
    </div>

    <!-- Members Tab -->
    <div v-if="activeTab === 'members'">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-800">{{ t('ai.squad.detail.memberList') }}</h2>
        <button @click="showAddMemberModal = true" class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm">
          {{ t('ai.squad.detail.addMember') }}
        </button>
      </div>
      <div v-if="squad?.members && squad.members.length > 0" class="bg-white rounded-lg shadow-sm border border-gray-200">
        <table class="w-full">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.squad.detail.agentId') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.squad.detail.role') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.status') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.squad.detail.assignedAt') }}</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="member in squad.members" :key="member.id" class="hover:bg-gray-50">
              <td class="px-6 py-4">
                <div class="flex items-center">
                  <div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center mr-3 text-sm font-medium text-blue-700">
                    {{ member.agent_id }}
                  </div>
                  <span class="text-sm text-gray-900">#{{ member.agent_id }}</span>
                </div>
              </td>
              <td class="px-6 py-4">
                <span :class="getRoleBadgeClass(member.role)" class="px-2 py-1 text-xs font-semibold rounded-full">
                  {{ member.role }}
                </span>
              </td>
              <td class="px-6 py-4">
                <span :class="member.status === 'active' ? 'text-green-600' : 'text-gray-400'" class="text-sm">
                  {{ member.status }}
                </span>
              </td>
              <td class="px-6 py-4">
                <span class="text-sm text-gray-500">{{ formatDate(member.assigned_at) }}</span>
              </td>
              <td class="px-6 py-4 text-right">
                <button @click="removeMemberConfirm(member)" class="text-gray-400 hover:text-red-600 text-sm">
                  {{ t('ai.squad.detail.remove') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="text-center py-12 text-gray-400">
        <div class="text-4xl mb-2">👥</div>
        <p>{{ t('ai.squad.detail.noMembers') }}</p>
      </div>
    </div>

    <!-- Execution Tab -->
    <div v-if="activeTab === 'execution'">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-800">{{ t('ai.squad.detail.execution') }}</h2>
        <button @click="showGoalModal = true" class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm">
          {{ t('ai.squad.detail.startExecution') }}
        </button>
      </div>

      <!-- Running Execution Card -->
      <div v-if="runningExecution" class="bg-white rounded-lg shadow-sm border border-blue-200 p-6 mb-4">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2">
            <div class="w-3 h-3 bg-blue-500 rounded-full animate-pulse"></div>
            <span class="font-semibold text-gray-900">{{ t('ai.squad.detail.running') }}</span>
          </div>
          <button @click="cancelExecutionConfirm" class="bg-red-600 hover:bg-red-700 text-white px-3 py-1.5 rounded-lg text-sm">
            {{ t('ai.squad.detail.cancelExecution') }}
          </button>
        </div>
        <p class="text-sm text-gray-600 mb-2"><strong>{{ t('ai.squad.goal') }}:</strong> {{ runningExecution.goal }}</p>
        <p class="text-sm text-gray-500"><strong>{{ t('ai.squad.detail.startedAt') }}:</strong> {{ formatDate(runningExecution.started_at || runningExecution.created_at) }}</p>
        <!-- SSE live logs -->
        <div v-if="liveLogs.length > 0" class="mt-3 bg-gray-50 rounded-lg p-3 max-h-48 overflow-y-auto">
          <div v-for="(log, idx) in liveLogs" :key="idx" class="text-xs text-gray-600 font-mono mb-1">
            {{ log }}
          </div>
        </div>
      </div>

      <div v-if="!runningExecution" class="text-center py-12 text-gray-400">
        <div class="text-4xl mb-2">🚀</div>
        <p>{{ t('ai.squad.detail.noRunningExecution') }}</p>
      </div>
    </div>

    <!-- History Tab -->
    <div v-if="activeTab === 'history'">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-800">{{ t('ai.squad.detail.history') }}</h2>
        <button @click="loadExecutions" class="text-sm text-blue-600 hover:text-blue-700">
          {{ t('common.refresh') }}
        </button>
      </div>
      <div v-if="executions.length > 0" class="space-y-3">
        <div v-for="exec in executions" :key="exec.id" class="bg-white rounded-lg shadow-sm border border-gray-200">
          <div class="flex items-center justify-between p-4 cursor-pointer" @click="toggleExecutionDetail(exec.id)">
            <div class="flex items-center gap-3">
              <div :class="getExecutionIconClass(exec.status)" class="w-8 h-8 rounded-full flex items-center justify-center text-sm">
                {{ getExecutionIcon(exec.status) }}
              </div>
              <div>
                <p class="font-medium text-gray-900">{{ exec.goal }}</p>
                <p class="text-xs text-gray-500">{{ formatDate(exec.created_at) }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <span :class="getExecutionStatusClass(exec.status)" class="px-2 py-1 text-xs font-semibold rounded-full">
                {{ getExecutionStatusText(exec.status) }}
              </span>
              <svg :class="expandedExecutions.has(exec.id) ? 'rotate-180' : ''" class="w-4 h-4 text-gray-400 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
              </svg>
            </div>
          </div>
          <!-- Expandable Details -->
          <div v-if="expandedExecutions.has(exec.id)" class="border-t border-gray-100 p-4 space-y-2">
            <div v-if="exec.started_at" class="text-sm text-gray-600">
              <strong>{{ t('ai.squad.detail.startedAt') }}:</strong> {{ formatDate(exec.started_at) }}
            </div>
            <div v-if="exec.completed_at" class="text-sm text-gray-600">
              <strong>{{ t('ai.squad.detail.completedAt') }}:</strong> {{ formatDate(exec.completed_at) }}
            </div>
            <div v-if="exec.failed_at" class="text-sm text-gray-600">
              <strong>{{ t('ai.squad.detail.failedAt') }}:</strong> {{ formatDate(exec.failed_at) }}
            </div>
            <div v-if="exec.error_info" class="text-sm text-red-600 bg-red-50 rounded p-2">
              <strong>{{ t('ai.squad.detail.error') }}:</strong> {{ exec.error_info }}
            </div>
            <div v-if="exec.logs && exec.logs.length > 0" class="mt-2">
              <p class="text-sm font-medium text-gray-700 mb-1">{{ t('ai.squad.detail.logs') }}:</p>
              <div class="bg-gray-50 rounded-lg p-3 max-h-48 overflow-y-auto">
                <div v-for="(log, idx) in exec.logs" :key="idx" class="text-xs text-gray-600 font-mono mb-1">
                  {{ typeof log === 'string' ? log : JSON.stringify(log) }}
                </div>
              </div>
            </div>
            <div v-if="exec.output_data" class="mt-2">
              <p class="text-sm font-medium text-gray-700 mb-1">{{ t('ai.squad.detail.output') }}:</p>
              <pre class="bg-gray-50 rounded-lg p-3 text-xs text-gray-600 overflow-x-auto">{{ JSON.stringify(exec.output_data, null, 2) }}</pre>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="text-center py-12 text-gray-400">
        <div class="text-4xl mb-2">📋</div>
        <p>{{ t('ai.squad.detail.noHistory') }}</p>
      </div>
    </div>

    <!-- Config Tab -->
    <div v-if="activeTab === 'config'">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-800">{{ t('ai.squad.detail.config') }}</h2>
      </div>
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 max-w-xl">
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.squad.detail.maxRetries') }}</label>
            <input v-model.number="configForm.max_retries" type="number" min="0" max="10"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            <p class="text-xs text-gray-400 mt-1">{{ t('ai.squad.detail.maxRetriesHint') }}</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.squad.detail.timeoutSeconds') }}</label>
            <input v-model.number="configForm.timeout_seconds" type="number" min="0"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
            <p class="text-xs text-gray-400 mt-1">{{ t('ai.squad.detail.timeoutSecondsHint') }}</p>
          </div>
          <div class="pt-2">
            <button @click="saveConfig" :disabled="savingConfig"
              class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm disabled:opacity-50">
              {{ savingConfig ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Goal Modal (Start Execution) -->
    <Transition name="modal">
      <div v-if="showGoalModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/50" @click="showGoalModal = false"></div>
        <div class="relative bg-white rounded-lg shadow-xl w-full max-w-lg mx-4">
          <div class="flex items-center justify-between p-4 border-b">
            <h2 class="text-lg font-semibold">{{ t('ai.squad.detail.startExecution') }}</h2>
            <button @click="showGoalModal = false" class="text-gray-400 hover:text-gray-600">✕</button>
          </div>
          <div class="p-4 space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.squad.goal') }}</label>
              <textarea v-model="executionGoal" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500" rows="4"
                :placeholder="t('ai.squad.detail.goalPlaceholder')"></textarea>
            </div>
          </div>
          <div class="flex items-center justify-end p-4 border-t space-x-2">
            <button @click="showGoalModal = false" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">{{ t('common.cancel') }}</button>
            <button @click="startExecution" :disabled="!executionGoal"
              class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg disabled:opacity-50">
              {{ t('ai.squad.detail.start') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Add Member Modal -->
    <Transition name="modal">
      <div v-if="showAddMemberModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/50" @click="showAddMemberModal = false"></div>
        <div class="relative bg-white rounded-lg shadow-xl w-full max-w-lg mx-4">
          <div class="flex items-center justify-between p-4 border-b">
            <h2 class="text-lg font-semibold">{{ t('ai.squad.detail.addMember') }}</h2>
            <button @click="showAddMemberModal = false" class="text-gray-400 hover:text-gray-600">✕</button>
          </div>
          <div class="p-4 space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.squad.detail.agentId') }}</label>
              <input v-model.number="newMember.agent_id" type="number" min="1"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                :placeholder="t('ai.squad.detail.agentIdPlaceholder')" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.squad.detail.role') }}</label>
              <input v-model="newMember.role" type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                :placeholder="t('ai.squad.detail.rolePlaceholder')" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.squad.detail.configId') }}</label>
              <input v-model.number="newMember.agent_config_id" type="number" min="1"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                :placeholder="t('ai.squad.detail.configIdPlaceholder')" />
            </div>
          </div>
          <div class="flex items-center justify-end p-4 border-t space-x-2">
            <button @click="showAddMemberModal = false" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">{{ t('common.cancel') }}</button>
            <button @click="addMember" :disabled="!newMember.agent_id || !newMember.role"
              class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg disabled:opacity-50">
              {{ t('ai.squad.detail.add') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { useSSE } from '@/composables/useSSE'
import * as squadApi from '@/api/squad'
import type { Squad, SquadExecution, SquadMember } from '@/api/squad'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { getWorkspaceId } = useWorkspaceId()
const { onEvent } = useSSE()

const workspaceId = ref(0)
const squadId = computed(() => Number(route.params.id))
const squad = ref<Squad | null>(null)
const executions = ref<SquadExecution[]>([])
const activeTab = ref<'members' | 'execution' | 'history' | 'config'>('members')

// Tabs
const tabs = computed(() => [
  { key: 'members' as const, label: t('ai.squad.detail.tabMembers') },
  { key: 'execution' as const, label: t('ai.squad.detail.tabExecution') },
  { key: 'history' as const, label: t('ai.squad.detail.tabHistory') },
  { key: 'config' as const, label: t('ai.squad.detail.tabConfig') },
])

// Execution state
const runningExecution = ref<SquadExecution | null>(null)
const showGoalModal = ref(false)
const executionGoal = ref('')
const liveLogs = ref<string[]>([])

// History state
const expandedExecutions = ref<Set<number>>(new Set())

// Config state
const configForm = reactive({ max_retries: 3, timeout_seconds: 300 })
const savingConfig = ref(false)

// Add member state
const showAddMemberModal = ref(false)
const newMember = reactive({ agent_id: 0, role: '', agent_config_id: 0 })

// SSE unsubscribe handle
let unsubSSE: (() => void) | null = null

// ─── Load Data ───────────────────────────────────────────

async function loadSquad() {
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    workspaceId.value = wsId
    squad.value = await squadApi.getSquad(wsId, squadId.value)

    // Load config from squad config
    const cfg = (squad.value as any).config
    if (cfg) {
      if (cfg.max_retries !== undefined) configForm.max_retries = cfg.max_retries
      if (cfg.timeout_seconds !== undefined) configForm.timeout_seconds = cfg.timeout_seconds
    }
  } catch (e) {
    console.error('Failed to load squad', e)
  }
}

async function loadExecutions() {
  try {
    if (!workspaceId.value) return
    const data = await squadApi.getExecutions(workspaceId.value, squadId.value)
    executions.value = data

    // Find running execution
    runningExecution.value = data.find(e => e.status === 'running') || null
  } catch (e) {
    console.error('Failed to load executions', e)
  }
}

// ─── SSE ─────────────────────────────────────────────────

function setupSSE() {
  unsubSSE = onEvent((event, data) => {
    if (!event.startsWith('squad.execution.')) return
    // Only handle events for this squad's execution
    if (data && data.squad_id !== squadId.value) return

    const execData = data as SquadExecution

    switch (event) {
      case 'squad.execution.started':
        runningExecution.value = execData
        loadExecutions()
        break
      case 'squad.execution.updated':
        if (runningExecution.value && runningExecution.value.id === execData.id) {
          runningExecution.value = execData
        }
        break
      case 'squad.execution.log':
        if (execData.logs) {
          for (const log of execData.logs) {
            liveLogs.value.push(typeof log === 'string' ? log : JSON.stringify(log))
          }
        }
        break
      case 'squad.execution.completed':
        runningExecution.value = null
        liveLogs.value = []
        loadExecutions()
        break
      case 'squad.execution.failed':
        runningExecution.value = null
        liveLogs.value = []
        loadExecutions()
        break
      case 'squad.execution.cancelled':
        runningExecution.value = null
        liveLogs.value = []
        loadExecutions()
        break
      case 'squad.execution.progress':
        if (runningExecution.value && runningExecution.value.id === execData.id) {
          runningExecution.value = execData
        }
        break
    }
  })
}

// ─── Actions ─────────────────────────────────────────────

async function startExecution() {
  if (!executionGoal.value || !workspaceId.value) return
  try {
    await squadApi.startExecution(workspaceId.value, squadId.value, { goal: executionGoal.value })
    showGoalModal.value = false
    executionGoal.value = ''
    await loadExecutions()
  } catch (e) {
    console.error('Failed to start execution', e)
  }
}

async function cancelExecutionConfirm() {
  if (!runningExecution.value || !workspaceId.value) return
  if (!confirm(t('ai.squad.detail.cancelConfirm'))) return
  try {
    await squadApi.cancelExecution(workspaceId.value, squadId.value, runningExecution.value.id)
    runningExecution.value = null
    liveLogs.value = []
    await loadExecutions()
  } catch (e) {
    console.error('Failed to cancel execution', e)
  }
}

async function addMember() {
  if (!newMember.agent_id || !newMember.role || !workspaceId.value) return
  try {
    await squadApi.addMember(workspaceId.value, squadId.value, {
      agent_id: newMember.agent_id,
      role: newMember.role,
      agent_config_id: newMember.agent_config_id || 0,
    })
    showAddMemberModal.value = false
    newMember.agent_id = 0
    newMember.role = ''
    newMember.agent_config_id = 0
    await loadSquad()
  } catch (e) {
    console.error('Failed to add member', e)
  }
}

async function removeMemberConfirm(member: SquadMember) {
  if (!confirm(t('ai.squad.detail.removeConfirm'))) return
  try {
    await squadApi.removeMember(workspaceId.value, squadId.value, member.id)
    await loadSquad()
  } catch (e) {
    console.error('Failed to remove member', e)
  }
}

async function saveConfig() {
  if (!workspaceId.value) return
  savingConfig.value = true
  try {
    await squadApi.updateSquad(workspaceId.value, squadId.value, {
      config: { max_retries: configForm.max_retries, timeout_seconds: configForm.timeout_seconds },
    })
  } catch (e) {
    console.error('Failed to save config', e)
  } finally {
    savingConfig.value = false
  }
}

function toggleExecutionDetail(id: number) {
  if (expandedExecutions.value.has(id)) {
    expandedExecutions.value.delete(id)
  } else {
    expandedExecutions.value.add(id)
  }
}

// ─── Helpers ─────────────────────────────────────────────

function getStatusClass(status: string) {
  switch (status) {
    case 'active': return 'bg-green-100 text-green-800'
    case 'inactive': return 'bg-gray-100 text-gray-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

function getStatusText(status: string) {
  switch (status) {
    case 'active': return t('common.active')
    case 'inactive': return t('common.inactive')
    default: return status
  }
}

function getRoleBadgeClass(role: string) {
  switch (role) {
    case 'leader': return 'bg-purple-100 text-purple-800'
    case 'executor': return 'bg-blue-100 text-blue-800'
    case 'reviewer': return 'bg-yellow-100 text-yellow-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

function getExecutionStatusClass(status: string) {
  switch (status) {
    case 'completed': return 'bg-green-100 text-green-800'
    case 'failed': return 'bg-red-100 text-red-800'
    case 'running': return 'bg-blue-100 text-blue-800'
    case 'cancelled': return 'bg-yellow-100 text-yellow-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

function getExecutionStatusText(status: string) {
  switch (status) {
    case 'completed': return t('ai.squad.detail.statusCompleted')
    case 'failed': return t('ai.squad.detail.statusFailed')
    case 'running': return t('ai.squad.detail.statusRunning')
    case 'cancelled': return t('ai.squad.detail.statusCancelled')
    default: return status
  }
}

function getExecutionIconClass(status: string) {
  switch (status) {
    case 'completed': return 'bg-green-100 text-green-600'
    case 'failed': return 'bg-red-100 text-red-600'
    case 'running': return 'bg-blue-100 text-blue-600'
    case 'cancelled': return 'bg-yellow-100 text-yellow-600'
    default: return 'bg-gray-100 text-gray-400'
  }
}

function getExecutionIcon(status: string) {
  switch (status) {
    case 'completed': return '✓'
    case 'failed': return '✗'
    case 'running': return '⟳'
    case 'cancelled': return '⊘'
    default: return '○'
  }
}

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString()
}

// ─── Lifecycle ───────────────────────────────────────────

onMounted(async () => {
  await loadSquad()
  await loadExecutions()
  setupSSE()
})

onUnmounted(() => {
  unsubSSE?.()
})
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
