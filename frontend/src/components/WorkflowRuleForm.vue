<template>
  <div class="workflow-rule-form">
    <form @submit.prevent="handleSubmit" class="space-y-6">
      <!-- 基本信息 -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h3 class="text-sm font-medium text-gray-700 mb-4">{{ t('automationForm.basicInfo') }}</h3>

        <!-- 规则名称 -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('workflowRule.ruleName') }} *
          </label>
          <input
            v-model="form.name"
            type="text"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            :placeholder="t('workflowRule.ruleNamePlaceholder')"
          />
        </div>

        <!-- 描述 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('workflowRule.descriptionLabel') }}
          </label>
          <textarea
            v-model="form.description"
            rows="2"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            :placeholder="t('workflowRule.descriptionPlaceholder')"
          />
        </div>
      </div>

      <!-- 触发器 -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h3 class="text-sm font-medium text-gray-700 mb-4">
          {{ t('workflowRule.trigger') }}
          <span class="text-red-500">*</span>
        </h3>

        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('workflowRule.triggerEvent') }}
          </label>
          <select
            v-model="form.trigger.type"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">{{ t('workflowRule.selectTrigger') }}</option>
            <option value="issue.created">{{ t('workflowRule.trigger_issue_created') }}</option>
            <option value="issue.updated">{{ t('workflowRule.trigger_issue_updated') }}</option>
            <option value="issue.deleted">{{ t('workflowRule.trigger_issue_deleted') }}</option>
            <option value="issue.assigned">{{ t('workflowRule.trigger_issue_assigned') }}</option>
            <option value="issue.state_changed">{{ t('workflowRule.trigger_state_changed') }}</option>
            <option value="issue.priority_changed">{{ t('workflowRule.trigger_priority_changed') }}</option>
            <option value="issue.due_soon">{{ t('workflowRule.trigger_due_soon') }}</option>
            <option value="issue.due_date_passed">{{ t('workflowRule.trigger_due_date_passed') }}</option>
            <option value="cycle.started">{{ t('workflowRule.trigger_cycle_started') }}</option>
            <option value="cycle.ended">{{ t('workflowRule.trigger_cycle_ended') }}</option>
            <option value="comment.added">{{ t('workflowRule.trigger_comment_added') }}</option>
          </select>
        </div>

        <!-- 触发器附加参数 -->
        <div v-if="form.trigger.type === 'issue.due_soon'" class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('workflowRule.daysBefore') }}
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
          <h3 class="text-sm font-medium text-gray-700">{{ t('workflowRule.conditions') }}</h3>
          <button
            type="button"
            @click="addCondition"
            class="text-sm text-indigo-600 hover:text-indigo-800 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>{{ t('workflowRule.addCondition') }}</span>
          </button>
        </div>

        <div v-if="form.conditions.length === 0" class="text-center py-4 text-gray-500 text-sm">
          {{ t('workflowRule.noConditions') }}
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
              <option value="">{{ t('workflowRule.selectField') }}</option>
              <option value="state">{{ t('workflowRule.state') }}</option>
              <option value="priority">{{ t('workflowRule.priority') }}</option>
              <option value="assignee">{{ t('workflowRule.assignee') }}</option>
              <option value="labels">{{ t('workflowRule.labels') }}</option>
              <option value="cycle">{{ t('workflowRule.cycle') }}</option>
              <option value="due_date">{{ t('workflowRule.dueDate') }}</option>
              <option value="start_date">{{ t('workflowRule.startDate') }}</option>
            </select>

            <select
              v-model="condition.operator"
              class="w-32 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="equals">{{ t('workflowRule.equals') }}</option>
              <option value="not_equals">{{ t('workflowRule.notEquals') }}</option>
              <option value="contains">{{ t('workflowRule.contains') }}</option>
              <option value="not_contains">{{ t('workflowRule.notContains') }}</option>
              <option value="in">{{ t('workflowRule.inList') }}</option>
              <option value="not_in">{{ t('workflowRule.notInList') }}</option>
              <option value="is_empty">{{ t('workflowRule.isEmpty') }}</option>
              <option value="is_not_empty">{{ t('workflowRule.isNotEmpty') }}</option>
            </select>

            <input
              v-if="condition.operator !== 'is_empty' && condition.operator !== 'is_not_empty'"
              v-model="condition.value"
              type="text"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              :placeholder="t('workflowRule.value')"
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
            {{ t('workflowRule.actions_label') }}
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
            <span>{{ t('workflowRule.addAction') }}</span>
          </button>
        </div>

        <div v-if="form.actions.length === 0" class="text-center py-4 text-gray-500 text-sm">
          {{ t('workflowRule.noActions') }}
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
                <option value="">{{ t('automationForm.selectAction') }}</option>
                <option value="change_state">{{ t('workflowRule.changeState') }}</option>
                <option value="set_priority">{{ t('workflowRule.setPriority') }}</option>
                <option value="assign_to">{{ t('workflowRule.assignIssue') }}</option>
                <option value="unassign">{{ t('workflowRule.unassign') }}</option>
                <option value="add_label">{{ t('workflowRule.addLabel') }}</option>
                <option value="remove_label">{{ t('workflowRule.removeLabel') }}</option>
                <option value="add_comment">{{ t('workflowRule.createNotification') }}</option>
                <option value="set_field">{{ t('workflowRule.updateIssue') }}</option>
                <option value="dispatch_agent">{{ t('agent.dispatchAgent') }}</option>
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
            <div v-if="action.type === 'set_field'" class="mt-2 grid grid-cols-2 gap-2">
              <input
                v-model="action.field"
                type="text"
                class="px-3 py-2 border border-gray-300 rounded-md text-sm"
                :placeholder="t('workflowRule.fieldName')"
              />
              <input
                v-model="action.value"
                type="text"
                class="px-3 py-2 border border-gray-300 rounded-md text-sm"
                :placeholder="t('workflowRule.newValue')"
              />
            </div>

            <div v-if="action.type === 'change_state'" class="mt-2">
              <select
                v-model="action.value"
                class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
              >
                <option value="">{{ t('workflowRule.selectState') }}</option>
                <option v-for="state in states" :key="state.id" :value="state.id">
                  {{ state.name }}
                </option>
              </select>
            </div>

            <div v-if="action.type === 'add_comment'" class="mt-2">
              <input
                v-model="action.value"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                :placeholder="t('automationForm.commentPlaceholder')"
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
          {{ t('workflowRule.cancel') }}
        </button>
        <button
          type="submit"
          :disabled="submitting || !isValid"
          class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
        >
          {{ submitting ? t('workflowRule.saving') : (isEdit ? t('workflowRule.save') : t('workflowRule.create')) }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { AutomationRule, AutomationRuleCreate, Trigger, Condition, Action } from '@/types/workflow'
import { TriggerTypeEnum, ConditionOperatorEnum } from '@/types/workflow'

const { t } = useI18n()

// Props
const props = defineProps<{
  projectId: number
  workspaceId: number
  rule?: AutomationRule
  states?: Array<{ id: number; name: string }>
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
  trigger: { type: 'issue.created' as TriggerTypeEnum },
  conditions: [],
  actions: []
})

