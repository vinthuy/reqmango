<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  message,
} from 'ant-design-vue'
import { useI18n } from '@/composables/useI18n'
import {
  budgetApi,
  slaApi,
  decisionApi,
  type ProjectBudget,
  type SLAConfig,
  type DecisionRecord,
} from '@/api/workflow'

const route = useRoute()
const { t } = useI18n()
const projectId = computed(() => Number(route.params.id))

const activeTab = ref('budget')

// ==================== Tab 1: Budget ====================
const budget = ref<ProjectBudget | null>(null)
const budgetLoading = ref(false)
const budgetSaving = ref(false)
const showBudgetEdit = ref(false)

const budgetForm = reactive({
  monthly_budget: 100,
  alert_threshold: 80, // 0-100
  auto_block: false,
})

const budgetUsagePercent = computed(() => {
  if (!budget.value) return 0
  return Math.round(budget.value.budget_usage || 0)
})

const budgetUsageStatus = computed(() => {
  if (budgetUsagePercent.value >= 100) return 'exception'
  const threshold = budget.value?.alert_threshold ?? 80
  if (budgetUsagePercent.value >= threshold) return 'active'
  return 'normal'
})

function formatDateTime(dateStr: string | null) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

async function loadBudget() {
  budgetLoading.value = true
  try {
    const res = await budgetApi.get(projectId.value)
    budget.value = res.data
  } catch (err) {
    console.error('Failed to load budget:', err)
    message.error(t('budgetSla.loadBudgetFailed'))
  } finally {
    budgetLoading.value = false
  }
}

function openEditBudget() {
  if (!budget.value) return
  budgetForm.monthly_budget = budget.value.monthly_budget
  budgetForm.alert_threshold = budget.value.alert_threshold
  budgetForm.auto_block = budget.value.auto_block
  showBudgetEdit.value = true
}

async function handleUpdateBudget() {
  if (!budgetForm.monthly_budget || budgetForm.monthly_budget <= 0) {
    message.warning(t('budgetSla.invalidMonthlyBudget'))
    return
  }
  budgetSaving.value = true
  try {
    await budgetApi.update(projectId.value, {
      monthly_budget: budgetForm.monthly_budget,
      alert_threshold: budgetForm.alert_threshold,
      auto_block: budgetForm.auto_block,
    })
    message.success(t('budgetSla.budgetUpdateSuccess'))
    showBudgetEdit.value = false
    await loadBudget()
  } catch (err: any) {
    message.error(err?.response?.data?.message || t('budgetSla.budgetUpdateFailed'))
  } finally {
    budgetSaving.value = false
  }
}

// ==================== Tab 2: SLA ====================
const sla = ref<SLAConfig | null>(null)
const slaLoading = ref(false)
const slaSaving = ref(false)
const showSLAEdit = ref(false)

const slaForm = reactive({
  normal_task_max: 1800,  // seconds
  complex_task_max: 7200, // seconds
  auto_escalation: true,
  enabled: true,
})

function secondsToHours(sec: number) {
  return (sec / 3600).toFixed(1)
}

async function loadSLA() {
  slaLoading.value = true
  try {
    const res = await slaApi.get(projectId.value)
    sla.value = res.data
  } catch (err) {
    console.error('Failed to load SLA:', err)
    message.error(t('budgetSla.loadSlaFailed'))
  } finally {
    slaLoading.value = false
  }
}

function openEditSLA() {
  if (!sla.value) return
  slaForm.normal_task_max = sla.value.normal_task_max
  slaForm.complex_task_max = sla.value.complex_task_max
  slaForm.auto_escalation = sla.value.auto_escalation
  slaForm.enabled = sla.value.enabled
  showSLAEdit.value = true
}

