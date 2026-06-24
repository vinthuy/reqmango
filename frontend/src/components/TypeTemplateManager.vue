<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-semibold text-gray-900">Type Templates</h1>
        <p class="text-sm text-gray-500 mt-1">Define work item types with fields and hierarchy</p>
      </div>
      <button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ Create Type</button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="tt in templates" :key="tt.id" class="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition">
        <div class="flex items-start justify-between">
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 rounded-lg flex items-center justify-center text-white text-lg" :style="{backgroundColor:tt.color}">{{ tt.icon || '📋' }}</div>
            <div>
              <h3 class="font-medium text-gray-900">{{ tt.name }}</h3>
              <span class="text-xs text-gray-500 bg-gray-100 px-1.5 py-0.5 rounded font-mono">L{{ tt.level }}</span>
              <span v-if="tt.parent_type_id" class="text-xs text-yellow-600 ml-1">→ parent constraint</span>
            </div>
          </div>
        </div>
        <p v-if="tt.description" class="mt-2 text-sm text-gray-500">{{ tt.description }}</p>
        <div class="mt-3 space-y-1">
          <div v-for="f in tt.fields" :key="f.field_id" class="text-xs text-gray-600 flex items-center space-x-1">
            <span class="w-1.5 h-1.5 rounded-full" :class="f.is_required ? 'bg-red-400' : 'bg-gray-300'"></span>
            <span>{{ f.field_name }}</span>
            <span class="text-gray-400">({{ f.field_type }})</span>
            <span v-if="f.is_required" class="text-red-400">*</span>
          </div>
          <div v-if="!tt.fields?.length" class="text-xs text-gray-400 italic">No fields bound</div>
        </div>
        <div class="mt-4 pt-3 border-t border-gray-100 flex space-x-2">
          <button @click="openBindFields(tt)" class="text-xs text-indigo-600 hover:text-indigo-800">+ Bind Fields</button>
          <button @click="openEdit(tt)" class="text-xs text-blue-600 hover:text-blue-800">Edit</button>
          <button @click="confirmDelete(tt)" class="text-xs text-red-500 hover:text-red-700">Delete</button>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="closeModal">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold mb-4">{{ editing ? 'Edit' : 'Create' }} Type Template</h3>
        <div class="space-y-3">
          <div><label class="block text-sm font-medium mb-1">Name *</label><input v-model="form.name" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div><label class="block text-sm font-medium mb-1">Description</label><input v-model="form.description" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div class="grid grid-cols-2 gap-3">
            <div><label class="block text-sm font-medium mb-1">Color</label><input v-model="form.color" type="color" class="w-full h-10 border rounded-lg" /></div>
            <div><label class="block text-sm font-medium mb-1">Icon</label><input v-model="form.icon" class="w-full px-3 py-2 border rounded-lg" placeholder="icon name" /></div>
          </div>
          <div><label class="block text-sm font-medium mb-1">Level</label><select v-model="form.level" class="w-full px-3 py-2 border rounded-lg"><option :value="0">L0 - Root</option><option :value="1">L1</option><option :value="2">L2</option><option :value="3">L3</option><option :value="4">L4</option><option :value="5">L5</option></select></div>
          <div v-if="form.level > 0"><label class="block text-sm font-medium mb-1">Parent Type</label><select v-model="form.parent_type_id" class="w-full px-3 py-2 border rounded-lg"><option :value="undefined">None</option><option v-for="t in parentOptions" :key="t.id" :value="t.id">{{ t.name }} (L{{ t.level }})</option></select></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="closeModal" class="px-4 py-2 border rounded-lg">Cancel</button>
          <button @click="save" class="px-4 py-2 bg-blue-600 text-white rounded-lg">{{ editing ? 'Update' : 'Create' }}</button>
        </div>
      </div>
    </div>

    <!-- Bind Fields Modal -->
    <div v-if="showFieldModal && selected" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showFieldModal=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-lg">
        <h3 class="text-lg font-semibold mb-4">Bind Fields to {{ selected.name }}</h3>
        <div class="space-y-2 max-h-64 overflow-y-auto">
          <div v-for="cf in availableFields" :key="cf.id" class="flex items-center justify-between p-2 border rounded">
            <div><span class="font-medium text-sm">{{ cf.name }}</span><span class="text-xs text-gray-400 ml-2">({{ cf.field_type }})</span></div>
            <div class="flex items-center space-x-2">
              <label class="text-xs flex items-center"><input type="checkbox" v-model="bindRequired[cf.id]" /> required</label>
              <button @click="bindField(cf)" class="text-xs text-indigo-600 hover:text-indigo-800">Bind</button>
            </div>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t">
          <h4 class="text-sm font-medium mb-2">Bound Fields</h4>
          <div v-for="f in selected.fields||[]" :key="f.field_id" class="flex items-center justify-between text-sm py-1">
            <span>{{ f.field_name }} <span v-if="f.is_required" class="text-red-400">*</span></span>
            <button @click="unbindField(f.field_id)" class="text-xs text-red-500">Remove</button>
          </div>
        </div>
        <div class="flex justify-end mt-4"><button @click="showFieldModal=false" class="px-4 py-2 border rounded-lg">Done</button></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import templateApi from '@/api/template'
