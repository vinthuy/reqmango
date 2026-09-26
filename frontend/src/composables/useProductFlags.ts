/**
 * Product-surface flags (local). Same pattern as Stage B Agent console demotion.
 * - rm_show_initiatives: show workspace 「战略目标」in nav (default off)
 * - rm_advanced_agents: show /agents console (default off; guarded in router)
 */
import { ref, onMounted, onUnmounted } from 'vue'

export const SHOW_INITIATIVES_KEY = 'rm_show_initiatives'
export const ADVANCED_AGENTS_KEY = 'rm_advanced_agents'
const FLAGS_EVENT = 'rm-product-flags'

export function isInitiativesEnabled(): boolean {
  try {
    return localStorage.getItem(SHOW_INITIATIVES_KEY) === '1'
  } catch {
    return false
  }
}

export function setInitiativesEnabled(on: boolean): void {
  try {
    if (on) localStorage.setItem(SHOW_INITIATIVES_KEY, '1')
    else localStorage.removeItem(SHOW_INITIATIVES_KEY)
  } catch { /* ignore */ }
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new Event(FLAGS_EVENT))
  }
}

export function isAdvancedAgentsEnabled(): boolean {
  try {
    return localStorage.getItem(ADVANCED_AGENTS_KEY) === '1'
  } catch {
    return false
  }
}

export function setAdvancedAgentsEnabled(on: boolean): void {
  try {
    if (on) localStorage.setItem(ADVANCED_AGENTS_KEY, '1')
    else localStorage.removeItem(ADVANCED_AGENTS_KEY)
  } catch { /* ignore */ }
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new Event(FLAGS_EVENT))
  }
}

export function useShowInitiatives() {
  const enabled = ref(isInitiativesEnabled())

  function refresh() {
    enabled.value = isInitiativesEnabled()
  }

  function setEnabled(on: boolean) {
    setInitiativesEnabled(on)
    enabled.value = on
  }

  onMounted(() => {
    refresh()
    window.addEventListener(FLAGS_EVENT, refresh)
    window.addEventListener('storage', refresh)
  })
  onUnmounted(() => {
    window.removeEventListener(FLAGS_EVENT, refresh)
    window.removeEventListener('storage', refresh)
  })

  return { enabled, setEnabled, refresh }
}
