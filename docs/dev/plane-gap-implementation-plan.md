# reqmango 对标 Plane �?工作项管理差距补全计�?
> **基线**: reqmango master 分支 (2026-06-25)
> **对标**: [Plane Enterprise 信息架构](docs/kb/architecture/plane-enterprise-info-architecture.md) + Plane 工作项文�?> **范围**: 聚焦工作项（Issue）管理全生命周期，不含平台级功能（Integrations/Plane Runner 等）
> **拆分原则**: 每个 Task 独立可验证、无�?Task 代码冲突、有明确的完成标�?
---

## 差距总览：工作项管理 14 个维�?
```
Plane 工作项管理全�?                         reqmango 现状

📊 视图�?  �?List View     �?列表视图                   �?已实�?  �?Kanban View   �?看板视图                   �?已实�?  �?Calendar View �?日历视图                   �?�?  �?Gantt View    �?甘特�?                    �?�?  �?Timeline      �?时间�?路线�?              �?�?
📝 创建�?  �?Manual Create                              �?已实�?  �?Work Item Templates �?工作项模�?           �?Task 1
  �?Quick Create     �?行内快速创�?            �?Task 2
  �?Bulk Import      �?CSV/JSON 批量导入        �?Task 3

📋 详情�?  �?Rich Text Description                      �?TipTap
  �?Comments (threaded)                        �?已实�?  �?Activity Log                               �?已实�?  �?Sub-items Panel  �?子工作项可展开面板        �?Task 4
  �?Page Linking     �?关联文档页面              �?Task 5
  �?Release Linking  �?关联 Release             �?Task 6
  �?Cover Image      �?封面�?                  �?Task 7

🔧 操作�?  �?Bulk Update / Delete                       �?已实�?  �?Bulk Copy/Move   �?跨项目复�?移动           �?Task 8
  �?Convert Type     �?转换工作项类�?           �?Task 8
  �?Merge Duplicates �?合并重复工作�?           �?Task 9

🎯 属性层 (Custom Fields)
  �?7 种字段类�?(text/number/dropdown/boolean/date/member/url)
  �?Release Picker  �?Release 类型字段          �?Task 6
  �?Mandatory Validation �?必填校验             �?Task 10
  �?Conditional Fields   �?条件显隐规则         �?Task 11

🏗�?层级�?  �?Parent/Child with depth                    �?已实�?  �?Type Hierarchy Rules �?类型嵌套规则强制      �?Task 12

⏱️ 时间�?  �?Time Tracking   �?工时记录                  �?Task 13
  �?Start Date / Target Date                   �?已实�?  �?Recurring Work Items �?重复工作�?          �?Task 14

📊 估算�?  �?Points (Fibonacci)                         �?已实�?  �?Categories (T-shirt sizes)                 �?Task 15
  �?Time Estimates                             �?Task 15

🔍 筛选层
  �?RQL Query Language                         �?已实�?  �?Saved Views                                �?刚实�?  �?Quick Filters �?预设筛选芯�?              �?Task 16

📎 附件�?  �?后端 Attachment model/service              �?Task 17
  �?前端 AttachmentManager (仅UI无后�?

🔔 通知�?  �?Notification model/API                     �?刚实�?  �?Auto-trigger �?自动触发通知                 �?Task 18

📥 接收�?  �?Intake Form �?外部提交表单                  �?Task 19
  �?Triage Mode �?收件分诊模式                  �?Task 19
```

---

## Task 清单总览 (工作项管理专�?

