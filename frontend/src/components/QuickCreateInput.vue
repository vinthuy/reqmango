<template>
  <div class="quick-create-input flex items-center space-x-2" :class="inline ? '' : 'px-4 py-2 bg-gray-50 border-b border-gray-200'">
    <div class="flex-1 flex items-center space-x-2">
      <select
        v-model="quickCreate.type_id"
        class="px-2 py-1.5 border border-gray-300 rounded-md text-sm bg-white"
      >
        <option v-for="t in issueTypes" :key="t.id" :value="t.id">{{ t.name }}</option>
      </select>
      <input
        ref="titleInput"
        v-model="quickCreate.title"
        type="text"
        :placeholder="placeholder"
        class="flex-1 px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
        @keydown.enter="handleCreate"
        @keydown.esc="handleCancel"
      />
      <select
        v-if="showPriority"
        v-model="quickCreate.priority"
        class="px-2 py-1.5 border border-gray-300 rounded-md text-sm bg-white"
      >
        <option value="none">无优先级</option>
        <option value="urgent">紧急</option>
        <option value="high">高</option>
        <option value="medium">中</option>
        <option value="low">低</option>
      </select>
    </div>
    <div class="flex items-center space-x-1">
      <button
        @click="handleCreate"
        :disabled="!quickCreate.title.trim() || creating"
        class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {{ creating ? '创建中...' : '创建' }}
      </button>
      <button
        v-if="showCancel"
        @click="handleCancel"
        class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
      >
        取消
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import issueApi from '@/api/issue'

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
  placeholder: '输入标题，按 Enter 快速创建...',
  showPriority: true,
  showCancel: false,
  autoFocus: false,
  inline: false
})

const emit = defineEmits<{
  (e: 'created', issue: any): void
  (e: 'cancel'): void
}>()

const titleInput = ref<HTMLInputElement | null>(null)
const creating = ref(false)

const quickCreate = reactive({
  title: '',
  type_id: props.defaultTypeId || (props.issueTypes[0]?.id ?? null),
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
