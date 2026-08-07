<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { useSSE } from '@/composables/useSSE'
import {
  cicdConfigApi,
  cicdBuildApi,
  type CICDConfig,
  type CICDConfigCreate,
  type CICDConfigUpdate,
  type CICDProvider,
  type BuildRecord,
  type BuildTriggerRequest,
  type BuildStatus,
  type BuildTrigger,
  type BuildStage
} from '@/api/cicd'

const { getWorkspaceId } = useWorkspaceId()
const { onEvent } = useSSE()

const workspaceId = ref(0)
const activeTab = ref<'configs' | 'builds'>('configs')

// ---- Configs state ----
const configs = ref<CICDConfig[]>([])
const loadingConfigs = ref(true)
const showConfigModal = ref(false)
const editingConfig = ref<CICDConfig | null>(null)
const savingConfig = ref(false)
const errorMsg = ref('')

const PROVIDERS: CICDProvider[] = ['github_actions', 'gitlab_ci', 'jenkins', 'generic']
const TRIGGER_EVENT_OPTIONS = ['push', 'pull_request', 'manual', 'schedule', 'webhook']

const configForm = ref<CICDConfigCreate>(emptyConfig())
const triggerEventsInput = ref<string>('')
const extraConfigInput = ref<string>('{}')

function emptyConfig(): CICDConfigCreate {
  return {
    name: '',
    provider: 'generic',
    api_endpoint: '',
    project_slug: '',
    default_branch: 'main',
    auth_token_ref: '',
    trigger_events: [],
    extra_config: {},
    enabled: true
  }
}

// ---- Builds state ----
const builds = ref<BuildRecord[]>([])
const loadingBuilds = ref(true)
const showTriggerModal = ref(false)
const showBuildDetailModal = ref(false)
const selectedBuild = ref<BuildRecord | null>(null)
const triggering = ref(false)
const cancelingId = ref<number | null>(null)

const STATUS_FILTERS: { value: BuildStatus | ''; label: string }[] = [
  { value: '', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'queued', label: 'Queued' },
  { value: 'running', label: 'Running' },
  { value: 'success', label: 'Success' },
  { value: 'failed', label: 'Failed' },
  { value: 'cancelled', label: 'Cancelled' }
]
const activeStatusFilter = ref<BuildStatus | ''>('')

const triggerForm = ref<BuildTriggerRequest>(emptyTrigger())
const triggerProjectIdInput = ref('')
const triggerIssueIdInput = ref('')
const triggerAgentTaskIdInput = ref('')

function emptyTrigger(): BuildTriggerRequest {
  return {
    cicd_config_id: 0,
    branch: '',
    commit_sha: '',
    trigger: 'manual'
  }
}

const filteredBuilds = computed<BuildRecord[]>(() => {
  if (!activeStatusFilter.value) return builds.value
  return builds.value.filter(b => b.status === activeStatusFilter.value)
})

let sseCleanup: (() => void) | null = null

// ============ Data loading ============

async function loadConfigs() {
  loadingConfigs.value = true
  errorMsg.value = ''
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    workspaceId.value = wsId
    const res = await cicdConfigApi.list(wsId)
    configs.value = res || []
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to load CI/CD configs'
    configs.value = []
  } finally {
    loadingConfigs.value = false
  }
}

async function loadBuilds() {
  loadingBuilds.value = true
  errorMsg.value = ''
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    workspaceId.value = wsId
    const res = await cicdBuildApi.list(wsId)
    builds.value = res || []
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to load builds'
    builds.value = []
  } finally {
    loadingBuilds.value = false
  }
}

// ============ Config modal handlers ============

function openCreateConfigModal() {
  editingConfig.value = null
  configForm.value = emptyConfig()
  triggerEventsInput.value = ''
  extraConfigInput.value = '{}'
  errorMsg.value = ''
  showConfigModal.value = true
}

