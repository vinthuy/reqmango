<template>
  <div class="absolute -top-3 right-2 hidden group-hover:flex bg-white border border-gray-200 rounded-md shadow-sm text-xs">
    <button
      v-if="canEdit"
      @click="$emit('edit')"
      class="px-2 py-1 hover:bg-gray-100"
      :title="t('chat.action.edit')"
    >✏️</button>
    <button
      v-if="canDelete"
      @click="$emit('delete')"
      class="px-2 py-1 hover:bg-gray-100 text-red-500"
      :title="t('chat.action.delete')"
    >🗑️</button>
    <button
      @click="$emit('reply')"
      class="px-2 py-1 hover:bg-gray-100"
      :title="t('chat.action.reply')"
    >↩️</button>
    <button
      @click="copyContent"
      class="px-2 py-1 hover:bg-gray-100"
      :title="t('chat.action.copy')"
    >📋</button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ChatMessage } from '@/types/chat'

const props = defineProps<{
  message: ChatMessage
  currentUserId: number
}>()
defineEmits<{
  (e: 'edit'): void
  (e: 'delete'): void
  (e: 'reply'): void
}>()

const { t } = useI18n()

const canEdit = computed(() =>
  props.message.sender_type === 'user' &&
  props.message.sender_id === props.currentUserId &&
  !props.message.deleted_at &&
  // 30-minute edit window
  Date.now() - new Date(props.message.created_at).getTime() < 30 * 60 * 1000
)

const canDelete = computed(() =>
  !props.message.deleted_at &&
  (props.message.sender_id === props.currentUserId || props.message.sender_type === 'agent')
)

async function copyContent() {
  try {
    await navigator.clipboard.writeText(props.message.content)
  } catch (err) {
    console.error('[MessageActions] copy failed:', err)
  }
}
</script>
