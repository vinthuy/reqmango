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

  // NOTE (BUG-58): state-transition management for agent workflows is not
  // implemented — POST/PUT/DELETE /workflows/:id/transitions answer with canned
  // messages, there is no GET, and `state_transitions` still references the legacy
  // `workflows` table. No approval-type transition can therefore be created through
  // the API, so the happy path (create approval → approve/reject) is unreachable.
  // This case pins the documented behaviour: the request is rejected, not crashed.
  test('API - 缺少合法审批转换时创建审批返回 400（BUG-58）', async ({ request }) => {
    const statesResponse = await request.get(`http://localhost:8000/api/v1/projects/${projectId}/settings/states`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    const statesBody = await statesResponse.json()
    const states = statesBody.data || statesBody
    const stateId = Array.isArray(states) && states.length > 0 ? states[0].id : undefined

    const issueResponse = await request.post(`http://localhost:8000/api/v1/issues?project_id=${projectId}&workspace_id=${workspaceId}`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        name: 'API测试审批工作项',
        description_html: '<p>测试审批API</p>',
        priority: 'medium',
        state_id: stateId,
      },
    })
    expect(issueResponse.ok()).toBeTruthy()
    const issue = await issueResponse.json()

    const approvalResponse = await request.post(`http://localhost:8000/api/v1/issues/${issue.id}/approvals`, {
      headers: { Authorization: `Bearer ${authToken}` },
      data: {
        transition_id: 1,
        approver_ids: [1],
        reason: 'API测试审批',
      },
    })
    expect(approvalResponse.status()).toBe(400)
    console.log('审批创建按预期被拒绝（无合法审批转换）:', approvalResponse.status())
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