<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useRoute } from 'vue-router'
import {
  Badge,
  Modal,
  Tag,
  Tabs,
  TabPane,
  Progress,
  Select,
  Input,
  Textarea,
  Form,
  FormItem,
  DatePicker,
  Spin,
  Space,
  Tooltip,
  Popconfirm,
  Dropdown,
  Menu,
  MenuItem,
  message
} from 'ant-design-vue'
import type { Dayjs } from 'dayjs'
import { useI18n } from '@/composables/useI18n'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { listIssues, type IssueResponse } from '@/api/issue'
import {
  issueAgentApi,
  type AgentStatus,
  type AssignAgentRequest,
  type CompleteWorkRequest,
  type EscalateRequest,
  type UpdateStatusRequest
} from '@/api/issue-agent'
import { agentMemberApi, type AgentMember } from '@/api/agent-member'

const route = useRoute()
const { t } = useI18n()
const { getWorkspaceId } = useWorkspaceId()

const projectId = computed(() => parseInt(route.params.projectId as string, 10) || parseInt(route.params.id as string, 10))

interface IssueAgentRow {
  issue: IssueResponse
  agentStatus: AgentStatus | null
}

const rows = ref<IssueAgentRow[]>([])
const agentMembers = ref<AgentMember[]>([])
const loading = ref(false)
const activeTab = ref('all')
const projectIdentifier = ref('')

// Assign modal
const assignModalVisible = ref(false)
const assignTarget = ref<IssueAgentRow | null>(null)
const assignForm = reactive({
  agent_member_id: undefined as number | undefined,
  priority: 'p2',
  deadline: null as Dayjs | null,
  notes: ''
})
const assignLoading = ref(false)

// Complete modal
const completeModalVisible = ref(false)
const completeTarget = ref<IssueAgentRow | null>(null)
const completeForm = reactive({
  summary: '',
  decision_ids: ''
})
const completeLoading = ref(false)

// Escalate modal
const escalateModalVisible = ref(false)
const escalateTarget = ref<IssueAgentRow | null>(null)
const escalateForm = reactive({
  reason: '',
  escalation_type: 'needs_human' as EscalateRequest['escalation_type']
})
const escalateLoading = ref(false)

// Quick status update modal
const statusModalVisible = ref(false)
const statusTarget = ref<IssueAgentRow | null>(null)
const statusForm = reactive({
  status: ''
})
const statusLoading = ref(false)

// Action loading states
const actionLoadingId = ref<number | null>(null)

const statusTabItems = [
  { key: 'all', tab: 'All' },
  { key: 'pending', tab: 'Pending' },
  { key: 'in_progress', tab: 'In Progress' },
  { key: 'completed', tab: 'Completed' },
  { key: 'escalated', tab: 'Escalated' },
  { key: 'failed', tab: 'Failed' },
  { key: 'cancelled', tab: 'Cancelled' }
]

const priorityOptions = [
  { value: 'p0', label: 'P0 - Critical' },
  { value: 'p1', label: 'P1 - High' },
  { value: 'p2', label: 'P2 - Medium' },
  { value: 'p3', label: 'P3 - Low' },
  { value: 'p4', label: 'P4 - Lowest' }
]

const escalationTypeOptions = [
  { value: 'needs_human', label: 'Needs Human Intervention' },
  { value: 'approval', label: 'Requires Approval' },
  { value: 'resource_limit', label: 'Resource Limit' },
  { value: 'deadline_risk', label: 'Deadline Risk' }
]

const quickStatusOptions = [
  { value: 'pending', label: 'Pending' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'cancelled', label: 'Cancelled' }
]

const activeAgentMembers = computed(() =>
  agentMembers.value.filter(m => m.is_active)
)

const filteredRows = computed(() => {
  if (activeTab.value === 'all') return rows.value
  return rows.value.filter(r => {
    const status = r.agentStatus?.task_status
    return status === activeTab.value
  })
})

