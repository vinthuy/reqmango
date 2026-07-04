/**
 * Agent API 单元测试 — 缓存与去重验证
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPut = vi.fn()
const mockDelete = vi.fn()
const mockPatch = vi.fn()

vi.mock('./index', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    put: (...args: any[]) => mockPut(...args),
    delete: (...args: any[]) => mockDelete(...args),
    patch: (...args: any[]) => mockPatch(...args),
  },
}))

import { agentApi, invalidateAgentCache } from './agent'

beforeEach(() => {
  vi.clearAllMocks()
  invalidateAgentCache()
})

describe('agentApi.list cache', () => {
  it('should deduplicate concurrent calls', async () => {
    mockGet.mockResolvedValue({ data: [{ id: 1, name: 'Agent A' }] })

    const [a, b] = await Promise.all([
      agentApi.list(1),
      agentApi.list(1),
    ])

    expect(mockGet).toHaveBeenCalledTimes(1)
    expect(a).toEqual([{ id: 1, name: 'Agent A' }])
    expect(b).toBe(a) // same reference
  })

  it('should serve from cache on second call', async () => {
    mockGet.mockResolvedValue({ data: [{ id: 2 }] })

    await agentApi.list(1)
    await agentApi.list(1)

    expect(mockGet).toHaveBeenCalledTimes(1)
  })

  it('should cache per workspace', async () => {
    mockGet.mockResolvedValue({ data: [] })

    await agentApi.list(1)
    await agentApi.list(2)

    expect(mockGet).toHaveBeenCalledTimes(2)
  })
})

describe('agentApi mutation invalidates cache', () => {
  it('create invalidates list cache', async () => {
    mockGet.mockResolvedValue({ data: [{ id: 1 }] })
    mockPost.mockResolvedValue({ data: { id: 2, name: 'New' } })

    await agentApi.list(1)
    await agentApi.create(1, { name: 'New', key: 'new', system_prompt: '' } as any)
    await agentApi.list(1)

    expect(mockGet).toHaveBeenCalledTimes(2)
  })

  it('update invalidates list cache', async () => {
    mockGet.mockResolvedValue({ data: [{ id: 1 }] })
    mockPut.mockResolvedValue({ data: { id: 1, name: 'Updated' } })

    await agentApi.list(1)
    await agentApi.update(1, 1, { name: 'Updated' })
    await agentApi.list(1)

    expect(mockGet).toHaveBeenCalledTimes(2)
  })

  it('delete invalidates list cache', async () => {
    mockGet.mockResolvedValue({ data: [{ id: 1 }] })
    mockDelete.mockResolvedValue({})

    await agentApi.list(1)
    await agentApi.delete(1, 1)
    await agentApi.list(1)

    expect(mockGet).toHaveBeenCalledTimes(2)
  })
})

describe('invalidateAgentCache', () => {
  it('invalidateAgentCache(ws) clears specific workspace', async () => {
    mockGet.mockResolvedValue({ data: [] })

    await agentApi.list(1)
    await agentApi.list(2)
    expect(mockGet).toHaveBeenCalledTimes(2)

    invalidateAgentCache(1)

    await agentApi.list(1)
    await agentApi.list(2)
    expect(mockGet).toHaveBeenCalledTimes(3) // only ws 1 re-fetches
  })

  it('invalidateAgentCache() clears all', async () => {
    mockGet.mockResolvedValue({ data: [] })

    await agentApi.list(1)
    await agentApi.list(2)
    expect(mockGet).toHaveBeenCalledTimes(2)

    invalidateAgentCache()

    await agentApi.list(1)
    await agentApi.list(2)
    expect(mockGet).toHaveBeenCalledTimes(4)
  })
})