function openEditConfigModal(cfg: CICDConfig) {
  editingConfig.value = cfg
  configForm.value = {
    name: cfg.name,
    provider: cfg.provider,
    api_endpoint: cfg.api_endpoint,
    project_slug: cfg.project_slug,
    default_branch: cfg.default_branch,
    auth_token_ref: cfg.auth_token_ref,
    trigger_events: [...(cfg.trigger_events || [])],
    extra_config: { ...(cfg.extra_config || {}) },
    enabled: cfg.enabled
  }
  triggerEventsInput.value = (cfg.trigger_events || []).join(', ')
  extraConfigInput.value = JSON.stringify(cfg.extra_config || {}, null, 2)
  errorMsg.value = ''
  showConfigModal.value = true
}

function parseTriggerEvents(): string[] {
  return triggerEventsInput.value
    .split(',')
    .map(s => s.trim())
    .filter(s => s.length > 0)
}

function parseExtraConfig(): Record<string, unknown> {
  try {
    const parsed = JSON.parse(extraConfigInput.value || '{}')
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>
    }
  } catch {
    // ignore parse errors; will be caught by backend
  }
  return {}
}

async function handleSaveConfig() {
  if (!configForm.value.name) {
    errorMsg.value = 'Name is required'
    return
  }
  savingConfig.value = true
  errorMsg.value = ''
  try {
    const events = parseTriggerEvents()
    const extra = parseExtraConfig()
    if (editingConfig.value) {
      const update: CICDConfigUpdate = {
        name: configForm.value.name,
        provider: configForm.value.provider,
        api_endpoint: configForm.value.api_endpoint,
        project_slug: configForm.value.project_slug,
        default_branch: configForm.value.default_branch,
        auth_token_ref: configForm.value.auth_token_ref,
        trigger_events: events,
        extra_config: extra,
        enabled: configForm.value.enabled
      }
      await cicdConfigApi.update(workspaceId.value, editingConfig.value.id, update)
    } else {
      const create: CICDConfigCreate = {
        ...configForm.value,
        trigger_events: events,
        extra_config: extra
      }
      await cicdConfigApi.create(workspaceId.value, create)
    }
    showConfigModal.value = false
    await loadConfigs()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to save config'
  } finally {
    savingConfig.value = false
  }
}

async function handleDeleteConfig(cfg: CICDConfig) {
  if (!confirm(`Delete config "${cfg.name}"? Existing build records are retained.`)) return
  try {
    await cicdConfigApi.delete(workspaceId.value, cfg.id)
    await loadConfigs()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to delete config'
  }
}

// ============ Build modal handlers ============

function openTriggerModal() {
  if (configs.value.length === 0) {
    errorMsg.value = 'Create a CI/CD config before triggering a build'
    activeTab.value = 'configs'
    return
  }
  triggerForm.value = emptyTrigger()
  triggerForm.value.cicd_config_id = configs.value[0].id
  triggerForm.value.branch = configs.value[0].default_branch || 'main'
  triggerProjectIdInput.value = ''
  triggerIssueIdInput.value = ''
  triggerAgentTaskIdInput.value = ''
  errorMsg.value = ''
  showTriggerModal.value = true
}

function onTriggerConfigChange() {
  const cfg = configs.value.find(c => c.id === triggerForm.value.cicd_config_id)
  if (cfg && !triggerForm.value.branch) {
    triggerForm.value.branch = cfg.default_branch || 'main'
  }
}

async function handleTrigger() {
  if (!triggerForm.value.cicd_config_id) {
    errorMsg.value = 'Select a CI/CD config'
    return
  }
  triggering.value = true
  errorMsg.value = ''
  try {
    const payload: BuildTriggerRequest = {
      ...triggerForm.value,
      project_id: triggerProjectIdInput.value ? Number(triggerProjectIdInput.value) : undefined,
      issue_id: triggerIssueIdInput.value ? Number(triggerIssueIdInput.value) : undefined,
      agent_task_id: triggerAgentTaskIdInput.value ? Number(triggerAgentTaskIdInput.value) : undefined
    }
    await cicdBuildApi.trigger(workspaceId.value, payload)
    showTriggerModal.value = false
    await loadBuilds()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to trigger build'
  } finally {
    triggering.value = false
  }
}

