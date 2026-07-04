/**
 * Workflow API 单元测试 - Mock axios 验证端点
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPut = vi.fn()
const mockDelete = vi.fn()
const mockPatch = vi.fn()

vi.mock('./index', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    put: (...args: any[]) => mockPut(...args),
    delete: (...args: any[]) => mockDelete(...args),
    patch: (...args: any[]) => mockPatch(...args),
  },
}))

import workflowApi, {
  listWorkflows, createWorkflow, updateWorkflow, deleteWorkflow,
  addTransition, updateTransition, deleteTransition,
  listAutomations, createAutomation, updateAutomation, deleteAutomation,
  listStateTransitions, createStateTransition,
  listAutomationTemplates, toggleAutomationRule,
} from './workflow'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('Workflow CRUD API', () => {
  it('listWorkflows should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listWorkflows(1)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/workflows')
  })

  it('createWorkflow should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1, name: 'Default' } })
    await createWorkflow(1, { name: 'Default' })
    expect(mockPost).toHaveBeenCalledWith('/projects/1/workflows', { name: 'Default' })
  })

  it('updateWorkflow should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 5 } })
    await updateWorkflow(1, 5, { name: 'Updated' })
    expect(mockPut).toHaveBeenCalledWith('/projects/1/workflows/5', { name: 'Updated' })
  })

  it('deleteWorkflow should DELETE', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await deleteWorkflow(1, 5)
    expect(mockDelete).toHaveBeenCalledWith('/projects/1/workflows/5')
  })
})

describe('Transition API', () => {
  it('addTransition should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1 } })
    await addTransition(1, 5, { source_state_id: 1, target_state_id: 2 })
    expect(mockPost).toHaveBeenCalledWith('/projects/1/workflows/5/transitions', { source_state_id: 1, target_state_id: 2 })
  })

  it('updateTransition should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 10 } })
    await updateTransition(1, 5, 10, { name: 'T' })
    expect(mockPut).toHaveBeenCalledWith('/projects/1/workflows/5/transitions/10', { name: 'T' })
  })

  it('deleteTransition should DELETE', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await deleteTransition(1, 5, 10)
    expect(mockDelete).toHaveBeenCalledWith('/projects/1/workflows/5/transitions/10')
  })
})

describe('Automation API', () => {
  it('listAutomations should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listAutomations(1)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/automations')
  })

  it('createAutomation should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1 } })
    await createAutomation(1, { name: 'Auto Assign' })
    expect(mockPost).toHaveBeenCalledWith('/projects/1/automations', { name: 'Auto Assign' })
  })

  it('updateAutomation should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 5 } })
    await updateAutomation(1, 5, { is_enabled: false })
    expect(mockPut).toHaveBeenCalledWith('/projects/1/automations/5', { is_enabled: false })
  })

  it('deleteAutomation should DELETE', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await deleteAutomation(1, 5)
    expect(mockDelete).toHaveBeenCalledWith('/projects/1/automations/5')
  })

  it('toggleAutomationRule should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 5, is_enabled: false } })
    await toggleAutomationRule(1, 5, false)
    expect(mockPut).toHaveBeenCalledWith('/projects/1/automations/5', { is_enabled: false })
  })
})

describe('State Transitions API', () => {
  it('listStateTransitions should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listStateTransitions(1, 1)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/workflows/1/transitions')
  })

  it('createStateTransition should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1 } })
    await createStateTransition(1, 1, { source_state_id: 1, target_state_id: 2 })
    expect(mockPost).toHaveBeenCalledWith('/projects/1/workflows/1/transitions', { source_state_id: 1, target_state_id: 2 })
  })
})

describe('Templates API', () => {
  it('listAutomationTemplates should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listAutomationTemplates()
    expect(mockGet).toHaveBeenCalledWith('/automation-templates')
  })
})

describe('workflowApi export', () => {
  it('should have expected method count', () => {
    // 17 exported methods
    expect(Object.keys(workflowApi).length).toBeGreaterThanOrEqual(15)
  })
})
