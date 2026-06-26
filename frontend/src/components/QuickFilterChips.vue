<template>
  <div class="flex flex-wrap items-center gap-2">
    <button
      v-for="filter in filters"
      :key="filter.key"
      @click="toggleFilter(filter)"
      :class="[
        'px-3 py-1.5 rounded-full text-xs font-medium transition-colors',
        activeFilters.includes(filter.key)
          ? 'bg-indigo-600 text-white'
          : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
      ]"
    >
      <span class="flex items-center space-x-1">
        <component :is="filter.icon" class="w-3 h-3" />
        <span>{{ filter.label }}</span>
      </span>
    </button>
    <button
      v-if="activeFilters.length > 0"
      @click="clearAll"
      class="px-3 py-1.5 rounded-full text-xs font-medium bg-gray-100 text-gray-500 hover:bg-gray-200"
    >
      清除筛选
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h } from 'vue'

interface QuickFilter {
  key: string
  label: string
  icon: () => any
  params: Record<string, any>
}

const emit = defineEmits<{
  (e: 'filter', filters: Record<string, any>): void
}>()

const activeFilters = ref<string[]>([])

const filters = computed<QuickFilter[]>(() => [
  {
    key: 'mine',
    label: '我的',
    icon: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' })),
    params: { assignee_id: 'me' }
  },
  {
    key: 'unassigned',
    label: '未分配',
    icon: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' })),
    params: { assignee_id: null }
  },
  {
    key: 'high_priority',
    label: '高优先级',
    icon: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z' }), h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M15 12a3 3 0 11-6 0 3 3 0 016 0z' })]),
    params: { priority: 'high' }
  },
  {
    key: 'today',
    label: '今日创建',
    icon: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' })),
    params: { created_date: 'today' }
  },
  {
    key: 'due_soon',
    label: '即将到期',
    icon: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' })),
    params: { target_date: 'week' }
  }
])

function toggleFilter(filter: QuickFilter) {
  const index = activeFilters.value.indexOf(filter.key)
  if (index === -1) {
    activeFilters.value.push(filter.key)
  } else {
    activeFilters.value.splice(index, 1)
  }
  emitFilters()
}

function clearAll() {
  activeFilters.value = []
  emitFilters()
}

function emitFilters() {
  const result: Record<string, any> = {}
  activeFilters.value.forEach(key => {
    const filter = filters.value.find(f => f.key === key)
    if (filter) {
      Object.assign(result, filter.params)
    }
  })
  emit('filter', result)
}
</script>