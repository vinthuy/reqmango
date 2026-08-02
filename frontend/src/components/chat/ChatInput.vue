<template>
  <div class="border-t border-gray-200 p-3 bg-white">
    <div class="relative">
      <textarea
        v-model="text"
        @keydown="onKeydown"
        @input="onInput"
        :placeholder="t('chat.placeholder')"
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm resize-none focus:outline-none focus:border-indigo-400"
        rows="1"
        ref="ta"
      ></textarea>

      <!-- @mention picker -->
      <div
        v-if="mentionOpen && mentionCandidates.length"
        class="absolute bottom-full mb-1 left-0 bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto w-56"
      >
        <button
          v-for="c in mentionCandidates"
          :key="c.id"
          @click="pickMention(c)"
          class="w-full text-left px-3 py-1.5 text-sm hover:bg-indigo-50 flex items-center gap-2"
        >
          <span>{{ c.type === 'agent' ? '🤖' : '👤' }}</span>
          <span>{{ c.name }}</span>
        </button>
      </div>
    </div>

    <div class="flex items-center justify-between mt-1.5">
      <span class="text-[10px] text-gray-400">{{ text.length }}/10000</span>
      <div class="flex items-center gap-2">
        <span v-if="replyTo" class="text-xs text-gray-500">
          {{ t('chat.replyingTo') }} #{{ replyTo.id }}
          <button @click="$emit('cancel-reply')" class="text-gray-400 hover:text-gray-600 ml-1">✕</button>
        </span>
        <button
          @click="send"
          :disabled="!text.trim() || sending"
          class="px-3 py-1 bg-indigo-600 text-white text-sm rounded-lg disabled:opacity-50 hover:bg-indigo-700"
        >{{ t('chat.send') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ChatMessage } from '@/types/chat'

defineProps<{
  sending: boolean
  replyTo: ChatMessage | null
  mentionCandidates: { id: number; name: string; type: 'user' | 'agent' }[]
}>()
const emit = defineEmits<{
  (e: 'send', content: string): void
  (e: 'cancel-reply'): void
}>()

const { t } = useI18n()
const text = ref('')
const ta = ref<HTMLTextAreaElement | null>(null)
const mentionOpen = ref(false)
const mentionQuery = ref('')

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}

function onInput() {
  // Auto-grow
  const el = ta.value
  if (el) {
    el.style.height = 'auto'
    el.style.height = Math.min(el.scrollHeight, 160) + 'px'
  }
  // Detect @mention pattern at caret
  const match = text.value.match(/@([\w\u4e00-\u9fa5-]*)$/)
  if (match) {
    mentionOpen.value = true
    mentionQuery.value = match[1]
  } else {
    mentionOpen.value = false
  }
}

function pickMention(c: { id: number; name: string; type: 'user' | 'agent' }) {
  // Replace the trailing @query with @name + space
  text.value = text.value.replace(/@([\w\u4e00-\u9fa5-]*)$/, `@${c.name} `)
  mentionOpen.value = false
  nextTick(() => ta.value?.focus())
}

function send() {
  const content = text.value.trim()
  if (!content) return
  emit('send', content)
  text.value = ''
  mentionOpen.value = false
  const el = ta.value
  if (el) el.style.height = 'auto'
}
</script>
