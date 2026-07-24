<template>
  <div class="issue-detail-page min-h-screen bg-white">
    <IssueDetailHeader :issue :saving :project-identifier="projectIdentifier" :is-watching="isWatching" @back="goBack" @save="saveIssue" @delete="deleteIssue" @toggle-watch="handleToggleWatch" />

    <div class="max-w-6xl mx-auto px-6 pt-4">
      <ApprovalPendingBanner
        :approval="activeApproval"
        :current-user-id="currentUserId"
        @decide="onDecideApproval"
        @cancel="onCancelApproval"
      />
    </div>
    <ApprovalSubmitDialog
      :show="showSubmitDialog"
      :issue-id="issueId"
      :transition-id="submitDialogData.transitionId"
      :from-state-name="submitDialogData.fromStateName"
      :approve-state-name="submitDialogData.approveStateName"
      :approver-names="submitDialogData.approverNames"
      :workflow-name="submitDialogData.workflowName"
      @close="showSubmitDialog = false"
      @submitted="onApprovalSubmitted"
    />
    <ApprovalDecisionDialog
      :show="showDecisionDialog"
      :approval-id="decisionDialogData?.approvalId || 0"
      :decision="decisionDialogData?.decision || 'approved'"
      @close="showDecisionDialog = false"
      @decided="onApprovalDecided"
    />

    <div class="max-w-6xl mx-auto px-6 py-6">
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
              <span v-if="tab.count !== undefined && tab.count > 0" class="ml-1.5 text-xs bg-gray-100 text-gray-500 px-1.5 py-0.5 rounded-full">{{ tab.count }}</span>
            </button>
          </div>

          <!-- Tab panels -->
          <IssueTabDetails
            v-if="activeTab === 'details'"
            v-bind="detailProps"
            @update:title="issueForm.name = $event"
            @update:description="issueForm.description = $event"
            @navigate="navigateToIssue"
          />
          <IssueTabRelations
            v-else-if="activeTab === 'relations'"
            ref="relationsTabRef"
            :issue-id="issueId"
            :project-id="projectId"
            :workspace-id="workspaceId"
            :slug="(route.params.slug as string)"
            :states
            :parent="issue?.parent"
            :sub-issues="issue?.sub_issues || []"
            :issue-types="issueTypes"
            :current-issue-type="issue?.issue_type"
            @navigate="navigateToIssue"
            @refresh="handleRelationsRefresh"
          />
          <IssueTabAttachments
            v-else-if="activeTab === 'attachments'"
            :issue-id="issueId"
            :project-id="projectId"
          />
          <IssueGitPanel
            v-else-if="activeTab === 'git'"
            :workspace-id="workspaceId"
            :issue-id="issueId"
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
          :releases
          :custom-fields="customFieldEntries"
          :workspace-id="workspaceId"
          :agent-dispatching="agentDispatching"
          :labels="projectLabels"
          :relation-summary="relationSidebarSummary"
          @update:state="(id: any) => handleStateChange(id)"
          @update:priority="(p: any) => instantUpdate('priority', p)"
          @update:assignee="instantUpdateAssignee"
          @update:cycle="handleCycleUpdate"
          @update:module="handleModuleUpdate"
          @update:release="handleReleaseUpdate"
          @update:start-date="(d: any) => instantUpdate('start_date', d + 'T00:00:00Z')"
          @update:target-date="(d: any) => instantUpdate('target_date', d + 'T00:00:00Z')"
          @update:labels="handleLabelsUpdate"
          @update:custom-field="updateCustomField"
          @dispatch-agent="(id: string) => dispatchAgent(id)"
        />
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
import { setIssueCycle, removeIssueCycle } from '@/api/issue'
import * as issueTypeApi from '@/api/issue-type'
import { releaseApi } from '@/api/release'
import projectApi from '@/api/project'
import api from '@/api'
import { agentApi } from '@/api/agent'
import { useConfirm } from '@/composables/useConfirm'
import { getIssueCustomFieldsWithDefinitions, updateIssueCustomFieldValue } from '@/api/custom-field'
import IssueDetailHeader from '@/components/IssueDetailHeader.vue'
import IssuePropertySidebar from '@/components/IssuePropertySidebar.vue'
import IssueTabDetails from '@/components/IssueTabDetails.vue'
import IssueTabRelations from '@/components/IssueTabRelations.vue'
import IssueTabAttachments from '@/components/IssueTabAttachments.vue'
import IssueTabTimeTracking from '@/components/IssueTabTimeTracking.vue'
import IssueTabActivity from '@/components/IssueTabActivity.vue'
import IssueGitPanel from '@/components/IssueGitPanel.vue'
import ApprovalSubmitDialog from '@/components/ApprovalSubmitDialog.vue'
import ApprovalDecisionDialog from '@/components/ApprovalDecisionDialog.vue'
import ApprovalPendingBanner from '@/components/ApprovalPendingBanner.vue'
import approvalApi, { type ApprovalResponse } from '@/api/approval'

