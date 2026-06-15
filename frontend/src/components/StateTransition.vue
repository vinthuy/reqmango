<template>
  <div class="state-transition-manager">
    <!-- 头部 -->
    <div class="bg-white border-b border-gray-200 px-4 py-3">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium text-gray-700">状态转换</h3>
        <button
          @click="showCreateModal = true"
          class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          <span>添加转换</span>
        </button>
      </div>
    </div>

    <!-- 状态转换列表 -->
    <div class="p-4">
      <!-- 加载状态 -->
      <div v-if="loading" class="text-center py-12">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>

      <!-- 空状态 -->
      <div v-else-if="transitions.length === 0" class="text-center py-12">
        <svg class="h-12 w-12 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
        </svg>
        <p class="mt-2 text-gray-500">暂无状态转换规则</p>
        <button @click="showCreateModal = true" class="mt-3 text-indigo-600 hover:text-indigo-800 text-sm">
          添加第一个转换
        </button>
      </div>

      <!-- 转换列表 -->
      <div v-else class="space-y-3">
        <div
          v-for="transition in transitions"
          :key="transition.id"
          class="bg-white border border-gray-200 rounded-lg p-4 hover:border-gray-300"
        >
          <div class="flex items-center justify-between">
            <!-- 转换箭头 -->
            <div class="flex items-center space-x-3 flex-1">
              <span class="px-2 py-1 text-xs bg-blue-100 text-blue-700 rounded">
                {{ getStateName(transition.source_state_id) }}
              </span>
              <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
              </svg>
              <span class="px-2 py-1 text-xs bg-green-100 text-green-700 rounded">
                {{ getStateName(transition.target_state_id) }}
              </span>

              <!-- 转换名称 -->
              <span class="text-sm font-medium text-gray-700">{{ transition.name }}</span>

              <!-- 自动转换标签 -->
              <span
                v-if="transition.is_auto"
                class="px-2 py-0.5 text-xs bg-purple-100 text-purple-700 rounded"
              >
                自动
              </span>
            </div>

            <!-- 操作 -->
            <div class="flex items-center space-x-2">
              <button
                @click="editTransition(transition)"
                class="p-1 text-gray-400 hover:text-indigo-600"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button
                @click="deleteTransition(transition)"
                class="p-1 text-gray-400 hover:text-red-600"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>

          <!-- 描述 -->
          <p v-if="transition.description" class="mt-2 text-xs text-gray-500">
            {{ transition.description }}
          </p>
        </div>
      </div>
    </div>

    <!-- 创建/编辑模态框 -->
    <div
      v-if="showCreateModal || showEditModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
    >
      <div class="bg-white rounded-lg p-6 w-full max-w-md mx-4">
        <h3 class="text-lg font-semibold mb-4">
          {{ showEditModal ? '编辑状态转换' : '添加状态转换' }}
        </h3>

        <form @submit.prevent="submitTransition" class="space-y-4">
          <!-- 转换名称 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              名称 *
            </label>
            <input
              v-model="transitionForm.name"
              type="text"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="例如：待办 → 进行中"
            />
          </div>

          <!-- 源状态 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              源状态 *
            </label>
            <select
              v-model="transitionForm.source_state_id"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">选择源状态</option>
              <option v-for="state in states" :key="state.id" :value="state.id">
                {{ state.name }}
              </option>
            </select>
          </div>

          <!-- 目标状态 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              目标状态 *
            </label>
            <select
              v-model="transitionForm.target_state_id"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="">选择目标状态</option>
              <option v-for="state in states" :key="state.id" :value="state.id">
                {{ state.name }}
              </option>
            </select>
          </div>

          <!-- 描述 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              描述
            </label>
            <textarea
              v-model="transitionForm.description"
              rows="2"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          <!-- 自动转换 -->
          <div>
            <label class="flex items-center space-x-2">
              <input
                v-model="transitionForm.is_auto"
                type="checkbox"
                class="w-4 h-4 text-indigo-600 border-gray-300 rounded"
              />
              <span class="text-sm text-gray-700">允许自动转换</span>
            </label>
          </div>

          <!-- 按钮 -->
          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              @click="closeModal"
              class="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              取消
            </button>
            <button
              type="submit"
              class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
            >
              {{ showEditModal ? '保存' : '创建' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import workflowApi from '@/api/workflow'
import type { StateTransition, StateTransitionCreate } from '@/types/workflow'

// Props
const props = defineProps<{
  projectId: string
  workspaceId: string
  states: Array<{ id: string; name: string }>
}>()

// Emits
const emit = defineEmits<{
  (e: 'created'): void
  (e: 'updated'): void
  (e: 'deleted'): void
}>()

// State
const transitions = ref<StateTransition[]>([])
const loading = ref(false)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingTransition = ref<StateTransition | null>(null)

const transitionForm = ref({
  name: '',
  description: '',
  source_state_id: '',
  target_state_id: '',
  is_auto: false
})

// Load transitions
onMounted(() => {
  loadTransitions()
})

async function loadTransitions() {
  loading.value = true
  try {
    transitions.value = await workflowApi.listStateTransitions(props.projectId)
  } catch (error) {
    console.error('Failed to load transitions:', error)
  } finally {
    loading.value = false
  }
}

// Get state name by ID
function getStateName(stateId: string): string {
  const state = props.states.find(s => s.id === stateId)
  return state?.name || '未知'
}

// Edit transition
function editTransition(transition: StateTransition) {
  editingTransition.value = transition
  transitionForm.value = {
    name: transition.name,
    description: transition.description || '',
    source_state_id: transition.source_state_id,
    target_state_id: transition.target_state_id,
    is_auto: transition.is_auto
  }
  showEditModal.value = true
}

// Delete transition
async function deleteTransition(transition: StateTransition) {
  if (!confirm('确定要删除此转换规则吗？')) return

  try {
    await workflowApi.deleteStateTransition(transition.id)
    transitions.value = transitions.value.filter(t => t.id !== transition.id)
    emit('deleted')
  } catch (error) {
    console.error('Failed to delete transition:', error)
  }
}

// Submit form
async function submitTransition() {
  try {
    if (showEditModal.value && editingTransition.value) {
      // Update
      await workflowApi.updateStateTransition(editingTransition.value.id, {
        name: transitionForm.value.name,
        description: transitionForm.value.description || undefined,
        is_auto: transitionForm.value.is_auto
      })
      emit('updated')
    } else {
      // Create
      const data: StateTransitionCreate = {
        name: transitionForm.value.name,
        description: transitionForm.value.description || undefined,
        source_state_id: transitionForm.value.source_state_id,
        target_state_id: transitionForm.value.target_state_id,
        is_auto: transitionForm.value.is_auto,
        project_id: props.projectId
      }
      await workflowApi.createStateTransition(props.projectId, props.workspaceId, data)
      emit('created')
    }
    closeModal()
    await loadTransitions()
  } catch (error) {
    console.error('Failed to submit transition:', error)
  }
}

// Close modal
function closeModal() {
  showCreateModal.value = false
  showEditModal.value = false
  editingTransition.value = null
  transitionForm.value = {
    name: '',
    description: '',
    source_state_id: '',
    target_state_id: '',
    is_auto: false
  }
}
</script>

<style scoped>
.state-transition-manager {
  @apply bg-white rounded-lg;
}
</style>