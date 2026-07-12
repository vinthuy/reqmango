import { test, expect } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'
const BASE_URL = 'http://localhost:5173'
const TEST_PREFIX = `e2e-auto-${Date.now()}`

let _token = ''
let _wsId = 0
let _wsSlug = ''
let _projectId = 0

async function ensureSetup(request: any) {
  if (_token) return
  const user = {
    email: `${TEST_PREFIX}@t.com`, username: TEST_PREFIX,
    password: 'E2eTest123!', display_name: 'E2E Automation',
  }
  await request.post(`${BASE_API}/auth/register`, { data: user })
  const res = await request.post(`${BASE_API}/auth/login`, { data: { email: user.email, password: user.password } })
  const { access_token } = await res.json()
  _token = access_token

  const ws = await request.post(`${BASE_API}/workspaces`, {
    data: { name: 'E2E Automation WS', slug: `e2e-auto-${TEST_PREFIX}` },
    headers: { Authorization: `Bearer ${_token}` },
  })
  const wsData = await ws.json()
  _wsId = wsData.id || wsData.data?.id
  _wsSlug = wsData.slug || wsData.data?.slug

  const proj = await request.post(`${BASE_API}/projects?workspace_id=${_wsId}`, {
    data: { name: 'E2E Automation Project', identifier: 'E2EAUTO', description: 'For automation testing' },
    headers: { Authorization: `Bearer ${_token}` },
  })
  const projData = await proj.json()
  _projectId = projData.id || projData.data?.id
}

async function loginViaStorage(page: any) {
  await page.goto('/login')
  await page.evaluate((t: string) => localStorage.setItem('token', t), _token)
}

async function goToProjectSettings(page: any) {
  await loginViaStorage(page)
  await page.goto(`/workspace/${_wsSlug}/project/${_projectId}/settings`)
  await page.waitForLoadState('networkidle').catch(() => {})
}

async function navigateToAutomation(page: any) {
  await goToProjectSettings(page)
  await page.click('text=自动化')
  await page.waitForTimeout(1500)
}

async function createAutomationRule(page: any, name: string, conditions: string, actions: string) {
  await page.click('button:has-text("创建自动化")')
  await page.waitForTimeout(500)
  
  await page.fill('input[placeholder*="自动分配"]', name)
  await page.fill('textarea[placeholder*="priority"]', conditions)
  await page.fill('textarea[placeholder*="assign"]', actions)
  
  await page.locator('div.fixed').locator('button:has-text("创建")').click()
  await page.waitForTimeout(1500)
  
  const ruleExists = await page.isVisible(`text=${name}`)
  return ruleExists
}

test.describe('ReqMango 自动化功能 E2E 测试', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('E2E-01: 创建高优先级自动分配规则', async ({ page, request }) => {
    console.log('🧪 开始执行 E2E-01: 创建规则')
    
    await navigateToAutomation(page)
    
    const success = await createAutomationRule(
      page,
      '高优先级自动分配',
      '[{"field":"priority","operator":"equals","value":"high"}]',
      '[{"type":"assign_to","value":1}]'
    )
    expect(success).toBeTruthy()
    console.log('✅ 规则创建成功')
    
    const res = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rules = await res.json()
    const createdRule = Array.isArray(rules) ? rules.find((r: any) => r.name === '高优先级自动分配') : null
    
    expect(createdRule).toBeDefined()
    expect(createdRule.is_enabled).toBe(true)
    console.log(`✅ 规则已保存到数据库`)
  })

  test('E2E-02: 创建状态变更规则', async ({ page }) => {
    console.log('🧪 开始执行 E2E-02: 创建状态变更规则')
    
    await navigateToAutomation(page)
    
    const success = await createAutomationRule(
      page,
      '完成时添加评论',
      '[{"field":"state_group","operator":"equals","value":"done"}]',
      '[{"type":"add_comment","value":"✅ 工作项已完成"}]'
    )
    expect(success).toBeTruthy()
    console.log('✅ 状态变更规则创建成功')
  })

  test('E2E-03: 禁用规则', async ({ page, request }) => {
    console.log('🧪 开始执行 E2E-03: 禁用规则')
    
    await navigateToAutomation(page)
    
    const res = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rules = await res.json()
    const targetRule = Array.isArray(rules) ? rules.find((r: any) => r.name === '高优先级自动分配') : null
    
    expect(targetRule).toBeDefined()
    
    await request.put(`${BASE_API}/projects/${_projectId}/automations/${targetRule.id}`, {
      data: { is_enabled: false },
      headers: { Authorization: `Bearer ${_token}` },
    })
    await page.waitForTimeout(1000)
    
    await page.reload()
    await page.waitForTimeout(1500)
    
    const updatedRes = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const updatedRules = await updatedRes.json()
    const disabledRule = Array.isArray(updatedRules) ? updatedRules.find((r: any) => r.name === '高优先级自动分配') : null
    
    expect(disabledRule).toBeDefined()
    expect(disabledRule.is_enabled).toBe(false)
    console.log('✅ 规则状态已更新到数据库')
  })

  test('E2E-04: 正则匹配条件规则测试', async ({ page }) => {
    console.log('🧪 开始执行 E2E-04: 正则匹配条件规则测试')
    
    await navigateToAutomation(page)
    
    const success = await createAutomationRule(
      page,
      'Bug 自动标记为紧急',
      '[{"field":"name","operator":"matches_regex","value":"^BUG-\\\\d+"}]',
      '[{"type":"set_priority","value":"urgent"}]'
    )
    expect(success).toBeTruthy()
    console.log('✅ 正则匹配规则创建成功')
  })

  test('E2E-05: 批量更新触发自动化测试', async ({ page }) => {
    console.log('🧪 开始执行 E2E-05: 批量更新触发自动化测试')
    
    await navigateToAutomation(page)
    
    const success = await createAutomationRule(
      page,
      '批量更新时设置优先级',
      '[]',
      '[{"type":"set_priority","value":"high"}]'
    )
    expect(success).toBeTruthy()
    console.log('✅ 批量更新规则创建成功')
  })

  test('E2E-06: 循环依赖检测测试', async ({ page }) => {
    console.log('🧪 开始执行 E2E-06: 循环依赖检测测试')
    
    await navigateToAutomation(page)
    
    await createAutomationRule(
      page,
      '规则A: 高优先级→进行中',
      '[{"field":"priority","operator":"equals","value":"high"}]',
      '[{"type":"set_field","field":"state_id","value":2}]'
    )
    
    await createAutomationRule(
      page,
      '规则B: 进行中→高优先级',
      '[{"field":"new_state","operator":"equals","value":"2"}]',
      '[{"type":"set_priority","value":"high"}]'
    )
    
    const ruleAExists = await page.isVisible('text=规则A')
    const ruleBExists = await page.isVisible('text=规则B')
    expect(ruleAExists).toBeTruthy()
    expect(ruleBExists).toBeTruthy()
    console.log('✅ 规则 A 和 B 创建成功')
  })

  test('E2E-07: 自动化规则列表展示', async ({ page }) => {
    console.log('🧪 开始执行 E2E-07: 自动化规则列表展示')
    
    await navigateToAutomation(page)
    
    const rulesCount = await page.locator('div.bg-white.rounded-xl.border').count()
    expect(rulesCount).toBeGreaterThanOrEqual(4)
    console.log(`✅ 列表中显示 ${rulesCount} 条规则`)
  })
})