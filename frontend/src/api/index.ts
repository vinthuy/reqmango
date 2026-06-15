import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json'
  }
})

api.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api

// Re-export custom field API
export { customFieldApi } from './custom-field'

// Re-export project settings API
export { projectSettingsApi } from './project-settings'

// Re-export workflow API
export { workflowApi } from './workflow'

// Re-export issue API
export { issueApi } from './issue'

// Re-export cycle API
export { cycleApi } from './cycle'

// Re-export module API
export { moduleApi } from './module'

// Re-export project API
export { projectApi } from './project'

// Re-export estimate point API
export { default as estimatePointApi } from './estimate-point'

// Re-export notification API
export { default as notificationApi } from './notification'

// Re-export comment API
export { default as commentApi } from './comment'

// Re-export attachment API
export { default as attachmentApi } from './attachment'