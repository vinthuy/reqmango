<template>
  <div class="quick-create-input" :class="inline ? '' : 'px-4 py-2 bg-gray-50 border-b border-gray-200'">
    <div class="flex items-center space-x-2">
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
          {{ creating ? t('issueList.creating') : (duplicates.length > 0 ? t('issue.createAnyway') : t('common.create')) }}
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
    <div
      v-if="duplicates.length > 0"
      class="mt-1.5 text-xs text-amber-800 bg-amber-50 border border-amber-200 rounded px-2 py-1.5"
    >
      <span class="font-medium">{{ t('issue.duplicateWarning') }}</span>
      <span class="ml-1">
        <template v-for="(dup, i) in duplicates" :key="dup.id">
          <router-link
            :to="`/workspaces/${workspaceId}/projects/${projectId}/issues/${dup.id}`"
            class="underline hover:text-amber-950"
            target="_blank"
          >#{{ dup.sequence_id }} {{ dup.name }}</router-link><span v-if="i < duplicates.length - 1"> · </span>
        </template>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onBeforeUnmount } from 'vue'
import issueApi from '@/api/issue'
import type { DuplicateIssueItem } from '@/api/issue'
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
const duplicates = ref<DuplicateIssueItem[]>([])
let duplicateTimer: ReturnType<typeof setTimeout> | null = null

const quickCreate = reactive({
  title: '',
  type_id: '' as string | number | null,
  priority: 'none' as string,
  state_id: props.defaultStateId || null
})

watch(() => props.defaultTypeId, (newVal) => {
  if (newVal && !quickCreate.type_id) {
    quickCreate.type_id = newVal
  }
}, { immediate: true })

async function runDuplicateCheck() {
  const name = quickCreate.title.trim()
  if (!props.projectId || name.length < 2) {
    duplicates.value = []
    return
  }
  try {
    const result = await issueApi.checkDuplicates(props.projectId, { name })
    duplicates.value = result.duplicates || []
  } catch {
    duplicates.value = []
  }
}

watch(() => quickCreate.title, () => {
  if (duplicateTimer) clearTimeout(duplicateTimer)
  duplicateTimer = setTimeout(runDuplicateCheck, 400)
})

onBeforeUnmount(() => {
  if (duplicateTimer) clearTimeout(duplicateTimer)
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
    duplicates.value = []
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
  duplicates.value = []
  emit('cancel')
}
</script>
