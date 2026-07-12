# reqmango 功能差距补全计划
> **基线**: reqmango master 分支 (2026-06-25)
> **范围**: 聚焦工作项（Issue）管理全生命周期
> **拆分原则**: 每个 Task 独立可验证、无 Task 代码冲突、有明确的完成标准

---

## 差距总览：工作项管理 14 个维度

```
工作项管理全景                              reqmango 现状

📊 视图层
  ├─ List View     列表视图                   ✅ 已实现
  ├─ Kanban View   看板视图                   ✅ 已实现
  ├─ Calendar View 日历视图                   ✅ 已实现
  ├─ Gantt View    甘特图                    ✅ 已实现
  └─ Timeline      时间线/路线图              ✅ 已实现

📝 创建层
  ├─ Manual Create                              ✅ 已实现
  ├─ Work Item Templates 工作项模板           Task 1
  ├─ Quick Create     行内快速创建            Task 2
  └─ Bulk Import      CSV/JSON 批量导入        Task 3

📋 详情层
  ├─ Rich Text Description                      ✅ TipTap
  ├─ Comments (threaded)                        ✅ 已实现
  ├─ Activity Log                               ✅ 已实现
  ├─ Sub-items Panel  子工作项可展开面板        Task 4
  ├─ Page Linking     关联文档页面              Task 5
  ├─ Release Linking  关联 Release             Task 6
  └─ Cover Image      封面图                   Task 7

🔧 操作层
  ├─ Bulk Update / Delete                       ✅ 已实现
  ├─ Bulk Copy/Move   跨项目复制/移动           Task 8
  ├─ Convert Type     转换工作项类型           Task 8
  └─ Merge Duplicates 合并重复工作项           Task 9

🎯 属性层 (Custom Fields)
  ├─ 7 种字段类型(text/number/dropdown/boolean/date/member/url)
  ├─ Release Picker  Release 类型字段          Task 6
  ├─ Mandatory Validation 必填校验             Task 10
  └─ Conditional Fields   条件显隐规则         Task 11

🏗️ 层级层
  ├─ Parent/Child with depth                    ✅ 已实现
  └─ Type Hierarchy Rules 类型嵌套规则强制      Task 12

⏱️ 时间层
  ├─ Time Tracking   工时记录                  Task 13
  ├─ Start Date / Target Date                   ✅ 已实现
  └─ Recurring Work Items 重复工作项          Task 14

📊 估算层
  ├─ Points (Fibonacci)                         ✅ 已实现
  ├─ Categories (T-shirt sizes)                 Task 15
  └─ Time Estimates                             Task 15

🔍 筛选层
  ├─ RQL Query Language                         ✅ 已实现
  ├─ Saved Views                                ✅ 已实现
  └─ Quick Filters 预设筛选芯片              Task 16

📎 附件层
  ├─ 后端 Attachment model/service              Task 17
  └─ 前端 AttachmentManager (仅UI无后端)

🔔 通知层
  ├─ Notification model/API                     ✅ 已实现
  └─ Auto-trigger 自动触发通知                 Task 18

📥 接收层
  ├─ Intake Form 外部提交表单                  Task 19
  └─ Triage Mode 收件分诊模式                  Task 19
```

---

## Task 清单总览 (工作项管理专项)

