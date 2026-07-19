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
                @change="onConditionFieldChange(index)"
                class="flex-1 min-w-[120px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              >
                <option value="" disabled>{{ t('automationBuilder.selectField') }}</option>
                <optgroup :label="t('automationBuilder.systemFields')">
                  <option value="state">{{ t('automationBuilder.state') }}</option>
                  <option value="priority">{{ t('automationBuilder.priority') }}</option>
                  <option value="assignee">{{ t('automationBuilder.assignee') }}</option>
                  <option value="labels">{{ t('automationBuilder.labels') }}</option>
                  <option value="cycle">{{ t('automationBuilder.cycle') }}</option>
                  <option value="due_date">{{ t('automationBuilder.dueDate') }}</option>
                  <option value="start_date">{{ t('automationBuilder.startDate') }}</option>
                  <option value="title">{{ t('automationBuilder.title') }}</option>
                  <option value="description">{{ t('automationBuilder.description') }}</option>
                </optgroup>
                <optgroup :label="t('automationBuilder.customFields')">
                  <option v-for="field in customFields" :key="field.id" :value="'custom_' + field.id">
                    {{ field.name }}
                  </option>
                </optgroup>
              </select>
              <select
                v-model="condition.operator"
                class="min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              >
                <option v-for="op in getOperatorsForField(condition.field)" :key="op.value" :value="op.value">
                  {{ op.label }}
                </option>
              </select>
              <template v-if="condition.operator !== 'is_empty' && condition.operator !== 'is_not_empty'">
                <select
                  v-if="isDropdownField(condition.field)"
                  v-model="condition.value"
                  class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value="" disabled>{{ t('automationBuilder.selectValue') }}</option>
                  <option v-for="option in getDropdownOptions(condition.field)" :key="option.id" :value="option.value">
                    {{ option.value }}
                  </option>
                </select>
                <select
                  v-else-if="condition.field === 'priority'"
                  v-model="condition.value"
                  class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value="" disabled>{{ t('automationBuilder.selectValue') }}</option>
                  <option value="urgent">{{ t('automationBuilder.urgent') }}</option>
                  <option value="high">{{ t('automationBuilder.high') }}</option>
                  <option value="medium">{{ t('automationBuilder.medium') }}</option>
                  <option value="low">{{ t('automationBuilder.low') }}</option>
                </select>
                <input
                  v-else-if="isNumberField(condition.field)"
                  v-model.number="condition.value"
                  type="number"
                  class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  :placeholder="t('automationBuilder.numberValue')"
                />
                <input
                  v-else-if="isDateField(condition.field)"
                  v-model="condition.value"
                  type="date"
                  class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                />
                <input
                  v-else-if="isBooleanField(condition.field)"
                  v-model="condition.value"
                  type="checkbox"
                  class="w-5 h-5 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                />
                <input
                  v-else
                  v-model="condition.value"
                  type="text"
                  class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  :placeholder="t('automationBuilder.value')"
                />
              </template>
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
              class="p-3 bg-gray-50 rounded-lg group"
            >
              <div class="flex items-center space-x-3">
                <div class="flex items-center space-x-2 shrink-0">
                  <span class="text-gray-400 text-sm font-medium">DO</span>
                </div>
                <select
                  v-model="action.type"
                  class="flex-1 min-w-[120px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value="" disabled>{{ t('automationBuilder.selectAction') }}</option>
                  <option value="change_state">{{ t('automationBuilder.changeState') }}</option>
                  <option value="set_priority">{{ t('automationBuilder.setPriority') }}</option>
                  <option value="assign_to">{{ t('automationBuilder.assignTo') }}</option>
                  <option value="unassign">{{ t('automationBuilder.unassign') }}</option>
                  <option value="add_comment">{{ t('automationBuilder.addComment') }}</option>
                  <option value="set_field">{{ t('automationBuilder.setField') }}</option>
                  <option value="call_webhook">{{ t('automationBuilder.callWebhook') }}</option>
                </select>
                <template v-if="action.type === 'change_state' || action.type === 'set_priority' || action.type === 'assign_to'">
                  <select
                    v-if="action.type === 'change_state'"
                    v-model="action.value"
                    class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  >
                    <option value="" disabled>{{ t('automationBuilder.selectState') }}</option>
                    <option v-for="state in states" :key="state.id" :value="state.name">
                      {{ state.name }}
                    </option>
                  </select>
                  <select
                    v-else-if="action.type === 'set_priority'"
                    v-model="action.value"
                    class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  >
                    <option value="" disabled>{{ t('automationBuilder.selectPriority') }}</option>
                    <option value="urgent">{{ t('automationBuilder.urgent') }}</option>
                    <option value="high">{{ t('automationBuilder.high') }}</option>
                    <option value="medium">{{ t('automationBuilder.medium') }}</option>
                    <option value="low">{{ t('automationBuilder.low') }}</option>
                  </select>
                  <select
                    v-else-if="action.type === 'assign_to'"
                    v-model="action.value"
                    class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  >
                    <option value="" disabled>{{ t('automationBuilder.selectAssignee') }}</option>
                    <option v-for="member in members" :key="member.user_id" :value="member.user_id">
                      {{ member.user?.display_name || member.display_name || `User #${member.user_id}` }}
                    </option>
                  </select>
                </template>
                <button
                  @click="removeAction(index)"
                  class="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-red-600 transition-all shrink-0"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div v-if="action.type === 'add_comment'" class="mt-3">
                <textarea
                  v-model="action.value"
                  rows="3"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
                  :placeholder="t('automationBuilder.commentText')"
                />
              </div>
              <div v-if="action.type === 'set_field'" class="mt-3 space-y-2">
                <select
                  v-model="action.field"
                  @change="onSetFieldChange(index)"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value="" disabled>{{ t('automationBuilder.selectField') }}</option>
                  <optgroup :label="t('automationBuilder.systemFields')">
                    <option value="state_id">{{ t('automationBuilder.state') }}</option>
                    <option value="priority">{{ t('automationBuilder.priority') }}</option>
                    <option value="target_date">{{ t('automationBuilder.dueDate') }}</option>
                    <option value="start_date">{{ t('automationBuilder.startDate') }}</option>
                  </optgroup>
                  <optgroup :label="t('automationBuilder.customFields')">
                    <option v-for="field in customFields" :key="field.id" :value="'custom_' + field.id">
                      {{ field.name }}
                    </option>
                  </optgroup>
                </select>
                <!-- Dynamic value input based on field type -->
                <template v-if="action.field">
                  <select
                    v-if="action.field === 'state_id'"
                    v-model="action.value"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  >
                    <option value="" disabled>{{ t('automationBuilder.selectState') }}</option>
                    <option v-for="state in states" :key="state.id" :value="state.id">
                      {{ state.name }}
                    </option>
                  </select>
                  <select
                    v-else-if="action.field === 'priority'"
                    v-model="action.value"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  >
                    <option value="" disabled>{{ t('automationBuilder.selectPriority') }}</option>
                    <option value="urgent">{{ t('automationBuilder.urgent') }}</option>
                    <option value="high">{{ t('automationBuilder.high') }}</option>
                    <option value="medium">{{ t('automationBuilder.medium') }}</option>
                    <option value="low">{{ t('automationBuilder.low') }}</option>
                  </select>
                  <input
                    v-else-if="action.field === 'target_date' || action.field === 'start_date'"
                    v-model="action.value"
                    type="date"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  />
                  <!-- Custom field: dropdown -->
                  <select
                    v-else-if="getCustomFieldType(action.field) === 'dropdown'"
                    v-model="action.value"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  >
                    <option value="" disabled>{{ t('automationBuilder.selectValue') }}</option>
                    <option v-for="opt in getCustomFieldOptions(action.field)" :key="opt.id" :value="opt.value">
                      {{ opt.value }}
                    </option>
                  </select>
                  <!-- Custom field: number -->
                  <input
                    v-else-if="getCustomFieldType(action.field) === 'number'"
                    v-model.number="action.value"
                    type="number"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                    :placeholder="t('automationBuilder.numberValue')"
                  />
                  <!-- Custom field: boolean -->
                  <select
                    v-else-if="getCustomFieldType(action.field) === 'boolean'"
                    v-model="action.value"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  >
                    <option value="" disabled>{{ t('automationBuilder.selectValue') }}</option>
                    <option value="true">True</option>
                    <option value="false">False</option>
                  </select>
                  <!-- Custom field: date -->
                  <input
                    v-else-if="getCustomFieldType(action.field) === 'date'"
                    v-model="action.value"
                    type="date"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  />
                  <!-- Custom field: member -->
                  <select
                    v-else-if="getCustomFieldType(action.field) === 'member'"
                    v-model="action.value"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  >
                    <option value="" disabled>{{ t('automationBuilder.selectMember') }}</option>
                    <option v-for="member in members" :key="member.user_id" :value="member.user_id">
                      {{ member.user?.display_name || member.display_name || `User #${member.user_id}` }}
                    </option>
                  </select>
                  <!-- Custom field: text (textarea for better UX) -->
                  <textarea
                    v-else-if="getCustomFieldType(action.field) === 'text'"
                    v-model="action.value"
                    rows="3"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
                    :placeholder="t('automationBuilder.fieldValue')"
                  />
                  <!-- Custom field: url -->
                  <input
                    v-else-if="getCustomFieldType(action.field) === 'url'"
                    v-model="action.value"
                    type="url"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                    placeholder="https://..."
                  />
                  <!-- Default: text input -->
                  <input
                    v-else
                    v-model="action.value"
                    type="text"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                    :placeholder="t('automationBuilder.fieldValue')"
                  />
                </template>
              </div>
              <div v-if="action.type === 'call_webhook'" class="mt-3 space-y-2">
                <input
                  v-model="action.field"
                  type="url"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  :placeholder="t('automationBuilder.webhookUrl')"
                />
                <div class="grid grid-cols-2 gap-2">
                  <select
                    v-model="webhookMethodCache[index]"
                    @change="updateWebhookValue(index)"
                    class="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  >
                    <option value="POST">POST</option>
                    <option value="GET">GET</option>
                    <option value="PUT">PUT</option>
                    <option value="PATCH">PATCH</option>
                    <option value="DELETE">DELETE</option>
                  </select>
                  <input
                    v-model="webhookHeadersCache[index]"
                    type="text"
                    class="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                    :placeholder="t('automationBuilder.webhookHeaders')"
                    @change="updateWebhookValue(index)"
                  />
                </div>
                <textarea
                  v-model="webhookBodyCache[index]"
                  rows="3"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none font-mono"
                  :placeholder="t('automationBuilder.webhookBody')"
                  @change="updateWebhookValue(index)"
                />
                <p class="text-xs text-gray-400">{{ t('automationBuilder.webhookVarsHint') }}</p>
              </div>
            </div>
            <div v-if="form.actions.length > 0" class="flex items-center justify-center text-gray-400 text-sm">
              <span>AND</span>
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
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { AutomationRule, AutomationRuleCreate, Condition, Action } from '@/types/workflow'
import { TriggerTypeOptions, ConditionOperatorEnum } from '@/types/workflow'
import type { CustomField } from '@/types/custom-field'
import { customFieldApi } from '@/api/custom-field'

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
const customFields = ref<CustomField[]>([])

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
      if (a.type === 'add_comment' && !a.value) return false
      if (a.type === 'set_field' && (!a.field || a.value === '' || a.value == null)) return false
      if (a.type === 'call_webhook' && !a.field) return false
      return true
    })
})

