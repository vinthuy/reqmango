import { test, expect } from '@playwright/test'

const ADMIN_USER = 'admin@reqmango.com'
const ADMIN_PASS = 'demo1234'
const WORKSPACE_SLUG = 'reqmango-dev'

let authToken: string

test.describe('审批功能 - API测试', () => {
  test.beforeAll(async ({ request }) => {
    const loginResponse = await request.post('http://localhost:8000/api/v1/auth/login', {
      data: { email: ADMIN_USER, password: ADMIN_PASS },
    })
    const loginData = await loginResponse.json()
    authToken = loginData.access_token
    console.log('Token obtained:', authToken ? 'yes' : 'no')
  })

  test('API - 审批计数', async ({ request }) => {
    const response = await request.get(`http://localhost:8000/api/v1/workspaces/${WORKSPACE_SLUG}/approvals/count`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    console.log('审批计数:', data)
    expect(data).toHaveProperty('pending_count')
  })

  test('API - 审批列表', async ({ request }) => {
    const response = await request.get(`http://localhost:8000/api/v1/workspaces/${WORKSPACE_SLUG}/approvals`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    console.log('审批列表:', data.length, '条记录')
    expect(Array.isArray(data)).toBeTruthy()
  })

  test('API - 项目审批列表', async ({ request }) => {
    const workspacesResponse = await request.get('http://localhost:8000/api/v1/workspaces', {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    const workspaces = await workspacesResponse.json()
    const workspace = workspaces.find((w: any) => w.slug === WORKSPACE_SLUG)
    
    const projectsResponse = await request.get('http://localhost:8000/api/v1/projects', {
      headers: { Authorization: `Bearer ${authToken}` },
      params: { workspace_id: workspace.id },
    })
    const projects = await projectsResponse.json()
    const project = projects.find((p: any) => p.identifier === 'CORE')
    
    const response = await request.get(`http://localhost:8000/api/v1/projects/${project.id}/approvals`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    console.log('项目审批列表:', data.length, '条记录')
    expect(Array.isArray(data)).toBeTruthy()
  })

  test('API - 获取审批详情', async ({ request }) => {
    const response = await request.get(`http://localhost:8000/api/v1/workspaces/${WORKSPACE_SLUG}/approvals`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    const approvals = await response.json()
    
    if (approvals.length > 0) {
      const approvalId = approvals[0].id
      const detailResponse = await request.get(`http://localhost:8000/api/v1/approvals/${approvalId}`, {
        headers: { Authorization: `Bearer ${authToken}` },
      })
      expect(detailResponse.ok()).toBeTruthy()
      const data = await detailResponse.json()
      console.log('审批详情:', data.id, data.status)
      expect(data).toHaveProperty('id')
      expect(data).toHaveProperty('status')
    } else {
      console.log('暂无审批记录，跳过详情测试')
    }
  })

  test('API - 审批决策（批准）', async ({ request }) => {
    const response = await request.get(`http://localhost:8000/api/v1/workspaces/${WORKSPACE_SLUG}/approvals`, {
      headers: { Authorization: `Bearer ${authToken}` },
      params: { status: 'pending' },
    })
    const approvals = await response.json()
    
    if (approvals.length > 0) {
      const approvalId = approvals[0].id
      const decideResponse = await request.post(`http://localhost:8000/api/v1/approvals/${approvalId}/decide`, {
        headers: { Authorization: `Bearer ${authToken}` },
        data: { decision: 'approved', note: 'API测试批准' },
      })
      expect(decideResponse.ok()).toBeTruthy()
      console.log('审批决策成功')
    } else {
      console.log('暂无待审批记录，跳过决策测试')
    }
  })

  test('API - 创建审批转换', async ({ request }) => {
    const workspacesResponse = await request.get('http://localhost:8000/api/v1/workspaces', {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    const workspaces = await workspacesResponse.json()
    const workspace = workspaces.find((w: any) => w.slug === WORKSPACE_SLUG)
    
    const projectsResponse = await request.get('http://localhost:8000/api/v1/projects', {
      headers: { Authorization: `Bearer ${authToken}` },
      params: { workspace_id: workspace.id },
    })
    const projects = await projectsResponse.json()
    const project = projects.find((p: any) => p.identifier === 'CORE')
    
    const workflowsResponse = await request.get(`http://localhost:8000/api/v1/projects/${project.id}/workflows`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    const workflows = await workflowsResponse.json()
    const workflow = workflows.find((w: any) => w.name === 'Default Workflow')
    
    const transitionResponse = await request.post(`http://localhost:8000/api/v1/projects/${project.id}/workflows/${workflow.id}/transitions`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        name: 'E2E审批转换',
        from_state_id: 4,
        to_state_id: 5,
        rule_type: 'approval',
        approver_ids: '[1]',
        approve_target_state_id: 5,
        reject_target_state_id: 3,
        approval_mode: 'any',
      },
    })
    if (!transitionResponse.ok()) {
      const errorText = await transitionResponse.text()
      console.log('创建转换失败:', transitionResponse.status(), errorText)
    }
    expect(transitionResponse.ok()).toBeTruthy()
    const transition = await transitionResponse.json()
    console.log('创建审批转换成功:', transition.id, transition.rule_type)
    expect(transition.rule_type).toBe('approval')
  })
})
