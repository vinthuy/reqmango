import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  listProjects,
  getProject,
  createProject,
  updateProject,
  deleteProject,
  archiveProject,
  restoreProject,
  getProjectStatistics,
  getProjectIssuesSummary,
} from '@/api/project'
import type { ProjectCreate, ProjectUpdate, ProjectResponse, ProjectStatistics } from '@/types/project'

export const useProjectStore = defineStore('project', () => {
  // ==================== State ====================
  const projects = ref<ProjectResponse[]>([])
  const currentProject = ref<ProjectResponse | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const statistics = ref<ProjectStatistics | null>(null)

  // ==================== List ====================
  async function fetchProjects(
    workspaceId: number,
    options?: { include_archived?: boolean; limit?: number; offset?: number }
  ) {
    isLoading.value = true
    error.value = null
    try {
      projects.value = await listProjects(workspaceId, options)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  // ==================== Get One ====================
  async function fetchProject(projectId: number) {
    isLoading.value = true
    error.value = null
    try {
      currentProject.value = await getProject(projectId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  // ==================== CRUD ====================
  async function createProjectAction(workspaceId: number, data: ProjectCreate) {
    isLoading.value = true
    error.value = null
    try {
      const created = await createProject(workspaceId, data)
      projects.value.unshift(created)
      return created
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function updateProjectAction(projectId: number, data: ProjectUpdate) {
    error.value = null
    try {
      const updated = await updateProject(projectId, data)
      const idx = projects.value.findIndex(p => p.id === projectId)
      if (idx !== -1) projects.value[idx] = updated
      if (currentProject.value?.id === projectId) currentProject.value = updated
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function deleteProjectAction(projectId: number) {
    error.value = null
    try {
      await deleteProject(projectId)
      projects.value = projects.value.filter(p => p.id !== projectId)
      if (currentProject.value?.id === projectId) currentProject.value = null
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ==================== Archive / Restore ====================
  async function archiveProjectAction(projectId: number) {
    error.value = null
    try {
      const archived = await archiveProject(projectId)
      const idx = projects.value.findIndex(p => p.id === projectId)
      if (idx !== -1) projects.value[idx] = archived
      if (currentProject.value?.id === projectId) currentProject.value = archived
      return archived
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function restoreProjectAction(projectId: number) {
    error.value = null
    try {
      const restored = await restoreProject(projectId)
      const idx = projects.value.findIndex(p => p.id === projectId)
      if (idx !== -1) projects.value[idx] = restored
      if (currentProject.value?.id === projectId) currentProject.value = restored
      return restored
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  // ==================== Statistics ====================
  async function fetchStatistics(projectId: number) {
    error.value = null
    try {
      statistics.value = await getProjectStatistics(projectId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  async function fetchIssuesSummary(projectId: number) {
    error.value = null
    try {
      return await getProjectIssuesSummary(projectId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  // ==================== Return ====================
  return {
    projects, currentProject, isLoading, error, statistics,
    fetchProjects, fetchProject,
    createProjectAction, updateProjectAction, deleteProjectAction,
    archiveProjectAction, restoreProjectAction,
    fetchStatistics, fetchIssuesSummary,
  }
})
