<template>
  <Transition name="fade">
    <div v-if="visible" class="fixed inset-0 bg-black bg-opacity-30 z-50 flex items-center justify-center" @click.self="$emit('close')">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-md p-6">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ isEdit ? t('module.editTitle') : t('module.createTitle') }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('module.name') }} <span class="text-red-500">*</span></label>
            <input v-model="form.name" type="text" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" :placeholder="t('module.namePlaceholder')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('module.description') }}</label>
            <textarea v-model="form.description" rows="3" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" :placeholder="t('module.descriptionPlaceholder')"></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('module.parentModule') }}</label>
            <select v-model="form.parent_id" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500">
              <option :value="undefined">{{ t('module.noParent') }}</option>
              <option v-for="m in moduleStore.modules" :key="m.id" :value="m.id" :disabled="m.id === editModule?.id">
                {{ m.name }}
              </option>
            </select>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="$emit('close')" class="px-4 py-2 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-50">{{ t('common.cancel') }}</button>
          <button @click="handleSubmit" :disabled="submitting" class="px-4 py-2 bg-indigo-600 text-white rounded-md text-sm hover:bg-indigo-700 disabled:opacity-50">
            {{ submitting ? t('module.saving') : (isEdit ? t('common.save') : t('common.create')) }}
          </button>
        </div>
        <div v-if="moduleStore.error" class="mt-3 text-sm text-red-600">{{ moduleStore.error }}</div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useModuleStore } from '@/stores/module'
import type { ModuleResponse } from '@/types/module'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  editModule: ModuleResponse | null
  workspaceId: number
  projectId: number
}>()

const emit = defineEmits<{ close: []; saved: [] }>()

const moduleStore = useModuleStore()
const isEdit = computed(() => !!props.editModule)
const submitting = ref(false)

const form = ref({ name: '', description: '', parent_id: undefined as number | undefined })

watch(() => props.visible, (v) => {
  if (v) {
    form.value.name = props.editModule?.name || ''
    form.value.description = props.editModule?.description || ''
    form.value.parent_id = (props.editModule?.parent_id as number) || undefined
  }
})

async function handleSubmit() {
  if (!form.value.name.trim()) return
  submitting.value = true

  const data: any = {
    name: form.value.name,
    description: form.value.description,
    project_id: props.projectId,
    workspace_id: props.workspaceId,
    parent_id: form.value.parent_id || undefined,
  }

  let result
  if (isEdit.value) {
    result = await moduleStore.updateModuleAction(props.editModule!.id, {
      name: form.value.name,
      description: form.value.description,
      parent_id: form.value.parent_id || undefined,
    } as any)
  } else {
    result = await moduleStore.createModule(props.workspaceId, data)
  }

  submitting.value = false
  if (result) { emit('saved'); emit('close') }
}
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
