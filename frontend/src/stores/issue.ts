import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  listIssues,
  getIssue,
  createIssue,
  updateIssue,
  deleteIssue,
  archiveIssue,
  restoreIssue,
  searchIssues,
  getIssueStatistics,
  bulkUpdateIssues,
  bulkDeleteIssues,
} from '@/api/issue'
import type { IssueCreate, IssueUpdate, IssueResponse, IssueStatistics, IssuePriority } from '@/types/issue'

export const useIssueStore = defineStore('issue', () => {
  // ==================== State ====================
  const issues = ref<IssueResponse[]>([])
  const currentIssue = ref<IssueResponse | null>(null)
  const total = ref(0)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const statistics = ref<IssueStatistics | null>(null)

  // ==================== List ====================
  async function fetchIssues(
    projectId: number,
    workspaceId: number,
    filters?: {
      state_id?: number
      priority?: IssuePriority
      assignee_id?: number
      parent_id?: number
      cycle_id?: number
      module_id?: number
      search?: string
      is_draft?: boolean
      issue_type_id?: number
      rql?: string
      sort_by?: string
      sort_dir?: string
      sort_config?: string
      limit?: number
      offset?: number
    }
  ) {
    isLoading.value = true
    error.value = null
    try {
      const result = await listIssues(projectId, workspaceId, filters)
      issues.value = result.items
      total.value = result.total
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  // ==================== Get One ====================
  async function fetchIssue(issueId: number) {
    isLoading.value = true
    error.value = null
    try {
      currentIssue.value = await getIssue(issueId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  // ==================== CRUD ====================
  async function createIssueAction(projectId: number, workspaceId: number, data: IssueCreate) {
    isLoading.value = true
    error.value = null
    try {
      const created = await createIssue(projectId, workspaceId, data)
      issues.value.unshift(created)
      total.value += 1
      return created
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    } finally {
      isLoading.value = false
    }
  }

  async function updateIssueAction(issueId: number, data: IssueUpdate) {
    error.value = null
    try {
      const updated = await updateIssue(issueId, data)
      const idx = issues.value.findIndex(i => i.id === issueId)
      if (idx !== -1) issues.value[idx] = updated
      if (currentIssue.value?.id === issueId) currentIssue.value = updated
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function deleteIssueAction(issueId: number) {
    error.value = null
    try {
      await deleteIssue(issueId)
      issues.value = issues.value.filter(i => i.id !== issueId)
      total.value -= 1
      if (currentIssue.value?.id === issueId) currentIssue.value = null
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ==================== Archive / Restore ====================
  async function archiveIssueAction(issueId: number) {
    error.value = null
    try {
      await archiveIssue(issueId)
      issues.value = issues.value.filter(i => i.id !== issueId)
      total.value -= 1
      if (currentIssue.value?.id === issueId) currentIssue.value = null
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  async function restoreIssueAction(issueId: number) {
    error.value = null
    try {
      const restored = await restoreIssue(issueId)
      issues.value.unshift(restored)
      total.value += 1
      return restored
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  // ==================== Search ====================
  async function searchIssuesAction(
    workspaceId: number,
    query: string,
    projectId?: number,
    limit?: number
  ) {
    isLoading.value = true
    error.value = null
    try {
      return await searchIssues(workspaceId, query, projectId, limit)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return []
    } finally {
      isLoading.value = false
    }
  }

  // ==================== Statistics ====================
  async function fetchStatistics(projectId: number) {
    error.value = null
    try {
      statistics.value = await getIssueStatistics(projectId)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ==================== Bulk Operations ====================
  async function bulkUpdateAction(projectId: number, issueIds: number[], data: IssueUpdate) {
    isLoading.value = true
    error.value = null
    try {
      const updated = await bulkUpdateIssues(projectId, issueIds, data)
      for (const u of updated) {
        const idx = issues.value.findIndex(i => i.id === u.id)
        if (idx !== -1) issues.value[idx] = u
      }
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return []
    } finally {
      isLoading.value = false
    }
  }

  async function bulkDeleteAction(issueIds: number[]) {
    error.value = null
    try {
      await bulkDeleteIssues(issueIds)
      issues.value = issues.value.filter(i => !issueIds.includes(i.id))
      total.value -= issueIds.length
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ==================== Return ====================
  return {
    issues, currentIssue, total, isLoading, error, statistics,
    fetchIssues, fetchIssue,
    createIssueAction, updateIssueAction, deleteIssueAction,
    archiveIssueAction, restoreIssueAction,
    searchIssuesAction, fetchStatistics,
    bulkUpdateAction, bulkDeleteAction,
  }
})
