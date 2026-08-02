<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { workspaceApi } from '@/api/workspace'
import { projectApi } from '@/api/project'
import { useI18n } from '@/composables/useI18n'
import type { Workspace } from '@/types'
import type { ProjectResponse } from '@/types/project'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

const workspaces = ref<Workspace[]>([])
const projects = ref<ProjectResponse[]>([])
const collapsed = ref(false)
const showWorkspaceMenu = ref(false)

const isWorkspaceContext = computed(() => !!route.params.slug)
const workspaceSlug = computed(() => route.params.slug as string || '')
const currentWorkspace = computed(() => workspaces.value.find(w => w.slug === workspaceSlug.value))

const isActive = (path: string) => {
  const full = `/workspace/${workspaceSlug.value}${path}`
  return route.path === full || route.path.startsWith(full + '/') || 
    (path === '' && route.path === `/workspace/${workspaceSlug.value}`)
}

onMounted(async () => {
  try {
    workspaces.value = await workspaceApi.list()
  } catch { /* ignore */ }
})

watch(workspaceSlug, async (slug) => {
  if (!slug || !currentWorkspace.value) return
  try {
    projects.value = await projectApi.listProjects(currentWorkspace.value.id)
  } catch { projects.value = [] }
}, { immediate: true })

function goHome() { router.push('/') }
function goToWorkspace(slug: string) {
  showWorkspaceMenu.value = false
  router.push(`/workspace/${slug}`)
}
function logout() {
  authStore.logout()
  router.push('/login')
}

