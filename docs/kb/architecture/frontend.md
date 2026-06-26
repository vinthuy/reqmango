# Frontend Architecture（前端架构）

**最后更新**: 2026-06-26

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
├── api/ (22 模块)
│   ├── index.ts          — Axios 实例 + JWT 拦截器
│   ├── ai.ts             — AI Chat SSE / Search / Create
│   ├── auth.ts, workspace.ts, project.ts
│   ├── issue.ts, cycle.ts, module.ts
│   ├── project-settings.ts, custom-field.ts
│   ├── workflow.ts, issue-type.ts, estimate-point.ts
│   ├── comment.ts, notification.ts, attachment.ts
│   ├── saved-view.ts, page.ts, template.ts
│   ├── release.ts, work-item-template.ts, relation.ts
│   └── rql.ts
│
├── types/ (19 模块)
│   ├── index.ts
│   ├── ai.ts, attachment.ts, comment.ts
│   ├── custom-field.ts, cycle.ts, estimate-point.ts
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
├── router/index.ts — 14 条路由
│
├── views/ (15 视图)
│   ├── Login.vue, Register.vue, Home.vue
│   ├── Workspace.vue, WorkspaceSettings.vue
│   ├── Project.vue, ProjectPages.vue, ProjectSettings.vue
│   ├── IssueCreate.vue, IssueDetail.vue
│   ├── CycleCreate.vue, CycleDetail.vue
│   ├── CustomFields.vue, IssueTypeList.vue
│   └── WorkflowDetail.vue
│
├── components/ (50+ 组件)
│   ├── AI: AIChatSidebar.vue, AICreateDialog.vue
│   ├── Issue: IssueCard.vue, IssueList.vue, IssueKanban.vue,
│   │          IssueDetailPanel.vue, SubIssuePanel.vue
│   ├── Cycle: CycleCard.vue, CycleList.vue, CycleDetailPanel.vue,
│   │          CycleBurndownChart.vue, CycleProgressCard.vue
│   ├── Module: ModuleCard.vue, ModuleList.vue, ModuleTree.vue,
│   │           ModuleDetailPanel.vue, ModuleFormModal.vue
│   ├── Config: CustomFieldForm.vue, CustomFieldList.vue,
│   │           CustomFieldManager.vue, CustomFieldValueInput.vue
│   ├── Workflow: WorkflowManager.vue, WorkflowRuleCard.vue,
│   │             WorkflowRuleForm.vue, WorkflowRuleList.vue,
│   │             StateTransition.vue
│   ├── Automation: AutomationForm.vue, AutomationList.vue,
│   │                AutomationManager.vue, AutomationTemplateList.vue
│   ├── Template: ProjectTemplateManager.vue,
│   │              WorkItemTemplateManager.vue,
│   │              WorkspaceIssueTypeManager.vue
│   ├── Release: ReleaseList.vue
│   ├── Estimate: EstimatePointManager.vue
│   ├── Page: PageTree.vue
│   ├── View: SavedViewSelector.vue, QuickFilterChips.vue
│   ├── Import: ImportIssuesModal.vue, QuickCreateInput.vue
│   ├── Common: CommentList.vue, AttachmentManager.vue,
│   │           NotificationCenter.vue, RichTextEditor.vue,
│   │           UserSelect.vue, LabelSelector.vue,
│   │           ConfirmDialog.vue, ProjectDetail.vue,
│   │           ProjectMemberList.vue, ProjectIssueTypeManager.vue,
│   │           RelationTypeManager.vue, WorkspaceMemberList.vue
│   └── RQL: RQLHistory.vue, RQLInput.vue
│
└── composables/ (3 组合式函数)
    ├── useConfirm.ts — 确认对话框
    ├── useRQL.ts     — RQL 查询历史
    └── useAI.ts      — AI SSE 消费
```

## 路由表

| 路径 | 视图 | 说明 |
|------|------|------|
| `/` | Home | 工作空间列表 |
| `/login` | Login | 登录 |
| `/register` | Register | 注册 |
| `/workspace/:slug` | Workspace | 工作空间详情 + 项目列表 |
| `/workspace/:slug/settings` | WorkspaceSettings | 工作空间配置 |
| `/workspace/:slug/project/:id` | Project | 项目主页 (Issues/Cycles/Modules/Pages) |
| `/workspace/:slug/project/:id/pages` | ProjectPages | 页面文档 |
| `/workspace/:slug/project/:id/settings` | ProjectSettings | 项目配置 |
| `.../settings/workflows/:workflowId` | WorkflowDetail | 工作流详情 |
| `/workspaces/:wid/projects/:pid/issues/:iid` | IssueDetail | 工作项详情 |
| `/workspaces/:wid/projects/:pid/issues/new` | IssueCreate | 创建工作项 |
| `/workspaces/:wid/projects/:pid/custom-fields` | CustomFields | 自定义字段 |
| `/workspaces/:wid/projects/:pid/issue-types` | IssueTypeList | 工作项类型 |
| `/workspaces/:wid/projects/:pid/cycles/new` | CycleCreate | 创建周期 |
| `/workspaces/:wid/projects/:pid/cycles/:cid` | CycleDetail | 周期详情 |

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
