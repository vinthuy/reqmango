import { ref, watchEffect } from 'vue'

const isDark = ref(false)

export function useDarkMode() {
  const stored = localStorage.getItem('reqmanpy-dark-mode')
  if (stored !== null) {
    isDark.value = stored === 'true'
  } else {
    isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
  }

  watchEffect(() => {
    document.documentElement.classList.toggle('dark', isDark.value)
    localStorage.setItem('reqmanpy-dark-mode', String(isDark.value))
  })

  function toggle() { isDark.value = !isDark.value }

  return { isDark, toggle }
}
