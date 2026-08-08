import { ref, onUnmounted } from 'vue'

const connections: Record<string, EventSource> = {}
const listeners: Record<string, Set<(event: string, data: any) => void>> = {}

function getSSEUrl(): string {
  const token = localStorage.getItem('token')
  return `/api/v1/sse?token=${encodeURIComponent(token || '')}`
}

function ensureConnection(key: string): EventSource {
  if (connections[key]) return connections[key]

  const url = getSSEUrl()
  const es = new EventSource(url, { withCredentials: true })

  es.onopen = () => {
    console.debug('[SSE] Connected:', key)
  }

  es.onerror = (e) => {
    console.debug('[SSE] Error:', key, e)
  }

  es.onmessage = () => {
    // Generic message handler - individual listeners handle specific events
  }

  // Register event types
  es.addEventListener('agent_task.created', (e: MessageEvent) => {
    dispatch(key, 'agent_task.created', parseData(e.data))
  })
  es.addEventListener('agent_task.updated', (e: MessageEvent) => {
    dispatch(key, 'agent_task.updated', parseData(e.data))
  })
  es.addEventListener('agent_task.started', (e: MessageEvent) => {
    dispatch(key, 'agent_task.started', parseData(e.data))
  })
  es.addEventListener('agent_task.completed', (e: MessageEvent) => {
    dispatch(key, 'agent_task.completed', parseData(e.data))
  })
  es.addEventListener('agent_task.failed', (e: MessageEvent) => {
    dispatch(key, 'agent_task.failed', parseData(e.data))
  })
  es.addEventListener('agent_task.cancelled', (e: MessageEvent) => {
    dispatch(key, 'agent_task.cancelled', parseData(e.data))
  })
  // Developer Agent job lifecycle (PRD P4-001)
  es.addEventListener('developer_job.created', (e: MessageEvent) => {
    dispatch(key, 'developer_job.created', parseData(e.data))
  })
  es.addEventListener('developer_job.updated', (e: MessageEvent) => {
    dispatch(key, 'developer_job.updated', parseData(e.data))
  })
  es.addEventListener('developer_job.completed', (e: MessageEvent) => {
    dispatch(key, 'developer_job.completed', parseData(e.data))
  })
  es.addEventListener('developer_job.failed', (e: MessageEvent) => {
    dispatch(key, 'developer_job.failed', parseData(e.data))
  })
  es.addEventListener('developer_job.cancelled', (e: MessageEvent) => {
    dispatch(key, 'developer_job.cancelled', parseData(e.data))
  })
  // Tester Agent job lifecycle (PRD P4-002)
  es.addEventListener('tester_job.created', (e: MessageEvent) => {
    dispatch(key, 'tester_job.created', parseData(e.data))
  })
  es.addEventListener('tester_job.updated', (e: MessageEvent) => {
    dispatch(key, 'tester_job.updated', parseData(e.data))
  })
  es.addEventListener('tester_job.completed', (e: MessageEvent) => {
    dispatch(key, 'tester_job.completed', parseData(e.data))
  })
  es.addEventListener('tester_job.failed', (e: MessageEvent) => {
    dispatch(key, 'tester_job.failed', parseData(e.data))
  })
  es.addEventListener('tester_job.cancelled', (e: MessageEvent) => {
    dispatch(key, 'tester_job.cancelled', parseData(e.data))
  })
  // CI/CD build lifecycle (PRD P4-004)
  es.addEventListener('cicd_config.created', (e: MessageEvent) => {
    dispatch(key, 'cicd_config.created', parseData(e.data))
  })
  es.addEventListener('cicd_config.updated', (e: MessageEvent) => {
    dispatch(key, 'cicd_config.updated', parseData(e.data))
  })
  es.addEventListener('cicd_build.created', (e: MessageEvent) => {
    dispatch(key, 'cicd_build.created', parseData(e.data))
  })
  es.addEventListener('cicd_build.updated', (e: MessageEvent) => {
    dispatch(key, 'cicd_build.updated', parseData(e.data))
  })
  es.addEventListener('cicd_build.completed', (e: MessageEvent) => {
    dispatch(key, 'cicd_build.completed', parseData(e.data))
  })
  es.addEventListener('cicd_build.failed', (e: MessageEvent) => {
    dispatch(key, 'cicd_build.failed', parseData(e.data))
  })
  es.addEventListener('cicd_build.cancelled', (e: MessageEvent) => {
    dispatch(key, 'cicd_build.cancelled', parseData(e.data))
  })
  // SDLC workflow & stage lifecycle (PRD P4-006)
  es.addEventListener('sdlc_workflow.created', (e: MessageEvent) => {
    dispatch(key, 'sdlc_workflow.created', parseData(e.data))
  })
  es.addEventListener('sdlc_workflow.updated', (e: MessageEvent) => {
    dispatch(key, 'sdlc_workflow.updated', parseData(e.data))
  })
  es.addEventListener('sdlc_workflow.completed', (e: MessageEvent) => {
    dispatch(key, 'sdlc_workflow.completed', parseData(e.data))
  })
  es.addEventListener('sdlc_workflow.failed', (e: MessageEvent) => {
    dispatch(key, 'sdlc_workflow.failed', parseData(e.data))
  })
  es.addEventListener('sdlc_workflow.cancelled', (e: MessageEvent) => {
    dispatch(key, 'sdlc_workflow.cancelled', parseData(e.data))
  })
  es.addEventListener('sdlc_workflow.resumed', (e: MessageEvent) => {
    dispatch(key, 'sdlc_workflow.resumed', parseData(e.data))
  })
  es.addEventListener('sdlc_stage.updated', (e: MessageEvent) => {
    dispatch(key, 'sdlc_stage.updated', parseData(e.data))
  })
  es.addEventListener('sdlc_stage.completed', (e: MessageEvent) => {
    dispatch(key, 'sdlc_stage.completed', parseData(e.data))
  })
  es.addEventListener('sdlc_stage.failed', (e: MessageEvent) => {
    dispatch(key, 'sdlc_stage.failed', parseData(e.data))
  })
  // Autopilot task & execution lifecycle (PRD P4-008)
  es.addEventListener('autopilot_task.created', (e: MessageEvent) => {
    dispatch(key, 'autopilot_task.created', parseData(e.data))
  })
  es.addEventListener('autopilot_task.updated', (e: MessageEvent) => {
    dispatch(key, 'autopilot_task.updated', parseData(e.data))
  })
  es.addEventListener('autopilot_task.deleted', (e: MessageEvent) => {
    dispatch(key, 'autopilot_task.deleted', parseData(e.data))
  })
  es.addEventListener('autopilot_task.toggled', (e: MessageEvent) => {
    dispatch(key, 'autopilot_task.toggled', parseData(e.data))
  })
  es.addEventListener('autopilot_execution.started', (e: MessageEvent) => {
    dispatch(key, 'autopilot_execution.started', parseData(e.data))
  })
  es.addEventListener('autopilot_execution.completed', (e: MessageEvent) => {
    dispatch(key, 'autopilot_execution.completed', parseData(e.data))
  })
  es.addEventListener('autopilot_execution.failed', (e: MessageEvent) => {
    dispatch(key, 'autopilot_execution.failed', parseData(e.data))
  })
  // Tool calling lifecycle (T7 Tool Calling Hardening)
  es.addEventListener('tool_call.completed', (e: MessageEvent) => {
    dispatch(key, 'tool_call.completed', parseData(e.data))
  })
  es.addEventListener('tool_call.failed', (e: MessageEvent) => {
    dispatch(key, 'tool_call.failed', parseData(e.data))
  })
  es.addEventListener('tool_call.rate_limited', (e: MessageEvent) => {
    dispatch(key, 'tool_call.rate_limited', parseData(e.data))
  })
  // Squad execution lifecycle (T6 Squads Multi-Agent)
  es.addEventListener('squad.execution.started', (e: MessageEvent) => {
    dispatch(key, 'squad.execution.started', parseData(e.data))
  })
  es.addEventListener('squad.execution.updated', (e: MessageEvent) => {
    dispatch(key, 'squad.execution.updated', parseData(e.data))
  })
  es.addEventListener('squad.execution.log', (e: MessageEvent) => {
    dispatch(key, 'squad.execution.log', parseData(e.data))
  })
  es.addEventListener('squad.execution.progress', (e: MessageEvent) => {
    dispatch(key, 'squad.execution.progress', parseData(e.data))
  })
  es.addEventListener('squad.execution.completed', (e: MessageEvent) => {
    dispatch(key, 'squad.execution.completed', parseData(e.data))
  })
  es.addEventListener('squad.execution.failed', (e: MessageEvent) => {
    dispatch(key, 'squad.execution.failed', parseData(e.data))
  })
  es.addEventListener('squad.execution.cancelled', (e: MessageEvent) => {
    dispatch(key, 'squad.execution.cancelled', parseData(e.data))
  })
  es.addEventListener('notification', (e: MessageEvent) => {
    dispatch(key, 'notification', parseData(e.data))
  })
  es.addEventListener('connected', (e: MessageEvent) => {
    dispatch(key, 'connected', parseData(e.data))
  })

  connections[key] = es
  listeners[key] = new Set()
  return es
}

function parseData(data: string): any {
  try {
    return JSON.parse(data)
  } catch {
    return data
  }
}

function dispatch(key: string, event: string, data: any) {
  const set = listeners[key]
  if (!set) return
  for (const fn of set) {
    try {
      fn(event, data)
    } catch (e) {
      console.error('[SSE] Listener error:', e)
    }
  }
}

export function useSSE() {
  const connected = ref(false)
  const key = 'default'

  function onEvent(handler: (event: string, data: any) => void) {
    const es = ensureConnection(key)
    if (es.readyState === EventSource.OPEN) {
      connected.value = true
    }
    listeners[key].add(handler)
    return () => listeners[key]?.delete(handler)
  }

  function onAgentTask(handler: (event: string, task: any) => void) {
    return onEvent((event, data) => {
      if (event.startsWith('agent_task.')) {
        handler(event, data)
      }
    })
  }

  function close() {
    const es = connections[key]
    if (es) {
      es.close()
      delete connections[key]
      delete listeners[key]
      connected.value = false
    }
  }

  onUnmounted(() => {
    // Don't close the connection globally on unmount - it may be shared
    // Individual listeners are cleaned up via their returned cleanup fn
  })

  return { connected, onEvent, onAgentTask, close }
}
