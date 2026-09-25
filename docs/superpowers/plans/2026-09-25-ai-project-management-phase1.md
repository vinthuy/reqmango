# AI 项目管理 Phase 1 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 AI 从「Agent 控制台」收成 Plane 式项目管理日常面：Ask/Build Sidecar、Issue 指派 Agent、@mention、审计、Intake 分诊、创建前去重。

**Architecture:** 以前端接线与轻量 API 增量为主；复用现有 `ai/chat|create`、`issue_agent`、`comment` mention、`intake` analyze、`AgentAuditLog`/`TriagePanel` 组件。不删 Harness/Loop 代码，只退出默认导航。

**Tech Stack:** Vue 3 + TypeScript + Pinia、Go/Gin、现有 SSE chat。

**Spec:** `docs/superpowers/specs/2026-09-25-ai-project-management-redesign.md`

## Global Constraints

- 对标 Plane AI 项目管理主路径；禁止引入 Multica / MCP 分发 / Agent 市场叙事到 UI 文案。
- 默认侧栏不得再把 `/agents` 控制台当主入口；高级入口文案须含「实验」或「高级」。
- Build/Create：**无用户确认不得落库写操作**。
- Ask 模式：工具层禁止静默 create/update（若已有防护则加固文案与前端切换提示）。
- 去重：**只警告不拦截**（提供「仍要创建」）。
- 每完成一个 Task：相关单测/手测通过后再勾选；小步提交（用户未要求推远程则只本地 commit 若用户要求 commit）。
- 中英文 i18n 键同步（`frontend/src/locales` 或项目现有 i18n 文件）。

---

## 文件结构总览

| 区域 | 文件 | 动作 |
|------|------|------|
| 导航 | `frontend/src/components/AppSidebar.vue`、`TopBar.vue`（若有 Agents 链） | 隐藏默认入口；可选高级链 |
| Copilot | `frontend/src/components/AICopilot.vue`、`composables/useAI.ts`、`types/ai.ts` | Ask/Build 收敛 + 上下文条 |
| 父页传上下文 | `Project.vue`、`IssueDetail.vue`、`Workspace.vue`（若打开 Copilot） | 传入 issue/cycle/project |
| Issue 指派 | `IssuePropertySidebar.vue`、`IssueDetail.vue`、`api/issue-agent.ts` | assign 主路径；dispatch 次要 |
| @mention | `CommentList.vue`、可选 `ChatInput.vue` | 成员 + Agent 候选 |
| 审计 | `IssueDetail.vue` 或活动 Tab + `AgentAuditLog.vue` | 挂载 |
| Intake | 项目路由/设置 + `TriagePanel.vue` | 挂载 |
| 去重 API | `backend/internal/handler` + `service` + `router`；`frontend/src/api` | duplicate-check |
| 创建 UI | Issue 创建页 / 快速创建 / Copilot Create 预览 | 展示相似项 |
| i18n | 中英文 locale | 新文案 |
| 文档 | `docs/dev/pipeline-status.md`、`active/README.md` | 勾选进度 |

---

## Task 1: 导航降权 — 隐藏默认「AI Agents」

**Files:**
- Modify: `frontend/src/components/AppSidebar.vue`
- Modify: `frontend/src/components/TopBar.vue`（若存在 Agents 链接）
- Modify: locale 文件（若 label 硬编码）

- [x] **Step 1:** 从 `AppSidebar` 默认 `navItems`（或等价数组）移除指向 `/agents` 的「AI Agents」项。

- [x] **Step 2:** 在工作区设置页或侧栏底部「高级」折叠区增加链接：文案 `高级 Agent（实验）` / `Advanced Agents (experimental)`，`to` 仍为 `/workspace/:slug/agents` 或现有 agents 路由（保持路由可用）。

- [x] **Step 3:** 手测：登录后侧栏第一屏无 AI Agents；高级入口仍能打开原控制台。

- [ ] **Step 4:** Commit（若用户要求提交）：`feat(ai): demote agents console from primary nav`

---

## Task 2: Copilot 收敛为 Ask | Build + 上下文条

**Files:**
- Modify: `frontend/src/components/AICopilot.vue`
- Modify: `frontend/src/composables/useAI.ts`
- Modify: `frontend/src/types/ai.ts`
- Modify: `frontend/src/views/Project.vue`、`IssueDetail.vue`（传 props）
- Modify: i18n

