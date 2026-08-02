<template>
  <div class="workflows-page">
    <a-card>
      <template #title>
        <div style="display: flex; align-items: center; justify-content: space-between;">
          <span style="font-size: 20px; font-weight: 600;">工作流管理</span>
          <a-button type="primary" @click="openCreateModal">
            创建工作流
          </a-button>
        </div>
      </template>

      <a-tabs v-model:activeKey="activeTab" @change="onTabChange">
        <a-tab-pane key="workflows" tab="工作流列表">
          <div style="margin-bottom: 16px;">
            <a-space>
              <a-select
                v-model:value="statusFilter"
                placeholder="筛选状态"
                allowClear
                style="width: 160px;"
                @change="loadWorkflows"
              >
                <a-select-option value="draft">草稿</a-select-option>
                <a-select-option value="active">激活</a-select-option>
                <a-select-option value="archived">归档</a-select-option>
              </a-select>
              <a-button @click="loadWorkflows">刷新</a-button>
            </a-space>
          </div>

          <a-table
            :dataSource="filteredWorkflows"
            :columns="workflowColumns"
            :loading="loadingWorkflows"
            rowKey="id"
            :pagination="{ pageSize: 10, showSizeChanger: true, showTotal: (total: number) => `共 ${total} 条` }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <a-space>
                  <span style="font-weight: 500;">{{ record.name }}</span>
                </a-space>
              </template>
              <template v-if="column.key === 'status'">
                <a-tag :color="getStatusColor(record.is_active)">
                  {{ record.is_active ? '激活' : '草稿' }}
                </a-tag>
              </template>
              <template v-if="column.key === 'nodeCount'">
                <a-badge :count="record.node_count" :numberStyle="{ backgroundColor: '#1890ff' }" />
              </template>
              <template v-if="column.key === 'createdAt'">
                {{ formatDate(record.created_at) }}
              </template>
              <template v-if="column.key === 'actions'">
                <a-space>
                  <a-button size="small" @click="openEditModal(record)">编辑</a-button>
                  <a-button size="small" type="primary" ghost @click="goToDesign(record)">设计</a-button>
                  <a-button size="small" type="primary" @click="openRunModal(record)">运行</a-button>
                  <a-button
                    v-if="!record.is_active"
                    size="small"
                    type="success"
                    ghost
                    @click="handleActivate(record)"
                  >
                    激活
                  </a-button>
                  <a-button
                    v-if="record.is_active"
                    size="small"
                    warning
                    ghost
                    @click="handleArchive(record)"
                  >
                    归档
                  </a-button>
                  <a-popconfirm
                    title="确定要删除这个工作流吗？"
                    okText="确定"
                    cancelText="取消"
                    @confirm="handleDelete(record)"
                  >
                    <a-button size="small" danger>删除</a-button>
                  </a-popconfirm>
                </a-space>
              </template>
            </template>
          </a-table>
        </a-tab-pane>

        <a-tab-pane key="runs" tab="执行记录">
          <a-table
            :dataSource="allRuns"
            :columns="runColumns"
            :loading="loadingRuns"
            rowKey="id"
            :pagination="{ pageSize: 10, showSizeChanger: true, showTotal: (total: number) => `共 ${total} 条` }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'workflowName'">
                {{ getWorkflowName(record.workflow_id) }}
              </template>
              <template v-if="column.key === 'status'">
                <a-tag :color="getRunStatusColor(record.status)">
                  {{ getRunStatusLabel(record.status) }}
                </a-tag>
              </template>
              <template v-if="column.key === 'triggerType'">
                <a-tag>{{ getTriggerTypeLabel(record.trigger_type) }}</a-tag>
              </template>
              <template v-if="column.key === 'startedAt'">
                {{ record.started_at ? formatDate(record.started_at) : '-' }}
              </template>
              <template v-if="column.key === 'completedAt'">
                {{ record.completed_at ? formatDate(record.completed_at) : '-' }}
              </template>
              <template v-if="column.key === 'errorMessage'">
                <a-tooltip v-if="record.error_info" :title="record.error_info">
                  <span style="color: #ff4d4f; cursor: pointer;">查看错误</span>
                </a-tooltip>
                <span v-else>-</span>
              </template>
            </template>
          </a-table>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <!-- Create / Edit Workflow Modal -->
    <a-modal
      v-model:open="formModalVisible"
      :title="isEditing ? '编辑工作流' : '创建工作流'"
      @ok="handleFormSubmit"
      :confirmLoading="formSubmitting"
      okText="确定"
      cancelText="取消"
    >
      <a-form
        :model="formState"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 16 }"
      >
        <a-form-item label="名称" required>
          <a-input
            v-model:value="formState.name"
            placeholder="请输入工作流名称"
          />
        </a-form-item>
        <a-form-item label="描述">
          <a-input
            v-model:value="formState.description"
            type="textarea"
            :rows="3"
            placeholder="请输入工作流描述"
          />
        </a-form-item>
        <a-form-item label="触发方式">
          <a-select v-model:value="formState.trigger_type" placeholder="选择触发方式">
            <a-select-option value="manual">手动</a-select-option>
            <a-select-option value="scheduled">定时</a-select-option>
            <a-select-option value="webhook">Webhook</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Run Workflow Modal -->
    <a-modal
      v-model:open="runModalVisible"
      title="运行工作流"
      @ok="handleRunSubmit"
      :confirmLoading="runSubmitting"
      okText="运行"
      cancelText="取消"
    >
      <a-form
        :model="runFormState"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 16 }"
      >
        <a-form-item label="工作流">
          <a-input :value="runningWorkflow?.name" disabled />
        </a-form-item>
        <a-form-item label="触发类型">
          <a-select v-model:value="runFormState.trigger_type" placeholder="选择触发类型">
            <a-select-option value="manual">手动</a-select-option>
            <a-select-option value="scheduled">定时</a-select-option>
            <a-select-option value="webhook">Webhook</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Issue ID">
          <a-input-number
            v-model:value="runFormState.issue_id"
            :min="1"
            placeholder="可选，关联的 Issue ID"
            style="width: 100%;"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { workflowApi } from '@/api/workflow'
