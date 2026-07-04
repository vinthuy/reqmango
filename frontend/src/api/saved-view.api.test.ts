/**
 * Saved View API 单元测试 - Mock axios 验证端点
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

import savedViewApi, {
  listSavedViews, getSavedView, createSavedView, updateSavedView,
  deleteSavedView, setDefaultView, duplicateSavedView,
} from './saved-view'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('SavedView CRUD API', () => {
  it('listSavedViews should GET for project', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listSavedViews(1)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/views')
  })

  it('getSavedView should GET by project/view id', async () => {
    mockGet.mockResolvedValue({ data: { id: 42, name: 'My View' } })
    await getSavedView(1, 42)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/views/42')
  })

  it('createSavedView should POST to project', async () => {
    const data = { name: 'New View', view_type: 'list' as const, filters: {} }
    mockPost.mockResolvedValue({ data: { id: 1, ...data } })
    await createSavedView(1, data)
    expect(mockPost).toHaveBeenCalledWith('/projects/1/views', data)
  })

  it('updateSavedView should PUT by id', async () => {
    const data = { name: 'Updated View' }
    mockPut.mockResolvedValue({ data: { id: 42, name: 'Updated View' } })
    await updateSavedView(1, 42, data)
    expect(mockPut).toHaveBeenCalledWith('/projects/1/views/42', data)
  })

  it('deleteSavedView should DELETE by id', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await deleteSavedView(1, 42)
    expect(mockDelete).toHaveBeenCalledWith('/projects/1/views/42')
  })
})

describe('SavedView special actions API', () => {
  it('setDefaultView should POST to set-default', async () => {
    mockPost.mockResolvedValue({ data: { id: 42, is_default: true } })
    await setDefaultView(1, 42)
    expect(mockPost).toHaveBeenCalledWith('/projects/1/views/42/set-default')
  })

  it('duplicateSavedView should POST to duplicate', async () => {
    mockPost.mockResolvedValue({ data: { id: 43, name: 'My View (copy)' } })
    await duplicateSavedView(1, 42)
    expect(mockPost).toHaveBeenCalledWith('/projects/1/views/42/duplicate')
  })
})

describe('savedViewApi export', () => {
  it('should export all 7 methods', () => {
    expect(Object.keys(savedViewApi)).toHaveLength(7)
    expect(savedViewApi.listSavedViews).toBeDefined()
    expect(savedViewApi.setDefaultView).toBeDefined()
    expect(savedViewApi.duplicateSavedView).toBeDefined()
  })
})
