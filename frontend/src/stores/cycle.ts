import { defineStore } from 'pinia'
import { ref } from 'vue'
import { cycleApi } from '@/api/cycle'
import type { CycleResponse, CycleCreate, CycleUpdate, CycleProgress, CycleStatistics, BurndownData } from '@/types/cycle'

export const useCycleStore = defineStore('cycle', () => {
  // ==================== State ====================
  const cycles = ref<CycleResponse[]>([])
  const currentCycle = ref<CycleResponse | null>(null)
  const progress = ref<CycleProgress | null>(null)
  const statistics = ref<CycleStatistics | null>(null)
  const burndown = ref<BurndownData | null>(null)
  const cycleIssues = ref<any[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // ==================== List ====================
  async function fetchCycles(projectId: number, status?: string) {
    isLoading.value = true
    error.value = null
    try {
      const result = await cycleApi.listCycles(projectId, { status })
      cycles.value = result.items
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  // ==================== Get One ====================
  async function fetchCycle(cycleId: number) {
    isLoading.value = true
    error.value = null
    try {
      currentCycle.value = await cycleApi.getCycle(cycleId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  // ==================== CRUD ====================
  async function createCycleAction(projectId: number, workspaceId: number, data: CycleCreate) {
    isLoading.value = true
    error.value = null
    try {
      const created = await cycleApi.createCycle(projectId, workspaceId, data)
      cycles.value.unshift(created)
      return created
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function updateCycleAction(cycleId: number, data: CycleUpdate) {
    error.value = null
    try {
      const updated = await cycleApi.updateCycle(cycleId, data)
      const idx = cycles.value.findIndex(c => c.id === cycleId)
      if (idx !== -1) cycles.value[idx] = updated
      if (currentCycle.value?.id === cycleId) currentCycle.value = updated
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function deleteCycleAction(cycleId: number) {
    error.value = null
    try {
      await cycleApi.deleteCycle(cycleId)
      cycles.value = cycles.value.filter(c => c.id !== cycleId)
      if (currentCycle.value?.id === cycleId) currentCycle.value = null
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ==================== Status Transitions ====================
  async function startCycle(cycleId: number) {
    error.value = null
    try {
      const updated = await cycleApi.startCycle(cycleId)
      updateCycleInList(updated)
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function endCycle(cycleId: number) {
    error.value = null
    try {
      const updated = await cycleApi.endCycle(cycleId)
      updateCycleInList(updated)
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function cancelCycle(cycleId: number) {
    error.value = null
    try {
      const updated = await cycleApi.cancelCycle(cycleId)
      updateCycleInList(updated)
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  // ==================== Issue Association ====================
  async function addIssueToCycle(cycleId: number, issueId: number) {
    error.value = null
    try {
      const result = await cycleApi.addIssueToCycle(cycleId, issueId)
      await fetchCycleIssues(cycleId)
      await fetchProgress(cycleId)
      return result
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function removeIssueFromCycle(cycleId: number, issueId: number) {
    error.value = null
    try {
      const result = await cycleApi.removeIssueFromCycle(cycleId, issueId)
      cycleIssues.value = cycleIssues.value.filter((i: any) => i.id !== issueId)
      await fetchProgress(cycleId)
      return result
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function fetchCycleIssues(cycleId: number, filters?: { state_id?: number; priority?: string }) {
    isLoading.value = true
    error.value = null
    try {
      cycleIssues.value = await cycleApi.getCycleIssues(cycleId, filters)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  // ==================== Analysis ====================
  async function fetchProgress(cycleId: number) {
    try {
      progress.value = await cycleApi.getCycleProgress(cycleId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  async function fetchStatistics(cycleId: number) {
    try {
      statistics.value = await cycleApi.getCycleStatistics(cycleId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  async function fetchBurndown(cycleId: number) {
    try {
      burndown.value = await cycleApi.getBurndownData(cycleId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ==================== Helper ====================
  function updateCycleInList(updated: CycleResponse) {
    const idx = cycles.value.findIndex(c => c.id === updated.id)
    if (idx !== -1) cycles.value[idx] = updated
    if (currentCycle.value?.id === updated.id) currentCycle.value = updated
  }

  // ==================== Return ====================
  return {
    cycles, currentCycle, progress, statistics, burndown, cycleIssues, isLoading, error,
    fetchCycles, fetchCycle,
    createCycleAction, updateCycleAction, deleteCycleAction,
    startCycle, endCycle, cancelCycle,
    addIssueToCycle, removeIssueFromCycle, fetchCycleIssues,
    fetchProgress, fetchStatistics, fetchBurndown,
  }
})
