<template>
  <div class="p-6 max-w-5xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">⚙️ Pipeline Builder</h1>
      <button @click="showCreate=true" class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700">+ New Pipeline</button>
    </div>
    <div v-if="store.loading" class="text-center py-12 text-gray-400">Loading...</div>
    <div v-else-if="store.pipelines.length===0" class="text-center py-12 text-gray-400 border rounded-lg dark:border-gray-700">No pipelines yet. Create one to get started.</div>
    <div v-else class="space-y-3">
      <div v-for="p in store.pipelines" :key="p.id" class="border rounded-lg p-4 hover:border-indigo-300 transition-colors dark:border-gray-700">
        <div class="flex items-center justify-between">
          <div>
            <span class="font-medium">{{ p.name }}</span>
            <span class="text-xs text-gray-400 ml-2">v{{ p.version }} · {{ p.status }}</span>
          </div>
          <div class="flex gap-2">
            <button @click="runPipeline(p.id)" class="px-3 py-1 bg-green-600 text-white rounded text-xs hover:bg-green-700">▶ Run</button>
            <button @click="viewPipeline(p.id)" class="px-3 py-1 bg-gray-200 dark:bg-gray-700 rounded text-xs hover:bg-gray-300">View</button>
          </div>
        </div>
        <PipelineDAG v-if="p.pipeline_def?.pipeline?.stages" :stages="p.pipeline_def.pipeline.stages" :mode="p.pipeline_def.pipeline.mode||'sequential'" class="mt-3" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAgentPipelineStore } from '@/stores/agentPipeline'
import PipelineDAG from '@/components/agents/PipelineDAG.vue'

const store = useAgentPipelineStore()
const router = useRouter()
const showCreate = ref(false)

function getWS(): number|null {
  const m = window.location.pathname.match(/\/workspaces\/(\d+)/)
  if (m) return Number(m[1])
  const s = localStorage.getItem('currentWorkspaceId')
  return s ? Number(s) : null
}

onMounted(()=>{const ws=getWS();if(ws)store.fetchPipelines(ws)})

async function runPipeline(id:number){const ws=getWS();if(ws){const r=await store.runPipeline(ws,id);router.push(`/agents/pipelines/runs/${r.id}`)}}
function viewPipeline(id:number){router.push(`/agents/pipelines/${id}`)}
</script>
