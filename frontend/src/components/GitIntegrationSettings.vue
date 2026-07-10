<template>
  <div class="git-integration-settings">
    <div class="mb-4">
      <h3 class="text-lg font-semibold text-gray-800">{{ t('gitIntegration.title') }}</h3>
      <p class="text-sm text-gray-500 mt-1">{{ t('gitIntegration.description') }}</p>
    </div>

    <div v-if="!integration" class="bg-white border border-gray-200 rounded-lg p-6">
      <div class="text-center">
        <div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg class="w-8 h-8 text-gray-400" fill="currentColor" viewBox="0 0 24 24">
            <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.7-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.026A9.578 9.578 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.295 2.747-1.026 2.747-1.026.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .268.18.578.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
          </svg>
        </div>
        <h4 class="text-sm font-medium text-gray-700 mb-2">{{ t('gitIntegration.noIntegration') }}</h4>
        <p class="text-sm text-gray-500 mb-4">{{ t('gitIntegration.noIntegrationDesc') }}</p>
        <button
          class="px-4 py-2 bg-neutral-900 text-white text-sm rounded-md hover:bg-neutral-800 transition-colors"
          @click="showForm = true"
        >
          {{ t('gitIntegration.connect') }}
        </button>
      </div>
    </div>

    <div v-else class="space-y-4">
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 bg-gray-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-gray-600" fill="currentColor" viewBox="0 0 24 24">
                <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.7-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.026A9.578 9.578 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.295 2.747-1.026 2.747-1.026.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .268.18.578.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
              </svg>
            </div>
            <div>
              <div class="font-medium text-gray-800">{{ integration.repo_name }}</div>
              <div class="text-sm text-gray-500">{{ integration.repo_url }}</div>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span :class="integration.active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'" class="px-2 py-1 text-xs font-medium rounded-full">
              {{ integration.active ? t('gitIntegration.connected') : t('gitIntegration.disconnected') }}
            </span>
            <button class="text-gray-400 hover:text-red-500" @click="handleDelete">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h4 class="text-sm font-medium text-gray-700 mb-3">{{ t('gitIntegration.syncOptions') }}</h4>
        <div class="space-y-3">
          <label class="flex items-center gap-3">
            <input type="checkbox" :checked="integration.sync_prs" class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" @change="handleSyncChange('sync_prs', $event)" />
            <span class="text-sm text-gray-700">{{ t('gitIntegration.syncPRs') }}</span>
          </label>
          <label class="flex items-center gap-3">
            <input type="checkbox" :checked="integration.sync_commits" class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" @change="handleSyncChange('sync_commits', $event)" />
            <span class="text-sm text-gray-700">{{ t('gitIntegration.syncCommits') }}</span>
          </label>
          <label class="flex items-center gap-3">
            <input type="checkbox" :checked="integration.sync_branches" class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" @change="handleSyncChange('sync_branches', $event)" />
            <span class="text-sm text-gray-700">{{ t('gitIntegration.syncBranches') }}</span>
          </label>
        </div>
      </div>

      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h4 class="text-sm font-medium text-gray-700 mb-3">{{ t('gitIntegration.webhookSetup') }}</h4>
        <div class="text-sm text-gray-500 mb-2">{{ t('gitIntegration.webhookDesc') }}</div>
        <div class="bg-gray-50 rounded-md p-3 font-mono text-xs text-gray-600 break-all">
          {{ webhookUrl }}
        </div>
        <div class="mt-2 text-xs text-gray-400">{{ t('gitIntegration.webhookSecret') }}: {{ integration.webhook_secret || '-' }}</div>
      </div>
    </div>

    <div v-if="showForm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showForm = false">
      <div class="bg-white rounded-lg w-full max-w-md mx-4 p-6">
        <h4 class="text-lg font-semibold text-gray-800 mb-4">{{ t('gitIntegration.connect') }}</h4>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('gitIntegration.provider') }}</label>
            <select v-model="form.provider" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500">
              <option value="github">GitHub</option>
              <option value="gitlab">GitLab</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('gitIntegration.repoUrl') }}</label>
            <input type="text" v-model="form.repo_url" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" placeholder="https://github.com/username/repo" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('gitIntegration.repoName') }}</label>
            <input type="text" v-model="form.repo_name" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" placeholder="username/repo" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('gitIntegration.accessToken') }}</label>
            <input type="password" v-model="form.access_token" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" placeholder="ghp_..." />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('gitIntegration.webhookSecret') }}</label>
            <input type="text" v-model="form.webhook_secret" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" placeholder="random-secret" />
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-6">
          <button class="px-4 py-2 text-sm text-gray-600 hover:text-gray-800" @click="showForm = false">
            {{ t('common.cancel') }}
          </button>
          <button class="px-4 py-2 bg-neutral-900 text-white text-sm rounded-md hover:bg-neutral-800 transition-colors" @click="handleConnect">
            {{ t('gitIntegration.connect') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
import { getGitIntegration, createGitIntegration, updateGitIntegration, deleteGitIntegration, type GitIntegration } from '@/api/git-integration'

const props = defineProps<{
  workspaceId: number
  projectId: number
}>()

const { t } = useI18n()
const toast = useToast()
const { confirm } = useConfirm()

const integration = ref<GitIntegration | null>(null)
const showForm = ref(false)
const loading = ref(false)

const form = ref({
  provider: 'github' as const,
  repo_url: '',
  repo_name: '',
  access_token: '',
  webhook_secret: '',
})

const webhookUrl = computed(() => {
  const baseUrl = window.location.origin
  return `${baseUrl}/api/v1/webhook/git/${props.projectId}`
})

async function loadIntegration() {
  try {
    integration.value = await getGitIntegration(props.workspaceId, props.projectId)
  } catch {
    integration.value = null
  }
}

async function handleConnect() {
  loading.value = true
  try {
    await createGitIntegration(props.workspaceId, props.projectId, {
      provider: form.value.provider,
      repo_url: form.value.repo_url,
      repo_name: form.value.repo_name,
      access_token: form.value.access_token,
      webhook_secret: form.value.webhook_secret,
    })
    showForm.value = false
    await loadIntegration()
    toast.success(t('gitIntegration.connectSuccess'))
  } catch (error: any) {
    toast.error(error?.response?.data?.message || t('gitIntegration.connectFailed'))
  } finally {
    loading.value = false
  }
}

async function handleSyncChange(key: keyof GitIntegration, event: Event) {
  if (!integration.value) return
  const target = event.target as HTMLInputElement
  try {
    integration.value = await updateGitIntegration(props.workspaceId, props.projectId, {
      [key]: target.checked,
    })
  } catch (error: any) {
    toast.error(error?.response?.data?.message || t('common.error'))
  }
}

async function handleDelete() {
  const confirmed = await confirm(t('gitIntegration.deleteConfirm'))
  if (!confirmed) return
  try {
    await deleteGitIntegration(props.workspaceId, props.projectId)
    integration.value = null
    toast.success(t('gitIntegration.deleteSuccess'))
  } catch (error: any) {
    toast.error(error?.response?.data?.message || t('gitIntegration.deleteFailed'))
  }
}

onMounted(loadIntegration)
</script>