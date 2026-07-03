<template>
  <div class="relation-type-manager">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-semibold text-gray-900">{{ t('relationType.title') }}</h1>
        <p class="text-sm text-gray-500 mt-1">{{ t('relationType.desc') }}</p>
      </div>
      <button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">{{ t('relationType.create') }}</button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <!-- Empty State -->
    <div v-else-if="types.length === 0" class="text-center py-12">
      <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-gray-100 flex items-center justify-center">
        <span class="text-3xl">🔗</span>
      </div>
      <h3 class="text-lg font-medium text-gray-700 mb-2">{{ t('relationType.noTypes') }}</h3>
      <p class="text-gray-500 mb-4">{{ t('relationType.noTypesDesc') }}</p>
      <button @click="openCreate" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
        {{ t('relationType.create') }}
      </button>
    </div>

    <!-- Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="rt in types" :key="rt.id" class="bg-white rounded-xl border border-gray-200 overflow-hidden hover:shadow-lg transition-shadow">
        <!-- Card Header -->
        <div class="p-4">
          <div class="flex items-start justify-between">
            <div class="flex items-center space-x-3">
              <div class="w-12 h-12 bg-gradient-to-br from-emerald-500 to-teal-600 rounded-xl flex items-center justify-center text-white text-xl shadow-md">
                🔗
              </div>
              <div>
                <h3 class="font-semibold text-gray-900">{{ rt.name }}</h3>
                <span class="text-xs text-gray-500 mt-1">{{ t('relationType.typeLabel') }}</span>
              </div>
            </div>
            <div class="flex items-center space-x-1">
              <button @click="openEdit(rt)" class="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors" :title="t('common.edit')">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/></svg>
              </button>
              <button @click="confirmDelete(rt)" class="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors" :title="t('common.delete')">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Relation Details -->
        <div class="px-4 pb-4">
          <div class="bg-gray-50 rounded-lg p-3 space-y-2">
            <div class="flex items-center justify-between text-sm">
              <span class="text-gray-500 flex items-center">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6"/></svg>
                {{ t('relationType.inward') }}
              </span>
              <span class="font-mono bg-white px-2 py-1 rounded text-gray-700 text-xs border border-gray-200">
                {{ rt.inward_name || '—' }}
              </span>
            </div>
            <div class="flex items-center justify-between text-sm">
              <span class="text-gray-500 flex items-center">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7"/></svg>
                {{ t('relationType.outward') }}
              </span>
              <span class="font-mono bg-white px-2 py-1 rounded text-gray-700 text-xs border border-gray-200">
                {{ rt.outward_name || '—' }}
              </span>
            </div>
          </div>
          
          <!-- Preview -->
          <div class="mt-3 pt-3 border-t border-gray-100">
            <div class="flex items-center justify-center space-x-2 text-xs text-gray-500">
              <span class="bg-blue-100 text-blue-700 px-2 py-1 rounded">{{ t('relationType.itemA') }}</span>
              <span class="text-gray-400">→</span>
              <span class="bg-emerald-100 text-emerald-700 px-2 py-1 rounded">{{ rt.inward_name || rt.name }}</span>
              <span class="text-gray-400">→</span>
              <span class="bg-blue-100 text-blue-700 px-2 py-1 rounded">{{ t('relationType.itemB') }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showModal = false">
      <div class="bg-white rounded-xl shadow-xl w-full max-w-md">
        <div class="p-6 border-b border-gray-100">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900">{{ editing ? t('relationType.edit') : t('relationType.create') }}</h3>
            <button @click="showModal = false" class="text-gray-400 hover:text-gray-600 p-1">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
            </button>
          </div>
        </div>
        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('relationType.name') }} <span class="text-red-500">*</span></label>
            <input v-model="form.name" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent" :placeholder="t('relationType.namePlaceholder')" />
          </div>
          
          <!-- Inward -->
          <div class="bg-blue-50 rounded-lg p-4 space-y-2">
            <div class="flex items-center space-x-2">
              <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6"/></svg>
              <span class="text-sm font-medium text-blue-800">{{ t('relationType.inwardRelation') }}</span>
              <span class="text-xs text-blue-600">({{ t('relationType.inwardHint') }})</span>
            </div>
            <div>
              <input v-model="form.inward_name" class="w-full px-4 py-2 border border-blue-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent" :placeholder="t('relationType.inwardPlaceholder')" />
              <p class="text-xs text-blue-600 mt-1">{{ t('relationType.inwardExample') }}</p>
            </div>
          </div>
          
          <!-- Outward -->
          <div class="bg-emerald-50 rounded-lg p-4 space-y-2">
            <div class="flex items-center space-x-2">
              <svg class="w-5 h-5 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7"/></svg>
              <span class="text-sm font-medium text-emerald-800">{{ t('relationType.outwardRelation') }}</span>
              <span class="text-xs text-emerald-600">({{ t('relationType.outwardHint') }})</span>
            </div>
            <div>
              <input v-model="form.outward_name" class="w-full px-4 py-2 border border-emerald-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent" :placeholder="t('relationType.outwardPlaceholder')" />
              <p class="text-xs text-emerald-600 mt-1">{{ t('relationType.outwardExample') }}</p>
            </div>
          </div>
        </div>
        <div class="px-6 py-4 bg-gray-50 border-t border-gray-100 flex justify-end space-x-3 rounded-b-xl">
          <button @click="showModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-100 text-sm font-medium">{{ t('common.cancel') }}</button>
          <button @click="save" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium">{{ editing ? t('common.update') : t('common.create') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import relationApi from '@/api/relation'
