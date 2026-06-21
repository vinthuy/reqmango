<template>
  <div class="min-h-screen bg-gray-50">
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center space-x-4">
        <button @click="goBack" class="text-gray-500 hover:text-gray-700">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 class="text-lg font-semibold text-gray-900">创建新周期</h1>
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
          <label class="block text-sm font-medium text-gray-700">名称 <span class="text-red-500">*</span></label>
          <input v-model="form.name" type="text" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" placeholder="如: Sprint 1" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">描述</label>
          <textarea v-model="form.description" rows="3" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" placeholder="描述此周期的目标..."></textarea>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">开始日期 <span class="text-red-500">*</span></label>
            <input v-model="form.start_date" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">结束日期</label>
            <input v-model="form.end_date" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" />
          </div>
        </div>
      </div>

      <!-- Step 2: Select issues -->
      <div v-if="currentStep === 1" class="bg-white rounded-lg shadow-sm p-6">
        <p class="text-sm text-gray-500 mb-4">从 Backlog 中选择要加入此周期的工作项（可跳过）</p>
        <div class="max-h-96 overflow-y-auto space-y-2">
          <label v-for="issue in backlogIssues" :key="issue.id" class="flex items-center p-2 hover:bg-gray-50 rounded cursor-pointer">
            <input type="checkbox" :value="issue.id" v-model="selectedIssueIds" class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" />
            <span class="ml-3 text-sm text-gray-900">{{ issue.name }}</span>
            <span class="ml-auto text-xs text-gray-400">#{{ issue.sequence_id }}</span>
          </label>
          <p v-if="backlogIssues.length === 0" class="text-sm text-gray-400 py-4 text-center">暂无 Backlog 工作项</p>
        </div>
      </div>

      <!-- Step 3: Confirm -->
      <div v-if="currentStep === 2" class="bg-white rounded-lg shadow-sm p-6">
        <h3 class="text-lg font-medium text-gray-900 mb-4">确认创建</h3>
        <div class="space-y-2 text-sm">
          <div class="flex"><span class="text-gray-500 w-20">名称:</span><span class="text-gray-900">{{ form.name }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20">描述:</span><span class="text-gray-900">{{ form.description || '-' }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20">开始:</span><span class="text-gray-900">{{ form.start_date }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20">结束:</span><span class="text-gray-900">{{ form.end_date || '-' }}</span></div>
          <div class="flex"><span class="text-gray-500 w-20">工作项:</span><span class="text-gray-900">{{ selectedIssueIds.length }} 个</span></div>
        </div>
      </div>

      <!-- Navigation -->
      <div class="flex justify-between mt-6">
        <button v-if="currentStep > 0" @click="currentStep--" class="px-4 py-2 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-50">上一步</button>
        <div v-else></div>
        <button v-if="currentStep < 2" @click="nextStep" class="px-4 py-2 bg-indigo-600 text-white rounded-md text-sm hover:bg-indigo-700">下一步</button>
        <button v-if="currentStep === 2" @click="submitCycle" :disabled="submitting" class="px-4 py-2 bg-indigo-600 text-white rounded-md text-sm hover:bg-indigo-700 disabled:opacity-50">
          {{ submitting ? '创建中...' : '创建周期' }}
        </button>
      </div>

      <div v-if="cycleStore.error" class="mt-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-600">{{ cycleStore.error }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCycleStore } from '@/stores/cycle'
import { issueApi } from '@/api/issue'
import type { IssueResponse } from '@/types/issue'

const route = useRoute()
const router = useRouter()
const cycleStore = useCycleStore()

const workspaceId = Number(route.params.workspaceId)
const projectId = Number(route.params.projectId)
const steps = ['基本信息', '选择工作项', '确认']
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
  try {
    const result = await issueApi.listIssues(projectId, workspaceId, { limit: 100 })
    backlogIssues.value = (result?.items || []).filter(
      (i: any) => !i.cycle_id && i.state_group !== 'completed' && i.state_group !== 'cancelled'
    )
  } catch {
    // Silently ignore if backlog fails to load
  }
})

function nextStep() {
  if (currentStep.value === 0 && !form.value.name.trim()) {
    alert('请输入周期名称')
    return
  }
  currentStep.value++
}

const wsSlug = (route.query.ws as string) || ''

function goBack() {
  if (wsSlug) {
    router.push(`/workspace/${wsSlug}/project/${projectId}`)
  } else {
    router.back()
  }
}

async function submitCycle() {
  submitting.value = true
  const created = await cycleStore.createCycleAction(projectId, workspaceId, {
    name: form.value.name,
    description: form.value.description || undefined,
    start_date: new Date(form.value.start_date).toISOString(),
    end_date: form.value.end_date ? new Date(form.value.end_date).toISOString() : undefined,
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
    router.push(`/workspace/${wsSlug}/project/${projectId}`)
  } else {
    router.back()
  }
}
</script>
