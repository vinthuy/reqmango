<template>
  <div class="project-page min-h-screen bg-gray-50">
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
            <button @click="goBack" class="text-gray-400 hover:text-gray-600">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <div
              class="w-12 h-12 rounded-lg flex items-center justify-center text-white text-xl font-bold"
              style="background-color: #6366f1"
            >
              {{ project.name?.charAt(0)?.toUpperCase() || 'P' }}
            </div>
            <div>
              <h1 class="text-xl font-semibold text-gray-900">{{ project.name }}</h1>
              <div class="flex items-center space-x-3 mt-1">
                <span class="text-xs text-gray-500">{{ project.identifier }}</span>
                <span v-if="project.description" class="text-sm text-gray-500">{{ project.description }}</span>
              </div>
            </div>
          </div>

        </div>

        <!-- 标签页导航 -->
        <div class="mt-6 border-b border-gray-200">
          <nav class="-mb-px flex space-x-6">
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
        <!-- 视图切换 -->
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-3">
            <div class="inline-flex bg-gray-100 rounded-lg p-0.5">
              <button
                @click="issueView = 'list'"
                class="px-3 py-1.5 text-sm rounded-md transition-colors"
                :class="issueView === 'list' ? 'bg-white shadow text-gray-900 font-medium' : 'text-gray-500 hover:text-gray-700'"
              >
                列表视图
              </button>
              <button
                @click="issueView = 'kanban'"
                class="px-3 py-1.5 text-sm rounded-md transition-colors"
                :class="issueView === 'kanban' ? 'bg-white shadow text-gray-900 font-medium' : 'text-gray-500 hover:text-gray-700'"
              >
                看板视图
              </button>
              <button
                @click="issueView = 'tree'"
                class="px-3 py-1.5 text-sm rounded-md transition-colors"
                :class="issueView === 'tree' ? 'bg-white shadow text-gray-900 font-medium' : 'text-gray-500 hover:text-gray-700'"
              >
                树形视图
              </button>
              <button
                @click="issueView = 'calendar'"
                class="px-3 py-1.5 text-sm rounded-md transition-colors"
                :class="issueView === 'calendar' ? 'bg-white shadow text-gray-900 font-medium' : 'text-gray-500 hover:text-gray-700'"
              >
                日历视图
              </button>
              <button
                @click="issueView = 'gantt'"
                class="px-3 py-1.5 text-sm rounded-md transition-colors"
                :class="issueView === 'gantt' ? 'bg-white shadow text-gray-900 font-medium' : 'text-gray-500 hover:text-gray-700'"
              >
                甘特图
              </button>
            </div>
            <SavedViewSelector
              :project-id="projectId"
              :view-type="(issueView as 'list' | 'kanban' | 'tree' | 'calendar' | 'gantt')"
              @select="handleViewSelect"
            />
          </div>
          <div class="flex items-center gap-2">
            <button
              @click="showAICreate = true"
              class="px-3 py-1.5 bg-gradient-to-r from-indigo-500 to-purple-600 text-white text-sm rounded-md hover:from-indigo-600 hover:to-purple-700 transition flex items-center gap-1"
            >
              🤖 AI Create
            </button>
            <button
              @click="router.push(`/workspaces/${workspaceId}/projects/${projectId}/issues/new?view=${issueView}`)"
              class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700"
            >
              新建工作项
            </button>
          </div>
        </div>

        <IssueList
          v-if="issueView === 'list'"
          :key="'list-' + issueRefreshKey"
          :project-id="projectId"
          :workspace-id="workspaceId"
          @select="openDetailPanel"
          @delete="handleDeleteIssue"
        />
        <IssueKanban
          v-else-if="issueView === 'kanban'"
          :key="'kanban-' + issueRefreshKey"
          :project-id="projectId"
          :workspace-id="workspaceId"
          @select="openDetailPanel"
        />
        <IssueTreeView
          v-else-if="issueView === 'tree'"
          :key="'tree-' + issueRefreshKey"
          :project-id="projectId"
          :workspace-id="workspaceId"
          @select="openDetailPanel"
        />
        <IssueCalendar
          v-else-if="issueView === 'calendar'"
          :project-id="projectId"
          :workspace-id="workspaceId"
          @select="openDetailPanel"
        />
        <IssueGantt
          v-else-if="issueView === 'gantt'"
          :project-id="projectId"
          :workspace-id="workspaceId"
          @select="openDetailPanel"
          @create="router.push(`/workspaces/${workspaceId}/projects/${projectId}/issues/new?view=tree`)"
        />
      </div>

      <!-- 侧滑详情面板 -->
      <IssueDetailPanel
        :issue-id="detailIssueId"
        :visible="detailPanelVisible"
        :workspace-id="workspaceId"
        :project-id="projectId"
        @close="detailPanelVisible = false"
        @delete="handleDetailDelete"
        @refresh="handleDetailRefresh"
      />

      <CycleDetailPanel
        :cycle="selectedCycle"
        :visible="cyclePanelVisible"
        @close="cyclePanelVisible = false"
      />

      <ModuleDetailPanel
        :module="selectedModule"
        :visible="modulePanelVisible"
        :project-id="projectId"
        :workspace-id="workspaceId"
        @close="modulePanelVisible = false"
        @edit="handleModuleEdit"
      />

      <ModuleFormModal
        :visible="moduleFormVisible"
        :edit-module="editingModule"
        :workspace-id="workspaceId"
        :project-id="projectId"
        @close="moduleFormVisible = false; editingModule = null"
        @saved="moduleFormVisible = false; editingModule = null; modulePanelVisible = false"
      />

      <div v-if="activeTab === 'cycles'">
        <CycleList
          :project-id="projectId"
          :workspace-id="workspaceId"
          @select="openCyclePanel"
          @create="goToCycleCreate"
        />
      </div>

      <div v-if="activeTab === 'modules'">
        <ModuleList
          :project-id="projectId"
          :workspace-id="workspaceId"
          @select="openModulePanel"
          @create="moduleFormVisible = true"
          @delete="handleModuleDelete"
        />
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="text-center py-12">
      <p class="text-gray-500">项目不存在或加载失败</p>
      <button @click="goBack" class="mt-4 text-indigo-600 hover:text-indigo-800 text-sm">返回</button>
    </div>
  </div>

  <!-- AI Chat Sidebar -->
  <AIChatSidebar
    :visible="showAIChat"
    :project-id="projectId"
    :workspace-id="workspaceId"
    :project-name="project?.name"
    @close="showAIChat = false"
  />
  <AICreateDialog
    :visible="showAICreate"
    :project-id="projectId"
    :workspace-id="workspaceId"
    @close="showAICreate = false"
    @created="issueRefreshKey++"
  />
  <CommandPalette
    :visible="showCommandPalette"
    :workspace-slug="slug"
    :project-id="projectId"
    :workspace-id="workspaceId"
    @close="showCommandPalette = false"
  />
  <button
    @click="showAIChat = true"
    class="fixed bottom-6 right-6 w-14 h-14 bg-gradient-to-r from-indigo-500 to-purple-600 text-white rounded-full shadow-lg hover:shadow-xl hover:scale-105 transition flex items-center justify-center text-2xl z-40"
    title="AI Assistant"
  >🤖</button>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { workspaceApi } from '@/api/workspace'
