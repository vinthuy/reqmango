import { ref, onUnmounted } from 'vue'
import type { ChatMessage } from '@/types/chat'

export interface AgentTypingPayload {
  chat_id: number
  agent_id: number
  agent_name?: string
}

interface ReactionPayload {
  message_id: number
  user_id: number
  emoji: string
}

interface DeletedPayload {
  id: number
  deleted_at: string
}

interface EditedPayload extends ChatMessage {}

/**
 * Subscribe to a single chat's SSE stream. Returns reactive event refs.
 * Auto-reconnects on error with a 3s backoff. Cleans up on unmount.
 */
export function useChatSSE(chatId: number) {
  const newMessages = ref<ChatMessage[]>([])
  const editedMessages = ref<EditedPayload[]>([])
  const deletedMessages = ref<DeletedPayload[]>([])
  const reactionsAdded = ref<ReactionPayload[]>([])
  const reactionsRemoved = ref<ReactionPayload[]>([])
  const agentTyping = ref<AgentTypingPayload | null>(null)
  const connected = ref(false)

  let es: EventSource | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let typingTimer: ReturnType<typeof setTimeout> | null = null

  function connect() {
    const token = localStorage.getItem('token') || ''
    const url = `/api/v1/chats/${chatId}/stream?token=${encodeURIComponent(token)}`
    es = new EventSource(url)

    es.addEventListener('connected', () => {
      connected.value = true
    })

    es.addEventListener('message_new', (e: MessageEvent) => {
      try {
        const msg = JSON.parse(e.data) as ChatMessage
        newMessages.value.push(msg)
        agentTyping.value = null // a new message clears the typing indicator
      } catch (err) {
        console.error('[useChatSSE] message_new parse error:', err)
      }
    })

    es.addEventListener('message_edited', (e: MessageEvent) => {
      try {
        editedMessages.value.push(JSON.parse(e.data) as EditedPayload)
      } catch (err) {
        console.error('[useChatSSE] message_edited parse error:', err)
      }
    })

    es.addEventListener('message_deleted', (e: MessageEvent) => {
      try {
        deletedMessages.value.push(JSON.parse(e.data) as DeletedPayload)
      } catch (err) {
        console.error('[useChatSSE] message_deleted parse error:', err)
      }
    })

    es.addEventListener('reaction_added', (e: MessageEvent) => {
      try {
        reactionsAdded.value.push(JSON.parse(e.data) as ReactionPayload)
      } catch (err) {
        console.error('[useChatSSE] reaction_added parse error:', err)
      }
    })

    es.addEventListener('reaction_removed', (e: MessageEvent) => {
      try {
        reactionsRemoved.value.push(JSON.parse(e.data) as ReactionPayload)
      } catch (err) {
        console.error('[useChatSSE] reaction_removed parse error:', err)
      }
    })

    es.addEventListener('agent_typing', (e: MessageEvent) => {
      try {
        agentTyping.value = JSON.parse(e.data) as AgentTypingPayload
        // Clear typing indicator after 10s if no message arrives
        if (typingTimer) clearTimeout(typingTimer)
        typingTimer = setTimeout(() => {
          agentTyping.value = null
        }, 10000)
      } catch (err) {
        console.error('[useChatSSE] agent_typing parse error:', err)
      }
    })

    es.onerror = () => {
      connected.value = false
      es?.close()
      es = null
      // Auto-reconnect with 3s backoff
      if (reconnectTimer) clearTimeout(reconnectTimer)
      reconnectTimer = setTimeout(connect, 3000)
    }
  }

  function disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    if (typingTimer) clearTimeout(typingTimer)
    es?.close()
    es = null
    connected.value = false
  }

  connect()

  onUnmounted(disconnect)

  return {
    connected,
    newMessages,
    editedMessages,
    deletedMessages,
    reactionsAdded,
    reactionsRemoved,
    agentTyping,
    disconnect,
  }
}
