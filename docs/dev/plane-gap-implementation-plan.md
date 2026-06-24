# ReqManPy 对标 Plane 信息架构 — 差距补全计划

> **基线**: ReqManPy master 分支 (2026-06-25)
> **对标**: [Plane Enterprise 信息架构](docs/kb/architecture/plane-enterprise-info-architecture.md)
> **拆分原则**: 每个 Task 独立可验证、无跨 Task 代码冲突、有明确的完成标准

---

## Task 清单总览

| # | Task | 优先级 | 层级 | 预估文件数 | 依赖 |
|---|------|--------|------|-----------|------|
| 1 | Project Labels (工作空间级项目分类标签) | P0 | 工作空间 | 4 | 无 |
| 2 | Project Features 功能开关 | P0 | 项目 | 5 | 无 |
| 3 | Workspace General Settings (工作空间基本设置) | P3 | 工作空间 | 3 | 无 |
| 4 | Import/Export 数据导入导出 | P1 | 项目 | 5 | 无 |
| 5 | Work Item Templates (工作项模板) | P1 | 项目 | 7 | 无 |
| 6 | RBAC 角色权限系统 | P1 | 工作空间 | 8 | 无 |
| 7 | Estimates 估算系统增强 (Categories + Time) | P2 | 项目 | 4 | 无 |
| 8 | Type Hierarchy 可视化编辑器 | P2 | 工作空间 | 2 | 无 |
| 9 | Project States (项目健康度状态) | P3 | 工作空间 | 4 | Task 1 |
| 10 | Automations 分层 (工作空间/项目分离) | P2 | 混合 | 4 | 无 |

> **P3 项目**: Plane Runner 脚本引擎、Integrations、Recurring Work Items 暂不纳入 — 工作量大且依赖外部系统。

---

## Task 1: Project Labels — 工作空间级项目分类标签

**Plane 对应**: `Workspace Settings → Projects → Project Labels (Business)`
**ReqManPy 现状**: 完全没有。当前 Labels 仅在项目内作用于工作项。
**设计**: 工作空间管理员创建项目分类标签池，各项目从中选择标签标记自身。

### 1.1 数据模型

```go
// backend-go/internal/model/project_label.go

type ProjectLabel struct {
    BaseModel
    Name        string `gorm:"size:50;not null" json:"name"`
    Color       string `gorm:"size:7;default:#6B7280" json:"color"`
    Description *string `gorm:"size:255" json:"description"`
    WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
    // M:N with projects via project_project_labels
}

type ProjectProjectLabel struct {
    ProjectID      uint64 `gorm:"primaryKey;autoIncrement:false"`
    ProjectLabelID uint64 `gorm:"primaryKey;autoIncrement:false"`
}
```

### 1.2 API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/workspaces/:wsParam/project-labels` | 列出工作空间所有项目标签 |
| POST | `/workspaces/:wsParam/project-labels` | 创建项目标签 |
| PUT | `/workspaces/:wsParam/project-labels/:id` | 更新 |
| DELETE | `/workspaces/:wsParam/project-labels/:id` | 删除 |
| POST | `/projects/:projectId/project-labels` | 给项目打标签 |
| DELETE | `/projects/:projectId/project-labels/:labelId` | 移除项目标签 |

### 1.3 前端

- 新增 `components/ProjectLabelManager.vue` — 工作空间 Settings 中管理标签池
- 修改 `views/WorkspaceSettings.vue` — 新增 "Project Labels" tab
- 修改 `components/ProjectDetail.vue` — 显示和编辑项目标签
- 修改 `views/Home.vue` — 按标签筛选/分组项目

### 1.4 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `backend-go/internal/model/project_label.go` |
| 新建 | `backend-go/internal/dto/request/project_label.go` |
| 新建 | `backend-go/internal/dto/response/project_label.go` |
| 新建 | `backend-go/internal/service/project_label_service.go` |
| 新建 | `backend-go/internal/handler/project_label_handler.go` |
| 新建 | `frontend/src/components/ProjectLabelManager.vue` |
| 新建 | `frontend/src/types/project-label.ts` |
| 新建 | `frontend/src/api/project-label.ts` |
| 修改 | `backend-go/internal/router/router.go` — 注册路由 |
| 修改 | `backend-go/cmd/server/main.go` — AutoMigrate |
| 修改 | `frontend/src/views/WorkspaceSettings.vue` — 新增 tab |
| 修改 | `frontend/src/views/Home.vue` — 标签展示 |

---

## Task 2: Project Features 功能开关

