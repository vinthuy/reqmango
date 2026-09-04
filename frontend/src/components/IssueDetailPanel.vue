<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex">
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-black/30" @click="close"></div>

      <!-- Slide-in panel - wide enough for tabs + sidebar -->
      <div class="relative ml-auto w-[90vw] max-w-5xl bg-white shadow-2xl flex flex-col h-full overflow-hidden animate-slide-in">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center h-full">
          <svg class="animate-spin h-8 w-8 text-indigo-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
          </svg>
        </div>

        <template v-else-if="issue">
          <!-- Header bar -->
          <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-200 shrink-0 bg-gray-50/50">
            <button @click="close" class="p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-200 rounded transition-colors shrink-0">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
            </button>
            <span class="text-[11px] text-gray-400 font-mono shrink-0">{{ projectIdentifier }}-{{ issue.sequence_id }}</span>
            <!-- Issue type badge (before title) -->
            <span
              v-if="issue.issue_type"
              class="px-1.5 py-0.5 rounded text-[10px] font-medium shrink-0"
              :style="{ backgroundColor: (issue.issue_type.color || '#e5e7eb') + '20', color: issue.issue_type.color || '#6b7280' }"
            >{{ issue.issue_type.name }}</span>
            <input
              :value="issue.name"
              class="flex-1 text-sm font-semibold text-gray-800 bg-transparent border-0 outline-none focus:bg-white focus:border focus:border-indigo-300 focus:rounded focus:px-1.5 focus:py-0.5 min-w-0"
              @blur="(e: FocusEvent) => handleTitleChange((e.target as HTMLInputElement).value)"
              @keydown.enter="(e: KeyboardEvent) => { (e.target as HTMLInputElement).blur() }"
            />
            <span v-if="saving" class="text-[10px] text-indigo-500 animate-pulse shrink-0">{{ t('issue.saving') }}</span>
            <!-- Open in full page -->
            <a
              :href="`/workspace/${workspaceSlug}/project/${projectId}/issues/${issue.id}`"
              class="p-1 text-gray-400 hover:text-indigo-500 rounded transition-colors shrink-0"
              :title="t('issue.openFullPage') || 'Open full page'"
              @click.prevent="navigateTo(issue.id)"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/></svg>
            </a>
          </div>

          <ApprovalPendingBanner
            :approval="activeApproval"
            :current-user-id="currentUserId"
            @decide="onDecideApproval"
            @cancel="onCancelApproval"
          />

          <!-- Body: tabs + sidebar -->
          <div class="flex-1 flex overflow-hidden">
            <!-- Main content area -->
            <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
              <!-- Tab buttons -->
              <div class="flex border-b border-gray-200 shrink-0 px-4">
                <button
                  v-for="tab in tabs"
                  :key="tab.key"
                  @click="activeTab = tab.key"
                  :class="activeTab === tab.key ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500'"
                  class="px-3 py-2 text-xs font-medium border-b-2 -mb-[1px] transition-colors"
                >
                  {{ tab.label }}
                  <span v-if="tab.count !== undefined && tab.count > 0" class="ml-1 text-[10px] bg-gray-100 text-gray-500 px-1 py-0.5 rounded-full">{{ tab.count }}</span>
                </button>
              </div>

              <!-- Tab panels -->
              <div class="flex-1 overflow-y-auto px-4 py-3">
                <IssueTabDetails
                  v-if="activeTab === 'details'"
                  :issue-id="issue.id"
                  :issue="issue"
                  :workspace-id="workspaceId"
                  :project-id="projectId"
                  :issue-type-id="issue.issue_type?.id"
                  :members="projectMembers"
                  :show-title="false"
                  @update:title="(v: string) => quickUpdate('name', v)"
                  @update:description="(v: string) => quickUpdate('description_html', v)"
                />
                <IssueTabRelations
                  v-else-if="activeTab === 'relations'"
                  ref="relationsTabRef"
                  :issue-id="issue.id"
                  :project-id="projectId"
                  :workspace-id="workspaceId"
                  :slug="workspaceSlug"
                  :states="stateOptions"
                  :parent="issue.parent"
                  :sub-issues="issue.sub_issues || []"
                  :issue-types="issueTypeOptions"
                  @navigate="navigateTo"
                  @refresh="reloadIssue"
                />
                <IssueTabAttachments
                  v-else-if="activeTab === 'attachments'"
                  :issue-id="issue.id"
                  :project-id="projectId"
                />
                <IssueTabTimeTracking
                  v-else-if="activeTab === 'timetrack'"
                  :issue-id="issue.id"
                />
                <IssueTabActivity
                  v-else-if="activeTab === 'activity'"
                  :issue-id="issue.id"
                />
              </div>
            </div>

            <!-- Right sidebar -->
            <div class="shrink-0 border-l border-gray-200 overflow-y-auto px-3 py-3">
              <IssuePropertySidebar
                v-if="issue"
                :issue="issue"
                :states="stateOptions"
                :members="projectMembers"
                :cycles="cycleOptions"
                :modules="moduleOptions"
                :releases="releaseOptions"
                :custom-fields="customFieldEntries"
                :workspace-id="workspaceId"
                :agent-dispatching="agentDispatching"
                :labels="projectLabels"
                :relation-summary="relationSidebarSummary"
                @update:state="handleStateChange"
                @update:priority="(p: any) => quickUpdate('priority', p)"
                @update:assignee="quickUpdateAssignee"
                @update:cycle="quickUpdateCycle"
                @update:module="quickUpdateModule"
                @update:release="quickUpdateRelease"
                @update:start-date="(d: any) => quickUpdate('start_date', d + 'T00:00:00Z')"
                @update:target-date="(d: any) => quickUpdate('target_date', d + 'T00:00:00Z')"
                @update:labels="handleLabelsUpdate"
                @update:custom-field="updateCustomField"
                @dispatch-agent="dispatchAgent"
              />
            </div>
          </div>
        </template>

        <!-- Approval Submit Dialog -->
        <ApprovalSubmitDialog
          :show="showSubmitDialog"
          :issue-id="issue?.id || 0"
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
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import issueApi from '@/api/issue'
import * as issueTypeApi from '@/api/issue-type'
import * as stateApi from '@/api/project-settings'
import * as cycleApi from '@/api/cycle'
import * as moduleApi from '@/api/module'
import { releaseApi } from '@/api/release'
import api from '@/api'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { agentApi } from '@/api/agent'
import { getIssueCustomFieldsWithDefinitions, updateIssueCustomFieldValue } from '@/api/custom-field'
import IssuePropertySidebar from '@/components/IssuePropertySidebar.vue'
import IssueTabDetails from '@/components/IssueTabDetails.vue'
import IssueTabRelations from '@/components/IssueTabRelations.vue'
import IssueTabAttachments from '@/components/IssueTabAttachments.vue'
import IssueTabTimeTracking from '@/components/IssueTabTimeTracking.vue'
import IssueTabActivity from '@/components/IssueTabActivity.vue'
import ApprovalSubmitDialog from '@/components/ApprovalSubmitDialog.vue'
import ApprovalDecisionDialog from '@/components/ApprovalDecisionDialog.vue'
import ApprovalPendingBanner from '@/components/ApprovalPendingBanner.vue'
import approvalApi, { type ApprovalResponse } from '@/api/approval'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const toast = useToast()

