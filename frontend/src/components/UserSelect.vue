<template>
  <div class="user-select" ref="containerRef">
    <div class="user-select-input" @click="toggleOpen">
      <span v-if="selectedUser" class="selected-text">
        {{ selectedUser.display_name || selectedUser.username || selectedUser.email }}
      </span>
      <span v-else class="placeholder-text">{{ placeholder }}</span>
      <svg v-if="selectedUser && clearable" @click.stop="clear" class="clear-icon" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
      </svg>
      <svg class="chevron-icon" :class="{ open: isOpen }" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
      </svg>
    </div>
    <div v-if="isOpen" class="user-select-dropdown">
      <div class="search-wrapper">
        <input
          ref="searchInput"
          v-model="searchText"
          type="text"
          class="search-input"
          placeholder="搜索用户..."
          @keydown.escape="close"
          @keydown.enter.prevent="selectFirst"
        />
      </div>
      <ul class="options-list">
        <li
          v-for="user in filteredUsers"
          :key="user.id"
          class="option-item"
          :class="{ selected: (user.id === (modelValue || 0)) }"
          @click="select(user)"
        >
          <div class="user-avatar" :style="{ backgroundColor: getAvatarColor(user.id) }">
            {{ getInitials(user.display_name || user.username || user.email) }}
          </div>
          <div class="user-info">
            <span class="user-name">{{ user.display_name || user.username }}</span>
            <span class="user-email">{{ user.email }}</span>
          </div>
        </li>
        <li v-if="filteredUsers.length === 0" class="no-results">无匹配用户</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'

interface UserOption {
  id: number
  display_name?: string
  username?: string
  email?: string
}

const props = defineProps<{
  modelValue?: number | string
  users: UserOption[]
  placeholder?: string
  clearable?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number | string | undefined): void
}>()

const isOpen = ref(false)
const searchText = ref('')
const containerRef = ref<HTMLElement | null>(null)
const searchInput = ref<HTMLInputElement | null>(null)

const selectedUser = computed(() => {
  if (!props.modelValue) return null
  return props.users.find(u => u.id === props.modelValue) || null
})

const filteredUsers = computed(() => {
  if (!searchText.value.trim()) return props.users
  const q = searchText.value.toLowerCase()
  return props.users.filter(u => {
    return (u.display_name || '').toLowerCase().includes(q) ||
           (u.username || '').toLowerCase().includes(q) ||
           (u.email || '').toLowerCase().includes(q)
  })
})

function toggleOpen() {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    searchText.value = ''
    nextTick(() => searchInput.value?.focus())
  }
}

function close() {
  isOpen.value = false
  searchText.value = ''
}

function select(user: UserOption) {
  emit('update:modelValue', user.id)
  close()
}

function clear() {
  emit('update:modelValue', undefined)
}

function selectFirst() {
  if (filteredUsers.value.length > 0) {
    select(filteredUsers.value[0])
  }
}

function getInitials(name: string | undefined) {
  return name?.charAt(0)?.toUpperCase() || '?'
}

function getAvatarColor(id: number) {
  const colors = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16']
  return colors[id % colors.length]
}

function handleClickOutside(e: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    close()
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.user-select {
  position: relative;
  width: 100%;
}

.user-select-input {
  display: flex;
  align-items: center;
  padding: 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  cursor: pointer;
  background: #fff;
  min-height: 2.25rem;
  font-size: 0.875rem;
}

.user-select-input:hover {
  border-color: #9ca3af;
}

.selected-text {
  flex: 1;
  color: #111827;
}

.placeholder-text {
  flex: 1;
  color: #9ca3af;
}

.clear-icon {
  width: 1rem;
  height: 1rem;
  color: #9ca3af;
  margin-left: 0.25rem;
}

.clear-icon:hover {
  color: #6b7280;
}

.chevron-icon {
  width: 1.25rem;
  height: 1.25rem;
  color: #9ca3af;
  margin-left: 0.25rem;
  transition: transform 0.15s;
}

.chevron-icon.open {
  transform: rotate(180deg);
}

.user-select-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 0.25rem;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1);
  z-index: 50;
  max-height: 260px;
  display: flex;
  flex-direction: column;
}

.search-wrapper {
  padding: 0.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.search-input {
  width: 100%;
  padding: 0.375rem 0.5rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.25rem;
  font-size: 0.8125rem;
  outline: none;
  box-sizing: border-box;
}

.search-input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
}

.options-list {
  list-style: none;
  margin: 0;
  padding: 0;
  overflow-y: auto;
  max-height: 200px;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  cursor: pointer;
  font-size: 0.8125rem;
}

.option-item:hover {
  background-color: #f3f4f6;
}

.option-item.selected {
  background-color: #eef2ff;
}

.user-avatar {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 600;
  flex-shrink: 0;
}

.user-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.user-name {
  color: #111827;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-email {
  color: #9ca3af;
  font-size: 0.75rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.no-results {
  padding: 1rem;
  text-align: center;
  color: #9ca3af;
  font-size: 0.8125rem;
}
</style>
