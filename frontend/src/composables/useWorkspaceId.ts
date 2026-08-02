import { useRoute } from 'vue-router'
import { workspaceApi } from '@/api/workspace'

/**
 * Resolves workspace ID from route params (wsParam or slug) or localStorage.
 * Works with both /workspaces/:wsParam and /workspace/:slug URL patterns.
 */
export function useWorkspaceId() {
  const route = useRoute()

  async function getWorkspaceId(): Promise<number | null> {
    // Pattern 1: /workspaces/:wsParam (numeric ID)
    const wsParam = route.params.wsParam
    if (wsParam) {
      const id = parseInt(wsParam as string, 10)
      if (!isNaN(id)) return id
    }

    // Pattern 2: /workspace/:slug (resolve via API)
    const slug = route.params.slug
    if (slug) {
      try {
        const ws = await workspaceApi.getBySlug(slug as string)
        if (ws?.id) {
          localStorage.setItem('currentWorkspaceId', String(ws.id))
          return ws.id
        }
      } catch { /* ignore */ }
    }

    // Pattern 3: localStorage fallback
    const stored = localStorage.getItem('currentWorkspaceId')
    if (stored) {
      const id = parseInt(stored, 10)
      if (!isNaN(id)) return id
    }

    return null
  }

  return { getWorkspaceId }
}
