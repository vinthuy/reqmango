<template>
  <div
    v-if="pending.length > 0"
    data-test="comment-suggestions"
    class="mt-2 rounded-lg border border-indigo-200 bg-indigo-50/60 px-3 py-2"
  >
    <div class="flex items-center justify-between gap-2 mb-1.5">
      <span class="text-[11px] font-semibold text-indigo-700">
        {{ t('comment.suggestionsTitle', { count: pending.length }) }}
      </span>
      <button
        v-if="applicable.length > 0"
        data-test="apply-all-suggestions"
        :disabled="busy !== null"
        class="px-2 py-0.5 text-[11px] font-medium rounded bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 transition-colors"
        @click="applyAll()"
      >
        {{ busy === '__all__' ? t('comment.applying') : t('comment.applyAll') }}
      </button>
      <span v-else class="text-[11px] text-gray-400">{{ t('comment.noChangesNeeded') }}</span>
    </div>

    <ul class="space-y-1">
      <li
        v-for="sg in pending"
        :key="sg.field"
        data-test="suggestion-row"
        class="flex items-start gap-2 text-[11px] leading-5"
      >
        <span class="shrink-0 w-14 text-gray-500">{{ fieldLabel(sg.field) }}</span>
        <span class="flex-1 min-w-0 text-gray-700">
          <template v-if="sg.noop">
            <span class="font-medium text-gray-600">{{ t('comment.keepValue', { value: sg.current || sg.label }) }}</span>
          </template>
          <template v-else>
            <span v-if="sg.current" class="text-gray-400 line-through mr-1">{{ sg.current }}</span>
            <span class="font-medium text-indigo-700">{{ sg.label }}</span>
          </template>
          <span v-if="sg.reason" class="block text-gray-500">{{ sg.reason }}</span>
        </span>
        <button
          v-if="!sg.noop"
          data-test="apply-suggestion"
          :disabled="busy !== null"
          class="shrink-0 px-1.5 py-0.5 rounded border border-indigo-300 text-indigo-700 hover:bg-indigo-100 disabled:opacity-50 transition-colors"
          @click="apply([sg.field])"
        >
          {{ busy === sg.field ? t('comment.applying') : t('comment.apply') }}
        </button>
        <span v-else data-test="suggestion-noop" class="shrink-0 px-1.5 py-0.5 text-gray-400">
          {{ t('comment.noChange') }}
        </span>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import commentApi from '@/api/comment'
import { useToast } from '@/composables/useToast'
import { useI18n } from '@/composables/useI18n'
import type { Comment, SuggestionField } from '@/types/comment'

const props = defineProps<{
  issueId: number
  comment: Comment
}>()

const emit = defineEmits<{
  (e: 'applied', applied: string[]): void
}>()

const { t } = useI18n()
const toast = useToast()

/** Field currently being applied, '__all__' for the bulk action, null when idle. */
const busy = ref<string | null>(null)
/** Locally hides the block as soon as the server confirms the proposals were adopted. */
const settled = ref(false)

const pending = computed(() => {
  if (settled.value || props.comment.suggestions_applied_at) return []
  return props.comment.suggestions ?? []
})

/** Proposals that would actually change something; no-ops only carry reasoning. */
const applicable = computed(() => pending.value.filter(sg => !sg.noop))

function fieldLabel(field: SuggestionField): string {
  return t(`comment.suggestionField.${field}`)
}

/** Apply every proposal that would actually change a field. */
function applyAll() {
  return apply(applicable.value.map(sg => sg.field))
}

async function apply(fields: SuggestionField[]) {
  if (busy.value !== null || fields.length === 0) return
  busy.value = fields.length === 1 ? fields[0] : '__all__'
  try {
    const result = await commentApi.applySuggestions(props.issueId, props.comment.id, fields)
    // Drop the adopted rows locally; the server marks the comment as settled too.
    const applied = result?.applied ?? []
    // A proposal can be refused on its own (a type this project cannot use, a
    // transition its workflow forbids). The rest still landed, so report the
    // refusal instead of pretending the whole batch succeeded.
    const skipped = result?.skipped ?? []
    const rest = (props.comment.suggestions ?? []).filter(
      sg => !applied.includes(sg.field) && !skipped.some(sk => sk.field === sg.field)
    )
    // Always reflect the remainder on the comment itself, so the block stays
    // settled even if the component is remounted before the data is refetched.
    props.comment.suggestions = rest
    if (rest.length === 0) {
      settled.value = true
    }
    if (applied.length > 0) {
      toast.success(t('comment.applySuccess', { count: applied.length }))
    }
    if (skipped.length > 0) {
      toast.warning(skipped.map(sk => `${fieldLabel(sk.field)}：${sk.reason}`).join('；'))
    }
    emit('applied', applied)
  } catch (e: any) {
    const status = e?.response?.status
    // 409 means someone already adopted these; treat it as settled, not an error.
    if (status === 409) {
      settled.value = true
      emit('applied', [])
      return
    }
    toast.error(e?.response?.data?.message || e?.message || t('comment.applyFailed'))
  } finally {
    busy.value = null
  }
}
</script>
