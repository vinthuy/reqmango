import { defineStore } from 'pinia'
import { ref } from 'vue'
import moduleApi from '@/api/module'
import type { ModuleResponse, ModuleCreate, ModuleUpdate, ModuleProgress, ModuleTreeNode, ModuleOverrideRequest } from '@/types/module'

export const useModuleStore = defineStore('module', () => {
  const modules = ref<ModuleResponse[]>([])
  const moduleTree = ref<ModuleTreeNode[]>([])
  const currentModule = ref<ModuleResponse | null>(null)
  const progress = ref<ModuleProgress | null>(null)
  const moduleIssues = ref<any[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  async function fetchModules(projectId: number, workspaceId: number) {
    isLoading.value = true; error.value = null
    try { modules.value = await moduleApi.listModules(projectId, workspaceId) }
    catch (e: any) { error.value = e.response?.data?.message || e.message }
    finally { isLoading.value = false }
  }

  async function fetchModuleTree(projectId: number, workspaceId: number) {
    try { moduleTree.value = await moduleApi.getModuleTree(projectId, workspaceId) }
    catch (e: any) { error.value = e.response?.data?.message || e.message }
  }

  async function createModule(workspaceId: number, data: ModuleCreate) {
    error.value = null
    try { const created = await moduleApi.createModule(workspaceId, data); modules.value.unshift(created); await fetchModuleTree(data.project_id, data.workspace_id); return created }
    catch (e: any) { error.value = e.response?.data?.message || e.message; return null }
  }

  async function updateModuleAction(id: number, data: ModuleUpdate) {
    error.value = null
    try {
      const updated = await moduleApi.updateModule(id, data)
      const idx = modules.value.findIndex(m => m.id === id)
      if (idx !== -1) modules.value[idx] = updated
      if (currentModule.value?.id === id) currentModule.value = updated
      if (updated.project_id) await fetchModuleTree(updated.project_id, updated.workspace_id)
      return updated
    } catch (e: any) { error.value = e.response?.data?.message || e.message; return null }
  }

  async function deleteModuleAction(id: number) {
    error.value = null
    try {
      const m = modules.value.find(m => m.id === id)
      await moduleApi.deleteModule(id)
      modules.value = modules.value.filter(m => m.id !== id)
      if (currentModule.value?.id === id) currentModule.value = null
      if (m?.project_id) await fetchModuleTree(m.project_id, m.workspace_id)
    } catch (e: any) { error.value = e.response?.data?.message || e.message }
  }

  async function createOrUpdateOverride(projectId: number, moduleId: number, data: ModuleOverrideRequest) {
    error.value = null
    try {
      const updated = await moduleApi.createOrUpdateModuleOverride(projectId, moduleId, data)
      const idx = modules.value.findIndex(m => m.id === moduleId)
      if (idx !== -1) modules.value[idx] = updated
      await fetchModuleTree(projectId, updated.workspace_id)
      return updated
    } catch (e: any) { error.value = e.response?.data?.message || e.message; return null }
  }

  async function deleteOverride(projectId: number, moduleId: number, workspaceId: number) {
    error.value = null
    try {
      await moduleApi.deleteModuleOverride(projectId, moduleId)
      await fetchModules(projectId, workspaceId)
      await fetchModuleTree(projectId, workspaceId)
    } catch (e: any) { error.value = e.response?.data?.message || e.message }
  }

  async function addIssueToModule(moduleId: number, issueId: number) {
    error.value = null
    try { const result = await moduleApi.addIssueToModule(moduleId, issueId); await fetchModuleIssues(moduleId); await fetchProgress(moduleId); return result }
    catch (e: any) { error.value = e.response?.data?.message || e.message; return null }
  }

  async function removeIssueFromModule(moduleId: number, issueId: number) {
    error.value = null
    try {
      await moduleApi.removeIssueFromModule(moduleId, issueId)
      moduleIssues.value = moduleIssues.value.filter((i: any) => i.id !== issueId)
      await fetchProgress(moduleId)
    } catch (e: any) { error.value = e.response?.data?.message || e.message }
  }

  async function fetchModuleIssues(moduleId: number, filters?: { state_id?: number; priority?: string }) {
    isLoading.value = true
    try { moduleIssues.value = await moduleApi.getModuleIssues(moduleId, filters) }
    catch (e: any) { error.value = e.response?.data?.message || e.message }
    finally { isLoading.value = false }
  }

  async function fetchProgress(moduleId: number) {
    try { progress.value = await moduleApi.getModuleProgress(moduleId) }
    catch (e: any) { error.value = e.response?.data?.message || e.message }
  }

  return {
    modules, moduleTree, currentModule, progress, moduleIssues, isLoading, error,
    fetchModules, fetchModuleTree, createModule, updateModuleAction, deleteModuleAction,
    addIssueToModule, removeIssueFromModule, fetchModuleIssues, fetchProgress,
    createOrUpdateOverride, deleteOverride,
  }
})
