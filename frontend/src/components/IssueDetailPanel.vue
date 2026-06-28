<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex">
      <!-- 遮罩层 -->
      <div class="absolute inset-0 bg-black/30" @click="close"></div>
      <!-- 侧滑面板 -->
      <div class="relative ml-auto w-full max-w-xl bg-white shadow-xl flex flex-col h-full overflow-hidden" :class="visible && 'animate-slide-in'">
        <div v-if="loading" class="flex items-center justify-center h-full">
          <div class="text-gray-400 text-sm">{{ t('common.loading') }}</div>
        </div>
        <template v-else-if="issue">
          <!-- 头部 -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 shrink-0">
            <div class="flex items-center space-x-3 min-w-0 flex-1">
              <span class="text-sm text-gray-400 font-mono shrink-0">#{{ issue.sequence_id }}</span>
              <input
                v-if="editing"
                v-model="editForm.name"
                class="flex-1 text-lg font-medium text-gray-900 border border-gray-300 rounded px-2 py-0.5 focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
              <span v-else class="text-lg font-medium text-gray-900 truncate">{{ issue.name }}</span>
            </div>
            <div class="flex items-center space-x-2 shrink-0">
              <button
                v-if="!editing"
                @click="startEdit"
                class="px-3 py-1.5 text-sm text-indigo-600 border border-indigo-300 rounded-md hover:bg-indigo-50"
              >{{ t('common.edit') }}</button>
              <button @click="close" class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-md">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
              </button>
            </div>
          </div>

          <!-- 内容区 -->
          <div class="flex-1 overflow-y-auto px-6 py-4 space-y-5">
            <!-- 编辑表单 -->
            <template v-if="editing">
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs text-gray-500 mb-1">{{ t('issue.state') }}</label>
                  <select v-model="editForm.state_id" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm">
                    <option v-for="s in stateOptions" :key="s.id" :value="s.id">{{ s.name }}</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs text-gray-500 mb-1">{{ t('issue.priority') }}</label>
                  <select v-model="editForm.priority" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm">
                    <option value="urgent">{{ t('issue.priorityUrgent') }}</option>
                    <option value="high">{{ t('issue.priorityHigh') }}</option>
                    <option value="medium">{{ t('issue.priorityMedium') }}</option>
                    <option value="low">{{ t('issue.priorityLow') }}</option>
                    <option value="none">{{ t('issue.priorityNone') }}</option>
                  </select>
                </div>
              </div>
              <div>
                <label class="block text-xs text-gray-500 mb-1">{{ t('issue.type') }}</label>
                <select v-model="editForm.type_id" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm">
                  <option :value="0">{{ t('issueDetail.unset') }}</option>
                  <option v-for="t in issueTypeOptions" :key="t.id" :value="t.id">{{ t.name }}</option>
                </select>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs text-gray-500 mb-1">{{ t('issue.startDate') }}</label>
                  <input v-model="editForm.start_date" type="date" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm" />
                </div>
                <div>
                  <label class="block text-xs text-gray-500 mb-1">{{ t('issue.targetDate') }}</label>
                  <input v-model="editForm.target_date" type="date" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm" />
                </div>
              </div>
              <div>
                <label class="block text-xs text-gray-500 mb-1">{{ t('issue.cycle') }}</label>
                <select v-model="editForm.cycle_id" class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm">
                  <option :value="0">{{ t('issueDetail.unset') }}</option>
                  <option v-for="c in cycleOptions" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>
              <div>
                <label class="block text-xs text-gray-500 mb-2">{{ t('issue.description') }}</label>
                <RichTextEditor v-model="editForm.description_html" :placeholder="t('issueDetail.descriptionPlaceholder')" />
              </div>
              <div v-if="issue">
                <label class="block text-xs text-gray-500 mb-2">{{ t('issue.customFields') }}</label>
                <CustomFieldManager
                  ref="panelCustomFieldRef"
                  :workspace-id="workspaceId"
                  :project-id="projectId"
                  :issue-id="issue.id"
                  :issue-type-id="issue.issue_type?.id"
                  mode="display"
                  :members="[]"
                />
              </div>
            </template>

            <!-- 查看模式 -->
            <template v-else>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs text-gray-500 mb-1">{{ t('issue.state') }}</label>
                  <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium" :style="{ backgroundColor: (issue.state_color || '#6366f1') + '20', color: issue.state_color || '#6366f1' }">
                    <span class="w-2 h-2 rounded-full mr-1.5" :style="{ backgroundColor: issue.state_color || '#6366f1' }"></span>
                    {{ issue.state_name || '-' }}
                  </span>
                </div>
                <div>
                  <label class="block text-xs text-gray-500 mb-1">{{ t('issue.priority') }}</label>
                  <span :class="priorityClass(issue.priority)" class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium">
                    <span class="w-1.5 h-1.5 rounded-full mr-1.5" :class="priorityDot(issue.priority)"></span>
                    {{ priorityLabel(issue.priority) }}
                  </span>
                </div>
              </div>
              <div v-if="issue.issue_type" class="grid grid-cols-1 gap-3">
                <div>
                  <label class="block text-xs text-gray-500 mb-1">{{ t('issue.type') }}</label>
                  <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-700">
                    <span class="w-2 h-2 rounded-full mr-1.5" :style="{ backgroundColor: issue.issue_type.color }"></span>
                    {{ issue.issue_type.name }}
                  </span>
                </div>
              </div>
              <div v-if="issue.assignees && issue.assignees.length > 0">
                <label class="block text-xs text-gray-500 mb-1">{{ t('issue.assignee') }}</label>
                <div class="flex flex-wrap gap-1.5">
                  <span v-for="a in issue.assignees" :key="a.id" class="inline-flex items-center px-2 py-0.5 bg-gray-100 rounded text-sm text-gray-700">
                    {{ a.display_name || a.username || 'User #' + a.id }}
                  </span>
                </div>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div><label class="block text-xs text-gray-500 mb-1">{{ t('issue.startDate') }}</label><span class="text-sm text-gray-800">{{ issue.start_date || '-' }}</span></div>
                <div><label class="block text-xs text-gray-500 mb-1">{{ t('issue.targetDate') }}</label><span class="text-sm text-gray-800">{{ issue.target_date || '-' }}</span></div>
              </div>
              <div v-if="issue.cycle_id">
                <label class="block text-xs text-gray-500 mb-1">{{ t('issue.cycle') }}</label>
                <span class="inline-flex items-center px-2 py-0.5 bg-indigo-50 text-indigo-700 rounded text-sm">Cycle #{{ issue.cycle_id }}</span>
              </div>
              <div v-if="issue.module_ids && issue.module_ids.length > 0">
                <label class="block text-xs text-gray-500 mb-1">{{ t('issueDetail.module') }}</label>
                <div class="flex flex-wrap gap-1.5">
                  <span v-for="mid in issue.module_ids" :key="mid" class="inline-flex items-center px-2 py-0.5 bg-gray-100 rounded text-sm text-gray-700">#{{ mid }}</span>
                </div>
              </div>
              <div>
                <label class="block text-xs text-gray-500 mb-2">{{ t('issue.description') }}</label>
                <div v-if="issue.description_html" class="prose prose-sm max-w-none text-gray-700 bg-gray-50 rounded-md p-3" v-html="issue.description_html"></div>
                <div v-else class="text-sm text-gray-400 italic">{{ t('issueDetail.noDescription') }}</div>
              </div>
              <div class="border-t border-gray-100 pt-3">
                <div class="text-xs text-gray-400 space-y-1">
                  <div>{{ t('issueDetail.createdAt').replace('{0}', formatDate(issue.created_at)) }}</div>
                  <div>{{ t('issueDetail.updatedAt').replace('{0}', formatDate(issue.updated_at)) }}</div>
                </div>
              </div>
            </template>
          </div>

          <!-- 底部操作栏 -->
          <div class="border-t border-gray-200 px-6 py-3 flex items-center justify-between shrink-0">
            <div class="text-xs text-gray-400">{{ t('issueDetail.createdAt').replace('{0}', formatDate(issue.created_at)) }}</div>
            <div class="flex items-center space-x-2">
              <template v-if="editing">
                <button @click="cancelEdit" class="px-4 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50">{{ t('common.cancel') }}</button>
                <button @click="saveEdit" :disabled="saving" class="px-4 py-1.5 text-sm text-white bg-indigo-600 rounded-md hover:bg-indigo-700 disabled:opacity-50">
                  {{ saving ? t('issue.saving') : t('common.save') }}
                </button>
              </template>
              <template v-else>
                <button @click="deleteIssue" class="px-4 py-1.5 text-sm text-red-600 border border-red-300 rounded-md hover:bg-red-50">{{ t('common.delete') }}</button>
              </template>
            </div>
          </div>
        </template>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import issueApi from '@/api/issue'