const workspaceSlug = computed(() => (route.params as any).slug as string || '')

const props = defineProps<{
  issueId: number | null
  visible: boolean
  workspaceId: number
  projectId: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'delete', issue: any): void
  (e: 'refresh'): void
}>()

const issue = ref<any>(null)
const loading = ref(false)
const activeTab = ref('details')

const stateOptions = ref<any[]>([])
const cycleOptions = ref<any[]>([])
const moduleOptions = ref<any[]>([])
const releaseOptions = ref<{ id: number; name: string; version: string }[]>([])
const issueTypeOptions = ref<any[]>([])
const projectMembers = ref<any[]>([])
const projectIdentifier = ref('')
const customFieldEntries = ref<Array<{ field: any; value: string | null }>>([])
const relationsTabRef = ref<InstanceType<typeof IssueTabRelations> | null>(null)
const agentDispatching = ref(false)
const saving = ref(false)
const projectLabels = ref<Array<{ id: number; name: string; color: string }>>([])
const showSubmitDialog = ref(false)
const submitDialogData = ref<{ transitionId: number; fromStateName: string; approveStateName: string; approverNames: string[]; workflowName?: string }>({
  transitionId: 0, fromStateName: '', approveStateName: '', approverNames: []
})
const activeApproval = ref<ApprovalResponse | null>(null)
const currentUserId = parseInt(localStorage.getItem('user_id') || '0', 10)
const showDecisionDialog = ref(false)
const decisionDialogData = ref<{ approvalId: number; decision: 'approved' | 'rejected' } | null>(null)

