import { describe, it, expect, vi, beforeEach } from 'vitest'
import { defineCrudApi } from './useApi'

// Mock API module
function createMockApi() {
  return {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  }
}

describe('useApi (defineCrudApi)', () => {
  let mockApi: ReturnType<typeof createMockApi>
  let useApi: ReturnType<typeof defineCrudApi<{ id: number; name: string }>>

  beforeEach(() => {
    vi.clearAllMocks()
    mockApi = createMockApi()
    useApi = defineCrudApi<{ id: number; name: string }>(mockApi)
  })

  it('should initialize with empty state', () => {
    const { items, currentItem, loading, error } = useApi()
    expect(items.value).toEqual([])
    expect(currentItem.value).toBeNull()
    expect(loading.value).toBe(false)
    expect(error.value).toBeNull()
  })

  it('should fetch items from array response', async () => {
    mockApi.list.mockResolvedValue([{ id: 1, name: 'A' }, { id: 2, name: 'B' }])
    const { items, loading, error, fetchItems } = useApi()

    await fetchItems()

    expect(loading.value).toBe(false)
    expect(error.value).toBeNull()
    expect(items.value).toEqual([{ id: 1, name: 'A' }, { id: 2, name: 'B' }])
  })

  it('should fetch items from { items } response', async () => {
    mockApi.list.mockResolvedValue({ items: [{ id: 1, name: 'A' }] })
    const { items, fetchItems } = useApi()

    await fetchItems()

    expect(items.value).toEqual([{ id: 1, name: 'A' }])
  })

  it('should handle fetch errors', async () => {
    mockApi.list.mockRejectedValue({ response: { data: { message: 'Not found' } } })
    const { items, error, fetchItems } = useApi()

    await fetchItems()

    expect(items.value).toEqual([])
    expect(error.value).toBe('Not found')
  })

  it('should create an item and prepend to list', async () => {
    mockApi.create.mockResolvedValue({ id: 3, name: 'C' })
    const { items, createItem } = useApi()

    const result = await createItem({ name: 'C' })

    expect(result).toEqual({ id: 3, name: 'C' })
    expect(items.value).toEqual([{ id: 3, name: 'C' }])
  })

  it('should update an item in the list', async () => {
    const { items, fetchItems, updateItem } = useApi()
    mockApi.list.mockResolvedValue([{ id: 1, name: 'A' }, { id: 2, name: 'B' }])
    await fetchItems()

    mockApi.update.mockResolvedValue({ id: 1, name: 'A-updated' })
    const result = await updateItem(1, { name: 'A-updated' })

    expect(result).toEqual({ id: 1, name: 'A-updated' })
    expect(items.value[0].name).toBe('A-updated')
  })

  it('should remove an item from the list', async () => {
    const { items, fetchItems, removeItem } = useApi()
    mockApi.list.mockResolvedValue([{ id: 1, name: 'A' }, { id: 2, name: 'B' }])
    await fetchItems()

    mockApi.delete.mockResolvedValue(undefined)
    const result = await removeItem(1)

    expect(result).toBe(true)
    expect(items.value).toEqual([{ id: 2, name: 'B' }])
  })

  it('should clear error', async () => {
    mockApi.list.mockRejectedValue(new Error('fail'))
    const { error, fetchItems, clearError } = useApi()

    await fetchItems()
    expect(error.value).toBe('fail')

    clearError()
    expect(error.value).toBeNull()
  })

  it('should pass arguments to list', async () => {
    mockApi.list.mockResolvedValue([])
    const { fetchItems } = useApi()

    await fetchItems(123, { status: 'active' })

    expect(mockApi.list).toHaveBeenCalledWith(123, { status: 'active' })
  })
})