| # | Task | 优先级 | 类型 | 文件数 | 依赖 |
|---|------|--------|------|--------|------|
| 1 | Work Item Templates 工作项模板 | **P0** | 新建 | 9 | - |
| 2 | Quick Create 行内快速创建 | **P0** | 前端 | 3 | - |
| 3 | Bulk Import CSV/JSON | **P0** | 新建 | 5 | - |
| 4 | Sub-items Panel 子工作项面板 | P1 | 前端 | 2 | - |
| 5 | Page Linking 关联文档页面 | P1 | 修改 | 3 | 已有 Pages |
| 6 | Release Management + Release Picker | P1 | 新建 | 9 | - |
| 7 | Cover Image 封面图 | P2 | 修改 | 2 | - |
| 8 | Bulk Copy/Move + Convert Type | P1 | 修改 | 3 | - |
| 9 | Merge Duplicates 合并重复工作项 | P2 | 修改 | 3 | - |
| 10 | Mandatory Field Validation 必填校验 | P1 | 修改 | 4 | - |
| 11 | Conditional Fields 条件显隐 | P2 | 修改 | 3 | Task 10 |
| 12 | Type Hierarchy Rules 类型嵌套规则 | P1 | 修改 | 4 | - |
| 13 | Time Tracking 工时记录 | P2 | 新建 | 7 | - |
| 14 | Recurring Work Items 重复工作项 | P2 | 新建 | 7 | - |
| 15 | Estimates Categories + Time | P1 | 修改 | 3 | - |
| 16 | Quick Filters 预设筛选芯片 | P1 | 前端 | 2 | - |
| 17 | Attachment Backend 附件后端 | P1 | 新建 | 5 | - |
| 18 | Notification Auto-trigger 自动通知 | P1 | 修改 | 3 | 已有 Notif |
| 19 | Intake & Triage 接收与分诊 | P3 | 新建 | 8 | - |

---

## Task 1: Work Item Templates 工作项模板
**功能**: 项目管理员为每种 Issue Type 预设创建模板（默认 title、description、assignee、labels、priority 等）。创建 Issue 时选择模板一键预填充。

### 数据模型

```go
// backend/internal/model/work_item_template.go

type WorkItemTemplate struct {
    BaseModel
    Name         string          `gorm:"size:100;not null" json:"name"`
    IssueTypeID  *uint64         `gorm:"index" json:"issue_type_id"`
    Defaults     json.RawMessage `gorm:"type:jsonb" json:"defaults"`
    // {"name":"Bug: [summary]","priority":"high","state_id":1,"assignee_ids":[1,2],"label_ids":[3,5]}
    IsDefault    bool            `gorm:"default:false" json:"is_default"`
    ProjectID    uint64          `gorm:"not null;index" json:"project_id"`
    WorkspaceID  uint64          `gorm:"not null" json:"workspace_id"`
}
```

### API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/projects/:projectId/work-item-templates` | 列表 |
| POST | `/projects/:projectId/work-item-templates` | 创建 |
| PUT | `/projects/:projectId/work-item-templates/:id` | 更新 |
| DELETE | `/projects/:projectId/work-item-templates/:id` | 删除 |

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `model/work_item_template.go` |
| 新建 | `dto/request/work_item_template.go` |
| 新建 | `dto/response/work_item_template.go` |
| 新建 | `service/work_item_template_service.go` |
| 新建 | `handler/work_item_template_handler.go` |
| 新建 | `frontend/src/components/WorkItemTemplateManager.vue` |
| 新建 | `frontend/src/types/work-item-template.ts` |
| 新建 | `frontend/src/api/work-item-template.ts` |
| 修改 | `router/router.go`, `cmd/server/main.go` |
| 修改 | `frontend/src/views/ProjectSettings.vue` 添加 "Templates" tab |
| 修改 | `frontend/src/views/IssueCreate.vue` 添加模板选择下拉 |

---

## Task 2: Quick Create 行内快速创建
**功能**: 在列表/看板顶部提供一行极简输入框，输入标题 + 回车即可创建，无需跳转完整表单。

### 前端改动

- 修改 `IssueList.vue` 在列表顶部新增行内输入框（title + type + priority 下拉 + 回车创建）
- 修改 `IssueKanban.vue` 在每列顶部新增 "+" 按钮 → 行内输入
- 新建 `components/QuickCreateInput.vue` 可复用行内创建组件

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/QuickCreateInput.vue` |
| 修改 | `frontend/src/components/IssueList.vue` |
| 修改 | `frontend/src/components/IssueKanban.vue` |

---

## Task 3: Bulk Import CSV/JSON

**功能**: 从 CSV/JSON 文件批量导入工作项，含字段映射 UI。

### API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/projects/:projectId/import/csv` | multipart/form-data |
| POST | `/projects/:projectId/import/json` | JSON body |
| GET | `/projects/:projectId/export/csv` | 导出 CSV |
| GET | `/projects/:projectId/export/json` | 导出 JSON |

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `service/import_export_service.go` |
| 新建 | `frontend/src/components/ImportDialog.vue` |
| 修改 | `handler/issue_handler.go` 添加 4 个方法 |
| 修改 | `router/router.go` |
| 修改 | `frontend/src/components/IssueList.vue` 添加导入/导出按钮 |

