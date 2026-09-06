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
import { useI18n } from '@/composables/useI18n'
import { agentMemberApi, type AgentMember } from '@/api/agent-member'
import { agentApi, type Agent } from '@/api/agent'

const { t } = useI18n()

const route = useRoute()

const projectId = computed(() => parseInt(route.params.projectId as string, 10) || parseInt(route.params.id as string, 10))

const members = ref<AgentMember[]>([])
const agents = ref<Agent[]>([])
const loading = ref(false)
const searchName = ref('')
const filterAgentType = ref<string | undefined>(undefined)
const filterStatus = ref<string | undefined>(undefined)

const agentTypeOptions = [
  { value: 'builtin', labelKey: 'agent.typeBuiltin' },
  { value: 'custom', labelKey: 'agent.typeCustom' }
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
  { value: 'admin', labelKey: 'agentMembers.roleAdmin' },
  { value: 'lead', labelKey: 'agentMembers.roleLead' },
  { value: 'developer', labelKey: 'agentMembers.roleDeveloper' },
  { value: 'reviewer', labelKey: 'agentMembers.roleReviewer' },
  { value: 'tester', labelKey: 'agentMembers.roleTester' },
  { value: 'contributor', labelKey: 'agentMembers.roleContributor' },
  { value: 'observer', labelKey: 'agentMembers.roleObserver' }
]

const availableSkillOptions = computed(() => [
  t('agentMembers.skillCodeGeneration'), t('agentMembers.skillCodeReview'),
  t('agentMembers.skillRequirementAnalysis'), t('agentMembers.skillTestWriting'),
  t('agentMembers.skillDocumentation'), t('agentMembers.skillBugFix'),
  t('agentMembers.skillRefactoring'), t('agentMembers.skillPerformanceAnalysis'),
  t('agentMembers.skillSecurityAudit'), t('agentMembers.skillDevOps'),
  t('agentMembers.skillDatabaseManagement'), t('agentMembers.skillApiDesign'),
  t('agentMembers.skillUiDesign'), t('agentMembers.skillProjectManagement'),
  t('agentMembers.skillDataAnalysis')
])

const columns = computed(() => [
  { title: t('agentMembers.columnName'), key: 'name', dataIndex: 'agent_name', width: 200 },
  { title: t('agentMembers.columnAgentType'), key: 'agentType', dataIndex: 'agent_type', width: 130 },
  { title: t('agentMembers.columnRole'), key: 'role', dataIndex: 'role', width: 120 },
  { title: t('agentMembers.columnStatus'), key: 'status', dataIndex: 'is_active', width: 100 },
  { title: t('agentMembers.columnSkills'), key: 'skills', width: 260 },
  { title: t('agentMembers.columnActions'), key: 'actions', width: 200 }
])

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
  return type === 'builtin' ? t('agent.typeBuiltin') : t('agent.typeCustom')
}

function getStatusLabel(isActive: boolean): string {
  return isActive ? t('agent.statusActive') : t('agent.statusInactive')
}

function getRoleLabel(role: string): string {
  const found = roleOptions.find(r => r.value === role)
  return found ? t(found.labelKey) : role
}

async function loadMembers() {
  loading.value = true
  try {
    const res = await agentMemberApi.list(projectId.value)
    const data = res.data?.data || res.data || []
    members.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    message.error(t('agentMembers.loadFailed') + (e.response?.data?.message || e.message || t('agentMembers.unknownError')))
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
    message.warning(t('agentMembers.selectRoleWarning'))
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
      message.success(t('agentMembers.updateSuccess'))
    } else {
      if (!formState.agent_id) {
        message.warning(t('agentMembers.selectAgentWarning'))
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
      message.success(t('agentMembers.addSuccess'))
    }
    modalVisible.value = false
    await loadMembers()
  } catch (e: any) {
    message.error(t('agentMembers.operationFailed') + (e.response?.data?.message || e.message || t('agentMembers.unknownError')))
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
    message.success(newStatus ? t('agentMembers.activated') : t('agentMembers.deactivated'))
    await loadMembers()
  } catch (e: any) {
    message.error(t('agentMembers.statusToggleFailed') + (e.response?.data?.message || e.message || t('agentMembers.unknownError')))
  }
}

async function handleDelete(member: AgentMember) {
  try {
    await agentMemberApi.remove(projectId.value, member.agent_id)
    message.success(t('agentMembers.deleteSuccess'))
    await loadMembers()
  } catch (e: any) {
    message.error(t('agentMembers.deleteFailed') + (e.response?.data?.message || e.message || t('agentMembers.unknownError')))
  }
}

onMounted(() => {
  loadMembers()
  loadAgents()
})
</script>

