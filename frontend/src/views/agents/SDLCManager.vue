<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { useSSE } from '@/composables/useSSE'
import {
  sdlcWorkflowApi,
  sdlcStageApi,
  SDLC_CANONICAL_STAGES,
  type SDLCWorkflow,
  type SDLCWorkflowCreate,
  type SDLCWorkflowStatus,
  type SDLCStage,
  type SDLCStageStatus
} from '@/api/sdlc'

const { getWorkspaceId } = useWorkspaceId()
const { onEvent } = useSSE()

const workspaceId = ref(0)
const workflows = ref<SDLCWorkflow[]>([])
const loading = ref(true)
const errorMsg = ref('')

// Detail panel
const selectedWorkflow = ref<SDLCWorkflow | null>(null)
const selectedStages = ref<SDLCStage[]>([])
const loadingDetail = ref(false)
const cancelingId = ref<number | null>(null)
const retryingStageId = ref<number | null>(null)

// Create modal
const showCreateModal = ref(false)
const creating = ref(false)
const createForm = ref<SDLCWorkflowCreate>(emptyCreate())
const projectIdInput = ref('')
const squadIdInput = ref('')
const configInput = ref('{}')

function emptyCreate(): SDLCWorkflowCreate {
  return {
    title: '',
    requirement: '',
    stages: [...SDLC_CANONICAL_STAGES],
    fail_fast: true
  }
}

// Status filter chips
const STATUS_FILTERS: { value: SDLCWorkflowStatus | ''; label: string }[] = [
  { value: '', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'running', label: 'Running' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'partial_failed', label: 'Partial' },
  { value: 'cancelled', label: 'Cancelled' }
]
const activeStatusFilter = ref<SDLCWorkflowStatus | ''>('')

const filteredWorkflows = computed<SDLCWorkflow[]>(() => {
  if (!activeStatusFilter.value) return workflows.value
  return workflows.value.filter(w => w.status === activeStatusFilter.value)
})

// Stage metadata lookup (key → display name from canonical list, fallback to key)
const STAGE_DISPLAY: Record<string, string> = {
  requirement_analysis: 'Requirement Analysis',
  requirement_design: 'Requirement Design',
  dispatch_feature: 'Dispatch Feature',
  feature_design: 'Feature Design',
  breakdown_us: 'Breakdown User Stories',
  sprint_planning: 'Sprint Planning',
  development: 'Development',
  code_review: 'Code Review',
  us_testing: 'US Testing',
  fe_testing: 'FE Testing',
  deploy: 'Deploy'
}

function stageDisplayName(key: string): string {
  return STAGE_DISPLAY[key] || key.replace(/_/g, ' ')
}

let sseCleanup: (() => void) | null = null

// ============ Data loading ============

async function loadWorkflows() {
  loading.value = true
  errorMsg.value = ''
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    workspaceId.value = wsId
    const res = await sdlcWorkflowApi.list(wsId)
    workflows.value = res || []
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to load SDLC workflows'
    workflows.value = []
  } finally {
    loading.value = false
  }
}

async function loadDetail(workflow: SDLCWorkflow) {
  selectedWorkflow.value = workflow
  selectedStages.value = workflow.stages || []
  showDetail.value = true
  if (!workspaceId.value) return
  loadingDetail.value = true
  try {
    // Re-fetch full detail (with stages) for freshness
    const fresh = await sdlcWorkflowApi.get(workspaceId.value, workflow.id)
    selectedWorkflow.value = fresh
    selectedStages.value = fresh.stages || (await sdlcStageApi.list(workspaceId.value, workflow.id).catch(() => []))
  } catch (err: any) {
    // Fall back to list view data
    selectedStages.value = await sdlcStageApi.list(workspaceId.value, workflow.id).catch(() => [])
  } finally {
    loadingDetail.value = false
  }
}

const showDetail = ref(false)

function closeDetail() {
  showDetail.value = false
  selectedWorkflow.value = null
  selectedStages.value = []
}

// ============ Create modal ============

function openCreateModal() {
  createForm.value = emptyCreate()
  projectIdInput.value = ''
  squadIdInput.value = ''
  configInput.value = '{}'
  errorMsg.value = ''
  showCreateModal.value = true
}

