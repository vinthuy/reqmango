<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import ConfirmDialog from '@/components/ConfirmDialog.vue'

const authStore = useAuthStore()

onMounted(() => {
  if (authStore.token) {
    authStore.fetchUser()
  }
})
</script>

<template>
  <router-view />
  <ConfirmDialog />
</template>
