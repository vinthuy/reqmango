import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'

const { listActivity, sseHandlers } = vi.hoisted(() => ({
  listActivity: vi.fn(),
  sseHandlers: new Set<(event: string, data: any) => void>(),
}))

vi.mock('@/api/agent', () => ({
  agentApi: { listWorkspaceActivity: (...args: any[]) => listActivity(...args) },
}))
vi.mock('@/composables/useSSE', () => ({
  useSSE: () => ({
    onEvent: (fn: (event: string, data: any) => void) => {
      sseHandlers.add(fn)
      return () => sseHandlers.delete(fn)
    },
  }),
}))

import { useAutoTriageWatch } from './useAutoTriageWatch'
import { markAutoTriagePending, isAutoTriagePending } from '@/utils/autoTriagePref'

function push(event: string, data: any) {
  for (const fn of [...sseHandlers]) fn(event, data)
}

function mountWatch(issueId: number) {
  const onRefreshComments = vi.fn()
  const onDone = vi.fn()
  let api!: ReturnType<typeof useAutoTriageWatch>
  const wrapper = mount(defineComponent({
    setup() {
      api = useAutoTriageWatch({ workspaceId: ref(1), issueId: ref(issueId), onRefreshComments, onDone })
      return () => h('div')
    },
  }))
  return { wrapper, api: () => api, onRefreshComments, onDone }
}

describe('useAutoTriageWatch', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.stubGlobal('EventSource', class {})
    listActivity.mockReset().mockResolvedValue([])
    sseHandlers.clear()
    sessionStorage.clear()
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('finishes as soon as the auto_triage activity is pushed for this issue', async () => {
    markAutoTriagePending(42)
    const { api, onDone, onRefreshComments } = mountWatch(42)
    await flushPromises()
    expect(api().pending.value).toBe(true)

    push('agent_activity.created', { issue_id: 7, action: 'auto_triage' })
    push('agent_activity.created', { issue_id: 42, action: 'dispatch' })
    expect(api().pending.value).toBe(true)

    push('agent_activity.created', { issue_id: 42, action: 'auto_triage' })
    expect(api().pending.value).toBe(false)
    expect(onDone).toHaveBeenCalledWith(true)
    expect(onRefreshComments).toHaveBeenCalled()
    expect(isAutoTriagePending(42)).toBe(false)
    expect(sseHandlers.size).toBe(0)
  })

  it('still finishes through the fallback poll when no push arrives', async () => {
    markAutoTriagePending(42)
    const { api, onDone } = mountWatch(42)
    await flushPromises()
    expect(listActivity).toHaveBeenCalledTimes(1)

    listActivity.mockResolvedValue([{ id: 1 }])
    await vi.advanceTimersByTimeAsync(15000)

    expect(api().pending.value).toBe(false)
    expect(onDone).toHaveBeenCalledWith(true)
  })

  it('unsubscribes on unmount', async () => {
    markAutoTriagePending(42)
    const { wrapper } = mountWatch(42)
    await flushPromises()
    expect(sseHandlers.size).toBe(1)

    wrapper.unmount()
    expect(sseHandlers.size).toBe(0)
  })
})