import { projectApi } from '@/api/project'
import { issueApi } from '@/api/issue'
import type { Workspace } from '@/types'
import type { ProjectResponse } from '@/types/project'
import { useConfirm } from '@/composables/useConfirm'

import IssueList from '@/components/IssueList.vue'
import IssueKanban from '@/components/IssueKanban.vue'
import IssueTreeView from '@/components/IssueTreeView.vue'
import IssueCalendar from '@/components/IssueCalendar.vue'
import IssueGantt from '@/components/IssueGantt.vue'
import IssueDetailPanel from '@/components/IssueDetailPanel.vue'
import CycleList from '@/components/CycleList.vue'
import CycleDetailPanel from '@/components/CycleDetailPanel.vue'
import ModuleList from '@/components/ModuleList.vue'
import ModuleDetailPanel from '@/components/ModuleDetailPanel.vue'
import ModuleFormModal from '@/components/ModuleFormModal.vue'
import SavedViewSelector from '@/components/SavedViewSelector.vue'
import AIChatSidebar from '@/components/AIChatSidebar.vue'
import AICreateDialog from '@/components/AICreateDialog.vue'
import CommandPalette from '@/components/CommandPalette.vue'
import type { SavedView } from '@/types/saved-view'
import type { CycleResponse } from '@/types/cycle'
import type { ModuleResponse } from '@/types/module'
import { useModuleStore } from '@/stores/module'

