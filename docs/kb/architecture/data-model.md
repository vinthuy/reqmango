# Data Model（数据模型总览）

**最后更新**: 2026-06-25

---

## 通用字段

所有 Go 后端模型嵌入 `BaseModel`：

```go
type BaseModel struct {
    ID        uint64         `gorm:"primaryKey;autoIncrement"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

统一 `uint64` 主键、自动时间戳、GORM 软删除。

---

## 核心实体关系图

```
Workspace 1──N WorkspaceMember N──1 User
Workspace 1──N Project
Project  1──N ProjectMember N──1 User
Project  1──N State
Project  1──N Label
Project  1──N Issue
Project  1──N Cycle
Project  1──N Module

Issue    N──M Label      (issue_labels)
Issue    N──M User       (issue_assignees)
Issue    N──1 Cycle      (issue_cycles, via Issue.CycleID)
Issue    N──M Module     (module_issues)
Issue    1──1 Issue      (parent_id, 子任务)

State    1──N StateTransition (from_state, to_state)
Module   1──1 Module     (parent_id, 树形结构)
```

---

## 表清单

### 用户与组织

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `users` | email, username, display_name, password_hash, is_active | 用户 |
| `workspaces` | name, slug, owner_id | 工作空间 |
| `workspace_members` | workspace_id, user_id, role | 空间成员 |
| `projects` | name, identifier, workspace_id, archived_at | 项目 |
| `project_members` | project_id, user_id, role | 项目成员 |

### 项目配置

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `states` | name, color, group, sequence, project_id | 工作项状态 |
| `state_transitions` | from_state_id, to_state_id, project_id | 状态流转规则 |
| `labels` | name, color, project_id | 标签 |

### 工作项

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `issues` | name, description, priority, state_id, project_id, sequence_id | 核心工作项 |
| `issue_assignees` | issue_id, user_id | 工作项分配 |
| `issue_labels` | issue_id, label_id | 工作项标签 |

### 周期

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `cycles` | name, description, start_date, end_date, status, project_id | 迭代周期 |

### 模块

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `modules` | name, description, parent_id, order, project_id | 功能模块(树形) |
| `module_issues` | module_id, issue_id | 模块-工作项关联 |

### 活动

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `issue_activities` | issue_id, actor_id, verb, field, old_value, new_value | 操作历史 |

---

## 关键设计决策

### 1. 软删除

所有表通过 `DeletedAt` (GORM soft delete) 实现软删除，查询自动排除已删除记录。

### 2. uint64 主键

使用 `uint64` 而非 UUID，兼顾性能和可读性。序列号 `sequence_id` 在项目内自增，用于 Issue 外部标识。

### 3. 多对多关联

Issue-Label 和 Issue-User 通过显式关联表 (`issue_labels`, `issue_assignees`) 实现多对多，而非 GORM many2many 标签，方便查询和扩展。

### 4. 树形结构

Module 通过 `parent_id` 自引用实现树形层级，前端通过 `ModuleTreeNode` 递归渲染。

### 通知

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `notifications` | title, message, type (info/warning/error/success), priority (low/medium/high/urgent), is_read, recipient_id, sender_id, project_id, issue_id | 用户通知 |

### 保存视图

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `saved_views` | name, view_type (list/kanban), filters (JSONB), sort_config (JSONB), columns (JSONB), group_by, is_default, is_shared, owner_id, project_id | 保存的筛选视图 |

### 页面文档

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `pages` | title, content, content_json (JSONB), parent_id (树形层级，max depth 5), sequence, published, archived_at, project_id | Wiki/文档页面 |

---

## 关键设计决策

### 1. 软删除

所有表通过 `DeletedAt` (GORM soft delete) 实现软删除，查询自动排除已删除记录。

### 2. uint64 主键

使用 `uint64` 而非 UUID，兼顾性能和可读性。序列号 `sequence_id` 在项目内自增，用于 Issue 外部标识。

### 3. 多对多关联

Issue-Label 和 Issue-User 通过显式关联表 (`issue_labels`, `issue_assignees`) 实现多对多，而非 GORM many2many 标签，方便查询和扩展。

### 4. 树形结构

Module 和 Page 通过 `parent_id` 自引用实现树形层级，前端通过递归组件渲染。

### 5. Saved Views JSONB 存储

`filters`, `sort_config`, `columns` 使用 PostgreSQL JSONB 存储，无需 schema 变更即可灵活扩展筛选维度。

---

## 尚未实现

| 模型 | 表 | 说明 |
|------|----|------|
| EstimatePoint | `estimate_points` | 估算点（仅前端） |
| Attachment | `attachments` | 附件元数据（仅前端） |
| AIThread | `ai_threads` | AI 对话线程 |
| AIMessage | `ai_messages` | AI 消息历史 |
