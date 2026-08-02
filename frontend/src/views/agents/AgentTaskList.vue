<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { useSSE } from '@/composables/useSSE'
import { agentTaskApi, type AgentTaskResponse, type AgentTaskCreate } from '@/api/agent-task'
import { agentTemplateApi, type AgentTemplateResponse } from '@/api/agent-template'
import { agentConfigApi, type AgentConfigResponse } from '@/api/agent-config'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { onAgentTask } = useSSE()
const { getWorkspaceId } = useWorkspaceId()

const workspaceId = ref(0)
const tasks = ref<AgentTaskResponse[]>([])
const templates = ref<AgentTemplateResponse[]>([])
const configs = ref<AgentConfigResponse[]>([])
const loading = ref(true)
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const executingTaskId = ref<number | null>(null)
const cancelingTaskId = ref<number | null>(null)
let sseCleanup: (() => void) | null = null

const newTask = ref<AgentTaskCreate>({
  title: '',
  description: '',
  agent_template_id: undefined,
  agent_config_id: undefined,
  input_data: '{}',
  priority: 'normal'
})

const selectedTask = ref<AgentTaskResponse | null>(null)

const priorities = [
  { value: 'low', label: 'Low', color: 'bg-gray-100 text-gray-700' },
  { value: 'normal', label: 'Normal', color: 'bg-blue-100 text-blue-700' },
  { value: 'high', label: 'High', color: 'bg-amber-100 text-amber-700' },
  { value: 'critical', label: 'Critical', color: 'bg-red-100 text-red-700' }
]

const filteredTasks = computed(() => {
  return tasks.value.sort((a, b) => {
    const priorityOrder: Record<string, number> = { critical: 0, high: 1, normal: 2, low: 3 }
    const orderA = priorityOrder[a.priority] ?? 4
    const orderB = priorityOrder[b.priority] ?? 4
    if (orderA !== orderB) {
      return orderA - orderB
    }
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  })
})

async function loadData() {
  loading.value = true
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    workspaceId.value = wsId
    
    const [tasksRes, templatesRes, configsRes] = await Promise.all([
      agentTaskApi.list(wsId),
      agentTemplateApi.list(wsId),
      agentConfigApi.list(wsId)
    ])
    
    tasks.value = tasksRes || []
    templates.value = templatesRes || []
    configs.value = configsRes || []
  } catch (err) {
    console.error('Failed to load data:', err)
    tasks.value = []
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  newTask.value = {
    title: '',
    description: '',
    agent_template_id: undefined,
    agent_config_id: undefined,
    input_data: '{}',
    priority: 'normal'
  }
  showCreateModal.value = true
}

function openDetailModal(task: AgentTaskResponse) {
  selectedTask.value = task
  showDetailModal.value = true
}

async function handleCreate() {
  if (!newTask.value.title) {
    alert(t('ai.agentTask.nameRequired') || 'Title is required')
    return
  }
  try {
    await agentTaskApi.create(workspaceId.value, newTask.value)
    showCreateModal.value = false
    await loadData()
  } catch (err: any) {
    console.error('Failed to create task:', err)
    alert(err.response?.data?.message || t('ai.agentTask.createFailed') || 'Failed to create task')
  }
}

async function handleClaim(taskId: number) {
  try {
    await agentTaskApi.claim(workspaceId.value, taskId)
    await loadData()
  } catch (err: any) {
    console.error('Failed to claim task:', err)
    alert(err.response?.data?.message || 'Failed to claim task')
  }
}

async function handleExecute(taskId: number) {
  executingTaskId.value = taskId
  try {
    await agentTaskApi.start(workspaceId.value, taskId)
    await loadData()
  } catch (err: any) {
    console.error('Failed to execute task:', err)
    alert(err.response?.data?.message || t('ai.agentTask.executeFailed') || 'Failed to execute task')
  } finally {
    executingTaskId.value = null
  }
}

