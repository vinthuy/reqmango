<template>
  <div class="p-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('ai.memory.title') }}</h1>
        <p class="text-gray-500 mt-1">{{ t('ai.memory.description') }}</p>
      </div>
      <button
        @click="openCreateModal"
        class="bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white px-4 py-2 rounded-lg font-medium transition"
      >
        {{ t('ai.common.create') }}
      </button>
    </div>

    <!-- Statistics Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-white border border-gray-200 rounded-xl p-4">
        <div class="text-2xl font-bold text-gray-900">{{ totalCount }}</div>
        <div class="text-sm text-gray-500">{{ t('ai.memory.total') }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-xl p-4">
        <div class="text-2xl font-bold text-blue-600">{{ shortTermCount }}</div>
        <div class="text-sm text-gray-500">{{ t('ai.memory.shortTerm') }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-xl p-4">
        <div class="text-2xl font-bold text-green-600">{{ mediumTermCount }}</div>
        <div class="text-sm text-gray-500">{{ t('ai.memory.mediumTerm') }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-xl p-4">
        <div class="text-2xl font-bold text-purple-600">{{ longTermCount }}</div>
        <div class="text-sm text-gray-500">{{ t('ai.memory.longTerm') }}</div>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="bg-white border border-gray-200 rounded-xl p-4 mb-6">
      <div class="flex flex-wrap gap-4 items-center">
        <div class="flex-1 min-w-[200px]">
          <input
            v-model="searchQuery"
            @input="handleSearch"
            type="text"
            :placeholder="t('ai.memory.searchPlaceholder')"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
        </div>
        <select
          v-model="filters.memory_type"
          @change="loadMemories"
          class="px-3 py-2 border border-gray-300 rounded-lg text-sm"
        >
          <option value="">{{ t('ai.memory.allTypes') }}</option>
          <option value="short_term">{{ t('ai.memory.shortTerm') }}</option>
          <option value="medium_term">{{ t('ai.memory.mediumTerm') }}</option>
          <option value="long_term">{{ t('ai.memory.longTerm') }}</option>
        </select>
        <select
          v-model="filters.scope"
          @change="loadMemories"
          class="px-3 py-2 border border-gray-300 rounded-lg text-sm"
        >
          <option value="">{{ t('ai.memory.allScopes') }}</option>
          <option value="workspace">{{ t('ai.memory.scopeWorkspace') }}</option>
          <option value="project">{{ t('ai.memory.scopeProject') }}</option>
          <option value="issue">{{ t('ai.memory.scopeIssue') }}</option>
          <option value="agent">{{ t('ai.memory.scopeAgent') }}</option>
        </select>
        <button
          @click="resetFilters"
          class="px-3 py-2 text-gray-600 hover:text-gray-900 text-sm"
        >
          {{ t('ai.common.reset') }}
        </button>
      </div>
    </div>

    <!-- Memory List -->
    <div class="bg-white border border-gray-200 rounded-xl overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="bg-gray-50 border-b border-gray-200">
            <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">{{ t('ai.memory.content') }}</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">{{ t('ai.memory.type') }}</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">{{ t('ai.memory.scope') }}</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">{{ t('ai.memory.relevance') }}</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">{{ t('ai.memory.tags') }}</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">{{ t('ai.common.createdAt') }}</th>
            <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 uppercase">{{ t('ai.common.actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-for="memory in memories" :key="memory.id" class="hover:bg-gray-50">
            <td class="px-4 py-3">
              <div class="max-w-md truncate text-sm text-gray-900" :title="memory.content">{{ memory.content }}</div>
              <div v-if="memory.context_name" class="text-xs text-gray-400 mt-0.5">{{ memory.context_name }}</div>
            </td>
            <td class="px-4 py-3">
              <span :class="['inline-flex items-center px-2 py-0.5 rounded text-xs font-medium', getTypeBadgeClass(memory.memory_type)]">
                {{ getTypeLabel(memory.memory_type) }}
              </span>
            </td>
            <td class="px-4 py-3">
              <span class="text-sm text-gray-600">{{ getScopeLabel(memory.scope) }}</span>
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <div class="w-24 h-2 bg-gray-200 rounded-full overflow-hidden">
                  <div
                    class="h-full bg-gradient-to-r from-green-400 to-green-600 transition-all"
                    :style="{ width: `${memory.relevance_score * 100}%` }"
                  ></div>
                </div>
                <span class="text-xs text-gray-500">{{ (memory.relevance_score * 100).toFixed(1) }}%</span>
              </div>
            </td>
            <td class="px-4 py-3">
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="tag in (memory.tags || [])"
                  :key="tag"
                  class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-gray-100 text-gray-600"
                >
                  {{ tag }}
                </span>
              </div>
            </td>
            <td class="px-4 py-3">
              <span class="text-xs text-gray-500">{{ formatDate(memory.created_at) }}</span>
            </td>
            <td class="px-4 py-3">
              <div class="flex justify-end gap-2">
                <button
                  @click="openEditModal(memory)"
                  class="text-gray-400 hover:text-indigo-600 p-1"
                  :title="t('ai.common.edit')"
                >
                  ✏️
                </button>
                <button
                  @click="confirmDelete(memory)"
                  class="text-gray-400 hover:text-red-600 p-1"
                  :title="t('ai.common.delete')"
                >
                  🗑️
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="memories.length === 0">
            <td colspan="7" class="px-4 py-8 text-center text-gray-500">
              <div class="text-4xl mb-2">🧠</div>
              <div>{{ t('ai.memory.empty') }}</div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Modal -->
    <Transition name="fade">
      <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div class="bg-white rounded-xl w-full max-w-lg mx-4">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
            <h3 class="text-lg font-semibold text-gray-900">{{ isEditing ? t('ai.memory.edit') : t('ai.memory.create') }}</h3>
            <button @click="closeModal" class="text-gray-400 hover:text-gray-600">✕</button>
          </div>
          <div class="p-6 space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.memory.content') }}</label>
              <textarea
                v-model="formData.content"
                rows="4"
                :placeholder="t('ai.memory.contentPlaceholder')"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
              ></textarea>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.memory.type') }}</label>
                <select
                  v-model="formData.memory_type"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                >
                  <option value="short_term">{{ t('ai.memory.shortTerm') }}</option>
                  <option value="medium_term">{{ t('ai.memory.mediumTerm') }}</option>
                  <option value="long_term">{{ t('ai.memory.longTerm') }}</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.memory.scope') }}</label>
                <select
                  v-model="formData.scope"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                >
                  <option value="workspace">{{ t('ai.memory.scopeWorkspace') }}</option>
                  <option value="project">{{ t('ai.memory.scopeProject') }}</option>
                  <option value="issue">{{ t('ai.memory.scopeIssue') }}</option>
                  <option value="agent">{{ t('ai.memory.scopeAgent') }}</option>
                </select>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.memory.tags') }}</label>
              <input
                v-model="tagInput"
                type="text"
                :placeholder="t('ai.memory.tagsPlaceholder')"
                @keydown.enter.prevent="addTag"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
              />
              <div class="flex flex-wrap gap-1 mt-2">
                <span
                  v-for="tag in formData.tags"
                  :key="tag"
                  class="inline-flex items-center px-2 py-0.5 rounded text-xs bg-indigo-100 text-indigo-600"
                >
                  {{ tag }}
                  <button @click="removeTag(tag)" class="ml-1 hover:text-indigo-800">✕</button>
                </span>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.memory.contextName') }}</label>
              <input
                v-model="formData.context_name"
                type="text"
                :placeholder="t('ai.memory.contextNamePlaceholder')"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
              />
            </div>
          </div>
          <div class="flex justify-end gap-3 px-6 py-4 border-t border-gray-200">
            <button @click="closeModal" class="px-4 py-2 text-gray-600 hover:text-gray-900">{{ t('ai.common.cancel') }}</button>
            <button
              @click="saveMemory"
              :disabled="!formData.content.trim()"
              class="px-4 py-2 bg-gradient-to-r from-indigo-500 to-purple-600 text-white rounded-lg disabled:opacity-50"
            >
              {{ t('ai.common.save') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useRoute } from 'vue-router'
import * as memoryApi from '@/api/memory'
import type { MemoryEntry, MemoryListFilters } from '@/api/memory'

const { t } = useI18n()
const route = useRoute()

const getWorkspaceId = () => {
  const id = route.params.wsParam
  if (Array.isArray(id)) {
    return parseInt(id[0])
  }
  return typeof id === 'string' ? parseInt(id) : 0
}

const memories = ref<MemoryEntry[]>([])
const searchQuery = ref('')
const filters = ref<MemoryListFilters>({})
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const formData = ref({
  content: '',
  memory_type: 'medium_term' as 'short_term' | 'medium_term' | 'long_term',
  scope: 'workspace' as 'workspace' | 'project' | 'issue' | 'agent',
  tags: [] as string[],
  context_name: '',
})
const tagInput = ref('')

const totalCount = computed(() => memories.value.length)
const shortTermCount = computed(() => memories.value.filter(m => m.memory_type === 'short_term').length)
const mediumTermCount = computed(() => memories.value.filter(m => m.memory_type === 'medium_term').length)
const longTermCount = computed(() => memories.value.filter(m => m.memory_type === 'long_term').length)

async function loadMemories() {
  try {
    memories.value = await memoryApi.listMemories(getWorkspaceId(), {
      ...filters.value,
      limit: 100,
    })
  } catch (e) {
    console.error('Failed to load memories', e)
  }
}

async function handleSearch() {
  if (!searchQuery.value.trim()) {
    await loadMemories()
    return
  }
  try {
    memories.value = await memoryApi.searchMemories(getWorkspaceId(), searchQuery.value, 100)
  } catch (e) {
    console.error('Search failed', e)
  }
}

function resetFilters() {
  filters.value = {}
  searchQuery.value = ''
  loadMemories()
}

function openCreateModal() {
  isEditing.value = false
  editingId.value = null
  formData.value = {
    content: '',
    memory_type: 'medium_term',
    scope: 'workspace',
    tags: [],
    context_name: '',
  }
  tagInput.value = ''
  showModal.value = true
}

function openEditModal(memory: MemoryEntry) {
  isEditing.value = true
  editingId.value = memory.id
  formData.value = {
    content: memory.content,
    memory_type: memory.memory_type,
    scope: memory.scope,
    tags: memory.tags || [],
    context_name: memory.context_name || '',
  }
  tagInput.value = ''
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

function addTag() {
  const tag = tagInput.value.trim()
  if (tag && !formData.value.tags.includes(tag)) {
    formData.value.tags.push(tag)
    tagInput.value = ''
  }
}

function removeTag(tag: string) {
  formData.value.tags = formData.value.tags.filter(t => t !== tag)
}

async function saveMemory() {
  if (!formData.value.content.trim()) return
  try {
    if (isEditing.value && editingId.value) {
      await memoryApi.updateMemory(getWorkspaceId(), editingId.value, {
        content: formData.value.content,
        memory_type: formData.value.memory_type,
        scope: formData.value.scope,
        tags: formData.value.tags,
        context_name: formData.value.context_name,
      })
    } else {
      await memoryApi.createMemory(getWorkspaceId(), {
        content: formData.value.content,
        memory_type: formData.value.memory_type,
        scope: formData.value.scope,
        tags: formData.value.tags,
        context_name: formData.value.context_name,
      })
    }
    closeModal()
    await loadMemories()
  } catch (e) {
    console.error('Save failed', e)
  }
}

async function confirmDelete(memory: MemoryEntry) {
  if (confirm(t('ai.memory.confirmDelete'))) {
    try {
      await memoryApi.deleteMemory(getWorkspaceId(), memory.id)
      await loadMemories()
    } catch (e) {
      console.error('Delete failed', e)
    }
  }
}

function getTypeLabel(type: string) {
  const labels: Record<string, string> = {
    short_term: t('ai.memory.shortTerm'),
    medium_term: t('ai.memory.mediumTerm'),
    long_term: t('ai.memory.longTerm'),
  }
  return labels[type] || type
}

function getTypeBadgeClass(type: string) {
  const classes: Record<string, string> = {
    short_term: 'bg-blue-100 text-blue-700',
    medium_term: 'bg-green-100 text-green-700',
    long_term: 'bg-purple-100 text-purple-700',
  }
  return classes[type] || 'bg-gray-100 text-gray-700'
}

function getScopeLabel(scope: string) {
  const labels: Record<string, string> = {
    workspace: t('ai.memory.scopeWorkspace'),
    project: t('ai.memory.scopeProject'),
    issue: t('ai.memory.scopeIssue'),
    agent: t('ai.memory.scopeAgent'),
  }
  return labels[scope] || scope
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleDateString()
}

onMounted(() => {
  loadMemories()
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>