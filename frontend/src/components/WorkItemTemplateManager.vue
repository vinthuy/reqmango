<template>
  <div class="work-item-template-manager">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h3 class="text-lg font-semibold text-gray-800">{{ t('workItemTemplate.title') }}</h3>
        <p class="text-sm text-gray-500 mt-1">{{ t('workItemTemplate.description') }}</p>
      </div>
      <button
        class="px-4 py-2 bg-neutral-900 text-white text-sm rounded-md hover:bg-neutral-800 transition-colors"
        @click="showForm = true"
      >
        {{ t('common.add') }}
      </button>
    </div>

    <div v-if="templates.length === 0" class="bg-white border border-gray-200 rounded-lg p-6">
      <div class="text-center">
        <div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
        <h4 class="text-sm font-medium text-gray-700 mb-2">{{ t('workItemTemplate.noTemplates') }}</h4>
        <p class="text-sm text-gray-500">{{ t('workItemTemplate.noTemplatesDesc') }}</p>
      </div>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="template in templates"
        :key="template.id"
        class="bg-white border border-gray-200 rounded-lg p-4 hover:border-gray-300 transition-colors"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="flex items-center gap-2">
              <span class="font-medium text-gray-800">{{ template.name }}</span>
              <span v-if="template.is_default" class="px-2 py-0.5 bg-blue-100 text-blue-700 text-xs font-medium rounded-full">
                {{ t('workItemTemplate.default') }}
              </span>
            </div>
            <p v-if="template.description" class="text-sm text-gray-500 mt-1">{{ template.description }}</p>
            <div class="flex items-center gap-4 mt-2">
              <span v-if="template.issue_type" class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium" :style="{ backgroundColor: template.issue_type.color + '20', color: template.issue_type.color }">
                {{ template.issue_type.name }}
              </span>
              <span class="text-xs text-gray-400">{{ t('workItemTemplate.fields') }}: {{ Object.keys(template.defaults).length }}</span>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button class="text-gray-400 hover:text-blue-500" @click="editTemplate(template)">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
            </button>
            <button v-if="!template.is_default" class="text-gray-400 hover:text-red-500" @click="deleteTemplate(template)">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showForm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="closeForm">
      <div class="bg-white rounded-lg w-full max-w-lg mx-4 p-6 max-h-[90vh] overflow-y-auto">
        <h4 class="text-lg font-semibold text-gray-800 mb-4">{{ editingTemplate ? t('workItemTemplate.edit') : t('workItemTemplate.create') }}</h4>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('common.name') }}</label>
            <input type="text" v-model="form.name" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" :placeholder="t('workItemTemplate.namePlaceholder')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('common.description') }}</label>
            <textarea v-model="form.description" rows="2" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" :placeholder="t('workItemTemplate.descriptionPlaceholder')"></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('workItemTemplate.issueType') }}</label>
            <select v-model="form.issue_type_id" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500">
              <option :value="null">{{ t('common.all') }}</option>
              <option v-for="type in issueTypes" :key="type.id" :value="type.id">{{ type.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('workItemTemplate.defaultFields') }}</label>
            <div class="grid grid-cols-2 gap-3">
              <label class="flex items-center gap-2 p-2 border border-gray-200 rounded-md cursor-pointer hover:bg-gray-50">
                <input type="text" v-model="form.defaults.name_prefix" class="flex-1 px-2 py-1 text-sm border-0 focus:ring-0" :placeholder="t('workItemTemplate.title')" />
              </label>
              <label class="flex items-center gap-2 p-2 border border-gray-200 rounded-md cursor-pointer hover:bg-gray-50">
                <input type="text" v-model="form.defaults.priority" class="flex-1 px-2 py-1 text-sm border-0 focus:ring-0" :placeholder="t('workItemTemplate.priority')" />
              </label>
              <label class="flex items-center gap-2 p-2 border border-gray-200 rounded-md cursor-pointer hover:bg-gray-50">
                <input type="text" v-model="form.defaults.state_id" class="flex-1 px-2 py-1 text-sm border-0 focus:ring-0" :placeholder="t('workItemTemplate.stateId')" />
              </label>
            </div>
            <div class="mt-3">
              <textarea v-model="form.defaults.description_html" class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" rows="3" :placeholder="t('workItemTemplate.description')"></textarea>
            </div>
          </div>
          <div>
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="form.is_default" class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" />
              <span class="text-sm text-gray-700">{{ t('workItemTemplate.setAsDefault') }}</span>
            </label>
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-6">
          <button class="px-4 py-2 text-sm text-gray-600 hover:text-gray-800" @click="closeForm">
            {{ t('common.cancel') }}
          </button>
          <button class="px-4 py-2 bg-neutral-900 text-white text-sm rounded-md hover:bg-neutral-800 transition-colors" @click="saveTemplate">
            {{ t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
import { listWorkItemTemplates, createWorkItemTemplate, updateWorkItemTemplate, deleteWorkItemTemplate } from '@/api/work-item-template'
import type { WorkItemTemplate } from '@/types/work-item-template'
import { getIssueTypes } from '@/api/issue-type'
import type { IssueType } from '@/types/issue-type'

const props = defineProps<{
  projectId: number
  workspaceId: number
}>()

const { t } = useI18n()
const toast = useToast()
const { confirm } = useConfirm()

const templates = ref<WorkItemTemplate[]>([])
const issueTypes = ref<IssueType[]>([])
const showForm = ref(false)
const editingTemplate = ref<WorkItemTemplate | null>(null)

const form = ref({
  name: '',
  description: '',
  issue_type_id: null as number | null,
  defaults: {
    name_prefix: '',
    description_html: '',
    priority: '',
    state_id: '',
  },
  is_default: false,
})

async function loadTemplates() {
  templates.value = await listWorkItemTemplates(props.projectId)
}

async function loadIssueTypes() {
  issueTypes.value = await getIssueTypes(props.workspaceId, props.projectId)
}

function closeForm() {
  showForm.value = false
  editingTemplate.value = null
  form.value = {
    name: '',
    description: '',
    issue_type_id: null,
    defaults: {
      name_prefix: '',
      description_html: '',
      priority: '',
      state_id: '',
    },
    is_default: false,
  }
}

function editTemplate(template: WorkItemTemplate) {
  editingTemplate.value = template
  form.value = {
    name: template.name,
    description: template.description || '',
    issue_type_id: template.issue_type_id || null,
    defaults: {
      name_prefix: template.defaults.name_prefix || '',
      description_html: template.defaults.description_html || '',
      priority: template.defaults.priority || '',
      state_id: template.defaults.state_id || '',
    },
    is_default: template.is_default,
  }
  showForm.value = true
}

async function saveTemplate() {
  if (!form.value.name.trim()) {
    toast.error(t('workItemTemplate.nameRequired'))
    return
  }

  try {
    const data = {
      name: form.value.name.trim(),
      description: form.value.description.trim() || undefined,
      issue_type_id: form.value.issue_type_id || undefined,
      defaults: Object.fromEntries(
        Object.entries(form.value.defaults).filter(([, v]) => v.trim())
      ),
      is_default: form.value.is_default,
    }

    if (editingTemplate.value) {
      await updateWorkItemTemplate(props.projectId, editingTemplate.value.id, data)
      toast.success(t('workItemTemplate.updateSuccess'))
    } else {
      await createWorkItemTemplate(props.projectId, data)
      toast.success(t('workItemTemplate.createSuccess'))
    }

    closeForm()
    await loadTemplates()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || t('common.error'))
  }
}

async function deleteTemplate(template: WorkItemTemplate) {
  const confirmed = await confirm(t('workItemTemplate.deleteConfirm', { name: template.name }))
  if (!confirmed) return
  try {
    await deleteWorkItemTemplate(props.projectId, template.id)
    toast.success(t('workItemTemplate.deleteSuccess'))
    await loadTemplates()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || t('common.error'))
  }
}

onMounted(() => {
  loadTemplates()
  loadIssueTypes()
})
</script>
