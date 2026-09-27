<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">{{ t('intake.title') }}</h2>
        <p class="text-sm text-gray-500 mt-1">{{ t('intake.desc') }}</p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="$emit('showForm')" class="px-3 py-1.5 text-sm rounded-lg border border-gray-300 bg-white text-gray-700 hover:bg-gray-50">
          {{ t('intake.formLink') }}
        </button>
        <router-link :to="hubLink" class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700" data-testid="triage-open-hub">
          {{ t('intakeHub.openHub') }}
        </router-link>
      </div>
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-400">{{ t('intake.loading') }}</div>
    <div v-else class="grid grid-cols-2 md:grid-cols-6 gap-3">
      <div v-for="s in statuses" :key="s" class="rounded-lg border border-gray-200 p-3">
        <div class="text-xs text-gray-500">{{ t('intakeHub.status.' + s) }}</div>
        <div class="text-xl font-semibold text-gray-900 mt-1">{{ counts[s] ?? 0 }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { intakeApi, type IntakeStatus } from '@/api/intake'
import { useToast } from '@/composables/useToast'
import { useI18n } from '@/composables/useI18n'

const props = defineProps<{ projectId: number }>()
defineEmits<{ (e: 'showForm'): void }>()
const toast = useToast()
const { t } = useI18n()
const route = useRoute()

const statuses: IntakeStatus[] = ['pending', 'snoozed', 'spec_review', 'accepted', 'rejected', 'duplicate']
const counts = ref<Record<string, number>>({})
const loading = ref(false)
const hubLink = computed(() => `/workspace/${route.params.slug}/project/${props.projectId}/intake`)

onMounted(async () => {
  loading.value = true
  try {
    counts.value = (await intakeApi.list(props.projectId, { limit: 1 })).counts
  } catch (e: any) {
    toast.error(e?.response?.data?.message || e?.message || t('intake.loadFailed'))
  } finally {
    loading.value = false
  }
})
</script>