const route = useRoute()
const router = useRouter()

const workspace = ref<Workspace | null>(null)
const project = ref<ProjectResponse | null>(null)
const loading = ref(false)
const activeTab = ref((route.query.tab as string) || 'issues')
const issueView = ref((route.query.view as string) || 'list')
const issueRefreshKey = ref(0)

const detailIssueId = ref<number | null>(null)
const detailPanelVisible = ref(false)

const selectedCycle = ref<CycleResponse | null>(null)
const cyclePanelVisible = ref(false)

const selectedModule = ref<ModuleResponse | null>(null)
const modulePanelVisible = ref(false)
const moduleFormVisible = ref(false)
const editingModule = ref<ModuleResponse | null>(null)

// AI state
const showAIChat = ref(false)
const showAICreate = ref(false)
const showCommandPalette = ref(false)

watch(activeTab, (tab) => {
  if (tab === 'settings') {
    router.push(`/workspace/${route.params.slug}/project/${projectId.value}/settings`)
    return
  }
  if (tab === 'pages') {
    router.push(`/workspace/${route.params.slug}/project/${projectId.value}/pages`)
    return
  }
  detailPanelVisible.value = false
  cyclePanelVisible.value = false
  modulePanelVisible.value = false
})

function openDetailPanel(issue: any) {
  detailIssueId.value = issue.id
  detailPanelVisible.value = true
}

function handleDetailDelete(issue: any) {
  handleDeleteIssue(issue)
}

function handleDetailRefresh() {
  issueRefreshKey.value++
}

function openCyclePanel(cycle: CycleResponse) {
  selectedCycle.value = cycle
  cyclePanelVisible.value = true
}

function goToCycleCreate() {
  const slug = route.params.slug as string
  router.push(`/workspaces/${workspaceId.value}/projects/${projectId.value}/cycles/new?ws=${slug}`)
}

function openModulePanel(module: ModuleResponse | any) {
  selectedModule.value = module as ModuleResponse
  modulePanelVisible.value = true
}

function handleModuleEdit(module: ModuleResponse) {
  editingModule.value = module
  moduleFormVisible.value = true
}

const moduleStore = useModuleStore()
const { confirm } = useConfirm()

async function handleModuleDelete(module: ModuleResponse | any) {
  if (await confirm(`确定要删除模块 "${module.name}" 吗？`)) {
    try {
      await moduleStore.deleteModuleAction(module.id)
    } catch (err) {
      console.error('Failed to delete module:', err)
      alert('删除失败')
    }
  }
}

const workspaceId = ref(0)
const projectId = ref(0)
const slug = ref('')

const tabs = [
  { id: 'issues', name: '工作项管理' },
  { id: 'cycles', name: '周期' },
  { id: 'modules', name: '模块' },
  { id: 'pages', name: '文档' },
  { id: 'settings', name: '设置' },
]

function goBack() {
  router.push(`/workspace/${route.params.slug}`)
}

function handleViewSelect(view: SavedView) {
  if (view.view_type && (view.view_type === 'list' || view.view_type === 'kanban' || view.view_type === 'tree')) {
    issueView.value = view.view_type
  }
  // Future: apply filters, sort, columns, groupBy from the view
  // These would be passed as props to IssueList/IssueKanban
  console.log('View selected:', view.name, view)
}

async function handleDeleteIssue(issue: any) {
  if (await confirm(`确定要删除工作项 "${issue.name}" 吗？`)) {
    try {
      await issueApi.deleteIssue(issue.id)
      window.location.reload()
    } catch (err) {
      console.error('Failed to delete issue:', err)
      alert('删除失败')
    }
  }
}

onMounted(async () => {
  loading.value = true
  slug.value = route.params.slug as string
  const id = parseInt(route.params.id as string)

  try {
    workspace.value = await workspaceApi.getBySlug(slug.value)
    workspaceId.value = workspace.value.id
    projectId.value = id
    project.value = await projectApi.getProject(id)
  } catch (err) {
    console.error('Failed to load project:', err)
  } finally {
    loading.value = false
  }
  document.addEventListener('keydown', (e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'j') {
      e.preventDefault()
      showAIChat.value = !showAIChat.value
    }
    if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
      e.preventDefault()
      showCommandPalette.value = !showCommandPalette.value
    }
  })
})
</script>

<style scoped>
.project-page {
  @apply p-6;
}
</style>
