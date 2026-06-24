# ReqManPy 对标 Plane — 工作项管理差距补全计划

> **基线**: ReqManPy master 分支 (2026-06-25)
> **对标**: [Plane Enterprise 信息架构](docs/kb/architecture/plane-enterprise-info-architecture.md) + Plane 工作项文档
> **范围**: 聚焦工作项（Issue）管理全生命周期，不含平台级功能（Integrations/Plane Runner 等）
> **拆分原则**: 每个 Task 独立可验证、无跨 Task 代码冲突、有明确的完成标准

---

## 差距总览：工作项管理 14 个维度

```
Plane 工作项管理全貌                          ReqManPy 现状

📊 视图层
  ✅ List View     — 列表视图                   ✅ 已实现
  ✅ Kanban View   — 看板视图                   ✅ 已实现
  ❌ Calendar View — 日历视图                   ❌ 无
  ❌ Gantt View    — 甘特图                     ❌ 无
  ❌ Timeline      — 时间线/路线图               ❌ 无

📝 创建层
  ✅ Manual Create                              ✅ 已实现
  ❌ Work Item Templates — 工作项模板            ← Task 1
  ❌ Quick Create     — 行内快速创建             ← Task 2
  ❌ Bulk Import      — CSV/JSON 批量导入        ← Task 3

📋 详情层
  ✅ Rich Text Description                      ✅ TipTap
  ✅ Comments (threaded)                        ✅ 已实现
  ✅ Activity Log                               ✅ 已实现
  ❌ Sub-items Panel  — 子工作项可展开面板        ← Task 4
  ❌ Page Linking     — 关联文档页面              ← Task 5
  ❌ Release Linking  — 关联 Release             ← Task 6
  ❌ Cover Image      — 封面图                   ← Task 7

🔧 操作层
  ✅ Bulk Update / Delete                       ✅ 已实现
  ❌ Bulk Copy/Move   — 跨项目复制/移动           ← Task 8
  ❌ Convert Type     — 转换工作项类型            ← Task 8
  ❌ Merge Duplicates — 合并重复工作项            ← Task 9

🎯 属性层 (Custom Fields)
  ✅ 7 种字段类型 (text/number/dropdown/boolean/date/member/url)
  ❌ Release Picker  — Release 类型字段          ← Task 6
  ❌ Mandatory Validation — 必填校验             ← Task 10
  ❌ Conditional Fields   — 条件显隐规则         ← Task 11

🏗️ 层级层
  ✅ Parent/Child with depth                    ✅ 已实现
  ❌ Type Hierarchy Rules — 类型嵌套规则强制      ← Task 12

⏱️ 时间层
  ❌ Time Tracking   — 工时记录                  ← Task 13
  ✅ Start Date / Target Date                   ✅ 已实现
  ❌ Recurring Work Items — 重复工作项           ← Task 14

📊 估算层
  ✅ Points (Fibonacci)                         ✅ 已实现
  ❌ Categories (T-shirt sizes)                 ← Task 15
  ❌ Time Estimates                             ← Task 15

🔍 筛选层
  ✅ RQL Query Language                         ✅ 已实现
  ✅ Saved Views                                ✅ 刚实现
  ❌ Quick Filters — 预设筛选芯片               ← Task 16

📎 附件层
  ❌ 后端 Attachment model/service              ← Task 17
  ✅ 前端 AttachmentManager (仅UI无后端)

🔔 通知层
  ✅ Notification model/API                     ✅ 刚实现
  ❌ Auto-trigger — 自动触发通知                 ← Task 18

📥 接收层
  ❌ Intake Form — 外部提交表单                  ← Task 19
  ❌ Triage Mode — 收件分诊模式                  ← Task 19
```

---

## Task 清单总览 (工作项管理专项)

