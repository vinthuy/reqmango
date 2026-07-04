<template>
  <div class="min-h-screen bg-gray-50 flex">
    <!-- Left Sidebar: Page Tree -->
    <aside class="w-64 bg-white border-r border-gray-200 flex flex-col shrink-0">
      <div class="p-4 border-b border-gray-200">
        <div class="flex items-center justify-between">
          <h2 class="font-semibold text-gray-800 text-sm">{{ t('pages.title') }}</h2>
          <button @click="createRootPage" class="text-indigo-600 hover:text-indigo-700 text-sm font-medium">{{ t('pages.new') }}</button>
        </div>
      </div>
      <div class="flex-1 overflow-y-auto p-2">
        <PageTree
          :pages="pageTree"
          :selected-id="selectedPageId"
          @select="selectPage"
          @add-child="addChildPage"
          @delete="promptDelete"
        />
        <div v-if="pageTree.length === 0 && !loading" class="text-center text-gray-400 py-8 text-sm">
          {{ t('pages.empty') }}
        </div>
      </div>
    </aside>

    <!-- Main Content: Editor -->
    <main class="flex-1 overflow-auto">
      <div v-if="loading" class="flex items-center justify-center h-64">
        <div class="animate-spin h-8 w-8 border-4 border-indigo-500 border-t-transparent rounded-full"></div>
      </div>

      <div v-else-if="selectedPage" class="p-6">
        <div class="flex items-center justify-between mb-4">
          <div class="flex-1">
            <input
              v-model="editForm.title"
              @input="debouncedSave"
              class="text-2xl font-bold text-gray-900 w-full border-none focus:outline-none focus:ring-0 p-0"
              placeholder="Page title"
            />
          </div>
          <div class="flex items-center gap-2">
            <span v-if="saving" class="text-xs text-indigo-500 animate-pulse">Saving...</span>
            <span v-if="saved" class="text-xs text-green-600">Saved</span>
            <span v-if="saveError" class="text-xs text-red-500">{{ saveError }}</span>
            <button @click="showVersionPanel = !showVersionPanel" class="text-sm text-gray-500 hover:text-gray-700" :class="{ 'text-indigo-600': showVersionPanel }">
              {{ t('pages.history') }}
            </button>
            <button @click="archiveCurrent" class="text-sm text-gray-500 hover:text-gray-700">
              {{ selectedPage.archived_at ? 'Restore' : 'Archive' }}
            </button>
            <button @click="goBack" class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50">Back</button>
          </div>
        </div>

        <!-- AI Toolbar -->
        <div class="flex items-center gap-2 mb-2">
          <span class="text-xs text-gray-400 mr-1">🤖 {{ t('ai.title') }}:</span>
          <button @click="aiAction('summarize')" :disabled="aiLoading" class="px-2 py-1 text-xs border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50">{{ t('ai.summarize') }}</button>
          <button @click="aiAction('improve')" :disabled="aiLoading" class="px-2 py-1 text-xs border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50">{{ t('ai.improve') }}</button>
          <button @click="aiAction('generate')" :disabled="aiLoading" class="px-2 py-1 text-xs border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50">{{ t('ai.generate') }}</button>
          <span v-if="aiLoading" class="text-xs text-indigo-500 animate-pulse">{{ t('ai.testing') }}</span>
        </div>

        <div class="flex gap-4">
          <div class="flex-1 bg-white rounded-lg border border-gray-200 p-4 min-h-[400px]">
            <TipTapEditor
              v-model="editForm.content"
              @update:modelValue="onContentChange"
              :placeholder="t('pages.placeholder')"
              class="min-h-[400px]"
            />
          </div>
          
          <!-- Version History Panel -->
          <div v-if="showVersionPanel && selectedPage" class="w-72 shrink-0">
            <PageVersionPanel
              :page-id="selectedPage.id"
              :project-id="projectId"
              @restore="onVersionRestore"
            />
          </div>
        </div>

        <div class="mt-4 text-xs text-gray-400">
          {{ t('pages.lastUpdated') }}: {{ selectedPage.updated_at }}
          <span v-if="selectedPage.archived_at" class="ml-2 text-amber-500">({{ t('pages.archived') }})</span>
        </div>
      </div>

      <div v-else class="flex items-center justify-center h-64 text-gray-400">
        {{ t('pages.selectHint') }}
      </div>
    </main>

    <!-- Create Page Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showCreateModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('pages.create') }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('pages.title') }}</label>
            <input v-model="newPageForm.title" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" :placeholder="t('pages.titlePlaceholder')" />
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showCreateModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('common.cancel') }}</button>
          <button @click="doCreate" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ t('common.create') }}</button>
        </div>
      </div>
    </div>

    <!-- Delete Confirm -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showDeleteConfirm = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-sm">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">{{ t('pages.delete') }}</h3>
        <p class="text-sm text-gray-500 mb-4">{{ t('pages.deleteConfirm', { title: deletingPage?.title || '' }) }}</p>
        <div class="flex justify-end space-x-3">
          <button @click="showDeleteConfirm = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('common.cancel') }}</button>
          <button @click="doDelete" class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700">{{ t('common.delete') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as pageApi from '@/api/page'
import type { Page } from '@/types/page'
import PageTree from '@/components/PageTree.vue'
import TipTapEditor from '@/components/TipTapEditor.vue'
import PageVersionPanel from '@/components/PageVersionPanel.vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'

