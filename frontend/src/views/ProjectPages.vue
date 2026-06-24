<template>
  <div class="min-h-screen bg-gray-50 flex">
    <!-- Left Sidebar: Page Tree -->
    <aside class="w-64 bg-white border-r border-gray-200 flex flex-col shrink-0">
      <div class="p-4 border-b border-gray-200">
        <div class="flex items-center justify-between">
          <h2 class="font-semibold text-gray-800 text-sm">Pages</h2>
          <button @click="createRootPage" class="text-indigo-600 hover:text-indigo-700 text-sm font-medium">+ New</button>
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
          No pages yet. Click "+ New" to create.
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
              @blur="savePage"
              class="text-2xl font-bold text-gray-900 w-full border-none focus:outline-none focus:ring-0 p-0"
              placeholder="Page title"
            />
          </div>
          <div class="flex items-center gap-2">
            <button @click="archiveCurrent" class="text-sm text-gray-500 hover:text-gray-700">
              {{ selectedPage.archived_at ? 'Restore' : 'Archive' }}
            </button>
            <button @click="goBack" class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50">Back</button>
          </div>
        </div>

        <div class="bg-white rounded-lg border border-gray-200 p-4 min-h-[400px]">
          <textarea
            v-model="editForm.content"
            @blur="savePage"
            class="w-full min-h-[400px] border-none focus:outline-none focus:ring-0 resize-none text-gray-700 leading-relaxed"
            placeholder="Start writing... (Markdown supported)"
          ></textarea>
        </div>

        <div class="mt-4 text-xs text-gray-400">
          Last updated: {{ selectedPage.updated_at }}
          <span v-if="selectedPage.archived_at" class="ml-2 text-amber-500">(Archived)</span>
        </div>
      </div>

      <div v-else class="flex items-center justify-center h-64 text-gray-400">
        Select a page from the sidebar or create a new one.
      </div>
    </main>

    <!-- Create Page Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showCreateModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">Create Page</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Title</label>
            <input v-model="newPageForm.title" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" placeholder="Page title" />
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showCreateModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="doCreate" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">Create</button>
        </div>
      </div>
    </div>

    <!-- Delete Confirm -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showDeleteConfirm = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-sm">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">Delete Page</h3>
        <p class="text-sm text-gray-500 mb-4">Delete "{{ deletingPage?.title }}"? This cannot be undone.</p>
        <div class="flex justify-end space-x-3">
          <button @click="showDeleteConfirm = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
          <button @click="doDelete" class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as pageApi from '@/api/page'
import type { Page } from '@/types/page'
import PageTree from '@/components/PageTree.vue'

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

// Simple computed helper (inline)
function computed<T>(fn: () => T) {
  return ref<T>(fn()) as any
}

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

function savePage() {
  if (!selectedPage.value) return
  if (saveTimeout) clearTimeout(saveTimeout)
  saveTimeout = setTimeout(async () => {
    try {
      await pageApi.updatePage(projectId.value, selectedPage.value!.id, {
        title: editForm.title,
        content: editForm.content,
      })
    } catch (e) { console.error('Failed to save page:', e) }
  }, 500)
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

function goBack() {
  router.push(`/workspace/${slug.value}/project/${projectId.value}`)
}
</script>