// Check if edit mode
const isEdit = computed(() => !!props.rule)

// Validate form
const isValid = computed(() => {
  return form.value.name.trim() !== '' &&
    form.value.trigger !== null &&
    form.value.actions.length > 0 &&
    form.value.actions.every(a => a.type !== undefined)
})

// Initialize form with existing data
if (props.rule) {
  let parsedTrigger: Trigger = { type: TriggerTypeEnum.ISSUE_CREATED }
  try { parsedTrigger = JSON.parse(props.rule.trigger_type || '{}') } catch { parsedTrigger = { type: (props.rule.trigger_type as TriggerTypeEnum) || TriggerTypeEnum.ISSUE_CREATED } }
  if (typeof parsedTrigger === 'string') parsedTrigger = { type: parsedTrigger as TriggerTypeEnum }

  let parsedConditions: Condition[] = []
  try { parsedConditions = JSON.parse(props.rule.conditions || '[]') } catch { /* use empty */ }

  let parsedActions: Action[] = []
  try { parsedActions = JSON.parse(props.rule.actions || '[]') } catch { /* use empty */ }

  form.value = {
    name: props.rule.name,
    description: props.rule.description || '',
    trigger: parsedTrigger,
    conditions: parsedConditions,
    actions: parsedActions
  }
}

// Add condition
function addCondition() {
  form.value.conditions.push({
    field: '',
    operator: ConditionOperatorEnum.EQUALS,
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
    field: '',
    value: ''
  })
}

// Remove action
function removeAction(index: number) {
  form.value.actions.splice(index, 1)
}

// Handle submit
async function handleSubmit() {
  if (!isValid.value || !form.value.trigger) return

  submitting.value = true
  try {
    const data: AutomationRuleCreate = {
      name: form.value.name,
      description: form.value.description || undefined,
      trigger_type: JSON.stringify(form.value.trigger),
      conditions: form.value.conditions.length > 0 ? JSON.stringify(form.value.conditions) : undefined,
      actions: JSON.stringify(form.value.actions),
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