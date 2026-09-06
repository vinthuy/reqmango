<template>
  <div class="automation-form max-w-3xl mx-auto">
    <!-- 基本信息 -->
    <div class="bg-white rounded-xl border border-gray-200 p-5 mb-6">
      <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-1">
          {{ t('automationForm.ruleName') }} <span class="text-red-500">*</span>
        </label>
        <input
          v-model="form.name"
          type="text"
          class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          :placeholder="t('automationForm.namePlaceholder')"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">
          {{ t('automationForm.descriptionLabel') }}
        </label>
        <textarea
          v-model="form.description"
          rows="2"
          class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
          :placeholder="t('automationForm.descPlaceholder')"
        />
      </div>
    </div>

    <!-- 项目作用域（仅工作区级规则显示） -->
    <div v-if="scopeEnabled" class="bg-white rounded-xl border border-gray-200 overflow-hidden mb-6">
      <div class="px-5 py-4 bg-gradient-to-r from-purple-50 to-pink-50 border-b border-gray-200">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 rounded-lg bg-purple-100 flex items-center justify-center">
            <span class="text-xl">📂</span>
          </div>
          <div>
            <h4 class="font-semibold text-gray-900">{{ t('automationForm.projectScope') }}</h4>
            <p class="text-sm text-gray-500">{{ t('automationForm.projectScopeHint') }}</p>
          </div>
        </div>
      </div>
      <div class="p-5">
        <div class="space-y-2">
          <label class="flex items-center space-x-2 p-3 bg-gray-50 rounded-lg cursor-pointer hover:bg-gray-100 transition-colors">
            <input type="radio" v-model="form.scope" value="all" class="text-indigo-600" />
            <span class="text-sm text-gray-700">{{ t('automationForm.scopeAllProjects') }}</span>
          </label>
          <label class="flex items-center space-x-2 p-3 bg-gray-50 rounded-lg cursor-pointer hover:bg-gray-100 transition-colors">
            <input type="radio" v-model="form.scope" value="specific" class="text-indigo-600" />
            <span class="text-sm text-gray-700">{{ t('automationForm.scopeSpecificProjects') }}</span>
          </label>
          <div v-if="form.scope === 'specific'" class="pl-8 space-y-1">
            <label v-for="proj in availableProjects" :key="proj.id" class="flex items-center space-x-2 py-1 cursor-pointer">
              <input
                type="checkbox"
                :value="proj.id"
                v-model="selectedProjectIds"
                class="text-indigo-600 rounded"
              />
              <span class="text-sm text-gray-600">{{ proj.name }}</span>
            </label>
            <p v-if="availableProjects.length === 0" class="text-sm text-gray-400 italic">{{ t('automationForm.noProjectsAvailable') }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 触发器 -->
    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden mb-6">
      <div class="px-5 py-4 bg-gradient-to-r from-indigo-50 to-purple-50 border-b border-gray-200">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 rounded-lg bg-indigo-100 flex items-center justify-center">
            <span class="text-xl">⚡</span>
          </div>
          <div>
            <h4 class="font-semibold text-gray-900">{{ t('automationForm.trigger') }}</h4>
            <p class="text-sm text-gray-500">{{ t('automationForm.triggerHint') }}</p>
          </div>
        </div>
      </div>
      <div class="p-5">
        <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
          <button
            v-for="trigger in sortedTriggerOptions"
            :key="trigger.value"
            @click="form.trigger = trigger.value"
            :class="[
              'flex flex-col items-center p-4 border rounded-lg transition-all',
              form.trigger === trigger.value
                ? 'border-indigo-300 bg-indigo-50 text-indigo-700'
                : 'border-gray-200 hover:border-indigo-300 hover:bg-indigo-50'
            ]"
          >
            <span class="text-2xl mb-2">{{ trigger.icon }}</span>
            <span class="text-sm font-medium">{{ trigger.label }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 定时触发器配置（仅当选择 scheduled 时显示） -->
    <div v-if="form.trigger === 'scheduled'" class="bg-white rounded-xl border border-gray-200 overflow-hidden mb-6">
      <div class="px-5 py-4 bg-gradient-to-r from-sky-50 to-blue-50 border-b border-gray-200">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 rounded-lg bg-sky-100 flex items-center justify-center">
            <span class="text-xl">⏱️</span>
          </div>
          <div>
            <h4 class="font-semibold text-gray-900">{{ t('automationForm.scheduleConfig') }}</h4>
          </div>
        </div>
      </div>
      <div class="p-5 space-y-3">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('automationForm.scheduleFrequency') }}</label>
          <select v-model="scheduleForm.frequency" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500">
            <option value="hourly">{{ t('automationForm.freqHourly') }}</option>
            <option value="daily">{{ t('automationForm.freqDaily') }}</option>
            <option value="weekly">{{ t('automationForm.freqWeekly') }}</option>
            <option value="monthly">{{ t('automationForm.freqMonthly') }}</option>
          </select>
        </div>
        <div v-if="scheduleForm.frequency === 'hourly'">
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('automationForm.atMinute') }}</label>
          <select v-model="scheduleForm.minute" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500">
            <option :value="0">{{ t('automationForm.minute00') }}</option>
            <option :value="15">{{ t('automationForm.minute15') }}</option>
            <option :value="30">{{ t('automationForm.minute30') }}</option>
            <option :value="45">{{ t('automationForm.minute45') }}</option>
          </select>
        </div>
        <div v-if="scheduleForm.frequency !== 'hourly'">
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('automationForm.atTime') }}</label>
          <input v-model="scheduleForm.time" type="time" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500" />
        </div>
        <div v-if="scheduleForm.frequency === 'weekly'">
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('automationForm.onDays') }}</label>
          <div class="flex flex-wrap gap-2">
            <label v-for="day in weekDays" :key="day.value" class="flex items-center space-x-1 cursor-pointer">
              <input type="checkbox" :value="day.value" v-model="scheduleForm.days" class="text-indigo-600 rounded" />
              <span class="text-sm text-gray-600">{{ day.label }}</span>
            </label>
          </div>
        </div>
        <div v-if="scheduleForm.frequency === 'monthly'">
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('automationForm.onDayOfMonth') }}</label>
          <select v-model="scheduleForm.day" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500">
            <option v-for="d in 28" :key="d" :value="d">{{ t('automationForm.dayNumber', { day: d }) }}</option>
          </select>
        </div>
      </div>
    </div>

    <!-- 条件 -->
    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden mb-6">
      <div class="px-5 py-4 bg-gradient-to-r from-amber-50 to-orange-50 border-b border-gray-200">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 rounded-lg bg-amber-100 flex items-center justify-center">
              <span class="text-xl">📋</span>
            </div>
            <div>
              <h4 class="font-semibold text-gray-900">{{ t('automationForm.condition') }}</h4>
              <p class="text-sm text-gray-500">{{ t('automationForm.conditionHint') }}</p>
            </div>
          </div>
          <button
            @click="addCondition"
            class="flex items-center space-x-1 px-3 py-1.5 bg-amber-100 text-amber-700 rounded-lg hover:bg-amber-200 transition-colors text-sm font-medium"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
            <span>{{ t('automationForm.addCondition') }}</span>
          </button>
        </div>
      </div>
      <div class="p-5">
        <div v-if="conditions.length === 0" class="text-center py-8 text-gray-500">
          <span class="text-4xl block mb-2">🎯</span>
          <p class="text-sm">{{ t('automationForm.noConditions') }}</p>
        </div>
        <div v-else class="space-y-3">
          <div v-for="(cond, index) in conditions" :key="index" class="flex items-center space-x-2 p-3 bg-gray-50 rounded-lg group">
            <span class="text-gray-400 text-sm font-medium shrink-0">IF</span>
            <!-- 字段选择 -->
            <select v-model="cond.field" class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
              <option value="" disabled>{{ t('automationForm.selectField') }}</option>
              <option value="state_group">{{ t('automationForm.stateGroup') }}</option>
              <option value="priority">{{ t('automationForm.priority') }}</option>
              <option value="assignee">{{ t('automationForm.assignee') }}</option>
              <option value="labels">{{ t('automationForm.labels') }}</option>
              <option value="issue_type">{{ t('automationForm.issueType') }}</option>
            </select>
            <!-- 操作符选择 -->
            <select v-model="cond.operator" class="min-w-[90px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
              <option v-for="op in operatorOptions" :key="op.value" :value="op.value">{{ op.label }}</option>
            </select>
            <!-- 值选择 -->
            <template v-if="cond.operator !== 'is_empty' && cond.operator !== 'is_not_empty'">
              <select v-if="cond.field === 'state_group'" v-model="cond.value" class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                <option value="" disabled>{{ t('automationForm.selectState') }}</option>
                <option v-for="s in stateGroupOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
              </select>
              <select v-else-if="cond.field === 'priority'" v-model="cond.value" class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                <option value="" disabled>{{ t('automationForm.selectPriority') }}</option>
                <option v-for="p in priorityOptions" :key="p.value" :value="p.value">{{ p.label }}</option>
              </select>
              <input v-else v-model="cond.value" type="text" class="flex-1 min-w-[100px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent" :placeholder="t('automationForm.value')" />
            </template>
            <span v-else class="flex-1 text-sm text-gray-500 text-center">{{ cond.operator === 'is_empty' ? t('automationForm.isEmpty') : t('automationForm.isNotEmpty') }}</span>
            <button @click="removeCondition(index)" class="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-red-600 transition-all shrink-0">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
            </button>
          </div>
          <div v-if="conditions.length > 0" class="flex items-center justify-center text-gray-400 text-sm">
            <span>AND</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 动作 -->
    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden mb-6">
      <div class="px-5 py-4 bg-gradient-to-r from-emerald-50 to-teal-50 border-b border-gray-200">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 rounded-lg bg-emerald-100 flex items-center justify-center">
              <span class="text-xl">🎬</span>
            </div>
            <div>
              <h4 class="font-semibold text-gray-900">{{ t('automationForm.action') }}</h4>
              <p class="text-sm text-gray-500">{{ t('automationForm.actionHint') }}</p>
            </div>
          </div>
          <button
            @click="addAction"
            class="flex items-center space-x-1 px-3 py-1.5 bg-emerald-100 text-emerald-700 rounded-lg hover:bg-emerald-200 transition-colors text-sm font-medium"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
            <span>{{ t('automationForm.addAction') }}</span>
          </button>
        </div>
      </div>
      <div class="p-5">
        <div v-if="actions.length === 0 || !actions.some(a => a.type)" class="text-center py-8 text-gray-500">
          <span class="text-4xl block mb-2">✨</span>
          <p class="text-sm">{{ t('automationForm.noActions') }}</p>
        </div>

        <div v-else class="space-y-3">
          <div v-for="(action, index) in actions" :key="index" class="p-3 bg-gray-50 rounded-lg group">
            <div class="flex items-center space-x-2 mb-3">
              <span class="text-gray-400 text-sm font-medium shrink-0">DO</span>
              <!-- 动作类型 -->
              <select v-model="action.type" class="flex-1 min-w-[120px] px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                <option value="" disabled>{{ t('automationForm.selectAction') }}</option>
                <option v-for="a in actionOptions" :key="a.value" :value="a.value">{{ a.icon }} {{ a.label }}</option>
              </select>
              <button @click="removeAction(index)" class="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-red-600 transition-all shrink-0">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
              </button>
            </div>
            
            <!-- 动作参数 -->
            <div class="pl-6 ml-1 border-l-2 border-gray-200 space-y-2">
              <template v-if="action.type === 'change_state'">
                <select v-model="action.value" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                  <option value="" disabled>{{ t('automationForm.selectState') }}</option>
                  <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
                </select>
              </template>
              <template v-else-if="action.type === 'set_priority'">
                <select v-model="action.value" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                  <option value="" disabled>{{ t('automationForm.selectPriority') }}</option>
                  <option v-for="p in priorityOptions" :key="p.value" :value="p.value">{{ p.label }}</option>
                </select>
              </template>
              <template v-else-if="action.type === 'assign_to'">
                <select v-model="action.value" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                  <option value="" disabled>{{ t('automationForm.selectPerson') }}</option>
                  <option v-for="m in members" :key="m.user_id || m.id" :value="m.user_id || m.id">
                    {{ m.user?.display_name || m.display_name || 'Unknown' }}
                  </option>
                </select>
              </template>
              <template v-else-if="action.type === 'add_comment'">
                <textarea v-model="action.value" rows="2" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none" :placeholder="t('automationForm.commentPlaceholder')" />
              </template>
              <template v-else-if="action.type === 'set_field'">
                <select
                  v-model="action.field"
                  @change="onSetFieldChange(index)"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent mb-2"
                >
                  <option value="" disabled>{{ t('automationForm.selectField') }}</option>
                  <optgroup :label="t('automationForm.systemFieldGroup')">
                    <option value="state_id">{{ t('automationForm.state') }}</option>
                    <option value="priority">{{ t('automationForm.priority') }}</option>
                    <option value="target_date">{{ t('automationForm.dueDate') }}</option>
                    <option value="start_date">{{ t('automationForm.startDate') }}</option>
                  </optgroup>
                  <optgroup :label="t('automationForm.customFieldGroup')">
                    <option v-for="field in customFields" :key="field.id" :value="'custom_' + field.id">{{ field.name }}</option>
                  </optgroup>
                </select>
                <template v-if="action.field">
                  <select v-if="action.field === 'state_id'" v-model="action.value" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                    <option value="" disabled>{{ t('automationForm.selectState') }}</option>
                    <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
                  </select>
                  <select v-else-if="action.field === 'priority'" v-model="action.value" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                    <option value="" disabled>{{ t('automationForm.selectPriority') }}</option>
                    <option v-for="p in priorityOptions" :key="p.value" :value="p.value">{{ p.label }}</option>
                  </select>
                  <input v-else-if="action.field === 'target_date' || action.field === 'start_date'" v-model="action.value" type="date" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
                  <select v-else-if="getCustomFieldType(action.field) === 'dropdown'" v-model="action.value" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                    <option value="" disabled>{{ t('automationForm.selectValue') }}</option>
                    <option v-for="opt in getCustomFieldOptions(action.field)" :key="opt.id" :value="opt.value">{{ opt.value }}</option>
                  </select>
                  <input v-else-if="getCustomFieldType(action.field) === 'number'" v-model.number="action.value" type="number" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent" :placeholder="t('automationForm.numberValue')" />
                  <select v-else-if="getCustomFieldType(action.field) === 'boolean'" v-model="action.value" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                    <option value="" disabled>{{ t('automationForm.selectValue') }}</option>
                    <option value="true">True</option>
                    <option value="false">False</option>
                  </select>
                  <input v-else-if="getCustomFieldType(action.field) === 'date'" v-model="action.value" type="date" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
                  <select v-else-if="getCustomFieldType(action.field) === 'member'" v-model="action.value" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                    <option value="" disabled>{{ t('automationForm.selectPerson') }}</option>
                    <option v-for="m in members" :key="m.user_id || m.id" :value="m.user_id || m.id">{{ m.user?.display_name || m.display_name || 'Unknown' }}</option>
                  </select>
                  <textarea v-else-if="getCustomFieldType(action.field) === 'text'" v-model="action.value" rows="3" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none" :placeholder="t('automationForm.value')" />
                  <input v-else-if="getCustomFieldType(action.field) === 'url'" v-model="action.value" type="url" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent" placeholder="https://..." />
                  <input v-else v-model="action.value" type="text" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent" :placeholder="t('automationForm.value')" />
                </template>
              </template>
              <template v-else-if="action.type === 'call_webhook'">
                <input v-model="action.field" type="url" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent mb-2" :placeholder="t('automationForm.webhookUrl')" />
                <div class="grid grid-cols-2 gap-2 mb-2">
                  <select v-model="webhookMethodCache[index]" @change="updateWebhookValue(index)" class="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                    <option value="POST">POST</option>
                    <option value="GET">GET</option>
                    <option value="PUT">PUT</option>
                    <option value="PATCH">PATCH</option>
                    <option value="DELETE">DELETE</option>
                  </select>
                  <input v-model="webhookHeadersCache[index]" type="text" class="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent" :placeholder="t('automationForm.webhookHeaders')" @change="updateWebhookValue(index)" />
                </div>
                <textarea v-model="webhookBodyCache[index]" rows="2" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm mb-1 font-mono focus:ring-2 focus:ring-indigo-500 focus:border-transparent" :placeholder="t('automationForm.webhookBody')" @change="updateWebhookValue(index)" />
                <p class="text-xs text-gray-400">{{ t('automationForm.webhookVarsHint') }}</p>
              </template>
              <template v-else-if="action.type === 'dispatch_agent'">
                <input v-model="action.value" type="text" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent" :placeholder="t('automationForm.agentNamePlaceholder')" />
              </template>
              <template v-else-if="action.type === 'rollup_to_parent'">
                <div class="space-y-2">
                  <div v-for="(rule, ri) in (action.value?.rules || [])" :key="ri" class="flex items-center gap-1 group">
                    <select v-model="rule.condition" class="px-2 py-1.5 border border-gray-200 rounded text-xs focus:ring-2 focus:ring-indigo-500">
                      <option value="all">{{ t('automationBuilder.rollupAll') }}</option>
                      <option value="any">{{ t('automationBuilder.rollupAny') }}</option>
                    </select>
                    <span class="text-xs text-gray-400">{{ t('automationBuilder.rollupChildrenAre') }}</span>
                    <select v-model="rule.child_state" class="flex-1 px-2 py-1.5 border border-gray-200 rounded text-xs focus:ring-2 focus:ring-indigo-500">
                      <option value="" disabled>{{ t('automationForm.selectState') }}</option>
                      <option v-for="s in states" :key="s.id" :value="s.name">{{ s.name }}</option>
                    </select>
                    <span class="text-xs text-gray-400">→</span>
                    <select v-model="rule.parent_state" class="flex-1 px-2 py-1.5 border border-gray-200 rounded text-xs focus:ring-2 focus:ring-indigo-500">
                      <option value="" disabled>{{ t('automationBuilder.rollupSetParentTo') }}</option>
                      <option v-for="s in states" :key="s.id" :value="s.name">{{ s.name }}</option>
                    </select>
                    <button @click="removeRollupRule(index, ri)" class="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-red-500 text-xs transition-all">✕</button>
                  </div>
                  <button @click="addRollupRule(index)" class="text-xs text-indigo-600 hover:text-indigo-800 flex items-center">
                    <svg class="w-3 h-3 mr-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
                    {{ t('automationBuilder.rollupAddRule') }}
                  </button>
                </div>
              </template>
              <template v-else>
                <input v-model="action.value" type="text" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent" :placeholder="t('automationForm.value')" />
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="flex justify-end space-x-3 pt-5 border-t border-gray-200">
      <button @click="$emit('cancel')" class="px-5 py-2.5 border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700 font-medium transition-colors">{{ t('common.cancel') }}</button>
      <button @click="handleSubmit" class="px-5 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-medium transition-colors">{{ t('common.save') }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted, computed } from 'vue';