async function handleTitleChange(newTitle: string) {
  if (!issue.value || newTitle === issue.value.name) return
  saving.value = true
  try {
    const updated = await issueApi.updateIssue(issue.value.id, { name: newTitle })
    if (issue.value) Object.assign(issue.value, updated)
  } catch (e: any) {
    toast.error(e?.response?.data?.message || t('issue.saveFailed'))
  } finally { saving.value = false }
}

const relationSidebarSummary = computed(() => {
  return relationsTabRef.value?.relationSummary ?? null
})

const tabs = computed(() => [
  { key: 'details', label: t('issue.tabDetails'), count: undefined },
  { key: 'relations', label: t('issue.tabRelations'), count: relationSidebarSummary.value?.total ?? undefined },
  { key: 'attachments', label: t('issue.tabAttachments'), count: issue.value?.attachment_count || undefined },
  { key: 'timetrack', label: t('issue.tabTimetrack'), count: undefined },
  { key: 'activity', label: t('issue.tabActivity'), count: undefined },
])

function navigateTo(issueId: number) {
  const slug = workspaceSlug.value
  if (slug) {
    router.push(`/workspace/${slug}/project/${props.projectId}/issues/${issueId}`)
  }
}

// ---- Data loading ----
watch(() => [props.issueId, props.visible] as const, async ([id, vis]) => {
  if (id && vis) {
    loading.value = true
    activeTab.value = 'details'
    try {
      const result = await issueApi.getIssue(id)
      issue.value = result
      projectIdentifier.value = result.project?.identifier || ''
      await Promise.all([
        loadStates(),
        loadCycles(),
        loadModules(),
        loadReleases(),
        loadIssueTypes(),
        loadMembers(),
        loadCustomFields(),
        loadLabels(),
      ])
      await loadActiveApproval()
    } catch (e) {
      console.error('Failed to load issue:', e)
      issue.value = null
    } finally {
      loading.value = false
    }
  } else if (!vis) {
    issue.value = null
    activeApproval.value = null
  }
}, { immediate: true })

async function loadStates() {
  try { stateOptions.value = await stateApi.listStates(props.projectId) } catch { /* */ }
}
async function loadCycles() {
  try { const d = await cycleApi.listCycles(props.projectId); cycleOptions.value = d?.items || d || [] } catch { /* */ }
}
async function loadModules() {
  try { moduleOptions.value = await moduleApi.listModules(props.projectId, props.workspaceId) } catch { /* */ }
}
async function loadReleases() {
  try { releaseOptions.value = await releaseApi.list(props.projectId) } catch { /* */ }
}
async function loadIssueTypes() {
  try { issueTypeOptions.value = await issueTypeApi.getIssueTypes(props.workspaceId, props.projectId) } catch { /* */ }
}
async function loadMembers() {
  try {
    const d = await api.get(`/projects/${props.projectId}/members`)
    projectMembers.value = (d.data || []).map((m: any) => m.user || m)
  } catch { /* */ }
}
async function loadCustomFields() {
  try {
    const resp = await getIssueCustomFieldsWithDefinitions(props.issueId!)
    if (resp?.fields) {
      customFieldEntries.value = resp.fields.map((item: any) => ({
        field: { id: item.id, name: item.name, field_type: item.field_type, options: item.options || [] },
        value: item.value ?? null,
      }))
    }
  } catch { /* */ }
}

async function reloadIssue() {
  if (!issue.value) return
  try {
    const data = await issueApi.getIssue(issue.value.id)
    Object.assign(issue.value, data)
  } catch { /* */ }
}

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

