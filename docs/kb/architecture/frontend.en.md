# Frontend Architecture

**Last Updated**: 2026-07-13

---

## Tech Stack

| Component | Selection |
|-----------|-----------|
| Framework | Vue 3 (Composition API + `<script setup>`) |
| Build | Vite |
| Language | TypeScript |
| State Management | Pinia |
| Router | Vue Router 4 |
| CSS | Tailwind CSS |
| HTTP | Axios |
| Rich Text | TipTap |
| SSE | fetch + ReadableStream (AI streaming) |

## Project Structure

```
frontend/src/
├── main.ts
├── App.vue
├── style.css
│
├── api/ (41 modules)
│   ├── index.ts          — Axios instance + JWT interceptors
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
├── types/ (29 modules)
│   ├── index.ts
│   ├── ai.ts, attachment.ts, comment.ts
│   ├── agent.ts, initiative.ts, plugin.ts
│   ├── search-template.ts, dashboard.ts
│   ├── project-page-tab.ts, role.ts
│   ├── recurrence.ts, time-track.ts
│   ├── custom-field.ts, cycle.ts, estimate-point.ts
│   ├── filters.ts          — Filter types + FILTER_FIELDS + buildRQL/parseRQL
│   ├── issue.ts, issue-type.ts, module.ts
│   ├── notification.ts, page.ts, project.ts
│   ├── project-settings.ts, release.ts, saved-view.ts
│   ├── template.ts, workflow.ts, work-item-template.ts
│
├── stores/ (3 stores)
│   ├── auth.ts   — Authentication state
│   ├── cycle.ts  — Cycle CRUD + progress
│   └── module.ts — Module CRUD + tree
│
├── router/index.ts — 26 routes

├── views/ (23 views)
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
├── components/ (92 components)
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
│   ├── Filter: FilterBar.vue — Unified filter bar (replaces old filter components)
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
└── composables/ (12 composables)
    ├── useConfirm.ts     — Confirm dialog
    ├── useRQL.ts         — RQL query history
    ├── useAI.ts          — AI SSE consumption
    ├── useFilters.ts     — Filter state management (Provide/Inject architecture, RQL bidirectional sync)
    ├── usePermission.ts  — RBAC permission checks
    ├── useToast.ts       — Toast notifications
    ├── useDashboard.ts   — Dashboard state
    ├── useReportChart.ts — Report charts
    ├── useIssueFilters.ts — Issue filtering
    ├── useI18n.ts        — Internationalization
    ├── useMarkdown.ts    — Markdown processing
    └── useDarkMode.ts    — Dark mode
```

## Route Table

| Path | View | Description |
|------|------|-------------|
| `/` | Home | Workspace list |
| `/login` | Login | Login |
| `/register` | Register | Register |
| `/workspace/:slug` | Workspace | Workspace detail + project list |
| `/workspace/:slug/settings` | WorkspaceSettings | Workspace configuration |
| `/workspace/:slug/overview` | WorkspaceOverview | Workspace overview |
| `/workspace/:slug/roadmap` | Roadmap | Roadmap |
| `/workspace/:slug/initiatives` | Initiatives | Strategic goals |
| `/workspace/:slug/analytics` | WorkspaceAnalytics | Workspace analytics |
| `/workspace/:slug/project/:id` | Project | Project homepage |
| `/workspace/:slug/project/:id/pages` | ProjectPages | Pages/documents |
| `/workspace/:slug/project/:id/settings` | ProjectSettings | Project configuration |
| `.../settings/workflows/:workflowId` | WorkflowDetail | Workflow detail |
| `/workspace/:slug/project/:id/analytics` | Analytics | Project analytics |
| `/workspace/:slug/project/:id/dashboards` | Dashboard | Dashboard |
| `/workspaces/:wid/projects/:pid/issues/:iid` | IssueDetail | Issue detail |
| `/workspaces/:wid/projects/:pid/issues/new` | IssueCreate | Create issue |
| `/workspaces/:wid/projects/:pid/custom-fields` | CustomFields | Custom fields |
| `/workspaces/:wid/projects/:pid/issue-types` | IssueTypeList | Issue types |
| `/workspaces/:wid/projects/:pid/cycles/new` | CycleCreate | Create cycle |
| `/workspaces/:wid/projects/:pid/cycles/:cid` | CycleDetail | Cycle detail |
| `/intake/:projectId` | IntakeForm | Public intake |
| `/workspace/:slug/project/:id/issues/:iid` | IssueDetail | Issue detail (slug-style) |
| `/workspace/:slug/project/:id/issues/new` | IssueCreate | Create issue (slug-style) |
| `/workspace/:slug/project/:id/cycles/new` | CycleCreate | Create cycle (slug-style) |
| `/workspace/:slug/project/:id/cycles/:cid` | CycleDetail | Cycle detail (slug-style) |

## Architecture Patterns

### API Layer

```typescript
// api/index.ts — Axios instance
const api = axios.create({ baseURL: '/api/v1' })
// Request interceptor: auto-attach JWT Bearer token
// Response interceptor: 401 → redirect to login
```

### AI SSE Consumption

```typescript
// composables/useAI.ts
// fetch + ReadableStream to consume SSE
// Event types: text | tool_call | tool_result | done | error
```

### Component Patterns

- All components use `<script setup lang="ts">`
- Props: `defineProps<T>()`
- Emits: `defineEmits<T>()`
- State management: page-level uses local ref, cross-component uses Pinia store

---

## 🌐 Language

- **English** (this document)
- [中文文档](frontend.md)