| # | Task | 优先级 | 类型 | 文件数 | 依赖 |
|---|------|--------|------|--------|------|
| 1 | Work Item Templates 工作项模板 | **P0** | 新建 | 9 | 无 |
| 2 | Quick Create 行内快速创建 | **P0** | 前端 | 3 | 无 |
| 3 | Bulk Import CSV/JSON | **P0** | 新建 | 5 | 无 |
| 4 | Sub-items Panel 子工作项面板 | P1 | 前端 | 2 | 无 |
| 5 | Page Linking 关联文档页面 | P1 | 修改 | 3 | 已有 Pages |
| 6 | Release Management + Release Picker | P1 | 新建 | 9 | 无 |
| 7 | Cover Image 封面图 | P2 | 修改 | 2 | 无 |
| 8 | Bulk Copy/Move + Convert Type | P1 | 修改 | 3 | 无 |
| 9 | Merge Duplicates 合并重复工作项 | P2 | 修改 | 3 | 无 |
| 10 | Mandatory Field Validation 必填校验 | P1 | 修改 | 4 | 无 |
| 11 | Conditional Fields 条件显隐 | P2 | 修改 | 3 | Task 10 |
| 12 | Type Hierarchy Rules 类型嵌套规则 | P1 | 修改 | 4 | 无 |
| 13 | Time Tracking 工时记录 | P2 | 新建 | 7 | 无 |
| 14 | Recurring Work Items 重复工作项 | P2 | 新建 | 7 | 无 |
| 15 | Estimates Categories + Time | P1 | 修改 | 3 | 无 |
| 16 | Quick Filters 预设筛选芯片 | P1 | 前端 | 2 | 无 |
| 17 | Attachment Backend 附件后端 | P1 | 新建 | 5 | 无 |
| 18 | Notification Auto-trigger 自动通知 | P1 | 修改 | 3 | 已有 Notif |
| 19 | Intake & Triage 接收与分诊 | P3 | 新建 | 8 | 无 |

---

## Task 1: Work Item Templates 工作项模板

**功能**: 项目管理员为每种 Issue Type 预设创建模板（默认 title、description、assignee、labels、priority 等）。创建 Issue 时选择模板一键预填。

### 数据模型

```go
// backend-go/internal/model/work_item_template.go

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
| 修改 | `frontend/src/views/ProjectSettings.vue` — "Templates" tab |
| 修改 | `frontend/src/views/IssueCreate.vue` — 模板选择下拉 |

---

## Task 2: Quick Create 行内快速创建

**功能**: 在列表/看板顶部提供一行极简输入框，输入标题 + 回车即可创建，无需跳转完整表单。

### 前端改动

- 修改 `IssueList.vue` — 列表顶部新增行内输入框（title + type + priority 下拉 + 回车创建）
- 修改 `IssueKanban.vue` — 每列顶部新增 "+" 按钮 → 行内输入
- 新建 `components/QuickCreateInput.vue` — 可复用行内创建组件

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
| 修改 | `handler/issue_handler.go` — 4 个方法 |
| 修改 | `router/router.go` |
| 修改 | `frontend/src/components/IssueList.vue` — 导入/导出按钮 |

---

## Task 4: Sub-items Panel 子工作项可展开面板

**功能**: IssueDetail 中展示子工作项列表，支持行内展开/折叠、拖拽排序。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/SubIssuesPanel.vue` |
| 修改 | `frontend/src/views/IssueDetail.vue` — 集成面板 |

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
| 修改 | `model/issue.go` — 新增 IssuePage 关联 |
| 新建 | `handler/issue_handler.go` — AddPage/RemovePage/ListPages 方法 |
| 修改 | `frontend/src/views/IssueDetail.vue` — Pages 关联区 |

---

## Task 6: Release Management + Release Picker

**功能**: 第一方 Release/Version 管理。Custom Field 新增 "Release" 类型。

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
| 修改 | `frontend/src/components/CustomFieldValueInput.vue` — Release picker |
| 修改 | `frontend/src/views/Project.vue` — Releases tab |

---

## Task 7: Cover Image 封面图