async function loadCustomFields() {
  try {
    const fields = await customFieldApi.listCustomFields(props.workspaceId, props.projectId)
    customFields.value = fields.filter(f => f.is_active)
  } catch (e) {
    console.error('Failed to load custom fields:', e)
  }
}

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

function onConditionFieldChange(index: number) {
  const condition = form.value.conditions[index]
  if (!condition.field) return
  condition.operator = getDefaultOperator(condition.field)
}

function getDefaultOperator(field: string): ConditionOperatorEnum {
  if (isNumberField(field)) return ConditionOperatorEnum.EQUALS
  if (isDateField(field)) return ConditionOperatorEnum.EQUALS
  if (isBooleanField(field)) return ConditionOperatorEnum.EQUALS
  return ConditionOperatorEnum.EQUALS
}

function getOperatorsForField(field: string) {
  const baseOperators = [
    { value: 'equals', label: t('automationBuilder.equals') },
    { value: 'not_equals', label: t('automationBuilder.notEquals') },
    { value: 'is_empty', label: t('automationBuilder.isEmpty') },
    { value: 'is_not_empty', label: t('automationBuilder.isNotEmpty') },
  ]
  
  if (isTextField(field)) {
    return [
      ...baseOperators,
      { value: 'contains', label: t('automationBuilder.contains') },
      { value: 'not_contains', label: t('automationBuilder.notContains') },
    ]
  }
  
  if (isNumberField(field)) {
    return [
      ...baseOperators,
      { value: 'greater_than', label: t('automationBuilder.greaterThan') },
      { value: 'less_than', label: t('automationBuilder.lessThan') },
    ]
  }
  
  if (isDateField(field)) {
    return [
      ...baseOperators,
      { value: 'greater_than', label: t('automationBuilder.greaterThan') },
      { value: 'less_than', label: t('automationBuilder.lessThan') },
    ]
  }
  
  return baseOperators
}

