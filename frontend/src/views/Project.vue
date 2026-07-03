<template>
  <div class="project-page min-h-screen bg-gray-50 dark:bg-gray-900">
    <!-- 加载状态 -->
    <div v-if="loading" class="flex items-center justify-center h-64">
      <svg class="animate-spin h-8 w-8 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </div>

    <!-- 项目内容 -->
    <div v-else-if="project" class="space-y-4">
      <!-- 项目头部 - clean top bar -->
      <div class="px-2">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-9 h-9 rounded-lg flex items-center justify-center text-white text-sm font-bold shrink-0" :style="{ backgroundColor: project.color || '#6366f1' }">
              {{ project.name?.charAt(0)?.toUpperCase() || 'P' }}
            </div>
            <div class="min-w-0">
              <h1 class="text-lg font-semibold text-gray-900 dark:text-gray-100 truncate">{{ project.name }}</h1>
              <div class="flex items-center gap-2 mt-0.5">
                <span class="text-xs text-gray-400 dark:text-gray-500 font-mono">{{ project.identifier }}</span>
                <span v-if="latestUpdateStatus" class="text-xs px-2 py-0.5 rounded-full" :class="getUpdateStatusColor(latestUpdateStatus)">{{ getUpdateStatusLabel(latestUpdateStatus) }}</span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <button @click="showAICreate = true" class="px-3 py-1.5 bg-gradient-to-r from-indigo-500 to-purple-600 text-white text-xs font-medium rounded-md hover:from-indigo-600 hover:to-purple-700 transition shadow-sm">🤖 {{ t('project.aiCreate') }}</button>
            <button @click="router.push(`/workspaces/${workspaceId}/projects/${projectId}/issues/new?view=${issueView}`)" class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 text-white text-xs font-medium rounded-md hover:bg-indigo-700 transition shadow-sm">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
              {{ t('project.create') }}
            </button>
          </div>
        </div>

        <!-- 标签页导航 - horizontal pills -->
        <nav class="flex items-center gap-1 border-b border-gray-200 dark:border-gray-800 pb-2">
          <button v-for="tab in computedTabs" :key="tab.id" @click="activeTab = tab.id"
            class="px-3.5 py-1.5 text-sm font-medium rounded-md transition-colors"
            :class="activeTab === tab.id ? 'bg-gray-200/70 dark:bg-gray-800 text-gray-900 dark:text-gray-100' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800/50'">
            {{ tab.name }}
          </button>
          <button @click="showPageConfig = true" class="px-2 py-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded hover:bg-gray-100 dark:hover:bg-gray-800 transition text-sm" :title="t('project.pageConfig')">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          </button>
        </nav>
      </div>

      <!-- 标签页内容 -->
      <div v-if="activeTab === 'issues'">
        <FilterBar
          :project-id="projectId"
          :workspace-id="workspaceId"
          :current-view="issueView"
          :project-identifier="project?.identifier || ''"
          @view-change="handleViewChange"
          @filters-changed="handleFiltersChanged"
          @columns-changed="handleColumnsChanged"
        />

        <IssueList
          v-if="issueView === 'list'"
          :key="'list-' + issueRefreshKey"
          :project-id="projectId"
          :workspace-id="workspaceId"
          :rql="currentRQL"
          :filter-sort-by="currentSortBy?.key"
          :filter-sort-dir="currentSortBy?.direction"
          :filter-group-by="currentGroupBy?.key"
          :search-term="searchTerm"
          :columns="currentColumns"
          @select="openDetailPanel"
          @delete="handleDeleteIssue"
        />
        <IssueKanban
          v-else-if="issueView === 'kanban'"
          :key="'kanban-' + issueRefreshKey"
          :project-id="projectId"
          :workspace-id="workspaceId"
          :rql="currentRQL"
          :filter-sort-by="currentSortBy?.key"
          :filter-sort-dir="currentSortBy?.direction"
          @select="openDetailPanel"
        />
        <IssueTreeView
          v-else-if="issueView === 'tree'"
          :key="'tree-' + issueRefreshKey"
          :project-id="projectId"
          :workspace-id="workspaceId"
          :rql="currentRQL"
          :filter-sort-by="currentSortBy?.key"
          :filter-sort-dir="currentSortBy?.direction"
          :filter-group-by="currentGroupBy?.key"
          @select="openDetailPanel"
        />
        <IssueCalendar
          v-else-if="issueView === 'calendar'"
          :project-id="projectId"
          :workspace-id="workspaceId"
          :rql="currentRQL"
          :filter-sort-by="currentSortBy?.key"
          :filter-sort-dir="currentSortBy?.direction"
          @select="openDetailPanel"
        />
        <IssueGantt
          v-else-if="issueView === 'gantt'"
          :project-id="projectId"
          :workspace-id="workspaceId"
          :rql="currentRQL"
          :filter-sort-by="currentSortBy?.key"
          :filter-sort-dir="currentSortBy?.direction"
          @select="openDetailPanel"
          @create="router.push(`/workspace/${route.params.slug}/project/${projectId}/issues/new?view=tree`)"
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

      <div v-if="activeTab === 'reports'">
        <ReportBuilder :project-id="projectId" />
      </div>

      <div v-if="activeTab === 'updates'">
        <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-semibold">{{ t('project.updates.title') }}</h3>
            <button
              @click="showUpdateForm = true"
              class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700"
            >{{ t('project.updates.publish') }}</button>
          </div>

          <div v-if="updatesLoading" class="text-gray-500 text-sm">{{ t('common.loading') }}</div>

          <div v-else-if="projectUpdates.length === 0" class="text-center py-8 text-gray-400 text-sm">
            {{ t('project.updates.empty') }}
          </div>

          <div v-else class="space-y-4">
            <div v-for="update in projectUpdates" :key="update.id" class="border border-gray-100 dark:border-gray-700 rounded-lg p-4">
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium">{{ update.author?.display_name || t('common.unknown') }}</span>
                  <span class="text-xs px-2 py-0.5 rounded-full" :class="getUpdateStatusColor(update.status)">{{ getUpdateStatusLabel(update.status) }}</span>
                </div>
                <span class="text-xs text-gray-400">{{ new Date(update.created_at).toLocaleString(locale === 'zh-CN' ? 'zh-CN' : 'en-US') }}</span>
              </div>
              <p class="text-sm text-gray-700 dark:text-gray-300">{{ update.content }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="text-center py-12">
      <p class="text-gray-500">{{ t('project.notFound') }}</p>
      <button @click="goBack" class="mt-4 text-indigo-600 hover:text-indigo-800 text-sm">{{ t('common.back') }}</button>
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
    @created="triggerRefresh()"
  />
  <CommandPalette
    :visible="showCommandPalette"
    :workspace-slug="slug"
    :project-id="projectId"
    :workspace-id="workspaceId"
    @close="showCommandPalette = false"
  />
  <PageTabConfig v-if="showPageConfig" :project-id="projectId" @close="showPageConfig = false" @saved="tabs => { pageTabs = tabs }" />

  <!-- Update Form Modal -->
  <div v-if="showUpdateForm" class="fixed inset-0 bg-black/30 z-50 flex items-center justify-center">
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-md mx-4">
      <h2 class="text-lg font-semibold mb-4">{{ t('project.updates.publish') }}</h2>
      <div class="space-y-3">
        <div>
          <label class="text-sm text-gray-600 mb-1 block">{{ t('common.status') }}</label>
          <select v-model="updateForm.status" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600">
            <option value="on_track">{{ t('project.updates.statusOnTrack') }}</option>
            <option value="at_risk">{{ t('project.updates.statusAtRisk') }}</option>
            <option value="off_track">{{ t('project.updates.statusOffTrack') }}</option>
          </select>
        </div>
        <div>
          <label class="text-sm text-gray-600 mb-1 block">{{ t('project.updates.content') }}</label>
          <textarea v-model="updateForm.content" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" rows="4" :placeholder="t('project.updates.contentPlaceholder')"></textarea>
        </div>
      </div>
      <div class="flex justify-end gap-2 mt-5">
        <button @click="showUpdateForm = false" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">{{ t('project.updates.cancel') }}</button>
        <button @click="submitUpdate" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ t('project.updates.submit') }}</button>
      </div>
    </div>
  </div>

  <button
    @click="showAIChat = true"
    class="fixed bottom-16 right-4 w-12 h-12 bg-gradient-to-r from-indigo-500 to-purple-600 text-white rounded-full shadow-lg hover:shadow-xl hover:scale-105 transition flex items-center justify-center text-xl z-30"
    title="AI Assistant (Ctrl+J)"
  >🤖</button>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { workspaceApi } from '@/api/workspace'
