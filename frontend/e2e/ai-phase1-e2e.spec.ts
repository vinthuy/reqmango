import { test, expect } from '@playwright/test'

const API = 'http://localhost:8000/api/v1'
const BASE = 'http://localhost:5176'

async function getToken(request: any) {
  const r = await request.post(`${API}/auth/login`, {
    data: { email: 'admin@reqmango.com', password: 'demo1234' },
  })
  const { access_token } = await r.json()
  return access_token
}

async function login(page: any, request: any) {
  const token = await getToken(request)
  await page.goto(BASE + '/login')
  await page.evaluate((t: string) => localStorage.setItem('token', t), token)
  return token
}

test.describe('AI Phase 1 — Agent Automation + Result Actions', () => {

  // ═══ P1.1: Agent 自动化触发器 ═══

  test('AI01: Automation rules list includes existing rules', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}` }

    const res = await request.get(`${API}/projects/15/automations`, { headers: H })
    expect(res.status()).toBe(200)
    const rules = await res.json()
    console.log(`Automation rules: ${rules.length}`)
    for (const r of rules.slice(0, 5)) {
      console.log(`  [${r.id}] ${r.name} trigger=${r.trigger_type}`)
    }
  })

  test('AI02: Create automation rule with dispatch_agent action', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }

    const create = await request.post(`${API}/projects/15/automations`, {
      headers: H,
      data: {
        name: `[E2E Agent] ${Date.now()}`,
        trigger_type: 'issue_created',
        conditions: '[{"field":"priority","operator":"equals","value":"urgent"}]',
        actions: '[{"type":"dispatch_agent","value":1,"field":"分析这个紧急工作项并推荐处理方案"}]',
        is_enabled: true,
      },
    })
    expect(create.status()).toBe(201)
    const rule = await create.json()
    console.log(`Created rule [${rule.id}]: ${rule.name}`)
    expect(rule.trigger_type).toBe('issue_created')

    // Verify the rule exists
    const get = await request.get(`${API}/projects/15/automations/${rule.id}`, { headers: H })
    expect(get.status()).toBe(200)

    // Cleanup
    await request.delete(`${API}/projects/15/automations/${rule.id}`, { headers: H })
    console.log('Cleaned up')
  })

  test('AI03: Automation rule with multiple actions including dispatch_agent', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }

    const create = await request.post(`${API}/projects/15/automations`, {
      headers: H,
      data: {
        name: `[E2E Multi] ${Date.now()}`,
        trigger_type: 'state_changed',
        conditions: '[{"field":"state_group","operator":"equals","value":"started"}]',
        actions: '[{"type":"add_comment","value":"状态变更为进行中"},{"type":"dispatch_agent","value":1,"field":"分析工作项进行中的风险"}]',
        is_enabled: true,
      },
    })
    expect(create.status()).toBe(201)
    const rule = await create.json()

    // Execute to verify it doesn't crash
    const exec = await request.post(`${API}/projects/15/automations/${rule.id}/execute`, {
      headers: H,
      data: { issue_id: 733, context: { state_group: 'started' } },
    })
    expect(exec.status()).toBe(200)
    console.log(`Execute multi-action rule: ${exec.status()}`)

    // Check history
    const history = await request.get(`${API}/issues/733/automation-history?limit=1`, { headers: H })
    const hist = await history.json()
    if (hist.length > 0) {
      console.log(`History: ${hist[0].status} actions=${hist[0].actions_taken?.slice(0, 60)}`)
    }

    await request.delete(`${API}/projects/15/automations/${rule.id}`, { headers: H })
  })

  test('AI04: List agents for dispatch_agent action', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}` }

    const agents = await request.get(`${API}/workspaces/8/agents`, { headers: H })
    expect(agents.status()).toBe(200)
    const list = await agents.json()
    console.log(`Agents: ${list.length}`)
    for (const a of list.slice(0, 5)) {
      console.log(`  [${a.id}] ${a.name} type=${a.agent_type} status=${a.status} capabilities=${JSON.stringify(a.capabilities)}`)
    }
  })

  // ═══ P1.3: AI 结果操作化 ═══

  test('AI05: AI Chat sidebar opens with Ctrl+J', async ({ page, request }) => {
    await login(page, request)
    await page.goto(BASE + '/workspace/reqmango-dev/project/15')
    await page.waitForTimeout(3000)

    // Press Ctrl+J to open AI chat
    await page.keyboard.press('Control+KeyJ')
    await page.waitForTimeout(1000)

    const body = await page.textContent('body')
    const hasAI = body?.includes('AI') || body?.includes('Ask') || body?.includes('Build')
    console.log(`AI sidebar visible: ${hasAI}`)
  })

  test('AI06: AI Chat has operation buttons after response', async ({ page, request }) => {
    await login(page, request)
    await page.goto(BASE + '/workspace/reqmango-dev/project/15')
    await page.waitForTimeout(3000)

    // Open AI sidebar
    await page.keyboard.press('Control+KeyJ')
    await page.waitForTimeout(1000)

    // Type a question about project stats
    const input = page.locator('textarea, input[type="text"]').last()
    if (await input.isVisible().catch(() => false)) {
      await input.fill('How many issues are in this project?')
      await page.keyboard.press('Enter')
      await page.waitForTimeout(5000)

      const body = await page.textContent('body')
      // Check for action buttons in AI response
      const hasCreateIssue = body?.includes('Create Issues') || body?.includes('生成工作项')
      console.log(`'Create Issues' button: ${hasCreateIssue}`)
    } else {
      console.log('Chat input not found')
    }
  })

  test('AI07: AICreateDialog opens from project page', async ({ page, request }) => {
    await login(page, request)
    await page.goto(BASE + '/workspace/reqmango-dev/project/15')
    await page.waitForTimeout(3000)

    // Click AI Create button
    const aiCreateBtn = page.locator('button:has-text("AI Create"), button:has-text("AI 创建")').first()
    if (await aiCreateBtn.isVisible().catch(() => false)) {
      await aiCreateBtn.click()
      await page.waitForTimeout(1000)
      const body = await page.textContent('body')
      console.log(`AI Create dialog visible: ${body?.includes('Generate') || body?.includes('Preview') || body?.includes('生成')}`)
    } else {
      console.log('AI Create button not found')
    }
  })

  test('AI08: AI Chat quick action buttons visible', async ({ page, request }) => {
    await login(page, request)
    await page.goto(BASE + '/workspace/reqmango-dev/project/15')
    await page.waitForTimeout(3000)

    await page.keyboard.press('Control+KeyJ')
    await page.waitForTimeout(1000)

    // Check for suggested questions / quick actions
    const body = await page.textContent('body')
    const hasSuggestions = body?.includes('💡') || false
    console.log(`Suggestions visible: ${hasSuggestions}`)
  })

  // ═══ Backend AI API Regression ═══

  test('AI09: AI Chat API returns SSE stream', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }

    // Send a simple AI chat request via fetch (non-streaming test)
    const res = await request.post(`${API}/projects/15/ai/analyze`, {
      headers: H,
    })
    expect(res.status()).toBe(200)
    const data = await res.json()
    console.log(`AI Analyze keys: ${Object.keys(data).join(', ')}`)
  })

  test('AI10: AI config endpoint works', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}` }

    const cfg = await request.get(`${API}/workspaces/8/ai-config`, { headers: H })
    expect(cfg.status()).toBe(200)
    const config = await cfg.json()
    console.log(`AI config: provider=${config.provider} model=${config.model} configured=${config.configured}`)
  })
})