import { useI18n } from '@/composables/useI18n';
import { useToast } from '@/composables/useToast';
import api from '@/api';
import { TriggerTypeOptions, StateGroupOptions, PriorityOptions, ConditionOperatorOptions, ActionTypeOptions } from '@/types/workflow';
import type { CustomField } from '@/types/custom-field';
import { customFieldApi } from '@/api/custom-field';

const { t } = useI18n();
const toast = useToast();

const props = defineProps<{
  projectId: number;
  workspaceId: number;
  automation?: any;
  projects?: any[];      // workspace projects for scope selection
  scopeEnabled?: boolean; // whether to show the project scope selector
}>();

const emit = defineEmits(['submit', 'cancel']);

const triggerOptions = TriggerTypeOptions;

// Sort triggers so that "scheduled" appears at the end of the event-based triggers section
const sortedTriggerOptions = computed(() => {
  return [...triggerOptions].sort((a, b) => {
    if (a.value === 'scheduled') return 1;
    if (b.value === 'scheduled') return -1;
    return 0;
  });
});

const stateGroupOptions = StateGroupOptions;
const priorityOptions = PriorityOptions;
const operatorOptions = ConditionOperatorOptions;
const actionOptions = ActionTypeOptions;

const states = ref<any[]>([]);
const labels = ref<any[]>([]);
const members = ref<any[]>([]);
const customFields = ref<CustomField[]>([]);

