<template>
  <div
    class="issue-card bg-white border border-gray-200 rounded-lg p-4 hover:border-indigo-300 hover:shadow-sm cursor-pointer transition-all"
    @click="$emit('click')"
  >
    <div class="flex items-start justify-between">
      <!-- 左侧：优先级和标题 -->
      <div class="flex items-start space-x-3 flex-1 min-w-0">
        <!-- 优先级图标 -->
        <div class="shrink-0 mt-0.5">
          <span>
            {{ getPriorityIcon(issue.priority) }}
          </span>
        </div>

        <!-- 标题和序列号 -->
        <div class="min-w-0 flex-1">
          <div class="flex items-center space-x-2">
            <span class="text-xs text-gray-500">#{{ issue.sequence_id }}</span>
            <span v-if="issue.state_name" class="px-1.5 py-0.5 text-xs rounded" :class="getStateClass(issue.state_group)">
              {{ issue.state_name }}
            </span>
            <span v-if="issue.issue_type" class="inline-flex items-center space-x-0.5 px-1.5 py-0.5 text-xs rounded bg-gray-100 text-gray-600">
              <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: issue.issue_type.color }"></span>
              <span>{{ issue.issue_type.name }}</span>
            </span>
          </div>
          <h3 class="text-sm font-medium text-gray-900 mt-0.5 truncate">
            {{ issue.name }}
          </h3>
        </div>
      </div>

      <!-- 右侧：标签和操作 -->
      <div class="flex items-center space-x-2 ml-4">
        <!-- 标签 -->
        <div v-if="issue.labels && issue.labels.length > 0" class="flex items-center space-x-1">
          <span
            v-for="label in issue.labels.slice(0, 2)"
            :key="label.id"
            class="px-1.5 py-0.5 text-xs rounded"
            :style="{ backgroundColor: label.color + '20', color: label.color }"
          >
            {{ label.name }}
          </span>
          <span v-if="issue.labels.length > 2" class="text-xs text-gray-400">
            +{{ issue.labels.length - 2 }}
          </span>
        </div>

        <!-- 负责人头像 -->
        <div v-if="issue.assignees && issue.assignees.length > 0" class="flex -space-x-1">
          <div
            v-for="(assignee, index) in issue.assignees.slice(0, 2)"
            :key="assignee.id"
            class="w-6 h-6 rounded-full border-2 border-white flex items-center justify-center text-xs font-medium"
            :class="getAvatarClass(index)"
            :title="assignee.display_name || assignee.username"
          >
            {{ getInitials(assignee.display_name || assignee.username) }}
          </div>
          <div
            v-if="issue.assignees.length > 2"
            class="w-6 h-6 rounded-full border-2 border-white bg-gray-200 flex items-center justify-center text-xs"
          >
            +{{ issue.assignees.length - 2 }}
          </div>
        </div>
        <div v-else class="w-6 h-6 rounded-full border-2 border-dashed border-gray-300 flex items-center justify-center">
          <svg class="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
        </div>

        <!-- 周期标签 -->
        <span v-if="issue.cycle" class="px-1.5 py-0.5 text-xs bg-blue-100 text-blue-800 rounded">
          {{ issue.cycle.name }}
        </span>

        <!-- 操作菜单 -->
        <div class="relative" @click.stop>
          <button
            @click="showMenu = !showMenu"
            class="p-1 text-gray-400 hover:text-gray-600 rounded"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
            </svg>
          </button>

          <!-- 下拉菜单 -->
          <div
            v-if="showMenu"
            class="absolute right-0 mt-1 w-32 bg-white border border-gray-200 rounded-md shadow-lg z-10"
          >
            <button
              @click="$emit('click'); showMenu = false"
              class="w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-50"
            >
              {{ t('cycleCard.viewDetails') }}
            </button>
            <button
              @click="$emit('archive', issue); showMenu = false"
              class="w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-50"
            >
              {{ t('projectDetail.archive') }}
            </button>
            <button
              @click="$emit('delete', issue); showMenu = false"
              class="w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50"
            >
              {{ t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部：日期和模块 -->
    <div class="mt-3 flex items-center justify-between text-xs text-gray-500">
      <div class="flex items-center space-x-4">
        <span v-if="issue.start_date">
          <svg class="w-3 h-3 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          {{ formatDate(issue.start_date) }}
        </span>
        <span v-if="issue.target_date" :class="{ 'text-red-500': isOverdue(issue.target_date) }">
          <svg class="w-3 h-3 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ formatDate(issue.target_date) }}
        </span>
      </div>
      <span v-if="issue.module_ids && issue.module_ids.length > 0">
        <svg class="w-3 h-3 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        {{ issue.module_ids.length }} {{ t('projectDetail.modules') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
const { t } = useI18n()

// Props
defineProps<{
  issue: {
    id: string
    name: string
    sequence_id: number
    priority: string
    state_name?: string
    state_group?: string
    labels?: Array<{ id: string; name: string; color: string }>
    assignees?: Array<{ id: string; display_name?: string; username: string }>
    cycle?: { id: string; name: string }
    start_date?: string
    target_date?: string
    module_ids?: string[]
    issue_type?: { id: number; name: string; color: string; icon: string }
  }
}>()

// Emits
defineEmits<{
  (e: 'click'): void
  (e: 'archive', issue: any): void
  (e: 'delete', issue: any): void
}>()

// State
const showMenu = ref(false)

// Priority icons
function getPriorityIcon(priority: string): string {
  const icons: Record<string, string> = {
    urgent: '🔴',
    high: '🟠',
    medium: '🟡',
    low: '🟢',
    none: '⚪'
  }
  return icons[priority] || '⚪'
}



// State classes
function getStateClass(group?: string): string {
  const classes: Record<string, string> = {
    backlog: 'bg-gray-100 text-gray-600',
    todo: 'bg-blue-100 text-blue-700',
    in_progress: 'bg-yellow-100 text-yellow-700',
    done: 'bg-green-100 text-green-700',
    cancelled: 'bg-red-100 text-red-700'
  }
  return classes[group || ''] || 'bg-gray-100 text-gray-600'
}

// Avatar classes
function getAvatarClass(index: number): string {
  const classes = [
    'bg-indigo-500 text-white',
    'bg-green-500 text-white',
    'bg-yellow-500 text-white',
    'bg-red-500 text-white'
  ]
  return classes[index % classes.length]
}

// Get initials
function getInitials(name: string): string {
  return name
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

// Format date
function formatDate(dateStr?: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()}`
}

// Check if overdue
function isOverdue(dateStr?: string): boolean {
  if (!dateStr) return false
  return new Date(dateStr) < new Date()
}
</script>

<style scoped>
.issue-card:hover {
  transform: translateY(-1px);
}
</style>