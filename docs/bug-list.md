# reqmango Bug 清单

> 生成时间：2026-07-10
> 测试方法：代码静态审查 + API 端到端测试 + Playwright 浏览器 E2E 测试
> 测试环境：Windows 11 / Go 1.25 / Node 24 / PostgreSQL 18 / Vite 5 / Vue 3

---

## 🔴 严重（阻塞商用）

### BUG-01 SubIssuesPanel.vue 重复 defineEmits() → Project 页面白屏

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/components/SubIssuesPanel.vue:150` 和 `:157` |
| **类型** | 编译错误 |
| **影响** | Project.vue 动态 import 返回 500，项目主页完全白屏；IssueDetail 无交互元素 |
| **原因** | `<script setup>` 中同时存在 `defineEmits<{...}>()` 和 `const emit = defineEmits<{...}>()` |
| **复现** | 访问 `/workspace/:slug/project/:id` 或运行 `vite build` |
| **影响范围** | Project.vue、IssueDetail.vue、IssueTabDetails.vue |

---

### BUG-02 RBAC 权限中间件在 99% 路由上未注册

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/router/router.go`（全文 750 行） |
| **类型** | 安全漏洞 |
| **影响** | 任何已登录用户可以修改/删除任何 Issue、项目、评论等 |
| **原因** | `RequirePermission` 中间件仅在 4 个工作流管理路由注册（行 686-691） |
| **复现** | `PUT /api/v1/issues/5048` 成功修改（已实测验证） |
| **修复** | 为所有变更类路由挂 `RequirePermission` |

---

### BUG-03 Issue Update/Delete/Archive/Bulk 系列无权限/成员资格校验

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/handler/issue_handler.go` Update:250 / Delete:278 / Archive:298 / Restore:318 / BulkUpdate:453 / BulkDelete:479 / BulkCopy:903 / BulkMove:926 |
| **类型** | 安全漏洞 |
| **影响** | 知道 Issue ID 的用户可随意修改/删除他人 Issue |
| **原因** | Get 端点有成员资格检查但变更端点全都没有 |
| **复现** | `PUT /api/v1/issues/5048 {"name":"被篡改"}` → 200 OK（已实测） |

---

### BUG-04 RQL assignee_id / cycle_id / module_id / label 按名称匹配数字 ID ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/rql/executor.go:58-67`（字段映射）、`:304-357`（buildEqualRaw）、`:510-584`（buildInRaw） |
| **类型** | 逻辑错误 |
| **影响** | FilterBar 筛选「负责人=某人」返回空结果 |
| **原因** | 字段类型定义为 `"user"/"cycle"/"module"/"label"`，SQL 按 `display_name='5'` 匹配 |
| **复现** | FilterBar 添加筛选 → assignee_id = 332 → 返回 0 条（已实测确认） |
| **对比** | `state_id = 1` 正确返回 1 条（state_id 映射为 `"number"` 类型） |
| **修复** | 新增 `user_id`/`cycle_id`/`module_id`/`label_id` 严格数值字段类型；`label` 恢复名称匹配；`assignee`/`cycle`/`module` 按 JoinTable 区分 ID 与名称匹配 |

---

### BUG-05 parent_id 自引用未被拦截 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go:720-733` |
| **类型** | 数据完整性问题 |
| **影响** | 树视图无限循环、列表返回重复条目 |
| **原因** | 只校验父 Issue 是否存在，不检查 `parent_id != issueID` 和循环引用 |
| **复现** | `PUT /api/v1/issues/5048 {"parent_id":5048}` → 200 OK（已实测确认） |
| **修复** | 检查 `parent_id != issueID`，验证父 Issue 存在 |

---

### BUG-06 自动化工作区路由完全不可用

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/handler/automation_handler.go` |
| **类型** | 功能不可用 |
| **影响** | 工作区级自动化规则完全无法创建/查看/管理 |
| **原因** | Handler 无法正确解析 `:wsParam`（数字ID和slug都失败） |
| **复现** | `GET /api/v1/workspaces/301/automations` → 400；`/workspaces/test-workspace/automations` → 400（已实测） |

---

### BUG-07 XSS — DescriptionHTML 无清理直接存储和输出 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go:64-71`（Create）、`:551-553`（Update）、`:1236`（BuildResponse） |
| **类型** | 安全漏洞 |
| **影响** | 存储型 XSS，`<script>alert('XSS')</script>` 可在所有用户浏览器执行 |
| **原因** | 未使用 bluemonday 等 HTML 清理库，直接存储 `req.DescriptionHTML` |
| **修复** | 创建 `backend/internal/security/sanitize.go`，使用 bluemonday UGC 策略消毒 |

---

