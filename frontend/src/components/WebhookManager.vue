<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div><h2 class="text-lg font-semibold text-gray-900">{{ t('webhook.title') }}</h2><p class="text-sm text-gray-500 mt-1">{{ t('webhook.desc') }}</p></div>
      <button @click="showForm=true;editing=null" class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700">{{ t('webhook.add') }}</button>
    </div>

    <div v-if="webhooks.length===0" class="text-center py-12 bg-gray-50 rounded-lg border text-gray-400 text-sm">{{ t('webhook.noWebhooks') }}</div>

    <div v-else class="space-y-3">
      <div v-for="w in webhooks" :key="w.id" class="bg-white rounded-lg border border-gray-200 p-4">
        <div class="flex items-center justify-between">
          <div>
            <div class="flex items-center gap-2"><span class="font-medium text-gray-900 text-sm">{{ w.name }}</span><span :class="w.is_active?'text-green-600 bg-green-50':'text-gray-400 bg-gray-100'" class="px-1.5 py-0.5 text-xs rounded-full">{{ w.is_active?t('webhook.active'):t('webhook.paused') }}</span></div>
            <div class="text-xs text-gray-400 mt-1 truncate max-w-md">{{ w.url }}</div>
            <div class="flex gap-1 mt-1"><span v-for="(ev, i) in w.events.split(',').map((e: string) => e.trim())" :key="i" class="px-1.5 py-0.5 bg-indigo-50 text-indigo-600 text-[10px] rounded">{{ t(`webhook.events.${ev}`) || ev }}</span></div>
          </div>
          <div class="flex gap-2">
            <button @click="editWebhook(w)" class="text-xs text-indigo-600 hover:text-indigo-800">{{ t('common.edit') }}</button>
            <button @click="toggle(w)" class="text-xs text-amber-600 hover:text-amber-800">{{ w.is_active?t('webhook.pause'):t('webhook.resume') }}</button>
            <button @click="del(w.id)" class="text-xs text-red-500 hover:text-red-700">{{ t('common.delete') }}</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showForm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showForm=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold mb-4">{{ editing?t('webhook.edit'):t('webhook.add') }} {{ t('webhook.title') }}</h3>
        <div class="space-y-3">
          <div><label class="block text-xs font-medium mb-1">{{ t('webhook.name') }}</label><input v-model="form.name" class="w-full px-3 py-2 border rounded-lg text-sm" :placeholder="t('webhook.namePlaceholder')" /></div>
          <div><label class="block text-xs font-medium mb-1">{{ t('webhook.url') }}</label><input v-model="form.url" class="w-full px-3 py-2 border rounded-lg text-sm" :placeholder="t('webhook.urlPlaceholder')" /></div>
          <div><label class="block text-xs font-medium mb-1">{{ t('webhook.secret') }}</label><input v-model="form.secret" class="w-full px-3 py-2 border rounded-lg text-sm" :placeholder="t('webhook.secretPlaceholder')" /></div>
          <div><label class="block text-xs font-medium mb-1">{{ t('webhook.events') }}</label>
            <div class="flex flex-wrap gap-2">
              <label v-for="e in ['issue_created','issue_updated','state_changed']" :key="e" class="flex items-center gap-1 text-xs"><input type="checkbox" :value="e" v-model="form.eventsList" class="rounded" />{{ t(`webhook.events.${e}`) || e }}</label>
            </div>
          </div>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button @click="showForm=false" class="px-4 py-2 border rounded-lg text-sm">{{ t('common.cancel') }}</button>
          <button @click="save" :disabled="saving" class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm disabled:opacity-50">{{ saving?t('common.saving'):t('common.save') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import api from '@/api'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'

const props = defineProps<{ projectId: number; workspaceId: number }>()
const { t } = useI18n()
const toast = useToast()
const webhooks = ref<any[]>([])
const showForm = ref(false); const editing = ref<any>(null); const saving = ref(false)
const form = reactive({ name:'', url:'', secret:'', eventsList:['issue_created','issue_updated','state_changed'] })

onMounted(()=>load())
async function load(){ const r=await api.get(`/projects/${props.projectId}/webhooks`); webhooks.value=r.data||[] }

function editWebhook(w:any){ editing.value=w; form.name=w.name; form.url=w.url; form.secret=''; form.eventsList=w.events.split(',').map((e:string)=>e.trim()); showForm.value=true }

async function save(){
  saving.value=true
  const payload={name:form.name,url:form.url,secret:form.secret,events:form.eventsList.join(',')}
  try{
    if(editing.value) await api.put(`/projects/${props.projectId}/webhooks/${editing.value.id}`,payload)
    else await api.post(`/projects/${props.projectId}/webhooks?workspace_id=${props.workspaceId}`,payload)
    showForm.value=false; load()
  }catch(e:any){toast.error(e.response?.data?.message||'Failed')}
  finally{saving.value=false}
}

async function toggle(w:any){ await api.put(`/projects/${props.projectId}/webhooks/${w.id}`,{is_active:!w.is_active}); load() }
async function del(id:number){ if(!confirm(t('common.deleteConfirm')))return; await api.delete(`/projects/${props.projectId}/webhooks/${id}`); load() }
</script>
