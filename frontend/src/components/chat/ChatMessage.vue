<template>
  <div
    :class="[
      'group relative flex gap-2 px-3 py-1.5 rounded-lg',
      isSelf ? 'flex-row-reverse' : 'flex-row',
      message.deleted_at ? 'opacity-60' : '',
    ]"
  >
    <!-- Avatar -->
    <div
      :class="[
        'flex-shrink-0 w-7 h-7 rounded-full flex items-center justify-center text-sm',
        message.sender_type === 'agent' ? 'bg-indigo-100 text-indigo-600' : 'bg-gray-200 text-gray-600',
      ]"
    >
      {{ message.sender_type === 'agent' ? '🤖' : '👤' }}
    </div>

    <!-- Bubble -->
    <div :class="['max-w-[75%]', isSelf ? 'items-end' : 'items-start']">
      <div
        :class="[
          'inline-block px-3 py-1.5 rounded-2xl text-sm break-words',
          message.deleted_at
            ? 'italic text-gray-400 bg-gray-50'
            : isSelf
              ? 'bg-indigo-500 text-white'
              : message.sender_type === 'agent'
                ? 'bg-gray-100 text-gray-800'
                : 'bg-gray-100 text-gray-800',
        ]"
      >
        <span v-if="message.deleted_at">{{ t('chat.deletedPlaceholder') }}</span>
        <div v-else>
          <span v-if="renderedHtml" v-html="renderedHtml"></span>
          <span v-else>{{ message.content }}</span>
        </div>
      </div>

      <!-- Edited + timestamp -->
      <div :class="['flex gap-1 mt-0.5 text-[10px] text-gray-400', isSelf ? 'justify-end' : 'justify-start']">
        <span v-if="message.edited_at">{{ t('chat.editedMarker') }}</span>
        <span>{{ formatTime(message.created_at) }}</span>
      </div>

      <!-- Reactions -->
      <MessageReactions
        v-if="!message.deleted_at"
        :message="message"
        :current-user-id="currentUserId"
        @refresh="$emit('refresh')"
      />
    </div>

    <!-- Hover actions -->
    <MessageActions
      v-if="!message.deleted_at"
      :message="message"
      :current-user-id="currentUserId"
      @edit="$emit('edit')"
      @delete="$emit('delete')"
      @reply="$emit('reply')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { renderMarkdown } from '@/composables/useMarkdown'
import type { ChatMessage } from '@/types/chat'
import MessageReactions from './MessageReactions.vue'
import MessageActions from './MessageActions.vue'

const props = defineProps<{
  message: ChatMessage
  currentUserId: number
}>()
defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
  (e: 'reply'): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()

const isSelf = computed(() =>
  props.message.sender_type === 'user' && props.message.sender_id === props.currentUserId
)

const renderedHtml = computed(() => {
  if (!props.message.content) return ''
  return renderMarkdown(props.message.content)
})

function formatTime(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch {
    return ''
  }
}
</script>
