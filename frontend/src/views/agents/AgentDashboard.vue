<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-xl font-semibold text-gray-800">{{ t('ai.dashboard.title') }}</h1>
          <p class="text-sm text-gray-500 mt-0.5">{{ t('ai.dashboard.subtitle') }}</p>
        </div>
      </div>
    </header>

    <main class="p-6">
      <div class="max-w-5xl mx-auto">
        <!-- Quick Stats -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
          <div class="bg-white border border-gray-200 rounded-xl p-4">
            <div class="text-2xl font-bold text-indigo-600">{{ templatesCount }}</div>
            <div class="text-xs text-gray-500 mt-1">{{ t('ai.dashboard.templates') }}</div>
          </div>
          <div class="bg-white border border-gray-200 rounded-xl p-4">
            <div class="text-2xl font-bold text-blue-600">{{ configsCount }}</div>
            <div class="text-xs text-gray-500 mt-1">{{ t('ai.dashboard.configs') }}</div>
          </div>
          <div class="bg-white border border-gray-200 rounded-xl p-4">
            <div class="text-2xl font-bold text-purple-600">{{ skillsCount }}</div>
            <div class="text-xs text-gray-500 mt-1">{{ t('ai.dashboard.skills') }}</div>
          </div>
          <div class="bg-white border border-gray-200 rounded-xl p-4">
            <div class="text-2xl font-bold text-teal-600">{{ tasksCount }}</div>
            <div class="text-xs text-gray-500 mt-1">{{ t('ai.dashboard.tasks') }}</div>
          </div>
        </div>

        <!-- Active Loops -->
        <section class="mb-8">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-lg font-semibold text-gray-800">{{ t('ai.dashboard.activeLoops') }}</h2>
            <router-link
              :to="agentLink('/loops')"
              class="text-sm text-indigo-600 hover:text-indigo-800 font-medium"
            >
              {{ t('ai.dashboard.viewAll') }}
            </router-link>
          </div>
          <div v-if="activeRuns.length === 0" class="text-sm text-gray-400 py-8 text-center border border-gray-200 rounded-xl bg-white">
            {{ t('ai.dashboard.noActiveLoops') }}
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="run in activeRuns"
              :key="run.id"
              class="bg-white border border-gray-200 rounded-xl p-4 hover:border-indigo-300 cursor-pointer transition-colors"
              @click="$router.push(agentLink(`/loops/runs/${run.id}`))"
            >
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900">{{ run.goal }}</span>
                <LoopStateBadge :status="run.status" />
              </div>
              <BudgetGauge
                :max-tokens="50000"
                :used-tokens="run.tokens_used"
                :max-iterations="run.max_iterations"
                :current-iteration="run.current_iteration"
              />
            </div>
          </div>
        </section>

        <!-- Navigation Grid -->
        <section class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <router-link
            :to="agentLink('/templates')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center text-white text-xl">
                🎭
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.agentTemplates') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.agentTemplatesDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/configs')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center text-white text-xl">
                ⚙️
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.modelConfigs') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.modelConfigsDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/skills')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white text-xl">
                🛠�?              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.skillsTitle') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.skillsDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="`/workspaces/${getWorkspaceId()}/agents/tasks`"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-teal-500 to-cyan-500 flex items-center justify-center text-white text-xl">
                📋
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.tasksTitle') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.tasksDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/runtimes')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-emerald-500 to-green-500 flex items-center justify-center text-white text-xl">
                🖥�?              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.runtimes') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.runtimesDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="`/workspaces/${getWorkspaceId()}/agents/loops`"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-orange-500 to-amber-500 flex items-center justify-center text-white text-xl">
                🔄
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.loopConfigs') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.loopConfigsDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/monitor')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-red-500 to-rose-500 flex items-center justify-center text-white text-xl">
                📈
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.monitor') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.monitorDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/performance')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-fuchsia-500 to-pink-500 flex items-center justify-center text-white text-xl">
                📊
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.performance') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.performanceDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/sessions')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-green-500 to-emerald-500 flex items-center justify-center text-white text-xl">
                📊
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.sessions') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.sessionsDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/memories')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-violet-500 to-purple-500 flex items-center justify-center text-white text-xl">
                🧠
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.memory') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.memoryDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/autopilot')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-cyan-500 to-blue-500 flex items-center justify-center text-white text-xl">
                �?              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.autopilot') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.autopilotDesc') }}</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/developer')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-slate-700 to-gray-900 flex items-center justify-center text-white text-xl">
                👨‍💻
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.developer') }}</div>
                <div class="text-sm text-gray-500 mt-1">Generate code, commit &amp; open pull requests</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/tester')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-500 flex items-center justify-center text-white text-xl">
                🧪
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.tester') }}</div>
                <div class="text-sm text-gray-500 mt-1">Generate test cases, execute &amp; report bugs</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/cicd')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-sky-500 to-blue-500 flex items-center justify-center text-white text-xl">
                🚀
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.cicd') }}</div>
                <div class="text-sm text-gray-500 mt-1">Configure pipelines &amp; monitor builds</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/sdlc')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-violet-500 to-fuchsia-500 flex items-center justify-center text-white text-xl">
                🧩
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.sdlc') }}</div>
                <div class="text-sm text-gray-500 mt-1">End-to-end delivery pipeline</div>
              </div>
            </div>
          </router-link>

          <router-link
            :to="agentLink('/tools')"
            class="bg-white border border-gray-200 rounded-xl p-5 hover:border-indigo-300 hover:shadow-md transition-all"
          >
            <div class="flex items-start gap-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-500 to-orange-500 flex items-center justify-center text-white text-xl">
                🛠️
              </div>
              <div>
                <div class="font-semibold text-gray-900">{{ t('ai.dashboard.tools') }}</div>
                <div class="text-sm text-gray-500 mt-1">{{ t('ai.dashboard.footer') }}</div>
              </div>
            </div>
          </router-link>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { loopApi, type LoopRun } from '@/api/agent-loop'
