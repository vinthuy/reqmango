<template>
  <div class="issue-detail-page min-h-screen bg-gray-50">
    <!-- 头部导航 -->
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <button @click="goBack" class="text-gray-500 hover:text-gray-700">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1 class="text-lg font-semibold text-gray-900">
            {{ issue?.issue_type?.name || '工作项' }} #{{ issue?.sequence_id }}
          </h1>
        </div>
        <div class="flex items-center space-x-3">
          <button
            @click="saveIssue"
            :disabled="saving"
            class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
          >
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="max-w-7xl mx-auto px-6 py-6">
      <div class="grid grid-cols-3 gap-6">
        <!-- 左侧：主要内容 -->
        <div class="col-span-2 space-y-6">
          <!-- 标题 -->
          <div class="bg-white rounded-lg shadow-sm p-6">
            <input
              v-model="issueForm.name"
              type="text"
              placeholder="工作项标题"
              class="w-full text-xl font-semibold border-0 focus:outline-none focus:ring-0"
            />
          </div>

          <!-- 描述 -->
          <div class="bg-white rounded-lg shadow-sm p-6">
            <h3 class="text-sm font-medium text-gray-700 mb-3">描述</h3>
            <RichTextEditor
              v-model="issueForm.description"
              placeholder="添加描述..."
            />
          </div>

          <!-- 自定义字段区域 -->
          <div class="bg-white rounded-lg shadow-sm p-6">
            <h3 class="text-sm font-medium text-gray-700 mb-4">自定义属性</h3>
            <CustomFieldManager
              ref="customFieldManagerRef"
              :workspace-id="workspaceId"
              :project-id="projectId"
              :issue-id="issueId"
              :issue-type-id="issue?.issue_type?.id"
              mode="display"
              :members="projectMembers"
              @update:values="handleValuesUpdate"
            />
          </div>

          <!-- 子工作项 -->
          <div class="bg-white rounded-lg shadow-sm p-6">
            <h3 class="text-sm font-medium text-gray-700 mb-3">子工作项</h3>
            <div v-if="subIssues.length > 0" class="space-y-2">
              <div
                v-for="sub in subIssues"
                :key="sub.id"
                class="flex items-center justify-between p-2 border border-gray-200 rounded"
              >
                <span class="text-sm">{{ sub.name }}</span>
                <span class="text-xs text-gray-500">#{{ sub.sequence_id }}</span>
              </div>
            </div>
            <button
              @click="createSubIssue"
              class="text-sm text-indigo-600 hover:text-indigo-800"
            >
              + 添加子工作项
            </button>
          </div>
        </div>

        <!-- 右侧：属性面板 -->
        <div class="space-y-6">
          <!-- 状态 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">状态</label>
            <select
              v-model="issueForm.state_id"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option v-for="state in states" :key="state.id" :value="state.id">
                {{ state.name }}
              </option>
            </select>
          </div>

          <!-- 类型 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">类型</label>
            <select
              v-model="issueForm.type_id"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">未设置</option>
              <option v-for="it in issueTypes" :key="it.id" :value="it.id">
                {{ it.name }}
              </option>
            </select>
          </div>

          <!-- 优先级 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">优先级</label>
            <select
              v-model="issueForm.priority"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="urgent">紧急</option>
              <option value="high">高</option>
              <option value="medium">中</option>
              <option value="low">低</option>
              <option value="none">无</option>
            </select>
          </div>

          <!-- 负责人 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">负责人</label>
            <select
              v-model="issueForm.assignee_id"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">未分配</option>
              <option v-for="member in projectMembers" :key="member.id" :value="member.id">
                {{ member.display_name || member.email }}
              </option>
            </select>
          </div>

          <!-- 周期 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">周期</label>
            <select
              v-model="issueForm.cycle_id"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">无周期</option>
              <option v-for="cycle in cycles" :key="cycle.id" :value="cycle.id">
                {{ cycle.name }}
              </option>
            </select>
          </div>

          <!-- 模块 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">模块</label>
            <select
              v-model="issueForm.module_id"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">无模块</option>
              <option v-for="module in modules" :key="module.id" :value="module.id">
                {{ module.name }}
              </option>
            </select>
          </div>

          <!-- 开始日期 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">开始日期</label>
            <input
              v-model="issueForm.start_date"
              type="date"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          <!-- 目标日期 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">目标日期</label>
            <input
              v-model="issueForm.target_date"
              type="date"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          <!-- 标签 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">标签</label>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="label in issueForm.labels"
                :key="label"
                class="px-2 py-1 text-xs rounded bg-indigo-100 text-indigo-800"
              >
                {{ label }}
              </span>
            </div>
            <input
              v-model="newLabel"
              @keyup.enter="addLabel"
              type="text"
              placeholder="添加标签"
              class="w-full mt-2 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CustomFieldManager from '@/components/CustomFieldManager.vue'
import RichTextEditor from '@/components/RichTextEditor.vue'
import * as issueApi from '@/api/issue'
import * as issueTypeApi from '@/api/issue-type'
import * as stateApi from '@/api/project-settings'
import * as cycleApi from '@/api/cycle'
import * as moduleApi from '@/api/module'
import projectApi from '@/api/project'

