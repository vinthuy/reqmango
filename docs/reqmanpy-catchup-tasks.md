# ReqManPy 追平 PlaneAI �?开发任务清�?
> 基于 [ReqManPy vs PlaneAI 全面功能对比分析](./reqmanpy-vs-planeai.md)
> 所有任务可直接导入 Trae 进行开�?
---

## 概述

| 指标 | 数�?|
|------|------|
| **总任务数** | 52 �?|
| **P0 工作�?* | 20.5 人天 |
| **P1 工作�?* | 28.5 人天 |
| **P2 工作�?* | 19.5 人天 |
| **P3 工作�?* | 19.5 人天 |
| **总工作量** | **�?8 人天（≈4.5 人月�?* |
| **推荐执行顺序** | MCP �?Agent �?GitHub �?Wiki �?Slack �?P2 �?P3 |

---

## P0 �?核心差距（必须追赶）

### 1. AI Agent 指派体系 [10d]

**目标�?* �?AI Agent 像真实成员一样被 @提及、分配工作项、自动执行操�?
#### 1.1 后端：Agent 模型 [1d]

- **文件�?* `backend/internal/model/agent.go`
- **内容�?*
  - 定义 Agent 表结构：`id`, `workspace_id`, `name`, `avatar`, `type` (builtin/custom), `capabilities` (JSON), `status` (active/inactive), `model`, `system_prompt`, `created_at`, `updated_at`
  - 定义 Agent 能力枚举：`create_issue`, `update_issue`, `search`, `comment`, `label`, `assign`, `triage`, `summarize`, `plan`

#### 1.2 后端：Agent 执行引擎 [3d]

- **文件�?* `backend/internal/service/agent_service.go`
- **内容�?*
  - `ExecuteTask(agentID, taskType, context)` �?根据任务类型调用 LLM，执行操�?  - `TriageIssue(issueID)` �?自动分析 Issue，建议类�?优先�?标签/指派�?  - `AutoLabel(issueID)` �?根据内容自动打标�?  - `AutoAssign(issueID)` �?根据负载和能力自动分�?  - `SummarizeCycle(cycleID)` �?总结 Sprint 进展
  - 所�?Agent 操作写入 IssueActivity 审计日志
  - 支持 `@agent-name` 触发关键�?
#### 1.3 后端：Agent 分配 API [1d]

- **文件�?* `backend/internal/handler/agent_handler.go`
- **路由�?*
  - `GET /api/v1/workspaces/:slug/agents` �?列出可用 Agent
  - `POST /api/v1/workspaces/:slug/agents/:id/dispatch` �?派发任务�?Agent
  - `POST /api/v1/workspaces/:slug/agents/:id/assign-to-issue` �?�?Agent 设为 Issue 指派�?  - `GET /api/v1/workspaces/:slug/agents/:id/activity` �?查看 Agent 操作历史

#### 1.4 后端：@mention 解析增强 [1d]

- **文件�?* `backend/internal/service/comment_service.go`
- **内容�?*
  - 扩展现有 @mention 解析逻辑，支�?`@agent-name` 模式
  - 检测到 Agent 提及时，自动触发 Agent 执行与该评论上下文相关的操作
  - 返回 Agent 响应作为评论回复

#### 1.5 前端：Agent 选择�?[1d]

- **文件�?* `frontend/src/components/AgentSelector.vue`
- **内容�?*
  - 在指派人选择器中混入 AI Agent 列表（用不同颜色/图标区分人类和Agent�?  - Agent 头像显示�?AI 图标
  - Hover 显示 Agent 能力列表

#### 1.6 前端：Agent 审计面板 [1d]

- **文件�?* `frontend/src/components/AgentAuditLog.vue`
- **内容�?*
  - 时间线展�?Agent 的所有操�?  - 每条记录显示：操作类型、目标、结果、耗时
  - 筛选：�?Agent / 操作类型 / 时间范围

#### 1.7 AI：Agent 自动分诊增强 [2d]

- **文件�?* `backend/internal/service/ai_service.go`（增强现�?TriageAnalyze�?- **内容�?*
  - 增强分诊逻辑：自动分�?�?自动标签 �?自动分配 �?自动路由
  - 支持自定义分诊规则（按条件路由到不同项目/Cycle/指派人）
  - Webhook 回调：分诊完成后触发通知

