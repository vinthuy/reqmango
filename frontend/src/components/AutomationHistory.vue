<template>
  <div class="automation-history">
    <h4 class="text-sm font-semibold text-gray-700 mb-3">{{ t('automationHistory.title') }}</h4>
    
    <div v-if="loading" class="text-center py-4">
      <span class="inline-block animate-spin rounded-full h-4 w-4 border-b-2 border-indigo-600"></span>
    </div>
    
    <div v-else-if="history.length === 0" class="text-center py-4 text-gray-400 text-sm">
      {{ t('automationHistory.noHistory') }}
    </div>
    
    <div v-else class="space-y-2">
      <div
        v-for="item in history"
        :key="item.id"
        class="p-3 bg-gray-50 rounded-lg border border-gray-100"
      >
        <div class="flex items-center justify-between mb-2">
          <span class="text-xs font-medium" :class="item.status === 'success' ? 'text-green-600' : 'text-red-600'">
            {{ item.status === 'success' ? t('automationHistory.success') : t('automationHistory.failed') }}
          </span>
          <span class="text-xs text-gray-400">{{ formatTime(item.executed_at) }}</span>
        </div>
        <div class="text-xs text-gray-600 mb-1">
          <span class="font-medium">{{ t('automationHistory.trigger') }}:</span> {{ getTriggerLabel(item.trigger_type) }}
        </div>
        <div v-if="item.context_json" class="text-xs text-gray-500 mb-1">
          <span class="font-medium">{{ t('automationHistory.context') }}:</span>
          <pre class="mt-1 p-2 bg-white rounded text-xs overflow-auto max-h-20">{{ item.context_json }}</pre>
        </div>
        <div v-if="item.actions_taken && item.actions_taken !== '[]'" class="text-xs text-gray-500">
          <span class="font-medium">{{ t('automationHistory.actions') }}:</span> {{ parseActions(item.actions_taken) }}
        </div>
        <div v-if="item.error" class="text-xs text-red-500 mt-1">
          <span class="font-medium">{{ t('automationHistory.error') }}:</span> {{ item.error }}
        </div>
        <div class="text-xs text-gray-400 mt-1">
          {{ t('automationHistory.duration') }}: {{ item.duration }}ms
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { automationApi, type AutomationExecution } from '@/api/automation'

const { t } = useI18n()

const props = defineProps<{
  issueId: number
}>()

const loading = ref(false)
const history = ref<AutomationExecution[]>([])

async function loadHistory() {
  loading.value = true
  try {
    history.value = await automationApi.getExecutionHistory(props.issueId, 20)
  } catch (e) {
    console.error('Failed to load automation history:', e)
  } finally {
    loading.value = false
  }
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

function parseActions(actionsJson: string) {
  try {
    const actions = JSON.parse(actionsJson)
    return actions.map((a: any) => {
      const actionLabels: Record<string, string> = {
        'change_state': t('automationHistory.actionChangeState'),
        'set_priority': t('automationHistory.actionSetPriority'),
        'assign_to': t('automationHistory.actionAssign'),
        'add_comment': t('automationHistory.actionAddComment'),
        'add_label': t('automationHistory.actionAddLabel'),
      }
      return actionLabels[a.type] || a.type
    }).join(', ')
  } catch {
    return actionsJson
  }
}

onMounted(loadHistory)
</script>

<style scoped>
.automation-history {
  @apply py-2;
}
</style>
