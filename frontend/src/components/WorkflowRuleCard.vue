<template>
  <div class="workflow-rule-card bg-white border border-gray-200 rounded-lg p-4 hover:border-gray-300 transition-colors">
    <div class="flex items-start justify-between">
      <!-- 左侧：规则信息 -->
      <div class="flex items-start space-x-3 flex-1">
        <!-- 启用/禁用开关 -->
        <button
          @click="$emit('toggle')"
          class="mt-1 relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
          :class="rule.is_enabled ? 'bg-indigo-600' : 'bg-gray-200'"
        >
          <span
            class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
            :class="rule.is_enabled ? 'translate-x-5' : 'translate-x-0'"
          />
        </button>

        <!-- 规则详情 -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center space-x-2">
            <h4 class="text-sm font-medium text-gray-900">{{ rule.name }}</h4>
            <span
              class="px-2 py-0.5 text-xs rounded-full"
              :class="rule.is_enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'"
            >
              {{ rule.is_enabled ? '已启用' : '已禁用' }}
            </span>
          </div>

          <p v-if="rule.description" class="text-sm text-gray-500 mt-0.5 line-clamp-1">
            {{ rule.description }}
          </p>

          <!-- 触发器信息 -->
          <div class="mt-2 flex items-center space-x-2">
            <span class="inline-flex items-center px-2 py-0.5 text-xs font-medium bg-blue-100 text-blue-700 rounded">
              <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              {{ getTriggerName(rule.trigger.type) }}
            </span>

            <!-- 条件数量 -->
            <span v-if="rule.conditions && rule.conditions.length > 0" class="text-xs text-gray-500">
              {{ rule.conditions.length }} 个条件
            </span>

            <!-- 动作数量 -->
            <span class="text-xs text-gray-500">
              {{ rule.actions?.length || 0 }} 个动作
            </span>
          </div>
        </div>
      </div>

      <!-- 右侧：操作 -->
      <div class="flex items-center space-x-2 ml-4">
        <!-- 执行统计 -->
        <div v-if="rule.execution_count > 0" class="text-right mr-2">
          <div class="text-xs text-gray-500">执行 {{ rule.execution_count }} 次</div>
          <div v-if="rule.last_executed_at" class="text-xs text-gray-400">
            {{ formatDate(rule.last_executed_at) }}
          </div>
        </div>

        <!-- 查看日志 -->
        <button
          @click="$emit('view-logs')"
          class="p-1.5 text-gray-400 hover:text-gray-600 rounded"
          title="查看日志"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
        </button>

        <!-- 编辑 -->
        <button
          @click="$emit('edit')"
          class="p-1.5 text-gray-400 hover:text-indigo-600 rounded"
          title="编辑"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
        </button>

        <!-- 删除 -->
        <button
          @click="$emit('delete')"
          class="p-1.5 text-gray-400 hover:text-red-600 rounded"
          title="删除"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 条件和动作预览 -->
    <div v-if="showDetails" class="mt-4 pt-4 border-t border-gray-100">
      <!-- 条件 -->
      <div v-if="rule.conditions && rule.conditions.length > 0" class="mb-3">
        <h5 class="text-xs font-medium text-gray-500 mb-1">条件</h5>
        <div class="space-y-1">
          <div
            v-for="(condition, index) in rule.conditions"
            :key="index"
            class="text-xs text-gray-600 flex items-center space-x-1"
          >
            <span v-if="index > 0" class="text-gray-400">且</span>
            <span class="font-medium">{{ condition.field }}</span>
            <span class="text-gray-400">{{ getOperatorName(condition.operator) }}</span>
            <span v-if="condition.value !== undefined" class="text-indigo-600">{{ condition.value }}</span>
          </div>
        </div>
      </div>

      <!-- 动作 -->
      <div v-if="rule.actions && rule.actions.length > 0">
        <h5 class="text-xs font-medium text-gray-500 mb-1">动作</h5>
        <div class="space-y-1">
          <div
            v-for="(action, index) in rule.actions"
            :key="index"
            class="text-xs text-gray-600"
          >
            {{ getActionName(action.type) }}
            <span v-if="action.field" class="text-indigo-600">{{ action.field }}</span>
            <span v-if="action.value !== undefined" class="text-gray-400"> → {{ action.value }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 展开/收起详情 -->
    <button
      @click="showDetails = !showDetails"
      class="mt-3 text-xs text-indigo-600 hover:text-indigo-800 flex items-center space-x-1"
    >
      <span>{{ showDetails ? '收起' : '查看' }}详情</span>
      <svg
        class="w-3 h-3 transition-transform"
        :class="{ 'rotate-180': showDetails }"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { AutomationRule, TriggerTypeEnum, ConditionOperatorEnum, ActionTypeEnum } from '@/types/workflow'
import { getTriggerDisplayName, getActionDisplayName, getOperatorDisplayName } from '@/types/workflow'

// Props
defineProps<{
  rule: AutomationRule
}>()

// Emits
defineEmits<{
  (e: 'toggle'): void
  (e: 'edit'): void
  (e: 'delete'): void
  (e: 'view-logs'): void
}>()

// State
const showDetails = ref(false)

// Get trigger name
function getTriggerName(type: string): string {
  return getTriggerDisplayName(type as TriggerTypeEnum) || type
}

// Get action name
function getActionName(type: string): string {
  return getActionDisplayName(type as ActionTypeEnum) || type
}

// Get operator name
function getOperatorName(operator: string): string {
  return getOperatorDisplayName(operator as ConditionOperatorEnum) || operator
}

// Format date
function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${date.getMinutes()}`
}
</script>

<style scoped>
.workflow-rule-card:hover {
  @apply shadow-sm;
}

.line-clamp-1 {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>