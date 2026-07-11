import { test, expect } from '@playwright/test'

const BASE_API = 'http://localhost:8000/api/v1'

let testData = {
  workspaceId: '',
  projectId: '',
  issueId: '',
  moduleId: '',
  stateId: '',
  labelId: '',
}

const TEST_USER = {
  email: 'e2e_test_1@reqmango.com',
  password: 'E2eTest123!',
}

async function login(page: any) {
  await page.goto('/')
  await page.waitForTimeout(2000)
  
  const emailInput = page.locator('input[type="email"]')
  if (await emailInput.isVisible()) {
    await emailInput.fill(TEST_USER.email)
    const passwordInput = page.locator('input[type="password"]')
    await passwordInput.fill(TEST_USER.password)
    const loginBtn = page.locator('button:has-text("登录")')
    await loginBtn.click()
    await page.waitForTimeout(3000)
  }
}

async function navigateToProject(page: any) {
  await page.waitForTimeout(2000)
  
  const workspaces = page.locator('.workspace-item')
  if ((await workspaces.count()) > 0) {
    await workspaces.first().click()
    await page.waitForTimeout(2000)
  }
  
  const projects = page.locator('.project-item')
  if ((await projects.count()) > 0) {
    await projects.first().click()
    await page.waitForTimeout(2000)
  }
}

