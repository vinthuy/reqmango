<template>
  <div class="issue-create-page">
    <!-- 头部 -->
    <div class="page-header">
      <button @click="goBack" class="back-btn">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <h1 class="page-title">创建工作项</h1>
    </div>

    <div class="page-content">
      <!-- 左侧表单 -->
      <div class="form-section">
        <!-- 类型选择 -->
        <div class="form-group">
          <label class="form-label">类型</label>
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

        <!-- 标题 -->
        <div class="form-group">
          <label class="form-label">标题 <span class="required">*</span></label>
          <input
            v-model="formData.name"
            type="text"
            class="form-input title-input"
            placeholder="输入工作项标题"
            ref="titleInput"
          />
        </div>

        <!-- 描述 -->
        <div class="form-group">
          <label class="form-label">描述</label>
          <RichTextEditor
            v-model="formData.description"
            placeholder="添加描述..."
          />
        </div>

        <!-- 自定义字段 -->
        <div v-if="linkedFields.length > 0" class="form-group">
          <label class="form-label">自定义属性</label>
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
          <label class="property-label">状态</label>
          <select v-model="formData.state_id" class="property-select">
            <option value="">选择状态</option>
            <option v-for="state in states" :key="state.id" :value="state.id">
              {{ state.name }}
            </option>
          </select>
        </div>

        <div class="property-group">
          <label class="property-label">优先级</label>
          <select v-model="formData.priority" class="property-select">
            <option value="">选择优先级</option>
            <option value="urgent">紧急</option>
            <option value="high">高</option>
            <option value="medium">中</option>
            <option value="low">低</option>
            <option value="none">无</option>
          </select>
        </div>

        <div class="property-group">
          <label class="property-label">负责人</label>
          <UserSelect
            v-model="formData.assignee_id"
            :users="projectMembers"
            placeholder="选择负责人"
            :clearable="true"
          />
        </div>

        <div class="property-group">
          <label class="property-label">周期</label>
          <select v-model="formData.cycle_id" class="property-select">
            <option value="">选择周期</option>
            <option v-for="cycle in cycles" :key="cycle.id" :value="cycle.id">
              {{ cycle.name }}
            </option>
          </select>
        </div>

        <div class="property-group">
          <label class="property-label">模块</label>
          <select v-model="formData.module_id" class="property-select">
            <option value="">选择模块</option>
            <option v-for="module in modules" :key="module.id" :value="module.id">
              {{ module.name }}
            </option>
          </select>
        </div>

        <div class="property-group">
          <label class="property-label">开始日期</label>
          <input v-model="formData.start_date" type="date" class="property-input" />
        </div>

        <div class="property-group">
          <label class="property-label">截止日期</label>
          <input v-model="formData.target_date" type="date" class="property-input" />
        </div>
      </div>
    </div>

    <!-- 底部操作 -->
    <div class="page-footer">
      <button @click="goBack" class="btn btn-secondary">取消</button>
      <button @click="submitForm" class="btn btn-primary" :disabled="!canSubmit">
        创建工作项
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CustomFieldValueInput from '@/components/CustomFieldValueInput.vue'
import RichTextEditor from '@/components/RichTextEditor.vue'
import UserSelect from '@/components/UserSelect.vue'
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
import customFieldApi from '@/api/custom-field'

const route = useRoute()
const router = useRouter()

const projectId = computed(() => parseInt(route.params.projectId as string, 10))
const workspaceId = computed(() => parseInt(route.params.workspaceId as string, 10))
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
  target_date: ''
})

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

// 判断字段是否必填
function isFieldRequired(field: any): boolean {
  return field.is_required === true
}

// 判断字段是否有值
function hasFieldValue(fieldId: number): boolean {
  const value = customFieldValues.value[fieldId]
  if (!value) return false
  return !!(value.text_value || value.number_value !== undefined || value.boolean_value ||
    value.date_value || value.url_value || (value.json_value && value.json_value.length > 0))
}