async function handleUpdateSLA() {
  if (!slaForm.normal_task_max || slaForm.normal_task_max <= 0) {
    message.warning(t('budgetSla.invalidNormalTaskMax'))
    return
  }
  if (slaForm.complex_task_max < slaForm.normal_task_max) {
    message.warning(t('budgetSla.complexTaskMaxLessThanNormal'))
    return
  }
  slaSaving.value = true
  try {
    await slaApi.update(projectId.value, {
      normal_task_max: slaForm.normal_task_max,
      complex_task_max: slaForm.complex_task_max,
      auto_escalation: slaForm.auto_escalation,
      enabled: slaForm.enabled,
    })
    message.success(t('budgetSla.slaUpdateSuccess'))
    showSLAEdit.value = false
    await loadSLA()
  } catch (err: any) {
    message.error(err?.response?.data?.message || t('budgetSla.slaUpdateFailed'))
  } finally {
    slaSaving.value = false
  }
}

async function handleToggleSLAEnabled() {
  if (!sla.value) return
  try {
    await slaApi.update(projectId.value, { enabled: !sla.value.enabled })
    message.success(t('budgetSla.slaToggleSuccess'))
    await loadSLA()
  } catch (err: any) {
    message.error(err?.response?.data?.message || t('budgetSla.statusUpdateFailed'))
  }
}

// ==================== Tab 3: Decision Records ====================
const decisionList = ref<DecisionRecord[]>([])
const decisionLoading = ref(false)
const showDecisionDrawer = ref(false)
const selectedDecision = ref<DecisionRecord | null>(null)

const decisionColumns = computed(() => [
  { title: 'Agent', dataIndex: 'agent_name', key: 'agent_name', width: 140 },
  { title: t('budgetSla.nodeType'), dataIndex: 'node_type', key: 'node_type', width: 140 },
  { title: t('budgetSla.decision'), dataIndex: 'decision', key: 'decision', ellipsis: true },
  { title: t('budgetSla.confidence'), dataIndex: 'confidence', key: 'confidence', width: 140 },
  { title: t('budgetSla.relatedIssue'), dataIndex: 'issue_id', key: 'issue_id', width: 100 },
  { title: t('budgetSla.createdAt'), dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: t('budgetSla.action'), key: 'actions', width: 100 },
])

function getNodeTypeColor(type: string) {
  const map: Record<string, string> = {
    issue_assignment: 'green',
    workflow_node: 'blue',
    triage: 'cyan',
    assign: 'green',
    prioritize: 'orange',
    estimate: 'purple',
    close: 'red',
    escalate: 'volcano',
  }
  return map[type] || 'default'
}

function getConfidenceColor(confidence: number) {
  if (confidence >= 0.9) return '#52c41a'
  if (confidence >= 0.7) return '#faad14'
  return '#ff4d4f'
}

async function loadDecisions() {
  decisionLoading.value = true
  try {
    const res = await decisionApi.list(projectId.value)
    decisionList.value = res.data.data || []
  } catch (err) {
    console.error('Failed to load decisions:', err)
    message.error(t('budgetSla.loadDecisionsFailed'))
  } finally {
    decisionLoading.value = false
  }
}

function openDecisionDetail(record: DecisionRecord) {
  selectedDecision.value = record
  showDecisionDrawer.value = true
}

// ==================== Init ====================
onMounted(async () => {
  await Promise.all([loadBudget(), loadSLA(), loadDecisions()])
})

watch(activeTab, (tab) => {
  if (tab === 'budget') {
    loadBudget()
  } else if (tab === 'sla') {
    loadSLA()
  } else if (tab === 'decision') {
    loadDecisions()
  }
})
</script>