| # | Task | 优先�?| 类型 | 文件�?| 依赖 |
|---|------|--------|------|--------|------|
| 1 | Work Item Templates 工作项模�?| **P0** | 新建 | 9 | �?|
| 2 | Quick Create 行内快速创�?| **P0** | 前端 | 3 | �?|
| 3 | Bulk Import CSV/JSON | **P0** | 新建 | 5 | �?|
| 4 | Sub-items Panel 子工作项面板 | P1 | 前端 | 2 | �?|
| 5 | Page Linking 关联文档页面 | P1 | 修改 | 3 | 已有 Pages |
| 6 | Release Management + Release Picker | P1 | 新建 | 9 | �?|
| 7 | Cover Image 封面�?| P2 | 修改 | 2 | �?|
| 8 | Bulk Copy/Move + Convert Type | P1 | 修改 | 3 | �?|
| 9 | Merge Duplicates 合并重复工作�?| P2 | 修改 | 3 | �?|
| 10 | Mandatory Field Validation 必填校验 | P1 | 修改 | 4 | �?|
| 11 | Conditional Fields 条件显隐 | P2 | 修改 | 3 | Task 10 |
| 12 | Type Hierarchy Rules 类型嵌套规则 | P1 | 修改 | 4 | �?|
| 13 | Time Tracking 工时记录 | P2 | 新建 | 7 | �?|
| 14 | Recurring Work Items 重复工作�?| P2 | 新建 | 7 | �?|
| 15 | Estimates Categories + Time | P1 | 修改 | 3 | �?|
| 16 | Quick Filters 预设筛选芯�?| P1 | 前端 | 2 | �?|
| 17 | Attachment Backend 附件后端 | P1 | 新建 | 5 | �?|
| 18 | Notification Auto-trigger 自动通知 | P1 | 修改 | 3 | 已有 Notif |
| 19 | Intake & Triage 接收与分�?| P3 | 新建 | 8 | �?|

---

## Task 1: Work Item Templates 工作项模�?
**功能**: 项目管理员为每种 Issue Type 预设创建模板（默�?title、description、assignee、labels、priority 等）。创�?Issue 时选择模板一键预填�?
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
| 修改 | `frontend/src/views/ProjectSettings.vue` �?"Templates" tab |
| 修改 | `frontend/src/views/IssueCreate.vue` �?模板选择下拉 |

---

## Task 2: Quick Create 行内快速创�?
**功能**: 在列�?看板顶部提供一行极简输入框，输入标题 + 回车即可创建，无需跳转完整表单�?
### 前端改动

- 修改 `IssueList.vue` �?列表顶部新增行内输入框（title + type + priority 下拉 + 回车创建�?- 修改 `IssueKanban.vue` �?每列顶部新增 "+" 按钮 �?行内输入
- 新建 `components/QuickCreateInput.vue` �?可复用行内创建组�?
### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/QuickCreateInput.vue` |
| 修改 | `frontend/src/components/IssueList.vue` |
| 修改 | `frontend/src/components/IssueKanban.vue` |

---

## Task 3: Bulk Import CSV/JSON

**功能**: �?CSV/JSON 文件批量导入工作项，含字段映�?UI�?
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
| 修改 | `handler/issue_handler.go` �?4 个方�?|
| 修改 | `router/router.go` |
| 修改 | `frontend/src/components/IssueList.vue` �?导入/导出按钮 |

---

## Task 4: Sub-items Panel 子工作项可展开面板

**功能**: IssueDetail 中展示子工作项列表，支持行内展开/折叠、拖拽排序�?
### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/SubIssuesPanel.vue` |
| 修改 | `frontend/src/views/IssueDetail.vue` �?集成面板 |

---

## Task 5: Page Linking 关联文档页面

**功能**: 工作项可关联到项目内�?Pages 文档。Issue model 已有扩展能力�?
### 数据模型

新增 M:N 关联�?`issue_pages`:

```go
type IssuePage struct {
    IssueID uint64 `gorm:"primaryKey;autoIncrement:false"`
    PageID  uint64 `gorm:"primaryKey;autoIncrement:false"`
}
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue.go` �?新增 IssuePage 关联 |
| 新建 | `handler/issue_handler.go` �?AddPage/RemovePage/ListPages 方法 |
| 修改 | `frontend/src/views/IssueDetail.vue` �?Pages 关联�?|

---

## Task 6: Release Management + Release Picker

**功能**: 第一�?Release/Version 管理。Custom Field 新增 "Release" 类型�?
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
| 修改 | `frontend/src/components/CustomFieldValueInput.vue` �?Release picker |
| 修改 | `frontend/src/views/Project.vue` �?Releases tab |

---

## Task 7: Cover Image 封面�?
**功能**: Issue 可设置封面图，在工作项详情和卡片中展示�?
### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue.go` �?新增 CoverImageURL 字段 |
| 修改 | `frontend/src/components/IssueCard.vue` �?展示封面 |
| 修改 | `frontend/src/views/IssueDetail.vue` �?上传/编辑封面 |

---

## Task 8: Bulk Copy/Move + Convert Type