function isTextField(field: string): boolean {
  if (['title', 'description', 'labels'].includes(field)) return true
  if (field.startsWith('custom_')) {
    const fieldId = parseInt(field.replace('custom_', ''))
    const customField = customFields.value.find(f => f.id === fieldId)
    return customField?.field_type === 'text' || customField?.field_type === 'url'
  }
  return false
}

function isNumberField(field: string): boolean {
  if (field.startsWith('custom_')) {
    const fieldId = parseInt(field.replace('custom_', ''))
    const customField = customFields.value.find(f => f.id === fieldId)
    return customField?.field_type === 'number'
  }
  return false
}

function isDateField(field: string): boolean {
  if (['due_date', 'start_date'].includes(field)) return true
  if (field.startsWith('custom_')) {
    const fieldId = parseInt(field.replace('custom_', ''))
    const customField = customFields.value.find(f => f.id === fieldId)
    return customField?.field_type === 'date'
  }
  return false
}

function isBooleanField(field: string): boolean {
  if (field.startsWith('custom_')) {
    const fieldId = parseInt(field.replace('custom_', ''))
    const customField = customFields.value.find(f => f.id === fieldId)
    return customField?.field_type === 'boolean'
  }
  return false
}

function isDropdownField(field: string): boolean {
  if (field === 'state') return true
  if (field.startsWith('custom_')) {
    const fieldId = parseInt(field.replace('custom_', ''))
    const customField = customFields.value.find(f => f.id === fieldId)
    return customField?.field_type === 'dropdown'
  }
  return false
}

