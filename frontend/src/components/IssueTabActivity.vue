<template>
  <div class="activity-timeline">
    <!-- Loading -->
    <div v-if="loading" class="py-8 text-center text-gray-400 text-sm">
      {{ t('common.loading') }}
    </div>

    <!-- Error -->
    <div v-else-if="error" class="py-8 text-center text-red-400 text-sm">
      {{ t('common.error') }}
    </div>

    <!-- Empty -->
    <div v-else-if="groupedActivities.length === 0" class="py-12 text-center">
      <div class="text-3xl mb-3 opacity-30">📋</div>
      <p class="text-sm text-gray-400">{{ t('issue.noActivity') }}</p>
    </div>

    <!-- Timeline -->
    <div v-else class="relative pl-8">
      <!-- Vertical line -->
      <div class="absolute left-[11px] top-2 bottom-2 w-0.5 bg-gray-200"></div>

      <div v-for="group in groupedActivities" :key="group.date" class="mb-6">
        <!-- Date header -->
        <div class="flex items-center gap-3 mb-3 -ml-8">
          <span class="w-6 h-6 rounded-full bg-indigo-100 text-indigo-600 flex items-center justify-center text-[10px] font-bold shrink-0">●</span>
          <span class="text-xs font-semibold text-gray-500 uppercase tracking-wide">{{ group.label }}</span>
        </div>

        <!-- Activity items for this date -->
        <div class="space-y-1">
          <div
            v-for="activity in group.items"
            :key="activity.id"
            class="group relative flex items-start gap-3 py-2.5 -ml-8 pl-8 rounded-lg hover:bg-gray-50/70 transition-colors"
          >
            <!-- Timeline dot -->
            <div
              class="absolute left-[8px] top-[14px] w-2 h-2 rounded-full border-2 border-white ring-2 shrink-0"
              :style="{ backgroundColor: getActivityColor(activity) }"
              :class="getActivityRingColor(activity)"
            ></div>

            <!-- Actor avatar -->
            <div class="w-6 h-6 rounded-full shrink-0 flex items-center justify-center text-[10px] font-bold text-white"
              :style="{ backgroundColor: getActorColor(activity.actor_id || 0) }"
            >
              {{ getActorInitial(activity) }}
            </div>

            <!-- Activity content -->
            <div class="flex-1 min-w-0">
              <div class="flex items-baseline gap-2 flex-wrap">
                <!-- Actor name -->
                <span class="text-sm font-medium text-gray-800">
                  {{ activity.actor_display_name || t('issue.unknownUser') }}
                </span>

                <!-- Activity message -->
                <span class="text-sm text-gray-600">
                  {{ getActivityMessage(activity) }}
                </span>

                <!-- Timestamp -->
                <span
                  class="text-xs text-gray-400 whitespace-nowrap"
                  :title="formatFullDate(activity.created_at)"
                >{{ formatRelativeTime(activity.created_at) }}</span>
              </div>

              <!-- Diff: old → new value (for state/priority changes) -->
              <div v-if="shouldShowDiff(activity)" class="mt-1 flex items-center gap-1.5 text-xs">
                <span class="px-1.5 py-0.5 rounded bg-red-50 text-red-600 line-through">{{ activity.old_value || '—' }}</span>
                <span class="text-gray-400">→</span>
                <span class="px-1.5 py-0.5 rounded bg-green-50 text-green-600 font-medium">{{ activity.new_value || '—' }}</span>
              </div>

              <!-- Comment content -->
              <div v-if="activity.comment" class="mt-1.5 text-sm text-gray-600 bg-white border border-gray-200 rounded-lg p-2.5">
                {{ activity.comment }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Load more -->
      <div v-if="hasMore" class="text-center py-4 -ml-8">
        <button
          @click="loadMore"
          :disabled="loadingMore"
          class="text-xs text-indigo-600 hover:text-indigo-700 font-medium disabled:opacity-50"
        >
          {{ loadingMore ? t('common.loading') : t('issue.loadMoreActivity') }}
        </button>
      </div>
    </div>

    <!-- Automation execution history -->
    <div v-if="!loading && groupedActivities.length > 0" class="mt-8 pt-6 border-t border-gray-200">
      <AutomationHistory :issue-id="props.issueId" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { getIssueActivities } from '@/api/issue'
import AutomationHistory from '@/components/AutomationHistory.vue'

const { t } = useI18n()

interface Activity {
  id: number
  created_at: string
  verb: string
  field?: string | null
  old_value?: string | null
  new_value?: string | null
  comment?: string | null
  actor_id?: number | null
  actor_display_name?: string
  actor_avatar?: string
}

const props = defineProps<{
  issueId: number
}>()

const loading = ref(true)
const loadingMore = ref(false)
const error = ref(false)
const activities = ref<Activity[]>([])
const pageSize = 30
const offset = ref(0)
const hasMore = ref(true)

// Actor avatar colors (consistent hash)
const AVATAR_COLORS = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316']

function getActorColor(id: number): string {
  return AVATAR_COLORS[id % AVATAR_COLORS.length]
}

function getActorInitial(activity: Activity): string {
  const name = activity.actor_display_name || ''
  return name.charAt(0).toUpperCase() || '?'
}

// Activity type colors
function getActivityColor(activity: Activity): string {
  if (activity.verb === 'created') return '#10b981'
  if (activity.verb === 'moved' || activity.verb === 'converted' || activity.verb === 'merged') return '#06b6d4'
  if (activity.verb === 'relation_added') return '#6366f1'
  if (activity.verb === 'relation_removed') return '#ef4444'
  if (activity.field === 'state_id') return '#8b5cf6'
  if (activity.field === 'priority') return '#f59e0b'
  if (activity.field === 'assignees') return '#6366f1'
  if (activity.field === 'name') return '#14b8a6'
  if (activity.field === 'description') return '#ec4899'
  if (activity.field === 'start_date' || activity.field === 'target_date') return '#f97316'
  if (activity.field === 'cycle') return '#8b5cf6'
  if (activity.field === 'module') return '#14b8a6'
  if (activity.field === 'labels') return '#f59e0b'
  if (activity.field === 'custom_field') return '#06b6d4'
  return '#94a3b8'
}

function getActivityRingColor(activity: Activity): string {
  if (activity.verb === 'created') return 'ring-green-200'
  if (activity.verb === 'relation_added' || activity.verb === 'relation_removed') return 'ring-indigo-200'
  if (activity.verb === 'moved' || activity.verb === 'converted' || activity.verb === 'merged') return 'ring-cyan-200'
  if (activity.field === 'state_id') return 'ring-purple-200'
  if (activity.field === 'priority' || activity.field === 'labels') return 'ring-amber-200'
  if (activity.field === 'assignees') return 'ring-indigo-200'
  if (activity.field === 'description') return 'ring-pink-200'
  if (activity.field === 'start_date' || activity.field === 'target_date') return 'ring-orange-200'
  return 'ring-gray-200'
}

// Activity message builder
function getActivityMessage(activity: Activity): string {
  switch (activity.verb) {
    case 'created': return t('activity.created')
    case 'updated': return getFieldMessage(activity)
    case 'changed': return getFieldMessage(activity)
    case 'state_changed': return t('activity.changedState')
    case 'priority_changed': return t('activity.changedPriority')
    case 'assigned': return t('activity.changedAssignees')
    case 'labeled': return t('activity.changedLabels')
    case 'mentioned': return t('activity.mentioned')
    case 'commented': return t('activity.commented')
    case 'added_to_cycle': return t('activity.addedToCycle')
    case 'added_to_module': return t('activity.addedToModule')
    case 'relation_added': return t('activity.relationAdded')
    case 'relation_removed': return t('activity.relationRemoved')
    case 'moved': return t('activity.moved')
    case 'converted': return t('activity.converted')
    case 'merged': return t('activity.merged')
    default: return t('activity.updated')
  }
}

function getFieldMessage(activity: Activity): string {
  const field = activity.field
  switch (field) {
    case 'state_id': return t('activity.changedState')
    case 'priority': return t('activity.changedPriority')
    case 'assignees': return t('activity.changedAssignees')
    case 'name': return t('activity.changedTitle')
    case 'description': return t('activity.changedDescription')
    case 'parent': return t('activity.changedParent')
    case 'cycle': return t('activity.changedCycle')
    case 'labels': return t('activity.changedLabels')
    case 'module': return t('activity.changedModule')
    case 'start_date': return t('activity.changedStartDate')
    case 'target_date': return t('activity.changedTargetDate')
    case 'issue_type': return t('activity.changedType')
    case 'relation': return t('activity.changedRelation')
    case 'custom_field': return (activity.comment ? activity.comment + ' ' : '') + t('activity.changedCustomField')
    default: return t('activity.updated')
  }
}

function shouldShowDiff(activity: Activity): boolean {
  return !!(activity.old_value || activity.new_value) &&
    activity.field !== 'description' &&
    activity.field !== 'comment'
}

// Date handling
interface ActivityGroup {
  date: string
  label: string
  items: Activity[]
}

const groupedActivities = computed<ActivityGroup[]>(() => {
  const groups: Map<string, ActivityGroup> = new Map()

  for (const a of activities.value) {
    const d = new Date(a.created_at)
    const today = new Date()
    const yesterday = new Date(today)
    yesterday.setDate(yesterday.getDate() - 1)

    let dateKey: string
    let label: string

    if (isSameDay(d, today)) {
      dateKey = 'today'
      label = t('activity.today')
    } else if (isSameDay(d, yesterday)) {
      dateKey = 'yesterday'
      label = t('activity.yesterday')
    } else {
      dateKey = d.toISOString().split('T')[0]
      label = formatDateLabel(d)
    }

    if (!groups.has(dateKey)) {
      groups.set(dateKey, { date: dateKey, label, items: [] })
    }
    groups.get(dateKey)!.items.push(a)
  }

  return Array.from(groups.values())
})

function isSameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
}