**功能**: 批量跨项目复�?移动工作项。单个工作项可转�?Issue Type�?
### API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/issues/bulk/copy` | 批量复制到目标项�?|
| POST | `/issues/bulk/move` | 批量移动到目标项�?|
| POST | `/issues/:issueId/convert-type` | 转换类型 |

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `handler/issue_handler.go` �?3 个新方法 |
| 修改 | `service/issue_service.go` |
| 修改 | `router/router.go` |

---

## Task 9: Merge Duplicates 合并重复工作�?
**功能**: 将两�?Issue 合并为一个，保留 A 的数据，移动 B 的关联（assignees/labels/comments/relations）到 A，然后删�?B�?
### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `handler/issue_handler.go` �?Merge 方法 |
| 修改 | `service/issue_service.go` |
| 修改 | `frontend/src/views/IssueDetail.vue` �?"Merge" 按钮 |

---

## Task 10: Mandatory Field Validation 必填校验

**功能**: Custom Field 标记 `is_required` 后，�?Issue 创建/更新时强制验证�?
### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `service/issue_service.go` �?Create/Update 中检查必填字�?|
| 修改 | `service/custom_field_service.go` |
| 修改 | `frontend/src/components/CustomFieldValueInput.vue` �?必填标记 + 提示 |
| 修改 | `frontend/src/views/IssueCreate.vue` �?提交前校�?|

---

## Task 11: Conditional Fields 条件显隐

**功能**: 根据其他字段值动态显�?隐藏 Custom Field。如"是否需要审�?= Yes"时才显示"审批�?字段�?
### 数据模型

�?`custom_fields` 表新增字�?
```go
VisibilityRules json.RawMessage `gorm:"type:jsonb" json:"visibility_rules"`
// [{"field_id":5,"operator":"equals","value":"Yes"}]
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/custom_field.go` |
| 修改 | `frontend/src/components/CustomFieldValueInput.vue` �?条件显隐逻辑 |
| 修改 | `frontend/src/components/CustomFieldForm.vue` �?配置 UI |

---

## Task 12: Type Hierarchy Rules 类型嵌套规则强制

**功能**: 在工作空间级 Type Hierarchy 中定�?Epic 只能包含 Story/Feature"，API 层强制校验�?
### 数据模型

�?`issue_types` 表新�?
```go
AllowedChildTypeIDs json.RawMessage `gorm:"type:jsonb" json:"allowed_child_type_ids"`
// [2, 3, 5] �?允许作为子类型的 Type ID 列表
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue_type.go` �?AllowedChildTypeIDs |
| 修改 | `service/issue_service.go` �?Create/Update 时校�?parent �?Type |
| 修改 | `handler/issue_type_handler.go` �?Update 时设置规�?|
| 修改 | `frontend/src/components/WorkspaceIssueTypeManager.vue` �?Hierarchy tab UI |

---

## Task 13: Time Tracking 工时记录

**功能**: 在工作项上记录工时（开�?暂停/停止），支持汇总统计�?
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
| POST | `/issues/:issueId/time-tracks` | 开始计�?|
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
| 修改 | `frontend/src/views/IssueDetail.vue` �?集成面板 |

---

## Task 14: Recurring Work Items 重复工作�?
**功能**: 设置工作项按周期自动创建副本（每�?每周/每月/Cron）�?
### 数据模型

```go
type RecurrenceRule struct {
    BaseModel
    IssueID     uint64     `gorm:"not null;uniqueIndex" json:"issue_id"`
    Frequency   string     `gorm:"size:20;not null" json:"frequency"` // daily/weekly/monthly/cron
    Interval    int        `gorm:"default:1" json:"interval"`          // �?N 个周�?    CronExpr    *string    `gorm:"size:100" json:"cron_expr"`
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

**功能**: 扩展现有 EstimatePoint 为三种模式：Points / Categories / Time�?
### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `frontend/src/components/EstimatePointManager.vue` �?模式切换 |
| 修改 | `frontend/src/types/estimate-point.ts` �?新增模式类型 |
| 修改 | `frontend/src/api/estimate-point.ts` |

---

## Task 16: Quick Filters 预设筛选芯�?
**功能**: 列表顶部一行可点击的筛选芯片（"我的"�?未分�?�?高优先级"�?今日创建"），点击即应用�?
### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/QuickFilterChips.vue` |
| 修改 | `frontend/src/views/Project.vue` �?集成到列表上�?|

---

## Task 17: Attachment Backend 附件后端

**功能**: 补齐后端 Attachment model/service/handler，接上已有的前端 AttachmentManager�?
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
| 修改 | `frontend/src/components/AttachmentManager.vue` �?接上真实 API |