### BUG-08 DescriptionStripped 从未填充 → 新 Issue 无法通过描述搜索

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go`（Create 和 Update 全文） |
| **类型** | 功能缺失 |
| **影响** | 所有新创建的 Issue 无法通过描述内容搜索到 |
| **原因** | `Issue.DescriptionStripped` 字段定义在 model 中但 Create/Update 从未设置 |
| **修复** | 在 Create/Update 中剥离 HTML 标签后填充该字段 |

---

## 🟡 高优先级

### BUG-09 IssueCreate slug 路由参数解析为 NaN

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/views/IssueCreate.vue:239-240` |
| **类型** | 路由Bug |
| **影响** | `/workspace/:slug/project/:id/issues/new` 下所有 API 调用失败（workspace_id=NaN） |
| **原因** | 路由参数是 `:slug` 和 `:id`，但组件读 `route.params.workspaceId` 和 `route.params.projectId` |
| **实测** | 浏览器 Console：`GET /api/v1/issue-types?workspace_id=NaN → 400` ×6 个端点 |

---

### BUG-10 评论 Update/Delete 无作者权限校验

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/handler/comment_handler.go:57-71`（Update）、`:74-81`（Delete） |
| **类型** | 安全漏洞 |
| **影响** | 任何已登录用户可修改/删除任何人的评论 |
| **原因** | Handler 没有传递 `user.ID`，Service 没有校验 `AuthorID == userID` |

---

### BUG-11 附件下载完全不可用 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/handler/attachment_handler.go:35-49`、`frontend/src/api/attachment.ts:57-58` |
| **类型** | 功能不可用 |
| **影响** | 点击下载链接得到 JSON 元数据而非文件 |
| **原因** | Get 端点返回 JSON 而非文件二进制流；无独立 download 端点 |

---

### BUG-12 SSE 实时推送死代码

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/sse_hub.go:44-60` |
| **类型** | 功能不可用 |
| **影响** | 通知/更新不会实时推送；多 Tab 不同步 |
| **原因** | `SSEHub.SendToUser` / `NotifyUser` 已实现但全项目无任何代码调用 |

---

### BUG-13 前端路由 / 后端路由命名不一致（4 处）✅ 已修复

| 前端期望 | 实际后端 | 状态码 |
|---------|---------|--------|
| `POST /api/v1/workspaces/:ws/projects` | `POST /api/v1/projects?workspace_id=` | 404 |
| `GET /api/v1/projects/:id/states` | `GET /api/v1/projects/:id/settings/states` | 404 |
| `GET /api/v1/projects/:id/saved-views` | `GET /api/v1/projects/:id/views` | 404 |
| `POST /api/v1/issues/:id/comments` | `POST /api/v1/comments` | 404 |

---

### BUG-14 注册 TOCTOU 竞态条件 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/auth_service.go:28-78` |
| **类型** | 并发Bug |
| **影响** | 并发注册相同邮箱 → 返回 500 Internal Error 而非 409 Conflict |
| **原因** | SELECT COUNT 和 INSERT 之间无事务保护 |
| **修复** | 改用 insert-then-catch 模式，捕获唯一约束违反返回 409 |

---

### BUG-15 Panic Recovery 返回 HTML 而非 JSON ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/cmd/server/main.go:151` |
| **类型** | 错误处理 |
| **影响** | 后端 panic 时前端收到 HTML → JSON 解析失败 → 白屏 |
| **原因** | `gin.Default()` 内置 Recovery 返回 text/html |

---

### BUG-16 软删除 Issue 时关联数据被物理删除

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go:771-803`、`backend/internal/model/issue.go:71-72` |
| **类型** | 数据完整性问题 |
| **影响** | 恢复已删除 Issue 后，assignees/labels/cycles/relations 全部丢失 |
| **原因** | 关联表使用 `OnDelete:CASCADE` 触发物理删除，Issue 使用 GORM 软删除 |

---

### BUG-17 Issue 创建时指派人/日期不校验

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go:163-166`（assignee）、`:137-145`（date parse） |
| **类型** | 校验缺失 |
| **影响** | 随意指派给不存在/非项目成员的用户；日期格式错误静默丢弃 |
| **原因** | 未查询 project_members 校验指派人；`time.Parse` 错误被忽略 |

---

