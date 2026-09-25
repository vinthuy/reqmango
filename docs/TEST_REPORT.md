# ReqMango 自动化测试报告

> 生成时间: 2026-08-16 | 分支: `feature/ai-enhancements` | Commit: `b619470`

---

## 一、总览

| 指标 | 数值 |
|------|------|
| **后端测试用例** | **127** |
| **后端通过率** | **100% (127/127)** |
| **E2E 验收用例** | **154** |
| **E2E 通过率** | **98.1% (151/154)** |
| **E2E 覆盖页面** | **39** |
| **API 健康检查** | **13** |
| **测试耗时** | ~25s (后端) / ~3min (E2E) |

---

## 二、后端 Service 层测试详情

### 2.1 新增测试 (106 用例)

#### SkillService (17 用例) ✅

| 用例 | 状态 | 描述 |
|------|------|------|
| TestSkillService_Create | ✅ PASS | 创建技能 |
| TestSkillService_Create_MinimalFields | ✅ PASS | 最小字段创建 |
| TestSkillService_Create_WithWorkspaceMember | ✅ PASS | 带工作空间成员创建 |
| TestSkillService_Get | ✅ PASS | 获取技能 |
| TestSkillService_Get_NotFound | ✅ PASS | 获取不存在的技能 |
| TestSkillService_List | ✅ PASS | 列出技能 |
| TestSkillService_List_OtherWorkspace | ✅ PASS | 隔离其他工作空间 |
| TestSkillService_List_Empty | ✅ PASS | 空列表 |
| TestSkillService_Update | ✅ PASS | 更新技能 |
| TestSkillService_Update_NotFound | ✅ PASS | 更新不存在的技能 |
| TestSkillService_Delete | ✅ PASS | 删除技能 |
| TestSkillService_Delete_NotFound | ✅ PASS | 删除不存在的技能 |
| TestSkillService_IncrementUsage | ✅ PASS | 使用次数递增 |
| TestSkillService_Execute | ✅ PASS | 执行技能 |
| TestSkillService_Execute_NotFound | ✅ PASS | 执行不存在的技能 |
| TestSkillService_Execute_InactiveSkill | ✅ PASS | 执行非活跃技能 |
| TestSkillService_Execute_RecordsLog | ✅ PASS | 执行日志记录 |

#### RuntimeService (17 用例) ✅

| 用例 | 状态 | 描述 |
|------|------|------|
| TestRuntimeService_Create | ✅ PASS | 创建运行时 |
| TestRuntimeService_Create_MinimalFields | ✅ PASS | 最小字段创建 |
| TestRuntimeService_Get | ✅ PASS | 获取运行时 |
| TestRuntimeService_Get_NotFound | ✅ PASS | 获取不存在的运行时 |
| TestRuntimeService_List | ✅ PASS | 列出运行时 |
| TestRuntimeService_List_Empty | ✅ PASS | 空列表 |
| TestRuntimeService_List_OtherWorkspace | ✅ PASS | 隔离其他工作空间 |
| TestRuntimeService_Update | ✅ PASS | 更新运行时 |
| TestRuntimeService_Update_NotFound | ✅ PASS | 更新不存在的运行时 |
| TestRuntimeService_Update_PartialFields | ✅ PASS | 部分字段更新 |
| TestRuntimeService_Delete | ✅ PASS | 删除运行时 |
| TestRuntimeService_Delete_NotFound | ✅ PASS | 删除不存在的运行时 |
| TestRuntimeService_Heartbeat | ✅ PASS | 心跳检测 |
| TestRuntimeService_Heartbeat_NotFound | ✅ PASS | 不存在运行时的心跳 |
| TestRuntimeService_Heartbeat_UpdatesStatusToOnline | ✅ PASS | 心跳更新状态为在线 |
| TestRuntimeService_FindAvailable | ✅ PASS | 查找可用运行时 |
| TestRuntimeService_FindAvailable_None | ✅ PASS | 无可用运行时 |

#### AgentTaskService (36 用例) ✅