function formatDateLabel(d: Date): string {
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  return `${months[d.getMonth()]} ${d.getDate()}, ${d.getFullYear()}`
}

// Time formatting
function formatRelativeTime(dateStr: string): string {
  const now = Date.now()
  const date = new Date(dateStr).getTime()
  const diff = now - date
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (seconds < 60) return t('activity.justNow')
  if (minutes < 60) return t('activity.minutesAgo', { n: minutes })
  if (hours < 24) return t('activity.hoursAgo', { n: hours })
  if (days < 7) return t('activity.daysAgo', { n: days })
  return formatDateLabel(new Date(dateStr))
}

function formatFullDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleString()
}

// Data loading
async function loadActivities(reset = true) {
  if (reset) {
    loading.value = true
    offset.value = 0
  } else {
    loadingMore.value = true
  }

  try {
    const data = await getIssueActivities(props.issueId, pageSize, offset.value)
    if (reset) {
      activities.value = data || []
    } else {
      activities.value = [...activities.value, ...(data || [])]
    }
    hasMore.value = (data || []).length >= pageSize
    if (!reset) offset.value += pageSize
  } catch (err) {
    console.error('Failed to load activities:', err)
    if (reset) error.value = true
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

async function loadMore() {
  offset.value += pageSize
  await loadActivities(false)
}

onMounted(() => loadActivities())
</script>
