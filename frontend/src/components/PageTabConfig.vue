，继续<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { listPageTabs, batchSavePageTabs } from '@/api/project'
import type { ProjectPageTab, BuiltInTabDef } from '@/types/project-page-tab'

const { t } = useI18n()
const props = defineProps<{ projectId: number }>()
const emit = defineEmits<{ close: []; saved: [tabs: ProjectPageTab[]] }>()

// Built-in tabs that can be toggled
const builtInTabs: BuiltInTabDef[] = [
  { tab_type: 'issues', name: '工作项', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2', route_key: 'issues' },
  { tab_type: 'cycles', name: '周期', icon: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15', route_key: 'cycles' },
  { tab_type: 'modules', name: '模块', icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10', route_key: 'modules' },
  { tab_type: 'updates', name: '更新', icon: 'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9', route_key: 'updates' },
  { tab_type: 'pages', name: '文档', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z', route_key: 'pages' },
  { tab_type: 'analytics', name: '分析', icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z', route_key: 'analytics' },
  { tab_type: 'dashboards', name: '仪表盘', icon: 'M4 5a1 1 0 011-1h4a1 1 0 011 1v5a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm10 0a1 1 0 011-1h4a1 1 0 011 1v2a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zm0 6a1 1 0 011-1h4a1 1 0 011 1v5a1 1 0 01-1 1h-4a1 1 0 01-1-1v-5zm-10 0a1 1 0 011-1h4a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1v-2z', route_key: 'dashboards' },
  { tab_type: 'releases', name: '版本', icon: 'M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z', route_key: 'releases' },
  { tab_type: 'roadmap', name: '路线图', icon: 'M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l5.447 2.724A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7', route_key: 'roadmap' },
  { tab_type: 'settings', name: '设置', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z', route_key: 'settings' },
]

function getTabDisplayName(tab: { tab_type: string, name: string }): string {
  if (tab.tab_type !== 'custom') {
    return t(`pageTab.builtInTabs.${tab.tab_type}`)
  }
  return tab.name
}

const configuredTabs = ref<ProjectPageTab[]>([])
const loading = ref(false)
const saving = ref(false)
const showAddCustom = ref(false)
const customName = ref('')
const customTabType = ref<'saved_view' | 'url'>('url')
const customUrl = ref('')
const customViewId = ref<number | null>(null)

// Computed: which built-in tabs are enabled
const enabledBuiltIns = computed(() => {
  const set = new Set(configuredTabs.value.filter(t => t.tab_type !== 'custom').map(t => t.tab_type))
  return set
})

function toggleBuiltIn(tab: BuiltInTabDef) {
  const existing = configuredTabs.value.find(t => t.tab_type === tab.tab_type)
  if (existing) {
    configuredTabs.value = configuredTabs.value.filter(t => t.tab_type !== tab.tab_type)
  } else {
    configuredTabs.value.push({
      id: 0, project_id: props.projectId, owner_id: 0,
      name: tab.name, icon: '', tab_type: tab.tab_type, route_key: tab.route_key,
      target_type: '', target_url: '', visible: true, sort_order: configuredTabs.value.length,
      created_at: '', updated_at: ''
    })
  }
}

function addCustomTab() {
  if (!customName.value) return
  configuredTabs.value.push({
    id: 0, project_id: props.projectId, owner_id: 0,
    name: customName.value, icon: '', tab_type: 'custom', route_key: '',
    target_type: customTabType.value,
    target_url: customTabType.value === 'url' ? customUrl.value : '',
    target_id: customTabType.value === 'saved_view' ? (customViewId.value ?? undefined) : undefined,
    visible: true, sort_order: configuredTabs.value.length,
    created_at: '', updated_at: ''
  })
  customName.value = ''
  customUrl.value = ''
  customViewId.value = null
  showAddCustom.value = false
}

function removeCustom(index: number) {
  configuredTabs.value.splice(index, 1)
}

function moveUp(index: number) {
  if (index === 0) return
  const arr = configuredTabs.value
  ;[arr[index - 1], arr[index]] = [arr[index], arr[index - 1]]
  configuredTabs.value = [...arr]
}

function moveDown(index: number) {
  if (index === configuredTabs.value.length - 1) return
  const arr = configuredTabs.value
  ;[arr[index], arr[index + 1]] = [arr[index + 1], arr[index]]
  configuredTabs.value = [...arr]
}

async function save() {
  saving.value = true
  try {
    const tabsToSave = configuredTabs.value.map((t, i) => ({
      ...t,
      sort_order: i,
      id: t.id || 0
    }))
    await batchSavePageTabs(props.projectId, tabsToSave)
    const result = await listPageTabs(props.projectId)
    emit('saved', result)
    emit('close')
  } catch (e) {
    console.error('Failed to save page tabs:', e)
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  loading.value = true
  try {
    const tabs = await listPageTabs(props.projectId)
    if (tabs.length > 0) {
      configuredTabs.value = tabs
    } else {
      // Default: issues, cycles, modules, updates, pages, settings
      const defaults = ['issues', 'cycles', 'modules', 'updates', 'pages', 'settings']
      configuredTabs.value = defaults.map((type, i) => {
        const def = builtInTabs.find(b => b.tab_type === type)!
        return {
          id: 0, project_id: props.projectId, owner_id: 0,
          name: def.name, icon: '', tab_type: def.tab_type as ProjectPageTab['tab_type'],
          route_key: def.route_key, target_type: '', target_url: '',
          visible: true, sort_order: i, created_at: '', updated_at: ''
        }
      })
    }
  } catch {
    // Use defaults on error
    const defaults = ['issues', 'cycles', 'modules', 'updates', 'pages', 'settings']
    configuredTabs.value = defaults.map((type, i) => {
      const def = builtInTabs.find(b => b.tab_type === type)!
      return {
        id: 0, project_id: props.projectId, owner_id: 0,
        name: def.name, icon: '', tab_type: def.tab_type as ProjectPageTab['tab_type'],
        route_key: def.route_key, target_type: '', target_url: '',
        visible: true, sort_order: i, created_at: '', updated_at: ''
      }
    })
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="emit('close')">
    <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl w-full max-w-lg mx-4 max-h-[80vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between p-5 border-b border-gray-200 dark:border-gray-700 shrink-0">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">{{ t('pageTab.title') }}</h3>
        <button @click="emit('close')" class="p-1 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-400">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-5 space-y-5">
        <div v-if="loading" class="flex items-center justify-center py-8">
          <div class="animate-spin rounded-full h-6 w-6 border-2 border-indigo-600 border-t-transparent"></div>
        </div>

        <template v-else>
          <!-- Built-in tabs -->
          <div>
            <p class="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-2">{{ t('pageTab.builtInPages') }}</p>
            <div class="space-y-1">
              <div v-for="tab in builtInTabs" :key="tab.tab_type"
                class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors">
                <svg class="w-5 h-5 text-gray-400 dark:text-gray-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="tab.icon" />
                </svg>
                <span class="flex-1 text-sm text-gray-700 dark:text-gray-300">{{ t('pageTab.builtInTabs.' + tab.tab_type) }}</span>
                <button @click="toggleBuiltIn(tab)"
                  :class="['relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent transition-colors',
                    enabledBuiltIns.has(tab.tab_type) ? 'bg-indigo-600' : 'bg-gray-300 dark:bg-gray-600']">
                  <span :class="['inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform',
                    enabledBuiltIns.has(tab.tab_type) ? 'translate-x-4' : 'translate-x-0']" />
                </button>
              </div>
            </div>
          </div>

          <!-- Custom tabs -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <p class="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider">{{ t('pageTab.customPages') }}</p>
              <button @click="showAddCustom = !showAddCustom"
                class="text-xs text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 font-medium flex items-center gap-1">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                {{ t('pageTab.add') }}
              </button>
            </div>

            <!-- Add custom form -->
            <div v-if="showAddCustom" class="mb-3 p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg space-y-3">
              <input v-model="customName" type="text" :placeholder="t('pageTab.name')"
                class="w-full px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-indigo-500" />
              <div class="flex gap-2">
                <select v-model="customTabType"
                  class="text-sm border border-gray-200 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-2 py-2 outline-none">
                  <option value="url">{{ t('pageTab.urlLink') }}</option>
                  <option value="saved_view">{{ t('pageTab.savedView') }}</option>
                </select>
                <input v-if="customTabType === 'url'" v-model="customUrl" type="text" placeholder="https://..."
                  class="flex-1 px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <button @click="addCustomTab" :disabled="!customName"
                class="px-3 py-1.5 bg-indigo-600 text-white text-xs rounded-md hover:bg-indigo-700 disabled:opacity-50 transition">
                {{ t('pageTab.confirmAdd') }}
              </button>
            </div>

            <!-- Configured tabs - reorderable list -->
            <div class="space-y-1">
              <div v-for="(tab, index) in configuredTabs" :key="index"
                class="flex items-center gap-2 px-3 py-2 rounded-lg bg-gray-50 dark:bg-gray-700/50 group">
                <!-- Drag handle -->
                <div class="flex flex-col gap-0.5 shrink-0">
                  <button @click="moveUp(index)" :disabled="index === 0"
                    class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-20 transition p-0.5">
                    <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
                    </svg>
                  </button>
                  <button @click="moveDown(index)" :disabled="index === configuredTabs.length - 1"
                    class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-20 transition p-0.5">
                    <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                  </button>
                </div>

                <!-- Icon -->
                <svg class="w-4 h-4 shrink-0"
                  :class="tab.tab_type === 'custom' ? 'text-purple-500' : 'text-indigo-500'"
                  fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path v-if="tab.tab_type === 'custom'" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                  <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                    :d="builtInTabs.find(b => b.tab_type === tab.tab_type)?.icon || ''" />
                </svg>

                <span class="flex-1 text-sm text-gray-700 dark:text-gray-300">{{ getTabDisplayName(tab) }}</span>
                <span class="text-[10px] px-1.5 py-0.5 rounded bg-gray-200 dark:bg-gray-600 text-gray-500 dark:text-gray-400 shrink-0"
                  :class="tab.tab_type === 'custom' ? 'bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400' : ''">
                  {{ tab.tab_type === 'custom' ? t('pageTab.custom') : t('pageTab.builtIn') }}
                </span>

                <!-- Remove custom -->
                <button v-if="tab.tab_type === 'custom'" @click="removeCustom(index)"
                  class="opacity-0 group-hover:opacity-100 p-1 text-red-400 hover:text-red-600 transition shrink-0">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>

            <p v-if="configuredTabs.length === 0" class="text-sm text-gray-400 dark:text-gray-500 text-center py-4">
              {{ t('pageTab.noTabs') }}
            </p>
          </div>
        </template>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between p-5 border-t border-gray-200 dark:border-gray-700 shrink-0">
        <button @click="emit('close')"
          class="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 transition">
          {{ t('common.cancel') }}
        </button>
        <button @click="save" :disabled="saving"
          class="px-5 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition">
          {{ saving ? t('pageTab.saving') : t('pageTab.saveConfig') }}
        </button>
      </div>
    </div>
  </div>
</template>
