<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { initiativeApi, type Initiative } from '@/api/initiative'

const route = useRoute()
const slug = route.params.slug as string
const workspaceId = ref<number>(0)
const initiatives = ref<Initiative[]>([])
const loading = ref(true)
const showForm = ref(false)
const editing = ref<Initiative | null>(null)
const form = ref({ name: '', description: '', color: '#3b82f6', status: 'active', target_date: '', start_date: '', project_ids: [] as number[] })
const progressData = ref<Record<number, any>>({})

onMounted(async () => {
  // Get workspace info
  try {
    const wsResp = await fetch(`/api/v1/workspaces/${slug}`)
    if (!wsResp.ok) throw new Error(`Workspace fetch failed: ${wsResp.status}`)
    const body = await wsResp.json()
    const ws = body.data
    if (ws) { workspaceId.value = ws.id; await load(ws.id) }
  } catch (e) {
    console.error('Initiatives: failed to load workspace', e)
  } finally {
    loading.value = false
  }
})

async function load(wsId: number) {
  loading.value = true
  initiatives.value = await initiativeApi.list(wsId)
  // Load progress for each
  for (const ini of initiatives.value) {
    try { progressData.value[ini.id] = await initiativeApi.getProgress(ini.id) } catch(e) {}
  }
  loading.value = false
}

async function save() {
  if (!form.value.name.trim()) {
      alert('请输入名称')
      return
    }
  if (editing.value) {
    await initiativeApi.update(editing.value.id, { ...form.value })
  } else {
    await initiativeApi.create(workspaceId.value, { ...form.value })
  }
  showForm.value = false; editing.value = null
  form.value = { name: '', description: '', color: '#3b82f6', status: 'active', target_date: '', start_date: '', project_ids: [] }
  await load(workspaceId.value)
}

function edit(ini: Initiative) {
  editing.value = ini
  form.value = { name: ini.name, description: ini.description || '', color: ini.color || '#3b82f6', status: ini.status, target_date: ini.target_date || '', start_date: ini.start_date || '', project_ids: (ini.projects || []).map((p: any) => p.id) }
  showForm.value = true
}

async function remove(ini: Initiative) {
  if (!confirm(`确定删除 Initiative "${ini.name}"？`)) return
  await initiativeApi.delete(ini.id)
  await load(workspaceId.value)
}

function getStatusLabel(s: string) {
  const map: Record<string, string> = { active: '进行中', completed: '已完成', paused: '已暂停', at_risk: '有风险', off_track: '偏离轨道' }
  return map[s] || s
}
function getStatusColor(s: string) {
  const map: Record<string, string> = { active: 'bg-green-100 text-green-700', completed: 'bg-blue-100 text-blue-700', paused: 'bg-gray-100 text-gray-600', at_risk: 'bg-yellow-100 text-yellow-700', off_track: 'bg-red-100 text-red-700' }
  return map[s] || 'bg-gray-100'
}
</script>

<template>
  <div class="p-6 max-w-6xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Initiatives</h1>
      <button @click="showForm = true; editing = null; form = { name: '', description: '', color: '#3b82f6', status: 'active', target_date: '', start_date: '', project_ids: [] }" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">+ 新建 Initiative</button>
    </div>

    <div v-if="loading" class="text-gray-500">加载中...</div>

    <div v-else-if="initiatives.length === 0" class="text-center py-12 text-gray-400">
      <div class="text-4xl mb-2">🎯</div>
      <p>暂无 Initiative</p>
      <p class="text-sm mt-1">创建跨项目的战略目标，追踪整体进展</p>
    </div>

    <div v-else class="grid gap-4">
      <div v-for="ini in initiatives" :key="ini.id" class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-5 hover:shadow-md transition">
        <div class="flex items-start justify-between">
          <div class="flex items-center gap-3">
            <div class="w-3 h-3 rounded-full flex-shrink-0" :style="{ backgroundColor: ini.color || '#3b82f6' }"></div>
            <div>
              <h3 class="font-semibold text-lg">{{ ini.name }}</h3>
              <p v-if="ini.description" class="text-gray-500 text-sm mt-1">{{ ini.description }}</p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs px-2 py-0.5 rounded-full" :class="getStatusColor(ini.status)">{{ getStatusLabel(ini.status) }}</span>
            <button @click="edit(ini)" class="text-gray-400 hover:text-gray-600 text-sm">编辑</button>
            <button @click="remove(ini)" class="text-gray-400 hover:text-red-500 text-sm">删除</button>
          </div>
        </div>

        <!-- Progress bar -->
        <div v-if="progressData[ini.id]" class="mt-4">
          <div class="flex justify-between text-sm mb-1">
            <span class="text-gray-500">进度</span>
            <span class="font-medium">{{ Math.round(progressData[ini.id].progress) }}%</span>
          </div>
          <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
            <div class="h-2 rounded-full transition-all" :style="{ width: progressData[ini.id].progress + '%', backgroundColor: ini.color || '#3b82f6' }"></div>
          </div>
          <div class="flex justify-between text-xs text-gray-400 mt-1">
            <span>{{ progressData[ini.id].completed_issues }}/{{ progressData[ini.id].total_issues }} 项完成</span>
            <span>{{ progressData[ini.id].project_count }} 个项目</span>
          </div>
        </div>

        <!-- Linked projects -->
        <div v-if="ini.projects && ini.projects.length > 0" class="mt-3 flex flex-wrap gap-1">
          <span v-for="p in ini.projects" :key="p.id" class="text-xs bg-gray-100 dark:bg-gray-700 px-2 py-0.5 rounded">{{ p.name }}</span>
        </div>

        <!-- Date range -->
        <div v-if="ini.start_date || ini.target_date" class="mt-2 flex gap-3 text-xs text-gray-400">
          <span v-if="ini.start_date">开始: {{ ini.start_date }}</span>
          <span v-if="ini.target_date">目标: {{ ini.target_date }}</span>
        </div>
      </div>
    </div>

    <!-- Form Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black/30 z-50 flex items-center justify-center">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-lg mx-4">
        <h2 class="text-lg font-semibold mb-4">{{ editing ? '编辑' : '新建' }} Initiative</h2>
        <div class="space-y-3">
          <div>
            <label class="text-sm text-gray-600 mb-1 block">名称 *</label>
            <input v-model="form.name" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" placeholder="例如：Q3 国际化">
          </div>
          <div>
            <label class="text-sm text-gray-600 mb-1 block">描述</label>
            <textarea v-model="form.description" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" rows="3" placeholder="简要描述这个 Initiative 的目标"></textarea>
          </div>
          <div class="flex gap-3">
            <div class="flex-1">
              <label class="text-sm text-gray-600 mb-1 block">开始日期</label>
              <input type="date" v-model="form.start_date" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600">
            </div>
            <div class="flex-1">
              <label class="text-sm text-gray-600 mb-1 block">目标日期</label>
              <input type="date" v-model="form.target_date" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600">
            </div>
          </div>
          <div>
            <label class="text-sm text-gray-600 mb-1 block">颜色</label>
            <input type="color" v-model="form.color" class="w-10 h-8 rounded cursor-pointer border">
          </div>
          <div>
            <label class="text-sm text-gray-600 mb-1 block">状态</label>
            <select v-model="form.status" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600">
              <option value="active">进行中</option>
              <option value="completed">已完成</option>
              <option value="at_risk">有风险</option>
              <option value="off_track">偏离轨道</option>
              <option value="paused">已暂停</option>
            </select>
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-5">
          <button @click="showForm = false" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">取消</button>
          <button @click="save" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ editing ? '保存' : '创建' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
