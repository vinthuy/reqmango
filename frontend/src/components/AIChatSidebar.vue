<template>
  <Transition name="slide">
    <div v-if="visible" class="fixed right-0 top-0 h-full w-96 bg-white dark:bg-gray-850 shadow-2xl border-l border-gray-200 dark:border-gray-700 z-50 flex flex-col">
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
          <button
            @click="mode = mode === 'chart' ? 'ask' : 'chart'"
            :class="['px-2 py-1 text-xs rounded transition', mode === 'chart' ? 'bg-emerald-500 text-white' : 'bg-white/20 text-white']"
            title="AI 图表生成"
          >
            📊
          </button>
          <button
            @click="mode = mode === 'agent' ? 'ask' : 'agent'"
            :class="['px-2 py-1 text-xs rounded transition', mode === 'agent' ? 'bg-violet-500 text-white' : 'bg-white/20 text-white']"
            title="Agent 模式"
          >
            👥
          </button>
          <button @click="close" class="p-1 hover:bg-white/20 rounded text-white">✕</button>
        </div>
      </div>

      <!-- Agent Selector (agent mode only) -->
      <div v-if="mode === 'agent'" class="px-3 py-2 border-b border-gray-200 dark:border-gray-700">
        <div v-if="!selectedAgent" class="relative">
          <button
            @click="showAgentPicker = !showAgentPicker"
            class="w-full text-left px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:border-violet-400 transition-colors"
          >
            🤖 Choose an AI Agent...
          </button>
          <div
            v-if="showAgentPicker"
            class="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10 max-h-48 overflow-y-auto"
          >
            <div v-if="!agents.length" class="px-3 py-2 text-sm text-gray-400">No agents available</div>
            <button
              v-for="agent in agents.filter(a => a.status === 'active')"
              :key="agent.id"
              @click="selectAgent(agent)"
              class="w-full text-left px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2 text-sm"
            >
              <span>{{ agent.avatar }}</span>
              <span class="text-gray-800 dark:text-gray-200">{{ agent.name }}</span>
              <span class="ml-auto text-xs text-gray-400">{{ (agent.capabilities || []).length || 'all' }} skills</span>
            </button>
          </div>
        </div>
        <div v-else class="flex items-center gap-2">
          <span class="text-xl">{{ selectedAgent.avatar }}</span>
          <div class="flex-1">
            <p class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ selectedAgent.name }}</p>
            <p class="text-xs text-gray-400">{{ selectedAgent.agent_type }} · {{ (selectedAgent.capabilities || []).join(', ') || 'all capabilities' }}</p>
          </div>
          <button @click="selectedAgent = null; showAgentPicker = false" class="text-xs text-gray-400 hover:text-red-500">Change</button>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="flex gap-1.5 px-3 py-2 border-b border-gray-200 dark:border-gray-700 flex-wrap">
        <template v-if="mode === 'agent'">
          <button @click="sendAgent('分析当前项目的健康状况')" class="text-xs px-2 py-1 rounded-full bg-violet-50 dark:bg-violet-900/20 text-violet-600 dark:text-violet-300 hover:bg-violet-100">📋 项目分析</button>
          <button @click="sendAgent('总结当前Sprint的进展')" class="text-xs px-2 py-1 rounded-full bg-violet-50 dark:bg-violet-900/20 text-violet-600 dark:text-violet-300 hover:bg-violet-100">📊 Sprint总结</button>
          <button @click="sendAgent('对未分类的Issue进行分诊')" class="text-xs px-2 py-1 rounded-full bg-violet-50 dark:bg-violet-900/20 text-violet-600 dark:text-violet-300 hover:bg-violet-100">🏥 分诊</button>
        </template>
        <template v-else-if="mode !== 'chart'">
          <button @click="sendQuickAction('总结当前工作项的关键信息')" class="text-xs px-2 py-1 rounded-full bg-indigo-50 dark:bg-indigo-800 text-indigo-600 dark:text-indigo-300 hover:bg-indigo-100">📝 总结</button>
          <button @click="sendQuickAction('分析这个工作项的风险和阻塞点')" class="text-xs px-2 py-1 rounded-full bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-300 hover:bg-amber-100">⚠️ 风险</button>
          <button @click="sendQuickAction('给出处理这个工作项的建议步骤')" class="text-xs px-2 py-1 rounded-full bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-300 hover:bg-green-100">💡 建议</button>
        </template>
        <template v-else>
          <button @click="sendChartQuery('按状态分布饼图')" class="text-xs px-2 py-1 rounded-full bg-purple-50 dark:bg-purple-900/20 text-purple-600 dark:text-purple-300 hover:bg-purple-100">🥧 状态分布</button>
          <button @click="sendChartQuery('按优先级柱状图')" class="text-xs px-2 py-1 rounded-full bg-purple-50 dark:bg-purple-900/20 text-purple-600 dark:text-purple-300 hover:bg-purple-100">📊 优先级</button>
          <button @click="sendChartQuery('近30天趋势折线图')" class="text-xs px-2 py-1 rounded-full bg-purple-50 dark:bg-purple-900/20 text-purple-600 dark:text-purple-300 hover:bg-purple-100">📈 趋势</button>
        </template>
      </div>

      <!-- Messages -->
      <div ref="msgContainer" class="flex-1 overflow-y-auto p-4 space-y-3">
        <div v-if="messages.length === 0" class="text-center text-gray-400 mt-8">
          <div class="text-4xl mb-3">{{ mode === 'agent' ? '👥' : '🤖' }}</div>
          <p class="text-sm font-medium">{{ mode === 'agent' ? 'Agent Mode' : 'AI Assistant Ready' }}</p>
          <p class="text-xs mt-1">{{ mode === 'agent' ? 'Select an agent and give it a task' : 'Ask me about your project data!' }}</p>
          <div v-if="mode !== 'agent' && mode !== 'chart'" class="mt-4 space-y-2 text-xs text-left">
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
            <div v-if="msg.chartConfig" class="mt-2 bg-white dark:bg-gray-900 rounded-lg p-3 border border-gray-200 dark:border-gray-700" style="min-width:250px">
              <AIChartRenderer :config="msg.chartConfig" />
            </div>
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
      <div class="border-t border-gray-200 dark:border-gray-700 p-3 relative">
        <!-- @mention popup -->
        <div
          v-if="showMentionPopup && filteredAgents.length > 0"
          class="absolute bottom-full left-3 right-3 mb-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10 max-h-32 overflow-y-auto"
        >
          <button
            v-for="agent in filteredAgents"
            :key="agent.id"
            @click="insertMention(agent)"
            class="w-full text-left px-3 py-1.5 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 flex items-center gap-2 text-sm"
          >
            <span>{{ agent.avatar }}</span>
            <span class="text-gray-800 dark:text-gray-200">{{ agent.name }}</span>
          </button>
        </div>
        <div class="flex items-center gap-2">
          <span v-if="mode === 'agent' && selectedAgent" class="text-lg">{{ selectedAgent.avatar }}</span>
          <input
            ref="inputRef"
            v-model="input"
            @keydown.enter.prevent="handleSend"
            @input="handleInputKeydown"
            :disabled="isStreaming || loadingChart"
            :placeholder="inputPlaceholder"
            class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent disabled:bg-gray-100 dark:disabled:bg-gray-700 outline-none bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
          <button
            @click="handleSend"
            :disabled="isStreaming || loadingChart || !input.trim() || (mode === 'agent' && !selectedAgent)"
            :class="['px-3 py-2 text-white text-sm rounded-lg disabled:opacity-50 transition', mode === 'agent' ? 'bg-violet-600 hover:bg-violet-700' : mode === 'chart' ? 'bg-emerald-600 hover:bg-emerald-700' : 'bg-indigo-600 hover:bg-indigo-700']"
          >
            {{ isStreaming || loadingChart ? '⏳' : '→' }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, computed } from 'vue'
