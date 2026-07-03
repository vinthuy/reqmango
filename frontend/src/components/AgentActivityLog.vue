<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { agentApi } from '@/api/agent'
import type { AgentActivity } from '@/types/agent'
import { useI18n } from '@/composables/useI18n'

const props = defineProps<{
  agentId: number
  workspaceId: number
  issueId?: number
}>()

const { t } = useI18n()

const activities = ref<AgentActivity[]>([])
const loading = ref(false)

const actionLabels = computed(() => ({
  dispatch: t('agent.dispatch'),
  auto_triage: t('agent.autoTriage'),
  auto_assign: t('agent.autoAssign'),
  mention: t('agent.mention'),
  summarize: t('agent.summarize'),
  custom: t('agent.customTask'),
}))

const actionColors: Record<string, string> = {
  dispatch: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  auto_triage: 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300',
  auto_assign: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  mention: 'bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300',
  summarize: 'bg-teal-100 text-teal-700 dark:bg-teal-900 dark:text-teal-300',
  custom: 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
}

async function fetchActivities() {
  loading.value = true
  try {
    activities.value = await agentApi.getActivity(props.workspaceId, props.agentId)
  } catch (e) {
    console.error('Failed to fetch agent activities', e)
  } finally {
    loading.value = false
  }
}

watch(() => props.agentId, fetchActivities, { immediate: true })
</script>

<template>
  <div class="agent-activity-log">
    <h4 class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-3 uppercase tracking-wider">{{ t('agent.activity') }}</h4>

    <div v-if="loading" class="flex justify-center py-4">
      <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-indigo-600"></div>
    </div>

    <div v-else-if="activities.length === 0" class="text-center py-4 text-gray-400 text-sm">
      {{ t('agent.noActivity') }}
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="act in activities"
        :key="act.id"
        class="flex gap-3 text-sm"
      >
        <!-- Timeline dot -->
        <div class="flex flex-col items-center">
          <div class="w-2 h-2 rounded-full bg-indigo-400 mt-1.5"></div>
          <div class="w-px flex-1 bg-gray-200 dark:bg-gray-700"></div>
        </div>

        <!-- Content -->
        <div class="flex-1 pb-3">
          <div class="flex items-center gap-2 mb-0.5">
            <span
              :class="actionColors[act.action] || 'bg-gray-100 text-gray-600'"
              class="text-xs px-1.5 py-0.5 rounded-full font-medium"
            >
              {{ (actionLabels as Record<string, string>)[act.action] || act.action }}
            </span>
            <span class="text-xs text-gray-400">
              {{ new Date(act.executed_at).toLocaleString() }}
            </span>
          </div>
          <p class="text-gray-600 dark:text-gray-400 whitespace-pre-wrap line-clamp-3">
            {{ act.result_summary }}
          </p>
          <p
            v-if="act.task_context"
            class="text-xs text-gray-400 mt-1 line-clamp-2"
          >
            Task: {{ act.task_context }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