import { projectApi, listPageTabs } from '@/api/project'
import { issueApi } from '@/api/issue'
import { projectUpdateApi, type ProjectUpdate } from '@/api/project-update'
import type { Workspace } from '@/types'
import type { ProjectResponse } from '@/types/project'
import type { ProjectPageTab } from '@/types/project-page-tab'
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
import FilterBar from '@/components/FilterBar.vue'
import AIChatSidebar from '@/components/AIChatSidebar.vue'
import AICreateDialog from '@/components/AICreateDialog.vue'
import CommandPalette from '@/components/CommandPalette.vue'
import PageTabConfig from '@/components/PageTabConfig.vue'
import ReportBuilder from '@/components/ReportBuilder.vue'
import type { CycleResponse } from '@/types/cycle'
import type { ModuleResponse } from '@/types/module'
import { useModuleStore } from '@/stores/module'
import { extractSearchTerm } from '@/utils/highlight'

const route = useRoute()
const router = useRouter()
const { t, locale } = useI18n()

const workspace = ref<Workspace | null>(null)
const project = ref<ProjectResponse | null>(null)
const loading = ref(false)
const activeTab = ref((route.query.tab as string) || 'issues')
const issueView = ref<'list' | 'kanban' | 'tree' | 'calendar' | 'gantt'>((route.query.view as any) || 'list')
const issueRefreshKey = ref(0)
function triggerRefresh() { issueRefreshKey.value++ }

