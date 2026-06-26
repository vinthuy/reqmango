<template>
  <div class="automation-list">
    <div v-if="automations.length === 0" class="text-center py-12">
      <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-gray-100 flex items-center justify-center">
        <span class="text-3xl">🤖</span>
      </div>
      <h3 class="text-lg font-medium text-gray-700 mb-2">暂无自动化规则</h3>
      <p class="text-gray-500 mb-4">创建自动化规则来简化您的工作流程</p>
      <button @click="$emit('create')" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
        + 创建自动化
      </button>
    </div>

    <div v-else class="space-y-4">
      <div 
        v-for="automation in automations" 
        :key="automation.id" 
        class="bg-white rounded-xl border border-gray-200 overflow-hidden hover:shadow-md transition-shadow"
      >
        <!-- 头部 -->
        <div class="p-4 flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <div :class="['w-10 h-10 rounded-lg flex items-center justify-center', automation.is_enabled ? 'bg-green-100' : 'bg-gray-100']">
              <span class="text-lg">{{ automation.is_enabled ? '✅' : '⏸️' }}</span>
            </div>
            <div>
              <h3 class="font-medium text-gray-900">{{ automation.name }}</h3>
              <p v-if="automation.description" class="text-sm text-gray-500">{{ automation.description }}</p>
            </div>
          </div>
          <div class="flex items-center space-x-2">
            <button 
              @click="toggleEnabled(automation)"
              :class="[
                'px-3 py-1 rounded-full text-xs font-medium transition-colors cursor-pointer',
                automation.is_enabled 
                  ? 'bg-green-100 text-green-700 hover:bg-green-200' 
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              ]"
            >
              {{ automation.is_enabled ? '已启用' : '已停用' }}
            </button>
            <button @click="$emit('edit', automation)" class="p-1 text-gray-400 hover:text-blue-600">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/></svg>
            </button>
            <button @click="handleDelete(automation)" class="p-1 text-gray-400 hover:text-red-500">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
            </button>
          </div>
        </div>

        <!-- 规则详情 -->
        <div class="px-4 pb-4">
          <div class="flex flex-wrap items-center gap-2 text-sm">
            <!-- 触发器 -->
            <span class="inline-flex items-center px-2 py-1 rounded bg-purple-100 text-purple-700">
              <span class="mr-1">⚡</span>
              {{ getTriggerLabel(automation.trigger_type) }}
            </span>

            <!-- 条件 -->
            <template v-if="getConditions(automation).length > 0">
              <span class="text-gray-400">|</span>
              <span class="inline-flex items-center px-2 py-1 rounded bg-blue-50 text-blue-700">
                <span class="mr-1">🎯</span>
                {{ getConditions(automation).length }} 个条件
              </span>
            </template>

            <!-- 动作 -->
            <template v-if="getActions(automation).length > 0">
              <span class="text-gray-400">|</span>
              <span class="inline-flex items-center px-2 py-1 rounded bg-green-50 text-green-700">
                <span class="mr-1">🎬</span>
                {{ getActions(automation).length }} 个动作
              </span>
            </template>

            <!-- 执行统计 -->
            <template v-if="automation.execution_count > 0">
              <span class="text-gray-400">|</span>
              <span class="text-gray-500 text-xs">
                已执行 {{ automation.execution_count }} 次
              </span>
            </template>
          </div>

          <!-- 条件详情预览 -->
          <div v-if="getConditions(automation).length > 0" class="mt-3 text-xs text-gray-500">
            <span v-for="(cond, i) in getConditions(automation)" :key="i">
              {{ getFieldLabel(cond.field) }} {{ getOperatorLabel(cond.operator) }} {{ cond.value }}
              <span v-if="i < getConditions(automation).length - 1"> 且 </span>
            </span>
          </div>

          <!-- 动作详情预览 -->
          <div v-if="getActions(automation).length > 0" class="mt-2 text-xs text-gray-500">
            <span v-for="(act, i) in getActions(automation)" :key="i">
              {{ getActionLabel(act.type) }}
              <span v-if="i < getActions(automation).length - 1"> → </span>
            </span>
          </div>
        </div>
      </div>

      <button @click="$emit('create')" class="w-full py-4 border-2 border-dashed border-gray-300 rounded-xl text-gray-500 hover:border-blue-400 hover:text-blue-600 transition-colors">
        + 添加自动化规则
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { TriggerTypeOptions, ConditionOperatorOptions, ActionTypeOptions } from '@/types/workflow';

defineProps<{
  automations: any[];
}>();

const emit = defineEmits(['create', 'edit', 'delete', 'toggle']);

function getTriggerLabel(triggerType: string): string {
  const trigger = TriggerTypeOptions.find(t => t.value === triggerType);
  return trigger ? trigger.label : triggerType;
}

function getFieldLabel(field: string): string {
  const labels: Record<string, string> = {
    state_group: '状态分组',
    priority: '优先级',
    assignee: '分配人',
    labels: '标签',
    issue_type: '工作项类型',
  };
  return labels[field] || field;
}

function getOperatorLabel(op: string): string {
  const option = ConditionOperatorOptions.find(o => o.value === op);
  return option ? option.label : op;
}

function getActionLabel(actionType: string): string {
  const action = ActionTypeOptions.find(a => a.value === actionType);
  return action ? action.label : actionType;
}

function getConditions(automation: any): any[] {
  try {
    if (!automation.conditions) return [];
    const parsed = typeof automation.conditions === 'string' ? JSON.parse(automation.conditions) : automation.conditions;
    return Array.isArray(parsed) ? parsed : [];
  } catch { return []; }
}

function getActions(automation: any): any[] {
  try {
    if (!automation.actions) return [];
    const parsed = typeof automation.actions === 'string' ? JSON.parse(automation.actions) : automation.actions;
    return Array.isArray(parsed) ? parsed : [];
  } catch { return []; }
}

function toggleEnabled(automation: any) {
  emit('toggle', automation);
}

function handleDelete(automation: any) {
  emit('delete', automation);
}
</script>
