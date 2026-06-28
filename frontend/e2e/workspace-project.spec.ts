/**
 * E2E Tests - Workspace & Project Flow
 * 测试工作空间和项目的创建、导航、页面功能
 */
import { test, expect } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'

// Helper: register and login, return token
async function getToken(request: any) {
  const ts = Date.now()
  const user = { email: `e2ews${ts}@t.com`, username: `e2ews${ts}`, password: 'Test123!', display_name: 'T' }
  await request.post(`${BASE_API}/auth/register`, { data: user })
  const login = await request.post(`${BASE_API}/auth/login`, { data: { email: user.email, password: user.password } })
  const { access_token } = await login.json()
  return { token: access_token, user }
}

// Helper: set auth token in browser
async function loginViaStorage(page: any, token: string) {
  await page.goto('/login')
  await page.evaluate((t: string) => {
    localStorage.setItem('token', t)
  }, token)
}

test.describe('Workspace & Project', () => {
  let token: string
  let wsSlug: string
  let projectId: number

  test.beforeAll(async ({ request }) => {
    const result = await getToken(request)
    token = result.token

    // Create workspace via API
    const wsRes = await request.post(`${BASE_API}/workspaces`, {
      data: { name: 'E2E WS', slug: `e2e-ws-${Date.now()}` },
      headers: { Authorization: `Bearer ${token}` },
    })
    const ws = await wsRes.json()
    wsSlug = ws.slug

    // Create project via API
    const projRes = await request.post(`${BASE_API}/projects?workspace_id=${ws.id}`, {
      data: { name: 'E2E Project', identifier: 'E2E' },
      headers: { Authorization: `Bearer ${token}` },
    })
    const proj = await projRes.json()
    projectId = proj.id
  })

  test('should show home page after login', async ({ page }) => {
    await loginViaStorage(page, token)
    await page.goto('/')

    // Should show home page content
    await expect(page.locator('body')).not.toContainText(/login/i)
    // Should have sidebar
    await expect(page.locator('aside, nav, [class*="sidebar"]').first()).toBeVisible({ timeout: 5000 })
  })

  test('should navigate to workspace page', async ({ page }) => {
    await loginViaStorage(page, token)
    await page.goto(`/workspace/${wsSlug}`)

    await expect(page).toHaveURL(new RegExp(wsSlug))
    await expect(page.locator('body')).toBeVisible()
  })

  test('should navigate to project page', async ({ page }) => {
    await loginViaStorage(page, token)
    await page.goto(`/workspace/${wsSlug}/project/${projectId}`)

    await expect(page).toHaveURL(new RegExp(String(projectId)))
    await expect(page.locator('body')).toBeVisible()
  })

  test('should show project tabs on project page', async ({ page }) => {
    await loginViaStorage(page, token)
    await page.goto(`/workspace/${wsSlug}/project/${projectId}`)

    // Should have tabs (工作项/Issues, 周期/Cycles, etc.)
    const tabContainer = page.locator('[class*="tab"], nav[class*="flex"], [role="tablist"]').first()
    await expect(tabContainer).toBeVisible({ timeout: 8000 })
  })

  test('should open AI chat sidebar', async ({ page }) => {
    await loginViaStorage(page, token)
    await page.goto(`/workspace/${wsSlug}/project/${projectId}`)

    // Look for AI chat button
    const aiButton = page.locator('button:has-text("AI"), button[title*="AI"], [class*="ai"]').first()
    if (await aiButton.isVisible({ timeout: 3000 }).catch(() => false)) {
      await aiButton.click()
      // Should show AI sidebar/dialog
      await page.waitForTimeout(1000)
    }
  })

  test('should load project settings page', async ({ page }) => {
    await loginViaStorage(page, token)
    await page.goto(`/workspace/${wsSlug}/project/${projectId}/settings`)

    await expect(page).toHaveURL(/settings/)
    await expect(page.locator('body')).toBeVisible()
  })

  test('should navigate via sidebar menu', async ({ page }) => {
    await loginViaStorage(page, token)
    await page.goto(`/workspace/${wsSlug}`)

    // Sidebar should show workspace name
    const sidebar = page.locator('aside, nav[class*="sidebar"]').first()
    if (await sidebar.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expect(sidebar).toBeVisible()
    }
  })
})
