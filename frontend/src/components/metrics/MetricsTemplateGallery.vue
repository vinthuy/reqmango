<template>
  <div class="space-y-4">
    <!-- Category Tabs -->
    <div class="flex items-center gap-1 border-b border-gray-200">
      <button
        v-for="cat in categories"
        :key="cat.id"
        @click="activeCategory = cat.id"
        :class="[
          'px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px',
          activeCategory === cat.id
            ? 'border-indigo-600 text-indigo-600'
            : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
        ]"
      >{{ cat.name }}</button>
    </div>

    <!-- Template Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <div
        v-for="tpl in currentTemplates"
        :key="tpl.id"
        @click="emit('use-template', tpl)"
        class="bg-white border border-gray-100 rounded-xl p-5 cursor-pointer hover:shadow-md hover:border-indigo-200 transition-all group"
      >
        <div class="flex items-start gap-3">
          <span class="text-2xl shrink-0 mt-0.5">{{ tpl.icon || '📊' }}</span>
          <div class="flex-1 min-w-0">
            <h4 class="text-sm font-medium text-gray-800 group-hover:text-indigo-600 transition-colors">{{ tpl.name }}</h4>
            <p class="text-xs text-gray-400 mt-1 line-clamp-2">{{ tpl.description }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty -->
    <div v-if="currentTemplates.length === 0" class="flex flex-col items-center justify-center py-12 text-xs text-gray-400">
      该分类下暂无模板
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { TemplateCategory, MetricTemplate } from '@/types/metrics'

const props = defineProps<{
  categories: TemplateCategory[]
}>()

const emit = defineEmits<{
  (e: 'use-template', template: MetricTemplate): void
}>()

const activeCategory = ref(props.categories[0]?.id || '')

const currentTemplates = computed(() => {
  const cat = props.categories.find(c => c.id === activeCategory.value)
  return cat?.templates || []
})
</script>
