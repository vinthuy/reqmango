import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import type { User, LoginRequest, UserCreate, Token } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token') || null)
  
  const isLoggedIn = computed(() => !!token.value)
  
  const login = async (data: LoginRequest): Promise<void> => {
    const result: Token = await authApi.login(data)
    token.value = result.access_token
    localStorage.setItem('token', result.access_token)
    user.value = await authApi.getCurrentUser()
  }
  
  const register = async (data: UserCreate): Promise<User> => {
    const result = await authApi.register(data)
    return result
  }
  
  const logout = (): void => {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
  }
  
  const fetchUser = async (): Promise<void> => {
    if (token.value) {
      try {
        user.value = await authApi.getCurrentUser()
      } catch {
        logout()
      }
    }
  }
  
  return {
    user,
    token,
    isLoggedIn,
    login,
    register,
    logout,
    fetchUser
  }
})