// Webhook action caches
const webhookMethodCache = ref<Record<number, string>>({});
const webhookHeadersCache = ref<Record<number, string>>({});
const webhookBodyCache = ref<Record<number, string>>({});

const isEditing = ref(false);

const form = reactive({
  name: '',
  description: '',
  trigger: 'issue.created',
  scope: 'all',
});

// Project scope multi-select
const selectedProjectIds = ref<number[]>([]);
const availableProjects = computed(() => props.projects || []);

// Schedule config
const scheduleForm = reactive({
  frequency: 'daily',
  time: '09:00',
  minute: 0,
  days: [] as string[],
  day: 1,
});

const weekDays = computed(() => [
  { value: 'mon', label: t('automationForm.weekMonday') },
  { value: 'tue', label: t('automationForm.weekTuesday') },
  { value: 'wed', label: t('automationForm.weekWednesday') },
  { value: 'thu', label: t('automationForm.weekThursday') },
  { value: 'fri', label: t('automationForm.weekFriday') },
  { value: 'sat', label: t('automationForm.weekSaturday') },
  { value: 'sun', label: t('automationForm.weekSunday') },
]);

const conditions = ref<any[]>([]);
const actions = ref<any[]>([]);

async function loadData() {
  try {
    const [statesRes, labelsRes, membersRes] = await Promise.all([
      api.get(`/projects/${props.projectId}/settings/states`).then(r => r.data),
      api.get(`/projects/${props.projectId}/settings/labels`).then(r => r.data),
      api.get(`/projects/${props.projectId}/members`).then(r => r.data),
    ]);
    states.value = Array.isArray(statesRes) ? statesRes : (statesRes?.data || []);
    labels.value = Array.isArray(labelsRes) ? labelsRes : (labelsRes?.data || []);
    members.value = Array.isArray(membersRes) ? membersRes : (membersRes?.data || []);
    // 加载自定义字段
    try {
      const fields = await customFieldApi.listCustomFields(props.workspaceId, props.projectId);
      customFields.value = fields.filter((f: CustomField) => f.is_active);
    } catch (e) {
      console.error('Failed to load custom fields:', e);
    }
  } catch (e) {
    console.error('Failed to load data:', e);
  }
}

