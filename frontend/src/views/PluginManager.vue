<script setup lang="ts">
/**
 * PluginManager - Workspace-level plugin management page
 * Allows browsing catalog, installing/enabling/disabling/uninstalling plugins
 */
import { ref, computed, onMounted } from 'vue'
import { pluginApi } from '@/api/plugin'
import type { Plugin, PluginInfo } from '@/types/plugin'
import { PLUGIN_TYPES, PLUGIN_TYPE_ICONS, SYSTEM_EVENTS } from '@/types/plugin'
import { useI18n } from '@/composables/useI18n'

const props = defineProps<{ workspaceId: number }>()

const { t } = useI18n()

// State
const catalog = ref<PluginInfo[]>([])
const installed = ref<Plugin[]>([])
const loading = ref(true)
const selectedPlugin = ref<Plugin | null>(null)
const selectedLogs = ref<any[]>([])
const viewingLogs = ref(false)
const showConfigModal = ref(false)
const configPlugin = ref<Plugin | null>(null)
const configValues = ref<string>('{}')
const configError = ref('')
const testResult = ref('')

// Computed
const installedSlugs = computed(() => new Set(installed.value.map(p => p.slug)))

function isInstalled(slug: string): boolean {
  return installedSlugs.value.has(slug)
}

function getInstalledPlugin(slug: string): Plugin | undefined {
  return installed.value.find(p => p.slug === slug)
}

// Load data
async function loadData() {
  loading.value = true
  try {
    const [cat, inst] = await Promise.all([
      pluginApi.catalog(props.workspaceId),
      pluginApi.list(props.workspaceId),
    ])
    catalog.value = cat
    installed.value = inst
  } catch (e: any) {
    console.error('Failed to load plugins:', e)
  } finally {
    loading.value = false
  }
}

// Install
async function installPlugin(slug: string) {
  try {
    const p = await pluginApi.install(props.workspaceId, { slug })
    installed.value.push(p)
  } catch (e: any) {
    alert(e?.response?.data?.message || t('plugin.installFailed'))
  }
}

// Uninstall
async function uninstallPlugin(id: number) {
  if (!confirm(t('plugin.confirmUninstall'))) return
  try {
    await pluginApi.uninstall(props.workspaceId, id)
    installed.value = installed.value.filter(p => p.id !== id)
    if (selectedPlugin.value?.id === id) selectedPlugin.value = null
  } catch (e: any) {
    alert(e?.response?.data?.message || t('plugin.uninstallFailed'))
  }
}

// Toggle enable/disable
async function toggleEnabled(plugin: Plugin) {
  try {
    if (plugin.enabled) {
      const updated = await pluginApi.disable(props.workspaceId, plugin.id)
      Object.assign(plugin, updated)
    } else {
      const updated = await pluginApi.enable(props.workspaceId, plugin.id)
      Object.assign(plugin, updated)
    }
  } catch (e: any) {
    alert(e?.response?.data?.message || t('plugin.operationFailed'))
  }
}

// View logs
async function viewLogs(plugin: Plugin) {
  selectedPlugin.value = plugin
  viewingLogs.value = true
  try {
    selectedLogs.value = await pluginApi.logs(props.workspaceId, plugin.id, 50)
  } catch (e: any) {
    console.error('Failed to load logs:', e)
    selectedLogs.value = []
  }
}

// Config modal
function openConfig(plugin: Plugin) {
  configPlugin.value = plugin
  configValues.value = JSON.stringify(plugin.config || {}, null, 2)
  configError.value = ''
  showConfigModal.value = true
}

function closeConfig() {
  showConfigModal.value = false
  configPlugin.value = null
  configValues.value = '{}'
  configError.value = ''
}

async function saveConfig() {
  if (!configPlugin.value) return
  try {
    const parsed = JSON.parse(configValues.value)
    const updated = await pluginApi.update(props.workspaceId, configPlugin.value.id, { config: parsed })
    const idx = installed.value.findIndex(p => p.id === configPlugin.value!.id)
    if (idx !== -1) installed.value[idx] = updated
    closeConfig()
  } catch (e: any) {
    if (e instanceof SyntaxError) {
      configError.value = t('plugin.invalidJson')
    } else {
      configError.value = e?.response?.data?.message || t('plugin.saveConfigFailed')
    }
  }
}

