<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useRoute } from 'vue-router'
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  FormItem,
  Input,
  Select,
  SelectOption,
  Tag,
  Badge,
  Space,
  Popconfirm,
  message
} from 'ant-design-vue'
import { agentMemberApi, type AgentMember } from '@/api/agent-member'
import { agentApi, type Agent } from '@/api/agent'

const route = useRoute()

const projectId = computed(() => parseInt(route.params.projectId as string, 10) || parseInt(route.params.id as string, 10))

const members = ref<AgentMember[]>([])
const agents = ref<Agent[]>([])
const loading = ref(false)
const searchName = ref('')
const filterAgentType = ref<string | undefined>(undefined)
const filterStatus = ref<string | undefined>(undefined)

const agentTypeOptions = [
  { value: 'builtin', label: '内置 Agent' },
  { value: 'custom', label: '自定义 Agent' }
]

const modelOptions = [
  { value: 'gpt-4o', label: 'GPT-4o' },
  { value: 'gpt-4o-mini', label: 'GPT-4o Mini' },
  { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
  { value: 'claude-3-opus', label: 'Claude 3 Opus' },
  { value: 'claude-3-sonnet', label: 'Claude 3 Sonnet' },
  { value: 'claude-3-haiku', label: 'Claude 3 Haiku' },
  { value: 'deepseek-chat', label: 'DeepSeek Chat' },
  { value: 'deepseek-coder', label: 'DeepSeek Coder' },
  { value: 'gemini-pro', label: 'Gemini Pro' }
]

const roleOptions = [
  { value: 'admin', label: '管理员' },
  { value: 'lead', label: '负责人' },
  { value: 'developer', label: '开发者' },
  { value: 'reviewer', label: '审查者' },
  { value: 'tester', label: '测试者' },
  { value: 'contributor', label: '贡献者' },
  { value: 'observer', label: '观察者' }
]

const availableSkillOptions = [
  '代码生成', '代码审查', '需求分析', '测试编写', '文档撰写',
  'Bug修复', '重构优化', '性能分析', '安全审计', '部署运维',
  '数据库管理', 'API设计', 'UI/UX设计', '项目管理', '数据分析'
]

const columns = [
  { title: '名称', key: 'name', dataIndex: 'agent_name', width: 200 },
  { title: 'Agent 类型', key: 'agentType', dataIndex: 'agent_type', width: 130 },
  { title: '角色', key: 'role', dataIndex: 'role', width: 120 },
  { title: '状态', key: 'status', dataIndex: 'is_active', width: 100 },
  { title: '技能', key: 'skills', width: 260 },
  { title: '操作', key: 'actions', width: 200 }
]

// Form state
const modalVisible = ref(false)
const isEditing = ref(false)
const editingMember = ref<AgentMember | null>(null)
const formSubmitting = ref(false)

interface MemberFormState {
  agent_id: number | undefined
  name: string
  model_id: string | undefined
  agent_type: string | undefined
  role: string | undefined
  skills: string[]
  system_prompt: string
  max_tokens: number
  temperature: number
  cost_per_task: number
  sla_hours: number
}

const formState = reactive<MemberFormState>({
  agent_id: undefined,
  name: '',
  model_id: undefined,
  agent_type: undefined,
  role: undefined,
  skills: [],
  system_prompt: '',
  max_tokens: 4096,
  temperature: 0.7,
  cost_per_task: 0,
  sla_hours: 24
})

const selectedSkill = ref<string | undefined>(undefined)

// Local storage for extended agent config (skills, system_prompt, etc.)
// since the API only stores agent_id and role
const agentConfigMap = ref<Record<number, Partial<MemberFormState>>>({})

function getAgentConfig(agentId: number): Partial<MemberFormState> {
  return agentConfigMap.value[agentId] || {}
}

function getAgentSkills(agentId: number): string[] {
  return agentConfigMap.value[agentId]?.skills || []
}

const filteredMembers = computed(() => {
  let result = members.value
  if (searchName.value) {
    const keyword = searchName.value.toLowerCase()
    result = result.filter(m => m.agent_name.toLowerCase().includes(keyword))
  }
  if (filterAgentType.value) {
    result = result.filter(m => m.agent_type === filterAgentType.value)
  }
  if (filterStatus.value !== undefined) {
    const isActive = filterStatus.value === 'active'
    result = result.filter(m => m.is_active === isActive)
  }
  return result
})

const availableAgents = computed(() => {
  const memberAgentIds = new Set(members.value.map(m => m.agent_id))
  return agents.value.filter(a => !memberAgentIds.has(a.id))
})

function resetForm() {
  formState.agent_id = undefined
  formState.name = ''
  formState.model_id = undefined
  formState.agent_type = undefined
  formState.role = undefined
  formState.skills = []
  selectedSkill.value = undefined
  formState.system_prompt = ''
  formState.max_tokens = 4096
  formState.temperature = 0.7
  formState.cost_per_task = 0
  formState.sla_hours = 24
}

function openCreateModal() {
  isEditing.value = false
  editingMember.value = null
  resetForm()
  modalVisible.value = true
}

function openEditModal(member: AgentMember) {
  isEditing.value = true
  editingMember.value = member
  const config = getAgentConfig(member.agent_id)
  formState.agent_id = member.agent_id
  formState.name = member.agent_name
  formState.model_id = config.model_id
  formState.agent_type = member.agent_type
  formState.role = member.role
  formState.skills = [...(config.skills || [])]
  formState.system_prompt = config.system_prompt || ''
  formState.max_tokens = config.max_tokens ?? 4096
  formState.temperature = config.temperature ?? 0.7
  formState.cost_per_task = config.cost_per_task ?? 0
  formState.sla_hours = config.sla_hours ?? 24
  modalVisible.value = true
}

function onAgentSelect(agentId: number) {
  const agent = agents.value.find(a => a.id === agentId)
  if (agent) {
    formState.name = agent.name
    formState.agent_type = agent.agent_type
  }
}

function handleSkillAdd(skill: string) {
  if (skill && !formState.skills.includes(skill)) {
    formState.skills.push(skill)
  }
  selectedSkill.value = undefined
}

function handleSkillRemove(skill: string) {
  formState.skills = formState.skills.filter(s => s !== skill)
}

function getAgentTypeColor(type: string): string {
  return type === 'builtin' ? 'blue' : 'green'
}

function getAgentTypeLabel(type: string): string {
  return type === 'builtin' ? '内置' : '自定义'
}

function getStatusLabel(isActive: boolean): string {
  return isActive ? '活跃' : '停用'
}

function getRoleLabel(role: string): string {
  const found = roleOptions.find(r => r.value === role)
  return found?.label || role
}

async function loadMembers() {
  loading.value = true
  try {
    const res = await agentMemberApi.list(projectId.value)
    const data = res.data?.data || res.data || []
    members.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    message.error('加载 Agent 成员失败：' + (e.response?.data?.message || e.message || '未知错误'))
    members.value = []
  } finally {
    loading.value = false
  }
}

async function loadAgents() {
  try {
    const data = await agentApi.list(1)
    agents.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    agents.value = []
  }
}

async function handleFormSubmit() {
  if (!formState.role) {
    message.warning('请选择角色')
    return
  }

  formSubmitting.value = true
  try {
    if (isEditing.value && editingMember.value) {
      await agentMemberApi.updateRole(projectId.value, editingMember.value.agent_id, {
        role: formState.role
      })
      agentConfigMap.value[editingMember.value.agent_id] = {
        model_id: formState.model_id,
        skills: [...formState.skills],
        system_prompt: formState.system_prompt,
        max_tokens: formState.max_tokens,
        temperature: formState.temperature,
        cost_per_task: formState.cost_per_task,
        sla_hours: formState.sla_hours
      }
      message.success('成员更新成功')
    } else {
      if (!formState.agent_id) {
        message.warning('请选择要添加的 Agent')
        return
      }
      await agentMemberApi.add(projectId.value, {
        agent_id: formState.agent_id,
        role: formState.role
      })
      agentConfigMap.value[formState.agent_id] = {
        model_id: formState.model_id,
        skills: [...formState.skills],
        system_prompt: formState.system_prompt,
        max_tokens: formState.max_tokens,
        temperature: formState.temperature,
        cost_per_task: formState.cost_per_task,
        sla_hours: formState.sla_hours
      }
      message.success('成员添加成功')
    }
    modalVisible.value = false
    await loadMembers()
  } catch (e: any) {
    message.error('操作失败：' + (e.response?.data?.message || e.message || '未知错误'))
  } finally {
    formSubmitting.value = false
  }
}

async function handleToggleStatus(member: AgentMember) {
  try {
    const newStatus = !member.is_active
    await agentMemberApi.updateRole(projectId.value, member.agent_id, {
      role: member.role
    })
    member.is_active = newStatus
    message.success(`成员已${newStatus ? '激活' : '停用'}`)
    await loadMembers()
  } catch (e: any) {
    message.error('状态切换失败：' + (e.response?.data?.message || e.message || '未知错误'))
  }
}

async function handleDelete(member: AgentMember) {
  try {
    await agentMemberApi.remove(projectId.value, member.agent_id)
    message.success('成员已删除')
    await loadMembers()
  } catch (e: any) {
    message.error('删除失败：' + (e.response?.data?.message || e.message || '未知错误'))
  }
}

onMounted(() => {
  loadMembers()
  loadAgents()
})
</script>

<template>
  <Card title="Agent 团队成员" :bordered="false">
    <template #extra>
      <Space>
        <Input
          v-model:value="searchName"
          placeholder="搜索成员名称"
          allowClear
          style="width: 200px;"
        />
        <Select
          v-model:value="filterAgentType"
          placeholder="筛选类型"
          allowClear
          style="width: 140px;"
        >
          <SelectOption
            v-for="opt in agentTypeOptions"
            :key="opt.value"
            :value="opt.value"
          >
            {{ opt.label }}
          </SelectOption>
        </Select>
        <Select
          v-model:value="filterStatus"
          placeholder="筛选状态"
          allowClear
          style="width: 120px;"
        >
          <SelectOption value="active">活跃</SelectOption>
          <SelectOption value="inactive">停用</SelectOption>
        </Select>
        <Button type="primary" @click="openCreateModal">
          添加 Agent 成员
        </Button>
      </Space>
    </template>

    <Table
      :dataSource="filteredMembers"
      :columns="columns"
      :loading="loading"
      rowKey="id"
      :pagination="{ pageSize: 10, showSizeChanger: true, showTotal: (total: number) => `共 ${total} 条` }"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <Space>
            <span style="font-size: 18px;">{{ record.avatar }}</span>
            <span style="font-weight: 500;">{{ record.agent_name }}</span>
          </Space>
        </template>

        <template v-if="column.key === 'agentType'">
          <Badge :color="record.agent_type === 'builtin' ? '#1890ff' : '#52c41a'" />
          <Tag :color="getAgentTypeColor(record.agent_type)" style="margin-left: 6px;">
            {{ getAgentTypeLabel(record.agent_type) }}
          </Tag>
        </template>

        <template v-if="column.key === 'role'">
          <Tag color="geekblue">
            {{ getRoleLabel(record.role) }}
          </Tag>
        </template>

        <template v-if="column.key === 'status'">
          <Badge :status="record.is_active ? 'success' : 'default'" />
          <span style="margin-left: 6px;">{{ getStatusLabel(record.is_active) }}</span>
        </template>

        <template v-if="column.key === 'skills'">
          <Space wrap :size="[4, 4]">
            <Tag
              v-for="skill in getAgentSkills(record.agent_id)"
              :key="skill"
              color="processing"
              style="border-radius: 12px;"
            >
              {{ skill }}
            </Tag>
            <span v-if="getAgentSkills(record.agent_id).length === 0" style="color: #bfbfbf;">
              暂无技能
            </span>
          </Space>
        </template>

        <template v-if="column.key === 'actions'">
          <Space>
            <Button size="small" @click="openEditModal(record as AgentMember)">
              编辑
            </Button>
            <Button
              size="small"
              :type="(record as AgentMember).is_active ? 'default' : 'primary'"
              ghost
              @click="handleToggleStatus(record as AgentMember)"
            >
              {{ (record as AgentMember).is_active ? '停用' : '激活' }}
            </Button>
            <Popconfirm
              title="确定要删除该 Agent 成员吗？"
              okText="确定"
              cancelText="取消"
              @confirm="handleDelete(record as AgentMember)"
            >
              <Button size="small" danger>
                删除
              </Button>
            </Popconfirm>
          </Space>
        </template>
      </template>
    </Table>

    <!-- Create / Edit Modal -->
    <Modal
      v-model:open="modalVisible"
      :title="isEditing ? '编辑 Agent 成员' : '添加 Agent 成员'"
      @ok="handleFormSubmit"
      :confirmLoading="formSubmitting"
      okText="确定"
      cancelText="取消"
      width="640px"
    >
      <Form
        :model="formState"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 16 }"
        style="margin-top: 16px;"
      >
        <FormItem v-if="!isEditing" label="选择 Agent" required>
          <Select
            v-model:value="formState.agent_id"
            placeholder="请选择要添加的 Agent"
            @change="(val: any) => onAgentSelect(val)"
          >
            <SelectOption
              v-for="agent in availableAgents"
              :key="agent.id"
              :value="agent.id"
            >
              {{ agent.avatar }} {{ agent.name }}（{{ agent.agent_type === 'builtin' ? '内置' : '自定义' }}）
            </SelectOption>
          </Select>
        </FormItem>

        <FormItem v-if="isEditing" label="名称">
          <Input :value="formState.name" disabled />
        </FormItem>

        <FormItem label="模型">
          <Select
            v-model:value="formState.model_id"
            placeholder="选择模型（可选）"
            allowClear
          >
            <SelectOption
              v-for="model in modelOptions"
              :key="model.value"
              :value="model.value"
            >
              {{ model.label }}
            </SelectOption>
          </Select>
        </FormItem>

        <FormItem label="Agent 类型">
          <Select
            v-model:value="formState.agent_type"
            placeholder="选择 Agent 类型"
            disabled
          >
            <SelectOption
              v-for="opt in agentTypeOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>

        <FormItem label="角色" required>
          <Select
            v-model:value="formState.role"
            placeholder="选择角色"
          >
            <SelectOption
              v-for="opt in roleOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>

        <FormItem label="技能">
          <Space wrap :size="[4, 4]" style="margin-bottom: 8px;">
            <Tag
              v-for="skill in formState.skills"
              :key="skill"
              closable
              color="processing"
              @close="handleSkillRemove(skill)"
              style="border-radius: 12px;"
            >
              {{ skill }}
            </Tag>
          </Space>
          <Select
            v-model:value="selectedSkill"
            placeholder="选择技能"
            showSearch
            :filter-option="(input: string, option: any) => option.value.toLowerCase().includes(input.toLowerCase())"
            style="width: 200px;"
            @change="(val: any) => val && handleSkillAdd(val)"
            allowClear
          >
            <SelectOption
              v-for="skill in availableSkillOptions.filter(s => !formState.skills.includes(s))"
              :key="skill"
              :value="skill"
            >
              {{ skill }}
            </SelectOption>
          </Select>
        </FormItem>

        <FormItem label="系统提示词">
          <Input.Textarea
            v-model:value="formState.system_prompt"
            :rows="3"
            placeholder="设置 Agent 的系统提示词（可选）"
          />
        </FormItem>

        <FormItem label="最大 Token">
          <Input
            v-model:value="formState.max_tokens"
            type="number"
            placeholder="4096"
          />
        </FormItem>

        <FormItem label="温度">
          <Input
            v-model:value="formState.temperature"
            type="number"
            step="0.1"
            min="0"
            max="2"
            placeholder="0.7"
          />
        </FormItem>

        <FormItem label="单次任务成本">
          <Input
            v-model:value="formState.cost_per_task"
            type="number"
            min="0"
            step="0.01"
            placeholder="0.00"
          />
        </FormItem>

        <FormItem label="SLA (小时)">
          <Input
            v-model:value="formState.sla_hours"
            type="number"
            min="1"
            placeholder="24"
          />
        </FormItem>
      </Form>
    </Modal>
  </Card>
</template>

<style scoped>
:deep(.ant-card) {
  border-radius: 8px;
}

:deep(.ant-card-head) {
  border-bottom: 1px solid #f0f0f0;
}

:deep(.ant-table) {
  border-radius: 0 0 8px 8px;
}

:deep(.ant-table-thead > tr > th) {
  background: #fafafa;
  font-weight: 600;
}
</style>