---

## Task 4: Sub-items Panel 子工作项可展开面板

**功能**: IssueDetail 中展示子工作项列表，支持行内展开/折叠、拖拽排序。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/SubIssuesPanel.vue` |
| 修改 | `frontend/src/views/IssueDetail.vue` 集成面板 |

---

## Task 5: Page Linking 关联文档页面

**功能**: 工作项可关联到项目内的 Pages 文档。Issue model 已有扩展能力。

### 数据模型

新增 M:N 关联表 `issue_pages`:

```go
type IssuePage struct {
    IssueID uint64 `gorm:"primaryKey;autoIncrement:false"`
    PageID  uint64 `gorm:"primaryKey;autoIncrement:false"`
}
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue.go` 新增 IssuePage 关联 |
| 新建 | `handler/issue_handler.go` 添加 AddPage/RemovePage/ListPages 方法 |
| 修改 | `frontend/src/views/IssueDetail.vue` 添加 Pages 关联区 |

---

## Task 6: Release Management + Release Picker

**功能**: 第一版 Release/Version 管理。Custom Field 新增 "Release" 类型。

### 数据模型

```go
type Release struct {
    BaseModel
    Name        string     `gorm:"size:100;not null" json:"name"`
    Version     string     `gorm:"size:50;not null" json:"version"`
    Description string     `gorm:"type:text" json:"description"`
    Status      string     `gorm:"size:30;default:planned" json:"status"` // planned/in_progress/released/cancelled
    ReleaseDate *time.Time `json:"release_date"`
    ProjectID   uint64     `gorm:"not null;index" json:"project_id"`
}

type ReleaseIssue struct {
    ReleaseID uint64 `gorm:"primaryKey;autoIncrement:false"`
    IssueID   uint64 `gorm:"primaryKey;autoIncrement:false"`
}
```

### API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/projects/:projectId/releases` | 列表/创建 |
| GET/PUT/DELETE | `/projects/:projectId/releases/:id` | 单个 CRUD |
| POST/DELETE | `/projects/:projectId/releases/:id/issues` | 关联/取消关联 Issue |
| GET | `/projects/:projectId/releases/:id/progress` | 进度统计 |

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `model/release.go` |
| 新建 | `service/release_service.go` |
| 新建 | `handler/release_handler.go` |
| 新建 | `frontend/src/components/ReleaseList.vue` |
| 新建 | `frontend/src/components/ReleaseDetailPanel.vue` |
| 新建 | `frontend/src/types/release.ts` |
| 新建 | `frontend/src/api/release.ts` |
| 修改 | `router/router.go`, `cmd/server/main.go` |
| 修改 | `frontend/src/components/CustomFieldValueInput.vue` 添加 Release picker |
| 修改 | `frontend/src/views/Project.vue` 添加 Releases tab |

---

## Task 7: Cover Image 封面图
**功能**: Issue 可设置封面图，在工作项详情和卡片中展示。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue.go` 新增 CoverImageURL 字段 |
| 修改 | `frontend/src/components/IssueCard.vue` 展示封面 |
| 修改 | `frontend/src/views/IssueDetail.vue` 添加上传/编辑封面 |

---

## Task 8: Bulk Copy/Move + Convert Type

**功能**: 批量跨项目复制/移动工作项。单个工作项可转换 Issue Type。

### API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/issues/bulk/copy` | 批量复制到目标项目 |
| POST | `/issues/bulk/move` | 批量移动到目标项目 |
| POST | `/issues/:issueId/convert-type` | 转换类型 |

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `handler/issue_handler.go` 添加 3 个新方法 |
| 修改 | `service/issue_service.go` |
| 修改 | `router/router.go` |

---

## Task 9: Merge Duplicates 合并重复工作项
**功能**: 将两个 Issue 合并为一个，保留 A 的数据，移动 B 的关联（assignees/labels/comments/relations）到 A，然后删除 B。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `handler/issue_handler.go` 添加 Merge 方法 |
| 修改 | `service/issue_service.go` |
| 修改 | `frontend/src/views/IssueDetail.vue` 添加 "Merge" 按钮 |

