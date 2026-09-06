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

import {
  workflowApi,
  listWorkspaceWorkflows, listWorkspaceAutomations,
  createWorkspaceAutomation, updateWorkspaceAutomation, deleteWorkspaceAutomation,
  createWorkspaceWorkflow, deleteWorkspaceWorkflow,
  addWorkspaceTransition, deleteWorkspaceTransition,
  listStateTransitions, createStateTransition, updateStateTransition, deleteStateTransition,
  listAutomationRules, toggleAutomationRule, listAutomationTemplates,
  budgetApi, slaApi, decisionApi,
} from './workflow'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('Workflow CRUD API', () => {
  it('workflowApi.list should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await workflowApi.list(1)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/workflows')
  })

  it('workflowApi.create should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1, name: 'Default' } })
    await workflowApi.create(1, { name: 'Default' })
    expect(mockPost).toHaveBeenCalledWith('/projects/1/workflows', { name: 'Default' })
  })

  it('workflowApi.get should GET', async () => {
    mockGet.mockResolvedValue({ data: { id: 5 } })
    await workflowApi.get(1, 5)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/workflows/5')
  })

  it('workflowApi.update should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 5 } })
    await workflowApi.update(1, 5, { name: 'Updated' })
    expect(mockPut).toHaveBeenCalledWith('/projects/1/workflows/5', { name: 'Updated' })
  })

  it('workflowApi.delete should DELETE', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await workflowApi.delete(1, 5)
    expect(mockDelete).toHaveBeenCalledWith('/projects/1/workflows/5')
  })
})

describe('Workflow Run API', () => {
  it('workflowApi.execute should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1 } })
    await workflowApi.execute(1, 5, 10)
    expect(mockPost).toHaveBeenCalledWith('/projects/1/workflows/5/execute', { issue_id: 10 })
  })

  it('workflowApi.listRuns should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await workflowApi.listRuns(1, 5)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/workflows/5/runs')
  })

  it('workflowApi.getRun should GET', async () => {
    mockGet.mockResolvedValue({ data: { id: 7 } })
    await workflowApi.getRun(1, 5, 7)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/workflows/5/runs/7')
  })

  it('workflowApi.cancelRun should POST', async () => {
    mockPost.mockResolvedValue({ data: null })
    await workflowApi.cancelRun(1, 5, 7)
    expect(mockPost).toHaveBeenCalledWith('/projects/1/workflows/5/runs/7/cancel')
  })
})

describe('Workflow Node/Edge API', () => {
  it('workflowApi.addNode should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1 } })
    await workflowApi.addNode(1, 5, { agent_id: 3, name: 'N' })
    expect(mockPost).toHaveBeenCalledWith('/projects/1/workflows/5/nodes', { agent_id: 3, name: 'N' })
  })

  it('workflowApi.updateNode should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 9 } })
    await workflowApi.updateNode(1, 5, 9, { name: 'Updated' })
    expect(mockPut).toHaveBeenCalledWith('/projects/1/workflows/5/nodes/9', { name: 'Updated' })
  })

  it('workflowApi.deleteNode should DELETE', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await workflowApi.deleteNode(1, 5, 9)
    expect(mockDelete).toHaveBeenCalledWith('/projects/1/workflows/5/nodes/9')
  })

  it('workflowApi.addEdge should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1 } })
    await workflowApi.addEdge(1, 5, { source_node_id: 1, target_node_id: 2 })
    expect(mockPost).toHaveBeenCalledWith('/projects/1/workflows/5/edges', { source_node_id: 1, target_node_id: 2 })
  })

  it('workflowApi.updateEdge should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 9 } })
    await workflowApi.updateEdge(1, 5, 9, { condition: 'always' })
    expect(mockPut).toHaveBeenCalledWith('/projects/1/workflows/5/edges/9', { condition: 'always' })
  })

  it('workflowApi.deleteEdge should DELETE', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await workflowApi.deleteEdge(1, 5, 9)
    expect(mockDelete).toHaveBeenCalledWith('/projects/1/workflows/5/edges/9')
  })
})

