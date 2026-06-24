<template>
  <div class="page-tree">
    <div v-for="page in pages" :key="page.id">
      <div
        @click="$emit('select', page)"
        class="flex items-center justify-between px-3 py-1.5 rounded text-sm cursor-pointer group transition-colors"
        :class="selectedId === page.id ? 'bg-indigo-50 text-indigo-700' : 'text-gray-700 hover:bg-gray-50'"
        :style="{ paddingLeft: (page.depth * 16 + 12) + 'px' }"
      >
        <span class="flex items-center gap-2 truncate">
          <span class="text-xs">{{ page.children && page.children.length > 0 ? '📂' : '📄' }}</span>
          <span class="truncate">{{ page.title }}</span>
        </span>
        <span class="hidden group-hover:flex items-center gap-1">
          <button @click.stop="$emit('add-child', page.id)" title="Add child" class="text-gray-400 hover:text-indigo-600 text-xs">+</button>
          <button @click.stop="$emit('delete', page)" title="Delete" class="text-gray-400 hover:text-red-500 text-xs">✕</button>
        </span>
      </div>
      <!-- Recursive children -->
      <PageTree
        v-if="page.children && page.children.length > 0"
        :pages="page.children"
        :selected-id="selectedId"
        @select="(p: any) => $emit('select', p)"
        @add-child="(id: number) => $emit('add-child', id)"
        @delete="(p: any) => $emit('delete', p)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Page } from '@/types/page'

defineProps<{
  pages: Page[]
  selectedId: number | null
}>()

defineEmits<{
  (e: 'select', page: Page): void
  (e: 'add-child', parentId: number): void
  (e: 'delete', page: Page): void
}>()
</script>