import customFieldApi from '@/api/custom-field'
import { useConfirm } from '@/composables/useConfirm'

const props = defineProps<{ workspaceId: number }>()

const { confirm } = useConfirm()

const templates = ref<any[]>([])
const allFields = ref<any[]>([])
const showModal = ref(false)
const showFieldModal = ref(false)
const editing = ref(false)
const selected = ref<any>(null)
const bindRequired = ref<Record<number,boolean>>({})

const form = ref({ name:'', description:'', color:'#6366F1', icon:'circle', level:0, parent_type_id: undefined as number|undefined })

const parentOptions = computed(() => templates.value.filter(t => (t.level||0) === (form.value.level||0) - 1))
const availableFields = computed(() => allFields.value.filter((f:any) => !(selected.value?.fields||[]).some((bf:any) => bf.field_id === f.id)))

async function load() {
  try {
    const [tt, cf] = await Promise.all([
      templateApi.listTypeTemplates?.(props.workspaceId) || (await fetch(`/api/v1/type-templates?workspace_id=${props.workspaceId}`).then(r=>r.json())),
      customFieldApi.listCustomFields(props.workspaceId),
    ])
    templates.value = tt; allFields.value = cf
  } catch(e) { console.error(e) }
}

function openCreate() { editing.value = false; selected.value = null; form.value = { name:'', description:'', color:'#6366F1', icon:'circle', level:0, parent_type_id:undefined }; showModal.value = true }
function openEdit(tt: any) { editing.value = true; selected.value = tt; form.value = { name:tt.name, description:tt.description||'', color:tt.color, icon:tt.icon, level:tt.level||0, parent_type_id:tt.parent_type_id }; showModal.value = true }
function closeModal() { showModal.value = false }
async function save() {
  const data: any = { name:form.value.name, description:form.value.description, color:form.value.color, icon:form.value.icon, level:form.value.level, parent_type_id:form.value.parent_type_id || null }
  try {
    if (editing.value && selected.value) {
      await templateApi.updateTypeTemplate(selected.value.id, data)
    } else {
      await templateApi.createTypeTemplate(props.workspaceId, data)
    }
    closeModal(); load()
  } catch (e) {
    console.error('Failed to save type template:', e)
    alert('保存失败：' + (e as any).message)
  }
}

function openBindFields(tt: any) { selected.value = tt; showFieldModal.value = true }
async function bindField(cf: any) {
  if (!selected.value) return
  await fetch(`/api/v1/type-templates/${selected.value.id}/fields`, { method:'POST', headers:{'Content-Type':'application/json',Authorization:'Bearer '+localStorage.getItem('token')}, body:JSON.stringify({field_id:cf.id,is_required:!!bindRequired.value[cf.id],sequence:1}) })
  bindRequired.value[cf.id] = false; load().then(() => { selected.value = templates.value.find(t=>t.id===selected.value!.id) })
}
async function unbindField(fid: number) {
  if (!selected.value) return
  await fetch(`/api/v1/type-templates/${selected.value.id}/fields/${fid}`, { method:'DELETE', headers:{Authorization:'Bearer '+localStorage.getItem('token')} })
  load().then(() => { selected.value = templates.value.find(t=>t.id===selected.value!.id) })
}
async function confirmDelete(tt: any) { 
  if (!(await confirm({ title: '删除类型模板', message: `确定要删除 "${tt.name}" 吗？此操作不可恢复。`, danger: true, confirmText: '删除' }))) return
  try {
    await templateApi.deleteTypeTemplate(tt.id)
    load()
  } catch (e) {
    console.error('Failed to delete type template:', e)
    alert('删除失败：' + (e as any).message)
  }
}

onMounted(load)
watch(() => props.workspaceId, load)
</script>