const currentRQL = ref('')
const currentSortBy = ref<any>(null)
const currentGroupBy = ref<any>(null)
const currentColumns = ref<string[]>([])

const searchTerm = computed(() => extractSearchTerm(currentRQL.value))

function handleFiltersChanged(rql: string, sortBy: any = null, groupBy: any = null) {
  currentRQL.value = rql
  currentSortBy.value = sortBy
  currentGroupBy.value = groupBy
  triggerRefresh()
}

function handleColumnsChanged(columns: string[]) {
  currentColumns.value = columns
}

function handleViewChange(view: 'list' | 'kanban' | 'tree' | 'calendar' | 'gantt') {
  issueView.value = view
}

// Note: FilterBar handles its own data loading (states, cycles, members, modules, issueTypes, labels, customFields)


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

// Project updates
const projectUpdates = ref<ProjectUpdate[]>([])
const showUpdateForm = ref(false)
const updateForm = ref({ status: 'on_track', content: '' })
const latestUpdateStatus = ref<string>('')
const updatesLoading = ref(false)

async function loadUpdates() {
  updatesLoading.value = true
  try {
    projectUpdates.value = await projectUpdateApi.list(projectId.value, 10)
    if (projectUpdates.value.length > 0) {
      latestUpdateStatus.value = projectUpdates.value[0].status
    }
  } catch { projectUpdates.value = [] }
  finally { updatesLoading.value = false }
}

