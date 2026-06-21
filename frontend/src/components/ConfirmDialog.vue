<template>
  <Transition name="fade">
    <div v-if="dialogVisible" class="fixed inset-0 bg-black bg-opacity-40 z-[100] flex items-center justify-center" @click.self="onCancel">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-sm p-6 mx-4">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">{{ dialogOptions.title || '确认操作' }}</h3>
        <p class="text-sm text-gray-600 mb-6">{{ dialogOptions.message }}</p>
        <div class="flex justify-end space-x-3">
          <button
            @click="onCancel"
            class="px-4 py-2 text-sm border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
          >
            {{ dialogOptions.cancelText || '取消' }}
          </button>
          <button
            @click="onConfirm"
            class="px-4 py-2 text-sm text-white rounded-md"
            :class="dialogOptions.danger ? 'bg-red-600 hover:bg-red-700' : 'bg-indigo-600 hover:bg-indigo-700'"
          >
            {{ dialogOptions.confirmText || '确认' }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { useConfirm } from '@/composables/useConfirm'

const { dialogVisible, dialogOptions, onConfirm, onCancel } = useConfirm()
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
