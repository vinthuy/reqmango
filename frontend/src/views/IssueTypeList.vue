<template>
  <div class="issue-type-page">
    <!-- 头部 -->
    <div class="page-header">
      <div class="header-left">
        <button @click="goBack" class="back-btn">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div>
          <h1 class="page-title">工作项类型</h1>
          <p class="page-subtitle">管理项目中的工作项类型</p>
        </div>
      </div>
      <button @click="showCreateModal = true" class="create-btn">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        创建类型
      </button>
    </div>

    <!-- 类型列表 -->
    <div class="type-list" v-if="issueTypes.length > 0">
      <div
        v-for="type in issueTypes"
        :key="type.id"
        class="type-card"
        :class="{ 'is-default': type.is_default }"
      >
        <div class="type-header">
          <div class="type-icon" :style="{ backgroundColor: type.color }">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path v-if="type.icon === 'circle'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              <path v-else-if="type.icon === 'bug'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
          </div>
          <div class="type-info">
            <h3 class="type-name">{{ type.name }}</h3>
            <div class="type-badges">
              <span class="badge badge-level">L{{ type.level || 0 }}</span>
              <span v-if="type.is_default" class="badge badge-default">默认</span>
              <span v-if="!type.is_active" class="badge badge-inactive">已禁用</span>
              <span v-if="type.parent_type_id" class="badge badge-parent">父类型约束</span>
            </div>
          </div>
        </div>

        <div class="type-actions">
          <button @click="openEditModal(type)" class="action-btn" title="编辑">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
          </button>
          <button @click="openFieldsModal(type)" class="action-btn" title="管理字段">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
            </svg>
          </button>
          <button
            v-if="!type.is_default"
            @click="toggleActive(type)"
            class="action-btn"
            :title="type.is_active ? '禁用' : '启用'"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path v-if="type.is_active" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
              <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>
          <button
            v-if="!type.is_default"
            @click="confirmDelete(type)"
            class="action-btn action-btn-danger"
            title="删除"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <svg class="w-16 h-16 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <h3 class="text-lg font-medium text-gray-900">暂无工作项类型</h3>
      <p class="text-gray-500">点击上方按钮创建第一个工作项类型</p>
    </div>

    <!-- 创建/编辑模态框 -->
    <div v-if="showCreateModal || showEditModal" class="modal-overlay" @click.self="closeModals">
      <div class="modal">
        <div class="modal-header">
          <h2 class="modal-title">{{ showEditModal ? '编辑类型' : '创建类型' }}</h2>
          <button @click="closeModals" class="close-btn">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">名称 <span class="required">*</span></label>
            <input
              v-model="formData.name"
              type="text"
              class="form-input"
              placeholder="输入类型名称"
            />
          </div>

          <div class="form-group">
            <label class="form-label">图标</label>
            <div class="icon-grid">
              <button
                v-for="icon in ISSUE_TYPE_ICONS"
                :key="icon"
                class="icon-btn"
                :class="{ active: formData.icon === icon }"
                @click="formData.icon = icon"
                :title="getIconName(icon)"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </button>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">颜色</label>
            <div class="color-grid">
              <button
                v-for="color in ISSUE_TYPE_COLORS"
                :key="color"
                class="color-btn"
                :class="{ active: formData.color === color }"
                :style="{ backgroundColor: color }"
                @click="formData.color = color"
              />
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">描述</label>
            <input v-model="formData.description" type="text" class="form-input" placeholder="类型描述" />
          </div>

          <div class="form-group">
            <label class="form-label">层级 (Level)</label>
            <select v-model="formData.level" class="form-input">
              <option :value="0">L0 - 根层级 (Epic级)</option>
              <option :value="1">L1 - 第1层子级</option>
              <option :value="2">L2 - 第2层子级</option>
              <option :value="3">L3 - 第3层子级</option>
              <option :value="4">L4 - 第4层子级</option>
              <option :value="5">L5 - 第5层子级</option>
            </select>
          </div>

          <div class="form-group" v-if="(formData.level || 0) > 0">
            <label class="form-label">父类型 (Parent Type)</label>
            <select v-model="formData.parent_type_id" class="form-input">
              <option :value="undefined">无约束</option>
              <option v-for="t in parentTypeOptions" :key="t.id" :value="t.id">
                {{ t.name }} (L{{ t.level }})
              </option>
            </select>
            <p class="text-xs text-gray-500 mt-1">限定此类型只能挂在选定的父类型下</p>
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="formData.is_default" type="checkbox" class="checkbox" />
              <span>设为默认类型</span>
            </label>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="closeModals" class="btn btn-secondary">取消</button>
          <button @click="submitForm" class="btn btn-primary">
            {{ showEditModal ? '保存' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 字段管理模态框 -->
    <div v-if="showFieldsModal" class="modal-overlay" @click.self="showFieldsModal = false">
      <div class="modal modal-lg">
        <div class="modal-header">
          <h2 class="modal-title">管理字段 - {{ selectedType?.name }}</h2>
          <button @click="showFieldsModal = false" class="close-btn">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="modal-body">
          <div v-if="availableFields.length > 0" class="available-fields">
            <h4 class="section-title">可添加的字段</h4>
            <div class="field-list">
              <div
                v-for="field in availableFields"
                :key="field.id"
                class="field-item"
              >
                <div class="field-info">
                  <span class="field-name">{{ field.name }}</span>
                  <span class="field-type">{{ getFieldTypeName(field.field_type) }}</span>
                </div>
                <button @click="addField(field)" class="add-btn">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                  </svg>
                  添加
                </button>
              </div>
            </div>
          </div>

          <div v-if="typeFields.length > 0" class="linked-fields">
            <h4 class="section-title">已关联的字段</h4>
            <div class="field-list">
              <div
                v-for="tf in typeFields"
                :key="tf.field_id"
                class="field-item"
              >
                <div class="field-info">
                  <span class="field-name">{{ tf.name }}</span>
                  <span class="field-type">{{ getFieldTypeName(tf.field_type as any) }}</span>
                </div>
                <label class="required-toggle">
                  <input
                    type="checkbox"
                    :checked="tf.is_required"
                    @change="toggleFieldRequired(tf)"
                  />
                  <span>必填</span>
                </label>
                <button @click="removeField(tf.field_id)" class="remove-btn">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <div v-if="availableFields.length === 0 && typeFields.length === 0" class="empty-fields">
            <p>暂无可用的自定义字段</p>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="showFieldsModal = false" class="btn btn-secondary">完成</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { IssueType, IssueTypeCreate, IssueTypeUpdate, IssueTypeField } from '@/types/issue-type'
import { ISSUE_TYPE_ICONS, ISSUE_TYPE_COLORS, getIconName } from '@/types/issue-type'
import type { CustomField } from '@/types/custom-field'
import { getFieldTypeName } from '@/types/custom-field'
import { useConfirm } from '@/composables/useConfirm'
import * as issueTypeApi from '@/api/issue-type'
import * as customFieldApi from '@/api/custom-field'

const route = useRoute()
const router = useRouter()
const { confirm } = useConfirm()

const projectId = computed(() => parseInt(route.params.projectId as string, 10))
const workspaceId = computed(() => parseInt(route.params.workspaceId as string, 10))

// 状态
const issueTypes = ref<IssueType[]>([])
const customFields = ref<CustomField[]>([])
const loading = ref(false)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showFieldsModal = ref(false)
const selectedType = ref<IssueType | null>(null)
const typeFields = ref<(IssueTypeField & { name?: string; field_type?: string })[]>([])

// 表单数据
const formData = ref<IssueTypeCreate & IssueTypeUpdate>({
  name: '',
  color: ISSUE_TYPE_COLORS[0],
  icon: ISSUE_TYPE_ICONS[0],
  description: '',
  level: 0,
  parent_type_id: undefined,
  is_default: false,
  sequence: 1
})

// 父类型选项: 只显示比当前层级低一级的类型
const parentTypeOptions = computed(() => {
  const targetLevel = (formData.value.level || 0) - 1
  if (targetLevel < 0) return []
  return issueTypes.value.filter(t => (t.level || 0) === targetLevel)
})

// 可用字段（未关联的字段）
const availableFields = computed(() => {
  const linkedFieldIds = typeFields.value.map(tf => tf.field_id)
  return customFields.value.filter(f => !linkedFieldIds.includes(f.id) && f.is_active)
})

// 加载数据
async function loadData() {
  loading.value = true
  try {
    const [typesRes, fieldsRes] = await Promise.all([
      issueTypeApi.getIssueTypes(workspaceId.value, projectId.value),
      customFieldApi.listCustomFields(workspaceId.value)
    ])
    issueTypes.value = typesRes
    customFields.value = fieldsRes
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    loading.value = false
  }
}

// 打开编辑模态框
function openEditModal(type: IssueType) {
  formData.value = {
    name: type.name,
    color: type.color,
    icon: type.icon,
    description: type.description || '',
    level: type.level || 0,
    parent_type_id: type.parent_type_id || undefined,
    is_default: type.is_default,
    sequence: type.sequence
  }
  selectedType.value = type
  showEditModal.value = true
}

// 打开字段管理模态框
async function openFieldsModal(type: IssueType) {
  selectedType.value = type
  try {
    const fields = await issueTypeApi.getIssueTypeFields(type.id)
    typeFields.value = fields
    showFieldsModal.value = true
  } catch (error) {
    console.error('Failed to load type fields:', error)
  }
}

// 关闭所有模态框
function closeModals() {
  showCreateModal.value = false
  showEditModal.value = false
  selectedType.value = null
}

// 提交表单
async function submitForm() {
  if (!formData.value.name) return

  try {
    if (showEditModal.value && selectedType.value) {
      await issueTypeApi.updateIssueType(selectedType.value.id, formData.value)
    } else {
      await issueTypeApi.createIssueType(workspaceId.value, formData.value)
    }
    closeModals()
    await loadData()
  } catch (error) {
    console.error('Failed to submit form:', error)
  }
}

// 切换启用状态
async function toggleActive(type: IssueType) {
  try {
    await issueTypeApi.disableIssueType(type.id, !type.is_active)
    await loadData()
  } catch (error) {
    console.error('Failed to toggle active:', error)
  }
}

// 确认删除
async function confirmDelete(type: IssueType) {
  if (await confirm(`确定要删除类型 "${type.name}" 吗？此操作不可撤销。`)) {
    deleteType(type)
  }
}

// 删除类型
async function deleteType(type: IssueType) {
  try {
    await issueTypeApi.deleteIssueType(type.id)
    await loadData()
  } catch (error) {
    console.error('Failed to delete type:', error)
  }
}

// 添加字段
async function addField(field: CustomField) {
  if (!selectedType.value) return
  try {
    await issueTypeApi.addFieldToIssueType(selectedType.value.id, {
      field_id: field.id,
      is_required: false,
      sequence: typeFields.value.length + 1
    })
    await openFieldsModal(selectedType.value)
  } catch (error) {
    console.error('Failed to add field:', error)
  }
}

// 移除字段
async function removeField(fieldId: number) {
  if (!selectedType.value) return
  try {
    await issueTypeApi.removeFieldFromIssueType(selectedType.value.id, fieldId)
    await openFieldsModal(selectedType.value)
  } catch (error) {
    console.error('Failed to remove field:', error)
  }
}

// 切换字段必填状态
async function toggleFieldRequired(_tf: IssueTypeField & { name?: string; field_type?: string }) {
  // 这里需要调用更新 API，暂时先刷新列表
  if (selectedType.value) {
    await openFieldsModal(selectedType.value)
  }
}

// 返回
function goBack() {
  router.back()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.issue-type-page {
  @apply min-h-screen bg-gray-50 p-6;
}

.page-header {
  @apply flex items-center justify-between mb-6;
}

.header-left {
  @apply flex items-center space-x-4;
}

.back-btn {
  @apply p-2 rounded-lg hover:bg-gray-200 transition;
}

.page-title {
  @apply text-2xl font-bold text-gray-900;
}

.page-subtitle {
  @apply text-sm text-gray-500;
}

.create-btn {
  @apply flex items-center space-x-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition;
}

.type-list {
  @apply grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4;
}

.type-card {
  @apply bg-white rounded-lg border border-gray-200 p-4 hover:border-indigo-300 transition;
}

.type-card.is-default {
  @apply border-indigo-200 bg-indigo-50;
}

.type-header {
  @apply flex items-start space-x-3;
}

.type-icon {
  @apply w-10 h-10 rounded-lg flex items-center justify-center text-white;
}

.type-info {
  @apply flex-1;
}

.type-name {
  @apply font-semibold text-gray-900;
}

.type-badges {
  @apply flex items-center space-x-2 mt-1;
}

.badge {
  @apply px-2 py-0.5 text-xs rounded-full;
}

.badge-default {
  @apply bg-indigo-100 text-indigo-700;
}

.badge-inactive {
  @apply bg-gray-100 text-gray-600;
}

.badge-level {
  @apply bg-blue-100 text-blue-700 font-mono;
}

.badge-parent {
  @apply bg-yellow-100 text-yellow-700;
}

.type-actions {
  @apply flex items-center space-x-2 mt-4 pt-4 border-t border-gray-100;
}

.action-btn {
  @apply p-2 rounded hover:bg-gray-100 text-gray-600 hover:text-gray-900 transition;
}

.action-btn-danger {
  @apply hover:bg-red-50 hover:text-red-600;
}

.empty-state {
  @apply flex flex-col items-center justify-center py-16 text-center;
}

.modal-overlay {
  @apply fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50;
}

.modal {
  @apply bg-white rounded-xl shadow-xl w-full max-w-md;
}

.modal-lg {
  @apply max-w-2xl;
}

.modal-header {
  @apply flex items-center justify-between px-6 py-4 border-b;
}

.modal-title {
  @apply text-lg font-semibold text-gray-900;
}

.close-btn {
  @apply p-2 rounded-lg hover:bg-gray-100 transition;
}

.modal-body {
  @apply px-6 py-4 max-h-96 overflow-y-auto;
}

.modal-footer {
  @apply flex items-center justify-end space-x-3 px-6 py-4 border-t;
}

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

.icon-grid {
  @apply grid grid-cols-7 gap-2;
}

.icon-btn {
  @apply p-2 rounded-lg border border-gray-200 hover:border-indigo-300 hover:bg-indigo-50 transition;
}

.icon-btn.active {
  @apply border-indigo-500 bg-indigo-100;
}

.color-grid {
  @apply flex flex-wrap gap-2;
}

.color-btn {
  @apply w-8 h-8 rounded-full border-2 border-transparent hover:border-gray-400 transition;
}

.color-btn.active {
  @apply border-gray-900;
}

.checkbox-label {
  @apply flex items-center space-x-2 cursor-pointer;
}

.checkbox {
  @apply w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500;
}

.btn {
  @apply px-4 py-2 rounded-lg font-medium transition;
}

.btn-primary {
  @apply bg-indigo-600 text-white hover:bg-indigo-700;
}

.btn-secondary {
  @apply bg-gray-100 text-gray-700 hover:bg-gray-200;
}

.section-title {
  @apply text-sm font-medium text-gray-700 mb-3;
}

.field-list {
  @apply space-y-2;
}

.field-item {
  @apply flex items-center justify-between p-3 bg-gray-50 rounded-lg;
}

.field-info {
  @apply flex items-center space-x-3;
}

.field-name {
  @apply font-medium text-gray-900;
}

.field-type {
  @apply text-xs text-gray-500 bg-gray-200 px-2 py-0.5 rounded;
}

.add-btn {
  @apply flex items-center space-x-1 px-3 py-1 text-sm text-indigo-600 hover:bg-indigo-50 rounded-lg transition;
}

.remove-btn {
  @apply p-1.5 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded transition;
}

.required-toggle {
  @apply flex items-center space-x-1 text-sm text-gray-600;
}

.empty-fields {
  @apply text-center py-8 text-gray-500;
}
</style>
