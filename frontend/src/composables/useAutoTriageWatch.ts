import { ref, watch, onBeforeUnmount, type Ref } from 'vue'
import { agentApi } from '@/api/agent'
import { useSSE } from '@/composables/useSSE'
import { clearAutoTriagePending, isAutoTriagePending } from '@/utils/autoTriagePref'

/** Fallback poll; the SSE push normally ends the wait first. */
const POLL_MS_WITH_SSE = 15000
const POLL_MS = 4000
const MAX_MS = 3 * 60 * 1000

/**
 * After create-time auto-triage, wait for the auto_triage activity on the
 * issue — pushed over SSE, with a slower poll as a fallback — then refresh
 * comments so CommentSuggestions can show one-click apply.
 */
export function useAutoTriageWatch(opts: {
  workspaceId: Ref<number>
  issueId: Ref<number | null | undefined>
  onRefreshComments: () => void
  onActivity?: () => void
  onDone?: (ok: boolean) => void
}) {
  const pending = ref(false)
  let timer: ReturnType<typeof setInterval> | null = null
  let startedAt = 0

  const sse = typeof EventSource !== 'undefined' ? useSSE() : null
  let unsubscribe: (() => void) | null = null

  function stop() {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
    unsubscribe?.()
    unsubscribe = null
  }

  function finish(issueId: number, ok: boolean) {
    pending.value = false
    stop()
    clearAutoTriagePending(issueId)
    if (ok) {
      opts.onRefreshComments()
      opts.onActivity?.()
    }
    opts.onDone?.(ok)
  }

  async function tick() {
    const issueId = opts.issueId.value
    const workspaceId = opts.workspaceId.value
    if (!issueId || !workspaceId || !pending.value) return

    if (Date.now() - startedAt > MAX_MS) {
      finish(issueId, false)
      return
    }

    try {
      const list = await agentApi.listWorkspaceActivity(workspaceId, {
        issue_id: issueId,
        action: 'auto_triage',
        limit: 1,
      })
      if (!pending.value) return
      if (list?.length) {
        finish(issueId, true)
      } else {
        // Comment may land slightly before activity is visible; keep refreshing.
        opts.onRefreshComments()
      }
    } catch {
      /* keep polling */
    }
  }

  function start(issueId?: number | null) {
    const id = issueId ?? opts.issueId.value
    if (!id || !opts.workspaceId.value) return
    if (!isAutoTriagePending(id) && !pending.value) return
    pending.value = true
    startedAt = Date.now()
    stop()
    if (sse) {
      unsubscribe = sse.onEvent((event, data) => {
        if (event !== 'agent_activity.created' || !pending.value) return
        if (Number(data?.issue_id) === id && data?.action === 'auto_triage') {
          finish(id, true)
        }
      })
    }
    void tick()
    timer = setInterval(tick, sse ? POLL_MS_WITH_SSE : POLL_MS)
  }

  function maybeStartFromFlag() {
    const id = opts.issueId.value
    if (id && isAutoTriagePending(id)) start(id)
  }

  watch(
    () => [opts.issueId.value, opts.workspaceId.value] as const,
    () => maybeStartFromFlag(),
    { immediate: true },
  )

  onBeforeUnmount(stop)

  return { pending, start, stop, maybeStartFromFlag }
}
