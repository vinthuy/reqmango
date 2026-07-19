<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import approvalApi, { type ApprovalResponse } from '@/api/approval'

const props = defineProps<{
  workspaceId: number
  slug: string
}>()

const { t } = useI18n()

const pendingCount = ref(0)
const approvals = ref<ApprovalResponse[]>([])
const isOpen = ref(false)
let pollTimer: ReturnType<typeof setInterval> | undefined

async function loadCount() {
  if (props.workspaceId <= 0) return
  try {
    pendingCount.value = await approvalApi.countPending(props.workspaceId)
  } catch (error) {
    console.error('Failed to load approval count:', error)
  }
}

async function loadApprovals() {
  if (props.workspaceId <= 0) return
  try {
    const list = await approvalApi.listByWorkspace(props.workspaceId, { status: 'pending' })
    approvals.value = (list || []).slice(0, 5)
  } catch (error) {
    console.error('Failed to load approvals:', error)
  }
}

async function refresh() {
  await Promise.all([loadCount(), loadApprovals()])
}

function togglePanel() {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    refresh()
  }
}

function onDocClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.approval-badge')) {
    isOpen.value = false
  }
}

function formatDate(s: string): string {
  const d = new Date(s)
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString().slice(0, 5)
}

onMounted(() => {
  if (props.workspaceId <= 0) return
  refresh()
  document.addEventListener('click', onDocClick)
  pollTimer = setInterval(refresh, 60000)
})

onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<template>
  <div v-if="workspaceId > 0" class="approval-badge relative">
    <button
      @click="togglePanel"
      class="relative p-1.5 rounded-md text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
      :title="t('approvals.title')"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
      </svg>
      <span
        v-if="pendingCount > 0"
        class="absolute -top-1 -right-1 bg-red-500 text-white text-[10px] font-medium rounded-full min-w-[18px] h-[18px] flex items-center justify-center px-1"
      >
        {{ pendingCount > 99 ? '99+' : pendingCount }}
      </span>
    </button>

    <!-- Dropdown panel -->
    <div
      v-if="isOpen"
      class="absolute right-0 mt-2 w-80 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-50"
    >
      <!-- Header -->
      <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
        <h3 class="font-semibold text-gray-900 dark:text-gray-100 text-sm">{{ t('approvals.title') }}</h3>
        <button
          @click="togglePanel"
          class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- List -->
      <div class="max-h-96 overflow-y-auto">
        <!-- Empty state -->
        <div v-if="approvals.length === 0" class="p-8 text-center">
          <svg class="h-10 w-10 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
          </svg>
          <p class="mt-2 text-gray-500 dark:text-gray-400 text-sm">{{ t('approvals.noApprovals') }}</p>
        </div>

        <!-- Items -->
        <div v-else>
          <router-link
            v-for="a in approvals"
            :key="a.id"
            :to="`/workspace/${slug}/project/${a.project_id}/issues/${a.issue_id}`"
            class="block px-4 py-3 hover:bg-gray-50 dark:hover:bg-gray-700/50 border-b border-gray-100 dark:border-gray-700 transition-colors"
            @click="isOpen = false"
          >
            <p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
              {{ a.issue_key }}: {{ a.issue_title }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate">
              {{ a.requester_name }} · {{ formatDate(a.created_at) }}
            </p>
          </router-link>
        </div>
      </div>

      <!-- Footer -->
      <div class="px-4 py-3 border-t border-gray-200 dark:border-gray-700 text-center">
        <router-link
          :to="`/workspace/${slug}/approvals`"
          class="text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300"
          @click="isOpen = false"
        >
          {{ t('approvals.viewAll') }}
        </router-link>
      </div>
    </div>
  </div>
</template>
