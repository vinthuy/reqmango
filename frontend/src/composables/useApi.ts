import { ref, type Ref } from 'vue'

/**
 * Generic API composable factory.
 *
 * Wraps any API module that exposes at minimum a `list()` method and
 * optional `get()`, `create()`, `update()`, `delete()` methods into a
 * reactive CRUD composable with loading / error state.
 *
 * Usage:
 *   import { defineCrudApi } from '@/composables/useApi'
 *   import { workspaceApi } from '@/api/workspace'
 *   export const useWorkspaceApi = defineCrudApi(workspaceApi)
 *
 * Then in a component:
 *   const { items, loading, error, fetchItems, createItem } = useWorkspaceApi()
 */

export interface CrudApi<Item> {
  list: (...args: any[]) => Promise<Item[] | { items: Item[] }>
  get?: (...args: any[]) => Promise<Item>
  create?: (...args: any[]) => Promise<Item>
  update?: (...args: any[]) => Promise<Item>
  delete?: (...args: any[]) => Promise<void>
}

export interface UseApiReturn<Item> {
  items: Ref<Item[]>
  currentItem: Ref<Item | null>
  loading: Ref<boolean>
  error: Ref<string | null>
  fetchItems: (...args: any[]) => Promise<void>
  fetchItem: (...args: any[]) => Promise<void>
  createItem: (...args: any[]) => Promise<Item | null>
  updateItem: (...args: any[]) => Promise<Item | null>
  removeItem: (...args: any[]) => Promise<boolean>
  clearError: () => void
}

export function defineCrudApi<Item>(api: CrudApi<Item>) {
  return function useApi(): UseApiReturn<Item> {
    const items = ref<Item[]>([]) as Ref<Item[]>
    const currentItem = ref<Item | null>(null) as Ref<Item | null>
    const loading = ref(false)
    const error = ref<string | null>(null)

    function clearError() {
      error.value = null
    }

    async function fetchItems(...args: any[]) {
      loading.value = true
      error.value = null
      try {
        const result = await api.list(...args)
        // Handle both { items: [...] } and bare array responses
        if (Array.isArray(result)) {
          items.value = result
        } else if (result && 'items' in result) {
          items.value = (result as { items: Item[] }).items
        }
      } catch (e: any) {
        error.value = e.response?.data?.message || e.message
      } finally {
        loading.value = false
      }
    }

    async function fetchItem(...args: any[]) {
      if (!api.get) return
      loading.value = true
      error.value = null
      try {
        currentItem.value = await api.get(...args)
      } catch (e: any) {
        error.value = e.response?.data?.message || e.message
      } finally {
        loading.value = false
      }
    }

    async function createItem(...args: any[]): Promise<Item | null> {
      if (!api.create) return null
      error.value = null
      try {
        const created = await api.create(...args)
        items.value.unshift(created)
        return created
      } catch (e: any) {
        error.value = e.response?.data?.message || e.message
        return null
      }
    }

    async function updateItem(...args: any[]): Promise<Item | null> {
      if (!api.update) return null
      error.value = null
      try {
        const updated = await api.update(...args)
        // Try to update in list if the first arg is an ID-like value
        const id = args[0]
        if (id != null) {
          const idx = items.value.findIndex((item: any) => item.id === id)
          if (idx !== -1) items.value[idx] = updated
          if (currentItem.value && (currentItem.value as any).id === id) {
            currentItem.value = updated
          }
        }
        return updated
      } catch (e: any) {
        error.value = e.response?.data?.message || e.message
        return null
      }
    }

    async function removeItem(...args: any[]): Promise<boolean> {
      if (!api.delete) return false
      error.value = null
      try {
        await api.delete(...args)
        const id = args[0]
        if (id != null) {
          items.value = items.value.filter((item: any) => item.id !== id)
          if (currentItem.value && (currentItem.value as any).id === id) {
            currentItem.value = null
          }
        }
        return true
      } catch (e: any) {
        error.value = e.response?.data?.message || e.message
        return false
      }
    }

    return {
      items,
      currentItem,
      loading,
      error,
      fetchItems,
      fetchItem,
      createItem,
      updateItem,
      removeItem,
      clearError,
    }
  }
}
