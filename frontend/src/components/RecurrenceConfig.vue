<template>
  <div class="border-t border-gray-100 pt-4 mt-4">
    <div class="flex items-center justify-between mb-3">
      <h4 class="text-sm font-semibold text-gray-700">🔄 {{ t('recurrence.title') }}</h4>
    </div>

    <div v-if="!rule" class="text-xs text-gray-400">
      <button @click="showForm = true" class="text-indigo-600 hover:text-indigo-700">{{ t('recurrence.setup') }}</button>
    </div>

    <div v-else class="bg-gray-50 rounded-lg p-3 text-xs space-y-2">
      <div class="flex items-center justify-between">
        <span>{{ t('recurrence.every') }} {{ rule.interval }} {{ t(`recurrence.frequency.${rule.frequency}`) || rule.frequency }}</span>
        <span :class="rule.is_active ? 'text-green-600' : 'text-gray-400'">{{ rule.is_active ? t('common.active') : t('recurrence.paused') }}</span>
      </div>
      <div>{{ t('recurrence.next') }}: {{ formatDate(rule.next_run) }}</div>
      <div v-if="rule.end_date">{{ t('recurrence.until') }}: {{ formatDate(rule.end_date) }}</div>
      <div class="flex gap-2 pt-1">
        <button @click="edit" class="text-indigo-600 hover:text-indigo-700">{{ t('common.edit') }}</button>
        <button @click="toggle" class="text-amber-600 hover:text-amber-700">{{ rule.is_active ? t('recurrence.pause') : t('recurrence.resume') }}</button>
        <button @click="remove" class="text-red-500 hover:text-red-700">{{ t('recurrence.remove') }}</button>
      </div>
    </div>

    <!-- Form Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showForm=false">
      <div class="bg-white rounded-xl p-6 w-full max-w-sm">
        <h3 class="text-lg font-semibold mb-4">{{ rule ? t('common.edit') : t('common.create') }} {{ t('recurrence.title') }}</h3>
        <div class="space-y-3">
          <div class="grid grid-cols-2 gap-2">
            <div>
              <label class="block text-xs font-medium mb-1">{{ t('recurrence.frequencyLabel') }}</label>
              <select v-model="form.frequency" class="w-full px-2 py-1.5 border rounded text-sm">
                <option value="daily">{{ t('recurrence.frequency.daily') }}</option>
                <option value="weekly">{{ t('recurrence.frequency.weekly') }}</option>
                <option value="monthly">{{ t('recurrence.frequency.monthly') }}</option>
                <option value="cron">{{ t('recurrence.frequency.cron') }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1">{{ t('recurrence.interval') }}</label>
              <input v-model.number="form.interval" type="number" min="1" class="w-full px-2 py-1.5 border rounded text-sm" />
            </div>
          </div>
          <div>
            <label class="block text-xs font-medium mb-1">{{ t('recurrence.nextRun') }}</label>
            <input v-model="form.next_run" type="datetime-local" class="w-full px-2 py-1.5 border rounded text-sm" />
          </div>
          <div v-if="form.frequency === 'cron'">
            <label class="block text-xs font-medium mb-1">{{ t('recurrence.cronExpr') }}</label>
            <input v-model="form.cron_expr" :placeholder="t('recurrence.cronPlaceholder')" class="w-full px-2 py-1.5 border rounded text-sm" />
          </div>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button @click="showForm=false" class="px-4 py-2 border rounded-lg text-sm">{{ t('common.cancel') }}</button>
          <button @click="save" :disabled="saving" class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm disabled:opacity-50">{{ saving?t('common.saving'):t('common.save') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import * as recApi from '@/api/recurrence'
import type { RecurrenceRule } from '@/types/recurrence'
import { useI18n } from '@/composables/useI18n'

const props = defineProps<{ issueId: number }>()
const { t } = useI18n()
const rule = ref<RecurrenceRule | null>(null)
const showForm = ref(false)
const saving = ref(false)
const form = reactive({ frequency: 'weekly', interval: 1, cron_expr: '', next_run: '', end_date: '' })

onMounted(() => load())

async function load() {
  try { rule.value = await recApi.getRecurrence(props.issueId) } catch (_) { rule.value = null }
}

function edit() {
  if (rule.value) {
    form.frequency = rule.value.frequency; form.interval = rule.value.interval
    form.cron_expr = rule.value.cron_expr || ''
    form.next_run = rule.value.next_run ? new Date(rule.value.next_run).toISOString().slice(0,16) : ''
  }
  showForm.value = true
}

async function save() {
  saving.value = true
  try {
    const payload: any = { frequency: form.frequency, interval: form.interval }
    if (form.next_run) payload.next_run = new Date(form.next_run).toISOString()
    if (form.cron_expr) payload.cron_expr = form.cron_expr
    if (form.end_date) payload.end_date = new Date(form.end_date).toISOString()
    if (rule.value) {
      await recApi.updateRecurrence(props.issueId, payload)
    } else {
      if (!payload.next_run) payload.next_run = new Date(Date.now()+86400000).toISOString()
      await recApi.createRecurrence(props.issueId, payload)
    }
    showForm.value = false; load()
  } catch (e: any) { alert(e.response?.data?.message || 'Failed') }
  finally { saving.value = false }
}

async function toggle() {
  if (!rule.value) return
  try {
    await recApi.updateRecurrence(props.issueId, { is_active: !rule.value.is_active })
    await load()
  } catch (e: any) { alert(e.response?.data?.message || 'Failed to toggle recurrence') }
}

async function remove() {
  if (!confirm(t('recurrence.removeConfirm'))) return
  try {
    await recApi.deleteRecurrence(props.issueId)
    rule.value = null
  } catch (e: any) { alert(e.response?.data?.message || 'Failed to remove recurrence') }
}

function formatDate(d: string) { return new Date(d).toLocaleDateString() }
</script>