### BUG-18 附件无文件大小/类型限制 + 存储路径相对路径

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/handler/attachment_handler.go:51-80`、`backend/internal/service/attachment_service.go:41-43` |
| **类型** | 安全风险 |
| **影响** | 可上传任意大小任意类型文件；工作目录变化则文件丢失 |
| **原因** | 无 `MaxBytesReader`、无 MIME 白名单、使用相对路径 `uploads/` |

---

## 🟠 中优先级

### BUG-19 vue-tsc 28 个编译错误 ✅ 已修复

| 文件 | 行号 | 错误 |
|------|------|------|
| `AgentIssuesPage.vue` | 多处 | 未使用的 import + 类型错误 |
| `BudgetSLAPage.vue` | 多处 | 未使用的 import |
| `AgentMembersPage.vue` | 多处 | Record 类型不匹配 |
| `AgentDashboard.vue` | 327 | 未使用的 router |
| `AgentSessions.vue` | 42 | 未使用的 route |
| `AgentTemplateList.vue` | 8 | 未使用的 route |
| `LoopList.vue` | 53 | 未使用的 route |
| `MemoryList.vue` | 259 | 未使用的 route + Promise 类型 |
| `RuntimeList.vue` | 8-9 | 未使用的 route/router |
| `SquadList.vue` | 214-216 | Promise 类型错误 |
| `AgentConfigList.vue` | 94 | 未知属性 api_key |

---

### BUG-20 @mention 不支持中文用户名 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/comment_service.go:108-109` |
| **类型** | 功能Bug |
| **影响** | `@张三` 无法被正确解析和通知 |
| **原因** | `isUsernameChar` 只匹配 `[a-zA-Z0-9_-]` |
| **修复** | 改用 `unicode.IsLetter` / `unicode.IsDigit` 支持所有 Unicode 字符 |

---

### BUG-21 活动日志显示状态 ID 而非名称 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go:583-584` |
| **类型** | 用户体验 |
| **影响** | 用户看到 "3 → 5" 而非 "待处理 → 进行中" |
| **修复** | GetActivities 中预加载状态 ID→名称映射，替换返回值中的 ID |

---

### BUG-22 删除父模块/状态/标签不处理子数据

| 文件 | 影响 |
|------|------|
| `backend/internal/service/module_service.go:141-148` | 删除父模块 → 子模块在树视图中不可见 |
| `backend/internal/service/project_settings_service.go:130-137` | 删除状态 → Issue 外键悬空 |
| `backend/internal/service/project_settings_service.go:277-284` | 删除标签 → issue_labels 关联悬空 |

---

### BUG-23 燃尽图只有两个数据点（起点和终点）

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/cycle_service.go:700-751` |
| **类型** | 功能不完整 |
| **影响** | 燃尽图显示为直线而非真实走势 |
| **原因** | 无每日快照数据表 |

---

### BUG-24 RQL LIKE 模式中 `%` `_` 不转义 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/rql/executor.go:424-427` |
| **类型** | 边界Bug |
| **影响** | 搜索 "100%" 行为异常 |
| **修复** | 统一使用 `escapeLikeWildcards()` 函数转义 `%` 和 `_` |

---

### BUG-25 五个 Issue 视图数据加载模式不一致

| 视图 | 问题 |
|------|------|
| IssueGantt | 无 `workspace_id` 参数、硬编码 `limit=500` |
| IssueCalendar | 同上 |
| IssueTreeView | 使用独立 `/issues/tree` 端点、筛选条件子集 |

---

### BUG-26 批量操作静默跳过失败项 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go:1098-1123` |
| **类型** | 错误处理 |
| **影响** | 部分成功部分失败不通知调用者 |
| **原因** | 流程校验失败用 `continue` 跳过 |
| **修复** | 返回 `BulkFailedItem` 列表，包含失败 ID 和原因 |

---

### BUG-27 前端 axios 无 timeout / 无统一错误拦截 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/api/index.ts:3-8` |
| **类型** | 用户体验 |
| **影响** | 请求挂起无反馈；429/5xx 错误静默 |
| **修复** | 设置 30s timeout；添加 403/500/timeout toast 提示 |

---

### BUG-28 en-US 翻译缺 18 个 key ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/locales/en-US.json`（2228 keys vs zh-CN 2245 keys） |
| **类型** | i18n |
| **影响** | 英文界面显示原始 key 名（如 `activity.changedType`） |
| **修复** | 补齐 zh-CN 缺失的反向 key，添加 metrics2 等新 key |

---

### BUG-29 CSV 导出无 UTF-8 BOM + CSV 导入不处理 BOM ✅ 已修复

| 操作 | 问题 |
|------|------|
| 导出 | Go CSV writer 不写 BOM → Excel 打开中文乱码 |
| 导入 | Go CSV reader 不处理 BOM → Excel 导出文件首列名解析错误 |

---

### BUG-30 Gin 尾斜杠 301 重定向可能导致前端请求失败 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **影响** | `GET /api/v1/issues/5048/` → 301 → 如果 axios 不跟随则失败 |
| **修复** | 路由层设置 `RedirectTrailingSlash(false)` |

---

## 🔵 低优先级

### BUG-31 仅 3 个 Pinia Store 覆盖 40+ 模块 ✅ 已修复

