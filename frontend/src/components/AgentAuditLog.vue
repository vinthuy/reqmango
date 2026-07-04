<template>
  <div class="agent-audit-log">
    <!-- Header with filters -->
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-sm font-semibold text-gray-700">{{ t('agent.auditTitle') || 'AI Agent Audit Log' }}</h3>
      <button @click="refresh" class="text-xs text-blue-600 hover:text-blue-800">
        {{ t('common.refresh') || 'Refresh' }}
      </button>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-2 mb-4">
      <select
        v-model="filterAgentId"
        @change="fetchActivities"
        class="px-2 py-1 text-xs border border-gray-200 rounded-md focus:outline-none focus:ring-1 focus:ring-violet-400"
      >
        <option value="">{{ t('agent.auditAllAgents') || 'All Agents' }}</option>
        <option v-for="a in agents" :key="a.id" :value="a.id">{{ a.avatar || '🤖' }} {{ a.name }}</option>
      </select>

      <select
        v-model="filterAction"
        @change="fetchActivities"
        class="px-2 py-1 text-xs border border-gray-200 rounded-md focus:outline-none focus:ring-1 focus:ring-violet-400"
      >
        <option value="">{{ t('agent.auditAllActions') || 'All Actions' }}</option>
        <option value="dispatch">{{ t('agent.dispatch') || 'Dispatch' }}</option>
        <option value="auto_triage">{{ t('agent.autoTriage') || 'Auto Triage' }}</option>
        <option value="auto_assign">{{ t('agent.autoAssign') || 'Auto Assign' }}</option>
        <option value="mention">{{ t('agent.mention') || 'Mention' }}</option>
        <option value="summarize">{{ t('agent.summarize') || 'Summarize' }}</option>
        <option value="custom">{{ t('agent.customTask') || 'Custom Task' }}</option>
      </select>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="flex justify-center py-8">
      <div class="animate-spin w-5 h-5 border-2 border-violet-500 border-t-transparent rounded-full"></div>
    </div>

    <!-- Empty state -->
    <div v-else-if="activities.length === 0" class="text-center py-8 text-sm text-gray-400">
      {{ t('agent.auditNoRecords') || 'No audit records found' }}
    </div>

    <!-- Timeline -->
    <div v-else class="relative">
      <!-- Timeline line -->
      <div class="absolute left-3 top-2 bottom-2 w-px bg-gray-200"></div>

      <div v-for="act in activities" :key="act.id" class="relative pl-8 pb-4">
        <!-- Dot -->
        <div
          :class="[
            'absolute left-1.5 top-1.5 w-3.5 h-3.5 rounded-full border-2 border-white',
            actionColor(act.action)
          ]"
        ></div>

        <!-- Card -->
        <div class="p-3 bg-white border border-gray-100 rounded-lg shadow-sm hover:shadow transition-shadow">
          <div class="flex items-center justify-between mb-1">
            <div class="flex items-center gap-1.5">
              <span class="text-sm">{{ act.agent_name || 'Agent #' + act.agent_id }}</span>
              <span
                :class="['px-1.5 py-0.5 rounded text-xs', actionBadge(act.action)]"
              >{{ formatAction(act.action) }}</span>
            </div>
            <span class="text-xs text-gray-400">{{ formatTime(act.executed_at) }}</span>
          </div>
          <p class="text-xs text-gray-600 line-clamp-2">{{ act.result_summary }}</p>
          <div class="flex items-center justify-between mt-1.5">
            <span v-if="act.issue_id" class="text-xs text-gray-400">
              {{ t('agent.auditIssueId') || 'Issue' }}: #{{ act.issue_id }}
            </span>
            <span v-else></span>
            <div class="flex items-center gap-1">
              <button
                @click.stop="submitFeedback(act.id, 1)"
                :disabled="feedbackSubmitting === act.id"
                :class="['p-1 rounded transition', act.rating === 1 ? 'text-emerald-500 bg-emerald-50' : 'text-gray-300 hover:text-emerald-500 hover:bg-emerald-50']"
                :title="t('agent.feedbackUp') || 'Helpful'"
              ><svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5"/></svg></button>
              <button
                @click.stop="submitFeedback(act.id, -1)"
                :disabled="feedbackSubmitting === act.id"
                :class="['p-1 rounded transition', act.rating === -1 ? 'text-red-500 bg-red-50' : 'text-gray-300 hover:text-red-500 hover:bg-red-50']"
                :title="t('agent.feedbackDown') || 'Not helpful'"
              ><svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14H5.236a2 2 0 01-1.789-2.894l3.5-7A2 2 0 018.736 3h4.018a2 2 0 01.485.06L17 4m-7 10v5a2 2 0 002 2h.095c.5 0 .905-.405.905-.905 0-.714.211-1.412.608-2.006L17 13V4m-7 10h2m5-10h2a2 2 0 012 2v6a2 2 0 01-2 2h-2.5"/></svg></button>
            </div>
          </div>
          <div v-if="act.task_context" class="mt-1.5 pt-1.5 border-t border-gray-50">
            <p class="text-xs text-gray-400 line-clamp-3 italic">"{{ act.task_context }}"</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { agentApi } from '@/api/agent'
