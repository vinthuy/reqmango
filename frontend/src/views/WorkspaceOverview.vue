<template>
  <div class="workspace-overview p-6">
    <h1 class="text-xl font-bold text-gray-800 dark:text-gray-100 mb-6">工作空间总览</h1>

    <!-- Project summary cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
      <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
        <div class="text-2xl font-bold text-indigo-600">{{ projects.length }}</div>
        <div class="text-sm text-gray-500">项目数</div>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
        <div class="text-2xl font-bold text-green-600">{{ totalIssues }}</div>
        <div class="text-sm text-gray-500">工作项总数</div>
      </div>
      <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
        <div class="text-2xl font-bold text-orange-600">{{ issues.length }}</div>
        <div class="text-sm text-gray-500">已加载</div>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 mb-4">
      <div class="flex flex-wrap gap-3 items-center">
        <select v-model="filters.projectId" @change="loadIssues" class="px-3 py-1.5 border border-gray-300 rounded-md text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200">
          <option :value="0">所有项目</option>
          <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <select v-model="filters.priority" @change="loadIssues" class="px-3 py-1.5 border border-gray-300 rounded-md text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200">
          <option value="">所有优先级</option>
          <option value="urgent">紧急</option><option value="high">高</option><option value="medium">中</option><option value="low">低</option><option value="none">无</option>
        </select>
        <input v-model="filters.search" @keydown.enter="loadIssues" placeholder="搜索..." class="px-3 py-1.5 border border-gray-300 rounded-md text-sm flex-1 max-w-xs dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200" />
        <button @click="loadIssues" class="px-4 py-1.5 text-sm bg-indigo-600 text-white rounded-md hover:bg-indigo-700">查询</button>
      </div>
    </div>

    <!-- Issues table -->
    <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
      <div v-if="loading" class="text-center py-8 text-gray-400 text-sm">加载中...</div>
      <table v-else class="w-full text-sm">
        <thead class="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th class="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">ID</th>
            <th class="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">名称</th>
            <th class="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">项目</th>
            <th class="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">状态</th>
            <th class="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">优先级</th>
            <th class="text-left px-4 py-2 text-gray-600 dark:text-gray-300 font-medium">负责人</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="issue in issues" :key="issue.id" class="border-t border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-750 cursor-pointer" @click="openIssue(issue)">
            <td class="px-4 py-2 text-gray-400 font-mono text-xs">{{ issue.project?.identifier || '?' }}-{{ issue.sequence_id }}</td>
            <td class="px-4 py-2 text-gray-800 dark:text-gray-200 max-w-xs truncate">{{ issue.name }}</td>
            <td class="px-4 py-2 text-gray-500 text-xs">{{ issue.project?.name || '-' }}</td>
            <td class="px-4 py-2">
              <span class="inline-flex items-center gap-1 text-xs">
                <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: stateById[issue.state_id]?.color || '#9ca3af' }"></span>
                {{ stateById[issue.state_id]?.name || '-' }}
              </span>
            </td>
            <td class="px-4 py-2">
              <span :class="priorityClass(issue.priority)" class="text-xs px-1.5 py-0.5 rounded">{{ priorityLabel(issue.priority) }}</span>
            </td>
            <td class="px-4 py-2 text-gray-500 text-xs">{{ (issue.assignees || []).map((a: any) => a.display_name || a.username).join(', ') || '-' }}</td>
          </tr>
          <tr v-if="issues.length === 0">
            <td colspan="6" class="text-center py-8 text-gray-400 text-sm">暂无工作项</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import issueApi from '@/api/issue'
import { workspaceApi } from '@/api/workspace'
import { projectApi } from '@/api/project'
import api from '@/api'

const route = useRoute()
const router = useRouter()

const workspaceSlug = route.params.slug as string
const projects = ref<any[]>([])
const issues = ref<any[]>([])
const totalIssues = ref(0)
const loading = ref(false)
const stateById = ref<Record<number, any>>({})

const filters = ref({ projectId: 0, priority: '', search: '' })

const workspaceId = ref(0)

function priorityClass(p: string) {
  const m: Record<string, string> = { urgent: 'bg-red-100 text-red-700', high: 'bg-orange-100 text-orange-700', medium: 'bg-yellow-100 text-yellow-700', low: 'bg-green-100 text-green-700', none: 'bg-gray-100 text-gray-500' }
  return m[p] || m.none
}
function priorityLabel(p: string) {
  const m: Record<string, string> = { urgent: '紧急', high: '高', medium: '中', low: '低', none: '无' }
  return m[p] || p
}

async function loadWorkspace() {
  try {
    const ws = await workspaceApi.getBySlug(workspaceSlug)
    workspaceId.value = ws.id
  } catch { return }
  try {
    projects.value = await projectApi.listProjects(workspaceId.value)
    for (const p of projects.value) {
      const r = await api.get(`/projects/${p.id}/settings/states`)
      for (const s of r.data) { stateById.value[s.id] = s }
    }
  } catch { /* */ }
  loadIssues()
}

async function loadIssues() {
  loading.value = true
  try {
    const params: any = { limit: 200 }
    if (filters.value.projectId) params.project_id = filters.value.projectId
    else params.workspace_id = workspaceId.value
    if (filters.value.priority) params.priority = filters.value.priority
    if (filters.value.search) params.search = filters.value.search
    const result = await issueApi.listIssues(filters.value.projectId || workspaceId.value, workspaceId.value, params)
    issues.value = result.items
    totalIssues.value = issues.value.length
  } catch { issues.value = [] }
  loading.value = false
}

function openIssue(issue: any) {
  const pid = issue.project_id || issue.project?.id
  if (pid) router.push(`/projects/${pid}/issues/${issue.id}`)
}

onMounted(loadWorkspace)
</script>