---

## Task 18: Notification Auto-trigger 自动通知触发

**功能**: 在关键业务事件中自动创建通知�?
### 触发�?
| 事件 | 触发位置 | 接收�?|
|------|---------|--------|
| Issue 被分�?| `issue_service.go` AddAssignee | 被分配人 |
| State 变更 | `issue_service.go` Update | 关注�?|
| Comment 添加 | `comment_service.go` Create | Issue 参与�?|
| Cycle 开�?| `cycle_service.go` Start | 项目成员 |

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `service/issue_service.go` �?调用 notificationSvc.Create |
| 修改 | `service/comment_service.go` |
| 修改 | `service/cycle_service.go` |

---

## Task 19: Intake & Triage 接收与分�?
**功能**: 生成外部提交链接（Intake Form），非项目成员可提交工作项。管理员�?Triage 视图中审�?分配/拒绝�?
### 数据模型

```go
// Issue 新增字段
IntakeSource  *string `gorm:"size:50" json:"intake_source"`  // "form" | "email" | "api"
IntakeStatus  *string `gorm:"size:30" json:"intake_status"`  // "pending" | "accepted" | "rejected"
```

### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `model/issue.go` �?新增 Intake 字段 |
| 新建 | `handler/intake_handler.go` �?公开提交端点 |
| 新建 | `frontend/src/views/IntakeForm.vue` �?公开表单 |
| 新建 | `frontend/src/components/TriagePanel.vue` �?分诊面板 |
| 新建 | `frontend/src/types/intake.ts` |
| 新建 | `frontend/src/api/intake.ts` |
| 修改 | `router/router.go` �?公开端点（无认证�?|
| 修改 | `frontend/src/views/Project.vue` �?Triage tab |

---

## Part B: Plane AI 智能助手差距分析

> **reqmango 现状**: �?AI 功能。无 LLM 集成、无 AI 模型、无聊天界面�?> **Plane AI 现状**: 内置 AI 助手，支�?Ask/Build 双模式，上下文感知�?
### AI 功能对比矩阵

| Plane AI 能力 | 描述 | reqmango 现状 | 难度 |
|---------------|------|--------------|------|
| **AI Chat 对话** | 聊天界面，自然语言交互项目数据 | �?�?| �?|
| **NL Search 自然语言搜索** | "上周未解决的紧急Bug" �?RQL | �?�?| �?|
| **Smart Create 智能创建** | "创建一个登录页面的Bug，P0" �?Issue | �?�?| �?|
| **Smart Update 智能更新** | "�?42标记为完�? �?执行操作 | �?�?| �?|
| **Data Analysis 数据分析** | "分析本周项目进度" �?洞察报告 | �?�?| �?|
| **Context Awareness 上下文感�?* | AI 知道当前项目/页面/选中�?| �?�?| �?|
| **Page AI 文档AI** | �?Page 编辑器中 AI 生成/总结/翻译 | �?�?| �?|
| **AI Triage 智能分诊** | AI 自动分类/优先级建议新提交�?Issue | �?�?| �?|
| **AI Sprint Planning** | AI 根据历史数据建议 Sprint 容量 | �?�?| �?|
| **Command Palette** | ⌘K 快速导�?搜索/操作 | �?�?| �?|

---

### AI 架构设计

```
┌─────────────────────────────────────────────────────�?�?                   前端 (Vue 3)                       �?�? ┌──────────�? ┌──────────�? ┌──────────────────�? �?�? �?AI Chat  �? �?⌘K Cmd   �? �?Page AI (Editor) �? �?�? �?Sidebar  �? �?Palette  �? �?生成/总结/翻译    �? �?�? └────┬─────�? └────┬─────�? └────────┬─────────�? �?�?      �?            �?               �?             �?�?      └─────────────┼────────────────�?             �?�?                    �?SSE (Server-Sent Events)       �?└─────────────────────┼───────────────────────────────�?                      �?┌─────────────────────┼───────────────────────────────�?�?               Go 后端 (Gin)                          �?�? ┌──────────────────▼───────────────────────────�?  �?�? �?          AI Service (ai_service.go)          �?  �?�? �? ┌─────────�?┌──────────�?┌──────────────�? �?  �?�? �? �?Intent  �?�?Context  �?�?Tool Calling �? �?  �?�? �? �?Parser  �?�?Builder  �?�?(Function)   �? �?  �?�? �? └─────────�?└──────────�?└──────────────�? �?  �?�? �?                                             �?  �?�? �? Tools: search_issues, create_issue,         �?  �?�? �? update_issue, get_project_stats,            �?  �?�? �? get_cycle_progress, list_assignees...       �?  �?�? └──────────────────┬───────────────────────────�?  �?�?                    �?LLM API Call                    �?└─────────────────────┼───────────────────────────────�?                      �?              ┌───────▼───────�?              �? Claude API   �?              �? (�?OpenAI)   �?              �? + Tool Use   �?              └───────────────�?```

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
    // 3. If LLM requests tool calls �?execute �?send results back
    // 4. Stream response via SSE channel
}

