<template>
  <div class="min-h-screen bg-gray-50 p-6 space-y-5" data-testid="intake-page">
    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('intakeHub.title') }}</h1>
        <p class="text-sm text-gray-500 mt-1">{{ projectName }} · {{ t('intakeHub.subtitle') }}</p>
      </div>
      <div class="flex items-center gap-2">
        <div class="inline-flex rounded-lg border border-gray-300 overflow-hidden" data-testid="in-range">
          <button
            v-for="d in [7, 14, 30, 90]"
            :key="d"
            :class="['px-3 py-1.5 text-sm', days === d ? 'bg-indigo-600 text-white' : 'bg-white text-gray-700 hover:bg-gray-50']"
            @click="setDays(d)"
          >{{ t('intakeHub.lastDays', { n: String(d) }) }}</button>
        </div>
        <button :class="btnClass" data-testid="in-refresh" @click="reloadAll">{{ t('intakeHub.refresh') }}</button>
      </div>
    </div>

    <!-- Metrics -->
    <div v-if="metrics" class="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-8 gap-3" data-testid="in-metrics">
      <div v-for="k in kpis" :key="k.key" class="bg-white rounded-lg border border-gray-200 p-3" :data-testid="'in-kpi-' + k.key">
        <div class="text-xs text-gray-500">{{ k.label }}</div>
        <div :class="['text-xl font-semibold mt-1', k.warn ? 'text-red-600' : 'text-gray-900']">{{ k.value }}</div>
      </div>
    </div>

    <div v-if="metrics" class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <div class="bg-white rounded-lg border border-gray-200 p-4 lg:col-span-2">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-sm font-semibold text-gray-800">{{ t('intakeHub.trend') }}</h3>
          <div class="flex items-center gap-3 text-xs text-gray-500">
            <span class="flex items-center gap-1"><span class="w-2.5 h-2.5 rounded-sm bg-indigo-400 inline-block"></span>{{ t('intakeHub.received') }}</span>
            <span class="flex items-center gap-1"><span class="w-2.5 h-2.5 rounded-sm bg-emerald-400 inline-block"></span>{{ t('intakeHub.triaged') }}</span>
          </div>
        </div>
        <div class="flex items-end gap-px h-28" data-testid="in-trend">
          <div
            v-for="p in metrics.trend"
            :key="p.date"
            class="flex-1 flex items-end gap-px h-full"
            :title="`${p.date}  ${t('intakeHub.received')} ${p.received} / ${t('intakeHub.triaged')} ${p.triaged}`"
          >
            <div class="flex-1 bg-indigo-400 rounded-t-sm" :style="{ height: barH(p.received) }"></div>
            <div class="flex-1 bg-emerald-400 rounded-t-sm" :style="{ height: barH(p.triaged) }"></div>
          </div>
        </div>
        <div class="flex justify-between text-[10px] text-gray-400 mt-1">
          <span>{{ metrics.trend[0]?.date }}</span><span>{{ metrics.trend[metrics.trend.length - 1]?.date }}</span>
        </div>
      </div>
      <div class="bg-white rounded-lg border border-gray-200 p-4" data-testid="in-by-source">
        <h3 class="text-sm font-semibold text-gray-800 mb-3">{{ t('intakeHub.bySource') }}</h3>
        <div v-if="!metrics.by_source.length" class="text-xs text-gray-400">{{ t('intakeHub.noData') }}</div>
        <div v-for="s in metrics.by_source" :key="s.source" class="mb-2">
          <div class="flex justify-between text-xs text-gray-600">
            <span>{{ t('intakeHub.source.' + s.source) }}</span>
            <span>{{ s.accepted }} / {{ s.received }}</span>
          </div>
          <div class="h-2 bg-gray-100 rounded mt-1 overflow-hidden">
            <div class="h-full bg-indigo-500" :style="{ width: (s.received ? (s.accepted / s.received) * 100 : 0) + '%' }"></div>
          </div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-5 gap-4">
      <!-- Queue -->
      <div class="xl:col-span-3 bg-white rounded-lg border border-gray-200">
        <div class="flex items-center gap-1 border-b border-gray-200 px-3 pt-2 overflow-x-auto" data-testid="in-tabs">
          <button
            v-for="s in statuses"
            :key="s"
            :data-testid="'in-tab-' + s"
            :class="['px-3 py-2 text-sm border-b-2 -mb-px whitespace-nowrap', status === s ? 'border-indigo-600 text-indigo-700 font-medium' : 'border-transparent text-gray-500 hover:text-gray-700']"
            @click="setStatus(s)"
          >
            {{ t('intakeHub.status.' + s) }}
            <span class="ml-1 text-xs px-1.5 rounded-full bg-gray-100 text-gray-600">{{ counts[s] ?? 0 }}</span>
          </button>
        </div>
        <div class="flex items-center gap-2 p-3 border-b border-gray-100">
          <input
            v-model="search"
            data-testid="in-search"
            :placeholder="t('intakeHub.searchPh')"
            class="px-3 py-1.5 text-sm border border-gray-300 rounded-lg"
            style="width: 260px"
            @input="onSearch"
          />
          <select v-model="source" data-testid="in-source" class="px-3 py-1.5 text-sm border border-gray-300 rounded-lg" style="width: 140px" @change="loadList()">
            <option value="">{{ t('intakeHub.allSources') }}</option>
            <option v-for="s in sources" :key="s" :value="s">{{ t('intakeHub.source.' + s) }}</option>
          </select>
          <span class="ml-auto text-xs text-gray-400">{{ t('intakeHub.slaHint', { h: String(slaHours) }) }}</span>
        </div>
        <div v-if="loading" class="p-8 text-center text-sm text-gray-400">{{ t('intake.loading') }}</div>
        <div v-else-if="!items.length" class="p-10 text-center" data-testid="in-empty">
          <p class="text-sm text-gray-500">{{ t('intakeHub.empty') }}</p>
          <p class="text-xs text-gray-400 mt-1">{{ t('intake.emptyHint') }}</p>
        </div>
        <ul v-else class="divide-y divide-gray-100" data-testid="in-list">
          <li
            v-for="it in items"
            :key="it.id"
            :data-testid="'in-item-' + it.id"
            :class="['px-4 py-3 cursor-pointer', selected?.id === it.id ? 'bg-indigo-50' : 'hover:bg-gray-50']"
            @click="select(it)"
          >
            <div class="flex items-center gap-2">
              <span class="text-xs text-gray-400">#{{ it.sequence_id }}</span>
              <span class="text-sm font-medium text-gray-900 truncate flex-1">{{ it.name }}</span>
              <span :class="['text-[11px] px-1.5 py-0.5 rounded', sourceClass(it.source)]">{{ t('intakeHub.source.' + it.source) }}</span>
              <span v-if="it.sla_overdue" class="text-[11px] px-1.5 py-0.5 rounded bg-red-100 text-red-700" data-testid="in-overdue">{{ t('intakeHub.overdue') }}</span>
            </div>
            <div class="flex items-center gap-3 mt-1 text-xs text-gray-500">
              <span>{{ it.submitter || t('intakeHub.anonymous') }}<template v-if="it.email"> · {{ it.email }}</template></span>
              <span>{{ ageLabel(it.age_hours) }}</span>
              <span v-if="it.status === 'snoozed' && it.snoozed_until">{{ t('intakeHub.snoozedUntil', { d: fmtDateTime(it.snoozed_until) }) }}</span>
              <span v-if="it.status === 'duplicate' && it.duplicate_sequence_id">{{ t('intakeHub.duplicateOf', { n: String(it.duplicate_sequence_id) }) }}</span>
              <span v-if="it.triaged_by_name && it.status !== 'snoozed'">{{ t('intakeHub.triagedBy', { u: it.triaged_by_name }) }}</span>
            </div>
          </li>
        </ul>
        <div v-if="total > items.length" class="p-3 text-center">
          <button :class="btnClass" @click="loadMore">{{ t('intakeHub.loadMore') }}</button>
        </div>
      </div>

      <!-- Detail -->
      <div class="xl:col-span-2 space-y-4">
        <div class="bg-white rounded-lg border border-gray-200 p-4" data-testid="in-detail">
          <div v-if="!selected" class="text-sm text-gray-400 py-10 text-center">{{ t('intakeHub.selectHint') }}</div>
          <template v-else>
            <div class="flex items-start justify-between gap-2">
              <div>
                <div class="text-xs text-gray-400">#{{ selected.sequence_id }} · {{ t('intakeHub.status.' + selected.status) }}</div>
                <h3 class="text-base font-semibold text-gray-900 mt-0.5">{{ selected.name }}</h3>
              </div>
              <button :class="btnClass" :disabled="analyzing" data-testid="in-ai" @click="analyze">
                {{ analyzing ? t('intake.analyzing') : t('intakeHub.aiAnalyze') }}
              </button>
            </div>
            <dl class="grid grid-cols-3 gap-y-1 text-xs mt-3">
              <dt class="text-gray-500">{{ t('intakeHub.submitter') }}</dt>
              <dd class="col-span-2 text-gray-800">{{ selected.submitter || t('intakeHub.anonymous') }}<template v-if="selected.email"> &lt;{{ selected.email }}&gt;</template></dd>
              <dt class="text-gray-500">{{ t('intakeHub.channel') }}</dt>
              <dd class="col-span-2 text-gray-800">{{ t('intakeHub.source.' + selected.source) }}</dd>
              <dt class="text-gray-500">{{ t('intakeHub.receivedAt') }}</dt>
              <dd class="col-span-2 text-gray-800">{{ fmtDateTime(selected.created_at) }}</dd>
              <dt class="text-gray-500">{{ t('intakeHub.priority') }}</dt>
              <dd class="col-span-2 text-gray-800">{{ selected.priority }}</dd>
              <template v-if="selected.note">
                <dt class="text-gray-500">{{ t('intakeHub.reason') }}</dt>
                <dd class="col-span-2 text-gray-800" data-testid="in-note">{{ selected.note }}</dd>
              </template>
            </dl>
            <p class="text-sm text-gray-700 whitespace-pre-wrap mt-3 p-3 bg-gray-50 rounded max-h-48 overflow-auto" data-testid="in-desc">{{ plainText(selected.description_html) || t('intakeHub.noDescription') }}</p>

            <div v-if="ai" class="mt-3 p-3 bg-indigo-50 rounded text-xs space-y-1" data-testid="in-ai-result">
              <div class="flex items-center gap-2">
                <span class="font-medium text-indigo-700">{{ t('intake.aiLabel') }}</span>
                <span>{{ ai.suggested_type }}</span>
                <span class="px-1 rounded bg-amber-500 text-white">{{ ai.suggested_priority }}</span>
              </div>
              <div class="text-gray-700">{{ ai.summary }}</div>
              <div v-if="ai.duplicate_ids?.length" class="text-amber-700">
                {{ t('intakeHub.aiDuplicates') }}
                <button
                  v-for="id in ai.duplicate_ids"
                  :key="id"
                  class="ml-1 underline"
                  @click="pickDuplicateById(id)"
                >#{{ id }}</button>
              </div>
            </div>

            <!-- Actions for open items -->
            <div v-if="selected.status === 'pending' || selected.status === 'snoozed'" class="mt-4 space-y-3">
              <div class="border border-gray-200 rounded-lg p-3">
                <div class="text-xs font-medium text-gray-700 mb-2">{{ t('intakeHub.acceptTitle') }}</div>
                <div class="flex gap-2">
                  <select v-model="acceptState" data-testid="in-accept-state" class="flex-1 px-2 py-1.5 text-sm border border-gray-300 rounded-lg">
                    <option :value="null">{{ t('intakeHub.keepState') }}</option>
                    <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
                  </select>
                  <select v-model="acceptAssignee" data-testid="in-accept-assignee" class="flex-1 px-2 py-1.5 text-sm border border-gray-300 rounded-lg">
                    <option :value="null">{{ t('intakeHub.noAssignee') }}</option>
                    <option v-for="m in members" :key="m.user_id" :value="m.user_id">{{ memberName(m) }}</option>
                  </select>
                </div>
                <button class="mt-2 w-full px-3 py-1.5 text-sm rounded-lg bg-green-600 text-white hover:bg-green-700 disabled:opacity-50" :disabled="acting" data-testid="in-accept" @click="doAccept">{{ t('intake.accept') }}</button>
              </div>

              <div class="border border-gray-200 rounded-lg p-3">
                <div class="text-xs font-medium text-gray-700 mb-2">{{ t('intakeHub.snoozeTitle') }}</div>
                <div class="flex gap-2">
                  <button v-for="o in snoozeOptions" :key="o.h" :class="btnClass + ' flex-1'" :disabled="acting" :data-testid="'in-snooze-' + o.h" @click="doSnooze(o.h)">{{ o.label }}</button>
                </div>
              </div>

              <div class="border border-gray-200 rounded-lg p-3">
                <div class="text-xs font-medium text-gray-700 mb-2">{{ t('intakeHub.duplicateTitle') }}</div>
                <input
                  v-model="dupQuery"
                  data-testid="in-dup-search"
                  :placeholder="t('intakeHub.dupSearchPh')"
                  class="w-full px-2 py-1.5 text-sm border border-gray-300 rounded-lg"
                  @input="onDupSearch"
                />
                <ul v-if="dupResults.length && !dupTarget" class="mt-1 border border-gray-200 rounded max-h-40 overflow-auto" data-testid="in-dup-results">
                  <li v-for="r in dupResults" :key="r.id" class="px-2 py-1 text-sm hover:bg-gray-50 cursor-pointer" @click="dupTarget = r">#{{ r.sequence_id }} {{ r.name }}</li>
                </ul>
                <div v-if="dupTarget" class="mt-2 flex items-center gap-2 text-sm" data-testid="in-dup-target">
                  <span class="flex-1 truncate">→ #{{ dupTarget.sequence_id }} {{ dupTarget.name }}</span>
                  <button class="text-xs text-gray-500 underline" @click="dupTarget = null">{{ t('intakeHub.change') }}</button>
                </div>
                <button class="mt-2 w-full px-3 py-1.5 text-sm rounded-lg bg-amber-500 text-white hover:bg-amber-600 disabled:opacity-50" :disabled="acting || !dupTarget" data-testid="in-duplicate" @click="doDuplicate">{{ t('intakeHub.mergeDuplicate') }}</button>
              </div>

              <div class="border border-gray-200 rounded-lg p-3">
                <div class="text-xs font-medium text-gray-700 mb-2">{{ t('intakeHub.rejectTitle') }}</div>
                <input v-model="rejectReason" data-testid="in-reject-reason" :placeholder="t('intakeHub.reasonPh')" class="w-full px-2 py-1.5 text-sm border border-gray-300 rounded-lg" />
                <button class="mt-2 w-full px-3 py-1.5 text-sm rounded-lg bg-red-500 text-white hover:bg-red-600 disabled:opacity-50" :disabled="acting" data-testid="in-reject" @click="doReject">{{ t('intake.reject') }}</button>
              </div>
            </div>

            <div v-else class="mt-4 flex items-center gap-2">
              <router-link
                v-if="selected.status === 'accepted'"
                :to="`/workspace/${slug}/project/${projectId}/issues/${selected.id}`"
                class="px-3 py-1.5 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700"
                data-testid="in-open-issue"
              >{{ t('intakeHub.openIssue') }}</router-link>
              <button v-else :class="btnClass" :disabled="acting" data-testid="in-reopen" @click="doReopen">{{ t('intakeHub.reopen') }}</button>
            </div>
          </template>
        </div>

        <!-- Channels -->
        <div v-if="settings" class="bg-white rounded-lg border border-gray-200 p-4 space-y-3" data-testid="in-channels">
          <h3 class="text-sm font-semibold text-gray-800">{{ t('intakeHub.channels') }}</h3>
          <div v-for="ch in channels" :key="ch.key" class="text-xs">
            <div class="flex items-center justify-between">
              <label class="flex items-center gap-2 font-medium text-gray-700">
                <input type="checkbox" :checked="ch.enabled" :data-testid="'in-toggle-' + ch.key" @change="toggleChannel(ch.key, ($event.target as HTMLInputElement).checked)" />
                {{ t('intakeHub.source.' + ch.key) }}
              </label>
              <button class="text-indigo-600 hover:underline" :data-testid="'in-copy-' + ch.key" @click="copy(ch.url)">{{ t('intakeHub.copy') }}</button>
            </div>
            <div class="mt-1 font-mono text-[11px] text-gray-500 break-all bg-gray-50 rounded px-2 py-1" :data-testid="'in-url-' + ch.key">{{ ch.url }}</div>
          </div>
          <div class="flex items-center gap-2 text-xs pt-1">
            <span class="text-gray-600">{{ t('intakeHub.slaLabel') }}</span>
            <input v-model.number="slaInput" type="number" min="1" max="720" data-testid="in-sla-input" class="px-2 py-1 border border-gray-300 rounded" style="width: 80px" />
            <button :class="btnClass" data-testid="in-sla-save" @click="saveSLA">{{ t('intakeHub.save') }}</button>
            <button class="ml-auto text-red-600 hover:underline" data-testid="in-rotate" @click="rotateToken">{{ t('intakeHub.rotateToken') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { intakeApi, type IntakeItem, type IntakeMetrics, type IntakeSettings, type IntakeStatus, type IntakeSource } from '@/api/intake'
import * as projectApi from '@/api/project'
import { listStates } from '@/api/project-settings'
import { listIssues } from '@/api/issue'
import { useToast } from '@/composables/useToast'
import { useI18n } from '@/composables/useI18n'

const { t, locale } = useI18n()
const toast = useToast()
const route = useRoute()

const btnClass = 'px-3 py-1.5 text-sm rounded-lg border border-gray-300 bg-white text-gray-700 hover:bg-gray-50 disabled:opacity-50'
const statuses: IntakeStatus[] = ['pending', 'snoozed', 'accepted', 'rejected', 'duplicate']
const sources: IntakeSource[] = ['form', 'webhook', 'email']

const slug = computed(() => route.params.slug as string)
const projectId = computed(() => parseInt(route.params.id as string, 10))
const projectName = ref('')
const workspaceId = ref(0)

const days = ref(30)
const metrics = ref<IntakeMetrics | null>(null)
const settings = ref<IntakeSettings | null>(null)
const slaInput = ref(48)

const status = ref<IntakeStatus>('pending')
const source = ref('')
const search = ref('')
const items = ref<IntakeItem[]>([])
const total = ref(0)
const counts = ref<Record<string, number>>({})
const slaHours = ref(48)
const loading = ref(false)

const selected = ref<IntakeItem | null>(null)
const ai = ref<any>(null)
const analyzing = ref(false)
const acting = ref(false)
const states = ref<{ id: number; name: string }[]>([])
const members = ref<any[]>([])
const acceptState = ref<number | null>(null)
const acceptAssignee = ref<number | null>(null)
const rejectReason = ref('')
const dupQuery = ref('')
const dupResults = ref<{ id: number; sequence_id: number; name: string }[]>([])
const dupTarget = ref<{ id: number; sequence_id: number; name: string } | null>(null)

const snoozeOptions = computed(() => [
  { h: 24, label: t('intakeHub.snooze1d') },
  { h: 72, label: t('intakeHub.snooze3d') },
  { h: 168, label: t('intakeHub.snooze1w') },
])

function fmtHours(h: number | null | undefined) {
  if (h == null) return '—'
  return h >= 48 ? t('intakeHub.daysShort', { n: (h / 24).toFixed(1) }) : t('intakeHub.hoursShort', { n: h.toFixed(1) })
}
function pct(v: number | null | undefined) {
  return v == null ? '—' : `${v}%`
}

const kpis = computed(() => {
  const m = metrics.value
  if (!m) return []
  return [
    { key: 'received', label: t('intakeHub.kpi.received'), value: String(m.received) },
    { key: 'accepted', label: t('intakeHub.kpi.accepted'), value: String(m.accepted) },
    { key: 'acceptance', label: t('intakeHub.kpi.acceptance'), value: pct(m.acceptance_rate) },
    { key: 'avg', label: t('intakeHub.kpi.avgTriage'), value: fmtHours(m.avg_triage_hours) },
    { key: 'median', label: t('intakeHub.kpi.medianTriage'), value: fmtHours(m.median_triage_hours) },
    { key: 'sla', label: t('intakeHub.kpi.slaMet'), value: pct(m.sla_met_rate) },
    { key: 'pending', label: t('intakeHub.kpi.pending'), value: String(m.pending_now) },
    { key: 'overdue', label: t('intakeHub.kpi.overdue', { h: String(m.sla_hours) }), value: String(m.overdue_now), warn: m.overdue_now > 0 },
  ]
})

const maxTrend = computed(() => Math.max(1, ...(metrics.value?.trend || []).map((p) => Math.max(p.received, p.triaged))))
function barH(v: number) {
  return v ? `${Math.max(4, (v / maxTrend.value) * 100)}%` : '0'
}

const channels = computed(() => {
  const s = settings.value
  if (!s) return []
  const origin = window.location.origin
  const apiBase = `${origin}/api/v1`
  return [
    { key: 'form' as const, enabled: s.form_enabled, url: `${origin}/intake/${projectId.value}` },
    { key: 'webhook' as const, enabled: s.webhook_enabled, url: `${apiBase}/intake-channels/${s.token}/webhook` },
    { key: 'email' as const, enabled: s.email_enabled, url: `${apiBase}/intake-channels/${s.token}/email` },
  ]
})

function sourceClass(s: string) {
  return s === 'webhook' ? 'bg-purple-100 text-purple-700' : s === 'email' ? 'bg-sky-100 text-sky-700' : 'bg-gray-100 text-gray-700'
}
function ageLabel(h: number) {
  return t('intakeHub.age', { t: fmtHours(h) })
}
function fmtDateTime(d: string) {
  return new Date(d).toLocaleString(locale.value === 'zh-CN' ? 'zh-CN' : 'en-US', { dateStyle: 'short', timeStyle: 'short' })
}
function plainText(html: string) {
  if (!html) return ''
  const doc = new DOMParser().parseFromString(html.replace(/<br\s*\/?>/gi, '\n'), 'text/html')
  return (doc.body.textContent || '').trim()
}
function memberName(m: any) {
  return m.display_name || m.user?.display_name || m.username || m.email || `#${m.user_id}`
}
function errMsg(e: any) {
  return e?.response?.data?.message || e?.message || t('intake.actionFailed')
}

async function loadMetrics() {
  try {
    metrics.value = await intakeApi.metrics(projectId.value, days.value)
  } catch (e) {
    toast.error(errMsg(e))
  }
}

async function loadList(append = false) {
  if (!append) loading.value = true
  try {
    const res = await intakeApi.list(projectId.value, {
      status: status.value,
      source: source.value || undefined,
      q: search.value || undefined,
      offset: append ? items.value.length : 0,
    })
    items.value = append ? [...items.value, ...res.items] : res.items
    total.value = res.total
    counts.value = res.counts
    slaHours.value = res.sla_hours
    if (!append) {
      const keep = selected.value && items.value.find((i) => i.id === selected.value!.id)
      selectItem(keep || items.value[0] || null)
    }
  } catch (e) {
    toast.error(errMsg(e))
  } finally {
    loading.value = false
  }
}
function loadMore() {
  loadList(true)
}

async function loadSettings() {
  try {
    settings.value = await intakeApi.getSettings(projectId.value)
    slaInput.value = settings.value.sla_hours
  } catch {
    settings.value = null
  }
}

function reloadAll() {
  loadMetrics()
  loadList()
}
function setDays(d: number) {
  days.value = d
  loadMetrics()
}
function setStatus(s: IntakeStatus) {
  status.value = s
  loadList()
}

let searchTimer: ReturnType<typeof setTimeout> | undefined
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadList(), 300)
}

