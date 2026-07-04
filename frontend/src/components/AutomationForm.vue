<template>
  <div class="automation-form">
    <!-- 基本信息 -->
    <div class="mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">{{ t('automationForm.basicInfo') }}</h3>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('automationForm.ruleName') }}</label>
          <input
            v-model="form.name"
            type="text"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            :placeholder="t('automationForm.namePlaceholder')"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('automationForm.descriptionLabel') }}</label>
          <input
            v-model="form.description"
            type="text"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            :placeholder="t('automationForm.descPlaceholder')"
          />
        </div>
      </div>
    </div>

    <!-- 触发器 -->
    <div class="mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4 flex items-center">
        <span class="text-xl mr-2">⚡</span>
        {{ t('automationForm.trigger') }}
        <span class="text-sm font-normal text-gray-500 ml-2">{{ t('automationForm.triggerHint') }}</span>
      </h3>
      <div class="grid grid-cols-2 md:grid-cols-3 gap-2">
        <button
          v-for="trigger in triggerOptions"
          :key="trigger.value"
          @click="form.trigger = trigger.value"
          :class="[
            'flex items-center space-x-2 px-4 py-3 rounded-lg border transition-all',
            form.trigger === trigger.value 
              ? 'border-blue-500 bg-blue-50 text-blue-700' 
              : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
          ]"
        >
          <span>{{ trigger.icon }}</span>
          <span class="text-sm">{{ trigger.label }}</span>
        </button>
      </div>
    </div>

    <!-- 条件 -->
    <div class="mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4 flex items-center">
        <span class="text-xl mr-2">🎯</span>
        {{ t('automationForm.condition') }}
        <span class="text-sm font-normal text-gray-500 ml-2">{{ t('automationForm.conditionHint') }}</span>
      </h3>
      
      <div v-if="conditions.length === 0" class="text-center py-8 text-gray-400 bg-gray-50 rounded-lg border-2 border-dashed border-gray-200">
        <p>{{ t('automationForm.noConditions') }}</p>
        <button @click="addCondition" class="mt-2 text-blue-600 hover:text-blue-700 text-sm">+ {{ t('automationForm.addCondition') }}</button>
      </div>

      <div v-else class="space-y-3">
        <div v-for="(cond, index) in conditions" :key="index" class="flex items-center space-x-2 p-3 bg-gray-50 rounded-lg">
          <!-- 字段选择 -->
          <select v-model="cond.field" class="px-3 py-2 border border-gray-300 rounded-lg text-sm">
            <option value="">{{ t('automationForm.selectField') }}</option>
            <option value="state_group">{{ t('automationForm.stateGroup') }}</option>
            <option value="priority">{{ t('automationForm.priority') }}</option>
            <option value="assignee">{{ t('automationForm.assignee') }}</option>
            <option value="labels">{{ t('automationForm.labels') }}</option>
            <option value="issue_type">{{ t('automationForm.issueType') }}</option>
          </select>
          
          <!-- 操作符选择 -->
          <select v-model="cond.operator" class="px-3 py-2 border border-gray-300 rounded-lg text-sm">
            <option v-for="op in operatorOptions" :key="op.value" :value="op.value">{{ op.label }}</option>
          </select>
          
          <!-- 值选择 -->
          <template v-if="cond.operator !== 'is_empty' && cond.operator !== 'is_not_empty'">
            <select v-if="cond.field === 'state_group'" v-model="cond.value" class="px-3 py-2 border border-gray-300 rounded-lg text-sm flex-1">
              <option value="">{{ t('automationForm.selectState') }}</option>
              <option v-for="s in stateGroupOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
            </select>
            <select v-else-if="cond.field === 'priority'" v-model="cond.value" class="px-3 py-2 border border-gray-300 rounded-lg text-sm flex-1">
              <option value="">{{ t('automationForm.selectPriority') }}</option>
              <option v-for="p in priorityOptions" :key="p.value" :value="p.value">{{ p.label }}</option>
            </select>
            <input v-else v-model="cond.value" type="text" class="px-3 py-2 border border-gray-300 rounded-lg text-sm flex-1" :placeholder="t('automationForm.value')" />
          </template>
          
          <span v-else class="flex-1 text-sm text-gray-500 text-center">{{ cond.operator === 'is_empty' ? t('automationForm.isEmpty') : t('automationForm.isNotEmpty') }}</span>
          
          <button @click="removeCondition(index)" class="text-gray-400 hover:text-red-500 p-1">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
          </button>
        </div>
        <button @click="addCondition" class="text-blue-600 hover:text-blue-700 text-sm flex items-center">
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          {{ t('automationForm.addCondition') }}
        </button>
      </div>
    </div>

    <!-- 动作 -->
    <div class="mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4 flex items-center">
        <span class="text-xl mr-2">🎬</span>
        {{ t('automationForm.action') }}
        <span class="text-sm font-normal text-gray-500 ml-2">{{ t('automationForm.actionHint') }}</span>
      </h3>
      
      <div v-if="actions.length === 0" class="text-center py-8 text-gray-400 bg-gray-50 rounded-lg border-2 border-dashed border-gray-200">
        <p>{{ t('automationForm.noActions') }}</p>
      </div>

      <div v-else class="space-y-3">
        <div v-for="(action, index) in actions" :key="index" class="p-4 bg-gray-50 rounded-lg">
          <div class="flex items-center justify-between mb-3">
            <span class="text-sm font-medium text-gray-700">{{ t('automationForm.actionLabel') }} {{ index + 1 }}</span>
            <button @click="removeAction(index)" class="text-gray-400 hover:text-red-500 p-1">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
            </button>
          </div>
          
          <!-- 动作类型 -->
          <select v-model="action.type" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-3">
            <option value="">{{ t('automationForm.selectAction') }}</option>
            <option v-for="a in actionOptions" :key="a.value" :value="a.value">{{ a.icon }} {{ a.label }}</option>
          </select>
          
          <!-- 动作参数 -->
          <div class="pl-4 border-l-2 border-blue-200">
            <template v-if="action.type === 'change_state'">
              <select v-model="action.value" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                <option value="">{{ t('automationForm.selectState') }}</option>
                <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
              </select>
            </template>
            <template v-else-if="action.type === 'set_priority'">
              <select v-model="action.value" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                <option value="">{{ t('automationForm.selectPriority') }}</option>
                <option v-for="p in priorityOptions" :key="p.value" :value="p.value">{{ p.label }}</option>
              </select>
            </template>
            <template v-else-if="action.type === 'assign_to'">
              <select v-model="action.value" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                <option value="">{{ t('automationForm.selectPerson') }}</option>
                <option v-for="m in members" :key="m.user_id || m.id" :value="m.user_id || m.id">
                  {{ m.user?.display_name || m.display_name || 'Unknown' }}
                </option>
              </select>
            </template>
            <template v-else-if="action.type === 'add_label' || action.type === 'remove_label'">
              <select v-model="action.value" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                <option value="">{{ t('automationForm.selectLabel') }}</option>
                <option v-for="l in labels" :key="l.id" :value="l.id">{{ l.name }}</option>
              </select>
            </template>
            <template v-else-if="action.type === 'add_comment'">
              <textarea v-model="action.value" rows="2" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" :placeholder="t('automationForm.commentPlaceholder')"></textarea>
            </template>
            <template v-else-if="action.type === 'set_field'">
              <input v-model="action.field" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-2" :placeholder="t('automationForm.fieldName')" />
              <input v-model="action.value" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" :placeholder="t('automationForm.value')" />
            </template>
            <template v-else-if="action.type === 'dispatch_agent'">
              <input v-model="action.value" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" :placeholder="t('automationForm.agentNamePlaceholder')" />
            </template>
            <template v-else>
              <input v-model="action.value" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" :placeholder="t('automationForm.value')" />
            </template>
          </div>
        </div>
        <button @click="addAction" class="text-blue-600 hover:text-blue-700 text-sm flex items-center">
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          {{ t('automationForm.addAction') }}
        </button>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="flex justify-end space-x-3 pt-4 border-t">
      <button @click="$emit('cancel')" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('common.cancel') }}</button>
      <button @click="handleSubmit" class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">{{ t('common.save') }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue';
