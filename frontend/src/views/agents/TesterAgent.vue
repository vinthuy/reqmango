<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { useSSE } from '@/composables/useSSE'
import {
  testerAgentApi,
  type TesterJob,
  type TesterJobCreate,
  type TesterJobStatus
} from '@/api/tester-agent'

const { getWorkspaceId } = useWorkspaceId()
const { onEvent } = useSSE()

const workspaceId = ref(0)
const jobs = ref<TesterJob[]>([])
const loading = ref(true)
const creating = ref(false)
const cancelingId = ref<number | null>(null)
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const selectedJob = ref<TesterJob | null>(null)
const errorMsg = ref('')
let sseCleanup: (() => void) | null = null

const statusFilters: { value: TesterJobStatus | ''; label: string }[] = [
  { value: '', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'generating_cases', label: 'Generating Cases' },
  { value: 'executing', label: 'Executing' },
  { value: 'reporting', label: 'Reporting' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'cancelled', label: 'Cancelled' }
]
const activeFilter = ref<TesterJobStatus | ''>('')

const projectIdInput = ref('')
const issueIdInput = ref('')

const newJob = ref<TesterJobCreate>(emptyJob())

function emptyJob(): TesterJobCreate {
  return {
    title: '',
    requirement_text: '',
    acceptance_criteria: '',
    test_scope: 'unit',
    project_id: undefined,
    issue_id: undefined
  }
}

const filteredJobs = ref<TesterJob[]>([])
function applyFilter() {
  if (!activeFilter.value) {
    filteredJobs.value = jobs.value
  } else {
    filteredJobs.value = jobs.value.filter(j => j.status === activeFilter.value)
  }
}

async function loadData() {
  loading.value = true
  errorMsg.value = ''
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    workspaceId.value = wsId
    const jobsRes = await testerAgentApi.list(wsId)
    jobs.value = jobsRes || []
    applyFilter()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to load tester jobs'
    jobs.value = []
    filteredJobs.value = []
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  newJob.value = emptyJob()
  projectIdInput.value = ''
  issueIdInput.value = ''
  showCreateModal.value = true
}

function openDetailModal(job: TesterJob) {
  selectedJob.value = job
  showDetailModal.value = true
}

async function handleCreate() {
  if (!newJob.value.title) {
    errorMsg.value = 'Title is required'
    return
  }
  creating.value = true
  errorMsg.value = ''
  try {
    const payload: TesterJobCreate = {
      ...newJob.value,
      project_id: projectIdInput.value ? Number(projectIdInput.value) : undefined,
      issue_id: issueIdInput.value ? Number(issueIdInput.value) : undefined
    }
    await testerAgentApi.create(workspaceId.value, payload)
    showCreateModal.value = false
    await loadData()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to start tester job'
  } finally {
    creating.value = false
  }
}

async function handleCancel(jobId: number) {
  if (!confirm('Cancel this tester job?')) return
  cancelingId.value = jobId
  try {
    await testerAgentApi.cancel(workspaceId.value, jobId)
    await loadData()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to cancel job'
  } finally {
    cancelingId.value = null
  }
}

async function handleDelete(jobId: number) {
  if (!confirm('Delete this tester job record?')) return
  try {
    await testerAgentApi.delete(workspaceId.value, jobId)
    await loadData()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to delete job'
  }
}

function statusBadgeClass(status: TesterJobStatus): string {
  switch (status) {
    case 'completed': return 'bg-green-100 text-green-700'
    case 'failed': return 'bg-red-100 text-red-700'
    case 'cancelled': return 'bg-amber-100 text-amber-700'
    case 'pending': return 'bg-gray-100 text-gray-700'
    case 'generating_cases': return 'bg-blue-100 text-blue-700'
    case 'executing': return 'bg-indigo-100 text-indigo-700'
    case 'reporting': return 'bg-purple-100 text-purple-700'
    default: return 'bg-gray-100 text-gray-700'
  }
}