import { useConfirm } from '@/composables/useConfirm'
import { useI18n } from '@/composables/useI18n'

const props = defineProps<{ workspaceId: number }>()
const { confirm } = useConfirm()
const { t } = useI18n()

const types = ref<any[]>([])
const loading = ref(false)
const showModal = ref(false)
const editing = ref(false)
const editId = ref(0)
const form = ref({ name: '', inward_name: '', outward_name: '' })

async function load() {
  loading.value = true
  try { 
    types.value = await relationApi.listRelationTypes(props.workspaceId)
  } catch (e) { 
    console.error(e) 
  } finally {
    loading.value = false
  }
}

function openCreate() { 
  editing.value = false; 
  editId.value = 0; 
  form.value = { name: '', inward_name: '', outward_name: '' }; 
  showModal.value = true 
}

function openEdit(rt: any) { 
  editing.value = true; 
  editId.value = rt.id; 
  form.value = { name: rt.name, inward_name: rt.inward_name, outward_name: rt.outward_name }; 
  showModal.value = true 
}

async function save() {
  if (!form.value.name.trim() || !form.value.inward_name.trim() || !form.value.outward_name.trim()) {
    alert(t('relationType.fillRequired'))
    return
  }
  try {
    if (editing.value) { 
      await relationApi.updateRelationType(editId.value, form.value) 
    } else { 
      await relationApi.createRelationType(props.workspaceId, form.value) 
    }
    showModal.value = false
    load()
  } catch (e) {
    console.error('Failed to save:', e)
    alert(t('relationType.saveFailed') + (e as any).message)
  }
}

async function confirmDelete(rt: any) { 
  if (!(await confirm({ 
    title: t('relationType.deleteTitle'), 
    message: t('relationType.deleteConfirm').replace('{name}', rt.name), 
    danger: true, 
    confirmText: t('common.delete') 
  }))) return
  try {
    await relationApi.deleteRelationType(rt.id)
    load()
  } catch (e) {
    console.error('Failed to delete:', e)
    alert(t('relationType.deleteFailed') + (e as any).message)
  }
}

onMounted(load)
watch(() => props.workspaceId, load)
</script>
