import { ref } from 'vue'

export interface Toast {
  id: number
  message: string
  type: 'success' | 'error' | 'warning' | 'info'
  duration: number
}

const toasts = ref<Toast[]>([])
let nextId = 0

export function useToast() {
  const show = (message: string, type: Toast['type'] = 'info', duration = 3000) => {
    const id = nextId++
    const toast: Toast = { id, message, type, duration }
    toasts.value.push(toast)
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, duration)
  }

  const success = (message: string, duration = 3000) => show(message, 'success', duration)
  const error = (message: string, duration = 5000) => show(message, 'error', duration)
  const warning = (message: string, duration = 4000) => show(message, 'warning', duration)
  const info = (message: string, duration = 3000) => show(message, 'info', duration)

  return { toasts, show, success, error, warning, info }
}