const navItems = computed(() => {
  if (!isWorkspaceContext.value) return []
  return [
    { label: t('sidebar.projects'), icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', path: '' },
    { label: t('sidebar.initiatives'), icon: 'M13 7h8m0 0v8m0-8l-8 8-4-4-6 6', path: '/initiatives' },
    { label: 'AI Agents', icon: 'M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5', path: '/agents' },
    { label: t('sidebar.analytics'), icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0v-6', path: '/analytics' },
    { label: t('sidebar.settings'), icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z', path: '/settings' },
  ]
})
</script>

<template>
  <aside
    :class="[
      'flex flex-col h-screen bg-gray-50/95 dark:bg-gray-900/95 border-r border-gray-200 dark:border-gray-800 transition-all duration-200',
      collapsed ? 'w-16' : 'w-60'
    ]"
  >
    <!-- Logo -->
    <div class="flex items-center gap-2 px-4 h-14 border-b border-gray-200 dark:border-gray-800">
      <button @click="goHome" class="flex items-center gap-2 min-w-0 flex-1" v-if="!collapsed">
        <div class="w-6 h-6 bg-indigo-600 rounded flex items-center justify-center shrink-0">
          <span class="text-white text-xs font-bold">R</span>
        </div>
        <span class="font-semibold text-sm text-gray-900 dark:text-gray-100 truncate">ReqMango</span>
      </button>
      <button @click="goHome" class="mx-auto" v-else>
        <div class="w-7 h-7 bg-indigo-600 rounded flex items-center justify-center">
          <span class="text-white text-xs font-bold">R</span>
        </div>
      </button>
      <button @click="collapsed = !collapsed" class="shrink-0 ml-auto p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-400">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="collapsed ? 'M13 5l7 7-7 7M5 5l7 7-7 7' : 'M11 19l-7-7 7-7m8 14l-7-7 7-7'" />
        </svg>
      </button>
    </div>

    <!-- Workspace Switcher -->
    <div v-if="isWorkspaceContext" class="px-2 pt-2">
      <button @click="showWorkspaceMenu = !showWorkspaceMenu"
        class="w-full flex items-center gap-2 px-3 py-2 rounded-md hover:bg-gray-200/70 dark:hover:bg-gray-800 text-sm text-gray-700 dark:text-gray-300">
        <div class="w-6 h-6 rounded bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center shrink-0">
          <span class="text-indigo-600 dark:text-indigo-400 text-xs font-bold">{{ currentWorkspace?.name?.charAt(0) || 'W' }}</span>
        </div>
        <span v-if="!collapsed" class="truncate flex-1 text-left font-medium">{{ currentWorkspace?.name || 'Workspace' }}</span>
        <svg v-if="!collapsed" class="w-4 h-4 shrink-0 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      <!-- Workspace dropdown -->
      <div v-if="showWorkspaceMenu && !collapsed" class="mt-1 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-1 z-50">
        <div class="px-3 py-1.5 text-xs text-gray-400 uppercase tracking-wider">{{ t('topbar.workspace') }}</div>
        <button v-for="ws in workspaces" :key="ws.id" @click="goToWorkspace(ws.slug)"
          :class="['w-full text-left px-3 py-2 text-sm hover:bg-gray-50 dark:hover:bg-gray-700/50 flex items-center gap-2',
            ws.slug === workspaceSlug ? 'text-indigo-600 dark:text-indigo-400 font-medium bg-indigo-50 dark:bg-indigo-900/20' : 'text-gray-700 dark:text-gray-300']">
          <div class="w-5 h-5 rounded bg-gray-200 dark:bg-gray-700 flex items-center justify-center shrink-0">
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ ws.name.charAt(0) }}</span>
          </div>
          {{ ws.name }}
        </button>
      </div>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 overflow-y-auto px-2 py-2 space-y-0.5">
      <!-- Workspace nav items -->
      <template v-if="isWorkspaceContext">
        <router-link v-for="item in navItems" :key="item.path"
          :to="`/workspace/${workspaceSlug}${item.path}`"
          :class="[
            'flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors',
            isActive(item.path) ? 'bg-gray-200/80 dark:bg-gray-800 text-gray-900 dark:text-gray-100 font-medium' 
                               : 'text-gray-600 dark:text-gray-400 hover:bg-gray-200/50 dark:hover:bg-gray-800/50'
          ]"
          :title="collapsed ? item.label : ''"
        >
          <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="item.icon" />
          </svg>
          <span v-if="!collapsed" class="truncate">{{ item.label }}</span>
        </router-link>

        <!-- Projects divider -->
        <div v-if="!collapsed" class="px-3 pt-3 pb-1">
          <p class="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider">{{ t('sidebar.projects') }}</p>
        </div>

        <!-- Project list -->
        <router-link v-for="p in projects" :key="p.id"
          :to="`/workspace/${workspaceSlug}/project/${p.id}`"
          :class="[
            'flex items-center gap-3 px-3 py-1.5 rounded-md text-sm transition-colors',
            route.params.id === String(p.id) ? 'bg-gray-200/80 dark:bg-gray-800 text-gray-900 dark:text-gray-100 font-medium'
                                            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-200/50 dark:hover:bg-gray-800/50'
          ]"
          :title="collapsed ? p.name : ''"
        >
          <div class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: p.color || '#6366f1' }"></div>
          <span v-if="!collapsed" class="truncate">{{ p.name }}</span>
        </router-link>
      </template>

      <!-- Global nav (no workspace context) -->
      <template v-else>
        <router-link to="/" class="flex items-center gap-3 px-3 py-2 rounded-md text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-200/50 dark:hover:bg-gray-800/50 transition-colors">
          <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
          </svg>
          <span v-if="!collapsed">{{ t('app.home') }}</span>
        </router-link>
      </template>
    </nav>

    <!-- User -->
    <div class="border-t border-gray-200 dark:border-gray-800 p-2">
      <button @click="logout"
        :class="['w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-200/50 dark:hover:bg-gray-800 transition-colors']"
        :title="collapsed ? t('app.logout') : ''"
      >
        <div class="w-6 h-6 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center shrink-0">
          <span class="text-xs text-gray-600 dark:text-gray-300">{{ authStore.user?.email?.charAt(0)?.toUpperCase() || 'U' }}</span>
        </div>
        <span v-if="!collapsed" class="truncate text-xs">{{ authStore.user?.email }}</span>
      </button>
    </div>
  </aside>
</template>
