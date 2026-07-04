# Frontend Architecture（前端架构）

**最后更新**: 2026-07-04

---

## 技术栈

| 组件 | 选型 |
|------|------|
| 框架 | Vue 3 (Composition API + `<script setup>`) |
| 构建 | Vite |
| 语言 | TypeScript |
| 状态管理 | Pinia |
| 路由 | Vue Router 4 |
| CSS | Tailwind CSS |
| HTTP | Axios |
| 富文本 | TipTap |
| SSE | fetch + ReadableStream (AI streaming) |

## 项目结构

```
frontend/src/
├── main.ts
├── App.vue
├── style.css
│
├── api/ (41 模块)
│   ├── index.ts          — Axios 实例 + JWT 拦截器
│   ├── ai.ts             — AI Chat SSE / Search / Create
│   ├── agent.ts, initiative.ts, plugin.ts
│   ├── search-template.ts, report.ts, dashboard.ts
│   ├── role.ts, webhook.ts, type-template.ts
│   ├── slack.ts, project-update.ts, mcp.ts
│   ├── github.ts, intake.ts, conditional-field.ts
│   ├── automation.ts, recurrence.ts, time-track.ts
│   ├── auth.ts, workspace.ts, project.ts
│   ├── issue.ts, cycle.ts, module.ts
│   ├── project-settings.ts, custom-field.ts
│   ├── workflow.ts, issue-type.ts, estimate-point.ts
│   ├── comment.ts, notification.ts, attachment.ts
│   ├── saved-view.ts, page.ts, template.ts
│   ├── release.ts, work-item-template.ts, relation.ts
│   └── rql.ts
│
├── types/ (29 模块)
│   ├── index.ts
│   ├── ai.ts, attachment.ts, comment.ts
│   ├── agent.ts, initiative.ts, plugin.ts
│   ├── search-template.ts, dashboard.ts
│   ├── project-page-tab.ts, role.ts
│   ├── recurrence.ts, time-track.ts
│   ├── custom-field.ts, cycle.ts, estimate-point.ts
│   ├── filters.ts          — 筛选条件类型 + FILTER_FIELDS + buildRQL/parseRQL
│   ├── issue.ts, issue-type.ts, module.ts
│   ├── notification.ts, page.ts, project.ts
│   ├── project-settings.ts, release.ts, saved-view.ts
│   ├── template.ts, workflow.ts, work-item-template.ts
│
├── stores/ (3 stores)
│   ├── auth.ts   — 认证状态
│   ├── cycle.ts  — 周期 CRUD + 进度
│   └── module.ts — 模块 CRUD + 树形
│
├── router/index.ts — 26 条路由

├── views/ (23 视图)
│   ├── Login.vue, Register.vue, Home.vue
│   ├── Workspace.vue, WorkspaceSettings.vue, WorkspaceOverview.vue
│   ├── WorkspaceAnalytics.vue, Analytics.vue
│   ├── Project.vue, ProjectPages.vue, ProjectSettings.vue
│   ├── IssueCreate.vue, IssueDetail.vue
│   ├── CycleCreate.vue, CycleDetail.vue
│   ├── CustomFields.vue, IssueTypeList.vue
│   ├── IntakeForm.vue, WorkflowDetail.vue
│   ├── Roadmap.vue, Initiatives.vue
│   ├── Dashboard.vue, PluginManager.vue
│

├── components/ (92 组件)
│   ├── Agent: AgentAuditLog.vue, AgentSelector.vue, AgentList.vue, AgentActivityLog.vue
│   ├── AI: AIChatSidebar.vue, AICreateDialog.vue, AISettingsPanel.vue,
│   │        AICopilot.vue, AIChartRenderer.vue, AIResultActions.vue
│   ├── Dashboard: DashboardGrid.vue, DashboardSidebar.vue,
│   │              WidgetCard.vue, WidgetConfigPanel.vue
│   ├── Palette: CommandPalette.vue
│   ├── Issue: IssueCard.vue, IssueList.vue, IssueKanban.vue,
│   │          IssueDetailPanel.vue, SubIssuePanel.vue,
│   │          IssueTreeView.vue, IssueGantt.vue, IssueCalendar.vue
│   ├── Cycle: CycleCard.vue, CycleList.vue, CycleDetailPanel.vue,
│   │          CycleBurndownChart.vue, CycleProgressCard.vue
│   ├── Module: ModuleCard.vue, ModuleList.vue, ModuleTree.vue,
│   │           ModuleDetailPanel.vue, ModuleFormModal.vue
│   ├── Config: CustomFieldForm.vue, CustomFieldList.vue,
│   │           CustomFieldManager.vue, CustomFieldValueInput.vue
│   ├── Workflow: WorkflowManager.vue, WorkflowRuleCard.vue,
│   │             WorkflowRuleForm.vue, WorkflowRuleList.vue,
│   │             StateTransition.vue, WorkflowVisualization.vue
│   ├── Automation: AutomationForm.vue, AutomationList.vue,
│   │                AutomationManager.vue, AutomationTemplateList.vue
│   ├── Template: ProjectTemplateManager.vue,
│   │              WorkItemTemplateManager.vue,
│   │              WorkspaceIssueTypeManager.vue
│   ├── Release: ReleaseList.vue, ReleaseRoadmap.vue
│   ├── Estimate: EstimatePointManager.vue
│   ├── Page: PageTree.vue, PageVersionDiff.vue, PageVersionPanel.vue,
│   │          PageTemplateSelector.vue, PageTabConfig.vue
│   ├── Filter: FilterBar.vue — 统一筛选栏（替代旧筛选组件）
│   ├── View: SavedViewSelector.vue, QuickFilterChips.vue
│   ├── Import: ImportIssuesModal.vue, QuickCreateInput.vue
│   ├── Integration: WorkspaceIntegrations.vue, WebhookManager.vue,
│   │                 ReportBuilder.vue
│   ├── Triage: TriagePanel.vue
│   ├── Search: SearchTemplateSelector.vue
│   ├── Common: CommentList.vue, AttachmentManager.vue, AttachmentPreview.vue,
│   │           NotificationCenter.vue, RichTextEditor.vue, TipTapEditor.vue,
│   │           UserSelect.vue, LabelSelector.vue, MultiSelectDropdown.vue,
│   │           ConfirmDialog.vue, ProjectDetail.vue,
│   │           ProjectMemberList.vue, ProjectIssueTypeManager.vue,
│   │           RelationTypeManager.vue, WorkspaceMemberList.vue,
│   │           ToastContainer.vue, ShortcutsPanel.vue,
│   │           LanguageSwitcher.vue, TopBar.vue, AppSidebar.vue,
│   │           RoleManagement.vue, TimeTrackPanel.vue,
│   │           RecurrenceConfig.vue, TreeNodeItem.vue
│   └── RQL: RQLHistory.vue, RQLInput.vue
│
└── composables/ (12 组合式函数)
    ├── useConfirm.ts     — 确认对话框
    ├── useRQL.ts         — RQL 查询历史
    ├── useAI.ts          — AI SSE 消费
    ├── useFilters.ts     — 筛选状态管理（Provide/Inject 架构，RQL 双向同步）
    ├── usePermission.ts  — RBAC 权限检查
    ├── useToast.ts       — Toast 通知
    ├── useDashboard.ts   — 仪表盘状态
    ├── useReportChart.ts — 报表图表
    ├── useIssueFilters.ts — 工作项筛选
    ├── useI18n.ts        — 国际化
    ├── useMarkdown.ts    — Markdown 处理
    └── useDarkMode.ts    — 深色模式
```