// ---- State change with approval handling ----
async function handleStateChange(newStateId: number) {
  if (!issue.value) return
  if (issue.value.approval_status === 'pending') {
    toast.error(t('approvals.stateDisabledHint'))
    return
  }
  try {
    const updated = await issueApi.updateIssue(issue.value.id, { state_id: newStateId })
    if (issue.value) Object.assign(issue.value, updated)
    emit('refresh')
  } catch (e: any) {
    if (e?.response?.status === 409 && e?.response?.data?.message === 'approval_required') {
      const transitionId = e.response.data.transition_id
      const sourceStateId = e.response.data.source_state_id
      const targetStateId = e.response.data.target_state_id
      const workflowName = e.response.data.workflow_name
      const sourceStateName = e.response.data.source_state_name
      const targetStateName = e.response.data.target_state_name
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
      toast.error(e?.response?.data?.message || t('issue.saveFailed'))
    }
  }
}

async function onApprovalSubmitted() {
  showSubmitDialog.value = false
  await reloadIssue()
  await loadActiveApproval()
}

function onDecideApproval(approval: ApprovalResponse, decision: 'approved' | 'rejected') {
  decisionDialogData.value = { approvalId: approval.id, decision }
  showDecisionDialog.value = true
}

async function onApprovalDecided() {
  showDecisionDialog.value = false
  decisionDialogData.value = null
  await reloadIssue()
  await loadActiveApproval()
}

async function onCancelApproval(approval: ApprovalResponse) {
  if (!confirm(t('approvals.cancelApproval'))) return
  try {
    await approvalApi.cancel(approval.id)
    await reloadIssue()
    await loadActiveApproval()
  } catch (e: any) {
    alert(e?.response?.data?.message || 'Failed to cancel approval')
  }
}

// ---- Quick updates (same pattern as IssueDetail.vue) ----
async function quickUpdate(field: string, value: any) {
  try {
    const updated = await issueApi.updateIssue(issue.value.id, { [field]: value })
    if (issue.value) Object.assign(issue.value, updated)
    emit('refresh')
  } catch (e: any) {
    toast.error(e?.response?.data?.message || t('issue.saveFailed'))
  }
}

async function quickUpdateAssignee(userId: any) {
  try {
    const assigneeIds = userId ? [Number(userId)] : []
    const updated = await issueApi.updateIssue(issue.value.id, { assignee_ids: assigneeIds })
    if (issue.value) Object.assign(issue.value, updated)
    emit('refresh')
  } catch (e: any) { toast.error(t('issue.saveFailed')) }
}

async function quickUpdateCycle(cycleId: number | null) {
  try {
    if (cycleId) {
      await issueApi.setIssueCycle(issue.value.id, cycleId)
    } else {
      await issueApi.removeIssueCycle(issue.value.id)
    }
    await reloadIssue()
  } catch (e: any) { toast.error(t('issue.saveFailed')) }
}

async function quickUpdateModule(moduleId: number | null) {
  await quickUpdate('module_ids', moduleId ? [moduleId] : [])
}

async function quickUpdateRelease(releaseId: number | null) {
  await quickUpdate('release_id', releaseId ?? 0)
}

async function updateCustomField(fieldId: number, value: string) {
  try {
    await updateIssueCustomFieldValue(issue.value.id, fieldId, { value })
    const entry = customFieldEntries.value.find(e => e.field.id === fieldId)
    if (entry) entry.value = value
  } catch { toast.error(t('issue.saveFailed')) }
}

async function loadLabels() {
  try {
    const resp = await api.get(`/projects/${props.projectId}/settings/labels`)
    projectLabels.value = Array.isArray(resp.data) ? resp.data : (resp.data?.items || [])
  } catch { /* */ }
}

async function handleLabelsUpdate(labelIds: number[]) {
  await quickUpdate('label_ids', labelIds)
}

async function dispatchAgent(agentSelectorId: string) {
  if (!agentSelectorId || !agentSelectorId.startsWith('agent:')) return
  const agentId = parseInt(agentSelectorId.replace('agent:', ''))
  if (!agentId) return
  agentDispatching.value = true
  try {
    await agentApi.dispatch(props.workspaceId, agentId, {
      task: `Analyze issue #${issue.value?.sequence_id}: ${issue.value?.name || 'Untitled'}`,
      issue_id: issue.value.id,
      project_id: props.projectId,
    })
    toast.success(t('agent.dispatch'))
  } catch (e: any) {
    toast.error(e?.response?.data?.message || e?.message || 'Failed to dispatch agent')
  } finally { agentDispatching.value = false }
}

function close() {
  emit('close')
}
</script>

<style scoped>
@keyframes slideIn {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}
.animate-slide-in {
  animation: slideIn 0.25s ease-out;
}
</style>
