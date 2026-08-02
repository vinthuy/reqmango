import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json'
  },
  timeout: 30000
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config
    if (error.response?.status === 307 && !originalRequest._retry) {
      originalRequest._retry = true
      const location = error.response.headers.location
      if (location) {
        const baseURL = '/api/v1'
        originalRequest.url = location.startsWith(baseURL) ? location.slice(baseURL.length) : location
      }
      return api(originalRequest)
    }
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
      return Promise.reject(error)
    }
    if (error.response?.status === 429) {
      console.warn('Rate limited, please try again later')
      return Promise.reject(error)
    }
    if (error.response?.status >= 500) {
      console.error('Server error:', error.response.data)
    }
    if (error.code === 'ECONNABORTED') {
      console.error('Request timeout')
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

// Re-export RQL API
export { rqlApi } from './rql'

// Re-export chat API
export { default as chatApi } from './chat'