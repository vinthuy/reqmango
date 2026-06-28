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
            {{ issue?.issue_type?.name || t('issue.unknown') }} #{{ issue?.sequence_id }}
          </h1>
        </div>
        <div class="flex items-center space-x-3">
          <button
            @click="saveIssue"
            :disabled="saving"
            class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
          >
            {{ saving ? t('issue.saving') : t('issue.save') }}
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
              :placeholder="t('issue.titlePlaceholder')"
              class="w-full text-xl font-semibold border-0 focus:outline-none focus:ring-0"
            />
          </div>

          <!-- 描述 -->
          <div class="bg-white rounded-lg shadow-sm p-6">
            <h3 class="text-sm font-medium text-gray-700 mb-3">{{ t('issue.description') }}</h3>
            <RichTextEditor
              v-model="issueForm.description"
              :placeholder="t('issue.descriptionPlaceholder')"
            />
          </div>

          <!-- 评论区 -->
          <div class="bg-white rounded-lg shadow-sm p-6">
            <h3 class="text-sm font-medium text-gray-700 mb-4">{{ t('issue.comments') }}</h3>
            <CommentList :issue-id="issueId" />
            <TimeTrackPanel :issue-id="issueId" />
            <RecurrenceConfig :issue-id="issueId" />
          </div>

          <!-- 活动历史 -->
          <div v-if="activities.length > 0" class="bg-white rounded-lg shadow-sm p-6">
            <h3 class="text-sm font-medium text-gray-700 mb-4">{{ t('issue.activity') }}</h3>
            <div class="space-y-2">
              <div v-for="act in activities" :key="act.id" class="flex items-start space-x-3 text-sm">
                <span class="text-xs text-gray-400 w-20 shrink-0">{{ formatActivityTime(act.created_at) }}</span>
                <span class="text-gray-600">{{ formatActivity(act) }}</span>
              </div>
            </div>
          </div>

          <!-- 自定义字段区域 -->
          <div class="bg-white rounded-lg shadow-sm p-6">
            <h3 class="text-sm font-medium text-gray-700 mb-4">{{ t('issue.customFields') }}</h3>
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

          <!-- 子工作项面板 -->
          <div class="bg-white rounded-lg shadow-sm p-6">
            <SubIssuePanel
              :sub-issues="subIssues"
              @create="createSubIssue"
              @edit="editSubIssue"
            />
          </div>
        </div>

        <!-- 右侧：属性面板 -->
        <div class="space-y-6">
          <!-- 状态 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.state') }}</label>
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
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.type') }}</label>
            <select
              v-model="issueForm.type_id"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">{{ t('issue.notSet') }}</option>
              <option v-for="it in issueTypes" :key="it.id" :value="it.id">
                {{ it.name }}
              </option>
            </select>
          </div>

          <!-- 优先级 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.priority') }}</label>
            <select
              v-model="issueForm.priority"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="urgent">{{ t('issue.priorityUrgent') }}</option>
              <option value="high">{{ t('issue.priorityHigh') }}</option>
              <option value="medium">{{ t('issue.priorityMedium') }}</option>
              <option value="low">{{ t('issue.priorityLow') }}</option>
              <option value="none">{{ t('issue.priorityNone') }}</option>
            </select>
          </div>

          <!-- 负责人 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.assignee') }}</label>
            <select
              v-model="issueForm.assignee_id"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">{{ t('issue.unassigned') }}</option>
              <option v-for="member in projectMembers" :key="member.id" :value="member.id">
                {{ member.display_name || member.email }}
              </option>
            </select>
          </div>

          <!-- 周期 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.cycle') }}</label>
            <select
              v-model="issueForm.cycle_id"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">{{ t('issue.noCycle') }}</option>
              <option v-for="cycle in cycles" :key="cycle.id" :value="cycle.id">
                {{ cycle.name }}
              </option>
            </select>
          </div>

          <!-- 模块 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.module') }}</label>
            <select
              v-model="issueForm.module_id"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">{{ t('issue.noModule') }}</option>
              <option v-for="module in modules" :key="module.id" :value="module.id">
                {{ module.name }}
              </option>
            </select>
          </div>

          <!-- 开始日期 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.startDate') }}</label>
            <input
              v-model="issueForm.start_date"
              type="date"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          <!-- 目标日期 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.targetDate') }}</label>
            <input
              v-model="issueForm.target_date"
              type="date"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          <!-- 标签 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.labels') }}</label>
            <LabelSelector
              :project-id="projectId"
              v-model="selectedLabelIds"
              @change="saveLabels"
            />
          </div>

          <!-- 关联关系 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.relations') }}</label>
            <div v-if="relations.length > 0" class="space-y-2 mb-3">
              <div v-for="rel in relations" :key="rel.id" class="flex items-center justify-between text-sm p-2 bg-gray-50 rounded">
                <div class="flex items-center space-x-2">
                  <span class="text-xs px-1.5 py-0.5 bg-blue-100 text-blue-700 rounded">{{ rel.relation_type?.outward_name || rel.relation_type?.name || t('issue.relations') }}</span>
                  <span>#{{ rel.related_issue?.sequence_id || rel.related_issue_id }}</span>
                  <span class="text-gray-700 truncate max-w-[120px]">{{ rel.related_issue?.name || '' }}</span>
                </div>
                <button @click="deleteRelation(rel.id)" class="text-gray-400 hover:text-red-500">&times;</button>
              </div>
            </div>
            <div v-if="!showAddRelation" class="text-xs text-gray-400">{{ t('issue.noRelations') }}</div>
            <button v-if="!showAddRelation" @click="showAddRelation = true" class="text-xs text-indigo-600 hover:text-indigo-800 mt-1">{{ t('issue.addRelation') }}</button>
            <div v-if="showAddRelation" class="mt-2 space-y-2">
              <select v-model="newRelation.type_id" class="w-full px-2 py-1 border rounded text-xs">
                <option value="">{{ t('issue.selectRelationType') }}</option>
                <option v-for="rt in relationTypes" :key="rt.id" :value="rt.id">{{ rt.outward_name }} ({{ rt.name }})</option>
              </select>
              <input v-model="newRelation.search" @input="searchRelatedIssues" type="text" :placeholder="t('issue.searchIssuesPlaceholder')" class="w-full px-2 py-1 border rounded text-xs" />
              <div v-if="relationSearchResults.length > 0" class="max-h-24 overflow-y-auto border rounded">
                <div v-for="r in relationSearchResults" :key="r.id" @click="addRelation(r)" class="px-2 py-1 hover:bg-indigo-50 cursor-pointer text-xs">
                  #{{ r.sequence_id }} {{ r.name?.substring(0, 30) }}
                </div>
              </div>
              <button @click="showAddRelation = false" class="text-xs text-gray-500">{{ t('issue.cancel') }}</button>
            </div>
          </div>

          <!-- 关联文档页面 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('issue.relatedPages') }}</label>
            <div v-if="pages.length > 0" class="space-y-2 mb-3">
              <div v-for="page in pages" :key="page.id" class="flex items-center justify-between text-sm p-2 bg-gray-50 rounded">
                <div class="flex items-center space-x-2">
                  <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  <span class="text-gray-700 truncate max-w-[120px]">{{ page.title }}</span>
                </div>
                <button @click="removePage(page.id)" class="text-gray-400 hover:text-red-500">&times;</button>
              </div>
            </div>
            <div v-if="!showAddPage" class="text-xs text-gray-400">{{ t('issue.noPages') }}</div>
            <button v-if="!showAddPage" @click="showAddPage = true" class="text-xs text-indigo-600 hover:text-indigo-800 mt-1">{{ t('issue.addPage') }}</button>
            <div v-if="showAddPage" class="mt-2 space-y-2">
              <input v-model="newPage.search" @input="searchPages" type="text" :placeholder="t('issue.searchPagesPlaceholder')" class="w-full px-2 py-1 border rounded text-xs" />
              <div v-if="pageSearchResults.length > 0" class="max-h-24 overflow-y-auto border rounded">
                <div v-for="p in pageSearchResults" :key="p.id" @click="addPage(p)" class="px-2 py-1 hover:bg-indigo-50 cursor-pointer text-xs">
                  {{ p.title?.substring(0, 40) }}
                </div>
              </div>
              <button @click="showAddPage = false" class="text-xs text-gray-500">{{ t('issue.cancel') }}</button>
            </div>
          </div>

          <!-- 附件 -->
          <div class="bg-white rounded-lg shadow-sm p-4">
            <AttachmentManager :issue-id="issueId" :project-id="projectId" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useRoute, useRouter } from 'vue-router'
