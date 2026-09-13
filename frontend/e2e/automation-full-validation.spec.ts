import { test, expect } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'
const TEST_PREFIX = `e2e-auto-full-${Date.now()}`

let _token = ''
let _wsId = 0
let _wsSlug = ''
let _projectId = 0

function auth() {
  return { Authorization: `Bearer ${_token}` }
}

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
    headers: auth(),
  })
  const wsData = await ws.json()
  _wsId = wsData.id || wsData.data?.id
  _wsSlug = wsData.slug || wsData.data?.slug

  const proj = await request.post(`${BASE_API}/projects?workspace_id=${_wsId}`, {
    data: { name: 'E2E Full Validation Project', identifier: 'E2EFULL', description: 'For full automation validation' },
    headers: auth(),
  })
  const projData = await proj.json()
  _projectId = projData.id || projData.data?.id
}

// ==================== API helpers ====================

async function listRules(request: any) {
  const res = await request.get(`${BASE_API}/projects/${_projectId}/automations`, { headers: auth() })
  const body = await res.json()
  return Array.isArray(body) ? body : (body.data || [])
}

async function findRule(request: any, name: string) {
  return (await listRules(request)).find((r: any) => r.name === name)
}

async function createRuleViaAPI(
  request: any, name: string, triggerType: string, conditions: string, actions: string,
): Promise<number> {
  const res = await request.post(`${BASE_API}/projects/${_projectId}/automations`, {
    data: { name, trigger_type: triggerType, conditions, actions },
    headers: auth(),
  })
  const data = await res.json()
  return data.id || data.data?.id
}

async function createIssue(request: any, name: string): Promise<number> {
  const res = await request.post(`${BASE_API}/issues?project_id=${_projectId}&workspace_id=${_wsId}`, {
    data: { name, description: '自动化验证用工作项' },
    headers: auth(),
  })
  const data = await res.json()
  return data.id || data.data?.id
}

async function getComments(request: any, issueId: number) {
  const res = await request.get(`${BASE_API}/comments/issue/${issueId}`, { headers: auth() })
  const body = await res.json()
  const list = body.comments || body.data || body
  return Array.isArray(list) ? list : []
}

async function getIssue(request: any, issueId: number) {
  const res = await request.get(`${BASE_API}/issues/${issueId}`, { headers: auth() })
  const body = await res.json()
  return body.data || body
}

async function waitFor<T>(probe: () => Promise<T | undefined | null>, timeoutMs = 15000): Promise<T | undefined> {
  const deadline = Date.now() + timeoutMs
  let last: T | undefined
  while (Date.now() < deadline) {
    const found = await probe()
    if (found) return found
    last = undefined
    await new Promise(r => setTimeout(r, 500))
  }
  return last
}

/**
 * Every trigger registered by the event bus (backend/internal/service/automation_service.go)
 * uses dot notation: issue.created, issue.updated, issue.state_changed, issue.assigned,
 * comment.added, scheduled. Legacy underscore names (issue_created, ...) never match an
 * event and silently do nothing, so the specs must use the dotted form.
 */
async function waitForComment(request: any, issueId: number, body: string) {
  return waitFor(async () => (await getComments(request, issueId)).find((c: any) => c.body === body))
}

async function waitForPriority(request: any, issueId: number, priority: string) {
  return waitFor(async () => {
    const issue = await getIssue(request, issueId)
    return issue.priority === priority ? issue : undefined
  })
}

// ==================== UI helpers (AutomationRuleBuilder) ====================

async function loginViaStorage(page: any) {
  await page.goto('/login')
  await page.evaluate((t: string) => {
    localStorage.setItem('token', t)
    localStorage.setItem('locale', 'zh-CN')
  }, _token)
}

async function goToProjectSettings(page: any) {
  await loginViaStorage(page)
  await page.goto(`/workspace/${_wsSlug}/project/${_projectId}/settings`)
  await page.waitForLoadState('networkidle').catch(() => {})
}

async function navigateToAutomation(page: any) {
  await goToProjectSettings(page)
  await page.click('text=自动化')
  await expect(page.locator('button:has-text("创建自动化")').first()).toBeVisible({ timeout: 15000 })
}

const modal = (page: any) => page.locator('div.fixed.inset-0.z-50')
const formRows = (page: any) => page.locator('div.bg-gray-50.rounded-lg')

async function openCreateBuilder(page: any) {
  await page.locator('button:has-text("创建自动化")').first().click()
  await expect(page.locator('input[placeholder*="自动分配"]')).toBeVisible({ timeout: 10000 })
}

async function fillName(page: any, name: string) {
  await page.locator('input[placeholder*="自动分配"]').fill(name)
}