import type { Workflow, WorkflowRun } from '@/api/workflow'

const route = useRoute()
const router = useRouter()

const projectId = computed(() => {
  return Number(route.params.projectId || route.params.id)
})

const activeTab = ref<string>('workflows')
const loadingWorkflows = ref(false)
const loadingRuns = ref(false)
const workflows = ref<Workflow[]>([])
const allRuns = ref<(WorkflowRun & { trigger_type?: string })[]>([])
const statusFilter = ref<string | undefined>(undefined)

const workflowColumns = [
  { title: '名称', key: 'name', dataIndex: 'name' },
  { title: '描述', key: 'description', dataIndex: 'description', ellipsis: true },
  { title: '状态', key: 'status', dataIndex: 'is_active', width: 100 },
  { title: '节点数', key: 'nodeCount', dataIndex: 'node_count', width: 80 },
  { title: '创建时间', key: 'createdAt', dataIndex: 'created_at', width: 160 },
  { title: '操作', key: 'actions', width: 360 },
]

const runColumns = [
  { title: '工作流', key: 'workflowName', dataIndex: 'workflow_id', width: 160 },
  { title: '状态', key: 'status', dataIndex: 'status', width: 100 },
  { title: '触发类型', key: 'triggerType', dataIndex: 'trigger_type', width: 120 },
  { title: '开始时间', key: 'startedAt', dataIndex: 'started_at', width: 160 },
  { title: '完成时间', key: 'completedAt', dataIndex: 'completed_at', width: 160 },
  { title: '错误信息', key: 'errorMessage', dataIndex: 'error_info', ellipsis: true },
]

const filteredWorkflows = computed(() => {
  if (!statusFilter.value) return workflows.value
  return workflows.value.filter(w => {
    if (statusFilter.value === 'active') return w.is_active
    if (statusFilter.value === 'archived') return !w.is_active
    if (statusFilter.value === 'draft') return !w.is_active
    return true
  })
})

// Form state
const formModalVisible = ref(false)
const isEditing = ref(false)
const editingWorkflowId = ref<number | null>(null)
const formSubmitting = ref(false)
const formState = reactive({
  name: '',
  description: '',
  trigger_type: 'manual' as string,
})

// Run form state
const runModalVisible = ref(false)
const runSubmitting = ref(false)
const runningWorkflow = ref<Workflow | null>(null)
const runFormState = reactive({
  trigger_type: 'manual' as string,
  issue_id: undefined as number | undefined,
})

function getStatusColor(isActive: boolean): string {
  return isActive ? 'green' : 'default'
}

function getRunStatusColor(status: string): string {
  const map: Record<string, string> = {
    running: 'processing',
    completed: 'success',
    failed: 'error',
    cancelled: 'warning',
    pending: 'default',
  }
  return map[status] || 'default'
}

function getRunStatusLabel(status: string): string {
  const map: Record<string, string> = {
    running: '运行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
    pending: '等待中',
  }
  return map[status] || status
}

function getTriggerTypeLabel(type?: string): string {
  const map: Record<string, string> = {
    manual: '手动',
    scheduled: '定时',
    webhook: 'Webhook',
  }
  return map[type || ''] || type || '-'
}