import type { Agent, AgentActivity } from '@/types/agent'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  workspaceId: number
  issueId?: number
}>(), {})

const agents = ref<Agent[]>([])
const activities = ref<AgentActivity[]>([])
const loading = ref(false)
const filterAgentId = ref<string>('')
const filterAction = ref<string>('')
const feedbackSubmitting = ref<number | null>(null)

function actionColor(action: string): string {
  const map: Record<string, string> = {
    dispatch: 'bg-blue-400',
    auto_triage: 'bg-emerald-400',
    auto_assign: 'bg-violet-400',
    mention: 'bg-amber-400',
    summarize: 'bg-cyan-400',
    custom: 'bg-gray-400',
  }
  return map[action] || 'bg-gray-400'
}

function actionBadge(action: string): string {
  const map: Record<string, string> = {
    dispatch: 'bg-blue-50 text-blue-600',
    auto_triage: 'bg-emerald-50 text-emerald-600',
    auto_assign: 'bg-violet-50 text-violet-600',
    mention: 'bg-amber-50 text-amber-600',
    summarize: 'bg-cyan-50 text-cyan-600',
    custom: 'bg-gray-100 text-gray-600',
  }
  return map[action] || 'bg-gray-100 text-gray-600'
}

function formatAction(action: string): string {
  const labels: Record<string, string> = {
    dispatch: t('agent.dispatch') || 'Dispatch',
    auto_triage: t('agent.autoTriage') || 'Auto Triage',
    auto_assign: t('agent.autoAssign') || 'Auto Assign',
    mention: t('agent.mention') || 'Mention',
    summarize: t('agent.summarize') || 'Summarize',
    custom: t('agent.customTask') || 'Custom Task',
  }
  return labels[action] || action
}

function formatTime(ts: string): string {
  const d = new Date(ts)
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return t('common.justNow') || 'just now'
  if (diffMin < 60) return `${diffMin}m ago`
  const diffHr = Math.floor(diffMin / 60)
  if (diffHr < 24) return `${diffHr}h ago`
  return d.toLocaleDateString()
}

async function fetchAgents() {
  try {
    agents.value = await agentApi.list(props.workspaceId)
  } catch (e) {
    console.error('Failed to fetch agents', e)
  }
}

async function fetchActivities() {
  loading.value = true
  try {
    const params: { agent_id?: number; action?: string; limit: number } = {
      limit: 50,
    }
    if (filterAgentId.value) params.agent_id = parseInt(filterAgentId.value)
    if (filterAction.value) params.action = filterAction.value
    activities.value = await agentApi.listWorkspaceActivity(props.workspaceId, params)
  } catch (e) {
    console.error('Failed to fetch activities', e)
  } finally {
    loading.value = false
  }
}

async function submitFeedback(activityId: number, rating: 1 | -1) {
  feedbackSubmitting.value = activityId
  try {
    await agentApi.rateActivity(props.workspaceId, activityId, rating)
    const act = activities.value.find(a => a.id === activityId)
    if (act) act.rating = rating
  } catch (e) {
    console.error('Failed to submit feedback:', e)
  } finally {
    feedbackSubmitting.value = null
  }
}

function refresh() {
  fetchActivities()
}

onMounted(() => {
  fetchAgents()
  fetchActivities()
})
</script>
