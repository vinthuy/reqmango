<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import NotificationCenter from '@/components/NotificationCenter.vue'
import ApprovalBadge from '@/components/ApprovalBadge.vue'
import { workspaceApi } from '@/api/workspace'
import { useI18n } from '@/composables/useI18n'
import { useDarkMode } from '@/composables/useDarkMode'
import { projectApi } from '@/api/project'
import type { Workspace } from '@/types'
import type { ProjectResponse } from '@/types/project'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()
const { isDark, toggle: toggleTheme } = useDarkMode()

const workspaces = ref<Workspace[]>([])
const projects = ref<ProjectResponse[]>([])
const showWorkspaceMenu = ref(false)

const isWorkspaceContext = computed(() => !!route.params.slug)
const workspaceSlug = computed(() => route.params.slug as string || '')
const currentWorkspace = computed(() => workspaces.value.find(w => w.slug === workspaceSlug.value))
const isInProject = computed(() => !!route.params.id && isWorkspaceContext.value)

function isActive(path: string) {
  const full = `/workspace/${workspaceSlug.value}${path}`
  if (path === '' || path === '/') {
    return route.path === full
  }
  return route.path === full || route.path.startsWith(full + '/')
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
    { label: t('sidebar.projects'), path: '', icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' },
    { label: t('sidebar.initiatives'), path: '/initiatives', icon: 'M13 7h8m0 0v8m0-8l-8 8-4-4-6 6' },
    { label: t('sidebar.settings'), path: '/settings', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z' },
  ]
})

const projectNavItems = computed(() => {
  if (!isInProject.value) return []
  return [
    { label: t('project.tab.issues'), path: '', query: { tab: undefined } },
    { label: t('project.tab.cycles'), path: '', query: { tab: 'cycles' } },
    { label: t('project.tab.modules'), path: '', query: { tab: 'modules' } },
    { label: t('project.tab.updates'), path: '', query: { tab: 'updates' } },
    { label: t('project.tab.reports'), path: '', query: { tab: 'reports' } },
    { label: t('project.tab.pages'), path: '/pages', query: { tab: undefined } },
    { label: t('project.tab.dashboards'), path: '/dashboards', query: { tab: undefined } },
    { label: t('project.tab.settings'), path: '/settings', query: { tab: undefined } },
  ] as { label: string; path: string; query: { tab?: string } }[]
})

function projectNavLink(item: { path: string; query: { tab?: string } }) {
  const base = `/workspace/${workspaceSlug.value}/project/${route.params.id}`
  if (item.path) return base + item.path
  if (item.query.tab) {
    return base + '?tab=' + item.query.tab
  }
  return base
}

function isProjectNavActive(item: { path: string; query: { tab?: string } }) {
  if (item.path) return route.path.startsWith(`/workspace/${workspaceSlug.value}/project/${route.params.id}${item.path}`)
  if (item.query.tab) return route.query.tab === item.query.tab
  return route.path === `/workspace/${workspaceSlug.value}/project/${route.params.id}` && !route.query.tab
}

// Close dropdown on outside click
function onDocClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.workspace-switcher')) {
    showWorkspaceMenu.value = false
  }
}
onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>

