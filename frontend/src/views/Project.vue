<template>
  <div class="project-page">
    <!-- 加载状态 -->
    <div v-if="loading" class="flex items-center justify-center h-64">
      <svg class="animate-spin h-8 w-8 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </div>

    <!-- 项目内容 -->
    <div v-else-if="project" class="space-y-6">
      <!-- 项目头部 -->
      <div class="bg-white rounded-lg border border-gray-200 p-6">
        <div class="flex items-start justify-between">
          <div class="flex items-center space-x-4">
            <div
              class="w-12 h-12 rounded-lg flex items-center justify-center text-white text-xl font-bold"
              :style="{ backgroundColor: project.color || '#6366f1' }"
            >
              {{ project.name?.charAt(0)?.toUpperCase() || 'P' }}
            </div>
            <div>
              <h1 class="text-xl font-semibold text-gray-900">{{ project.name }}</h1>
              <p v-if="project.description" class="text-sm text-gray-500 mt-1">{{ project.description }}</p>
            </div>
          </div>

          <div class="flex items-center space-x-2">
            <button
              @click="$emit('settings')"
              class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              设置
            </button>
          </div>
        </div>

        <!-- 标签页 -->
        <div class="mt-6 border-b border-gray-200">
          <nav class="-mb-px flex space-x-4">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              @click="activeTab = tab.id"
              class="py-2 px-1 border-b-2 text-sm font-medium transition-colors"
              :class="activeTab === tab.id
                ? 'border-indigo-500 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
            >
              {{ tab.name }}
            </button>
          </nav>
        </div>
      </div>

      <!-- 标签页内容 -->
      <div v-if="activeTab === 'issues'">
        <IssueList
          :project-id="projectId"
          :workspace-id="workspaceId"
          @create="showIssueModal = true"
          @select="selectIssue"
        />
      </div>

      <div v-if="activeTab === 'cycles'">
        <CycleList
          :project-id="projectId"
          :workspace-id="workspaceId"
          @create="showCycleModal = true"
          @select="selectCycle"
        />
      </div>

      <div v-if="activeTab === 'modules'">
        <ModuleList
          :project-id="projectId"
          :workspace-id="workspaceId"
          @create="showModuleModal = true"
          @select="selectModule"
        />
      </div>

      <div v-if="activeTab === 'workflow'">
        <WorkflowRuleList
          :project-id="projectId"
          :workspace-id="workspaceId"
          @create="showRuleForm = true"
          @open-templates="showTemplates = true"
        />
      </div>

      <div v-if="activeTab === 'custom-fields'">
        <CustomFieldList
          :project-id="projectId"
          @create="showFieldForm = true"
        />
      </div>

      <div v-if="activeTab === 'estimate-points'">
        <EstimatePointManager
          :project-id="projectId"
          @create="showEstimateForm = true"
        />
      </div>

      <div v-if="activeTab === 'members'">
        <ProjectMemberList
          :project-id="projectId"
          :workspace-id="workspaceId"
        />
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="text-center py-12">
      <svg class="h-12 w-12 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <p class="mt-2 text-gray-500">项目不存在或加载失败</p>
      <a href="/" class="mt-4 text-indigo-600 hover:text-indigo-800 text-sm">返回首页</a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import projectApi from '@/api/project'
import type { ProjectResponse } from '@/types/project'

// Components
import IssueList from '@/components/IssueList.vue'
import CycleList from '@/components/CycleList.vue'
import ModuleList from '@/components/ModuleList.vue'
import WorkflowRuleList from '@/components/WorkflowRuleList.vue'
import CustomFieldList from '@/components/CustomFieldList.vue'
import EstimatePointManager from '@/components/EstimatePointManager.vue'
import ProjectMemberList from '@/components/ProjectMemberList.vue'

// Props
const props = defineProps<{
  projectId: string
  workspaceId: string
}>()

// Emits
defineEmits<{
  (e: 'settings'): void
}>()

// State
const project = ref<ProjectResponse | null>(null)
const loading = ref(false)
const activeTab = ref('issues')

// Tabs
const tabs = [
  { id: 'issues', name: '工作项' },
  { id: 'cycles', name: '周期' },
  { id: 'modules', name: '模块' },
  { id: 'workflow', name: '自动化' },
  { id: 'custom-fields', name: '自定义字段' },
  { id: 'estimate-points', name: '估算点' },
  { id: 'members', name: '成员' }
]

// Modals
const showIssueModal = ref(false)
const showCycleModal = ref(false)
const showModuleModal = ref(false)
const showRuleForm = ref(false)
const showTemplates = ref(false)
const showFieldForm = ref(false)
const showEstimateForm = ref(false)

// Load project
onMounted(() => {
  loadProject()
})

async function loadProject() {
  loading.value = true
  try {
    project.value = await projectApi.getProject(props.projectId, props.workspaceId)
  } catch (error) {
    console.error('Failed to load project:', error)
  } finally {
    loading.value = false
  }
}

// Selectors
function selectIssue(issue: any) {
  console.log('Selected issue:', issue)
}

function selectCycle(cycle: any) {
  console.log('Selected cycle:', cycle)
}

function selectModule(module: any) {
  console.log('Selected module:', module)
}
</script>

<style scoped>
.project-page {
  @apply p-6;
}
</style>