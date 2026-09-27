<template>
  <div class="space-y-4" data-test="ai-tab">
    <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">🤖 {{ t('ai.title') }}</h3>
        <button
          data-test="ai-open-copilot"
          @click="emit('open-copilot')"
          class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white text-sm rounded-lg flex items-center gap-1.5"
        >
          <span>💬</span> {{ t('ai.copilotTitle') }}
        </button>
      </div>
      <div class="grid grid-cols-2 gap-3">
        <button
          v-for="action in aiActions"
          :key="action.key"
          :data-test="`ai-action-${action.key}`"
          :disabled="loading && action.key !== 'copilot'"
          @click="executeAction(action.key)"
          class="flex items-start gap-3 p-3 border border-gray-200 dark:border-gray-700 rounded-lg hover:border-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 transition-colors text-left disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span class="text-xl">{{ action.icon }}</span>
          <div>
            <div class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ action.title }}</div>
            <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ action.description }}</div>
          </div>
        </button>
      </div>

      <div class="mt-4 space-y-3" data-test="ai-results">
        <p v-if="loading" class="text-sm text-indigo-600 dark:text-indigo-300" data-test="ai-loading">
          {{ t('ai.tabLoading') }}
        </p>
        <p v-else-if="error" class="text-sm text-red-600" data-test="ai-error">{{ error }}</p>

        <template v-if="!loading && analyzeResult">
          <div class="rounded-lg border border-gray-200 dark:border-gray-700 p-3 space-y-3" data-test="ai-analyze-result">
            <div v-if="analyzeResult.summary">
              <h4 class="text-sm font-medium text-gray-800 dark:text-gray-200 mb-1">{{ t('ai.analysisSummary') }}</h4>
              <p class="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap">{{ analyzeResult.summary }}</p>
            </div>
            <div v-if="analyzeResult.insights?.length">
              <h4 class="text-sm font-medium text-gray-800 dark:text-gray-200 mb-1">{{ t('ai.insights') }}</h4>
              <ul class="list-disc list-inside text-sm text-gray-700 dark:text-gray-300 space-y-1">
                <li v-for="(insight, idx) in analyzeResult.insights" :key="idx">{{ insight }}</li>
              </ul>
            </div>
            <div v-if="analyzeResult.mode === 'risk'" data-test="ai-risks">
              <h4 class="text-sm font-medium text-gray-800 dark:text-gray-200 mb-1">{{ t('ai.risksTitle') }}</h4>
              <p v-if="!analyzeResult.risks?.length" class="text-sm text-emerald-700 dark:text-emerald-300">{{ t('ai.noRisks') }}</p>
              <ul v-else class="space-y-2">
                <li
                  v-for="(risk, idx) in analyzeResult.risks"
                  :key="idx"
                  data-test="ai-risk-item"
                  class="flex items-start gap-2 text-sm rounded-md px-2 py-1.5"
                  :class="RISK_STYLES[risk.level]?.row || RISK_STYLES.medium.row"
                >
                  <span
                    class="shrink-0 px-1.5 py-0.5 rounded text-[10px] font-semibold"
                    :class="RISK_STYLES[risk.level]?.badge || RISK_STYLES.medium.badge"
                  >{{ t(`ai.riskLevel.${risk.level}`) }}</span>
                  <div class="min-w-0">
                    <div class="font-medium text-gray-900 dark:text-gray-100">{{ risk.title }}</div>
                    <div v-if="risk.detail" class="text-xs text-gray-600 dark:text-gray-400 mt-0.5">{{ risk.detail }}</div>
                  </div>
                </li>
              </ul>
            </div>
            <div v-if="analyzeResult.next_steps?.length">
              <h4 class="text-sm font-medium text-gray-800 dark:text-gray-200 mb-1">{{ t('ai.nextStepsTitle') }}</h4>
              <ol class="list-decimal list-inside text-sm text-gray-700 dark:text-gray-300 space-y-1">
                <li v-for="(step, idx) in analyzeResult.next_steps" :key="idx" data-test="ai-next-step">{{ step }}</li>
              </ol>
            </div>
            <div v-if="analyzeResult.bottlenecks?.length">
              <h4 class="text-sm font-medium text-gray-800 dark:text-gray-200 mb-1">{{ t('ai.bottlenecks') }}</h4>
              <ul class="space-y-2">
                <li
                  v-for="bn in analyzeResult.bottlenecks"
                  :key="bn.issue_id"
                  class="text-sm text-amber-800 dark:text-amber-200 bg-amber-50 dark:bg-amber-900/20 rounded px-2 py-1.5"
                >
                  <span class="font-medium">{{ bn.issue_name }}</span>
                  <span class="text-xs text-amber-700 dark:text-amber-300 ml-2">
                    {{ t('ai.daysInState', { days: bn.days_in_state, state: bn.state_name }) }}
                  </span>
                </li>
              </ul>
            </div>
          </div>
        </template>

        <template v-if="!loading && labelSuggestionsVisible">
          <div class="rounded-lg border border-gray-200 dark:border-gray-700 p-3 space-y-2" data-test="ai-label-suggestions">
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ t('ai.suggestedLabelsTitle') }}</h4>
            <p v-if="!labelSuggestions.length" class="text-sm text-gray-500" data-test="ai-labels-empty">
              {{ t('ai.noLabelSuggestions') }}
            </p>
            <ul v-else class="space-y-2">
              <li
                v-for="(item, idx) in labelSuggestions"
                :key="labelSuggestionKey(item, idx)"
                class="flex items-start justify-between gap-3 text-sm border border-gray-100 dark:border-gray-700 rounded-md px-2 py-2"
              >
                <div class="min-w-0">
                  <div class="font-medium text-gray-900 dark:text-gray-100">{{ labelSuggestionName(item) }}</div>
                  <div v-if="labelSuggestionReason(item)" class="text-xs text-gray-500 mt-0.5">{{ labelSuggestionReason(item) }}</div>
                </div>
                <button
                  v-if="labelSuggestionId(item)"
                  type="button"
                  class="shrink-0 px-2 py-1 text-xs bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                  :disabled="applyingLabelId === labelSuggestionId(item)"
                  :data-test="`ai-apply-label-${labelSuggestionId(item)}`"
                  @click="applySuggestedLabel(labelSuggestionId(item)!)"
                >
                  {{ t('ai.applyToIssue') }}
                </button>
              </li>
            </ul>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import issueApi from '@/api/issue'
