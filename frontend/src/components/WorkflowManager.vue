<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6"><div><h1 class="text-xl font-semibold text-gray-900">{{ t('workflow.title') }}</h1><p class="text-sm text-gray-500 mt-1">{{ t('workflow.desc') }}</p></div><button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ {{ t('workflow.create') }}</button></div>
    <div class="space-y-4">
      <div v-for="w in workflows" :key="w.id" class="bg-white rounded-xl border p-4">
        <div class="flex items-center justify-between mb-3">
          <div>
            <h3 class="font-medium">{{ w.name }}</h3>
            <p class="text-sm text-gray-500">{{ w.description }}</p>
            <div v-if="w.issue_type_ids" class="mt-1 flex flex-wrap gap-1">
              <span v-for="tid in parseIssueTypeIds(w.issue_type_ids)" :key="tid" class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-700">
                {{ getIssueTypeName(tid) }}
              </span>
            </div>
            <div v-else-if="w.issue_type_id" class="mt-1">
              <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-700">
                {{ getIssueTypeName(w.issue_type_id) }}
              </span>
            </div>
          </div>
          <div class="flex space-x-2"><button @click="openAddTrans(w)" class="text-xs text-indigo-600">+ {{ t('workflow.transition') }}</button><button @click="confirmDel(w)" class="text-xs text-red-500">{{ t('common.delete') }}</button></div>
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
      <div class="bg-white rounded-xl p-6 w-full max-w-md"><h3 class="text-lg font-semibold mb-4">{{ t('workflow.createTitle') }}</h3>
        <div class="space-y-3">
          <div><label class="block text-sm font-medium mb-1">{{ t('workflow.name') }}</label><input v-model="form.name" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div><label class="block text-sm font-medium mb-1">{{ t('workflow.description') }}</label><input v-model="form.desc" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div>
            <label class="block text-sm font-medium mb-1">{{ t('workflow.issueType') }}</label>
            <div class="border rounded-lg p-2 max-h-40 overflow-y-auto space-y-1">
              <label v-for="it in issueTypes" :key="it.id" class="flex items-center space-x-2 cursor-pointer hover:bg-gray-50 rounded px-1 py-0.5">
                <input type="checkbox" :value="it.id" v-model="form.selectedTypeIds" class="rounded text-purple-600" />
                <span class="text-sm">{{ it.name }}</span>
              </label>
              <p v-if="!issueTypes.length" class="text-xs text-gray-400">{{ t('workflow.noIssueType') }}</p>
            </div>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6"><button @click="showModal=false" class="px-4 py-2 border rounded-lg">{{ t('common.cancel') }}</button><button @click="save" class="px-4 py-2 bg-blue-600 text-white rounded-lg">{{ t('common.create') }}</button></div>
      </div>
    </div>

    <div v-if="showTrans" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showTrans=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md"><h3 class="text-lg font-semibold mb-4">{{ t('workflow.addTransition') }}</h3>
        <div class="space-y-3">
          <div><label class="block text-sm font-medium mb-1">{{ t('workflow.fromState') }}</label><select v-model="trans.from" class="w-full px-3 py-2 border rounded-lg"><option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option></select></div>
          <div><label class="block text-sm font-medium mb-1">{{ t('workflow.toState') }}</label><select v-model="trans.to" class="w-full px-3 py-2 border rounded-lg"><option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option></select></div>
          <div><label class="block text-sm font-medium mb-1">{{ t('workflow.description') }}</label><input v-model="trans.desc" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div><label class="block text-sm font-medium mb-1">{{ t('workflow.ruleType') }}</label><select v-model="trans.rule_type" class="w-full px-3 py-2 border rounded-lg"><option value="allow">{{ t('workflow.allow') }}</option><option value="approval">{{ t('workflow.approval') }}</option></select></div>
          <div v-if="trans.rule_type==='approval'"><label class="block text-sm font-medium mb-1">{{ t('workflow.approverIds') }}</label><input v-model="trans.approver_ids" class="w-full px-3 py-2 border rounded-lg" :placeholder="t('workflow.approverIdsPlaceholder')" /></div>
          <div v-if="trans.rule_type==='approval'"><label class="block text-sm font-medium mb-1">{{ t('workflow.roleAllowed') }}</label><input v-model="trans.role_allowed" class="w-full px-3 py-2 border rounded-lg" :placeholder="t('workflow.roleAllowedPlaceholder')" /></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6"><button @click="showTrans=false" class="px-4 py-2 border rounded-lg">{{ t('common.cancel') }}</button><button @click="saveTrans" class="px-4 py-2 bg-blue-600 text-white rounded-lg">{{ t('workflow.add') }}</button></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { workflowApi, listWorkspaceWorkflows, createWorkspaceWorkflow, deleteWorkspaceWorkflow, addWorkspaceTransition, deleteWorkspaceTransition } from '@/api/workflow'
