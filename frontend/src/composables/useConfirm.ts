import { ref } from 'vue'

interface ConfirmOptions {
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
}

const visible = ref(false)
const options = ref<ConfirmOptions>({ message: '' })
let resolvePromise: ((value: boolean) => void) | null = null

export function useConfirm() {
  function confirm(message: string): Promise<boolean>
  function confirm(opts: ConfirmOptions): Promise<boolean>
  function confirm(messageOrOpts: string | ConfirmOptions): Promise<boolean> {
    const opts: ConfirmOptions = typeof messageOrOpts === 'string'
      ? { message: messageOrOpts }
      : messageOrOpts

    options.value = opts
    visible.value = true

    return new Promise((resolve) => {
      resolvePromise = resolve
    })
  }

  function onConfirm() {
    visible.value = false
    resolvePromise?.(true)
    resolvePromise = null
  }

  function onCancel() {
    visible.value = false
    resolvePromise?.(false)
    resolvePromise = null
  }

  return {
    confirm,
    dialogVisible: visible,
    dialogOptions: options,
    onConfirm,
    onCancel,
  }
}
