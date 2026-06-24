<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6"><div><h1 class="text-xl font-semibold text-gray-900">Automations</h1><p class="text-sm text-gray-500 mt-1">Trigger→Condition→Action rules</p></div><button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ Create Rule</button></div>
    <div class="space-y-3">
      <div v-for="a in automations" :key="a.id" class="bg-white rounded-xl border p-4">
        <div class="flex items-center justify-between">
          <div><h3 class="font-medium">{{ a.name }}</h3><p class="text-sm text-gray-500">{{ a.description }}</p>
            <div class="flex items-center space-x-2 mt-1"><span class="text-xs bg-purple-100 text-purple-700 px-1.5 py-0.5 rounded font-mono">{{ a.trigger_type }}</span><span v-if="!a.is_enabled" class="text-xs bg-gray-100 text-gray-500 px-1.5 py-0.5 rounded">Disabled</span></div>
          </div>
          <div class="flex space-x-2"><button @click="confirmDel(a)" class="text-xs text-red-500">Delete</button></div>
        </div>
      </div>
    </div>

    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showModal=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md"><h3 class="text-lg font-semibold mb-4">Create Automation Rule</h3>
        <div class="space-y-3">
          <div><label class="block text-sm font-medium mb-1">Name</label><input v-model="form.name" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div><label class="block text-sm font-medium mb-1">Description</label><input v-model="form.desc" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div><label class="block text-sm font-medium mb-1">Trigger</label><select v-model="form.trigger" class="w-full px-3 py-2 border rounded-lg"><option value="issue_created">Issue Created</option><option value="issue_updated">Issue Updated</option><option value="state_changed">State Changed</option><option value="assignee_changed">Assignee Changed</option><option value="comment_added">Comment Added</option></select></div>
          <div><label class="block text-sm font-medium mb-1">Conditions (JSON)</label><textarea v-model="form.conditions" rows="2" class="w-full px-3 py-2 border rounded-lg text-xs font-mono" placeholder='[{"field":"priority","operator":"equals","value":"urgent"}]'></textarea></div>
          <div><label class="block text-sm font-medium mb-1">Actions (JSON)</label><textarea v-model="form.actions" rows="2" class="w-full px-3 py-2 border rounded-lg text-xs font-mono" placeholder='[{"type":"assign","field":"assignee","value":"1"}]'></textarea></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6"><button @click="showModal=false" class="px-4 py-2 border rounded-lg">Cancel</button><button @click="save" class="px-4 py-2 bg-blue-600 text-white rounded-lg">Create</button></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import workflowApi from '@/api/workflow'
import { useConfirm } from '@/composables/useConfirm'
const props = defineProps<{ projectId: number }>()
const { confirm } = useConfirm()
const automations = ref<any[]>([])
const showModal = ref(false)
const form = ref({ name:'', desc:'', trigger:'issue_created', conditions:'[]', actions:'[]' })
async function load() { try { automations.value = await workflowApi.listAutomations(props.projectId) } catch(e){ console.error(e) } }
function openCreate() { form.value = { name:'', desc:'', trigger:'issue_created', conditions:'[]', actions:'[]' }; showModal.value = true }
async function save() { await workflowApi.createAutomation(props.projectId, { name:form.value.name, description:form.value.desc, trigger_type:form.value.trigger, conditions:form.value.conditions, actions:form.value.actions }); showModal.value = false; load() }
async function confirmDel(a:any) { if(await confirm('确定要删除此自动化规则吗？')) { await workflowApi.deleteAutomation(props.projectId, a.id); load() } }
onMounted(load)
</script>