---

## Task 10: Mandatory Field Validation 必填校验

**功能**: Custom Field 标记 `is_required` 后，在 Issue 创建/更新时强制验证。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `service/issue_service.go` 在 Create/Update 中检查必填字段 |
| 修改 | `service/custom_field_service.go` |
| 修改 | `frontend/src/components/CustomFieldValueInput.vue` 添加必填标记 + 提示 |
| 修改 | `frontend/src/views/IssueCreate.vue` 添加提交前校验 |

---

## Task 11: Conditional Fields 条件显隐

**功能**: 根据其他字段值动态显示/隐藏 Custom Field。如"是否需要审批 = Yes"时才显示"审批人"字段。

### 数据模型

在 `custom_fields` 表新增字段：

```go
VisibilityRules json.RawMessage `gorm:"type:jsonb" json:"visibility_rules"`
// [{"field_id":5,"operator":"equals","value":"Yes"}]
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/custom_field.go` |
| 修改 | `frontend/src/components/CustomFieldValueInput.vue` 添加条件显隐逻辑 |
| 修改 | `frontend/src/components/CustomFieldForm.vue` 添加配置 UI |

---

## Task 12: Type Hierarchy Rules 类型嵌套规则强制

**功能**: 在工作空间级 Type Hierarchy 中定义"Epic 只能包含 Story/Feature"，API 层强制校验。

### 数据模型

在 `issue_types` 表新增：

```go
AllowedChildTypeIDs json.RawMessage `gorm:"type:jsonb" json:"allowed_child_type_ids"`
// [2, 3, 5] 允许作为子类型的 Type ID 列表
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue_type.go` 添加 AllowedChildTypeIDs |
| 修改 | `service/issue_service.go` 在 Create/Update 时校验 parent 的 Type |
| 修改 | `handler/issue_type_handler.go` 在 Update 时设置规则 |
| 修改 | `frontend/src/components/WorkspaceIssueTypeManager.vue` 添加 Hierarchy tab UI |

---

## Task 13: Time Tracking 工时记录

**功能**: 在工作项上记录工时（开始/暂停/停止），支持汇总统计。

### 数据模型

```go
type TimeTrack struct {
    BaseModel
    IssueID     uint64     `gorm:"not null;index" json:"issue_id"`
    UserID      uint64     `gorm:"not null" json:"user_id"`
    Description *string    `gorm:"size:500" json:"description"`
    StartedAt   time.Time  `gorm:"not null" json:"started_at"`
    EndedAt     *time.Time `json:"ended_at"`
    Duration    *int64     `json:"duration"` // seconds
}
```

### API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/issues/:issueId/time-tracks` | 开始计时 |
| PATCH | `/issues/:issueId/time-tracks/:id/stop` | 停止计时 |
| GET | `/issues/:issueId/time-tracks` | 列表 |
| DELETE | `/issues/:issueId/time-tracks/:id` | 删除记录 |

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `model/time_track.go` |
| 新建 | `service/time_track_service.go` |
| 新建 | `handler/time_track_handler.go` |
| 新建 | `frontend/src/components/TimeTrackPanel.vue` |
| 新建 | `frontend/src/types/time-track.ts` |
| 新建 | `frontend/src/api/time-track.ts` |
| 修改 | `router/router.go`, `cmd/server/main.go` |
| 修改 | `frontend/src/views/IssueDetail.vue` 集成面板 |

---

## Task 14: Recurring Work Items 重复工作项
**功能**: 设置工作项按周期自动创建副本（每天/每周/每月/Cron）。

### 数据模型