import * as issueTypeApi from '@/api/issue-type'
import api from '@/api'
import { useConfirm } from '@/composables/useConfirm'
import { useI18n } from '@/composables/useI18n'
import RichTextEditor from '@/components/RichTextEditor.vue'
import CustomFieldManager from '@/components/CustomFieldManager.vue'

const { confirm } = useConfirm()
const { t, locale } = useI18n()

const props = defineProps<{
  issueId: number | null
  visible: boolean
  workspaceId: number
  projectId: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'delete', issue: any): void
  (e: 'refresh'): void
}>()

const issue = ref<any>(null)
const loading = ref(false)
const editing = ref(false)
const saving = ref(false)

const stateOptions = ref<any[]>([])
const cycleOptions = ref<any[]>([])
const issueTypeOptions = ref<any[]>([])
const panelCustomFieldRef = ref<InstanceType<typeof CustomFieldManager> | null>(null)

const editForm = ref({
  name: '',
  state_id: 0,
  priority: '',
  type_id: 0,
  start_date: '',
  target_date: '',
  cycle_id: 0,
  description_html: ''
})

watch(() => props.issueId, async (id) => {
  if (id && props.visible) {
    loading.value = true
    editing.value = false
    try {
      const result = await issueApi.getIssue(id)
      issue.value = result
    } catch (e) {
      console.error('Failed to load issue:', e)
      issue.value = null
    } finally {
      loading.value = false
    }
  } else if (!props.visible) {
    issue.value = null
  }
}, { immediate: true })