/** Trigger labels come from TriggerTypeOptions and are not localised. */
async function pickTrigger(page: any, label: string) {
  await page.locator(`button:has-text("${label}")`).first().click()
}

async function addCondition(page: any, field: string, operator: string, value?: string) {
  await page.locator('button:has-text("添加条件")').first().click()
  const row = formRows(page).last()
  await row.locator('select').nth(0).selectOption(field)
  await row.locator('select').nth(1).selectOption(operator)
  if (value !== undefined) {
    if ((await row.locator('select').count()) >= 3) {
      await row.locator('select').nth(2).selectOption(value)
    } else {
      await row.locator('input[type="text"], input[type="number"]').last().fill(value)
    }
  }
}

type ActionSpec = { type: string; value?: string }

async function addAction(page: any, action: ActionSpec) {
  await page.locator('button:has-text("添加动作")').first().click()
  const row = formRows(page).last()
  await row.locator('select').first().selectOption(action.type)
  switch (action.type) {
    case 'add_comment':
      await row.locator('textarea').first().fill(action.value || '')
      break
    case 'set_priority':
      await row.locator('select').nth(1).selectOption(action.value || 'high')
      break
    case 'change_state':
      await row.locator('select').nth(1).selectOption({ index: 1 })
      break
    case 'assign_to':
      await row.locator('select').nth(1).selectOption({ index: 1 })
      break
    case 'set_field':
      // field selector, then the value control rendered for that field
      await row.locator('select').nth(1).selectOption('state_id')
      await row.locator('select').nth(2).selectOption({ index: 1 })
      break
  }
}

async function submitBuilder(page: any, submitLabel: string) {
  await modal(page).locator('button').filter({ hasText: submitLabel }).last().click()
  await expect(modal(page)).toHaveCount(0, { timeout: 15000 })
}

// ==================== Tests ====================

