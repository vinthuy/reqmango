/**
 * Cycle API 单元测试 - Mock axios 验证端点 URL 和参数
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock the API client
const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPut = vi.fn()
const mockDelete = vi.fn()

vi.mock('./index', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    put: (...args: any[]) => mockPut(...args),
    delete: (...args: any[]) => mockDelete(...args),
  },
}))

import {
  createCycle, listCycles, getCycle, updateCycle, deleteCycle,
  startCycle, endCycle, cancelCycle,
  addIssueToCycle, removeIssueFromCycle, getCycleIssues,
  getCycleProgress, getCycleStatistics, getBurndownData,
  cycleApi,
} from './cycle'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('cycle CRUD API', () => {
  it('createCycle should POST to correct URL', async () => {
    mockPost.mockResolvedValue({ data: { id: 1, name: 'Sprint 1' } })
    const result = await createCycle(1, 1, { name: 'Sprint 1', project_id: 1 })
    expect(mockPost).toHaveBeenCalledWith('/projects/1/cycles?workspace_id=1', { name: 'Sprint 1', project_id: 1 })
    expect(result.name).toBe('Sprint 1')
  })

  it('listCycles should GET without status filter', async () => {
    mockGet.mockResolvedValue({ data: { items: [], total: 0, limit: 50, offset: 0 } })
    await listCycles(1)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/cycles?')
  })

  it('listCycles should GET with status filter', async () => {
    mockGet.mockResolvedValue({ data: { items: [], total: 0, limit: 50, offset: 0 } })
    await listCycles(1, { status: 'active' })
    expect(mockGet).toHaveBeenCalledWith('/projects/1/cycles?status=active')
  })

  it('listCycles should GET with limit', async () => {
    mockGet.mockResolvedValue({ data: { items: [], total: 0, limit: 10, offset: 0 } })
    await listCycles(1, { limit: 10 })
    const url = mockGet.mock.calls[0][0]
    expect(url).toContain('/projects/1/cycles?')
    expect(url).toContain('limit=10')
  })

  it('listCycles should GET with offset', async () => {
    mockGet.mockResolvedValue({ data: { items: [], total: 0, limit: 50, offset: 20 } })
    await listCycles(1, { offset: 20 })
    expect(mockGet).toHaveBeenCalledWith('/projects/1/cycles?offset=20')
  })

  it('getCycle should GET by id', async () => {
    mockGet.mockResolvedValue({ data: { id: 42 } })
    const result = await getCycle(42)
    expect(mockGet).toHaveBeenCalledWith('/cycles/42')
    expect(result.id).toBe(42)
  })

  it('updateCycle should PUT to correct URL', async () => {
    mockPut.mockResolvedValue({ data: { id: 42, name: 'Updated' } })
    await updateCycle(42, { name: 'Updated' })
    expect(mockPut).toHaveBeenCalledWith('/cycles/42', { name: 'Updated' })
  })

  it('deleteCycle should DELETE by id', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await deleteCycle(42)
    expect(mockDelete).toHaveBeenCalledWith('/cycles/42')
  })
})

describe('cycle status transitions API', () => {
  it('startCycle should POST to start endpoint', async () => {
    mockPost.mockResolvedValue({ data: { id: 1, status: 'active' } })
    await startCycle(1)
    expect(mockPost).toHaveBeenCalledWith('/cycles/1/start')
  })

  it('endCycle should POST to end endpoint', async () => {
    mockPost.mockResolvedValue({ data: { id: 1, status: 'completed' } })
    await endCycle(1)
    expect(mockPost).toHaveBeenCalledWith('/cycles/1/end')
  })

  it('cancelCycle should POST to cancel endpoint', async () => {
    mockPost.mockResolvedValue({ data: { id: 1, status: 'cancelled' } })
    await cancelCycle(1)
    expect(mockPost).toHaveBeenCalledWith('/cycles/1/cancel')
  })
})

describe('cycle issue association API', () => {
  it('addIssueToCycle should POST with issue_id query', async () => {
    mockPost.mockResolvedValue({ data: { cycle_id: 1, issue_id: 42, action: 'add' } })
    await addIssueToCycle(1, 42)
    expect(mockPost).toHaveBeenCalledWith('/cycles/1/issues?issue_id=42')
  })

  it('removeIssueFromCycle should DELETE with issue id in path', async () => {
    mockDelete.mockResolvedValue({ data: { cycle_id: 1, issue_id: 42, action: 'remove' } })
    await removeIssueFromCycle(1, 42)
    expect(mockDelete).toHaveBeenCalledWith('/cycles/1/issues/42')
  })

  it('getCycleIssues should GET with optional filters', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await getCycleIssues(1, { state_id: 3, priority: 'high' })
    expect(mockGet).toHaveBeenCalledWith('/cycles/1/issues?state_id=3&priority=high')
  })
})

describe('cycle analysis API', () => {
  it('getCycleProgress should GET progress', async () => {
    mockGet.mockResolvedValue({ data: { cycle_id: 1, progress: 50 } })
    await getCycleProgress(1)
    expect(mockGet).toHaveBeenCalledWith('/cycles/1/progress')
  })

  it('getCycleStatistics should GET statistics', async () => {
    mockGet.mockResolvedValue({ data: { cycle_id: 1 } })
    await getCycleStatistics(1)
    expect(mockGet).toHaveBeenCalledWith('/cycles/1/statistics')
  })

  it('getBurndownData should GET burndown', async () => {
    mockGet.mockResolvedValue({ data: { cycle_id: 1 } })
    await getBurndownData(1)
    expect(mockGet).toHaveBeenCalledWith('/cycles/1/burndown')
  })
})

describe('cycleApi export', () => {
  it('should export all methods', () => {
    expect(cycleApi.createCycle).toBeDefined()
    expect(cycleApi.listCycles).toBeDefined()
    expect(cycleApi.getCycle).toBeDefined()
    expect(cycleApi.updateCycle).toBeDefined()
    expect(cycleApi.deleteCycle).toBeDefined()
    expect(cycleApi.startCycle).toBeDefined()
    expect(cycleApi.endCycle).toBeDefined()
    expect(cycleApi.cancelCycle).toBeDefined()
    expect(cycleApi.addIssueToCycle).toBeDefined()
    expect(cycleApi.removeIssueFromCycle).toBeDefined()
    expect(cycleApi.getCycleIssues).toBeDefined()
    expect(cycleApi.getCycleProgress).toBeDefined()
    expect(cycleApi.getCycleStatistics).toBeDefined()
    expect(cycleApi.getBurndownData).toBeDefined()
    expect(cycleApi.applyAutoAddRules).toBeDefined()
    expect(cycleApi.applyAutoCloseRules).toBeDefined()
    expect(Object.keys(cycleApi)).toHaveLength(16)
  })
})
