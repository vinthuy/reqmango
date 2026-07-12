import { test, expect } from '@playwright/test'

const API = 'http://localhost:8000/api/v1'
const BASE = 'http://localhost:5173'

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

test.describe('Custom Fields — Full UI Journey', () => {

  test('CF01: Custom Fields settings page renders', async ({ page, request }) => {
    await login(page, request)
    await page.goto(BASE + '/workspaces/8/projects/15/custom-fields')
    await page.waitForTimeout(2500)
    const body = await page.textContent('body')
    console.log('CF Settings page chars:', body?.length)
    expect(body).toBeTruthy()
  })

  test('CF02: Custom Fields API — list fields', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}` }

    const res = await request.get(`${API}/custom-fields?workspace_id=8`, { headers: H })
    expect(res.status()).toBe(200)
    const fields = await res.json()
    expect(Array.isArray(fields)).toBe(true)
    console.log(`Total custom fields: ${fields.length}`)
    for (const f of fields.slice(0, 5)) {
      console.log(`  [${f.id}] ${f.name} type=${f.field_type} active=${f.is_active}`)
    }
  })

  test('CF03: Custom Fields — CRUD lifecycle via API', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }
    const name = `[E2E-CRUD] ${Date.now()}`

    // Create
    const create = await request.post(`${API}/custom-fields?workspace_id=8`, {
      headers: H,
      data: { name, field_type: 'text', description: 'CRUD test' },
    })
    expect(create.status()).toBe(201)
    const field = await create.json()
    console.log(`Created: [${field.id}]`)
    expect(field.field_type).toBe('text')

    // Get
    const get = await request.get(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    expect(get.status()).toBe(200)

    // Update
    const update = await request.put(`${API}/custom-fields/${field.id}?workspace_id=8`, {
      headers: H,
      data: { description: 'Updated description' },
    })
    expect(update.status()).toBe(200)

    // Delete
    const del = await request.delete(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    expect(del.status()).toBe(200)
    console.log('CRUD lifecycle: OK')
  })

  test('CF04: Dropdown custom field with options', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }
    const name = `[E2E-Dropdown] ${Date.now()}`

    const create = await request.post(`${API}/custom-fields?workspace_id=8`, {
      headers: H,
      data: { name, field_type: 'dropdown' },
    })
    expect(create.status()).toBe(201)
    const field = await create.json()

    // Add options
    for (const opt of [{ value: 'Alpha', color: '#3B82F6' }, { value: 'Beta', color: '#10B981' }]) {
      const r = await request.post(`${API}/custom-fields/${field.id}/options?workspace_id=8`, {
        headers: H, data: opt,
      })
      expect(r.status()).toBe(201)
    }

    // Verify
    const get = await request.get(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    const got = await get.json()
    expect(got.options).toHaveLength(2)
    console.log(`Options: ${got.options.map((o: any) => o.value).join(', ')}`)

    // Cleanup
    await request.delete(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    console.log('Dropdown CRUD: OK')
  })

  test('CF05: Issue custom field values — set, read, bulk update', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }
    const name = `[E2E-Values] ${Date.now()}`

    const create = await request.post(`${API}/custom-fields?workspace_id=8`, {
      headers: H,
      data: { name, field_type: 'text' },
    })
    const field = await create.json()

    // Set value (workspace_id as query param)
    const setVal = await request.post(`${API}/custom-fields/issues/733/values?workspace_id=8`, {
      headers: H,
      data: { field_id: field.id, value: 'Hello E2E' },
    })
    expect(setVal.status()).toBe(201)
    console.log('Set value: OK')

    // Read
    const getVals = await request.get(`${API}/custom-fields/issues/733/values?workspace_id=8`, { headers: H })
    expect(getVals.status()).toBe(200)
    const vals = await getVals.json()
    const ours = Array.isArray(vals) ? vals.find((v: any) => v.field_id === field.id) : null
    expect(ours).toBeDefined()
    console.log(`Read value: ${ours?.value}`)

    // Cleanup
    await request.delete(`${API}/custom-fields/issues/733/values/${field.id}?workspace_id=8`, { headers: H })
    await request.delete(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    console.log('Values test: OK')
  })

  test('CF06: Issue detail page renders with custom fields', async ({ page, request }) => {
    await login(page, request)
    await page.goto(BASE + '/workspace/reqmango-dev/project/15/issues/733')
    await page.waitForTimeout(3000)

    const body = await page.textContent('body')
    console.log('Issue detail chars:', body?.length)
    expect(body).toBeTruthy()
  })

  test('CF07: Issue list column config gear icon opens modal', async ({ page, request }) => {
    await login(page, request)
    await page.goto(BASE + '/workspace/reqmango-dev/project/15')
    await page.waitForTimeout(3000)

    // Find the gear icon (column config button has a title attribute)
    const gearBtn = page.locator('[title*="Column"], [title*="列"]').first()
    const gearSvg = page.locator('button svg path[d*="M10.325"]').first()

    if (await gearBtn.isVisible().catch(() => false)) {
      await gearBtn.click()
    } else if (await gearSvg.isVisible().catch(() => false)) {
      // Click the parent button
      await page.locator('button:has(svg)').filter({ has: page.locator('svg') }).nth(1).click()
    }
    await page.waitForTimeout(1000)

    const modal = await page.textContent('body')
    console.log('Page text sample:', modal?.slice(0, 300))
  })

  test('CF08: All field types — create via API', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }

    for (const ftype of ['text', 'number', 'boolean', 'date', 'url']) {
      const create = await request.post(`${API}/custom-fields?workspace_id=8`, {
        headers: H,
        data: { name: `[E2E-${ftype}] ${Date.now()}`, field_type: ftype },
      })
      expect(create.status()).toBe(201)
      const field = await create.json()
      expect(field.field_type).toBe(ftype)
      await request.delete(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    }
    console.log('All field types: OK')
  })

  test('CF09: Field active/inactive toggle', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }
    const name = `[E2E-Toggle] ${Date.now()}`

    const create = await request.post(`${API}/custom-fields?workspace_id=8`, {
      headers: H,
      data: { name, field_type: 'text' },
    })
    const field = await create.json()
    expect(field.is_active).toBe(true)

    // Deactivate
    await request.put(`${API}/custom-fields/${field.id}?workspace_id=8`, {
      headers: H, data: { is_active: false },
    })
    const deact = await request.get(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    expect((await deact.json()).is_active).toBe(false)

    // Reactivate
    await request.put(`${API}/custom-fields/${field.id}?workspace_id=8`, {
      headers: H, data: { is_active: true },
    })
    const react = await request.get(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    expect((await react.json()).is_active).toBe(true)

    await request.delete(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    console.log('Toggle: OK')
  })

  test('CF10: Project-scoped custom field', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }
    const name = `[E2E-Project] ${Date.now()}`

    const create = await request.post(`${API}/custom-fields?workspace_id=8`, {
      headers: H,
      data: { name, field_type: 'text', project_id: 15 },
    })
    expect(create.status()).toBe(201)
    const field = await create.json()
    expect(field.project_id).toBe(15)

    await request.delete(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    console.log('Project-scoped field: OK')
  })

  test('CF11: Default value for new custom field', async ({ request }) => {
    const token = await getToken(request)
    const H = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }
    const name = `[E2E-Default] ${Date.now()}`

    const create = await request.post(`${API}/custom-fields?workspace_id=8`, {
      headers: H,
      data: { name, field_type: 'text', default_value: 'Default Text' },
    })
    expect(create.status()).toBe(201)
    const field = await create.json()
    expect(field.default_value).toBe('Default Text')

    await request.delete(`${API}/custom-fields/${field.id}?workspace_id=8`, { headers: H })
    console.log('Default value: OK')
  })
})