async function handleCancel(taskId: number) {
  if (!confirm(t('ai.agentTask.cancelConfirm') || 'Are you sure you want to cancel this task?')) {
    return
  }
  cancelingTaskId.value = taskId
  try {
    await agentTaskApi.cancel(workspaceId.value, taskId)
    await loadData()
  } catch (err: any) {
    console.error('Failed to cancel task:', err)
    alert(err.response?.data?.message || t('ai.agentTask.cancelFailed') || 'Failed to cancel task')
  } finally {
    cancelingTaskId.value = null
  }
}

async function handleDelete(taskId: number) {
  if (!confirm(t('ai.agentTask.deleteConfirm') || 'Are you sure you want to delete this task?')) {
    return
  }
  try {
    await agentTaskApi.delete(workspaceId.value, taskId)
    await loadData()
  } catch (err: any) {
    console.error('Failed to delete task:', err)
    alert(err.response?.data?.message || t('ai.agentTask.deleteFailed') || 'Failed to delete task')
  }
}

function goBack() {
  const slug = route.params.slug as string
  if (slug) {
    router.push(`/workspace/${slug}/agents`)
  } else {
    router.push(`/workspaces/${workspaceId.value}/agents`)
  }
}

function getStatusColor(status: string) {
  switch (status) {
    case 'enqueue': return 'bg-gray-100 text-gray-700'
    case 'claimed': return 'bg-blue-100 text-blue-700'
    case 'running': return 'bg-indigo-100 text-indigo-700'
    case 'completed': return 'bg-green-100 text-green-700'
    case 'failed': return 'bg-red-100 text-red-700'
    case 'cancelled': return 'bg-amber-100 text-amber-700'
    default: return 'bg-blue-100 text-blue-700'
  }
}

function getStatusText(status: string) {
  switch (status) {
    case 'enqueue': return 'Queued'
    case 'claimed': return 'Claimed'
    case 'running': return 'Running'
    case 'completed': return 'Completed'
    case 'failed': return 'Failed'
    case 'cancelled': return 'Cancelled'
    default: return status
  }
}

function getPriorityInfo(priority: string) {
  return priorities.find(p => p.value === priority) || priorities[1]
}

function formatDuration(seconds?: number) {
  if (!seconds) return '-'
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
}

function parseInputData(value: any) {
  try {
    return typeof value === 'string' ? JSON.parse(value) : value
  } catch {
    return value
  }
}

function handleTaskEvent(event: string, task: AgentTaskResponse) {
  const idx = tasks.value.findIndex(t => t.id === task.id)
  if (idx !== -1) {
    tasks.value[idx] = task
  } else if (event === 'agent_task.created') {
    tasks.value.unshift(task)
  }
  if (selectedTask.value && selectedTask.value.id === task.id) {
    selectedTask.value = task
  }
}

function setupSSE() {
  sseCleanup = onAgentTask((event, data) => {
    handleTaskEvent(event, data as AgentTaskResponse)
  })
}

onMounted(() => {
  loadData()
  setupSSE()
})