const tabCounts = computed(() => {
  const counts: Record<string, number> = { all: rows.value.length }
  for (const row of rows.value) {
    const status = row.agentStatus?.task_status || 'pending'
    counts[status] = (counts[status] || 0) + 1
  }
  return counts
})

const tableColumns = [
  {
    title: 'Issue ID',
    key: 'issueId',
    width: 140
  },
  {
    title: 'Title',
    key: 'title',
    ellipsis: true
  },
  {
    title: 'Priority',
    key: 'priority',
    width: 120,
    sorter: (a: IssueAgentRow, b: IssueAgentRow) => {
      const order: Record<string, number> = { urgent: 0, high: 1, medium: 2, low: 3, none: 4 }
      return (order[a.issue.priority] ?? 5) - (order[b.issue.priority] ?? 5)
    }
  },
  {
    title: 'Assigned Agent',
    key: 'agent',
    width: 200
  },
  {
    title: 'Status',
    key: 'status',
    width: 130
  },
  {
    title: 'Progress',
    key: 'progress',
    width: 180
  },
  {
    title: 'Actions',
    key: 'actions',
    width: 280,
    fixed: 'right' as const
  }
]

function getPriorityColor(priority: string): string {
  const colors: Record<string, string> = {
    urgent: 'red',
    high: 'orange',
    medium: 'blue',
    low: 'green',
    none: 'default'
  }
  return colors[priority] || 'default'
}

function getPriorityLabel(priority: string): string {
  const labels: Record<string, string> = {
    urgent: 'Urgent',
    high: 'High',
    medium: 'Medium',
    low: 'Low',
    none: 'None'
  }
  return labels[priority] || priority
}

function getStatusColor(status: string | undefined): string {
  const colors: Record<string, string> = {
    pending: 'default',
    in_progress: 'processing',
    completed: 'success',
    escalated: 'warning',
    failed: 'error',
    cancelled: 'default'
  }
  return colors[status || 'pending'] || 'default'
}

function getStatusLabel(status: string | undefined): string {
  const labels: Record<string, string> = {
    pending: 'Pending',
    in_progress: 'In Progress',
    completed: 'Completed',
    escalated: 'Escalated',
    failed: 'Failed',
    cancelled: 'Cancelled'
  }
  return labels[status || 'pending'] || status || 'Unknown'
}

function getAgentTypeColor(type: string): string {
  const colors: Record<string, string> = {
    autopilot: 'purple',
    loop: 'cyan',
    squad: 'geekblue',
    skill: 'magenta'
  }
  return colors[type] || 'default'
}

async function loadData() {
  loading.value = true
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) {
      message.error('Could not resolve workspace')
      return
    }
    const issueResult = await listIssues(projectId.value, wsId, {
      limit: 200,
      sort_by: 'created_at',
      sort_dir: 'desc'
    })
    const issues = issueResult.items || []
    if (issues.length > 0 && issues[0].project?.identifier) {
      projectIdentifier.value = issues[0].project.identifier
    }

    const membersRes = await agentMemberApi.list(projectId.value)
    agentMembers.value = membersRes?.data || []

    const statusPromises = issues.map(issue =>
      issueAgentApi.getStatus(issue.id).then(res => res.data).catch(() => null)
    )
    const statuses = await Promise.all(statusPromises)

    rows.value = issues.map((issue, idx) => ({
      issue,
      agentStatus: statuses[idx]
    }))
  } catch (err) {
    console.error('Failed to load issue-agent data:', err)
    message.error('Failed to load data')
  } finally {
    loading.value = false
  }
}

function openAssignModal(row: IssueAgentRow) {
  assignTarget.value = row
  assignForm.agent_member_id = undefined
  assignForm.priority = 'p2'
  assignForm.deadline = null
  assignForm.notes = ''
  assignModalVisible.value = true
}