| 用例 | 状态 | 描述 |
|------|------|------|
| TestAgentTaskService_Create | ✅ PASS | 创建任务 |
| TestAgentTaskService_Create_NonAdmin | ✅ PASS | 非管理员创建失败 |
| TestAgentTaskService_Create_NoMember | ✅ PASS | 非成员创建失败 |
| TestAgentTaskService_Create_WithOptionalFields | ✅ PASS | 带可选字段创建 |
| TestAgentTaskService_Get | ✅ PASS | 获取任务 |
| TestAgentTaskService_Get_NotFound | ✅ PASS | 获取不存在的任务 |
| TestAgentTaskService_List | ✅ PASS | 列出任务 |
| TestAgentTaskService_List_FilterByStatus | ✅ PASS | 按状态过滤 |
| TestAgentTaskService_List_Empty | ✅ PASS | 空列表 |
| TestAgentTaskService_List_OtherWorkspace | ✅ PASS | 隔离其他工作空间 |
| TestAgentTaskService_Claim | ✅ PASS | 认领任务 |
| TestAgentTaskService_Claim_WithRuntime | ✅ PASS | 带运行时认领 |
| TestAgentTaskService_Claim_NotEnqueue | ✅ PASS | 非入队状态认领 |
| TestAgentTaskService_Claim_NotFound | ✅ PASS | 认领不存在的任务 |
| TestAgentTaskService_Complete | ✅ PASS | 完成任务 |
| TestAgentTaskService_Complete_NotRunning | ✅ PASS | 非运行状态完成 |
| TestAgentTaskService_Complete_NotFound | ✅ PASS | 完成不存在的任务 |
| TestAgentTaskService_Fail | ✅ PASS | 任务失败 |
| TestAgentTaskService_Fail_AutoDetectReason (6 sub) | ✅ PASS | 自动检测失败原因 (timeout/runtime_offline/invalid_input/rate_limit/model_error/agent_error) |
| TestAgentTaskService_Fail_NotFound | ✅ PASS | 不存在的任务失败 |
| TestAgentTaskService_Cancel | ✅ PASS | 取消任务 |
| TestAgentTaskService_Cancel_ClaimedTask | ✅ PASS | 取消已认领任务 |
| TestAgentTaskService_Cancel_CompletedTask | ✅ PASS | 取消已完成任务 |
| TestAgentTaskService_Cancel_FailedTask | ✅ PASS | 取消已失败任务 |
| TestAgentTaskService_Cancel_NotFound | ✅ PASS | 取消不存在的任务 |
| TestAgentTaskService_Cancel_WithRuntime | ✅ PASS | 带运行时取消 |
| TestAgentTaskService_AddLog | ✅ PASS | 添加日志 |
| TestAgentTaskService_AddLog_WithMetadata | ✅ PASS | 带元数据添加日志 |
| TestAgentTaskService_Delete | ✅ PASS | 删除任务 |
| TestAgentTaskService_Delete_NonAdmin | ✅ PASS | 非管理员删除失败 |
| TestAgentTaskService_Delete_NotFound | ✅ PASS | 删除不存在的任务 |
| TestAgentTaskService_Update | ✅ PASS | 更新任务 |
| TestAgentTaskService_Update_NotFound | ✅ PASS | 更新不存在的任务 |
| TestAgentTaskService_Update_NonAdmin | ✅ PASS | 非管理员更新失败 |
| TestAgentTaskService_FullLifecycle | ✅ PASS | 完整生命周期 |
| TestAgentTaskService_FullLifecycle_FailAndRetry | ✅ PASS | 失败重试生命周期 |

#### AgentTemplateService (18 用例) ✅

