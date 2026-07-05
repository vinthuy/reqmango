<template>
  <div class="issue-detail-page min-h-screen bg-white">
    <IssueDetailHeader :issue :saving @back="goBack" @save="saveIssue" />

    <div class="max-w-5xl mx-auto px-6 py-6">
      <div class="flex gap-8">
        <!-- Main content: tab buttons + tab panels -->
        <div class="flex-1 min-w-0">
          <!-- 5 tab buttons -->
          <div class="flex border-b-2 border-gray-200 mb-4">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              data-test="tab-btn"
              @click="activeTab = tab.key"
              :class="activeTab === tab.key
                ? 'border-indigo-500 text-indigo-600'
                : 'border-transparent text-gray-500'"
              class="px-4 py-2 text-sm font-medium border-b-2 -mb-0.5 transition-colors"
            >
              {{ tab.label }}
            </button>
          </div>

          <!-- Tab panels -->
          <IssueTabDetails
            v-if="activeTab === 'details'"
            v-bind="detailProps"
            @update:title="issueForm.name = $event"
            @update:description="issueForm.description = $event"
          />
          <IssueTabRelations
            v-else-if="activeTab === 'relations'"
            :issue-id="issueId"
            :project-id="projectId"
            :workspace-id="workspaceId"
            :parent="issue?.parent"
            :sub-issues="subIssues"
            :issue-types="issueTypes"
            @navigate="navigateToIssue"
          />
          <IssueTabAttachments
            v-else-if="activeTab === 'attachments'"
            :issue-id="issueId"
            :project-id="projectId"
          />
          <IssueTabTimeTracking
            v-else-if="activeTab === 'timetrack'"
            :issue-id="issueId"
          />
          <IssueTabActivity
            v-else-if="activeTab === 'activity'"
            :issue-id="issueId"
          />
        </div>

        <!-- Right sidebar -->
        <IssuePropertySidebar
          v-if="issue"
          :issue
          :states
          :members="projectMembers"
          :cycles
          :modules
          :selected-agent-id="selectedAgentId"
          :agent-dispatching="agentDispatching"
          @update:state="(id: any) => instantUpdate('state_id', id)"
          @update:priority="(p: any) => instantUpdate('priority', p)"
          @update:assignee="instantUpdateAssignee"
          @update:cycle="(id: any) => instantUpdate('cycle_id', id)"
          @update:module="(id: any) => instantUpdate('module_id', id)"
          @update:start-date="(d: any) => instantUpdate('start_date', d + 'T00:00:00Z')"
          @update:target-date="(d: any) => instantUpdate('target_date', d + 'T00:00:00Z')"
        >
          <template #agent>
            <AgentSelector v-model="selectedAgentId" :workspace-id="workspaceId" />
            <button
              v-if="selectedAgentId"
              @click="dispatchAgent"
              :disabled="agentDispatching"
              class="mt-2 w-full px-3 py-1.5 text-xs font-medium rounded-md bg-violet-500 hover:bg-violet-600 text-white disabled:opacity-50"
            >
              {{ agentDispatching ? t('agent.dispatching') : t('agent.dispatch') }}
            </button>
          </template>
        </IssuePropertySidebar>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { useRoute, useRouter } from 'vue-router'
import * as issueApi from '@/api/issue'
import * as stateApi from '@/api/project-settings'
import * as cycleApi from '@/api/cycle'
import * as moduleApi from '@/api/module'
import * as issueTypeApi from '@/api/issue-type'
import projectApi from '@/api/project'
import { agentApi } from '@/api/agent'
import IssueDetailHeader from '@/components/IssueDetailHeader.vue'
import IssuePropertySidebar from '@/components/IssuePropertySidebar.vue'
import IssueTabDetails from '@/components/IssueTabDetails.vue'
import IssueTabRelations from '@/components/IssueTabRelations.vue'
import IssueTabAttachments from '@/components/IssueTabAttachments.vue'
import IssueTabTimeTracking from '@/components/IssueTabTimeTracking.vue'
import IssueTabActivity from '@/components/IssueTabActivity.vue'
import AgentSelector from '@/components/AgentSelector.vue'

// Route params
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const toast = useToast()
const issueId = parseInt(route.params.issueId as string, 10) || 0
const projectId = ref(0)
const workspaceId = ref(0)

// Reactive state
const issue = ref<any>(null)
const saving = ref(false)
const issueForm = ref({ name: '', description: '' })
const states = ref<any[]>([])
const cycles = ref<any[]>([])
const modules = ref<any[]>([])
const projectMembers = ref<any[]>([])
const issueTypes = ref<any[]>([])
const subIssues = ref<any[]>([])
const activeTab = ref('details')
const selectedAgentId = ref('')
const agentDispatching = ref(false)

