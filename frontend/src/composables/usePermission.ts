import { computed, ref, toValue, watch, type MaybeRefOrGetter, type Ref } from 'vue'
import api from '@/api'

interface Permissions {
  permissions: string[]
  role_level: number
}

// One entry per project or workspace, shared by every component on the page.
// The cache is dropped when the signed-in token changes so another user never
// sees it.
const cache = new Map<string, Ref<Permissions | null>>()
const inflight = new Map<string, Promise<void>>()
let cacheToken: string | null = null

function entry(key: string): Ref<Permissions | null> {
  const token = localStorage.getItem('token')
  if (token !== cacheToken) {
    cache.clear()
    inflight.clear()
    cacheToken = token
  }
  let e = cache.get(key)
  if (!e) {
    e = ref(null)
    cache.set(key, e)
  }
  return e
}

function load(key: string, url: string): Promise<void> {
  const e = entry(key)
  if (e.value) return Promise.resolve()
  let p = inflight.get(key)
  if (!p) {
    p = api
      .get(url)
      .then((resp) => {
        e.value = {
          permissions: resp.data?.permissions ?? [],
          role_level: resp.data?.role_level ?? 0,
        }
      })
      .catch(() => {
        // Leave it unloaded so the next mount retries; actions stay hidden.
      })
      .finally(() => inflight.delete(key))
    inflight.set(key, p)
  }
  return p
}

function usePermissions(
  id: MaybeRefOrGetter<number | string | null | undefined>,
  scope: 'projects' | 'workspaces',
) {
  const key = (v: number | string) => `${scope}:${v}`
  const state = computed(() => {
    const v = toValue(id)
    return v ? entry(key(v)).value : null
  })
  watch(
    () => toValue(id),
    (v) => {
      if (v) load(key(v), `/${scope}/${v}/my-permissions`)
    },
    { immediate: true },
  )
  return {
    loaded: computed(() => state.value !== null),
    roleLevel: computed(() => state.value?.role_level ?? 0),
    can: (code: string) => !!state.value?.permissions.includes(code),
  }
}

/**
 * The caller's permission codes in a project, as the API enforces them.
 * `can` is false until loaded, so guarded actions never flash in and out.
 */
export function useProjectPermissions(projectId: MaybeRefOrGetter<number | null | undefined>) {
  return usePermissions(projectId, 'projects')
}

/** Workspace-level counterpart; accepts a numeric ID or slug. */
export function useWorkspacePermissions(workspace: MaybeRefOrGetter<number | string | null | undefined>) {
  return usePermissions(workspace, 'workspaces')
}

export function resetProjectPermissionsCache() {
  cache.clear()
  inflight.clear()
  cacheToken = null
}