// Available AI Tools (Function Calling):
// - search_issues(query, filters) �?[]Issue
// - create_issue(project_id, type, title, description, priority, ...) �?Issue
// - update_issue(issue_id, fields) �?Issue
// - get_project_stats(project_id) �?ProjectStatistics
// - get_cycle_progress(cycle_id) �?CycleProgress
// - get_issue_detail(issue_id) �?IssueDetail
// - list_assignees(project_id) �?[]User
// - add_comment(issue_id, body) �?Comment
```

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/projects/:projectId/ai/chat` | AI 对话（SSE 流式�?|
| POST | `/projects/:projectId/ai/search` | NL 搜索 �?返回 Issue 列表 |
| POST | `/projects/:projectId/ai/create` | NL 创建 �?返回预览 + 确认 |
| POST | `/projects/:projectId/ai/analyze` | 数据分析 �?返回洞察 |
| POST | `/pages/:pageId/ai/generate` | Page AI 生成内容 |
| POST | `/pages/:pageId/ai/summarize` | Page AI 总结 |
| GET | `/ai/models` | 列出可用 AI 模型 |

### SSE 流式响应格式

```
data: {"type":"thinking","content":"正在搜索最近一周的Bug..."}
data: {"type":"tool_call","name":"search_issues","args":{"priority":"urgent","days":7}}
data: {"type":"tool_result","result":[{"id":42,"name":"登录失败","priority":"urgent"}]}
data: {"type":"text","content":"找到 3 个紧急Bug�?42 登录失败�?58 支付超时�?99 数据丢失"}
data: {"type":"done"}
```

### 上下文注�?
每个 AI 请求自动注入当前上下文：

```go
type AIContext struct {
    Workspace   *WorkspaceLite   // 当前工作空间
    Project     *ProjectLite     // 当前项目
    Page        *PageLite        // 当前页面（如�?Pages 中）
    Issue       *IssueLite       // 当前工作项（如在 IssueDetail 中）
    User        *UserLite        // 当前用户
    RecentItems []RecentItem     // 最近访�?操作
}
```

---

## Task 20-28: AI 功能拆分

### Task 20: AI Infrastructure �?LLM Client + Config

**功能**: 建立 AI 基础设施。LLM 客户端封装、模型配置、API Key 管理�?
#### 数据模型

```go
// backend/internal/model/ai_config.go

type AIConfig struct {
    BaseModel
    Provider   string `gorm:"size:20;default:anthropic" json:"provider"` // anthropic | openai
    Model      string `gorm:"size:50;default:claude-sonnet-4-6" json:"model"`
    APIKey     string `gorm:"size:500;not null" json:"-"`  // 不序列化�?JSON
    BaseURL    string `gorm:"size:255" json:"base_url"`
    MaxTokens  int    `gorm:"default:4096" json:"max_tokens"`
    IsActive   bool   `gorm:"default:true" json:"is_active"`
    WorkspaceID uint64 `gorm:"not null;uniqueIndex" json:"workspace_id"`
}

type AIThread struct {
    BaseModel
    Title       string  `gorm:"size:255" json:"title"`
    WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
    ProjectID   *uint64 `gorm:"index" json:"project_id"`
    UserID      uint64  `gorm:"not null;index" json:"user_id"`
    Messages    []AIMessage `gorm:"foreignKey:ThreadID" json:"messages"`
}

type AIMessage struct {
    BaseModel
    ThreadID   uint64  `gorm:"not null;index" json:"thread_id"`
    Role       string  `gorm:"size:20;not null" json:"role"` // user | assistant | system | tool
    Content    string  `gorm:"type:text;not null" json:"content"`
    ToolCalls  json.RawMessage `gorm:"type:jsonb" json:"tool_calls"`
    ToolName   *string `gorm:"size:50" json:"tool_name"`
}
```

#### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `model/ai_config.go`, `model/ai_thread.go`, `model/ai_message.go` |
| 新建 | `service/llm_client.go` �?LLM SDK 封装 |
| 新建 | `service/ai_service.go` �?AI 核心服务 |
| 新建 | `handler/ai_handler.go` |
| 新建 | `handler/ai_config_handler.go` |
| 新建 | `frontend/src/components/AIChatSidebar.vue` |
| 新建 | `frontend/src/composables/useAI.ts` |
| 新建 | `frontend/src/types/ai.ts` |
| 新建 | `frontend/src/api/ai.ts` |
| 修改 | `router/router.go`, `cmd/server/main.go` |

---

### Task 21: AI Chat �?对话�?AI 助手

**功能**: 右侧滑出�?AI 聊天面板，支�?Ask 模式（查询）�?Build 模式（执行操作）。SSE 流式响应�?
#### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `frontend/src/components/AIChatSidebar.vue` �?完整聊天 UI |
| 修改 | `frontend/src/views/Project.vue` �?集成 AI 按钮 |
| 修改 | `frontend/src/views/IssueDetail.vue` �?集成 AI 按钮 |
| 修改 | `handler/ai_handler.go` �?Chat SSE 端点 |

---

### Task 22: NL Search �?自然语言搜索

**功能**: 在搜索框支持自然语言输入�?上周未解决的紧急Bug" �?RQL 查询 �?结果列表�?
```
用户输入: "分配给张三的高优先级任务"
    �?AI Service (NL �?RQL)
RQL: priority = "high" AND assignee = "张三" AND type = "Task"
    �?RQL Executor
结果: [Issue#42, Issue#58, Issue#99]
```

#### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `frontend/src/components/RQL/RQLInput.vue` �?NL 开�?|
| 修改 | `handler/ai_handler.go` �?Search 端点 |
| 修改 | `service/ai_service.go` �?NL→RQL 转换 |

---

### Task 23: Smart Create �?智能创建工作�?
**功能**: 自然语言描述需�?�?AI 解析为结构化 Issue �?预览确认 �?创建�?
```
输入: "创建一个登录页面的Bug，P0紧急，分配给张三，截止下周五，需要在Safari上复�?
    �?AI 解析
预览: {
  name: "登录页面Bug - Safari兼容性问�?,
  type: "Bug",
  priority: "urgent",
  assignee: "张三 (id=3)",
  target_date: "2026-07-03",
  description: "## 复现步骤\n1. 打开Safari浏览器\n..."
}
    �?用户确认
创建 Issue #101
```

#### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/AICreateDialog.vue` �?预览确认弹窗 |
| 修改 | `handler/ai_handler.go` �?Create 端点 |
| 修改 | `service/ai_service.go` �?Parse→Preview→Create 流程 |

---

### Task 24: Data Analysis �?AI 数据分析