// Tabs definition
const tabs = computed(() => [
  { key: 'details', label: t('issue.tabDetails') },
  { key: 'relations', label: t('issue.tabRelations') },
  { key: 'attachments', label: t('issue.tabAttachments') },
  { key: 'timetrack', label: t('issue.tabTimetrack') },
  { key: 'activity', label: t('issue.tabActivity') },
])

// Props for IssueTabDetails
const detailProps = computed(() => ({
  issueId,
  issue: issue.value,
  workspaceId: workspaceId.value,
  projectId: projectId.value,
  issueTypeId: issue.value?.issue_type?.id,
  members: projectMembers.value,
}))

// Data loading
onMounted(async () => {
  if (!issueId) return
  try {
    const issueData = await issueApi.getIssue(issueId)
    issue.value = issueData
    if (issueData.workspace_id) workspaceId.value = issueData.workspace_id
    if (issueData.project_id) projectId.value = issueData.project_id
    issueForm.value = {
      name: issueData.name || '',
      description: issueData.description_html || '',
    }
    subIssues.value = issueData.sub_issues || []

    // Load auxiliary data
    await Promise.all([
      loadStates(),
      loadCycles(),
      loadModules(),
      loadProjectMembers(),
      loadIssueTypes(),
    ])
  } catch (error) {
    console.error('Failed to load issue:', error)
  }
})

async function loadStates() {
  try {
    const data = await stateApi.listStates(projectId.value)
    states.value = data
  } catch (error) {
    console.error('Failed to load states:', error)
  }
}

async function loadCycles() {
  try {
    const data = await cycleApi.listCycles(projectId.value)
    cycles.value = data.items || data
  } catch (error) {
    console.error('Failed to load cycles:', error)
  }
}

async function loadModules() {
  try {
    const data = await moduleApi.listModules(projectId.value, workspaceId.value)
    modules.value = data
  } catch (error) {
    console.error('Failed to load modules:', error)
  }
}

async function loadIssueTypes() {
  try {
    const types = await issueTypeApi.getIssueTypes(workspaceId.value, projectId.value)
    issueTypes.value = types
  } catch (error) {
    console.error('Failed to load issue types:', error)
  }
}

async function loadProjectMembers() {
  try {
    const data = await projectApi.listProjectMembers(projectId.value)
    projectMembers.value = data.map((m: any) => m.user || m)
  } catch (error) {
    console.error('Failed to load project members:', error)
  }
}

// Save issue (batch save title + description)
async function saveIssue() {
  saving.value = true
  try {
    await issueApi.updateIssue(issueId, {
      name: issueForm.value.name,
      description_html: issueForm.value.description || undefined,
    })
    toast.success(t('issue.saveSuccess'))
  } catch (error) {
    console.error('Failed to save issue:', error)
    toast.error(t('issue.saveFailed'))
  } finally {
    saving.value = false
  }
}

// Instant update for sidebar field changes
function instantUpdate(field: string, value: any) {
  issueApi.updateIssue(issueId, { [field]: value }).catch((e: any) => {
    console.error('Failed to update field:', field, e)
    toast.error(t('issue.saveFailed'))
  })
}

// Instant update for assignee
function instantUpdateAssignee(userId: any) {
  const assigneeIds = userId ? [Number(userId)] : []
  issueApi.updateIssue(issueId, { assignee_ids: assigneeIds }).catch((e: any) => {
    console.error('Failed to update assignee:', e)
    toast.error(t('issue.saveFailed'))
  })
}

// Navigation
function goBack() {
  router.back()
}

function navigateToIssue(id: number) {
  router.push({
    path: `/workspace/${route.params.slug}/project/${projectId.value}/issues/${id}`,
  })
}

// Agent dispatch
async function dispatchAgent() {
  if (!selectedAgentId.value || !selectedAgentId.value.startsWith('agent:')) return
  const agentId = parseInt(selectedAgentId.value.replace('agent:', ''))
  if (!agentId) return
  agentDispatching.value = true
  try {
    await agentApi.dispatch(workspaceId.value, agentId, {
      task: `Analyze issue #${issue.value?.sequence_id || issueId}: ${issue.value?.name || 'Untitled'}`,
      issue_id: issueId,
      project_id: projectId.value,
    })
    toast.success('Agent dispatched successfully!')
    selectedAgentId.value = ''
  } catch (e: any) {
    toast.error(e?.response?.data?.message || e?.message || 'Failed to dispatch agent')
  } finally {
    agentDispatching.value = false
  }
}
</script>

<style scoped>
.issue-detail-page {
  min-height: 100vh;
}
</style>
