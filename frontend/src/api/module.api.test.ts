/**
 * Module API 单元测试 - Mock axios 验证端点
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'

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

import moduleApi, {
  createModule, listModules, getModule, updateModule, deleteModule,
  addIssueToModule, removeIssueFromModule, getModuleIssues,
  getModuleProgress, getModuleStatistics, getModuleTree,
} from './module'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('Module CRUD API', () => {
  it('createModule should POST with workspace_id', async () => {
    mockPost.mockResolvedValue({ data: { id: 1, name: 'Core' } })
    await createModule(1, { name: 'Core', project_id: 1, workspace_id: 1 })
    expect(mockPost).toHaveBeenCalledWith('/modules?workspace_id=1', { name: 'Core', project_id: 1, workspace_id: 1 })
  })

  it('listModules should GET with project_id and workspace_id', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listModules(1, 1)
    expect(mockGet).toHaveBeenCalledWith('/modules?project_id=1&workspace_id=1')
  })

  it('listModules should append optional filters', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listModules(1, 2, { parent_id: 5, include_archived: true })
    expect(mockGet).toHaveBeenCalledWith('/modules?project_id=1&workspace_id=2&parent_id=5&include_archived=true')
  })

  it('listModules should append limit and offset', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listModules(1, 1, { limit: 20, offset: 40 })
    expect(mockGet).toHaveBeenCalledWith('/modules?project_id=1&workspace_id=1&limit=20&offset=40')
  })

  it('getModule should GET by id', async () => {
    mockGet.mockResolvedValue({ data: { id: 42 } })
    await getModule(42)
    expect(mockGet).toHaveBeenCalledWith('/modules/42')
  })

  it('updateModule should PUT by id', async () => {
    mockPut.mockResolvedValue({ data: { id: 42, name: 'Updated' } })
    await updateModule(42, { name: 'Updated' })
    expect(mockPut).toHaveBeenCalledWith('/modules/42', { name: 'Updated' })
  })

  it('deleteModule should DELETE by id', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await deleteModule(42)
    expect(mockDelete).toHaveBeenCalledWith('/modules/42')
  })
})

describe('Module Issue API', () => {
  it('addIssueToModule should POST with issue_id', async () => {
    mockPost.mockResolvedValue({ data: { module_id: 1, issue_id: 42, action: 'add' } })
    await addIssueToModule(1, 42)
    expect(mockPost).toHaveBeenCalledWith('/modules/1/issues?issue_id=42')
  })

  it('removeIssueFromModule should DELETE by module/issue id', async () => {
    mockDelete.mockResolvedValue({ data: { module_id: 1, issue_id: 42, action: 'remove' } })
    await removeIssueFromModule(1, 42)
    expect(mockDelete).toHaveBeenCalledWith('/modules/1/issues/42')
  })

  it('getModuleIssues should GET with optional state/priority filters', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await getModuleIssues(1, { state_id: 3, priority: 'high' })
    expect(mockGet).toHaveBeenCalledWith('/modules/1/issues?state_id=3&priority=high')
  })
})

describe('Module Analysis API', () => {
  it('getModuleProgress should GET progress', async () => {
    mockGet.mockResolvedValue({ data: { module_id: 1, progress: 60 } })
    await getModuleProgress(1)
    expect(mockGet).toHaveBeenCalledWith('/modules/1/progress')
  })

  it('getModuleStatistics should GET statistics', async () => {
    mockGet.mockResolvedValue({ data: { module_id: 1 } })
    await getModuleStatistics(1)
    expect(mockGet).toHaveBeenCalledWith('/modules/1/statistics')
  })

  it('getModuleTree should GET tree with project_id', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await getModuleTree(1)
    expect(mockGet).toHaveBeenCalledWith('/modules/tree?project_id=1')
  })
})

describe('moduleApi export', () => {
  it('should export all methods', () => {
    const keys = Object.keys(moduleApi)
    // CRUD: 5, Issues: 3, Stats: 2, Tree: 1 = 11
    expect(keys).toHaveLength(11)
  })
})