**功能**: Issue 可设置封面图，在工作项详情和卡片中展示。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue.go` — 新增 CoverImageURL 字段 |
| 修改 | `frontend/src/components/IssueCard.vue` — 展示封面 |
| 修改 | `frontend/src/views/IssueDetail.vue` — 上传/编辑封面 |

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
| 修改 | `handler/issue_handler.go` — 3 个新方法 |
| 修改 | `service/issue_service.go` |
| 修改 | `router/router.go` |

---

## Task 9: Merge Duplicates 合并重复工作项

**功能**: 将两个 Issue 合并为一个，保留 A 的数据，移动 B 的关联（assignees/labels/comments/relations）到 A，然后删除 B。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `handler/issue_handler.go` — Merge 方法 |
| 修改 | `service/issue_service.go` |
| 修改 | `frontend/src/views/IssueDetail.vue` — "Merge" 按钮 |

---

## Task 10: Mandatory Field Validation 必填校验

**功能**: Custom Field 标记 `is_required` 后，在 Issue 创建/更新时强制验证。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `service/issue_service.go` — Create/Update 中检查必填字段 |
| 修改 | `service/custom_field_service.go` |
| 修改 | `frontend/src/components/CustomFieldValueInput.vue` — 必填标记 + 提示 |
| 修改 | `frontend/src/views/IssueCreate.vue` — 提交前校验 |

---

## Task 11: Conditional Fields 条件显隐

**功能**: 根据其他字段值动态显示/隐藏 Custom Field。如"是否需要审批 = Yes"时才显示"审批人"字段。

### 数据模型

在 `custom_fields` 表新增字段:
```go
VisibilityRules json.RawMessage `gorm:"type:jsonb" json:"visibility_rules"`
// [{"field_id":5,"operator":"equals","value":"Yes"}]
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/custom_field.go` |
| 修改 | `frontend/src/components/CustomFieldValueInput.vue` — 条件显隐逻辑 |
| 修改 | `frontend/src/components/CustomFieldForm.vue` — 配置 UI |

---

## Task 12: Type Hierarchy Rules 类型嵌套规则强制

**功能**: 在工作空间级 Type Hierarchy 中定义"Epic 只能包含 Story/Feature"，API 层强制校验。

### 数据模型

在 `issue_types` 表新增:
```go
AllowedChildTypeIDs json.RawMessage `gorm:"type:jsonb" json:"allowed_child_type_ids"`
// [2, 3, 5] — 允许作为子类型的 Type ID 列表
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue_type.go` — AllowedChildTypeIDs |
| 修改 | `service/issue_service.go` — Create/Update 时校验 parent 的 Type |
| 修改 | `handler/issue_type_handler.go` — Update 时设置规则 |
| 修改 | `frontend/src/components/WorkspaceIssueTypeManager.vue` — Hierarchy tab UI |

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
| 修改 | `frontend/src/views/IssueDetail.vue` — 集成面板 |

---

## Task 14: Recurring Work Items 重复工作项

**功能**: 设置工作项按周期自动创建副本（每日/每周/每月/Cron）。

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
| 修改 | `frontend/src/components/EstimatePointManager.vue` — 模式切换 |
| 修改 | `frontend/src/types/estimate-point.ts` — 新增模式类型 |
| 修改 | `frontend/src/api/estimate-point.ts` |

---

## Task 16: Quick Filters 预设筛选芯片

**功能**: 列表顶部一行可点击的筛选芯片（"我的"、"未分配"、"高优先级"、"今日创建"），点击即应用。

### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/QuickFilterChips.vue` |
| 修改 | `frontend/src/views/Project.vue` — 集成到列表上方 |

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
| 修改 | `frontend/src/components/AttachmentManager.vue` — 接上真实 API |

---

## Task 18: Notification Auto-trigger 自动通知触发

**功能**: 在关键业务事件中自动创建通知。

### 触发点

