<template>
  <div class="p-6 max-w-5xl mx-auto">
    <h1 class="text-2xl font-bold mb-6 text-gray-900 dark:text-white">📋 Agent Sessions</h1>

    <div v-if="loading" class="text-center py-12 text-gray-400">Loading...</div>

    <div v-else-if="sessions.length === 0" class="text-center py-12 text-gray-400 border rounded-lg dark:border-gray-700">
      No agent sessions recorded yet
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="session in sessions"
        :key="session.id"
        class="border rounded-lg p-3 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors dark:border-gray-700"
      >
        <div class="flex items-center justify-between">
          <div>
            <span class="text-xs font-mono text-gray-400 mr-2">{{ session.id.slice(0, 8) }}</span>
            <span class="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300">{{ session.agent_type }}</span>
          </div>
          <LoopStateBadge :status="session.status" />
        </div>
        <div v-if="session.input_summary" class="text-sm text-gray-600 dark:text-gray-400 mt-1">{{ session.input_summary }}</div>
        <div class="flex gap-4 text-xs text-gray-400 mt-1">
          <span>T: {{ (session.tokens_input + session.tokens_output).toLocaleString() }}</span>
          <span>${{ session.cost_usd.toFixed(4) }}</span>
          <span>{{ new Date(session.started_at).toLocaleString() }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { sessionApi, type AgentSession } from '@/api/agent-session'
import LoopStateBadge from '@/components/agents/LoopStateBadge.vue'

const route = useRoute()
const { getWorkspaceId } = useWorkspaceId()

const sessions = ref<AgentSession[]>([])
const loading = ref(true)

onMounted(async () => {
  const wsId = await getWorkspaceId()
  if (!wsId) { loading.value = false; return }
  try {
    sessions.value = await sessionApi.list(wsId, { limit: 50 })
  } finally {
    loading.value = false
  }
})
</script>
