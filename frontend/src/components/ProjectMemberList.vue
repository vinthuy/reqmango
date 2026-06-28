<template>
  <div class="project-member-list">
    <div class="bg-white rounded-lg border border-gray-200">
      <!-- 头部 -->
      <div class="px-4 py-3 border-b border-gray-200">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-medium text-gray-700">{{ t('projectMember.title') }}</h3>
          <button
            @click="openInviteModal"
            class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>{{ t('projectMember.addMember') }}</span>
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
          <p class="mt-2 text-gray-500 text-sm">{{ t('projectMember.noMembers') }}</p>
        </div>

        <!-- 成员列表 -->
        <div v-else v-for="member in members" :key="member.id" class="p-4 flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <!-- 头像 -->
            <div class="w-8 h-8 rounded-full bg-indigo-500 flex items-center justify-center text-white text-sm font-medium">
              {{ getInitials(member.user?.display_name || member.user?.email || '') }}
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900">{{ member.user?.display_name || member.user?.email }}</p>
              <p class="text-xs text-gray-500">{{ member.user?.email }}</p>
            </div>
          </div>

          <div class="flex items-center space-x-3">
            <!-- 角色 -->
            <select
              :value="getRoleLabel(member.role)"
              @change="updateMemberRole(member.user_id, ($event.target as HTMLSelectElement).value)"
              class="px-2 py-1 text-xs border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="member">{{ t('projectMember.roleMember') }}</option>
              <option value="admin">{{ t('projectMember.roleAdmin') }}</option>
              <option value="viewer">{{ t('projectMember.roleViewer') }}</option>
            </select>

            <!-- 移除 -->
            <button
              @click="removeMember(member.user_id)"
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
        <h3 class="text-lg font-semibold mb-4">{{ t('projectMember.addMember') }}</h3>

        <form @submit.prevent="inviteMember" class="space-y-4">
          <!-- 用户选择 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ t('projectMember.selectUser') }}
            </label>
            <div v-if="selectedUser" class="mb-2 p-2 bg-indigo-50 border border-indigo-200 rounded-md flex items-center justify-between">
              <div class="flex items-center space-x-2">
                <div class="w-6 h-6 rounded-full bg-indigo-500 flex items-center justify-center text-white text-xs font-medium">
                  {{ getInitials(selectedUser.display_name || selectedUser.email || '') }}
                </div>
                <span class="text-sm text-gray-700">{{ selectedUser.display_name }} ({{ selectedUser.email }})</span>
              </div>
              <button type="button" @click="clearSelectedUser" class="text-gray-400 hover:text-red-500">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <!-- 搜索输入框 -->
            <div class="relative">
              <input
                v-model="userSearchQuery"
                type="text"
                :placeholder="t('projectMember.searchUsers')"
                class="w-full px-3 py-2 pl-9 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
              <svg class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <!-- 用户列表（最多显示20条） -->
            <div v-if="userSearchQuery || filteredUsers.length > 0" class="mt-2 max-h-48 overflow-y-auto border border-gray-200 rounded-md bg-white">
              <div
                v-for="user in filteredUsers.slice(0, 20)"
                :key="user.id"
                @click="selectUser(user)"
                class="px-3 py-2 hover:bg-gray-50 cursor-pointer flex items-center space-x-2 border-b border-gray-100 last:border-b-0"
              >
                <div class="w-6 h-6 rounded-full bg-indigo-500 flex items-center justify-center text-white text-xs font-medium">
                  {{ getInitials(user.display_name || user.email || '') }}
                </div>
                <div>
                  <p class="text-sm text-gray-900">{{ user.display_name }}</p>
                  <p class="text-xs text-gray-500">{{ user.email }}</p>
                </div>
              </div>
              <div v-if="filteredUsers.length === 0" class="px-3 py-4 text-center text-gray-500 text-sm">
                {{ t('projectMember.noUsersFound') }}
              </div>
              <div v-if="filteredUsers.length > 20" class="px-3 py-2 text-center text-gray-500 text-xs bg-gray-50">
                {{ t('projectMember.showingTop20') }}
              </div>
            </div>
          </div>

          <!-- 角色 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ t('projectMember.role') }}
            </label>
            <select
              v-model="inviteForm.role"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="member">{{ t('projectMember.roleMember') }}</option>
              <option value="admin">{{ t('projectMember.roleAdmin') }}</option>
              <option value="viewer">{{ t('projectMember.roleViewer') }}</option>
            </select>
          </div>

          <!-- 按钮 -->
          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              @click="closeInviteModal"
              class="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              {{ t('common.cancel') }}
            </button>
            <button
              type="submit"
              :disabled="submitting"
              class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
            >
              {{ submitting ? t('projectMember.adding') : t('projectMember.add') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import projectApi from '@/api/project'
import { authApi } from '@/api/auth'
import type { ProjectMember } from '@/types/project'
import { useConfirm } from '@/composables/useConfirm'
import { useI18n } from '@/composables/useI18n'
import type { UserLite } from '@/types'

const { t } = useI18n()

// Props
const props = defineProps<{
  projectId: number
  workspaceId: number
}>()

// Emits
const emit = defineEmits<{
  (e: 'refresh'): void
}>()

// State
const { confirm } = useConfirm()
const members = ref<ProjectMember[]>([])
const allUsers = ref<UserLite[]>([])
const loading = ref(false)
const showInviteModal = ref(false)
const submitting = ref(false)
const userSearchQuery = ref('')
const selectedUserId = ref<number | null>(null)

const inviteForm = ref({
  userId: '',
  role: 'member'
})

// Open invite modal and reset state
function openInviteModal() {
  showInviteModal.value = true
  userSearchQuery.value = ''
  selectedUserId.value = null
}

function closeInviteModal() {
  showInviteModal.value = false
  userSearchQuery.value = ''
  selectedUserId.value = null
}

// Load members
onMounted(() => {
  loadMembers()
})

async function loadMembers() {
  loading.value = true
  try {
    members.value = await projectApi.listProjectMembers(props.projectId)
    await loadUsers()
  } catch (error) {
    console.error('Failed to load members:', error)
  } finally {
    loading.value = false
  }
}

async function loadUsers() {
  try {
    allUsers.value = await authApi.listUsers()
  } catch (error) {
    console.error('Failed to load users:', error)
  }
}

// Available users (not already in project)
const availableUsers = computed(() => {
  const memberUserIds = new Set(members.value.map(m => m.user_id))
  return allUsers.value.filter(u => !memberUserIds.has(u.id))
})

// Filtered users based on search query
const filteredUsers = computed(() => {
  if (!userSearchQuery.value.trim()) {
    return availableUsers.value
  }
  const query = userSearchQuery.value.toLowerCase()
  return availableUsers.value.filter(user => {
    const displayName = user.display_name?.toLowerCase() || ''
    const email = user.email?.toLowerCase() || ''
    return displayName.includes(query) || email.includes(query)
  })
})

// Selected user
const selectedUser = computed(() => {
  if (!selectedUserId.value) return null
  return allUsers.value.find(u => u.id === selectedUserId.value) || null
})

// Select a user
function selectUser(user: UserLite) {
  selectedUserId.value = user.id
}

// Clear selected user
function clearSelectedUser() {
  selectedUserId.value = null
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

// Get role label from numeric value
function getRoleLabel(role: number): string {
  const roleMap: Record<number, string> = {
    5: 'viewer',
    10: 'member',
    15: 'admin'
  }
  return roleMap[role] || 'member'
}

// Invite member
async function inviteMember() {
  if (!selectedUserId.value) {
    alert(t('projectMember.pleaseSelectUser'))
    return
  }
  
  submitting.value = true
  try {
    await projectApi.addProjectMember(props.projectId, {
      user_id: selectedUserId.value,
      role: inviteForm.value.role as any
    })
    showInviteModal.value = false
    inviteForm.value = { userId: '', role: 'member' }
    selectedUserId.value = null
    userSearchQuery.value = ''
    await loadMembers()
    emit('refresh')
    alert(t('projectMember.addSuccess'))
  } catch (error: any) {
    console.error('Failed to invite member:', error)
    const errorMsg = error.response?.data?.message || error.message || t('projectMember.addFailed')
    alert(t('projectMember.addFailed') + errorMsg)
  } finally {
    submitting.value = false
  }
}

// Update member role
async function updateMemberRole(memberId: number, role: string) {
  try {
    const roleMap: Record<string, number> = {
      viewer: 5,
      member: 10,
      admin: 15
    }
    await projectApi.updateProjectMember(props.projectId, memberId, roleMap[role] || 10)
    emit('refresh')
  } catch (error: any) {
    console.error('Failed to update member role:', error)
    const errorMsg = error.response?.data?.message || error.message || t('projectMember.updateFailed')
    alert(t('projectMember.updateFailed') + errorMsg)
  }
}

// Remove member
async function removeMember(memberId: number) {
  if (!(await confirm(t('projectMember.confirmRemove')))) return

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