import api from '@/api'
import { useConfirm } from '@/composables/useConfirm'

const props = defineProps<{ projectId?: number; workspaceId?: number }>()
const { t } = useI18n()
const { confirm } = useConfirm()
const workflows = ref<any[]>([])
const states = ref<any[]>([])
const issueTypes = ref<any[]>([])
const showModal = ref(false); const showTrans = ref(false); const selWid = ref(0)
const form = ref<{ name: string; desc: string; selectedTypeIds: number[] }>({ name: '', desc: '', selectedTypeIds: [] })
const trans = ref({ from: 0, to: 0, desc: '', rule_type: 'allow', approver_ids: '', role_allowed: '' })

const isWorkspaceMode = computed(() => !!props.workspaceId && !props.projectId)

function getIssueTypeName(id: number) {
  const it = issueTypes.value.find((i: any) => i.id === id)
  return it ? it.name : `#${id}`
}

function parseIssueTypeIds(ids: string | null | undefined): number[] {
  if (!ids) return []
  try { return JSON.parse(ids) } catch { return [] }
}

async function load() { 
  try { 
    const [w, s] = await Promise.all([
      isWorkspaceMode.value 
        ? listWorkspaceWorkflows(props.workspaceId!)
        : workflowApi.list(props.projectId!),
      isWorkspaceMode.value
        ? api.get(`/workspaces/${props.workspaceId}/settings/states`)
        : api.get(`/projects/${props.projectId}/settings/states`)
    ]); 
    workflows.value = Array.isArray(w) ? w : (w?.data ?? []); 
    const statesRaw = s?.data ?? s; 
    states.value = Array.isArray(statesRaw) ? statesRaw : []
    
    // Load issue types
    try {
      const itRes = await api.get('/issue-types', { params: { workspace_id: props.workspaceId || undefined, project_id: props.projectId || undefined } })
      issueTypes.value = itRes?.data?.data || itRes?.data || []
    } catch { issueTypes.value = [] }
  } catch(e){ console.error(e) } 
}

function openCreate() { form.value = { name: '', desc: '', selectedTypeIds: [] }; showModal.value = true }

async function save() { 
  const data: any = { name: form.value.name, description: form.value.desc }
  if (form.value.selectedTypeIds.length > 0) {
    data.issue_type_ids = form.value.selectedTypeIds
  }
  if (isWorkspaceMode.value) {
    await createWorkspaceWorkflow(props.workspaceId!, data)
  } else {
    await workflowApi.create(props.projectId!, data)
  }
  showModal.value = false; load() 
}

async function confirmDel(w:any) { 
  if(await confirm(t('workflow.confirmDelete'))) { 
    if (isWorkspaceMode.value) {
      await deleteWorkspaceWorkflow(props.workspaceId!, w.id)
    } else {
      await workflowApi.delete(props.projectId!, w.id)
    }
    load() 
  } 
}

function openAddTrans(w:any) { selWid.value = w.id; trans.value = { from:0, to:0, desc:'', rule_type:'allow', approver_ids:'', role_allowed:'' }; showTrans.value = true }

async function saveTrans() { 
  const data = { from_state_id:trans.value.from, to_state_id:trans.value.to, description:trans.value.desc, rule_type:trans.value.rule_type, approver_ids:trans.value.approver_ids || undefined, role_allowed:trans.value.role_allowed || undefined }
  if (isWorkspaceMode.value) {
    await addWorkspaceTransition(props.workspaceId!, selWid.value, data)
  } else {
    await workflowApi.addEdge(props.projectId!, selWid.value, data as any)
  }
  showTrans.value = false; load() 
}

async function delTrans(tid:number) { 
  const wid = selWid.value || (workflows.value.find(w=>w.transitions?.some((t:any)=>t.id===tid))?.id||0)
  if (isWorkspaceMode.value) {
    await deleteWorkspaceTransition(props.workspaceId!, wid, tid)
  } else {
    await workflowApi.deleteEdge(props.projectId!, wid, tid)
  }
  load() 
}

onMounted(load)
</script>
