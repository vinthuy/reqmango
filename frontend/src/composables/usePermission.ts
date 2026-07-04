import { ref } from 'vue'
import api from '@/api'

export function usePermission() {
  const permissions = ref<string[]>([])
  const loaded = ref(false)
  const userLevel = ref<number>(0)

  async function loadPermissions(workspaceId?: number, projectId?: number) {
    try {
      const params: Record<string, number> = {}
      if (workspaceId) params.workspace_id = workspaceId
      if (projectId) params.project_id = projectId
      
      const resp = await api.get('/permissions', { params })
      const data = resp.data
      
      // Handle different response formats
      if (data && data.data && Array.isArray(data.data.permissions)) {
        permissions.value = data.data.permissions
        userLevel.value = data.data.role_level || 0
      } else if (Array.isArray(data)) {
        permissions.value = data.map((p: any) => p.code || p).filter(Boolean)
      }
      
      loaded.value = true
    } catch (e) {
      console.error('Failed to load permissions:', e)
      // Fallback to level-based check
      loaded.value = true
    }
  }

  function can(permission: string): boolean {
    // If we have actual permissions loaded, use them
    if (permissions.value.length > 0) {
      return permissions.value.includes(permission)
    }
    // Fallback to level-based check
    const level = userLevel.value || 15
    return checkPermissionByLevel(permission, level)
  }

  function checkPermissionByLevel(permission: string, level: number): boolean {
    if (level >= 20) return true
    
    const readPerms = ['view', 'view_all', 'view_project']
    const permAction = permission.split(':')[1] || ''
    
    if (level <= 5) {
      return readPerms.some(rp => permAction.includes(rp))
    }
    
    const adminOnly = ['workspace:manage', 'workspace:delete', 'member:manage', 
      'settings:manage', 'project:delete', 'agent:manage', 'automation:manage',
      'role:manage', 'issue:delete', 'cycle:delete', 'module:delete',
      'page:delete', 'comment:delete', 'attachment:delete', 'time_track:delete']
    
    if (level >= 15 && !adminOnly.includes(permission)) {
      return true
    }
    
    return false
  }

  return { permissions, loaded, can, loadPermissions }
}