---

### 2. MCP Server（Model Context Protocol�?[10.5d]

**目标�?* �?Claude/Cursor/VS Code �?AI 编程工具直接操作 ReqManPy

#### 2.1 新建 MCP Server 子项�?[2d]

- **文件�?* `mcp-server/` (新建目录)
- **内容�?*
  - Go module 初始�?(`mcp-server/go.mod`)
  - MCP 协议实现：Stdio 传输�?  - `main.go` 入口，注册所�?Tools
  - 连接�?ReqManPy REST API 的客户端封装

#### 2.2 Project Tools [1d]

- **文件�?* `mcp-server/tools/project.go`
- **Tools�?*
  - `list_projects` �?列出工作空间所有项�?  - `get_project` �?获取项目详情和统�?
#### 2.3 Issue Tools [2d]

- **文件�?* `mcp-server/tools/issue.go`
- **Tools�?*
  - `create_issue` �?创建 Issue（title, description, priority, state, type, assignee, labels�?  - `list_issues` �?列表查询（filters, pagination�?  - `get_issue` �?获取单个 Issue 详情含评�?  - `update_issue` �?更新 Issue 状�?优先�?指派�?标签

#### 2.4 Search Tools [1d]

- **文件�?* `mcp-server/tools/search.go`
- **Tools�?*
  - `search_issues` �?自然语言搜索 + RQL 查询
  - `ai_search` �?通过 AI 端点进行语义搜索

#### 2.5 Cycle Tools [1d]

- **文件�?* `mcp-server/tools/cycle.go`
- **Tools�?*
  - `list_cycles` �?列出项目所�?Sprint
  - `get_cycle_progress` �?获取 Sprint 进展和燃尽数�?  - `add_issue_to_cycle` �?�?Issue 加入 Sprint

#### 2.6 Meta Tools [1d]

- **文件�?* `mcp-server/tools/meta.go`
- **Tools�?*
  - `list_members` �?列出工作空间成员
  - `get_states` �?获取项目工作流状�?  - `get_labels` �?获取项目标签
  - `list_issue_types` �?获取 Issue 类型

#### 2.7 HTTP 传输支持 [2d]

- **文件�?* `mcp-server/transport/http.go`
- **内容�?*
  - HTTP + OAuth 传输模式（Plane Cloud 同款�?  - PAT Token 认证模式（CI/CD 场景�?  - 端点：`/http/mcp`

#### 2.8 配置模板与文�?[0.5d]

- **文件�?* `mcp-server/README.md`
- **内容�?*
  - Claude Desktop 配置示例 (`claude_desktop_config.json`)
  - Claude Code 配置命令 (`claude mcp add reqman ...`)
  - VS Code / Cursor 配置说明
  - 示例对话：常见使用场�?
---

## P1 �?关键集成（直接影响采用率�?
### 3. GitHub / GitLab 原生集成 [10d]

#### 3.1 Git 集成模型 [1d]

- **文件�?* `backend/internal/model/git_integration.go`
- **内容�?*
  - `GitIntegration` 表：`id`, `project_id`, `provider` (github/gitlab), `repo_url`, `repo_name`, `access_token` (加密存储), `webhook_secret`, `active`, `sync_prs`, `sync_commits`, `sync_branches`

#### 3.2 GitHub Webhook 接收�?[2d]

- **文件�?* `backend/internal/handler/git_webhook_handler.go`
- **内容�?*
  - 接收 GitHub push / pull_request / issues / issue_comment 事件
  - 签名验证（HMAC-SHA256�?  - 事件分发到对应处理器

#### 3.3 Smart Commit 解析 [1.5d]

- **文件�?* `backend/internal/service/git_service.go`
- **内容�?*
  - 解析 commit message 中的关键词：`fixes DEMO-42`, `closes DEMO-42`, `ref DEMO-42`
  - 自动更新对应 Issue 状�?  - 自动添加 commit 引用评论

#### 3.4 PR / Branch 链接 [1d]

- **文件�?* `backend/internal/service/git_service.go`
- **内容�?*
  - PR 创建时自动关联对�?Issue
  - Branch �?Issue 的关联显�?  - PR 合并时自动关�?Issue
  - Issue 详情展示关联 PR 状态（Open/Merged/Closed�?
