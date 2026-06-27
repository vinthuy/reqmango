<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import ShortcutsPanel from '@/components/ShortcutsPanel.vue'

const authStore = useAuthStore()
const dark = ref(false)
const showShortcuts = ref(false)

function onKeydown(e: KeyboardEvent) {
  if (e.key === '?' && !e.ctrlKey && !e.metaKey && document.activeElement?.tagName !== 'INPUT' && document.activeElement?.tagName !== 'TEXTAREA') {
    e.preventDefault(); showShortcuts.value = !showShortcuts.value
  }
  if (e.key === 'Escape') showShortcuts.value = false
}

onMounted(() => {
  if (authStore.token) authStore.fetchUser()
  dark.value = localStorage.getItem('theme') === 'dark'
  if (dark.value) document.documentElement.classList.add('dark')
  document.addEventListener('keydown', onKeydown)
})
onUnmounted(() => document.removeEventListener('keydown', onKeydown))

function toggleTheme() {
  dark.value = !dark.value
  document.documentElement.classList.toggle('dark', dark.value)
  localStorage.setItem('theme', dark.value ? 'dark' : 'light')
}
</script>

<template>
  <router-view />
  <ConfirmDialog />
  <ShortcutsPanel :visible="showShortcuts" @close="showShortcuts = false" />
  <button
    @click="toggleTheme"
    class="fixed bottom-6 left-6 w-10 h-10 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-full shadow-lg hover:shadow-xl transition flex items-center justify-center text-lg z-40"
    :title="dark ? 'Switch to Light' : 'Switch to Dark'"
  >{{ dark ? '☀️' : '🌙' }}</button>
</template>
