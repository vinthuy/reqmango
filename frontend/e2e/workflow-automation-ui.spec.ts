import { test, expect } from '@playwright/test'

const API = 'http://localhost:8000/api/v1'
const BASE = 'http://localhost:5173'

// Helper: get a fresh token
async function getToken(request: any) {
  const r = await request.post(`${API}/auth/login`, {
    data: { email: 'admin@reqmango.com', password: 'demo1234' },
  })
  const { access_token } = await r.json()
  return access_token
}

test.describe('Workflow & Automation — Full UI Journey', () => {

  test('1. Login page → workspace → project settings renders', async ({ page, request }) => {
    const token = await getToken(request)

    // Set auth and navigate
    await page.goto(BASE + '/login')
    await page.evaluate((t) => localStorage.setItem('token', t), token)

    // Home page
    await page.goto(BASE + '/')
    await page.waitForTimeout(2000)
    console.log('Home title:', await page.title())

    // Workspace Settings
    await page.goto(BASE + '/workspace/reqmango-dev/settings')
    await page.waitForTimeout(1500)
    const wsBody = await page.textContent('body')
    console.log('Workspace settings:', wsBody?.length, 'chars')

    // Project Settings
    await page.goto(BASE + '/workspace/reqmango-dev/project/15/settings')
    await page.waitForTimeout(2000)
    const psBody = await page.textContent('body')
    console.log('Project settings:', psBody?.length, 'chars')
  })

  test('2. States API — verify format matches WorkflowManager expectation', async ({ request }) => {
    const token = await getToken(request)
    const headers = { Authorization: `Bearer ${token}` }

    // WorkflowManager.vue line 56 expects: states.value = s.data
    const res = await request.get(`${API}/projects/15/settings/states`, { headers })
    expect(res.status()).toBe(200)
    const body = await res.json()

    console.log('States response keys:', Object.keys(body).join(', '))
    if ('data' in body) {
      console.log('✅ States wrap in {data: [...]} — WorkflowManager expects this')
      console.log('States count:', body.data.length)
      for (const s of body.data.slice(0, 5)) {
        console.log(`  State: ${s.name} (id=${s.id}, group=${s.group})`)
      }
    } else if (Array.isArray(body)) {
      console.log('⚠️ States return as raw array — WorkflowManager expects {data: [...]}')
      console.log('   This means WorkflowManager.load() will fail because states.value = s.data is undefined')
      console.log('   WorkflowManager line 56: states.value = s.data')
    } else {
      console.log('⚠️ Unknown format:', JSON.stringify(body).slice(0, 200))
    }
  })

  test('3. WorkflowManager — full CRUD via API (simulating UI actions)', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }

    // Step 1: List workflows (WorkflowManager.load)
    console.log('--- Step 1: List workflows ---')
    const list1 = await request.get(`${API}/projects/15/workflows`, { headers: H })
    expect(list1.status()).toBe(200)
    const wfs1 = await list1.json()
    console.log(`Existing workflows: ${wfs1.length}`)
    for (const w of wfs1) {
      console.log(`  [${w.id}] ${w.name} active=${w.is_active} transitions=${w.transitions?.length || 0}`)
    }

    // Step 2: Create workflow (WorkflowManager.save)
    console.log('\n--- Step 2: Create workflow ---')
    const create = await request.post(`${API}/projects/15/workflows`, {
      headers: H,
      data: { name: `[UI Flow Test] ${Date.now()}`, description: 'Created simulating UI flow' },
    })
    expect(create.status()).toBe(201)
    const wf = await create.json()
    console.log(`Created: [${wf.id}] ${wf.name}`)

    // Step 3: Add transition (WorkflowManager.saveTrans)
    console.log('\n--- Step 3: Add transition ---')
    const trans = await request.post(`${API}/projects/15/workflows/${wf.id}/transitions`, {
      headers: H,
      data: { from_state_id: 85, to_state_id: 86, description: 'Backlog→Todo' },
    })
    expect(trans.status()).toBe(201)
    const tr = await trans.json()
    console.log(`Created transition: ${tr.from_name} → ${tr.to_name} (rule=${tr.rule_type})`)

    // Verify transition appears in workflow detail
    const wfDetail = await request.get(`${API}/projects/15/workflows/${wf.id}`, { headers: H })
    const detail = await wfDetail.json()
    expect(detail.transitions?.length).toBeGreaterThanOrEqual(1)
    console.log(`Workflow detail: ${detail.transitions.length} transitions`)

    // Step 4: Cleanup (WorkflowManager.confirmDel)
    console.log('\n--- Step 4: Delete workflow (cleanup) ---')
    await request.delete(`${API}/projects/15/workflows/${wf.id}`, { headers: H })
    console.log('Deleted test workflow')
  })

  test('4. AutomationManager — full CRUD via API (simulating UI actions)', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }

    // List automations (AutomationManager.load)
    console.log('--- List automations ---')
    const list = await request.get(`${API}/projects/15/automations`, { headers: H })
    expect(list.status()).toBe(200)
    const autos = await list.json()
    console.log(`Existing automations: ${autos.length}`)
    for (const a of autos) {
      console.log(`  [${a.id}] ${a.name} trigger=${a.trigger_type} enabled=${a.is_enabled} exec=${a.execution_count}`)
    }

    // Create automation (AutomationManager.save)
    console.log('\n--- Create automation ---')
    const create = await request.post(`${API}/projects/15/automations`, {
      headers: H,
      data: {
        name: `[UI Flow Test] ${Date.now()}`,
        trigger_type: 'issue_created',
        conditions: '[{"field":"priority","operator":"equals","value":"high"}]',
        actions: '[{"type":"add_comment","value":"[Auto] High priority issue created"}]',
      },
    })
    expect(create.status()).toBe(201)
    const auto = await create.json()
    console.log(`Created: [${auto.id}] ${auto.name}`)

    // Execute manually (AutomationHandler.Execute endpoint)
    console.log('\n--- Execute automation ---')
    const exec = await request.post(`${API}/projects/15/automations/${auto.id}/execute`, {
      headers: H,
      data: { issue_id: 1227, context: { priority: 'high' } },
    })
    expect(exec.status()).toBe(200)
    const execResult = await exec.json()
    console.log('Execute results:', execResult.results)

    // Cleanup
    console.log('\n--- Delete automation (cleanup) ---')
    await request.delete(`${API}/projects/15/automations/${auto.id}`, { headers: H })
    console.log('Deleted test automation')
  })

  test('5. ⚠️ GAP: WorkflowManager missing approval fields', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }

    console.log('=== WorkflowManager.vue Transition Form Audit ===')
    console.log('Current form fields (from source code lines 31-36):')
    console.log('  1. From State dropdown  ✅')
    console.log('  2. To State dropdown    ✅')
    console.log('  3. Description input    ✅')
    console.log('')
    console.log('MISSING from the UI form:')
    console.log('  ❌ Rule Type dropdown (allow/approval)')
    console.log('  ❌ Approver IDs input')
    console.log('  ❌ Role Allowed input')
    console.log('')
    console.log('Impact: Users cannot create approval transitions from the UI.')
    console.log('They can only create "allow" transitions.')
    console.log('')

    // Verify API supports it but UI doesn't expose it
    const wfCreate = await request.post(`${API}/projects/15/workflows`, {
      headers: H,
      data: { name: `[Gap Test] ${Date.now()}`, description: 'Approval gap test' },
    })
    const wf = await wfCreate.json()

    // API supports approval rule_type + approver_ids
    const tr = await request.post(`${API}/projects/15/workflows/${wf.id}/transitions`, {
      headers: H,
      data: {
        from_state_id: 85, to_state_id: 86,
        rule_type: 'approval', approver_ids: '49',
        description: 'Requires admin approval',
      },
    })
    expect(tr.status()).toBe(201)
    const t = await tr.json()
    console.log(`API: Created approval transition (rule=${t.rule_type}, approvers=${t.approver_ids})`)
    console.log('UI: Cannot create this — form has no rule_type or approver_ids fields')
    console.log('')
    console.log('FIX: Add to WorkflowManager.vue add-transition form:')
    console.log('  - <select v-model="trans.rule_type"> with options allow/approval')
    console.log('  - <input v-model="trans.approver_ids"> for comma-separated user IDs')
    console.log('  - <input v-model="trans.role_allowed"> for role-based approval')

    await request.delete(`${API}/projects/15/workflows/${wf.id}`, { headers: H })
  })

  test('6. ⚠️ States API response format check', async ({ request }) => {
    const token = await getToken(request)
    const headers = { Authorization: `Bearer ${token}` }

    const res = await request.get(`${API}/projects/15/settings/states`, { headers })
    const body = await res.json()

    // WorkflowManager line 56: states.value = s.data
    // If the API returns a raw array instead of {data: [...]}, the UI breaks
    const hasWrapper = 'data' in body && Array.isArray(body.data)
    const isArray = Array.isArray(body)

    console.log(`States API format: ${hasWrapper ? '{data:[...]}' : isArray ? 'raw array' : 'other'}`)

    if (!hasWrapper) {
      console.log('⚠️ ISSUE: WorkflowManager.vue line 56 expects states.value = s.data')
      console.log('   But the API returns a different format.')
      console.log('   This means the UI will silently fail to load states,')
      console.log('   and the "Add Transition" form will show empty dropdowns.')
    } else {
      console.log('✅ Format matches WorkflowManager expectation')
    }
  })
})
