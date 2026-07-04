/**
 * useConfirm Composable 单元测试
 * Module-level state is shared — reset in beforeEach
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useConfirm } from './useConfirm'

describe('useConfirm', () => {
  let dialog: ReturnType<typeof useConfirm>

  beforeEach(() => {
    vi.clearAllMocks()
    dialog = useConfirm()
    // Reset module-level state
    dialog.dialogVisible.value = false
    dialog.dialogOptions.value = { message: '' }
  })

  it('should initialize with dialog hidden', () => {
    expect(dialog.dialogVisible.value).toBe(false)
    expect(dialog.dialogOptions.value.message).toBe('')
  })

  it('should open dialog with string message', () => {
    dialog.confirm('Are you sure?')
    expect(dialog.dialogVisible.value).toBe(true)
    expect(dialog.dialogOptions.value.message).toBe('Are you sure?')
  })

  it('should open dialog with ConfirmOptions', () => {
    dialog.confirm({
      message: 'Delete this item?',
      title: 'Confirm Deletion',
      confirmText: 'Yes, delete',
      cancelText: 'Keep',
      danger: true,
    })
    expect(dialog.dialogOptions.value.message).toBe('Delete this item?')
    expect(dialog.dialogOptions.value.title).toBe('Confirm Deletion')
    expect(dialog.dialogOptions.value.danger).toBe(true)
  })

  it('should resolve true on confirm', async () => {
    const promise = dialog.confirm('Sure?')
    dialog.onConfirm()
    const result = await promise
    expect(result).toBe(true)
    expect(dialog.dialogVisible.value).toBe(false)
  })

  it('should resolve false on cancel', async () => {
    const promise = dialog.confirm('Sure?')
    dialog.onCancel()
    const result = await promise
    expect(result).toBe(false)
    expect(dialog.dialogVisible.value).toBe(false)
  })

  it('should handle multiple sequential confirms', async () => {
    const p1 = dialog.confirm('First?')
    dialog.onConfirm()
    expect(await p1).toBe(true)

    const p2 = dialog.confirm('Second?')
    dialog.onCancel()
    expect(await p2).toBe(false)
  })
})
