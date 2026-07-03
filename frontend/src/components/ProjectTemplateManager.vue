<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div><h1 class="text-xl font-semibold text-gray-900">{{ t('template.title') }}</h1><p class="text-sm text-gray-500 mt-1">{{ t('template.desc') }}</p></div>
      <button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">{{ t('template.create') }}</button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="tmpl in templates" :key="tmpl.id" class="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition">
        <div class="flex items-center space-x-3 mb-3">
          <div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center text-white text-lg">📋</div>
          <div>
            <h3 class="font-medium text-gray-900">{{ tmpl.name }}</h3>
            <div class="flex items-center space-x-2 text-xs text-gray-400">
              <span>{{ tmpl.types?.length || 0 }} {{ t('template.types') }}</span>
              <span v-if="parsedStates(tmpl).length" class="text-gray-300">·</span>
              <span v-if="parsedStates(tmpl).length">{{ parsedStates(tmpl).length }} {{ t('template.states') }}</span>
              <span v-if="parsedLabels(tmpl).length" class="text-gray-300">·</span>
              <span v-if="parsedLabels(tmpl).length">{{ parsedLabels(tmpl).length }} {{ t('template.labels') }}</span>
            </div>
          </div>
        </div>
        <p v-if="tmpl.description" class="text-sm text-gray-500 mb-3">{{ tmpl.description }}</p>
        <div class="space-y-1 mb-3">
          <div v-for="t in tmpl.types" :key="t.type_template_id" class="flex items-center text-xs text-gray-600">
            <span class="w-2 h-2 rounded-full mr-1.5" :style="{backgroundColor:t.type_color||'#6366F1'}"></span>
            <span>{{ t.type_name }}</span>
            <span class="text-gray-400 ml-1">L{{ t.type_level || 0 }}</span>
            <span v-if="t.is_required" class="text-red-400 ml-1">*</span>
          </div>
        </div>
        <!-- States preview pills -->
        <div v-if="parsedStates(tmpl).length" class="flex flex-wrap gap-1 mb-3">
          <span
            v-for="s in parsedStates(tmpl)"
            :key="s.name"
            class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
            :style="{ backgroundColor: s.color + '20', color: s.color, border: '1px solid ' + s.color + '40' }"
          >{{ s.name }}</span>
        </div>
        <div class="pt-3 border-t border-gray-100 flex space-x-2">
          <button @click="openAddTypes(tmpl)" class="text-xs text-indigo-600 hover:text-indigo-800">{{ t('template.addTypes') }}</button>
          <button @click="openApplyModal(tmpl)" class="text-xs text-green-600 hover:text-green-800">{{ t('template.apply') }}</button>
          <button @click="confirmDelete(tmpl)" class="text-xs text-red-500 hover:text-red-700">{{ t('common.delete') }}</button>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showModal=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-lg max-h-[85vh] overflow-y-auto">
        <h3 class="text-lg font-semibold mb-4">{{ t('template.create') }}</h3>
        <div class="space-y-3">
          <div><label class="block text-sm font-medium mb-1">{{ t('template.name') }} *</label><input v-model="form.name" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div><label class="block text-sm font-medium mb-1">{{ t('template.description') }}</label><input v-model="form.description" class="w-full px-3 py-2 border rounded-lg" /></div>

          <!-- States Section -->
          <div class="border-t pt-4 mt-4">
            <h4 class="text-sm font-semibold text-gray-700 mb-2">{{ t('template.preConfiguredStates') }}</h4>
            <div class="space-y-2">
              <div v-for="(st, idx) in templateStates" :key="idx" class="flex items-center space-x-2 bg-gray-50 p-2 rounded-lg">
                <input v-model="st.name" :placeholder="t('template.stateName')" class="flex-1 px-2 py-1.5 border rounded text-sm" />
                <input type="color" v-model="st.color" class="w-8 h-8 border rounded cursor-pointer p-0.5" />
                <select v-model="st.group" class="px-2 py-1.5 border rounded text-sm">
                  <option value="backlog">{{ t('template.stateGroup.backlog') }}</option>
                  <option value="unstarted">{{ t('template.stateGroup.unstarted') }}</option>
                  <option value="started">{{ t('template.stateGroup.started') }}</option>
                  <option value="completed">{{ t('template.stateGroup.completed') }}</option>
                  <option value="cancelled">{{ t('template.stateGroup.cancelled') }}</option>
                </select>
                <button @click="removeTemplateState(idx)" class="text-red-400 hover:text-red-600 text-lg leading-none px-1">&times;</button>
              </div>
            </div>
            <button @click="addTemplateState" class="mt-2 text-xs text-blue-600 hover:text-blue-800">{{ t('template.addState') }}</button>
          </div>

          <!-- Labels Section -->
          <div class="border-t pt-4 mt-4">
            <h4 class="text-sm font-semibold text-gray-700 mb-2">{{ t('template.preConfiguredLabels') }}</h4>
            <div class="space-y-2">
              <div v-for="(lb, idx) in templateLabels" :key="idx" class="flex items-center space-x-2 bg-gray-50 p-2 rounded-lg">
                <input v-model="lb.name" :placeholder="t('template.labelName')" class="flex-1 px-2 py-1.5 border rounded text-sm" />
                <input type="color" v-model="lb.color" class="w-8 h-8 border rounded cursor-pointer p-0.5" />
                <button @click="removeTemplateLabel(idx)" class="text-red-400 hover:text-red-600 text-lg leading-none px-1">&times;</button>
              </div>
            </div>
            <button @click="addTemplateLabel" class="mt-2 text-xs text-blue-600 hover:text-blue-800">{{ t('template.addLabel') }}</button>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showModal=false" class="px-4 py-2 border rounded-lg">{{ t('common.cancel') }}</button>
          <button @click="saveTemplate" class="px-4 py-2 bg-blue-600 text-white rounded-lg">{{ t('common.create') }}</button>
        </div>
      </div>
    </div>

    <!-- Apply Template Modal -->
    <div v-if="showApplyModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="closeApplyModal">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold mb-4">{{ t('template.createFromTemplate') }}: {{ templateToApply?.name }}</h3>
        <div v-if="applySuccess" class="text-center py-4">
          <div class="text-green-500 text-4xl mb-2">✓</div>
          <p class="text-gray-700 font-medium mb-4">{{ applySuccessMessage }}</p>
          <div class="flex justify-center space-x-3">
            <button @click="goToProject" class="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm">{{ t('template.goToProject') }}</button>
            <button @click="closeApplyModal" class="px-4 py-2 border rounded-lg text-sm">{{ t('common.close') }}</button>
          </div>
        </div>
        <div v-else>
          <div class="space-y-3">
            <div><label class="block text-sm font-medium mb-1">{{ t('template.projectName') }} *</label><input v-model="applyProjectName" class="w-full px-3 py-2 border rounded-lg" /></div>
            <div><label class="block text-sm font-medium mb-1">{{ t('template.projectIdentifier') }} *</label><input v-model="applyProjectIdentifier" class="w-full px-3 py-2 border rounded-lg" :placeholder="t('template.projectIdentifierPlaceholder')" /></div>
          </div>
          <div class="flex justify-end space-x-3 mt-6">
            <button @click="closeApplyModal" class="px-4 py-2 border rounded-lg" :disabled="applyLoading">{{ t('common.cancel') }}</button>
            <button @click="doApplyTemplate" class="px-4 py-2 bg-green-600 text-white rounded-lg disabled:opacity-50" :disabled="applyLoading || !applyProjectName.trim() || !applyProjectIdentifier.trim()">
              {{ applyLoading ? t('common.creating') : t('template.createProject') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Add Types Modal -->
    <div v-if="showTypeModal && selected" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showTypeModal=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-lg">
        <h3 class="text-lg font-semibold mb-4">{{ t('template.addTypesTo') }} {{ selected.name }}</h3>
        <div class="space-y-2 max-h-64 overflow-y-auto">
          <div v-for="tt in availableTypeTemplates" :key="tt.id" class="flex items-center justify-between p-2 border rounded">
            <div class="flex items-center space-x-2">
              <span class="w-2 h-2 rounded-full" :style="{backgroundColor:tt.color}"></span>
              <span class="text-sm font-medium">{{ tt.name }}</span>
              <span class="text-xs text-gray-400">L{{ tt.level }}</span>
            </div>
            <div class="flex items-center space-x-2">
              <label class="text-xs"><input type="checkbox" v-model="typeRequired[tt.id]" /> {{ t('template.required') }}</label>
              <button @click="addType(tt)" class="text-xs text-indigo-600 hover:text-indigo-800">{{ t('common.add') }}</button>
            </div>
          </div>
          <div v-if="!availableTypeTemplates.length" class="text-sm text-gray-400 text-center py-4">{{ t('template.allTypesAdded') }}</div>
        </div>
        <div class="flex justify-end mt-4"><button @click="showTypeModal=false" class="px-4 py-2 border rounded-lg">{{ t('common.done') }}</button></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import templateApi from '@/api/template'
import * as issueTypeApi from '@/api/issue-type'
import { createProject } from '@/api/project'
import { useConfirm } from '@/composables/useConfirm'
import { useI18n } from '@/composables/useI18n'

const props = defineProps<{ workspaceId: number }>()

const router = useRouter()
const { confirm } = useConfirm()
const { t } = useI18n()

interface StateEntry {
  name: string
  color: string
  group: string
  sequence: number
}

interface LabelEntry {
  name: string
  color: string
}

const DEFAULT_STATES: StateEntry[] = [
  { name: t('template.backlog'), color: '#6B7280', group: 'backlog', sequence: 1 },
  { name: t('template.todo'), color: '#3B82F6', group: 'unstarted', sequence: 2 },
  { name: t('template.inProgress'), color: '#F59E0B', group: 'started', sequence: 3 },
  { name: t('template.inReview'), color: '#8B5CF6', group: 'started', sequence: 4 },
  { name: t('template.done'), color: '#10B981', group: 'completed', sequence: 5 },
  { name: t('template.cancelled'), color: '#EF4444', group: 'cancelled', sequence: 6 },
]

const templates = ref<any[]>([])
const typeTemplates = ref<any[]>([])
const showModal = ref(false)
const showTypeModal = ref(false)
const selected = ref<any>(null)
const typeRequired = ref<Record<number,boolean>>({})
const form = ref({ name:'', description:'' })
const templateStates = ref<StateEntry[]>([])
const templateLabels = ref<LabelEntry[]>([])

// Apply template modal state
const showApplyModal = ref(false)
const templateToApply = ref<any>(null)
const applyProjectName = ref('')
const applyProjectIdentifier = ref('')
const applyLoading = ref(false)
const applySuccess = ref(false)
const applySuccessMessage = ref('')
const applyCreatedProjectId = ref<number | null>(null)

function parsedStates(tmpl: any): StateEntry[] {
  try {
    if (tmpl.states) return JSON.parse(tmpl.states)
  } catch {}
  return []
}

function parsedLabels(tmpl: any): LabelEntry[] {
  try {
    if (tmpl.labels) return JSON.parse(tmpl.labels)
  } catch {}
  return []
}

const availableTypeTemplates = computed(() =>
  typeTemplates.value.filter((tt:any) =>
    !(selected.value?.types||[]).some((t:any) => (t.type_template_id || t.issue_type_id) === tt.id)
  )
)

async function load() {
  try {
    const [tmpl, types] = await Promise.all([
      templateApi.listTemplates(props.workspaceId),
      issueTypeApi.getIssueTypes(props.workspaceId),
    ])
    templates.value = tmpl; typeTemplates.value = types
  } catch(e) { console.error(e) }
}

function openCreate() {
  form.value = { name:'', description:'' }
  templateStates.value = DEFAULT_STATES.map(s => ({ ...s }))
  templateLabels.value = []
  showModal.value = true
}

function addTemplateState() {
  templateStates.value.push({ name: '', color: '#3B82F6', group: 'unstarted', sequence: templateStates.value.length + 1 })
}

function removeTemplateState(idx: number) {
  templateStates.value.splice(idx, 1)
}

function addTemplateLabel() {
  templateLabels.value.push({ name: '', color: '#EF4444' })
}

function removeTemplateLabel(idx: number) {
  templateLabels.value.splice(idx, 1)
}

async function saveTemplate() {
  const payload: any = { ...form.value }
  if (templateStates.value.length) {
    payload.states = JSON.stringify(templateStates.value.map((s, i) => ({ ...s, sequence: i + 1 })))
  }
  if (templateLabels.value.length) {
    payload.labels = JSON.stringify(templateLabels.value)
  }
  await templateApi.createTemplate(props.workspaceId, payload)
  showModal.value = false; load()
}

function openAddTypes(tmpl: any) { selected.value = tmpl; showTypeModal.value = true }
async function addType(tt: any) {
  if (!selected.value) return
  await templateApi.addTypeToTemplate(selected.value.id, { issue_type_id: tt.id, is_required: !!typeRequired.value[tt.id], sequence: 1 })
  typeRequired.value[tt.id] = false; load().then(() => { selected.value = templates.value.find(t=>t.id===selected.value!.id) })
}

// Apply template - opens modal instead of prompt()
function openApplyModal(tmpl: any) {
  templateToApply.value = tmpl
  applyProjectName.value = ''
  applyProjectIdentifier.value = ''
  applyLoading.value = false
  applySuccess.value = false
  applySuccessMessage.value = ''
  applyCreatedProjectId.value = null
  showApplyModal.value = true
}

function closeApplyModal() {
  showApplyModal.value = false
  applyCreatedProjectId.value = null
}

async function doApplyTemplate() {
  if (!templateToApply.value || !applyProjectName.value.trim() || !applyProjectIdentifier.value.trim()) return
  applyLoading.value = true
  try {
    const project = await createProject(props.workspaceId, {
      name: applyProjectName.value.trim(),
      identifier: applyProjectIdentifier.value.trim(),
      template_id: templateToApply.value.id,
    })
    applyCreatedProjectId.value = project.id
    applySuccessMessage.value = `Project "${project.name}" has been created successfully.`
    applySuccess.value = true
  } catch (e) {
    console.error(e)
    applySuccessMessage.value = 'Failed to create project. Please try again.'
    applySuccess.value = true
  } finally {
    applyLoading.value = false
  }
}

function goToProject() {
  if (applyCreatedProjectId.value) {
    router.push(`/workspaces/${props.workspaceId}/projects/${applyCreatedProjectId.value}/issues`)
  }
  closeApplyModal()
}

async function confirmDelete(tmpl: any) {
  if (await confirm({
    title: t('template.deleteTitle'),
    message: t('relationType.deleteConfirm').replace('{name}', tmpl.name),
    confirmText: t('common.delete'),
    danger: true,
  })) {
    await templateApi.deleteTemplate(tmpl.id); load()
  }
}

onMounted(load)
</script>