async function handleAssign() {
  if (!assignTarget.value || !assignForm.agent_member_id) {
    message.warning('Please select an agent')
    return
  }
  assignLoading.value = true
  try {
    const agentMember = agentMembers.value.find(m => m.id === assignForm.agent_member_id)
    const data: AssignAgentRequest = {
      agent_id: agentMember?.agent_id || assignForm.agent_member_id,
      priority: assignForm.priority,
      deadline: assignForm.deadline ? assignForm.deadline.toISOString() : undefined,
      notes: assignForm.notes || undefined
    }
    await issueAgentApi.assign(assignTarget.value.issue.id, data)
    message.success('Agent assigned successfully')
    assignModalVisible.value = false
    await loadData()
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to assign agent')
  } finally {
    assignLoading.value = false
  }
}

async function handleStartWork(row: IssueAgentRow) {
  actionLoadingId.value = row.issue.id
  try {
    await issueAgentApi.startWork(row.issue.id)
    message.success('Work started')
    await loadData()
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to start work')
  } finally {
    actionLoadingId.value = null
  }
}

function openCompleteModal(row: IssueAgentRow) {
  completeTarget.value = row
  completeForm.summary = ''
  completeForm.decision_ids = ''
  completeModalVisible.value = true
}

async function handleComplete() {
  if (!completeTarget.value) return
  completeLoading.value = true
  try {
    const decisionIds = completeForm.decision_ids
      ? completeForm.decision_ids.split(',').map(s => parseInt(s.trim(), 10)).filter(n => !isNaN(n))
      : []
    const data: CompleteWorkRequest = {
      summary: completeForm.summary || undefined,
      decision_ids: decisionIds.length > 0 ? decisionIds : undefined
    }
    await issueAgentApi.completeWork(completeTarget.value.issue.id, data)
    message.success('Work completed')
    completeModalVisible.value = false
    await loadData()
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to complete work')
  } finally {
    completeLoading.value = false
  }
}

function openEscalateModal(row: IssueAgentRow) {
  escalateTarget.value = row
  escalateForm.reason = ''
  escalateForm.escalation_type = 'needs_human'
  escalateModalVisible.value = true
}

async function handleEscalate() {
  if (!escalateTarget.value || !escalateForm.reason) {
    message.warning('Please provide a reason')
    return
  }
  escalateLoading.value = true
  try {
    const data: EscalateRequest = {
      reason: escalateForm.reason,
      escalation_type: escalateForm.escalation_type
    }
    await issueAgentApi.escalate(escalateTarget.value.issue.id, data)
    message.success('Issue escalated')
    escalateModalVisible.value = false
    await loadData()
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to escalate')
  } finally {
    escalateLoading.value = false
  }
}

function openStatusModal(row: IssueAgentRow) {
  statusTarget.value = row
  statusForm.status = row.agentStatus?.task_status || 'pending'
  statusModalVisible.value = true
}

async function handleStatusUpdate() {
  if (!statusTarget.value || !statusForm.status) {
    message.warning('Please select a status')
    return
  }
  statusLoading.value = true
  try {
    const data: UpdateStatusRequest = { status: statusForm.status }
    await issueAgentApi.updateStatus(statusTarget.value.issue.id, data)
    message.success('Status updated')
    statusModalVisible.value = false
    await loadData()
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to update status')
  } finally {
    statusLoading.value = false
  }
}