async function submitUpdate() {
  if (!updateForm.value.content?.trim()) {
      alert(t('project.updates.contentRequired'))
      return
    }
  try {
    await projectUpdateApi.create(projectId.value, updateForm.value.status, updateForm.value.content)
    updateForm.value = { status: 'on_track', content: '' }
    showUpdateForm.value = false
    await loadUpdates()
  } catch (err) { console.error('Failed to submit update:', err) }
}

function getUpdateStatusLabel(s: string) {
  const map: Record<string, string> = { on_track: t('project.updates.statusOnTrack'), at_risk: t('project.updates.statusAtRisk'), off_track: t('project.updates.statusOffTrack') }
  return map[s] || s
}
function getUpdateStatusColor(s: string) {
  const map: Record<string, string> = { on_track: 'bg-green-100 text-green-700', at_risk: 'bg-yellow-100 text-yellow-700', off_track: 'bg-red-100 text-red-700' }
  return map[s] || 'bg-gray-100'
}

// Reset active tab when switching between projects (Vue Router reuses the component)
watch(() => route.params.id, () => {
  activeTab.value = (route.query.tab as string) || 'issues'
  issueView.value = (route.query.view as any) || 'list'
})

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
  triggerRefresh()
}

function openCyclePanel(cycle: CycleResponse) {
  selectedCycle.value = cycle
  cyclePanelVisible.value = true
}

function goToCycleCreate() {
  const slug = route.params.slug as string
  router.push(`/workspace/${route.params.slug}/project/${projectId.value}/cycles/new?ws=${slug}`)
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
  if (await confirm(t('project.deleteModuleConfirm', { name: module.name }))) {
    try {
      await moduleStore.deleteModuleAction(module.id)
    } catch (err) {
      console.error('Failed to delete module:', err)
      alert(t('project.deleteFailed'))
    }
  }
}

const workspaceId = ref(0)
const projectId = ref(0)
const slug = ref('')

const pageTabs = ref<ProjectPageTab[]>([])
const showPageConfig = ref(false)
const defaultTabs = computed(() => [
  { id: 'issues', name: t('project.tab.issues') },
  { id: 'cycles', name: t('project.tab.cycles') },
  { id: 'modules', name: t('project.tab.modules') },
  { id: 'updates', name: t('project.tab.updates') },
  { id: 'pages', name: t('project.tab.pages') },
  { id: 'reports', name: t('project.tab.reports') },
  { id: 'settings', name: t('project.tab.settings') },
])

async function loadPageTabs() {
  try {
    const tabs = await listPageTabs(projectId.value)
    if (tabs.length > 0) pageTabs.value = tabs
  } catch { /* use defaults */ }
}

const computedTabs = computed(() => {
  if (pageTabs.value.length > 0) {
    return pageTabs.value.filter(t => t.visible).map(t => ({ id: t.route_key || `custom_${t.id}`, name: t.name }))
  }
  return defaultTabs.value
})

function goBack() {
  router.push(`/workspace/${route.params.slug}`)
}

async function handleDeleteIssue(issue: any) {
  if (await confirm(t('project.deleteIssueConfirm', { name: issue.name }))) {
    try {
      await issueApi.deleteIssue(issue.id)
      detailPanelVisible.value = false
      triggerRefresh()
    } catch (err) {
      console.error('Failed to delete issue:', err)
      alert(t('project.deleteFailed'))
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
    // Filter data is loaded by FilterBar component internally
    await loadUpdates()
    await loadPageTabs()
  } catch (err) {
    console.error('Failed to load project:', err)
  } finally {
    loading.value = false
  }
  const handleKeydown = (e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'j') {
      e.preventDefault()
      showAIChat.value = !showAIChat.value
    }
    if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
      e.preventDefault()
      showCommandPalette.value = !showCommandPalette.value
    }
  }
  document.addEventListener('keydown', handleKeydown)
  onUnmounted(() => {
    document.removeEventListener('keydown', handleKeydown)
  })
})
</script>

<style scoped>
.project-page {
  @apply p-6;
}
</style>