// Route params
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { confirm } = useConfirm()
const toast = useToast()
const issueId = parseInt(route.params.issueId as string, 10) || 0
const projectId = ref(0)
const workspaceId = ref(0)
const projectIdentifier = ref('')
const projectLabels = ref<Array<{ id: number; name: string; color: string }>>([])

// Reactive state
const issue = ref<any>(null)
const saving = ref(false)
const issueForm = ref({ name: '', description: '' })
const states = ref<any[]>([])
const cycles = ref<any[]>([])
const modules = ref<any[]>([])
const releases = ref<{ id: number; name: string; version: string }[]>([])
const projectMembers = ref<any[]>([])
const issueTypes = ref<any[]>([])
const activeTab = ref('details')
const isWatching = ref(false)
const customFieldEntries = ref<Array<{ field: any; value: string | null }>>([])
const relationsTabRef = ref<InstanceType<typeof IssueTabRelations> | null>(null)
const agentDispatching = ref(false)
const activeApproval = ref<ApprovalResponse | null>(null)
const currentUserId = parseInt(localStorage.getItem('user_id') || '0', 10)
const showSubmitDialog = ref(false)
const showDecisionDialog = ref(false)
const submitDialogData = ref<{ transitionId: number; fromStateName: string; approveStateName: string; approverNames: string[]; workflowName?: string }>({
  transitionId: 0, fromStateName: '', approveStateName: '', approverNames: []
})
const decisionDialogData = ref<{ approvalId: number; decision: 'approved' | 'rejected' } | null>(null)

// Relation summary for sidebar
const relationSidebarSummary = computed(() => {
  return relationsTabRef.value?.relationSummary ?? null
})