#### 3.5 前端：Git 集成设置�?[1.5d]

- **文件�?* `frontend/src/views/GitIntegration.vue`
- **内容�?*
  - OAuth 授权流程（GitHub App / GitLab App�?  - 仓库选择器（搜索 �?绑定�?  - 同步选项：PR / Commit / Branch
  - Webhook 状态和重试按钮

#### 3.6 前端：Issue Git 面板 [1d]

- **文件�?* `frontend/src/components/IssueGitPanel.vue`
- **内容�?*
  - Issue 详情右侧面板
  - 显示关联�?PR 列表�?Title, 状�? 作�? 更新时间�?  - 显示关联�?Branch 列表
  - 显示关联�?Commit 列表（最�?条）

#### 3.7 GitLab 支持 [2d]

- **内容�?*
  - 复用 GitHub 架构
  - GitLab API 适配�?  - GitLab Webhook 格式适配

---

### 4. 实时协作 Wiki [13d]

#### 4.1 WebSocket 服务 [2d]

- **文件�?* `backend/internal/service/ws_hub.go`
- **内容�?*
  - WebSocket Hub 管理所有连�?  - 房间机制（按 Page ID 分组�?  - 客户端注�?注销/广播
  - JWT 认证升级

#### 4.2 协作算法 [3d]

- **文件�?* `backend/internal/service/collab_service.go`
- **内容�?*
  - 基于 OT (Operational Transformation) 或简�?CRDT
  - 操作类型：insert, delete, replace
  - 冲突解决策略（Last-Write-Wins + 自动合并�?  - 操作广播（同一房间其他客户端）

#### 4.3 页面版本历史 [1d]

- **文件�?* `backend/internal/model/page_version.go`
- **内容�?*
  - `PageVersion` 表：`id`, `page_id`, `content`, `created_by`, `created_at`, `version_number`
  - 每次保存（非每次按键）创建版本快�?  - API：`GET /api/v1/pages/:id/versions`, `GET /api/v1/pages/:id/versions/:vid`

#### 4.4 多人 Diff 视图 API [1d]

- **文件�?* `backend/internal/handler/page_handler.go`
- **内容�?*
  - `GET /api/v1/pages/:id/versions/:vid/diff` �?返回两个版本之间�?diff
  - 支持多用户修改标记（不同颜色区分不同作者）

#### 4.5 实时协作编辑�?[3d]

- **文件�?* `frontend/src/components/CollabEditor.vue`
- **内容�?*
  - 基于 TipTap / Quill 的富文本编辑�?  - 多人光标显示（不同颜色）
  - WebSocket 连接管理
  - 离线编辑队列 + 重连恢复

#### 4.6 版本历史 Diff 面板 [1.5d]

- **文件�?* `frontend/src/components/PageVersionDiff.vue`
- **内容�?*
  - 双栏 Diff 视图（左旧右新）
  - 版本时间线侧边栏
  - 回滚到特定版本（带确认对话框�?
#### 4.7 斜杠命令 [1.5d]

- **文件�?* `frontend/src/components/SlashCommand.vue`
- **命令列表�?*
  - `/table` �?插入表格
  - `/code` �?插入代码�?  - `/image` �?插入图片
  - `/mermaid` �?插入 Mermaid 图表
  - `/equation` �?插入数学公式
  - `/embed` �?嵌入外部链接预览
  - `/issue` �?引用 Issue
  - `/divider` �?分割�?
---

### 5. Slack 集成 [5.5d]

#### 5.1 Slack Bot 接收�?[2d]

- **文件�?* `backend/internal/handler/slack_handler.go`
- **内容�?*
  - 处理 Slack Slash Commands：`/reqman create`, `/reqman search`, `/reqman status`
  - 处理 Slack Events API（消息事件，消息操作�?  - 签名验证

#### 5.2 Slack �?Issue [1.5d]

- **文件�?* `backend/internal/service/slack_service.go`
- **内容�?*
  - `CreateIssueFromThread(channelID, threadTS)` �?�?Slack 线程转为 Issue
  - 消息线程映射表存�?  - 支持双向同步：Issue 评论 �?Slack 回复

#### 5.3 Issue �?Slack 通知 [1d]

