<template>
  <Transition name="slide">
    <div v-if="visible" class="fixed right-0 top-0 h-full w-96 bg-white shadow-2xl border-l border-gray-200 z-50 flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 bg-gradient-to-r from-indigo-500 to-purple-600 text-white">
        <div class="flex items-center gap-2">
          <span class="text-lg">🤖</span>
          <div>
            <h3 class="font-semibold text-sm">AI Assistant</h3>
            <p class="text-xs text-indigo-100">{{ projectName || 'Project' }}</p>
          </div>
        </div>
        <div class="flex items-center gap-1">
          <button
            @click="mode = mode === 'ask' ? 'build' : 'ask'"
            :class="['px-2 py-1 text-xs rounded transition', mode === 'build' ? 'bg-amber-500 text-white' : 'bg-white/20 text-white']"
            :title="mode === 'ask' ? '切换到 Build 模式' : '切换到 Ask 模式'"
          >
            {{ mode === 'ask' ? '💬 Ask' : '🔧 Build' }}
          </button>
          <button @click="close" class="p-1 hover:bg-white/20 rounded text-white">✕</button>
        </div>
      </div>

      <!-- Messages -->
      <div ref="msgContainer" class="flex-1 overflow-y-auto p-4 space-y-3">
        <div v-if="messages.length === 0" class="text-center text-gray-400 mt-8">
          <div class="text-4xl mb-3">🤖</div>
          <p class="text-sm font-medium">AI Assistant Ready</p>
          <p class="text-xs mt-1">Ask me about your project data!</p>
          <div class="mt-4 space-y-2 text-xs text-left">
            <div class="bg-gray-50 rounded-lg p-2 hover:bg-gray-100 cursor-pointer" @click="send('有哪些紧急的Bug？')">💡 "有哪些紧急的Bug？"</div>
            <div class="bg-gray-50 rounded-lg p-2 hover:bg-gray-100 cursor-pointer" @click="send('项目进展如何？')">💡 "项目进展如何？"</div>
            <div class="bg-gray-50 rounded-lg p-2 hover:bg-gray-100 cursor-pointer" @click="send('列出所有进行中的任务')">💡 "列出所有进行中的任务"</div>
          </div>
        </div>

        <div v-for="(msg, idx) in messages" :key="idx" class="flex" :class="msg.role === 'user' ? 'justify-end' : 'justify-start'">
          <div
            :class="['max-w-[85%] rounded-xl px-3 py-2 text-sm', msg.role === 'user' ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-800']"
          >
            <div v-if="msg.role === 'assistant' && msg.toolCalls?.length" class="mb-2">
              <div v-for="tc in msg.toolCalls" :key="tc.id" class="text-xs bg-amber-100 text-amber-700 px-2 py-1 rounded mb-1">
                🔧 {{ tc.name }}
              </div>
            </div>
            <div class="whitespace-pre-wrap" v-text="msg.content || (isStreaming && idx === messages.length - 1 ? '...' : '')"></div>
          </div>
        </div>

        <div v-if="isStreaming && messages.length > 0 && !messages[messages.length - 1].content" class="flex justify-start">
          <div class="bg-gray-100 rounded-xl px-3 py-2">
            <span class="inline-flex gap-1">
              <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay:0ms"></span>
              <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay:150ms"></span>
              <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay:300ms"></span>
            </span>
          </div>
        </div>

        <div v-if="error" class="bg-red-50 text-red-600 text-xs px-3 py-2 rounded-lg">{{ error }}</div>
      </div>

      <!-- Input -->
      <div class="border-t border-gray-200 p-3">
        <div class="flex items-center gap-2">
          <input
            ref="inputRef"
            v-model="input"
            @keydown.enter="send(input)"
            :disabled="isStreaming"
            :placeholder="mode === 'ask' ? 'Ask anything about the project...' : 'Describe what to build...'"
            class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent disabled:bg-gray-100"
          />
          <button
            @click="send(input)"
            :disabled="isStreaming || !input.trim()"
            class="px-3 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition"
          >
            {{ isStreaming ? '⏳' : '→' }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useAI } from '@/composables/useAI'

const props = defineProps<{
  visible: boolean
  projectId: number
  workspaceId: number
  projectName?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { messages, isStreaming, error, sendMessage, cancel, clear } = useAI()
const input = ref('')
const mode = ref<'ask' | 'build'>('ask')
const inputRef = ref<HTMLInputElement | null>(null)
const msgContainer = ref<HTMLDivElement | null>(null)

function close() {
  cancel()
  emit('close')
}

function send(text: string) {
  if (!text.trim() || isStreaming.value) return
  const msg = text.trim()
  input.value = ''
  sendMessage(msg, props.projectId, props.workspaceId, mode.value)
}

function scrollToBottom() {
  nextTick(() => {
    if (msgContainer.value) {
      msgContainer.value.scrollTop = msgContainer.value.scrollHeight
    }
  })
}

watch(() => messages.value.length, scrollToBottom)
watch(() => props.visible, (v) => {
  if (v) {
    nextTick(() => inputRef.value?.focus())
  } else {
    clear()
    mode.value = 'ask'
  }
})
</script>

<style scoped>
.slide-enter-active, .slide-leave-active {
  transition: transform 0.25s ease;
}
.slide-enter-from, .slide-leave-to {
  transform: translateX(100%);
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}
</style>
