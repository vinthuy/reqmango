<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import ShortcutsPanel from '@/components/ShortcutsPanel.vue'
import AppSidebar from '@/components/AppSidebar.vue'
import { useDarkMode } from '@/composables/useDarkMode'

const route = useRoute()
const authStore = useAuthStore()
const { isDark, toggle: toggleTheme } = useDarkMode()
const showShortcuts = ref(false)

const showSidebar = computed(() => {
  const name = (route.name as string || '').toLowerCase()
  return name && !['login', 'register'].includes(name)
})

function onKeydown(e: KeyboardEvent) {
  if (e.key === '?' && !e.ctrlKey && !e.metaKey && document.activeElement?.tagName !== 'INPUT' && document.activeElement?.tagName !== 'TEXTAREA') {
    e.preventDefault(); showShortcuts.value = !showShortcuts.value
  }
  if (e.key === 'Escape') showShortcuts.value = false
}

onMounted(() => {
  if (authStore.token) authStore.fetchUser()
  document.addEventListener('keydown', onKeydown)
})
onUnmounted(() => document.removeEventListener('keydown', onKeydown))
</script>

<template>
  <div class="flex h-screen bg-gray-50 dark:bg-gray-900">
    <AppSidebar v-if="showSidebar" />
    <main class="flex-1 overflow-auto">
      <router-view />
    </main>
  </div>

  <ConfirmDialog />
  <ShortcutsPanel :visible="showShortcuts" @close="showShortcuts = false" />

  <button
    v-if="showSidebar"
    @click="toggleTheme"
    class="fixed bottom-4 right-4 w-9 h-9 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-full shadow-md hover:shadow-lg transition flex items-center justify-center text-sm z-40"
    :title="isDark ? '切换浅色' : '切换深色'"
  >{{ isDark ? '☀️' : '🌙' }}</button>
</template>
