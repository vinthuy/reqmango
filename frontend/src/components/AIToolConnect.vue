<template>
  <div class="p-6 max-w-4xl space-y-6" data-testid="ai-connect">
    <div>
      <h1 class="text-xl font-semibold text-gray-900">{{ t('aiConnect.title') }}</h1>
      <p class="text-sm text-gray-500 mt-1">{{ t('aiConnect.desc') }}</p>
    </div>

    <!-- Step 1: token -->
    <section class="bg-white border border-gray-200 rounded-xl p-5 space-y-3">
      <h2 class="text-sm font-semibold text-gray-800">1. {{ t('aiConnect.step1') }}</h2>
      <p class="text-xs text-gray-500">{{ t('aiConnect.step1Hint') }}</p>
      <div class="flex gap-2">
        <input
          v-model="newName"
          data-testid="aic-name"
          maxlength="100"
          class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          :placeholder="t('aiConnect.namePlaceholder')"
        />
        <button
          data-testid="aic-create"
          :disabled="creating || !newName.trim()"
          class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50"
          @click="createToken"
        >{{ creating ? t('common.loading') : t('aiConnect.create') }}</button>
      </div>
      <div v-if="createdToken" class="rounded-lg border border-amber-200 bg-amber-50 p-3 space-y-2" data-testid="aic-created">
        <p class="text-xs text-amber-800">{{ t('aiConnect.onceWarning') }}</p>
        <div class="flex items-center gap-2">
          <code class="flex-1 text-xs font-mono break-all bg-white border border-amber-200 rounded px-2 py-1.5" data-testid="aic-token">{{ createdToken }}</code>
          <button class="px-3 py-1.5 text-xs border border-amber-300 rounded-lg bg-white hover:bg-amber-100" data-testid="aic-copy-token" @click="copy(createdToken)">{{ t('aiConnect.copy') }}</button>
        </div>
      </div>
      <details class="text-xs text-gray-500">
        <summary class="cursor-pointer select-none">{{ t('aiConnect.pasteExisting') }}</summary>
        <input
          v-model="pastedToken"
          data-testid="aic-paste"
          class="mt-2 w-full px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono"
          placeholder="reqmango_pat_..."
        />
      </details>
    </section>

    <!-- Step 2: client config -->
    <section class="bg-white border border-gray-200 rounded-xl p-5 space-y-3">
      <h2 class="text-sm font-semibold text-gray-800">2. {{ t('aiConnect.step2') }}</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <label class="text-xs text-gray-600 space-y-1">
          <span>{{ t('aiConnect.binaryPath') }}</span>
          <input v-model="binaryPath" data-testid="aic-bin" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono" />
        </label>
        <label class="text-xs text-gray-600 space-y-1">
          <span>{{ t('aiConnect.apiUrl') }}</span>
          <input :value="apiUrl" readonly data-testid="aic-api-url" class="w-full px-3 py-2 border border-gray-100 bg-gray-50 rounded-lg text-sm font-mono text-gray-600" />
        </label>
      </div>
      <p class="text-xs text-gray-500">
        {{ t('aiConnect.buildHint') }}
        <code class="font-mono bg-gray-100 px-1 rounded">cd sdk &amp;&amp; go build -o bin/reqmango-mcp ./cmd/reqmango-mcp</code>
      </p>

      <div class="flex gap-1 border-b border-gray-200">
        <button
          v-for="c in clients"
          :key="c.id"
          :data-testid="`aic-tab-${c.id}`"
          class="px-3 py-2 text-sm -mb-px border-b-2"
          :class="client === c.id ? 'border-indigo-600 text-indigo-700 font-medium' : 'border-transparent text-gray-500 hover:text-gray-700'"
          @click="client = c.id"
        >{{ c.label }}</button>
      </div>

      <p class="text-xs text-gray-500">{{ clientHint }}</p>
      <div class="relative">
        <pre class="text-xs font-mono bg-gray-900 text-gray-100 rounded-lg p-3 overflow-x-auto whitespace-pre" data-testid="aic-snippet">{{ snippet }}</pre>
        <button
          class="absolute top-2 right-2 px-2 py-1 text-[11px] rounded bg-white/10 text-gray-100 hover:bg-white/20"
          data-testid="aic-copy-snippet"
          @click="copy(snippet)"
        >{{ t('aiConnect.copy') }}</button>
      </div>
      <div v-if="client === 'cursor'" class="flex items-center gap-3">
        <a
          :href="activeToken ? cursorDeeplink : undefined"
          data-testid="aic-deeplink"
          class="inline-flex items-center gap-1.5 px-4 py-2 text-sm rounded-lg"
          :class="activeToken ? 'bg-gray-900 text-white hover:bg-gray-800' : 'bg-gray-100 text-gray-400 pointer-events-none'"
          :aria-disabled="!activeToken"
        >{{ t('aiConnect.addToCursor') }}</a>
        <span v-if="!activeToken" class="text-xs text-gray-400">{{ t('aiConnect.needToken') }}</span>
      </div>
    </section>

    <!-- Step 3: verify -->
    <section class="bg-white border border-gray-200 rounded-xl p-5 space-y-3">
      <h2 class="text-sm font-semibold text-gray-800">3. {{ t('aiConnect.step3') }}</h2>
      <div class="flex items-center gap-3">
        <button
          data-testid="aic-test"
          :disabled="testing || !activeToken"
          class="px-4 py-2 text-sm border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
          @click="runTest"
        >{{ testing ? t('common.loading') : t('aiConnect.test') }}</button>
        <span
          v-if="testResult"
          data-testid="aic-test-result"
          :data-ok="testResult.ok ? '1' : '0'"
          class="text-sm"
          :class="testResult.ok ? 'text-green-600' : 'text-red-600'"
        >{{ testResult.ok ? t('aiConnect.testOk', { name: testResult.name || '' }) : t('aiConnect.testFail', { status: testResult.status }) }}</span>
      </div>
      <div class="text-xs text-gray-500 space-y-1">
        <p class="font-medium text-gray-600">{{ t('aiConnect.tryPrompt') }}</p>
        <code class="block font-mono bg-gray-50 border border-gray-100 rounded px-2 py-1.5">{{ t('aiConnect.examplePrompt') }}</code>
      </div>
    </section>

    <!-- PR write-back -->
    <section class="bg-white border border-gray-200 rounded-xl p-5 space-y-2" data-testid="aic-writeback">
      <h2 class="text-sm font-semibold text-gray-800">{{ t('aiConnect.writebackTitle') }}</h2>
      <p class="text-xs text-gray-500">{{ t('aiConnect.writebackDesc') }}</p>
      <label v-if="projects && projects.length > 1" class="flex items-center gap-2 text-xs text-gray-600">
        <span>{{ t('aiConnect.project') }}</span>
        <select
          :value="selectedProject?.id"
          data-testid="aic-project"
          class="px-2 py-1 border border-gray-200 rounded-lg text-xs"
          @change="webhookProjectId = Number(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </label>
      <ul class="text-xs text-gray-600 list-disc pl-5 space-y-1">
        <li>{{ t('aiConnect.writeback1') }} <code class="font-mono bg-gray-100 px-1 rounded" data-testid="aic-webhook-url">{{ webhookUrl }}</code></li>
        <li>{{ t('aiConnect.writeback2', { key: `${selectedProject?.identifier || 'PROJ'}-123` }) }}</li>
        <li>{{ t('aiConnect.writeback3') }}</li>
      </ul>
    </section>

    <!-- Tokens -->
    <section class="bg-white border border-gray-200 rounded-xl p-5 space-y-3">
      <div class="flex items-center justify-between">
        <h2 class="text-sm font-semibold text-gray-800">{{ t('aiConnect.tokens') }}</h2>
        <button class="text-xs text-indigo-600 hover:text-indigo-700" data-testid="aic-refresh" @click="load">{{ t('aiConnect.refresh') }}</button>
      </div>
      <div v-if="loading" class="text-sm text-gray-400">{{ t('common.loading') }}</div>
      <div v-else-if="tokens.length === 0" class="text-sm text-gray-400" data-testid="aic-empty">{{ t('aiConnect.noTokens') }}</div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="text-left text-xs text-gray-500 border-b border-gray-100">
            <th class="py-2 font-medium">{{ t('aiConnect.colName') }}</th>
            <th class="py-2 font-medium">{{ t('aiConnect.colPrefix') }}</th>
            <th class="py-2 font-medium">{{ t('aiConnect.colStatus') }}</th>
            <th class="py-2 font-medium">{{ t('aiConnect.colCreated') }}</th>
            <th class="py-2"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="tk in tokens" :key="tk.id" class="border-b border-gray-50" :data-testid="`aic-row-${tk.id}`">
            <td class="py-2 text-gray-800">{{ tk.name }}</td>
            <td class="py-2 font-mono text-xs text-gray-500">{{ tk.token_prefix }}…</td>
            <td class="py-2">
              <span class="inline-flex items-center gap-1.5 text-xs" :data-testid="`aic-status-${tk.id}`" :data-status="statusOf(tk)">
                <span class="w-1.5 h-1.5 rounded-full" :class="statusDot(tk)"></span>
                {{ statusLabel(tk) }}
              </span>
            </td>
            <td class="py-2 text-xs text-gray-500">{{ formatDate(tk.created_at) }}</td>
            <td class="py-2 text-right">
              <button
                v-if="!tk.revoked_at"
                :data-testid="`aic-revoke-${tk.id}`"
                class="text-xs text-red-600 hover:text-red-700"
                @click="revoke(tk)"
              >{{ t('aiConnect.revoke') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
import { patApi, testPAT, type PersonalAccessToken } from '@/api/pat'

const props = defineProps<{ workspaceId: number; projects?: { id: number; name: string; identifier?: string }[] }>()

const { t } = useI18n()
const toast = useToast()
const { confirm } = useConfirm()

const tokens = ref<PersonalAccessToken[]>([])
const loading = ref(false)
const newName = ref('Cursor')
const creating = ref(false)
const createdToken = ref('')
const pastedToken = ref('')
const binaryPath = ref('reqmango-mcp')
const client = ref<'cursor' | 'claude-code' | 'claude-desktop'>('cursor')
const testing = ref(false)
const testResult = ref<{ ok: boolean; status: number; name?: string } | null>(null)

const clients = [
  { id: 'cursor' as const, label: 'Cursor' },
  { id: 'claude-code' as const, label: 'Claude Code' },
  { id: 'claude-desktop' as const, label: 'Claude Desktop' },
]

const apiUrl = computed(() => `${window.location.origin}/api/v1`)
const activeToken = computed(() => pastedToken.value.trim() || createdToken.value)
const tokenForSnippet = computed(() => activeToken.value || '<REQMANGO_PAT>')
const webhookProjectId = ref<number | null>(null)
const selectedProject = computed(() =>
  props.projects?.find(p => p.id === webhookProjectId.value) || props.projects?.[0])
const webhookUrl = computed(() => `${apiUrl.value}/webhook/git/${selectedProject.value?.id ?? '<project_id>'}`)

const serverConfig = computed(() => ({
  command: binaryPath.value.trim() || 'reqmango-mcp',
  env: { REQMANGO_PAT: tokenForSnippet.value, REQMANGO_API_URL: apiUrl.value },
}))

function shellQuote(v: string): string {
  return /^[\w@%+=:,./-]+$/.test(v) ? v : `'${v.replace(/'/g, `'\\''`)}'`
}

const snippet = computed(() => {
  const cfg = serverConfig.value
  if (client.value === 'claude-code') {
    return [
      'claude mcp add reqmango \\',
      `  --env REQMANGO_PAT=${shellQuote(cfg.env.REQMANGO_PAT)} \\`,
      `  --env REQMANGO_API_URL=${shellQuote(cfg.env.REQMANGO_API_URL)} \\`,
      `  -- ${shellQuote(cfg.command)}`,
    ].join('\n')
  }
  return JSON.stringify({ mcpServers: { reqmango: cfg } }, null, 2)
})

const clientHint = computed(() => {
  if (client.value === 'claude-code') return t('aiConnect.hintClaudeCode')
  if (client.value === 'claude-desktop') return t('aiConnect.hintClaudeDesktop')
  return t('aiConnect.hintCursor')
})

function toBase64(s: string): string {
  const bytes = new TextEncoder().encode(s)
  let bin = ''
  bytes.forEach(b => { bin += String.fromCharCode(b) })
  return btoa(bin)
}

const cursorDeeplink = computed(() =>
  `cursor://anysphere.cursor-deeplink/mcp/install?name=reqmango&config=${encodeURIComponent(toBase64(JSON.stringify(serverConfig.value)))}`)

async function load() {
  loading.value = true
  try {
    tokens.value = await patApi.list()
  } catch {
    toast.error(t('aiConnect.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function createToken() {
  const name = newName.value.trim()
  if (!name) return
  creating.value = true
  try {
    const res = await patApi.create(name)
    createdToken.value = res.token
    pastedToken.value = ''
    testResult.value = null
    toast.success(t('aiConnect.created'))
    await load()
  } catch {
    toast.error(t('aiConnect.createFailed'))
  } finally {
    creating.value = false
  }
}

async function revoke(tk: PersonalAccessToken) {
  const ok = await confirm({ title: t('aiConnect.revoke'), message: t('aiConnect.revokeConfirm', { name: tk.name }), danger: true })
  if (!ok) return
  try {
    await patApi.revoke(tk.id)
    toast.success(t('aiConnect.revoked'))
    await load()
  } catch {
    toast.error(t('aiConnect.revokeFailed'))
  }
}

async function runTest() {
  if (!activeToken.value) return
  testing.value = true
  try {
    testResult.value = await testPAT(activeToken.value)
    await load()
  } catch {
    testResult.value = { ok: false, status: 0 }
  } finally {
    testing.value = false
  }
}

async function copy(text: string) {
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    const ta = document.createElement('textarea')
    ta.value = text
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    ta.remove()
  }
  toast.success(t('aiConnect.copied'))
}

function statusOf(tk: PersonalAccessToken): 'revoked' | 'expired' | 'connected' | 'unused' {
  if (tk.revoked_at) return 'revoked'
  if (tk.expires_at && new Date(tk.expires_at).getTime() < Date.now()) return 'expired'
  return tk.last_used_at ? 'connected' : 'unused'
}

function statusDot(tk: PersonalAccessToken): string {
  switch (statusOf(tk)) {
    case 'connected': return 'bg-green-500'
    case 'unused': return 'bg-gray-300'
    default: return 'bg-red-400'
  }
}

function relative(iso: string): string {
  const mins = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 60000))
  if (mins < 1) return t('aiConnect.justNow')
  if (mins < 60) return t('aiConnect.minutesAgo', { n: mins })
  const hours = Math.round(mins / 60)
  if (hours < 24) return t('aiConnect.hoursAgo', { n: hours })
  return t('aiConnect.daysAgo', { n: Math.round(hours / 24) })
}

function statusLabel(tk: PersonalAccessToken): string {
  const s = statusOf(tk)
  if (s === 'connected') return t('aiConnect.statusConnected', { when: relative(tk.last_used_at!) })
  return t(`aiConnect.status_${s}`)
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString()
}

onMounted(load)
</script>