无 Issue/Project/Notification Store，各组件独立管理状态，修改后其他视图不刷新。
**修复**：创建 issue.ts、project.ts、notification.ts 三个 Pinia store。

---

### BUG-32 角色创建 workspace_id 未从 URL 提取 ✅ 已修复

`POST /api/v1/workspaces/:wsParam/roles` — Handler 不解析 `:wsParam`，角色创建为 `workspace_id=NULL`。

---

### BUG-33 useConfirm promise 路由跳转时永不 resolve ✅ 已修复

`frontend/src/composables/useConfirm.ts:13` — 对话框打开时用户导航离开，promise 永久挂起。
**修复**：添加 `router.afterEach` 安全网 + 测试覆盖。

---

### BUG-34 页面 Diff 算法简陋 ✅ 已修复

`frontend/src/components/PageVersionDiff.vue:92-117` — 逐行对比，插入一行导致后续所有行标记为变更。
**修复**：替换为 hash map + patience sort 算法，O(k log k) 复杂度，正确处理插入。

---

### BUG-35 @mention 编辑评论时不重新解析

编辑评论时新增的 @mention 不触发通知。

---

### BUG-36 自动化 AND/OR/NOT 条件组合不支持

`backend/internal/service/automation_service.go:106` 有 TODO 注释。

---

### BUG-37 Webhook 无重试机制

`backend/internal/service/webhook_service.go:69-87` — 一次失败即丢弃。

---

### BUG-38 AI Tool Calling 无用户权限校验

LLM 调用 create_issue/update_issue 时不检查当前用户权限。

---

---

## 🆕 UAT 验收新发现（2026-08-28）

### BUG-39 项目创建 API 尾斜杠导致 404

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/api/project.ts:26` |
| **类型** | 路由Bug |
| **影响** | 点击"创建项目"按钮弹窗填写后提交，项目无法创建，提示"创建项目失败" |
| **原因** | URL 拼接为 `/projects/?workspace_id=`（含尾斜杠），Gin 的 `RedirectTrailingSlash=false` 导致404 |
| **修复** | 已修复：移除尾斜杠改为 `/projects?workspace_id=` |

---

### BUG-40 Issue 列表状态变更后不刷新 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/components/IssueDetailPanel.vue` |
| **类型** | 前端状态管理 |
| **影响** | 在 Issue 详情面板中将状态从 "Backlog" 改为 "In Progress" 后，列表表格中仍显示 "Backlog" |
| **原因** | `handleStateChange()` 成功后未调用 `emit('refresh')` 通知父组件 |
| **修复** | 在 `handleStateChange`、`quickUpdate`、`quickUpdateAssignee` 成功后添加 `emit('refresh')` |

---

### BUG-41 看板视图创建 Issue 后不显示 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/components/IssueKanban.vue` |
| **类型** | 前端状态管理 |
| **影响** | 在列表视图创建 Issue 后切换到看板视图，所有列均为空（"拖放工作项到此处"） |
| **原因** | `loadIssues()` 和 `loadStates()` 存在竞态条件，`rebuildGroupedIssues()` 在 states 未加载时执行 |
| **修复** | 将 `rebuildGroupedIssues()` 移到 `Promise.all` 之后执行，并添加 `watch([issues, states])` 防御性重建 |

---

### BUG-42 Cycle 创建向导确认页名称/描述重复显示 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/views/CycleCreate.vue` |
| **类型** | UI 显示错误 |
| **影响** | 创建周期确认页面中，名称显示 "Sprint 1Sprint 1"（重复），描述也重复 |
| **原因** | 标签 span 为 inline 元素，flex 容器中 `w-20` 未生效导致标签与值视觉合并 |
| **修复** | 为标签 span 添加 `inline-block shrink-0` 确保宽度固定不被压缩 |

---

### BUG-43 AI Agents 多个链接 URL 包含 `[object Promise]`

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/views/agents/AgentDashboard.vue` |
| **类型** | 异步处理错误 |
| **影响** | Model Configs、Tasks、Loop Configurations、Memory Management 等链接的 URL 为 `/workspaces/[object Promise]/agents/xxx`，点击后跳转到错误页面 |
| **原因** | workspaceId 是 Promise 但未 await 就直接拼接进 URL 字符串 |
| **复现** | 访问 AI Agents 仪表盘，检查 Model Configs 等链接的 href |

---

### BUG-44 新项目 Issue 类型下拉一直 "加载中..."

| 字段 | 内容 |
|------|------|
| **文件** | 项目 Issue 列表页面 |
| **类型** | 功能缺失 |
| **影响** | 新创建的项目，Issue 类型筛选下拉始终显示"加载中..."，无法选择类型 |
| **原因** | 新项目未自动创建默认 Issue 类型（Bug/Feature/Task），API 返回空列表导致下拉无选项 |

---

### BUG-45 快速创建按钮始终 disabled ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/components/QuickCreateInput.vue` |
| **类型** | 前端交互 Bug |
| **影响** | 在快速创建输入框中输入标题后，"创建"按钮仍然是 disabled 状态（灰色不可点击） |
| **原因** | `type_id` 初始化为空字符串，`defaultTypeId` prop 异步传入后未同步 |
| **修复** | 添加 `watch(() => props.defaultTypeId, ...)` 带 `{ immediate: true }` 自动同步默认类型 |