import CustomFieldManager from '@/components/CustomFieldManager.vue'
import LabelSelector from '@/components/LabelSelector.vue'
import CommentList from '@/components/CommentList.vue'
import RichTextEditor from '@/components/RichTextEditor.vue'
import SubIssuePanel from '@/components/SubIssuePanel.vue'
import TimeTrackPanel from '@/components/TimeTrackPanel.vue'
import RecurrenceConfig from '@/components/RecurrenceConfig.vue'
import AttachmentManager from '@/components/AttachmentManager.vue'
import * as issueApi from '@/api/issue'
import * as issueTypeApi from '@/api/issue-type'
import * as stateApi from '@/api/project-settings'
import * as cycleApi from '@/api/cycle'
import * as moduleApi from '@/api/module'
import projectApi from '@/api/project'
import * as relationApi from '@/api/relation'
import * as pageApi from '@/api/page'
import * as issueApiSearch from '@/api/issue'

// Route params
const route = useRoute()
const router = useRouter()
const { t, locale } = useI18n()
const issueId = parseInt(route.params.issueId as string, 10) || 0

let workspaceId = 0
let projectId = 0

// State
const issue = ref<any>(null)
const states = ref<any[]>([])
const cycles = ref<any[]>([])
const modules = ref<any[]>([])
const projectMembers = ref<any[]>([])
const issueTypes = ref<any[]>([])
const subIssues = ref<any[]>([])
const saving = ref(false)
const selectedLabelIds = ref<number[]>([])
const relations = ref<any[]>([])
const relationTypes = ref<any[]>([])
const activities = ref<any[]>([])
const showAddRelation = ref(false)
const newRelation = ref({ type_id: '', search: '' })
const relationSearchResults = ref<any[]>([])
const customFieldManagerRef = ref<InstanceType<typeof CustomFieldManager> | null>(null)
const customFieldValues = ref<Record<string, any>>({})
const pages = ref<any[]>([])
const showAddPage = ref(false)
const newPage = ref({ search: '' })
const pageSearchResults = ref<any[]>([])

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
  target_date: ''
})