- **文件�?* `backend/internal/service/slack_service.go`
- **内容�?*
  - 配置频道映射（项�?�?Slack 频道�?  - Issue 创建/状态变�?评论 实时推送到对应频道
  - 使用 Slack Block Kit 美化消息格式
  - 支持项目更新摘要定时推�?
#### 5.4 前端：Slack 配置�?[1d]

- **文件�?* `frontend/src/views/SlackIntegration.vue`
- **内容�?*
  - OAuth "Add to Slack" 按钮
  - Bot Token 配置
  - 频道映射表格
  - 通知事件勾选（创建/状态变�?评论�?
---

## P2 �?体验增强

### 6. AI 交互式图�?[6d]

#### 6.1 后端：AI Chart 生成 [2d]

- **文件�?* `backend/internal/service/ai_service.go`（新�?Chart 方法�?- **内容�?*
  - 自然语言 �?图表查询 �?结构化图表定�?  - 支持的图表类型：bar, line, pie, doughnut, burndown, cumulative_flow
  - 返回格式：`{ type, title, labels, datasets, options }`

#### 6.2 前端：AI Chart 渲染�?[2d]

- **文件�?* `frontend/src/components/AIChart.vue`
- **内容�?*
  - 基于 Chart.js / ECharts 的动态图表渲�?  - 类型切换（柱状图 �?饼图 �?折线图）
  - 导出 PNG / SVG

#### 6.3 图表交互 [1d]

- **文件�?* `frontend/src/components/AIChart.vue`
- **内容�?*
  - 点击数据�?�?跳转到对�?Issue 列表（带自动过滤�?  - Tooltip 显示详情
  - 时间范围缩放

#### 6.4 AI Chat 集成 [1d]

- **文件�?* `frontend/src/components/AIChatSidebar.vue`
- **内容�?*
  - 对话中识别图表意图，AI 返回图表定义
  - 聊天内嵌图表渲染
  - 支持追问调整�?换成饼图" / "只看本月"�?
---

### 7. 多渠�?Intake [5.5d]

#### 7.1 Email Intake [2d]

- **文件�?* `backend/internal/handler/email_intake_handler.go`
- **内容�?*
  - 方案 A：Mailgun/Postmark Webhook 接收
  - 方案 B：IMAP 定时拉取
  - 邮件解析：From→提交人, Subject→标�? Body→描�? 附件→Issue 附件
  - 去重检测（Message-ID�?
#### 7.2 API Intake 增强 [1d]

- **文件�?* `backend/internal/handler/intake_handler.go`（增强现有）
- **内容�?*
  - API Key 认证模式（用于外部系统集成）
  - 自定义字段填充支�?  - 回调 URL 配置（创建成功后通知外部系统�?
#### 7.3 Intake 管理面板 [1.5d]

- **文件�?* `frontend/src/views/IntakeManagement.vue`
- **内容�?*
  - 多渠道统一管理 UI
  - 各通道统计（提交数/接受�?平均处理时间�?  - 通道启用/禁用开�?
#### 7.4 Intake 规则配置 [1d]

- **文件�?* `frontend/src/components/IntakeRuleConfig.vue`
- **内容�?*
  - 按来�?关键�?类型自动路由
  - 条件：if 来源=Email �?主题�?Bug" �?项目=工程�? 类型=Bug
  - 规则优先级排�?
---

### 8. 看板 Swimlane 增强 [3.5d]

#### 8.1 Swimlane 逻辑 [2d]

- **文件�?* `frontend/src/components/IssueKanban.vue`（增强）
- **内容�?*
  - 分组维度：指派人 / 类型 / 优先�?/ 自定义字�?  - 水平泳道 + 垂直�?= 矩阵视图
  - 折叠/展开单个泳道
  - "无分�?泳道用于未分�?Issue

#### 8.2 泳道拖拽 [1.5d]

- **文件�?* `frontend/src/components/IssueKanban.vue`
- **内容�?*
  - 跨列拖拽（状态变更）
  - 跨泳道拖拽（属性变更：改指派人/类型等）
  - 拖拽预览 + 动画
  - API 调用 + 乐观更新

---

### 9. 工作空间跨项目视�?[4.5d]

#### 9.1 跨项�?API [1d]

- **文件�?* `backend/internal/handler/workspace_handler.go`
- **内容�?*
  - `GET /api/v1/workspaces/:slug/issues` �?跨项�?Issue 聚合
  - 支持按项�?状�?指派�?类型过滤
  - 分页支持