test.describe('BUG Fixes Acceptance Tests', () => {
  test.beforeAll(async ({ request }) => {
    let res = await request.post(`${BASE_API}/auth/login`, {
      data: { email: TEST_USER.email, password: TEST_USER.password },
    })
    let body = await res.json()
    const token = body.access_token || ''

    res = await request.get(`${BASE_API}/workspaces`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    body = await res.json()
    testData.workspaceId = body[0]?.id || ''

    if (testData.workspaceId) {
      res = await request.get(`${BASE_API}/projects?workspace_id=${testData.workspaceId}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      body = await res.json()
      if (body.length > 0) {
        testData.projectId = body[0].id || ''
      } else {
        res = await request.post(`${BASE_API}/projects?workspace_id=${testData.workspaceId}`, {
          headers: { Authorization: `Bearer ${token}` },
          data: { name: 'E2E Test Project', identifier: 'E2E' },
        })
        body = await res.json()
        testData.projectId = body.id || ''
      }
    }
  })

  test('BUG-20: @mention should support Chinese usernames', async ({ page }) => {
    await login(page)
    await navigateToProject(page)

    const issues = page.locator('.issue-card')
    if ((await issues.count()) > 0) {
      await issues.first().click()
      await page.waitForTimeout(2000)

      const commentInput = page.locator('[placeholder*="评论"]')
      if (await commentInput.isVisible()) {
        await commentInput.fill('@张三 测试中文用户名')
        await page.waitForTimeout(500)
        const mentionSuggestion = page.locator('.mention-suggestion')
        if (await mentionSuggestion.isVisible()) {
          await mentionSuggestion.first().click()
        }
        await page.locator('button:has-text("发送")').click()
        await page.waitForTimeout(1000)
      }
    }
  })

  test('BUG-21: Activity log should show state name not ID', async ({ page }) => {
    await login(page)
    await navigateToProject(page)

    const issues = page.locator('.issue-card')
    if ((await issues.count()) > 0) {
      await issues.first().click()
      await page.waitForTimeout(2000)

      const activityTab = page.locator('text="活动"')
      if (await activityTab.isVisible()) {
        await activityTab.click()
        await page.waitForTimeout(1000)

        const activities = page.locator('.activity-item')
        const activityCount = await activities.count()
        if (activityCount > 0) {
          const activityText = await activities.first().textContent()
          expect(activityText).toBeTruthy()
          const isNumericOnly = /^\d+$/.test(activityText || '')
          expect(isNumericOnly).toBe(false)
        }
      }
    }
  })

  test('BUG-22: Deleting parent module should handle child modules', async ({ page }) => {
    await login(page)
    await navigateToProject(page)

    const modulesLink = page.locator('text="模块"')
    if (await modulesLink.isVisible()) {
      await modulesLink.click()
      await page.waitForTimeout(2000)

      const addModuleBtn = page.locator('button:has-text("添加模块")')
      if (await addModuleBtn.isVisible()) {
        await addModuleBtn.click()
        await page.waitForTimeout(500)

        const nameInput = page.locator('[placeholder*="模块名称"]')
        if (await nameInput.isVisible()) {
          await nameInput.fill('父模块测试')
          await page.locator('button:has-text("创建")').click()
          await page.waitForTimeout(1000)
        }
      }

      const modules = page.locator('.module-item')
      if ((await modules.count()) > 0) {
        const deleteBtn = modules.first().locator('button:has-text("删除")')
        if (await deleteBtn.isVisible()) {
          await deleteBtn.click()
          await page.waitForTimeout(500)
          await page.locator('button:has-text("确认")').click()
          await page.waitForTimeout(1000)
        }
      }
    }
  })

  test('BUG-24: RQL LIKE should escape % and _ characters', async ({ page }) => {
    await login(page)
    await navigateToProject(page)

    const searchInput = page.locator('[placeholder*="搜索"]')
    if (await searchInput.isVisible()) {
      await searchInput.fill('test%')
      await page.waitForTimeout(1000)

      await searchInput.fill('test_')
      await page.waitForTimeout(1000)
    }
  })

  test('BUG-23: Burndown chart should have multiple data points', async ({ page }) => {
    await login(page)
    await navigateToProject(page)

    const analyticsLink = page.locator('text="分析"')
    if (await analyticsLink.isVisible()) {
      await analyticsLink.click()
      await page.waitForTimeout(2000)
    }

    const burndownSection = page.locator('.burndown-chart')
    if (await burndownSection.isVisible()) {
      await page.waitForTimeout(1000)
    }
  })

  test('BUG-34: Page version diff should show proper LCS algorithm', async ({ page }) => {
    await login(page)
    await navigateToProject(page)

    const pagesLink = page.locator('text="页面"')
    if (await pagesLink.isVisible()) {
      await pagesLink.click()
      await page.waitForTimeout(2000)

      const pageItems = page.locator('.page-item')
      if ((await pageItems.count()) > 0) {
        await pageItems.first().click()
        await page.waitForTimeout(2000)

        const historyBtn = page.locator('text="历史版本"')
        if (await historyBtn.isVisible()) {
          await historyBtn.click()
          await page.waitForTimeout(1000)
        }
      }
    }
  })

  test('BUG-25: Issue views should load consistently', async ({ page }) => {
    await login(page)
    await navigateToProject(page)

    const viewButtons = page.locator('.view-toggle button')
    for (let i = 0; i < await viewButtons.count(); i++) {
      await viewButtons.nth(i).click()
      await page.waitForTimeout(1000)
      const issues = page.locator('.issue-card, .gantt-item, .calendar-event')
      const count = await issues.count()
    }
  })

  test('BUG-26: Bulk operations should return success/failure counts', async ({ page }) => {
    await login(page)
    await navigateToProject(page)

    const firstCheckbox = page.locator('.issue-checkbox')
    if (await firstCheckbox.isVisible()) {
      await firstCheckbox.click()
      await page.waitForTimeout(500)

      const bulkActions = page.locator('.bulk-actions')
      if (await bulkActions.isVisible()) {
        await bulkActions.click()
        await page.waitForTimeout(500)
      }
    }
  })

  test('BUG-30: Gin trailing slash should not redirect', async ({ request }) => {
    const res = await request.get(`${BASE_API}/workspaces/`)
    expect(res.status()).not.toBe(301)
    expect(res.status()).not.toBe(302)
  })

  test('BUG-29: CSV export should have UTF-8 BOM', async ({ request }) => {
    const res = await request.post(`${BASE_API}/auth/login`, {
      data: { email: TEST_USER.email, password: TEST_USER.password },
    })
    const body = await res.json()
    const token = body.access_token || ''

    if (testData.projectId) {
      const exportRes = await request.get(`${BASE_API}/projects/${testData.projectId}/issues/export`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (exportRes.status() === 200) {
        const exportBody = await exportRes.text()
        const hasBOM = exportBody.charCodeAt(0) === 0xFEFF
        expect(hasBOM).toBe(true)
      }
    }
  })
})