import { defineStore } from 'pinia'
import { ref } from 'vue'
import { pipelineApi, type Pipeline, type PipelineRun } from '@/api/agent-pipeline'

export const useAgentPipelineStore = defineStore('agentPipeline', () => {
  const pipelines = ref<Pipeline[]>([])
  const currentRun = ref<PipelineRun | null>(null)
  const loading = ref(false)

  async function fetchPipelines(ws: number) { loading.value=true; try{pipelines.value=await pipelineApi.list(ws)}finally{loading.value=false} }
  async function runPipeline(ws: number, id: number) { const r = await pipelineApi.run(ws, id); currentRun.value = r; return r }
  async function fetchRun(ws: number, runId: number) { currentRun.value = await pipelineApi.getRun(ws, runId); return currentRun.value }

  function watchRun(ws: number, runId: number, interval=3000) {
    const iv = setInterval(async()=>{try{const r=await pipelineApi.getRun(ws,runId);currentRun.value=r;if(r.status!=='running')clearInterval(iv)}catch{clearInterval(iv)}},interval)
    return ()=>clearInterval(iv)
  }

  return { pipelines, currentRun, loading, fetchPipelines, runPipeline, fetchRun, watchRun }
})
