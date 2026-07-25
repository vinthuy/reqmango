<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { useSSE } from '@/composables/useSSE'
import { agentTaskApi, type AgentTaskResponse } from '@/api/agent-task'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { onAgentTask } = useSSE()

const workspaceId = ref(0)
const tasks = ref<AgentTaskResponse[]>([])
const loading = ref(true)
const selectedTask = ref<AgentTaskResponse | null>(null)
const autoRefresh = ref(true)
const sseConnected = ref(false)
const showDetailPanel = ref(false)
let refreshInterval: ReturnType<typeof setInterval> | null = null
let sseCleanup: (() => void) | null = null

const statusFilter = ref<string>('all')

const statusOptions = [
  { value: 'all', label: 'All', color: 'bg-gray-100 text-gray-700' },
  { value: 'enqueue', label: 'Queued', color: 'bg-gray-100 text-gray-700' },
  { value: 'claimed', label: 'Claimed', color: 'bg-blue-100 text-blue-700' },
  { value: 'running', label: 'Running', color: 'bg-indigo-100 text-indigo-700' },
  { value: 'completed', label: 'Completed', color: 'bg-green-100 text-green-700' },
  { value: 'failed', label: 'Failed', color: 'bg-red-100 text-red-700' },
  { value: 'cancelled', label: 'Cancelled', color: 'bg-amber-100 text-amber-700' }
]

const filteredTasks = computed(() => {
  let filtered = tasks.value
  if (statusFilter.value !== 'all') {
    filtered = filtered.filter(t => t.status === statusFilter.value)
  }
  return filtered.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
})

const stats = computed(() => {
  const total = tasks.value.length
  const running = tasks.value.filter(t => t.status === 'running').length
  const completed = tasks.value.filter(t => t.status === 'completed').length
  const failed = tasks.value.filter(t => t.status === 'failed').length
  const pending = tasks.value.filter(t => t.status === 'enqueue' || t.status === 'claimed').length
  const successRate = (completed + failed) > 0 ? ((completed / (completed + failed)) * 100).toFixed(0) : '0'

  return {
    total,
    running,
    completed,
    failed,
    pending,
    avgDuration: 0,
    successRate
  }
})

async function loadData() {
  loading.value = true
  try {
    const wsId = parseInt(route.params.wsParam as string, 10)
    workspaceId.value = wsId
    tasks.value = await agentTaskApi.list(wsId) || []
  } catch (err) {
    console.error('Failed to load tasks:', err)
    tasks.value = []
  } finally {
    loading.value = false
  }
}

function handleTaskEvent(event: string, task: AgentTaskResponse) {
  const idx = tasks.value.findIndex(t => t.id === task.id)

  if (event === 'agent_task.created') {
    if (idx === -1) {
      tasks.value.unshift(task)
    }
  } else if (idx !== -1) {
    tasks.value[idx] = task
  } else {
    // Task not in list, add it
    tasks.value.unshift(task)
  }

  // Update selected task if it's the one being updated
  if (selectedTask.value && selectedTask.value.id === task.id) {
    selectedTask.value = task
  }
}

