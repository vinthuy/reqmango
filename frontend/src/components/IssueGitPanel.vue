<template>
  <div class="issue-git-panel">
    <div class="flex items-center justify-between mb-4">
      <h4 class="text-sm font-medium text-gray-700">{{ t('gitIntegration.title') }}</h4>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-8">
      <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-gray-900"></div>
    </div>

    <div v-else-if="links.length === 0" class="text-center py-8">
      <div class="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-3">
        <svg class="w-6 h-6 text-gray-400" fill="currentColor" viewBox="0 0 24 24">
          <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.7-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.026A9.578 9.578 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.295 2.747-1.026 2.747-1.026.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .268.18.578.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
        </svg>
      </div>
      <p class="text-sm text-gray-500">{{ t('gitIntegration.noIntegration') }}</p>
      <p class="text-xs text-gray-400 mt-1">{{ t('gitIntegration.noIntegrationDesc') }}</p>
    </div>

    <div v-else class="space-y-4">
      <div v-if="pullRequests.length > 0">
        <h5 class="text-xs font-medium text-gray-500 uppercase tracking-wider mb-2">{{ t('gitIntegration.pullRequests') }}</h5>
        <div class="space-y-2">
          <div v-for="link in pullRequests" :key="link.id" class="bg-white border border-gray-200 rounded-lg p-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
                </svg>
                <a :href="link.git_url" target="_blank" rel="noopener" class="text-sm font-medium text-gray-700 hover:text-indigo-600 truncate">
                  {{ link.git_title }}
                </a>
              </div>
              <span :class="getStatusClass(link.git_state)" class="px-2 py-0.5 text-xs font-medium rounded-full">
                {{ link.git_state }}
              </span>
            </div>
            <div class="flex items-center gap-4 mt-2 text-xs text-gray-400">
              <span>{{ t('gitIntegration.author') }}: {{ link.git_author }}</span>
              <span>{{ t('gitIntegration.branch') }}: {{ link.git_branch }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="commits.length > 0">
        <h5 class="text-xs font-medium text-gray-500 uppercase tracking-wider mb-2">{{ t('gitIntegration.commits') }}</h5>
        <div class="space-y-2">
          <div v-for="link in commits" :key="link.id" class="bg-white border border-gray-200 rounded-lg p-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
                <a :href="link.git_url" target="_blank" rel="noopener" class="text-sm font-medium text-gray-700 hover:text-indigo-600 truncate">
                  {{ link.git_title }}
                </a>
              </div>
              <span class="text-xs text-gray-400">{{ link.git_author }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="branches.length > 0">
        <h5 class="text-xs font-medium text-gray-500 uppercase tracking-wider mb-2">{{ t('gitIntegration.branches') }}</h5>
        <div class="space-y-2">
          <div v-for="link in branches" :key="link.id" class="bg-white border border-gray-200 rounded-lg p-3">
            <div class="flex items-center gap-2">
              <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
              </svg>
              <span class="text-sm font-medium text-gray-700">{{ link.git_branch }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { getIssueGitLinks, type GitIssueLink } from '@/api/git-integration'

const props = defineProps<{
  workspaceId: number
  issueId: number
}>()

const { t } = useI18n()

const links = ref<GitIssueLink[]>([])
const loading = ref(false)

const pullRequests = computed(() => links.value.filter(link => link.git_type === 'pull_request'))
const commits = computed(() => links.value.filter(link => link.git_type === 'commit'))
const branches = computed(() => links.value.filter(link => link.git_type === 'branch'))

function getStatusClass(state: string) {
  switch (state.toLowerCase()) {
    case 'open':
      return 'bg-green-100 text-green-700'
    case 'closed':
      return 'bg-gray-100 text-gray-700'
    case 'merged':
      return 'bg-purple-100 text-purple-700'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}

async function loadLinks() {
  loading.value = true
  try {
    links.value = await getIssueGitLinks(props.workspaceId, props.issueId)
  } catch {
    links.value = []
  } finally {
    loading.value = false
  }
}

onMounted(loadLinks)

watch(() => props.issueId, loadLinks)
</script>