<template>
  <div>
    <div
      v-if="visible"
      class="fixed inset-0 bg-black/50 z-50"
      @click.self="$emit('close')"
    ></div>
    <div
      v-if="visible"
      class="fixed top-0 right-0 h-full w-full max-w-md bg-white shadow-xl z-50 flex flex-col transition-transform duration-300"
      :class="visible ? 'translate-x-0' : 'translate-x-full'"
    >
      <div class="flex items-center justify-between p-4 border-b border-gray-100">
        <div>
          <h2 class="text-lg font-semibold text-gray-900">{{ t('automationExecutionLog.title') }}</h2>
          <p class="text-sm text-gray-500 mt-0.5">{{ t('automationExecutionLog.total') }}: {{ total }}</p>
        </div>
        <button
          @click="$emit('close')"
          class="text-gray-500 hover:text-gray-700 transition-colors"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="p-4 border-b border-gray-100 bg-gray-50">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex items-center gap-2">
            <label class="text-xs font-medium text-gray-600">{{ t('automationExecutionLog.startDate') }}</label>
            <input
              v-model="filters.startDate"
              type="date"
              @change="loadLogs"
              class="px-2 py-1.5 border border-gray-200 rounded text-xs focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
          <div class="flex items-center gap-2">
            <label class="text-xs font-medium text-gray-600">{{ t('automationExecutionLog.endDate') }}</label>
            <input
              v-model="filters.endDate"
              type="date"
              @change="loadLogs"
              class="px-2 py-1.5 border border-gray-200 rounded text-xs focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            />
          </div>
          <button
            @click="resetFilters"
            class="px-2 py-1 text-xs text-gray-600 hover:text-gray-800 transition-colors"
          >
            {{ t('automationExecutionLog.reset') }}
          </button>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto p-4 space-y-2">
        <div v-if="loading" class="text-center py-12">
          <span class="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-indigo-600"></span>
        </div>

        <div v-else-if="logs.length === 0" class="text-center py-12">
          <div class="w-12 h-12 mx-auto mb-3 rounded-full bg-gray-100 flex items-center justify-center">
            <span class="text-2xl">📝</span>
          </div>
          <p class="text-sm text-gray-500">{{ t('automationExecutionLog.emptyDescription') }}</p>
        </div>

        <div
          v-for="log in logs"
          :key="log.id"
          class="bg-white rounded-lg border border-gray-200 overflow-hidden cursor-pointer hover:border-indigo-300 hover:shadow-md transition-all"
          @click="toggleDetail(log.id)"
        >
          <div class="p-3" :class="getStatusBg(log.status)">
            <div class="flex items-start justify-between">
              <div class="flex items-center space-x-2">
                <span :class="['w-7 h-7 rounded flex items-center justify-center text-sm', getStatusIconBg(log.status)]">
                  {{ getStatusIcon(log.status) }}
                </span>
                <div class="min-w-0">
                  <div class="flex items-center space-x-2">
                    <span :class="['font-medium text-sm', getStatusTextColor(log.status)]">
                      {{ getStatusLabel(log.status) }}
                    </span>
                    <span class="text-xs text-gray-400">{{ formatTime(log.executed_at) }}</span>
                  </div>
                  <div class="flex items-center space-x-1 mt-1 flex-wrap">
                    <span class="text-xs bg-purple-100 text-purple-700 px-1.5 py-0.5 rounded">
                      {{ getTriggerLabel(log.trigger_type) }}
                    </span>
                    <span v-if="log.issue_id" class="text-xs text-gray-500">
                      #{{ log.issue_id }}
                    </span>
                    <span v-if="log.duration > 0" class="text-xs text-gray-400">
                      {{ log.duration }}ms
                    </span>
                  </div>
                </div>
              </div>
              <svg
                class="w-4 h-4 text-gray-400 flex-shrink-0 transition-transform"
                :class="{ 'rotate-180': expandedId === log.id }"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </div>
          </div>

          <div v-if="expandedId === log.id" class="p-3 border-t border-gray-100 bg-gray-50">
            <div v-if="log.context_json" class="mb-3">
              <h4 class="text-xs font-semibold text-gray-500 uppercase mb-1.5">{{ t('automationExecutionLog.context') }}</h4>
              <div class="bg-white rounded p-2 text-xs font-mono overflow-auto max-h-32 border border-gray-200">
                <pre>{{ formatContext(log.context_json) }}</pre>
              </div>
            </div>

            <div v-if="log.actions_taken && log.actions_taken !== '[]'" class="mb-3">
              <h4 class="text-xs font-semibold text-gray-500 uppercase mb-1.5">{{ t('automationExecutionLog.actionsTaken') }}</h4>
              <div class="flex flex-wrap gap-1.5">
                <span
                  v-for="(action, index) in parseActions(log.actions_taken)"
                  :key="index"
                  class="text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded"
                >
                  {{ action }}
                </span>
              </div>
            </div>

            <div v-if="log.error" :class="getErrorClass(log)">
              <h4 class="text-xs font-semibold uppercase mb-1" :class="log.status === 'skipped' && log.error === 'Conditions not met' ? 'text-yellow-600' : 'text-red-600'">
                {{ log.status === 'skipped' && log.error === 'Conditions not met' ? t('automationExecutionLog.conditionNotMet') : t('automationExecutionLog.error') }}
              </h4>
              <p class="text-xs" :class="log.status === 'skipped' && log.error === 'Conditions not met' ? 'text-yellow-700' : 'text-red-700'">{{ log.error }}</p>
            </div>

            <div v-if="log.rule_id" class="mt-2 pt-2 border-t border-gray-200">
              <span class="text-xs text-gray-500">{{ t('automationExecutionLog.ruleId') }}: {{ log.rule_id }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="total > limit" class="p-3 border-t border-gray-100 bg-white">
        <div class="flex items-center justify-center space-x-2">
          <button
            @click="prevPage"
            :disabled="currentPage === 1"
            class="px-3 py-1 border border-gray-200 rounded text-xs hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ t('automationExecutionLog.prev') }}
          </button>
          <span class="text-xs text-gray-600">{{ t('automationExecutionLog.page') }} {{ currentPage }} / {{ totalPages }}</span>
          <button
            @click="nextPage"
            :disabled="currentPage >= totalPages"
            class="px-3 py-1 border border-gray-200 rounded text-xs hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ t('automationExecutionLog.next') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { automationApi, type AutomationExecution } from '@/api/automation'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  ruleId?: number
  projectId?: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const loading = ref(false)
const logs = ref<AutomationExecution[]>([])
const total = ref(0)
const limit = 20
const currentPage = ref(1)
const expandedId = ref<number | null>(null)

const filters = ref({
  startDate: '',
  endDate: '',
})

const totalPages = computed(() => Math.ceil(total.value / limit))

function getOffset() {
  return (currentPage.value - 1) * limit
}

function getTimeParams() {
  const params: { startTime?: string; endTime?: string } = {}
  if (filters.value.startDate) {
    params.startTime = new Date(filters.value.startDate).toISOString()
  }
  if (filters.value.endDate) {
    const endDate = new Date(filters.value.endDate)
    endDate.setDate(endDate.getDate() + 1)
    params.endTime = endDate.toISOString()
  }
  return params
}

async function loadLogs() {
  loading.value = true
  currentPage.value = 1
  expandedId.value = null
  try {
    const timeParams = getTimeParams()
    if (props.ruleId) {
      const res = await automationApi.getRuleExecutionHistory(props.ruleId, {
        limit,
        offset: 0,
        ...timeParams,
      })
      logs.value = res.data
      total.value = res.total
    } else if (props.projectId) {
      const res = await automationApi.getProjectExecutionHistory(props.projectId, {
        limit,
        offset: 0,
        ...timeParams,
      })
      logs.value = res.data
      total.value = res.total
    }
  } catch (e) {
    console.error('Failed to load automation execution logs:', e)
  } finally {
    loading.value = false
  }
}

async function loadPage(page: number) {
  loading.value = true
  currentPage.value = page
  expandedId.value = null
  try {
    const timeParams = getTimeParams()
    if (props.ruleId) {
      const res = await automationApi.getRuleExecutionHistory(props.ruleId, {
        limit,
        offset: getOffset(),
        ...timeParams,
      })
      logs.value = res.data
      total.value = res.total
    } else if (props.projectId) {
      const res = await automationApi.getProjectExecutionHistory(props.projectId, {
        limit,
        offset: getOffset(),
        ...timeParams,
      })
      logs.value = res.data
      total.value = res.total
    }
  } catch (e) {
    console.error('Failed to load automation execution logs:', e)
  } finally {
    loading.value = false
  }
}

function prevPage() {
  if (currentPage.value > 1) {
    loadPage(currentPage.value - 1)
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    loadPage(currentPage.value + 1)
  }
}

function resetFilters() {
  filters.value = { startDate: '', endDate: '' }
  loadLogs()
}

function toggleDetail(id: number) {
  expandedId.value = expandedId.value === id ? null : id
}

function formatTime(timeStr: string) {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

function getStatusIcon(status: string) {
  switch (status) {
    case 'success': return '✅'
    case 'failed': return '❌'
    case 'skipped': return '⏭️'
    default: return 'ℹ️'
  }
}

function getStatusIconBg(status: string) {
  switch (status) {
    case 'success': return 'bg-green-100'
    case 'failed': return 'bg-red-100'
    case 'skipped': return 'bg-yellow-100'
    default: return 'bg-gray-100'
  }
}

function getStatusBg(status: string) {
  switch (status) {
    case 'success': return 'bg-green-50/50'
    case 'failed': return 'bg-red-50/50'
    case 'skipped': return 'bg-yellow-50/50'
    default: return 'bg-gray-50/50'
  }
}

function getStatusTextColor(status: string) {
  switch (status) {
    case 'success': return 'text-green-700'
    case 'failed': return 'text-red-700'
    case 'skipped': return 'text-yellow-700'
    default: return 'text-gray-700'
  }
}

function getStatusLabel(status: string) {
  switch (status) {
    case 'success': return t('automationExecutionLog.success')
    case 'failed': return t('automationExecutionLog.failed')
    case 'skipped': return t('automationExecutionLog.skipped')
    default: return status
  }
}

function getErrorClass(log: AutomationExecution) {
  if (log.status === 'skipped' && log.error === 'Conditions not met') {
    return 'p-2 bg-yellow-50 rounded-lg border border-yellow-100'
  }
  return 'p-2 bg-red-50 rounded-lg border border-red-100'
}

function getTriggerLabel(triggerType: string) {
  const labels: Record<string, string> = {
    'issue.created': t('automationHistory.triggerCreated'),
    'issue.updated': t('automationHistory.triggerUpdated'),
    'issue.state_changed': t('automationHistory.triggerStateChanged'),
    'issue.assigned': t('automationHistory.triggerAssigned'),
    'comment.added': t('automationHistory.triggerCommentAdded'),
  }
  return labels[triggerType] || triggerType
}

function formatContext(contextJson: string) {
  try {
    const context = JSON.parse(contextJson)
    return JSON.stringify(context, null, 2)
  } catch {
    return contextJson
  }
}

function parseActions(actionsJson: string) {
  try {
    const actions = JSON.parse(actionsJson)
    return actions.map((a: any) => {
      const actionLabels: Record<string, string> = {
        'change_state': t('automationHistory.actionChangeState'),
        'set_priority': t('automationHistory.actionSetPriority'),
        'assign_to': t('automationHistory.actionAssign'),
        'add_comment': t('automationHistory.actionAddComment'),
        'set_field': t('automationHistory.actionSetField'),
        'call_webhook': t('automationHistory.actionCallWebhook'),
      }
      return actionLabels[a.type] || a.type
    })
  } catch {
    return []
  }
}

watch(() => [props.ruleId, props.projectId], () => {
  loadLogs()
})

watch(() => props.visible, (newVal) => {
  if (newVal) {
    loadLogs()
  }
})
</script>
