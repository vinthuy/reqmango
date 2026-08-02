<template>
  <div
    ref="container"
    class="flex-1 overflow-y-auto py-3 space-y-1"
    @scroll="onScroll"
  >
    <!-- Load older button -->
    <div v-if="hasMore" class="text-center py-2">
      <button
        @click="$emit('load-older')"
        :disabled="loadingOlder"
        class="text-xs text-indigo-600 hover:underline disabled:opacity-50"
      >{{ loadingOlder ? t('chat.loading') : t('chat.loadOlder') }}</button>
    </div>

    <ChatMessage
      v-for="msg in messages"
      :key="msg.id"
      :message="msg"
      :current-user-id="currentUserId"
      @edit="$emit('edit', msg)"
      @delete="$emit('delete', msg)"
      @reply="$emit('reply', msg)"
      @refresh="$emit('refresh')"
    />

    <!-- Agent typing indicator -->
    <div v-if="agentTyping" class="flex gap-2 px-3 py-1.5">
      <div class="flex-shrink-0 w-7 h-7 rounded-full bg-indigo-100 text-indigo-600 flex items-center justify-center text-sm">🤖</div>
      <div class="bg-gray-100 rounded-2xl px-3 py-2 text-sm text-gray-500 flex items-center gap-1">
        <span class="animate-bounce">·</span>
        <span class="animate-bounce" style="animation-delay:0.1s">·</span>
        <span class="animate-bounce" style="animation-delay:0.2s">·</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ChatMessage as ChatMessageType } from '@/types/chat'
import ChatMessage from './ChatMessage.vue'

const props = defineProps<{
  messages: ChatMessageType[]
  currentUserId: number
  hasMore: boolean
  loadingOlder: boolean
  agentTyping: { agent_id: number } | null
}>()
defineEmits<{
  (e: 'load-older'): void
  (e: 'edit', msg: ChatMessageType): void
  (e: 'delete', msg: ChatMessageType): void
  (e: 'reply', msg: ChatMessageType): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()
const container = ref<HTMLDivElement | null>(null)
let wasAtBottom = true

function isAtBottom(): boolean {
  const el = container.value
  if (!el) return true
  return el.scrollHeight - el.scrollTop - el.clientHeight < 50
}

function onScroll() {
  wasAtBottom = isAtBottom()
}

// Auto-scroll to bottom when new messages arrive (only if user was at bottom)
watch(
  () => props.messages.length,
  async () => {
    if (wasAtBottom) {
      await nextTick()
      const el = container.value
      if (el) el.scrollTop = el.scrollHeight
    }
  }
)

// Initial scroll to bottom
watch(
  () => props.messages,
  async () => {
    await nextTick()
    const el = container.value
    if (el) el.scrollTop = el.scrollHeight
    wasAtBottom = true
  },
  { immediate: true }
)
</script>