// 加载数据
async function loadData() {
  try {
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
  } catch (error) {
    console.error('Failed to load data:', error)
  }
}

// 加载类型的关联字段
async function loadLinkedFields(typeId: number) {
  try {
    const fields = await issueTypeApi.getIssueTypeFields(typeId)
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
    alert('请输入工作项标题')
    return
  }
  
  // 检查类型
  if (!selectedTypeId.value) {
    alert('请选择工作项类型')
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

    console.log('Creating issue with data:', data)
    const issue = await issueApi.createIssue(projectId.value, workspaceId.value, data)
    console.log('Created issue:', issue)

    // 保存自定义字段值
    if (Object.keys(customFieldValues.value).length > 0) {
      try {
        const values = Object.entries(customFieldValues.value)
          .filter(([_, v]) => v)
          .map(([fieldId, v]) => ({
            field_id: parseInt(fieldId),
            value: (v as any).text_value ?? (v as any).date_value ?? (v as any).url_value ??
                   ((v as any).number_value !== undefined ? String((v as any).number_value) : '') ??
                   ((v as any).boolean_value !== undefined ? String((v as any).boolean_value) : '') ??
                   ((v as any).json_value?.length ? JSON.stringify((v as any).json_value) : '') ?? ''
          }))
          .filter((v: any) => v.value !== '')
        if (values.length > 0) {
          await customFieldApi.bulkUpdateIssueCustomFieldValues(issue.id, values as any)
        }
      } catch (e) {
        console.error('Failed to save custom field values:', e)
      }
    }

    // 刷新工作项列表
    emit('created', issue)

    // 返回到项目页面，并保持当前的视图状态
    router.push(`/workspaces/${workspaceId.value}/projects/${projectId.value}?view=${returnView.value}`)
  } catch (error: any) {
    console.error('Failed to create issue:', error)
    const errorMsg = error.response?.data?.message || error.message || '未知错误'
    alert('创建失败: ' + errorMsg)
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
  }
})

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.issue-create-page {
  @apply min-h-screen bg-gray-50 flex flex-col;
}

.page-header {
  @apply flex items-center space-x-4 px-6 py-4 bg-white border-b;
}

.back-btn {
  @apply p-2 rounded-lg hover:bg-gray-100 transition;
}

.page-title {
  @apply text-xl font-semibold text-gray-900;
}

.page-content {
  @apply flex flex-1 gap-6 p-6;
}

.form-section {
  @apply flex-1 bg-white rounded-lg border p-6;
}

.properties-section {
  @apply w-72 bg-white rounded-lg border p-6 h-fit;
}

.form-group {
  @apply mb-6;
}

.form-label {
  @apply block text-sm font-medium text-gray-700 mb-2;
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
  @apply w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500;
}

.form-textarea {
  @apply resize-none;
}

.type-selector {
  @apply flex flex-wrap gap-2;
}

.type-option {
  @apply flex items-center space-x-2 px-3 py-2 rounded-lg border border-gray-200 hover:border-indigo-300 transition;
}

.type-option.active {
  @apply border-indigo-500 bg-indigo-50;
}

.type-icon {
  @apply w-6 h-6 rounded flex items-center justify-center text-white;
}

.property-group {
  @apply mb-4;
}

.property-label {
  @apply block text-sm font-medium text-gray-700 mb-1;
}

.custom-fields {
  @apply space-y-4;
}

.field-item {
  @apply border border-gray-200 rounded-lg p-4;
}

.field-label {
  @apply block text-sm font-medium text-gray-700 mb-2;
}

.page-footer {
  @apply flex items-center justify-end space-x-4 px-6 py-4 bg-white border-t;
}

.btn {
  @apply px-4 py-2 rounded-lg font-medium transition;
}

.btn-primary {
  @apply bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed;
}

.btn-secondary {
  @apply bg-gray-100 text-gray-700 hover:bg-gray-200;
}
</style>
