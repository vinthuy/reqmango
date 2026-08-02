<template>
  <div class="p-6 max-w-5xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">🔄 Agent Loops</h1>
      <button
        @click="showCreate = true"
        class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700"
      >
        + New Loop
      </button>
    </div>

    <div v-if="store.loading" class="text-center py-12 text-gray-400">Loading...</div>

    <div v-else-if="store.loops.length === 0" class="text-center py-12 text-gray-400 border rounded-lg dark:border-gray-700">
      No loops configured yet. Create one to get started.
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="loop in store.loops"
        :key="loop.id"
        class="border rounded-lg p-4 hover:border-indigo-300 transition-colors dark:border-gray-700"
      >
        <div class="flex items-center justify-between">
          <div>
            <span class="font-medium text-gray-900 dark:text-white">{{ loop.name }}</span>
            <span v-if="loop.description" class="text-sm text-gray-400 ml-2">{{ loop.description }}</span>
          </div>
          <div class="flex gap-2">
            <button
              @click="startLoop(loop.id)"
              class="px-3 py-1 bg-green-600 text-white rounded text-xs hover:bg-green-700"
            >
              ▶ Run
            </button>
          </div>
        </div>
        <div class="text-xs text-gray-400 mt-1">
          Status: {{ loop.status }} · v{{ loop.version }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { useAgentLoopStore } from '@/stores/agentLoop'

const route = useRoute()
const router = useRouter()
const store = useAgentLoopStore()
const { getWorkspaceId } = useWorkspaceId()
const showCreate = ref(false)

onMounted(async () => {
  const wsId = await getWorkspaceId()
  if (wsId) store.fetchLoops(wsId)
})

async function startLoop(loopId: number) {
  const wsId = await getWorkspaceId()
  if (!wsId) return
  try {
    const run = await store.startLoop(wsId, loopId)
    router.push(`/agents/loops/runs/${run.id}`)
  } catch (e) {
    console.error('Failed to start loop:', e)
  }
}
</script>
