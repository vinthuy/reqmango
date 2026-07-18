# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: automation-full-validation.spec.ts >> 自动化功能全面验证 - 前后端联动 >> 验证-09: 删除规则功能
- Location: e2e\automation-full-validation.spec.ts:313:3

# Error details

```
Error: expect(received).toBeDefined()

Received: undefined
```

# Test source

```ts
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
  297 |   test('验证-08: 规则执行历史记录', async ({ request }) => {
  298 |     console.log('🧪 验证-08: 规则执行历史')
  299 |     
  300 |     const historyRes = await request.get(`${BASE_API}/issues/${_issueId}/automation-executions`, {
  301 |       headers: { Authorization: `Bearer ${_token}` },
  302 |     })
  303 |     
  304 |     const status = historyRes.status()
  305 |     if (status === 200) {
  306 |       const history = await historyRes.json()
  307 |       console.log(`✅ 执行历史记录: ${Array.isArray(history) ? history.length : 0} 条记录`)
  308 |     } else {
  309 |       console.log(`⚠️ 执行历史API返回: ${status}`)
  310 |     }
  311 |   })
  312 | 
  313 |   test('验证-09: 删除规则功能', async ({ request }) => {
  314 |     console.log('🧪 验证-09: 删除规则')
  315 |     
  316 |     const res = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
  317 |       headers: { Authorization: `Bearer ${_token}` },
  318 |     })
  319 |     const rules = await res.json()
  320 |     const rulesArray = Array.isArray(rules) ? rules : (rules.data || [])
  321 |     const targetRule = rulesArray.find((r: any) => r.name === '创建时自动加评论')
  322 |     
> 323 |     expect(targetRule).toBeDefined()
      |                        ^ Error: expect(received).toBeDefined()
  324 |     
  325 |     await request.delete(`${BASE_API}/projects/${_projectId}/automations/${targetRule.id}`, {
  326 |       headers: { Authorization: `Bearer ${_token}` },
  327 |     })
  328 |     
  329 |     const res2 = await request.get(`${BASE_API}/projects/${_projectId}/automations`, {
  330 |       headers: { Authorization: `Bearer ${_token}` },
  331 |     })
  332 |     const rules2 = await res2.json()
  333 |     const rulesArray2 = Array.isArray(rules2) ? rules2 : (rules2.data || [])
  334 |     const deletedRule = rulesArray2.find((r: any) => r.id === targetRule.id)
  335 |     
  336 |     expect(deletedRule).toBeUndefined()
  337 |     console.log('✅ 删除规则功能正常')
  338 |   })
  339 | })
```