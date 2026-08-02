<template>
  <div class="flex flex-col h-[600px] border border-gray-200 rounded-lg overflow-hidden bg-white">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-gray-50">
      <h3 class="text-sm font-medium text-gray-700">{{ t('chat.title') }}</h3>
      <span :class="connected ? 'text-green-500' : 'text-gray-400'" class="text-xs">●</span>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="flex-1 flex items-center justify-center text-gray-400 text-sm">
      {{ t('chat.loading') }}
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="flex-1 flex items-center justify-center text-red-500 text-sm">
      {{ error }}
    </div>

    <!-- Message list -->
    <ChatMessageList
      v-else
      :messages="messages"
      :current-user-id="currentUserId"
      :has-more="hasMore"
      :loading-older="loadingOlder"
      :agent-typing="agentTyping"
      @load-older="loadOlder"
      @edit="onEdit"
      @delete="onDelete"
      @reply="onReply"
      @refresh="refresh"
    />

    <!-- Input -->
    <ChatInput
      :sending="sending"
      :reply-to="replyTo"
      :mention-candidates="mentionCandidates"
      @send="onSend"
      @cancel-reply="replyTo = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, type WatchStopHandle } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ChatMessage } from '@/types/chat'
import * as chatApi from '@/api/chat'
import { useChatSSE } from '@/composables/useChatSSE'
import ChatMessageList from './ChatMessageList.vue'
import ChatInput from './ChatInput.vue'

const props = defineProps<{
  issueId: number
  workspaceId: number
  currentUserId: number
  mentionCandidates: { id: number; name: string; type: 'user' | 'agent' }[]
}>()

const { t } = useI18n()

const loading = ref(true)
const error = ref('')
const sending = ref(false)
const chatId = ref(0)
const messages = ref<ChatMessage[]>([])
const hasMore = ref(false)
const loadingOlder = ref(false)
const nextCursor = ref('')
const replyTo = ref<ChatMessage | null>(null)
// @ts-expect-error TS6133 — editingId retained as plan-accepted dead code (future inline editing)
const editingId = ref<number | null>(null)

const connected = ref(false)
const agentTyping = ref<{ agent_id: number } | null>(null)

let sse: ReturnType<typeof useChatSSE> | null = null
let isUnmounted = false
const watcherStops: WatchStopHandle[] = []
let reactionRefreshTimer: ReturnType<typeof setTimeout> | null = null

// Registered SYNCHRONOUSLY in setup scope so it fires on unmount even though
// `sse` and the watchers are created later (after `await` in loadChat).
onUnmounted(() => {
  isUnmounted = true
  for (const stop of watcherStops) stop()
  watcherStops.length = 0
  if (reactionRefreshTimer != null) {
    clearTimeout(reactionRefreshTimer)
    reactionRefreshTimer = null
  }
  sse?.disconnect()
  sse = null
})

onMounted(async () => {
  await loadChat()
})

async function loadChat() {
  loading.value = true
  error.value = ''
  try {
    const chat = await chatApi.getIssueChat(props.issueId)
    // If the component unmounted while waiting for the chat, bail out —
    // creating an SSE connection now would leak (no live component to clean up).
    if (isUnmounted) return
    chatId.value = chat.id
    messages.value = chat.messages || []
    loading.value = false
    // Open SSE now that we have a chatId. useChatSSE's internal onUnmounted
    // won't register (we're after an await), so we manage cleanup ourselves
    // via the onUnmounted hook above + watcherStops array.
    sse = useChatSSE(chat.id)
    connected.value = sse.connected.value
    agentTyping.value = sse.agentTyping.value
    watchSSE()
  } catch (err: any) {
    if (isUnmounted) return
    error.value = err?.response?.data?.message || t('chat.loadFailed')
    loading.value = false
  }
}

