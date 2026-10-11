import { computed, ref, toValue, watch, type MaybeRefOrGetter, type Ref } from 'vue'
import api from '@/api'

interface ProjectPermissions {
  permissions: string[]
  role_level: number
}

// One entry per project, shared by every component on the page. The cache is
// dropped when the signed-in token changes so another user never sees it.
const cache = new Map<number, Ref<ProjectPermissions | null>>()
const inflight = new Map<number, Promise<void>>()
let cacheToken: string | null = null

function entry(projectId: number): Ref<ProjectPermissions | null> {
  const token = localStorage.getItem('token')
  if (token !== cacheToken) {
    cache.clear()
    inflight.clear()
    cacheToken = token
  }
  let e = cache.get(projectId)
  if (!e) {
    e = ref(null)
    cache.set(projectId, e)
  }
  return e
}

function load(projectId: number): Promise<void> {
  const e = entry(projectId)
  if (e.value) return Promise.resolve()
  let p = inflight.get(projectId)
  if (!p) {
    p = api
      .get(`/projects/${projectId}/my-permissions`)
      .then((resp) => {
        e.value = {
          permissions: resp.data?.permissions ?? [],
          role_level: resp.data?.role_level ?? 0,
        }
      })
      .catch(() => {
        // Leave it unloaded so the next mount retries; actions stay hidden.
      })
      .finally(() => inflight.delete(projectId))
    inflight.set(projectId, p)
  }
  return p
}

/**
 * The caller's permission codes in a project, as the API enforces them.
 * `can` is false until loaded, so guarded actions never flash in and out.
 */
export function useProjectPermissions(projectId: MaybeRefOrGetter<number | null | undefined>) {
  const state = computed(() => {
    const id = toValue(projectId)
    return id ? entry(id).value : null
  })
  watch(
    () => toValue(projectId),
    (id) => {
      if (id) load(id)
    },
    { immediate: true },
  )
  return {
    loaded: computed(() => state.value !== null),
    roleLevel: computed(() => state.value?.role_level ?? 0),
    can: (code: string) => !!state.value?.permissions.includes(code),
  }
}

export function resetProjectPermissionsCache() {
  cache.clear()
  inflight.clear()
  cacheToken = null
}
