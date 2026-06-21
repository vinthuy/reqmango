<template>
  <div class="custom-field-manager">
    <!-- 字段列表显示模式 -->
    <div v-if="mode === 'display'" class="space-y-4">
      <div
        v-for="item in fieldsWithValues"
        :key="item.field.id"
        class="flex items-start space-x-4 py-2"
      >
        <label class="text-sm font-medium text-gray-600 w-32 shrink-0">
          {{ item.field.name }}
          <span v-if="item.field.is_required" class="text-red-500">*</span>
        </label>
        <div class="flex-1">
          <CustomFieldValueInput
            :field="item.field"
            :value="getFieldValue(item)"
            :members="members"
            @update:value="updateFieldValue(item.field.id, $event)"
          />
        </div>
      </div>
      
      <!-- 空状态 -->
      <div v-if="fieldsWithValues.length === 0" class="text-center py-8 text-gray-500">
        暂无自定义字段
      </div>
    </div>

    <!-- 字段管理模式 -->
    <div v-else-if="mode === 'manage'" class="space-y-6">
      <!-- 创建新字段按钮 -->
      <button
        @click="showCreateModal = true"
        class="w-full px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 transition"
      >
        + 创建自定义字段
      </button>

      <!-- 字段列表 -->
      <div class="space-y-4">
        <div
          v-for="field in fields"
          :key="field.id"
          class="border border-gray-200 rounded-lg p-4 hover:border-gray-300 transition"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center space-x-2">
                <h4 class="font-medium text-gray-900">{{ field.name }}</h4>
                <span
                  class="px-2 py-0.5 text-xs rounded"
                  :class="getFieldTypeBadgeClass(field.field_type)"
                >
                  {{ getFieldTypeName(field.field_type) }}
                </span>
                <span v-if="field.is_required" class="text-xs text-red-500">必填</span>
                <span v-if="!field.is_active" class="text-xs text-gray-400">已禁用</span>
              </div>
              <p v-if="field.description" class="text-sm text-gray-500 mt-1">
                {{ field.description }}
              </p>
              
              <!-- 下拉选项显示 -->
              <div v-if="field.field_type === 'dropdown' && field.options.length > 0" class="mt-2">
                <p class="text-xs text-gray-500 mb-1">选项:</p>
                <div class="flex flex-wrap gap-2">
                  <span
                    v-for="option in field.options"
                    :key="option.id"
                    class="px-2 py-1 text-xs rounded"
                    :style="{ backgroundColor: option.color || '#E5E7EB' }"
                  >
                    {{ option.value }}
                  </span>
                </div>
              </div>
            </div>
            
            <!-- 操作按钮 -->
            <div class="flex items-center space-x-2">
              <button
                @click="editField(field)"
                class="text-gray-500 hover:text-indigo-600 transition"
                title="编辑"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button
                @click="deleteField(field.id)"
                class="text-gray-500 hover:text-red-600 transition"
                title="删除"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建/编辑字段模态框 -->
    <div
      v-if="showCreateModal || showEditModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
    >
      <div class="bg-white rounded-lg p-6 w-full max-w-lg mx-4">
        <h3 class="text-lg font-semibold mb-4">
          {{ showEditModal ? '编辑字段' : '创建自定义字段' }}
        </h3>
        
        <form @submit.prevent="submitFieldForm" class="space-y-4">
          <!-- 字段名称 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">字段名称 *</label>
            <input
              v-model="fieldForm.name"
              type="text"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          
          <!-- 字段描述 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">描述</label>
            <input
              v-model="fieldForm.description"
              type="text"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          
          <!-- 字段类型 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">字段类型 *</label>
            <select
              v-model="fieldForm.field_type"
              required
              :disabled="showEditModal"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="text">文本</option>
              <option value="number">数字</option>
              <option value="dropdown">下拉选择</option>
              <option value="boolean">布尔值</option>
              <option value="date">日期</option>
              <option value="member">成员选择</option>
              <option value="url">链接</option>
            </select>
          </div>
          
          <!-- 文本类型属性 -->
          <div v-if="fieldForm.field_type === 'text'">
            <label class="block text-sm font-medium text-gray-700 mb-1">文本类型</label>
            <select
              v-model="fieldForm.text_type"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="single">单行文本</option>
              <option value="paragraph">多行文本</option>
              <option value="readonly">只读文本</option>
            </select>
          </div>
          
          <!-- 数字类型属性 -->
          <div v-if="fieldForm.field_type === 'number'" class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">最小值</label>
              <input
                v-model.number="fieldForm.number_min"
                type="number"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">最大值</label>
              <input
                v-model.number="fieldForm.number_max"
                type="number"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
            </div>
          </div>
          
          <!-- 下拉类型属性 -->
          <div v-if="fieldForm.field_type === 'dropdown'">
            <label class="flex items-center space-x-2 mb-2">
              <input
                v-model="fieldForm.is_multi_select"
                type="checkbox"
                class="w-4 h-4 text-indigo-600 border-gray-300 rounded"
              />
              <span class="text-sm text-gray-700">允许多选</span>
            </label>
            
            <div class="space-y-2">
              <label class="block text-sm font-medium text-gray-700">选项列表</label>
              <div
                v-for="(option, index) in fieldForm.options"
                :key="index"
                class="flex items-center space-x-2"
              >
                <input
                  v-model="option.value"
                  type="text"
                  placeholder="选项值"
                  class="flex-1 px-3 py-2 border border-gray-300 rounded-md"
                />
                <input
                  v-model="option.color"
                  type="color"
                  class="w-10 h-10 border border-gray-300 rounded-md cursor-pointer"
                />
                <button
                  @click="removeOption(index)"
                  class="text-red-500 hover:text-red-700"
                >
                  ×
                </button>
              </div>
              <button
                @click="addOption"
                type="button"
                class="text-sm text-indigo-600 hover:text-indigo-800"
              >
                + 添加选项
              </button>
            </div>
          </div>
          
          <!-- 必填/只读 -->
          <div class="flex items-center space-x-4">
            <label class="flex items-center space-x-2">
              <input
                v-model="fieldForm.is_required"
                type="checkbox"
                class="w-4 h-4 text-indigo-600 border-gray-300 rounded"
              />
              <span class="text-sm text-gray-700">必填</span>
            </label>
            <label class="flex items-center space-x-2">
              <input
                v-model="fieldForm.is_readonly"
                type="checkbox"
                class="w-4 h-4 text-indigo-600 border-gray-300 rounded"
              />
              <span class="text-sm text-gray-700">只读</span>
            </label>
          </div>
          
          <!-- 操作按钮 -->
          <div class="flex justify-end space-x-3 pt-4">
            <button
              @click="closeModal"
              type="button"
              class="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              取消
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
            >
              {{ showEditModal ? '保存' : '创建' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import CustomFieldValueInput from './CustomFieldValueInput.vue'
import customFieldApi from '@/api/custom-field'
import type {
  CustomField,
  CustomFieldCreate,
  CustomFieldUpdate,
  IssueCustomFieldValueUpdate,
  CustomFieldWithValues,
  CustomFieldTypeEnum
} from '@/types/custom-field'
import { useConfirm } from '@/composables/useConfirm'
import { getFieldTypeName } from '@/types/custom-field'

// Props
const props = defineProps<{
  workspaceId: number
  projectId?: number
  issueId?: number
  mode?: 'display' | 'manage'  // display: 显示字段值, manage: 管理字段定义
  members?: Array<{ id: number; display_name?: string; email: string }>
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:values', values: Record<string, IssueCustomFieldValueUpdate>): void
  (e: 'field-created', field: CustomField): void
  (e: 'field-updated', field: CustomField): void
  (e: 'field-deleted', fieldId: number): void
}>()

// State
const { confirm } = useConfirm()
const fields = ref<CustomField[]>([])
const fieldsWithValues = ref<CustomFieldWithValues[]>([])
const fieldValues = ref<Record<string, IssueCustomFieldValueUpdate>>({})
const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingField = ref<CustomField | null>(null)

// Field form
const fieldForm = ref<CustomFieldCreate>({
  name: '',
  description: '',
  field_type: 'text' as CustomFieldTypeEnum,
  is_required: false,
  is_readonly: false,
  is_active: true,
  is_multi_select: false,
  options: []
})

// Load fields on mount
onMounted(async () => {
  await loadFields()
  if (props.issueId && props.mode === 'display') {
    await loadFieldValues()
  }
})

// Watch for prop changes
watch(() => props.projectId, async () => {
  await loadFields()
})

watch(() => props.issueId, async () => {
  if (props.issueId && props.mode === 'display') {
    await loadFieldValues()
  }
})

// Load fields
async function loadFields() {
  try {
    fields.value = await customFieldApi.listCustomFields(
      props.workspaceId,
      props.projectId,
      undefined,
      props.mode === 'manage'  // 管理模式下显示所有字段
    )
    
    if (props.mode === 'display') {
      // 转换为 fieldsWithValues 格式
      fieldsWithValues.value = fields.value.map(field => ({
        field: {
          id: field.id,
          name: field.name,
          field_type: field.field_type,
          is_required: field.is_required,
          is_readonly: field.is_readonly,
          options: field.options
        },
        value: fieldValues.value[field.id]
      }))
    }
  } catch (error) {
    console.error('Failed to load custom fields:', error)
  }
}

// Load field values for issue
async function loadFieldValues() {
  if (!props.issueId) return
  
  try {
    const response = await customFieldApi.getIssueCustomFieldsWithDefinitions(props.issueId)
    fieldsWithValues.value = response.fields
    
    // 更新 fieldValues
    response.fields.forEach(item => {
      if (item.value) {
        fieldValues.value[item.field.id] = convertValueToUpdate(item)
      }
    })
  } catch (error) {
    console.error('Failed to load field values:', error)
  }
}

// Convert display value to update format
function convertValueToUpdate(item: CustomFieldWithValues): IssueCustomFieldValueUpdate {
  const field = item.field
  const value = item.value
  
  const update: IssueCustomFieldValueUpdate = {}
  
  switch (field.field_type) {
    case 'text':
      update.text_value = value as string
      break
    case 'number':
      update.number_value = value as number
      break
    case 'boolean':
      update.boolean_value = value as boolean
      break
    case 'date':
      update.date_value = value as string
      break
    case 'url':
      update.url_value = value as string
      break
    case 'dropdown':
    case 'member':
      update.json_value = value as number[]
      break
  }
  
  return update
}

// Get field value from fieldsWithValues
function getFieldValue(item: CustomFieldWithValues): IssueCustomFieldValueUpdate {
  return fieldValues.value[item.field.id] || {}
}

// Update field value
function updateFieldValue(fieldId: number, value: IssueCustomFieldValueUpdate) {
  fieldValues.value[fieldId] = value
  emit('update:values', fieldValues.value)
}

// Field type badge class
function getFieldTypeBadgeClass(type: string): string {
  const classes: Record<string, string> = {
    text: 'bg-blue-100 text-blue-800',
    number: 'bg-green-100 text-green-800',
    dropdown: 'bg-purple-100 text-purple-800',
    boolean: 'bg-yellow-100 text-yellow-800',
    date: 'bg-orange-100 text-orange-800',
    member: 'bg-pink-100 text-pink-800',
    url: 'bg-cyan-100 text-cyan-800'
  }
  return classes[type] || 'bg-gray-100 text-gray-800'
}

// Edit field
function editField(field: CustomField) {
  editingField.value = field
  fieldForm.value = {
    name: field.name,
    description: field.description || '',
    field_type: field.field_type as CustomFieldTypeEnum,
    is_required: field.is_required,
    is_readonly: field.is_readonly,
    is_active: field.is_active,
    text_type: field.text_type,
    placeholder: field.placeholder,
    number_default: field.number_default,
    number_min: field.number_min,
    number_max: field.number_max,
    is_multi_select: field.is_multi_select,
    options: field.options.map(o => ({
      value: o.value,
      color: o.color,
      is_default: o.is_default,
      is_active: o.is_active
    }))
  }
  showEditModal.value = true
}

// Delete field
async function deleteField(fieldId: number) {
  if (!(await confirm('确定要删除此字段吗？'))) return
  
  try {
    await customFieldApi.deleteCustomField(fieldId)
    emit('field-deleted', fieldId)
    await loadFields()
  } catch (error) {
    console.error('Failed to delete field:', error)
  }
}

// Add option
function addOption() {
  fieldForm.value.options?.push({
    value: '',
    color: '#E5E7EB',
    is_default: false,
    is_active: true
  })
}

// Remove option
function removeOption(index: number) {
  fieldForm.value.options?.splice(index, 1)
}

// Submit field form
async function submitFieldForm() {
  try {
    if (showEditModal.value && editingField.value) {
      // Update existing field
      const updateData: CustomFieldUpdate = {
        name: fieldForm.value.name,
        description: fieldForm.value.description,
        is_required: fieldForm.value.is_required,
        is_readonly: fieldForm.value.is_readonly,
        is_active: fieldForm.value.is_active,
        is_multi_select: fieldForm.value.is_multi_select,
        number_min: fieldForm.value.number_min,
        number_max: fieldForm.value.number_max
      }
      const updated = await customFieldApi.updateCustomField(editingField.value.id, updateData)
      emit('field-updated', updated)
    } else {
      // Create new field
      const created = await customFieldApi.createCustomField(props.workspaceId, fieldForm.value)
      emit('field-created', created)
    }
    
    closeModal()
    await loadFields()
  } catch (error) {
    console.error('Failed to save field:', error)
  }
}

// Close modal
function closeModal() {
  showCreateModal.value = false
  showEditModal.value = false
  editingField.value = null
  fieldForm.value = {
    name: '',
    description: '',
    field_type: 'text' as CustomFieldTypeEnum,
    is_required: false,
    is_readonly: false,
    is_active: true,
    is_multi_select: false,
    options: []
  }
}

// Expose methods for parent component
defineExpose({
  saveValues: async () => {
    if (!props.issueId) return
    
    const values = Object.entries(fieldValues.value).map(([fieldId, value]) => ({
      field_id: fieldId,
      ...value
    }))
    
    if (values.length > 0) {
      await customFieldApi.bulkUpdateIssueCustomFieldValues(props.issueId, values)
    }
  },
  getValues: () => fieldValues.value
})
</script>

<style scoped>
.custom-field-manager {
  width: 100%;
}
</style>