import { useAI } from '@/composables/useAI'
import { generateChart } from '@/api/ai'
import { agentApi } from '@/api/agent'
import type { AIChartData } from '@/api/ai'
import type { Agent } from '@/types/agent'
import AIChartRenderer from '@/components/AIChartRenderer.vue'

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
const mode = ref<'ask' | 'build' | 'chart' | 'agent'>('ask')
const loadingChart = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)
const msgContainer = ref<HTMLDivElement | null>(null)

// Agent mode state
const agents = ref<Agent[]>([])
const selectedAgent = ref<Agent | null>(null)
const showAgentPicker = ref(false)
const mentionFilter = ref('')
const showMentionPopup = ref(false)
const mentionStartIdx = ref(-1)

const filteredAgents = computed(() => {
  if (!mentionFilter.value) return agents.value.filter(a => a.status === 'active')
  const q = mentionFilter.value.toLowerCase()
  return agents.value.filter(a => a.status === 'active' && a.name.toLowerCase().includes(q))
})

const inputPlaceholder = computed(() => {
  if (mode.value === 'agent') return selectedAgent.value ? `Ask ${selectedAgent.value.name} to do something...` : 'Select an agent first...'
  if (mode.value === 'chart') return '输入图表需求，如：按状态分布饼图...'
  if (mode.value === 'ask') return 'Ask anything about the project...'
  return 'Describe what to build...'
})

