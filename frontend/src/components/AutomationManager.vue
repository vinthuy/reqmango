<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6"><div><h1 class="text-xl font-semibold text-gray-900">{{ t('automation.title') }}</h1><p class="text-sm text-gray-500 mt-1">{{ t('automation.desc') }}</p></div><button @click="openCreate" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">+ {{ t('automation.create') }}</button></div>
    <div class="space-y-3">
      <div v-for="a in automations" :key="a.id" class="bg-white rounded-xl border p-4">
        <div class="flex items-center justify-between">
          <div><h3 class="font-medium">{{ a.name }}</h3><p class="text-sm text-gray-500">{{ a.description }}</p>
            <div class="flex items-center space-x-2 mt-1"><span class="text-xs bg-purple-100 text-purple-700 px-1.5 py-0.5 rounded font-mono">{{ t(`automation.triggers.${a.trigger_type}`) || a.trigger_type }}</span><span v-if="!a.is_enabled" class="text-xs bg-gray-100 text-gray-500 px-1.5 py-0.5 rounded">{{ t('automation.disabled') }}</span></div>
          </div>
          <div class="flex space-x-2"><button @click="confirmDel(a)" class="text-xs text-red-500">{{ t('common.delete') }}</button></div>
        </div>
      </div>
    </div>

    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showModal=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md"><h3 class="text-lg font-semibold mb-4">{{ t('automation.createTitle') }}</h3>
        <div class="space-y-3">
          <div><label class="block text-sm font-medium mb-1">{{ t('automation.name') }}</label><input v-model="form.name" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div><label class="block text-sm font-medium mb-1">{{ t('automation.description') }}</label><input v-model="form.description" class="w-full px-3 py-2 border rounded-lg" /></div>
          <div><label class="block text-sm font-medium mb-1">{{ t('automation.trigger') }}</label><select v-model="form.trigger" class="w-full px-3 py-2 border rounded-lg"><option value="issue_created">{{ t('automation.triggers.issue_created') }}</option><option value="issue_updated">{{ t('automation.triggers.issue_updated') }}</option><option value="state_changed">{{ t('automation.triggers.state_changed') }}</option><option value="assignee_changed">{{ t('automation.triggers.assignee_changed') }}</option><option value="comment_added">{{ t('automation.triggers.comment_added') }}</option></select></div>
          <div><label class="block text-sm font-medium mb-1">{{ t('automation.conditions') }}</label><textarea v-model="form.conditions" rows="2" class="w-full px-3 py-2 border rounded-lg text-xs font-mono" :placeholder="t('automation.conditionsPlaceholder')"></textarea></div>
          <div><label class="block text-sm font-medium mb-1">{{ t('automation.actions') }}</label><textarea v-model="form.actions" rows="2" class="w-full px-3 py-2 border rounded-lg text-xs font-mono" :placeholder="t('automation.actionsPlaceholder')"></textarea></div>
        </div>
        <div class="flex justify-end space-x-3 mt-6"><button @click="showModal=false" class="px-4 py-2 border rounded-lg">{{ t('common.cancel') }}</button><button @click="save" class="px-4 py-2 bg-blue-600 text-white rounded-lg">{{ t('common.create') }}</button></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { automationApi } from '@/api/automation'
import { useConfirm } from '@/composables/useConfirm'
const props = defineProps<{ projectId: number }>()
const { t } = useI18n()
const { confirm } = useConfirm()
const automations = ref<any[]>([])
const showModal = ref(false)
const form = ref({ name:'', description:'', trigger:'issue_created', conditions:'[]', actions:'[]' })
async function load() { try { automations.value = await automationApi.list(props.projectId) } catch(e){ console.error(e) } }
function openCreate() { form.value = { name:'', description:'', trigger:'issue_created', conditions:'[]', actions:'[]' }; showModal.value = true }
async function save() { await automationApi.create(props.projectId, { name:form.value.name, description:form.value.description, trigger_type:form.value.trigger, conditions:form.value.conditions, actions:form.value.actions }); showModal.value = false; load() }
async function confirmDel(a:any) { if(await confirm(t('automation.deleteConfirm'))) { await automationApi.delete(props.projectId, a.id); load() } }
onMounted(load)
</script>