async function handleCancelBuild(buildId: number) {
  if (!confirm('Cancel this build?')) return
  cancelingId.value = buildId
  try {
    await cicdBuildApi.cancel(workspaceId.value, buildId)
    await loadBuilds()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to cancel build'
  } finally {
    cancelingId.value = null
  }
}

async function handleDeleteBuild(buildId: number) {
  if (!confirm('Delete this build record?')) return
  try {
    await cicdBuildApi.delete(workspaceId.value, buildId)
    await loadBuilds()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to delete build'
  }
}

function openBuildDetailModal(build: BuildRecord) {
  selectedBuild.value = build
  showBuildDetailModal.value = true
}

// ============ SSE ============

function upsertBuild(build: BuildRecord) {
  const idx = builds.value.findIndex(b => b.id === build.id)
  if (idx >= 0) {
    builds.value[idx] = build
    if (selectedBuild.value && selectedBuild.value.id === build.id) {
      selectedBuild.value = build
    }
  } else {
    builds.value.unshift(build)
  }
}

function upsertConfig(cfg: CICDConfig) {
  const idx = configs.value.findIndex(c => c.id === cfg.id)
  if (idx >= 0) {
    configs.value[idx] = cfg
  } else {
    configs.value.unshift(cfg)
  }
}

// ============ Helpers ============

function statusBadgeClass(status: BuildStatus): string {
  switch (status) {
    case 'success': return 'bg-green-100 text-green-700'
    case 'failed': return 'bg-red-100 text-red-700'
    case 'cancelled': return 'bg-amber-100 text-amber-700'
    case 'pending': return 'bg-gray-100 text-gray-700'
    case 'queued': return 'bg-blue-100 text-blue-700'
    case 'running': return 'bg-indigo-100 text-indigo-700'
    case 'unknown': return 'bg-gray-100 text-gray-500'
    default: return 'bg-gray-100 text-gray-700'
  }
}

function triggerBadgeClass(trigger: BuildTrigger): string {
  switch (trigger) {
    case 'manual': return 'bg-gray-100 text-gray-700'
    case 'push': return 'bg-blue-100 text-blue-700'
    case 'pull_request': return 'bg-purple-100 text-purple-700'
    case 'schedule': return 'bg-cyan-100 text-cyan-700'
    case 'agent': return 'bg-indigo-100 text-indigo-700'
    case 'webhook': return 'bg-teal-100 text-teal-700'
    default: return 'bg-gray-100 text-gray-600'
  }
}

function providerBadgeClass(provider: CICDProvider): string {
  switch (provider) {
    case 'github_actions': return 'bg-gray-900 text-white'
    case 'gitlab_ci': return 'bg-orange-100 text-orange-700'
    case 'jenkins': return 'bg-red-100 text-red-700'
    case 'generic': return 'bg-slate-100 text-slate-700'
    default: return 'bg-gray-100 text-gray-600'
  }
}

function stageStatusClass(status: BuildStage['status']): string {
  switch (status) {
    case 'success': return 'bg-green-500'
    case 'running': return 'bg-indigo-500 animate-pulse'
    case 'failed': return 'bg-red-500'
    case 'pending': return 'bg-gray-300'
    case 'skipped': return 'bg-gray-200'
    default: return 'bg-gray-300'
  }
}

function stageStatusBadgeClass(status: BuildStage['status']): string {
  switch (status) {
    case 'success': return 'bg-green-100 text-green-700'
    case 'running': return 'bg-indigo-100 text-indigo-700'
    case 'failed': return 'bg-red-100 text-red-700'
    case 'pending': return 'bg-gray-100 text-gray-500'
    case 'skipped': return 'bg-gray-100 text-gray-400'
    default: return 'bg-gray-100 text-gray-500'
  }
}