function selectItem(it: IntakeItem | null) {
  if (selected.value?.id !== it?.id) {
    ai.value = null
    acceptState.value = null
    acceptAssignee.value = null
    rejectReason.value = ''
    dupQuery.value = ''
    dupResults.value = []
    dupTarget.value = null
  }
  selected.value = it
}
function select(it: IntakeItem) {
  selectItem(it)
}

async function analyze() {
  if (!selected.value) return
  analyzing.value = true
  try {
    ai.value = await intakeApi.aiAnalyze(projectId.value, selected.value.id)
  } catch (e) {
    toast.error(errMsg(e))
  } finally {
    analyzing.value = false
  }
}

async function act(payload: Parameters<typeof intakeApi.triage>[2], okKey: string) {
  if (!selected.value) return
  acting.value = true
  try {
    await intakeApi.triage(projectId.value, selected.value.id, payload)
    toast.success(t(okKey))
    selected.value = null
    await Promise.all([loadList(), loadMetrics()])
  } catch (e) {
    toast.error(errMsg(e))
  } finally {
    acting.value = false
  }
}
function doAccept() {
  act({ action: 'accept', state_id: acceptState.value ?? undefined, assignee_id: acceptAssignee.value ?? undefined }, 'intakeHub.toast.accepted')
}
function doReject() {
  act({ action: 'reject', reason: rejectReason.value.trim() || undefined }, 'intakeHub.toast.rejected')
}
function doSnooze(h: number) {
  act({ action: 'snooze', snooze_hours: h }, 'intakeHub.toast.snoozed')
}
function doDuplicate() {
  if (!dupTarget.value) return
  act({ action: 'duplicate', duplicate_of: dupTarget.value.id }, 'intakeHub.toast.merged')
}
function doReopen() {
  act({ action: 'reopen' }, 'intakeHub.toast.reopened')
}