<template>
  <header class="h-12 border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 flex items-center px-3 gap-2 shrink-0">
    <!-- Logo -->
    <button @click="goHome" class="flex items-center gap-2 shrink-0 mr-2">
      <div class="w-6 h-6 bg-indigo-600 rounded flex items-center justify-center">
        <span class="text-white text-xs font-bold">R</span>
      </div>
      <span class="font-semibold text-sm text-gray-900 dark:text-gray-100 hidden sm:inline">ReqMango</span>
    </button>

    <!-- Workspace nav items -->
    <nav v-if="isWorkspaceContext" class="flex items-center gap-1">
      <router-link v-for="item in navItems" :key="item.path"
        :to="`/workspace/${workspaceSlug}${item.path}`"
        :class="[
          'px-3 py-1.5 rounded-md text-sm font-medium transition-colors',
          isActive(item.path) ? 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100'
                             : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800/50'
        ]">
        {{ item.label }}
      </router-link>
    </nav>

    <!-- Project context indicator -->
    <div v-if="isInProject" class="flex items-center gap-1 ml-4 pl-4 border-l border-gray-200 dark:border-gray-700">
      <span class="text-xs text-gray-400 dark:text-gray-500 mr-1">{{ t('topbar.project') }}</span>
      <router-link v-for="item in projectNavItems" :key="item.label"
        :to="projectNavLink(item)"
        :class="[
          'px-2.5 py-1 text-xs rounded transition-colors',
          isProjectNavActive(item) ? 'bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 font-medium'
                                   : 'text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300'
        ]">
        {{ item.label }}
      </router-link>
    </div>

    <!-- Spacer -->
    <div class="flex-1"></div>

    <!-- Workspace Switcher -->
    <div v-if="isWorkspaceContext" class="workspace-switcher relative">
      <button @click="showWorkspaceMenu = !showWorkspaceMenu"
        class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 text-sm text-gray-700 dark:text-gray-300 transition-colors">
        <div class="w-5 h-5 rounded bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center shrink-0">
          <span class="text-indigo-600 dark:text-indigo-400 text-[10px] font-bold">{{ currentWorkspace?.name?.charAt(0) || 'W' }}</span>
        </div>
        <span class="text-xs font-medium max-w-[100px] truncate hidden sm:inline">{{ currentWorkspace?.name || 'Workspace' }}</span>
        <svg class="w-3.5 h-3.5 shrink-0 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      <!-- Workspace dropdown -->
      <div v-if="showWorkspaceMenu" class="absolute right-0 mt-1 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-1 z-50 min-w-[180px]">
        <div class="px-3 py-1.5 text-[11px] text-gray-400 uppercase tracking-wider">{{ t('topbar.workspace') }}</div>
        <button v-for="ws in workspaces" :key="ws.id" @click="goToWorkspace(ws.slug)"
          :class="['w-full text-left px-3 py-2 text-sm hover:bg-gray-50 dark:hover:bg-gray-700/50 flex items-center gap-2',
            ws.slug === workspaceSlug ? 'text-indigo-600 dark:text-indigo-400 font-medium bg-indigo-50 dark:bg-indigo-900/20' : 'text-gray-700 dark:text-gray-300']">
          <div class="w-5 h-5 rounded bg-gray-200 dark:bg-gray-700 flex items-center justify-center shrink-0">
            <span class="text-[10px] text-gray-500 dark:text-gray-400">{{ ws.name.charAt(0) }}</span>
          </div>
          {{ ws.name }}
        </button>
      </div>
    </div>

    <!-- Language Switcher -->
    <LanguageSwitcher />

    <!-- Dark Mode Toggle -->
    <button @click="toggleTheme"
      class="p-1.5 rounded-md text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
      :title="isDark ? t('app.lightMode') : t('app.darkMode')">
      <span class="text-sm">{{ isDark ? '☀️' : '🌙' }}</span>
    </button>

    <!-- Approval Badge (workspace context only) -->
    <ApprovalBadge v-if="isWorkspaceContext && currentWorkspace"
      :workspace-id="currentWorkspace.id"
      :slug="workspaceSlug" />

    <!-- Notification Center -->
    <NotificationCenter />

    <!-- User / Logout -->
    <button @click="logout"
      class="ml-1 p-1.5 rounded-md text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
      :title="t('topbar.logout')">
      <div class="w-6 h-6 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
        <span class="text-[11px] text-gray-600 dark:text-gray-300">{{ authStore.user?.email?.charAt(0)?.toUpperCase() || 'U' }}</span>
      </div>
    </button>
  </header>
</template>
