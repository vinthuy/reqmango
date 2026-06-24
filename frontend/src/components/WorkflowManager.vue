<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6"><div><h1 class="text-xl font-semibold text-gray-900">Workflows</h1><p class="text-sm text-gray-500 mt-1">Define state transition rules</p></div><button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ Create Workflow</button></div>
    <div class="space-y-4">
      <div v-for="w in workflows" :key="w.id" class="bg-white rounded-xl border p-4">
        <div class="flex items-center justify-between mb-3">
          <div><h3 class="font-medium">{{ w.name }}</h3><p class="text-sm text-gray-500">{{ w.description }}</p></div>
          <div class="flex space-x-2"><button @click="openAddTrans(w)" class="text-xs text-indigo-600">+ Transition</button><button @click="confirmDel(w)" class="text-xs text-red-500">Delete</button></div>
        </div>
        <div v-if="w.transitions?.length" class="space-y-1">
          <div v-for="t in w.transitions" :key="t.id" class="flex items-center text-sm text-gray-600 bg-gray-50 rounded px-3 py-1.5">
            <span class="font-medium">{{ t.from_name || t.source_name || '#'+t.from_state_id }}</span>
            <span class="mx-2 text-gray-400">→</span>
            <span class="font-medium">{{ t.to_name || t.target_name || '#'+t.to_state_id }}</span>
            <span v-if="t.rule_type" class="ml-2 text-xs bg-blue-100 text-blue-700 px-1 rounded">{{ t.rule_type }}</span>
            <button @click="delTrans(t.id)" class="ml-auto text-xs text-red-400 hover:text-red-600">×</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showModal=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md"><h3 class="text-lg font-semibold mb-4">Create Workflow</h3>
        <div class="space-y-3"><div><label class="block text-sm font-medium mb-1">Name</label><input v-model="form.name" class="w-full px-3 py-2 border rounded-lg" /></div>
        <div><label class="block text-sm font-medium mb-1">Description</label><input v-model="form.desc" class="w-full px-3 py-2 border rounded-lg" /></div></div>
        <div class="flex justify-end space-x-3 mt-6"><button @click="showModal=false" class="px-4 py-2 border rounded-lg">Cancel</button><button @click="save" class="px-4 py-2 bg-blue-600 text-white rounded-lg">Create</button></div>
      </div>
    </div>

    <div v-if="showTrans" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showTrans=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md"><h3 class="text-lg font-semibold mb-4">Add Transition</h3>
        <div class="space-y-3"><div><label class="block text-sm font-medium mb-1">From State</label><select v-model="trans.from" class="w-full px-3 py-2 border rounded-lg"><option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option></select></div>
        <div><label class="block text-sm font-medium mb-1">To State</label><select v-model="trans.to" class="w-full px-3 py-2 border rounded-lg"><option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option></select></div>
        <div><label class="block text-sm font-medium mb-1">Description</label><input v-model="trans.desc" class="w-full px-3 py-2 border rounded-lg" /></div></div>
        <div class="flex justify-end space-x-3 mt-6"><button @click="showTrans=false" class="px-4 py-2 border rounded-lg">Cancel</button><button @click="saveTrans" class="px-4 py-2 bg-blue-600 text-white rounded-lg">Add</button></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import workflowApi from '@/api/workflow'
import api from '@/api'
import { useConfirm } from '@/composables/useConfirm'

const props = defineProps<{ projectId: number }>()
const { confirm } = useConfirm()
const workflows = ref<any[]>([])
const states = ref<any[]>([])
const showModal = ref(false); const showTrans = ref(false); const selWid = ref(0)
const form = ref({ name:'', desc:'' }); const trans = ref({ from:0, to:0, desc:'' })

async function load() { try { const [w,s] = await Promise.all([workflowApi.listWorkflows(props.projectId), api.get(`/projects/${props.projectId}/settings/states`)]); workflows.value = w; states.value = s.data } catch(e){ console.error(e) } }
function openCreate() { form.value = { name:'', desc:'' }; showModal.value = true }
async function save() { await workflowApi.createWorkflow(props.projectId, { name:form.value.name, description:form.value.desc }); showModal.value = false; load() }
async function confirmDel(w:any) { if(await confirm('确定要删除此工作流吗？')) { await workflowApi.deleteWorkflow(props.projectId, w.id); load() } }
function openAddTrans(w:any) { selWid.value = w.id; trans.value = { from:0, to:0, desc:'' }; showTrans.value = true }
async function saveTrans() { await workflowApi.addTransition(props.projectId, selWid.value, { from_state_id:trans.value.from, to_state_id:trans.value.to, description:trans.value.desc }); showTrans.value = false; load() }
async function delTrans(tid:number) { await workflowApi.deleteTransition(props.projectId, selWid.value || (workflows.value.find(w=>w.transitions?.some((t:any)=>t.id===tid))?.id||0), tid); load() }
onMounted(load)
</script>