**Plane 对应**: `Project Settings → Features` (Cycles/Modules/Views/Pages/Intake/Time Tracking)
**ReqManPy 现状**: 所有功能默认启用，项目无法选择性关闭。
**设计**: Project 表新增 `features` JSONB 字段，存储各功能的启用/禁用状态。

### 2.1 数据模型

```go
// 在现有 model/project.go 的 Project struct 中新增字段:
Features json.RawMessage `gorm:"type:jsonb" json:"features"`
// 示例: {"cycles":true,"modules":true,"views":true,"pages":true,"intake":false,"time_tracking":false}
```

### 2.2 API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/projects/:projectId/features` | 获取功能开关状态 |
| PUT | `/projects/:projectId/features` | 更新功能开关 |

### 2.3 前端

- 修改 `views/ProjectSettings.vue` — 新增 "Features" tab（开关列表）

### 2.4 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `backend-go/internal/model/project.go` — 新增 Features 字段 |
| 修改 | `backend-go/internal/service/project_service.go` — GetFeatures/UpdateFeatures |
| 修改 | `backend-go/internal/handler/project_handler.go` — 2 个新方法 |
| 修改 | `backend-go/internal/router/router.go` — 注册路由 |
| 修改 | `frontend/src/views/ProjectSettings.vue` — Features tab |

---

## Task 3: Workspace General Settings

**Plane 对应**: `Workspace Settings → General` (名称/URL/删除)
**ReqManPy 现状**: 工作空间创建后无设置页面修改基本属性。
**设计**: WorkspaceSettings 新增 "General" tab — 名称、描述、时区编辑 + 删除工作空间。

### 3.1 前端

- 修改 `views/WorkspaceSettings.vue` — 新增 "General" tab

### 3.2 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `frontend/src/views/WorkspaceSettings.vue` — General tab |
| 无需 | 后端 API 已支持 (PATCH/DELETE /workspaces/:id) |

---

## Task 4: Import/Export 数据导入导出

**Plane 对应**: `Workspace Settings → Import/Export`
**ReqManPy 现状**: 完全没有导入导出。
**设计**: CSV/JSON 双格式支持。导出直接用现有查询；导入需字段映射 UI。

### 4.1 API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/projects/:projectId/export/csv` | 导出工作项为 CSV |
| GET | `/projects/:projectId/export/json` | 导出工作项为 JSON |
| POST | `/projects/:projectId/import/csv` | 从 CSV 导入（multipart） |
| POST | `/projects/:projectId/import/json` | 从 JSON 导入 |

### 4.2 前端

- 新增 `components/ImportDialog.vue` — 文件上传 + 字段映射
- 修改 `components/IssueList.vue` — "Export" 按钮

### 4.3 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `backend-go/internal/service/import_export_service.go` |
| 修改 | `backend-go/internal/handler/issue_handler.go` — 4 个新方法 |
| 修改 | `backend-go/internal/router/router.go` |
| 新建 | `frontend/src/components/ImportDialog.vue` |
| 修改 | `frontend/src/components/IssueList.vue` |

---

## Task 5: Work Item Templates 工作项模板

**Plane 对应**: `Project Settings → Templates` (工作项创建模板)
**ReqManPy 现状**: 每次创建需手动填写所有字段。
**设计**: 项目管理员为每种 Issue Type 预设一个模板（默认 title、description、assignee、labels、priority 等）。

### 5.1 数据模型

```go
// backend-go/internal/model/work_item_template.go

type WorkItemTemplate struct {
    BaseModel
    Name         string          `gorm:"size:100;not null" json:"name"`
    Description  *string         `gorm:"size:255" json:"description"`
    IssueTypeID  uint64          `gorm:"not null;index" json:"issue_type_id"`
    Defaults     json.RawMessage `gorm:"type:jsonb" json:"defaults"`
    // defaults 示例: {"name":"Bug: ","priority":"high","assignee_ids":[1,2],"label_ids":[3]}
    ProjectID    uint64          `gorm:"not null;index" json:"project_id"`
    WorkspaceID  uint64          `gorm:"not null" json:"workspace_id"`
    IsDefault    bool            `gorm:"default:false" json:"is_default"`
}
```

### 5.2 API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/projects/:projectId/work-item-templates` | 列出 |
| POST | `/projects/:projectId/work-item-templates` | 创建 |
| PUT | `/projects/:projectId/work-item-templates/:id` | 更新 |
| DELETE | `/projects/:projectId/work-item-templates/:id` | 删除 |

### 5.3 前端

- 新建 `components/WorkItemTemplateManager.vue` — 模板 CRUD
- 修改 `views/ProjectSettings.vue` — "Templates" tab
- 修改 `views/IssueCreate.vue` — 创建时选择模板预填

