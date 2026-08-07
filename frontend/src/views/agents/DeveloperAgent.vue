<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { useSSE } from '@/composables/useSSE'
import {
  developerAgentApi,
  type DeveloperJob,
  type DeveloperJobCreate,
  type DeveloperJobStatus
} from '@/api/developer-agent'
import { githubApi, type GitHubConnection } from '@/api/github'

const { getWorkspaceId } = useWorkspaceId()
const { onEvent } = useSSE()

const workspaceId = ref(0)
const jobs = ref<DeveloperJob[]>([])
const connections = ref<GitHubConnection[]>([])
const loading = ref(true)
const creating = ref(false)
const cancelingId = ref<number | null>(null)
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const selectedJob = ref<DeveloperJob | null>(null)
const errorMsg = ref('')
let sseCleanup: (() => void) | null = null

const statusFilters: { value: DeveloperJobStatus | ''; label: string }[] = [
  { value: '', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'analyzing', label: 'Analyzing' },
  { value: 'generating', label: 'Generating' },
  { value: 'committing', label: 'Committing' },
  { value: 'opening_pr', label: 'Opening PR' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'cancelled', label: 'Cancelled' }
]
const activeFilter = ref<DeveloperJobStatus | ''>('')

const newJob = ref<DeveloperJobCreate>(emptyJob())

function emptyJob(): DeveloperJobCreate {
  return {
    title: '',
    requirement_text: '',
    design_doc_url: undefined,
    git_connection_id: 0,
    git_provider: 'github',
    branch_name: '',
    base_branch: 'main',
    commit_message: '',
    pr_title: '',
    pr_body: '',
    language: '',
    files: []
  }
}

const filteredJobs = ref<DeveloperJob[]>([])
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
    const [jobsRes, connsRes] = await Promise.all([
      developerAgentApi.list(wsId),
      githubApi.list(wsId).catch(() => [])
    ])
    jobs.value = jobsRes || []
    connections.value = connsRes || []
    applyFilter()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to load developer jobs'
    jobs.value = []
    filteredJobs.value = []
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  newJob.value = emptyJob()
  if (connections.value.length > 0) {
    newJob.value.git_connection_id = connections.value[0].id
  }
  showCreateModal.value = true
}

function openDetailModal(job: DeveloperJob) {
  selectedJob.value = job
  showDetailModal.value = true
}

async function handleCreate() {
  if (!newJob.value.title) {
    errorMsg.value = 'Title is required'
    return
  }
  if (!newJob.value.git_connection_id) {
    errorMsg.value = 'A GitHub connection is required'
    return
  }
  creating.value = true
  errorMsg.value = ''
  try {
    await developerAgentApi.create(workspaceId.value, newJob.value)
    showCreateModal.value = false
    await loadData()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to start developer job'
  } finally {
    creating.value = false
  }
}

async function handleCancel(jobId: number) {
  if (!confirm('Cancel this developer job? In-flight GitHub API calls may still complete.')) {
    return
  }
  cancelingId.value = jobId
  try {
    await developerAgentApi.cancel(workspaceId.value, jobId)
    await loadData()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to cancel job'
  } finally {
    cancelingId.value = null
  }
}

async function handleDelete(jobId: number) {
  if (!confirm('Delete this developer job record?')) return
  try {
    await developerAgentApi.delete(workspaceId.value, jobId)
    await loadData()
  } catch (err: any) {
    errorMsg.value = err.response?.data?.message || 'Failed to delete job'
  }
}

function statusBadgeClass(status: DeveloperJobStatus): string {
  switch (status) {
    case 'completed': return 'bg-green-100 text-green-700'
    case 'failed': return 'bg-red-100 text-red-700'
    case 'cancelled': return 'bg-gray-100 text-gray-600'
    case 'pending': return 'bg-gray-100 text-gray-700'
    case 'analyzing':
    case 'generating':
    case 'committing':
    case 'opening_pr': return 'bg-blue-100 text-blue-700'
    default: return 'bg-gray-100 text-gray-700'
  }
}

function isTerminal(status: DeveloperJobStatus): boolean {
  return status === 'completed' || status === 'failed' || status === 'cancelled'
}

function isRunning(status: DeveloperJobStatus): boolean {
  return status === 'pending' || status === 'analyzing' || status === 'generating' ||
    status === 'committing' || status === 'opening_pr'
}

