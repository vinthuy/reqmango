<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { skillApi, type SkillResponse, type SkillCreate, type SkillUpdate } from '@/api/skill'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const workspaceId = ref(0)
const skills = ref<SkillResponse[]>([])
const loading = ref(true)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const saving = ref(false)

const newSkill = ref<SkillCreate>({
  name: '',
  description: '',
  category: 'general',
  parameters: [],
  steps: [],
  output_format: '',
  version: '1.0.0',
  is_shared: false
})

const editingSkill = ref<SkillResponse | null>(null)
const editForm = ref<SkillUpdate>({})

const categories = [
  { value: 'general', label: 'General' },
  { value: 'code', label: 'Code' },
  { value: 'test', label: 'Testing' },
  { value: 'analysis', label: 'Analysis' },
  { value: 'document', label: 'Documentation' },
  { value: 'devops', label: 'DevOps' }
]

const filteredSkills = computed(() => {
  return skills.value.sort((a, b) => {
    if (a.is_shared && !b.is_shared) return -1
    if (!a.is_shared && b.is_shared) return 1
    if (a.status === 'active' && b.status !== 'active') return -1
    if (a.status !== 'active' && b.status === 'active') return 1
    return b.usage_count - a.usage_count
  })
})

async function loadSkills() {
  loading.value = true
  try {
    const wsId = parseInt(route.params.wsParam as string, 10)
    workspaceId.value = wsId
    skills.value = await skillApi.list(wsId) || []
  } catch (err) {
    console.error('Failed to load skills:', err)
    skills.value = []
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  newSkill.value = {
    name: '',
    description: '',
    category: 'general',
    parameters: [],
    steps: [],
    output_format: '',
    version: '1.0.0',
    is_shared: false
  }
  showCreateModal.value = true
}

function openEditModal(skill: SkillResponse) {
  editingSkill.value = skill
  editForm.value = {
    name: skill.name,
    description: skill.description,
    category: skill.category,
    parameters: skill.parameters,
    steps: skill.steps,
    output_format: skill.output_format,
    version: skill.version,
    is_shared: skill.is_shared,
    status: skill.status
  }
  showEditModal.value = true
}

async function handleCreate() {
  if (!newSkill.value.name) {
    alert(t('ai.skill.nameRequired') || 'Name is required')
    return
  }
  saving.value = true
  try {
    await skillApi.create(workspaceId.value, newSkill.value)
    showCreateModal.value = false
    await loadSkills()
  } catch (err: any) {
    console.error('Failed to create skill:', err)
    alert(err.response?.data?.message || t('ai.skill.createFailed') || 'Failed to create skill')
  } finally {
    saving.value = false
  }
}

async function handleUpdate() {
  if (!editingSkill.value) return
  if (!editForm.value.name) {
    alert(t('ai.skill.nameRequired') || 'Name is required')
    return
  }
  saving.value = true
  try {
    await skillApi.update(workspaceId.value, editingSkill.value.id, editForm.value)
    showEditModal.value = false
    await loadSkills()
  } catch (err: any) {
    console.error('Failed to update skill:', err)
    alert(err.response?.data?.message || t('ai.skill.updateFailed') || 'Failed to update skill')
  } finally {
    saving.value = false
  }
}

async function handleDelete(skillId: number) {
  if (!confirm(t('ai.skill.deleteConfirm') || 'Are you sure you want to delete this skill?')) {
    return
  }
  try {
    await skillApi.delete(workspaceId.value, skillId)
    await loadSkills()
  } catch (err: any) {
    console.error('Failed to delete skill:', err)
    alert(err.response?.data?.message || t('ai.skill.deleteFailed') || 'Failed to delete skill')
  }
}

function goBack() {
  router.push(`/workspaces/${workspaceId.value}/agents`)
}

function getStatusColor(status: string) {
  switch (status) {
    case 'active': return 'bg-green-100 text-green-700'
    case 'inactive': return 'bg-gray-100 text-gray-500'
    default: return 'bg-blue-100 text-blue-700'
  }
}

function getStatusText(status: string) {
  switch (status) {
    case 'active': return t('ai.skill.active') || 'Active'
    case 'inactive': return t('ai.skill.inactive') || 'Inactive'
    default: return status
  }
}

function getCategoryLabel(category: string) {
  return categories.find(c => c.value === category)?.label || category
}

function parseParameters(value: any) {
  try {
    return typeof value === 'string' ? JSON.parse(value) : value
  } catch {
    return value
  }
}

function parseSteps(value: any) {
  try {
    return typeof value === 'string' ? JSON.parse(value) : value
  } catch {
    return value
  }
}

onMounted(loadSkills)
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <button @click="goBack" class="text-gray-500 hover:text-gray-700">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h1 class="text-xl font-semibold text-gray-800">{{ t('ai.skill.title') || 'Skills' }}</h1>
            <p class="text-sm text-gray-500">{{ t('ai.skill.description') || 'Manage reusable AI skills and tools' }}</p>
          </div>
        </div>
        <button
          @click="openCreateModal"
          class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors shadow-sm"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          {{ t('ai.skill.create') || 'Create Skill' }}
        </button>
      </div>
    </header>

    <!-- Main Content -->
    <main class="p-6">
      <div class="max-w-5xl mx-auto">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-20">
          <div class="animate-spin rounded-full h-8 w-8 border-2 border-indigo-600 border-t-transparent"></div>
        </div>

        <!-- Empty State -->
        <div v-else-if="skills.length === 0" class="text-center py-20">
          <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h3 class="text-base font-medium text-gray-900 dark:text-gray-100 mb-1">{{ t('ai.skill.noSkills') || 'No Skills' }}</h3>
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{{ t('ai.skill.noSkillsHint') || 'Create your first skill to get started' }}</p>
          <button
            @click="openCreateModal"
            class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
          >
            {{ t('ai.skill.createFirst') || 'Create First Skill' }}
          </button>
        </div>

        <!-- Skill Grid -->
        <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div
            v-for="skill in filteredSkills"
            :key="skill.id"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start justify-between mb-3">
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white text-xl">
                  🛠️
                </div>
                <div class="min-w-0">
                  <div class="flex items-center space-x-2">
                    <h3 class="font-semibold text-gray-900 truncate">{{ skill.name }}</h3>
                    <span v-if="skill.is_shared" class="px-2 py-0.5 bg-indigo-100 text-indigo-700 rounded text-xs font-medium">
                      {{ t('ai.skill.shared') || 'Shared' }}
                    </span>
                  </div>
                  <div class="flex items-center space-x-2 mt-0.5">
                    <span class="text-xs text-gray-400">{{ getCategoryLabel(skill.category) }}</span>
                    <span class="text-gray-400">|</span>
                    <span class="text-xs text-gray-400">v{{ skill.version }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-1">
                <span :class="['px-2 py-0.5 rounded text-xs font-medium', getStatusColor(skill.status)]">
                  {{ getStatusText(skill.status) }}
                </span>
                <button
                  @click="openEditModal(skill)"
                  class="p-2 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                  :title="t('common.edit') || 'Edit'"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </button>
                <button
                  @click="handleDelete(skill.id)"
                  class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  :title="t('common.delete') || 'Delete'"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
            <p v-if="skill.description" class="text-sm text-gray-500 line-clamp-2 mb-3">{{ skill.description }}</p>
            <div class="flex flex-wrap gap-2 mb-3">
              <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600">
                {{ skill.parameters.length }} {{ t('ai.skill.params') || 'params' }}
              </span>
              <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600">
                {{ skill.steps.length }} {{ t('ai.skill.stepsCount') || 'steps' }}
              </span>
              <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600">
                {{ skill.usage_count }} {{ t('ai.skill.usage') || 'uses' }}
              </span>
            </div>
            <div class="mt-3 pt-3 border-t border-gray-100 text-xs text-gray-400">
              {{ t('ai.skill.updatedAt') || 'Updated' }}: {{ new Date(skill.updated_at).toLocaleDateString() }}
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showCreateModal = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold text-gray-900">{{ t('ai.skill.create') || 'Create Skill' }}</h3>
          <button @click="showCreateModal = false" class="text-gray-400 hover:text-gray-600">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.name') || 'Name' }}</label>
            <input
              v-model="newSkill.name"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              :placeholder="t('ai.skill.namePlaceholder') || 'Enter skill name'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.description') || 'Description' }}</label>
            <textarea
              v-model="newSkill.description"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              rows="2"
              :placeholder="t('ai.skill.descriptionPlaceholder') || 'Enter description'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.skill.category') || 'Category' }}</label>
            <select
              v-model="newSkill.category"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            >
              <option v-for="c in categories" :key="c.value" :value="c.value">{{ c.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.skill.parameters') || 'Parameters (JSON)' }}</label>
            <textarea
              v-model="newSkill.parameters"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              rows="3"
              :placeholder="t('ai.skill.parametersPlaceholder') || 'Enter JSON array'"
              @blur="newSkill.parameters = parseParameters(newSkill.parameters)"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.skill.steps') || 'Steps (JSON)' }}</label>
            <textarea
              v-model="newSkill.steps"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              rows="3"
              :placeholder="t('ai.skill.stepsPlaceholder') || 'Enter JSON array'"
              @blur="newSkill.steps = parseSteps(newSkill.steps)"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.skill.outputFormat') || 'Output Format' }}</label>
            <textarea
              v-model="newSkill.output_format"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              rows="2"
              :placeholder="t('ai.skill.outputFormatPlaceholder') || 'JSON schema or description'"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.skill.version') || 'Version' }}</label>
            <input
              v-model="newSkill.version"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              :placeholder="t('ai.skill.versionPlaceholder') || '1.0.0'"
            />
          </div>
          <div class="flex items-center">
            <input
              v-model="newSkill.is_shared"
              type="checkbox"
              class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
            />
            <label class="ml-2 text-sm text-gray-700">{{ t('ai.skill.isShared') || 'Make this skill shared' }}</label>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showCreateModal = false"
            class="px-4 py-2.5 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition text-sm font-medium"
          >
            {{ t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="handleCreate"
            :disabled="saving"
            class="px-4 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition text-sm font-medium"
          >
            {{ saving ? (t('common.saving') || 'Saving...') : (t('common.create') || 'Create') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showEditModal = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-semibold text-gray-900">{{ t('ai.skill.edit') || 'Edit Skill' }}</h3>
          <button @click="showEditModal = false" class="text-gray-400 hover:text-gray-600">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.name') || 'Name' }}</label>
            <input
              v-model="editForm.name"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.description') || 'Description' }}</label>
            <textarea
              v-model="editForm.description"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
              rows="2"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.skill.category') || 'Category' }}</label>
            <select
              v-model="editForm.category"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            >
              <option v-for="c in categories" :key="c.value" :value="c.value">{{ c.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.skill.parameters') || 'Parameters (JSON)' }}</label>
            <textarea
              v-model="editForm.parameters"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              rows="3"
              @blur="editForm.parameters = parseParameters(editForm.parameters)"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.skill.steps') || 'Steps (JSON)' }}</label>
            <textarea
              v-model="editForm.steps"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              rows="3"
              @blur="editForm.steps = parseSteps(editForm.steps)"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.skill.outputFormat') || 'Output Format' }}</label>
            <textarea
              v-model="editForm.output_format"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm font-mono"
              rows="2"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('ai.skill.version') || 'Version' }}</label>
            <input
              v-model="editForm.version"
              type="text"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('common.status') || 'Status' }}</label>
            <select
              v-model="editForm.status"
              class="w-full px-3.5 py-2.5 border border-gray-300 rounded-lg bg-white text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm"
            >
              <option value="active">{{ t('ai.skill.active') || 'Active' }}</option>
              <option value="inactive">{{ t('ai.skill.inactive') || 'Inactive' }}</option>
            </select>
          </div>
          <div class="flex items-center">
            <input
              v-model="editForm.is_shared"
              type="checkbox"
              class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
            />
            <label class="ml-2 text-sm text-gray-700">{{ t('ai.skill.isShared') || 'Make this skill shared' }}</label>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="showEditModal = false"
            class="px-4 py-2.5 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition text-sm font-medium"
          >
            {{ t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="handleUpdate"
            :disabled="saving"
            class="px-4 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition text-sm font-medium"
          >
            {{ saving ? (t('common.saving') || 'Saving...') : (t('common.save') || 'Save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>