### 5.4 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `backend-go/internal/model/work_item_template.go` |
| 新建 | `backend-go/internal/dto/request/work_item_template.go` |
| 新建 | `backend-go/internal/dto/response/work_item_template.go` |
| 新建 | `backend-go/internal/service/work_item_template_service.go` |
| 新建 | `backend-go/internal/handler/work_item_template_handler.go` |
| 新建 | `frontend/src/components/WorkItemTemplateManager.vue` |
| 新建 | `frontend/src/types/work-item-template.ts` |
| 新建 | `frontend/src/api/work-item-template.ts` |
| 修改 | `backend-go/internal/router/router.go` |
| 修改 | `backend-go/cmd/server/main.go` |
| 修改 | `frontend/src/views/ProjectSettings.vue` |
| 修改 | `frontend/src/views/IssueCreate.vue` |

---

## Task 6: RBAC 角色权限系统

**Plane 对应**: `Workspace Settings → Roles & Permissions`
**ReqManPy 现状**: 硬编码 3 级角色 (Guest=5, Member=15, Admin=20)，无自定义权限。
**设计**: 新增 `Role` 和 `Permission` 模型，支持自定义角色 + 权限方案。

### 6.1 数据模型

```go
// backend-go/internal/model/role.go

type Role struct {
    BaseModel
    Name          string `gorm:"size:50;not null" json:"name"`
    Description   string `gorm:"size:255" json:"description"`
    WorkspaceID   uint64 `gorm:"not null;index" json:"workspace_id"`
    IsSystem      bool   `gorm:"default:false" json:"is_system"` // 内置角色不可删除
    Permissions   []RolePermission `gorm:"foreignKey:RoleID" json:"permissions"`
}

type RolePermission struct {
    RoleID       uint64 `gorm:"primaryKey;autoIncrement:false"`
    Resource     string `gorm:"primaryKey;size:50"` // "project", "issue", "cycle", "member"...
    Action       string `gorm:"primaryKey;size:20"` // "create", "read", "update", "delete", "manage"
}

type MemberRole struct {
    MemberID     uint64 `gorm:"primaryKey"`          // workspace_member.id or project_member.id
    MemberType   string `gorm:"primaryKey;size:20"`  // "workspace" | "project"
    RoleID       uint64 `gorm:"not null"`
}
```

### 6.2 API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/workspaces/:wsParam/roles` | 列出所有角色 |
| POST | `/workspaces/:wsParam/roles` | 创建自定义角色 |
| PUT | `/workspaces/:wsParam/roles/:id` | 更新角色权限 |
| DELETE | `/workspaces/:wsParam/roles/:id` | 删除（非系统角色） |
| PUT | `/workspaces/:wsParam/members/:userId/role` | 变更成员角色 |

### 6.3 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `backend-go/internal/model/role.go` |
| 新建 | `backend-go/internal/service/role_service.go` |
| 新建 | `backend-go/internal/handler/role_handler.go` |
| 新建 | `frontend/src/components/RoleManager.vue` |
| 新建 | `frontend/src/types/role.ts` |
| 新建 | `frontend/src/api/role.ts` |
| 修改 | `backend-go/internal/middleware/auth.go` — 权限检查中间件 |
| 修改 | `backend-go/internal/router/router.go` |
| 修改 | `backend-go/cmd/server/main.go` |
| 修改 | `frontend/src/views/WorkspaceSettings.vue` — "Roles" tab |

---

## Task 7: Estimates 估算系统增强

**Plane 对应**: `Project Settings → Estimates` (Points/Categories/Time)
**ReqManPy 现状**: 仅有 `EstimatePointManager` (Points 模式 — Fibonacci 数列)。
**设计**: 扩展现有 EstimatePoint 为统一的 Estimate 系统，支持 Points / Categories / Time 三种模式。

### 7.1 修改范围

- 不动数据模型（现有 `estimate_points` 表已足够）
- 新增 `estimate_type` 字段区分模式
- 前端 `EstimatePointManager.vue` 增加模式选择和对应配置 UI

### 7.2 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `frontend/src/components/EstimatePointManager.vue` — 模式切换 + Categories/Time 配置 |
| 修改 | `frontend/src/types/estimate-point.ts` |
| 修改 | `frontend/src/api/estimate-point.ts` |
| 修改 | `frontend/src/components/IssueCard.vue` — 展示适配 |

---

## Task 8: Type Hierarchy 可视化编辑器