// Route params
const route = useRoute()
const router = useRouter()
const workspaceId = parseInt(route.params.workspaceId as string, 10)
const projectId = parseInt(route.params.projectId as string, 10)
const issueId = parseInt(route.params.issueId as string, 10)

// State
const issue = ref<any>(null)
const states = ref<any[]>([])
const cycles = ref<any[]>([])
const modules = ref<any[]>([])
const projectMembers = ref<any[]>([])
const issueTypes = ref<any[]>([])
const subIssues = ref<any[]>([])
const saving = ref(false)
const newLabel = ref('')
const customFieldManagerRef = ref<InstanceType<typeof CustomFieldManager> | null>(null)
const customFieldValues = ref<Record<string, any>>({})

// Form
const issueForm = ref({
  name: '',
  description: '',
  state_id: '',
  priority: 'none',
  type_id: '' as number | string,
  assignee_id: '',
  cycle_id: '',
  module_id: '',
  start_date: '',
  target_date: '',
  labels: [] as string[]
})

// Load issue data
onMounted(async () => {
  await Promise.all([
    loadIssue(),
    loadStates(),
    loadCycles(),
    loadModules(),
    loadProjectMembers(),
    loadIssueTypes()
  ])
})

// Load issue
async function loadIssue() {
  try {
    const issueData = await issueApi.getIssue(issueId)
    issue.value = issueData

    // Populate form
    issueForm.value = {
      name: issue.value.name,
      description: issue.value.description_html || '',
      state_id: issue.value.state_id,
      priority: issue.value.priority,
      type_id: issue.value.issue_type?.id || '',
      assignee_id: issue.value.assignees?.[0]?.id || '',
      cycle_id: issue.value.cycle_id || '',
      module_id: issue.value.module_ids?.[0] || '',
      start_date: issue.value.start_date?.split('T')[0] || '',
      target_date: issue.value.target_date?.split('T')[0] || '',
      labels: issue.value.label_details?.map((l: any) => l.name) || []
    }
    
    // Load sub issues
    if (issue.value.sub_issues) {
      subIssues.value = issue.value.sub_issues
    }
  } catch (error) {
    console.error('Failed to load issue:', error)
  }
}

// Load states
async function loadStates() {
  try {
    const data = await stateApi.listStates(projectId)
    states.value = data
  } catch (error) {
    console.error('Failed to load states:', error)
  }
}

// Load cycles
async function loadCycles() {
  try {
    const data = await cycleApi.listCycles(projectId)
    cycles.value = data.items || data
  } catch (error) {
    console.error('Failed to load cycles:', error)
  }
}

// Load modules
async function loadModules() {
  try {
    const data = await moduleApi.listModules(projectId, workspaceId)
    modules.value = data
  } catch (error) {
    console.error('Failed to load modules:', error)
  }
}

// Load issue types
async function loadIssueTypes() {
  try {
    const types = await issueTypeApi.getIssueTypes(workspaceId, projectId)
    issueTypes.value = types
  } catch (error) {
    console.error('Failed to load issue types:', error)
  }
}

// Load project members
async function loadProjectMembers() {
  try {
    const data = await projectApi.listProjectMembers(projectId)
    projectMembers.value = data.map((m: any) => m.user || m)
  } catch (error) {
    console.error('Failed to load project members:', error)
  }
}

// Handle custom field values update
function handleValuesUpdate(values: Record<string, any>) {
  customFieldValues.value = values
}

// Save issue
async function saveIssue() {
  saving.value = true
  try {
    // Build update payload
    const data: any = {
      name: issueForm.value.name,
      priority: issueForm.value.priority,
      description_html: issueForm.value.description || undefined,
    }
    if (issueForm.value.state_id) data.state_id = Number(issueForm.value.state_id)
    if (issueForm.value.type_id) data.type_id = Number(issueForm.value.type_id)
    if (issueForm.value.assignee_id) data.assignee_ids = [Number(issueForm.value.assignee_id)]
    if (issueForm.value.cycle_id) data.cycle_id = Number(issueForm.value.cycle_id)
    if (issueForm.value.start_date) data.start_date = issueForm.value.start_date + 'T00:00:00Z'
    if (issueForm.value.target_date) data.target_date = issueForm.value.target_date + 'T00:00:00Z'

    await issueApi.updateIssue(issueId, data)

    // Save custom field values
    if (customFieldManagerRef.value) {
      await customFieldManagerRef.value.saveValues()
    }

    alert('保存成功')
  } catch (error) {
    console.error('Failed to save issue:', error)
    alert('保存失败')
  } finally {
    saving.value = false
  }
}

// Add label
function addLabel() {
  if (newLabel.value && !issueForm.value.labels.includes(newLabel.value)) {
    issueForm.value.labels.push(newLabel.value)
    newLabel.value = ''
  }
}

// Create sub issue
function createSubIssue() {
  // Navigate to create issue page with parent_id
  router.push({
    path: `/workspaces/${workspaceId}/projects/${projectId}/issues/new`,
    query: { parent_id: issueId }
  })
}

// Go back
function goBack() {
  router.back()
}
</script>

<style scoped>
.issue-detail-page {
  min-height: 100vh;
}
</style>