// Tabs definition
const tabs = computed(() => [
  { key: 'details', label: t('issue.tabDetails'), count: undefined },
  { key: 'relations', label: t('issue.tabRelations'), count: relationSidebarSummary.value?.total ?? undefined },
  { key: 'attachments', label: t('issue.tabAttachments'), count: issue.value?.attachment_count || undefined },
  { key: 'git', label: t('gitIntegration.title'), count: undefined },
  { key: 'timetrack', label: t('issue.tabTimetrack'), count: undefined },
  { key: 'activity', label: t('issue.tabActivity'), count: undefined },
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
    projectIdentifier.value = issueData.project?.identifier || ''
    issueForm.value = {
      name: issueData.name || '',
      description: issueData.description_html || '',
    }

    // Load auxiliary data
    await Promise.all([
      loadStates(),
      loadCycles(),
      loadModules(),
      loadReleases(),
      loadProjectMembers(),
      loadIssueTypes(),
      loadCustomFields(),
      loadLabels(),
      loadWatchers(),
    ])
    // Load active approval if pending
    await loadActiveApproval()
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

async function loadReleases() {
  try {
    const data = await releaseApi.list(projectId.value)
    releases.value = data
  } catch (error) {
    console.error('Failed to load releases:', error)
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

async function loadCustomFields() {
  try {
    const response = await getIssueCustomFieldsWithDefinitions(issueId)
    if (response?.fields) {
      customFieldEntries.value = response.fields.map((item: any) => ({
        field: {
          id: item.id,
          name: item.name,
          field_type: item.field_type,
          options: item.options || [],
        },
        value: item.value ?? null,
      }))
    }
  } catch (error) {
    console.error('Failed to load custom fields:', error)
  }
}

async function loadLabels() {
  try {
    const resp = await api.get(`/projects/${projectId.value}/settings/labels`)
    projectLabels.value = Array.isArray(resp.data) ? resp.data : (resp.data?.items || [])
  } catch { /* */ }
}

async function handleLabelsUpdate(labelIds: number[]) {
  await instantUpdate('label_ids', labelIds)
}

async function deleteIssue() {
  if (!issue.value || !(await confirm(t('issueDetail.confirmDelete', { 0: issue.value.name })))) return
  try {
    await issueApi.deleteIssue(issueId)
    router.back()
  } catch (e: any) {
    toast.error(e?.response?.data?.message || t('issue.saveFailed'))
  }
}

async function updateCustomField(fieldId: number, value: string) {
  try {
    await updateIssueCustomFieldValue(issueId, fieldId, { value })
    const entry = customFieldEntries.value.find((e) => e.field.id === fieldId)
    if (entry) entry.value = value
  } catch (error) {
    console.error('Failed to update custom field:', error)
    toast.error(t('issue.saveFailed'))
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
async function instantUpdate(field: string, value: any) {
  try {
    const updated = await issueApi.updateIssue(issueId, { [field]: value })
    if (issue.value) Object.assign(issue.value, updated)
  } catch (e: any) {
    console.error('Failed to update field:', field, e)
    toast.error(t('issue.saveFailed'))
  }
}

// Load active approval if issue has pending approval
async function loadActiveApproval() {
  if (!issue.value) return
  if (issue.value.approval_status === 'pending' && issue.value.active_approval_id) {
    try {
      activeApproval.value = await approvalApi.get(issue.value.active_approval_id)
    } catch (e) {
      console.error('Failed to load active approval:', e)
      activeApproval.value = null
    }
  } else {
    activeApproval.value = null
  }
}

// Intercept state change for approval transitions
async function handleStateChange(newStateId: number) {
  if (!issue.value) return
  if (issue.value.approval_status === 'pending') {
    toast.error(t('approvals.stateDisabledHint'))
    return
  }
  try {
    const updated = await issueApi.updateIssue(issueId, { state_id: newStateId })
    if (issue.value) Object.assign(issue.value, updated)
    await loadActiveApproval()
  } catch (e: any) {
    // Check for approval_required response (HTTP 409)
    if (e?.response?.status === 409 && e?.response?.data?.message === 'approval_required') {
      const transitionId = e.response.data.transition_id
      const sourceStateId = e.response.data.source_state_id
      const targetStateId = e.response.data.target_state_id
      const workflowName = e.response.data.workflow_name
      const sourceStateName = e.response.data.source_state_name
      const targetStateName = e.response.data.target_state_name
      // Fetch the transition details to get approver names
      try {
        const wfRes = await api.get(`/projects/${issue.value.project_id}/workflows`)
        const workflows = wfRes.data || []
        let transition: any = null
        for (const w of workflows) {
          const found = (w.transitions || []).find((tr: any) => tr.id === transitionId)
          if (found) {
            transition = found
            break
          }
        }
        // Parse approver IDs and resolve names from projectMembers
        let approverNames: string[] = []
        if (transition?.approver_ids) {
          try {
            const ids: number[] = JSON.parse(transition.approver_ids)
            approverNames = ids.map(id => {
              const m = projectMembers.value.find((m: any) => m.user_id === id || m.id === id)
              return m?.user?.display_name || m?.display_name || `#${id}`
            })
          } catch { /* ignore parse errors */ }
        }
        submitDialogData.value = {
          transitionId,
          fromStateName: sourceStateName || `#${sourceStateId}`,
          approveStateName: targetStateName || `#${targetStateId}`,
          approverNames,
          workflowName,
        }
        showSubmitDialog.value = true
      } catch (fetchErr) {
        console.error('Failed to fetch transition details:', fetchErr)
        toast.error(t('issue.saveFailed'))
      }
    } else {
      console.error('Failed to update state:', e)
      toast.error(t('issue.saveFailed'))
    }
  }
}

async function onApprovalSubmitted() {
  showSubmitDialog.value = false
  // Reload issue to get updated approval_status
  try {
    const updated = await issueApi.getIssue(issueId)
    issue.value = updated
    await loadActiveApproval()
  } catch (e) {
    console.error('Failed to reload issue:', e)
  }
}

function onDecideApproval(approval: ApprovalResponse, decision: 'approved' | 'rejected') {
  decisionDialogData.value = { approvalId: approval.id, decision }
  showDecisionDialog.value = true
}

async function onApprovalDecided() {
  showDecisionDialog.value = false
  try {
    const updated = await issueApi.getIssue(issueId)
    issue.value = updated
    await loadActiveApproval()
  } catch (e) {
    console.error('Failed to reload issue:', e)
  }
}

async function onCancelApproval(approval: ApprovalResponse) {
  if (!confirm(t('approvals.cancelApproval'))) return
  try {
    await approvalApi.cancel(approval.id)
    const updated = await issueApi.getIssue(issueId)
    issue.value = updated
    await loadActiveApproval()
  } catch (e: any) {
    alert(e?.response?.data?.message || 'Failed to cancel approval')
  }
}

// Instant update for assignee
async function instantUpdateAssignee(userId: any) {
  try {
    const assigneeIds = userId ? [Number(userId)] : []
    const updated = await issueApi.updateIssue(issueId, { assignee_ids: assigneeIds })
    if (issue.value) Object.assign(issue.value, updated)
  } catch (e: any) {
    console.error('Failed to update assignee:', e)
    toast.error(t('issue.saveFailed'))
  }
}

// Cycle update uses dedicated API
async function handleCycleUpdate(cycleId: number | null) {
  try {
    if (cycleId) {
      await setIssueCycle(issueId, cycleId)
    } else {
      await removeIssueCycle(issueId)
    }
    const updated = await issueApi.getIssue(issueId)
    if (issue.value) Object.assign(issue.value, updated)
  } catch (e: any) {
    console.error('Failed to update cycle:', e)
    toast.error(t('issue.saveFailed'))
  }
}

// Module update — backend Update handles module_ids atomically with delete+recreate
async function handleModuleUpdate(moduleId: number | null) {
  await instantUpdate('module_ids', moduleId ? [moduleId] : [])
}

// Release update
async function handleReleaseUpdate(releaseId: number | null) {
  await instantUpdate('release_id', releaseId ?? 0)
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

async function handleRelationsRefresh() {
  try {
    const data = await issueApi.getIssue(issueId)
    if (issue.value) {
      Object.assign(issue.value, data)
    }
  } catch (err) {
    console.error('Failed to refresh issue after relations change:', err)
  }
}

async function loadWatchers() {
  try {
    const { watchers } = await issueApi.listWatchers(issueId)
    const currentUserId = parseInt(localStorage.getItem('user_id') || '0', 10)
    isWatching.value = watchers.includes(currentUserId)
  } catch (error) {
    console.error('Failed to load watchers:', error)
  }
}

async function handleToggleWatch() {
  try {
    if (isWatching.value) {
      await issueApi.removeWatcher(issueId)
      isWatching.value = false
      toast.success(t('issue.unwatch'))
    } else {
      await issueApi.addWatcher(issueId)
      isWatching.value = true
      toast.success(t('issue.watch'))
    }
  } catch (error: any) {
    console.error('Failed to toggle watch:', error)
    toast.error(error?.response?.data?.message || t('common.error'))
  }
}

// Agent dispatch
async function dispatchAgent(agentSelectorId: string) {
  if (!agentSelectorId || !agentSelectorId.startsWith('agent:')) return
  const agentId = parseInt(agentSelectorId.replace('agent:', ''))
  if (!agentId) return
  agentDispatching.value = true
  try {
    await agentApi.dispatch(workspaceId.value, agentId, {
      task: `Analyze issue #${issue.value?.sequence_id || issueId}: ${issue.value?.name || 'Untitled'}`,
      issue_id: issueId,
      project_id: projectId.value,
    })
    toast.success(t('agent.dispatch'))
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
