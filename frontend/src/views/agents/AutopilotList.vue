<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('ai.autopilot.title') }}</h1>
        <p class="text-gray-500 mt-1">{{ t('ai.autopilot.description') }}</p>
      </div>
      <button @click="openCreateModal" class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg">
        {{ t('ai.autopilot.create') }}
      </button>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('common.total') }}</p>
            <p class="text-2xl font-bold text-gray-900">{{ tasks.length }}</p>
          </div>
          <div class="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">⏰</div>
        </div>
      </div>
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('ai.autopilot.enabled') }}</p>
            <p class="text-2xl font-bold text-green-600">{{ enabledCount }}</p>
          </div>
          <div class="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">✅</div>
        </div>
      </div>
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('ai.autopilot.cron') }}</p>
            <p class="text-2xl font-bold text-purple-600">{{ cronCount }}</p>
          </div>
          <div class="w-10 h-10 bg-purple-100 rounded-full flex items-center justify-center">🔄</div>
        </div>
      </div>
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('ai.autopilot.webhook') }}</p>
            <p class="text-2xl font-bold text-orange-600">{{ webhookCount }}</p>
          </div>
          <div class="w-10 h-10 bg-orange-100 rounded-full flex items-center justify-center">🔗</div>
        </div>
      </div>
    </div>

    <!-- Task List -->
    <div class="bg-white rounded-lg shadow-sm border border-gray-200">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.name') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.autopilot.taskType') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.autopilot.trigger') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.autopilot.schedule') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.status') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.autopilot.lastRun') }}</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="task in tasks" :key="task.id" class="hover:bg-gray-50">
              <td class="px-6 py-4">
                <div>
                  <p class="font-medium text-gray-900">{{ task.name }}</p>
                  <p class="text-sm text-gray-500">{{ task.description }}</p>
                </div>
              </td>
              <td class="px-6 py-4">
                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-indigo-100 text-indigo-800">
                  {{ getTaskTypeText(task.task_type) }}
                </span>
              </td>
              <td class="px-6 py-4">
                <span :class="getTriggerClass(task.trigger_type)" class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full">
                  {{ getTriggerText(task.trigger_type) }}
                </span>
              </td>
              <td class="px-6 py-4">
                <p v-if="task.trigger_type === 'cron'" class="text-sm text-gray-600 font-mono">{{ task.cron_expression }}</p>
                <p v-else-if="task.trigger_type === 'webhook'" class="text-sm text-gray-600 font-mono truncate max-w-xs" :title="task.trigger_url">{{ task.trigger_url }}</p>
                <p v-else class="text-sm text-gray-600">{{ t('ai.autopilot.manual') }}</p>
              </td>
              <td class="px-6 py-4">
                <span :class="task.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'" class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full">
                  {{ task.enabled ? t('common.enabled') : t('common.disabled') }}
                </span>
              </td>
              <td class="px-6 py-4">
                <p v-if="task.last_run_at" class="text-sm text-gray-500">{{ formatDate(task.last_run_at) }}</p>
                <p v-else class="text-sm text-gray-400">{{ t('ai.autopilot.never') }}</p>
              </td>
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end space-x-2">
                  <button @click="viewExecutions(task)" class="text-gray-600 hover:text-indigo-600" :title="t('ai.autopilot.executionHistory')">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"/></svg>
                  </button>
                  <button @click="executeTask(task)" class="text-gray-600 hover:text-blue-600" :title="t('ai.autopilot.execute')">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
                  </button>
                  <button @click="toggleTask(task)" class="text-gray-600 hover:text-green-600" :title="task.enabled ? t('ai.autopilot.disable') : t('ai.autopilot.enable')">
                    <svg v-if="task.enabled" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
                    <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/></svg>
                  </button>
                  <button @click="openEditModal(task)" class="text-gray-600 hover:text-blue-600">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
                  </button>
                  <button @click="deleteTaskConfirm(task)" class="text-gray-600 hover:text-red-600">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="tasks.length === 0">
              <td colspan="7" class="px-6 py-12 text-center">
                <div class="text-gray-400">
                  <div class="text-4xl mb-2">⏰</div>
                  <p>{{ t('ai.autopilot.empty') }}</p>
                  <button @click="openCreateModal" class="mt-4 text-blue-600 hover:text-blue-700 font-medium">{{ t('common.create') }}</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Transition name="modal">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/50" @click="closeModal"></div>
        <div class="relative bg-white rounded-lg shadow-xl w-full max-w-lg mx-4">
          <div class="flex items-center justify-between p-4 border-b">
            <h2 class="text-lg font-semibold">{{ isEditing ? t('ai.autopilot.edit') : t('ai.autopilot.create') }}</h2>
            <button @click="closeModal" class="text-gray-400 hover:text-gray-600">✕</button>
          </div>
          <div class="p-4 space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('common.name') }}</label>
              <input v-model="form.name" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500" :placeholder="t('ai.autopilot.namePlaceholder')" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('common.description') }}</label>
              <textarea v-model="form.description" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500" rows="2" :placeholder="t('common.descriptionPlaceholder')"></textarea>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.autopilot.taskType') }}</label>
              <select v-model="form.task_type" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500" :disabled="isEditing">
                <option value="report">{{ t('ai.autopilot.taskTypes.report') }}</option>
                <option value="sdlc">{{ t('ai.autopilot.taskTypes.sdlc') }}</option>
                <option value="test">{{ t('ai.autopilot.taskTypes.test') }}</option>
                <option value="deploy">{{ t('ai.autopilot.taskTypes.deploy') }}</option>
                <option value="cleanup">{{ t('ai.autopilot.taskTypes.cleanup') }}</option>
              </select>
              <p v-if="isEditing" class="text-xs text-gray-400 mt-1">{{ t('ai.autopilot.taskTypeLocked') }}</p>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.autopilot.trigger') }}</label>
              <select v-model="form.trigger_type" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500" :disabled="isEditing">
                <option value="cron">{{ t('ai.autopilot.cron') }}</option>
                <option value="webhook">{{ t('ai.autopilot.webhook') }}</option>
                <option value="manual">{{ t('ai.autopilot.manual') }}</option>
              </select>
            </div>
            <div v-if="form.trigger_type === 'cron'">
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.autopilot.cronExpression') }}</label>
              <input v-model="form.cron_expression" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 font-mono" placeholder="0 9 * * 1-5" />
              <p class="text-xs text-gray-400 mt-1">{{ t('ai.autopilot.cronHint') }}</p>
            </div>
            <div v-if="form.trigger_type === 'webhook' && isEditing">
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.autopilot.webhookUrl') }}</label>
              <input :value="webhookUrlPreview" type="text" readonly class="w-full px-3 py-2 border border-gray-200 rounded-lg bg-gray-50 font-mono text-xs text-gray-600" />
              <button @click="copyWebhookUrl" class="mt-1 text-xs text-blue-600 hover:text-blue-700">{{ t('common.copy') }}</button>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('common.status') }}</label>
              <label class="flex items-center gap-2 cursor-pointer">
                <input v-model="form.enabled" type="checkbox" class="rounded text-blue-600 focus:ring-blue-500" />
                <span class="text-sm text-gray-700">{{ t('ai.autopilot.enableTask') }}</span>
              </label>
            </div>
          </div>
          <div class="flex items-center justify-end p-4 border-t space-x-2">
            <button @click="closeModal" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">{{ t('common.cancel') }}</button>
            <button @click="saveTask" :disabled="!canSave" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg disabled:opacity-50">{{ t('common.save') }}</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Execution History Drawer -->
    <Transition name="drawer">
      <div v-if="showExecutionsPanel" class="fixed inset-0 z-50 flex justify-end">
        <div class="absolute inset-0 bg-black/50" @click="closeExecutionsPanel"></div>
        <div class="relative bg-white shadow-xl w-full max-w-2xl h-full overflow-y-auto">
          <div class="flex items-center justify-between p-4 border-b sticky top-0 bg-white z-10">
            <div>
              <h2 class="text-lg font-semibold">{{ t('ai.autopilot.executionHistory') }}</h2>
              <p class="text-sm text-gray-500">{{ currentTask?.name }}</p>
            </div>
            <button @click="closeExecutionsPanel" class="text-gray-400 hover:text-gray-600">✕</button>
          </div>
          <div class="p-4">
            <div v-if="executionsLoading" class="text-center py-8 text-gray-400">{{ t('common.loading') }}</div>
            <div v-else-if="executions.length === 0" class="text-center py-8 text-gray-400">
              <div class="text-4xl mb-2">📋</div>
              <p>{{ t('ai.autopilot.noExecutions') }}</p>
            </div>
            <ul v-else class="space-y-3">
              <li v-for="exec in executions" :key="exec.id" class="border border-gray-200 rounded-lg p-3 hover:bg-gray-50">
                <div class="flex items-center justify-between mb-2">
                  <div class="flex items-center gap-2">
                    <span :class="getExecutionStatusClass(exec.status)" class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full">
                      {{ getExecutionStatusText(exec.status) }}
                    </span>
                    <span class="text-xs text-gray-500">#{{ exec.id }}</span>
                    <span class="text-xs text-gray-400">{{ t('ai.autopilot.trigger') }}: {{ exec.trigger_type }}</span>
                  </div>
                  <span class="text-xs text-gray-400">{{ formatDate(exec.created_at) }}</span>
                </div>
                <div class="grid grid-cols-2 gap-2 text-xs text-gray-600 mb-2">
                  <div>
                    <span class="text-gray-400">{{ t('ai.autopilot.startedAt') }}:</span>
                    <span class="ml-1">{{ exec.started_at ? formatDate(exec.started_at) : '-' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-400">{{ t('ai.autopilot.completedAt') }}:</span>
                    <span class="ml-1">{{ exec.completed_at ? formatDate(exec.completed_at) : '-' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-400">{{ t('ai.autopilot.duration') }}:</span>
                    <span class="ml-1">{{ exec.duration_ms ? `${exec.duration_ms} ms` : '-' }}</span>
                  </div>
                  <div>
                    <span class="text-gray-400">{{ t('ai.autopilot.retryCount') }}:</span>
                    <span class="ml-1">{{ exec.retry_count }}</span>
                  </div>
                </div>
                <div v-if="exec.error_info" class="mt-2 text-xs bg-red-50 text-red-700 rounded p-2 font-mono whitespace-pre-wrap">
                  {{ exec.error_info }}
                </div>
                <div v-if="exec.output_data" class="mt-2">
                  <p class="text-xs text-gray-400 mb-1">{{ t('ai.autopilot.output') }}</p>
                  <pre class="text-xs bg-gray-50 rounded p-2 overflow-x-auto">{{ JSON.stringify(exec.output_data, null, 2) }}</pre>
                </div>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { useSSE } from '@/composables/useSSE'
import { autopilotApi, type AutopilotTask, type AutopilotExecution, type AutopilotCreateRequest, type AutopilotUpdateRequest } from '@/api/autopilot'

const { t } = useI18n()
const { getWorkspaceId } = useWorkspaceId()
const { onEvent } = useSSE()

const workspaceId = ref(0)

const tasks = ref<AutopilotTask[]>([])
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const editingTask = ref<AutopilotTask | null>(null)

// Executions panel state
const showExecutionsPanel = ref(false)
const currentTask = ref<AutopilotTask | null>(null)
const executions = ref<AutopilotExecution[]>([])
const executionsLoading = ref(false)

const form = ref({
  name: '',
  description: '',
  trigger_type: 'cron' as 'cron' | 'webhook' | 'manual',
  cron_expression: '',
  task_type: 'report',
  enabled: true,
})

const enabledCount = computed(() => tasks.value.filter(t => t.enabled).length)
const cronCount = computed(() => tasks.value.filter(t => t.trigger_type === 'cron').length)
const webhookCount = computed(() => tasks.value.filter(t => t.trigger_type === 'webhook').length)

// Save button is disabled when name is empty, or cron trigger lacks expression
const canSave = computed(() => {
  if (!form.value.name.trim()) return false
  if (form.value.trigger_type === 'cron' && !form.value.cron_expression.trim()) return false
  return true
})

const webhookUrlPreview = computed(() => {
  if (!editingTask.value?.trigger_url) return ''
  const origin = window.location.origin
  return `${origin}${editingTask.value.trigger_url}`
})

async function loadTasks() {
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    workspaceId.value = wsId
    const data = await autopilotApi.list(wsId)
    tasks.value = data || []
  } catch (err) {
    console.error('Failed to load autopilot tasks:', err)
    tasks.value = []
  }
}

function openCreateModal() {
  isEditing.value = false
  editingId.value = null
  editingTask.value = null
  form.value = { name: '', description: '', trigger_type: 'cron', cron_expression: '', task_type: 'report', enabled: true }
  showModal.value = true
}

function openEditModal(task: AutopilotTask) {
  isEditing.value = true
  editingId.value = task.id
  editingTask.value = task
  form.value = {
    name: task.name,
    description: task.description,
    trigger_type: task.trigger_type,
    cron_expression: task.cron_expression,
    task_type: task.task_type || 'report',
    enabled: task.enabled,
  }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

async function saveTask() {
  if (!form.value.name) return

  // Cron trigger requires a valid cron_expression (backend enforces this)
  if (form.value.trigger_type === 'cron' && !form.value.cron_expression.trim()) {
    alert(t('ai.autopilot.cronRequired'))
    return
  }

  const wsId = await getWorkspaceId()
  if (!wsId) return

  try {
    if (isEditing.value && editingId.value) {
      const data: AutopilotUpdateRequest = {
        name: form.value.name,
        description: form.value.description,
        cron_expression: form.value.cron_expression,
        enabled: form.value.enabled,
      }
      await autopilotApi.update(wsId, editingId.value, data)
    } else {
      const data: AutopilotCreateRequest = {
        name: form.value.name,
        description: form.value.description,
        trigger_type: form.value.trigger_type,
        cron_expression: form.value.cron_expression,
        task_type: form.value.task_type,
        enabled: form.value.enabled,
      }
      await autopilotApi.create(wsId, data)
    }
    closeModal()
    await loadTasks()
  } catch (err: any) {
    console.error('Failed to save autopilot task:', err)
    // Surface backend error message when available
    const backendMsg = err?.response?.data?.message
    alert(backendMsg ? `${t('common.saveFailed')}: ${backendMsg}` : t('common.saveFailed'))
  }
}

async function executeTask(task: AutopilotTask) {
  if (!workspaceId.value) return
  try {
    await autopilotApi.execute(workspaceId.value, task.id)
    // SSE will refresh; also reload tasks to update last_run_at
    await loadTasks()
  } catch (err) {
    console.error('Failed to execute task:', err)
    alert(t('ai.autopilot.executeFailed'))
  }
}

async function toggleTask(task: AutopilotTask) {
  const wsId = await getWorkspaceId()
  if (!wsId) return
  try {
    await autopilotApi.toggle(wsId, task.id)
    // SSE will refresh; optimistic local update as fallback
    task.enabled = !task.enabled
  } catch (err) {
    console.error('Failed to toggle task:', err)
    alert(t('common.operationFailed'))
  }
}

function deleteTaskConfirm(task: AutopilotTask) {
  if (!workspaceId.value) return
  if (confirm(t('ai.autopilot.deleteConfirm', { name: task.name }))) {
    autopilotApi.delete(workspaceId.value, task.id).then(() => {
      // SSE will refresh; local removal as fallback
      tasks.value = tasks.value.filter(t => t.id !== task.id)
    }).catch(err => {
      console.error('Failed to delete task:', err)
      alert(t('common.deleteFailed'))
    })
  }
}

async function viewExecutions(task: AutopilotTask) {
  currentTask.value = task
  showExecutionsPanel.value = true
  executions.value = []
  executionsLoading.value = true
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    const data = await autopilotApi.listExecutions(wsId, task.id)
    executions.value = data || []
  } catch (err) {
    console.error('Failed to load executions:', err)
    executions.value = []
  } finally {
    executionsLoading.value = false
  }
}

function closeExecutionsPanel() {
  showExecutionsPanel.value = false
  currentTask.value = null
  executions.value = []
}

function copyWebhookUrl() {
  if (!webhookUrlPreview.value) return
  navigator.clipboard.writeText(webhookUrlPreview.value).then(() => {
    alert(t('common.copied'))
  }).catch(() => {
    alert(webhookUrlPreview.value)
  })
}

// SSE handler - refreshes task list and executions panel on autopilot events
function handleSSEEvent(event: string, data: any) {
  if (!event.startsWith('autopilot_')) return

  if (event.startsWith('autopilot_task.')) {
    // Task lifecycle: refresh list to reflect changes
    loadTasks()
    return
  }

  if (event.startsWith('autopilot_execution.')) {
    const exec = data as AutopilotExecution
    // If executions panel is open for this task, refresh it
    if (showExecutionsPanel.value && currentTask.value && exec.task_id === currentTask.value.id) {
      viewExecutions(currentTask.value)
    }
    // Also refresh task list to update last_run_at / next_run_at
    loadTasks()
  }
}

function getTriggerClass(triggerType: string) {
  switch (triggerType) {
    case 'cron': return 'bg-purple-100 text-purple-800'
    case 'webhook': return 'bg-orange-100 text-orange-800'
    case 'manual': return 'bg-gray-100 text-gray-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

function getTriggerText(triggerType: string) {
  switch (triggerType) {
    case 'cron': return t('ai.autopilot.cron')
    case 'webhook': return t('ai.autopilot.webhook')
    case 'manual': return t('ai.autopilot.manual')
    default: return triggerType
  }
}

function getTaskTypeText(taskType: string) {
  const key = `ai.autopilot.taskTypes.${taskType}`
  const translated = t(key)
  // Fallback to raw value if no translation
  return translated === key ? taskType : translated
}

function getExecutionStatusClass(status: string) {
  switch (status) {
    case 'completed': return 'bg-green-100 text-green-800'
    case 'failed': return 'bg-red-100 text-red-800'
    case 'running': return 'bg-blue-100 text-blue-800'
    case 'pending': return 'bg-yellow-100 text-yellow-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

function getExecutionStatusText(status: string) {
  switch (status) {
    case 'completed': return t('ai.autopilot.statusCompleted')
    case 'failed': return t('ai.autopilot.statusFailed')
    case 'running': return t('ai.autopilot.statusRunning')
    case 'pending': return t('ai.autopilot.statusPending')
    default: return status
  }
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleString()
}

let cleanupSSE: (() => void) | null = null

onMounted(() => {
  loadTasks()
  cleanupSSE = onEvent(handleSSEEvent)
})

onUnmounted(() => {
  if (cleanupSSE) cleanupSSE()
})
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .relative,
.modal-leave-to .relative {
  transform: scale(0.95);
}

.drawer-enter-active,
.drawer-leave-active {
  transition: all 0.3s ease;
}

.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}
</style>
