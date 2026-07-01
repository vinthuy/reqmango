<template>
  <div>
    <!-- This node -->
    <div
      class="tree-node flex items-center px-4 py-2 hover:bg-gray-50 cursor-pointer transition-colors border-b border-gray-50 group"
      :class="{
        'bg-amber-50/60': isSearchMatched,
        'bg-blue-50/30': isInSearchPath && !isSearchMatched
      }"
      :style="{ paddingLeft: (16 + depth * 24) + 'px' }"
      @click="$emit('select', node)"
    >
      <!-- Expand/Collapse toggle -->
      <div class="w-8 shrink-0 flex items-center justify-center">
        <input
          type="checkbox"
          :checked="selectedIds.has(node.id)"
          @click.stop
          @change="$emit('toggle-select', node.id)"
          class="rounded border-gray-300 dark:border-gray-500 w-3.5 h-3.5"
        />
        <button
          v-if="node.has_children"
          @click.stop="$emit('toggle', node.id)"
          class="w-5 h-5 flex items-center justify-center rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors ml-1"
        >
          <svg
            v-if="loadingChildren.has(node.id)"
            class="animate-spin w-3.5 h-3.5 text-gray-400"
            xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
          >
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
          </svg>
          <svg
            v-else
            class="w-3.5 h-3.5 text-gray-400 transition-transform"
            :class="{ 'rotate-90': expandedNodes.has(node.id) }"
            fill="none" stroke="currentColor" viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
        <span v-else class="w-5 h-5"></span>
      </div>

      <!-- Node content -->
      <div class="flex-1 min-w-0 flex items-center gap-2">
        <!-- Identifier -->
        <span class="text-xs text-gray-400 font-mono shrink-0">{{ projectIdentifier }}-{{ node.sequence_id }}</span>

        <!-- Type icon -->
        <span v-if="node.issue_type" class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: node.issue_type.color }"></span>

        <!-- Name -->
        <span
          class="text-sm truncate"
          :class="{
            'text-gray-800 font-medium': depth === 0,
            'text-gray-700': depth > 0,
            'text-indigo-700 font-semibold': isSearchMatched
          }"
        >
          <svg v-if="isSearchMatched" class="inline w-3.5 h-3.5 text-amber-500 mr-1 -mt-0.5" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
          </svg>
          {{ node.name }}
        </span>
      </div>

      <!-- Priority -->
      <div class="w-16 text-center shrink-0">
        <span :class="priorityClass(node.priority)" class="text-xs px-1.5 py-0.5 rounded whitespace-nowrap">{{ priorityLabel(node.priority) }}</span>
      </div>

      <!-- State -->
      <div class="w-20 text-center shrink-0">
        <span class="text-xs text-gray-500 dark:text-gray-400">{{ node.state_name || '-' }}</span>
      </div>

      <!-- Sub-issues count -->
      <div class="w-16 text-center shrink-0">
        <span v-if="node.sub_issues_count > 0" class="text-xs text-gray-500">
          <svg class="inline w-3 h-3 mr-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          {{ node.sub_issues_count }}
        </span>
        <span v-else class="text-xs text-gray-300">-</span>
      </div>

      <!-- Add child button (show on hover) -->
      <div class="w-8 shrink-0 flex items-center justify-center">
        <button
          @click.stop="showChildForm = !showChildForm"
          class="w-5 h-5 flex items-center justify-center rounded text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 transition-colors opacity-0 group-hover:opacity-100"
          :title="t('treeNode.createChild')"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Inline child creation form -->
    <div v-if="showChildForm" class="flex items-center gap-2 py-2 px-4 bg-gray-50 border-b border-gray-100" :style="{ paddingLeft: (16 + (depth + 1) * 24) + 'px' }">
      <input
        v-model="childName"
        type="text"
        :placeholder="t('treeNode.childNamePlaceholder')"
        class="flex-1 min-w-0 border border-gray-300 rounded text-sm px-2 py-1 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
        @keydown.enter="submitChild"
        @keydown.escape="cancelChild"
      />
      <select v-model="childPriority" class="border border-gray-300 rounded text-sm px-1.5 py-1 focus:outline-none focus:ring-2 focus:ring-indigo-500 shrink-0">
        <option value="none">{{ t('treeNode.none') }}</option>
        <option value="low">{{ t('treeNode.low') }}</option>
        <option value="medium">{{ t('treeNode.medium') }}</option>
        <option value="high">{{ t('treeNode.high') }}</option>
        <option value="urgent">{{ t('treeNode.urgent') }}</option>
      </select>
      <button @click.stop="submitChild" :disabled="!childName.trim()" class="px-2 py-1 bg-indigo-600 text-white text-xs rounded hover:bg-indigo-700 disabled:opacity-50 shrink-0">{{ t('treeNode.create') }}</button>
      <button @click.stop="cancelChild" class="px-2 py-1 border border-gray-300 text-gray-600 text-xs rounded hover:bg-gray-100 shrink-0">{{ t('treeNode.cancel') }}</button>
    </div>

    <!-- Children (recursive) -->
    <template v-if="expandedNodes.has(node.id) && childrenMap.has(node.id)">
      <TreeNodeItem
        v-for="child in childrenMap.get(node.id)"
        :key="child.id"
        :node="child"
        :depth="depth + 1"
        :expanded-nodes="expandedNodes"
        :children-map="childrenMap"
        :loading-children="loadingChildren"
        :project-identifier="projectIdentifier"
        :selected-ids="selectedIds"
        :search-matched-path="searchMatchedPath"
        :search-matched-id="searchMatchedId"
        @toggle="$emit('toggle', $event)"
        @select="$emit('select', $event)"
        @toggle-select="$emit('toggle-select', $event)"
        @create-child="(payload) => $emit('create-child', payload)"
      />
    </template>

    <!-- Loading placeholder for expanding node -->
    <div v-else-if="expandedNodes.has(node.id) && !childrenMap.has(node.id)" class="flex items-center gap-2 py-2 text-xs text-gray-400" :style="{ paddingLeft: (16 + (depth + 1) * 24) + 'px' }">
      <svg class="animate-spin w-3 h-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
      </svg>
      {{ t('treeNode.loadingChildren') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { TreeIssueResponse } from '@/types/issue'

const { t } = useI18n()

const props = defineProps<{
  node: TreeIssueResponse
  depth: number
  expandedNodes: Set<number>
  childrenMap: Map<number, TreeIssueResponse[]>
  loadingChildren: Set<number>
  projectIdentifier: string
  searchMatchedPath?: number[]
  searchMatchedId?: number
  selectedIds: Set<number>
}>()

const emit = defineEmits<{
  (e: 'toggle', nodeId: number): void
  (e: 'select', issue: any): void
  (e: 'toggle-select', issueId: number): void
  (e: 'create-child', payload: { parentId: number; name: string; priority: string }): void
}>()

// Check if this node is in the search-matched path or is the matched node itself
const isSearchMatched = computed(() => props.searchMatchedId === props.node.id)
const isInSearchPath = computed(() => props.searchMatchedPath?.includes(props.node.id) ?? false)

// Child creation form state
const showChildForm = ref(false)
const childName = ref('')
const childPriority = ref('medium')

function submitChild() {
  const name = childName.value.trim()
  if (!name) return
  emit('create-child', { parentId: props.node.id, name, priority: childPriority.value })
  childName.value = ''
  childPriority.value = 'medium'
  showChildForm.value = false
}

function cancelChild() {
  childName.value = ''
  childPriority.value = 'medium'
  showChildForm.value = false
}

function priorityClass(p: string) {
  const m: Record<string, string> = {
    urgent: 'bg-red-100 text-red-700',
    high: 'bg-orange-100 text-orange-700',
    medium: 'bg-yellow-100 text-yellow-700',
    low: 'bg-green-100 text-green-700',
    none: 'bg-gray-100 text-gray-500'
  }
  return m[p] || m.none
}

function priorityLabel(p: string) {
  const m: Record<string, string> = { urgent: t('treeNode.urgent'), high: t('treeNode.high'), medium: t('treeNode.medium'), low: t('treeNode.low'), none: t('treeNode.none') }
  return m[p] || p
}
</script>
