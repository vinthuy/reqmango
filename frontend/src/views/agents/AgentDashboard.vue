<template>
  <div class="p-6 max-w-5xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">🤖 Agent Dashboard</h1>
        <p class="text-sm text-gray-500 mt-1">Monitor and manage autonomous agent loops</p>
      </div>
    </div>

    <section class="mb-8">
      <h2 class="text-lg font-semibold mb-3">Active Loops</h2>
      <div v-if="activeRuns.length === 0" class="text-sm text-gray-400 py-8 text-center border rounded-lg dark:border-gray-700">
        No active loop runs. Start one from the Loops tab.
      </div>
      <div v-else class="space-y-3">
        <div
          v-for="run in activeRuns"
          :key="run.id"
          class="border rounded-lg p-4 hover:border-indigo-300 cursor-pointer transition-colors dark:border-gray-700"
          @click="$router.push(`/agents/loops/runs/${run.id}`)"
        >
          <div class="flex items-center justify-between mb-2">
            <span class="font-medium text-gray-900 dark:text-white">{{ run.goal }}</span>
            <LoopStateBadge :status="run.status" />
          </div>
          <BudgetGauge
            :max-tokens="50000"
            :used-tokens="run.tokens_used"
            :max-iterations="run.max_iterations"
            :current-iteration="run.current_iteration"
          />
        </div>
      </div>
    </section>

    <section class="grid grid-cols-2 gap-4">
      <router-link
        to="/agents/loops"
        class="border rounded-lg p-4 hover:border-indigo-400 hover:shadow-sm transition-all dark:border-gray-700"
      >
        <div class="text-xl mb-1">🔄</div>
        <div class="font-medium text-sm text-gray-900 dark:text-white">Loop Configurations</div>
        <div class="text-xs text-gray-400 mt-1">Create and manage autonomous loops</div>
      </router-link>
      <router-link
        to="/agents/sessions"
        class="border rounded-lg p-4 hover:border-indigo-400 hover:shadow-sm transition-all dark:border-gray-700"
      >
        <div class="text-xl mb-1">📋</div>
        <div class="font-medium text-sm text-gray-900 dark:text-white">Agent Sessions</div>
        <div class="text-xs text-gray-400 mt-1">View agent execution history and costs</div>
      </router-link>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { loopApi, type LoopRun } from '@/api/agent-loop'
import LoopStateBadge from '@/components/agents/LoopStateBadge.vue'
import BudgetGauge from '@/components/agents/BudgetGauge.vue'

const activeRuns = ref<LoopRun[]>([])

onMounted(async () => {
  try {
    // Get workspace ID from localStorage or router
    const wsId = getWorkspaceId()
    if (!wsId) return
    const loops = await loopApi.list(wsId)
    for (const loop of loops) {
      if (loop.status === 'active') {
        const runs = await loopApi.getRuns(wsId, loop.id, 5)
        activeRuns.value.push(...runs.filter(r => r.status === 'running'))
      }
    }
  } catch { /* no loops yet */ }
})

function getWorkspaceId(): number | null {
  const match = window.location.pathname.match(/\/workspaces\/(\d+)/)
  if (match) return Number(match[1])
  // fallback: check localStorage
  const stored = localStorage.getItem('currentWorkspaceId')
  return stored ? Number(stored) : null
}
</script>