## 路由表

| 路径 | 视图 | 说明 |
|------|------|------|
| `/` | Home | 工作空间列表 |
| `/login` | Login | 登录 |
| `/register` | Register | 注册 |
| `/workspace/:slug` | Workspace | 工作空间详情 + 项目列表 |
| `/workspace/:slug/settings` | WorkspaceSettings | 工作空间配置 |
| `/workspace/:slug/overview` | WorkspaceOverview | 工作空间概览 |
| `/workspace/:slug/roadmap` | Roadmap | 路线图 |
| `/workspace/:slug/initiatives` | Initiatives | 战略目标 |
| `/workspace/:slug/analytics` | WorkspaceAnalytics | 工作空间分析 |
| `/workspace/:slug/project/:id` | Project | 项目主页 |
| `/workspace/:slug/project/:id/pages` | ProjectPages | 页面文档 |
| `/workspace/:slug/project/:id/settings` | ProjectSettings | 项目配置 |
| `.../settings/workflows/:workflowId` | WorkflowDetail | 工作流详情 |
| `/workspace/:slug/project/:id/analytics` | Analytics | 项目分析 |
| `/workspace/:slug/project/:id/dashboards` | Dashboard | 仪表盘 |
| `/workspaces/:wid/projects/:pid/issues/:iid` | IssueDetail | 工作项详情 |
| `/workspaces/:wid/projects/:pid/issues/new` | IssueCreate | 创建工作项 |
| `/workspaces/:wid/projects/:pid/custom-fields` | CustomFields | 自定义字段 |
| `/workspaces/:wid/projects/:pid/issue-types` | IssueTypeList | 工作项类型 |
| `/workspaces/:wid/projects/:pid/cycles/new` | CycleCreate | 创建周期 |
| `/workspaces/:wid/projects/:pid/cycles/:cid` | CycleDetail | 周期详情 |
| `/intake/:projectId` | IntakeForm | 公开提交 |
| `/workspace/:slug/project/:id/issues/:iid` | IssueDetail | 工作项详情（slug风格） |
| `/workspace/:slug/project/:id/issues/new` | IssueCreate | 创建工作项（slug风格） |
| `/workspace/:slug/project/:id/cycles/new` | CycleCreate | 创建周期（slug风格） |
| `/workspace/:slug/project/:id/cycles/:cid` | CycleDetail | 周期详情（slug风格） |

## 架构模式

### API 层

```typescript
// api/index.ts — Axios 实例
const api = axios.create({ baseURL: '/api/v1' })
// 请求拦截器：自动附加 JWT Bearer token
// 响应拦截器：401 → 跳转 login
```

### AI SSE 消费

```typescript
// composables/useAI.ts
// fetch + ReadableStream 消费 SSE
// 事件类型: text | tool_call | tool_result | done | error
```

### 组件模式

- 所有组件使用 `<script setup lang="ts">`
- Props: `defineProps<T>()`
- Emits: `defineEmits<T>()`
- 状态管理: 页面级用 local ref，跨组件用 Pinia store
