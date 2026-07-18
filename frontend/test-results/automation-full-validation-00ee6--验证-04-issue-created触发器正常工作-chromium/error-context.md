# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: automation-full-validation.spec.ts >> 自动化功能全面验证 - 前后端联动 >> 验证-04: issue_created触发器正常工作
- Location: e2e\automation-full-validation.spec.ts:169:3

# Error details

```
Error: expect(received).toBeDefined()

Received: undefined
```

# Test source

```ts
  96  |   })
  97  | 
  98  |   test('验证-02: 编辑规则功能正常', async ({ page, request }) => {
  99  |     console.log('🧪 验证-02: 编辑规则功能')
  100 |     
  101 |     await navigateToAutomation(page)
  102 |     
  103 |     const res = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
  104 |       headers: { Authorization: `Bearer ${_token}` },
  105 |     })
  106 |     const rules = await res.json()
  107 |     const targetRule = Array.isArray(rules) ? rules.find((r: any) => r.name === 'Trigger类型验证规则') : null
  108 |     expect(targetRule).toBeDefined()
  109 |     
  110 |     await page.click('button:text("✏️")')
  111 |     await page.waitForTimeout(500)
  112 |     
  113 |     await page.fill('input[placeholder*="自动分配"]', '编辑后的规则名称')
  114 |     await page.fill('textarea[placeholder*="priority"]', '[{"field":"priority","operator":"equals","value":"urgent"}]')
  115 |     
  116 |     await page.locator('div.fixed').locator('button:has-text("更新")').click()
  117 |     await page.waitForTimeout(1500)
  118 |     
  119 |     const res2 = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
  120 |       headers: { Authorization: `Bearer ${_token}` },
  121 |     })
  122 |     const rules2 = await res2.json()
  123 |     const updatedRule = Array.isArray(rules2) ? rules2.find((r: any) => r.name === '编辑后的规则名称') : null
  124 |     
  125 |     expect(updatedRule).toBeDefined()
  126 |     expect(updatedRule.conditions).toContain('urgent')
  127 |     console.log('✅ 编辑规则功能正常')
  128 |   })
  129 | 
  130 |   test('验证-03: 启用/禁用切换按钮功能', async ({ request }) => {
  131 |     console.log('🧪 验证-03: 启用/禁用切换')
  132 |     
  133 |     const res = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
  134 |       headers: { Authorization: `Bearer ${_token}` },
  135 |     })
  136 |     const rules = await res.json()
  137 |     const targetRule = Array.isArray(rules) ? rules.find((r: any) => r.name === '编辑后的规则名称') : null
  138 |     expect(targetRule).toBeDefined()
  139 |     
  140 |     await request.put(`${BASE_API}/projects/${_projectId}/automations/${targetRule.id}`, {
  141 |       data: { is_enabled: false },
  142 |       headers: { Authorization: `Bearer ${_token}` },
  143 |     })
  144 |     
  145 |     const res2 = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
  146 |       headers: { Authorization: `Bearer ${_token}` },
  147 |     })
  148 |     const rules2 = await res2.json()
  149 |     const disabledRule = Array.isArray(rules2) ? rules2.find((r: any) => r.id === targetRule.id) : null
  150 |     
  151 |     expect(disabledRule.is_enabled).toBe(false)
  152 |     console.log('✅ 禁用规则成功')
  153 |     
  154 |     await request.put(`${BASE_API}/projects/${_projectId}/automations/${targetRule.id}`, {
  155 |       data: { is_enabled: true },
  156 |       headers: { Authorization: `Bearer ${_token}` },
  157 |     })
  158 |     
  159 |     const res3 = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
  160 |       headers: { Authorization: `Bearer ${_token}` },
  161 |     })
  162 |     const rules3 = await res3.json()
  163 |     const enabledRule = Array.isArray(rules3) ? rules3.find((r: any) => r.id === targetRule.id) : null
  164 |     
  165 |     expect(enabledRule.is_enabled).toBe(true)
  166 |     console.log('✅ 启用规则成功')
  167 |   })
  168 | 
  169 |   test('验证-04: issue_created触发器正常工作', async ({ request }) => {
  170 |     console.log('🧪 验证-04: issue_created触发器')
  171 |     
  172 |     await createRuleViaAPI(
  173 |       request,
  174 |       '创建时自动加评论',
  175 |       'issue_created',
  176 |       '[]',
  177 |       '[{"type":"add_comment","value":"🎉 新工作项已创建"}]'
  178 |     )
  179 |     
  180 |     const issueRes = await request.post(`${BASE_API}/issues?project_id=${_projectId}`, {
  181 |       data: { name: '测试issue_created触发', description: '测试触发器' },
  182 |       headers: { Authorization: `Bearer ${_token}` },
  183 |     })
  184 |     const issueData = await issueRes.json()
  185 |     _issueId = issueData.id || issueData.data?.id
  186 |     
  187 |     await new Promise(resolve => setTimeout(resolve, 2000))
  188 |     
  189 |     const commentsRes = await request.get(`${BASE_API}/comments/issue/${_issueId}`, {
  190 |       headers: { Authorization: `Bearer ${_token}` },
  191 |     })
  192 |     const commentsData = await commentsRes.json()
  193 |     const comments = commentsData.comments || commentsData.data || commentsData
  194 |     
  195 |     const autoComment = Array.isArray(comments) ? comments.find((c: any) => c.body === '🎉 新工作项已创建') : null
> 196 |     expect(autoComment).toBeDefined()
      |                         ^ Error: expect(received).toBeDefined()
  197 |     console.log('✅ issue_created触发器正常工作')
  198 |   })
  199 | 
  200 |   test('验证-05: comment_added触发器正常工作', async ({ request }) => {
  201 |     console.log('🧪 验证-05: comment_added触发器')
  202 |     
  203 |     await createRuleViaAPI(
  204 |       request,
  205 |       '评论时自动标记',
  206 |       'comment_added',
  207 |       '[{"field":"comment","operator":"contains","value":"bug"}]',
  208 |       '[{"type":"set_priority","value":"high"}]'
  209 |     )
  210 |     
  211 |     await request.post(`${BASE_API}/comments`, {
  212 |       data: { issue_id: _issueId, body: '发现一个bug需要修复' },
  213 |       headers: { Authorization: `Bearer ${_token}` },
  214 |     })
  215 |     
  216 |     await new Promise(resolve => setTimeout(resolve, 2000))
  217 |     
  218 |     const issueRes = await request.get(`${BASE_API}/issues/${_issueId}`, {
  219 |       headers: { Authorization: `Bearer ${_token}` },
  220 |     })
  221 |     const issueData = await issueRes.json()
  222 |     const issue = issueData.data || issueData
  223 |     
  224 |     expect(issue.priority).toBe('high')
  225 |     console.log('✅ comment_added触发器正常工作')
  226 |   })
  227 | 
  228 |   test('验证-06: issue_updated触发器（非状态变化）正常工作', async ({ request }) => {
  229 |     console.log('🧪 验证-06: issue_updated触发器（非状态变化）')
  230 |     
  231 |     await createRuleViaAPI(
  232 |       request,
  233 |       '更新时自动加评论',
  234 |       'issue_updated',
  235 |       '[]',
  236 |       '[{"type":"add_comment","value":"🔄 工作项已更新"}]'
  237 |     )
  238 |     
  239 |     await request.put(`${BASE_API}/issues/${_issueId}`, {
  240 |       data: { name: '测试自动化触发 - 已更新' },
  241 |       headers: { Authorization: `Bearer ${_token}` },
  242 |     })
  243 |     
  244 |     await new Promise(resolve => setTimeout(resolve, 2000))
  245 |     
  246 |     const commentsRes = await request.get(`${BASE_API}/comments/issue/${_issueId}`, {
  247 |       headers: { Authorization: `Bearer ${_token}` },
  248 |     })
  249 |     const commentsData = await commentsRes.json()
  250 |     const comments = commentsData.comments || commentsData.data || commentsData
  251 |     
  252 |     const updateComment = Array.isArray(comments) ? comments.find((c: any) => c.body === '🔄 工作项已更新') : null
  253 |     expect(updateComment).toBeDefined()
  254 |     console.log('✅ issue_updated触发器（非状态变化）正常工作')
  255 |   })
  256 | 
  257 |   test('验证-07: state_changed触发器正常工作', async ({ request }) => {
  258 |     console.log('🧪 验证-07: state_changed触发器')
  259 |     
  260 |     await createRuleViaAPI(
  261 |       request,
  262 |       '状态变更时自动完成',
  263 |       'state_changed',
  264 |       '[{"field":"state_group","operator":"equals","value":"done"}]',
  265 |       '[{"type":"add_comment","value":"✅ 工作项已完成"}]'
  266 |     )
  267 |     
  268 |     const statesRes = await request.get(`${BASE_API}/projects/${_projectId}/settings/states`, {
  269 |       headers: { Authorization: `Bearer ${_token}` },
  270 |     })
  271 |     const statesData = await statesRes.json()
  272 |     const states = statesData.data || statesData
  273 |     const doneState = Array.isArray(states) ? states.find((s: any) => s.group === 'done') : null
  274 |     
  275 |     if (doneState) {
  276 |       await request.put(`${BASE_API}/issues/${_issueId}`, {
  277 |         data: { state_id: doneState.id },
  278 |         headers: { Authorization: `Bearer ${_token}` },
  279 |       })
  280 |       
  281 |       await new Promise(resolve => setTimeout(resolve, 2000))
  282 |       
  283 |       const commentsRes = await request.get(`${BASE_API}/comments/issue/${_issueId}`, {
  284 |         headers: { Authorization: `Bearer ${_token}` },
  285 |       })
  286 |       const commentsData = await commentsRes.json()
  287 |       const comments = commentsData.comments || commentsData.data || commentsData
  288 |       
  289 |       const doneComment = Array.isArray(comments) ? comments.find((c: any) => c.body === '✅ 工作项已完成') : null
  290 |       expect(doneComment).toBeDefined()
  291 |       console.log('✅ state_changed触发器正常工作')
  292 |     } else {
  293 |       console.log('⚠️ 未找到done状态，跳过此测试')
  294 |     }
  295 |   })
  296 | 
```