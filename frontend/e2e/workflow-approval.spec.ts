import { test, expect } from '@playwright/test'

const ADMIN_USER = 'admin@reqmango.com'
const ADMIN_PASS = 'demo1234'
const WORKSPACE_SLUG = 'reqmango-dev'
const PROJECT_IDENTIFIER = 'CORE'

let authToken: string
let projectId: number
let workspaceId: number

test.describe('工作流审批功能', () => {
  test.beforeAll(async ({ request }) => {
    const loginResponse = await request.post('http://localhost:8000/api/v1/auth/login', {
      data: { email: ADMIN_USER, password: ADMIN_PASS },
    })
    const loginData = await loginResponse.json()
    authToken = loginData.access_token

    const workspacesResponse = await request.get('http://localhost:8000/api/v1/workspaces', {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    const workspacesData = await workspacesResponse.json()
    const workspaces = Array.isArray(workspacesData) ? workspacesData : (workspacesData?.data || [])
    const workspace = workspaces.find((w: any) => w.slug === WORKSPACE_SLUG)
    workspaceId = workspace.id

    const projectsResponse = await request.get(`http://localhost:8000/api/v1/projects?workspace_id=${workspaceId}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    const projectsData = await projectsResponse.json()
    const projects = Array.isArray(projectsData) ? projectsData : (projectsData?.data || [])
    const project = projects.find((p: any) => p.identifier === PROJECT_IDENTIFIER)
    projectId = project.id
  })

  test('API - 审批计数接口', async ({ request }) => {
    const response = await request.get(`http://localhost:8000/api/v1/workspaces/${WORKSPACE_SLUG}/approvals/count`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    console.log('审批计数:', data)
    expect(data).toHaveProperty('pending_count')
  })

  test('API - 工作空间审批列表接口', async ({ request }) => {
    const response = await request.get(`http://localhost:8000/api/v1/workspaces/${WORKSPACE_SLUG}/approvals`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    console.log('工作空间审批列表:', data)
  })

  test('API - 项目审批列表接口', async ({ request }) => {
    const response = await request.get(`http://localhost:8000/api/v1/projects/${projectId}/approvals?workspace_id=${workspaceId}`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    console.log('项目审批列表:', data)
  })

  test('API - 创建审批并批准', async ({ request }) => {
    const issueResponse = await request.post(`http://localhost:8000/api/v1/issues?project_id=${projectId}&workspace_id=${workspaceId}`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        name: 'API测试审批工作项',
        description_html: '<p>测试审批API</p>',
        priority: 'medium',
        state_id: 4,
      },
    })
    expect(issueResponse.ok()).toBeTruthy()
    const issue = await issueResponse.json()
    console.log('创建Issue成功:', issue.id)

    const approvalResponse = await request.post(`http://localhost:8000/api/v1/issues/${issue.id}/approvals`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        transition_id: 1,
        approver_ids: [1],
        reason: 'API测试审批',
      },
    })
    expect(approvalResponse.ok()).toBeTruthy()
    const approval = await approvalResponse.json()
    console.log('创建审批成功:', approval.id)

    const decideResponse = await request.post(`http://localhost:8000/api/v1/approvals/${approval.id}/decide`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { decision: 'approved', note: 'API批准' },
    })
    expect(decideResponse.ok()).toBeTruthy()
    console.log('审批批准成功')
  })

  test('API - 创建审批并拒绝', async ({ request }) => {
    const issueResponse = await request.post(`http://localhost:8000/api/v1/issues?project_id=${projectId}&workspace_id=${workspaceId}`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        name: 'API测试拒绝审批工作项',
        description_html: '<p>测试拒绝审批API</p>',
        priority: 'medium',
        state_id: 4,
      },
    })
    const issue = await issueResponse.json()

    const approvalResponse = await request.post(`http://localhost:8000/api/v1/issues/${issue.id}/approvals`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        transition_id: 1,
        approver_ids: [1],
        reason: 'API测试拒绝审批',
      },
    })
    const approval = await approvalResponse.json()

    const decideResponse = await request.post(`http://localhost:8000/api/v1/approvals/${approval.id}/decide`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: { decision: 'rejected', note: 'API拒绝' },
    })
    expect(decideResponse.ok()).toBeTruthy()
    console.log('审批拒绝成功')
  })

  test('审批中心页面 - 查看待审批列表', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    await page.evaluate((token) => {
      localStorage.setItem('token', token)
    }, authToken)
    
    await page.reload()
    await page.waitForLoadState('networkidle')
    
    await page.goto(`/workspace/${WORKSPACE_SLUG}/approvals`)
    await page.waitForLoadState('networkidle')
    
    await expect(page.locator('h1:text("审批中心")')).toBeVisible()
    console.log('审批中心页面加载成功')
  })

  test('TopBar - 审批图标显示', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    await page.evaluate((token) => {
      localStorage.setItem('token', token)
    }, authToken)
    
    await page.reload()
    await page.waitForLoadState('networkidle')
    
    await page.goto(`/workspace/${WORKSPACE_SLUG}`)
    await page.waitForLoadState('networkidle')
    
    const badge = page.locator('.approval-badge')
    await expect(badge).toBeVisible()
    console.log('审批图标显示成功')
  })
})