// Load issue data
onMounted(async () => {
  await loadIssue()
  await Promise.all([
    loadStates(),
    loadCycles(),
    loadModules(),
    loadProjectMembers(),
    loadIssueTypes(),
    loadRelations(),
    loadActivities(),
    loadPages()
  ])
})

// Load issue
async function loadIssue() {
  try {
    const issueData = await issueApi.getIssue(issueId)
    issue.value = issueData
    
    // Set workspaceId and projectId from issue data
    if (issueData.workspace_id) workspaceId = issueData.workspace_id
    if (issueData.project_id) projectId = issueData.project_id

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
    }
    selectedLabelIds.value = issue.value.labels || []
    
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
  if (!issueForm.value.name?.trim()) {
    if ((window as any).$toast) (window as any).$toast.error(t('issue.nameRequired'))
    return
  }
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

    alert(t('issue.saveSuccess'))
  } catch (error) {
    console.error('Failed to save issue:', error)
    alert(t('issue.saveFailed'))
  } finally {
    saving.value = false
  }
}

// Save labels via API
async function saveLabels(labelIds: number[]) {
  try {
    const currentIds: Set<number> = new Set(issue.value?.labels || [])
    const newIds: Set<number> = new Set(labelIds)

    // Add new labels
    for (const id of labelIds) {
      if (!currentIds.has(id)) {
        await issueApi.addIssueLabel(issueId, id)
      }
    }
    // Remove deselected labels
    for (const id of currentIds) {
      if (!newIds.has(id)) {
        await issueApi.removeIssueLabel(issueId, id)
      }
    }
  } catch (e) {
    console.error('Failed to save labels:', e)
  }
}

