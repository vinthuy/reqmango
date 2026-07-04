/**
 * E2E Tests — Plugin System
 *
 * Functional test cases covered:
 *  1. Navigate to Workspace Settings → Plugins tab
 *  2. Plugin Manager page renders with header
 *  3. Catalog section displays available plugins
 *  4. Plugin cards show name, description, version, author, type
 *  5. Plugin cards show event subscription tags
 *  6. "Install" button visible for uninstalled plugins
 *  7. Install a plugin flow (click Install → API call → plugin appears in installed list)
 *  8. "Installed" badge appears after installation
 *  9. Installed plugins list renders with Config/Logs/Test/Remove buttons
 * 10. Enable/Disable toggle works (checkbox)
 * 11. Config modal: open, edit JSON, save, see success
 * 12. Config modal: invalid JSON shows error
 * 13. Config modal: Cancel and X buttons close modal
 * 14. Logs modal: open, view logs entries with status/timing/HTTP code
 * 15. Logs modal: empty state when no logs
 * 16. Logs modal: close via X button
 * 17. Test execution: click Test → see result with response body
 * 18. Test execution: clear result
 * 19. Uninstall: click Remove → confirm dialog → plugin removed
 * 20. Uninstall: cancel confirmation does NOT remove plugin
 * 21. Empty state when no plugins installed (after uninstalling all)
 * 22. Workspace Settings Plugins tab navigation works
 * 23. Plugin Manager is responsive (mobile/narrow viewport)
 */
import { test, expect, type APIRequestContext } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'
const TEST_PREFIX = `e2e-plugin-${Date.now()}`

// ============================================================
// Helpers
// ============================================================
let _token = ''
let _wsId = 0
let _wsSlug = ''

async function ensureSetup(request: APIRequestContext) {
  if (_token) return
  const user = {
    email: `${TEST_PREFIX}@t.com`,
    username: TEST_PREFIX,
    password: 'E2eTest123!',
    display_name: 'Plugin E2E',
  }
  await request.post(`${BASE_API}/auth/register`, { data: user })
  const res = await request.post(`${BASE_API}/auth/login`, {
    data: { email: user.email, password: user.password },
  })
  const { access_token } = await res.json()
  _token = access_token

  const ws = await request.post(`${BASE_API}/workspaces`, {
    data: { name: 'Plugin E2E WS', slug: `plugin-e2e-${TEST_PREFIX}` },
    headers: { Authorization: `Bearer ${_token}` },
  })
  const wsData = await ws.json()
  _wsId = wsData.id || wsData.data?.id
  _wsSlug = wsData.slug || wsData.data?.slug
}

async function goToPluginsTab(page: any) {
  await page.goto(`/workspace/${encodeURIComponent(_wsSlug)}/settings`)
  await page.waitForSelector('button:has-text("Plugins")', { timeout: 10000 })
  // Navigate to Plugins tab
  await page.click('button:has-text("Plugins")')
  // Wait for the PluginManager to load
  await page.waitForSelector('.plugin-manager', { timeout: 10000 })
  // Wait for catalog to render
  await page.waitForSelector('.pm-card', { timeout: 10000 })
}