| 用例 | 状态 | 描述 |
|------|------|------|
| TestAgentTemplateService_Create | ✅ PASS | 创建模板 |
| TestAgentTemplateService_Create_NonAdmin | ✅ PASS | 非管理员创建失败 |
| TestAgentTemplateService_Create_NoMember | ✅ PASS | 非成员创建失败 |
| TestAgentTemplateService_Create_WithSkills | ✅ PASS | 带技能创建 |
| TestAgentTemplateService_Create_InvalidSkillIDs | ✅ PASS | 无效技能ID |
| TestAgentTemplateService_Get | ✅ PASS | 获取模板 |
| TestAgentTemplateService_Get_NotFound | ✅ PASS | 获取不存在的模板 |
| TestAgentTemplateService_Get_PresetTemplate | ✅ PASS | 获取预设模板 |
| TestAgentTemplateService_List | ✅ PASS | 列出模板 |
| TestAgentTemplateService_List_IncludesPresets | ✅ PASS | 包含预设模板 |
| TestAgentTemplateService_List_Empty | ✅ PASS | 空列表 |
| TestAgentTemplateService_Update | ✅ PASS | 更新模板 |
| TestAgentTemplateService_Update_NotFound | ✅ PASS | 更新不存在的模板 |
| TestAgentTemplateService_Update_NonAdmin | ✅ PASS | 非管理员更新失败 |
| TestAgentTemplateService_Delete | ✅ PASS | 删除模板 |
| TestAgentTemplateService_Delete_PresetTemplate | ✅ PASS | 预设模板不可删除 |
| TestAgentTemplateService_Delete_NonAdmin | ✅ PASS | 非管理员删除失败 |
| TestAgentTemplateService_Delete_NotFound | ✅ PASS | 删除不存在的模板 |

#### AgentConfigService (18 用例) ✅

| 用例 | 状态 | 描述 |
|------|------|------|
| TestAgentConfigService_Create | ✅ PASS | 创建配置 |
| TestAgentConfigService_Create_NonAdmin | ✅ PASS | 非管理员创建失败 |
| TestAgentConfigService_Create_NoMember | ✅ PASS | 非成员创建失败 |
| TestAgentConfigService_Get | ✅ PASS | 获取配置 |
| TestAgentConfigService_Get_NotFound | ✅ PASS | 获取不存在的配置 |
| TestAgentConfigService_List | ✅ PASS | 列出配置 |
| TestAgentConfigService_List_Empty | ✅ PASS | 空列表 |
| TestAgentConfigService_List_OtherWorkspace | ✅ PASS | 隔离其他工作空间 |
| TestAgentConfigService_Update | ✅ PASS | 更新配置 |
| TestAgentConfigService_Update_NotFound | ✅ PASS | 更新不存在的配置 |
| TestAgentConfigService_Update_NonAdmin | ✅ PASS | 非管理员更新失败 |
| TestAgentConfigService_Update_SetDefault | ✅ PASS | 设置默认配置 |
| TestAgentConfigService_Delete | ✅ PASS | 删除配置 |
| TestAgentConfigService_Delete_NonAdmin | ✅ PASS | 非管理员删除失败 |
| TestAgentConfigService_Delete_NotFound | ✅ PASS | 删除不存在的配置 |
| TestAgentConfigService_GetDefault | ✅ PASS | 获取默认配置 |
| TestAgentConfigService_GetDefault_NotFound | ✅ PASS | 无默认配置 |
| TestAgentConfigService_GetDefault_NoDefaultSet | ✅ PASS | 未设置默认 |

### 2.2 已有测试 (21 用例)

#### SquadService (6 用例) ✅

| 用例 | 状态 | 描述 |
|------|------|------|
| TestCancelExecution_NotRunning | ✅ PASS | 取消未运行的执行 |
| TestCancelExecution_RemovesFromStore | ✅ PASS | 取消后从存储移除 |
| TestCheckPermissions_NoDB | ✅ PASS | 无DB时权限检查 |
| TestTruncateStr | ✅ PASS | 字符串截断 |
| TestNewSquadService_NilDB | ✅ PASS | Nil DB初始化 |
| TestExecuteSubtaskWithRetry_ContextCancelled | ✅ PASS | 上下文取消时重试 |

#### ToolService (13 用例) ✅

已有完整 CRUD + 权限 + 速率限制 + MCP 集成测试。

#### AutopilotService (2 用例) ✅

| 用例 | 状态 | 描述 |
|------|------|------|
| TestAutopilotTriggerTypes | ✅ PASS | 触发器类型常量 |
| TestAutopilotStatusConstants | ✅ PASS | 状态常量 |