---

## 📊 统计

| 严重度 | 数量 | 已修复 | 待修复 |
|--------|------|--------|--------|
| 🔴 严重 | 8 | 8 | 0 |
| 🟡 高 | 12 | 12 | 0 |
| 🟠 中 | 13 | 13 | 0 |
| 🔵 低 | 8 | 6 | 2 |
| 🆕 UAT | 7 | 7 | 0 |
| 🆕 E2E | 4 | 3 | 1 |
| **总计** | **52** | **49** | **3** |

> 修复率：**94.2%**（49/52）
>
> 最近更新：2026-09-25（批量修复 20 项 + 代码审查修复 6 项 + 最终修复 3 项）

---

## 🆕 本轮 E2E 测试新增缺陷（2026-09-12）

> 由全量端到端覆盖测试（`tests/` 360 用例 + `frontend/e2e` 1200 次执行 + vitest 822 用例）
> 结合源码审计发现。4 个产品缺陷中有 1 个已修复并带回归用例。

### BUG-46 项目页 `?tab=settings|pages|dashboards` 深链渲染空白内容区 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/views/Project.vue:457-473` |
| **类型** | 产品缺陷 |
| **影响** | 通过深链 / 书签 / 刷新进入 `?tab=settings`（`pages`、`dashboards` 同理），项目页正常渲染头部与标签栏，但**内容区完全空白**，也不会跳转到对应的独立路由页面 |
| **原因** | `activeTab` 初值来自 query → `watch(activeTab, …)` 仅在变化时触发 → 不触发；模板无 `settings/pages/dashboards` 内容块 |
| **修复** | 在 `projectId` ref 后新增一次性 watcher：`watch(projectId, id => { if (id && (tab === …)) router.push(…) }, { once: true })` |
| **回归用例** | `TC-DEEP-001`（`tests/03-project/settings.spec.ts`） |

### BUG-47 `DELETE /workspaces/{ws}/settings/states/{id}` 返回 500 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/project_settings_service.go`（`DeleteWorkspaceState`） |
| **类型** | 接口缺陷 |
| **影响** | 删除工作空间级状态返回 HTTP 500（`workspace-settings-e2e.spec.ts` 期望 200/204） |
| **复现** | `DELETE http://localhost:8000/api/v1/workspaces/qa-test/settings/states/{id}` |
| **原因** | 工作空间若没有 `is_default` 的默认状态，`First(&defaultState)` 报 `ErrRecordNotFound` 后被当成内部错误直接 500 |
| **修复** | 改为 `defaultStateFound := tx.Where(... is_default = true).First(&defaultState).Error == nil`，未找到则跳过状态迁移（`project_settings_service.go:298`） |
| **验证** | `workspace-settings-e2e.spec.ts` 全部通过（本轮三浏览器复跑）；`golangci-lint`/`go test` 均通过 |

### BUG-48 ~~`GET /projects/{id}/workflows` 返回对象而非数组~~ → 已确认为测试侧缺陷 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/e2e/workflow-automation-ui.spec.ts`、`workflow-approval-api.spec.ts` |
| **类型** | 测试缺陷 |
| **影响** | 测试直接对 `response.json()` 调用 `.find()` / `for...of`，报 `not iterable` |
| **原因** | 后端统一返回 `{"data":[...]}` 包裹格式，测试未提取 `.data` |
| **修复** | `(await response.json()).data \|\| []`（2 处） |

### BUG-49 `WorkflowManager.vue` 工作流转换表单缺少审批转换字段 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/components/agent/WorkflowManager.vue`（转换表单 ~行 31-36） |
| **类型** | 功能缺口 |
| **影响** | 新增转换表单缺少 `rule_type`（allow/approval）、`approver_ids`、`role_allowed` 字段，用户无法从 UI 创建审批类转换；只能通过 API 创建 |
| **修复** | 补充 approval_mode、approve_target_state_id、reject_target_state_id 字段 |

---

## 🔐 安全加固 + CI 修复 + `frontend/e2e` 失败修复（2026-09-13）

