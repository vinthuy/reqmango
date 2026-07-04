<template>
  <div class="fixed top-4 right-4 z-[9999] flex flex-col gap-2 pointer-events-none">
    <TransitionGroup name="toast">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :class="[
          'pointer-events-auto px-4 py-3 rounded-lg shadow-lg text-sm font-medium max-w-sm transition-all',
          toast.type === 'success' && 'bg-green-50 border border-green-200 text-green-700',
          toast.type === 'error' && 'bg-red-50 border border-red-200 text-red-700',
          toast.type === 'warning' && 'bg-yellow-50 border border-yellow-200 text-yellow-700',
          toast.type === 'info' && 'bg-blue-50 border border-blue-200 text-blue-700'
        ]"
      >
        <div class="flex items-center gap-2">
          <span v-if="toast.type === 'success'" class="text-green-500">✓</span>
          <span v-else-if="toast.type === 'error'" class="text-red-500">✕</span>
          <span v-else-if="toast.type === 'warning'" class="text-yellow-500">⚠</span>
          <span v-else class="text-blue-500">ℹ</span>
          <span>{{ toast.message }}</span>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { useToast } from '../composables/useToast'

const { toasts } = useToast()
</script>

<style scoped>
.toast-enter-active {
  animation: toast-in 0.3s ease-out;
}
.toast-leave-active {
  animation: toast-out 0.2s ease-in;
}
.toast-move {
  transition: transform 0.3s ease;
}
@keyframes toast-in {
  from { opacity: 0; transform: translateX(100%); }
  to { opacity: 1; transform: translateX(0); }
}
@keyframes toast-out {
  from { opacity: 1; transform: translateX(0); }
  to { opacity: 0; transform: translateX(100%); }
}
</style>