function toggleStage(key: string) {
  const stages = createForm.value.stages || []
  const idx = stages.indexOf(key)
  if (idx >= 0) {
    // Preserve canonical order
    createForm.value.stages = SDLC_CANONICAL_STAGES.filter(k => k !== key && stages.includes(k))
  } else {
    createForm.value.stages = SDLC_CANONICAL_STAGES.filter(k => stages.includes(k) || k === key)
  }
}

function isStageSelected(key: string): boolean {
  return (createForm.value.stages || []).includes(key)
}

function parseConfig(): Record<string, unknown> {
  try {
    const parsed = JSON.parse(configInput.value || '{}')
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>
    }
  } catch {
    // ignore parse errors; backend will validate
  }
  return {}
}

async function handleCreate() {
  if (!createForm.value.title) {
    errorMsg.value = 'Title is required'
    return
  }
  if (!createForm.value.requirement) {
    errorMsg.value = 'Requirement is required'
    return
  }
  creating.value = true
  errorMsg.value = ''
  try {
    const payload: SDLCWorkflowCreate = {
      ...createForm.value,
      project_id: projectIdInput.value ? Number(projectIdInput.value) : undefined,
      squad_id: squadIdInput.value ? Number(squadIdInput.value) : undefined,
      config: parseConfig()
    }
    await sdlcWorkflowApi.create(workspaceId.value, payload)
    showCreateModal.value = false
    await loadWorkflows()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to create workflow'
  } finally {
    creating.value = false
  }
}

// ============ Workflow actions ============

async function handleCancel(workflowId: number) {
  if (!confirm('Cancel this workflow? Pending/running stages will be marked skipped.')) return
  cancelingId.value = workflowId
  try {
    const updated = await sdlcWorkflowApi.cancel(workspaceId.value, workflowId)
    upsertWorkflow(updated)
    if (selectedWorkflow.value && selectedWorkflow.value.id === workflowId) {
      selectedWorkflow.value = updated
      selectedStages.value = updated.stages || selectedStages.value
    }
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to cancel workflow'
  } finally {
    cancelingId.value = null
  }
}

async function handleDelete(workflowId: number) {
  if (!confirm('Delete this workflow and all its stage records?')) return
  try {
    await sdlcWorkflowApi.delete(workspaceId.value, workflowId)
    workflows.value = workflows.value.filter(w => w.id !== workflowId)
    if (selectedWorkflow.value && selectedWorkflow.value.id === workflowId) {
      closeDetail()
    }
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to delete workflow'
  }
}

async function handleRetryStage(workflowId: number, stageId: number) {
  retryingStageId.value = stageId
  try {
    const updated = await sdlcWorkflowApi.retry(workspaceId.value, workflowId, { stage_id: stageId })
    upsertWorkflow(updated)
    if (selectedWorkflow.value && selectedWorkflow.value.id === workflowId) {
      selectedWorkflow.value = updated
      selectedStages.value = updated.stages || selectedStages.value
    }
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to retry from stage'
  } finally {
    retryingStageId.value = null
  }
}

// ============ SSE ============

function upsertWorkflow(wf: SDLCWorkflow) {
  const idx = workflows.value.findIndex(w => w.id === wf.id)
  if (idx >= 0) {
    workflows.value[idx] = wf
  } else {
    workflows.value.unshift(wf)
  }
  if (selectedWorkflow.value && selectedWorkflow.value.id === wf.id) {
    selectedWorkflow.value = wf
    selectedStages.value = wf.stages || selectedStages.value
  }
}

function upsertStage(stage: SDLCStage) {
  if (!selectedWorkflow.value || selectedWorkflow.value.id !== stage.workflow_id) return
  const idx = selectedStages.value.findIndex(s => s.id === stage.id)
  if (idx >= 0) {
    selectedStages.value[idx] = stage
  } else {
    selectedStages.value.push(stage)
    selectedStages.value.sort((a, b) => a.order - b.order)
  }
}

// ============ Helpers ============

function statusBadgeClass(status: SDLCWorkflowStatus): string {
  switch (status) {
    case 'completed': return 'bg-green-100 text-green-700'
    case 'failed': return 'bg-red-100 text-red-700'
    case 'partial_failed': return 'bg-amber-100 text-amber-700'
    case 'cancelled': return 'bg-gray-100 text-gray-500'
    case 'running': return 'bg-indigo-100 text-indigo-700'
    case 'pending': return 'bg-blue-100 text-blue-700'
    default: return 'bg-gray-100 text-gray-700'
  }
}

