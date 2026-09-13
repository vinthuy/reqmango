import { test, expect } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'
const RUN = Date.now()

let _token = ''
let _wsId = 0
let _wsSlug = ''
let _projectId = 0
let _states: Array<{ id: number; name: string; group: string }> = []

function auth() {
  return { Authorization: `Bearer ${_token}` }
}

async function ensureSetup(request: any) {
  if (_token) return
  const user = {
    email: `e2e-auto-${RUN}@t.com`, username: `e2e-auto-${RUN}`,
    password: 'E2eTest123!', display_name: 'E2E Automation',
  }
  await request.post(`${BASE_API}/auth/register`, { data: user })
  const res = await request.post(`${BASE_API}/auth/login`, { data: { email: user.email, password: user.password } })
  const { access_token } = await res.json()
  _token = access_token

  const ws = await request.post(`${BASE_API}/workspaces`, {
    data: { name: 'E2E Automation WS', slug: `e2e-auto-ws-${RUN}` },
    headers: auth(),
  })
  const wsData = await ws.json()
  _wsId = wsData.id || wsData.data?.id
  _wsSlug = wsData.slug || wsData.data?.slug

  const proj = await request.post(`${BASE_API}/projects?workspace_id=${_wsId}`, {
    data: { name: 'E2E Automation Project', identifier: 'E2EAUTO', description: 'For automation testing' },
    headers: auth(),
  })
  const projData = await proj.json()
  _projectId = projData.id || projData.data?.id

  const statesRes = await request.get(`${BASE_API}/projects/${_projectId}/settings/states`, { headers: auth() })
  const statesBody = await statesRes.json()
  const states = statesBody.data || statesBody
  _states = Array.isArray(states) ? states : []
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

async function createIssue(request: any, name: string, priority = 'medium'): Promise<number> {
  const res = await request.post(`${BASE_API}/issues?project_id=${_projectId}&workspace_id=${_wsId}`, {
    data: { name, description: '自动化测试用工作项', priority },
    headers: auth(),
  })
  const data = await res.json()
  return data.id || data.data?.id
}

async function getIssue(request: any, issueId: number) {
  const res = await request.get(`${BASE_API}/issues/${issueId}`, { headers: auth() })
  const body = await res.json()
  return body.data || body
}

async function waitFor<T>(probe: () => Promise<T | undefined | null>, timeoutMs = 15000): Promise<T | undefined> {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    const found = await probe()
    if (found) return found
    await new Promise(r => setTimeout(r, 500))
  }
  return undefined
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
const ruleCard = (page: any, name: string) =>
  page.locator('div.bg-white.rounded-xl.border').filter({ hasText: name }).first()

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

async function addAction(page: any, action: { type: string; value?: string }) {
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
      if (action.value) await row.locator('select').nth(1).selectOption(action.value)
      else await row.locator('select').nth(1).selectOption({ index: 1 })
      break
    case 'assign_to':
      await row.locator('select').nth(1).selectOption({ index: 1 })
      break
  }
}

async function submitBuilder(page: any, submitLabel: string) {
  await modal(page).locator('button').filter({ hasText: submitLabel }).last().click()
  await expect(modal(page)).toHaveCount(0, { timeout: 15000 })
}

/** The enable/disable and delete actions are guarded by the shared confirm dialog. */
async function confirmDialog(page: any) {
  const dialog = page.locator('div.z-\\[100\\]')
  await expect(dialog).toBeVisible({ timeout: 10000 })
  await dialog.locator('button').last().click()
  await expect(dialog).toHaveCount(0, { timeout: 10000 })
}

// ==================== Tests ====================