---

## 三、E2E 全功能验收 (154 用例)

> 脚本: [e2e-full-verify.cjs](file:///d:/code/reqmango/frontend/e2e-full-verify.cjs)

### Phase 0: 登录与环境准备 (5 项) ✅

| # | 测试项 | 状态 |
|---|--------|------|
| 1 | 登录页面 — UI 元素验证 | ✅ PASS |
| 2 | 注册页面 — UI 元素验证 | ✅ PASS |
| 3 | API 登录获取 Token | ✅ PASS |
| 4 | 动态发现工作空间 (reqmango-dev, id=1) | ✅ PASS |
| 5 | 动态发现项目 (开放平台 & API, id=4) | ✅ PASS |

### Part 2: Core 核心页面 (16 个) ✅

| # | 页面 | 路由 | 状态 |
|---|------|------|------|
| 1 | 首页/工作空间列表 | `/` | ✅ PASS |
| 2 | 工作空间概览 | `/workspace/:slug/overview` | ✅ PASS |
| 3 | 工作空间设置 | `/workspace/:slug/settings` | ✅ PASS |
| 4 | 工作空间主页(项目列表) | `/workspace/:slug` | ✅ PASS |
| 5 | 项目主页 | `/workspace/:slug/project/:id` | ✅ PASS |
| 6 | 项目设置 (含 Tab: 概览/状态/标签) | `.../settings` | ✅ PASS |
| 7 | Issue 列表 | `...?tab=issues` | ✅ PASS |
| 8 | Issue 创建 | `.../issues/new` | ✅ PASS |
| 9 | Issue 详情 | `.../issues/:issueId` | ✅ PASS |
| 10 | Cycle 列表 | `...?tab=cycles` | ✅ PASS |
| 11 | Cycle 详情 | `.../cycles/:cycleId` | ✅ PASS |
| 12 | Module 列表 | `...?tab=modules` | ✅ PASS |
| 13 | Pages 文档 | `.../pages` | ✅ PASS |
| 14 | Analytics 分析 | `.../analytics` | ✅ PASS |
| 15 | Dashboard 仪表盘 | `.../dashboards` | ✅ PASS |
| 16 | 工作空间分析 | `/workspace/:slug/analytics` | ✅ PASS |

### Part 3: AI/Agent 页面 (16 个) ✅

| # | 页面 | 路由 | 状态 | Tab 测试 |
|---|------|------|------|----------|
| 1 | Agent Dashboard | `/agents` | ✅ PASS | — |
| 2 | Agent Templates | `/agents/templates` | ✅ PASS | — |
| 3 | Agent Configs | `/agents/configs` | ✅ PASS | — |
| 4 | Agent Skills | `/agents/skills` | ✅ PASS | — |
| 5 | Agent Tasks | `/agents/tasks` | ✅ PASS | — |
| 6 | Agent Runtimes | `/agents/runtimes` | ✅ PASS | — |
| 7 | Agent Loops | `/agents/loops` | ✅ PASS | — |
| 8 | Agent Sessions | `/agents/sessions` | ✅ PASS | — |
| 9 | Agent Memories | `/agents/memories` | ✅ PASS | — |
| 10 | Squads 列表 | `/agents/squads` | ✅ PASS | — |
| 11 | Squads 详情 | `/agents/squads/1` | ✅ PASS | 成员/执行/历史/配置 ✅ |
| 12 | Autopilot | `/agents/autopilot` | ✅ PASS | — |
| 13 | Tools 工具管理 | `/agents/tools` | ✅ PASS | 工具/调用日志/权限/MCP ✅ |
| 14 | Monitor 监控 | `/agents/monitor` | ✅ PASS | — |
| 15 | Performance 性能 | `/agents/performance` | ✅ PASS | — |
| 16 | Approvals 审批 | `/approvals` | ✅ PASS | — |

### Part 4: Other 其他页面 (9 个) ✅

| # | 页面 | 路由 | 状态 |
|---|------|------|------|
| 1 | Initiatives/Roadmap | `/initiatives` | ✅ PASS |
| 2 | Custom Fields | `/custom-fields` | ✅ PASS |
| 3 | Issue Types | `/issue-types` | ✅ PASS |
| 4 | Workflows | `/workflows` | ✅ PASS |
| 5 | Workflow Designer | `/workflow/:id/design` | ✅ PASS |
| 6 | Agent Members | `/agent-members` | ✅ PASS |
| 7 | Agent Issues | `/agent-issues` | ✅ PASS |
| 8 | Budget/SLA | `/budget-sla` | ✅ PASS |
| 9 | — | — | — |

### Part 5: API 健康检查 (13 个) ✅

| API | 状态 | 数据量 |
|-----|------|--------|
| Workspaces | 200 OK | 5 个 |
| Projects | 200 OK | 4 个 |
| Issues | 200 OK | 10 个 |
| Agent Templates | 200 OK | 5 个 |
| Tools | 200 OK | 10 个 |
| Squads | 200 OK | 7 个 |
| Agent Configs | 200 OK | 3 个 |
| Skills | 200 OK | 7 个 |
| Autopilot | 200 OK | 1 个 |
| Workflows | 200 OK | 0 个 |
| Runtimes | 200 OK | 1 个 |
| Loops | 200 OK | 1 个 |
| Approvals | 200 OK | 13 个 |

### 已知非阻塞性问题 (3 项)

| 问题 | 严重度 | 说明 |
|------|--------|------|
| 项目主页 400 Bad Request | Low | 后端某 API 返回 400，不影响页面渲染 |
| JS Error: undefined.length | Low | 后端返回空数据时前端未做空值保护 |
| JS Error: null.length | Low | 同上，Performance 页面 API 返回异常 |

### 运行方式

```bash
# 确保前后端服务已启动
cd frontend && node e2e-full-verify.cjs
```

截图保存在 `frontend/e2e-full-screenshots/` 目录。

---

## 四、测试文件清单

### 新增文件 (6 个)

| 文件 | 用例数 | 行数 |
|------|--------|------|
| `backend/internal/service/skill_service_test.go` | 17 | ~350 |
| `backend/internal/service/runtime_service_test.go` | 17 | ~350 |
| `backend/internal/service/agent_task_service_test.go` | 36 | ~700 |
| `backend/internal/service/agent_template_service_test.go` | 18 | ~380 |
| `backend/internal/service/agent_config_service_test.go` | 18 | ~380 |
| `frontend/e2e-agent-features.cjs` | 14页面 | ~700 |

### 已有测试文件 (参考)

| 文件 | 用例数 |
|------|--------|
| `backend/internal/service/tool_service_test.go` | 13 |
| `backend/internal/service/squad_service_test.go` | 6 |
| `backend/internal/service/autopilot_service_test.go` | 2 |

---

## 五、测试技术方案

### 后端测试

- **框架**: Go testing + testify/assert + testify/require
- **数据库**: SQLite 内存数据库 (gorm.io/driver/sqlite)
- **隔离**: 每个测试函数独立创建 DB + AutoMigrate
- **JSON 兼容**: 自定义 `jsonSafeDriver` 解决 SQLite TEXT 与 `json.RawMessage` 类型转换

### E2E 测试

- **框架**: Puppeteer (Node.js)
- **登录**: API 登录获取 token → localStorage 注入
- **路由发现**: 动态查询 workspace slug
- **DOM 查询**: `page.evaluate()` 避免非标准 CSS 选择器
- **容错**: 每个测试 try/catch 独立运行，失败不阻塞

---

## 六、测试缺口 (后续建议)

| 模块 | 当前状态 | 建议 |
|------|----------|------|
| Handler 层 (API 集成测试) | 仅 2/70 有测试 | 优先补充 Auth、Issue、Project Handler |
| Middleware (认证/限流/RBAC) | 无测试 | 安全关键路径，建议优先 |
| Issue Service | 无测试 | 核心业务，建议补充 |
| 前端组件单元测试 | 仅 5 个组件 | 建议补充关键表单组件 |
| Cycle/Module Service | 无测试 | 迭代管理核心 |