```go
type RecurrenceRule struct {
    BaseModel
    IssueID     uint64     `gorm:"not null;uniqueIndex" json:"issue_id"`
    Frequency   string     `gorm:"size:20;not null" json:"frequency"` // daily/weekly/monthly/cron
    Interval    int        `gorm:"default:1" json:"interval"`          // 每 N 个周期
    CronExpr    *string    `gorm:"size:100" json:"cron_expr"`
    NextRun     time.Time  `gorm:"not null" json:"next_run"`
    EndDate     *time.Time `json:"end_date"`
    IsActive    bool       `gorm:"default:true" json:"is_active"`
}
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `model/recurrence_rule.go` |
| 新建 | `service/recurrence_service.go` |
| 新建 | `handler/recurrence_handler.go` |
| 新建 | `frontend/src/components/RecurrenceConfig.vue` |
| 新建 | `frontend/src/types/recurrence.ts` |
| 新建 | `frontend/src/api/recurrence.ts` |
| 修改 | `router/router.go`, `cmd/server/main.go` |

---

## Task 15: Estimates Categories + Time 模式

**功能**: 扩展现有 EstimatePoint 为三种模式：Points / Categories / Time。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `frontend/src/components/EstimatePointManager.vue` 添加模式切换 |
| 修改 | `frontend/src/types/estimate-point.ts` 新增模式类型 |
| 修改 | `frontend/src/api/estimate-point.ts` |

---

## Task 16: Quick Filters 预设筛选芯片
**功能**: 列表顶部一行可点击的筛选芯片（"我的"/"未分配"/"高优先级"/"今日创建"），点击即应用。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/QuickFilterChips.vue` |
| 修改 | `frontend/src/views/Project.vue` 集成到列表上方 |

---

## Task 17: Attachment Backend 附件后端

**功能**: 补齐后端 Attachment model/service/handler，接上已有的前端 AttachmentManager。

### 数据模型

```go
type Attachment struct {
    BaseModel
    Name     string `gorm:"size:255;not null" json:"name"`
    FilePath string `gorm:"size:500;not null" json:"file_path"`
    FileSize int64  `json:"file_size"`
    MimeType string `gorm:"size:100" json:"mime_type"`
    IssueID  uint64 `gorm:"not null;index" json:"issue_id"`
    UploaderID *uint64 `json:"uploader_id"`
}
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `model/attachment.go` |
| 新建 | `service/attachment_service.go` |
| 新建 | `handler/attachment_handler.go` |
| 修改 | `router/router.go`, `cmd/server/main.go` |
| 修改 | `frontend/src/components/AttachmentManager.vue` 接上真实 API |

---

## Task 18: Notification Auto-trigger 自动通知触发

**功能**: 在关键业务事件中自动创建通知。

### 触发器

| 事件 | 触发位置 | 接收人 |
|------|---------|--------|
| Issue 被分配 | `issue_service.go` AddAssignee | 被分配人 |
| State 变更 | `issue_service.go` Update | 关注者 |
| Comment 添加 | `comment_service.go` Create | Issue 参与者 |
| Cycle 开始 | `cycle_service.go` Start | 项目成员 |

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `service/issue_service.go` 调用 notificationSvc.Create |
| 修改 | `service/comment_service.go` |
| 修改 | `service/cycle_service.go` |

---

## Task 19: Intake & Triage 接收与分诊
**功能**: 生成外部提交链接（Intake Form），非项目成员可提交工作项。管理员在 Triage 视图中审核、分配/拒绝。

### 数据模型

```go
// Issue 新增字段
IntakeSource  *string `gorm:"size:50" json:"intake_source"`  // "form" | "email" | "api"
IntakeStatus  *string `gorm:"size:30" json:"intake_status"`  // "pending" | "accepted" | "rejected"
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue.go` 新增 Intake 字段 |
| 新建 | `handler/intake_handler.go` 添加公开提交端点 |
| 新建 | `frontend/src/views/IntakeForm.vue` 添加公开表单 |
| 新建 | `frontend/src/components/TriagePanel.vue` 添加分诊面板 |
| 新建 | `frontend/src/types/intake.ts` |
| 新建 | `frontend/src/api/intake.ts` |
| 修改 | `router/router.go` 添加公开端点（无认证） |
| 修改 | `frontend/src/views/Project.vue` 添加 Triage tab |

---

## Part B: AI 智能助手功能规划

> **reqmango 现状**: 已有 AI 功能。包含 LLM 集成、AI 模型、聊天界面。

### AI 功能矩阵

| AI 能力 | 描述 | reqmango 现状 | 难度 |
|---------|------|--------------|------|
| **AI Chat 对话** | 聊天界面，自然语言交互项目数据 | ✅ | 低 |
| **NL Search 自然语言搜索** | "上周未解决的紧急Bug" → RQL | ✅ | 中 |
| **Smart Create 智能创建** | "创建一个登录页面的Bug，P0" → Issue | ✅ | 中 |
| **Smart Update 智能更新** | "把 #42 标记为完成" → 执行操作 | ✅ | 中 |
| **Data Analysis 数据分析** | "分析本周项目进度" → 洞察报告 | ✅ | 高 |
| **Context Awareness 上下文感知** | AI 知道当前项目/页面/选中项 | ✅ | 中 |
| **Page AI 文档AI** | 在 Page 编辑器中 AI 生成/总结/翻译 | ✅ | 中 |
| **AI Triage 智能分诊** | AI 自动分类/优先级建议新提交的 Issue | ✅ | 中 |
| **AI Sprint Planning** | AI 根据历史数据建议 Sprint 容量 | ✅ | 高 |
| **Command Palette** | ⌘K 快速导航/搜索/操作 | ✅ | 低 |

---

### AI 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                   前端 (Vue 3)                              │
│  ┌───────────┐  ┌───────────┐  ┌───────────────────────┐  │
│  │ AI Chat   │  │ ⌘K Cmd   │  │ Page AI (Editor)      │  │
│  │ Sidebar   │  │ Palette   │  │ 生成/总结/翻译        │  │
│  └────┬──────┘  └────┬──────┘  └────────┬────────────────┘  │
│       │              │                   │                  │
│       └──────────────┼───────────────────┘                  │
│                      │                                     │
│              SSE (Server-Sent Events)                      │
└──────────────────────┼─────────────────────────────────────┘
                       │
┌──────────────────────┼─────────────────────────────────────┐
│               Go 后端 (Gin)                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │          AI Service (ai_service.go)                  │   │
│  │  ┌─────────┐┌──────────┐┌──────────────┐           │   │
│  │  │ Intent  ││ Context  ││ Tool Calling │           │   │
│  │  │ Parser  ││ Builder  ││(Function)    │           │   │
│  │  └─────────┘└──────────┘└──────────────┘           │   │
│  │                                                     │   │
│  │  Tools: search_issues, create_issue,               │   │
│  │         update_issue, get_project_stats,           │   │
│  │         get_cycle_progress, list_assignees...       │   │
│  └─────────────────────────────────────────────────────┘   │
│                      │                                     │
│              LLM API Call                                  │
└──────────────────────┼─────────────────────────────────────┘
                       │
              ┌────────┴────────┐
              │  LLM API        │
              │  (Claude/OpenAI)│
              │  + Tool Use     │
              └─────────────────┘
```

