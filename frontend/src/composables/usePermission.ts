import { ref } from 'vue'

export function usePermission() {
  const permissions = ref<string[]>([])
  const loaded = ref(false)

  async function loadPermissions(_workspaceId?: number, _projectId?: number) {
    loaded.value = true
  }

  function can(permission: string): boolean {
    const level = getCurrentRoleLevel()
    return checkPermissionByLevel(permission, level)
  }

  function getCurrentRoleLevel(): number {
    return 15
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
