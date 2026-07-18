<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <button @click="$router.back()" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">← Back</button>
      <h1 class="text-xl font-bold text-gray-900 dark:text-white">Loop Run #{{ runId }}</h1>
      <LoopStateBadge v-if="detail?.run.status" :status="detail.run.status" />
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-400">Loading...</div>

    <template v-else-if="detail">
      <div class="border rounded-lg p-4 mb-6 dark:border-gray-700">
        <div class="font-medium mb-2 text-gray-900 dark:text-white">{{ detail.run.goal }}</div>
        <BudgetGauge
          :max-tokens="50000"
          :used-tokens="detail.run.tokens_used"
          :max-iterations="detail.run.max_iterations"
          :current-iteration="detail.run.current_iteration"
        />
        <div class="text-xs text-gray-400 mt-2">
          Started {{ new Date(detail.run.started_at).toLocaleString() }}
          <span v-if="detail.run.stopped_reason"> · {{ detail.run.stopped_reason }}</span>
        </div>
      </div>

      <h2 class="font-semibold mb-3 text-gray-900 dark:text-white">Iterations</h2>
      <div class="space-y-3">
        <div
          v-for="iter in detail.iterations"
          :key="iter.id"
          class="border rounded-lg p-4 dark:border-gray-700"
        >
          <div class="flex items-center justify-between mb-2">
            <span class="text-sm font-medium text-gray-600 dark:text-gray-300">Iteration #{{ iter.iteration_num }}</span>
            <span
              :class="{
                'text-green-600': iter.decision === 'stop',
                'text-blue-600': iter.decision === 'continue',
                'text-yellow-600': iter.decision === 'escalate',
              }"
              class="text-xs font-medium uppercase"
            >
              {{ iter.decision }}
            </span>
          </div>
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-2">{{ iter.reasoning }}</p>
          <div class="flex gap-4 text-xs text-gray-400">
            <span>Tokens: {{ iter.tokens_used }}</span>
            <span v-if="iter.duration_ms">Duration: {{ (iter.duration_ms / 1000).toFixed(1) }}s</span>
          </div>
        </div>
        <div v-if="detail.iterations.length === 0" class="text-center py-8 text-gray-400 text-sm">
          No iterations recorded yet
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAgentLoopStore } from '@/stores/agentLoop'
import LoopStateBadge from '@/components/agents/LoopStateBadge.vue'
import BudgetGauge from '@/components/agents/BudgetGauge.vue'

const route = useRoute()
const runId = Number(route.params.runId)
const store = useAgentLoopStore()
const detail = ref(store.runDetail)
const loading = ref(true)
let stopWatch: (() => void) | null = null

function getWorkspaceId(): number | null {
  const match = window.location.pathname.match(/\/workspaces\/(\d+)/)
  if (match) return Number(match[1])
  const stored = localStorage.getItem('currentWorkspaceId')
  return stored ? Number(stored) : null
}

onMounted(async () => {
  const wsId = getWorkspaceId()
  if (!wsId) { loading.value = false; return }

  await store.fetchRunDetail(wsId, runId)
  detail.value = store.runDetail
  loading.value = false

  if (detail.value?.run.status === 'running') {
    stopWatch = store.watchRun(wsId, runId, 3000)
  }
})

onUnmounted(() => {
  stopWatch?.()
})
</script>