### AI Service 核心设计

```go
// backend/internal/service/ai_service.go

type AIService struct {
    db       *gorm.DB
    llmClient *LLMClient  // Claude/OpenAI SDK wrapper
    issueSvc  *IssueService
    projectSvc *ProjectService
    // ... 其他 service 引用
}

// Chat handles a conversational AI request.
func (s *AIService) Chat(ctx context.Context, req *AIChatRequest) (*AIChatResponse, error) {
    // 1. Build system prompt with project/page/issue context
    // 2. Send message + tool definitions to LLM
    // 3. If LLM requests tool calls → execute → send results back
    // 4. Stream response via SSE channel
}

// Available AI Tools (Function Calling):
// - search_issues(query, filters) → []Issue
// - create_issue(project_id, type, title, description, priority, ...) → Issue
// - update_issue(issue_id, fields) → Issue
// - get_project_stats(project_id) → ProjectStatistics
// - get_cycle_progress(cycle_id) → CycleProgress
// - get_issue_detail(issue_id) → IssueDetail
// - list_assignees(project_id) → []User
// - add_comment(issue_id, body) → Comment
```

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/projects/:projectId/ai/chat` | AI 对话（SSE 流式） |
| POST | `/projects/:projectId/ai/search` | NL 搜索 → 返回 Issue 列表 |
| POST | `/projects/:projectId/ai/create` | NL 创建 → 返回预览 + 确认 |
| POST | `/projects/:projectId/ai/analyze` | 数据分析 → 返回洞察 |
| POST | `/pages/:pageId/ai/generate` | Page AI 生成内容 |
| POST | `/pages/:pageId/ai/summarize` | Page AI 总结 |
| GET | `/ai/models` | 列出可用 AI 模型 |

### SSE 流式响应格式

```
data: {"type":"thinking","content":"正在搜索最近一周的Bug..."}
data: {"type":"tool_call","name":"search_issues","args":{"priority":"urgent","days":7}}
data: {"type":"tool_result","result":[{"id":42,"name":"登录失败","priority":"urgent"}]}
data: {"type":"text","content":"找到 3 个紧急Bug：#42 登录失败，#58 支付超时，#99 数据丢失"}
data: {"type":"done"}
```

### 上下文注入

每个 AI 请求自动注入当前上下文：

```go
type AIContext struct {
    Workspace   *WorkspaceLite   // 当前工作空间
    Project     *ProjectLite     // 当前项目
    Page        *PageLite        // 当前页面（如果在 Pages 中）
    Issue       *IssueLite       // 当前工作项（如在 IssueDetail 中）
    User        *UserLite        // 当前用户
    RecentItems []RecentItem     // 最近访问/操作
}
```

---

## 更新后的实施节奏

```
Phase 1 — 工作项基础能力 (Week 1-3):
  Task 1-3:    Templates + Quick Create + Import
  Task 15,10,12: Estimates + Validation + Hierarchy