#### 9.2 工作空间看板 [2d]

- **文件�?* `frontend/src/views/WorkspaceKanban.vue`
- **内容�?*
  - 多项�?Issue 同一看板
  - 泳道按项目分�?  - 项目颜色标签区分
  - 全局拖拽

#### 9.3 工作空间日历 [1.5d]

- **文件�?* `frontend/src/views/WorkspaceCalendar.vue`
- **内容�?*
  - 多项�?Issue 同一日历
  - 项目颜色编码
  - 日期范围过滤
  - Sprint 时间线叠�?
---

## P3 �?基础设施与生�?
### 10. Kubernetes Helm 部署 [6d]

#### 10.1 Helm Chart 骨架 [2d]

- **目录�?* `deploy/helm/reqmanpy/`
- **内容�?*
  - `Chart.yaml` �?元数�?  - `values.yaml` �?默认配置�?  - `values-dev.yaml` �?开发环境覆�?  - `values-prod.yaml` �?生产环境覆盖

#### 10.2 K8s 资源模板 [1.5d]

- **目录�?* `deploy/helm/reqmanpy/templates/`
- **内容�?*
  - `backend-deployment.yaml` + `backend-service.yaml`
  - `frontend-deployment.yaml` + `frontend-service.yaml`
  - `postgres-statefulset.yaml` + `postgres-service.yaml` + `postgres-pvc.yaml`
  - `redis-deployment.yaml` + `redis-service.yaml`（用�?Session 和协作）
  - `ingress.yaml`
  - `configmap.yaml` + `secret.yaml`

#### 10.3 水平扩展 [2d]

- **文件�?* `backend/internal/config/`（增强）+ Helm values
- **内容�?*
  - Backend 无状态化确认
  - Redis Session 存储（替代内存）
  - HPA 配置（CPU/Memory 自动扩缩�?  - 就绪探针 / 存活探针

#### 10.4 部署文档 [0.5d]

- **文件�?* `deploy/helm/README.md`
- **内容�?*
  - 前置条件（K8s 1.25+, Helm 3.12+, kubectl�?  - 快速开始：`helm install reqmanpy ./reqmanpy -f values-prod.yaml`
  - 升级：`helm upgrade reqmanpy ./reqmanpy`
  - 卸载：`helm uninstall reqmanpy`

---

### 11. SDK 发布 [7d]

#### 11.1 Python SDK [3d]

- **目录�?* `sdk/python/`
- **内容�?*
  - `setup.py` / `pyproject.toml`
  - `reqman/__init__.py` �?Client 类，支持 API Key �?JWT 认证
  - `reqman/workspace.py` �?Workspace 操作
  - `reqman/project.py` �?Project CRUD
  - `reqman/issue.py` �?Issue CRUD + 搜索 + 批量操作
  - `reqman/cycle.py` �?Cycle 操作
  - `reqman/comment.py` �?Comment 操作
  - `reqman/webhook.py` �?Webhook 管理
  - `reqman/ai.py` �?AI 端点（chat, search, analyze�?  - 类型标注 + Docstring
  - 单元测试（pytest + vcrpy 录制模式�?
#### 11.2 Node.js SDK [3d]

- **目录�?* `sdk/node/`
- **内容�?*
  - `package.json` �?npm 包配�?  - `src/index.ts` �?ReqManClient �?  - `src/workspace.ts`
  - `src/project.ts`
  - `src/issue.ts`
  - `src/cycle.ts`
  - `src/comment.ts`
  - `src/webhook.ts`
  - `src/ai.ts`
  - TypeScript 完整类型定义
  - 单元测试（Vitest + nock�?
#### 11.3 SDK 文档 [1d]

- **文件�?* `sdk/python/README.md`, `sdk/node/README.md`
- **内容�?*
  - 安装说明
  - 快速开始（5行代码创�?Issue�?  - API 参�?  - 错误处理
  - 示例代码（常见使用场景）

---

### 12. 导入器扩�?[5d]

#### 12.1 Jira Cloud 导入 [2d]

- **文件�?* `backend/internal/service/import_service.go`
- **内容�?*
  - Jira REST API v3 调用
  - 映射：Project �?Project, Issue �?Issue, Sprint �?Cycle, Epic �?Issue(L0), Comments �?Comments, Attachments �?Attachments
  - 自定义字段映射配�?  - 用户映射配置
  - 增量导入支持

