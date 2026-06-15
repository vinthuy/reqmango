<template>
  <div class="project-member-list">
    <div class="bg-white rounded-lg border border-gray-200">
      <!-- 头部 -->
      <div class="px-4 py-3 border-b border-gray-200">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-medium text-gray-700">成员管理</h3>
          <button
            @click="showInviteModal = true"
            class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>添加成员</span>
          </button>
        </div>
      </div>

      <!-- 成员列表 -->
      <div class="divide-y divide-gray-100">
        <!-- 加载状态 -->
        <div v-if="loading" class="p-8 text-center">
          <svg class="animate-spin h-6 w-6 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>

        <!-- 空状态 -->
        <div v-else-if="members.length === 0" class="p-8 text-center">
          <svg class="h-10 w-10 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
          </svg>
          <p class="mt-2 text-gray-500 text-sm">暂无成员</p>
        </div>

        <!-- 成员列表 -->
        <div v-else v-for="member in members" :key="member.id" class="p-4 flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <!-- 头像 -->
            <div class="w-8 h-8 rounded-full bg-indigo-500 flex items-center justify-center text-white text-sm font-medium">
              {{ getInitials(member.display_name || member.username) }}
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900">{{ member.display_name || member.username }}</p>
              <p class="text-xs text-gray-500">{{ member.email }}</p>
            </div>
          </div>

          <div class="flex items-center space-x-3">
            <!-- 角色 -->
            <select
              :value="member.role"
              @change="updateMemberRole(member.id, ($event.target as HTMLSelectElement).value)"
              class="px-2 py-1 text-xs border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="member">成员</option>
              <option value="admin">管理员</option>
              <option value="viewer">查看者</option>
            </select>

            <!-- 移除 -->
            <button
              @click="removeMember(member.id)"
              class="p-1 text-gray-400 hover:text-red-600"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 邀请成员模态框 -->
    <div
      v-if="showInviteModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
    >
      <div class="bg-white rounded-lg p-6 w-full max-w-md mx-4">
        <h3 class="text-lg font-semibold mb-4">添加成员</h3>

        <form @submit.prevent="inviteMember" class="space-y-4">
          <!-- 邮箱 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              邮箱地址 *
            </label>
            <input
              v-model="inviteForm.email"
              type="email"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="user@example.com"
            />
          </div>

          <!-- 角色 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              角色
            </label>
            <select
              v-model="inviteForm.role"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="member">成员</option>
              <option value="admin">管理员</option>
              <option value="viewer">查看者</option>
            </select>
          </div>

          <!-- 按钮 -->
          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              @click="showInviteModal = false"
              class="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              取消
            </button>
            <button
              type="submit"
              :disabled="submitting"
              class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
            >
              {{ submitting ? '添加中...' : '添加' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import projectApi from '@/api/project'
import type { ProjectMember } from '@/types/project'

// Props
const props = defineProps<{
  projectId: string
  workspaceId: string
}>()

// Emits
const emit = defineEmits<{
  (e: 'refresh'): void
}>()

// State
const members = ref<ProjectMember[]>([])
const loading = ref(false)
const showInviteModal = ref(false)
const submitting = ref(false)

const inviteForm = ref({
  email: '',
  role: 'member'
})

// Load members
onMounted(() => {
  loadMembers()
})

async function loadMembers() {
  loading.value = true
  try {
    members.value = await projectApi.listProjectMembers(props.projectId)
  } catch (error) {
    console.error('Failed to load members:', error)
  } finally {
    loading.value = false
  }
}

// Get initials
function getInitials(name: string): string {
  return name
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

// Invite member
async function inviteMember() {
  submitting.value = true
  try {
    await projectApi.addProjectMember(props.projectId, {
      email: inviteForm.value.email,
      role: inviteForm.value.role as any
    })
    showInviteModal.value = false
    inviteForm.value = { email: '', role: 'member' }
    await loadMembers()
    emit('refresh')
  } catch (error) {
    console.error('Failed to invite member:', error)
  } finally {
    submitting.value = false
  }
}

// Update member role
async function updateMemberRole(memberId: string, role: string) {
  try {
    await projectApi.updateProjectMember(props.projectId, memberId, { role: role as any })
    emit('refresh')
  } catch (error) {
    console.error('Failed to update member role:', error)
  }
}

// Remove member
async function removeMember(memberId: string) {
  if (!confirm('确定要移除此成员吗？')) return

  try {
    await projectApi.removeProjectMember(props.projectId, memberId)
    await loadMembers()
    emit('refresh')
  } catch (error) {
    console.error('Failed to remove member:', error)
  }
}
</script>

<style scoped>
.project-member-list {
  @apply space-y-4;
}
</style>