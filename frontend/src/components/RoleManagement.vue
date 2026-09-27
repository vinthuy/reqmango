<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100">{{ t('role.title') }}</h2>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ t('role.description') }}</p>
      </div>
      <button
        @click="showCreateModal = true"
        class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium"
      >
        {{ t('role.create') }}
      </button>
    </div>

    <!-- Role List -->
    <div class="space-y-4">
      <div
        v-for="role in roles"
        :key="role.id"
        class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div
              class="w-3 h-3 rounded-full"
              :class="{
                'bg-green-500': role.level >= 20,
                'bg-blue-500': role.level >= 15 && role.level < 20,
                'bg-gray-400': role.level < 15,
              }"
            />
            <div>
              <div class="flex items-center gap-2">
                <span class="font-medium text-gray-900 dark:text-gray-100">{{ roleName(role) }}</span>
                <span
                  v-if="role.is_system"
                  class="text-xs px-2 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400"
                >
                  {{ t('role.system') }}
                </span>
              </div>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ roleDescription(role) }}</p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-400 dark:text-gray-500">{{ role.permissions.length }} {{ t('role.permissionsCount') }}</span>
            <button
              v-if="!role.is_system"
              @click="editRole(role)"
              class="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400"
            >
              {{ t('common.edit') }}
            </button>
            <button
              v-if="!role.is_system"
              @click="confirmDelete(role)"
              class="text-sm text-red-600 hover:text-red-700 dark:text-red-400"
            >
              {{ t('common.delete') }}
            </button>
          </div>
        </div>

        <!-- Permission chips (expandable) -->
        <div v-if="expandedRole === role.id" class="mt-3 pt-3 border-t border-gray-100 dark:border-gray-700 space-y-2">
          <div v-for="group in groupPermissions(role.permissions)" :key="group.resource" class="flex flex-wrap items-center gap-1.5">
            <span class="text-xs font-medium text-gray-500 dark:text-gray-400 w-24 shrink-0">{{ resourceLabel(group.resource) }}</span>
            <span
              v-for="perm in group.permissions"
              :key="perm.id"
              :title="perm.code"
              class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300"
            >
              {{ permissionLabel(perm) }}
            </span>
          </div>
        </div>
        <button
          @click="expandedRole = expandedRole === role.id ? null : role.id"
          class="mt-2 text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
        >
          {{ expandedRole === role.id ? t('role.collapse') : t('role.expand') }}
        </button>
      </div>
    </div>

    <!-- Permissions Reference -->
    <div class="mt-8">
      <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-3">{{ t('role.allPermissions') }}</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        <div
          v-for="group in allPermissionGroups"
          :key="group.resource"
          class="rounded-lg border border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 p-3"
        >
          <div class="flex items-center justify-between mb-2">
            <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ resourceLabel(group.resource) }}</span>
            <span class="text-xs text-gray-400">{{ group.permissions.length }}</span>
          </div>
          <div class="flex flex-wrap gap-1.5">
            <span
              v-for="perm in group.permissions"
              :key="perm.id"
              :title="perm.code"
              class="text-xs px-2 py-0.5 rounded bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-300"
            >
              {{ permissionLabel(perm) }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Role Modal -->
    <div
      v-if="showCreateModal || editingRole"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="closeModal"
    >
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-lg mx-4 max-h-[80vh] overflow-y-auto">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          {{ editingRole ? t('role.edit') : t('role.create') }}
        </h3>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('role.name') }}</label>
            <input
              v-model="form.name"
              type="text"
              :placeholder="t('role.namePlaceholder')"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('role.descriptionLabel') }}</label>
            <input
              v-model="form.description"
              type="text"
              :placeholder="t('role.descriptionPlaceholder')"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('role.level') }}</label>
            <select
              v-model="form.level"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm"
            >
              <option :value="5">{{ t('role.guest') }}</option>
              <option :value="15">{{ t('role.memberRole') }}</option>
              <option :value="20">{{ t('role.adminRole') }}</option>
              <option :value="10">{{ t('role.custom') }}</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{{ t('role.permissions') }}</label>
            <div class="max-h-72 overflow-y-auto border border-gray-200 dark:border-gray-600 rounded-lg p-2 space-y-3">
              <div v-for="group in allPermissionGroups" :key="group.resource">
                <label class="flex items-center gap-2 px-2 py-1 text-xs font-semibold text-gray-600 dark:text-gray-300 cursor-pointer">
                  <input
                    type="checkbox"
                    :checked="isGroupFullySelected(group.permissions)"
                    :indeterminate.prop="isGroupPartiallySelected(group.permissions)"
                    @change="toggleGroup(group.permissions)"
                    class="rounded border-gray-300 text-blue-600"
                  />
                  {{ resourceLabel(group.resource) }}
                </label>
                <div class="grid grid-cols-2 gap-x-2 pl-6">
                  <label
                    v-for="perm in group.permissions"
                    :key="perm.id"
                    :title="perm.code"
                    class="flex items-center gap-2 py-1 px-2 rounded hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer text-sm"
                  >
                    <input
                      type="checkbox"
                      :value="perm.id"
                      v-model="form.permissions"
                      class="rounded border-gray-300 text-blue-600"
                    />
                    <span class="text-gray-700 dark:text-gray-300">{{ permissionLabel(perm) }}</span>
                  </label>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end gap-2 mt-6">
          <button
            @click="closeModal"
            class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            @click="saveRole"
            :disabled="!form.name"
            class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {{ editingRole ? t('common.save') : t('common.create') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirm -->
    <div
      v-if="deletingRole"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="deletingRole = null"
    >
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm mx-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">{{ t('role.confirmDeleteTitle') }}</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
          {{ t('role.confirmDeleteMessage', { name: deletingRole.name }) }}{{ t('role.irreversible') }}
        </p>
        <div class="flex justify-end gap-2">
          <button
            @click="deletingRole = null"
            class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            @click="doDelete(deletingRole)"
            class="px-4 py-2 text-sm bg-red-600 text-white rounded-lg hover:bg-red-700"
          >
            {{ t('common.delete') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { roleApi } from '../api/role'
import type { Role, Permission } from '../types/role'
import { useI18n } from '@/composables/useI18n'

const { t, isZh } = useI18n()

const RESOURCE_ORDER = [
  'workspace', 'member', 'role', 'project', 'initiative', 'issue', 'cycle', 'module',
  'page', 'comment', 'attachment', 'workflow', 'automation', 'settings', 'report',
  'release', 'time_track', 'ai', 'agent',
]

interface PermissionGroup {
  resource: string
  permissions: Permission[]
}

function groupPermissions(perms: Permission[]): PermissionGroup[] {
  const map = new Map<string, Permission[]>()
  for (const p of perms) {
    const key = p.resource || p.code.split(':')[0]
    if (!map.has(key)) map.set(key, [])
    map.get(key)!.push(p)
  }
  const rank = (r: string) => {
    const i = RESOURCE_ORDER.indexOf(r)
    return i === -1 ? RESOURCE_ORDER.length : i
  }
  return [...map.entries()]
    .sort(([a], [b]) => rank(a) - rank(b) || a.localeCompare(b))
    .map(([resource, permissions]) => ({ resource, permissions }))
}

function resourceLabel(resource: string): string {
  return t(`role.resources.${resource}`, resource)
}

function permissionLabel(perm: Permission): string {
  if (!isZh.value) return perm.name || perm.code
  const action = t(`role.actions.${perm.action}`, '')
  const resource = t(`role.resources.${perm.resource}`, '')
  return action && resource ? `${action}${resource}` : (perm.name || perm.code)
}

function roleName(role: Role): string {
  return role.is_system ? t(`role.systemRoles.${role.name}.name`, role.name) : role.name
}

function roleDescription(role: Role): string {
  return role.is_system ? t(`role.systemRoles.${role.name}.desc`, role.description) : role.description
}

const route = useRoute()
const workspaceSlug = route.params.slug as string

const roles = ref<Role[]>([])
const allPermissions = ref<Permission[]>([])
const expandedRole = ref<number | null>(null)
const showCreateModal = ref(false)
const editingRole = ref<Role | null>(null)
const deletingRole = ref<Role | null>(null)

const form = ref({
  name: '',
  description: '',
  level: 15,
  permissions: [] as number[],
})

const allPermissionGroups = computed(() => groupPermissions(allPermissions.value))

function isGroupFullySelected(perms: Permission[]): boolean {
  return perms.length > 0 && perms.every(p => form.value.permissions.includes(p.id))
}

function isGroupPartiallySelected(perms: Permission[]): boolean {
  const n = perms.filter(p => form.value.permissions.includes(p.id)).length
  return n > 0 && n < perms.length
}

function toggleGroup(perms: Permission[]) {
  const ids = perms.map(p => p.id)
  if (isGroupFullySelected(perms)) {
    form.value.permissions = form.value.permissions.filter(id => !ids.includes(id))
  } else {
    form.value.permissions = [...new Set([...form.value.permissions, ...ids])]
  }
}

onMounted(async () => {
  await Promise.all([loadRoles(), loadPermissions()])
})

async function loadRoles() {
  try {
    const res = await roleApi.listRoles(workspaceSlug)
    roles.value = res.data.data || []
  } catch { /* workspace slug may need numeric ID */ }
}

async function loadPermissions() {
  try {
    const res = await roleApi.listPermissions()
    allPermissions.value = res.data.data || []
  } catch { /* leave empty */ }
}

function editRole(role: Role) {
  editingRole.value = role
  form.value = {
    name: role.name,
    description: role.description,
    level: role.level,
    permissions: role.permissions.map(p => p.id),
  }
}

function closeModal() {
  showCreateModal.value = false
  editingRole.value = null
  form.value = { name: '', description: '', level: 15, permissions: [] }
}

async function saveRole() {
  try {
    if (editingRole.value) {
      await roleApi.updateRole(workspaceSlug, editingRole.value.id, {
        name: form.value.name,
        description: form.value.description,
        level: form.value.level,
        permissions: form.value.permissions,
      })
    } else {
      await roleApi.createRole(workspaceSlug, {
        name: form.value.name,
        description: form.value.description,
        scope: 'workspace',
        level: form.value.level,
        permissions: form.value.permissions,
      })
    }
    closeModal()
    await loadRoles()
  } catch { /* handle error */ }
}

function confirmDelete(role: Role) {
  deletingRole.value = role
}

async function doDelete(role: Role) {
  try {
    await roleApi.deleteRole(workspaceSlug, role.id)
    deletingRole.value = null
    await loadRoles()
  } catch { /* handle error */ }
}
</script>
