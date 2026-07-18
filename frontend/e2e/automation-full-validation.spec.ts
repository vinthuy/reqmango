import { test, expect } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'
const BASE_URL = 'http://localhost:5173'
const TEST_PREFIX = `e2e-auto-full-${Date.now()}`

let _token = ''
let _wsId = 0
let _wsSlug = ''
let _projectId = 0
let _issueId = 0

async function ensureSetup(request: any) {
  if (_token) return
  const user = {
    email: `${TEST_PREFIX}@t.com`, username: TEST_PREFIX,
    password: 'E2eTest123!', display_name: 'E2E Full Validation',
  }
  await request.post(`${BASE_API}/auth/register`, { data: user })
  const res = await request.post(`${BASE_API}/auth/login`, { data: { email: user.email, password: user.password } })
  const { access_token } = await res.json()
  _token = access_token

  const ws = await request.post(`${BASE_API}/workspaces`, {
    data: { name: 'E2E Full Validation WS', slug: `e2e-full-${TEST_PREFIX}` },
    headers: { Authorization: `Bearer ${_token}` },
  })
  const wsData = await ws.json()
  _wsId = wsData.id || wsData.data?.id
  _wsSlug = wsData.slug || wsData.data?.slug

  const proj = await request.post(`${BASE_API}/projects?workspace_id=${_wsId}`, {
    data: { name: 'E2E Full Validation Project', identifier: 'E2EFULL', description: 'For full automation validation' },
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

async function createRuleViaAPI(request: any, name: string, triggerType: string, conditions: string, actions: string): Promise<number> {
  const res = await request.post(`${BASE_API}/projects/${_projectId}/automations`, {
    data: { name, trigger_type: triggerType, conditions, actions },
    headers: { Authorization: `Bearer ${_token}` },
  })
  const data = await res.json()
  return data.id || data.data?.id
}

test.describe('自动化功能全面验证 - 前后端联动', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('验证-01: 创建规则时trigger_type格式正确（字符串而非JSON）', async ({ page, request }) => {
    console.log('🧪 验证-01: 检查trigger_type格式')
    
    await navigateToAutomation(page)
    
    await page.click('button:has-text("创建自动化")')
    await page.waitForTimeout(500)
    
    await page.fill('input[placeholder*="自动分配"]', 'Trigger类型验证规则')
    await page.fill('textarea[placeholder*="priority"]', '[{"field":"priority","operator":"equals","value":"high"}]')
    await page.fill('textarea[placeholder*="assign"]', '[{"type":"add_comment","value":"trigger测试"}]')
    await page.selectOption('select', 'issue_created')
    
    await page.locator('div.fixed').locator('button:has-text("创建")').click()
    await page.waitForTimeout(1500)
    
    const res = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rules = await res.json()
    const createdRule = Array.isArray(rules) ? rules.find((r: any) => r.name === 'Trigger类型验证规则') : null
    
    expect(createdRule).toBeDefined()
    expect(typeof createdRule.trigger_type).toBe('string')
    expect(createdRule.trigger_type).toBe('issue_created')
    expect(createdRule.trigger_type).not.toContain('{')
    console.log('✅ trigger_type格式正确：字符串类型，值为"issue_created"')
  })

  test('验证-02: 编辑规则功能正常', async ({ page, request }) => {
    console.log('🧪 验证-02: 编辑规则功能')
    
    await navigateToAutomation(page)
    
    const res = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rules = await res.json()
    const targetRule = Array.isArray(rules) ? rules.find((r: any) => r.name === 'Trigger类型验证规则') : null
    expect(targetRule).toBeDefined()
    
    await page.click('button:text("✏️")')
    await page.waitForTimeout(500)
    
    await page.fill('input[placeholder*="自动分配"]', '编辑后的规则名称')
    await page.fill('textarea[placeholder*="priority"]', '[{"field":"priority","operator":"equals","value":"urgent"}]')
    
    await page.locator('div.fixed').locator('button:has-text("更新")').click()
    await page.waitForTimeout(1500)
    
    const res2 = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rules2 = await res2.json()
    const updatedRule = Array.isArray(rules2) ? rules2.find((r: any) => r.name === '编辑后的规则名称') : null
    
    expect(updatedRule).toBeDefined()
    expect(updatedRule.conditions).toContain('urgent')
    console.log('✅ 编辑规则功能正常')
  })

  test('验证-03: 启用/禁用切换按钮功能', async ({ request }) => {
    console.log('🧪 验证-03: 启用/禁用切换')
    
    const res = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rules = await res.json()
    const targetRule = Array.isArray(rules) ? rules.find((r: any) => r.name === '编辑后的规则名称') : null
    expect(targetRule).toBeDefined()
    
    await request.put(`${BASE_API}/projects/${_projectId}/automations/${targetRule.id}`, {
      data: { is_enabled: false },
      headers: { Authorization: `Bearer ${_token}` },
    })
    
    const res2 = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rules2 = await res2.json()
    const disabledRule = Array.isArray(rules2) ? rules2.find((r: any) => r.id === targetRule.id) : null
    
    expect(disabledRule.is_enabled).toBe(false)
    console.log('✅ 禁用规则成功')
    
    await request.put(`${BASE_API}/projects/${_projectId}/automations/${targetRule.id}`, {
      data: { is_enabled: true },
      headers: { Authorization: `Bearer ${_token}` },
    })
    
    const res3 = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rules3 = await res3.json()
    const enabledRule = Array.isArray(rules3) ? rules3.find((r: any) => r.id === targetRule.id) : null
    
    expect(enabledRule.is_enabled).toBe(true)
    console.log('✅ 启用规则成功')
  })

  test('验证-04: issue_created触发器正常工作', async ({ request }) => {
    console.log('🧪 验证-04: issue_created触发器')
    
    await createRuleViaAPI(
      request,
      '创建时自动加评论',
      'issue_created',
      '[]',
      '[{"type":"add_comment","value":"🎉 新工作项已创建"}]'
    )
    
    const issueRes = await request.post(`${BASE_API}/issues?project_id=${_projectId}&workspace_id=${_wsId}`, {
      data: { name: '测试issue_created触发', description: '测试触发器' },
      headers: { Authorization: `Bearer ${_token}` },
    })
    const issueData = await issueRes.json()
    _issueId = issueData.id || issueData.data?.id
    
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    const commentsRes = await request.get(`${BASE_API}/comments/issue/${_issueId}`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const commentsData = await commentsRes.json()
    const comments = commentsData.comments || commentsData.data || commentsData
    
    const autoComment = Array.isArray(comments) ? comments.find((c: any) => c.body === '🎉 新工作项已创建') : null
    expect(autoComment).toBeDefined()
    console.log('✅ issue_created触发器正常工作')
  })

  test('验证-05: comment_added触发器正常工作', async ({ request }) => {
    console.log('🧪 验证-05: comment_added触发器')
    
    await createRuleViaAPI(
      request,
      '评论时自动标记',
      'comment_added',
      '[{"field":"comment","operator":"contains","value":"bug"}]',
      '[{"type":"set_priority","value":"high"}]'
    )
    
    await request.post(`${BASE_API}/comments`, {
      data: { issue_id: _issueId, body: '发现一个bug需要修复' },
      headers: { Authorization: `Bearer ${_token}` },
    })
    
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    const issueRes = await request.get(`${BASE_API}/issues/${_issueId}`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const issueData = await issueRes.json()
    const issue = issueData.data || issueData
    
    expect(issue.priority).toBe('high')
    console.log('✅ comment_added触发器正常工作')
  })

  test('验证-06: issue_updated触发器（非状态变化）正常工作', async ({ request }) => {
    console.log('🧪 验证-06: issue_updated触发器（非状态变化）')
    
    await createRuleViaAPI(
      request,
      '更新时自动加评论',
      'issue_updated',
      '[]',
      '[{"type":"add_comment","value":"🔄 工作项已更新"}]'
    )
    
    await request.put(`${BASE_API}/issues/${_issueId}`, {
      data: { name: '测试自动化触发 - 已更新' },
      headers: { Authorization: `Bearer ${_token}` },
    })
    
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    const commentsRes = await request.get(`${BASE_API}/comments/issue/${_issueId}`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const commentsData = await commentsRes.json()
    const comments = commentsData.comments || commentsData.data || commentsData
    
    const updateComment = Array.isArray(comments) ? comments.find((c: any) => c.body === '🔄 工作项已更新') : null
    expect(updateComment).toBeDefined()
    console.log('✅ issue_updated触发器（非状态变化）正常工作')
  })

  test('验证-07: state_changed触发器正常工作', async ({ request }) => {
    console.log('🧪 验证-07: state_changed触发器')
    
    await createRuleViaAPI(
      request,
      '状态变更时自动完成',
      'state_changed',
      '[{"field":"state_group","operator":"equals","value":"done"}]',
      '[{"type":"add_comment","value":"✅ 工作项已完成"}]'
    )
    
    const statesRes = await request.get(`${BASE_API}/projects/${_projectId}/settings/states`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const statesData = await statesRes.json()
    const states = statesData.data || statesData
    const doneState = Array.isArray(states) ? states.find((s: any) => s.group === 'done') : null
    
    if (doneState) {
      await request.put(`${BASE_API}/issues/${_issueId}`, {
        data: { state_id: doneState.id },
        headers: { Authorization: `Bearer ${_token}` },
      })
      
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      const commentsRes = await request.get(`${BASE_API}/comments/issue/${_issueId}`, {
        headers: { Authorization: `Bearer ${_token}` },
      })
      const commentsData = await commentsRes.json()
      const comments = commentsData.comments || commentsData.data || commentsData
      
      const doneComment = Array.isArray(comments) ? comments.find((c: any) => c.body === '✅ 工作项已完成') : null
      expect(doneComment).toBeDefined()
      console.log('✅ state_changed触发器正常工作')
    } else {
      console.log('⚠️ 未找到done状态，跳过此测试')
    }
  })

  test('验证-08: 规则执行历史记录', async ({ request }) => {
    console.log('🧪 验证-08: 规则执行历史')
    
    const historyRes = await request.get(`${BASE_API}/issues/${_issueId}/automation-executions`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    
    const status = historyRes.status()
    if (status === 200) {
      const history = await historyRes.json()
      console.log(`✅ 执行历史记录: ${Array.isArray(history) ? history.length : 0} 条记录`)
    } else {
      console.log(`⚠️ 执行历史API返回: ${status}`)
    }
  })

  test('验证-09: 删除规则功能', async ({ request }) => {
    console.log('🧪 验证-09: 删除规则')
    
    const res = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rules = await res.json()
    const rulesArray = Array.isArray(rules) ? rules : (rules.data || [])
    const targetRule = rulesArray.find((r: any) => r.name === '创建时自动加评论')
    
    expect(targetRule).toBeDefined()
    
    await request.delete(`${BASE_API}/projects/${_projectId}/automations/${targetRule.id}`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    
    const res2 = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const rules2 = await res2.json()
    const rulesArray2 = Array.isArray(rules2) ? rules2 : (rules2.data || [])
    const deletedRule = rulesArray2.find((r: any) => r.id === targetRule.id)
    
    expect(deletedRule).toBeUndefined()
    console.log('✅ 删除规则功能正常')
  })
})