onUnmounted(() => {
  if (sseCleanup) {
    sseCleanup()
  }
})
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
            <h1 class="text-xl font-semibold text-gray-800">{{ t('ai.agentTask.title') || 'Agent Tasks' }}</h1>
            <p class="text-sm text-gray-500">{{ t('ai.agentTask.description') || 'Monitor and manage agent execution tasks' }}</p>
          </div>
        </div>
        <button
          @click="openCreateModal"
          class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors shadow-sm"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          {{ t('ai.agentTask.create') || 'Create Task' }}
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
        <div v-else-if="tasks.length === 0" class="text-center py-20">
          <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
            </svg>
          </div>
          <h3 class="text-base font-medium text-gray-900 dark:text-gray-100 mb-1">{{ t('ai.agentTask.noTasks') || 'No Tasks' }}</h3>
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{{ t('ai.agentTask.noTasksHint') || 'Create your first agent task to get started' }}</p>
          <button
            @click="openCreateModal"
            class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
          >
            {{ t('ai.agentTask.createFirst') || 'Create First Task' }}
          </button>
        </div>

        <!-- Task List -->
        <div v-else class="space-y-4">
          <div
            v-for="task in filteredTasks"
            :key="task.id"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center gap-4">
                <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-teal-500 to-cyan-500 flex items-center justify-center text-white text-lg font-semibold">
                  {{ (task.title || '').charAt(0).toUpperCase() }}
                </div>
                <div>
                  <div class="flex items-center space-x-2">
                    <h3 class="font-semibold text-gray-900">{{ task.title }}</h3>
                    <span :class="['px-2 py-0.5 rounded text-xs font-medium', getStatusColor(task.status)]">
                      {{ getStatusText(task.status) }}
                    </span>
                    <span :class="['px-2 py-0.5 rounded text-xs font-medium', getPriorityInfo(task.priority).color]">
                      {{ getPriorityInfo(task.priority).label }}
                    </span>
                  </div>
                  <div class="flex items-center space-x-3 mt-1">
                    <span v-if="task.task_type" class="text-sm text-gray-600">🎯 {{ task.task_type }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-xs text-gray-400">{{ formatDuration(task.actual_time) }}</span>
                <button
                  @click="openDetailModal(task)"
                  class="p-2 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                  :title="t('common.view') || 'View'"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                </button>
                <button
                  v-if="task.status === 'enqueue'"
                  @click="handleClaim(task.id)"
                  class="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                  :title="'Claim'"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                </button>
                <button
                  v-if="task.status === 'claimed'"
                  @click="handleExecute(task.id)"
                  :disabled="executingTaskId === task.id"
                  class="p-2 text-gray-400 hover:text-green-600 hover:bg-green-50 rounded-lg transition-colors disabled:opacity-50"
                  :title="t('ai.agentTask.execute') || 'Execute'"
                >
                  <svg v-if="executingTaskId !== task.id" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <svg v-else class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                </button>
                <button
                  v-if="task.status === 'running'"
                  @click="handleCancel(task.id)"
                  :disabled="cancelingTaskId === task.id"
                  class="p-2 text-gray-400 hover:text-amber-600 hover:bg-amber-50 rounded-lg transition-colors disabled:opacity-50"
                  :title="t('ai.agentTask.cancel') || 'Cancel'"
                >
                  <svg v-if="cancelingTaskId !== task.id" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  <svg v-else class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                </button>
                <button
                  @click="handleDelete(task.id)"
                  class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  :title="t('common.delete') || 'Delete'"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
            <p v-if="task.description" class="text-sm text-gray-500 mb-3">{{ task.description }}</p>
            <div class="flex flex-wrap gap-4 text-sm text-gray-500">
              <span>{{ t('ai.agentTask.createdAt') || 'Created' }}: {{ new Date(task.created_at).toLocaleString() }}</span>
              <span v-if="task.started_at">{{ t('ai.agentTask.startedAt') || 'Started' }}: {{ new Date(task.started_at).toLocaleString() }}</span>
              <span v-if="task.completed_at">{{ t('ai.agentTask.completedAt') || 'Completed' }}: {{ new Date(task.completed_at).toLocaleString() }}</span>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showCreateModal = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold text-gray-900">{{ t('ai.agentTask.create') || 'Create Task' }}</h3>
          <button @click="showCreateModal = false" class="text-gray-400 hover:text-gray-600">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.name') || 'Title' }}</label>
            <input
              v-model="newTask.title"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              :placeholder="t('ai.agentTask.namePlaceholder') || 'Enter task title'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.description') || 'Description' }}</label>
            <textarea
              v-model="newTask.description"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              rows="2"
              :placeholder="t('ai.agentTask.descriptionPlaceholder') || 'Enter description'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.agentTask.template') || 'Agent Template' }}</label>
            <select
              v-model.number="newTask.agent_template_id"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            >
              <option :value="undefined">{{ t('common.none') || 'None' }}</option>
              <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.agentTask.config') || 'Model Config' }}</label>
            <select
              v-model.number="newTask.agent_config_id"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            >
              <option :value="undefined">{{ t('common.none') || 'None' }}</option>
              <option v-for="c in configs" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.agentTask.inputParams') || 'Input Data (JSON)' }}</label>
            <textarea
              v-model="newTask.input_data"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              rows="3"
              :placeholder="t('ai.agentTask.inputParamsPlaceholder') || 'Enter JSON object'"
              @blur="newTask.input_data = parseInputData(newTask.input_data)"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.agentTask.priority') || 'Priority' }}</label>
            <select
              v-model="newTask.priority"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            >
              <option v-for="p in priorities" :key="p.value" :value="p.value">{{ p.label }}</option>
            </select>
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
            class="px-4 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition text-sm font-medium"
          >
            {{ t('common.create') || 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Detail Modal -->
    <div v-if="showDetailModal && selectedTask" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showDetailModal = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold text-gray-900">{{ t('ai.agentTask.detail') || 'Task Detail' }}</h3>
          <button @click="showDetailModal = false" class="text-gray-400 hover:text-gray-600">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="space-y-4">
          <div>
            <div class="text-xs text-gray-400 mb-1">{{ t('common.name') || 'Title' }}</div>
            <div class="font-semibold text-gray-900">{{ selectedTask.title }}</div>
          </div>
          <div>
            <div class="text-xs text-gray-400 mb-1">{{ t('common.description') || 'Description' }}</div>
            <div class="text-gray-700">{{ selectedTask.description || '-' }}</div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-xs text-gray-400 mb-1">{{ t('common.status') || 'Status' }}</div>
              <span :class="['px-2 py-0.5 rounded text-xs font-medium', getStatusColor(selectedTask.status)]">
                {{ getStatusText(selectedTask.status) }}
              </span>
            </div>
            <div>
              <div class="text-xs text-gray-400 mb-1">{{ t('ai.agentTask.priority') || 'Priority' }}</div>
              <span :class="['px-2 py-0.5 rounded text-xs font-medium', getPriorityInfo(selectedTask.priority).color]">
                {{ getPriorityInfo(selectedTask.priority).label }}
              </span>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-xs text-gray-400 mb-1">Task Type</div>
              <div class="text-gray-700">{{ selectedTask.task_type || '-' }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-400 mb-1">Progress</div>
              <div class="text-gray-700">{{ selectedTask.progress || 0 }}%</div>
            </div>
          </div>
          <div>
            <div class="text-xs text-gray-400 mb-1">{{ t('ai.agentTask.inputParams') || 'Input Data' }}</div>
            <pre class="text-sm text-gray-600 font-mono bg-gray-50 p-3 rounded-lg max-h-40 overflow-y-auto">{{ typeof selectedTask.input_data === 'string' ? selectedTask.input_data : JSON.stringify(selectedTask.input_data, null, 2) }}</pre>
          </div>
          <div v-if="selectedTask.output_data">
            <div class="text-xs text-gray-400 mb-1">{{ t('ai.agentTask.outputResult') || 'Output Data' }}</div>
            <pre class="text-sm text-gray-600 font-mono bg-gray-50 p-3 rounded-lg max-h-40 overflow-y-auto">{{ typeof selectedTask.output_data === 'string' ? selectedTask.output_data : JSON.stringify(selectedTask.output_data, null, 2) }}</pre>
          </div>
          <div v-if="selectedTask.error_info" class="p-3 bg-red-50 border border-red-200 rounded-lg">
            <div class="text-xs text-red-600 font-medium mb-1">{{ t('ai.agentTask.error') || 'Error' }}</div>
            <div class="text-sm text-red-700">{{ selectedTask.error_info }}</div>
          </div>
          <div class="grid grid-cols-3 gap-4 text-sm">
            <div>
              <div class="text-xs text-gray-400 mb-1">{{ t('ai.agentTask.duration') || 'Duration' }}</div>
              <div class="text-gray-700">{{ formatDuration(selectedTask.actual_time) }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-400 mb-1">Estimated</div>
              <div class="text-gray-700">{{ formatDuration(selectedTask.estimated_time) }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-400 mb-1">{{ t('ai.agentTask.createdAt') || 'Created' }}</div>
              <div class="text-gray-700">{{ new Date(selectedTask.created_at).toLocaleString() }}</div>
            </div>
          </div>
        </div>
        <div class="flex justify-end mt-6">
          <button
            @click="showDetailModal = false"
            class="px-4 py-2.5 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition text-sm font-medium"
          >
            {{ t('common.close') || 'Close' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>