// ===== Relations =====
async function loadRelations() {
  try {
    const [rels, types] = await Promise.all([
      relationApi.listIssueRelations(issueId),
      relationApi.listRelationTypes(workspaceId)
    ])
    relations.value = rels || []
    relationTypes.value = types || []
  } catch (e) { console.error('Failed to load relations:', e) }
}

let relationSearchTimer: any = null
async function searchRelatedIssues() {
  if (relationSearchTimer) clearTimeout(relationSearchTimer)
  if (!newRelation.value.search.trim()) { relationSearchResults.value = []; return }
  relationSearchTimer = setTimeout(async () => {
    try {
      const result: any = await issueApiSearch.searchIssues(workspaceId, newRelation.value.search)
      relationSearchResults.value = (Array.isArray(result) ? result : (result.items || [])).slice(0, 8)
    } catch (e) { console.error('Search failed:', e) }
  }, 300)
}

async function addRelation(related: any) {
  if (!newRelation.value.type_id) return
  try {
    await relationApi.createIssueRelation(issueId, {
      related_issue_id: related.id,
      relation_type_id: parseInt(newRelation.value.type_id)
    })
    showAddRelation.value = false
    newRelation.value = { type_id: '', search: '' }
    relationSearchResults.value = []
    await loadRelations()
  } catch (e) { console.error('Failed to add relation:', e) }
}

async function deleteRelation(relationId: number) {
  try {
    await relationApi.deleteIssueRelation(relationId)
    await loadRelations()
  } catch (e) { console.error('Failed to delete relation:', e) }
}

// ===== Activity History =====
async function loadActivities() {
  try {
    const result: any = await issueApi.getIssueActivities(issueId)
    activities.value = Array.isArray(result) ? result.slice(0, 20) : (result?.activities || []).slice(0, 20)
  } catch (e) { console.error('Failed to load activities:', e) }
}

function formatActivityTime(timeStr: string): string {
  const d = new Date(timeStr)
  return d.toLocaleDateString(locale.value === 'zh-CN' ? 'zh-CN' : 'en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function formatActivity(act: any): string {
  const fieldMap: Record<string, string> = {
    name: t('activity.name'), state_id: t('activity.state_id'), priority: t('activity.priority'),
    assignees: t('activity.assignees'), labels: t('activity.labels'),
    cycle: t('activity.cycle'), description: t('activity.description'), created: t('activity.created'),
  }
  const verb = act.verb === 'created' ? t('activity.createdIssue') : act.verb === 'updated' ? t('activity.updated') : act.verb
  const field = fieldMap[act.field || ''] || act.field || ''
  if (act.verb === 'created') return t('activity.createdIssue')
  if (act.old_value && act.new_value) return `${verb}${field}: ${act.old_value} → ${act.new_value}`
  return `${verb}${field}`
}

// Create sub issue
function createSubIssue() {
  // Navigate to create issue page with parent_id
  router.push({
    path: `/workspaces/${workspaceId}/projects/${projectId}/issues/new`,
    query: { parent_id: issueId }
  })
}

// Edit sub issue
function editSubIssue(issue: any) {
  router.push(`/workspaces/${workspaceId}/projects/${projectId}/issues/${issue.id}`)
}

// Pages
async function loadPages() {
  try {
    pages.value = await issueApi.listIssuePages(issueId)
  } catch (e) { console.error('Failed to load pages:', e) }
}

let pageSearchTimer: any = null
async function searchPages() {
  if (pageSearchTimer) clearTimeout(pageSearchTimer)
  if (!newPage.value.search.trim()) { pageSearchResults.value = []; return }
  pageSearchTimer = setTimeout(async () => {
    try {
      const pages: any[] = await pageApi.searchPages(projectId, newPage.value.search)
      pageSearchResults.value = pages.slice(0, 8)
    } catch (e) { console.error('Search pages failed:', e) }
  }, 300)
}

async function addPage(page: any) {
  try {
    await issueApi.addIssuePage(issueId, page.id)
    showAddPage.value = false
    newPage.value = { search: '' }
    pageSearchResults.value = []
    await loadPages()
  } catch (e) { console.error('Failed to add page:', e) }
}

async function removePage(pageId: number) {
  try {
    await issueApi.removeIssuePage(issueId, pageId)
    await loadPages()
  } catch (e) { console.error('Failed to remove page:', e) }
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