async function handleUnassign(row: IssueAgentRow) {
  actionLoadingId.value = row.issue.id
  try {
    await issueAgentApi.unassign(row.issue.id)
    message.success('Agent unassigned')
    await loadData()
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to unassign agent')
  } finally {
    actionLoadingId.value = null
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="p-6 max-w-7xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Issue-Agent Assignments</h1>
        <p class="text-sm text-gray-500 mt-1">Manage agent assignments and track task progress</p>
      </div>
      <a-button type="primary" @click="loadData" :loading="loading">
        Refresh
      </a-button>
    </div>

    <a-tabs v-model:activeKey="activeTab" class="mb-4">
      <a-tab-pane
        v-for="item in statusTabItems"
        :key="item.key"
      >
        <template #tab>
          {{ item.tab }}
          <a-badge
            :count="tabCounts[item.key] || 0"
            :number-style="{ backgroundColor: item.key === 'all' ? '#1890ff' : undefined }"
            show-zero
          />
        </template>
      </a-tab-pane>
    </a-tabs>

    <a-spin :spinning="loading">
      <a-table
        :columns="tableColumns"
        :data-source="filteredRows"
        :row-key="(record: IssueAgentRow) => record.issue.id"
        :pagination="{ pageSize: 20, showSizeChanger: true, showTotal: (total: number) => `Total ${total} issues` }"
        :scroll="{ x: 1200 }"
        size="middle"
        :locale="{ emptyText: 'No issues found' }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'issueId'">
            <span class="font-mono text-sm font-medium text-indigo-600">
              {{ projectIdentifier }}-{{ (record as IssueAgentRow).issue.sequence_id }}
            </span>
          </template>

          <template v-if="column.key === 'title'">
            <span class="text-gray-900">{{ (record as IssueAgentRow).issue.name }}</span>
          </template>

          <template v-if="column.key === 'priority'">
            <a-tag :color="getPriorityColor((record as IssueAgentRow).issue.priority)">
              {{ getPriorityLabel((record as IssueAgentRow).issue.priority) }}
            </a-tag>
          </template>

          <template v-if="column.key === 'agent'">
            <template v-if="(record as IssueAgentRow).agentStatus?.agent_name">
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-900">{{ (record as IssueAgentRow).agentStatus!.agent_name }}</span>
                <a-tag :color="getAgentTypeColor((record as IssueAgentRow).agentStatus!.task_status)" size="small">
                  agent
                </a-tag>
              </div>
            </template>
            <template v-else>
              <span class="text-sm text-gray-400 italic">Unassigned</span>
            </template>
          </template>

          <template v-if="column.key === 'status'">
            <a-tag :color="getStatusColor((record as IssueAgentRow).agentStatus?.task_status)">
              {{ getStatusLabel((record as IssueAgentRow).agentStatus?.task_status) }}
            </a-tag>
          </template>

          <template v-if="column.key === 'progress'">
            <a-progress
              :percent="(record as IssueAgentRow).agentStatus?.task_progress || 0"
              :size="'small'"
              :status="(record as IssueAgentRow).agentStatus?.task_status === 'failed' ? 'exception' : (record as IssueAgentRow).agentStatus?.task_status === 'completed' ? 'success' : 'active'"
            />
          </template>

          <template v-if="column.key === 'actions'">
            <a-space :size="4">
              <a-tooltip title="Assign Agent" v-if="!(record as IssueAgentRow).agentStatus?.agent_id">
                <a-button
                  type="primary"
                  size="small"
                  @click="openAssignModal(record as IssueAgentRow)"
                >
                  Assign Agent
                </a-button>
              </a-tooltip>

              <a-tooltip title="Start Work" v-if="(record as IssueAgentRow).agentStatus?.agent_id && (record as IssueAgentRow).agentStatus?.task_status === 'pending'">
                <a-button
                  type="primary"
                  size="small"
                  ghost
                  :loading="actionLoadingId === (record as IssueAgentRow).issue.id"
                  @click="handleStartWork(record as IssueAgentRow)"
                >
                  Start Work
                </a-button>
              </a-tooltip>

              <a-tooltip title="Complete" v-if="(record as IssueAgentRow).agentStatus?.task_status === 'in_progress'">
                <a-button
                  type="primary"
                  size="small"
                  style="background-color: #52c41a; border-color: #52c41a"
                  @click="openCompleteModal(record as IssueAgentRow)"
                >
                  Complete
                </a-button>
              </a-tooltip>

              <a-tooltip title="Escalate" v-if="(record as IssueAgentRow).agentStatus?.task_status === 'in_progress'">
                <a-button
                  size="small"
                  danger
                  @click="openEscalateModal(record as IssueAgentRow)"
                >
                  Escalate
                </a-button>
              </a-tooltip>

              <a-tooltip title="Unassign" v-if="(record as IssueAgentRow).agentStatus?.agent_id">
                <a-popconfirm
                  title="Are you sure you want to unassign this agent?"
                  @confirm="handleUnassign(record as IssueAgentRow)"
                >
                  <a-button
                    size="small"
                    :loading="actionLoadingId === (record as IssueAgentRow).issue.id"
                  >
                    Unassign
                  </a-button>
                </a-popconfirm>
              </a-tooltip>

              <a-dropdown>
                <a-button size="small">
                  Status
                </a-button>
                <template #overlay>
                  <a-menu @click="({ key }) => { statusTarget = record as IssueAgentRow; statusForm.status = key as string; handleStatusUpdate(); }">
                    <a-menu-item key="pending">Pending</a-menu-item>
                    <a-menu-item key="in_progress">In Progress</a-menu-item>
                    <a-menu-item key="completed">Completed</a-menu-item>
                    <a-menu-item key="escalated">Escalated</a-menu-item>
                    <a-menu-item key="failed">Failed</a-menu-item>
                    <a-menu-item key="cancelled">Cancelled</a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-spin>

    <!-- Assign Agent Modal -->
    <a-modal
      v-model:open="assignModalVisible"
      title="Assign Agent"
      :confirm-loading="assignLoading"
      @ok="handleAssign"
      @cancel="assignModalVisible = false"
      width="520px"
    >
      <a-form layout="vertical" class="mt-4">
        <a-form-item label="Agent" required>
          <a-select
            v-model:value="assignForm.agent_member_id"
            placeholder="Select an agent"
            :options="activeAgentMembers.map(m => ({ value: m.id, label: `${m.agent_name} (${m.agent_type})` }))"
          />
        </a-form-item>
        <a-form-item label="Priority">
          <a-select
            v-model:value="assignForm.priority"
            :options="priorityOptions"
          />
        </a-form-item>
        <a-form-item label="Deadline">
          <a-date-picker
            v-model:value="assignForm.deadline"
            show-time
            format="YYYY-MM-DD HH:mm"
            placeholder="Select deadline"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="Notes">
          <a-textarea
            v-model:value="assignForm.notes"
            :rows="3"
            placeholder="Optional notes about this assignment"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Complete Work Modal -->
    <a-modal
      v-model:open="completeModalVisible"
      title="Complete Work"
      :confirm-loading="completeLoading"
      @ok="handleComplete"
      @cancel="completeModalVisible = false"
      ok-text="Mark Complete"
      width="520px"
    >
      <a-form layout="vertical" class="mt-4">
        <a-form-item label="Summary">
          <a-textarea
            v-model:value="completeForm.summary"
            :rows="4"
            placeholder="Provide a summary of the work completed"
          />
        </a-form-item>
        <a-form-item label="Decision IDs">
          <a-input
            v-model:value="completeForm.decision_ids"
            placeholder="Comma-separated decision IDs (e.g. 1,2,3)"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Escalate Modal -->
    <a-modal
      v-model:open="escalateModalVisible"
      title="Escalate Issue"
      :confirm-loading="escalateLoading"
      @ok="handleEscalate"
      @cancel="escalateModalVisible = false"
      ok-text="Escalate"
      :ok-button-props="{ danger: true }"
      width="520px"
    >
      <a-form layout="vertical" class="mt-4">
        <a-form-item label="Escalation Type" required>
          <a-select
            v-model:value="escalateForm.escalation_type"
            :options="escalationTypeOptions"
          />
        </a-form-item>
        <a-form-item label="Reason" required>
          <a-textarea
            v-model:value="escalateForm.reason"
            :rows="4"
            placeholder="Explain why this issue needs to be escalated"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Quick Status Update Modal -->
    <a-modal
      v-model:open="statusModalVisible"
      title="Update Status"
      :confirm-loading="statusLoading"
      @ok="handleStatusUpdate"
      @cancel="statusModalVisible = false"
      width="420px"
    >
      <a-form layout="vertical" class="mt-4">
        <a-form-item label="Status">
          <a-select
            v-model:value="statusForm.status"
            :options="quickStatusOptions"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>
