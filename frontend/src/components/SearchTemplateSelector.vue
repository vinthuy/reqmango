<template>
  <div class="search-template-selector flex items-center gap-2">
    <div class="relative" v-click-outside="closeDropdown">
      <button
        @click="open = !open"
        class="flex items-center gap-2 px-3 py-1.5 text-sm border border-gray-300 rounded-lg bg-white hover:bg-gray-50 transition"
      >
        <span>🎯</span>
        <span class="max-w-[120px] truncate">{{ currentTemplateName }}</span>
        <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      <div v-if="open" class="absolute left-0 top-full mt-1 w-72 bg-white rounded-lg shadow-lg border border-gray-200 z-50 py-1">
        <div class="px-3 py-2 text-xs font-semibold text-gray-500 uppercase">{{ t('filter.searchTemplates') }}</div>
        <div v-if="templates.length === 0" class="px-3 py-2 text-sm text-gray-400">{{ t('filter.noSearchTemplates') }}</div>
        
        <div v-for="template in groupedTemplates" :key="template.group" class="border-t border-gray-100 mt-1">
          <div class="px-3 py-1.5 text-xs font-medium text-gray-400">{{ template.group }}</div>
          <button
            v-for="t in template.items"
            :key="t.id"
            @click="applyTemplate(t)"
            class="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-50 transition text-left"
            :class="{ 'bg-indigo-50 text-indigo-700': selectedTemplateId === t.id }"
          >
            <span class="text-lg">{{ t.icon }}</span>
            <div class="flex-1 min-w-0">
              <div class="truncate">{{ t.name }}</div>
              <div class="text-xs text-gray-400 truncate">{{ t.description }}</div>
            </div>
          </button>
        </div>

        <div class="border-t border-gray-100 mt-1 pt-1">
          <button @click="saveAsTemplate" class="w-full flex items-center gap-2 px-3 py-2 text-sm text-indigo-600 hover:bg-indigo-50 transition">
            <span>💾</span> {{ t('filter.saveAsTemplate') }}
          </button>
        </div>
      </div>
    </div>
  </div>

  <div v-if="showSaveModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showSaveModal = false">
    <div class="bg-white rounded-xl p-6 w-full max-w-md">
      <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('filter.saveAsTemplate') }}</h3>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('filter.templateNameLabel') }}</label>
          <input v-model="saveForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" :placeholder="t('filter.templateNamePlaceholder')" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('filter.templateDescriptionLabel') }}</label>
          <textarea v-model="saveForm.description" rows="2" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" :placeholder="t('filter.templateDescriptionPlaceholder')"></textarea>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('filter.templateIconLabel') }}</label>
          <input v-model="saveForm.icon" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" placeholder="🎯" />
        </div>
      </div>
      <div class="flex justify-end space-x-3 mt-6">
        <button @click="showSaveModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('filter.cancel') }}</button>
        <button @click="doSave" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ t('filter.save') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import * as searchTemplateApi from '@/api/search-template'
import type { SearchTemplate, SearchTemplateCreate } from '@/types/search-template'

const { t } = useI18n()

const props = defineProps<{
  projectId: number
  currentRQL?: string
  currentView?: 'list' | 'kanban' | 'tree' | 'gantt' | 'calendar'
}>()

const emit = defineEmits<{
  (e: 'apply', template: SearchTemplate): void
}>()

const templates = ref<SearchTemplate[]>([])
const selectedTemplateId = ref<number | null>(null)
const open = ref(false)
const showSaveModal = ref(false)

const saveForm = ref({
  name: '',
  description: '',
  icon: '🎯',
})

const groupedTemplates = computed(() => {
  const builtIn = templates.value.filter(t => t.is_built_in)
  const custom = templates.value.filter(t => !t.is_built_in)
  
  const groups: { group: string; items: SearchTemplate[] }[] = []
  
  if (builtIn.length > 0) {
    groups.push({ group: t('filter.builtInTemplates'), items: builtIn })
  }
  if (custom.length > 0) {
    groups.push({ group: t('filter.customTemplates'), items: custom })
  }
  
  return groups
})

const currentTemplateName = computed(() => {
  if (selectedTemplateId.value) {
    const t = templates.value.find(x => x.id === selectedTemplateId.value)
    if (t) return t.name
  }
  return t('filter.searchTemplates')
})

onMounted(() => loadTemplates())
watch(() => props.projectId, () => { loadTemplates(); selectedTemplateId.value = null })

async function loadTemplates() {
  try {
    templates.value = await searchTemplateApi.listSearchTemplates(props.projectId)
  } catch (e) { console.error('Failed to load search templates:', e) }
}

function applyTemplate(template: SearchTemplate) {
  selectedTemplateId.value = template.id
  open.value = false
  emit('apply', template)
}

function saveAsTemplate() {
  saveForm.value = {
    name: '',
    description: '',
    icon: '🎯',
  }
  open.value = false
  showSaveModal.value = true
}

async function doSave() {
  if (!saveForm.value.name.trim()) return
  try {
    const data: SearchTemplateCreate = {
      name: saveForm.value.name.trim(),
      description: saveForm.value.description.trim() || undefined,
      icon: saveForm.value.icon,
      rql_template: props.currentRQL || '',
      view_type: props.currentView || 'list',
    }
    const created = await searchTemplateApi.createSearchTemplate(props.projectId, data)
    templates.value.push(created)
    showSaveModal.value = false
  } catch (e) { console.error('Failed to save template:', e) }
}

function closeDropdown() {
  open.value = false
}

const vClickOutside = {
  mounted(el: any, binding: any) {
    el.__clickOutside__ = (e: MouseEvent) => {
      if (!el.contains(e.target)) binding.value()
    }
    document.addEventListener('click', el.__clickOutside__)
  },
  unmounted(el: any) {
    document.removeEventListener('click', el.__clickOutside__)
  },
}
</script>