const { t } = useI18n()
const toast = useToast()

const route = useRoute()
const router = useRouter()

const projectId = computed(() => parseInt((route.params as any).id as string, 10))
const slug = computed(() => (route.params as any).slug as string)
const workspaceId = ref(0)

const pageTree = ref<Page[]>([])
const selectedPage = ref<Page | null>(null)
const selectedPageId = ref<number | null>(null)
const loading = ref(false)

const showCreateModal = ref(false)
const showDeleteConfirm = ref(false)
const newPageForm = ref({ title: '' })
const creatingParentId = ref<number | undefined>(undefined)
const deletingPage = ref<Page | null>(null)

const editForm = reactive({ title: '', content: '' })
let saveTimeout: ReturnType<typeof setTimeout> | null = null
const saving = ref(false)
const saved = ref(false)
const saveError = ref('')
const showVersionPanel = ref(false)

onMounted(() => loadPages())

async function loadPages() {
  loading.value = true
  try {
    // Get workspace ID from the project context in route params
    const slug = (route.params as any).slug as string
    const ws = await import('@/api/workspace').then(m => m.workspaceApi.getBySlug(slug))
    workspaceId.value = ws.id
    pageTree.value = await pageApi.getPageTree(projectId.value)
  } catch (e) { console.error('Failed to load pages:', e) }
  finally { loading.value = false }
}

function createRootPage() {
  creatingParentId.value = undefined
  newPageForm.value.title = ''
  showCreateModal.value = true
}

function addChildPage(parentId: number) {
  creatingParentId.value = parentId
  newPageForm.value.title = ''
  showCreateModal.value = true
}

async function doCreate() {
  if (!newPageForm.value.title.trim()) return
  try {
    const created = await pageApi.createPage(projectId.value, workspaceId.value, {
      title: newPageForm.value.title.trim(),
      parent_id: creatingParentId.value,
    })
    showCreateModal.value = false
    await loadPages()
    selectPage(created)
  } catch (e) { console.error('Failed to create page:', e) }
}

async function selectPage(page: Page) {
  try {
    const full = await pageApi.getPage(projectId.value, page.id)
    selectedPage.value = full
    selectedPageId.value = full.id
    editForm.title = full.title
    editForm.content = full.content || ''
  } catch (e) { console.error('Failed to load page:', e) }
}

function onContentChange(content: string) {
  editForm.content = content
  debouncedSave()
}

function onVersionRestore() {
  // Reload the current page after version restore
  if (selectedPage.value) {
    selectPage(selectedPage.value)
  }
}

function debouncedSave() {
  if (!selectedPage.value) return
  if (saveTimeout) clearTimeout(saveTimeout)
  saveError.value = ''
  saveTimeout = setTimeout(() => performSave(), 800)
}

async function performSave() {
  if (!selectedPage.value) return
  saving.value = true
  saved.value = false
  try {
    const updated = await pageApi.updatePage(projectId.value, selectedPage.value!.id, {
      title: editForm.title,
      content: editForm.content,
    })
    // Sync back to selectedPage so UI reflects saved state
    selectedPage.value = { ...selectedPage.value, ...updated }
    saved.value = true
    setTimeout(() => { saved.value = false }, 2000)
  } catch (e: any) {
    saveError.value = e?.message || 'Save failed'
    setTimeout(() => { saveError.value = '' }, 4000)
  } finally {
    saving.value = false
  }
}

// Keep savePage for backward compatibility (textarea @blur)
function savePage() {
  debouncedSave()
}

async function archiveCurrent() {
  if (!selectedPage.value) return
  try {
    if (selectedPage.value.archived_at) {
      await pageApi.restorePage(projectId.value, selectedPage.value.id)
    } else {
      await pageApi.archivePage(projectId.value, selectedPage.value.id)
    }
    await loadPages()
    selectPage(selectedPage.value)
  } catch (e) { console.error('Failed to toggle archive:', e) }
}

function promptDelete(page: Page) {
  deletingPage.value = page
  showDeleteConfirm.value = true
}

async function doDelete() {
  if (!deletingPage.value) return
  try {
    await pageApi.deletePage(projectId.value, deletingPage.value.id)
    if (selectedPageId.value === deletingPage.value.id) {
      selectedPage.value = null
      selectedPageId.value = null
    }
    showDeleteConfirm.value = false
    deletingPage.value = null
    await loadPages()
  } catch (e) { console.error('Failed to delete page:', e) }
}

const aiLoading = ref(false)

async function aiAction(action: string) {
  if (!selectedPage.value) return
  aiLoading.value = true
  try {
    const content = editForm.content || selectedPage.value.content || ''
    const res = await fetch(`/api/v1/pages/${selectedPage.value.id}/ai`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
      },
      body: JSON.stringify({ action, content, context: editForm.title }),
    })
    if (!res.ok) {
      const err = await res.json()
      toast.error(t('ai.failed') + (err.message || t('common.unknown')))
      return
    }
    const data = await res.json()
    if (action === 'summarize' || action === 'improve') {
      editForm.content = data.result
    } else {
      editForm.content = (editForm.content || '') + '\n\n' + data.result
    }
    savePage()
  } catch (e: any) {
    toast.error(t('ai.requestFailed') + e.message)
  } finally { aiLoading.value = false }
}

function goBack() {
  router.push(`/workspace/${slug.value}/project/${projectId.value}`)
}
</script>