function close() {
  cancel()
  emit('close')
}

function sendQuickAction(msg: string) {
  send(msg)
}

function handleSend() {
  if (mode.value === 'agent') {
    sendAgent(input.value)
  } else if (mode.value === 'chart') {
    sendChartQuery(input.value)
  } else {
    send(input.value)
  }
}

async function fetchAgents() {
  try {
    agents.value = await agentApi.list(props.workspaceId)
  } catch (e) {
    console.error('Failed to fetch agents', e)
  }
}

function selectAgent(agent: Agent) {
  selectedAgent.value = agent
  showAgentPicker.value = false
}

function handleInputKeydown() {
  const val = input.value
  const cursorPos = (inputRef.value?.selectionStart || 0) - 1
  if (cursorPos >= 0 && val[cursorPos] === '@') {
    mentionStartIdx.value = cursorPos
    mentionFilter.value = ''
    showMentionPopup.value = true
  } else if (showMentionPopup.value && mentionStartIdx.value >= 0) {
    const fragment = val.slice(mentionStartIdx.value + 1)
    const spaceIdx = fragment.indexOf(' ')
    if (spaceIdx >= 0) {
      showMentionPopup.value = false
    } else {
      mentionFilter.value = fragment
    }
  }
}

function insertMention(agent: Agent) {
  const before = input.value.slice(0, mentionStartIdx.value)
  const after = input.value.slice(mentionStartIdx.value + 1 + mentionFilter.value.length)
  input.value = before + '@' + agent.name + ' ' + after
  showMentionPopup.value = false
  mentionStartIdx.value = -1
  mentionFilter.value = ''
  nextTick(() => inputRef.value?.focus())
}

function sendAgent(text: string) {
  if (!text.trim()) return
  const msg = text.trim()
  input.value = ''

  // Add user message
  messages.value.push({
    role: 'user',
    content: selectedAgent.value ? `@${selectedAgent.value.name} ${msg}` : msg,
  })

  if (!selectedAgent.value) {
    messages.value.push({ role: 'assistant', content: 'Please select an agent first.' })
    return
  }

  // Add loading placeholder
  messages.value.push({ role: 'assistant', content: '', toolCalls: [] })
  isStreaming.value = true

  agentApi
    .dispatch(props.workspaceId, selectedAgent.value.id, { task: msg })
    .then((activity) => {
      messages.value[messages.value.length - 1].content =
        `${selectedAgent.value?.avatar || '🤖'} **${selectedAgent.value?.name}**:\n\n${activity.result_summary}`
    })
    .catch((e) => {
      messages.value[messages.value.length - 1].content =
        `❌ Agent error: ${e?.response?.data?.message || e.message || 'Unknown error'}`
    })
    .finally(() => {
      isStreaming.value = false
      scrollToBottom()
    })
}

async function sendChartQuery(query: string) {
  if (!query.trim() || loadingChart.value) return
  const q = query.trim()
  input.value = ''
  loadingChart.value = true

  messages.value.push({ role: 'user', content: `📊 ${q}` })

  try {
    const chartData = await generateChart(props.projectId, q)
    messages.value.push({
      role: 'assistant',
      content: `**${chartData.title}**\n\n${chartData.chart_type} 图表，${chartData.labels.length} 个数据维度`,
      chartConfig: chartData as AIChartData,
    })
    scrollToBottom()
  } catch (e: any) {
    messages.value.push({
      role: 'assistant',
      content: `❌ 图表生成失败: ${e?.response?.data?.message || e.message || '请重试'}`,
    })
  } finally {
    loadingChart.value = false
    scrollToBottom()
  }
}

function send(text: string) {
  if (!text.trim() || isStreaming.value) return
  const msg = text.trim()
  input.value = ''
  // Only pass 'ask' or 'build' to sendMessage — agent mode uses sendAgent()
  const chatMode = (mode.value === 'agent' ? 'ask' : mode.value) as 'ask' | 'build'
  sendMessage(msg, props.projectId, props.workspaceId, chatMode)
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
    fetchAgents()
  } else {
    clear()
    mode.value = 'ask'
    selectedAgent.value = null
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