#### 12.2 Linear 导入 [1.5d]

- **文件�?* `backend/internal/service/import_service.go`
- **内容�?*
  - Linear CSV 解析 + Linear API 模式
  - 映射：Team �?Project, Issue �?Issue, Cycle �?Cycle, Project �?Module, Comment �?Comment
  - Linear 优先�?状态映射表
  - Linear 标签映射

#### 12.3 导入向导 [1.5d]

- **文件�?* `frontend/src/components/ImportWizard.vue`
- **内容�?*
  - 步骤1：选择源（Jira Cloud / Jira DC / Linear / Asana / CSV�?  - 步骤2：连接配置（URL + Token�?  - 步骤3：数据映射预览（字段对应表）
  - 步骤4：确认并执行（进度条 + 实时日志�?  - 步骤5：结果报告（成功/失败/跳过数量�?
---

### 13. 投票功能 [1.5d]

#### 13.1 后端：Vote 模型 + API [1d]

- **文件�?* `backend/internal/model/vote.go`
- **内容�?*
  - `Vote` 表：`id`, `issue_id`, `user_id`, `vote` (+1/-1), `created_at`
  - 唯一约束�?issue_id, user_id)
- **Handler�?*
  - `POST /api/v1/issues/:id/vote` �?投票（body: `{vote: +1|-1}`�?  - `DELETE /api/v1/issues/:id/vote` �?取消投票
  - `GET /api/v1/issues/:id/votes` �?获取投票统计

#### 13.2 前端：投票按�?[0.5d]

- **文件�?* `frontend/src/components/VoteButton.vue`
- **内容�?*
  - 上下箭头按钮组（�?赞同 / �?反对�?  - 显示净投票�?  - 已投票状态高�?  - Issue 列表中也显示投票数徽�?
---

## 汇�?
### 按优先级

| 优先�?| 模块 | 任务�?| 工作�?| 状�?|
|--------|------|:---:|:---:|:---:|
| **P0** | 1. AI Agent 指派体系 | 7 | 10.0d | 🔴 |
| **P0** | 2. MCP Server | 8 | 10.5d | 🔴 |
| **P1** | 3. GitHub/GitLab 集成 | 7 | 10.0d | 🟡 |
| **P1** | 4. 实时协作 Wiki | 7 | 13.0d | 🟡 |
| **P1** | 5. Slack 集成 | 4 | 5.5d | 🟡 |
| **P2** | 6. AI 交互式图�?| 4 | 6.0d | 🟢 |
| **P2** | 7. 多渠�?Intake | 4 | 5.5d | 🟢 |
| **P2** | 8. 看板 Swimlane | 2 | 3.5d | 🟢 |
| **P2** | 9. 工作空间跨项目视�?| 3 | 4.5d | 🟢 |
| **P3** | 10. K8s Helm 部署 | 4 | 6.0d | 🔵 |
| **P3** | 11. SDK (Python+Node) | 3 | 7.0d | 🔵 |
| **P3** | 12. 导入器扩�?| 3 | 5.0d | 🔵 |
| **P3** | 13. 投票功能 | 2 | 1.5d | 🔵 |

### 工作量汇�?
| 类别 | 工作�?|
|------|:---:|
| P0 合计 | **20.5d** |
| P1 合计 | **28.5d** |
| P2 合计 | **19.5d** |
| P3 合计 | **19.5d** |
| **总计** | **�?8d (�?.5 人月)** |

### 推荐执行甘特

```
Week 1-2:   ████████████████ P0.2 MCP Server (10.5d)
Week 2-4:                   ████████████████████ P0.1 AI Agent (10d)
Week 4-6:                                       ████████████████████ P1.3 GitHub (10d)
Week 6-8:                                                                       ██████████████████████████ P1.4 Wiki (13d)
Week 8-9:                                                                                                 ██████████�?P1.5 Slack (5.5d)
Week 9-10:                                                                                                              ████████████ P2 (19.5d)
Week 10-12:                                                                                                                          ████████████ P3 (19.5d)
```

> **说明�?* 以上为单人力估算�?-3 人团队可将时间缩�?6-8 周�?