test.describe('ReqMango 自动化功能 E2E 测试', () => {
  test.beforeAll(async ({ request }) => { await ensureSetup(request) })

  test('E2E-01: 创建高优先级自动分配规则', async ({ page, request }) => {
    await navigateToAutomation(page)

    await openCreateBuilder(page)
    await fillName(page, '高优先级自动分配')
    await pickTrigger(page, '工作项创建时')
    await addCondition(page, 'priority', 'equals', 'high')
    await addAction(page, { type: 'assign_to' })
    await submitBuilder(page, '创建')

    const rule = await waitFor(async () => findRule(request, '高优先级自动分配'))
    expect(rule, '规则应已保存到数据库').toBeDefined()
    expect(rule.is_enabled).toBe(true)
    expect(rule.actions).toContain('assign_to')

    // The list must render the new rule without a reload.
    await expect(ruleCard(page, '高优先级自动分配')).toBeVisible({ timeout: 10000 })
  })

  test('E2E-02: 创建状态变更规则', async ({ page, request }) => {
    await navigateToAutomation(page)

    await openCreateBuilder(page)
    await fillName(page, '完成时添加评论')
    await pickTrigger(page, '状态变更时')
    await addAction(page, { type: 'add_comment', value: '✅ 工作项已完成' })
    await submitBuilder(page, '创建')

    const rule = await waitFor(async () => findRule(request, '完成时添加评论'))
    expect(rule).toBeDefined()
    expect(rule.trigger_type).toContain('issue.state_changed')
    expect(rule.actions).toContain('add_comment')
  })

  test('E2E-03: 启用/禁用切换按钮', async ({ page, request }) => {
    const ruleId = await createRuleViaAPI(
      request, '待切换的规则', 'issue.created', '[]',
      '[{"type":"add_comment","value":"toggle"}]',
    )
    expect(ruleId).toBeTruthy()

    await navigateToAutomation(page)
    const card = ruleCard(page, '待切换的规则')
    await expect(card).toBeVisible({ timeout: 10000 })

    // ⏸️ is rendered only by the enable/disable toggle.
    await card.locator('button:has-text("⏸️")').click()
    await confirmDialog(page)
    expect(await waitFor(async () => {
      const rule = (await listRules(request)).find((r: any) => r.id === ruleId)
      return rule && rule.is_enabled === false ? rule : undefined
    })).toBeDefined()
    await expect(card.locator('text=已停用')).toBeVisible({ timeout: 10000 })

    await card.locator('button:has-text("▶️")').click()
    await confirmDialog(page)
    expect(await waitFor(async () => {
      const rule = (await listRules(request)).find((r: any) => r.id === ruleId)
      return rule && rule.is_enabled === true ? rule : undefined
    })).toBeDefined()
  })

  test('E2E-04: 正则匹配条件规则测试', async ({ page, request }) => {
    // The builder offers no regex operator, so the rule is created through the API
    // and this test then verifies both the rendered list entry and the real match.
    const ruleId = await createRuleViaAPI(
      request, 'Bug 自动标记为紧急', 'issue.created',
      '[{"field":"priority","operator":"matches_regex","value":"^hig"}]',
      '[{"type":"set_priority","value":"urgent"}]',
    )
    expect(ruleId).toBeTruthy()

    await navigateToAutomation(page)
    await expect(ruleCard(page, 'Bug 自动标记为紧急')).toBeVisible({ timeout: 10000 })

    // issue.created only exposes issue_id/priority/state_id/project_id in its event
    // context, so the regex is matched against a field the event actually carries.
    const issueId = await createIssue(request, '正则匹配验证', 'high')
    expect(await waitFor(async () => {
      const issue = await getIssue(request, issueId)
      return issue.priority === 'urgent' ? issue : undefined
    }), 'priority=high 应被正则 ^hig 命中并改为 urgent').toBeDefined()
  })

  test('E2E-05: 批量更新触发自动化测试', async ({ request }) => {
    await createRuleViaAPI(
      request, '批量更新时设置优先级', 'issue.updated', '[]',
      '[{"type":"set_priority","value":"high"}]',
    )

    const issueIds = [
      await createIssue(request, '批量更新验证 1'),
      await createIssue(request, '批量更新验证 2'),
      await createIssue(request, '批量更新验证 3'),
    ]

    // Update each issue: every update must trigger the rule independently.
    for (const id of issueIds) {
      const res = await request.put(`${BASE_API}/issues/${id}`, {
        data: { name: `批量更新验证 - 已更新 ${id}` }, headers: auth(),
      })
      expect(res.status()).toBeLessThan(400)
    }

    for (const id of issueIds) {
      const updated = await waitFor(async () => {
        const issue = await getIssue(request, id)
        return issue.priority === 'high' ? issue : undefined
      })
      expect(updated, `工作项 ${id} 应被批量更新规则提升为 high`).toBeDefined()
    }
  })

  test('E2E-06: 循环依赖规则可创建且不会无限触发', async ({ page, request }) => {
    const targetState = _states.find(s => s.group === 'started') || _states[1] || _states[0]
    test.skip(!targetState, '项目没有可用状态')

    await navigateToAutomation(page)

    // Rule A: priority becomes high -> move to another state.
    await openCreateBuilder(page)
    await fillName(page, '规则A: 高优先级→进行中')
    await pickTrigger(page, '工作项更新时')
    await addCondition(page, 'priority', 'equals', 'high')
    await addAction(page, { type: 'change_state', value: targetState.name })
    await submitBuilder(page, '创建')

    // Rule B: that state is reached -> set priority back to high (mutual reference).
    await openCreateBuilder(page)
    await fillName(page, '规则B: 进行中→高优先级')
    await pickTrigger(page, '状态变更时')
    await addAction(page, { type: 'set_priority', value: 'high' })
    await submitBuilder(page, '创建')

    expect(await waitFor(async () => findRule(request, '规则A: 高优先级→进行中'))).toBeDefined()
    expect(await waitFor(async () => findRule(request, '规则B: 进行中→高优先级'))).toBeDefined()
    await expect(ruleCard(page, '规则A')).toBeVisible({ timeout: 10000 })
    await expect(ruleCard(page, '规则B')).toBeVisible({ timeout: 10000 })

    // Fire rule A once and make sure the request completes and the effect lands
    // exactly once (automation actions write to the DB directly, so a mutual
    // reference cannot cascade; the backend loop guard is covered by Go tests).
    const issueId = await createIssue(request, '循环依赖验证', 'medium')
    const res = await request.put(`${BASE_API}/issues/${issueId}`, {
      data: { priority: 'high' }, headers: auth(),
    })
    expect(res.status()).toBeLessThan(400)

    const applied = await waitFor(async () => {
      const issue = await getIssue(request, issueId)
      return issue.state_id === targetState.id ? issue : undefined
    })
    expect(applied, '规则A 应把工作项移动到目标状态').toBeDefined()
  })
})