function watchSSE() {
  if (!sse) return
  // Cursor-based processing: useChatSSE accumulates events into arrays and
  // never clears them, so we track the last-processed index to avoid
  // re-processing (and duplicating) older events on each watcher fire.
  let newCursor = 0
  let editedCursor = 0
  let deletedCursor = 0

  watcherStops.push(
    watch(sse.newMessages, (arr) => {
      for (let i = newCursor; i < arr.length; i++) {
        const m = arr[i]
        // Dedup against messages already in the list (e.g. optimistic push
        // from onSend, or a message delivered via SSE after a manual send).
        if (!messages.value.find((x) => x.id === m.id)) {
          messages.value.push(m)
        }
      }
      newCursor = arr.length
    }, { deep: true })
  )

  watcherStops.push(
    watch(sse.editedMessages, (arr) => {
      for (let i = editedCursor; i < arr.length; i++) {
        const e = arr[i]
        const idx = messages.value.findIndex((m) => m.id === e.id)
        if (idx >= 0) messages.value[idx] = { ...messages.value[idx], ...e }
      }
      editedCursor = arr.length
    }, { deep: true })
  )

  watcherStops.push(
    watch(sse.deletedMessages, (arr) => {
      for (let i = deletedCursor; i < arr.length; i++) {
        const d = arr[i]
        const idx = messages.value.findIndex((m) => m.id === d.id)
        if (idx >= 0) messages.value[idx] = { ...messages.value[idx], deleted_at: d.deleted_at, content: '' }
      }
      deletedCursor = arr.length
    }, { deep: true })
  )

  // Live connection indicator (the initial snapshot in loadChat only covers
  // the moment of SSE creation; this keeps the dot in sync afterwards).
  watcherStops.push(
    watch(sse.connected, (v) => { connected.value = v })
  )

  // Debounce reaction events: the arrays accumulate, so without debouncing
  // each push would trigger a separate refresh() call.
  watcherStops.push(
    watch(sse.reactionsAdded, scheduleReactionRefresh, { deep: true })
  )
  watcherStops.push(
    watch(sse.reactionsRemoved, scheduleReactionRefresh, { deep: true })
  )

  watcherStops.push(
    watch(sse.agentTyping, (v) => { agentTyping.value = v })
  )
}

function scheduleReactionRefresh() {
  if (reactionRefreshTimer != null) clearTimeout(reactionRefreshTimer)
  reactionRefreshTimer = setTimeout(() => {
    refresh()
    reactionRefreshTimer = null
  }, 50)
}

async function refresh() {
  if (!chatId.value) return
  try {
    const resp = await chatApi.listMessages(chatId.value, '', 50)
    messages.value = resp.messages
  } catch (err) {
    console.error('[ChatPanel] refresh failed:', err)
  }
}

async function loadOlder() {
  if (!chatId.value || loadingOlder.value || !hasMore.value) return
  loadingOlder.value = true
  try {
    const resp = await chatApi.listMessages(chatId.value, nextCursor.value, 20)
    // Prepend older messages
    messages.value = [...resp.messages, ...messages.value]
    nextCursor.value = resp.next_cursor
    hasMore.value = !!resp.next_cursor
  } catch (err) {
    console.error('[ChatPanel] loadOlder failed:', err)
  } finally {
    loadingOlder.value = false
  }
}

async function onSend(content: string) {
  if (!chatId.value) return
  sending.value = true
  error.value = '' // clear stale error so the list isn't hidden on success
  try {
    const msg = await chatApi.sendMessage(chatId.value, {
      content,
      reply_to_id: replyTo.value?.id || null,
    })
    // The SSE event will normally deliver this; push optimistically if not dup
    if (!messages.value.find((m) => m.id === msg.id)) {
      messages.value.push(msg)
    }
    replyTo.value = null
  } catch (err: any) {
    error.value = err?.response?.data?.message || t('chat.sendFailed')
  } finally {
    sending.value = false
  }
}

async function onEdit(msg: ChatMessage) {
  // Simple prompt-based editor (v1); can be upgraded to inline editor later
  const newContent = window.prompt(t('chat.editPrompt'), msg.content)
  if (newContent == null || newContent === msg.content) return
  error.value = '' // clear stale error
  try {
    const updated = await chatApi.editMessage(msg.id, newContent)
    const idx = messages.value.findIndex((m) => m.id === updated.id)
    if (idx >= 0) messages.value[idx] = updated
  } catch (err: any) {
    error.value = err?.response?.data?.message || t('chat.editFailed')
  }
}

async function onDelete(msg: ChatMessage) {
  if (!window.confirm(t('chat.deleteConfirm'))) return
  error.value = '' // clear stale error
  try {
    await chatApi.deleteMessage(msg.id)
    // SSE will update; optimistic fallback:
    const idx = messages.value.findIndex((m) => m.id === msg.id)
    if (idx >= 0) {
      messages.value[idx] = { ...messages.value[idx], deleted_at: new Date().toISOString(), content: '' }
    }
  } catch (err: any) {
    error.value = err?.response?.data?.message || t('chat.deleteFailed')
  }
}

function onReply(msg: ChatMessage) {
  replyTo.value = msg
}
</script>
