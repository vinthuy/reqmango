import { test, expect } from '@playwright/test'

/**
 * ReqMango 自动化功能端到端测试
 * 
 * 前置条件:
 * 1. 后端服务运行在 http://localhost:8000
 * 2. 前端服务运行在 http://localhost:5173
 * 3. 数据库中已有测试数据（项目 ID=1, 工作空间 ID=1）
 */

// 测试配置
const BASE_URL = 'http://localhost:5173'
const API_URL = 'http://localhost:8000'
const TEST_TOKEN = 'YOUR_JWT_TOKEN_HERE' // 替换为实际 Token
const PROJECT_ID = 1
const WORKSPACE_SLUG = 'default'

test.describe('ReqMango 自动化功能 E2E 测试', () => {
  
  /**
   * E2E-01: 创建规则并验证 issue_created 触发
   */
  test('E2E-01: 创建高优先级自动分配规则并验证触发', async ({ page }) => {
    console.log('🧪 开始执行 E2E-01: 创建规则并验证 issue_created 触发')
    
    // 步骤 1: 登录（简化版，假设已有 Token）
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/settings`)
    await page.waitForLoadState('networkidle')
    
    // 步骤 2: 导航到自动化标签页
    await page.click('text=🤖 自动化')
    await page.waitForSelector('button:has-text("新建自动化规则")')
    
    // 步骤 3: 打开创建规则模态框
    await page.click('button:has-text("新建自动化规则")')
    await page.waitForSelector('text=创建自动化规则')
    
    // 步骤 4: 填写规则表单
    await page.fill('input[placeholder*="规则名称"]', '高优先级自动分配')
    await page.fill('input[placeholder*="描述"]', '当工作项优先级为 high 时，自动分配给项目负责人')
    await page.selectOption('select', 'issue_created')
    await page.fill('textarea[placeholder*="条件"]', '[{"field":"priority","operator":"equals","value":"high"}]')
    await page.fill('textarea[placeholder*="动作"]', '[{"type":"assign_to","value":1}]')
    
    // 步骤 5: 提交表单
    await page.click('button:has-text("创建")')
    await page.waitForTimeout(1000) // 等待异步操作完成
    
    // 验证 1: 检查 Toast 提示
    const toastVisible = await page.isVisible('text=自动化规则创建成功')
    expect(toastVisible).toBeTruthy()
    console.log('✅ 规则创建成功，Toast 提示显示')
    
    // 验证 2: 检查规则列表中新增记录
    const ruleRow = page.locator('tr:has-text("高优先级自动分配")')
    await expect(ruleRow).toBeVisible()
    console.log('✅ 规则已出现在列表中')
    
    // 验证 3: 通过 API 检查规则详情
    const response = await page.request.get(`${API_URL}/api/v1/projects/${PROJECT_ID}/automations`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const rules = await response.json()
    const createdRule = rules.find((r: any) => r.name === '高优先级自动分配')
    
    expect(createdRule).toBeDefined()
    expect(createdRule.is_enabled).toBe(true)
    expect(createdRule.execution_count).toBe(0)
    console.log(`✅ 规则已保存到数据库，ID=${createdRule.id}, is_enabled=true, execution_count=0`)
    
    const ruleId = createdRule.id
    
    // 步骤 6: 导航到工作项列表页面
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/issues`)
    await page.waitForLoadState('networkidle')
    
    // 步骤 7: 创建工作项触发规则
    await page.click('button:has-text("新建")')
    await page.waitForSelector('text=新建工作项')
    
    await page.fill('input[name="name"]', '测试自动化触发')
    await page.selectOption('select[name="priority"]', 'high')
    await page.selectOption('select[name="state_id"]', '1')
    
    await page.click('button:has-text("创建")')
    await page.waitForTimeout(1000)
    
    // 验证 4: 检查工作项创建成功
    const issueToast = await page.isVisible('text=工作项创建成功')
    expect(issueToast).toBeTruthy()
    console.log('✅ 工作项创建成功')
    
    // 验证 5: 通过 API 检查工作项的负责人
    const issuesResponse = await page.request.get(`${API_URL}/api/v1/issues?project_id=${PROJECT_ID}`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const issues = await issuesResponse.json()
    const createdIssue = issues.find((i: any) => i.name === '测试自动化触发')
    
    expect(createdIssue).toBeDefined()
    expect(createdIssue.assignees).toHaveLength(1)
    expect(createdIssue.assignees[0].user_id).toBe(1)
    console.log(`✅ 工作项已分配给用户 ID=1, assignees.length=${createdIssue.assignees.length}`)
    
    // 验证 6: 检查规则执行计数
    const ruleResponse = await page.request.get(`${API_URL}/api/v1/projects/${PROJECT_ID}/automations/${ruleId}`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const updatedRule = await ruleResponse.json()
    
    expect(updatedRule.execution_count).toBe(1)
    console.log(`✅ 规则执行计数从 0 增加到 1`)
    
    // 验证 7: 查询执行历史
    const historyResponse = await page.request.get(`${API_URL}/api/v1/issues/${createdIssue.id}/automation-history?limit=1`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const history = await historyResponse.json()
    
    expect(history).toHaveLength(1)
    expect(history[0].rule_id).toBe(ruleId)
    expect(history[0].trigger_type).toBe('issue_created')
    expect(history[0].status).toBe('success')
    expect(history[0].duration).toBeLessThan(100)
    console.log(`✅ 执行历史记录完整，status=success, duration=${history[0].duration}ms`)
    
    console.log('🎉 E2E-01 测试通过！')
  })
  
  /**
   * E2E-02: 状态变更规则触发测试
   */
  test('E2E-02: 创建状态变更规则并验证触发', async ({ page }) => {
    console.log('🧪 开始执行 E2E-02: 状态变更规则触发测试')
    
    // 步骤 1: 创建状态变更规则
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/settings?tab=automations`)
    await page.waitForLoadState('networkidle')
    
    await page.click('button:has-text("新建自动化规则")')
    await page.fill('input[placeholder*="规则名称"]', '完成时添加评论')
    await page.selectOption('select', 'state_changed')
    await page.fill('textarea[placeholder*="条件"]', '[{"field":"state_group","operator":"equals","value":"done"}]')
    await page.fill('textarea[placeholder*="动作"]', '[{"type":"add_comment","value":"✅ 工作项已完成"}]')
    await page.click('button:has-text("创建")')
    await page.waitForTimeout(1000)
    
    console.log('✅ 状态变更规则创建成功')
    
    // 步骤 2: 获取规则 ID
    const rulesResponse = await page.request.get(`${API_URL}/api/v1/projects/${PROJECT_ID}/automations`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const rules = await rulesResponse.json()
    const stateChangeRule = rules.find((r: any) => r.name === '完成时添加评论')
    expect(stateChangeRule).toBeDefined()
    
    // 步骤 3: 更新工作项状态到 done
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/issues`)
    await page.waitForLoadState('networkidle')
    
    // 点击第一个工作项的编辑按钮
    await page.click('tbody tr:nth-child(1) td:last-child button')
    await page.click('text=编辑')
    await page.waitForSelector('text=编辑工作项')
    
    await page.selectOption('select[name="state_id"]', '3') // done 状态
    await page.click('button:has-text("保存")')
    await page.waitForTimeout(1000)
    
    console.log('✅ 工作项状态更新成功')
    
    // 步骤 4: 验证评论已添加
    const issuesResponse = await page.request.get(`${API_URL}/api/v1/issues?project_id=${PROJECT_ID}`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const issues = await issuesResponse.json()
    const firstIssue = issues[0]
    
    const commentsResponse = await page.request.get(`${API_URL}/api/v1/comments/issue/${firstIssue.id}`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const comments = await commentsResponse.json()
    const lastComment = comments[comments.length - 1]
    
    expect(lastComment.body).toBe('✅ 工作项已完成')
    console.log(`✅ 评论已成功添加: "${lastComment.body}"`)
    
    // 步骤 5: 验证执行历史
    const historyResponse = await page.request.get(`${API_URL}/api/v1/issues/${firstIssue.id}/automation-history?limit=1`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const history = await historyResponse.json()
    
    expect(history[0].trigger_type).toBe('state_changed')
    expect(history[0].status).toBe('success')
    
    const context = JSON.parse(history[0].context_json)
    expect(context).toHaveProperty('old_state')
    expect(context).toHaveProperty('new_state')
    expect(context).toHaveProperty('state_group')
    console.log(`✅ 执行历史包含 old_state, new_state, state_group`)
    
    console.log('🎉 E2E-02 测试通过！')
  })
  
  /**
   * E2E-03: 禁用规则后不应触发
   */
  test('E2E-03: 禁用规则后验证不触发', async ({ page }) => {
    console.log('🧪 开始执行 E2E-03: 禁用规则后验证不触发')
    
    // 步骤 1: 禁用规则
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/settings?tab=automations`)
    await page.waitForLoadState('networkidle')
    
    const ruleRow = page.locator('tr:has-text("高优先级自动分配")')
    await ruleRow.locator('button:has-text("禁用")').click()
    await page.waitForTimeout(1000)
    
    console.log('✅ 规则已禁用')
    
    // 步骤 2: 创建工作项（priority=high）
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/issues`)
    await page.click('button:has-text("新建")')
    await page.fill('input[name="name"]', '测试禁用规则')
    await page.selectOption('select[name="priority"]', 'high')
    await page.click('button:has-text("创建")')
    await page.waitForTimeout(1000)
    
    console.log('✅ 工作项创建成功')
    
    // 步骤 3: 验证规则未执行
    const issuesResponse = await page.request.get(`${API_URL}/api/v1/issues?project_id=${PROJECT_ID}`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const issues = await issuesResponse.json()
    const testIssue = issues.find((i: any) => i.name === '测试禁用规则')
    
    expect(testIssue.assignees).toHaveLength(0)
    console.log(`✅ 工作项未被分配负责人，assignees.length=${testIssue.assignees.length}`)
    
    // 步骤 4: 检查规则执行计数未增加
    const rulesResponse = await page.request.get(`${API_URL}/api/v1/projects/${PROJECT_ID}/automations`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const rules = await rulesResponse.json()
    const disabledRule = rules.find((r: any) => r.name === '高优先级自动分配')
    
    expect(disabledRule.execution_count).toBe(1) // 仍为 1，未增加
    console.log(`✅ 规则执行计数保持为 1，未增加`)
    
    console.log('🎉 E2E-03 测试通过！')
  })
  
  /**
   * E2E-04: 复杂条件规则（正则匹配）测试
   */
  test('E2E-04: 正则匹配条件规则测试', async ({ page }) => {
    console.log('🧪 开始执行 E2E-04: 正则匹配条件规则测试')
    
    // 步骤 1: 创建正则匹配规则
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/settings?tab=automations`)
    await page.waitForLoadState('networkidle')
    
    await page.click('button:has-text("新建自动化规则")')
    await page.fill('input[placeholder*="规则名称"]', 'Bug 自动标记为紧急')
    await page.selectOption('select', 'issue_created')
    await page.fill('textarea[placeholder*="条件"]', '[{"field":"name","operator":"matches_regex","value":"^BUG-\\\\d+"}]')
    await page.fill('textarea[placeholder*="动作"]', '[{"type":"set_priority","value":"urgent"}]')
    await page.click('button:has-text("创建")')
    await page.waitForTimeout(1000)
    
    console.log('✅ 正则匹配规则创建成功')
    
    // 步骤 2: 创建工作项（名称以 BUG- 开头）
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/issues`)
    await page.click('button:has-text("新建")')
    await page.fill('input[name="name"]', 'BUG-123 登录失败')
    await page.selectOption('select[name="priority"]', 'medium')
    await page.click('button:has-text("创建")')
    await page.waitForTimeout(1000)
    
    console.log('✅ 工作项创建成功')
    
    // 步骤 3: 验证优先级已变更为 urgent
    const issuesResponse = await page.request.get(`${API_URL}/api/v1/issues?project_id=${PROJECT_ID}`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const issues = await issuesResponse.json()
    const bugIssue = issues.find((i: any) => i.name === 'BUG-123 登录失败')
    
    expect(bugIssue.priority).toBe('urgent')
    console.log(`✅ 优先级从 medium 自动变更为 urgent`)
    
    console.log('🎉 E2E-04 测试通过！')
  })
  
  /**
   * E2E-05: BulkUpdate 触发自动化测试
   */
  test('E2E-05: 批量更新触发自动化测试', async ({ page }) => {
    console.log('🧪 开始执行 E2E-05: 批量更新触发自动化测试')
    
    // 步骤 1: 创建批量更新规则
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/settings?tab=automations`)
    await page.waitForLoadState('networkidle')
    
    await page.click('button:has-text("新建自动化规则")')
    await page.fill('input[placeholder*="规则名称"]', '批量更新时设置优先级')
    await page.selectOption('select', 'state_changed')
    await page.fill('textarea[placeholder*="条件"]', '[]')
    await page.fill('textarea[placeholder*="动作"]', '[{"type":"set_priority","value":"high"}]')
    await page.click('button:has-text("创建")')
    await page.waitForTimeout(1000)
    
    console.log('✅ 批量更新规则创建成功')
    
    // 步骤 2: 创建 3 个工作项
    const issueIds: number[] = []
    for (let i = 1; i <= 3; i++) {
      await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/issues`)
      await page.click('button:has-text("新建")')
      await page.fill('input[name="name"]', `批量测试 ${i}`)
      await page.selectOption('select[name="state_id"]', '1')
      await page.click('button:has-text("创建")')
      await page.waitForTimeout(500)
      
      // 获取刚创建的工作项 ID
      const issuesResponse = await page.request.get(`${API_URL}/api/v1/issues?project_id=${PROJECT_ID}`, {
        headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
      })
      const issues = await issuesResponse.json()
      const newIssue = issues.find((i: any) => i.name === `批量测试 ${i}`)
      if (newIssue) {
        issueIds.push(newIssue.id)
      }
    }
    
    console.log(`✅ 创建了 3 个工作项，IDs: ${issueIds.join(', ')}`)
    
    // 步骤 3: 批量选中并更新状态
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/issues`)
    await page.waitForLoadState('networkidle')
    
    // 全选 3 个工作项
    for (let i = 1; i <= 3; i++) {
      await page.check(`tbody tr:nth-child(${i}) input[type="checkbox"]`)
    }
    
    // 点击"批量修改状态"
    await page.click('button:has-text("批量修改状态")')
    await page.click('button:has-text("进行中")')
    await page.waitForTimeout(1000)
    
    console.log('✅ 批量状态更新成功')
    
    // 步骤 4: 验证所有工作项优先级已变为 high
    for (const issueId of issueIds) {
      const issueResponse = await page.request.get(`${API_URL}/api/v1/issues/${issueId}`, {
        headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
      })
      const issue = await issueResponse.json()
      
      expect(issue.priority).toBe('high')
    }
    
    console.log('✅ 所有工作项 priority=high')
    
    // 步骤 5: 检查规则执行计数
    const rulesResponse = await page.request.get(`${API_URL}/api/v1/projects/${PROJECT_ID}/automations`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const rules = await rulesResponse.json()
    const bulkRule = rules.find((r: any) => r.name === '批量更新时设置优先级')
    
    expect(bulkRule.execution_count).toBe(3)
    console.log(`✅ 规则执行计数=3（每个工作项触发一次）`)
    
    console.log('🎉 E2E-05 测试通过！')
  })
  
  /**
   * E2E-06: 循环依赖检测测试
   */
  test('E2E-06: 循环依赖检测测试', async ({ page }) => {
    console.log('🧪 开始执行 E2E-06: 循环依赖检测测试')
    
    // 步骤 1: 创建规则 A
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/settings?tab=automations`)
    await page.waitForLoadState('networkidle')
    
    await page.click('button:has-text("新建自动化规则")')
    await page.fill('input[placeholder*="规则名称"]', '规则A: 高优先级→进行中')
    await page.selectOption('select', 'issue_created')
    await page.fill('textarea[placeholder*="条件"]', '[{"field":"priority","operator":"equals","value":"high"}]')
    await page.fill('textarea[placeholder*="动作"]', '[{"type":"set_field","field":"state_id","value":2}]')
    await page.click('button:has-text("创建")')
    await page.waitForTimeout(1000)
    
    console.log('✅ 规则 A 创建成功')
    
    // 步骤 2: 创建规则 B
    await page.click('button:has-text("新建自动化规则")')
    await page.fill('input[placeholder*="规则名称"]', '规则B: 进行中→高优先级')
    await page.selectOption('select', 'state_changed')
    await page.fill('textarea[placeholder*="条件"]', '[{"field":"new_state","operator":"equals","value":"2"}]')
    await page.fill('textarea[placeholder*="动作"]', '[{"type":"set_priority","value":"high"}]')
    await page.click('button:has-text("创建")')
    await page.waitForTimeout(1000)
    
    console.log('✅ 规则 B 创建成功')
    
    // 步骤 3: 创建工作项触发循环
    await page.goto(`${BASE_URL}/workspace/${WORKSPACE_SLUG}/project/${PROJECT_ID}/issues`)
    await page.click('button:has-text("新建")')
    await page.fill('input[name="name"]', '循环依赖测试')
    await page.selectOption('select[name="priority"]', 'high')
    await page.selectOption('select[name="state_id"]', '1')
    await page.click('button:has-text("创建")')
    await page.waitForTimeout(2000) // 等待循环检测生效
    
    console.log('✅ 工作项创建成功，等待循环检测')
    
    // 步骤 4: 查询执行历史
    const issuesResponse = await page.request.get(`${API_URL}/api/v1/issues?project_id=${PROJECT_ID}`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const issues = await issuesResponse.json()
    const cycleIssue = issues.find((i: any) => i.name === '循环依赖测试')
    
    const historyResponse = await page.request.get(`${API_URL}/api/v1/issues/${cycleIssue.id}/automation-history?limit=15`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const history = await historyResponse.json()
    
    // 验证：前 10 次成功，第 11 次被跳过
    const skippedCount = history.filter((h: any) => h.status === 'skipped').length
    const successCount = history.filter((h: any) => h.status === 'success').length
    
    expect(successCount).toBeLessThanOrEqual(10)
    expect(skippedCount).toBeGreaterThan(0)
    
    console.log(`✅ 循环检测生效：success=${successCount}, skipped=${skippedCount}`)
    
    console.log('🎉 E2E-06 测试通过！')
  })
  
  /**
   * E2E-07: 执行历史查询测试
   */
  test('E2E-07: 执行历史查询测试', async ({ page }) => {
    console.log('🧪 开始执行 E2E-07: 执行历史查询测试')
    
    // 步骤 1: 获取某工作项的执行历史
    const issuesResponse = await page.request.get(`${API_URL}/api/v1/issues?project_id=${PROJECT_ID}`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const issues = await issuesResponse.json()
    const firstIssue = issues[0]
    
    const historyResponse = await page.request.get(`${API_URL}/api/v1/issues/${firstIssue.id}/automation-history?limit=10`, {
      headers: { 'Authorization': `Bearer ${TEST_TOKEN}` }
    })
    const history = await historyResponse.json()
    
    // 验证 1: 返回记录数 ≤ limit
    expect(history.length).toBeLessThanOrEqual(10)
    console.log(`✅ 返回记录数=${history.length} ≤ 10`)
    
    // 验证 2: 每条记录包含必需字段
    for (const record of history) {
      expect(record).toHaveProperty('id')
      expect(record).toHaveProperty('rule_id')
      expect(record).toHaveProperty('issue_id')
      expect(record).toHaveProperty('trigger_type')
      expect(record).toHaveProperty('context_json')
      expect(record).toHaveProperty('actions_taken')
      expect(record).toHaveProperty('status')
      expect(record).toHaveProperty('duration')
      expect(record).toHaveProperty('executed_at')
    }
    
    console.log('✅ 所有记录包含必需字段')
    
    // 验证 3: duration < 100ms
    for (const record of history) {
      expect(record.duration).toBeLessThan(100)
    }
    
    console.log('✅ 所有记录 duration < 100ms')
    
    // 验证 4: context_json 格式正确
    for (const record of history) {
      const context = JSON.parse(record.context_json)
      expect(context).toHaveProperty('issue_id')
    }
    
    console.log('✅ context_json 格式正确')
    
    console.log('🎉 E2E-07 测试通过！')
  })
})