// Test
async function testPlugin(id: number) {
  testResult.value = t('plugin.runningTest')
  try {
    const result = await pluginApi.test(props.workspaceId, id)
    testResult.value = JSON.stringify(result, null, 2)
  } catch (e: any) {
    testResult.value = t('plugin.testFailed') + ' ' + (e?.response?.data?.message || e.message)
  }
}

// Subscribed events display
function eventLabel(eventType: string): string {
  const ev = SYSTEM_EVENTS.find(e => e.value === eventType)
  return ev?.label || eventType
}

onMounted(loadData)
</script>

<template>
  <div class="plugin-manager">
    <div class="pm-header">
      <h2>{{ t('plugin.title') }}</h2>
      <p class="pm-subtitle">{{ t('plugin.subtitle') }}</p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="pm-loading">
      <div class="spinner"></div>
      <span>{{ t('plugin.loading') }}</span>
    </div>

    <template v-else>
      <!-- Catalog -->
      <section class="pm-section">
        <h3>{{ t('plugin.availablePlugins') }}</h3>
        <div class="pm-catalog">
          <div
            v-for="info in catalog"
            :key="info.slug"
            class="pm-card"
            :class="{ installed: isInstalled(info.slug) }"
          >
            <div class="pm-card-header">
              <span class="pm-card-icon">{{ PLUGIN_TYPE_ICONS[info.type] || '🔌' }}</span>
              <div class="pm-card-title">
                <span class="pm-card-name">{{ info.name }}</span>
                <span class="pm-card-type">{{ PLUGIN_TYPES[info.type] || info.type }}</span>
              </div>
              <span v-if="isInstalled(info.slug)" class="pm-badge">{{ t('plugin.installed') }}</span>
            </div>
            <p class="pm-card-desc">{{ info.description }}</p>
            <div class="pm-card-meta">
              <span class="pm-meta-item">v{{ info.version }}</span>
              <span class="pm-meta-item">by {{ info.author }}</span>
            </div>
            <div class="pm-card-actions">
              <button
                v-if="!isInstalled(info.slug)"
                class="btn btn-sm btn-primary"
                @click="installPlugin(info.slug)"
              >
                {{ t('plugin.install') }}
              </button>
              <template v-else>
                <button
                  class="btn btn-sm"
                  :class="getInstalledPlugin(info.slug)?.enabled ? 'btn-warning' : 'btn-success'"
                  @click="toggleEnabled(getInstalledPlugin(info.slug)!)"
                >
                  {{ getInstalledPlugin(info.slug)?.enabled ? t('plugin.disable') : t('plugin.enable') }}
                </button>
                <button
                  class="btn btn-sm btn-outline"
                  @click="openConfig(getInstalledPlugin(info.slug)!)"
                >
                  {{ t('plugin.config') }}
                </button>
                <button
                  class="btn btn-sm btn-outline"
                  @click="viewLogs(getInstalledPlugin(info.slug)!)"
                >
                  {{ t('plugin.logs') }}
                </button>
                <button
                  class="btn btn-sm btn-danger"
                  @click="uninstallPlugin(getInstalledPlugin(info.slug)!.id)"
                >
                  {{ t('plugin.remove') }}
                </button>
              </template>
            </div>

            <!-- Event subscriptions -->
            <div v-if="info.subscribed_events?.length" class="pm-card-events">
              <span class="pm-events-label">{{ t('plugin.subscribesTo') }}</span>
              <span
                v-for="ev in info.subscribed_events"
                :key="ev"
                class="pm-event-tag"
              >{{ eventLabel(ev) }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Installed List -->
      <section v-if="installed.length" class="pm-section">
        <h3>{{ t('plugin.installedPlugins') }} ({{ installed.length }})</h3>
        <div class="pm-installed-list">
          <div
            v-for="p in installed"
            :key="p.id"
            class="pm-installed-item"
            :class="{ disabled: !p.enabled }"
          >
            <div class="pm-item-left">
              <span class="pm-status-dot" :class="{ active: p.enabled }"></span>
              <span class="pm-item-name">{{ p.name }}</span>
              <span class="pm-item-version">v{{ p.version }}</span>
            </div>
            <div class="pm-item-type">{{ PLUGIN_TYPES[p.type] || p.type }}</div>
            <div class="pm-item-actions">
              <label class="pm-toggle">
                <input
                  type="checkbox"
                  :checked="p.enabled"
                  @change="toggleEnabled(p)"
                />
                <span class="pm-toggle-slider"></span>
              </label>
              <button class="btn btn-xs btn-outline" @click="openConfig(p)">{{ t('plugin.config') }}</button>
              <button class="btn btn-xs btn-outline" @click="viewLogs(p)">{{ t('plugin.logs') }}</button>
              <button class="btn btn-xs btn-outline" @click="testPlugin(p.id)">{{ t('plugin.test') }}</button>
              <button class="btn btn-xs btn-danger-outline" @click="uninstallPlugin(p.id)">{{ t('plugin.remove') }}</button>
            </div>
          </div>
        </div>

        <!-- Test result -->
        <div v-if="testResult" class="pm-test-result">
          <pre>{{ testResult }}</pre>
          <button class="btn btn-xs btn-outline" @click="testResult = ''">{{ t('plugin.clear') }}</button>
        </div>
      </section>

      <!-- Empty state -->
      <section v-else class="pm-empty">
        <p>{{ t('plugin.noPlugins') }}</p>
      </section>
    </template>

    <!-- Logs Modal -->
    <div v-if="viewingLogs && selectedPlugin" class="pm-modal-overlay" @click.self="viewingLogs = false">
      <div class="pm-modal">
        <div class="pm-modal-header">
          <h3>{{ t('plugin.executionLogs') }}: {{ selectedPlugin.name }}</h3>
          <button class="pm-modal-close" @click="viewingLogs = false">&times;</button>
        </div>
        <div class="pm-modal-body">
          <div v-if="!selectedLogs.length" class="pm-logs-empty">{{ t('plugin.noLogs') }}</div>
          <div v-else class="pm-logs-list">
            <div
              v-for="log in selectedLogs"
              :key="log.id"
              class="pm-log-entry"
              :class="log.status"
            >
              <div class="pm-log-header">
                <span class="pm-log-status" :class="log.status">{{ t('plugin.' + log.status, log.status.toUpperCase()) }}</span>
                <span class="pm-log-event">{{ eventLabel(log.event_type) }}</span>
                <span class="pm-log-duration">{{ log.duration_ms }}ms</span>
                <span v-if="log.status_code" class="pm-log-code">HTTP {{ log.status_code }}</span>
                <span class="pm-log-time">{{ new Date(log.created_at).toLocaleString() }}</span>
              </div>
              <details v-if="log.response_body" class="pm-log-detail">
                <summary>{{ t('plugin.responseBody') }}</summary>
                <pre>{{ log.response_body }}</pre>
              </details>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Config Modal -->
    <div v-if="showConfigModal && configPlugin" class="pm-modal-overlay" @click.self="closeConfig">
      <div class="pm-modal pm-modal-config">
        <div class="pm-modal-header">
          <h3>{{ t('plugin.configuration') }}: {{ configPlugin.name }}</h3>
          <button class="pm-modal-close" @click="closeConfig">&times;</button>
        </div>
        <div class="pm-modal-body">
          <div v-if="configPlugin.config_schema" class="pm-config-schema">
            <h4>{{ t('plugin.configSchema') }}</h4>
            <pre>{{ JSON.stringify(configPlugin.config_schema, null, 2) }}</pre>
          </div>
          <div class="pm-config-editor">
            <h4>{{ t('plugin.currentConfig') }}</h4>
            <textarea
              v-model="configValues"
              class="pm-config-textarea"
              rows="12"
              placeholder="{}"
            ></textarea>
            <p v-if="configError" class="pm-config-error">{{ configError }}</p>
          </div>
        </div>
        <div class="pm-modal-footer">
          <button class="btn btn-outline" @click="closeConfig">{{ t('plugin.cancel') }}</button>
          <button class="btn btn-primary" @click="saveConfig">{{ t('plugin.save') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.plugin-manager {
  max-width: 960px;
  margin: 0 auto;
  padding: 24px;
}

.pm-header {
  margin-bottom: 32px;
}

.pm-header h2 {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0 0 4px;
  color: var(--text-primary, #1a1a2e);
}

.pm-subtitle {
  color: var(--text-secondary, #666);
  font-size: 0.875rem;
  margin: 0;
}

.pm-loading {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 48px 0;
  color: var(--text-secondary, #666);
  justify-content: center;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color, #e0e0e0);
  border-top-color: var(--primary-color, #4f46e5);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.pm-section {
  margin-bottom: 40px;
}

.pm-section h3 {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0 0 16px;
  color: var(--text-primary, #1a1a2e);
}

/* Catalog Cards */
.pm-catalog {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(420px, 1fr));
  gap: 16px;
}

.pm-card {
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 12px;
  padding: 20px;
  background: var(--bg-card, #fff);
  transition: box-shadow 0.15s, border-color 0.15s;
}

.pm-card:hover {
  border-color: var(--primary-color, #4f46e5);
  box-shadow: 0 2px 12px rgba(79, 70, 229, 0.08);
}

.pm-card.installed {
  border-color: #d1fae5;
  background: #f0fdf4;
}

.pm-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.pm-card-icon {
  font-size: 1.5rem;
}

.pm-card-title {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.pm-card-name {
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-primary, #1a1a2e);
}

.pm-card-type {
  font-size: 0.75rem;
  color: var(--text-secondary, #666);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.pm-badge {
  font-size: 0.7rem;
  padding: 2px 8px;
  border-radius: 10px;
  background: #d1fae5;
  color: #065f46;
  font-weight: 500;
}

.pm-card-desc {
  font-size: 0.85rem;
  color: var(--text-secondary, #666);
  margin: 0 0 12px;
  line-height: 1.5;
}

.pm-card-meta {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
}

.pm-meta-item {
  font-size: 0.75rem;
  color: var(--text-muted, #999);
}

.pm-card-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.pm-card-events {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color, #e0e0e0);
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.pm-events-label {
  font-size: 0.7rem;
  color: var(--text-muted, #999);
}

.pm-event-tag {
  font-size: 0.7rem;
  padding: 1px 8px;
  border-radius: 8px;
  background: var(--bg-tag, #f3f4f6);
  color: var(--text-secondary, #666);
}

/* Installed List */
.pm-installed-list {
  display: flex;
  flex-direction: column;
  gap: 1px;
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 12px;
  overflow: hidden;
}

.pm-installed-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 20px;
  background: var(--bg-card, #fff);
  border-bottom: 1px solid var(--border-color, #e0e0e0);
}

.pm-installed-item:last-child {
  border-bottom: none;
}

.pm-installed-item.disabled {
  opacity: 0.6;
}

.pm-item-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.pm-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ccc;
}

.pm-status-dot.active {
  background: #22c55e;
}

.pm-item-name {
  font-weight: 500;
  color: var(--text-primary, #1a1a2e);
}

.pm-item-version {
  font-size: 0.75rem;
  color: var(--text-muted, #999);
}

.pm-item-type {
  font-size: 0.75rem;
  color: var(--text-secondary, #666);
}

.pm-item-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Toggle */
.pm-toggle {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
}

.pm-toggle input {
  opacity: 0;
  width: 0;
  height: 0;
}

.pm-toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background: #ccc;
  border-radius: 22px;
  transition: 0.2s;
}

.pm-toggle-slider::before {
  content: '';
  position: absolute;
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background: #fff;
  border-radius: 50%;
  transition: 0.2s;
}

.pm-toggle input:checked + .pm-toggle-slider {
  background: #22c55e;
}

.pm-toggle input:checked + .pm-toggle-slider::before {
  transform: translateX(18px);
}

.pm-test-result {
  margin-top: 16px;
  padding: 16px;
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
  background: #1e1e2e;
  position: relative;
}

.pm-test-result pre {
  color: #a6e3a1;
  font-size: 0.8rem;
  white-space: pre-wrap;
  margin: 0;
  max-height: 200px;
  overflow-y: auto;
}

.pm-test-result button {
  position: static;
  margin-top: 8px;
}

.pm-empty {
  text-align: center;
  padding: 48px 0;
  color: var(--text-muted, #999);
}

/* Buttons */
.btn {
  padding: 6px 14px;
  border: 1px solid transparent;
  border-radius: 6px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.15s;
  font-weight: 500;
}

.btn-sm { padding: 4px 10px; font-size: 0.78rem; }
.btn-xs { padding: 3px 8px; font-size: 0.72rem; }

.btn-primary { background: var(--primary-color, #4f46e5); color: #fff; border-color: var(--primary-color, #4f46e5); }
.btn-primary:hover { opacity: 0.9; }

.btn-success { background: #22c55e; color: #fff; border-color: #22c55e; }
.btn-success:hover { opacity: 0.9; }

.btn-warning { background: #f59e0b; color: #fff; border-color: #f59e0b; }
.btn-warning:hover { opacity: 0.9; }

.btn-danger { background: #ef4444; color: #fff; border-color: #ef4444; }
.btn-danger:hover { opacity: 0.9; }

.btn-outline { background: transparent; color: var(--text-primary, #1a1a2e); border-color: var(--border-color, #e0e0e0); }
.btn-outline:hover { background: var(--bg-hover, #f5f5f5); }

.btn-danger-outline { background: transparent; color: #ef4444; border-color: #ef4444; }
.btn-danger-outline:hover { background: #fef2f2; }

/* Modals */
.pm-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.pm-modal {
  background: var(--bg-card, #fff);
  border-radius: 16px;
  width: min(680px, 90vw);
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 8px 40px rgba(0,0,0,0.12);
}

.pm-modal-config {
  width: min(720px, 90vw);
}

.pm-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px 16px;
  border-bottom: 1px solid var(--border-color, #e0e0e0);
}

.pm-modal-header h3 {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary, #1a1a2e);
}

.pm-modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--text-secondary, #666);
  padding: 0 4px;
  line-height: 1;
}

.pm-modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.pm-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 24px;
  border-top: 1px solid var(--border-color, #e0e0e0);
}

/* Logs */
.pm-logs-empty {
  text-align: center;
  padding: 32px;
  color: var(--text-muted, #999);
}

.pm-logs-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pm-log-entry {
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #e0e0e0);
  font-size: 0.82rem;
}

.pm-log-entry.success { border-left: 3px solid #22c55e; }
.pm-log-entry.error { border-left: 3px solid #ef4444; }
.pm-log-entry.skipped { border-left: 3px solid #f59e0b; }

.pm-log-header {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.pm-log-status {
  padding: 1px 8px;
  border-radius: 10px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
}

.pm-log-status.success { background: #d1fae5; color: #065f46; }
.pm-log-status.error { background: #fee2e2; color: #991b1b; }
.pm-log-status.skipped { background: #fef3c7; color: #92400e; }

.pm-log-event { color: var(--text-primary, #1a1a2e); font-weight: 500; }
.pm-log-duration { color: var(--text-secondary, #666); }
.pm-log-code { color: var(--text-muted, #999); }
.pm-log-time { color: var(--text-muted, #999); font-size: 0.72rem; margin-left: auto; }

.pm-log-detail {
  margin-top: 8px;
}

.pm-log-detail summary {
  font-size: 0.72rem;
  color: var(--text-secondary, #666);
  cursor: pointer;
}

.pm-log-detail pre {
  margin-top: 6px;
  padding: 10px;
  background: #1e1e2e;
  color: #a6e3a1;
  border-radius: 6px;
  font-size: 0.72rem;
  max-height: 150px;
  overflow-y: auto;
  white-space: pre-wrap;
}

/* Config Editor */
.pm-config-schema {
  margin-bottom: 16px;
}

.pm-config-schema h4 {
  font-size: 0.85rem;
  font-weight: 600;
  margin: 0 0 8px;
}

.pm-config-schema pre {
  padding: 12px;
  background: #1e1e2e;
  color: #a6e3a1;
  border-radius: 8px;
  font-size: 0.72rem;
  max-height: 160px;
  overflow-y: auto;
}

.pm-config-editor h4 {
  font-size: 0.85rem;
  font-weight: 600;
  margin: 0 0 8px;
}

.pm-config-textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--border-color, #e0e0e0);
  border-radius: 8px;
  font-family: 'SF Mono', 'Monaco', monospace;
  font-size: 0.8rem;
  resize: vertical;
  background: var(--bg-input, #fff);
  color: var(--text-primary, #1a1a2e);
}

.pm-config-error {
  color: #ef4444;
  font-size: 0.78rem;
  margin: 6px 0 0;
}
</style>