function setupSSE() {
  sseCleanup = onAgentTask((event, data) => {
    handleTaskEvent(event, data as AgentTaskResponse)
  })
  sseConnected.value = true
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

function startAutoRefresh() {
  stopAutoRefresh()
  refreshInterval = setInterval(loadData, 5000)
}

function stopAutoRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

function openDetail(task: AgentTaskResponse) {
  selectedTask.value = task
  showDetailPanel.value = true
}

function closeDetail() {
  showDetailPanel.value = false
  selectedTask.value = null
}

function goBack() {
  router.push(`/workspaces/${workspaceId.value}/agents`)
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

function formatDuration(seconds?: number) {
  if (!seconds) return '-'
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
}

function formatDateTime(dateStr?: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

onMounted(() => {
  loadData()
  startAutoRefresh()
  setupSSE()
})

onUnmounted(() => {
  stopAutoRefresh()
  if (sseCleanup) {
    sseCleanup()
  }
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex">
    <!-- Main Content -->
    <div class="flex-1">
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
              <h1 class="text-xl font-semibold text-gray-800">{{ t('ai.agentMonitor.title') || 'Agent Monitor' }}</h1>
              <p class="text-sm text-gray-500">{{ t('ai.agentMonitor.description') || 'Real-time monitoring for agent tasks' }}</p>
            </div>
          </div>
          <div class="flex items-center gap-3">
            <button
              @click="toggleAutoRefresh"
              :class="[
                'inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                autoRefresh ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-700'
              ]"
            >
              <svg v-if="autoRefresh" class="w-4 h-4 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              {{ autoRefresh ? (t('ai.agentMonitor.autoRefreshOn') || 'Auto-refresh On') : (t('ai.agentMonitor.autoRefreshOff') || 'Auto-refresh Off') }}
            </button>
            <button
              @click="loadData"
              class="inline-flex items-center gap-2 px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-200 transition-colors"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              {{ t('common.refresh') || 'Refresh' }}
            </button>
          </div>
        </div>
      </header>

      <!-- Stats Cards -->
      <main class="p-6">
        <div class="max-w-7xl mx-auto">
          <!-- Stats Grid -->
          <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4 mb-6">
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-2xl font-bold text-gray-900">{{ stats.total }}</div>
                  <div class="text-xs text-gray-500 mt-1">{{ t('ai.agentMonitor.totalTasks') || 'Total Tasks' }}</div>
                </div>
                <div class="w-10 h-10 rounded-lg bg-gray-100 flex items-center justify-center">
                  <svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
                  </svg>
                </div>
              </div>
            </div>
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-2xl font-bold text-indigo-600">{{ stats.running }}</div>
                  <div class="text-xs text-gray-500 mt-1">{{ t('ai.agentMonitor.running') || 'Running' }}</div>
                </div>
                <div class="w-10 h-10 rounded-lg bg-indigo-100 flex items-center justify-center">
                  <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
            </div>
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-2xl font-bold text-green-600">{{ stats.completed }}</div>
                  <div class="text-xs text-gray-500 mt-1">{{ t('ai.agentMonitor.completed') || 'Completed' }}</div>
                </div>
                <div class="w-10 h-10 rounded-lg bg-green-100 flex items-center justify-center">
                  <svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
            </div>
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-2xl font-bold text-red-600">{{ stats.failed }}</div>
                  <div class="text-xs text-gray-500 mt-1">{{ t('ai.agentMonitor.failed') || 'Failed' }}</div>
                </div>
                <div class="w-10 h-10 rounded-lg bg-red-100 flex items-center justify-center">
                  <svg class="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
            </div>
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-2xl font-bold text-blue-600">{{ stats.successRate }}%</div>
                  <div class="text-xs text-gray-500 mt-1">{{ t('ai.agentMonitor.successRate') || 'Success Rate' }}</div>
                </div>
                <div class="w-10 h-10 rounded-lg bg-blue-100 flex items-center justify-center">
                  <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                </div>
              </div>
            </div>
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-2xl font-bold text-purple-600">{{ formatDuration(stats.avgDuration) }}</div>
                  <div class="text-xs text-gray-500 mt-1">{{ t('ai.agentMonitor.avgDuration') || 'Avg Duration' }}</div>
                </div>
                <div class="w-10 h-10 rounded-lg bg-purple-100 flex items-center justify-center">
                  <svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <!-- Filter Bar -->
          <div class="flex items-center gap-3 mb-4">
            <span class="text-sm font-medium text-gray-700">{{ t('common.filter') || 'Filter:' }}</span>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="option in statusOptions"
                :key="option.value"
                @click="statusFilter = option.value"
                :class="[
                  'px-3 py-1.5 rounded-full text-xs font-medium transition-colors',
                  statusFilter === option.value ? option.color : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                ]"
              >
                {{ option.label }}
              </button>
            </div>
          </div>

          <!-- Loading -->
          <div v-if="loading" class="flex items-center justify-center py-20">
            <div class="animate-spin rounded-full h-8 w-8 border-2 border-indigo-600 border-t-transparent"></div>
          </div>

          <!-- Empty State -->
          <div v-else-if="filteredTasks.length === 0" class="text-center py-20">
            <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gray-100 flex items-center justify-center">
              <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
              </svg>
            </div>
            <h3 class="text-base font-medium text-gray-900 mb-1">{{ t('ai.agentMonitor.noTasks') || 'No Tasks' }}</h3>
            <p class="text-sm text-gray-500">{{ t('ai.agentMonitor.noTasksHint') || 'No agent tasks found for the selected filter' }}</p>
          </div>

          <!-- Task List -->
          <div v-else class="space-y-3">
            <div
              v-for="task in filteredTasks"
              :key="task.id"
              @click="openDetail(task)"
              class="bg-white border border-gray-200 rounded-xl p-4 hover:border-indigo-300 hover:shadow-md cursor-pointer transition-all"
            >
              <div class="flex items-start justify-between">
                <div class="flex items-center gap-4">
                  <div :class="[
                    'w-10 h-10 rounded-lg flex items-center justify-center text-white font-semibold',
                    task.status === 'running' ? 'bg-indigo-600' :
                    task.status === 'completed' ? 'bg-green-600' :
                    task.status === 'failed' ? 'bg-red-600' :
                    task.status === 'cancelled' ? 'bg-amber-600' : 'bg-gray-500'
                  ]">
                    {{ (task.title || '').charAt(0).toUpperCase() }}
                  </div>
                  <div>
                    <div class="flex items-center space-x-2">
                      <h3 class="font-semibold text-gray-900">{{ task.title }}</h3>
                      <span :class="['px-2 py-0.5 rounded text-xs font-medium', getStatusColor(task.status)]">
                        {{ getStatusText(task.status) }}
                      </span>
                    </div>
                    <div class="flex items-center space-x-3 mt-1 text-sm text-gray-500">
                      <span v-if="task.task_type">🎯 {{ task.task_type }}</span>
                      <span v-if="task.priority">⚡ {{ task.priority }}</span>
                      <span>{{ formatDuration(task.actual_time) }}</span>
                    </div>
                  </div>
                </div>
                <div class="text-right">
                  <div class="text-xs text-gray-400">{{ formatDateTime(task.created_at) }}</div>
                  <div v-if="task.started_at" class="text-xs text-gray-400 mt-0.5">{{ t('ai.agentMonitor.startedAt') || 'Started' }}: {{ formatDateTime(task.started_at) }}</div>
                </div>
              </div>
              <!-- Progress Bar for Running Tasks -->
              <div v-if="task.status === 'running'" class="mt-4">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs text-gray-500">{{ t('ai.agentMonitor.progress') || 'Progress' }}</span>
                  <span class="text-xs text-indigo-600 font-medium">Processing...</span>
                </div>
                <div class="h-2 bg-gray-100 rounded-full overflow-hidden">
                  <div class="h-full bg-indigo-600 animate-pulse rounded-full" style="width: 40%"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- Detail Panel -->
    <div v-if="showDetailPanel && selectedTask" class="w-full max-w-lg bg-white border-l border-gray-200 p-6 overflow-y-auto">
      <div class="flex items-center justify-between mb-6">
        <h3 class="text-lg font-semibold text-gray-900">{{ t('ai.agentMonitor.taskDetail') || 'Task Detail' }}</h3>
        <button @click="closeDetail" class="text-gray-400 hover:text-gray-600">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="space-y-4">
        <!-- Basic Info -->
        <div class="p-4 bg-gray-50 rounded-xl">
          <div class="flex items-center gap-3 mb-3">
            <div :class="[
              'w-10 h-10 rounded-lg flex items-center justify-center text-white font-semibold',
              selectedTask.status === 'running' ? 'bg-indigo-600' :
              selectedTask.status === 'completed' ? 'bg-green-600' :
              selectedTask.status === 'failed' ? 'bg-red-600' :
              selectedTask.status === 'cancelled' ? 'bg-amber-600' : 'bg-gray-500'
            ]">
              {{ (selectedTask.title || '').charAt(0).toUpperCase() }}
            </div>
            <div>
              <h4 class="font-semibold text-gray-900">{{ selectedTask.title }}</h4>
              <span :class="['px-2 py-0.5 rounded text-xs font-medium', getStatusColor(selectedTask.status)]">
                {{ getStatusText(selectedTask.status) }}
              </span>
            </div>
          </div>
          <p v-if="selectedTask.description" class="text-sm text-gray-600">{{ selectedTask.description }}</p>
        </div>

        <!-- Metadata -->
        <div class="grid grid-cols-2 gap-3">
          <div class="p-3 bg-gray-50 rounded-lg">
            <div class="text-xs text-gray-400 mb-1">Task Type</div>
            <div class="text-sm font-medium text-gray-900">{{ selectedTask.task_type || '-' }}</div>
          </div>
          <div class="p-3 bg-gray-50 rounded-lg">
            <div class="text-xs text-gray-400 mb-1">Priority</div>
            <div class="text-sm font-medium text-gray-900">{{ selectedTask.priority || '-' }}</div>
          </div>
          <div class="p-3 bg-gray-50 rounded-lg">
            <div class="text-xs text-gray-400 mb-1">{{ t('ai.agentMonitor.duration') || 'Duration' }}</div>
            <div class="text-sm font-medium text-gray-900">{{ formatDuration(selectedTask.actual_time) }}</div>
          </div>
          <div class="p-3 bg-gray-50 rounded-lg">
            <div class="text-xs text-gray-400 mb-1">Estimated</div>
            <div class="text-sm font-medium text-gray-900">{{ formatDuration(selectedTask.estimated_time) }}</div>
          </div>
        </div>

        <!-- Timestamps -->
        <div class="p-4 bg-gray-50 rounded-xl space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">{{ t('ai.agentMonitor.createdAt') || 'Created At' }}</span>
            <span class="text-gray-900">{{ formatDateTime(selectedTask.created_at) }}</span>
          </div>
          <div v-if="selectedTask.started_at" class="flex justify-between text-sm">
            <span class="text-gray-500">{{ t('ai.agentMonitor.startedAt') || 'Started At' }}</span>
            <span class="text-gray-900">{{ formatDateTime(selectedTask.started_at) }}</span>
          </div>
          <div v-if="selectedTask.completed_at" class="flex justify-between text-sm">
            <span class="text-gray-500">{{ t('ai.agentMonitor.completedAt') || 'Completed At' }}</span>
            <span class="text-gray-900">{{ formatDateTime(selectedTask.completed_at) }}</span>
          </div>
        </div>

        <!-- Input Params -->
        <div v-if="selectedTask.input_data">
          <div class="text-sm font-medium text-gray-700 mb-2">{{ t('ai.agentMonitor.inputParams') || 'Input Data' }}</div>
          <pre class="text-xs text-gray-600 font-mono bg-gray-50 p-3 rounded-lg max-h-40 overflow-y-auto">{{ typeof selectedTask.input_data === 'string' ? selectedTask.input_data : JSON.stringify(selectedTask.input_data, null, 2) }}</pre>
        </div>

        <!-- Output Result -->
        <div v-if="selectedTask.output_data">
          <div class="text-sm font-medium text-gray-700 mb-2">{{ t('ai.agentMonitor.outputResult') || 'Output Result' }}</div>
          <pre class="text-xs text-gray-600 font-mono bg-gray-50 p-3 rounded-lg max-h-60 overflow-y-auto">{{ typeof selectedTask.output_data === 'string' ? selectedTask.output_data : JSON.stringify(selectedTask.output_data, null, 2) }}</pre>
        </div>

        <!-- Error Message -->
        <div v-if="selectedTask.error_info" class="p-4 bg-red-50 border border-red-200 rounded-xl">
          <div class="text-sm font-medium text-red-700 mb-2">{{ t('ai.agentMonitor.error') || 'Error' }}</div>
          <pre class="text-xs text-red-600 font-mono max-h-40 overflow-y-auto">{{ selectedTask.error_info }}</pre>
        </div>

        <!-- Progress -->
        <div v-if="selectedTask.progress" class="mb-4">
          <div class="flex items-center justify-between mb-1">
            <span class="text-sm font-medium text-gray-700">{{ t('ai.agentMonitor.progress') || 'Progress' }}</span>
            <span class="text-sm text-indigo-600 font-medium">{{ selectedTask.progress }}%</span>
          </div>
          <div class="h-2 bg-gray-100 rounded-full overflow-hidden">
            <div class="h-full bg-indigo-600 rounded-full transition-all" :style="{ width: selectedTask.progress + '%' }"></div>
          </div>
        </div>

        <!-- Execution Log -->
        <div v-if="selectedTask.logs && selectedTask.logs.length > 0">
          <div class="text-sm font-medium text-gray-700 mb-2">{{ t('ai.agentMonitor.executionLog') || 'Execution Log' }}</div>
          <pre class="text-xs text-gray-600 font-mono bg-gray-50 p-3 rounded-lg max-h-60 overflow-y-auto">{{ typeof selectedTask.logs === 'string' ? selectedTask.logs : JSON.stringify(selectedTask.logs, null, 2) }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>