| 事件 | 触发位置 | 接收者 |
|------|---------|--------|
| Issue 被分配 | `issue_service.go` AddAssignee | 被分配人 |
| State 变更 | `issue_service.go` Update | 关注者 |
| Comment 添加 | `comment_service.go` Create | Issue 参与者 |
| Cycle 开始 | `cycle_service.go` Start | 项目成员 |

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `service/issue_service.go` — 调用 notificationSvc.Create |
| 修改 | `service/comment_service.go` |
| 修改 | `service/cycle_service.go` |

---

## Task 19: Intake & Triage 接收与分诊

**功能**: 生成外部提交链接（Intake Form），非项目成员可提交工作项。管理员在 Triage 视图中审批/分配/拒绝。

### 数据模型

```go
// Issue 新增字段
IntakeSource  *string `gorm:"size:50" json:"intake_source"`  // "form" | "email" | "api"
IntakeStatus  *string `gorm:"size:30" json:"intake_status"`  // "pending" | "accepted" | "rejected"
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue.go` — 新增 Intake 字段 |
| 新建 | `handler/intake_handler.go` — 公开提交端点 |
| 新建 | `frontend/src/views/IntakeForm.vue` — 公开表单 |
| 新建 | `frontend/src/components/TriagePanel.vue` — 分诊面板 |
| 新建 | `frontend/src/types/intake.ts` |
| 新建 | `frontend/src/api/intake.ts` |
| 修改 | `router/router.go` — 公开端点（无认证） |
| 修改 | `frontend/src/views/Project.vue` — Triage tab |

---

## 实施节奏建议

```
Week 1 (P0 — 核心创建工作流):
  Task 1:  Work Item Templates      → 创建效率翻倍
  Task 2:  Quick Create             → 极简创建
  Task 3:  Bulk Import CSV/JSON     → 数据迁移

Week 2 (P1 — 结构化增强):
  Task 15: Estimates Categories     → 估算多样性
  Task 10: Mandatory Validation     → 数据完整性
  Task 12: Type Hierarchy Rules     → 层级约束

Week 3 (P1 — 关联与操作):
  Task 6:  Release Management       → 发布追踪
  Task 5:  Page Linking             → 文档关联
  Task 8:  Bulk Copy/Move/Convert   → 灵活操作

Week 4 (P1 — 附件与通知):
  Task 17: Attachment Backend       → 文件上传
  Task 18: Notification Auto-trigger→ 自动通知
  Task 4:  Sub-items Panel          → 层级展示

Week 5 (P1 — 筛选与展示):
  Task 16: Quick Filters            → 快速筛选
  Task 7:  Cover Image              → 视觉增强

Week 6 (P2 — 深度功能):
  Task 13: Time Tracking            → 工时管理
  Task 14: Recurring Work Items     → 自动化创建
  Task 9:  Merge Duplicates         → 去重

Week 7 (P2-P3):
  Task 11: Conditional Fields       → 智能表单
  Task 19: Intake & Triage          → 外部接收
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
| Vue View | `views/ProjectSettings.vue` | `<script setup lang="ts">`, sidebar + tabs 布局 |
| Vue Component | `components/SavedViewSelector.vue` | `defineProps<T>()`, `defineEmits<T>()`, `onMounted(() => load())` |
| Vue API | `api/page.ts` | 函数导出, `api.get/post/put/delete`, `Promise<T>` |
| Vue Types | `types/page.ts` | `interface` 定义，字段名与 Go DTO 一致 |

### 关键约定

1. **getUserID**: 必须 `c.Get("currentUser").(*model.User).ID`（不是 `c.Get("user_id")`）
2. **AuthMiddleware**: 所有新路由挂载 `authMiddleware`（Intake 除外）
3. **AutoMigrate**: 新 Model 在 `cmd/server/main.go` 中注册
4. **路由**: workspace 级 `/workspaces/:wsParam/xxx`，project 级 `/projects/:projectId/xxx`
5. **前端 API base**: axios 实例 baseURL 已设 `/api/v1`
6. **JSONB 字段**: 使用 `json.RawMessage`，Go 端 `normalizeJSON()` 处理空值
7. **软删除**: 所有 Model 使用 GORM `DeletedAt`，不是物理删除
