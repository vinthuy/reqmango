<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div><h1 class="text-xl font-semibold text-gray-900">Project Templates</h1><p class="text-sm text-gray-500 mt-1">Create reusable project structures from type templates</p></div>
      <button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ Create Template</button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="tmpl in templates" :key="tmpl.id" class="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition">
        <div class="flex items-center space-x-3 mb-3">
          <div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center text-white text-lg">📋</div>
          <div><h3 class="font-medium text-gray-900">{{ tmpl.name }}</h3><span class="text-xs text-gray-400">{{ tmpl.types?.length || 0 }} types</span></div>
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
        <div class="pt-3 border-t border-gray-100 flex space-x-2">
          <button @click="openAddTypes(tmpl)" class="text-xs text-indigo-600 hover:text-indigo-800">+ Add Types</button>
          <button @click="applyTemplate(tmpl)" class="text-xs text-green-600 hover:text-green-800">Apply</button>
          <button @click="confirmDelete(tmpl)" class="text-xs text-red-500 hover:text-red-700">Delete</button>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showModal=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold mb-4">Create Project Template</h3>
        <div class="space-y-3">
          <div><label class="block text-sm font-medium mb-1">Name *</label><input v-model="form.name" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div><label class="block text-sm font-medium mb-1">Description</label><input v-model="form.description" class="w-full px-3 py-2 border rounded-lg" /></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showModal=false" class="px-4 py-2 border rounded-lg">Cancel</button>
          <button @click="saveTemplate" class="px-4 py-2 bg-blue-600 text-white rounded-lg">Create</button>
        </div>
      </div>
    </div>

    <!-- Add Types Modal -->
    <div v-if="showTypeModal && selected" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showTypeModal=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-lg">
        <h3 class="text-lg font-semibold mb-4">Add Types to {{ selected.name }}</h3>
        <div class="space-y-2 max-h-64 overflow-y-auto">
          <div v-for="tt in availableTypeTemplates" :key="tt.id" class="flex items-center justify-between p-2 border rounded">
            <div class="flex items-center space-x-2">
              <span class="w-2 h-2 rounded-full" :style="{backgroundColor:tt.color}"></span>
              <span class="text-sm font-medium">{{ tt.name }}</span>
              <span class="text-xs text-gray-400">L{{ tt.level }}</span>
            </div>
            <div class="flex items-center space-x-2">
              <label class="text-xs"><input type="checkbox" v-model="typeRequired[tt.id]" /> Req</label>
              <button @click="addType(tt)" class="text-xs text-indigo-600 hover:text-indigo-800">Add</button>
            </div>
          </div>
          <div v-if="!availableTypeTemplates.length" class="text-sm text-gray-400 text-center py-4">All type templates already added</div>
        </div>
        <div class="flex justify-end mt-4"><button @click="showTypeModal=false" class="px-4 py-2 border rounded-lg">Done</button></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import templateApi from '@/api/template'

const props = defineProps<{ workspaceId: number }>()

const templates = ref<any[]>([])
const typeTemplates = ref<any[]>([])
const showModal = ref(false)
const showTypeModal = ref(false)
const selected = ref<any>(null)
const typeRequired = ref<Record<number,boolean>>({})
const form = ref({ name:'', description:'' })

const availableTypeTemplates = computed(() =>
  typeTemplates.value.filter((tt:any) =>
    !(selected.value?.types||[]).some((t:any) => t.type_template_id === tt.id)
  )
)

async function load() {
  try {
    const [tmpl, tt] = await Promise.all([
      templateApi.listTemplates(props.workspaceId),
      templateApi.listTypeTemplates(props.workspaceId),
    ])
    templates.value = tmpl; typeTemplates.value = tt
  } catch(e) { console.error(e) }
}

function openCreate() { form.value = { name:'', description:'' }; showModal.value = true }
async function saveTemplate() {
  await templateApi.createTemplate(props.workspaceId, form.value)
  showModal.value = false; load()
}

function openAddTypes(tmpl: any) { selected.value = tmpl; showTypeModal.value = true }
async function addType(tt: any) {
  if (!selected.value) return
  await templateApi.addTypeToTemplate(selected.value.id, { type_template_id: tt.id, is_required: !!typeRequired.value[tt.id], sequence: 1 })
  typeRequired.value[tt.id] = false; load().then(() => { selected.value = templates.value.find(t=>t.id===selected.value!.id) })
}
async function applyTemplate(tmpl: any) {
  const pid = prompt('Enter project ID to apply this template to:')
  if (!pid) return
  await templateApi.applyTemplate(tmpl.id, parseInt(pid))
  alert('Template applied!')
}
async function confirmDelete(tmpl: any) {
  if (confirm(`Delete "${tmpl.name}"?`)) { await templateApi.deleteTemplate(tmpl.id); load() }
}

onMounted(load)
</script>