<template>
  <div class="budget-sla-page">
    <a-card bordered :body-style="{ padding: '0' }">
      <template #title>
        <span style="font-size: 16px; font-weight: 600">{{ t('budgetSla.title') }}</span>
      </template>

      <a-tabs v-model:activeKey="activeTab" style="padding: 0 24px">
        <!-- ==================== Tab 1: 成本预算 ==================== -->
        <a-tab-pane key="budget" :tab="t('budgetSla.tabBudget')">
          <div style="padding: 24px 0">
            <!-- Loading -->
            <div v-if="budgetLoading" style="text-align: center; padding: 60px 0">
              <a-spin size="large" />
            </div>

            <!-- Budget Display -->
            <div v-else-if="budget">
              <a-row :gutter="[24, 24]">
                <!-- Budget Overview -->
                <a-col :span="16">
                  <a-card :title="t('budgetSla.budgetOverview')" :bordered="true" style="border-radius: 8px">
                    <template #extra>
                      <a-button type="primary" size="small" @click="openEditBudget">
                        {{ t('budgetSla.edit') }}
                      </a-button>
                    </template>
                    <a-row :gutter="[32, 24]">
                      <a-col :span="8">
                        <a-statistic
                          :title="t('budgetSla.monthlyLimit')"
                          :value="budget.monthly_budget"
                          :precision="2"
                          prefix="$"
                          :value-style="{ fontSize: '24px', fontWeight: 600 }"
                        />
                      </a-col>
                      <a-col :span="8">
                        <a-statistic
                          :title="t('budgetSla.monthlyUsage')"
                          :value="budget.current_cost"
                          :precision="2"
                          prefix="$"
                          :value-style="{
                            fontSize: '24px',
                            fontWeight: 600,
                            color: budgetUsagePercent >= 100 ? '#ff4d4f' : budgetUsagePercent >= budget.alert_threshold ? '#faad14' : '#333'
                          }"
                        />
                      </a-col>
                      <a-col :span="8">
                        <a-statistic
                          :title="t('budgetSla.remaining')"
                          :value="Math.max(0, budget.monthly_budget - budget.current_cost)"
                          :precision="2"
                          prefix="$"
                          :value-style="{ fontSize: '24px', fontWeight: 600, color: '#52c41a' }"
                        />
                      </a-col>
                    </a-row>

                    <a-divider />

                    <!-- Usage Progress -->
                    <div style="margin-bottom: 16px">
                      <div style="display: flex; justify-content: space-between; margin-bottom: 8px">
                        <span style="font-weight: 500">{{ t('budgetSla.usageProgress') }}</span>
                        <span style="color: #666">{{ budgetUsagePercent }}%</span>
                      </div>
                      <a-progress
                        :percent="budgetUsagePercent"
                        :status="budgetUsageStatus"
                        :stroke-color="budgetUsagePercent >= 100 ? '#ff4d4f' : undefined"
                        :show-info="false"
                        size="large"
                      />
                    </div>

                    <!-- Threshold & Auto-block Indicators -->
                    <div style="margin-bottom: 16px; display: flex; align-items: center; gap: 8px">
                      <span style="color: #666; font-size: 13px">{{ t('budgetSla.alertThreshold') }}:</span>
                      <a-tag color="orange">
                        {{ Math.round(budget.alert_threshold) }}%
                      </a-tag>
                      <a-divider type="vertical" />
                      <span style="color: #666; font-size: 13px">{{ t('budgetSla.autoBlock') }}:</span>
                      <a-tag :color="budget.auto_block ? 'red' : 'default'">
                        {{ budget.auto_block ? t('budgetSla.enabled') : t('budgetSla.disabled') }}
                      </a-tag>
                    </div>

                    <!-- Budget Meter Visual -->
                    <div style="background: #f5f5f5; border-radius: 8px; padding: 12px; position: relative; height: 32px">
                      <div
                        style="position: absolute; left: 0; top: 0; bottom: 0; border-radius: 8px; transition: width 0.3s"
                        :style="{
                          width: `${Math.min(100, budgetUsagePercent)}%`,
                          background: budgetUsagePercent >= 100
                            ? 'linear-gradient(90deg, #faad14, #ff4d4f)'
                            : budgetUsagePercent >= budget.alert_threshold
                              ? 'linear-gradient(90deg, #52c41a, #faad14)'
                              : 'linear-gradient(90deg, #1890ff, #52c41a)',
                        }"
                      />
                      <div
                        style="position: absolute; top: -4px; bottom: -4px; width: 2px; background: #ff4d4f"
                        :style="{ left: `${budget.alert_threshold}%` }"
                        :title="t('budgetSla.alertThreshold')"
                      />
                    </div>
                  </a-card>
                </a-col>

                <!-- Budget Status -->
                <a-col :span="8">
                  <a-card :title="t('budgetSla.budgetStatus')" :bordered="true" style="border-radius: 8px">
                    <div style="text-align: center; padding: 20px 0">
                      <a-badge
                        :status="budgetUsagePercent >= 100 ? 'error' : budgetUsagePercent >= budget.alert_threshold ? 'warning' : 'success'"
                        :text="budgetUsagePercent >= 100 ? t('budgetSla.overLimit') : budgetUsagePercent >= budget.alert_threshold ? t('budgetSla.nearThreshold') : t('budgetSla.normal')"
                      />
                      <div style="margin-top: 16px">
                        <a-statistic
                          :title="t('budgetSla.budgetUsageRate')"
                          :value="budgetUsagePercent"
                          suffix="%"
                          :value-style="{
                            fontSize: '36px',
                            fontWeight: 700,
                            color: budgetUsagePercent >= 100 ? '#ff4d4f' : budgetUsagePercent >= budget.alert_threshold ? '#faad14' : '#52c41a'
                          }"
                        />
                      </div>
                    </div>
                    <a-divider />
                    <div style="font-size: 13px; color: #666">
                      <div style="display: flex; justify-content: space-between; margin-bottom: 8px">
                        <span>{{ t('budgetSla.lastReset') }}</span>
                        <span>{{ formatDateTime(budget.last_reset_at) }}</span>
                      </div>
                    </div>
                  </a-card>
                </a-col>
              </a-row>
            </div>

            <!-- No Budget -->
            <div v-else style="text-align: center; padding: 60px 0; color: #999">
              {{ t('budgetSla.noBudgetData') }}
            </div>
          </div>
        </a-tab-pane>

        <!-- ==================== Tab 2: SLA配置 ==================== -->
        <a-tab-pane key="sla" :tab="t('budgetSla.tabSla')">
          <div style="padding: 24px 0">
            <div v-if="slaLoading" style="text-align: center; padding: 60px 0">
              <a-spin size="large" />
            </div>

            <div v-else-if="sla" style="max-width: 720px">
              <a-card :title="t('budgetSla.slaConfig')" :bordered="true" style="border-radius: 8px">
                <template #extra>
                  <a-space>
                    <a-button type="primary" size="small" @click="openEditSLA">{{ t('budgetSla.edit') }}</a-button>
                    <a-button size="small" @click="handleToggleSLAEnabled">
                      {{ sla.enabled ? t('budgetSla.deactivate') : t('budgetSla.activate') }}
                    </a-button>
                  </a-space>
                </template>

                <a-row :gutter="[32, 24]">
                  <a-col :span="12">
                    <a-statistic
                      :title="t('budgetSla.normalTaskMax')"
                      :value="secondsToHours(sla.normal_task_max)"
                      :suffix="t('budgetSla.hours')"
                      :value-style="{ fontSize: '24px', fontWeight: 600 }"
                    />
                    <div style="color: #999; font-size: 12px; margin-top: 4px">
                      {{ sla.normal_task_max }} {{ t('budgetSla.seconds') }}
                    </div>
                  </a-col>
                  <a-col :span="12">
                    <a-statistic
                      :title="t('budgetSla.complexTaskMax')"
                      :value="secondsToHours(sla.complex_task_max)"
                      :suffix="t('budgetSla.hours')"
                      :value-style="{ fontSize: '24px', fontWeight: 600 }"
                    />
                    <div style="color: #999; font-size: 12px; margin-top: 4px">
                      {{ sla.complex_task_max }} {{ t('budgetSla.seconds') }}
                    </div>
                  </a-col>
                </a-row>

                <a-divider />

                <div style="display: flex; align-items: center; gap: 24px; flex-wrap: wrap">
                  <div style="display: flex; align-items: center; gap: 8px">
                    <span style="color: #666; font-size: 13px">{{ t('budgetSla.runningStatus') }}:</span>
                    <a-badge
                      :status="sla.enabled ? 'success' : 'default'"
                      :text="sla.enabled ? t('budgetSla.enabled') : t('budgetSla.disabled')"
                    />
                  </div>
                  <div style="display: flex; align-items: center; gap: 8px">
                    <span style="color: #666; font-size: 13px">{{ t('budgetSla.autoEscalation') }}:</span>
                    <a-tag :color="sla.auto_escalation ? 'orange' : 'default'">
                      {{ sla.auto_escalation ? t('budgetSla.enabled') : t('budgetSla.disabled') }}
                    </a-tag>
                  </div>
                </div>

                <a-divider />

                <div style="background: #f6f8fa; border-radius: 8px; padding: 16px; font-size: 13px; color: #666; line-height: 1.8">
                  <div style="font-weight: 600; color: #333; margin-bottom: 8px">{{ t('budgetSla.slaDescription') }}</div>
                  <ul style="margin: 0; padding-left: 20px">
                    <li>{{ t('budgetSla.slaNormalDesc', { hours: secondsToHours(sla.normal_task_max) }) }}</li>
                    <li>{{ t('budgetSla.slaComplexDesc', { hours: secondsToHours(sla.complex_task_max) }) }}</li>
                    <li v-if="sla.auto_escalation">{{ t('budgetSla.slaAutoEscalationDesc') }}</li>
                    <li v-else>{{ t('budgetSla.slaNoEscalationDesc') }}</li>
                  </ul>
                </div>
              </a-card>
            </div>

            <div v-else style="text-align: center; padding: 60px 0; color: #999">
              {{ t('budgetSla.noSlaData') }}
            </div>
          </div>
        </a-tab-pane>

        <!-- ==================== Tab 3: 决策记录 ==================== -->
        <a-tab-pane key="decision" :tab="t('budgetSla.tabDecision')">
          <div style="padding: 24px 0">
            <a-table
              :columns="decisionColumns"
              :data-source="decisionList"
              :loading="decisionLoading"
              :pagination="{ pageSize: 20, showSizeChanger: true, showTotal: (total: number) => t('budgetSla.totalItems', { total }) }"
              row-key="id"
              size="middle"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'node_type'">
                  <a-tag :color="getNodeTypeColor(record.node_type)">
                    {{ record.node_type || '-' }}
                  </a-tag>
                </template>
                <template v-if="column.key === 'confidence'">
                  <a-progress
                    :percent="Math.round((record.confidence || 0) * 100)"
                    :stroke-color="getConfidenceColor(record.confidence)"
                    :format="(percent: number) => `${percent}%`"
                    size="small"
                    style="width: 100px"
                  />
                </template>
                <template v-if="column.key === 'issue_id'">
                  {{ record.issue_id ? `#${record.issue_id}` : '-' }}
                </template>
                <template v-if="column.key === 'created_at'">
                  {{ formatDateTime(record.created_at) }}
                </template>
                <template v-if="column.key === 'actions'">
                  <a-button size="small" type="link" @click="openDecisionDetail(record)">
                    {{ t('budgetSla.details') }}
                  </a-button>
                </template>
              </template>
            </a-table>
          </div>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <!-- ==================== Budget Edit Modal ==================== -->
    <a-modal
      v-model:open="showBudgetEdit"
      :title="t('budgetSla.editBudgetConfig')"
      :confirm-loading="budgetSaving"
      @ok="handleUpdateBudget"
      @cancel="showBudgetEdit = false"
      width="520px"
      :mask-closable="false"
    >
      <a-form layout="vertical" style="margin-top: 16px">
        <a-form-item :label="t('budgetSla.monthlyLimit')" required>
          <a-input-number
            v-model:value="budgetForm.monthly_budget"
            :min="1"
            :step="100"
            :precision="2"
            style="width: 100%"
            :placeholder="t('budgetSla.monthlyBudgetPlaceholder')"
          />
        </a-form-item>
        <a-form-item :label="t('budgetSla.alertThresholdPercent')">
          <a-row :gutter="16" align="middle">
            <a-col :span="18">
              <a-slider
                v-model:value="budgetForm.alert_threshold"
                :min="0"
                :max="100"
                :step="5"
                :marks="{ 0: '0%', 50: '50%', 80: '80%', 100: '100%' }"
              />
            </a-col>
            <a-col :span="6">
              <a-input-number
                v-model:value="budgetForm.alert_threshold"
                :min="0"
                :max="100"
                :step="5"
                :precision="0"
                style="width: 100%"
              />
            </a-col>
          </a-row>
          <div style="color: #999; font-size: 12px; margin-top: 4px">
            {{ t('budgetSla.alertThresholdHint', { percent: Math.round(budgetForm.alert_threshold) }) }}
          </div>
        </a-form-item>
        <a-form-item :label="t('budgetSla.autoBlockLimit')">
          <a-switch v-model:checked="budgetForm.auto_block" />
          <span style="margin-left: 8px; color: #666; font-size: 13px">
            {{ budgetForm.auto_block ? t('budgetSla.autoBlockEnabled') : t('budgetSla.disabled') }}
          </span>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- ==================== SLA Edit Modal ==================== -->
    <a-modal
      v-model:open="showSLAEdit"
      :title="t('budgetSla.editSlaConfig')"
      :confirm-loading="slaSaving"
      @ok="handleUpdateSLA"
      @cancel="showSLAEdit = false"
      width="520px"
      :mask-closable="false"
    >
      <a-form layout="vertical" style="margin-top: 16px">
        <a-form-item :label="t('budgetSla.normalTaskMaxSeconds')" required>
          <a-input-number
            v-model:value="slaForm.normal_task_max"
            :min="60"
            :max="86400"
            :step="300"
            style="width: 100%"
            :placeholder="t('budgetSla.normalTaskMaxPlaceholder')"
          />
          <div style="color: #999; font-size: 12px; margin-top: 4px">
            {{ t('budgetSla.approxHours', { hours: secondsToHours(slaForm.normal_task_max) }) }}
          </div>
        </a-form-item>
        <a-form-item :label="t('budgetSla.complexTaskMaxSeconds')" required>
          <a-input-number
            v-model:value="slaForm.complex_task_max"
            :min="300"
            :max="604800"
            :step="1800"
            style="width: 100%"
            :placeholder="t('budgetSla.complexTaskMaxPlaceholder')"
          />
          <div style="color: #999; font-size: 12px; margin-top: 4px">
            {{ t('budgetSla.approxHours', { hours: secondsToHours(slaForm.complex_task_max) }) }}
          </div>
        </a-form-item>
        <a-form-item :label="t('budgetSla.autoEscalation')">
          <a-switch v-model:checked="slaForm.auto_escalation" />
          <span style="margin-left: 8px; color: #666; font-size: 13px">
            {{ slaForm.auto_escalation ? t('budgetSla.autoEscalationEnabled') : t('budgetSla.disabled') }}
          </span>
        </a-form-item>
        <a-form-item :label="t('budgetSla.enableSlaMonitoring')">
          <a-switch v-model:checked="slaForm.enabled" />
          <span style="margin-left: 8px; color: #666; font-size: 13px">
            {{ slaForm.enabled ? t('budgetSla.enabled') : t('budgetSla.disabled') }}
          </span>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- ==================== Decision Detail Drawer ==================== -->
    <a-drawer
      v-model:open="showDecisionDrawer"
      :title="t('budgetSla.decisionDetail')"
      placement="right"
      :width="640"
      :destroy-on-close="true"
    >
      <div v-if="selectedDecision">
        <a-row :gutter="[16, 16]">
          <a-col :span="12">
            <div style="font-size: 12px; color: #999; margin-bottom: 4px">{{ t('budgetSla.nodeType') }}</div>
            <a-tag :color="getNodeTypeColor(selectedDecision.node_type)" style="font-size: 14px; padding: 4px 12px">
              {{ selectedDecision.node_type || '-' }}
            </a-tag>
          </a-col>
          <a-col :span="12">
            <div style="font-size: 12px; color: #999; margin-bottom: 4px">Agent</div>
            <div style="font-size: 14px; font-weight: 500">{{ selectedDecision.agent_name || '-' }}</div>
          </a-col>
          <a-col :span="12">
            <div style="font-size: 12px; color: #999; margin-bottom: 4px">{{ t('budgetSla.relatedIssue') }}</div>
            <div style="font-size: 14px; font-weight: 500">{{ selectedDecision.issue_id ? `#${selectedDecision.issue_id}` : '-' }}</div>
          </a-col>
          <a-col :span="12">
            <div style="font-size: 12px; color: #999; margin-bottom: 4px">{{ t('budgetSla.relatedTask') }}</div>
            <div style="font-size: 14px; font-weight: 500">{{ selectedDecision.agent_task_id ? `#${selectedDecision.agent_task_id}` : '-' }}</div>
          </a-col>
        </a-row>

        <a-divider />

        <!-- Decision -->
        <div v-if="selectedDecision.decision" style="margin-bottom: 20px">
          <div style="font-size: 14px; font-weight: 600; margin-bottom: 8px">{{ t('budgetSla.decision') }}</div>
          <div style="background: #f5f5f5; border-radius: 6px; padding: 12px; font-size: 13px; line-height: 1.8; white-space: pre-wrap">
            {{ selectedDecision.decision }}
          </div>
        </div>

        <!-- Reasoning -->
        <div v-if="selectedDecision.reasoning" style="margin-bottom: 20px">
          <div style="font-size: 14px; font-weight: 600; margin-bottom: 8px">{{ t('budgetSla.reasoning') }}</div>
          <div style="background: #f5f5f5; border-radius: 6px; padding: 12px; font-size: 13px; line-height: 1.8; white-space: pre-wrap">
            {{ selectedDecision.reasoning }}
          </div>
        </div>

        <!-- Thinking -->
        <div v-if="selectedDecision.thinking" style="margin-bottom: 20px">
          <div style="font-size: 14px; font-weight: 600; margin-bottom: 8px">{{ t('budgetSla.thinking') }}</div>
          <div style="background: #f5f5f5; border-radius: 6px; padding: 12px; font-size: 13px; line-height: 1.8; white-space: pre-wrap">
            {{ selectedDecision.thinking }}
          </div>
        </div>

        <!-- Alternatives Considered -->
        <div v-if="selectedDecision.alternatives && selectedDecision.alternatives.length > 0" style="margin-bottom: 20px">
          <div style="font-size: 14px; font-weight: 600; margin-bottom: 8px">{{ t('budgetSla.alternatives') }}</div>
          <a-timeline>
            <a-timeline-item
              v-for="(alt, index) in selectedDecision.alternatives"
              :key="index"
              :color="index === 0 ? 'green' : 'gray'"
            >
              {{ alt }}
            </a-timeline-item>
          </a-timeline>
        </div>

        <!-- Confidence Score -->
        <div style="margin-bottom: 20px">
          <div style="font-size: 14px; font-weight: 600; margin-bottom: 8px">{{ t('budgetSla.confidence') }}</div>
          <div style="background: #f5f5f5; border-radius: 8px; padding: 20px; text-align: center">
            <a-progress
              type="dashboard"
              :percent="Math.round((selectedDecision.confidence || 0) * 100)"
              :stroke-color="getConfidenceColor(selectedDecision.confidence)"
              :width="120"
              :format="(percent: number) => `${percent}%`"
            />
            <div style="margin-top: 8px; font-size: 13px; color: #666">
              {{ (selectedDecision.confidence || 0) >= 0.9 ? t('budgetSla.highConfidence') : (selectedDecision.confidence || 0) >= 0.7 ? t('budgetSla.mediumConfidence') : t('budgetSla.lowConfidence') }}
            </div>
          </div>
        </div>

        <!-- Timestamp -->
        <div>
          <div style="font-size: 14px; font-weight: 600; margin-bottom: 8px">{{ t('budgetSla.time') }}</div>
          <a-timeline>
            <a-timeline-item color="blue">
              {{ t('budgetSla.createdAt') }}: {{ formatDateTime(selectedDecision.created_at) }}
            </a-timeline-item>
          </a-timeline>
        </div>
      </div>
    </a-drawer>
  </div>
</template>

<style scoped>
.budget-sla-page {
  padding: 0;
}
</style>