import { analyzeWithAI, suggestLabels } from '@/api/ai'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'

const props = defineProps<{
  issueId: number
  projectId: number
  issue: any
  labels: Array<{ id: number; name: string; color: string }>
}>()

const emit = defineEmits<{
  (e: 'open-copilot'): void
  (e: 'issue-updated', issue: any): void
}>()

const { t } = useI18n()
const toast = useToast()

const RISK_STYLES: Record<string, { row: string; badge: string }> = {
  high: { row: 'bg-red-50 dark:bg-red-900/20', badge: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300' },
  medium: { row: 'bg-amber-50 dark:bg-amber-900/20', badge: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300' },
  low: { row: 'bg-gray-50 dark:bg-gray-700/30', badge: 'bg-gray-200 text-gray-600 dark:bg-gray-700 dark:text-gray-300' },
}

const loading = ref(false)
const error = ref('')
const analyzeResult = ref<any>(null)
const labelSuggestions = ref<any[]>([])
const labelSuggestionsVisible = ref(false)
const applyingLabelId = ref<number | null>(null)

watch(() => props.issueId, () => {
  loading.value = false
  error.value = ''
  analyzeResult.value = null
  labelSuggestions.value = []
  labelSuggestionsVisible.value = false
})

const aiActions = computed(() => [
  { key: 'summarize', icon: '📝', title: t('ai.summarizeLabel'), description: t('ai.summarizeIssue') },
  { key: 'risk', icon: '⚠️', title: t('ai.riskLabel'), description: t('ai.riskAnalysis') },
  { key: 'suggest', icon: '💡', title: t('ai.suggestLabel'), description: t('ai.suggestSteps') },
  { key: 'copilot', icon: '🤖', title: t('ai.copilotTitle'), description: t('ai.readyHint') },
])

function normalizeLabelSuggestions(res: any): any[] {
  if (Array.isArray(res)) return res
  if (Array.isArray(res?.suggested_labels)) return res.suggested_labels
  if (Array.isArray(res?.labels)) return res.labels
  return []
}

function labelSuggestionName(item: any): string {
  if (typeof item === 'string') return item
  return item?.label_name || item?.name || String(item?.label_id ?? item?.id ?? '')
}

function labelSuggestionReason(item: any): string {
  if (typeof item === 'string') return ''
  return item?.reason || ''
}

function labelSuggestionId(item: any): number | null {
  if (typeof item === 'string') {
    const found = props.labels.find((l) => l.name === item)
    return found?.id ?? null
  }
  const id = item?.label_id ?? item?.id
  return typeof id === 'number' && id > 0 ? id : null
}

function labelSuggestionKey(item: any, idx: number): string | number {
  return labelSuggestionId(item) ?? (labelSuggestionName(item) || idx)
}

async function executeAction(action: string) {
  if (action === 'copilot') {
    emit('open-copilot')
    return
  }
  if (!props.projectId || !props.issueId) return
  const requestedIssueId = props.issueId
  loading.value = true
  error.value = ''
  try {
    if (action === 'summarize' || action === 'risk') {
      const res = await analyzeWithAI(props.projectId, requestedIssueId, action === 'risk' ? 'risk' : 'summary')
      if (requestedIssueId !== props.issueId) return
      analyzeResult.value = res
      labelSuggestions.value = []
      labelSuggestionsVisible.value = false
    } else if (action === 'suggest') {
      analyzeResult.value = null
      labelSuggestionsVisible.value = false
      const [steps, labels] = await Promise.allSettled([
        analyzeWithAI(props.projectId, requestedIssueId, 'next_steps'),
        suggestLabels(props.projectId, requestedIssueId, {
          name: props.issue?.name || `Issue #${requestedIssueId}`,
          description: props.issue?.description_html || props.issue?.description_text || '',
        }),
      ])
      if (requestedIssueId !== props.issueId) return
      if (steps.status === 'rejected' && labels.status === 'rejected') throw steps.reason
      analyzeResult.value = steps.status === 'fulfilled' ? steps.value : null
      if (labels.status === 'fulfilled') {
        labelSuggestions.value = normalizeLabelSuggestions(labels.value)
        labelSuggestionsVisible.value = true
      }
    }
  } catch (e: any) {
    if (requestedIssueId !== props.issueId) return
    error.value = e?.response?.data?.message || e?.message || t('ai.connectionFailed')
  } finally {
    if (requestedIssueId === props.issueId) loading.value = false
  }
}

async function applySuggestedLabel(labelId: number) {
  if (!labelId || !props.issueId) return
  applyingLabelId.value = labelId
  try {
    await issueApi.addIssueLabel(props.issueId, labelId)
    const updated = await issueApi.getIssue(props.issueId)
    emit('issue-updated', updated)
    toast.success(t('ai.labelApplied'))
  } catch (e: any) {
    toast.error(e?.response?.data?.message || t('ai.labelApplyFailed'))
  } finally {
    applyingLabelId.value = null
  }
}
</script>
