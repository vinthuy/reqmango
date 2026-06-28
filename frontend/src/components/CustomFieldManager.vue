<template>
  <div class="w-full">
    <!-- 头部 -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">自定义字段</h2>
        <p class="text-sm text-gray-500 mt-0.5">定义工作空间中可用的自定义字段（如优先级、版本号等）</p>
      </div>
      <button v-if="mode !== 'display'" @click="openCreateModal" class="create-btn">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        新建字段
      </button>
    </div>

    <!-- 字段列表 -->
    <div v-if="customFields.length > 0" class="space-y-3">
      <div
        v-for="field in customFields"
        :key="field.id"
        class="field-row"
        :class="{ 'is-default': field.is_required }"
      >
        <div class="field-row-main" @click="openEditModal(field)">
          <div class="field-icon" :class="getFieldTypeClass(field.field_type)">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path v-if="field.field_type === 'text'" stroke-linecap="round" stroke-linejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              <path v-else-if="field.field_type === 'number'" stroke-linecap="round" stroke-linejoin="round" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
              <path v-else-if="field.field_type === 'dropdown'" stroke-linecap="round" stroke-linejoin="round" d="M8 9l4-4 4 4m0 6l-4 4-4-4" />
              <path v-else-if="field.field_type === 'boolean'" stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              <path v-else-if="field.field_type === 'date'" stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              <path v-else-if="field.field_type === 'member'" stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              <path v-else stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center space-x-2">
              <span class="font-medium text-gray-900">{{ field.name }}</span>
              <span class="field-type-badge">{{ getFieldTypeName(field.field_type) }}</span>
              <span v-if="field.is_required" class="badge-required">必填</span>
              <span v-if="field.is_readonly" class="badge-readonly">只读</span>
              <span v-if="!field.is_active" class="badge-inactive">已禁用</span>
            </div>
            <p v-if="field.description" class="text-xs text-gray-500 mt-0.5 truncate">{{ field.description }}</p>
            <p v-if="field.field_type === 'dropdown' && field.options?.length > 0" class="text-xs text-gray-400 mt-0.5">
              {{ field.options?.length }} 个选项{{ field.is_multi_select ? ' (多选)' : '' }}
            </p>
          </div>
        </div>
        <div class="field-row-actions">
          <button
            @click.stop="toggleActive(field)"
            class="icon-action"
            :title="field.is_active ? '禁用' : '启用'"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path v-if="field.is_active" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
              <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>
          <button
            @click.stop="confirmDelete(field)"
            class="icon-action icon-action-danger"
            title="删除"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>

        <!-- 值编辑区域（仅在显示模式且有 issueId 时显示） -->
        <div v-if="mode === 'display' && issueId && field.is_active && !field.is_readonly" class="field-value-input">
          <input
            v-if="field.field_type === 'text'"
            type="text"
            :value="getFieldValue(field)"
            @input="(e: Event) => setFieldValue(field, (e.target as HTMLInputElement).value)"
            class="form-input text-sm"
            :placeholder="field.placeholder || '输入值...'"
          />
          <input
            v-else-if="field.field_type === 'number'"
            type="number"
            :value="getFieldValue(field)"
            @input="(e: Event) => setFieldValue(field, (e.target as HTMLInputElement).value)"
            :min="field.number_min"
            :max="field.number_max"
            class="form-input text-sm"
            placeholder="输入数字..."
          />
          <input
            v-else-if="field.field_type === 'url'"
            type="url"
            :value="getFieldValue(field)"
            @input="(e: Event) => setFieldValue(field, (e.target as HTMLInputElement).value)"
            class="form-input text-sm"
            placeholder="https://..."
          />
          <input
            v-else-if="field.field_type === 'date'"
            type="date"
            :value="getFieldValue(field)"
            @input="(e: Event) => setFieldValue(field, (e.target as HTMLInputElement).value)"
            class="form-input text-sm"
          />
          <label
            v-else-if="field.field_type === 'boolean'"
            class="flex items-center space-x-2 cursor-pointer"
          >
            <input
              type="checkbox"
              :checked="getFieldValue(field)"
              @change="(e: Event) => setFieldValue(field, (e.target as HTMLInputElement).checked)"
              class="w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
            />
            <span class="text-sm text-gray-700">{{ getFieldValue(field) ? '是' : '否' }}</span>
          </label>
          <select
            v-else-if="field.field_type === 'dropdown' && !field.is_multi_select"
            :value="getFieldValue(field)"
            @change="(e: Event) => setFieldValue(field, (e.target as HTMLSelectElement).value ? Number((e.target as HTMLSelectElement).value) : '')"
            class="form-input text-sm"
          >
            <option value="">-- 选择 --</option>
            <option v-for="opt in field.options" :key="opt.id" :value="opt.id">{{ opt.value }}</option>
          </select>
          <div v-else-if="field.field_type === 'dropdown' && field.is_multi_select" class="flex flex-wrap gap-2">
            <label v-for="opt in field.options" :key="opt.id" class="flex items-center space-x-1 text-sm">
              <input
                type="checkbox"
                :checked="(getFieldValue(field) || []).includes(opt.id)"
                @change="(e: Event) => {
                  const cur = [...(getFieldValue(field) || [])]
                  if ((e.target as HTMLInputElement).checked) { cur.push(opt.id) }
                  else { const idx = cur.indexOf(opt.id); if (idx !== -1) cur.splice(idx, 1) }
                  setFieldValue(field, cur)
                }"
                class="w-3.5 h-3.5 rounded border-gray-300 text-indigo-600"
              />
              <span class="text-xs text-gray-600">{{ opt.value }}</span>
            </label>
          </div>
          <select
            v-else-if="field.field_type === 'member'"
            :value="getFieldValue(field)?.[0] || ''"
            @change="(e: Event) => setFieldValue(field, [Number((e.target as HTMLSelectElement).value)])"
            class="form-input text-sm"
          >
            <option value="">-- 选择 --</option>
            <option v-for="m in members" :key="m.id || m.user_id" :value="m.id || m.user_id">
              {{ m.display_name || m.username || m.name || m.email }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <svg class="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
      </svg>
      <p class="mt-2 text-sm text-gray-500">暂无自定义字段，点击上方按钮创建</p>
    </div>

    <!-- 编辑抽屉 -->
    <Transition name="slide-fade">
      <div v-if="showEditDrawer" class="edit-drawer-overlay" @click.self="closeDrawer">
        <div class="edit-drawer">
          <div class="drawer-header">
            <div>
              <h3 class="drawer-title">
                {{ isCreating ? '新建自定义字段' : '编辑: ' + (selectedField?.name || '') }}
              </h3>
              <p class="drawer-subtitle">配置自定义字段的属性</p>
            </div>
            <button @click="closeDrawer" class="close-btn">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="drawer-body">
            <div class="form-group">
              <label class="form-label">名称 <span class="required">*</span></label>
              <input v-model="formData.name" type="text" class="form-input" placeholder="输入字段名称" />
            </div>

            <div class="form-group">
              <label class="form-label">类型 <span class="required">*</span></label>
              <select
                v-model="formData.field_type"
                :disabled="!isCreating"
                class="form-input"
                @change="handleFieldTypeChange"
              >
                <option v-for="type in fieldTypes" :key="type.value" :value="type.value">
                  {{ type.label }}
                </option>
              </select>
            </div>

            <div class="form-group">
              <label class="form-label">描述</label>
              <textarea
                v-model="formData.description"
                rows="2"
                class="form-input"
                placeholder="字段描述（可选）"
              ></textarea>
            </div>

            <!-- 文本类型属性 -->
            <div v-if="formData.field_type === 'text'" class="form-group">
              <label class="form-label">文本类型</label>
              <select v-model="formData.text_type" class="form-input">
                <option value="single">单行文本</option>
                <option value="paragraph">多行文本</option>
                <option value="readonly">只读文本</option>
              </select>
            </div>

            <!-- 数字类型属性 -->
            <div v-if="formData.field_type === 'number'" class="space-y-3">
              <div class="form-group">
                <label class="form-label">最小值</label>
                <input v-model.number="formData.number_min" type="number" class="form-input" placeholder="最小值" />
              </div>
              <div class="form-group">
                <label class="form-label">最大值</label>
                <input v-model.number="formData.number_max" type="number" class="form-input" placeholder="最大值" />
              </div>
            </div>

            <!-- 下拉类型属性 -->
            <div v-if="formData.field_type === 'dropdown'" class="space-y-4">
              <div class="form-group">
                <label class="checkbox-label">
                  <input v-model="formData.is_multi_select" type="checkbox" class="checkbox" />
                  <span>允许多选</span>
                </label>
              </div>

              <div class="form-group">
                <div class="flex items-center justify-between">
                  <label class="form-label">选项列表</label>
                  <button @click="addOption" class="text-sm text-indigo-600 hover:text-indigo-700">+ 添加选项</button>
                </div>
                <div class="space-y-2 mt-2">
                  <div
                    v-for="(option, index) in formData.options"
                    :key="index"
                    class="option-row"
                  >
                    <div class="flex items-center space-x-2">
                      <button
                        class="w-8 h-8 rounded-full border-2 border-gray-300 hover:border-gray-400 transition"
                        :style="{ backgroundColor: option.color }"
                        @click="openColorPicker(index)"
                      />
                      <input
                        v-model="option.value"
                        type="text"
                        class="form-input flex-1"
                        :placeholder="'选项 ' + (index + 1)"
                      />
                      <button
                        @click="removeOption(index)"
                        class="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </div>
                  </div>
                  <div v-if="!formData.options || formData.options.length === 0" class="text-sm text-gray-400 text-center py-4">
                    点击"添加选项"创建下拉选项
                  </div>
                </div>
              </div>
            </div>

            <!-- 通用属性 -->
            <div class="form-group">
              <label class="checkbox-label">
                <input v-model="formData.is_required" type="checkbox" class="checkbox" />
                <span>必填字段</span>
              </label>
            </div>

            <div class="form-group">
              <label class="checkbox-label">
                <input v-model="formData.is_readonly" type="checkbox" class="checkbox" />
                <span>只读字段</span>
              </label>
            </div>
          </div>

          <div class="drawer-footer">
            <button @click="closeDrawer" class="btn btn-secondary">取消</button>
            <button @click="submitForm" class="btn btn-primary" :disabled="!formData.name">
              {{ isCreating ? '创建' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 颜色选择器弹窗 -->
    <div v-if="showColorPicker" class="fixed inset-0 bg-black bg-opacity-30 z-50 flex items-center justify-center" @click.self="showColorPicker = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-sm">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">选择颜色</h3>
        <div class="color-grid">
          <button
            v-for="color in COLOR_PALETTE"
            :key="color"
            class="color-btn"
            :class="{ active: selectedColor === color }"
            :style="{ backgroundColor: color }"
            @click="selectColor(color)"
          />
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showColorPicker = false" class="btn btn-secondary">取消</button>
          <button @click="confirmColor" class="btn btn-primary">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { CustomField, CustomFieldCreate, CustomFieldUpdate, CustomFieldOptionCreate } from '@/types/custom-field'
import { CustomFieldTypeEnum, getFieldTypeName } from '@/types/custom-field'
import { useConfirm } from '@/composables/useConfirm'
import * as customFieldApi from '@/api/custom-field'
import api from '@/api'

const props = defineProps<{
  workspaceId: number
  projectId?: number
  issueId?: number
  issueTypeId?: number
  mode?: string
  members?: any[]
}>()

const emit = defineEmits<{
  (e: 'update:values', values: any[]): void
}>()

const { confirm } = useConfirm()

const COLOR_PALETTE = [
  '#6366F1', '#8B5CF6', '#EC4899', '#F43F5E', '#F97316',
  '#EAB308', '#22C55E', '#14B8A6', '#06B6D4', '#3B82F6',
  '#64748B', '#94A3B8', '#A855F7', '#D946EF', '#F472B6',
  '#FB923C', '#FACC15', '#4ADE80', '#2DD4BF', '#60A5FA',
]

const fieldTypes = [
  { value: CustomFieldTypeEnum.TEXT, label: '文本' },
  { value: CustomFieldTypeEnum.NUMBER, label: '数字' },
  { value: CustomFieldTypeEnum.DROPDOWN, label: '下拉选择' },
  { value: CustomFieldTypeEnum.BOOLEAN, label: '布尔值' },
  { value: CustomFieldTypeEnum.DATE, label: '日期' },
  { value: CustomFieldTypeEnum.MEMBER, label: '成员选择' },
  { value: CustomFieldTypeEnum.URL, label: '链接' },
]

const customFields = ref<CustomField[]>([])
const fieldValues = ref<Record<number, any>>({})
const showEditDrawer = ref(false)
const isCreating = ref(false)
const selectedField = ref<CustomField | null>(null)
const showColorPicker = ref(false)
const colorPickerIndex = ref(0)
const selectedColor = ref('#6366F1')

const formData = ref<CustomFieldCreate & CustomFieldUpdate>({
  name: '',
  field_type: CustomFieldTypeEnum.TEXT,
  description: '',
  is_required: false,
  is_readonly: false,
  is_active: true,
  options: [],
})

function getFieldTypeClass(type: string) {
  const classes: Record<string, string> = {
    text: 'bg-blue-100 text-blue-700',
    number: 'bg-green-100 text-green-700',
    dropdown: 'bg-purple-100 text-purple-700',
    boolean: 'bg-orange-100 text-orange-700',
    date: 'bg-pink-100 text-pink-700',
    member: 'bg-cyan-100 text-cyan-700',
    url: 'bg-indigo-100 text-indigo-700',
  }
  return classes[type] || 'bg-gray-100 text-gray-700'
}

async function loadData() {
  if (!props.workspaceId) {
    console.warn('CustomFieldManager: workspaceId is 0 or undefined')
    return
  }
  try {
    const result = await customFieldApi.listCustomFields(props.workspaceId)
    customFields.value = result
    // After loading fields, load existing values for this issue
    if (props.issueId) {
      await loadIssueValues()
    }
  } catch (error: any) {
    console.error('CustomFieldManager: Failed to load custom fields:', error?.message || error)
  }
}

function handleFieldTypeChange() {
  if (formData.value.field_type !== 'dropdown') {
    formData.value.options = []
    formData.value.is_multi_select = false
  }
}

function addOption() {
  if (!formData.value.options) {
    formData.value.options = []
  }
  const newOption: CustomFieldOptionCreate = {
    value: '',
    color: COLOR_PALETTE[formData.value.options.length % COLOR_PALETTE.length],
    sequence: formData.value.options.length + 1,
    is_default: false,
    is_active: true,
  }
  formData.value.options.push(newOption)
}

function removeOption(index: number) {
  if (!formData.value.options) return
  formData.value.options.splice(index, 1)
  formData.value.options.forEach((opt, i) => {
    opt.sequence = i + 1
  })
}

function openColorPicker(index: number) {
  colorPickerIndex.value = index
  selectedColor.value = formData.value.options?.[index]?.color || COLOR_PALETTE[0]
  showColorPicker.value = true
}

function selectColor(color: string) {
  selectedColor.value = color
}

function confirmColor() {
  if (formData.value.options?.[colorPickerIndex.value]) {
    formData.value.options[colorPickerIndex.value].color = selectedColor.value
  }
  showColorPicker.value = false
}

function openCreateModal() {
  isCreating.value = true
  selectedField.value = null
  formData.value = {
    name: '',
    field_type: CustomFieldTypeEnum.TEXT,
    description: '',
    is_required: false,
    is_readonly: false,
    is_active: true,
    options: [],
  }
  showEditDrawer.value = true
}

function openEditModal(field: CustomField) {
  isCreating.value = false
  selectedField.value = field
  formData.value = {
    name: field.name,
    field_type: field.field_type,
    description: field.description || '',
    is_required: field.is_required,
    is_readonly: field.is_readonly,
    is_active: field.is_active,
    options: (field.options || []).map((opt) => ({
      value: opt.value,
      color: opt.color || COLOR_PALETTE[0],
      sequence: opt.sequence,
      is_default: opt.is_default,
      is_active: opt.is_active,
    })),
    is_multi_select: field.is_multi_select,
    text_type: field.text_type,
    number_min: field.number_min,
    number_max: field.number_max,
  }
  showEditDrawer.value = true
}

function closeDrawer() {
  showEditDrawer.value = false
  selectedField.value = null
  isCreating.value = false
}

async function submitForm() {
  if (!formData.value.name) return

  try {
    if (isCreating.value) {
      await customFieldApi.createCustomField(props.workspaceId, formData.value)
    } else if (selectedField.value) {
      await customFieldApi.updateCustomField(selectedField.value.id, formData.value)
    }
    closeDrawer()
    await loadData()
  } catch (error) {
    console.error('Failed to submit form:', error)
  }
}

async function toggleActive(field: CustomField) {
  if (await confirm({
    title: field.is_active ? '禁用字段' : '启用字段',
    message: `确定要${field.is_active ? '禁用' : '启用'}字段 "${field.name}" 吗？`,
    confirmText: field.is_active ? '禁用' : '启用',
    danger: field.is_active,
  })) {
    try {
      await customFieldApi.updateCustomField(field.id, { is_active: !field.is_active })
      await loadData()
    } catch (error) {
      console.error('Failed to toggle active:', error)
    }
  }
}

async function confirmDelete(field: CustomField) {
  if (await confirm({
    title: '删除字段',
    message: `确定要删除字段 "${field.name}" 吗？此操作不可撤销。`,
    danger: true,
    confirmText: '删除'
  })) {
    try {
      await customFieldApi.deleteCustomField(field.id)
      await loadData()
    } catch (error) {
      console.error('Failed to delete field:', error)
    }
  }
}

async function loadIssueValues() {
  if (!props.issueId) return
  try {
    const response = await customFieldApi.getIssueCustomFieldsWithDefinitions(props.issueId)
    if (response?.fields) {
      const values: Record<number, any> = {}
      for (const item of response.fields) {
        if (item.field && item.value !== undefined) {
          values[item.field.id] = item.value
        }
      }
      fieldValues.value = values
    }
  } catch (error) {
    console.error('Failed to load issue custom field values:', error)
  }
}

function getFieldValue(field: CustomField): any {
  // Return the current UI value, falling back to default
  if (fieldValues.value[field.id] !== undefined) {
    return fieldValues.value[field.id]
  }
  // Provide defaults per type
  switch (field.field_type) {
    case 'number': return field.number_default ?? ''
    case 'boolean': return false
    case 'dropdown': return field.is_multi_select ? [] : ''
    case 'date': return ''
    default: return ''
  }
}

function setFieldValue(field: CustomField, val: any) {
  fieldValues.value = { ...fieldValues.value, [field.id]: val }
}

async function saveValues() {
  if (!props.issueId) {
    console.warn('CustomFieldManager.saveValues: no issueId')
    return
  }

  const updates: { field_id: number; value: string }[] = []
  for (const field of customFields.value) {
    if (!field.is_active || field.is_readonly) continue
    const val = fieldValues.value[field.id]
    if (val === undefined || val === null || val === '') continue

    let stringVal: string
    if (field.field_type === 'boolean') {
      stringVal = val ? 'true' : 'false'
    } else if (field.field_type === 'dropdown' || field.field_type === 'member') {
      stringVal = JSON.stringify(Array.isArray(val) ? val : [val])
    } else {
      stringVal = String(val)
    }

    updates.push({ field_id: field.id, value: stringVal })
  }

  if (updates.length === 0) return

  try {
    const response = await api.post(`/custom-fields/issues/${props.issueId}/values/bulk`, { issue_id: props.issueId, values: updates })
    emit('update:values', response.data)
  } catch (error) {
    console.error('Failed to save custom field values:', error)
  }
}

defineExpose({ loadData, saveValues, loadIssueValues })

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.create-btn {
  @apply flex items-center space-x-1 px-3 py-1.5 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition;
}

.field-row {
  @apply flex items-center justify-between px-3 py-2 bg-white border border-gray-200 rounded-lg hover:border-indigo-300 transition;
}

.field-row.is-default {
  @apply border-indigo-200 bg-indigo-50;
}

.field-row-main {
  @apply flex items-center space-x-3 flex-1 min-w-0 cursor-pointer;
}

.field-icon {
  @apply w-8 h-8 rounded-md flex items-center justify-center shrink-0;
}

.field-type-badge {
  @apply px-1.5 py-0.5 text-xs rounded bg-gray-100 text-gray-600;
}

.badge-required {
  @apply px-1.5 py-0.5 text-xs rounded bg-red-100 text-red-700;
}

.badge-readonly {
  @apply px-1.5 py-0.5 text-xs rounded bg-gray-100 text-gray-600;
}

.badge-inactive {
  @apply px-1.5 py-0.5 text-xs rounded bg-gray-100 text-gray-500;
}

.field-row-actions {
  @apply flex items-center space-x-1;
}

.icon-action {
  @apply p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-gray-700 transition;
}

.icon-action-danger {
  @apply hover:bg-red-50 hover:text-red-600;
}

.empty-state {
  @apply flex flex-col items-center justify-center py-12 text-center;
}

/* 编辑抽屉 */
.edit-drawer-overlay {
  @apply fixed inset-0 bg-black bg-opacity-30 z-40 flex justify-end;
}

.edit-drawer {
  @apply bg-white w-full max-w-lg h-full shadow-xl flex flex-col;
}

.drawer-header {
  @apply flex items-center justify-between px-6 py-4 border-b border-gray-200 shrink-0;
}

.drawer-title {
  @apply text-lg font-semibold text-gray-900;
}

.drawer-subtitle {
  @apply text-sm text-gray-500 mt-0.5;
}

.close-btn {
  @apply p-2 rounded-lg hover:bg-gray-100 text-gray-500 transition;
}

.drawer-body {
  @apply px-6 py-4 flex-1 overflow-y-auto;
}

.drawer-footer {
  @apply flex items-center justify-end space-x-3 px-6 py-4 border-t border-gray-200 shrink-0;
}

.option-row {
  @apply bg-gray-50 rounded-lg p-2;
}

/* 表单 */
.form-group {
  @apply mb-4;
}

.form-label {
  @apply block text-sm font-medium text-gray-700 mb-1;
}

.required {
  @apply text-red-500;
}

.form-input {
  @apply w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500;
}

.checkbox-label {
  @apply flex items-center space-x-2 cursor-pointer;
}

.checkbox {
  @apply w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500;
}

.color-grid {
  @apply grid grid-cols-10 gap-2;
}

.color-btn {
  @apply w-6 h-6 rounded-full border-2 border-transparent hover:border-gray-400 transition;
}

.color-btn.active {
  @apply border-gray-900 ring-2 ring-gray-300;
}

.btn {
  @apply px-4 py-2 rounded-lg font-medium transition disabled:opacity-50 disabled:cursor-not-allowed;
}

.btn-primary {
  @apply bg-indigo-600 text-white hover:bg-indigo-700;
}

.btn-secondary {
  @apply bg-gray-100 text-gray-700 hover:bg-gray-200;
}

.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.25s ease;
}

.slide-fade-enter-from .edit-drawer,
.slide-fade-leave-to .edit-drawer {
  transform: translateX(100%);
}

.field-value-input {
  @apply mt-3 pt-3 border-t border-gray-100;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  opacity: 0;
}
</style>
