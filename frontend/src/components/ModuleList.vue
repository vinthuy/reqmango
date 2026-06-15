<template>
  <div class="module-list">
    <!-- 头部工具栏 -->
    <div class="bg-white border-b border-gray-200 px-4 py-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <!-- 视图切换 -->
          <div class="flex items-center border border-gray-300 rounded-md">
            <button
              @click="viewMode = 'card'"
              :class="viewMode === 'card' ? 'bg-indigo-50 text-indigo-600' : 'text-gray-600 hover:bg-gray-50'"
              class="px-3 py-1.5 text-sm first:rounded-l-md last:rounded-r-md"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
            </button>
            <button
              @click="viewMode = 'tree'"
              :class="viewMode === 'tree' ? 'bg-indigo-50 text-indigo-600' : 'text-gray-600 hover:bg-gray-50'"
              class="px-3 py-1.5 text-sm first:rounded-l-md last:rounded-r-md"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
              </svg>
            </button>
          </div>
        </div>

        <div class="flex items-center space-x-3">
          <!-- 新建按钮 -->
          <button
            @click="$emit('create')"
            class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>新建模块</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 列表内容 -->
    <div class="p-4">
      <!-- 加载状态 -->
      <div v-if="loading" class="text-center py-12">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="mt-2 text-gray-500">加载中...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="modules.length === 0" class="text-center py-12">
        <svg class="h-12 w-12 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <p class="mt-2 text-gray-500">暂无模块</p>
        <button @click="$emit('create')" class="mt-3 text-indigo-600 hover:text-indigo-800 text-sm">
          创建第一个模块
        </button>
      </div>

      <!-- 卡片视图 -->
      <div v-else-if="viewMode === 'card'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <ModuleCard
          v-for="module in modules"
          :key="module.id"
          :module="module"
          @click="$emit('select', module)"
          @delete="$emit('delete', module)"
        />
      </div>

      <!-- 树形视图 -->
      <div v-else class="space-y-1">
        <ModuleTree
          :tree="moduleTree"
          @select="$emit('select', $event)"
          @delete="$emit('delete', $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import ModuleCard from './ModuleCard.vue'
import ModuleTree from './ModuleTree.vue'
import moduleApi from '@/api/module'
import type { ModuleResponse, ModuleTreeNode } from '@/types/module'

// Props
const props = defineProps<{
  projectId: string
  workspaceId: string
}>()

// Emits
defineEmits<{
  (e: 'create'): void
  (e: 'select', module: ModuleResponse): void
  (e: 'delete', module: ModuleResponse): void
}>()

// State
const modules = ref<ModuleResponse[]>([])
const moduleTree = ref<ModuleTreeNode[]>([])
const loading = ref(false)
const viewMode = ref<'card' | 'tree'>('card')

// Load modules
onMounted(() => {
  loadModules()
})

async function loadModules() {
  loading.value = true
  try {
    modules.value = await moduleApi.listModules(props.projectId, props.workspaceId)
    moduleTree.value = await moduleApi.getModuleTree(props.projectId)
  } catch (error) {
    console.error('Failed to load modules:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.module-list {
  @apply bg-white rounded-lg;
}
</style>