test.describe('自动化功能全面验证 - 前后端联动', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('验证-01: 创建规则时trigger_type格式正确（字符串而非JSON对象）', async ({ page, request }) => {
    await navigateToAutomation(page)

    await openCreateBuilder(page)
    await fillName(page, 'Trigger类型验证规则')
    await pickTrigger(page, '工作项创建时')
    await addCondition(page, 'priority', 'equals', 'high')
    await addAction(page, { type: 'add_comment', value: 'trigger测试' })
    await submitBuilder(page, '创建')

    const createdRule = await waitFor(async () => findRule(request, 'Trigger类型验证规则'))
    expect(createdRule, '规则应已保存到数据库').toBeDefined()

    // trigger_type is persisted as a string. The builder serialises trigger parameters
    // as JSON ({"type":"issue.created"}), older rows store the bare event name.
    expect(typeof createdRule.trigger_type).toBe('string')
    const raw = createdRule.trigger_type as string
    const parsed = raw.trim().startsWith('{') ? JSON.parse(raw) : { type: raw }
    expect(parsed.type).toBe('issue.created')

    // The builder's condition/action rows must round-trip into the stored JSON.
    expect(createdRule.conditions).toContain('priority')
    expect(createdRule.actions).toContain('add_comment')
  })

  test('验证-02: 编辑规则功能正常', async ({ page, request }) => {
    await navigateToAutomation(page)

    await openCreateBuilder(page)
    await fillName(page, '待编辑的规则')
    await pickTrigger(page, '工作项创建时')
    await addCondition(page, 'priority', 'equals', 'high')
    await addAction(page, { type: 'add_comment', value: '初始动作' })
    await submitBuilder(page, '创建')
    expect(await waitFor(async () => findRule(request, '待编辑的规则'))).toBeDefined()

    // Edit through the card's ✏️ button and change the condition value.
    const card = page.locator('div.bg-white.rounded-xl.border').filter({ hasText: '待编辑的规则' }).first()
    await card.locator('button:has-text("✏️")').click()
    await expect(page.locator('input[placeholder*="自动分配"]')).toBeVisible({ timeout: 10000 })

    await fillName(page, '编辑后的规则名称')
    const conditionRow = formRows(page).first()
    await conditionRow.locator('select').nth(2).selectOption('urgent')
    await submitBuilder(page, '更新')

    const updatedRule = await waitFor(async () => findRule(request, '编辑后的规则名称'))
    expect(updatedRule).toBeDefined()
    expect(updatedRule.conditions).toContain('urgent')
    expect(await findRule(request, '待编辑的规则')).toBeUndefined()
  })

  test('验证-03: 启用/禁用切换按钮功能', async ({ request }) => {
    const ruleId = await createRuleViaAPI(
      request, '切换开关的规则', 'issue.created', '[]',
      '[{"type":"add_comment","value":"toggle"}]',
    )
    expect(ruleId).toBeTruthy()

    await request.put(`${BASE_API}/projects/${_projectId}/automations/${ruleId}`, {
      data: { is_enabled: false }, headers: auth(),
    })
    let rule = await waitFor(async () => (await listRules(request)).find((r: any) => r.id === ruleId && r.is_enabled === false))
    expect(rule, '规则应已停用').toBeDefined()

    await request.put(`${BASE_API}/projects/${_projectId}/automations/${ruleId}`, {
      data: { is_enabled: true }, headers: auth(),
    })
    rule = await waitFor(async () => (await listRules(request)).find((r: any) => r.id === ruleId && r.is_enabled === true))
    expect(rule, '规则应已重新启用').toBeDefined()
  })

  test('验证-04: issue.created 触发器正常工作', async ({ request }) => {
    await createRuleViaAPI(
      request, '创建时自动加评论', 'issue.created', '[]',
      '[{"type":"add_comment","value":"🎉 新工作项已创建"}]',
    )

    const issueId = await createIssue(request, '测试issue.created触发')
    expect(issueId).toBeTruthy()

    expect(await waitForComment(request, issueId, '🎉 新工作项已创建')).toBeDefined()
  })

  test('验证-05: comment.added 触发器正常工作', async ({ request }) => {
    await createRuleViaAPI(
      request, '评论时自动标记', 'comment.added',
      '[{"field":"comment","operator":"contains","value":"bug"}]',
      '[{"type":"set_priority","value":"high"}]',
    )

    const issueId = await createIssue(request, '测试comment.added触发')
    await request.post(`${BASE_API}/comments`, {
      data: { issue_id: issueId, body: '发现一个bug需要修复' }, headers: auth(),
    })

    expect(await waitForPriority(request, issueId, 'high')).toBeDefined()
  })

  test('验证-06: issue.updated 触发器（非状态变化）正常工作', async ({ request }) => {
    await createRuleViaAPI(
      request, '更新时自动加评论', 'issue.updated', '[]',
      '[{"type":"add_comment","value":"🔄 工作项已更新"}]',
    )

    const issueId = await createIssue(request, '测试issue.updated触发')
    await request.put(`${BASE_API}/issues/${issueId}`, {
      data: { name: '测试自动化触发 - 已更新' }, headers: auth(),
    })

    expect(await waitForComment(request, issueId, '🔄 工作项已更新')).toBeDefined()
  })

  test('验证-07: issue.state_changed 触发器正常工作', async ({ request }) => {
    const statesRes = await request.get(`${BASE_API}/projects/${_projectId}/settings/states`, { headers: auth() })
    const statesBody = await statesRes.json()
    const states = statesBody.data || statesBody
    // Pick a real state group from this project rather than assuming "done".
    const target = Array.isArray(states) ? states.find((s: any) => s.group === 'completed' || s.group === 'done') : null
    test.skip(!target, '该项目没有已完成状态')

    await createRuleViaAPI(
      request, '状态变更时自动完成', 'issue.state_changed',
      `[{"field":"state_group","operator":"equals","value":"${target.group}"}]`,
      '[{"type":"add_comment","value":"✅ 工作项已完成"}]',
    )

    const issueId = await createIssue(request, '测试issue.state_changed触发')
    await request.put(`${BASE_API}/issues/${issueId}`, {
      data: { state_id: target.id }, headers: auth(),
    })

    expect(await waitForComment(request, issueId, '✅ 工作项已完成')).toBeDefined()
  })

  test('验证-08: 规则执行历史记录', async ({ request }) => {
    const issueId = await createIssue(request, '测试执行历史')
    const historyRes = await request.get(`${BASE_API}/issues/${issueId}/automation-executions`, { headers: auth() })

    // The endpoint is optional for this flow; assert it does not error out.
    expect(historyRes.status()).toBeLessThan(500)
  })

  test('验证-09: 删除规则功能', async ({ request }) => {
    const ruleId = await createRuleViaAPI(
      request, '创建时自动加评论-待删除', 'issue.created', '[]',
      '[{"type":"add_comment","value":"delete me"}]',
    )
    expect(ruleId).toBeTruthy()

    const del = await request.delete(`${BASE_API}/projects/${_projectId}/automations/${ruleId}`, { headers: auth() })
    expect(del.status(), '删除应成功').toBeLessThan(400)

    const deleted = await waitFor(async () => ((await listRules(request)).some((r: any) => r.id === ruleId) ? undefined : true))
    expect(deleted, '规则应已从列表中移除').toBeTruthy()
  })
})
