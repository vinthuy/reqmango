<template>
  <Transition name="slide">
    <div v-if="visible" class="fixed right-0 top-0 h-full w-[480px] max-w-[100vw] bg-white dark:bg-gray-850 shadow-2xl border-l border-gray-200 dark:border-gray-700 z-50 flex flex-col">
      <!-- Header with Tab Bar -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 bg-gradient-to-r from-indigo-500 to-purple-600 text-white">
        <div class="flex items-center gap-2">
          <span class="text-lg">🤖</span>
          <div>
            <h3 class="font-semibold text-sm">{{ t('ai.copilotTitle') }}</h3>
            <p class="text-xs text-indigo-100">{{ projectName || 'Project' }}</p>
          </div>
        </div>
        <button @click="close" class="p-1 hover:bg-white/20 rounded text-white" :title="t('common.close')">✕</button>
      </div>

      <!-- Tab Bar -->
      <div class="flex border-b border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="mode = tab.id"
          :class="[
            'flex-1 py-2.5 text-xs font-medium transition-colors relative',
            mode === tab.id
              ? 'text-indigo-600 dark:text-indigo-400'
              : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
          ]"
          :title="tab.label"
        >
          <span class="mr-1">{{ tab.icon }}</span>{{ tab.label }}
          <div
            v-if="mode === tab.id"
            class="absolute bottom-0 left-1/4 right-1/4 h-0.5 rounded-full"
            :class="tab.color"
          ></div>
        </button>
      </div>

      <!-- Agent Selector (agent mode only) -->
      <div v-if="mode === 'agent'" class="px-3 py-2 border-b border-gray-200 dark:border-gray-700">
        <div v-if="!selectedAgent" class="relative">
          <button
            @click="showAgentPicker = !showAgentPicker"
            class="w-full text-left px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:border-violet-400 transition-colors"
          >
            🤖 {{ t('ai.chooseAgent') }}
          </button>
          <div
            v-if="showAgentPicker"
            class="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10 max-h-48 overflow-y-auto"
          >
            <div v-if="!agents.length" class="px-3 py-2 text-sm text-gray-400">{{ t('ai.noAgents') }}</div>
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
          <button @click="selectedAgent = null; showAgentPicker = false" class="text-xs text-gray-400 hover:text-red-500">{{ t('ai.change') }}</button>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="flex gap-1.5 px-3 py-2 border-b border-gray-200 dark:border-gray-700 flex-wrap">
        <template v-if="mode === 'agent'">
          <button @click="sendAgent(t('ai.projectAnalysis'))" class="text-xs px-2 py-1 rounded-full bg-violet-50 dark:bg-violet-900/20 text-violet-600 dark:text-violet-300 hover:bg-violet-100">{{ t('ai.projectAnalysisLabel') }}</button>
          <button @click="sendAgent(t('ai.sprintSummary'))" class="text-xs px-2 py-1 rounded-full bg-violet-50 dark:bg-violet-900/20 text-violet-600 dark:text-violet-300 hover:bg-violet-100">{{ t('ai.sprintSummaryLabel') }}</button>
          <button @click="sendAgent(t('ai.triage'))" class="text-xs px-2 py-1 rounded-full bg-violet-50 dark:bg-violet-900/20 text-violet-600 dark:text-violet-300 hover:bg-violet-100">{{ t('ai.triageLabel') }}</button>
        </template>
        <template v-else-if="mode === 'chart'">
          <button @click="sendChartQuery(t('ai.chartPieState'))" class="text-xs px-2 py-1 rounded-full bg-purple-50 dark:bg-purple-900/20 text-purple-600 dark:text-purple-300 hover:bg-purple-100">{{ t('ai.chartPieLabel') }}</button>
          <button @click="sendChartQuery(t('ai.chartBarPriority'))" class="text-xs px-2 py-1 rounded-full bg-purple-50 dark:bg-purple-900/20 text-purple-600 dark:text-purple-300 hover:bg-purple-100">{{ t('ai.chartBarLabel') }}</button>
          <button @click="sendChartQuery(t('ai.chartLineTrend'))" class="text-xs px-2 py-1 rounded-full bg-purple-50 dark:bg-purple-900/20 text-purple-600 dark:text-purple-300 hover:bg-purple-100">{{ t('ai.chartLineLabel') }}</button>
        </template>
        <template v-else-if="mode === 'create'">
          <span class="text-xs text-gray-400 px-1 py-1">{{ t('ai.createQuickHint') }}</span>
        </template>
        <template v-else>
          <button @click="sendQuickAction(t('ai.summarizeIssue'))" class="text-xs px-2 py-1 rounded-full bg-indigo-50 dark:bg-indigo-800 text-indigo-600 dark:text-indigo-300 hover:bg-indigo-100">{{ t('ai.summarizeLabel') }}</button>
          <button @click="sendQuickAction(t('ai.riskAnalysis'))" class="text-xs px-2 py-1 rounded-full bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-300 hover:bg-amber-100">{{ t('ai.riskLabel') }}</button>
          <button @click="sendQuickAction(t('ai.suggestSteps'))" class="text-xs px-2 py-1 rounded-full bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-300 hover:bg-green-100">{{ t('ai.suggestLabel') }}</button>
        </template>
      </div>

      <!-- Messages / Content Area -->
      <div ref="msgContainer" class="flex-1 overflow-y-auto p-4 space-y-3">
        <!-- Empty state -->
        <div v-if="messages.length === 0 && mode !== 'create'" class="text-center text-gray-400 mt-8">
          <div class="text-4xl mb-3">{{ mode === 'agent' ? '👥' : mode === 'chart' ? '📊' : '🤖' }}</div>
          <p class="text-sm font-medium">{{ emptyTitle }}</p>
          <p class="text-xs mt-1">{{ emptyHint }}</p>
          <div v-if="mode === 'ask' || mode === 'build'" class="mt-4 space-y-2 text-xs text-left">
            <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer" @click="send(t('ai.suggestedQuestion1'))">💡 "{{ t('ai.suggestedQuestion1') }}"</div>
            <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer" @click="send(t('ai.suggestedQuestion2'))">💡 "{{ t('ai.suggestedQuestion2') }}"</div>
            <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer" @click="send(t('ai.suggestedQuestion3'))">💡 "{{ t('ai.suggestedQuestion3') }}"</div>
          </div>
        </div>

        <!-- Create Mode：AI Smart Create -->
        <div v-if="mode === 'create'" class="space-y-4">
          <div class="text-center">
            <div class="text-3xl mb-2">✨</div>
            <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('ai.createModeTitle') }}</p>
            <p class="text-xs text-gray-500 mt-1">{{ t('ai.createHint') }}</p>
          </div>

          <textarea
            v-model="createInput"
            rows="3"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            :placeholder="t('ai.createPlaceholder')"
          ></textarea>

          <button
            @click="generatePreview"
            :disabled="!createInput.trim() || isCreating"
            class="w-full py-2 bg-gradient-to-r from-indigo-500 to-purple-600 text-white text-sm font-medium rounded-lg hover:from-indigo-600 hover:to-purple-700 disabled:opacity-50 transition"
          >
            {{ isCreating ? t('common.loading') : t('ai.generatePreview') }}
          </button>

          <!-- Preview -->
          <div v-if="createPreview" class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700 space-y-3">
            <div class="flex items-center justify-between">
              <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">{{ t('ai.preview') }}</h4>
              <span class="text-xs text-gray-400">{{ createExplanation }}</span>
            </div>

            <div v-if="createError" class="text-xs text-red-500 bg-red-50 dark:bg-red-900/20 rounded p-2">{{ createError }}</div>

            <div class="space-y-2">
              <div>
                <label class="text-xs text-gray-500">{{ t('issue.name') }}</label>
                <input v-model="createPreview.name" class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100" />
              </div>
              <div class="flex gap-2">
                <div class="flex-1">
                  <label class="text-xs text-gray-500">{{ t('issue.priority') }}</label>
                  <select v-model="createPreview.priority" class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
                    <option value="urgent">{{ t('issue.priorityUrgent') }}</option>
                    <option value="high">{{ t('issue.priorityHigh') }}</option>
                    <option value="medium">{{ t('issue.priorityMedium') }}</option>
                    <option value="low">{{ t('issue.priorityLow') }}</option>
                    <option value="none">{{ t('issue.priorityNone') }}</option>
                  </select>
                </div>
                <div class="flex-1">
                  <label class="text-xs text-gray-500">{{ t('issue.type') }}</label>
                  <select v-model="createPreview.type_id" class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
                    <option value="">{{ t('issue.notSet') }}</option>
                    <option v-for="it in issueTypes" :key="it.id" :value="it.id">{{ it.name }}</option>
                  </select>
                </div>
              </div>
              <div>
                <label class="text-xs text-gray-500">{{ t('issue.description') }}</label>
                <textarea v-model="createPreview.description" rows="2" class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded text-sm resize-none bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"></textarea>
              </div>
            </div>

            <button
              @click="confirmCreate"
              :disabled="!createPreview.name?.trim() || isCreating"
              class="w-full py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 disabled:opacity-50 transition"
            >
              ✓ {{ t('ai.confirmCreate') }}
            </button>
          </div>
        </div>

        <!-- Chat Messages -->
        <div v-for="(msg, idx) in messages" :key="idx" class="flex" :class="msg.role === 'user' ? 'justify-end' : 'justify-start'">
          <div
            :class="[
              'rounded-2xl px-4 py-2.5 leading-relaxed',
              msg.role === 'user'
                ? 'rounded-br-md bg-indigo-50 dark:bg-indigo-900/20 text-gray-900 dark:text-gray-100 font-medium max-w-[85%] ml-auto'
                : 'rounded-bl-md bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200 max-w-[85%]',
              (msg.toolResults?.length || msg.chartConfig) ? 'max-w-[98%]' : 'max-w-[85%]'
            ]"
          >
            <!-- AI Thinking Process (collapsible) -->
            <div v-if="msg.role === 'assistant' && msg.toolResults?.length" class="mb-3">
              <button
                @click="toggleThinking(idx)"
                class="w-full flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors group"
              >
                <span :class="['transition-transform', expandedThinking.has(idx) ? 'rotate-90' : '']">▶</span>
                <span>{{ t('ai.thinkingProcess') }}</span>
                <span class="text-gray-300 dark:text-gray-600">—</span>
                <span>{{ summarizeThinking(msg.toolResults) }}</span>
                <span class="ml-auto text-gray-300 dark:text-gray-600 group-hover:text-gray-400">{{ expandedThinking.has(idx) ? t('ai.collapse') : t('ai.expand') }}</span>
              </button>
              <div v-if="expandedThinking.has(idx)" class="mt-2 space-y-2 border-l-2 border-amber-200 dark:border-amber-800 pl-3">
                <div v-for="(tr, ti) in msg.toolResults" :key="ti" class="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden text-[11px]">
                  <div class="px-2.5 py-1.5 bg-gray-50 dark:bg-gray-800 border-b border-gray-100 dark:border-gray-700 font-medium text-gray-500 dark:text-gray-400 flex items-center gap-1.5">
                    <span>📋</span><span>{{ formatToolName(tr.toolName || '') }}</span>
                    <span v-if="tr.rows?.length" class="text-gray-300">· {{ tr.rows.length }} {{ t('ai.chartCount') }}</span>
                  </div>
                  <div class="overflow-x-auto">
                    <table v-if="(tr.columns?.length || 0) > 1 || tr.columns?.[0] !== 'key'" class="w-full">
                      <thead>
                        <tr class="border-b border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-850">
                          <th v-for="col in (tr.columns || [])" :key="col" class="text-left px-2.5 py-1.5 text-gray-400 dark:text-gray-500 font-medium whitespace-nowrap">{{ formatColName(col) }}</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(row, ri) in (tr.rows || [])" :key="ri" class="border-b border-gray-50 dark:border-gray-800 last:border-0">
                          <td v-for="col in (tr.columns || [])" :key="col" class="px-2.5 py-1.5 whitespace-nowrap max-w-[200px] truncate" :title="String(row[col] ?? '')">
                            <span v-if="col === 'name' && row['color']" class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[11px] font-medium" :style="{ backgroundColor: row['color'] + '20', color: row['color'] }">
                              <span class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: row['color'] }"></span>{{ row[col] }}
                            </span>
                            <span v-else-if="col === 'priority' && row[col]" :class="['inline-block px-1.5 py-0.5 rounded text-[10px] font-medium', priorityBadge(row[col])]">{{ row[col] }}</span>
                            <span v-else :class="col === 'name' || col === 'title' ? 'text-gray-800 dark:text-gray-200 font-medium' : 'text-gray-500 dark:text-gray-400'">{{ row[col] ?? '' }}</span>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                    <div v-else class="divide-y divide-gray-50 dark:divide-gray-800">
                      <div v-for="(row, ri) in (tr.rows || [])" :key="ri" class="flex px-2.5 py-1.5 gap-3">
                        <span class="text-gray-400 shrink-0">{{ row.key }}</span>
                        <span class="text-gray-800 dark:text-gray-200 truncate">{{ row.value }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Text content -->
            <div v-if="msg.content" class="ai-content leading-relaxed" v-html="renderMarkdown(msg.content)"></div>
            <div v-else-if="isStreaming && idx === messages.length - 1 && !msg.toolResults?.length" class="italic text-gray-400">...</div>

            <!-- Chart -->
            <div v-if="msg.chartConfig" class="mt-2 bg-white dark:bg-gray-900 rounded-lg p-3 border border-gray-200 dark:border-gray-700" style="min-width:250px">
              <AIChartRenderer :config="msg.chartConfig" />
              <div class="flex items-center gap-2 mt-2 pt-2 border-t border-gray-100 dark:border-gray-700">
                <button @click="openQuickCreate(msg)" class="flex items-center gap-1 px-2 py-1 text-[11px] bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-300 rounded-md hover:bg-indigo-100 dark:hover:bg-indigo-900/40 transition-colors" :title="t('ai.createIssue')">
                  <span>📋</span> {{ t('ai.createIssue') }}
                </button>
                <button @click="saveChartToDashboard(msg)" class="flex items-center gap-1 px-2 py-1 text-[11px] bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-300 rounded-md hover:bg-green-100 dark:hover:bg-green-900/40 transition-colors" :title="t('ai.saveToDashboard')">
                  <span>📊</span> {{ t('ai.saveToDashboard') }}
                </button>
              </div>
            </div>

            <!-- Action buttons for AI responses -->
            <div v-if="msg.role === 'assistant' && msg.content && !msg.toolResults?.length" class="flex items-center gap-2 mt-2">
              <button @click="openQuickCreate(msg)" class="flex items-center gap-1 px-2 py-1 text-[11px] bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-300 rounded-md hover:bg-indigo-100 dark:hover:bg-indigo-900/40 transition-colors" :title="t('ai.createIssue')">
                <span>📋</span> {{ t('ai.createIssue') }}
              </button>
              <button @click="saveContentAsPage(msg)" class="flex items-center gap-1 px-2 py-1 text-[11px] bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-300 rounded-md hover:bg-amber-100 dark:hover:bg-amber-900/40 transition-colors" :title="t('ai.saveAsPage')">
                <span>📝</span> {{ t('ai.saveAsPage') }}
              </button>
            </div>
          </div>
        </div>

        <div v-if="isStreaming && messages.length > 0 && !messages[messages.length - 1].content" class="flex justify-start">
          <div class="bg-gray-100 dark:bg-gray-800 rounded-xl px-3 py-2">
            <span class="inline-flex gap-1">
              <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay:0ms"></span>
              <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay:150ms"></span>
              <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay:300ms"></span>
            </span>
          </div>
        </div>

        <div v-if="error" class="bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-xs px-3 py-2 rounded-lg">{{ error }}</div>
      </div>

      <!-- Input (hidden in create mode since it has its own) -->
      <div v-if="mode !== 'create'" class="border-t border-gray-200 dark:border-gray-700 p-3 relative">
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
import { useI18n } from '@/composables/useI18n'
import { useAI } from '@/composables/useAI'
import { renderMarkdown } from '@/composables/useMarkdown'
import { generateChart, createPreviewWithAI } from '@/api/ai'
import { agentApi } from '@/api/agent'
import issueApi from '@/api/issue'
import * as issueTypeApi from '@/api/issue-type'
import type { AIChartData } from '@/api/ai'
import type { Agent } from '@/types/agent'
import AIChartRenderer from '@/components/AIChartRenderer.vue'

const props = defineProps<{
  visible: boolean
  projectId: number
  workspaceId: number
  projectName?: string
  initialMode?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'quickCreate', preview: Record<string, any>): void
  (e: 'saveAsPage', data: { title: string; content: string }): void
  (e: 'issueCreated'): void
}>()

