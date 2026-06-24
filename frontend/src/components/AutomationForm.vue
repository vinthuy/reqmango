<template>
  <div class="automation-form">
    <!-- 基本信息 -->
    <div class="mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">基本信息</h3>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">规则名称</label>
          <input 
            v-model="form.name" 
            type="text" 
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            placeholder="例如：Bug自动分配给QA"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">描述（可选）</label>
          <input 
            v-model="form.description" 
            type="text" 
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
            placeholder="简要说明这个规则的作用"
          />
        </div>
      </div>
    </div>

    <!-- 触发器 -->
    <div class="mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4 flex items-center">
        <span class="text-xl mr-2">⚡</span>
        触发器
        <span class="text-sm font-normal text-gray-500 ml-2">（当以下事件发生时触发）</span>
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
        条件
        <span class="text-sm font-normal text-gray-500 ml-2">（满足以下所有条件时）</span>
      </h3>
      
      <div v-if="conditions.length === 0" class="text-center py-8 text-gray-400 bg-gray-50 rounded-lg border-2 border-dashed border-gray-200">
        <p>暂无条件限制</p>
        <button @click="addCondition" class="mt-2 text-blue-600 hover:text-blue-700 text-sm">+ 添加条件</button>
      </div>

      <div v-else class="space-y-3">
        <div v-for="(cond, index) in conditions" :key="index" class="flex items-center space-x-2 p-3 bg-gray-50 rounded-lg">
          <!-- 字段选择 -->
          <select v-model="cond.field" class="px-3 py-2 border border-gray-300 rounded-lg text-sm">
            <option value="">选择字段</option>
            <option value="state_group">状态分组</option>
            <option value="priority">优先级</option>
            <option value="assignee">分配人</option>
            <option value="labels">标签</option>
            <option value="issue_type">工作项类型</option>
          </select>
          
          <!-- 操作符选择 -->
          <select v-model="cond.operator" class="px-3 py-2 border border-gray-300 rounded-lg text-sm">
            <option v-for="op in operatorOptions" :key="op.value" :value="op.value">{{ op.label }}</option>
          </select>
          
          <!-- 值选择 -->
          <template v-if="cond.operator !== 'is_empty' && cond.operator !== 'is_not_empty'">
            <select v-if="cond.field === 'state_group'" v-model="cond.value" class="px-3 py-2 border border-gray-300 rounded-lg text-sm flex-1">
              <option value="">选择状态</option>
              <option v-for="s in stateGroupOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
            </select>
            <select v-else-if="cond.field === 'priority'" v-model="cond.value" class="px-3 py-2 border border-gray-300 rounded-lg text-sm flex-1">
              <option value="">选择优先级</option>
              <option v-for="p in priorityOptions" :key="p.value" :value="p.value">{{ p.label }}</option>
            </select>
            <input v-else v-model="cond.value" type="text" class="px-3 py-2 border border-gray-300 rounded-lg text-sm flex-1" placeholder="值" />
          </template>
          
          <span v-else class="flex-1 text-sm text-gray-500 text-center">{{ cond.operator === 'is_empty' ? '为空' : '不为空' }}</span>
          
          <button @click="removeCondition(index)" class="text-gray-400 hover:text-red-500 p-1">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
          </button>
        </div>
        <button @click="addCondition" class="text-blue-600 hover:text-blue-700 text-sm flex items-center">
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          添加条件
        </button>
      </div>
    </div>

    <!-- 动作 -->
    <div class="mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4 flex items-center">
        <span class="text-xl mr-2">🎬</span>
        动作
        <span class="text-sm font-normal text-gray-500 ml-2">（执行以下操作）</span>
      </h3>
      
      <div v-if="actions.length === 0" class="text-center py-8 text-gray-400 bg-gray-50 rounded-lg border-2 border-dashed border-gray-200">
        <p>请添加至少一个动作</p>
      </div>

      <div v-else class="space-y-3">
        <div v-for="(action, index) in actions" :key="index" class="p-4 bg-gray-50 rounded-lg">
          <div class="flex items-center justify-between mb-3">
            <span class="text-sm font-medium text-gray-700">动作 {{ index + 1 }}</span>
            <button @click="removeAction(index)" class="text-gray-400 hover:text-red-500 p-1">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
            </button>
          </div>
          
          <!-- 动作类型 -->
          <select v-model="action.type" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-3">
            <option value="">选择动作</option>
            <option v-for="a in actionOptions" :key="a.value" :value="a.value">{{ a.icon }} {{ a.label }}</option>
          </select>
          
          <!-- 动作参数 -->
          <div class="pl-4 border-l-2 border-blue-200">
            <template v-if="action.type === 'issue.change_state'">
              <select v-model="action.state_id" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                <option value="">选择状态</option>
                <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
              </select>
            </template>
            <template v-else-if="action.type === 'issue.set_priority'">
              <select v-model="action.priority" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                <option value="">选择优先级</option>
                <option v-for="p in priorityOptions" :key="p.value" :value="p.value">{{ p.label }}</option>
              </select>
            </template>
            <template v-else-if="action.type === 'issue.assign'">
              <select v-model="action.assignee_id" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                <option value="">选择人员</option>
                <option v-for="m in members" :key="m.user_id || m.id" :value="m.user_id || m.id">
                  {{ m.user?.display_name || m.display_name || 'Unknown' }}
                </option>
              </select>
            </template>
            <template v-else-if="action.type === 'issue.add_label' || action.type === 'issue.remove_label'">
              <select v-model="action.label_id" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                <option value="">选择标签</option>
                <option v-for="l in labels" :key="l.id" :value="l.id">{{ l.name }}</option>
              </select>
            </template>
            <template v-else-if="action.type === 'issue.add_comment'">
              <textarea v-model="action.comment" rows="2" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="输入评论内容"></textarea>
            </template>
            <template v-else-if="action.type === 'notification.create'">
              <input v-model="action.message" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="通知内容" />
            </template>
            <template v-else>
              <input v-model="action.value" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="值" />
            </template>
          </div>
        </div>
        <button @click="addAction" class="text-blue-600 hover:text-blue-700 text-sm flex items-center">
          <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          添加动作
        </button>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="flex justify-end space-x-3 pt-4 border-t">
      <button @click="$emit('cancel')" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">取消</button>
      <button @click="handleSubmit" class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">保存</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue';
