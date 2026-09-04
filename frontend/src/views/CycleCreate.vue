<template>
  <div class="min-h-screen bg-gray-50">
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center space-x-4">
        <button @click="goBack" class="text-gray-500 hover:text-gray-700">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 class="text-lg font-semibold text-gray-900">{{ t('cycle.create') }}</h1>
      </div>
    </div>

    <div class="max-w-3xl mx-auto px-6 py-6">
      <!-- Step indicator -->
      <div class="flex items-center justify-center mb-8">
        <div v-for="(step, i) in steps" :key="i" class="flex items-center">
          <div class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium" :class="currentStep >= i ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-500'">{{ i + 1 }}</div>
          <span class="ml-2 text-sm" :class="currentStep >= i ? 'text-indigo-600 font-medium' : 'text-gray-400'">{{ step }}</span>
          <div v-if="i < steps.length - 1" class="w-16 h-0.5 mx-3" :class="currentStep > i ? 'bg-indigo-600' : 'bg-gray-200'"></div>
        </div>
      </div>

      <!-- Step 1: Basic info -->
      <div v-if="currentStep === 0" class="bg-white rounded-lg shadow-sm p-6 space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700">{{ t('cycle.name') }} <span class="text-red-500">*</span></label>
          <input v-model="form.name" type="text" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" :placeholder="t('cycle.namePlaceholder')" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">{{ t('cycle.description') }}</label>
          <textarea v-model="form.description" rows="3" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" :placeholder="t('cycle.descriptionPlaceholder')"></textarea>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('cycle.startDate') }} <span class="text-red-500">*</span></label>
            <input v-model="form.start_date" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('cycle.endDate') }}</label>
            <input v-model="form.end_date" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" />
          </div>
        </div>
      </div>

      <!-- Step 2: Select issues -->
      <div v-if="currentStep === 1" class="bg-white rounded-lg shadow-sm p-6">
        <p class="text-sm text-gray-500 mb-4">{{ t('cycle.selectIssuesHint') }}</p>
        <div class="max-h-96 overflow-y-auto space-y-2">
          <label v-for="issue in backlogIssues" :key="issue.id" class="flex items-center p-2 hover:bg-gray-50 rounded cursor-pointer">
            <input type="checkbox" :value="issue.id" v-model="selectedIssueIds" class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" />
            <span class="ml-3 text-sm text-gray-900">{{ issue.name }}</span>
            <span class="ml-auto text-xs text-gray-400">#{{ issue.sequence_id }}</span>
          </label>
          <p v-if="backlogIssues.length === 0" class="text-sm text-gray-400 py-4 text-center">{{ t('cycle.noBacklogIssues') }}</p>
        </div>
      </div>

      <!-- Step 3: Confirm -->
      <div v-if="currentStep === 2" class="bg-white rounded-lg shadow-sm p-6">
        <h3 class="text-lg font-medium text-gray-900 mb-4">{{ t('cycle.confirmCreate') }}</h3>
        <div class="space-y-2 text-sm">
          <div class="flex"><span class="text-gray-500 w-20 inline-block shrink-0">{{ t('cycle.name') }}:</span><span class="text-gray-900">{{ form.name }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20 inline-block shrink-0">{{ t('cycle.description') }}:</span><span class="text-gray-900">{{ form.description || '-' }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20 inline-block shrink-0">{{ t('cycle.startDate') }}:</span><span class="text-gray-900">{{ form.start_date }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20 inline-block shrink-0">{{ t('cycle.endDate') }}:</span><span class="text-gray-900">{{ form.end_date || '-' }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20 inline-block shrink-0">{{ t('issue.issues') }}:</span><span class="text-gray-900">{{ t('cycle.issueCount', { count: selectedIssueIds.length }) }}</span></div>
        </div>
      </div>

      <!-- Navigation -->
      <div class="flex justify-between mt-6">
        <button v-if="currentStep > 0" @click="currentStep--" class="px-4 py-2 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-50">{{ t('cycle.previous') }}</button>
        <div v-else></div>
        <button v-if="currentStep < 2" @click="nextStep" class="px-4 py-2 bg-indigo-600 text-white rounded-md text-sm hover:bg-indigo-700">{{ t('cycle.next') }}</button>
        <button v-if="currentStep === 2" @click="submitCycle" :disabled="submitting" class="px-4 py-2 bg-indigo-600 text-white rounded-md text-sm hover:bg-indigo-700 disabled:opacity-50">
          {{ submitting ? t('cycle.creating') : t('cycle.createCycle') }}
        </button>
      </div>

      <div v-if="cycleStore.error" class="mt-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-600">{{ cycleStore.error }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { useRoute, useRouter } from 'vue-router'
import { useCycleStore } from '@/stores/cycle'
import { issueApi } from '@/api/issue'
import { workspaceApi } from '@/api/workspace'
import type { IssueResponse } from '@/types/issue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const toast = useToast()
const cycleStore = useCycleStore()

// Support both route styles: /workspace/:slug/project/:id and /workspaces/:workspaceId/projects/:projectId
const projectId = Number(route.params.id || route.params.projectId)
const wsSlug = ref(route.params.slug as string || route.query.ws as string || '')
const workspaceId = ref(Number(route.params.workspaceId) || 0)
const steps = [t('cycle.stepBasic'), t('cycle.stepSelectIssues'), t('cycle.stepConfirm')]
const currentStep = ref(0)
const submitting = ref(false)

const form = ref({
  name: '',
  description: '',
  start_date: new Date().toISOString().slice(0, 10),
  end_date: '',
})

const selectedIssueIds = ref<number[]>([])
const backlogIssues = ref<IssueResponse[]>([])

onMounted(async () => {
  // Fetch workspace ID from slug if not already set
  if (workspaceId.value === 0 && wsSlug.value) {
    try {
      const ws = await workspaceApi.getBySlug(wsSlug.value)
      workspaceId.value = ws.id
    } catch { /* ignore */ }
  }
  try {
    const result = await issueApi.listIssues(projectId, workspaceId.value, { limit: 100 })
    backlogIssues.value = (result?.items || []).filter(
      (i: any) => !i.cycle_id && i.state_group !== 'completed' && i.state_group !== 'cancelled'
    )
  } catch {
    // Silently ignore if backlog fails to load
  }
})

function nextStep() {
  if (currentStep.value === 0 && !form.value.name.trim()) {
    toast.warning(t('cycle.nameRequired'))
    return
  }
  currentStep.value++
}

function projectPage() {
  return `/workspace/${wsSlug.value}/project/${projectId}?tab=cycles`
}

function goBack() {
  if (wsSlug) {
    router.push(projectPage())
  } else {
    router.back()
  }
}

async function submitCycle() {
  submitting.value = true
  const created = await cycleStore.createCycleAction(projectId, workspaceId.value, {
    name: form.value.name,
    description: form.value.description || undefined,
    start_date: new Date(form.value.start_date + 'T00:00:00').toISOString(),
    end_date: form.value.end_date ? new Date(form.value.end_date + 'T23:59:59').toISOString() : undefined,
    project_id: projectId,
  })

  if (!created) {
    submitting.value = false
    return
  }

  for (const issueId of selectedIssueIds.value) {
    await cycleStore.addIssueToCycle(created.id, issueId)
  }

  submitting.value = false
  if (wsSlug) {
    router.push(projectPage())
  } else {
    router.back()
  }
}
</script>
