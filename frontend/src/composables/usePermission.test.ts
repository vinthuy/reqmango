import { describe, it, expect, beforeEach, vi } from 'vitest'
import { nextTick, ref } from 'vue'
import { flushPromises } from '@vue/test-utils'

const mockGet = vi.fn()
vi.mock('@/api', () => ({
  default: { get: (...args: any[]) => mockGet(...args) },
}))

import { useProjectPermissions, resetProjectPermissionsCache } from './usePermission'

describe('useProjectPermissions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    resetProjectPermissionsCache()
    localStorage.setItem('token', 'a')
  })

  it('denies everything until loaded, then reflects the API', async () => {
    mockGet.mockResolvedValue({ data: { permissions: ['issue:edit'], role_level: 15 } })
    const { can, loaded, roleLevel } = useProjectPermissions(1)
    expect(loaded.value).toBe(false)
    expect(can('issue:edit')).toBe(false)

    await flushPromises()
    expect(mockGet).toHaveBeenCalledWith('/projects/1/my-permissions')
    expect(loaded.value).toBe(true)
    expect(roleLevel.value).toBe(15)
    expect(can('issue:edit')).toBe(true)
    expect(can('issue:delete')).toBe(false)
  })

  it('shares one request per project across components', async () => {
    mockGet.mockResolvedValue({ data: { permissions: ['issue:delete'], role_level: 20 } })
    const a = useProjectPermissions(1)
    const b = useProjectPermissions(1)
    await flushPromises()
    expect(mockGet).toHaveBeenCalledTimes(1)
    expect(a.can('issue:delete')).toBe(true)
    expect(b.can('issue:delete')).toBe(true)
  })

  it('follows a reactive project id', async () => {
    mockGet.mockImplementation((url: string) =>
      Promise.resolve({ data: { permissions: url.includes('/2/') ? ['issue:delete'] : [], role_level: 15 } }),
    )
    const id = ref<number | null>(null)
    const { can } = useProjectPermissions(id)
    expect(mockGet).not.toHaveBeenCalled()

    id.value = 2
    await nextTick()
    await flushPromises()
    expect(can('issue:delete')).toBe(true)

    id.value = 3
    await nextTick()
    await flushPromises()
    expect(can('issue:delete')).toBe(false)
  })

  it('stays denied and retries later when the request fails', async () => {
    mockGet.mockRejectedValueOnce(new Error('boom'))
    const first = useProjectPermissions(1)
    await flushPromises()
    expect(first.loaded.value).toBe(false)
    expect(first.can('issue:edit')).toBe(false)

    mockGet.mockResolvedValue({ data: { permissions: ['issue:edit'], role_level: 15 } })
    const second = useProjectPermissions(1)
    await flushPromises()
    expect(second.can('issue:edit')).toBe(true)
    expect(first.can('issue:edit')).toBe(true)
  })

  it('refetches when the signed-in user changes', async () => {
    mockGet.mockResolvedValue({ data: { permissions: ['issue:delete'], role_level: 20 } })
    useProjectPermissions(1)
    await flushPromises()

    localStorage.setItem('token', 'b')
    mockGet.mockResolvedValue({ data: { permissions: [], role_level: 15 } })
    const { can } = useProjectPermissions(1)
    await flushPromises()
    expect(mockGet).toHaveBeenCalledTimes(2)
    expect(can('issue:delete')).toBe(false)
  })
})
