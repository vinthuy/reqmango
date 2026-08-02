<template>
  <div class="flex flex-wrap gap-1 mt-1">
    <button
      v-for="group in message.reactions"
      :key="group.emoji"
      @click="toggle(group.emoji)"
      :class="hasReacted(group.user_ids)
        ? 'bg-indigo-100 text-indigo-700 border-indigo-300'
        : 'bg-gray-100 text-gray-600 border-gray-200 hover:bg-gray-200'"
      class="flex items-center gap-1 px-1.5 py-0.5 text-xs rounded-full border transition-colors"
    >
      <span>{{ group.emoji }}</span>
      <span>{{ group.count }}</span>
    </button>
    <div class="relative" v-if="showPicker">
      <div class="absolute z-10 bg-white border border-gray-200 rounded-lg shadow-lg p-1 flex gap-0.5 -top-9 left-0">
        <button
          v-for="emoji in quickEmojis"
          :key="emoji"
          @click="add(emoji)"
          class="w-7 h-7 hover:bg-gray-100 rounded text-base"
        >{{ emoji }}</button>
      </div>
    </div>
    <button
      v-if="!showPicker"
      @click="showPicker = true"
      class="text-gray-400 hover:text-gray-600 text-xs px-1"
      :title="t('chat.reaction.add')"
    >😊+</button>
    <button
      v-else
      @click="showPicker = false"
      class="text-gray-400 hover:text-gray-600 text-xs px-1"
    >✕</button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ChatMessage } from '@/types/chat'
import * as chatApi from '@/api/chat'

const props = defineProps<{
  message: ChatMessage
  currentUserId: number
}>()
const emit = defineEmits<{ (e: 'refresh'): void }>()
const { t } = useI18n()

const showPicker = ref(false)
const quickEmojis = ['👍', '❤️', '🎉', '😢', '🚀', '👀', '✅', '❓']

function hasReacted(userIds: number[]): boolean {
  return userIds.includes(props.currentUserId)
}

async function toggle(emoji: string) {
  try {
    if (hasReacted(props.message.reactions.find((r) => r.emoji === emoji)?.user_ids || [])) {
      await chatApi.removeReaction(props.message.id, emoji)
    } else {
      await chatApi.addReaction(props.message.id, emoji)
    }
    emit('refresh')
  } catch (err) {
    console.error('[MessageReactions] toggle failed:', err)
  }
}

async function add(emoji: string) {
  showPicker.value = false
  try {
    await chatApi.addReaction(props.message.id, emoji)
    emit('refresh')
  } catch (err) {
    console.error('[MessageReactions] add failed:', err)
  }
}
</script>