**Plane 对应**: `Workspace Settings → Work Item Types → Hierarchy Tab`
**ReqManPy 现状**: 创建 Type 时手动设 level/parent_type_id，无可视化。
**设计**: 纯前端改进 — 树形拖拽编辑器，底层 API 不变。

### 8.1 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/TypeHierarchyEditor.vue` — 树形拖拽 |
| 修改 | `frontend/src/components/WorkspaceIssueTypeManager.vue` — 新增 "Hierarchy" tab |

---

## Task 9: Project States (工作空间级项目健康度)

**Plane 对应**: `Workspace Settings → Project States`
**ReqManPy 现状**: 无项目生命周期状态。
**设计**: 工作空间级定义项目状态，6 个阶段：Draft → Planning → Execution → Monitoring → Completed → Cancelled。

### 9.1 数据模型

```go
// 在 model/project.go 中新增 ProjectStateID 外键
ProjectStateID *uint64 `json:"project_state_id"`
```

工作空间级项目管理状态表 `project_states`。

### 9.2 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `backend-go/internal/model/project_state.go` |
| 新建 | `backend-go/internal/handler/project_state_handler.go` |
| 新建 | `frontend/src/components/ProjectStateManager.vue` |
| 修改 | `backend-go/internal/router/router.go` |
| 修改 | `backend-go/cmd/server/main.go` |
| 修改 | `frontend/src/views/WorkspaceSettings.vue` |
| 修改 | `frontend/src/views/Home.vue` — 显示项目状态 |

---

## Task 10: Automations 分层

**Plane 对应**: 工作空间 Automation + 项目 Automation 双层级
**ReqManPy 现状**: 工作空间和项目 Settings 各有一套 Automation，但数据混合存储。
**设计**: `automation_rules` 表增加 `scope` 字段，明确归属级别。工作空间级 Automation 在创建时可选适用范围。

### 10.1 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `backend-go/internal/model/workflow.go` — AutomationRule 新增 Scope 字段 |
| 修改 | `backend-go/internal/service/workflow_service.go` |
| 修改 | `frontend/src/components/AutomationForm.vue` — Scope 选择器 |
| 修改 | `frontend/src/views/WorkspaceSettings.vue` — 工作空间 Automation 列表适配 |

---

## 实施节奏建议

```
Week 1 (P0 快速见效):
  Task 1: Project Labels        ← 独立，4 文件
  Task 2: Features 功能开关      ← 独立，5 文件
  Task 3: Workspace General      ← 独立，仅前端

Week 2 (P1 核心能力):
  Task 4: Import/Export          ← 独立，5 文件
  Task 5: Work Item Templates    ← 独立，7 文件

Week 3 (P1 基础设施):
  Task 6: RBAC 权限系统          ← 独立但工作量大，8 文件

Week 4 (P2 体验增强):
  Task 7: Estimates 增强
  Task 8: Type Hierarchy 编辑器
  Task 10: Automations 分层

Week 5 (P3):
  Task 9: Project States
```

---

## 实现规范

所有 Task 遵循 [api-conventions.md](docs/kb/architecture/api-conventions.md) 和现有分层模式：

| 层 | 参考文件 | 模式要点 |
|---|---------|---------|
| Go Model | `model/notification.go` | 嵌入 BaseModel, uint64 ID, GORM tag, TableName() |
| Go Service | `service/notification_service.go` | 构造函数注入 db, AppError 错误, scope 过滤 |
| Go Handler | `handler/notification_handler.go` | `c.Get("currentUser")` 获取用户, ShouldBindJSON, Gin 响应 |
| Vue View | `views/WorkspaceSettings.vue` | `<script setup>` + Composition API, Tailwind, sidebar+tabs 布局 |
| Vue Component | `components/SavedViewSelector.vue` | Props/Emits 模式, defineProps<T>(), Axios 调用 |
| Vue API | `api/saved-view.ts` | 函数式导出, api.get/post/put/delete, 泛型返回 |
| Vue Types | `types/page.ts` | interface 定义, 与 Go DTO 字段对应 |

### 关键约定

1. **getUserID**: 必须用 `c.Get("currentUser").(*model.User).ID`，不能用 `c.Get("user_id")`
2. **AuthMiddleware**: 所有新路由必须挂载 `authMiddleware`
3. **AutoMigrate**: 新 Model 必须在 `cmd/server/main.go` 中注册
4. **路由注册**: 工作空间级用 `/workspaces/:wsParam/...`，项目级嵌套在 `projects.Group` 内
5. **前端 API base**: `api/index.ts` 的 axios 实例，baseURL 已设 `/api/v1`
