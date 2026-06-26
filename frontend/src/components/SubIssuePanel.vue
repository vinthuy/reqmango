<template>
  <div class="sub-issue-panel">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-sm font-medium text-gray-700 flex items-center">
        <span>子工作项</span>
        <span v-if="subIssues.length > 0" class="ml-2 px-2 py-0.5 bg-gray-100 text-gray-600 text-xs rounded-full">
          {{ completedCount }}/{{ subIssues.length }}
        </span>
      </h3>
      <button
        @click="$emit('create')"
        class="text-xs text-indigo-600 hover:text-indigo-800 flex items-center space-x-1"
      >
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        <span>添加</span>
      </button>
    </div>

    <div v-if="subIssues.length === 0" class="text-center py-8">
      <svg class="h-10 w-10 text-gray-300 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <p class="mt-2 text-sm text-gray-500">暂无子工作项</p>
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="issue in subIssues"
        :key="issue.id"
        class="border border-gray-200 rounded-lg overflow-hidden"
      >
        <div
          @click="toggleExpand(issue.id)"
          class="flex items-center justify-between p-3 bg-gray-50 hover:bg-gray-100 cursor-pointer transition-colors"
        >
          <div class="flex items-center space-x-3 flex-1">
            <button class="shrink-0">
              <svg
                class="w-4 h-4 text-gray-400 transition-transform"
                :class="{ 'rotate-90': expandedIds.includes(issue.id) }"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </button>
            
            <div class="flex-1 min-w-0">
              <div class="flex items-center space-x-2">
                <span
                  v-if="issue.issue_type"
                  class="px-2 py-0.5 text-xs font-medium rounded"
                  :style="{ backgroundColor: issue.issue_type.color + '20', color: issue.issue_type.color }"
                >
                  {{ issue.issue_type.name }}
                </span>
                <span class="text-sm font-medium text-gray-900 truncate">#{{ issue.sequence_id }} {{ issue.name }}</span>
              </div>
              
              <div class="flex items-center space-x-3 mt-1">
                <span
                  v-if="issue.state_name"
                  class="text-xs px-2 py-0.5 rounded"
                  :class="getStateClass(issue.state_group)"
                >
                  {{ issue.state_name }}
                </span>
                
                <div v-if="issue.assignees && issue.assignees.length > 0" class="flex items-center">
                  <img
                    v-if="issue.assignees?.[0]?.avatar_url"
                    :src="issue.assignees[0].avatar_url"
                    :alt="issue.assignees[0].display_name"
                    class="w-4 h-4 rounded-full"
                  />
                  <div v-else class="w-4 h-4 rounded-full bg-indigo-100 flex items-center justify-center text-xs text-indigo-600">
                    {{ (issue.assignees?.[0]?.display_name || issue.assignees?.[0]?.email || '?')[0] }}
                  </div>
                </div>
                
                <span
                  v-if="issue.priority && issue.priority !== 'none'"
                  class="text-xs"
                  :class="getPriorityClass(issue.priority)"
                >
                  {{ getPriorityLabel(issue.priority) }}
                </span>
              </div>
            </div>
          </div>
          
          <div class="flex items-center space-x-2 shrink-0 ml-2">
            <button
              v-if="issue.sub_issues && issue.sub_issues.length > 0"
              class="text-xs text-gray-400 hover:text-indigo-600"
            >
              {{ issue.sub_issues?.length }} 个子项
            </button>
            <button
              @click.stop="$emit('edit', issue)"
              class="p-1 text-gray-400 hover:text-indigo-600"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
            </button>
          </div>
        </div>

        <div
          v-show="expandedIds.includes(issue.id)"
          class="p-4 bg-white border-t border-gray-100"
        >
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span class="text-gray-500">开始日期</span>
              <p class="text-gray-900 mt-1">{{ formatDate(issue.start_date) || '-' }}</p>
            </div>
            <div>
              <span class="text-gray-500">目标日期</span>
              <p class="text-gray-900 mt-1">{{ formatDate(issue.target_date) || '-' }}</p>
            </div>
            <div>
              <span class="text-gray-500">估算</span>
              <p class="text-gray-900 mt-1">{{ issue.estimate_point_id ? '已设置' : '-' }}</p>
            </div>
            <div>
              <span class="text-gray-500">创建时间</span>
              <p class="text-gray-900 mt-1">{{ formatDate(issue.created_at) }}</p>
            </div>
          </div>
          
          <div v-if="issue.description_html" class="mt-4">
            <span class="text-gray-500 text-xs">描述</span>
            <p class="text-sm text-gray-700 mt-1 line-clamp-2" v-html="issue.description_html"></p>
          </div>

          <div v-if="issue.sub_issues && issue.sub_issues.length > 0" class="mt-4 pt-4 border-t border-gray-100">
            <span class="text-xs text-gray-500">孙工作项</span>
            <div class="mt-2 space-y-1">
              <div
                v-for="sub in issue.sub_issues"
                :key="sub.id"
                class="flex items-center justify-between p-2 bg-gray-50 rounded text-xs"
              >
                <div class="flex items-center space-x-2">
                  <span class="text-gray-400">#{{ sub.sequence_id }}</span>
                  <span class="text-gray-700 truncate">{{ sub.name }}</span>
                </div>
                <span :class="getStateClass(sub.state_group)" class="px-1.5 py-0.5 rounded">{{ sub.state_name }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface SubIssue {
  id: number
  sequence_id: number
  name: string
  issue_type?: { name: string; color: string }
  state_name?: string
  state_group?: string
  priority?: string
  assignees?: Array<{ display_name?: string; email?: string; avatar_url?: string }>
  start_date?: string
  target_date?: string
  created_at?: string
  description_html?: string
  estimate_point_id?: number
  sub_issues?: SubIssue[]
}

const props = defineProps<{
  subIssues: SubIssue[]
}>()

defineEmits<{
  (e: 'create'): void
  (e: 'edit', issue: SubIssue): void
}>()

const expandedIds = ref<number[]>([])

const completedCount = computed(() => {
  return props.subIssues.filter(issue => issue.state_group === 'done').length
})

function toggleExpand(id: number) {
  const index = expandedIds.value.indexOf(id)
  if (index > -1) {
    expandedIds.value.splice(index, 1)
  } else {
    expandedIds.value.push(id)
  }
}

function getStateClass(stateGroup?: string): string {
  switch (stateGroup) {
    case 'done':
      return 'bg-green-100 text-green-700'
    case 'in_progress':
      return 'bg-blue-100 text-blue-700'
    case 'backlog':
      return 'bg-gray-100 text-gray-700'
    default:
      return 'bg-gray-100 text-gray-500'
  }
}

function getPriorityClass(priority: string): string {
  switch (priority) {
    case 'urgent':
      return 'text-red-600'
    case 'high':
      return 'text-orange-600'
    case 'medium':
      return 'text-yellow-600'
    case 'low':
      return 'text-green-600'
    default:
      return 'text-gray-400'
  }
}

function getPriorityLabel(priority: string): string {
  switch (priority) {
    case 'urgent':
      return '紧急'
    case 'high':
      return '高'
    case 'medium':
      return '中'
    case 'low':
      return '低'
    default:
      return '-'
  }
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' })
}
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>