import api from '@/api';
import { TriggerTypeOptions, StateGroupOptions, PriorityOptions, ConditionOperatorOptions, ActionTypeOptions } from '@/types/workflow';

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
    form.trigger = props.automation.trigger_type || 'issue.created';
    
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
      actions.value = Array.isArray(acts) && acts.length > 0 ? acts : [{ type: '', state_id: '', priority: '', assignee_id: '', label_id: '', comment: '', message: '', value: '' }];
    } catch { 
      actions.value = [{ type: '', state_id: '', priority: '', assignee_id: '', label_id: '', comment: '', message: '', value: '' }]; 
    }
  } else {
    isEditing.value = false;
    form.name = '';
    form.description = '';
    form.trigger = 'issue.created';
    conditions.value = [];
    actions.value = [{ type: '', state_id: '', priority: '', assignee_id: '', label_id: '', comment: '', message: '', value: '' }];
  }
}

function addCondition() {
  conditions.value.push({ field: '', operator: 'equals', value: '' });
}

function removeCondition(index: number) {
  conditions.value.splice(index, 1);
}

function addAction() {
  actions.value.push({ type: '', state_id: '', priority: '', assignee_id: '', label_id: '', comment: '', message: '', value: '' });
}

function removeAction(index: number) {
  actions.value.splice(index, 1);
}

function handleSubmit() {
  if (!form.name.trim()) {
    alert('请输入规则名称');
    return;
  }
  if (actions.value.length === 0 || !actions.value.some(a => a.type)) {
    alert('请添加至少一个有效动作');
    return;
  }

  const validConditions = conditions.value.filter(c => c.field && (c.operator === 'is_empty' || c.operator === 'is_not_empty' || c.value));
  const validActions = actions.value.filter(a => a.type).map(a => {
    const action: any = { type: a.type };
    if (a.state_id) action.state_id = a.state_id;
    if (a.priority) action.priority = a.priority;
    if (a.assignee_id) action.assignee_id = a.assignee_id;
    if (a.label_id) action.label_id = a.label_id;
    if (a.comment) action.comment = a.comment;
    if (a.message) action.message = a.message;
    if (a.value) action.value = a.value;
    return action;
  });

  emit('submit', {
    name: form.name,
    description: form.description,
    trigger_type: form.trigger,
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