async function startEdit() {
  if (!issue.value) return
  editing.value = true

  if (!stateOptions.value.length || !cycleOptions.value.length) {
    try {
      const [statesRes, cyclesRes, typesRes] = await Promise.all([
        api.get(`/projects/${props.projectId}/settings/states`),
        api.get(`/projects/${props.projectId}/cycles`),
        issueTypeApi.getIssueTypes(props.workspaceId, props.projectId)
      ])
      stateOptions.value = Array.isArray(statesRes.data) ? statesRes.data : (statesRes.data?.states || [])
      cycleOptions.value = Array.isArray(cyclesRes.data) ? cyclesRes.data : (cyclesRes.data?.items || cyclesRes.data?.cycles || [])
      issueTypeOptions.value = typesRes
    } catch (e) {
      console.error('Failed to load options:', e)
    }
  }

  const i = issue.value
  editForm.value = {
    name: i.name || '',
    state_id: i.state_id || 0,
    priority: i.priority || '',
    type_id: i.issue_type?.id || 0,
    start_date: i.start_date?.split('T')[0] || '',
    target_date: i.target_date?.split('T')[0] || '',
    cycle_id: i.cycle_id || 0,
    description_html: i.description_html || ''
  }
}

function cancelEdit() {
  editing.value = false
}

async function saveEdit() {
  if (!issue.value) return
  saving.value = true
  try {
    const data: any = {
      name: editForm.value.name,
      priority: editForm.value.priority,
      state_id: editForm.value.state_id,
      cycle_id: editForm.value.cycle_id || undefined,
      description_html: editForm.value.description_html || undefined,
    }
    if (editForm.value.type_id) data.type_id = editForm.value.type_id
    if (editForm.value.start_date) data.start_date = editForm.value.start_date + 'T00:00:00Z'
    if (editForm.value.target_date) data.target_date = editForm.value.target_date + 'T00:00:00Z'
    const updated = await issueApi.updateIssue(issue.value.id, data)
    if (panelCustomFieldRef.value) {
      await panelCustomFieldRef.value.saveValues()
    }
    issue.value = { ...issue.value, ...updated }
    editing.value = false
    emit('refresh')
    emit('close')
  } catch (e) {
    console.error('Failed to save issue:', e)
    alert(t('issueDetail.saveFailed'))
  } finally {
    saving.value = false
  }
}

function close() {
  emit('close')
}

async function deleteIssue() {
  if (issue.value && await confirm(t('issueDetail.confirmDelete').replace('{0}', issue.value.name))) {
    emit('delete', issue.value)
  }
}

function formatDate(d: string) {
  if (!d) return '-'
  return new Date(d).toLocaleString(locale.value, { hour12: false })
}

function priorityClass(p: string) {
  const m: Record<string, string> = {
    urgent: 'bg-red-50 text-red-700', high: 'bg-orange-50 text-orange-700',
    medium: 'bg-yellow-50 text-yellow-700', low: 'bg-green-50 text-green-700',
    none: 'bg-gray-50 text-gray-600'
  }
  return m[p] || m.none
}

function priorityDot(p: string) {
  const m: Record<string, string> = {
    urgent: 'bg-red-500', high: 'bg-orange-500', medium: 'bg-yellow-500',
    low: 'bg-green-500', none: 'bg-gray-400'
  }
  return m[p] || m.none
}

function priorityLabel(p: string) {
  const m: Record<string, string> = {
    urgent: t('issue.priorityUrgent'), high: t('issue.priorityHigh'), medium: t('issue.priorityMedium'), low: t('issue.priorityLow'), none: t('issue.priorityNone')
  }
  return m[p] || p
}
</script>

<style scoped>
@keyframes slideIn {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}
.animate-slide-in {
  animation: slideIn 0.25s ease-out;
}
</style>