let dupTimer: ReturnType<typeof setTimeout> | undefined
function onDupSearch() {
  dupTarget.value = null
  clearTimeout(dupTimer)
  const q = dupQuery.value.trim()
  if (!q) {
    dupResults.value = []
    return
  }
  dupTimer = setTimeout(async () => {
    try {
      const res = await listIssues(projectId.value, workspaceId.value, { search: q, limit: 8 })
      dupResults.value = res.items
        .filter((i: any) => i.id !== selected.value?.id)
        .map((i: any) => ({ id: i.id, sequence_id: i.sequence_id, name: i.name }))
    } catch {
      dupResults.value = []
    }
  }, 250)
}
async function pickDuplicateById(id: number) {
  try {
    const res = await listIssues(projectId.value, workspaceId.value, { limit: 200 })
    const hit = res.items.find((i: any) => i.id === id)
    if (hit) {
      dupTarget.value = { id: hit.id, sequence_id: hit.sequence_id, name: hit.name }
      dupQuery.value = hit.name
    } else {
      toast.error(t('intakeHub.dupNotEligible'))
    }
  } catch (e) {
    toast.error(errMsg(e))
  }
}

async function updateSettings(data: Parameters<typeof intakeApi.updateSettings>[1], okKey: string) {
  try {
    settings.value = await intakeApi.updateSettings(projectId.value, data)
    slaInput.value = settings.value.sla_hours
    toast.success(t(okKey))
  } catch (e) {
    toast.error(errMsg(e))
    await loadSettings()
  }
}
function toggleChannel(key: IntakeSource, on: boolean) {
  updateSettings({ [`${key}_enabled`]: on } as any, 'intakeHub.toast.saved')
}
async function saveSLA() {
  const v = Number(slaInput.value)
  if (!Number.isInteger(v) || v < 1 || v > 720) {
    toast.error(t('intakeHub.slaInvalid'))
    return
  }
  await updateSettings({ sla_hours: v }, 'intakeHub.toast.saved')
  reloadAll()
}
function rotateToken() {
  if (!confirm(t('intakeHub.rotateConfirm'))) return
  updateSettings({ rotate_token: true }, 'intakeHub.toast.rotated')
}
async function copy(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    toast.success(t('intakeHub.toast.copied'))
  } catch {
    toast.error(text)
  }
}

onMounted(async () => {
  try {
    const p = await projectApi.getProject(projectId.value)
    projectName.value = p.name
    workspaceId.value = p.workspace_id
  } catch {
    /* header only */
  }
  reloadAll()
  loadSettings()
  listStates(projectId.value).then((s) => (states.value = s as any)).catch(() => {})
  projectApi.listProjectMembers(projectId.value).then((m) => (members.value = m)).catch(() => {})
})
</script>
