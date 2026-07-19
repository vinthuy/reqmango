<template>
  <div v-if="visible" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="close">
    <div class="bg-white rounded-xl shadow-xl w-full max-w-3xl mx-4 max-h-[92vh] overflow-hidden flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between px-5 py-3.5 border-b border-gray-100">
        <div class="flex items-center gap-3 min-w-0">
          <h3 class="text-base font-semibold text-gray-900 shrink-0">{{ t('decompose.title') }}</h3>
          <span class="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded truncate" :title="parentIssue?.name">
            {{ t('decompose.parentInfo') }}: {{ parentIssue?.name }}
          </span>
        </div>
        <button @click="close" class="text-gray-400 hover:text-gray-600 p-1 rounded hover:bg-gray-100 shrink-0">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
        </button>
      </div>

      <!-- Body -->
      <div v-if="loadingData" class="flex-1 flex items-center justify-center py-20">
        <svg class="animate-spin h-6 w-6 text-gray-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
      </div>

      <template v-else>
        <!-- Warning banner -->
        <div v-if="!canDecompose" class="bg-amber-50 border-b border-amber-200 px-5 py-2.5 text-sm text-amber-700">
          {{ depthLimitReached ? t('decompose.hierarchyLimit') : t('decompose.noChildTypes') }}
        </div>

        <div class="flex-1 overflow-y-auto flex divide-x divide-gray-100">
          <!-- Left column: type + name + description -->
          <div class="flex-1 p-5 space-y-4 min-w-0">
            <!-- Type selector -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-2">{{ t('decompose.childType') }}</label>
              <div class="flex flex-wrap gap-1.5">
                <button
                  v-for="type in allowedChildTypes"
                  :key="type.id"
                  @click="form.typeId = type.id"
                  class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-sm transition-colors"
                  :class="form.typeId === type.id
                    ? 'border-indigo-400 bg-indigo-50 text-indigo-700'
                    : 'border-gray-200 text-gray-600 hover:border-gray-300 hover:bg-gray-50'"
                >
                  <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: type.color }"></span>
                  {{ type.name }}
                </button>
                <span v-if="allowedChildTypes.length === 0 && !depthLimitReached" class="text-sm text-gray-400">{{ t('decompose.noChildTypes') }}</span>
              </div>
            </div>

            <!-- Name -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1.5">{{ t('issue.titleRequired') }}</label>
              <input
                v-model="form.name"
                type="text"
                ref="nameInput"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-indigo-400 focus:border-indigo-400 text-sm font-medium"
                :placeholder="t('decompose.childNamePlaceholder')"
              />
            </div>

            <!-- Description -->
            <div class="flex-1 flex flex-col min-h-0">
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1.5">{{ t('issue.description') }}</label>
              <div class="flex-1 min-h-[180px]">
                <RichTextEditor
                  v-model="form.description"
                  :placeholder="t('issue.descriptionPlaceholder')"
                />
              </div>
            </div>
          </div>

          <!-- Right column: properties sidebar -->
          <div class="w-56 shrink-0 p-5 space-y-4 flex flex-col">
            <div class="flex items-center justify-between mb-1">
              <span class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('issue.properties') }}</span>
              <button
                @click="prefillAllFromParent"
                class="text-xs text-indigo-500 hover:text-indigo-700 transition-colors"
                :title="t('decompose.prefillFromParent')"
              >
                {{ t('decompose.prefillFromParent') }}
              </button>
            </div>

            <!-- State -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.state') }}</label>
              <select v-model="form.stateId" class="w-full px-2.5 py-1.5 border border-gray-200 rounded-lg text-xs focus:outline-none focus:ring-1 focus:ring-indigo-400 focus:border-indigo-400">
                <option value="">{{ t('issue.statePlaceholder') }}</option>
                <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
              </select>
            </div>

            <!-- Priority -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.priority') }}</label>
              <select v-model="form.priority" class="w-full px-2.5 py-1.5 border border-gray-200 rounded-lg text-xs focus:outline-none focus:ring-1 focus:ring-indigo-400 focus:border-indigo-400">
                <option value="">{{ t('issue.priorityPlaceholder') }}</option>
                <option value="urgent">{{ t('issue.priorityUrgent') }}</option>
                <option value="high">{{ t('issue.priorityHigh') }}</option>
                <option value="medium">{{ t('issue.priorityMedium') }}</option>
                <option value="low">{{ t('issue.priorityLow') }}</option>
                <option value="none">{{ t('issue.priorityNone') }}</option>
              </select>
            </div>

            <!-- Assignee -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.assignee') }}</label>
              <UserSelect
                v-model="form.assigneeId"
                :users="projectMembers"
                :placeholder="t('issue.assigneePlaceholder')"
                :clearable="true"
              />
            </div>

            <!-- Cycle -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.cycle') }}</label>
              <select v-model="form.cycleId" class="w-full px-2.5 py-1.5 border border-gray-200 rounded-lg text-xs focus:outline-none focus:ring-1 focus:ring-indigo-400 focus:border-indigo-400">
                <option value="">{{ t('issue.cyclePlaceholder') }}</option>
                <option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
            </div>

            <!-- Module -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.module') }}</label>
              <select v-model="form.moduleId" class="w-full px-2.5 py-1.5 border border-gray-200 rounded-lg text-xs focus:outline-none focus:ring-1 focus:ring-indigo-400 focus:border-indigo-400">
                <option value="">{{ t('issue.modulePlaceholder') }}</option>
                <option v-for="m in modules" :key="m.id" :value="m.id">{{ m.name }}</option>
              </select>
            </div>

            <!-- Release -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.release') }}</label>
              <select v-model="form.releaseId" class="w-full px-2.5 py-1.5 border border-gray-200 rounded-lg text-xs focus:outline-none focus:ring-1 focus:ring-indigo-400 focus:border-indigo-400">
                <option value="">{{ t('issue.releasePlaceholder') }}</option>
                <option v-for="r in releases" :key="r.id" :value="r.id">{{ r.name }} ({{ r.version }})</option>
              </select>
            </div>

            <!-- Start date -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.startDate') }}</label>
              <input v-model="form.startDate" type="date" class="w-full px-2.5 py-1.5 border border-gray-200 rounded-lg text-xs focus:outline-none focus:ring-1 focus:ring-indigo-400 focus:border-indigo-400" />
            </div>

            <!-- Target date -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.targetDate') }}</label>
              <input v-model="form.targetDate" type="date" class="w-full px-2.5 py-1.5 border border-gray-200 rounded-lg text-xs focus:outline-none focus:ring-1 focus:ring-indigo-400 focus:border-indigo-400" />
            </div>

            <!-- Labels -->
            <div>
              <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('issue.labels') }}</label>
              <LabelSelector
                :project-id="projectId"
                v-model="form.labelIds"
              />
            </div>
          </div>
        </div>
      </template>

      <!-- Error -->
      <div v-if="error" class="bg-red-50 border-t border-red-200 px-5 py-2.5 text-sm text-red-700">
        {{ error }}
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end gap-3 px-5 py-3 border-t border-gray-100 bg-gray-50">
        <button @click="close" class="px-3 py-1.5 text-xs border border-gray-300 rounded-lg hover:bg-gray-100 transition-colors">
          {{ t('issue.cancel') }}
        </button>
        <button
          @click="create"
          :disabled="!canCreate || saving || !canDecompose"
          class="px-4 py-1.5 text-xs bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1.5"
        >
          <svg v-if="saving" class="animate-spin w-3 h-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
          {{ t('decompose.createChild') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { createIssue } from '@/api/issue'
import * as stateApi from '@/api/project-settings'
import * as cycleApi from '@/api/cycle'
import * as moduleApi from '@/api/module'
import { releaseApi } from '@/api/release'
import projectApi from '@/api/project'
import { getIssueTypes as fetchIssueTypes } from '@/api/issue-type'
import RichTextEditor from '@/components/RichTextEditor.vue'
import UserSelect from '@/components/UserSelect.vue'
import LabelSelector from '@/components/LabelSelector.vue'
import type { IssueResponse } from '@/types/issue'
import type { IssueType } from '@/types/issue-type'
import type { State } from '@/types/project-settings'
import type { CycleResponse } from '@/types/cycle'
import type { ModuleResponse } from '@/types/module'

const props = defineProps<{
  visible: boolean
  parentIssue: IssueResponse | null
  issueTypes: IssueType[]
  projectId: number
  workspaceId: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', issue: IssueResponse): void
}>()

const { t } = useI18n()
const toast = useToast()

const nameInput = ref<HTMLInputElement>()
const saving = ref(false)
const loadingData = ref(false)
const error = ref<string | null>(null)

// Dropdown data
const states = ref<State[]>([])
const cycles = ref<CycleResponse[]>([])
const modules = ref<ModuleResponse[]>([])
const releases = ref<{ id: number; name: string; version: string }[]>([])

// Self-loaded issue types (fallback when parent doesn't provide them)
const selfLoadedIssueTypes = ref<IssueType[]>([])

// Merge parent-provided types with self-loaded fallback
const allIssueTypes = computed(() => {
  if (props.issueTypes && props.issueTypes.length > 0) {
    return props.issueTypes
  }
  return selfLoadedIssueTypes.value
})

interface UserOption {
  id: number
  display_name?: string
  username?: string
  email?: string
}

const projectMembers = ref<UserOption[]>([])

// Form
const form = reactive({
  name: '',
  description: '',
  typeId: null as number | null,
  stateId: '' as number | string,
  priority: '',
  assigneeId: '' as number | string,
  cycleId: '' as number | string,
  moduleId: '' as number | string,
  releaseId: '' as number | string,
  startDate: '',
  targetDate: '',
  labelIds: [] as number[],
})

// Allowed child types
const allowedChildTypes = computed(() => {
  if (!props.parentIssue || !props.parentIssue.issue_type) {
    console.log('[Decompose] no parentIssue or issue_type:', { hasParent: !!props.parentIssue, issueType: props.parentIssue?.issue_type })
    return []
  }
  const issueType = props.parentIssue.issue_type
  const types = allIssueTypes.value

  console.log('[Decompose] computing allowedChildTypes:', {
    parentTypeId: issueType.id,
    parentTypeName: issueType.name,
    parentLevel: issueType.level,
    allowedChildIds: issueType.allowed_child_type_ids,
    availableTypesCount: types.length,
    availableTypes: types.map(t => ({ id: t.id, name: t.name, level: t.level })),
  })

  // Prefer metadata already embedded on the issue (level / allowed_child_type_ids)
  // so decomposition works even when the parent type is not in the project's visible list.
  if (issueType.level !== undefined || (issueType.allowed_child_type_ids && issueType.allowed_child_type_ids.length > 0)) {
    if (issueType.allowed_child_type_ids && issueType.allowed_child_type_ids.length > 0) {
      return types.filter(t => issueType.allowed_child_type_ids!.includes(t.id))
    }
    const parentLevel = issueType.level ?? 0
    return types.filter(t => (t.level ?? 0) > parentLevel)
  }

  // Fallback: look up the full type in the project type list
  const parentType = types.find(t => t.id === issueType.id)
  if (!parentType) return []

  if (parentType.allowed_child_type_ids && parentType.allowed_child_type_ids.length > 0) {
    return types.filter(t => parentType.allowed_child_type_ids!.includes(t.id))
  }

  const parentLevel = parentType.level ?? 0
  return types.filter(t => (t.level ?? 0) > parentLevel)
})

const depthLimitReached = computed(() => (props.parentIssue?.depth ?? 0) >= 5)

const canDecompose = computed(() => !depthLimitReached.value && allowedChildTypes.value.length > 0)

const canCreate = computed(() => form.name.trim() !== '' && form.typeId !== null)

// Date helper: strip time from ISO string to YYYY-MM-DD
function toDateInputValue(isoStr?: string): string {
  if (!isoStr) return ''
  return isoStr.substring(0, 10)
}

// Pre-fill form from parent issue
function prefillFromParent() {
  if (!props.parentIssue) return

  // Type: first allowed child type
  if (allowedChildTypes.value.length > 0 && form.typeId === null) {
    form.typeId = allowedChildTypes.value[0].id
  }

  // Priority
  if (props.parentIssue.priority && props.parentIssue.priority !== 'none') {
    form.priority = props.parentIssue.priority
  }

  // Description
  if (props.parentIssue.description_html) {
    form.description = props.parentIssue.description_html
  }

  // Assignee: first from parent
  if (props.parentIssue.assignees?.length && !form.assigneeId) {
    form.assigneeId = props.parentIssue.assignees[0].id
  }

  // Cycle
  if (props.parentIssue.cycle_id) {
    form.cycleId = props.parentIssue.cycle_id
  }

  // Module
  if (props.parentIssue.module_ids?.length) {
    form.moduleId = props.parentIssue.module_ids[0]
  }

  // Release
  if (props.parentIssue.release_id) {
    form.releaseId = props.parentIssue.release_id
  }

  // Dates
  form.startDate = toDateInputValue(props.parentIssue.start_date)
  form.targetDate = toDateInputValue(props.parentIssue.target_date)

  // Labels
  if (props.parentIssue.label_details?.length) {
    form.labelIds = props.parentIssue.label_details.map(l => l.id)
  }

  // State: try to match parent's state if loaded
  if (props.parentIssue.state_id && states.value.length > 0) {
    const matched = states.value.find(s => s.id === props.parentIssue!.state_id)
    if (matched) {
      form.stateId = matched.id
    }
  }
}

// "Fill from parent" button — override all fields with parent values
async function prefillAllFromParent() {
  prefillFromParent()

  // Force-override all fields even if already has value
  if (!props.parentIssue) return

  if (props.parentIssue.description_html) {
    form.description = props.parentIssue.description_html
  }
  if (props.parentIssue.priority) {
    form.priority = props.parentIssue.priority
  }
  if (props.parentIssue.assignees?.length) {
    form.assigneeId = props.parentIssue.assignees[0].id
  }
  if (props.parentIssue.cycle_id) {
    form.cycleId = props.parentIssue.cycle_id
  }
  if (props.parentIssue.module_ids?.length) {
    form.moduleId = props.parentIssue.module_ids[0]
  }
  if (props.parentIssue.release_id) {
    form.releaseId = props.parentIssue.release_id
  }
  form.startDate = toDateInputValue(props.parentIssue.start_date)
  form.targetDate = toDateInputValue(props.parentIssue.target_date)
  if (props.parentIssue.label_details?.length) {
    form.labelIds = [...props.parentIssue.label_details.map(l => l.id)]
  }
  if (props.parentIssue.state_id && states.value.length > 0) {
    const matched = states.value.find(s => s.id === props.parentIssue!.state_id)
    if (matched) form.stateId = matched.id
  }

  toast.success(t('decompose.prefilled'))
}

// Reset form to defaults
function resetForm() {
  form.name = ''
  form.description = ''
  form.typeId = null
  form.stateId = ''
  form.priority = ''
  form.assigneeId = ''
  form.cycleId = ''
  form.moduleId = ''
  form.releaseId = ''
  form.startDate = ''
  form.targetDate = ''
  form.labelIds = []
  error.value = null
}

// Load dropdown data
async function loadDropdownData() {
  try { states.value = await stateApi.listStates(props.projectId) } catch (e) { console.error('load states:', e) }
  try {
    const cyc = await cycleApi.listCycles(props.projectId)
    cycles.value = cyc.items || []
  } catch (e) { console.error('load cycles:', e) }
  try { modules.value = await moduleApi.listModules(props.projectId, props.workspaceId) } catch (e) { console.error('load modules:', e) }
  try { releases.value = await releaseApi.list(props.projectId) } catch (e) { console.error('load releases:', e) }
  try {
    const members = await projectApi.listProjectMembers(props.projectId)
    projectMembers.value = (members || []).map((m: any) => ({
      id: m.user_id,
      display_name: m.user?.display_name || m.user?.username,
      username: m.user?.username,
      email: m.user?.email,
    }))
  } catch (e) { console.error('load members:', e) }
}

// Watch visible -> load & prefill
watch(() => props.visible, async (v) => {
  if (v) {
    resetForm()
    loadingData.value = true
    error.value = null
    // If parent didn't provide issue types, load them ourselves
    if (!props.issueTypes || props.issueTypes.length === 0) {
      try { selfLoadedIssueTypes.value = await fetchIssueTypes(props.workspaceId, props.projectId) } catch (e) { /* */ }
    }
    await loadDropdownData()
    prefillFromParent()
    loadingData.value = false
    await nextTick()
    nameInput.value?.focus()
  }
})

// Create
async function create() {
  if (!canCreate.value || !props.parentIssue) return

  saving.value = true
  error.value = null

  try {
    const data: any = {
      name: form.name.trim(),
      type_id: form.typeId!,
      parent_id: props.parentIssue.id,
    }

    // Priority
    if (form.priority) {
      data.priority = form.priority
    }

    // Description
    if (form.description && form.description.trim() !== '<p></p>') {
      data.description_html = form.description
    }

    // State
    const stateId = typeof form.stateId === 'string' ? parseInt(form.stateId) : form.stateId
    if (stateId > 0) {
      data.state_id = stateId
    }

    // Assignee
    const assigneeId = typeof form.assigneeId === 'string' ? parseInt(form.assigneeId) : form.assigneeId
    if (assigneeId > 0) {
      data.assignee_ids = [assigneeId]
    }

    // Cycle
    const cycleId = typeof form.cycleId === 'string' ? parseInt(form.cycleId) : form.cycleId
    if (cycleId > 0) {
      data.cycle_id = cycleId
    }

    // Module
    const moduleId = typeof form.moduleId === 'string' ? parseInt(form.moduleId) : form.moduleId
    if (moduleId > 0) {
      data.module_ids = [moduleId]
    }

    // Release
    const releaseId = typeof form.releaseId === 'string' ? parseInt(form.releaseId) : form.releaseId
    if (releaseId > 0) {
      data.release_id = releaseId
    }

    // Dates
    if (form.startDate) {
      data.start_date = form.startDate + 'T00:00:00Z'
    }
    if (form.targetDate) {
      data.target_date = form.targetDate + 'T00:00:00Z'
    }

    // Labels
    if (form.labelIds.length > 0) {
      data.label_ids = form.labelIds
    }

    const issue = await createIssue(props.projectId, props.workspaceId, data)
    toast.success(t('decompose.createSuccess'))
    emit('created', issue)
    close()
  } catch (e: any) {
    const msg = e?.response?.data?.message || e?.message || t('decompose.createFailed')
    error.value = msg
  } finally {
    saving.value = false
  }
}

function close() {
  emit('close')
}
</script>