function loadAutomationData() {
  if (props.automation) {
    isEditing.value = true;
    form.name = props.automation.name || '';
    form.description = props.automation.description || '';
    // Parse trigger_type - it may be JSON like {"type":"issue.created"} or plain string
    try {
      const parsed = JSON.parse(props.automation.trigger_type || '{}')
      form.trigger = parsed?.type || props.automation.trigger_type || 'issue.created'
    } catch {
      form.trigger = props.automation.trigger_type || 'issue.created'
    }

    // Load scope
    if (props.automation.scope) {
      if (props.automation.scope === 'all') {
        form.scope = 'all';
      } else {
        try {
          const ids = JSON.parse(props.automation.scope);
          if (Array.isArray(ids) && ids.length > 0) {
            form.scope = 'specific';
            selectedProjectIds.value = ids;
          } else {
            form.scope = 'all';
          }
        } catch {
          form.scope = 'all';
        }
      }
    } else {
      form.scope = 'all';
    }

    // Load schedule config
    if (props.automation.schedule_config) {
      try {
        const sc = JSON.parse(props.automation.schedule_config);
        scheduleForm.frequency = sc.frequency || 'daily';
        scheduleForm.time = sc.time || '09:00';
        scheduleForm.minute = sc.minute || 0;
        scheduleForm.days = sc.days || [];
        scheduleForm.day = sc.day || 1;
      } catch { /* ignore */ }
    }
    
    // 解析条件
    try {
      const conds = props.automation.conditions 
        ? (typeof props.automation.conditions === 'string' ? JSON.parse(props.automation.conditions) : props.automation.conditions)
        : [];
      conditions.value = Array.isArray(conds) ? conds : [];
    } catch { conditions.value = []; }
    
    // 解析动作
    try {
      const acts = props.automation.actions
        ? (typeof props.automation.actions === 'string' ? JSON.parse(props.automation.actions) : props.automation.actions)
        : [];
      actions.value = Array.isArray(acts) && acts.length > 0 ? acts : [{ type: '', field: '', value: '' }];
      // 初始化 webhook 缓存和 rollup 规则
      actions.value.forEach((action, index) => {
        if (action.type === 'call_webhook') {
          initWebhookCache(index);
        }
        if (action.type === 'rollup_to_parent') {
          if (!action.value || typeof action.value !== 'object' || !Array.isArray(action.value.rules)) {
            action.value = { rules: [] }
          }
        }
      });
    } catch {
      actions.value = [{ type: '', field: '', value: '' }];
    }
  } else {
    isEditing.value = false;
    form.name = '';
    form.description = '';
    form.trigger = 'issue.created';
    form.scope = 'all';
    selectedProjectIds.value = [];
    scheduleForm.frequency = 'daily';
    scheduleForm.time = '09:00';
    scheduleForm.minute = 0;
    scheduleForm.days = [];
    scheduleForm.day = 1;
    conditions.value = [];
    actions.value = [{ type: '', field: '', value: '' }];
  }
}

