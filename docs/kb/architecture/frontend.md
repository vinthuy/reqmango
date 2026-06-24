# Frontend Architecture（前端架构）

**最后更新**: 2026-06-25

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

## 项目结构

```
frontend/src/
├── main.ts                         # 入口：创建 app, 挂载 Pinia/Router
├── App.vue                         # 根组件
├── style.css                       # Tailwind 入口
│
├── api/                            # API 调用层（17 个模块）
│   ├── index.ts                   # Axios 实例 + JWT 拦截器
│   ├── auth.ts, workspace.ts, project.ts
│   ├── issue.ts, cycle.ts, module.ts
│   ├── project-settings.ts, custom-field.ts
│   ├── workflow.ts, issue-type.ts, estimate-point.ts
│   ├── comment.ts, notification.ts, attachment.ts
│   ├── saved-view.ts              # 保存视图 API
│   └── page.ts                    # 页面文档 API
│
├── types/                          # TypeScript 类型定义（14 个模块）
│   ├── index.ts                   # 核心类型 (User, Workspace, Project, State, Label)
│   ├── issue.ts, cycle.ts, module.ts
│   ├── project-settings.ts, custom-field.ts
│   ├── workflow.ts, issue-type.ts, estimate-point.ts
│   ├── comment.ts, notification.ts, attachment.ts, project.ts
│   ├── saved-view.ts              # 保存视图类型
│   └── page.ts                    # 页面文档类型
│
├── stores/                         # Pinia 状态管理（3 个 store）
│   ├── auth.ts                    # 认证状态（login, logout, user）
│   ├── cycle.ts                   # 周期状态（CRUD, 进度, 燃尽图）
│   └── module.ts                  # 模块状态（CRUD, 树形, Issue 关联）
│
├── router/index.ts                # Vue Router 配置（13 条路由）
│
├── views/                          # 页面级组件（15 个视图）
│   ├── Login.vue, Register.vue, Home.vue
│   ├── Workspace.vue, WorkspaceSettings.vue
│   ├── Project.vue, ProjectPages.vue
│   ├── IssueCreate.vue, IssueDetail.vue
│   ├── CycleCreate.vue, CycleDetail.vue
│   ├── CustomFields.vue, IssueTypeList.vue
│   └── WorkflowDetail.vue
│
├── components/                     # 可复用组件（35 个）
│   ├── IssueCard.vue, IssueList.vue, IssueKanban.vue
│   ├── CycleCard.vue, CycleList.vue, CycleDetailPanel.vue
│   ├── ModuleCard.vue, ModuleList.vue, ModuleTree.vue
│   ├── CustomField*.vue, WorkflowRule*.vue
│   ├── CommentList.vue, AttachmentManager.vue
│   ├── NotificationCenter.vue, UserSelect.vue
│   ├── RichTextEditor.vue, ConfirmDialog.vue
│   ├── ProjectDetail.vue, ProjectIssueTypeManager.vue
│   ├── SavedViewSelector.vue       # 保存视图选择器
│   └── PageTree.vue                # 页面树形结构
│
└── composables/                   # 组合式函数
    └── useConfirm.ts              # 确认对话框
```

## 架构模式

### API 层

```typescript
// api/index.ts — Axios 实例
const api = axios.create({ baseURL: '/api/v1' })

// 请求拦截器：自动附加 JWT
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 响应拦截器：401 自动跳转登录
api.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) router.push('/login')
    return Promise.reject(err)
  }
)
```

### Store 模式 (Pinia)

采用 Composition API 风格：

```typescript
export const useCycleStore = defineStore('cycle', () => {
  const items = ref<Cycle[]>([])
  const loading = ref(false)

  async function fetchCycles(projectId: number) {
    loading.value = true
    const { data } = await cycleApi.list(projectId)
    items.value = data.data
    loading.value = false
  }

  return { items, loading, fetchCycles }
})
```

### 组件通信

- 父→子: Props
- 子→父: Emits
- 跨层级: Pinia Store
- 路由参数: `useRoute().params`

## 路由结构

```
/                                    → Home.vue
/login                               → Login.vue
/register                            → Register.vue
/workspace/:slug                     → Workspace.vue
/workspace/:slug/settings            → WorkspaceSettings.vue
/workspace/:slug/project/:id         → Project.vue
  /workspace/:slug/project/:id/issues/new     → IssueCreate.vue
  /workspace/:slug/project/:id/issues/:iid    → IssueDetail.vue
  /workspace/:slug/project/:id/cycles/new     → CycleCreate.vue
  /workspace/:slug/project/:id/cycles/:cid    → CycleDetail.vue
  /workspace/:slug/project/:id/custom-fields  → CustomFields.vue
  /workspace/:slug/project/:id/issue-types    → IssueTypeList.vue
/workspace/:slug/settings/workflows/:wid      → WorkflowDetail.vue
```

## 已实现组件数

| 功能域 | 组件数 | 视图数 |
|--------|--------|--------|
| Issue | 5 | 2 |
| Cycle | 5 | 2 |
| Module | 5 | 0（集成在 Project tab） |
| CustomField | 4 | 1 |
| Workflow | 3 | 1 |
| Comment | 1 | 0 |
| Attachment | 1 | 0 |
| Notification | 1 | 0 |
| Project | 3 | 1 |
| Common | 4 | 4 |