function scopeBadgeClass(scope: string): string {
  switch (scope) {
    case 'unit': return 'bg-teal-100 text-teal-700'
    case 'integration': return 'bg-cyan-100 text-cyan-700'
    case 'e2e': return 'bg-violet-100 text-violet-700'
    default: return 'bg-gray-100 text-gray-600'
  }
}

function resultBadgeClass(status: 'passed' | 'failed' | 'skipped'): string {
  switch (status) {
    case 'passed': return 'bg-green-100 text-green-700'
    case 'failed': return 'bg-red-100 text-red-700'
    case 'skipped': return 'bg-gray-100 text-gray-600'
    default: return 'bg-gray-100 text-gray-600'
  }
}

function isTerminal(status: TesterJobStatus): boolean {
  return status === 'completed' || status === 'failed' || status === 'cancelled'
}

function isRunning(status: TesterJobStatus): boolean {
  return status === 'pending' || status === 'generating_cases' ||
    status === 'executing' || status === 'reporting'
}

function formatTime(iso?: string): string {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function upsertJob(job: TesterJob) {
  const idx = jobs.value.findIndex(j => j.id === job.id)
  if (idx >= 0) {
    jobs.value[idx] = job
    // If the detail modal is showing this job, refresh it too.
    if (selectedJob.value && selectedJob.value.id === job.id) {
      selectedJob.value = job
    }
  } else {
    jobs.value.unshift(job)
  }
  applyFilter()
}

onMounted(async () => {
  await loadData()
  // Subscribe to SSE updates for live progress.
  sseCleanup = onEvent((event, data) => {
    if (!event.startsWith('tester_job.')) return
    if (data && typeof data === 'object' && data.id) {
      upsertJob(data as TesterJob)
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
          <h1 class="text-xl font-semibold text-gray-800">🧪 Tester Agent</h1>
          <p class="text-sm text-gray-500 mt-0.5">
            Generate test cases, execute tests, and report bugs (P4-002)
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button
            class="px-3 py-1.5 text-sm rounded-lg border border-gray-300 bg-white hover:bg-gray-50"
            @click="loadData"
          >
            Refresh
          </button>
          <button
            class="px-3 py-1.5 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700"
            @click="openCreateModal"
          >
            New Tester Job
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
            v-for="f in statusFilters"
            :key="f.value || 'all'"
            class="px-3 py-1 text-xs rounded-full border transition-colors"
            :class="activeFilter === f.value
              ? 'bg-indigo-600 text-white border-indigo-600'
              : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50'"
            @click="activeFilter = f.value; applyFilter()"
          >
            {{ f.label }}
          </button>
        </div>

        <!-- Job list -->
        <div v-if="loading" class="text-sm text-gray-400 py-8 text-center">
          Loading tester jobs…
        </div>
        <div v-else-if="filteredJobs.length === 0" class="text-sm text-gray-400 py-12 text-center border border-gray-200 rounded-xl bg-white">
          No tester jobs yet. Click "New Tester Job" to generate test cases and run tests.
        </div>
        <div v-else class="space-y-3">
          <div
            v-for="job in filteredJobs"
            :key="job.id"
            class="bg-white border border-gray-200 rounded-xl p-4 hover:border-indigo-300 cursor-pointer transition-colors"
            @click="openDetailModal(job)"
          >
            <div class="flex items-start justify-between gap-3 flex-wrap">
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="font-medium text-gray-900 truncate">{{ job.title }}</span>
                  <span class="text-xs px-2 py-0.5 rounded-full" :class="statusBadgeClass(job.status)">
                    {{ job.status }}
                  </span>
                  <span class="text-xs px-2 py-0.5 rounded-full" :class="scopeBadgeClass(job.test_scope)">
                    {{ job.test_scope }}
                  </span>
                </div>
                <div class="text-xs text-gray-500 mt-1 flex flex-wrap gap-x-3 gap-y-1">
                  <span>cases: {{ job.total_cases }}</span>
                  <span class="text-green-600">✓ {{ job.pass_count }}</span>
                  <span class="text-red-600">✗ {{ job.fail_count }}</span>
                  <span class="text-gray-500">⊘ {{ job.skip_count }}</span>
                  <span>created: {{ formatTime(job.created_at) }}</span>
                </div>
                <div v-if="job.current_step" class="text-xs text-blue-600 mt-1">step: {{ job.current_step }}</div>
                <div v-if="job.error_message" class="text-xs text-red-600 mt-1 truncate">
                  error: {{ job.error_message }}
                </div>
              </div>
              <div class="flex items-center gap-2" @click.stop>
                <button
                  v-if="isRunning(job.status)"
                  class="text-xs px-2 py-1 rounded border border-amber-300 text-amber-700 hover:bg-amber-50 disabled:opacity-50"
                  :disabled="cancelingId === job.id"
                  @click="handleCancel(job.id)"
                >
                  {{ cancelingId === job.id ? 'Cancelling…' : 'Cancel' }}
                </button>
                <button
                  v-if="isTerminal(job.status)"
                  class="text-xs px-2 py-1 rounded border border-red-300 text-red-700 hover:bg-red-50"
                  @click="handleDelete(job.id)"
                >
                  Delete
                </button>
              </div>
            </div>
            <!-- Progress bar -->
            <div v-if="isRunning(job.status)" class="mt-3">
              <div class="h-1.5 bg-gray-100 rounded-full overflow-hidden">
                <div
                  class="h-full bg-indigo-500 transition-all duration-300"
                  :style="{ width: `${job.progress}%` }"
                />
              </div>
              <div class="text-xs text-gray-400 mt-1">{{ job.progress }}%</div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Create modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" @click.self="showCreateModal = false">
      <div class="bg-white rounded-xl shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="px-6 py-4 border-b border-gray-200">
          <h2 class="text-lg font-semibold text-gray-800">New Tester Job</h2>
          <p class="text-xs text-gray-500 mt-0.5">
            The agent will generate test cases, execute them, and report any bugs found.
          </p>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div v-if="errorMsg" class="bg-red-50 border border-red-200 rounded p-2 text-sm text-red-700">
            {{ errorMsg }}
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Title *</label>
            <input v-model="newJob.title" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="e.g. Test user login flow" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Requirement Text</label>
            <textarea v-model="newJob.requirement_text" rows="4" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="User story or free-form requirement description" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Acceptance Criteria</label>
            <textarea v-model="newJob.acceptance_criteria" rows="3" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Acceptance criteria, one per line" />
          </div>
          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Test Scope</label>
              <select v-model="newJob.test_scope" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm">
                <option value="unit">unit</option>
                <option value="integration">integration</option>
                <option value="e2e">e2e</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Project ID (optional)</label>
              <input type="number" v-model="projectIdInput" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="—" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Issue ID (optional)</label>
              <input type="number" v-model="issueIdInput" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="—" />
            </div>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <button class="px-4 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50" @click="showCreateModal = false">
            Cancel
          </button>
          <button
            class="px-4 py-2 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50"
            :disabled="creating || !newJob.title"
            @click="handleCreate"
          >
            {{ creating ? 'Starting…' : 'Start Job' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Detail modal -->
    <div v-if="showDetailModal && selectedJob" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4" @click.self="showDetailModal = false">
      <div class="bg-white rounded-xl shadow-xl max-w-3xl w-full max-h-[90vh] overflow-y-auto">
        <div class="px-6 py-4 border-b border-gray-200 flex items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-800">{{ selectedJob.title }}</h2>
            <div class="text-xs text-gray-500 mt-1">
              Job #{{ selectedJob.id }} · {{ selectedJob.status }} · {{ selectedJob.test_scope }}
            </div>
          </div>
          <button class="text-gray-400 hover:text-gray-600" @click="showDetailModal = false">✕</button>
        </div>
        <div class="px-6 py-4 space-y-4 text-sm">
          <div v-if="selectedJob.error_message" class="bg-red-50 border border-red-200 rounded p-3 text-red-700">
            {{ selectedJob.error_message }}
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <div class="text-xs text-gray-500 uppercase">Total Cases</div>
              <span>{{ selectedJob.total_cases }}</span>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Pass / Fail / Skip</div>
              <span class="text-green-600">✓ {{ selectedJob.pass_count }}</span>
              · <span class="text-red-600">✗ {{ selectedJob.fail_count }}</span>
              · <span class="text-gray-500">⊘ {{ selectedJob.skip_count }}</span>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Started</div>
              <span>{{ formatTime(selectedJob.started_at) }}</span>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Completed</div>
              <span>{{ formatTime(selectedJob.completed_at) }}</span>
            </div>
          </div>
          <div v-if="selectedJob.requirement_text">
            <div class="text-xs text-gray-500 uppercase mb-1">Requirement</div>
            <pre class="bg-gray-50 border border-gray-200 rounded p-3 text-xs whitespace-pre-wrap">{{ selectedJob.requirement_text }}</pre>
          </div>
          <div v-if="selectedJob.acceptance_criteria">
            <div class="text-xs text-gray-500 uppercase mb-1">Acceptance Criteria</div>
            <pre class="bg-gray-50 border border-gray-200 rounded p-3 text-xs whitespace-pre-wrap">{{ selectedJob.acceptance_criteria }}</pre>
          </div>
          <div v-if="selectedJob.generated_cases && selectedJob.generated_cases.length > 0">
            <div class="text-xs text-gray-500 uppercase mb-1">
              Generated Cases ({{ selectedJob.generated_cases.length }})
            </div>
            <div class="space-y-2">
              <div
                v-for="tc in selectedJob.generated_cases"
                :key="tc.id"
                class="border border-gray-200 rounded-lg p-3"
              >
                <div class="flex items-center gap-2">
                  <code class="text-xs bg-gray-100 px-1 rounded">{{ tc.id }}</code>
                  <span class="font-medium text-gray-900 text-sm">{{ tc.name }}</span>
                </div>
                <div v-if="tc.description" class="text-xs text-gray-600 mt-1">{{ tc.description }}</div>
                <div v-if="tc.steps && tc.steps.length > 0" class="mt-2">
                  <div class="text-xs text-gray-500 mb-0.5">Steps:</div>
                  <ol class="list-decimal list-inside text-xs text-gray-700 space-y-0.5">
                    <li v-for="(step, idx) in tc.steps" :key="idx">{{ step }}</li>
                  </ol>
                </div>
                <div v-if="tc.expected" class="text-xs text-gray-600 mt-1">
                  <span class="text-gray-500">Expected:</span> {{ tc.expected }}
                </div>
              </div>
            </div>
          </div>
          <div v-if="selectedJob.test_results && selectedJob.test_results.length > 0">
            <div class="text-xs text-gray-500 uppercase mb-1">
              Test Results ({{ selectedJob.test_results.length }})
            </div>
            <ul class="border border-gray-200 rounded divide-y divide-gray-100">
              <li v-for="tr in selectedJob.test_results" :key="tr.case_id" class="px-3 py-2">
                <div class="flex items-center gap-2">
                  <span class="text-xs px-2 py-0.5 rounded-full" :class="resultBadgeClass(tr.status)">
                    {{ tr.status }}
                  </span>
                  <code class="text-xs bg-gray-100 px-1 rounded">{{ tr.case_id }}</code>
                  <span class="text-xs text-gray-700 flex-1 truncate">{{ tr.name }}</span>
                  <span class="text-xs text-gray-400">{{ tr.duration_ms }}ms</span>
                </div>
                <div v-if="tr.error" class="text-xs text-red-600 mt-1 pl-2">
                  {{ tr.error }}
                </div>
              </li>
            </ul>
          </div>
          <div v-if="selectedJob.bug_issue_ids && selectedJob.bug_issue_ids.length > 0">
            <div class="text-xs text-gray-500 uppercase mb-1">Bug Issues Created</div>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="bid in selectedJob.bug_issue_ids"
                :key="bid"
                class="text-xs px-2 py-1 rounded-full bg-red-100 text-red-700"
              >
                #{{ bid }}
              </span>
            </div>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <button class="px-4 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50" @click="showDetailModal = false">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
