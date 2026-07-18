<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <button @click="$router.back()" class="text-gray-400 hover:text-gray-600">← Back</button>
      <h1 class="text-xl font-bold">Pipeline Run #{{ runId }}</h1>
      <span v-if="run" :class="statusClass" class="px-2 py-0.5 rounded-full text-xs font-medium">{{ run.status }}</span>
    </div>
    <div v-if="!run" class="text-center py-12 text-gray-400">Loading...</div>
    <template v-else>
      <div class="border rounded-lg p-4 mb-6 dark:border-gray-700">
        <div class="text-sm text-gray-500">Tokens: {{ run.tokens_used.toLocaleString() }} · Cost: ${{ run.cost_usd.toFixed(4) }}</div>
        <div v-if="run.error_message" class="text-sm text-red-500 mt-1">{{ run.error_message }}</div>
      </div>
      <h2 class="font-semibold mb-3">Stage Results</h2>
      <div v-if="run.stages_result" class="space-y-3">
        <div v-for="(stage, i) in run.stages_result" :key="i" class="border rounded-lg p-4 dark:border-gray-700">
          <div class="flex items-center justify-between mb-2">
            <span class="font-medium text-sm">{{ stage.stage_name }}</span>
            <span class="text-xs text-gray-400">{{ stage.stage_type }} · {{ stage.tokens_used }} tokens · {{ stage.duration_ms }}ms</span>
          </div>
          <p v-if="stage.error" class="text-sm text-red-500">{{ stage.error }}</p>
          <p v-else class="text-sm text-gray-600 dark:text-gray-400 line-clamp-3">{{ stage.output }}</p>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAgentPipelineStore } from '@/stores/agentPipeline'
import type { PipelineRun } from '@/api/agent-pipeline'

const route = useRoute()
const runId = Number(route.params.runId)
const store = useAgentPipelineStore()
const run = ref<PipelineRun|null>(null)
let stop: (()=>void)|null=null

const statusClass = computed(()=>{
  switch(run.value?.status){case'running':return'bg-blue-100 text-blue-800';case'completed':return'bg-green-100 text-green-800';case'failed':return'bg-red-100 text-red-800';default:return'bg-gray-100'}
})

function getWS(){const m=window.location.pathname.match(/\/workspaces\/(\d+)/);return m?Number(m[1]):null}

onMounted(async()=>{const ws=getWS();if(ws){run.value=await store.fetchRun(ws,runId);if(run.value?.status==='running')stop=store.watchRun(ws,runId,3000)}})
onUnmounted(()=>stop?.())
</script>
