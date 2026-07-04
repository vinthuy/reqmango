<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-semibold text-gray-900">{{ t('integrations.title') }}</h1>
        <p class="text-sm text-gray-500 mt-1">{{ t('integrations.desc') }}</p>
      </div>
    </div>

    <!-- Sub-tabs -->
    <div class="flex space-x-1 mb-6 border-b">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        :class="[
          'px-4 py-2 text-sm font-medium rounded-t-lg transition-colors',
          activeTab === tab.id ? 'bg-white border border-b-0 text-blue-700' : 'text-gray-500 hover:text-gray-700'
        ]"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- MCP Servers tab -->
    <div v-if="activeTab === 'mcp'" class="bg-white rounded-xl border p-6">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h2 class="text-lg font-semibold">{{ t('integrations.mcp.title') }}</h2>
          <p class="text-sm text-gray-500">{{ t('integrations.mcp.desc') }}</p>
        </div>
        <button @click="openMcpCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ {{ t('integrations.mcp.createTitle') }}</button>
      </div>

      <!-- Empty state -->
      <div v-if="mcpConfigs.length === 0 && !mcpLoading" class="text-center py-12 text-gray-400">
        <p class="text-lg mb-2">No MCP servers configured</p>
        <p class="text-sm">Add an MCP server to enable AI tool access.</p>
      </div>

      <!-- Loading -->
      <div v-if="mcpLoading" class="text-center py-8">
        <div class="animate-spin h-6 w-6 border-2 border-blue-500 border-t-transparent rounded-full mx-auto"></div>
      </div>

      <!-- MCP list -->
      <div v-for="cfg in mcpConfigs" :key="cfg.id" class="border rounded-lg p-4 mb-3">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="font-medium">{{ cfg.name }}</h3>
            <p class="text-sm text-gray-500">{{ cfg.server_url }}</p>
            <div class="flex items-center space-x-2 mt-1">
              <span :class="['text-xs px-1.5 py-0.5 rounded', cfg.is_enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500']">{{ cfg.is_enabled ? t('integrations.mcp.enabled') : t('integrations.mcp.disabled') }}</span>
              <span class="text-xs text-gray-400">{{ cfg.transport_type }}</span>
              <span class="text-xs text-gray-400">{{ cfg.tools_count }} {{ t('integrations.mcp.tools') }}</span>
            </div>
          </div>
          <div class="flex space-x-2">
            <button @click="discoverTools(cfg)" class="text-xs px-2 py-1 bg-purple-50 text-purple-600 rounded hover:bg-purple-100">{{ t('integrations.mcp.discover') }}</button>
            <button @click="confirmDelMcp(cfg)" class="text-xs px-2 py-1 text-red-500 hover:bg-red-50 rounded">{{ t('common.delete') }}</button>
          </div>
        </div>
        <!-- Tools list -->
        <div v-if="expandedMcpId === cfg.id && mcpTools.length > 0" class="mt-3 pt-3 border-t">
          <h4 class="text-sm font-medium mb-2">{{ t('integrations.mcp.availableTools') }}</h4>
          <div v-for="tool in mcpTools" :key="tool.name" class="text-xs text-gray-600 py-1 flex items-center space-x-2">
            <span class="font-mono bg-gray-100 px-1 rounded">{{ tool.name }}</span>
            <span v-if="tool.description" class="text-gray-400">{{ tool.description }}</span>
          </div>
        </div>
      </div>

      <!-- Create MCP modal -->
      <div v-if="showMcpModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showMcpModal=false">
        <div class="bg-white rounded-xl p-6 w-full max-w-md">
          <h3 class="text-lg font-semibold mb-4">Add MCP Server</h3>
          <div class="space-y-3">
            <div><label class="block text-sm font-medium mb-1">Name</label><input v-model="mcpForm.name" class="w-full px-3 py-2 border rounded-lg" placeholder="My MCP Server" /></div>
            <div><label class="block text-sm font-medium mb-1">Server URL</label><input v-model="mcpForm.server_url" class="w-full px-3 py-2 border rounded-lg" placeholder="https://mcp.example.com" /></div>
            <div><label class="block text-sm font-medium mb-1">Transport</label><select v-model="mcpForm.transport_type" class="w-full px-3 py-2 border rounded-lg"><option value="sse">SSE</option><option value="stdio">STDIO</option></select></div>
            <div><label class="block text-sm font-medium mb-1">API Key (optional)</label><input v-model="mcpForm.api_key" class="w-full px-3 py-2 border rounded-lg" type="password" /></div>
          </div>
          <div class="flex justify-end space-x-3 mt-6">
            <button @click="showMcpModal=false" class="px-4 py-2 border rounded-lg">Cancel</button>
            <button @click="createMcp" class="px-4 py-2 bg-blue-600 text-white rounded-lg">Create</button>
          </div>
        </div>
      </div>
    </div>

    <!-- GitHub tab -->
    <div v-if="activeTab === 'github'" class="bg-white rounded-xl border p-6">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h2 class="text-lg font-semibold">{{ t('integrations.github.title') }}</h2>
          <p class="text-sm text-gray-500">{{ t('integrations.github.desc') }}</p>
        </div>
        <button @click="openGitHubCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ {{ t('integrations.github.createTitle') }}</button>
      </div>

      <div v-if="githubConns.length === 0 && !githubLoading" class="text-center py-12 text-gray-400">
        <p class="text-lg mb-2">{{ t('integrations.github.noConnections') }}</p>
        <p class="text-sm">{{ t('integrations.github.noConnectionsDesc') }}</p>
      </div>

      <div v-if="githubLoading" class="text-center py-8">
        <div class="animate-spin h-6 w-6 border-2 border-blue-500 border-t-transparent rounded-full mx-auto"></div>
      </div>

      <div v-for="conn in githubConns" :key="conn.id" class="border rounded-lg p-4 mb-3">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="font-medium">{{ conn.repo_owner }}/{{ conn.repo_name }}</h3>
            <div class="flex items-center space-x-2 mt-1">
              <span :class="['text-xs px-1.5 py-0.5 rounded', conn.is_enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500']">{{ conn.is_enabled ? t('integrations.github.enabled') : t('integrations.github.disabled') }}</span>
              <span class="text-xs text-gray-400" v-if="conn.sync_issues">{{ t('integrations.github.syncIssues') }}</span>
              <span class="text-xs text-gray-400" v-if="conn.sync_prs">{{ t('integrations.github.syncPrs') }}</span>
              <span class="text-xs text-gray-400" v-if="conn.last_sync_at">{{ t('integrations.github.lastSync') }}: {{ conn.last_sync_at }}</span>
            </div>
          </div>
          <div class="flex space-x-2">
            <button @click="syncGitHub(conn)" class="text-xs px-2 py-1 bg-purple-50 text-purple-600 rounded hover:bg-purple-100">{{ t('integrations.github.syncNow') }}</button>
            <button @click="confirmDelGitHub(conn)" class="text-xs px-2 py-1 text-red-500 hover:bg-red-50 rounded">{{ t('common.delete') }}</button>
          </div>
        </div>
      </div>

      <!-- Create GitHub modal -->
      <div v-if="showGitHubModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showGitHubModal=false">
        <div class="bg-white rounded-xl p-6 w-full max-w-md">
          <h3 class="text-lg font-semibold mb-4">{{ t('integrations.github.createTitle') }}</h3>
          <div class="space-y-3">
            <div><label class="block text-sm font-medium mb-1">{{ t('integrations.github.project') }}</label><select v-model="ghForm.project_id" class="w-full px-3 py-2 border rounded-lg"><option :value="0">{{ t('integrations.github.selectProject') }}</option><option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option></select></div>
            <div><label class="block text-sm font-medium mb-1">{{ t('integrations.github.repoOwner') }}</label><input v-model="ghForm.repo_owner" class="w-full px-3 py-2 border rounded-lg" :placeholder="t('integrations.github.repoOwnerPlaceholder')" /></div>
            <div><label class="block text-sm font-medium mb-1">{{ t('integrations.github.repoName') }}</label><input v-model="ghForm.repo_name" class="w-full px-3 py-2 border rounded-lg" :placeholder="t('integrations.github.repoNamePlaceholder')" /></div>
            <div><label class="block text-sm font-medium mb-1">{{ t('integrations.github.accessToken') }}</label><input v-model="ghForm.access_token" class="w-full px-3 py-2 border rounded-lg" type="password" /></div>
            <div class="flex items-center space-x-4">
              <label class="flex items-center space-x-2 text-sm"><input type="checkbox" v-model="ghForm.sync_issues" /> <span>{{ t('integrations.github.syncIssues') }}</span></label>
              <label class="flex items-center space-x-2 text-sm"><input type="checkbox" v-model="ghForm.sync_prs" /> <span>{{ t('integrations.github.syncPrs') }}</span></label>
            </div>
          </div>
          <div class="flex justify-end space-x-3 mt-6">
            <button @click="showGitHubModal=false" class="px-4 py-2 border rounded-lg">{{ t('common.cancel') }}</button>
            <button @click="createGitHub" class="px-4 py-2 bg-blue-600 text-white rounded-lg">{{ t('common.create') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Slack tab -->
    <div v-if="activeTab === 'slack'" class="bg-white rounded-xl border p-6">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h2 class="text-lg font-semibold">{{ t('integrations.slack.title') }}</h2>
          <p class="text-sm text-gray-500">{{ t('integrations.slack.desc') }}</p>
        </div>
        <button @click="openSlackCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ {{ t('integrations.slack.createTitle') }}</button>
      </div>

      <div v-if="slackConns.length === 0 && !slackLoading" class="text-center py-12 text-gray-400">
        <p class="text-lg mb-2">{{ t('integrations.slack.noConnections') }}</p>
        <p class="text-sm">{{ t('integrations.slack.noConnectionsDesc') }}</p>
      </div>

      <div v-if="slackLoading" class="text-center py-8">
        <div class="animate-spin h-6 w-6 border-2 border-blue-500 border-t-transparent rounded-full mx-auto"></div>
      </div>

      <div v-for="conn in slackConns" :key="conn.id" class="border rounded-lg p-4 mb-3">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="font-medium">#{{ conn.channel_name }}</h3>
            <div class="flex items-center space-x-2 mt-1">
              <span :class="['text-xs px-1.5 py-0.5 rounded', conn.is_enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500']">{{ conn.is_enabled ? t('integrations.slack.enabled') : t('integrations.slack.disabled') }}</span>
              <span class="text-xs text-gray-400" v-if="conn.notify_on_create">{{ t('integrations.slack.onCreate') }}</span>
              <span class="text-xs text-gray-400" v-if="conn.notify_on_update">{{ t('integrations.slack.onUpdate') }}</span>
              <span class="text-xs text-gray-400" v-if="conn.notify_on_comment">{{ t('integrations.slack.onComment') }}</span>
              <span class="text-xs text-gray-400" v-if="conn.notify_on_complete">{{ t('integrations.slack.onComplete') }}</span>
            </div>
          </div>
          <div class="flex space-x-2">
            <button @click="testSlack(conn)" class="text-xs px-2 py-1 bg-green-50 text-green-600 rounded hover:bg-green-100">{{ t('integrations.slack.test') }}</button>
            <button @click="confirmDelSlack(conn)" class="text-xs px-2 py-1 text-red-500 hover:bg-red-50 rounded">{{ t('common.delete') }}</button>
          </div>
        </div>
      </div>

      <!-- Create Slack modal -->
      <div v-if="showSlackModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showSlackModal=false">
        <div class="bg-white rounded-xl p-6 w-full max-w-md">
          <h3 class="text-lg font-semibold mb-4">{{ t('integrations.slack.createTitle') }}</h3>
          <div class="space-y-3">
            <div><label class="block text-sm font-medium mb-1">{{ t('integrations.slack.project') }}</label><select v-model="slForm.project_id" class="w-full px-3 py-2 border rounded-lg"><option :value="0">{{ t('integrations.slack.selectProject') }}</option><option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option></select></div>
            <div><label class="block text-sm font-medium mb-1">{{ t('integrations.slack.channelName') }}</label><input v-model="slForm.channel_name" class="w-full px-3 py-2 border rounded-lg" :placeholder="t('integrations.slack.channelNamePlaceholder')" /></div>
            <div><label class="block text-sm font-medium mb-1">{{ t('integrations.slack.webhookUrl') }}</label><input v-model="slForm.webhook_url" class="w-full px-3 py-2 border rounded-lg" :placeholder="t('integrations.slack.webhookUrlPlaceholder')" /></div>
            <div><label class="block text-sm font-medium mb-1">{{ t('integrations.slack.botToken') }}</label><input v-model="slForm.bot_token" class="w-full px-3 py-2 border rounded-lg" type="password" /></div>
            <div class="grid grid-cols-2 gap-2">
              <label class="flex items-center space-x-2 text-sm"><input type="checkbox" v-model="slForm.notify_on_create" /> <span>{{ t('integrations.slack.onCreate') }}</span></label>
              <label class="flex items-center space-x-2 text-sm"><input type="checkbox" v-model="slForm.notify_on_update" /> <span>{{ t('integrations.slack.onUpdate') }}</span></label>
              <label class="flex items-center space-x-2 text-sm"><input type="checkbox" v-model="slForm.notify_on_comment" /> <span>{{ t('integrations.slack.onComment') }}</span></label>
              <label class="flex items-center space-x-2 text-sm"><input type="checkbox" v-model="slForm.notify_on_complete" /> <span>{{ t('integrations.slack.onComplete') }}</span></label>
            </div>
          </div>
          <div class="flex justify-end space-x-3 mt-6">
            <button @click="showSlackModal=false" class="px-4 py-2 border rounded-lg">{{ t('common.cancel') }}</button>
            <button @click="createSlack" class="px-4 py-2 bg-blue-600 text-white rounded-lg">{{ t('common.create') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { mcpApi, type MCPConfig, type MCPTool } from '@/api/mcp'
import { githubApi, type GitHubConnection } from '@/api/github'
import { slackApi, type SlackConnection } from '@/api/slack'
import { workspaceApi } from '@/api/workspace'
import { listProjects } from '@/api/project'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'

const props = defineProps<{ workspaceId: number; slug: string }>()
const { t } = useI18n()
const toast = useToast()

const tabs = computed(() => [
  { id: 'mcp', label: t('integrations.tabs.mcp') },
  { id: 'github', label: t('integrations.tabs.github') },
  { id: 'slack', label: t('integrations.tabs.slack') },
])
const activeTab = ref('mcp')

// Common
const projects = ref<any[]>([])
const workspaceId = ref(0)

// MCP state
const mcpConfigs = ref<MCPConfig[]>([])
const mcpLoading = ref(false)
const showMcpModal = ref(false)
const expandedMcpId = ref(0)
const mcpTools = ref<MCPTool[]>([])
const mcpForm = ref({ name: '', server_url: '', transport_type: 'sse', api_key: '' })

// GitHub state
const githubConns = ref<GitHubConnection[]>([])
const githubLoading = ref(false)
const showGitHubModal = ref(false)
const ghForm = ref({ project_id: 0, repo_owner: '', repo_name: '', access_token: '', webhook_secret: '', sync_issues: true, sync_prs: true })

// Slack state
const slackConns = ref<SlackConnection[]>([])
const slackLoading = ref(false)
const showSlackModal = ref(false)
const slForm = ref({ project_id: 0, channel_name: '', webhook_url: '', bot_token: '', notify_on_create: true, notify_on_update: true, notify_on_comment: false, notify_on_complete: true })

async function init() {
  const ws = await workspaceApi.getBySlug(props.slug)
  workspaceId.value = ws.id
  projects.value = await listProjects(ws.id)
}

// === MCP ===
async function loadMcp() {
  mcpLoading.value = true
  try { mcpConfigs.value = await mcpApi.list(workspaceId.value) } catch (e) { console.error(e) }
  finally { mcpLoading.value = false }
}
function openMcpCreate() { mcpForm.value = { name: '', server_url: '', transport_type: 'sse', api_key: '' }; showMcpModal.value = true }
async function createMcp() {
  try { await mcpApi.create(workspaceId.value, mcpForm.value); showMcpModal.value = false; await loadMcp() } catch (e) { console.error(e) }
}
async function discoverTools(cfg: MCPConfig) {
  try { mcpTools.value = await mcpApi.discoverTools(workspaceId.value, cfg.id); expandedMcpId.value = cfg.id; await loadMcp() } catch (e) { console.error(e) }
}
async function confirmDelMcp(cfg: MCPConfig) {
  if (confirm(`Delete MCP server "${cfg.name}"?`)) {
    try { await mcpApi.delete(workspaceId.value, cfg.id); await loadMcp() } catch (e) { console.error(e) }
  }
}

// === GitHub ===
async function loadGitHub() {
  githubLoading.value = true
  try { githubConns.value = await githubApi.list(workspaceId.value) } catch (e) { console.error(e) }
  finally { githubLoading.value = false }
}
function openGitHubCreate() { ghForm.value = { project_id: 0, repo_owner: '', repo_name: '', access_token: '', webhook_secret: '', sync_issues: true, sync_prs: true }; showGitHubModal.value = true }
async function createGitHub() {
  try { await githubApi.create(workspaceId.value, ghForm.value); showGitHubModal.value = false; await loadGitHub() } catch (e) { console.error(e) }
}
async function syncGitHub(conn: GitHubConnection) {
  try { await githubApi.syncIssues(workspaceId.value, conn.id); toast.success(t('integrations.github.syncCompleted')); await loadGitHub() } catch (e) { console.error(e) }
}
async function confirmDelGitHub(conn: GitHubConnection) {
  if (confirm(t('integrations.github.deleteConfirm', { name: `${conn.repo_owner}/${conn.repo_name}` }))) {
    try { await githubApi.delete(workspaceId.value, conn.id); await loadGitHub() } catch (e) { console.error(e) }
  }
}

// === Slack ===
async function loadSlack() {
  slackLoading.value = true
  try { slackConns.value = await slackApi.list(workspaceId.value) } catch (e) { console.error(e) }
  finally { slackLoading.value = false }
}
function openSlackCreate() { slForm.value = { project_id: 0, channel_name: '', webhook_url: '', bot_token: '', notify_on_create: true, notify_on_update: true, notify_on_comment: false, notify_on_complete: true }; showSlackModal.value = true }
async function createSlack() {
  try { await slackApi.create(workspaceId.value, slForm.value); showSlackModal.value = false; await loadSlack() } catch (e) { console.error(e) }
}
async function testSlack(conn: SlackConnection) {
  try { const r = await slackApi.testNotification(workspaceId.value, conn.id); toast.success(t('integrations.slack.testSent', { channel: r.channel, status: r.status })) } catch (e: any) { toast.error(t('integrations.slack.testFailed') + (e?.message || t('common.unknown'))) }
}
async function confirmDelSlack(conn: SlackConnection) {
  if (confirm(t('integrations.slack.deleteConfirm', { channel: conn.channel_name }))) {
    try { await slackApi.delete(workspaceId.value, conn.id); await loadSlack() } catch (e) { console.error(e) }
  }
}

onMounted(async () => {
  await init()
  loadMcp()
  loadGitHub()
  loadSlack()
})
</script>
