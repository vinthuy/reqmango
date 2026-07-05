<template>
  <div v-if="visible" class="issue-picker-overlay" @click.self="$emit('close')">
    <div class="issue-picker-dialog bg-white rounded-xl shadow-2xl border border-gray-200 w-[540px] max-h-[560px] flex flex-col">
      <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 shrink-0">
        <h3 class="text-sm font-semibold text-gray-800">{{ title }}</h3>
        <button class="text-gray-400 hover:text-gray-600" @click="$emit('close')">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
        </button>
      </div>

      <div class="px-5 py-3 border-b border-gray-100 shrink-0">
        <input
          ref="searchInput"
          v-model="query"
          type="text"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          :placeholder="t('issue.searchIssuesPlaceholder')"
          @input="onSearch"
        />
      </div>

      <div class="flex-1 overflow-y-auto min-h-0">
        <div v-if="loading" class="flex justify-center py-12 text-sm text-gray-400">
          <svg class="animate-spin w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
          {{ t('common.loading') }}
        </div>

        <div v-else-if="query && results.length === 0" class="py-12 text-center text-sm text-gray-400">
          {{ t('common.noResults') }}
        </div>

        <div v-else-if="!query.trim()" class="py-12 text-center text-sm text-gray-400">
          {{ t('issue.searchIssuesPlaceholder') }}
        </div>

        <div v-else>
          <div
            v-for="item in results"
            :key="item.id"
            class="flex items-center gap-3 px-5 py-3 border-b border-gray-50 last:border-b-0 hover:bg-gray-50 cursor-pointer transition-colors"
            @click="selectIssue(item)"
          >
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <span
                  v-if="item.issue_type"
                  class="px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap shrink-0"
                  :style="{ backgroundColor: item.issue_type.color + '20', color: item.issue_type.color }"
                >{{ item.issue_type.name }}</span>
                <span class="text-[10px] text-gray-400 shrink-0 font-mono">#{{ item.sequence_id }}</span>
                <span class="text-sm font-medium text-gray-800 truncate">{{ item.name }}</span>
              </div>
              <div class="flex items-center gap-2 mt-1 text-[10px] text-gray-500">
                <span class="flex items-center gap-1">
                  <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: stateColor(item.state_group || '') }"></span>
                  {{ item.state_name }}
                </span>
                <span v-if="item.assignees?.[0]">{{ item.assignees[0].display_name || item.assignees[0].username }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { suggestIssues } from '@/api/issue'
import { stateColor } from '@/composables/useRelationHelpers'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  projectId: number
  excludeId?: number
  title: string
}>()

const emit = defineEmits<{
  close: []
  select: [issueId: number, issueName: string]
}>()

const query = ref('')
const results = ref<any[]>([])
const loading = ref(false)
const searchInput = ref<HTMLInputElement | null>(null)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function onSearch() {
  if (debounceTimer) clearTimeout(debounceTimer)
  const q = query.value.trim()
  if (!q) {
    results.value = []
    loading.value = false
    return
  }
  loading.value = true
  debounceTimer = setTimeout(async () => {
    try {
      const data = await suggestIssues(props.projectId, q, 20)
      results.value = (data || []).filter((item: any) => item.id !== (props.excludeId || 0))
    } catch (err) {
      console.error('Failed to search issues:', err)
      results.value = []
    } finally {
      loading.value = false
    }
  }, 300)
}

function selectIssue(item: any) {
  emit('select', item.id, item.name)
}

watch(() => props.visible, async (v) => {
  if (v) {
    await nextTick()
    searchInput.value?.focus()
  } else {
    query.value = ''
    results.value = []
  }
})
</script>

<style scoped>
.issue-picker-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}
.issue-picker-dialog {
  max-height: 80vh;
}
</style>
