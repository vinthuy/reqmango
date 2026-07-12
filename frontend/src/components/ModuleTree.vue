<template>
  <div class="module-tree">
    <div
      v-for="node in tree"
      :key="node.id"
      class="module-tree-node border-b border-gray-100 last:border-b-0"
    >
      <!-- 节点行 -->
      <div
        class="flex items-center justify-between p-2 hover:bg-gray-50 rounded cursor-pointer"
        :style="{ paddingLeft: ((node.level || 0) * 20 + 8) + 'px' }"
        @click="$emit('select', node)"
      >
        <div class="flex items-center space-x-2 flex-1 min-w-0">
          <!-- 展开/收起图标 -->
          <button
            v-if="hasChildren(node)"
            @click.stop="toggleExpand(node.id)"
            class="p-0.5 text-gray-400 hover:text-gray-600"
          >
            <svg
              class="w-4 h-4 transition-transform"
              :class="{ 'rotate-90': expandedIds.has(node.id) }"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
          <span v-else class="w-5"></span>

          <!-- 模块图标 -->
          <svg class="w-4 h-4 text-indigo-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>

          <!-- 模块名称 -->
          <span class="text-sm font-medium text-gray-900 truncate">{{ node.name }}</span>
          <span v-if="node.is_inherited" class="px-2 py-0.5 bg-green-100 text-green-600 rounded text-xs font-medium whitespace-nowrap">⚙️</span>

          <!-- 进度 -->
          <span class="text-xs text-gray-500">
            {{ node.completed_issues }}/{{ node.total_issues }}
          </span>

          <!-- 进度条 -->
          <div class="w-16 h-1.5 bg-gray-200 rounded-full overflow-hidden">
            <div
              class="h-full bg-indigo-500 rounded-full"
              :style="{ width: node.progress + '%' }"
            ></div>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="flex items-center space-x-2 mr-2" @click.stop>
          <button
            v-if="!node.is_inherited"
            @click="$emit('delete', node)"
            class="p-1 text-gray-400 hover:text-red-600 rounded"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>

      <!-- 子节点（递归渲染） -->
      <template v-if="node.children && node.children.length > 0 && expandedIds.has(node.id)">
        <div
          v-for="child in node.children"
          :key="child.id"
          class="border-b border-gray-100 last:border-b-0"
        >
          <div
            class="flex items-center justify-between p-2 hover:bg-gray-50 rounded cursor-pointer"
            :style="{ paddingLeft: (((node.level || 0) + 1) * 20 + 8) + 'px' }"
            @click="$emit('select', child)"
          >
            <div class="flex items-center space-x-2 flex-1 min-w-0">
              <span class="w-5"></span>
              <svg class="w-4 h-4 text-indigo-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
              <span class="text-sm font-medium text-gray-700 truncate">{{ child.name }}</span>
              <span class="text-xs text-gray-500">
                {{ child.completed_issues }}/{{ child.total_issues }}
              </span>
            </div>
            <div class="flex items-center mr-2" @click.stop>
              <button
                v-if="!child.is_inherited"
                @click="$emit('delete', child)"
                class="p-1 text-gray-400 hover:text-red-600 rounded"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ModuleTreeNode } from '@/types/module'

// Props
defineProps<{
  tree: (ModuleTreeNode & { level?: number })[]
}>()

// Emits
defineEmits<{
  (e: 'select', node: ModuleTreeNode): void
  (e: 'delete', node: ModuleTreeNode): void
}>()

// State
const expandedIds = ref<Set<number>>(new Set())

// Check if node has children
function hasChildren(node: ModuleTreeNode): boolean {
  return !!(node.children && node.children.length > 0)
}

// Toggle expand
function toggleExpand(id: number) {
  if (expandedIds.value.has(id)) {
    expandedIds.value.delete(id)
  } else {
    expandedIds.value.add(id)
  }
}
</script>

<style scoped>
</style>