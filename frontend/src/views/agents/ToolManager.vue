<template>
  <div class="p-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('ai.tools.title') }}</h1>
        <p class="text-gray-500 mt-1">{{ t('ai.tools.description') }}</p>
      </div>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 border-b border-gray-200 mb-6">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        @click="activeTab = tab.key"
        class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === tab.key
          ? 'border-indigo-600 text-indigo-600'
          : 'border-transparent text-gray-500 hover:text-gray-700'"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- ==================== TOOLS TAB ==================== -->
    <template v-if="activeTab === 'tools'">
      <!-- Filter Bar -->
      <div class="flex items-center gap-3 mb-4">
        <select
          v-model="toolFilterCategory"
          class="border border-gray-300 rounded-lg px-3 py-2 text-sm"
        >
          <option value="">{{ t('ai.tools.allCategories') }}</option>
          <option value="api">API</option>
          <option value="function">{{ t('ai.tools.typeFunction') }}</option>
          <option value="mcp">MCP</option>
          <option value="workflow">{{ t('ai.tools.typeWorkflow') }}</option>
        </select>
        <select
          v-model="toolFilterStatus"
          class="border border-gray-300 rounded-lg px-3 py-2 text-sm"
        >
          <option value="">{{ t('ai.tools.allStatus') }}</option>
          <option value="active">{{ t('common.active') }}</option>
          <option value="inactive">{{ t('ai.tools.statusInactive') }}</option>
        </select>
        <div class="flex-1"></div>
        <button
          @click="openToolModal()"
          class="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg text-sm font-medium"
        >
          {{ t('ai.tools.createTool') }}
        </button>
      </div>

      <!-- Tools Table -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.name') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.tools.type') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.tools.category') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.status') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.tools.rateLimit') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.tools.timeout') }}</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="tool in filteredTools" :key="tool.id" class="hover:bg-gray-50">
                <td class="px-6 py-4">
                  <div>
                    <p class="font-medium text-gray-900">{{ tool.name }}</p>
                    <p class="text-sm text-gray-500 line-clamp-1">{{ tool.description }}</p>
                  </div>
                </td>
                <td class="px-6 py-4">
                  <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full"
                    :class="toolTypeClass(tool.tool_type)">
                    {{ tool.tool_type }}
                  </span>
                </td>
                <td class="px-6 py-4 text-sm text-gray-600">{{ tool.category || '-' }}</td>
                <td class="px-6 py-4">
                  <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full"
                    :class="tool.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'">
                    {{ tool.status === 'active' ? t('common.active') : t('ai.tools.statusInactive') }}
                  </span>
                </td>
                <td class="px-6 py-4 text-sm text-gray-600">{{ tool.rate_limit }}/min</td>
                <td class="px-6 py-4 text-sm text-gray-600">{{ tool.timeout }}ms</td>
                <td class="px-6 py-4 text-right text-sm">
                  <button @click="openToolModal(tool)" class="text-indigo-600 hover:text-indigo-800 mr-3">{{ t('common.edit') }}</button>
                  <button @click="handleDeleteTool(tool)" class="text-red-600 hover:text-red-800">{{ t('common.delete') }}</button>
                </td>
              </tr>
              <tr v-if="filteredTools.length === 0">
                <td colspan="7" class="px-6 py-12 text-center text-sm text-gray-400">
                  {{ t('ai.tools.noTools') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <!-- ==================== LOGS TAB ==================== -->
    <template v-if="activeTab === 'logs'">
      <!-- Stats Cards -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500">{{ t('ai.tools.totalCalls') }}</p>
              <p class="text-2xl font-bold text-gray-900">{{ logStats.total }}</p>
            </div>
            <div class="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">📞</div>
          </div>
        </div>
        <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500">{{ t('ai.tools.successRate') }}</p>
              <p class="text-2xl font-bold text-green-600">{{ logStats.successRate }}%</p>
            </div>
            <div class="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">✅</div>
          </div>
        </div>
        <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500">{{ t('ai.tools.p95Latency') }}</p>
              <p class="text-2xl font-bold text-purple-600">{{ logStats.p95 }}ms</p>
            </div>
            <div class="w-10 h-10 bg-purple-100 rounded-full flex items-center justify-center">⚡</div>
          </div>
        </div>
        <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500">{{ t('ai.tools.rateLimited') }}</p>
              <p class="text-2xl font-bold text-orange-600">{{ logStats.rateLimited }}</p>
            </div>
            <div class="w-10 h-10 bg-orange-100 rounded-full flex items-center justify-center">🚫</div>
          </div>
        </div>
      </div>

      <!-- Log Filters -->
      <div class="flex items-center gap-3 mb-4">
        <select
          v-model="logFilterStatus"
          class="border border-gray-300 rounded-lg px-3 py-2 text-sm"
        >
          <option value="">{{ t('ai.tools.allStatus') }}</option>
          <option value="success">{{ t('ai.tools.statusSuccess') }}</option>
          <option value="failed">{{ t('ai.tools.statusFailed') }}</option>
          <option value="timeout">{{ t('ai.tools.statusTimeout') }}</option>
        </select>
        <div class="flex-1"></div>
        <span class="text-xs text-gray-400">{{ t('ai.tools.sseLive') }}</span>
      </div>

      <!-- Log Timeline -->
      <div class="bg-white rounded-lg shadow-sm border border-gray-200">
        <div v-for="log in callLogs" :key="log.id"
          class="border-b border-gray-100 last:border-0 px-6 py-4 hover:bg-gray-50 cursor-pointer"
          @click="toggleLogDetail(log.id)"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <span class="w-2 h-2 rounded-full flex-shrink-0"
                :class="log.status === 'success' ? 'bg-green-500' : log.status === 'failed' ? 'bg-red-500' : 'bg-yellow-500'">
              </span>
              <div>
                <p class="font-medium text-gray-900">{{ log.tool_name }}</p>
                <p class="text-xs text-gray-400">{{ formatTime(log.created_at) }} · {{ log.duration_ms }}ms
                  <span v-if="log.rate_limited" class="text-orange-500">· Rate Limited</span>
                </p>
              </div>
            </div>
            <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full"
              :class="log.status === 'success' ? 'bg-green-100 text-green-800' : log.status === 'failed' ? 'bg-red-100 text-red-800' : 'bg-yellow-100 text-yellow-800'">
              {{ log.status }}
            </span>
          </div>
          <!-- Expanded Detail -->
          <div v-if="expandedLogId === log.id" class="mt-3 pl-5 text-sm space-y-2">
            <div>
              <span class="text-gray-500">{{ t('ai.tools.input') }}:</span>
              <pre class="text-xs bg-gray-50 p-2 rounded mt-1 overflow-x-auto">{{ JSON.stringify(log.input_params, null, 2) }}</pre>
            </div>
            <div>
              <span class="text-gray-500">{{ t('ai.tools.output') }}:</span>
              <pre class="text-xs bg-gray-50 p-2 rounded mt-1 overflow-x-auto">{{ JSON.stringify(log.output_result, null, 2) }}</pre>
            </div>
            <div v-if="log.error_message">
              <span class="text-red-500">{{ t('ai.tools.error') }}:</span>
              <pre class="text-xs bg-red-50 p-2 rounded mt-1 text-red-700">{{ log.error_message }}</pre>
            </div>
          </div>
        </div>
        <div v-if="callLogs.length === 0" class="px-6 py-12 text-center text-sm text-gray-400">
          {{ t('ai.tools.noLogs') }}
        </div>
      </div>
    </template>

    <!-- ==================== PERMISSIONS TAB ==================== -->
    <template v-if="activeTab === 'permissions'">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Tool List (left) -->
        <div class="bg-white rounded-lg shadow-sm border border-gray-200">
          <div class="px-4 py-3 border-b border-gray-200">
            <h3 class="font-semibold text-gray-900 text-sm">{{ t('ai.tools.toolList') }}</h3>
          </div>
          <div class="max-h-96 overflow-y-auto">
            <button
              v-for="tool in tools"
              :key="tool.id"
              @click="selectedPermTool = tool"
              class="w-full text-left px-4 py-3 border-b border-gray-100 last:border-0 hover:bg-gray-50 transition-colors"
              :class="selectedPermTool?.id === tool.id ? 'bg-indigo-50 border-l-2 border-l-indigo-600' : ''"
            >
              <p class="font-medium text-sm text-gray-900">{{ tool.name }}</p>
              <p class="text-xs text-gray-500">{{ tool.tool_type }}</p>
            </button>
            <div v-if="tools.length === 0" class="px-4 py-8 text-center text-sm text-gray-400">
              {{ t('ai.tools.noTools') }}
            </div>
          </div>
        </div>

        <!-- Permissions (right) -->
        <div class="lg:col-span-2 bg-white rounded-lg shadow-sm border border-gray-200">
          <div class="px-4 py-3 border-b border-gray-200">
            <h3 class="font-semibold text-gray-900 text-sm">
              {{ selectedPermTool ? t('ai.tools.agentPermissions') + ': ' + selectedPermTool.name : t('ai.tools.selectToolFirst') }}
            </h3>
          </div>
          <div v-if="selectedPermTool" class="p-4">
            <div v-if="agentTemplates.length === 0" class="text-center text-sm text-gray-400 py-8">
              {{ t('ai.tools.noTemplates') }}
            </div>
            <div v-else class="space-y-3">
              <div v-for="tpl in agentTemplates" :key="tpl.id"
                class="flex items-center justify-between p-3 rounded-lg border border-gray-200 hover:bg-gray-50">
                <div class="flex items-center gap-3">
                  <span class="text-xl">{{ tpl.icon || '🎭' }}</span>
                  <div>
                    <p class="font-medium text-sm text-gray-900">{{ tpl.name }}</p>
                    <p class="text-xs text-gray-500">{{ tpl.description || '-' }}</p>
                  </div>
                </div>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    class="sr-only peer"
                    :checked="isTemplateAllowed(tpl.id)"
                    @change="togglePerm(tpl.id, ($event.target as HTMLInputElement).checked)"
                  />
                  <div class="w-9 h-5 bg-gray-200 peer-focus:ring-2 peer-focus:ring-indigo-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-indigo-600"></div>
                </label>
              </div>
            </div>
          </div>
          <div v-else class="p-8 text-center text-sm text-gray-400">
            {{ t('ai.tools.selectToolFromLeft') }}
          </div>
        </div>
      </div>
    </template>

    <!-- ==================== MCP TAB ==================== -->
    <template v-if="activeTab === 'mcp'">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="cfg in mcpConfigs" :key="cfg.id"
          class="bg-white rounded-lg shadow-sm border border-gray-200 p-5 hover:shadow-md transition">
          <div class="flex items-start justify-between mb-3">
            <div>
              <h3 class="font-semibold text-gray-900">{{ cfg.name }}</h3>
              <p class="text-sm text-gray-500 mt-1 line-clamp-2">{{ cfg.description || '-' }}</p>
            </div>
            <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full"
              :class="cfg.is_enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'">
              {{ cfg.is_enabled ? t('ai.tools.enabled') : t('ai.tools.disabled') }}
            </span>
          </div>
          <div class="text-xs text-gray-400 mb-3 space-y-1">
            <p>{{ t('ai.tools.serverUrl') }}: {{ cfg.server_url }}</p>
            <p>{{ t('ai.tools.toolsCount') }}: {{ cfg.tools_count }}</p>
            <p v-if="cfg.last_sync_at">{{ t('ai.tools.lastSync') }}: {{ formatTime(cfg.last_sync_at) }}</p>
          </div>
          <div class="flex gap-2">
            <button
              @click="handleSyncMCP(cfg)"
              :disabled="syncingId === cfg.id"
              class="flex-1 px-3 py-2 text-sm font-medium rounded-lg border border-indigo-300 text-indigo-700 hover:bg-indigo-50 disabled:opacity-50"
            >
              {{ syncingId === cfg.id ? t('ai.tools.syncing') : t('ai.tools.syncTools') }}
            </button>
          </div>
        </div>
        <div v-if="mcpConfigs.length === 0" class="col-span-full px-6 py-12 text-center text-sm text-gray-400">
          {{ t('ai.tools.noMcpConfigs') }}
        </div>
      </div>
    </template>

    <!-- ==================== TOOL CREATE/EDIT MODAL ==================== -->
    <Teleport to="body">
      <div v-if="showToolModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
        <div class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
            <h3 class="text-lg font-semibold text-gray-900">
              {{ editingTool ? t('ai.tools.editTool') : t('ai.tools.createTool') }}
            </h3>
            <button @click="showToolModal = false" class="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
          </div>
          <div class="px-6 py-4 space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('common.name') }} *</label>
              <input v-model="toolForm.name" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" :placeholder="t('ai.tools.namePlaceholder')" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('common.description') }}</label>
              <textarea v-model="toolForm.description" rows="2" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" :placeholder="t('ai.tools.descPlaceholder')"></textarea>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.tools.type') }}</label>
                <select v-model="toolForm.tool_type" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm">
                  <option value="api">API</option>
                  <option value="function">{{ t('ai.tools.typeFunction') }}</option>
                  <option value="mcp">MCP</option>
                  <option value="workflow">{{ t('ai.tools.typeWorkflow') }}</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.tools.category') }}</label>
                <input v-model="toolForm.category" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.tools.endpoint') }}</label>
              <input v-model="toolForm.endpoint" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="https://..." />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.tools.rateLimit') }}</label>
                <input v-model.number="toolForm.rate_limit" type="number" min="0" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.tools.timeout') }} (ms)</label>
                <input v-model.number="toolForm.timeout" type="number" min="0" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" />
              </div>
            </div>
          </div>
          <div class="flex justify-end gap-3 px-6 py-4 border-t border-gray-200">
            <button @click="showToolModal = false" class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg">
              {{ t('common.cancel') }}
            </button>
            <button
              @click="handleSaveTool"
              :disabled="!toolForm.name"
              class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg disabled:opacity-50"
            >
              {{ editingTool ? t('common.save') : t('common.create') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { useSSE } from '@/composables/useSSE'
import { toolApi, type Tool, type ToolCallLog, type ToolPermissionView } from '@/api/tool'
import { mcpApi, type MCPConfig } from '@/api/mcp'
import { agentTemplateApi, type AgentTemplateResponse } from '@/api/agent-template'

const { t } = useI18n()
const { getWorkspaceId } = useWorkspaceId()

// --- Tab state ---
type TabKey = 'tools' | 'logs' | 'permissions' | 'mcp'
const activeTab = ref<TabKey>('tools')

const tabs = computed(() => [
  { key: 'tools' as TabKey, label: t('ai.tools.tabTools') },
  { key: 'logs' as TabKey, label: t('ai.tools.tabLogs') },
  { key: 'permissions' as TabKey, label: t('ai.tools.tabPermissions') },
  { key: 'mcp' as TabKey, label: t('ai.tools.tabMcp') },
])

// --- Workspace ---
const wsId = ref<number | null>(null)
async function resolveWsId(): Promise<number | null> {
  if (wsId.value) return wsId.value
  const id = await getWorkspaceId()
  wsId.value = id
  return id
}

// --- Tools state ---
const tools = ref<Tool[]>([])
const toolFilterCategory = ref('')
const toolFilterStatus = ref('')

const filteredTools = computed(() => {
  return tools.value.filter(t => {
    if (toolFilterCategory.value && t.tool_type !== toolFilterCategory.value) return false
    if (toolFilterStatus.value && t.status !== toolFilterStatus.value) return false
    return true
  })
})

// --- Tool Modal ---
const showToolModal = ref(false)
const editingTool = ref<Tool | null>(null)
const toolForm = ref({
  name: '',
  description: '',
  tool_type: 'api' as Tool['tool_type'],
  category: '',
  endpoint: '',
  rate_limit: 60,
  timeout: 30000,
})

function openToolModal(tool?: Tool) {
  if (tool) {
    editingTool.value = tool
    toolForm.value = {
      name: tool.name,
      description: tool.description,
      tool_type: tool.tool_type,
      category: tool.category,
      endpoint: tool.endpoint || '',
      rate_limit: tool.rate_limit,
      timeout: tool.timeout,
    }
  } else {
    editingTool.value = null
    toolForm.value = { name: '', description: '', tool_type: 'api', category: '', endpoint: '', rate_limit: 60, timeout: 30000 }
  }
  showToolModal.value = true
}

async function handleSaveTool() {
  const id = await resolveWsId()
  if (!id) return
  try {
    if (editingTool.value) {
      await toolApi.update(id, editingTool.value.id, toolForm.value)
    } else {
      await toolApi.create(id, toolForm.value)
    }
    showToolModal.value = false
    await loadTools()
  } catch (e: any) {
    console.error('Save tool failed:', e)
  }
}

async function handleDeleteTool(tool: Tool) {
  if (!confirm(t('ai.tools.confirmDelete', { name: tool.name }))) return
  const id = await resolveWsId()
  if (!id) return
  try {
    await toolApi.delete(id, tool.id)
    await loadTools()
  } catch (e: any) {
    console.error('Delete tool failed:', e)
  }
}

function toolTypeClass(type: string): string {
  const map: Record<string, string> = {
    api: 'bg-blue-100 text-blue-800',
    function: 'bg-green-100 text-green-800',
    mcp: 'bg-purple-100 text-purple-800',
    workflow: 'bg-orange-100 text-orange-800',
  }
  return map[type] || 'bg-gray-100 text-gray-800'
}

// --- Logs state ---
const callLogs = ref<ToolCallLog[]>([])
const logFilterStatus = ref('')
const expandedLogId = ref<number | null>(null)

const logStats = computed(() => {
  const total = callLogs.value.length
  const success = callLogs.value.filter(l => l.status === 'success').length
  const successRate = total > 0 ? Math.round((success / total) * 100) : 0
  const sorted = [...callLogs.value].sort((a, b) => b.duration_ms - a.duration_ms)
  const p95Index = Math.floor(total * 0.05)
  const p95 = sorted[p95Index]?.duration_ms || 0
  const rateLimited = callLogs.value.filter(l => l.rate_limited).length
  return { total, successRate, p95, rateLimited }
})

function toggleLogDetail(id: number) {
  expandedLogId.value = expandedLogId.value === id ? null : id
}

function formatTime(ts: string): string {
  if (!ts) return '-'
  const d = new Date(ts)
  return d.toLocaleString()
}

// --- Permissions state ---
const agentTemplates = ref<AgentTemplateResponse[]>([])
const selectedPermTool = ref<Tool | null>(null)
const toolPermissions = ref<ToolPermissionView[]>([])

watch(selectedPermTool, async (tool) => {
  if (!tool) { toolPermissions.value = []; return }
  const id = await resolveWsId()
  if (!id) return
  try {
    toolPermissions.value = await toolApi.listPermissions(id, tool.id)
  } catch {
    toolPermissions.value = []
  }
})

function isTemplateAllowed(templateId: number): boolean {
  return toolPermissions.value.some(p => p.agent_template_id === templateId && p.allowed)
}

async function togglePerm(templateId: number, allowed: boolean) {
  if (!selectedPermTool.value) return
  const id = await resolveWsId()
  if (!id) return
  try {
    await toolApi.setPermission(id, selectedPermTool.value.id, {
      tool_id: selectedPermTool.value.id,
      agent_template_id: templateId,
      allowed,
    })
    // Update local state
    const idx = toolPermissions.value.findIndex(p => p.agent_template_id === templateId)
    if (idx >= 0) {
      toolPermissions.value[idx].allowed = allowed
    } else {
      toolPermissions.value.push({ tool_id: selectedPermTool.value.id, agent_template_id: templateId, allowed })
    }
  } catch (e: any) {
    console.error('Toggle permission failed:', e)
  }
}

// --- MCP state ---
const mcpConfigs = ref<MCPConfig[]>([])
const syncingId = ref<number | null>(null)

async function handleSyncMCP(cfg: MCPConfig) {
  const id = await resolveWsId()
  if (!id) return
  syncingId.value = cfg.id
  try {
    const result = await toolApi.syncMCP(id, cfg.id)
    // Reload tools to pick up newly synced tools
    await loadTools()
    alert(t('ai.tools.syncComplete', { added: result.added, updated: result.updated }))
  } catch (e: any) {
    console.error('MCP sync failed:', e)
  } finally {
    syncingId.value = null
  }
}

// --- SSE real-time for logs ---
const { onEvent } = useSSE()

function handleToolCallSSE(_event: string, _data: any) {
  // Refresh logs when a tool call event arrives
  loadLogs()
}

let unsubSSE: (() => void) | null = null

function subscribeSSE() {
  if (unsubSSE) unsubSSE()
  unsubSSE = onEvent((event, data) => {
    if (event.startsWith('tool_call.')) {
      handleToolCallSSE(event, data)
    }
  })
}

function unsubscribeSSE() {
  if (unsubSSE) { unsubSSE(); unsubSSE = null }
}

// --- Data loading ---
async function loadTools() {
  const id = await resolveWsId()
  if (!id) return
  try {
    tools.value = await toolApi.list(id)
  } catch {
    tools.value = []
  }
}

async function loadLogs() {
  const id = await resolveWsId()
  if (!id) return
  try {
    const params: any = { per_page: 50 }
    if (logFilterStatus.value) params.status = logFilterStatus.value
    const res = await toolApi.listCallLogs(id, params)
    callLogs.value = res.items || []
  } catch {
    callLogs.value = []
  }
}

async function loadPermissions() {
  const id = await resolveWsId()
  if (!id) return
  try {
    agentTemplates.value = await agentTemplateApi.list(id)
  } catch {
    agentTemplates.value = []
  }
}

async function loadMCPConfigs() {
  const id = await resolveWsId()
  if (!id) return
  try {
    mcpConfigs.value = await mcpApi.list(id)
  } catch {
    mcpConfigs.value = []
  }
}

async function loadTabData() {
  const id = await resolveWsId()
  if (!id) return

  if (activeTab.value === 'tools') {
    await loadTools()
  } else if (activeTab.value === 'logs') {
    await loadLogs()
    subscribeSSE()
  } else if (activeTab.value === 'permissions') {
    await Promise.all([loadTools(), loadPermissions()])
  } else if (activeTab.value === 'mcp') {
    await loadMCPConfigs()
  }
}

watch(activeTab, () => {
  unsubscribeSSE()
  loadTabData()
})

onMounted(async () => {
  await loadTabData()
})

onUnmounted(() => {
  unsubscribeSSE()
})
</script>
