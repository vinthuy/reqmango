<template>
  <div class="quick-create-input flex items-center space-x-2" :class="inline ? '' : 'px-4 py-2 bg-gray-50 border-b border-gray-200'">
    <div class="flex-1 flex items-center space-x-2">
      <select
        v-model="quickCreate.type_id"
        class="px-2 py-1.5 border border-gray-300 rounded-md text-sm bg-white"
      >
        <option v-if="issueTypes.length === 0" value="">{{ t('issueList.loading') }}</option>
        <option value="">{{ t('issue.allTypes') }}</option>
        <option v-for="t in issueTypes" :key="t.id" :value="t.id">{{ t.name }}</option>
      </select>
      <input
        ref="titleInput"
        v-model="quickCreate.title"
        type="text"
        :placeholder="effectivePlaceholder"
        class="flex-1 px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
        @keydown.enter="handleCreate"
        @keydown.esc="handleCancel"
      />
      <select
        v-if="showPriority"
        v-model="quickCreate.priority"
        class="px-2 py-1.5 border border-gray-300 rounded-md text-sm bg-white"
      >
        <option value="none">{{ t('issue.priorityNone') }}</option>
        <option value="urgent">{{ t('issue.priorityUrgent') }}</option>
        <option value="high">{{ t('issue.priorityHigh') }}</option>
        <option value="medium">{{ t('issue.priorityMedium') }}</option>
        <option value="low">{{ t('issue.priorityLow') }}</option>
      </select>
    </div>
    <div class="flex items-center space-x-1">
      <button
        @click="handleCreate"
        :disabled="!quickCreate.title.trim() || creating"
        class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {{ creating ? t('issueList.creating') : t('common.create') }}
      </button>
      <button
        v-if="showCancel"
        @click="handleCancel"
        class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
      >
        {{ t('common.cancel') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import issueApi from '@/api/issue'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  projectId: number
  workspaceId: number
  issueTypes: any[]
  defaultTypeId?: number
  defaultStateId?: number
  placeholder?: string
  showPriority?: boolean
  showCancel?: boolean
  autoFocus?: boolean
  inline?: boolean
}>(), {
  placeholder: '',
  showPriority: true,
  showCancel: false,
  autoFocus: false,
  inline: false
})

const effectivePlaceholder = computed(() => props.placeholder || t('quickCreate.placeholder'))

const emit = defineEmits<{
  (e: 'created', issue: any): void
  (e: 'cancel'): void
}>()

const titleInput = ref<HTMLInputElement | null>(null)
const creating = ref(false)

const quickCreate = reactive({
  title: '',
  type_id: '' as string | number | null,
  priority: 'none' as string,
  state_id: props.defaultStateId || null
})

async function handleCreate() {
  if (!quickCreate.title.trim() || creating.value) return

  creating.value = true
  try {
    const issueData: any = {
      name: quickCreate.title.trim(),
      issue_type_id: quickCreate.type_id,
      priority: quickCreate.priority
    }
    if (quickCreate.state_id) {
      issueData.state_id = quickCreate.state_id
    }
    const issue = await issueApi.createIssue(props.projectId, props.workspaceId, issueData)
    emit('created', issue)
    quickCreate.title = ''
    quickCreate.priority = 'none'
    titleInput.value?.focus()
  } catch (e) {
    console.error('Failed to quick create issue:', e)
  } finally {
    creating.value = false
  }
}

function handleCancel() {
  quickCreate.title = ''
  quickCreate.priority = 'none'
  emit('cancel')
}
</script>