**功能**: AI 分析项目/周期数据，生成洞察和建议�?
支持的分析类型：
- 项目健康度概�?("这个项目进展如何�?)
- 周期回顾 ("Sprint 3 完成率为什么低�?)
- 瓶颈检�?("哪些任务卡住了？")
- 工作量分�?("张三是否过载�?)

#### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `handler/ai_handler.go` �?Analyze 端点 |
| 修改 | `service/ai_service.go` �?数据聚合 + LLM 分析 |

---

### Task 25: Page AI �?文档 AI 能力

**功能**: �?Page 编辑器中集成 AI 能力：生成内容、总结要点、翻译语言、改进写作�?
```
Page Editor Toolbar:
  [🤖 Generate] [📝 Summarize] [🌐 Translate] [�?Improve]
```

#### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `frontend/src/views/ProjectPages.vue` �?AI 工具�?|
| 修改 | `frontend/src/components/RichTextEditor.vue` �?AI 按钮 |
| 修改 | `handler/ai_handler.go` �?Page AI 端点 |

---

### Task 26: AI Triage �?智能分诊

**功能**: 外部提交�?Issue �?AI 自动分类、优先级建议、可能重复检测�?
```
Intake Form 提交 �?  AI 分析:
    - 类型建议: Bug (置信�?0.92)
    - 优先级建�? High
    - 可能重复: #42 "登录问题" (相似�?0.78)
    - 建议 Assignee: 张三 (相关领域专家)
  �?  Triage 面板展示 AI 建议，管理员一键采�?```

#### 涉及文件

| 类型 | 文件 |
|------|------|
| 修改 | `handler/intake_handler.go` �?提交后触�?AI 分析 |
| 修改 | `service/ai_service.go` �?Triage 分析逻辑 |
| 修改 | `frontend/src/components/TriagePanel.vue` �?展示 AI 建议 |

---

### Task 27: Command Palette (⌘K)

**功能**: 键盘优先的快速导�?搜索/操作面板。类�?VS Code ⌘K�?
```
⌘K �?输入:
  "new bug"        �?快速创�?Bug
  "go to #42"      �?跳转�?Issue #42
  "my issues"      �?列出我的工作�?  "sprint progress"�?显示当前 Sprint 进度
  "page prd"       �?打开 PRD 页面
```

#### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/CommandPalette.vue` |
| 新建 | `frontend/src/composables/useCommandPalette.ts` |
| 修改 | `frontend/src/App.vue` �?全局 ⌘K 监听 |

---

### Task 28: AI Settings �?AI 配置管理

**功能**: 工作空间�?AI 设置页面。选择 Provider、Model、配�?API Key、管理使用量�?
#### 涉及文件

| 类型 | 文件 |
|------|------|
| 新建 | `frontend/src/components/AISettingsPanel.vue` |
| 修改 | `frontend/src/views/WorkspaceSettings.vue` �?"AI" tab |
| 新建 | `handler/ai_config_handler.go` |

---

## 更新后的实施节奏

```
Phase 1 �?工作项基础能力 (Week 1-3):
  Task 1-3:    Templates + Quick Create + Import
  Task 15,10,12: Estimates + Validation + Hierarchy

Phase 2 �?工作项关联与操作 (Week 4-6):
  Task 6,5,8:  Release + Page Link + Bulk Ops
  Task 17,18,4: Attachment + Notification + Sub-items

Phase 3 �?工作项展示与深度 (Week 7-9):
  Task 16,7,13: Quick Filters + Cover + Time Track
  Task 14,9,11,19: Recurring + Merge + Conditional + Intake

Phase 4 �?AI 基础设施 (Week 10-11):
  Task 20: AI Infrastructure     �?所�?AI 功能的前置依�?  Task 27: Command Palette       �?独立，无 LLM 依赖
  Task 28: AI Settings           �?依赖 Task 20

Phase 5 �?AI 核心能力 (Week 12-14):
  Task 21: AI Chat               �?依赖 Task 20
  Task 22: NL Search             �?依赖 Task 20
  Task 23: Smart Create          �?依赖 Task 20

Phase 6 �?AI 深度集成 (Week 15-17):
  Task 24: Data Analysis         �?依赖 Task 21
  Task 25: Page AI               �?依赖 Task 20
  Task 26: AI Triage             �?依赖 Task 19 + Task 20
```

### AI 功能依赖�?
```
Task 20 (AI Infrastructure)
  ├── Task 21 (AI Chat) ────── Task 24 (Data Analysis)
  ├── Task 22 (NL Search)
  ├── Task 23 (Smart Create)
  ├── Task 25 (Page AI)
  ├── Task 26 (AI Triage) ──── 需�?Task 19 (Intake)
  └── Task 28 (AI Settings)

Task 27 (Command Palette) �?独立，无 LLM 依赖
```

---

## 实现规范

### 代码模式速查

| �?| 参考文�?| 关键模式 |
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
| Vue Types | `types/page.ts` | `interface` 定义，字段名�?Go DTO 一�?|
| Vue Composable | `composables/useAI.ts` | SSE EventSource, `ref()`, `onMounted`/`onUnmounted` |

### AI 特有约定

1. **SSE 流式**: Chat 端点使用 Server-Sent Events，`text/event-stream` MIME type
2. **API Key 安全**: AI Config �?`api_key` 字段标记 `json:"-"` 永远不序列化
3. **Tool Calling**: 使用 Claude/OpenAI 原生 Function Calling，工具定义在 `ai_service.go` �?`getTools()` 方法
4. **上下文注�?*: 每个 AI 请求自动附带当前 workspace/project/page/issue 上下�?5. **速率限制**: AI 端点建议添加 per-user rate limiting（后�?middleware 实现�?6. **日志审计**: AI 对话历史存储�?`ai_threads` + `ai_messages` 表，用于审计和改�?