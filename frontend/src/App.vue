<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import ShortcutsPanel from '@/components/ShortcutsPanel.vue'
import TopBar from '@/components/TopBar.vue'
import ToastContainer from '@/components/ToastContainer.vue'

const route = useRoute()
const authStore = useAuthStore()
const showShortcuts = ref(false)

const showTopBar = computed(() => {
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
  <div class="flex flex-col h-screen bg-gray-50 dark:bg-gray-900">
    <TopBar v-if="showTopBar" />
    <main class="flex-1 overflow-auto">
      <router-view />
    </main>
  </div>

  <ConfirmDialog />
  <ShortcutsPanel :visible="showShortcuts" @close="showShortcuts = false" />
  <ToastContainer />
</template>
