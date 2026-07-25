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