function addCondition() {
  conditions.value.push({ field: '', operator: 'equals', value: '' });
}

function removeCondition(index: number) {
  conditions.value.splice(index, 1);
}

function addAction() {
  actions.value.push({ type: '', field: '', value: '' });
}

function removeAction(index: number) {
  actions.value.splice(index, 1);
  delete webhookMethodCache.value[index];
  delete webhookHeadersCache.value[index];
  delete webhookBodyCache.value[index];
}

function addRollupRule(actionIndex: number) {
  const action = actions.value[actionIndex]
  if (!action) return
  if (!action.value || typeof action.value !== 'object' || !Array.isArray(action.value.rules)) {
    action.value = { rules: [] }
  }
  action.value.rules.push({ condition: 'all', child_state: '', parent_state: '' })
}

function removeRollupRule(actionIndex: number, ruleIndex: number) {
  const action = actions.value[actionIndex]
  if (!action || !action.value || !Array.isArray(action.value.rules)) return
  action.value.rules.splice(ruleIndex, 1)
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
  const action = actions.value[index]
  if (!action) return
  action.field = action.field || ''
  action.value = ''
}

// ====== webhook action helpers ======

function initWebhookCache(index: number) {
  const action = actions.value[index]
  if (!action) return
  if (action.type === 'call_webhook') {
    let cached: { method: string; headers: string; body: string } = { method: 'POST', headers: '', body: '' }
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
  const action = actions.value[index]
  if (!action || action.type !== 'call_webhook') return

  const method = webhookMethodCache.value[index] || 'POST'
  const body = webhookBodyCache.value[index] || ''

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

function handleSubmit() {
  if (!form.name.trim()) {
    toast.warning(t('automationForm.nameRequired'));
    return;
  }
  if (actions.value.length === 0 || !actions.value.some(a => a.type)) {
    toast.warning(t('automationForm.actionRequired'));
    return;
  }

  if (form.trigger === 'scheduled' && !scheduleForm.time && scheduleForm.frequency !== 'hourly') {
    toast.warning(t('automationForm.scheduleTimeRequired'));
    return;
  }

  const validConditions = conditions.value.filter(c => c.field && (c.operator === 'is_empty' || c.operator === 'is_not_empty' || c.value));
  const validActions = actions.value.filter(a => a.type).map(a => {
    const action: any = { type: a.type };
    if (a.field) action.field = a.field;
    if (a.value !== '' && a.value != null) action.value = a.value;
    return action;
  });

  // Build scope
  let scope = 'all';
  if (props.scopeEnabled && form.scope === 'specific' && selectedProjectIds.value.length > 0) {
    scope = JSON.stringify(selectedProjectIds.value);
  }

  // Build schedule_config
  let scheduleConfig = '';
  if (form.trigger === 'scheduled') {
    const sc: any = { frequency: scheduleForm.frequency };
    if (scheduleForm.frequency === 'hourly') {
      sc.minute = scheduleForm.minute;
    } else {
      sc.time = scheduleForm.time;
    }
    if (scheduleForm.frequency === 'weekly') {
      sc.days = scheduleForm.days;
    }
    if (scheduleForm.frequency === 'monthly') {
      sc.day = scheduleForm.day;
    }
    scheduleConfig = JSON.stringify(sc);
  }

  emit('submit', {
    name: form.name,
    description: form.description,
    trigger_type: form.trigger === 'scheduled' ? 'scheduled' : JSON.stringify({ type: form.trigger }),
    conditions: JSON.stringify(validConditions),
    actions: JSON.stringify(validActions),
    scope,
    schedule_config: scheduleConfig,
  });
}

watch(() => props.automation, loadAutomationData, { immediate: true });

onMounted(() => {
  loadData();
  loadAutomationData();
});
</script>

<style scoped>
.automation-form {
  max-height: 80vh;
  overflow-y: auto;
  padding: 0.5rem;
}
</style>