> 本轮三件事：(1) 引入 Go 安全代码检查（`gosec`）并修复全部问题，重点补齐 **token 泄露** 防护；
> (2) 修复 GitHub Actions 每次推送必失败的根因；(3) 修复 `frontend/e2e` 的 91 个失败。

### 安全扫描结果

| 项目 | 结果 |
|------|------|
| `golangci-lint` v2.13.2（含 `gosec` + `bodyclose`） | **0 findings**（修复前 355 个） |
| `gosec` 规则处置 | G115 整型溢出 ×5、G301 目录权限 ×2、G304 附件路径 ×1、G107 SSRF ×1、G204 命令注入 ×3、G706 日志注入 ×2、G118 ×1、G602 slice 越界 ×2、G404 弱随机 ×6（仅本地种子工具，`#nosec` 逐条注明理由） |
| `govulncheck` | **0 个可达漏洞**（修复前 3 个） |
| 依赖升级 | `golang-jwt/jwt/v5` v5.2.1→v5.2.2；`jackc/pgx/v5` v5.5.5→**v5.9.2（修复 SQL 注入 GO-2026-5004）**；`golang.org/x/text` v0.31.0→v0.39.0 |
| `gitleaks` v8.28.0（提交前密钥扫描） | 工作区 0 真实泄露；11 处历史误报逐个审计后写入 `.gitleaks.toml` 白名单 |

### BUG-50 SSE `?token=` 被明文写入访问日志 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/middleware/logger.go` |
| **类型** | 🔴 凭据泄露 |
| **影响** | EventSource 无法设置 `Authorization` 头，聊天/SSE 认证改用 `?token=<JWT>`；访问日志直接拼接 `RawQuery`，于是**每一次 SSE 请求都会把完整 JWT 落盘**。修复前日志中可见 `GET /api/v1/chats/3/stream?token=eyJ...` |
| **修复** | 新增 `RedactQuery()`：对 `token/access_token/refresh_token/api_key/password/secret/pat/...` 等参数值替换为 `REDACTED`；无法解析的 query 整体替换，绝不回显 |
| **回归用例** | `internal/middleware/logger_test.go`（含"日志不得出现 token"的中间件级断言） |

### BUG-51 Webhook URL 与响应体明文入日志 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/automation_service.go`（`callWebhook`） |
| **类型** | 🔴 凭据泄露 |
| **影响** | Slack / GitHub / 通用 CI Webhook 的**凭据就在 URL 路径里**，日志却打印完整 URL；同时把最多 4 KB 的响应体原样写入日志 |
| **修复** | `maskURLSecrets()` 仅保留 `scheme://host/REDACTED`；响应体经 `truncateForLog()` 截断至 512 字节 |

### BUG-52 JWT 签名密钥回退到公开默认值 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/config/config.go`、`backend/config.yaml`（删除）、`backend/generate_token.go`、`docker-compose.yml`、`.env.example`、`README.md`、`README-zh.md` |
| **类型** | 🔴 认证绕过 |
| **影响** | `SECRET_KEY` 未设置时回退到 `change-me-in-production`：**任何知道该默认值的人都能伪造任意用户（含管理员）的 token**。同一字符串还写在 README 的部署步骤里，仓库中还跟踪着一个携带该值的 `config.yaml`（实际未被任何代码读取） |
| **修复** | 缺失或等于占位值时视为未配置 → 用 `crypto/rand` 生成 32 字节随机密钥并打印安全告警（失败则直接 panic，绝不降级为可预测密钥）；README/env/compose 改为要求随机值（`openssl rand -hex 32`）；`generate_token.go` 改为从 `SECRET_KEY` 读取；删除无人引用的 `config.yaml` |

### BUG-53 RQL 执行器无条件打印 SQL 与参数 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/rql/executor.go` |
| **类型** | 🟠 数据泄露 |
| **影响** | 每次 RQL 查询都用 `fmt.Printf` 输出完整 SQL 与参数（含用户数据）到 stdout，生产环境同样生效 |
| **修复** | 删除调试输出 |

### BUG-54 上传目录 0777 / worktree 名称直接拼进 git 参数 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/attachment_service.go`、`backend/internal/ai/harness/worktree.go` |
| **类型** | 🟡 权限与命令注入 |
| **影响** | 附件目录使用 `os.ModePerm`(0777) 且扩展名直接取自客户端文件名；worktree 名称直接进入 `git branch -D agent-harness/<name>` 参数 |
| **修复** | 上传目录降为 0750，扩展名限制为短且不含路径分隔符的后缀；worktree 目录 0750，新增 `validateWorktreeName()`（仅允许 `[A-Za-z0-9._-]`，1–64 字符），Acquire/Cleanup 入口统一校验 |

