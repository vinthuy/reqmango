<template>
  <div class="automation-rule-builder max-w-3xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h3 class="text-xl font-semibold text-gray-900">
        {{ isEdit ? t('settings.editAutomation') : t('settings.createAutomation') }}
      </h3>
      <button
        @click="$emit('cancel')"
        class="text-gray-500 hover:text-gray-700 transition-colors"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div class="space-y-6">
      <div class="bg-white rounded-xl border border-gray-200 p-5">
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('settings.name') }} <span class="text-red-500">*</span>
          </label>
          <input
            v-model="form.name"
            type="text"
            class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
            :placeholder="t('settings.automationNamePlaceholder')"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('settings.descriptionOptional') }}
          </label>
          <textarea
            v-model="form.description"
            rows="2"
            class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
            :placeholder="t('settings.automationDescriptionPlaceholder')"
          />
        </div>
      </div>

      <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
        <div class="px-5 py-4 bg-gradient-to-r from-indigo-50 to-purple-50 border-b border-gray-200">
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 rounded-lg bg-indigo-100 flex items-center justify-center">
              <span class="text-xl">⚡</span>
            </div>
            <div>
              <h4 class="font-semibold text-gray-900">{{ t('automationBuilder.when') }}</h4>
              <p class="text-sm text-gray-500">{{ t('automationBuilder.triggerDescription') }}</p>
            </div>
          </div>
        </div>
        <div class="p-5">
          <div v-if="!selectedTrigger" class="grid grid-cols-2 md:grid-cols-3 gap-3">
            <button
              v-for="trigger in triggerOptions"
              :key="trigger.value"
              @click="selectTrigger(trigger)"
              class="flex flex-col items-center p-4 border border-gray-200 rounded-lg hover:border-indigo-300 hover:bg-indigo-50 transition-all text-left group"
            >
              <span class="text-2xl mb-2">{{ trigger.icon }}</span>
              <span class="text-sm font-medium text-gray-700">{{ trigger.label }}</span>
            </button>
          </div>
          <div v-else class="relative">
            <div class="flex items-center justify-between p-4 bg-indigo-50 rounded-lg border border-indigo-200">
              <div class="flex items-center space-x-3">
                <span class="text-2xl">{{ selectedTrigger.icon }}</span>
                <span class="font-medium text-gray-900">{{ selectedTrigger.label }}</span>
              </div>
              <button @click="selectedTrigger = null" class="text-gray-400 hover:text-indigo-600">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div v-if="selectedTrigger.value === 'issue.due_soon'" class="mt-3">
              <label class="block text-sm font-medium text-gray-700 mb-1">
                {{ t('automationBuilder.daysBefore') }}
              </label>
              <input
                v-model.number="triggerParams.days_before"
                type="number"
                min="1"
                class="w-32 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
        <div class="px-5 py-4 bg-gradient-to-r from-amber-50 to-orange-50 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <div class="w-10 h-10 rounded-lg bg-amber-100 flex items-center justify-center">
                <span class="text-xl">📋</span>
              </div>
              <div>
                <h4 class="font-semibold text-gray-900">{{ t('automationBuilder.conditions') }}</h4>
                <p class="text-sm text-gray-500">{{ t('automationBuilder.conditionsOptional') }}</p>
              </div>
            </div>
            <button
              @click="addCondition"
              class="flex items-center space-x-1 px-3 py-1.5 bg-amber-100 text-amber-700 rounded-lg hover:bg-amber-200 transition-colors text-sm font-medium"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              <span>{{ t('automationBuilder.addCondition') }}</span>
            </button>
          </div>
        </div>
        <div class="p-5">
          <div v-if="form.conditions.length === 0" class="text-center py-8 text-gray-500">
            <span class="text-4xl block mb-2">🎯</span>
            <p class="text-sm">{{ t('automationBuilder.noConditions') }}</p>
            <p class="text-xs text-gray-400 mt-1">{{ t('automationBuilder.addConditionHint') }}</p>
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="(condition, index) in form.conditions"
              :key="index"
              class="flex items-center space-x-3 p-3 bg-gray-50 rounded-lg group"
            >
              <div class="flex items-center space-x-2 shrink-0">
                <span class="text-gray-400 text-sm font-medium">IF</span>
              </div>
              <select
                v-model="condition.field"
                class="flex-1 min-w-[120px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              >
                <option value="" disabled>{{ t('automationBuilder.selectField') }}</option>
                <option value="state">{{ t('automationBuilder.state') }}</option>
                <option value="priority">{{ t('automationBuilder.priority') }}</option>
                <option value="assignee">{{ t('automationBuilder.assignee') }}</option>
                <option value="labels">{{ t('automationBuilder.labels') }}</option>
                <option value="cycle">{{ t('automationBuilder.cycle') }}</option>
                <option value="due_date">{{ t('automationBuilder.dueDate') }}</option>
                <option value="start_date">{{ t('automationBuilder.startDate') }}</option>
              </select>
              <select
                v-model="condition.operator"
                class="min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              >
                <option value="equals">{{ t('automationBuilder.equals') }}</option>
                <option value="not_equals">{{ t('automationBuilder.notEquals') }}</option>
                <option value="contains">{{ t('automationBuilder.contains') }}</option>
                <option value="not_contains">{{ t('automationBuilder.notContains') }}</option>
                <option value="is_empty">{{ t('automationBuilder.isEmpty') }}</option>
                <option value="is_not_empty">{{ t('automationBuilder.isNotEmpty') }}</option>
              </select>
              <input
                v-if="condition.operator !== 'is_empty' && condition.operator !== 'is_not_empty'"
                v-model="condition.value"
                type="text"
                class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                :placeholder="t('automationBuilder.value')"
              />
              <button
                @click="removeCondition(index)"
                class="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-red-600 transition-all"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div v-if="form.conditions.length > 0" class="flex items-center justify-center text-gray-400 text-sm">
              <span>AND</span>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
        <div class="px-5 py-4 bg-gradient-to-r from-emerald-50 to-teal-50 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <div class="w-10 h-10 rounded-lg bg-emerald-100 flex items-center justify-center">
                <span class="text-xl">🎬</span>
              </div>
              <div>
                <h4 class="font-semibold text-gray-900">{{ t('automationBuilder.then') }}</h4>
                <p class="text-sm text-gray-500">{{ t('automationBuilder.actionsDescription') }}</p>
              </div>
            </div>
            <button
              @click="addAction"
              class="flex items-center space-x-1 px-3 py-1.5 bg-emerald-100 text-emerald-700 rounded-lg hover:bg-emerald-200 transition-colors text-sm font-medium"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              <span>{{ t('automationBuilder.addAction') }}</span>
            </button>
          </div>
        </div>
        <div class="p-5">
          <div v-if="form.actions.length === 0" class="text-center py-8 text-gray-500">
            <span class="text-4xl block mb-2">✨</span>
            <p class="text-sm">{{ t('automationBuilder.noActions') }}</p>
            <p class="text-xs text-gray-400 mt-1">{{ t('automationBuilder.addActionHint') }}</p>
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="(action, index) in form.actions"
              :key="index"
              class="p-4 bg-gray-50 rounded-lg border border-gray-200 group"
            >
              <div class="flex items-start space-x-3">
                <select
                  v-model="action.type"
                  class="flex-1 min-w-[140px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value="" disabled>{{ t('automationBuilder.selectAction') }}</option>
                  <option value="change_state">{{ t('automationBuilder.changeState') }}</option>
                  <option value="set_priority">{{ t('automationBuilder.setPriority') }}</option>
                  <option value="assign_to">{{ t('automationBuilder.assignTo') }}</option>
                  <option value="unassign">{{ t('automationBuilder.unassign') }}</option>
                  <option value="add_label">{{ t('automationBuilder.addLabel') }}</option>
                  <option value="remove_label">{{ t('automationBuilder.removeLabel') }}</option>
                  <option value="add_comment">{{ t('automationBuilder.addComment') }}</option>
                  <option value="set_field">{{ t('automationBuilder.setField') }}</option>
                </select>
                <button
                  @click="removeAction(index)"
                  class="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-red-600 transition-all"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div v-if="action.type === 'change_state'" class="mt-3">
                <select
                  v-model="action.value"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value="" disabled>{{ t('automationBuilder.selectState') }}</option>
                  <option v-for="state in states" :key="state.id" :value="state.name">
                    {{ state.name }}
                  </option>
                </select>
              </div>
              <div v-if="action.type === 'set_priority'" class="mt-3">
                <select
                  v-model="action.value"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value="" disabled>{{ t('automationBuilder.selectPriority') }}</option>
                  <option value="urgent">{{ t('automationBuilder.urgent') }}</option>
                  <option value="high">{{ t('automationBuilder.high') }}</option>
                  <option value="medium">{{ t('automationBuilder.medium') }}</option>
                  <option value="low">{{ t('automationBuilder.low') }}</option>
                </select>
              </div>
              <div v-if="action.type === 'assign_to'" class="mt-3">
                <select
                  v-model="action.value"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value="" disabled>{{ t('automationBuilder.selectAssignee') }}</option>
                  <option v-for="member in members" :key="member.user_id" :value="member.user_id">
                    {{ member.user?.display_name || member.display_name || `User #${member.user_id}` }}
                  </option>
                </select>
              </div>
              <div v-if="action.type === 'add_label' || action.type === 'remove_label'" class="mt-3">
                <input
                  v-model="action.value"
                  type="text"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  :placeholder="t('automationBuilder.labelName')"
                />
              </div>
              <div v-if="action.type === 'add_comment'" class="mt-3">
                <textarea
                  v-model="action.value"
                  rows="3"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
                  :placeholder="t('automationBuilder.commentText')"
                />
              </div>
              <div v-if="action.type === 'set_field'" class="mt-3 grid grid-cols-2 gap-3">
                <input
                  v-model="action.field"
                  type="text"
                  class="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  :placeholder="t('automationBuilder.fieldName')"
                />
                <input
                  v-model="action.value"
                  type="text"
                  class="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  :placeholder="t('automationBuilder.fieldValue')"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="flex justify-end space-x-3">
        <button
          @click="$emit('cancel')"
          class="px-5 py-2.5 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium"
        >
          {{ t('settings.cancel') }}
        </button>
        <button
          @click="handleSubmit"
          :disabled="!isValid || submitting"
          class="px-5 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
        >
          {{ submitting ? t('settings.saving') : (isEdit ? t('settings.update') : t('settings.create')) }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { AutomationRule, AutomationRuleCreate, Condition, Action } from '@/types/workflow'
import { TriggerTypeOptions, ConditionOperatorEnum } from '@/types/workflow'

const { t } = useI18n()

const props = defineProps<{
  projectId: number
  workspaceId: number
  rule?: AutomationRule
  states?: Array<{ id: number; name: string; color: string }>
  members?: Array<{ user_id: number; user?: { display_name: string }; display_name?: string }>
}>()

const emit = defineEmits<{
  (e: 'submit', data: AutomationRuleCreate): void
  (e: 'cancel'): void
}>()

const submitting = ref(false)
const selectedTrigger = ref<any>(null)
const triggerParams = ref({ days_before: 3 })

const triggerOptions = TriggerTypeOptions

const form = ref<{
  name: string
  description: string
  conditions: Condition[]
  actions: Action[]
}>({
  name: '',
  description: '',
  conditions: [],
  actions: []
})

const isEdit = computed(() => !!props.rule)

const isValid = computed(() => {
  return form.value.name.trim() !== '' &&
    selectedTrigger.value !== null &&
    form.value.actions.length > 0 &&
    form.value.actions.every(a => {
      if (!a.type) return false
      if (a.type === 'change_state' && !a.value) return false
      if (a.type === 'set_priority' && !a.value) return false
      if (a.type === 'assign_to' && !a.value) return false
      if (a.type === 'add_label' && !a.value) return false
      if (a.type === 'remove_label' && !a.value) return false
      if (a.type === 'add_comment' && !a.value) return false
      if (a.type === 'set_field' && (!a.field || !a.value)) return false
      return true
    })
})

watch(() => props.rule, (rule) => {
  if (!rule) return
  form.value.name = rule.name
  form.value.description = rule.description || ''
  try {
    const parsedTrigger = typeof rule.trigger_type === 'string' ? JSON.parse(rule.trigger_type) : rule.trigger_type
    if (parsedTrigger && typeof parsedTrigger.type === 'string') {
      selectedTrigger.value = triggerOptions.find(t => t.value === parsedTrigger.type) || null
      if (parsedTrigger.days_before) {
        triggerParams.value.days_before = parsedTrigger.days_before
      }
    }
  } catch {
    selectedTrigger.value = triggerOptions.find(t => t.value === rule.trigger_type) || null
  }
  try {
    form.value.conditions = JSON.parse(rule.conditions || '[]')
  } catch {
    form.value.conditions = []
  }
  try {
    form.value.actions = JSON.parse(rule.actions || '[]')
  } catch {
    form.value.actions = []
  }
}, { immediate: true })

function selectTrigger(trigger: any) {
  selectedTrigger.value = trigger
}

function addCondition() {
  form.value.conditions.push({
    field: '',
    operator: ConditionOperatorEnum.EQUALS,
    value: ''
  })
}

function removeCondition(index: number) {
  form.value.conditions.splice(index, 1)
}

function addAction() {
  form.value.actions.push({
    type: '',
    field: '',
    value: ''
  })
}

function removeAction(index: number) {
  form.value.actions.splice(index, 1)
}

async function handleSubmit() {
  if (!isValid.value) return
  submitting.value = true
  try {
    const triggerObj: any = { type: selectedTrigger.value.value }
    if (selectedTrigger.value.value === 'issue.due_soon' && triggerParams.value.days_before) {
      triggerObj.days_before = triggerParams.value.days_before
    }
    const data: AutomationRuleCreate = {
      name: form.value.name,
      description: form.value.description || undefined,
      trigger_type: JSON.stringify(triggerObj),
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
.automation-rule-builder {
  @apply py-4;
}
</style>