function getWorkflowName(workflowId: number): string {
  const wf = workflows.value.find(w => w.id === workflowId)
  return wf?.name || `#${workflowId}`
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function loadWorkflows() {
  loadingWorkflows.value = true
  try {
    const res = await workflowApi.list(projectId.value)
    workflows.value = res.data.data || []
  } catch (e: any) {
    message.error('加载工作流列表失败：' + (e.message || '未知错误'))
  } finally {
    loadingWorkflows.value = false
  }
}

async function loadAllRuns() {
  loadingRuns.value = true
  try {
    const runs: (WorkflowRun & { trigger_type?: string })[] = []
    for (const wf of workflows.value) {
      try {
        const res = await workflowApi.listRuns(projectId.value, wf.id)
        const wfRuns = res.data.data || []
        for (const r of wfRuns) {
          runs.push({ ...r, trigger_type: wf.trigger_type })
        }
      } catch {
        // skip failed workflow runs
      }
    }
    allRuns.value = runs.sort((a, b) => {
      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    })
  } catch (e: any) {
    message.error('加载执行记录失败：' + (e.message || '未知错误'))
  } finally {
    loadingRuns.value = false
  }
}

function onTabChange(key: string) {
  if (key === 'runs') {
    loadAllRuns()
  }
}

function openCreateModal() {
  isEditing.value = false
  editingWorkflowId.value = null
  formState.name = ''
  formState.description = ''
  formState.trigger_type = 'manual'
  formModalVisible.value = true
}

function openEditModal(workflow: Workflow) {
  isEditing.value = true
  editingWorkflowId.value = workflow.id
  formState.name = workflow.name
  formState.description = workflow.description || ''
  formState.trigger_type = workflow.trigger_type || 'manual'
  formModalVisible.value = true
}

async function handleFormSubmit() {
  if (!formState.name.trim()) {
    message.warning('请输入工作流名称')
    return
  }
  formSubmitting.value = true
  try {
    if (isEditing.value && editingWorkflowId.value) {
      await workflowApi.update(projectId.value, editingWorkflowId.value, {
        name: formState.name,
        description: formState.description,
        trigger_type: formState.trigger_type,
      })
      message.success('工作流更新成功')
    } else {
      await workflowApi.create(projectId.value, {
        name: formState.name,
        description: formState.description,
        trigger_type: formState.trigger_type,
      })
      message.success('工作流创建成功')
    }
    formModalVisible.value = false
    await loadWorkflows()
  } catch (e: any) {
    message.error('操作失败：' + (e.message || '未知错误'))
  } finally {
    formSubmitting.value = false
  }
}

async function handleDelete(workflow: Workflow) {
  try {
    await workflowApi.delete(projectId.value, workflow.id)
    message.success('工作流删除成功')
    await loadWorkflows()
  } catch (e: any) {
    message.error('删除失败：' + (e.message || '未知错误'))
  }
}

async function handleActivate(workflow: Workflow) {
  try {
    await workflowApi.update(projectId.value, workflow.id, { is_active: true })
    message.success('工作流已激活')
    await loadWorkflows()
  } catch (e: any) {
    message.error('激活失败：' + (e.message || '未知错误'))
  }
}

async function handleArchive(workflow: Workflow) {
  try {
    await workflowApi.update(projectId.value, workflow.id, { is_active: false })
    message.success('工作流已归档')
    await loadWorkflows()
  } catch (e: any) {
    message.error('归档失败：' + (e.message || '未知错误'))
  }
}

function goToDesign(workflow: Workflow) {
  // The route is /workspace/:slug/... — use the slug from the current route
  // rather than the numeric workspace id (which is NaN for slug-based routes).
  const slug = route.params.slug
  router.push(
    `/workspace/${slug}/project/${projectId.value}/workflow/${workflow.id}/design`
  )
}

function openRunModal(workflow: Workflow) {
  runningWorkflow.value = workflow
  runFormState.trigger_type = 'manual'
  runFormState.issue_id = undefined
  runModalVisible.value = true
}

async function handleRunSubmit() {
  if (!runningWorkflow.value) return
  runSubmitting.value = true
  try {
    await workflowApi.execute(projectId.value, runningWorkflow.value.id, runFormState.issue_id)
    message.success('工作流已启动')
    runModalVisible.value = false
  } catch (e: any) {
    message.error('运行失败：' + (e.message || '未知错误'))
  } finally {
    runSubmitting.value = false
  }
}

onMounted(() => {
  loadWorkflows()
})
</script>

<style scoped>
.workflows-page {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}
</style>