function getDropdownOptions(field: string): Array<{ id: number; value: string }> {
  if (field === 'state') {
    return props.states?.map(s => ({ id: s.id, value: s.name })) || []
  }
  if (field.startsWith('custom_')) {
    const fieldId = parseInt(field.replace('custom_', ''))
    const customField = customFields.value.find(f => f.id === fieldId)
    return (customField as any)?.options?.map((o: any) => ({ id: o.id, value: o.value })) || []
  }
  return []
}

// ====== set_field helpers ======

function getCustomFieldType(field: string): string {
  if (field.startsWith('custom_')) {
    const fieldId = parseInt(field.replace('custom_', ''))
    const customField = customFields.value.find(f => f.id === fieldId)
    return customField?.field_type || 'text'
  }
  return ''
}

function getCustomFieldOptions(field: string): Array<{ id: number; value: string }> {
  if (field.startsWith('custom_')) {
    const fieldId = parseInt(field.replace('custom_', ''))
    const customField = customFields.value.find(f => f.id === fieldId)
    return (customField as any)?.options?.map((o: any) => ({ id: o.id, value: o.value })) || []
  }
  return []
}

function onSetFieldChange(index: number) {
  const action = form.value.actions[index]
  if (!action) return
  action.field = action.field || ''
  action.value = ''
}

// ====== webhook action helpers ======

type WebhookCache = {
  method: string
  headers: string
  body: string
}

const webhookMethodCache = ref<Record<number, string>>({})
const webhookHeadersCache = ref<Record<number, string>>({})
const webhookBodyCache = ref<Record<number, string>>({})

function initWebhookCache(index: number) {
  const action = form.value.actions[index]
  if (!action) return
  if (action.type === 'call_webhook') {
    let cached: WebhookCache = { method: 'POST', headers: '', body: '' }
    if (typeof action.value === 'object' && action.value !== null) {
      cached = {
        method: (action.value as any).method || 'POST',
        headers: (action.value as any).headers
          ? Object.entries((action.value as any).headers).map(([k, v]) => `${k}: ${v}`).join('\n')
          : '',
        body: (action.value as any).body || '',
      }
    }
    webhookMethodCache.value[index] = cached.method
    webhookHeadersCache.value[index] = cached.headers
    webhookBodyCache.value[index] = cached.body
  }
}

function updateWebhookValue(index: number) {
  const action = form.value.actions[index]
  if (!action || action.type !== 'call_webhook') return

  const method = webhookMethodCache.value[index] || 'POST'
  const body = webhookBodyCache.value[index] || ''

  // Parse headers from text format "Key: Value" per line
  const headers: Record<string, string> = {}
  const headerLines = (webhookHeadersCache.value[index] || '').split('\n').filter(l => l.trim())
  for (const line of headerLines) {
    const colonIdx = line.indexOf(':')
    if (colonIdx > 0) {
      headers[line.substring(0, colonIdx).trim()] = line.substring(colonIdx + 1).trim()
    }
  }

  action.value = { method, headers, body }
}

// Watch for action changes to init webhook caches
watch(() => form.value.actions, (actions) => {
  actions.forEach((action, index) => {
    if (action.type === 'call_webhook') {
      initWebhookCache(index)
    }
  })
}, { deep: true })

onMounted(() => {
  loadCustomFields()
})

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