const { t } = useI18n()
const { messages, isStreaming, error, sendMessage, cancel, clear } = useAI()

type CopilotMode = 'ask' | 'build' | 'create' | 'chart' | 'agent'

const mode = ref<CopilotMode>((props.initialMode as CopilotMode) || 'ask')
const input = ref('')
const loadingChart = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)
const msgContainer = ref<HTMLDivElement | null>(null)

const tabs = [
  { id: 'ask' as const, icon: '💬', label: t('ai.tabAsk'), color: 'bg-indigo-500' },
  { id: 'build' as const, icon: '🔧', label: t('ai.tabBuild'), color: 'bg-amber-500' },
  { id: 'create' as const, icon: '✨', label: t('ai.tabCreate'), color: 'bg-green-500' },
  { id: 'chart' as const, icon: '📊', label: t('ai.tabChart'), color: 'bg-emerald-500' },
  { id: 'agent' as const, icon: '👥', label: t('ai.tabAgent'), color: 'bg-violet-500' },
]

const emptyTitle = computed(() => {
  if (mode.value === 'agent') return t('ai.agentMode')
  if (mode.value === 'chart') return t('ai.chartReady')
  return t('ai.ready')
})

const emptyHint = computed(() => {
  if (mode.value === 'agent') return t('ai.agentHint')
  if (mode.value === 'chart') return t('ai.chartHint')
  return t('ai.readyHint')
})

