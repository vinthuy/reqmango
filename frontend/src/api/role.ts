import api from './index'
import type { Role, Permission, CreateRoleRequest, UpdateRoleRequest } from '../types/role'

export const roleApi = {
  /** List roles for a workspace */
  listRoles(workspaceId: number | string) {
    return api.get<{ data: Role[] }>(`/workspaces/${workspaceId}/roles`)
  },

  /** List all permissions */
  listPermissions() {
    return api.get<{ data: Permission[] }>('/permissions')
  },

  /** Create a custom role */
  createRole(workspaceId: number | string, data: CreateRoleRequest) {
    return api.post<{ data: Role }>(`/workspaces/${workspaceId}/roles`, data)
  },

  /** Update a role */
  updateRole(workspaceId: number | string, roleId: number, data: UpdateRoleRequest) {
    return api.put<{ data: Role }>(`/workspaces/${workspaceId}/roles/${roleId}`, data)
  },

  /** Delete a custom role */
  deleteRole(workspaceId: number | string, roleId: number) {
    return api.delete<{ data: { deleted: boolean } }>(`/workspaces/${workspaceId}/roles/${roleId}`)
  },
}