describe('Workspace-level API', () => {
  it('listWorkspaceWorkflows should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listWorkspaceWorkflows(1)
    expect(mockGet).toHaveBeenCalledWith('/workspaces/1/workflows')
  })

  it('listWorkspaceAutomations should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listWorkspaceAutomations(1)
    expect(mockGet).toHaveBeenCalledWith('/workspaces/1/automations')
  })

  it('createWorkspaceAutomation should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1 } })
    await createWorkspaceAutomation(1, { name: 'Auto Assign' })
    expect(mockPost).toHaveBeenCalledWith('/workspaces/1/automations', { name: 'Auto Assign' })
  })

  it('updateWorkspaceAutomation should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 5 } })
    await updateWorkspaceAutomation(1, 5, { is_enabled: false })
    expect(mockPut).toHaveBeenCalledWith('/workspaces/1/automations/5', { is_enabled: false })
  })

  it('deleteWorkspaceAutomation should DELETE', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await deleteWorkspaceAutomation(1, 5)
    expect(mockDelete).toHaveBeenCalledWith('/workspaces/1/automations/5')
  })

  it('createWorkspaceWorkflow should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1 } })
    await createWorkspaceWorkflow(1, { name: 'Default' })
    expect(mockPost).toHaveBeenCalledWith('/workspaces/1/workflows', { name: 'Default' })
  })

  it('deleteWorkspaceWorkflow should DELETE', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await deleteWorkspaceWorkflow(1, 5)
    expect(mockDelete).toHaveBeenCalledWith('/workspaces/1/workflows/5')
  })

  it('addWorkspaceTransition should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1 } })
    await addWorkspaceTransition(1, 5, { source_node_id: 1, target_node_id: 2 })
    expect(mockPost).toHaveBeenCalledWith('/workspaces/1/workflows/5/edges', { source_node_id: 1, target_node_id: 2 })
  })

  it('deleteWorkspaceTransition should DELETE', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await deleteWorkspaceTransition(1, 5, 9)
    expect(mockDelete).toHaveBeenCalledWith('/workspaces/1/workflows/5/edges/9')
  })
})

describe('State Transitions API', () => {
  it('listStateTransitions should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listStateTransitions(1, 5)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/workflows/5/transitions')
  })

  it('createStateTransition should POST', async () => {
    mockPost.mockResolvedValue({ data: { id: 1 } })
    await createStateTransition(1, 5, { source_state_id: 1, target_state_id: 2 })
    expect(mockPost).toHaveBeenCalledWith('/projects/1/workflows/5/transitions', { source_state_id: 1, target_state_id: 2 })
  })

  it('updateStateTransition should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 9 } })
    await updateStateTransition(1, 5, 9, { name: 'T' })
    expect(mockPut).toHaveBeenCalledWith('/projects/1/workflows/5/transitions/9', { name: 'T' })
  })

  it('deleteStateTransition should DELETE', async () => {
    mockDelete.mockResolvedValue({ data: null })
    await deleteStateTransition(1, 5, 9)
    expect(mockDelete).toHaveBeenCalledWith('/projects/1/workflows/5/transitions/9')
  })
})

describe('Automation Rules API', () => {
  it('listAutomationRules should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listAutomationRules(1)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/automation-rules')
  })

  it('toggleAutomationRule should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 5, is_enabled: false } })
    await toggleAutomationRule(1, 5, false)
    expect(mockPut).toHaveBeenCalledWith('/projects/1/automation-rules/5', { is_enabled: false })
  })

  it('listAutomationTemplates should GET', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await listAutomationTemplates()
    expect(mockGet).toHaveBeenCalledWith('/automation-templates')
  })
})

describe('Budget API', () => {
  it('budgetApi.get should GET', async () => {
    mockGet.mockResolvedValue({ data: { id: 1 } })
    await budgetApi.get(1)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/budget')
  })

  it('budgetApi.update should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 1 } })
    await budgetApi.update(1, { monthly_budget: 100 })
    expect(mockPut).toHaveBeenCalledWith('/projects/1/budget', { monthly_budget: 100 })
  })
})

describe('SLA API', () => {
  it('slaApi.get should GET', async () => {
    mockGet.mockResolvedValue({ data: { id: 1 } })
    await slaApi.get(1)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/sla')
  })

  it('slaApi.update should PUT', async () => {
    mockPut.mockResolvedValue({ data: { id: 1 } })
    await slaApi.update(1, { normal_task_max: 300 })
    expect(mockPut).toHaveBeenCalledWith('/projects/1/sla', { normal_task_max: 300 })
  })
})

describe('Decision API', () => {
  it('decisionApi.list should GET with limit param', async () => {
    mockGet.mockResolvedValue({ data: [] })
    await decisionApi.list(1)
    expect(mockGet).toHaveBeenCalledWith('/projects/1/decisions', { params: { limit: 100 } })
  })
})

describe('workflowApi export', () => {
  it('should have expected method count', () => {
    // 15 methods: list/create/get/update/delete/execute/listRuns/getRun/cancelRun + 3 node + 3 edge
    expect(Object.keys(workflowApi).length).toBe(15)
  })
})
