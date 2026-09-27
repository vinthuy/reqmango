<template>
  <section class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 space-y-5" data-testid="loop-panel">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h2 class="text-sm font-semibold text-gray-800 dark:text-gray-200">{{ t('workspaceAnalytics.loop.title') }}</h2>
        <p class="text-xs text-gray-500 mt-0.5">{{ t('workspaceAnalytics.loop.subtitle', { n: String(days) }) }}</p>
      </div>
      <span
        v-if="data && data.bottleneck"
        class="inline-flex items-center gap-1 rounded-full bg-amber-50 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300 px-2.5 py-1 text-xs font-medium"
        data-testid="loop-bottleneck"
      >⏳ {{ t('workspaceAnalytics.loop.bottleneck', { stage: stageLabel(data.bottleneck) }) }}</span>
    </div>

    <div v-if="error" class="rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700 flex items-center justify-between" data-testid="loop-error">
      <span>{{ t('workspaceAnalytics.loop.loadFailed') }}</span>
      <button type="button" class="text-xs underline" @click="load">{{ t('workspaceAnalytics.retry') }}</button>
    </div>
    <div v-else-if="!data" class="h-40 rounded bg-gray-100 dark:bg-gray-700 animate-pulse" />

    <div v-else-if="data.summary.received === 0" class="rounded border border-dashed border-gray-300 p-8 text-center text-sm text-gray-500" data-testid="loop-empty">
      {{ t('workspaceAnalytics.loop.empty') }}
    </div>

    <div v-else :class="['space-y-5 transition-opacity', loading ? 'opacity-60' : '']">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3" data-testid="loop-kpis">
        <div v-for="k in kpis" :key="k.key" class="rounded-md bg-gray-50 dark:bg-gray-900/40 p-3" :title="k.hint" :data-testid="`loop-kpi-${k.key}`">
          <div class="text-xs text-gray-500">{{ k.label }}</div>
          <div :class="['text-xl font-semibold mt-1 tabular-nums', k.tone]">{{ k.value }}</div>
          <div v-if="k.sub" class="text-xs text-gray-400 mt-0.5">{{ k.sub }}</div>
        </div>
      </div>

      <div>
        <h3 class="text-xs font-semibold text-gray-600 dark:text-gray-300 mb-2">{{ t('workspaceAnalytics.loop.funnelTitle') }}</h3>
        <div class="space-y-1.5" data-testid="loop-funnel">
          <div v-for="(f, i) in data.funnel" :key="f.key" class="flex items-center gap-3 text-xs" :data-testid="`loop-funnel-${f.key}`">
            <span class="w-20 shrink-0 text-gray-600 dark:text-gray-300">{{ t(`workspaceAnalytics.loop.funnel.${f.key}`) }}</span>
            <div class="flex-1 h-5 rounded bg-gray-100 dark:bg-gray-700 overflow-hidden">
              <div class="h-5 rounded" :style="{ width: `${Math.max(f.rate * 100, f.count ? 0.8 : 0)}%`, background: FUNNEL_COLORS[i] }" />
            </div>
            <span class="w-14 text-right tabular-nums text-gray-800 dark:text-gray-100 font-medium">{{ f.count }}</span>
            <span class="w-12 text-right tabular-nums text-gray-500">{{ pct(f.rate) }}</span>
            <span class="w-24 text-right tabular-nums text-gray-400" :title="t('workspaceAnalytics.loop.stepRateHint')">
              <template v-if="i > 0">{{ f.key === 'pr_linked' ? '↳' : '→' }} {{ pct(f.step_rate) }}</template>
            </span>
          </div>
        </div>
      </div>

      <div class="overflow-x-auto">
        <h3 class="text-xs font-semibold text-gray-600 dark:text-gray-300 mb-2">{{ t('workspaceAnalytics.loop.stagesTitle') }}</h3>
        <table class="w-full text-sm" data-testid="loop-stages">
          <thead>
            <tr class="text-left text-xs text-gray-500 border-b border-gray-100 dark:border-gray-700">
              <th class="py-2 pr-3 font-medium">{{ t('workspaceAnalytics.loop.col.stage') }}</th>
              <th class="py-2 px-3 font-medium text-right">{{ t('workspaceAnalytics.loop.col.median') }}</th>
              <th class="py-2 px-3 font-medium text-right">P85</th>
              <th class="py-2 px-3 font-medium text-right">{{ t('workspaceAnalytics.loop.col.avg') }}</th>
              <th class="py-2 px-3 font-medium text-right">{{ t('workspaceAnalytics.loop.col.samples') }}</th>
              <th class="py-2 px-3 font-medium text-right" :title="t('workspaceAnalytics.loop.wipHint')">{{ t('workspaceAnalytics.loop.col.wip') }}</th>
              <th class="py-2 pl-3 font-medium text-right">{{ t('workspaceAnalytics.loop.col.wipAge') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="s in data.stages"
              :key="s.key"
              :class="['border-b border-gray-50 dark:border-gray-700/50', s.key === data.bottleneck ? 'bg-amber-50/60 dark:bg-amber-900/20' : '']"
              :data-testid="`loop-stage-${s.key}`"
            >
              <td class="py-2 pr-3">
                <div class="text-gray-800 dark:text-gray-100">{{ stageLabel(s.key) }}</div>
                <div class="text-xs text-gray-400">{{ t(`workspaceAnalytics.loop.stageDesc.${s.key}`) }}</div>
              </td>
              <td class="py-2 px-3 text-right tabular-nums font-medium">{{ dur(s.median_hours) }}</td>
              <td class="py-2 px-3 text-right tabular-nums">{{ dur(s.p85_hours) }}</td>
              <td class="py-2 px-3 text-right tabular-nums">{{ dur(s.avg_hours) }}</td>
              <td class="py-2 px-3 text-right tabular-nums text-gray-500">{{ s.samples }}</td>
              <td class="py-2 px-3 text-right tabular-nums">{{ s.wip }}</td>
              <td class="py-2 pl-3 text-right tabular-nums text-gray-500">{{ dur(s.wip_median_hours) }}</td>
            </tr>
            <tr v-for="d in extraDurations" :key="d.key" class="border-b border-gray-50 dark:border-gray-700/50 text-gray-600 dark:text-gray-300" :data-testid="`loop-stage-${d.key}`">
              <td class="py-2 pr-3">
                <div>{{ t(`workspaceAnalytics.loop.extra.${d.key}`) }}</div>
                <div class="text-xs text-gray-400">{{ t(`workspaceAnalytics.loop.extraDesc.${d.key}`) }}</div>
              </td>
              <td class="py-2 px-3 text-right tabular-nums font-medium">{{ dur(d.median_hours) }}</td>
              <td class="py-2 px-3 text-right tabular-nums">{{ dur(d.p85_hours) }}</td>
              <td class="py-2 px-3 text-right tabular-nums">{{ dur(d.avg_hours) }}</td>
              <td class="py-2 px-3 text-right tabular-nums text-gray-500">{{ d.samples }}</td>
              <td class="py-2 px-3 text-right text-gray-300">—</td>
              <td class="py-2 pl-3 text-right text-gray-300">—</td>
            </tr>
          </tbody>
        </table>
        <p class="text-xs text-gray-400 mt-2">{{ t('workspaceAnalytics.loop.note') }}</p>
      </div>

      <div>
        <h3 class="text-xs font-semibold text-gray-600 dark:text-gray-300 mb-2">{{ t('workspaceAnalytics.loop.stalledTitle') }}</h3>
        <div v-if="data.stalled.length === 0" class="text-sm text-gray-500" data-testid="loop-stalled-empty">{{ t('workspaceAnalytics.loop.stalledEmpty') }}</div>
        <ul v-else class="divide-y divide-gray-100 dark:divide-gray-700" data-testid="loop-stalled">
          <li v-for="item in data.stalled" :key="item.issue_id" class="flex items-center gap-3 py-2 text-sm" data-testid="loop-stalled-row">
            <span :class="['shrink-0 rounded px-1.5 py-0.5 text-xs', STAGE_BADGE[item.stage]]">{{ stageLabel(item.stage) }}</span>
            <router-link
              :to="`/workspace/${slug}/project/${item.project_id}/issues/${item.issue_id}`"
              class="min-w-0 flex-1 truncate text-gray-800 dark:text-gray-100 hover:text-indigo-600"
              :data-testid="`loop-stalled-link-${item.issue_id}`"
            >
              <span class="text-xs text-gray-400 mr-1.5">{{ item.key }}</span>{{ item.name }}
            </router-link>
            <span class="shrink-0 text-xs tabular-nums" :class="item.age_hours >= 24 * 14 ? 'text-red-600' : 'text-gray-500'">
              {{ t('workspaceAnalytics.loop.waiting', { d: dur(item.age_hours) }) }}
            </span>
          </li>
        </ul>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { getDeliveryLoop, type DeliveryLoop, type LoopStageKey } from '@/api/workspace-analytics'

const props = defineProps<{ slug: string; days: number; projectId: number; refreshKey: number }>()
const { t } = useI18n()

const data = ref<DeliveryLoop | null>(null)
const loading = ref(false)
const error = ref(false)
let seq = 0

const FUNNEL_COLORS = ['#a5b4fc', '#818cf8', '#f59e0b', '#38bdf8', '#34d399', '#10b981']
const STAGE_BADGE: Record<LoopStageKey, string> = {
  triage: 'bg-purple-50 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
  queue: 'bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  develop: 'bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
  release: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300',
}

function pct(v: number) {
  return `${Math.round((v || 0) * 100)}%`
}

function dur(hours: number | null | undefined) {
  if (hours == null) return '—'
  if (hours < 1) return t('workspaceAnalytics.loop.minutes', { n: String(Math.max(1, Math.round(hours * 60))) })
  if (hours < 48) return t('workspaceAnalytics.loop.hours', { n: hours.toFixed(1) })
  return t('workspaceAnalytics.kpi.days', { n: (hours / 24).toFixed(1) })
}

function stageLabel(key: string) {
  return t(`workspaceAnalytics.loop.stage.${key}`)
}

async function load() {
  const my = ++seq
  loading.value = true
  error.value = false
  try {
    const res = await getDeliveryLoop(props.slug, { days: props.days, project_id: props.projectId || undefined })
    if (my === seq) data.value = res
  } catch {
    if (my === seq) error.value = true
  } finally {
    if (my === seq) loading.value = false
  }
}

const kpis = computed(() => {
  const d = data.value
  if (!d) return []
  const s = d.summary
  return [
    { key: 'received', label: t('workspaceAnalytics.loop.kpi.received'), value: s.received, sub: t('workspaceAnalytics.loop.kpi.fromIntake', { n: String(s.from_intake), r: String(s.rejected) }), tone: 'text-gray-900 dark:text-gray-100', hint: '' },
    { key: 'ship', label: t('workspaceAnalytics.loop.kpi.shipRate'), value: pct(s.ship_rate), sub: t('workspaceAnalytics.loop.kpi.shipped', { n: String(s.shipped), total: String(s.accepted) }), tone: 'text-emerald-600', hint: t('workspaceAnalytics.loop.kpi.shipRateHint') },
    { key: 'lead', label: t('workspaceAnalytics.loop.kpi.leadToShip'), value: dur(d.lead_to_ship.median_hours), sub: d.lead_to_ship.p85_hours == null ? '' : `P85 ${dur(d.lead_to_ship.p85_hours)}`, tone: 'text-indigo-600', hint: t('workspaceAnalytics.loop.kpi.leadHint') },
    { key: 'spec', label: t('workspaceAnalytics.loop.kpi.specCoverage'), value: s.spec_coverage == null ? '—' : pct(s.spec_coverage), sub: '', tone: 'text-gray-900 dark:text-gray-100', hint: t('workspaceAnalytics.loop.kpi.specHint') },
  ]
})

const extraDurations = computed(() => (data.value ? [data.value.pr_review, data.value.lead_to_done] : []))

watch(() => [props.slug, props.days, props.projectId, props.refreshKey], load)
onMounted(load)
</script>
