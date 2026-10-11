import { ref, watch, onBeforeUnmount, type Ref } from 'vue'
import { agentApi } from '@/api/agent'
import { clearAutoTriagePending, isAutoTriagePending } from '@/utils/autoTriagePref'

const POLL_MS = 4000
const MAX_MS = 3 * 60 * 1000

/**
 * After create-time auto-triage, poll workspace activity until an
 * auto_triage row appears (or timeout), then refresh comments so
 * CommentSuggestions can show one-click apply.
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

  function stop() {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  async function tick() {
    const issueId = opts.issueId.value
    const workspaceId = opts.workspaceId.value
    if (!issueId || !workspaceId || !pending.value) return

    if (Date.now() - startedAt > MAX_MS) {
      pending.value = false
      stop()
      clearAutoTriagePending(issueId)
      opts.onDone?.(false)
      return
    }

    try {
      const list = await agentApi.listWorkspaceActivity(workspaceId, {
        issue_id: issueId,
        action: 'auto_triage',
        limit: 1,
      })
      if (list?.length) {
        pending.value = false
        stop()
        clearAutoTriagePending(issueId)
        opts.onRefreshComments()
        opts.onActivity?.()
        opts.onDone?.(true)
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
    void tick()
    timer = setInterval(tick, POLL_MS)
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
