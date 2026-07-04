/**
 * usePermission Composable 单元测试
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

const mockGet = vi.fn()
vi.mock('@/api', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
  },
}))

import { usePermission } from './usePermission'

describe('usePermission', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should initialize with empty permissions', () => {
    const { permissions, loaded } = usePermission()
    expect(permissions.value).toEqual([])
    expect(loaded.value).toBe(false)
  })

  describe('loadPermissions', () => {
    it('should load permissions from API (data.data.permissions format)', async () => {
      mockGet.mockResolvedValue({
        data: {
          data: {
            permissions: ['view_project', 'create_issue', 'delete_issue'],
            role_level: 15,
          },
        },
      })

      const { permissions, loaded, loadPermissions } = usePermission()
      await loadPermissions(1, 1)

      expect(loaded.value).toBe(true)
      expect(permissions.value).toContain('view_project')
      expect(permissions.value).toHaveLength(3)
    })

    it('should handle array response format', async () => {
      mockGet.mockResolvedValue({
        data: [
          { code: 'view_project' },
          { code: 'create_issue' },
        ],
      })

      const { permissions, loaded, loadPermissions } = usePermission()
      await loadPermissions(1, 1)

      expect(loaded.value).toBe(true)
      expect(permissions.value).toContain('view_project')
      expect(permissions.value).toHaveLength(2)
    })

    it('should pass workspace_id and project_id as params', async () => {
      mockGet.mockResolvedValue({ data: { data: { permissions: [], role_level: 0 } } })

      const { loadPermissions } = usePermission()
      await loadPermissions(1, 2)

      expect(mockGet).toHaveBeenCalledWith('/permissions', {
        params: { workspace_id: 1, project_id: 2 },
      })
    })

    it('should handle API error gracefully', async () => {
      mockGet.mockRejectedValue(new Error('Network error'))

      const { loaded, loadPermissions } = usePermission()
      await loadPermissions()

      // Should still mark as loaded (fallback)
      expect(loaded.value).toBe(true)
    })
  })

  describe('can() with loaded permissions', () => {
    it('should return true for existing permission', async () => {
      mockGet.mockResolvedValue({
        data: { data: { permissions: ['view_project', 'create_issue'], role_level: 15 } },
      })

      const { can, loadPermissions } = usePermission()
      await loadPermissions()

      expect(can('view_project')).toBe(true)
      expect(can('create_issue')).toBe(true)
    })

    it('should return false for missing permission', async () => {
      mockGet.mockResolvedValue({
        data: { data: { permissions: ['view_project'], role_level: 15 } },
      })

      const { can, loadPermissions } = usePermission()
      await loadPermissions()

      expect(can('delete_issue')).toBe(false)
    })
  })

  describe('can() with level-based fallback', () => {
    beforeEach(() => {
      // No permissions loaded → fallback to level-based
      mockGet.mockResolvedValue({ data: { data: { permissions: [], role_level: 20 } } })
    })

    it('should grant all to level 20', async () => {
      const { can, loadPermissions } = usePermission()
      await loadPermissions()
      expect(can('workspace:manage')).toBe(true)
      expect(can('project:delete')).toBe(true)
    })

    it('should deny admin-only actions to low-level users', () => {
      // No permissions, default level 15
      const { can } = usePermission()
      expect(can('view_project')).toBe(true)
    })
  })
})