// Agent mode state
const agents = ref<Agent[]>([])
const selectedAgent = ref<Agent | null>(null)
const showAgentPicker = ref(false)
const mentionFilter = ref('')
const showMentionPopup = ref(false)
const mentionStartIdx = ref(-1)

const expandedThinking = ref(new Set<number>())

function toggleThinking(idx: number) {
  if (expandedThinking.value.has(idx)) {
    expandedThinking.value.delete(idx)
  } else {
    expandedThinking.value.add(idx)
  }
  expandedThinking.value = new Set(expandedThinking.value)
}

function summarizeThinking(results: any[]): string {
  if (!results?.length) return ''
  const parts = results.map(tr => {
    const name = formatToolName(tr.toolName || '')
    const count = tr.rows?.length || 0
    return count ? `${name}(${count})` : name
  })
  return parts.join(' · ')
}

const filteredAgents = computed(() => {
  if (!mentionFilter.value) return agents.value.filter(a => a.status === 'active')
  const q = mentionFilter.value.toLowerCase()
  return agents.value.filter(a => a.status === 'active' && a.name.toLowerCase().includes(q))
})

const inputPlaceholder = computed(() => {
  if (mode.value === 'agent') return selectedAgent.value ? t('ai.agentTaskPlaceholder').replace('{name}', selectedAgent.value.name) : t('ai.selectAgentFirst')
  if (mode.value === 'chart') return t('ai.chartPlaceholder')
  if (mode.value === 'ask') return t('ai.placeholder')
  return t('ai.buildPlaceholder')
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

// ─── Create Mode ───
const createInput = ref('')
const createPreview = ref<any>(null)
const createExplanation = ref('')
const isCreating = ref(false)
const createError = ref('')
const issueTypes = ref<any[]>([])

async function loadIssueTypes() {
  if (issueTypes.value.length > 0) return
  try {
    const types = await issueTypeApi.getIssueTypes(props.workspaceId, props.projectId)
    issueTypes.value = Array.isArray(types) ? types : (types?.data || [])
  } catch (e) { /* ignore */ }
}

async function generatePreview() {
  if (!createInput.value.trim()) return
  isCreating.value = true
  createError.value = ''
  createPreview.value = null
  try {
    const result: any = await createPreviewWithAI(props.projectId, {
      description: createInput.value,
      workspace_id: props.workspaceId,
    })
    createPreview.value = { name: '', priority: 'medium', type_id: '', description: '', ...(result.preview || {}) }
    createExplanation.value = result.explanation || ''
  } catch (e: any) {
    createError.value = e?.message || 'Failed to generate preview'
  } finally {
    isCreating.value = false
  }
}

async function confirmCreate() {
  if (!createPreview.value?.name?.trim()) return
  isCreating.value = true
  try {
    await issueApi.createIssue(props.projectId, props.workspaceId, {
      name: createPreview.value.name,
      priority: createPreview.value.priority || 'medium',
      description_html: createPreview.value.description || '',
      state_id: createPreview.value.state_id,
    })
    createInput.value = ''
    createPreview.value = null
    createExplanation.value = ''
    emit('issueCreated')
  } catch (e: any) {
    createError.value = e?.message || 'Failed to create issue'
  } finally {
    isCreating.value = false
  }
}

// ─── Agent Mode ───
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

  messages.value.push({
    role: 'user',
    content: selectedAgent.value ? `@${selectedAgent.value.name} ${msg}` : msg,
  })

  if (!selectedAgent.value) {
    messages.value.push({ role: 'assistant', content: 'Please select an agent first.' })
    return
  }

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

function formatToolName(name: string): string {
  if (!name) return ''
  const key = 'ai.tool_' + name
  const result = t(key)
  return result !== key ? result : name.replace(/_/g, ' ')
}

function formatColName(col: string): string {
  if (!col) return ''
  const key = 'ai.col_' + col
  const result = t(key)
  return result !== key ? result : col
}

function priorityBadge(p: string): string {
  const map: Record<string, string> = {
    urgent: 'bg-red-100 text-red-700', high: 'bg-orange-100 text-orange-700',
    medium: 'bg-yellow-100 text-yellow-700', low: 'bg-green-100 text-green-700',
    none: 'bg-gray-100 text-gray-500',
  }
  return map[p] || 'bg-gray-100 text-gray-500'
}

// ─── AI Result Actions ───
function openQuickCreate(msg: any) {
  const preview: Record<string, any> = { name: '', description: msg.content?.slice(0, 200) || '' }
  if (msg.content) {
    const lines = msg.content.split('\n').filter((l: string) => l.trim())
    if (lines.length > 0) {
      const title = lines[0].replace(/^[#*\-\d.]+\s*/, '').trim()
      if (title.length > 3) preview.name = title.slice(0, 120)
    }
  }
  // Switch to create mode and pre-fill
  mode.value = 'create'
  nextTick(() => {
    createInput.value = preview.description || ''
  })
}

function saveChartToDashboard(msg: any) {
  if (msg.chartConfig) {
    try {
      const key = 'ai_chart_save_' + Date.now()
      localStorage.setItem(key, JSON.stringify(msg.chartConfig))
      alert(t('ai.chartSaved') || 'Chart saved! You can add it to your dashboard.')
    } catch (e) { /* ignore */ }
  }
}

function saveContentAsPage(msg: any) {
  if (msg.content) {
    emit('saveAsPage', { title: t('ai.aiReport') || 'AI Report', content: msg.content })
  }
}

watch(() => messages.value.length, scrollToBottom)
watch(() => props.visible, (v) => {
  if (v) {
    nextTick(() => {
      if (mode.value !== 'create') inputRef.value?.focus()
    })
    fetchAgents()
    if (mode.value === 'create') loadIssueTypes()
  } else {
    clear()
    mode.value = 'ask'
    selectedAgent.value = null
    createInput.value = ''
    createPreview.value = null
  }
})
watch(mode, (newMode) => {
  if (newMode === 'create') {
    loadIssueTypes()
  } else {
    nextTick(() => inputRef.value?.focus())
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
