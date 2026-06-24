import { test, expect } from '@playwright/test'

test.describe('后端API集成测试', () => {
  let token: string = ''

  test.beforeAll(async ({ request }) => {
    const response = await request.post('http://localhost:8000/api/v1/auth/login', {
      data: { email: 'demo@example.com', password: 'demo1234' }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    token = data.access_token
  })

  test('认证API - 登录', async ({ request }) => {
    const response = await request.post('http://localhost:8000/api/v1/auth/login', {
      data: { email: 'demo@example.com', password: 'demo1234' }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(data.access_token).toBeDefined()
    expect(data.token_type).toBe('bearer')
  })

  test('认证API - 错误密码', async ({ request }) => {
    const response = await request.post('http://localhost:8000/api/v1/auth/login', {
      data: { email: 'demo@example.com', password: 'wrong' }
    })
    expect(response.status()).toBe(401)
  })

  test('工作空间API - 获取列表', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/workspaces', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(Array.isArray(data)).toBe(true)
    expect(data.length).toBeGreaterThan(0)
  })

  test('项目API - 获取项目列表', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/projects?workspace_id=1', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(Array.isArray(data)).toBe(true)
  })

  test('工作项API - 获取工作项列表', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/issues?project_id=1', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(Array.isArray(data)).toBe(true)
  })

  test('工作项API - 获取单个工作项', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/issues/1', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(data.id).toBe(1)
    expect(data.name).toBeDefined()
  })

  test('项目设置API - 获取状态列表', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/projects/1/settings/states', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(Array.isArray(data)).toBe(true)
    expect(data.length).toBeGreaterThan(0)
  })

  test('项目设置API - 获取标签列表', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/projects/1/settings/labels', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(Array.isArray(data)).toBe(true)
  })

  test('模块API - 获取模块列表', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/modules?project_id=1&workspace_id=1', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(Array.isArray(data)).toBe(true)
  })

  test('周期API - 获取周期列表', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/projects/1/cycles', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(data.items).toBeDefined()
    expect(Array.isArray(data.items)).toBe(true)
  })

  test('自定义字段API - 获取字段列表', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/custom-fields?workspace_id=1', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(Array.isArray(data)).toBe(true)
  })

  test('工作项类型API - 获取类型列表', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/issue-types?workspace_id=1', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(Array.isArray(data)).toBe(true)
  })

  test('自动化规则API - 获取规则列表', async ({ request }) => {
    const response = await request.get('http://localhost:8000/api/v1/projects/1/automations', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(response.ok()).toBeTruthy()
    const data = await response.json()
    expect(Array.isArray(data)).toBe(true)
  })

})