**Props 约定（AICopilot）：**
```ts
projectId: number
workspaceId?: number
issueId?: number | null
cycleId?: number | null
pageId?: number | null
view?: 'project' | 'issue_detail' | 'cycle' | 'page' | 'workspace'
```

- [x] **Step 1:** 将 Tab 列表改为仅 `ask` | `build`。原 `create` UI 作为 Build 子面板或 Build 下「创建工作项」入口保留（预览+确认逻辑不删）。原 `chart` 快捷并入 Ask 建议芯片或 Build 结果动作（二选一，优先 Ask 芯片调用现有 chart API）。移除独立 `agent` Tab；若需「对某 Agent 说话」，在 Build 顶部用可选 `AgentSelector`（非主 Tab）。

- [x] **Step 2:** 在面板顶加上下文条：显示当前 Issue 标识 / Cycle 名 / 仅项目；提供「仅用项目上下文」清除 issue/cycle。

- [x] **Step 3:** `useAI` / chat 请求 body 增加 `issue_id`、`cycle_id`、`page_id`（后端若已忽略未知字段则先传；若绑定校验失败再改 DTO）。

- [x] **Step 4:** `IssueDetail` 打开 Copilot 时传入 `issueId`；`Project` 传入 `projectId`（及当前 cycle 若有）。

- [ ] **Step 5:** 手测：在 Issue 页 Ctrl+J，上下文条显示该 Issue；Ask 提问应能引用当前 Issue（抽查一次 SSE）。

- [ ] **Step 6:** Commit：`feat(ai): collapse copilot to Ask/Build with view context`

---

## Task 3: Issue 侧栏 — 指派 Agent 为主路径

**Files:**
- Modify: `frontend/src/components/IssuePropertySidebar.vue`
- Modify: `frontend/src/views/IssueDetail.vue`
- Use: `frontend/src/api/issue-agent.ts`

- [x] **Step 1:** 加载 Issue 时拉取 `issueAgentApi.getStatus(issueId)`（或详情里已有字段则绑定），展示当前 `agent_name` / `task_status`。

- [x] **Step 2:** `AgentSelector` 变更时调用 `assign-agent`（或清除时 `unassign-agent`），替代「仅改 localAgentId」。保存成功后刷新 status。

- [x] **Step 3:** 「Dispatch」按钮降为次要样式（text/outline），主按钮文案改为「指派」类 i18n；未选 Agent 时禁用。

- [ ] **Step 4:** 手测：指派后刷新页面 Agent 仍在；取消指派清空。

- [ ] **Step 5:** Commit：`feat(ai): persist agent assignee on issue sidebar`

---

## Task 4: 评论 @ — 候选含 Agent

**Files:**
- Modify: `frontend/src/components/CommentList.vue`
- Optionally: `frontend/src/components/chat/ChatInput.vue`（若 Issue Chat 仍缺 Agent）
- Use: 现有 agents list API（与 `AgentSelector` 同源）

- [x] **Step 1:** 加载项目/工作区 Agent 列表；与 members 合并为 mention 候选（Agent 项带 `kind: 'agent'` 与可解析的 `@name`，与后端 `HandleMention` 约定一致——通常为 agent name/slug）。

- [x] **Step 2:** 下拉展示区分「成员」与「Agent」（小标签或图标）；插入后的文本格式须能被后端识别。

- [ ] **Step 3:** 手测：评论 `@` 某 Agent 名，提交后出现 Agent 回复或活动（取决于后端现行为）。

- [ ] **Step 4:** Commit：`feat(ai): include agents in comment mention picker`

---

## Task 5: Issue 活动流挂载 Agent 审计

**Files:**
- Modify: `frontend/src/views/IssueDetail.vue`（或活动 Tab 子组件）
- Modify/Use: `frontend/src/components/AgentAuditLog.vue`（按 issue 过滤若支持；否则包一层只显示该 issue）

- [x] **Step 1:** 在 Issue 详情活动/时间线区域增加「Agent」过滤或子区块，渲染 `AgentAuditLog`（传入 `issueId` / `workspaceId`）。

- [x] **Step 2:** 若组件仅支持 workspace 级，增加 optional `issueId` prop 并在请求参数过滤；无 API 参数则前端 filter。

