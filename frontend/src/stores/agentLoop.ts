import { defineStore } from 'pinia'
import { ref } from 'vue'
import { loopApi, type Loop, type LoopRun, type LoopRunDetail } from '@/api/agent-loop'

export const useAgentLoopStore = defineStore('agentLoop', () => {
  const loops = ref<Loop[]>([])
  const currentRun = ref<LoopRun | null>(null)
  const runDetail = ref<LoopRunDetail | null>(null)
  const loading = ref(false)

  async function fetchLoops(workspaceId: number) {
    loading.value = true
    try {
      loops.value = await loopApi.list(workspaceId)
    } finally {
      loading.value = false
    }
  }

  async function startLoop(workspaceId: number, loopId: number) {
    const run = await loopApi.start(workspaceId, loopId)
    currentRun.value = run
    return run
  }

  async function stopLoop(workspaceId: number, runId: number) {
    await loopApi.stop(workspaceId, runId)
    if (currentRun.value?.id === runId) {
      currentRun.value = { ...currentRun.value, status: 'stopped' }
    }
  }

  async function fetchRunDetail(workspaceId: number, runId: number) {
    runDetail.value = await loopApi.getRun(workspaceId, runId)
    currentRun.value = runDetail.value.run
    return runDetail.value
  }

  function watchRun(workspaceId: number, runId: number, intervalMs = 5000) {
    const interval = setInterval(async () => {
      try {
        const detail = await loopApi.getRun(workspaceId, runId)
        runDetail.value = detail
        currentRun.value = detail.run
        if (detail.run.status !== 'running') {
          clearInterval(interval)
        }
      } catch {
        clearInterval(interval)
      }
    }, intervalMs)
    return () => clearInterval(interval)
  }

  return { loops, currentRun, runDetail, loading, fetchLoops, startLoop, stopLoop, fetchRunDetail, watchRun }
})