import { useI18n } from '@/composables/useI18n';
import { useToast } from '@/composables/useToast';
import api from '@/api';
import { TriggerTypeOptions, StateGroupOptions, PriorityOptions, ConditionOperatorOptions, ActionTypeOptions } from '@/types/workflow';

const { t } = useI18n();
const toast = useToast();

const props = defineProps<{
  projectId: number;
  workspaceId: number;
  automation?: any;
}>();

const emit = defineEmits(['submit', 'cancel']);

const triggerOptions = TriggerTypeOptions;
const stateGroupOptions = StateGroupOptions;
const priorityOptions = PriorityOptions;
const operatorOptions = ConditionOperatorOptions;
const actionOptions = ActionTypeOptions;

const states = ref<any[]>([]);
const labels = ref<any[]>([]);
const members = ref<any[]>([]);

const isEditing = ref(false);

const form = reactive({
  name: '',
  description: '',
  trigger: 'issue.created',
});

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
    } catch {
      actions.value = [{ type: '', field: '', value: '' }];
    }
  } else {
    isEditing.value = false;
    form.name = '';
    form.description = '';
    form.trigger = 'issue.created';
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

  const validConditions = conditions.value.filter(c => c.field && (c.operator === 'is_empty' || c.operator === 'is_not_empty' || c.value));
  const validActions = actions.value.filter(a => a.type).map(a => {
    const action: any = { type: a.type };
    if (a.field) action.field = a.field;
    if (a.value !== '' && a.value != null) action.value = a.value;
    return action;
  });

  emit('submit', {
    name: form.name,
    description: form.description,
    trigger_type: JSON.stringify({ type: form.trigger }),
    conditions: JSON.stringify(validConditions),
    actions: JSON.stringify(validActions),
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
}
</style>