import { agentTemplateApi } from '@/api/agent-template'
import { agentConfigApi } from '@/api/agent-config'
import { skillApi } from '@/api/skill'
import { agentTaskApi } from '@/api/agent-task'
import { useI18n } from '@/composables/useI18n'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import LoopStateBadge from '@/components/agents/LoopStateBadge.vue'
import BudgetGauge from '@/components/agents/BudgetGauge.vue'

const { t } = useI18n()
const route = useRoute()
const { getWorkspaceId } = useWorkspaceId()

const activeRuns = ref<LoopRun[]>([])
const templatesCount = ref(0)
const configsCount = ref(0)
const skillsCount = ref(0)
const tasksCount = ref(0)
const resolvedWsId = ref<number | null>(null)

function agentLink(subpath: string): string {
  const slug = route.params.slug as string
  if (slug) return `/workspace/${slug}/agents${subpath}`
  const wsId = resolvedWsId.value
  if (wsId) return `/workspaces/${wsId}/agents${subpath}`
  return '#'
}

onMounted(async () => {
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    resolvedWsId.value = wsId

    // Load stats
    const [templatesRes, configsRes, skillsRes, tasksRes] = await Promise.all([
      agentTemplateApi.list(wsId),
      agentConfigApi.list(wsId),
      skillApi.list(wsId),
      agentTaskApi.list(wsId)
    ])
    templatesCount.value = templatesRes?.length || 0
    configsCount.value = configsRes?.length || 0
    skillsCount.value = skillsRes?.length || 0
    tasksCount.value = tasksRes?.length || 0

    // Load active loops
    const loops = await loopApi.list(wsId)
    for (const loop of loops) {
      if (loop.status === 'active') {
        const runs = await loopApi.getRuns(wsId, loop.id, 5)
        activeRuns.value.push(...runs.filter(r => r.status === 'running'))
      }
    }
  } catch { /* no data yet */ }
})
</script>
