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

### BUG-04 RQL assignee_id / cycle_id / module_id / label 按名称匹配数字 ID

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/rql/executor.go:58-67`（字段映射）、`:304-357`（buildEqualRaw）、`:510-584`（buildInRaw） |
| **类型** | 逻辑错误 |
| **影响** | FilterBar 筛选「负责人=某人」返回空结果 |
| **原因** | 字段类型定义为 `"user"/"cycle"/"module"/"label"`，SQL 按 `display_name='5'` 匹配 |
| **复现** | FilterBar 添加筛选 → assignee_id = 332 → 返回 0 条（已实测确认） |
| **对比** | `state_id = 1` 正确返回 1 条（state_id 映射为 `"number"` 类型） |

---

### BUG-05 parent_id 自引用未被拦截

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go:720-733` |
| **类型** | 数据完整性问题 |
| **影响** | 树视图无限循环、列表返回重复条目 |
| **原因** | 只校验父 Issue 是否存在，不检查 `parent_id != issueID` 和循环引用 |
| **复现** | `PUT /api/v1/issues/5048 {"parent_id":5048}` → 200 OK（已实测确认） |

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

### BUG-07 XSS — DescriptionHTML 无清理直接存储和输出

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go:64-71`（Create）、`:551-553`（Update）、`:1236`（BuildResponse） |
| **类型** | 安全漏洞 |
| **影响** | 存储型 XSS，`<script>alert('XSS')</script>` 可在所有用户浏览器执行 |
| **原因** | 未使用 bluemonday 等 HTML 清理库，直接存储 `req.DescriptionHTML` |

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

### BUG-11 附件下载完全不可用

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

### BUG-13 前端路由 / 后端路由命名不一致（4 处）

| 前端期望 | 实际后端 | 状态码 |
|---------|---------|--------|
| `POST /api/v1/workspaces/:ws/projects` | `POST /api/v1/projects?workspace_id=` | 404 |
| `GET /api/v1/projects/:id/states` | `GET /api/v1/projects/:id/settings/states` | 404 |
| `GET /api/v1/projects/:id/saved-views` | `GET /api/v1/projects/:id/views` | 404 |
| `POST /api/v1/issues/:id/comments` | `POST /api/v1/comments` | 404 |

---

### BUG-14 注册 TOCTOU 竞态条件

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/auth_service.go:28-78` |
| **类型** | 并发Bug |
| **影响** | 并发注册相同邮箱 → 返回 500 Internal Error 而非 409 Conflict |
| **原因** | SELECT COUNT 和 INSERT 之间无事务保护 |

---

### BUG-15 Panic Recovery 返回 HTML 而非 JSON

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

### BUG-19 vue-tsc 28 个编译错误

| 文件 | 行号 | 错误 |
|------|------|------|
| `IssueTabActivity.vue` | 215 | `Cannot find name 'activity'` — **运行时错误** |
| `SubIssuesPanel.vue` | 69 | `Property 'username' does not exist on type 'UserLite'` |
| `ReportBuilder.vue` | 645 | `'generate' is declared but never read`（死代码） |
| `IssueDetailHeader.vue` | 78 | `'props' is declared but its value is never read` |

---

### BUG-20 @mention 不支持中文用户名

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/comment_service.go:108-109` |
| **类型** | 功能Bug |
| **影响** | `@张三` 无法被正确解析和通知 |
| **原因** | `isUsernameChar` 只匹配 `[a-zA-Z0-9_-]` |

---

### BUG-21 活动日志显示状态 ID 而非名称

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go:583-584` |
| **类型** | 用户体验 |
| **影响** | 用户看到 "3 → 5" 而非 "待处理 → 进行中" |

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

### BUG-24 RQL LIKE 模式中 `%` `_` 不转义

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/rql/executor.go:424-427` |
| **类型** | 边界Bug |
| **影响** | 搜索 "100%" 行为异常 |

---

### BUG-25 五个 Issue 视图数据加载模式不一致

| 视图 | 问题 |
|------|------|
| IssueGantt | 无 `workspace_id` 参数、硬编码 `limit=500` |
| IssueCalendar | 同上 |
| IssueTreeView | 使用独立 `/issues/tree` 端点、筛选条件子集 |

---

### BUG-26 批量操作静默跳过失败项

| 字段 | 内容 |
|------|------|
| **文件** | `backend/internal/service/issue_service.go:1098-1123` |
| **类型** | 错误处理 |
| **影响** | 部分成功部分失败不通知调用者 |
| **原因** | 流程校验失败用 `continue` 跳过 |

---

### BUG-27 前端 axios 无 timeout / 无统一错误拦截

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/api/index.ts:3-8` |
| **类型** | 用户体验 |
| **影响** | 请求挂起无反馈；429/5xx 错误静默 |

---

### BUG-28 en-US 翻译缺 18 个 key

| 字段 | 内容 |
|------|------|
| **文件** | `frontend/src/locales/en-US.json`（2228 keys vs zh-CN 2245 keys） |
| **类型** | i18n |
| **影响** | 英文界面显示原始 key 名（如 `activity.changedType`） |

---

### BUG-29 CSV 导出无 UTF-8 BOM + CSV 导入不处理 BOM

| 操作 | 问题 |
|------|------|
| 导出 | Go CSV writer 不写 BOM → Excel 打开中文乱码 |
| 导入 | Go CSV reader 不处理 BOM → Excel 导出文件首列名解析错误 |

---

### BUG-30 Gin 尾斜杠 301 重定向可能导致前端请求失败

| 字段 | 内容 |
|------|------|
| **影响** | `GET /api/v1/issues/5048/` → 301 → 如果 axios 不跟随则失败 |

---

## 🔵 低优先级

### BUG-31 仅 3 个 Pinia Store 覆盖 40+ 模块

无 Issue/Project/Notification Store，各组件独立管理状态，修改后其他视图不刷新。

---

### BUG-32 角色创建 workspace_id 未从 URL 提取

`POST /api/v1/workspaces/:wsParam/roles` — Handler 不解析 `:wsParam`，角色创建为 `workspace_id=NULL`。

---

### BUG-33 useConfirm promise 路由跳转时永不 resolve

`frontend/src/composables/useConfirm.ts:13` — 对话框打开时用户导航离开，promise 永久挂起。

---

### BUG-34 页面 Diff 算法简陋

`frontend/src/components/PageVersionDiff.vue:92-117` — 逐行对比，插入一行导致后续所有行标记为变更。

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

## 📊 统计

| 严重度 | 数量 | 类别 |
|--------|------|------|
| 🔴 严重 | 8 | 编译错误、安全漏洞、功能不可用、数据完整性 |
| 🟡 高 | 10 | 路由不匹配、权限缺失、错误处理 |
| 🟠 中 | 12 | i18n、UX、边界检查 |
| 🔵 低 | 8 | 架构改进、代码规范 |
| **总计** | **38** | |
