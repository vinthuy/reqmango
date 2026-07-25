/**
 * Memory API — AI记忆系统API模块
 */
import api from './index'

export interface MemoryEntry {
  id: number
  workspace_id: number
  project_id?: number
  issue_id?: number
  agent_id?: number
  memory_type: 'short_term' | 'medium_term' | 'long_term'
  scope: 'workspace' | 'project' | 'issue' | 'agent'
  content: string
  embedding?: number[]
  metadata?: Record<string, any>
  tags?: string[]
  context_key: string
  context_name?: string
  relevance_score: number
  expires_at?: string
  created_at: string
  updated_at: string
}

export interface MemorySession {
  id: string
  workspace_id: number
  session_type: string
  context_id?: string
  started_at: string
  updated_at: string
  closed_at?: string
  memory_count: number
}

export interface MemoryListFilters {
  project_id?: number
  issue_id?: number
  agent_id?: number
  memory_type?: string
  scope?: string
  context_key?: string
  tag?: string
  limit?: number
  offset?: number
}

// ==================== Memory CRUD ====================

export async function listMemories(
  workspaceId: number,
  filters?: MemoryListFilters
): Promise<MemoryEntry[]> {
  const params: Record<string, any> = {}
  if (filters) {
    if (filters.project_id) params.project_id = filters.project_id
    if (filters.issue_id) params.issue_id = filters.issue_id
    if (filters.agent_id) params.agent_id = filters.agent_id
    if (filters.memory_type) params.memory_type = filters.memory_type
    if (filters.scope) params.scope = filters.scope
    if (filters.context_key) params.context_key = filters.context_key
    if (filters.tag) params.tag = filters.tag
    if (filters.limit) params.limit = filters.limit
    if (filters.offset) params.offset = filters.offset
  }
  const response = await api.get(`/workspaces/${workspaceId}/memories`, { params })
  return response.data
}

export async function getMemory(workspaceId: number, memoryId: number): Promise<MemoryEntry> {
  const response = await api.get(`/workspaces/${workspaceId}/memories/${memoryId}`)
  return response.data
}

export async function createMemory(
  workspaceId: number,
  data: Partial<MemoryEntry>
): Promise<MemoryEntry> {
  const response = await api.post(`/workspaces/${workspaceId}/memories`, data)
  return response.data
}

export async function updateMemory(
  workspaceId: number,
  memoryId: number,
  data: Partial<MemoryEntry>
): Promise<MemoryEntry> {
  const response = await api.put(`/workspaces/${workspaceId}/memories/${memoryId}`, data)
  return response.data
}

export async function deleteMemory(workspaceId: number, memoryId: number): Promise<void> {
  await api.delete(`/workspaces/${workspaceId}/memories/${memoryId}`)
}

// ==================== Memory Search ====================

export async function searchMemories(
  workspaceId: number,
  query: string,
  limit?: number
): Promise<MemoryEntry[]> {
  const response = await api.post(`/workspaces/${workspaceId}/memories/search`, {
    query,
    limit: limit || 10,
  })
  return response.data
}

export async function semanticSearchMemories(
  workspaceId: number,
  embedding: number[],
  limit?: number
): Promise<MemoryEntry[]> {
  const response = await api.post(`/workspaces/${workspaceId}/memories/semantic-search`, {
    embedding,
    limit: limit || 10,
  })
  return response.data
}

export async function getContextMemories(
  workspaceId: number,
  contextKey: string
): Promise<MemoryEntry[]> {
  const response = await api.get(`/workspaces/${workspaceId}/memories/context/${contextKey}`)
  return response.data
}

// ==================== Memory Sessions ====================

export async function createMemorySession(
  workspaceId: number,
  sessionType?: string
): Promise<MemorySession> {
  const response = await api.post(`/workspaces/${workspaceId}/memory-sessions`, {
    session_type: sessionType || 'chat',
  })
  return response.data
}

export async function getMemorySession(
  workspaceId: number,
  sessionId: string
): Promise<MemorySession> {
  const response = await api.get(`/workspaces/${workspaceId}/memory-sessions/${sessionId}`)
  return response.data
}

export async function closeMemorySession(workspaceId: number, sessionId: string): Promise<void> {
  await api.put(`/workspaces/${workspaceId}/memory-sessions/${sessionId}/close`)
}

export default {
  listMemories,
  getMemory,
  createMemory,
  updateMemory,
  deleteMemory,
  searchMemories,
  semanticSearchMemories,
  getContextMemories,
  createMemorySession,
  getMemorySession,
  closeMemorySession,
}