<template>
  <Card :title="t('agentMembers.title')" :bordered="false">
    <template #extra>
      <Space>
        <Input
          v-model:value="searchName"
          :placeholder="t('agentMembers.searchPlaceholder')"
          allowClear
          style="width: 200px;"
        />
        <Select
          v-model:value="filterAgentType"
          :placeholder="t('agentMembers.filterType')"
          allowClear
          style="width: 140px;"
        >
          <SelectOption
            v-for="opt in agentTypeOptions"
            :key="opt.value"
            :value="opt.value"
          >
            {{ t(opt.labelKey) }}
          </SelectOption>
        </Select>
        <Select
          v-model:value="filterStatus"
          :placeholder="t('agentMembers.filterStatus')"
          allowClear
          style="width: 120px;"
        >
          <SelectOption value="active">{{ t('agent.statusActive') }}</SelectOption>
          <SelectOption value="inactive">{{ t('agent.statusInactive') }}</SelectOption>
        </Select>
        <Button type="primary" @click="openCreateModal">
          {{ t('agentMembers.addButton') }}
        </Button>
      </Space>
    </template>

    <Table
      :dataSource="filteredMembers"
      :columns="columns"
      :loading="loading"
      rowKey="id"
      :pagination="{ pageSize: 10, showSizeChanger: true, showTotal: (total: number) => t('agentMembers.totalItems', { total }) }"
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
              {{ t('agentMembers.noSkills') }}
            </span>
          </Space>
        </template>

        <template v-if="column.key === 'actions'">
          <Space>
            <Button size="small" @click="openEditModal(record as AgentMember)">
              {{ t('common.edit') }}
            </Button>
            <Button
              size="small"
              :type="(record as AgentMember).is_active ? 'default' : 'primary'"
              ghost
              @click="handleToggleStatus(record as AgentMember)"
            >
              {{ (record as AgentMember).is_active ? t('agentMembers.deactivate') : t('agentMembers.activate') }}
            </Button>
            <Popconfirm
              :title="t('agentMembers.deleteConfirm')"
              :okText="t('common.confirm')"
              :cancelText="t('common.cancel')"
              @confirm="handleDelete(record as AgentMember)"
            >
              <Button size="small" danger>
                {{ t('common.delete') }}
              </Button>
            </Popconfirm>
          </Space>
        </template>
      </template>
    </Table>

    <!-- Create / Edit Modal -->
    <Modal
      v-model:open="modalVisible"
      :title="isEditing ? t('agentMembers.editTitle') : t('agentMembers.createTitle')"
      @ok="handleFormSubmit"
      :confirmLoading="formSubmitting"
      :okText="t('common.confirm')"
      :cancelText="t('common.cancel')"
      width="640px"
    >
      <Form
        :model="formState"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 16 }"
        style="margin-top: 16px;"
      >
        <FormItem v-if="!isEditing" :label="t('agentMembers.formSelectAgent')" required>
          <Select
            v-model:value="formState.agent_id"
            :placeholder="t('agentMembers.formSelectAgentPlaceholder')"
            @change="(val: any) => onAgentSelect(val)"
          >
            <SelectOption
              v-for="agent in availableAgents"
              :key="agent.id"
              :value="agent.id"
            >
              {{ agent.avatar }} {{ agent.name }}（{{ agent.agent_type === 'builtin' ? t('agent.typeBuiltin') : t('agent.typeCustom') }}）
            </SelectOption>
          </Select>
        </FormItem>

        <FormItem v-if="isEditing" :label="t('agentMembers.formName')">
          <Input :value="formState.name" disabled />
        </FormItem>

        <FormItem :label="t('agentMembers.formModel')">
          <Select
            v-model:value="formState.model_id"
            :placeholder="t('agentMembers.formModelPlaceholder')"
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

        <FormItem :label="t('agentMembers.formAgentType')">
          <Select
            v-model:value="formState.agent_type"
            :placeholder="t('agentMembers.formAgentTypePlaceholder')"
            disabled
          >
            <SelectOption
              v-for="opt in agentTypeOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ t(opt.labelKey) }}
            </SelectOption>
          </Select>
        </FormItem>

        <FormItem :label="t('agentMembers.formRole')" required>
          <Select
            v-model:value="formState.role"
            :placeholder="t('agentMembers.formRolePlaceholder')"
          >
            <SelectOption
              v-for="opt in roleOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ t(opt.labelKey) }}
            </SelectOption>
          </Select>
        </FormItem>

        <FormItem :label="t('agentMembers.formSkills')">
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
            :placeholder="t('agentMembers.formSkillsPlaceholder')"
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

        <FormItem :label="t('agentMembers.formSystemPrompt')">
          <Input.Textarea
            v-model:value="formState.system_prompt"
            :rows="3"
            :placeholder="t('agentMembers.formSystemPromptPlaceholder')"
          />
        </FormItem>

        <FormItem :label="t('agentMembers.formMaxTokens')">
          <Input
            v-model:value="formState.max_tokens"
            type="number"
            placeholder="4096"
          />
        </FormItem>

        <FormItem :label="t('agentMembers.formTemperature')">
          <Input
            v-model:value="formState.temperature"
            type="number"
            step="0.1"
            min="0"
            max="2"
            placeholder="0.7"
          />
        </FormItem>

        <FormItem :label="t('agentMembers.formCostPerTask')">
          <Input
            v-model:value="formState.cost_per_task"
            type="number"
            min="0"
            step="0.01"
            placeholder="0.00"
          />
        </FormItem>

        <FormItem :label="t('agentMembers.formSlaHours')">
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
