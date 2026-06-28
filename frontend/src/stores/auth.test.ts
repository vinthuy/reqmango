import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from './auth'

// Mock the auth API module
vi.mock('@/api/auth', () => ({
  authApi: {
    login: vi.fn(),
    register: vi.fn(),
    getCurrentUser: vi.fn(),
  }
}))

import { authApi } from '@/api/auth'

describe('Auth Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should start with no user', () => {
      const store = useAuthStore()
      expect(store.user).toBeNull()
    })

    it('should start with no token in a clean localStorage', () => {
      const store = useAuthStore()
      expect(store.token).toBeNull()
    })

    it('should not be logged in initially', () => {
      const store = useAuthStore()
      expect(store.isLoggedIn).toBe(false)
    })

    it('should read token from localStorage on init', () => {
      localStorage.setItem('token', 'test-token-123')
      const store = useAuthStore()
      expect(store.token).toBe('test-token-123')
      expect(store.isLoggedIn).toBe(true)
    })
  })

  describe('login', () => {
    it('should set token and user on successful login', async () => {
      const mockToken = { access_token: 'jwt-token-xyz', token_type: 'Bearer', expires_at: '2099-12-31T23:59:59Z' }
      const mockUser = {
        id: 1,
        email: 'test@example.com',
        username: 'testuser',
        display_name: 'Test User',
      }

      vi.mocked(authApi.login).mockResolvedValue(mockToken)
      vi.mocked(authApi.getCurrentUser).mockResolvedValue(mockUser as any)

      const store = useAuthStore()
      await store.login({ email: 'test@example.com', password: 'password123' })

      expect(store.token).toBe('jwt-token-xyz')
      expect(store.isLoggedIn).toBe(true)
      expect(store.user).toEqual(mockUser)
      expect(localStorage.getItem('token')).toBe('jwt-token-xyz')
    })

    it('should throw on failed login', async () => {
      vi.mocked(authApi.login).mockRejectedValue(new Error('Invalid credentials'))

      const store = useAuthStore()
      await expect(
        store.login({ email: 'bad@example.com', password: 'wrong' })
      ).rejects.toThrow('Invalid credentials')

      expect(store.token).toBeNull()
      expect(store.user).toBeNull()
    })
  })

  describe('register', () => {
    it('should call register API and return user', async () => {
      const mockUser = {
        id: 2,
        email: 'new@example.com',
        username: 'newuser',
        display_name: 'New User',
      }

      vi.mocked(authApi.register).mockResolvedValue(mockUser as any)

      const store = useAuthStore()
      const result = await store.register({
        email: 'new@example.com',
        username: 'newuser',
        password: 'password123',
      })

      expect(result).toEqual(mockUser)
      // Register should NOT log the user in
      expect(store.token).toBeNull()
    })
  })

  describe('logout', () => {
    it('should clear user, token, and localStorage', async () => {
      // First login
      vi.mocked(authApi.login).mockResolvedValue({ access_token: 'token', token_type: 'Bearer', expires_at: '2099-12-31T23:59:59Z' })
      vi.mocked(authApi.getCurrentUser).mockResolvedValue({ id: 1 } as any)

      const store = useAuthStore()
      await store.login({ email: 'test@example.com', password: 'pass' })

      expect(store.isLoggedIn).toBe(true)

      // Then logout
      store.logout()

      expect(store.token).toBeNull()
      expect(store.user).toBeNull()
      expect(store.isLoggedIn).toBe(false)
      expect(localStorage.getItem('token')).toBeNull()
    })
  })

  describe('fetchUser', () => {
    it('should fetch user when token exists', async () => {
      localStorage.setItem('token', 'existing-token')
      const mockUser = { id: 3, email: 'existing@example.com', display_name: 'Existing' }

      vi.mocked(authApi.getCurrentUser).mockResolvedValue(mockUser as any)

      const store = useAuthStore()
      await store.fetchUser()

      expect(store.user).toEqual(mockUser)
    })

    it('should logout if fetchUser fails', async () => {
      localStorage.setItem('token', 'bad-token')

      vi.mocked(authApi.getCurrentUser).mockRejectedValue(new Error('Unauthorized'))

      const store = useAuthStore()
      await store.fetchUser()

      expect(store.user).toBeNull()
      expect(store.token).toBeNull()
      expect(store.isLoggedIn).toBe(false)
      expect(localStorage.getItem('token')).toBeNull()
    })

    it('should skip fetch when no token exists', async () => {
      const store = useAuthStore()
      await store.fetchUser()

      expect(authApi.getCurrentUser).not.toHaveBeenCalled()
    })
  })
})
