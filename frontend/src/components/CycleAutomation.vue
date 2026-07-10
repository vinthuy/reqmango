<template>
  <div class="cycle-automation space-y-4">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-sm font-semibold text-gray-800">{{ t('cycle.automation') }}</h3>
    </div>

    <div class="bg-gray-50 rounded-lg p-4 space-y-4">
      <div class="flex items-start justify-between">
        <div>
          <label class="flex items-center gap-2">
            <input
              type="checkbox"
              :checked="autoAddEnabled"
              @change="updateAutoAddEnabled"
              class="w-4 h-4 text-indigo-600 rounded"
            />
            <span class="text-sm font-medium text-gray-700">{{ t('cycle.autoAddTitle') }}</span>
          </label>
          <p class="text-xs text-gray-500 mt-1 ml-6">{{ t('cycle.autoAddDesc') }}</p>
        </div>
      </div>

      <div v-if="autoAddEnabled" class="ml-6 space-y-2">
        <textarea
          v-model="localRQL"
          rows="3"
          class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
          :placeholder="t('cycle.rqlPlaceholder')"
        ></textarea>
        <div class="flex items-center gap-2">
          <button
            @click="applyAutoAddNow"
            :disabled="applying"
            class="px-3 py-1 text-xs text-white bg-indigo-600 rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ applying ? t('cycle.applying') : t('cycle.applyNow') }}
          </button>
          <span v-if="applyResult" :class="['text-xs', applyResult.success ? 'text-green-600' : 'text-red-600']">
            {{ applyResult.message }}
          </span>
        </div>
      </div>

      <div class="border-t border-gray-200 pt-4">
        <div class="flex items-start justify-between">
          <div>
            <label class="flex items-center gap-2">
              <input
                type="checkbox"
                :checked="autoCloseEnabled"
                @change="updateAutoCloseEnabled"
                class="w-4 h-4 text-indigo-600 rounded"
              />
              <span class="text-sm font-medium text-gray-700">{{ t('cycle.autoCloseTitle') }}</span>
            </label>
            <p class="text-xs text-gray-500 mt-1 ml-6">{{ t('cycle.autoCloseDesc') }}</p>
          </div>
        </div>
      </div>

      <div class="border-t border-gray-200 pt-4">
        <div class="flex items-start justify-between">
          <div>
            <label class="flex items-center gap-2">
              <input
                type="checkbox"
                :checked="autoProgressEnabled"
                @change="updateAutoProgressEnabled"
                class="w-4 h-4 text-indigo-600 rounded"
              />
              <span class="text-sm font-medium text-gray-700">{{ t('cycle.autoProgressTitle') }}</span>
            </label>
            <p class="text-xs text-gray-500 mt-1 ml-6">{{ t('cycle.autoProgressDesc') }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import cycleApi from '@/api/cycle'

function debounce<T extends (...args: unknown[]) => void>(fn: T, delay: number): T {
  let timer: ReturnType<typeof setTimeout> | null = null
  return ((...args: unknown[]) => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }) as T
}

const props = defineProps<{
  cycleId: number
  autoAddEnabled: boolean
  autoAddRql: string
  autoCloseEnabled: boolean
  autoProgressEnabled: boolean
}>()

const emit = defineEmits<{
  (e: 'update', data: { autoAddEnabled: boolean; autoAddRQL: string; autoCloseEnabled: boolean; autoProgressEnabled: boolean }): void
}>()

const { t } = useI18n()

const applying = ref(false)
const applyResult = ref<{ success: boolean; message: string } | null>(null)
const localRQL = ref(props.autoAddRql)

watch(() => props.autoAddRql, (newVal) => {
  localRQL.value = newVal
})

const emitRQLUpdate = debounce(() => {
  emit('update', {
    autoAddEnabled: props.autoAddEnabled,
    autoAddRQL: localRQL.value,
    autoCloseEnabled: props.autoCloseEnabled,
    autoProgressEnabled: props.autoProgressEnabled,
  })
}, 500)

watch(localRQL, () => {
  emitRQLUpdate()
})

function updateAutoAddEnabled(e: Event) {
  const value = (e.target as HTMLInputElement).checked
  emit('update', {
    autoAddEnabled: value,
    autoAddRQL: localRQL.value,
    autoCloseEnabled: props.autoCloseEnabled,
    autoProgressEnabled: props.autoProgressEnabled,
  })
}

function updateAutoCloseEnabled(e: Event) {
  const value = (e.target as HTMLInputElement).checked
  emit('update', {
    autoAddEnabled: props.autoAddEnabled,
    autoAddRQL: localRQL.value,
    autoCloseEnabled: value,
    autoProgressEnabled: props.autoProgressEnabled,
  })
}

function updateAutoProgressEnabled(e: Event) {
  const value = (e.target as HTMLInputElement).checked
  emit('update', {
    autoAddEnabled: props.autoAddEnabled,
    autoAddRQL: localRQL.value,
    autoCloseEnabled: props.autoCloseEnabled,
    autoProgressEnabled: value,
  })
}

async function applyAutoAddNow() {
  applying.value = true
  applyResult.value = null

  try {
    await cycleApi.applyAutoAddRules(props.cycleId)
    applyResult.value = { success: true, message: t('cycle.applySuccess') }
  } catch (e: any) {
    applyResult.value = { success: false, message: e.response?.data?.message || t('cycle.applyFailed') }
  } finally {
    applying.value = false
    setTimeout(() => {
      applyResult.value = null
    }, 3000)
  }
}
</script>
