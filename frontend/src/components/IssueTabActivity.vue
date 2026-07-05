<template>
  <div class="issue-tab-activity">
    <div v-if="loading" class="py-4 text-center text-gray-400">
      {{ t('common.loading') }}
    </div>

    <div v-else-if="error" class="py-4 text-center text-red-400">
      {{ t('common.error') }}
    </div>

    <div v-else-if="activities.length === 0" class="py-4 text-center text-gray-400">
      {{ t('issue.noActivity') }}
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="activity in activities"
        :key="activity.id"
        class="flex items-start gap-3 p-3 bg-gray-50 rounded-lg"
      >
        <div class="text-sm text-gray-500 whitespace-nowrap min-w-[140px]">
          {{ formatDate(activity.created_at) }}
        </div>
        <div class="text-sm text-gray-700">
          {{ activity.verb }}
          <span v-if="activity.field" class="text-gray-400 ml-1">
            &mdash; {{ activity.field }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { getIssueActivities } from '@/api/issue'

const { t } = useI18n()

interface Activity {
  id: number
  created_at: string
  verb: string
  field?: string
  comment?: string
}

const props = defineProps<{
  issueId: number
}>()

const loading = ref(true)
const error = ref(false)
const activities = ref<Activity[]>([])

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString()
}

onMounted(async () => {
  try {
    const data = await getIssueActivities(props.issueId)
    activities.value = data || []
  } catch (err) {
    console.error('Failed to load activities:', err)
    error.value = true
  } finally {
    loading.value = false
  }
})
</script>