function stageStatusBadgeClass(status: SDLCStageStatus): string {
  switch (status) {
    case 'success': return 'bg-green-100 text-green-700'
    case 'running': return 'bg-indigo-100 text-indigo-700'
    case 'failed': return 'bg-red-100 text-red-700'
    case 'pending': return 'bg-gray-100 text-gray-500'
    case 'skipped': return 'bg-gray-100 text-gray-400'
    default: return 'bg-gray-100 text-gray-500'
  }
}

function stageDotClass(status: SDLCStageStatus): string {
  switch (status) {
    case 'success': return 'bg-green-500'
    case 'running': return 'bg-indigo-500 animate-pulse'
    case 'failed': return 'bg-red-500'
    case 'pending': return 'bg-gray-300'
    case 'skipped': return 'bg-gray-200'
    default: return 'bg-gray-300'
  }
}

function isTerminal(status: SDLCWorkflowStatus): boolean {
  return status === 'completed' || status === 'failed' || status === 'partial_failed' || status === 'cancelled'
}

function isRunning(status: SDLCWorkflowStatus): boolean {
  return status === 'pending' || status === 'running'
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

function formatJson(value: unknown): string {
  if (!value) return '—'
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

function artifactEntries(artifacts: Record<string, unknown>): { key: string; value: unknown }[] {
  if (!artifacts || typeof artifacts !== 'object') return []
  return Object.entries(artifacts).map(([key, value]) => ({ key, value }))
}

// ============ Lifecycle ============

onMounted(async () => {
  await loadWorkflows()
  sseCleanup = onEvent((event, data) => {
    if (!event.startsWith('sdlc_')) return
    if (!data || typeof data !== 'object') return
    if (event.startsWith('sdlc_workflow.')) {
      if (!data.id) return
      upsertWorkflow(data as SDLCWorkflow)
    } else if (event.startsWith('sdlc_stage.')) {
      if (!data.id || !data.workflow_id) return
      upsertStage(data as SDLCStage)
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
          <h1 class="text-xl font-semibold text-gray-800">🧩 SDLC Orchestrator</h1>
          <p class="text-sm text-gray-500 mt-0.5">
            End-to-end delivery pipeline · 11 canonical stages (P4-006)
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button
            class="px-3 py-1.5 text-sm rounded-lg border border-gray-300 bg-white hover:bg-gray-50"
            @click="loadWorkflows"
          >
            Refresh
          </button>
          <button
            class="px-3 py-1.5 text-sm rounded-lg bg-violet-600 text-white hover:bg-violet-700"
            @click="openCreateModal"
          >
            New Workflow
          </button>
        </div>
      </div>
    </header>

    <main class="p-6">
      <div class="max-w-7xl mx-auto space-y-4">
        <div v-if="errorMsg" class="bg-red-50 border border-red-200 rounded-xl p-3 text-sm text-red-700">
          {{ errorMsg }}
        </div>

        <!-- Status filter chips -->
        <div class="flex flex-wrap gap-2">
          <button
            v-for="f in STATUS_FILTERS"
            :key="f.value || 'all'"
            class="px-3 py-1 text-xs rounded-full border transition-colors"
            :class="activeStatusFilter === f.value
              ? 'bg-violet-600 text-white border-violet-600'
              : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50'"
            @click="activeStatusFilter = f.value"
          >
            {{ f.label }}
          </button>
        </div>

        <div v-if="loading" class="text-sm text-gray-400 py-8 text-center">
          Loading SDLC workflows…
        </div>
        <div v-else-if="filteredWorkflows.length === 0" class="text-sm text-gray-400 py-12 text-center border border-gray-200 rounded-xl bg-white">
          No workflows match this filter. Click "New Workflow" to orchestrate a delivery pipeline.
        </div>
        <div v-else class="space-y-3">
          <div
            v-for="wf in filteredWorkflows"
            :key="wf.id"
            class="bg-white border border-gray-200 rounded-xl p-4 hover:border-violet-300 cursor-pointer transition-colors"
            @click="loadDetail(wf)"
          >
            <div class="flex items-start justify-between gap-3 flex-wrap">
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="font-medium text-gray-900 truncate">{{ wf.title }}</span>
                  <span class="text-xs px-2 py-0.5 rounded-full" :class="statusBadgeClass(wf.status)">
                    {{ wf.status }}
                  </span>
                  <span v-if="wf.current_stage" class="text-xs text-indigo-600">
                    stage: {{ stageDisplayName(wf.current_stage) }}
                  </span>
                </div>
                <div class="text-xs text-gray-500 mt-1 line-clamp-2">{{ wf.requirement }}</div>
                <div class="text-xs text-gray-400 mt-1 flex flex-wrap gap-x-3 gap-y-1">
                  <span>created: {{ formatTime(wf.created_at) }}</span>
                  <span v-if="wf.completed_at">completed: {{ formatTime(wf.completed_at) }}</span>
                </div>
                <div v-if="wf.error_message" class="text-xs text-red-600 mt-1 truncate">
                  error: {{ wf.error_message }}
                </div>
              </div>
              <div class="flex items-center gap-2" @click.stop>
                <button
                  v-if="isRunning(wf.status)"
                  class="text-xs px-2 py-1 rounded border border-amber-300 text-amber-700 hover:bg-amber-50 disabled:opacity-50"
                  :disabled="cancelingId === wf.id"
                  @click="handleCancel(wf.id)"
                >
                  {{ cancelingId === wf.id ? 'Cancelling…' : 'Cancel' }}
                </button>
                <button
                  v-if="isTerminal(wf.status)"
                  class="text-xs px-2 py-1 rounded border border-red-300 text-red-700 hover:bg-red-50"
                  @click="handleDelete(wf.id)"
                >
                  Delete
                </button>
              </div>
            </div>

            <!-- Progress -->
            <div v-if="isRunning(wf.status)" class="mt-3">
              <div class="h-1.5 bg-gray-100 rounded-full overflow-hidden">
                <div
                  class="h-full bg-violet-500 transition-all duration-300"
                  :style="{ width: `${wf.progress}%` }"
                />
              </div>
              <div class="text-xs text-gray-400 mt-1">{{ wf.progress }}%</div>
            </div>

            <!-- Stage timeline (compact) -->
            <div v-if="wf.stages && wf.stages.length > 0" class="mt-3 flex items-center gap-1 flex-wrap">
              <div
                v-for="stage in wf.stages"
                :key="stage.id"
                class="flex items-center gap-1"
                :title="`${stageDisplayName(stage.key)}: ${stage.status}`"
              >
                <div class="h-2 w-8 rounded-full" :class="stageDotClass(stage.status)" />
                <span class="text-xs text-gray-500">{{ stageDisplayName(stage.key) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- ============ Create workflow modal ============ -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" @click.self="showCreateModal = false">
      <div class="bg-white rounded-xl shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="px-6 py-4 border-b border-gray-200">
          <h2 class="text-lg font-semibold text-gray-800">New SDLC Workflow</h2>
          <p class="text-xs text-gray-500 mt-0.5">
            Runs asynchronously. The workflow executes the selected stages in canonical order;
            watch the list for live progress via SSE.
          </p>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div v-if="errorMsg" class="bg-red-50 border border-red-200 rounded p-2 text-sm text-red-700">
            {{ errorMsg }}
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Title *</label>
            <input v-model="createForm.title" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="e.g. Payment Service v2 Delivery" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Requirement *</label>
            <textarea v-model="createForm.requirement" rows="4" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Describe the feature/requirement to be delivered through the SDLC pipeline…"></textarea>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Project ID</label>
              <input type="number" v-model="projectIdInput" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="—" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Squad ID</label>
              <input type="number" v-model="squadIdInput" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="—" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Stages <span class="text-xs text-gray-400">(selected stages run in canonical order)</span>
            </label>
            <div class="grid grid-cols-2 gap-1.5 mt-1">
              <label
                v-for="key in SDLC_CANONICAL_STAGES"
                :key="key"
                class="flex items-center gap-2 text-xs cursor-pointer"
              >
                <input
                  type="checkbox"
                  :checked="isStageSelected(key)"
                  class="rounded border-gray-300"
                  @change="toggleStage(key)"
                />
                <span class="text-gray-700">{{ stageDisplayName(key) }}</span>
              </label>
            </div>
            <p class="text-xs text-gray-400 mt-1">
              Default: all 11 stages. Unselect to skip; order is always canonical.
            </p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Config (JSON)</label>
            <textarea v-model="configInput" rows="3" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono" placeholder="{}" />
          </div>
          <label class="flex items-center gap-2">
            <input type="checkbox" v-model="createForm.fail_fast" class="rounded border-gray-300" />
            <span class="text-sm text-gray-700">Fail-fast (halt on first stage failure)</span>
          </label>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <button class="px-4 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50" @click="showCreateModal = false">
            Cancel
          </button>
          <button
            class="px-4 py-2 text-sm rounded-lg bg-violet-600 text-white hover:bg-violet-700 disabled:opacity-50"
            :disabled="creating || !createForm.title || !createForm.requirement"
            @click="handleCreate"
          >
            {{ creating ? 'Creating…' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ============ Workflow detail modal ============ -->
    <div v-if="showDetail && selectedWorkflow" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" @click.self="closeDetail">
      <div class="bg-white rounded-xl shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div class="px-6 py-4 border-b border-gray-200 flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-800">{{ selectedWorkflow.title }}</h2>
            <div class="text-xs text-gray-500 mt-1">
              Workflow #{{ selectedWorkflow.id }} ·
              <span class="px-2 py-0.5 rounded-full" :class="statusBadgeClass(selectedWorkflow.status)">
                {{ selectedWorkflow.status }}
              </span>
              <span v-if="selectedWorkflow.current_stage"> · stage: {{ stageDisplayName(selectedWorkflow.current_stage) }}</span>
            </div>
          </div>
          <button class="text-gray-400 hover:text-gray-600" @click="closeDetail">✕</button>
        </div>

        <div class="px-6 py-4 space-y-4 text-sm">
          <div v-if="selectedWorkflow.error_message" class="bg-red-50 border border-red-200 rounded p-3 text-red-700">
            {{ selectedWorkflow.error_message }}
          </div>

          <!-- Requirement -->
          <div>
            <div class="text-xs text-gray-500 uppercase mb-1">Requirement</div>
            <div class="bg-gray-50 border border-gray-200 rounded-lg p-3 text-gray-700 whitespace-pre-wrap">{{ selectedWorkflow.requirement }}</div>
          </div>

          <!-- Progress -->
          <div v-if="isRunning(selectedWorkflow.status)">
            <div class="text-xs text-gray-500 uppercase mb-1">Progress</div>
            <div class="h-2 bg-gray-100 rounded-full overflow-hidden">
              <div
                class="h-full bg-violet-500 transition-all duration-300"
                :style="{ width: `${selectedWorkflow.progress}%` }"
              />
            </div>
            <div class="text-xs text-gray-400 mt-1">{{ selectedWorkflow.progress }}%</div>
          </div>

          <!-- Meta grid -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div>
              <div class="text-xs text-gray-500 uppercase">Started</div>
              <span>{{ formatTime(selectedWorkflow.started_at) }}</span>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Completed</div>
              <span>{{ formatTime(selectedWorkflow.completed_at) }}</span>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Cancelled</div>
              <span>{{ formatTime(selectedWorkflow.cancelled_at) }}</span>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Project / Squad</div>
              <span>{{ selectedWorkflow.project_id || '—' }} / {{ selectedWorkflow.squad_id || '—' }}</span>
            </div>
          </div>

          <!-- Stage timeline (vertical) -->
          <div>
            <div class="text-xs text-gray-500 uppercase mb-2">
              Stages <span v-if="loadingDetail" class="text-gray-400">· loading…</span>
            </div>
            <div v-if="selectedStages.length === 0 && !loadingDetail" class="text-xs text-gray-400">
              No stage records yet.
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="stage in selectedStages"
                :key="stage.id"
                class="border border-gray-200 rounded-lg p-3"
              >
                <div class="flex items-center justify-between gap-2 flex-wrap">
                  <div class="flex items-center gap-2">
                    <span class="text-xs text-gray-400 font-mono">#{{ stage.order }}</span>
                    <span class="text-xs px-2 py-0.5 rounded-full" :class="stageStatusBadgeClass(stage.status)">
                      {{ stage.status }}
                    </span>
                    <span class="font-medium text-gray-900 text-sm">{{ stageDisplayName(stage.key) }}</span>
                    <span v-if="stage.agent_role" class="text-xs text-gray-400">· {{ stage.agent_role }}</span>
                  </div>
                  <div class="flex items-center gap-2">
                    <span class="text-xs text-gray-400">{{ formatDuration(stage.duration_ms) }}</span>
                    <button
                      v-if="stage.status === 'failed' || stage.status === 'skipped'"
                      class="text-xs px-2 py-1 rounded border border-violet-300 text-violet-700 hover:bg-violet-50 disabled:opacity-50"
                      :disabled="retryingStageId === stage.id"
                      @click="handleRetryStage(selectedWorkflow.id, stage.id)"
                    >
                      {{ retryingStageId === stage.id ? 'Retrying…' : 'Retry from here' }}
                    </button>
                  </div>
                </div>
                <div class="text-xs text-gray-500 mt-1 flex gap-3 flex-wrap">
                  <span>started: {{ formatTime(stage.started_at) }}</span>
                  <span>completed: {{ formatTime(stage.completed_at) }}</span>
                  <span>progress: {{ stage.progress }}%</span>
                </div>
                <div v-if="stage.error_message" class="text-xs text-red-600 mt-1">
                  error: {{ stage.error_message }}
                </div>
                <!-- Logs -->
                <div v-if="stage.logs && stage.logs.length > 0" class="mt-2">
                  <div class="text-xs text-gray-500 uppercase mb-1">Logs</div>
                  <pre class="bg-gray-900 text-gray-100 rounded p-2 text-xs overflow-x-auto max-h-40">{{ stage.logs.join('\n') }}</pre>
                </div>
                <!-- Output -->
                <details v-if="stage.output && Object.keys(stage.output).length > 0" class="mt-2">
                  <summary class="text-xs text-sky-600 cursor-pointer hover:underline">Output (JSON)</summary>
                  <pre class="bg-gray-50 border border-gray-200 rounded p-2 text-xs overflow-x-auto mt-1">{{ formatJson(stage.output) }}</pre>
                </details>
                <!-- Input -->
                <details v-if="stage.input && Object.keys(stage.input).length > 0" class="mt-1">
                  <summary class="text-xs text-gray-500 cursor-pointer hover:underline">Input (JSON)</summary>
                  <pre class="bg-gray-50 border border-gray-200 rounded p-2 text-xs overflow-x-auto mt-1">{{ formatJson(stage.input) }}</pre>
                </details>
              </div>
            </div>
          </div>

          <!-- Artifacts -->
          <div v-if="artifactEntries(selectedWorkflow.artifacts).length > 0">
            <div class="text-xs text-gray-500 uppercase mb-2">Artifacts</div>
            <div class="space-y-2">
              <div
                v-for="entry in artifactEntries(selectedWorkflow.artifacts)"
                :key="entry.key"
                class="border border-gray-200 rounded-lg p-3"
              >
                <div class="text-xs font-medium text-gray-900">{{ entry.key }}</div>
                <pre class="bg-gray-50 border border-gray-200 rounded p-2 text-xs overflow-x-auto mt-1">{{ formatJson(entry.value) }}</pre>
              </div>
            </div>
          </div>
        </div>

        <div class="px-6 py-4 border-t border-gray-200 flex justify-between gap-2">
          <div class="flex gap-2">
            <button
              v-if="isRunning(selectedWorkflow.status)"
              class="px-4 py-2 text-sm rounded-lg border border-amber-300 text-amber-700 hover:bg-amber-50 disabled:opacity-50"
              :disabled="cancelingId === selectedWorkflow.id"
              @click="handleCancel(selectedWorkflow.id)"
            >
              {{ cancelingId === selectedWorkflow.id ? 'Cancelling…' : 'Cancel Workflow' }}
            </button>
            <button
              v-if="isTerminal(selectedWorkflow.status)"
              class="px-4 py-2 text-sm rounded-lg border border-red-300 text-red-700 hover:bg-red-50"
              @click="handleDelete(selectedWorkflow.id); closeDetail()"
            >
              Delete
            </button>
          </div>
          <button class="px-4 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50" @click="closeDetail">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
