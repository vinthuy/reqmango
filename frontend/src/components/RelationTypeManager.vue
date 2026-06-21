<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div><h1 class="text-xl font-semibold text-gray-900">Relation Types</h1><p class="text-sm text-gray-500 mt-1">Define custom relations between work items</p></div>
      <button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ Create Type</button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="rt in types" :key="rt.id" class="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition">
        <div class="flex items-center space-x-3 mb-2">
          <div class="w-10 h-10 bg-gradient-to-br from-green-500 to-teal-600 rounded-xl flex items-center justify-center text-white text-lg">🔗</div>
          <div><h3 class="font-medium text-gray-900">{{ rt.name }}</h3></div>
        </div>
        <div class="space-y-1 text-sm text-gray-500">
          <div class="flex items-center space-x-1"><span class="text-gray-400">inward:</span><span class="font-mono bg-gray-50 px-1.5 py-0.5 rounded">{{ rt.inward_name }}</span></div>
          <div class="flex items-center space-x-1"><span class="text-gray-400">outward:</span><span class="font-mono bg-gray-50 px-1.5 py-0.5 rounded">{{ rt.outward_name }}</span></div>
        </div>
        <div class="mt-3 pt-3 border-t border-gray-100 flex space-x-2">
          <button @click="openEdit(rt)" class="text-xs text-blue-600 hover:text-blue-800">Edit</button>
          <button @click="confirmDelete(rt)" class="text-xs text-red-500 hover:text-red-700">Delete</button>
        </div>
      </div>
    </div>

    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showModal=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold mb-4">{{ editing ? 'Edit' : 'Create' }} Relation Type</h3>
        <div class="space-y-3">
          <div><label class="block text-sm font-medium mb-1">Name *</label><input v-model="form.name" class="w-full px-3 py-2 border rounded-lg" placeholder="e.g. Blocks, Tests" /></div>
          <div><label class="block text-sm font-medium mb-1">Inward name *</label><input v-model="form.inward_name" class="w-full px-3 py-2 border rounded-lg" placeholder="e.g. blocked by, tested by" /><p class="text-xs text-gray-400 mt-0.5">How it reads from the source item</p></div>
          <div><label class="block text-sm font-medium mb-1">Outward name *</label><input v-model="form.outward_name" class="w-full px-3 py-2 border rounded-lg" placeholder="e.g. blocks, tests" /><p class="text-xs text-gray-400 mt-0.5">How it reads from the target item</p></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showModal=false" class="px-4 py-2 border rounded-lg">Cancel</button>
          <button @click="save" class="px-4 py-2 bg-blue-600 text-white rounded-lg">{{ editing ? 'Update' : 'Create' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import relationApi from '@/api/relation'

const props = defineProps<{ workspaceId: number }>()
const types = ref<any[]>([])
const showModal = ref(false)
const editing = ref(false)
const editId = ref(0)
const form = ref({ name:'', inward_name:'', outward_name:'' })

async function load() {
  try { types.value = await relationApi.listRelationTypes(props.workspaceId) } catch(e) { console.error(e) }
}
function openCreate() { editing.value = false; editId.value = 0; form.value = { name:'', inward_name:'', outward_name:'' }; showModal.value = true }
function openEdit(rt: any) { editing.value = true; editId.value = rt.id; form.value = { name:rt.name, inward_name:rt.inward_name, outward_name:rt.outward_name }; showModal.value = true }
async function save() {
  if (editing.value) { await relationApi.updateRelationType(editId.value, form.value) }
  else { await relationApi.createRelationType(props.workspaceId, form.value) }
  showModal.value = false; load()
}
async function confirmDelete(rt: any) { if (confirm(`Delete "${rt.name}"?`)) { await relationApi.deleteRelationType(rt.id); load() } }
onMounted(load)
</script>