Phase 2 — 工作项关联与操作 (Week 4-6):
  Task 6,5,8:  Release + Page Link + Bulk Ops
  Task 17,18,4: Attachment + Notification + Sub-items

Phase 3 — 工作项展示与深度 (Week 7-9):
  Task 16,7,13: Quick Filters + Cover + Time Track
  Task 14,9,11,19: Recurring + Merge + Conditional + Intake
```

---

## 实现规范

### 代码模式速查

| 层 | 参考文件 | 关键模式 |
|---|---------|---------|
| Go Model | `model/notification.go` | 嵌入 `BaseModel`, `uint64` ID, `TableName()`, GORM tag |
| Go Service | `service/notification_service.go` | `NewXxxService(db *gorm.DB)`, 返回 `*response.XxxResponse`, `*common.AppError` |
| Go Handler | `handler/notification_handler.go` | `c.Get("currentUser").(*model.User).ID`, `ShouldBindJSON`, `strconv.ParseUint(c.Param(...))` |
| Go Router | `router/router.go` | 项目级嵌套在 `projects.Group` 内，挂载 `authMiddleware` |
| Go SSE | `handler/ai_handler.go` (Chat) | `c.Stream()` + `c.SSEvent()`, `Content-Type: text/event-stream` |
| LLM Client | `service/llm_client.go` | 封装 Anthropic SDK / OpenAI SDK, Tool Calling, Streaming |
| Vue View | `views/ProjectSettings.vue` | `<script setup lang="ts">`, sidebar + tabs 布局 |
| Vue Component | `components/SavedViewSelector.vue` | `defineProps<T>()`, `defineEmits<T>()`, `onMounted(() => load())` |
| Vue API | `api/page.ts` | 函数导出, `api.get/post/put/delete`, `Promise<T>` |
| Vue Types | `types/page.ts` | `interface` 定义，字段名与 Go DTO 一致 |
| Vue Composable | `composables/useAI.ts` | SSE EventSource, `ref()`, `onMounted`/`onUnmounted` |

### AI 特有约定

1. **SSE 流式**: Chat 端点使用 Server-Sent Events，`text/event-stream` MIME type
2. **API Key 安全**: AI Config 的 `api_key` 字段标记 `json:"-"` 永远不序列化
3. **Tool Calling**: 使用 Claude/OpenAI 原生 Function Calling，工具定义在 `ai_service.go` 的 `getTools()` 方法
4. **上下文注入**: 每个 AI 请求自动附带当前 workspace/project/page/issue 上下文
5. **速率限制**: AI 端点建议添加 per-user rate limiting（后续 middleware 实现）
6. **日志审计**: AI 对话历史存储在 `ai_threads` + `ai_messages` 表，用于审计和改进