- [ ] **Step 3:** 手测：对 Issue dispatch/assign 后活动区可见记录。

- [ ] **Step 4:** Commit：`feat(ai): show agent audit on issue activity`

---

## Task 6: 挂载 Intake Triage 面板

**Files:**
- Modify: `frontend/src/router/index.ts`（若需新路由）
- Modify: 项目设置或 Issues 布局视图（如 `Project.vue` tab / `ProjectSettings`）
- Use: `frontend/src/components/TriagePanel.vue`
- Modify: i18n + 侧栏/Tab 文案「分诊」

- [x] **Step 1:** 找齐 `TriagePanel` 所需 props（projectId 等），在项目内可达页面挂载（优先：项目设置「Intake/分诊」或 Issues 顶 Tab）。

- [x] **Step 2:** 确认 `ai-analyze` / intake API 路径与面板一致；修断链 import。

- [ ] **Step 3:** 手测：导航可进入；对一条 intake 点 AI 分析有返回。

- [ ] **Step 4:** Commit：`feat(ai): mount intake triage panel in project UI`

---

## Task 7: 创建前去重 API + UI

**Files:**
- Create/Modify: `backend/internal/service/issue_service.go`（或独立小函数）查重：同项目下 title ILIKE / 简单相似度，返回 top N
- Create/Modify: `backend/internal/handler/issue_handler.go` — `GET or POST /projects/:id/issues/duplicate-check` body `{ "name": "...", "description": "..." }`
- Modify: `backend/internal/router/router.go`
- Create: `*_test.go` 至少 1 个 sqlmock/表测或 handler 测
- Modify: `frontend/src/api/issue.ts`（或 `ai.ts`）
- Modify: Issue 创建页 / `QuickCreateInput.vue` / Copilot Create 预览区

- [x] **Step 1:** 写失败测试：同项目已有相似标题时接口返回非空 `duplicates`。

- [x] **Step 2:** 实现服务 + 路由；只读，权限与 list issues 同级（项目成员）。

- [x] **Step 3:** 前端防抖（300–500ms）调用；UI 列表展示相似 Issue 链接；主按钮旁「仍要创建」。

- [x] **Step 4:** Copilot Create 确认前同样展示 duplicates（若有）。

- [x] **Step 5:** `cd backend && go test` 相关包通过。

- [ ] **Step 6:** Commit：`feat(ai): warn on duplicate issues before create`

---

## Task 8: Build 模式写保护文案与切换提示

**Files:**
- Modify: `AICopilot.vue`、后端 `ai_service` system prompt（若 Ask 仍可能 tool 写）

- [x] **Step 1:** Ask 空态/发送区提示：只读；需要改数据请切 Build。

- [x] **Step 2:** 抽查 Ask 模式下模型若返回写计划，前端识别并提示切换（若已有则补文案）。

- [ ] **Step 3:** Commit：`feat(ai): clarify Ask vs Build write boundaries`

---

## Task 9: 文档与管线收尾

**Files:**
- Modify: `docs/dev/pipeline-status.md` — Phase 1 Implement 勾选进度
- Modify: `docs/dev/active/README.md` — 指向本 plan
- Modify: spec 文末「下一步」改为 Phase 1 进行中/完成

- [x] **Step 1:** 更新管线状态。

- [ ] **Step 2:** 自检本 plan 全部 checkbox；缺项补做或记入 Phase 2。

---

## 验收清单（对照 Spec §2.2）

| # | 标准 | Task |
|---|------|------|
| 1 | Issue 侧栏可持久指派 Agent | 3 |
| 2 | 评论 @ 含 Agent | 4 |
| 3 | Ctrl+J 带当前上下文 | 2 |
| 4 | 仅 Ask \| Build | 2 |
| 5 | 创建前疑似重复 | 7 |
| 6 | Intake 分诊可打开 | 6 |
| 7 | 活动流可见 Agent | 5 |
| 8 | 默认侧栏无 Agents 主入口 | 1 |

---

## 不在本 Plan

- Phase 2 开箱 Agent 模板、事件自动分诊产品化  
- Harness/Loop UI  
- MCP/CLI/PAT 设置  
- Slack  

---

## 执行建议

使用 `superpowers:subagent-driven-development` 按 Task 1→9 顺序执行；每 Task 结束再开下一 Task。
