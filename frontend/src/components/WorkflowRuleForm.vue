<template>
  <div class="workflow-rule-form">
    <form @submit.prevent="handleSubmit" class="space-y-6">
      <!-- 基本信息 -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h3 class="text-sm font-medium text-gray-700 mb-4">基本信息</h3>

        <!-- 规则名称 -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            规则名称 *
          </label>
          <input
            v-model="form.name"
            type="text"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="例如：工作项完成时通知负责人"
          />
        </div>

        <!-- 描述 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            描述
          </label>
          <textarea
            v-model="form.description"
            rows="2"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="描述此规则的用途..."
          />
        </div>
      </div>

      <!-- 触发器 -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h3 class="text-sm font-medium text-gray-700 mb-4">
          触发器
          <span class="text-red-500">*</span>
        </h3>

        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            触发事件
          </label>
          <select
            v-model="form.trigger.type"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">选择触发事件</option>
            <option value="issue.created">工作项创建时</option>
            <option value="issue.updated">工作项更新时</option>
            <option value="issue.deleted">工作项删除时</option>
            <option value="issue.assigned">工作项分配时</option>
            <option value="issue.state_changed">状态变更时</option>
            <option value="issue.priority_changed">优先级变更时</option>
            <option value="issue.due_soon">截止日期临近时</option>
            <option value="issue.due_date_passed">截止日期过期时</option>
            <option value="cycle.started">周期开始时</option>
            <option value="cycle.ended">周期结束时</option>
            <option value="comment.added">添加评论时</option>
          </select>
        </div>

        <!-- 触发器附加参数 -->
        <div v-if="form.trigger.type === 'issue.due_soon'" class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            提前天数
          </label>
          <input
            v-model.number="form.trigger.days_before"
            type="number"
            min="1"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
      </div>

      <!-- 条件 -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-medium text-gray-700">条件</h3>
          <button
            type="button"
            @click="addCondition"
            class="text-sm text-indigo-600 hover:text-indigo-800 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>添加条件</span>
          </button>
        </div>

        <div v-if="form.conditions.length === 0" class="text-center py-4 text-gray-500 text-sm">
          暂无条件（触发时将立即执行动作）
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="(condition, index) in form.conditions"
            :key="index"
            class="flex items-center space-x-2"
          >
            <select
              v-model="condition.field"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">选择字段</option>
              <option value="state">状态</option>
              <option value="priority">优先级</option>
              <option value="assignee">负责人</option>
              <option value="labels">标签</option>
              <option value="cycle">周期</option>
              <option value="due_date">截止日期</option>
              <option value="start_date">开始日期</option>
            </select>

            <select
              v-model="condition.operator"
              class="w-32 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="equals">等于</option>
              <option value="not_equals">不等于</option>
              <option value="contains">包含</option>
              <option value="not_contains">不包含</option>
              <option value="in">在列表中</option>
              <option value="not_in">不在列表中</option>
              <option value="is_empty">为空</option>
              <option value="is_not_empty">不为空</option>
            </select>

            <input
              v-if="condition.operator !== 'is_empty' && condition.operator !== 'is_not_empty'"
              v-model="condition.value"
              type="text"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="值"
            />

            <button
              type="button"
              @click="removeCondition(index)"
              class="p-1 text-gray-400 hover:text-red-600"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- 动作 -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-medium text-gray-700">
            动作
            <span class="text-red-500">*</span>
          </h3>
          <button
            type="button"
            @click="addAction"
            class="text-sm text-indigo-600 hover:text-indigo-800 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>添加动作</span>
          </button>
        </div>

        <div v-if="form.actions.length === 0" class="text-center py-4 text-gray-500 text-sm">
          请至少添加一个动作
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="(action, index) in form.actions"
            :key="index"
            class="p-3 bg-gray-50 rounded-lg"
          >
            <div class="flex items-start space-x-2">
              <select
                v-model="action.type"
                class="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option value="">选择动作</option>
                <option value="issue.update">更新工作项</option>
                <option value="issue.assign">分配工作项</option>
                <option value="issue.add_label">添加标签</option>
                <option value="issue.remove_label">移除标签</option>
                <option value="issue.change_state">改变状态</option>
                <option value="issue.set_priority">设置优先级</option>
                <option value="notification.create">创建通知</option>
                <option value="email.send">发送邮件</option>
              </select>

              <button
                type="button"
                @click="removeAction(index)"
                class="p-1 text-gray-400 hover:text-red-600"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- 动作参数 -->
            <div v-if="action.type === 'issue.update'" class="mt-2 grid grid-cols-2 gap-2">
              <input
                v-model="action.field"
                type="text"
                class="px-3 py-2 border border-gray-300 rounded-md text-sm"
                placeholder="字段名"
              />
              <input
                v-model="action.value"
                type="text"
                class="px-3 py-2 border border-gray-300 rounded-md text-sm"
                placeholder="新值"
              />
            </div>

            <div v-if="action.type === 'issue.change_state'" class="mt-2">
              <select
                v-model="action.state_id"
                class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
              >
                <option value="">选择状态</option>
                <option v-for="state in states" :key="state.id" :value="state.id">
                  {{ state.name }}
                </option>
              </select>
            </div>

            <div v-if="action.type === 'notification.create' || action.type === 'email.send'" class="mt-2">
              <input
                v-model="action.message"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                placeholder="通知内容"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 提交按钮 -->
      <div class="flex justify-end space-x-3">
        <button
          type="button"
          @click="$emit('cancel')"
          class="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50"
        >
          取消
        </button>
        <button
          type="submit"
          :disabled="submitting || !isValid"
          class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
        >
          {{ submitting ? '保存中...' : (isEdit ? '保存' : '创建') }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { AutomationRule, AutomationRuleCreate, AutomationRuleUpdate, Trigger, Condition, Action } from '@/types/workflow'

// Props
const props = defineProps<{
  projectId: string
  workspaceId: string
  rule?: AutomationRule
  states?: Array<{ id: string; name: string }>
}>()

// Emits
const emit = defineEmits<{
  (e: 'submit', data: AutomationRuleCreate): void
  (e: 'cancel'): void
}>()

// State
const submitting = ref(false)

// Form data
const form = ref<{
  name: string
  description: string
  trigger: Trigger
  conditions: Condition[]
  actions: Action[]
}>({
  name: '',
  description: '',
  trigger: { type: '' },
  conditions: [],
  actions: []
})

// Check if edit mode
const isEdit = computed(() => !!props.rule)

// Validate form
const isValid = computed(() => {
  return form.value.name.trim() !== '' &&
    form.value.trigger.type !== '' &&
    form.value.actions.length > 0 &&
    form.value.actions.every(a => a.type !== '')
})

// Initialize form with existing data
if (props.rule) {
  form.value = {
    name: props.rule.name,
    description: props.rule.description || '',
    trigger: { ...props.rule.trigger },
    conditions: [...(props.rule.conditions || [])],
    actions: [...(props.rule.actions || [])]
  }
}

// Add condition
function addCondition() {
  form.value.conditions.push({
    field: '',
    operator: 'equals',
    value: undefined
  })
}

// Remove condition
function removeCondition(index: number) {
  form.value.conditions.splice(index, 1)
}

// Add action
function addAction() {
  form.value.actions.push({
    type: '',
    field: undefined,
    value: undefined,
    state_id: undefined,
    message: undefined
  })
}

// Remove action
function removeAction(index: number) {
  form.value.actions.splice(index, 1)
}

// Handle submit
async function handleSubmit() {
  if (!isValid.value) return

  submitting.value = true
  try {
    const data: AutomationRuleCreate = {
      name: form.value.name,
      description: form.value.description || undefined,
      trigger: form.value.trigger,
      conditions: form.value.conditions.length > 0 ? form.value.conditions : undefined,
      actions: form.value.actions,
      project_id: props.projectId
    }

    emit('submit', data)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.workflow-rule-form {
  @apply bg-gray-50;
}
</style>