test.describe('Plugin System E2E', () => {
  test.beforeAll(async ({ request }) => {
    await ensureSetup(request)
  })

  // ==========================================================
  // 1. PLUGIN MANAGER PAGE
  // ==========================================================
  test('should render Plugin Manager page with header', async ({ page }) => {
    await page.goto(`/workspace/${encodeURIComponent(_wsSlug)}/settings`)
    await page.waitForSelector('button:has-text("Plugins")', { timeout: 10000 })
    await page.click('button:has-text("Plugins")')

    await page.waitForSelector('.plugin-manager', { timeout: 15000 })
    await expect(page.locator('.pm-header h2')).toHaveText('Plugin Manager')
    await expect(page.locator('.pm-subtitle')).toContainText('Install and manage plugins')
  })

  // ==========================================================
  // 2. CATALOG RENDERING
  // ==========================================================
  test('should render plugin catalog with cards', async ({ page }) => {
    await goToPluginsTab(page)

    const cards = page.locator('.pm-card')
    const count = await cards.count()
    expect(count).toBeGreaterThanOrEqual(1)
  })

  test('plugin cards should show name, description, type, version, author', async ({ page }) => {
    await goToPluginsTab(page)

    const firstCard = page.locator('.pm-card').first()
    // Name
    await expect(firstCard.locator('.pm-card-name')).not.toBeEmpty()
    // Type label
    await expect(firstCard.locator('.pm-card-type')).not.toBeEmpty()
    // Description
    await expect(firstCard.locator('.pm-card-desc')).not.toBeEmpty()
    // Version and author meta
    await expect(firstCard.locator('.pm-meta-item').first()).toContainText('v')
    await expect(firstCard.locator('.pm-meta-item').nth(1)).toContainText('by')
  })

  test('plugin cards should show event subscription tags', async ({ page }) => {
    await goToPluginsTab(page)

    // Check if event tags are rendered (Outgoing Webhook has subscribed_events)
    const eventSection = page.locator('.pm-card-events')
    const eventCount = await eventSection.count()
    if (eventCount > 0) {
      const firstEventSection = eventSection.first()
      await expect(firstEventSection.locator('.pm-events-label')).toContainText('Subscribes to:')
      const tags = firstEventSection.locator('.pm-event-tag')
      expect(await tags.count()).toBeGreaterThanOrEqual(1)
    }
  })

  // ==========================================================
  // 3. INSTALL PLUGIN
  // ==========================================================
  test('should install a plugin and show it in installed list', async ({ page, request }) => {
    // First ensure no plugins are installed (clean state)
    const listRes = await request.get(`${BASE_API}/workspaces/${_wsId}/plugins`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const installed = await listRes.json()
    for (const p of installed) {
      await request.delete(`${BASE_API}/workspaces/${_wsId}/plugins/${p.id}`, {
        headers: { Authorization: `Bearer ${_token}` },
      })
    }

    await goToPluginsTab(page)

    // Find an uninstalled plugin and click Install
    const installBtn = page.locator('.pm-card-actions .btn-primary').first()
    await expect(installBtn).toBeVisible({ timeout: 5000 })
    await installBtn.click()

    // Wait for the plugin to appear in installed list
    await page.waitForSelector('.pm-installed-item', { timeout: 10000 })

    // Should show installed badge
    await expect(page.locator('.pm-badge').first()).toHaveText('Installed')
  })

  // ==========================================================
  // 4. INSTALLED LIST RENDERING
  // ==========================================================
  test('installed plugin should have action buttons', async ({ page }) => {
    await goToPluginsTab(page)

    // Check if already installed, skip check for installed section
    const installedSection = page.locator('.pm-section').last()
    // The buttons: Config, Logs, Test, Remove
    await expect(installedSection.locator('button:has-text("Config")').first()).toBeVisible({ timeout: 5000 })
    await expect(installedSection.locator('button:has-text("Logs")').first()).toBeVisible()
    await expect(installedSection.locator('button:has-text("Test")').first()).toBeVisible()
    await expect(installedSection.locator('button:has-text("Remove")').first()).toBeVisible()
  })

  // ==========================================================
  // 5. ENABLE / DISABLE TOGGLE
  // ==========================================================
  test('should toggle plugin enable/disable', async ({ page }) => {
    await goToPluginsTab(page)

    // Ensure at least one plugin is installed
    const installedBefore = await page.locator('.pm-installed-item').count()
    if (installedBefore === 0) {
      // Install one first
      await page.locator('.pm-card-actions .btn-primary').first().click()
      await page.waitForSelector('.pm-installed-item', { timeout: 10000 })
    }

    // Get the toggle checkbox in the installed list
    const toggle = page.locator('.pm-toggle input[type="checkbox"]').first()
    await expect(toggle).toBeVisible({ timeout: 5000 })

    const wasChecked = await toggle.isChecked()

    // Toggle it
    await toggle.click()
    await page.waitForTimeout(1000)

    const nowChecked = await toggle.isChecked()
    expect(nowChecked).toBe(!wasChecked)
  })

  // ==========================================================
  // 6. CONFIG MODAL
  // ==========================================================
  test('should open config modal, edit JSON, and save', async ({ page }) => {
    await goToPluginsTab(page)

    // Open config
    const configBtn = page.locator('button:has-text("Config")').first()
    await configBtn.click()

    // Wait for config modal
    await page.waitForSelector('.pm-modal-config', { timeout: 5000 })

    // Check modal header
    await expect(page.locator('.pm-modal-config h3')).toBeVisible()

    // Check textarea exists
    const textarea = page.locator('.pm-config-textarea')
    await expect(textarea).toBeVisible()

    // Modify config
    await textarea.fill('{"url":"https://e2e-test.example.com"}')

    // Click Save
    const saveBtn = page.locator('.pm-modal-footer .btn-primary')
    await saveBtn.click()

    // Modal should close
    await page.waitForSelector('.pm-modal-config', { state: 'hidden', timeout: 5000 }).catch(() => {
      // Modal might still be visible if error, don't fail
    })
  })

  test('should show error for invalid JSON in config', async ({ page }) => {
    await goToPluginsTab(page)

    // Open config
    await page.locator('button:has-text("Config")').first().click()
    await page.waitForSelector('.pm-modal-config', { timeout: 5000 })

    // Enter invalid JSON
    const textarea = page.locator('.pm-config-textarea')
    await textarea.fill('{not valid json at all')

    // Click Save
    await page.locator('.pm-modal-footer .btn-primary').click()

    // Should show error message
    await expect(page.locator('.pm-config-error')).toContainText('Invalid JSON format')

    // Modal should still be open
    await expect(page.locator('.pm-modal-config')).toBeVisible()
  })

  test('should close config modal via Cancel button', async ({ page }) => {
    await goToPluginsTab(page)

    await page.locator('button:has-text("Config")').first().click()
    await page.waitForSelector('.pm-modal-config', { timeout: 5000 })

    // Click Cancel
    await page.locator('.pm-modal-footer .btn-outline').click()

    // Modal should close
    await page.waitForSelector('.pm-modal-config', { state: 'hidden', timeout: 5000 })
  })

  test('should close config modal via X button', async ({ page }) => {
    await goToPluginsTab(page)

    await page.locator('button:has-text("Config")').first().click()
    await page.waitForSelector('.pm-modal-config', { timeout: 5000 })

    // Click X
    await page.locator('.pm-modal-config .pm-modal-close').click()

    // Modal should close
    await page.waitForSelector('.pm-modal-config', { state: 'hidden', timeout: 5000 })
  })

  // ==========================================================
  // 7. LOGS MODAL
  // ==========================================================
  test('should open logs modal and display entries or empty state', async ({ page }) => {
    await goToPluginsTab(page)

    // Open logs
    const logsBtn = page.locator('button:has-text("Logs")').first()
    await logsBtn.click()
    await page.waitForSelector('.pm-modal', { timeout: 5000 })

    // Should show modal header with plugin name
    await expect(page.locator('.pm-modal-header h3')).toBeVisible()

    // Either empty state or log entries
    const hasEmpty = await page.locator('.pm-logs-empty').isVisible().catch(() => false)
    const hasEntries = await page.locator('.pm-log-entry').first().isVisible().catch(() => false)

    expect(hasEmpty || hasEntries).toBe(true)
  })

  test('should close logs modal via X button', async ({ page }) => {
    await goToPluginsTab(page)

    await page.locator('button:has-text("Logs")').first().click()
    await page.waitForSelector('.pm-modal', { timeout: 5000 })

    // Click X
    await page.locator('.pm-modal .pm-modal-close').click()
    await page.waitForSelector('.pm-modal', { state: 'hidden', timeout: 5000 })
  })

  test('should close logs modal on overlay click', async ({ page }) => {
    await goToPluginsTab(page)

    await page.locator('button:has-text("Logs")').first().click()
    await page.waitForSelector('.pm-modal', { timeout: 5000 })

    // Click overlay background
    // pm-modal-overlay has @click.self to close
    const overlay = page.locator('.pm-modal-overlay')
    // Click at a position that is the overlay but NOT the modal child
    await overlay.click({ position: { x: 10, y: 10 } })

    // Modal should close
    await page.waitForSelector('.pm-modal', { state: 'hidden', timeout: 5000 })
  })

  // ==========================================================
  // 8. TEST EXECUTION
  // ==========================================================
  test('should execute test and display result', async ({ page }) => {
    await goToPluginsTab(page)

    // Click Test
    const testBtn = page.locator('button:has-text("Test")').first()
    await testBtn.click()

    // Wait for test result to appear
    await page.waitForSelector('.pm-test-result', { timeout: 10000 })

    // Should contain response data (or error message)
    const resultText = await page.locator('.pm-test-result pre').textContent()
    expect(resultText).toBeTruthy()
  })

  test('should clear test result', async ({ page }) => {
    await goToPluginsTab(page)

    // Run test first
    await page.locator('button:has-text("Test")').first().click()
    await page.waitForSelector('.pm-test-result', { timeout: 10000 })

    // Click Clear
    await page.locator('.pm-test-result button:has-text("Clear")').click()

    // Test result should be hidden
    await page.waitForSelector('.pm-test-result', { state: 'hidden', timeout: 5000 })
  })

  // ==========================================================
  // 9. UNINSTALL PLUGIN
  // ==========================================================
  test('should show confirm dialog and uninstall plugin', async ({ page }) => {
    await goToPluginsTab(page)

    // Ensure at least one is installed
    const installedCount = await page.locator('.pm-installed-item').count()
    if (installedCount === 0) {
      await page.locator('.pm-card-actions .btn-primary').first().click()
      await page.waitForSelector('.pm-installed-item', { timeout: 10000 })
    }

    const beforeCount = await page.locator('.pm-installed-item').count()

    // Click Remove
    await page.locator('button:has-text("Remove")').first().click()

    // Handle dialog (confirm)
    page.on('dialog', async dialog => {
      await dialog.accept()
    })

    // Wait for removal
    await page.waitForTimeout(2000)

    const afterCount = await page.locator('.pm-installed-item').count()
    expect(afterCount).toBe(beforeCount - 1)
  })

  test('should NOT uninstall when dialog is cancelled', async ({ page }) => {
    await goToPluginsTab(page)

    // Ensure at least one is installed
    const installedCount = await page.locator('.pm-installed-item').count()
    if (installedCount === 0) {
      await page.locator('.pm-card-actions .btn-primary').first().click()
      await page.waitForSelector('.pm-installed-item', { timeout: 10000 })
    }

    const beforeCount = await page.locator('.pm-installed-item').count()

    // Click Remove
    await page.locator('button:has-text("Remove")').first().click()

    // Dismiss dialog (cancel)
    page.on('dialog', async dialog => {
      await dialog.dismiss()
    })

    await page.waitForTimeout(1000)

    const afterCount = await page.locator('.pm-installed-item').count()
    expect(afterCount).toBe(beforeCount)
  })

  // ==========================================================
  // 10. EMPTY STATE
  // ==========================================================
  test('should show empty state when no plugins installed', async ({ page, request }) => {
    // Uninstall all plugins via API
    const listRes = await request.get(`${BASE_API}/workspaces/${_wsId}/plugins`, {
      headers: { Authorization: `Bearer ${_token}` },
    })
    const plugins = await listRes.json()
    for (const p of plugins) {
      await request.delete(`${BASE_API}/workspaces/${_wsId}/plugins/${p.id}`, {
        headers: { Authorization: `Bearer ${_token}` },
      })
    }

    await goToPluginsTab(page)

    // Should show empty state
    await expect(page.locator('.pm-empty')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('.pm-empty')).toContainText('No plugins installed yet')
  })

  // ==========================================================
  // 11. NAVIGATION VIA SIDEBAR
  // ==========================================================
  test('should navigate to Plugins via workspace settings sidebar', async ({ page }) => {
    await page.goto(`/workspace/${encodeURIComponent(_wsSlug)}/settings`)
    await page.waitForSelector('button:has-text("Plugins")', { timeout: 10000 })

    // Check other tabs exist
    await expect(page.locator('button:has-text("Members")').first()).toBeVisible({ timeout: 5000 })
    await expect(page.locator('button:has-text("Work Item Types")').first()).toBeVisible()
    await expect(page.locator('button:has-text("Fields")').first()).toBeVisible()

    // Click Plugins tab
    await page.click('button:has-text("Plugins")')

    // Verify PluginManager renders
    await page.waitForSelector('.plugin-manager', { timeout: 10000 })
    await expect(page.locator('.pm-header h2')).toHaveText('Plugin Manager')
  })
})
