<template>
  <div class="issue-create-page">
    <!-- 头部 -->
    <div class="page-header">
      <button @click="goBack" class="back-btn">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <h1 class="page-title">{{ t('issue.createTitle') }}</h1>
    </div>

    <div class="page-content">
      <!-- 左侧表单 -->
      <div class="form-section">
        <!-- 类型选择 -->
        <div class="form-group">
          <label class="form-label">{{ t('issue.type') }}</label>
          <div class="type-selector">
            <button
              v-for="type in issueTypes"
              :key="type.id"
              class="type-option"
              :class="{ active: selectedTypeId === type.id }"
              @click="selectedTypeId = type.id"
            >
              <div class="type-icon" :style="{ backgroundColor: type.color }">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <span>{{ type.name }}</span>
            </button>
          </div>
        </div>

        <!-- 模板选择 -->
        <div v-if="availableTemplates.length > 0" class="form-group">
          <label class="form-label">{{ t('issue.template') }}</label>
          <select
            v-model.number="selectedTemplateId"
            @change="handleTemplateChange(selectedTemplateId)"
            class="form-input"
          >
            <option :value="null">{{ t('issue.templatePlaceholder') }}</option>
            <option v-for="tmpl in availableTemplates" :key="tmpl.id" :value="tmpl.id">
              {{ tmpl.name }}{{ tmpl.is_default ? t('issue.templateDefault') : '' }}
            </option>
          </select>
        </div>

        <!-- 标题 -->
        <div class="form-group">
          <label class="form-label">{{ t('issue.titleRequired') }}</label>
          <input
            v-model="formData.name"
            type="text"
            class="form-input title-input"
            :placeholder="t('issue.titlePlaceholder')"
            ref="titleInput"
          />
        </div>

        <!-- 描述 -->
        <div class="form-group">
          <label class="form-label">{{ t('issue.description') }}</label>
          <RichTextEditor
            v-model="formData.description"
            :placeholder="t('issue.descriptionPlaceholder')"
          />
        </div>

        <!-- 自定义字段 -->
        <div v-if="linkedFields.length > 0" class="form-group">
          <label class="form-label">{{ t('issue.customFields') }}</label>
          <div class="custom-fields">
            <div v-for="field in linkedFields" :key="field.field_id" class="field-item">
              <label class="field-label">
                {{ field.name }}
                <span v-if="isFieldRequired(field)" class="required">*</span>
              </label>
              <CustomFieldValueInput
                :field="field as any"
                :value="customFieldValues[field.field_id]"
                :members="projectMembers"
                @update:value="(v: any) => updateCustomFieldValue(field.field_id, v)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧属性 -->
      <div class="properties-section">
        <div class="property-group">
          <label class="property-label">{{ t('issue.state') }}</label>
          <select v-model="formData.state_id" class="property-select">
            <option value="">{{ t('issue.statePlaceholder') }}</option>
            <option v-for="state in states" :key="state.id" :value="state.id">
              {{ state.name }}
            </option>
          </select>
        </div>

        <div class="property-group">
          <label class="property-label">{{ t('issue.priority') }}</label>
          <select v-model="formData.priority" class="property-select">
            <option value="">{{ t('issue.priorityPlaceholder') }}</option>
            <option value="urgent">{{ t('issue.priorityUrgent') }}</option>
            <option value="high">{{ t('issue.priorityHigh') }}</option>
            <option value="medium">{{ t('issue.priorityMedium') }}</option>
            <option value="low">{{ t('issue.priorityLow') }}</option>
            <option value="none">{{ t('issue.priorityNone') }}</option>
          </select>
        </div>

        <div class="property-group">
          <label class="property-label">{{ t('issue.assignee') }}</label>
          <UserSelect
            v-model="formData.assignee_id"
            :users="projectMembers"
            :placeholder="t('issue.assigneePlaceholder')"
            :clearable="true"
          />
        </div>

        <div class="property-group">
          <label class="property-label">{{ t('issue.cycle') }}</label>
          <select v-model="formData.cycle_id" class="property-select">
            <option value="">{{ t('issue.cyclePlaceholder') }}</option>
            <option v-for="cycle in cycles" :key="cycle.id" :value="cycle.id">
              {{ cycle.name }}
            </option>
          </select>
        </div>

        <div class="property-group">
          <label class="property-label">{{ t('issue.module') }}</label>
          <select v-model="formData.module_id" class="property-select">
            <option value="">{{ t('issue.modulePlaceholder') }}</option>
            <option v-for="module in modules" :key="module.id" :value="module.id">
              {{ module.name }}
            </option>
          </select>
        </div>

        <div class="property-group">
          <label class="property-label">{{ t('issue.startDate') }}</label>
          <input v-model="formData.start_date" type="date" class="property-input" />
        </div>

        <div class="property-group">
          <label class="property-label">{{ t('issue.targetDate') }}</label>
          <input v-model="formData.target_date" type="date" class="property-input" />
        </div>

        <div class="property-group">
          <label class="property-label">{{ t('issue.parentIssue') }}</label>
          <div class="relative">
            <div v-if="selectedParent" class="flex items-center justify-between p-2 bg-gray-50 rounded-md mb-1">
              <span class="text-sm text-gray-700">#{{ selectedParent.sequence_id }} {{ selectedParent.name }}</span>
              <button @click="selectedParent = null; formData.parent_id = ''" class="text-gray-400 hover:text-red-500">&times;</button>
            </div>
            <input
              v-if="!selectedParent"
              v-model="parentSearch"
              @input="searchParents"
              type="text"
              :placeholder="t('issue.parentSearchPlaceholder')"
              class="w-full px-3 py-2 border border-gray-200 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-400 focus:border-blue-400 text-sm"
            />
            <div v-if="parentResults.length > 0 && !selectedParent" class="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-md shadow-lg max-h-36 overflow-y-auto">
              <div
                v-for="p in parentResults"
                :key="p.id"
                @click="selectParent(p)"
                class="px-3 py-2 hover:bg-gray-50 cursor-pointer text-sm"
              >
                <span class="font-medium">#{{ p.sequence_id }}</span> {{ p.name }}
              </div>
            </div>
          </div>
        </div>

        <div class="property-group">
          <label class="property-label">{{ t('issue.labels') }}</label>
          <LabelSelector
            :project-id="projectId"
            v-model="selectedLabelIds"
          />
        </div>
      </div>
    </div>

    <!-- 底部操作 -->
    <div class="page-footer">
      <button @click="goBack" class="btn btn-secondary">{{ t('issue.cancel') }}</button>
      <button @click="submitForm" class="btn btn-primary" :disabled="!canSubmit || saving">
        <svg v-if="saving" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
        </svg>
        {{ saving ? t('issue.creating') : t('issue.create') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { useRoute, useRouter } from 'vue-router'
import CustomFieldValueInput from '@/components/CustomFieldValueInput.vue'
import RichTextEditor from '@/components/RichTextEditor.vue'
import UserSelect from '@/components/UserSelect.vue'
import LabelSelector from '@/components/LabelSelector.vue'
import type { IssueType } from '@/types/issue-type'
import type { IssueCustomFieldValueUpdate } from '@/types/custom-field'
import type { State } from '@/types/project-settings'
import type { CycleResponse } from '@/types/cycle'
import type { ModuleResponse } from '@/types/module'
import type { User } from '@/types'
import * as issueTypeApi from '@/api/issue-type'
import * as issueApi from '@/api/issue'
import * as stateApi from '@/api/project-settings'
import * as cycleApi from '@/api/cycle'
import * as moduleApi from '@/api/module'
import projectApi from '@/api/project'
import { workspaceApi } from '@/api/workspace'

import * as workItemTemplateApi from '@/api/work-item-template'
import type { WorkItemTemplate } from '@/types/work-item-template'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const toast = useToast()

const projectId = computed(() => {
  const id = parseInt(route.params.id as string, 10) || parseInt(route.params.projectId as string, 10)
  return id || 0
})
const workspaceId = ref(0)
const slug = computed(() => route.params.slug as string || '')
const returnView = computed(() => route.query.view as string || 'list')

// 定义事件
const emit = defineEmits<{
  (e: 'created', issue: any): void
}>()

// 状态
const issueTypes = ref<IssueType[]>([])
const linkedFields = ref<any[]>([])
const states = ref<State[]>([])
const cycles = ref<CycleResponse[]>([])
const modules = ref<ModuleResponse[]>([])
const projectMembers = ref<User[]>([])
const selectedTypeId = ref<number | null>(null)
const customFieldValues = ref<Record<number, IssueCustomFieldValueUpdate>>({})
const saving = ref(false)

const templates = ref<WorkItemTemplate[]>([])
const selectedTemplateId = ref<number | null>(null)

// 表单数据
const formData = ref({
  name: '',
  description: '',
  state_id: '' as number | string,
  priority: '',
  assignee_id: '' as number | string,
  cycle_id: '' as number | string,
  module_id: '' as number | string,
  start_date: '',
  target_date: '',
  parent_id: '' as number | string
})
const selectedLabelIds = ref<number[]>([])
const parentSearch = ref('')
const parentResults = ref<any[]>([])
const selectedParent = ref<any>(null)
let parentSearchTimer: any = null

// 计算属性
const canSubmit = computed(() => {
  // 标题必填
  if (!formData.value.name.trim()) return false

  // 检查必填的自定义字段
  for (const field of linkedFields.value) {
    if (isFieldRequired(field) && !hasFieldValue(field.field_id)) {
      return false
    }
  }

  return true
})

const availableTemplates = computed(() => {
  if (!selectedTypeId.value) return templates.value
  return templates.value.filter(t => !t.issue_type_id || t.issue_type_id === selectedTypeId.value)
})

// 判断字段是否必填
function isFieldRequired(field: any): boolean {
  return field.is_required === true
}

// 判断字段是否有值
function hasFieldValue(fieldId: number): boolean {
  const value = customFieldValues.value[fieldId]
  if (!value) return false
  return !!(value.text_value || value.number_value !== undefined ||
    value.boolean_value !== undefined || value.date_value || value.url_value ||
    (value.json_value && value.json_value.length > 0))
}

async function loadTemplates() {
  try {
    templates.value = await workItemTemplateApi.listWorkItemTemplates(projectId.value)
  } catch (e) {
    console.error('Failed to load templates:', e)
  }
}

function applyTemplate(template: WorkItemTemplate | null) {
  if (!template) return

  const defaults = template.defaults

  if (defaults.name_prefix) {
    formData.value.name = defaults.name_prefix
  }

  if (defaults.priority) {
    formData.value.priority = defaults.priority
  }

  if (defaults.state_id) {
    formData.value.state_id = defaults.state_id
  }

  if (defaults.assignee_ids && defaults.assignee_ids.length > 0) {
    formData.value.assignee_id = defaults.assignee_ids[0]
  }

  if (defaults.label_ids && defaults.label_ids.length > 0) {
    selectedLabelIds.value = [...defaults.label_ids]
  }

  if (defaults.module_id) {
    formData.value.module_id = defaults.module_id
  }

  if (defaults.description_html) {
    formData.value.description = defaults.description_html
  }
}

function applyDefaultTemplateForType(typeId: number | null) {
  if (!typeId) {
    selectedTemplateId.value = null
    return
  }

  const defaultTemplate = templates.value.find(t => t.is_default && t.issue_type_id === typeId)
  if (defaultTemplate) {
    selectedTemplateId.value = defaultTemplate.id
    applyTemplate(defaultTemplate)
  } else {
    selectedTemplateId.value = null
  }
}

function handleTemplateChange(templateId: number | null) {
  if (!templateId) {
    selectedTemplateId.value = null
    return
  }
  const template = templates.value.find(t => t.id === templateId)
  if (template) {
    selectedTemplateId.value = templateId
    applyTemplate(template)
  }
}

// 加载数据
async function loadData() {
  try {
    if (slug.value) {
      try {
        const ws = await workspaceApi.getBySlug(slug.value)
        workspaceId.value = ws.id
      } catch (e) {
        console.error('Failed to load workspace:', e)
      }
    }
    
    // 加载工作项类型 - 如果 API 不可用则使用默认类型
    try {
      const typesRes = await issueTypeApi.getIssueTypes(workspaceId.value, projectId.value)
      issueTypes.value = typesRes
      
      // 设置默认类型
      const defaultType = typesRes.find((t: any) => t.is_default) || typesRes[0]
      if (defaultType) {
        selectedTypeId.value = defaultType.id
      }
    } catch (e) {
      // API 不可用时使用默认类型 ID = 1
      console.warn('Issue types API not available, using default type ID 1')
      selectedTypeId.value = 1
    }

    // 加载状态
    try {
      const statesRes = await stateApi.listStates(projectId.value)
      states.value = statesRes
      if (statesRes.length > 0) {
        const todoState = statesRes.find((s: any) => s.group === 'todo') || statesRes[0]
        formData.value.state_id = todoState.id
      }
    } catch (e) {
      console.error('Failed to load states:', e)
    }

    // 加载周期
    try {
      const cyclesRes = await cycleApi.listCycles(projectId.value)
      cycles.value = cyclesRes.items
    } catch (e) {
      console.error('Failed to load cycles:', e)
    }

    // 加载模块
    try {
      const modulesRes = await moduleApi.listModules(projectId.value, workspaceId.value)
      modules.value = modulesRes
    } catch (e) {
      console.error('Failed to load modules:', e)
    }

    // 加载项目成员
    try {
      const membersRes = await projectApi.listProjectMembers(projectId.value)
      projectMembers.value = membersRes.map((m: any) => m.user)
    } catch (e) {
      console.error('Failed to load project members:', e)
    }

    // 加载工作项模板
    try {
      await loadTemplates()
    } catch (e) {
      console.error('Failed to load templates:', e)
    }
  } catch (error) {
    console.error('Failed to load data:', error)
  }
}

// 搜索父工作项
async function searchParents() {
  if (parentSearchTimer) clearTimeout(parentSearchTimer)
  if (!parentSearch.value.trim()) {
    parentResults.value = []
    return
  }
  parentSearchTimer = setTimeout(async () => {
    try {
      const result: any = await issueApi.searchIssues(workspaceId.value, parentSearch.value, projectId.value)
      parentResults.value = (Array.isArray(result) ? result : (result.items || [])).slice(0, 10)
    } catch (e) {
      console.error('Parent search failed:', e)
    }
  }, 300)
}

function selectParent(p: any) {
  selectedParent.value = p
  parentResults.value = []
  parentSearch.value = ''
}

// 加载类型的关联字段
async function loadLinkedFields(typeId: number) {
  try {
    const fields = await issueTypeApi.getIssueTypeFields(typeId, projectId.value)
    linkedFields.value = fields
  } catch (error) {
    console.error('Failed to load linked fields:', error)
    linkedFields.value = []
  }
}

// 更新自定义字段值
function updateCustomFieldValue(fieldId: number, value: IssueCustomFieldValueUpdate) {
  customFieldValues.value[fieldId] = value
}

// 提交表单
async function submitForm() {
  // 检查标题
  if (!formData.value.name.trim()) {
    toast.warning(t('issue.titleRequiredMsg'))
    return
  }
  
  // 检查类型
  if (!selectedTypeId.value) {
    toast.warning(t('issue.typeRequiredMsg'))
    return
  }

  // 检查必填自定义字段
  const missingFields = linkedFields.value
    .filter(field => isFieldRequired(field) && !hasFieldValue(field.field_id))
    .map(field => field.name)
  if (missingFields.length > 0) {
    toast.warning(t('issue.requiredFieldsMsg', { fields: missingFields.join(', ') }))
    return
  }

  saving.value = true
  try {
    const data: any = {
      name: formData.value.name,
      type_id: selectedTypeId.value,
      priority: formData.value.priority || 'none'
    }
    
    // Add description if provided
    if (formData.value.description && formData.value.description.trim() !== '<p></p>') {
      data.description_html = formData.value.description
    }
    
    // Add state_id if selected
    if (formData.value.state_id) {
      const stateId = typeof formData.value.state_id === 'string' 
        ? parseInt(formData.value.state_id) 
        : formData.value.state_id
      if (stateId > 0) {
        data.state_id = stateId
      }
    }
    
    // Add dates in RFC3339 format
    if (formData.value.start_date) {
      data.start_date = formData.value.start_date + 'T00:00:00Z'
    }
    if (formData.value.target_date) {
      data.target_date = formData.value.target_date + 'T00:00:00Z'
    }
    
    // Add assignee if selected
    if (formData.value.assignee_id) {
      const assigneeId = typeof formData.value.assignee_id === 'string'
        ? parseInt(formData.value.assignee_id)
        : formData.value.assignee_id
      if (assigneeId > 0) {
        data.assignee_ids = [assigneeId]
      }
    }

    // Add labels if selected
    if (selectedLabelIds.value.length > 0) {
      data.label_ids = selectedLabelIds.value
    }

    // Add module if selected
    if (formData.value.module_id) {
      const moduleId = typeof formData.value.module_id === 'string'
        ? parseInt(formData.value.module_id)
        : formData.value.module_id
      if (moduleId > 0) {
        data.module_id = moduleId
      }
    }

    // Add parent if selected
    if (selectedParent.value) {
      data.parent_id = selectedParent.value.id
    }

    // Add custom field values
    const cfValues: Record<number, string> = {}
    for (const [fieldId, v] of Object.entries(customFieldValues.value)) {
      const val = v as any
      const field = linkedFields.value.find((f: any) => f.field_id === parseInt(fieldId))
      let value = ''
      switch (field?.field_type) {
        case 'text':  value = val.text_value || ''; break
        case 'number': value = val.number_value !== undefined ? String(val.number_value) : ''; break
        case 'boolean': value = val.boolean_value !== undefined ? String(val.boolean_value) : ''; break
        case 'date':  value = val.date_value || ''; break
        case 'url':   value = val.url_value || ''; break
        case 'dropdown':
        case 'member': value = val.json_value?.length ? JSON.stringify(val.json_value) : ''; break
        default: value = val.text_value || ''
      }
      if (value !== '') {
        cfValues[parseInt(fieldId)] = value
      }
    }
    if (Object.keys(cfValues).length > 0) {
      data.custom_field_values = cfValues
    }

    console.log('Creating issue with data:', data)
    const issue = await issueApi.createIssue(projectId.value, workspaceId.value, data)
    console.log('Created issue:', issue)

    // 显示成功提示
    toast.success(t('issue.createSuccess'))

    // 刷新工作项列表
    emit('created', issue)

    // 返回到项目页面，并保持当前的视图状态
    router.push(`/workspaces/${workspaceId.value}/projects/${projectId.value}?view=${returnView.value}`)
  } catch (error: any) {
    console.error('Failed to create issue:', error)
    const errorMsg = error.response?.data?.message || error.message || t('issue.unknownError')
    toast.error(t('issue.createFailed', { msg: errorMsg }))
  } finally {
    saving.value = false
  }
}

// 返回
function goBack() {
  router.back()
}

// 监听类型变化
watch(selectedTypeId, (newTypeId) => {
  if (newTypeId) {
    loadLinkedFields(newTypeId)
    applyDefaultTemplateForType(newTypeId)
  }
})

onMounted(async () => {
  await loadData()
  // Auto-populate parent from query param (e.g. ?parent_id=X from IssueDetail)
  const parentId = route.query.parent_id
  if (parentId) {
    try {
      const parent = await issueApi.getIssue(Number(parentId))
      if (parent) selectParent(parent)
    } catch (e) {
      console.error('Failed to load parent issue:', e)
    }
  }
})
</script>

<style scoped>
.issue-create-page {
  @apply min-h-screen bg-white flex flex-col;
}

.page-header {
  @apply flex items-center space-x-3 px-6 py-3 bg-white border-b border-gray-100;
}

.back-btn {
  @apply p-2 rounded-lg hover:bg-gray-100 transition;
}

.page-title {
  @apply text-base font-semibold text-gray-800;
}

.page-content {
  @apply flex flex-1;
}

.form-section {
  @apply flex-1 bg-white border-r border-gray-100 p-6;
}

.properties-section {
  @apply w-64 bg-white p-6 h-fit;
}

.form-group {
  @apply mb-5;
}

.form-label {
  @apply block text-xs font-medium text-gray-400 uppercase tracking-wide mb-2;
}

.required {
  @apply text-red-500;
}

.title-input {
  @apply text-lg font-medium;
}

.form-input,
.form-textarea,
.property-select,
.property-input {
  @apply w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-blue-400 focus:border-blue-400;
}

.form-textarea {
  @apply resize-none;
}

.type-selector {
  @apply flex flex-wrap gap-2;
}

.type-option {
  @apply flex items-center space-x-2 px-3 py-2 rounded-lg border border-gray-100 hover:border-gray-200 transition;
}

.type-option.active {
  @apply border-gray-300 bg-gray-50;
}

.type-icon {
  @apply w-6 h-6 rounded flex items-center justify-center text-white;
}

.property-group {
  @apply mb-4;
}

.property-label {
  @apply block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1.5;
}

.custom-fields {
  @apply space-y-3;
}

.field-item {
  @apply border border-gray-100 rounded-lg p-3;
}

.field-label {
  @apply block text-sm font-medium text-gray-600 mb-2;
}

.page-footer {
  @apply flex items-center justify-end space-x-3 px-6 py-3 bg-white border-t border-gray-100;
}

.btn {
  @apply px-4 py-2 rounded-lg font-medium transition;
}

.btn-primary {
  @apply bg-neutral-900 text-white hover:bg-neutral-800 disabled:opacity-50 disabled:cursor-not-allowed;
}

.btn-secondary {
  @apply bg-gray-100 text-gray-600 hover:bg-gray-200;
}
</style>