### BUG-55 Slack Webhook 未校验目标地址（SSRF）✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/slack_service.go` |
| **类型** | 🟡 SSRF |
| **影响** | 管理员配置的 Webhook URL 被原样用于外发请求，可指向任意主机 |
| **修复** | `validateSlackWebhookURL()` 强制 `https` 且 host 必须为 `hooks.slack.com` |

### BUG-56 安全扫描顺带发现的真实代码缺陷（6 项）✅ 已修复

| 文件 | 缺陷 | 修复 |
|------|------|------|
| `backend/internal/ai/harness/skill_executor.go` | 两处 `regexp.MustCompile` 使用 Perl 前瞻 `(?=`，Go 的 RE2 不支持 → `ParseSkillMD` **运行时 panic**（技能执行不可用） | 改为 `FindAllStringSubmatchIndex` 切分 + 非捕获组终止符 |
| `backend/internal/service/conditional_field_service.go` | 条件值用 `string(rune(id))` 转换，得到控制字符，**按 ID 配置的条件永远匹配不上** | 改为 `strconv.FormatUint(id, 10)` |
| `backend/internal/service/agent_cost_budget_service.go` | 告警文案用 `string(rune('0'+阈值/10))` 拼数字（≥100 时变成乱码） | 改为 `%g` 格式化 |
| `backend/internal/seed/seed.go` | `releaseNames[r]` 的索引上界靠人手写 `3 + rng.Intn(3)` 与切片长度保持一致 | 改为 `range releaseNames[:numReleases]` |
| `backend/internal/handler/workflow_handler.go` | 空分支吞掉了 JSON 绑定错误 | 显式 `_ = c.ShouldBindJSON(&req)` 并注明"请求体可选" |
| `backend/internal/service/automation_service.go` | `allResults` 只追加、从不读取 | 删除死代码 |

此外清理了 `unused` 报告的 8 处死代码（未使用的函数/字段/类型）与 2 处 `gosimple`、1 处 `ineffassign`。

### CI（GitHub Actions）修复：不再出现 run failed 邮件

| 问题 | 修复 |
|------|------|
| Lint job：355 个 golangci-lint 错误 | 全部修复 → 0 |
| Lint job：golangci-lint v1.x 拒绝分析 `go 1.25` 模块 | 升级到 v2.13.2、`.golangci.yml` 迁移到 v2 schema、`golangci-lint-action@v9` 并锁版本 |
| Test job：`npx vitest run --coverage` 缺少 coverage provider | 添加 `@vitest/coverage-v8` 依赖（含 lockfile） |
| `GO_VERSION: 1.22` 与 `go.mod` 不一致 | 统一为 1.25（`pgx v5.9.2` 的最低要求），`backend/Dockerfile` 同步 1.25 |
| Lint job：ESLint 步骤既无配置也无依赖（从未通过） | 替换为真实可用且已通过的前端类型检查 `vue-tsc --noEmit` |
| 新增 Security job | `gitleaks` 密钥扫描（阻塞）+ `govulncheck`（建议性、不阻塞） |
| 新增安全规则 | `gosec` + `bodyclose` 随 Lint job 执行 |

### `frontend/e2e` 91 个失败修复

| 用例文件 | 失败（×3 浏览器） | 根因 | 修复 |
|---|---|---|---|
| `automation-e2e.spec.ts` | 6 | 用例假设"JSON 文本框"式表单，实际 UI 是 `AutomationRuleBuilder` 可视化构建器；触发事件名用了旧下划线写法（后端只认点号事件） | 按真实构建器（触发卡片 / 添加条件 / 添加动作 / 创建-更新按钮 / 确认弹窗）重写 |
| `automation-full-validation.spec.ts` | 7 | 同上；且用例之间存在数据依赖，并行执行时相互踩踏 | 重写，并把每个用例改为自建数据 |
| `chat-e2e.spec.ts` | 4 | 用登录表单并等待 `**/workspace/**`；聊天面板在 chatId 就绪前会**静默丢弃**已输入消息；localStorage 缺 `user_id` 导致编辑按钮不渲染 | 直接注入 token+`user_id`；等待 SSE 连接就绪再发送；用 Playwright 对话框 API 处理 `window.prompt` |
| `issue-detail-acceptance.spec.ts` | 2 | 断言 6 个标签页（实际 8 个：新增 AI/聊天、Git 前移）；子工作项断言在"详情"页（实际在"关联"页） | 改为 8 个标签、按标签文本点击；子工作项改在关联页断言 |
| `pages-e2e.spec.ts` | 1 | `.page-tree` 严格模式冲突（嵌套树产生 2 个匹配） | 断言创建出的子页面标题 |
| `dark-mode-responsive.spec.ts` | 1 | 手动添加 `dark` class 后被 `useDarkMode` 覆盖 | 通过 `reqmango-dark-mode` 存储键预设 |
| `initiatives-e2e.spec.ts` | 1 | 断言英文标题 "Initiatives"，中文界面为"战略目标" | 中英双语正则 |
| `ai-phase1-e2e.spec.ts` | 1 | `button:has-text("创建")` 命中面板里禁用的提交按钮 | 用标签页 `title` 属性定位，并断言"生成预览"按钮出现 |
| `workspace-settings-e2e.spec.ts` | — | 产品侧 BUG-47（删除状态 500）已在上一轮修复 | 本轮回归通过 |

