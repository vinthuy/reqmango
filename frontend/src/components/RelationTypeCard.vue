<template>
  <div class="border rounded-lg overflow-hidden" :style="{ borderColor: color + '40', backgroundColor: color + '08' }">
    <!-- Card header -->
    <div class="flex items-center justify-between px-3 py-2 border-b" :style="{ backgroundColor: color + '15', borderColor: color + '30' }">
      <div class="flex items-center gap-2">
        <span class="text-xs font-semibold" :style="{ color }">{{ typeName.toUpperCase() }}</span>
        <span
          v-if="items.length > 0"
          class="px-1.5 py-0.5 rounded-full text-[10px] font-medium"
          :style="{ backgroundColor: color + '30', color }"
        >{{ items.length }}</span>
      </div>
      <!-- Add dropdown -->
      <div class="relative">
        <button
          data-test="add-relation"
          class="text-[10px] font-medium hover:underline"
          :style="{ color }"
          @click="showMenu = !showMenu"
        >+ {{ t('subIssue.add') }}</button>
        <div
          v-if="showMenu"
          class="absolute right-0 top-full mt-1 w-44 bg-white border border-gray-200 rounded-lg shadow-lg z-10 py-1"
        >
          <button
            data-test="add-existing-relation"
            class="w-full text-left px-3 py-2 text-xs text-gray-700 hover:bg-gray-50 flex items-center gap-2"
            @click="selectExisting"
          >
            <span class="text-blue-500">&#9744;</span>
            {{ t('subIssue.selectExisting') }}
          </button>
          <button
            data-test="quick-create-relation"
            class="w-full text-left px-3 py-2 text-xs text-gray-700 hover:bg-gray-50 flex items-center gap-2"
            @click="startQuickCreate"
          >
            <span class="text-green-500">+</span>
            {{ t('subIssue.quickCreate') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Quick-create inline input -->
    <div v-if="quickCreating" class="px-3 py-2 border-b bg-white" :style="{ borderColor: color + '30' }">
      <input
        ref="quickInput"
        v-model="quickName"
        type="text"
        class="w-full px-2 py-1.5 border rounded text-xs focus:outline-none"
        :style="{ borderColor: color + '60' }"
        :placeholder="t('subIssue.quickCreatePlaceholder')"
        @keydown.enter="submitQuickCreate"
        @keydown.escape="cancelQuickCreate"
        @blur="cancelQuickCreate"
      />
    </div>

    <!-- Empty state -->
    <div v-if="items.length === 0 && !quickCreating" class="px-4 py-6 text-center text-xs text-gray-400">
      {{ t('issueKanban.noRelations') }}
    </div>

    <!-- Table -->
    <div v-if="items.length > 0" class="overflow-x-auto">
      <table class="w-full text-[10px]">
        <thead>
          <tr class="border-b text-gray-500" :style="{ borderColor: color + '20' }">
            <th class="px-3 py-1.5 text-left">{{ t('issue.type') }}</th>
            <th class="px-2 py-1.5 text-left">ID</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.title') }}</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.status') }}</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.priority') }}</th>
            <th class="px-2 py-1.5 text-left">{{ t('issue.assignee') }}</th>
            <th class="px-2 py-1.5 text-right">{{ t('issue.targetDate') }}</th>
            <th class="px-2 py-1.5 w-8"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in items"
            :key="item.id"
            class="border-b last:border-b-0 transition-colors hover:bg-gray-50/50"
            :style="{ borderColor: color + '10' }"
          >
            <!-- Type badge -->
            <td class="px-3 py-2">
              <span
                class="px-1.5 py-0.5 rounded text-[10px] font-medium whitespace-nowrap"
                :style="{ backgroundColor: getIssueType(item.related_issue)?.color + '20', color: getIssueType(item.related_issue)?.color }"
              >{{ getIssueType(item.related_issue)?.name || '—' }}</span>
            </td>
            <!-- ID -->
            <td class="px-2 py-2 font-mono text-gray-400">#{{ item.related_issue.sequence_id }}</td>
            <!-- Title (clickable) -->
            <td class="px-2 py-2">
              <span
                class="relation-type-clickable text-xs font-medium text-gray-800 cursor-pointer hover:text-indigo-600"
                @click="$emit('navigate', item.related_issue_id)"
              >{{ item.related_issue.name }}</span>
            </td>
            <!-- State -->
            <td class="px-2 py-2">
              <span class="flex items-center gap-1">
                <span
                  class="w-1.5 h-1.5 rounded-full"
                  :style="{ backgroundColor: stateColor(item.related_issue.state_group) }"
                ></span>
                {{ item.related_issue.state_name }}
              </span>
            </td>
            <!-- Priority -->
            <td class="px-2 py-2">
              <span class="flex items-center gap-1">
                <span
                  class="w-1.5 h-1.5 rounded-full"
                  :style="{ backgroundColor: priorityColor(item.related_issue.priority) }"
                ></span>
                {{ t(`issue.priority${(item.related_issue.priority || 'none').charAt(0).toUpperCase() + (item.related_issue.priority || 'none').slice(1)}`) }}
              </span>
            </td>
            <!-- Assignee -->
            <td class="px-2 py-2 text-gray-500">
              {{ item.related_issue.assignees?.[0]?.display_name || item.related_issue.assignees?.[0]?.username || '—' }}
            </td>
            <!-- Due date -->
            <td class="px-2 py-2 text-gray-400 text-right">
              {{ item.related_issue.target_date ? formatDate(item.related_issue.target_date) : '—' }}
            </td>
            <!-- Remove -->
            <td class="px-2 py-2 text-center">
              <button
                data-test="remove-relation"
                class="text-gray-300 hover:text-red-500 text-sm leading-none"
                @click="$emit('remove', item.id)"
              >&times;</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { stateColor, priorityColor, formatDate } from '@/composables/useRelationHelpers'

const { t } = useI18n()

interface IssueType {
  id: number
  name: string
  color: string
}

interface RelationItem {
  id: number
  related_issue: {
    id: number
    sequence_id: number
    name: string
    state_name: string
    state_group: string
    priority: string
    assignees?: Array<{ id: number; display_name?: string; username?: string }>
    start_date?: string | null
    target_date?: string | null
    issue_type?: IssueType
  }
  related_issue_id: number
}

const props = defineProps<{
  typeName: string
  items: RelationItem[]
  color: string
  issueTypes: IssueType[]
}>()

const emit = defineEmits<{
  add: []
  remove: [relationId: number]
  navigate: [issueId: number]
  'add-existing': []
  'quick-create': [name: string]
}>()

const showMenu = ref(false)
const quickCreating = ref(false)
const quickName = ref('')
const quickInput = ref<HTMLInputElement | null>(null)

function selectExisting() {
  showMenu.value = false
  emit('add-existing')
}

async function startQuickCreate() {
  showMenu.value = false
  quickCreating.value = true
  quickName.value = ''
  await nextTick()
  quickInput.value?.focus()
}

function submitQuickCreate() {
  const name = quickName.value.trim()
  if (!name) return
  quickCreating.value = false
  quickName.value = ''
  emit('quick-create', name)
}

function cancelQuickCreate() {
  quickCreating.value = false
  quickName.value = ''
}

function getIssueType(issue: RelationItem['related_issue']) {
  if (!props.issueTypes) return null
  return props.issueTypes.find((t) => t.id === issue?.issue_type?.id) || issue?.issue_type || null
}
</script>