function formatTime(iso?: string): string {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function upsertJob(job: DeveloperJob) {
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
    if (!event.startsWith('developer_job.')) return
    if (data && typeof data === 'object' && data.id) {
      upsertJob(data as DeveloperJob)
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
          <h1 class="text-xl font-semibold text-gray-800">👨‍💻 Developer Agent</h1>
          <p class="text-sm text-gray-500 mt-0.5">
            Generate code, commit to repository, and open pull requests (P4-001)
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
            class="px-3 py-1.5 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50"
            :disabled="connections.length === 0"
            @click="openCreateModal"
          >
            New Developer Job
          </button>
        </div>
      </div>
    </header>

    <main class="p-6">
      <div class="max-w-7xl mx-auto space-y-4">
        <div v-if="connections.length === 0" class="bg-amber-50 border border-amber-200 rounded-xl p-4 text-sm text-amber-800">
          No GitHub connections configured. Create one under the project's GitHub integration before launching a Developer Job.
        </div>

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
          Loading developer jobs…
        </div>
        <div v-else-if="filteredJobs.length === 0" class="text-sm text-gray-400 py-12 text-center border border-gray-200 rounded-xl bg-white">
          No developer jobs yet. Click "New Developer Job" to generate code and open a PR.
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
                  <span v-if="job.pr_number" class="text-xs text-indigo-600">#{{ job.pr_number }}</span>
                </div>
                <div class="text-xs text-gray-500 mt-1 flex flex-wrap gap-x-3 gap-y-1">
                  <span>branch: <code class="bg-gray-100 px-1 rounded">{{ job.branch_name || '—' }}</code></span>
                  <span>base: <code class="bg-gray-100 px-1 rounded">{{ job.base_branch }}</code></span>
                  <span v-if="job.git_provider">provider: {{ job.git_provider }}</span>
                  <span>created: {{ formatTime(job.created_at) }}</span>
                </div>
                <div v-if="job.current_step" class="text-xs text-blue-600 mt-1">step: {{ job.current_step }}</div>
                <div v-if="job.error_message" class="text-xs text-red-600 mt-1 truncate">
                  error: {{ job.error_message }}
                </div>
              </div>
              <div class="flex items-center gap-2" @click.stop>
                <a
                  v-if="job.pr_url"
                  :href="job.pr_url"
                  target="_blank"
                  rel="noopener"
                  class="text-xs px-2 py-1 rounded border border-indigo-300 text-indigo-700 hover:bg-indigo-50"
                >
                  View PR ↗
                </a>
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
          <h2 class="text-lg font-semibold text-gray-800">New Developer Job</h2>
          <p class="text-xs text-gray-500 mt-0.5">
            The agent will analyze the requirement, generate code, commit to a branch, and open a PR.
          </p>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div v-if="errorMsg" class="bg-red-50 border border-red-200 rounded p-2 text-sm text-red-700">
            {{ errorMsg }}
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Title *</label>
            <input v-model="newJob.title" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="e.g. Add user login page" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Requirement</label>
            <textarea v-model="newJob.requirement_text" rows="4" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="User story, acceptance criteria, or free-form description" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">GitHub Connection *</label>
              <select v-model="newJob.git_connection_id" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm">
                <option v-for="c in connections" :key="c.id" :value="c.id">
                  {{ c.repo_owner }}/{{ c.repo_name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Language hint</label>
              <input v-model="newJob.language" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="e.g. go, typescript" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Branch name</label>
              <input v-model="newJob.branch_name" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="auto-generated if empty" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Base branch</label>
              <input v-model="newJob.base_branch" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="main" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Design doc URL (optional)</label>
            <input v-model="newJob.design_doc_url" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="https://…" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">PR title (optional)</label>
            <input v-model="newJob.pr_title" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="auto: feat: <title>" />
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <button class="px-4 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50" @click="showCreateModal = false">
            Cancel
          </button>
          <button
            class="px-4 py-2 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50"
            :disabled="creating || !newJob.title || !newJob.git_connection_id"
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
            <div class="text-xs text-gray-500 mt-1">Job #{{ selectedJob.id }} · {{ selectedJob.status }}</div>
          </div>
          <button class="text-gray-400 hover:text-gray-600" @click="showDetailModal = false">✕</button>
        </div>
        <div class="px-6 py-4 space-y-4 text-sm">
          <div v-if="selectedJob.error_message" class="bg-red-50 border border-red-200 rounded p-3 text-red-700">
            {{ selectedJob.error_message }}
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <div class="text-xs text-gray-500 uppercase">Branch</div>
              <code class="text-xs bg-gray-100 px-1 rounded">{{ selectedJob.branch_name || '—' }}</code>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Base</div>
              <code class="text-xs bg-gray-100 px-1 rounded">{{ selectedJob.base_branch }}</code>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">PR</div>
              <a v-if="selectedJob.pr_url" :href="selectedJob.pr_url" target="_blank" rel="noopener" class="text-indigo-600 hover:underline">
                #{{ selectedJob.pr_number }} ↗
              </a>
              <span v-else class="text-gray-400">—</span>
            </div>
            <div>
              <div class="text-xs text-gray-500 uppercase">Commit SHA</div>
              <code class="text-xs bg-gray-100 px-1 rounded">{{ selectedJob.commit_sha ? selectedJob.commit_sha.slice(0, 12) : '—' }}</code>
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
          <div v-if="selectedJob.generated_files && selectedJob.generated_files.length > 0">
            <div class="text-xs text-gray-500 uppercase mb-1">Generated files ({{ selectedJob.generated_files.length }})</div>
            <ul class="border border-gray-200 rounded divide-y divide-gray-100">
              <li v-for="f in selectedJob.generated_files" :key="f.path" class="px-3 py-2 text-xs font-mono">
                {{ f.path }}
              </li>
            </ul>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <a
            v-if="selectedJob.pr_url"
            :href="selectedJob.pr_url"
            target="_blank"
            rel="noopener"
            class="px-4 py-2 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700"
          >
            Open PR ↗
          </a>
          <button class="px-4 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50" @click="showDetailModal = false">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
</content>
</invoke>