function isTerminal(status: BuildStatus): boolean {
  return status === 'success' || status === 'failed' || status === 'cancelled'
}

function isRunning(status: BuildStatus): boolean {
  return status === 'pending' || status === 'queued' || status === 'running'
}

function formatTime(iso?: string): string {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function formatDuration(ms: number): string {
  if (!ms || ms <= 0) return '—'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  const m = Math.floor(ms / 60000)
  const s = Math.floor((ms % 60000) / 1000)
  return `${m}m ${s}s`
}

function shortSha(sha: string): string {
  if (!sha) return '—'
  return sha.length > 8 ? sha.slice(0, 8) : sha
}

function configName(build: BuildRecord): string {
  return build.cicd_config_name || `#${build.cicd_config_id}`
}

// ============ Lifecycle ============

onMounted(async () => {
  await Promise.all([loadConfigs(), loadBuilds()])
  sseCleanup = onEvent((event, data) => {
    if (!event.startsWith('cicd_')) return
    if (!data || typeof data !== 'object' || !data.id) return
    if (event.startsWith('cicd_build.')) {
      upsertBuild(data as BuildRecord)
    } else if (event.startsWith('cicd_config.')) {
      upsertConfig(data as CICDConfig)
    }
  })
})

onUnmounted(() => {
  if (sseCleanup) sseCleanup()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div>
          <h1 class="text-xl font-semibold text-gray-800">🚀 CI/CD Manager</h1>
          <p class="text-sm text-gray-500 mt-0.5">
            Configure pipelines &amp; monitor builds (P4-005)
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button
            class="px-3 py-1.5 text-sm rounded-lg border border-gray-300 bg-white hover:bg-gray-50"
            @click="activeTab === 'configs' ? loadConfigs() : loadBuilds()"
          >
            Refresh
          </button>
          <button
            v-if="activeTab === 'configs'"
            class="px-3 py-1.5 text-sm rounded-lg bg-sky-600 text-white hover:bg-sky-700"
            @click="openCreateConfigModal"
          >
            New Config
          </button>
          <button
            v-else
            class="px-3 py-1.5 text-sm rounded-lg bg-sky-600 text-white hover:bg-sky-700"
            @click="openTriggerModal"
          >
            Trigger Build
          </button>
        </div>
      </div>

      <!-- Tabs -->
      <div class="flex gap-1 mt-4 border-b border-gray-200">
        <button
          class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
          :class="activeTab === 'configs'
            ? 'border-sky-600 text-sky-700'
            : 'border-transparent text-gray-500 hover:text-gray-700'"
          @click="activeTab = 'configs'"
        >
          Configs ({{ configs.length }})
        </button>
        <button
          class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
          :class="activeTab === 'builds'
            ? 'border-sky-600 text-sky-700'
            : 'border-transparent text-gray-500 hover:text-gray-700'"
          @click="activeTab = 'builds'"
        >
          Builds ({{ builds.length }})
        </button>
      </div>
    </header>

    <main class="p-6">
      <div class="max-w-7xl mx-auto space-y-4">
        <div v-if="errorMsg" class="bg-red-50 border border-red-200 rounded-xl p-3 text-sm text-red-700">
          {{ errorMsg }}
        </div>

        <!-- ============ Configs tab ============ -->
        <template v-if="activeTab === 'configs'">
          <div v-if="loadingConfigs" class="text-sm text-gray-400 py-8 text-center">
            Loading CI/CD configs…
          </div>
          <div v-else-if="configs.length === 0" class="text-sm text-gray-400 py-12 text-center border border-gray-200 rounded-xl bg-white">
            No CI/CD configs yet. Click "New Config" to register a pipeline.
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="cfg in configs"
              :key="cfg.id"
              class="bg-white border border-gray-200 rounded-xl p-4 hover:border-sky-300 transition-colors"
            >
              <div class="flex items-start justify-between gap-3 flex-wrap">
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="font-medium text-gray-900 truncate">{{ cfg.name }}</span>
                    <span class="text-xs px-2 py-0.5 rounded-full" :class="providerBadgeClass(cfg.provider)">
                      {{ cfg.provider }}
                    </span>
                    <span
                      class="text-xs px-2 py-0.5 rounded-full"
                      :class="cfg.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'"
                    >
                      {{ cfg.enabled ? 'enabled' : 'disabled' }}
                    </span>
                  </div>
                  <div class="text-xs text-gray-500 mt-1 flex flex-wrap gap-x-3 gap-y-1">
                    <span v-if="cfg.project_slug">repo: {{ cfg.project_slug }}</span>
                    <span>branch: {{ cfg.default_branch || 'main' }}</span>
                    <span v-if="cfg.api_endpoint">endpoint: <code class="bg-gray-100 px-1 rounded">{{ cfg.api_endpoint }}</code></span>
                    <span v-if="cfg.auth_token_ref">token: <code class="bg-gray-100 px-1 rounded">{{ cfg.auth_token_ref }}</code></span>
                  </div>
                  <div v-if="cfg.trigger_events && cfg.trigger_events.length > 0" class="mt-2 flex flex-wrap gap-1">
                    <span
                      v-for="ev in cfg.trigger_events"
                      :key="ev"
                      class="text-xs px-2 py-0.5 rounded-full bg-blue-50 text-blue-700 border border-blue-200"
                    >
                      {{ ev }}
                    </span>
                  </div>
                </div>
                <div class="flex items-center gap-2">
                  <button
                    class="text-xs px-2 py-1 rounded border border-sky-300 text-sky-700 hover:bg-sky-50"
                    @click="openEditConfigModal(cfg)"
                  >
                    Edit
                  </button>
                  <button
                    class="text-xs px-2 py-1 rounded border border-red-300 text-red-700 hover:bg-red-50"
                    @click="handleDeleteConfig(cfg)"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          </div>
        </template>

        <!-- ============ Builds tab ============ -->
        <template v-else>
          <!-- Status filter chips -->
          <div class="flex flex-wrap gap-2">
            <button
              v-for="f in STATUS_FILTERS"
              :key="f.value || 'all'"
              class="px-3 py-1 text-xs rounded-full border transition-colors"
              :class="activeStatusFilter === f.value
                ? 'bg-sky-600 text-white border-sky-600'
                : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50'"
              @click="activeStatusFilter = f.value"
            >
              {{ f.label }}
            </button>
          </div>

          <div v-if="loadingBuilds" class="text-sm text-gray-400 py-8 text-center">
            Loading builds…
          </div>
          <div v-else-if="filteredBuilds.length === 0" class="text-sm text-gray-400 py-12 text-center border border-gray-200 rounded-xl bg-white">
            No builds match this filter. Click "Trigger Build" to start a new run.
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="build in filteredBuilds"
              :key="build.id"
              class="bg-white border border-gray-200 rounded-xl p-4 hover:border-sky-300 cursor-pointer transition-colors"
              @click="openBuildDetailModal(build)"
            >
              <div class="flex items-start justify-between gap-3 flex-wrap">
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="font-medium text-gray-900 truncate">{{ configName(build) }}</span>
                    <span class="text-xs px-2 py-0.5 rounded-full" :class="statusBadgeClass(build.status)">
                      {{ build.status }}
                    </span>
                    <span class="text-xs px-2 py-0.5 rounded-full" :class="triggerBadgeClass(build.trigger)">
                      {{ build.trigger }}
                    </span>
                    <span v-if="build.current_stage" class="text-xs text-indigo-600">
                      stage: {{ build.current_stage }}
                    </span>
                  </div>
                  <div class="text-xs text-gray-500 mt-1 flex flex-wrap gap-x-3 gap-y-1">
                    <span>branch: <code class="bg-gray-100 px-1 rounded">{{ build.branch || '—' }}</code></span>
                    <span>commit: <code class="bg-gray-100 px-1 rounded">{{ shortSha(build.commit_sha) }}</code></span>
                    <span v-if="build.duration_ms > 0">duration: {{ formatDuration(build.duration_ms) }}</span>
                    <span>created: {{ formatTime(build.created_at) }}</span>
                  </div>
                  <div v-if="build.error_message" class="text-xs text-red-600 mt-1 truncate">
                    error: {{ build.error_message }}
                  </div>
                </div>
                <div class="flex items-center gap-2" @click.stop>
                  <a
                    v-if="build.build_url"
                    :href="build.build_url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-xs px-2 py-1 rounded border border-gray-300 text-gray-700 hover:bg-gray-50"
                  >
                    Logs ↗
                  </a>
                  <button
                    v-if="isRunning(build.status)"
                    class="text-xs px-2 py-1 rounded border border-amber-300 text-amber-700 hover:bg-amber-50 disabled:opacity-50"
                    :disabled="cancelingId === build.id"
                    @click="handleCancelBuild(build.id)"
                  >
                    {{ cancelingId === build.id ? 'Cancelling…' : 'Cancel' }}
                  </button>
                  <button
                    v-if="isTerminal(build.status)"
                    class="text-xs px-2 py-1 rounded border border-red-300 text-red-700 hover:bg-red-50"
                    @click="handleDeleteBuild(build.id)"
                  >
                    Delete
                  </button>
                </div>
              </div>
              <!-- Progress + stage timeline -->
              <div v-if="isRunning(build.status)" class="mt-3">
                <div class="h-1.5 bg-gray-100 rounded-full overflow-hidden">
                  <div
                    class="h-full bg-sky-500 transition-all duration-300"
                    :style="{ width: `${build.progress}%` }"
                  />
                </div>
                <div class="text-xs text-gray-400 mt-1">{{ build.progress }}%</div>
              </div>
              <div v-if="build.stages && build.stages.length > 0" class="mt-3 flex items-center gap-1">
                <div
                  v-for="(stage, idx) in build.stages"
                  :key="idx"
                  class="flex items-center gap-1"
                >
                  <div
                    class="h-2 w-10 rounded-full"
                    :class="stageStatusClass(stage.status)"
                    :title="`${stage.name}: ${stage.status}`"
                  />
                  <span class="text-xs text-gray-500">{{ stage.name }}</span>
                </div>
              </div>
            </div>
          </div>
        </template>
      </div>
    </main>

    <!-- ============ Config create/edit modal ============ -->
    <div v-if="showConfigModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" @click.self="showConfigModal = false">
      <div class="bg-white rounded-xl shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="px-6 py-4 border-b border-gray-200">
          <h2 class="text-lg font-semibold text-gray-800">
            {{ editingConfig ? 'Edit Config' : 'New CI/CD Config' }}
          </h2>
          <p class="text-xs text-gray-500 mt-0.5">
            Configure a pipeline provider. Tokens are stored as opaque references, never as raw credentials.
          </p>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div v-if="errorMsg" class="bg-red-50 border border-red-200 rounded p-2 text-sm text-red-700">
            {{ errorMsg }}
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
            <input v-model="configForm.name" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="e.g. Backend CI" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Provider</label>
              <select v-model="configForm.provider" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm">
                <option v-for="p in PROVIDERS" :key="p" :value="p">{{ p }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Default Branch</label>
              <input v-model="configForm.default_branch" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="main" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">API Endpoint</label>
            <input v-model="configForm.api_endpoint" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="https://api.github.com / https://gitlab.com/api/v4 / http://jenkins…" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Project Slug</label>
            <input v-model="configForm.project_slug" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="owner/repo (GitHub) · path/to/project (GitLab) · job-name (Jenkins)" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Auth Token Reference</label>
            <input v-model="configForm.auth_token_ref" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="env-var name or secret key (e.g. GITHUB_TOKEN)" />
            <p class="text-xs text-gray-500 mt-1">Never paste the raw token here — only the reference name.</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Trigger Events (comma-separated)</label>
            <input v-model="triggerEventsInput" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="push, pull_request, manual" />
            <div class="mt-1 flex flex-wrap gap-1">
              <button
                v-for="ev in TRIGGER_EVENT_OPTIONS"
                :key="ev"
                type="button"
                class="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-600 hover:bg-gray-200"
                @click="triggerEventsInput = (triggerEventsInput ? triggerEventsInput + ', ' : '') + ev"
              >
                + {{ ev }}
              </button>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Extra Config (JSON)</label>
            <textarea v-model="extraConfigInput" rows="3" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono" placeholder="{}" />
          </div>
          <label class="flex items-center gap-2">
            <input type="checkbox" v-model="configForm.enabled" class="rounded border-gray-300" />
            <span class="text-sm text-gray-700">Enabled</span>
          </label>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <button class="px-4 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50" @click="showConfigModal = false">
            Cancel
          </button>
          <button
            class="px-4 py-2 text-sm rounded-lg bg-sky-600 text-white hover:bg-sky-700 disabled:opacity-50"
            :disabled="savingConfig || !configForm.name"
            @click="handleSaveConfig"
          >
            {{ savingConfig ? 'Saving…' : (editingConfig ? 'Save' : 'Create') }}
          </button>
        </div>
      </div>
    </div>

    <!-- ============ Trigger build modal ============ -->
    <div v-if="showTriggerModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" @click.self="showTriggerModal = false">
      <div class="bg-white rounded-xl shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="px-6 py-4 border-b border-gray-200">
          <h2 class="text-lg font-semibold text-gray-800">Trigger Build</h2>
          <p class="text-xs text-gray-500 mt-0.5">
            The build runs asynchronously. Watch the Builds tab for live progress.
          </p>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div v-if="errorMsg" class="bg-red-50 border border-red-200 rounded p-2 text-sm text-red-700">
            {{ errorMsg }}
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">CI/CD Config *</label>
            <select
              v-model="triggerForm.cicd_config_id"
              class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
              @change="onTriggerConfigChange"
            >
              <option v-for="cfg in configs" :key="cfg.id" :value="cfg.id">
                {{ cfg.name }} ({{ cfg.provider }})
              </option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Branch</label>
              <input v-model="triggerForm.branch" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="defaults to config default branch" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Commit SHA</label>
              <input v-model="triggerForm.commit_sha" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="optional, full or short SHA" />
            </div>
          </div>
          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Trigger</label>
              <select v-model="triggerForm.trigger" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm">
                <option value="manual">manual</option>
                <option value="push">push</option>
                <option value="pull_request">pull_request</option>
                <option value="schedule">schedule</option>
                <option value="agent">agent</option>
                <option value="webhook">webhook</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Project ID</label>
              <input type="number" v-model="triggerProjectIdInput" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="—" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Issue ID</label>
              <input type="number" v-model="triggerIssueIdInput" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="—" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Agent Task ID (optional)</label>
            <input type="number" v-model="triggerAgentTaskIdInput" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="—" />
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <button class="px-4 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50" @click="showTriggerModal = false">
            Cancel
          </button>
          <button
            class="px-4 py-2 text-sm rounded-lg bg-sky-600 text-white hover:bg-sky-700 disabled:opacity-50"
            :disabled="triggering || !triggerForm.cicd_config_id"
            @click="handleTrigger"
          >
            {{ triggering ? 'Triggering…' : 'Trigger' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ============ Build detail modal ============ -->
    <div v-if="showBuildDetailModal && selectedBuild" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" @click.self="showBuildDetailModal = false">
      <div class="bg-white rounded-xl shadow-xl max-w-3xl w-full max-h-[90vh] overflow-y-auto">
        <div class="px-6 py-4 border-b border-gray-200 flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-800">{{ configName(selectedBuild) }}</h2>
            <div class="text-xs text-gray-500 mt-1">
              Build #{{ selectedBuild.id }} · {{ selectedBuild.status }} · {{ selectedBuild.trigger }}
            </div>
          </div>
          <button class="text-gray-400 hover:text-gray-600" @click="showBuildDetailModal = false">✕</button>
        </div>
        <div class="px-6 py-4 space-y-4 text-sm">
          <div v-if="selectedBuild.error_message" class="bg-red-50 border border-red-200 rounded p-3 text-red-700">
            {{ selectedBuild.error_message }}
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <div class="text-xs text-gray-500 uppercase">Branch</div>
              <code class="text-xs bg-gray-100 px-1 rounded">{{ selectedBuild.branch || '—' }}</code>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Commit</div>
              <code class="text-xs bg-gray-100 px-1 rounded">{{ selectedBuild.commit_sha || '—' }}</code>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Started</div>
              <span>{{ formatTime(selectedBuild.started_at) }}</span>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Completed</div>
              <span>{{ formatTime(selectedBuild.completed_at) }}</span>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Duration</div>
              <span>{{ formatDuration(selectedBuild.duration_ms) }}</span>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">External ID</div>
              <code class="text-xs bg-gray-100 px-1 rounded">{{ selectedBuild.external_build_id || '—' }}</code>
            </div>
          </div>

          <!-- Stage timeline -->
          <div v-if="selectedBuild.stages && selectedBuild.stages.length > 0">
            <div class="text-xs text-gray-500 uppercase mb-2">Stages</div>
            <div class="space-y-2">
              <div
                v-for="(stage, idx) in selectedBuild.stages"
                :key="idx"
                class="border border-gray-200 rounded-lg p-3"
              >
                <div class="flex items-center justify-between gap-2">
                  <div class="flex items-center gap-2">
                    <span class="text-xs px-2 py-0.5 rounded-full" :class="stageStatusBadgeClass(stage.status)">
                      {{ stage.status }}
                    </span>
                    <span class="font-medium text-gray-900 text-sm">{{ stage.name }}</span>
                  </div>
                  <span class="text-xs text-gray-400">{{ formatDuration(stage.duration_ms) }}</span>
                </div>
                <div class="text-xs text-gray-500 mt-1 flex gap-3">
                  <span>started: {{ formatTime(stage.started_at) }}</span>
                  <span>completed: {{ formatTime(stage.completed_at) }}</span>
                </div>
                <a
                  v-if="stage.log_url"
                  :href="stage.log_url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-xs text-sky-600 hover:underline mt-1 inline-block"
                >
                  View logs ↗
                </a>
              </div>
            </div>
          </div>

          <div v-if="selectedBuild.build_url">
            <a
              :href="selectedBuild.build_url"
              target="_blank"
              rel="noopener noreferrer"
              class="text-sm text-sky-600 hover:underline"
            >
              Open build in provider ↗
            </a>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-between gap-2">
          <div class="flex gap-2">
            <button
              v-if="isRunning(selectedBuild.status)"
              class="px-4 py-2 text-sm rounded-lg border border-amber-300 text-amber-700 hover:bg-amber-50 disabled:opacity-50"
              :disabled="cancelingId === selectedBuild.id"
              @click="handleCancelBuild(selectedBuild.id)"
            >
              {{ cancelingId === selectedBuild.id ? 'Cancelling…' : 'Cancel Build' }}
            </button>
            <button
              v-if="isTerminal(selectedBuild.status)"
              class="px-4 py-2 text-sm rounded-lg border border-red-300 text-red-700 hover:bg-red-50"
              @click="handleDeleteBuild(selectedBuild.id); showBuildDetailModal = false"
            >
              Delete
            </button>
          </div>
          <button class="px-4 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50" @click="showBuildDetailModal = false">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
