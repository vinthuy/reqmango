import api from './index'
import type { User, LoginRequest, UserCreate, Token } from '@/types'

export const authApi = {
  register: async (data: UserCreate): Promise<User> => {
    const response = await api.post('/auth/register', data)
    return response.data
  },
  
  login: async (data: LoginRequest): Promise<Token> => {
    const response = await api.post('/auth/login', data)
    return response.data
  },
  
  getCurrentUser: async (): Promise<User> => {
    const response = await api.get('/auth/me')
    return response.data
  },

  listUsers: async (): Promise<UserLite[]> => {
    const response = await api.get('/users')
    return response.data
  }
}

export interface UserLite {
  id: number
  display_name: string
  email: string
  avatar_url?: string
}