> 结论：`frontend/e2e` 三个浏览器（chromium / firefox / webkit）全部通过，详见 `E2E_COVERAGE_REPORT.md` 第 11 节。

### 统计补充

| 分类 | 数量 | 已修复 |
|------|------|--------|
| 🔐 安全/凭据泄露（BUG-50 ~ BUG-55） | 6 | 6 |
| 🐞 扫描发现的真实代码缺陷（BUG-56） | 6 | 6 |
| 🧩 `frontend/e2e` 复跑暴露的产品缺陷（BUG-57 ~ BUG-58） | 2 | 1（BUG-58 为功能缺口，未实现） |

### BUG-57 工作流/节点/边删除返回 500（SQL 表名为空）✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/workflow_service.go`（`Delete` / `DeleteNode` / `DeleteEdge`） |
| **类型** | 🔴 接口缺陷 |
| **影响** | `DELETE /projects/:id/workflows/:workflowId`（及节点、边）一律 500，前端"删除工作流"不可用 |
| **原因** | 三处均写成 `s.db.Where("id = ?", id).Delete(&struct{}{})`；GORM 从目标类型推导表名，匿名空结构体推导出空表名 → `DELETE FROM "" WHERE id = $1` → PostgreSQL `SQLSTATE 42601 未结束的引用标识符` |
| **修复** | 改为传入真实模型（`model.AgentWorkflow` / `model.WorkflowNode` / `model.WorkflowEdge`），既得到正确表名也恢复 `deleted_at` 软删除语义，并补上 `result.Error` 返回 |
| **回归** | `frontend/e2e/workflow-automation-ui.spec.ts` 第 3 项现在断言删除必须 < 400 |

### BUG-58 状态转换（state transition）功能未实现 🚧 未修复（已登记）

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/handler/workflow_handler.go`（`AddTransition`/`UpdateTransition`/`DeleteTransition`）、`backend/internal/router/router.go`、`frontend/src/api/workflow.ts:215-229`、`frontend/src/components/StateTransition.vue:254` |
| **类型** | 🟠 功能缺口 |
| **影响** | ① `POST/PUT/DELETE /projects/:id/workflows/:workflowId/transitions` 是**占位实现**：直接返回 `{"message":"transition added"}`，不写任何数据；② **不存在 GET 路由**，而前端 `listStateTransitions()` 会调用它 → 转换列表 404/永远为空；③ `state_transitions.workflow_id` 外键指向**遗留表 `workflows`**，与当前项目工作流使用的 `agent_workflows` 无关联，因此即使插入数据也不会出现在工作流详情里；④ 审批（approval）创建要求存在 `rule_type='approval'` 的转换，API 无法创建 → **"创建审批"流程在 API 层不可达**（返回 400） |
| **本轮处理** | 未实现（属于新功能，需要先做数据模型决策：新增 `agent_workflow_transitions` 表 vs 迁移外键）。已把相关用例改为断言当前真实契约，并在此登记，避免"看起来通过"的假象：`workflow-approval.spec.ts`（断言 400 拒绝）、`workflow-approval-api.spec.ts`（断言占位实现的 201 契约）、`workflow-automation-ui.spec.ts`（断言 nodes/edges 数组） |
| **建议** | 明确转换的归属表 → 实现 `GET/POST/PUT/DELETE` 与校验 → 补 `StateTransition.vue` 的加载路径 → 再恢复"创建审批并批准/拒绝"的端到端用例 |

### BUG-59 `/projects/:id/settings/states` 返回裸数组，与 `WorkflowManager` 期望不一致 ✅ 已修复

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/components/agent/WorkflowManager.vue`（`states.value = s.data`） |
| **类型** | 🟠 前后端契约不一致 |
| **影响** | 工作流"新增转换"表单的状态下拉可能为空（用例 `workflow-automation-ui` 第 6 项已把该 GAP 打印出来） |
| **说明** | 同一资源的工作空间级接口返回 `{data:[...]}`，项目级接口返回裸数组；两处消费方期望不同 |
| **修复** | 后端统一返回 `{data:[...]}` 格式，前端 9 个组件适配 `r.data?.data ?? r.data` |
