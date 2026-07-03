# Data Model（数据模型总览）

**最后更新**: 2026-07-03

---

## 通用字段

所有 Go 后端模型嵌入 `BaseModel`：

```go
type BaseModel struct {
    ID          uint64         `gorm:"primaryKey;autoIncrement"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
    CreatedByID *uint64
    UpdatedByID *uint64
}
```

统一 `uint64` 主键、自动时间戳、GORM 软删除。

---

## 核心实体关系图

```
Workspace 1──N WorkspaceMember N──1 User
Workspace 1──N Project
Project  1──N ProjectMember N──1 User
Workspace 1──N Role
Role     N──M Permission (role_permissions)
Project  1──N State
Project  1──N Label
Project  1──N Issue
Project  1──N Cycle
Project  1──N Module
Project  1──N Page
Project  1──N Release
Project  1──N SavedView
Project  1──N WorkItemTemplate
Project  1──1 ProjectEstimateSettings

Issue    N──M Label      (issue_labels)
Issue    N──M User       (issue_assignees)
Issue    N──1 Cycle      (issue_cycles)
Issue    N──M Module     (module_issues)
Issue    1──1 Issue      (parent_id, 子任务)
Issue    N──M Page       (issue_pages)
Issue    N──M Release    (release_issues)
Issue    1──N Attachment
Issue    1──N Comment    (支持嵌套 reply)

State    1──N StateTransition (from_state, to_state)
Module   1──1 Module     (parent_id, 树形)
Page     1──1 Page       (parent_id, 树形, depth≤5)

Workspace 1──1 AIConfig
User     1──N AIThread
AIThread 1──N AIMessage
User     1──N Notification
```

---

## 表清单

### 用户与组织 (5 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `users` | email, username, display_name, password_hash, is_active | 用户 |
| `workspaces` | name, slug, owner_id | 工作空间 |
| `workspace_members` | workspace_id, user_id, role | 空间成员 |
| `projects` | name, identifier, workspace_id, color, archived_at, template_id, project_lead_id | 项目 |
| `project_members` | project_id, user_id, role | 项目成员 |
| `project_subscribers` | project_id, user_id | 项目订阅者 |

### 项目配置 (5 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `states` | name, color, group, sequence, is_default, project_id | 工作项状态 (5 固定分组) |
| `state_transitions` | from_state_id, to_state_id, workflow_id, rule_type, approver_ids | 状态流转规则 |
| `labels` | name, color, project_id | 标签 |
| `workflows` | name, description, project_id, issue_type_id | 工作流 |
| `automation_rules` | name, trigger_type, conditions (JSONB), actions (JSONB), project_id | 自动化规则 |

### 工作项 (7 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `issues` | name, description_html, description_json (JSONB), priority, sequence_id, sort_order, start/target_date, state_id, project_id, parent_id(depth≤5), issue_type_id, is_draft, archived_at | 核心工作项 |
| `issue_assignees` | issue_id, user_id | 工作项分配 |
| `issue_labels` | issue_id, label_id | 工作项标签 |
| `issue_cycles` | issue_id, cycle_id | 工作项-周期关联 |
| `issue_activities` | issue_id, actor_id, verb, field, old_value, new_value | 操作历史 |
| `issue_pages` | issue_id, page_id | 工作项-文档关联 |
| `release_issues` | release_id, issue_id | 发布-工作项关联 |

### 工作项配置 (4 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `issue_types` | name, color, icon, level, parent_type_id, project_id (nullable), workspace_id | 工作项类型 |
| `issue_type_fields` | type_id, field_id, is_required, sequence | 类型-字段绑定 |
| `issue_type_templates` | name, color, icon, level, parent_type_id, workspace_id | 类型蓝图 (工作空间级) |
| `issue_type_template_fields` | template_type_id, field_id, is_required | 蓝图-字段绑定 |

### 自定义字段 (3 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `custom_fields` | name, field_type (text/number/dropdown/boolean/date/member/url), is_required, project_id, workspace_id | 自定义字段定义 |
| `custom_field_options` | field_id, value, color, sequence | 下拉选项 |
| `issue_custom_field_values` | issue_id, field_id, value | 工作项字段值 |

### 模板 (3 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `project_templates` | name, description, workspace_id, is_default | 项目模板 |
| `project_template_types` | template_id, type_template_id, is_required, sequence | 模板-类型关联 |
| `work_item_templates` | name, issue_type_id, defaults (JSONB), project_id, workspace_id | 工作项创建模板 |

### 周期与模块 (3 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `cycles` | name, description, start_date, end_date, completed_at, cancelled_at, project_id | 迭代周期 |
| `modules` | name, description, parent_id, order, project_id | 功能模块 (树形) |
| `module_issues` | module_id, issue_id | 模块-工作项关联 |

### 评论与附件 (2 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `comments` | issue_id, author_id, body, parent_id (嵌套), is_resolved | 评论 |
| `attachments` | name, file_path, file_size, mime_type, issue_id, uploader_id | 附件 |

### 关联类型 (2 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `relation_types` | name, inward_name, outward_name, workspace_id | 自定义关联类型 |
| `issue_relations` | issue_id, related_issue_id, relation_type_id | 工作项关联 |

### 估算 (4 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `estimate_points` | name, value, mode, project_id | Points 估算 (Fibonacci 等) |
| `estimate_categories` | name, mode, project_id | Categories 估算 (T-shirt size 等) |
| `estimate_times` | name, minutes, mode, project_id | Time 估算 |
| `project_estimate_settings` | project_id, mode, points/categories/time_enabled | 项目估算配置 |

### 发布与保存视图 (2 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `releases` | name, version, status, release_date, project_id | 发布版本 |
| `saved_views` | name, view_type, filters/rql/sort_config/columns (JSONB), group_by, is_default, is_shared, owner_id, project_id | 保存的筛选视图（含完整恢复链路） |

### 通知 (1 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `notifications` | title, message, type, priority, is_read, recipient_id, sender_id, project_id, issue_id, action_url | 用户通知 |

### 文档 (1 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `pages` | title, content, content_json (JSONB), parent_id (depth≤5), sequence, published, archived_at, project_id | Wiki/文档页面 |

### 工时 (1 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `time_tracks` | issue_id, user_id, description, started_at, ended_at, duration (seconds) | 工时记录 |

### 重复工作项 (1 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `recurrence_rules` | issue_id (unique), frequency (daily/weekly/monthly/cron), interval, cron_expr, next_run, end_date | 周期自动创建规则 |

### 条件字段 (1 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `conditional_fields` | field_id, condition_type, operator, condition_values (JSON), workspace_id | 字段显隐规则 |

### AI (3 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `ai_configs` | provider, model, api_key, max_tokens, workspace_id | 工作空间 AI 配置 |
| `ai_threads` | title, workspace_id, project_id, user_id | AI 对话线程 |
| `ai_messages` | thread_id, role, content, tool_calls (JSONB), tool_name | AI 对话消息 |

### RBAC 权限系统 (3 表)

| 表 | 关键字段 | 说明 |
|----|----------|------|
| `roles` | name, description, scope, workspace_id, project_id, level(5/15/20), is_system, sort_order | 角色（系统默认 + 自定义） |
| `permissions` | code (unique, 如 issue:create), name, description, resource, action, scope | 权限枚举（55 个） |
| `role_permissions` | role_id + permission_id (复合主键) | 角色-权限多对多关联 |

---

## 关键设计决策

1. **软删除**: 所有表通过 GORM `DeletedAt` 实现软删除
2. **uint64 主键**: 统一使用 `uint64`，`sequence_id` 在项目内自增用于外部标识
3. **JSONB 存储**: 筛选条件、自动化条件/动作、AI Tool Calls、模板默认值等灵活结构使用 JSONB
4. **树形结构**: Module 和 Page 通过 `parent_id` 自引用，depth 字段冗余存储避免递归查询
5. **多对多关联**: 使用显式关联表而非 GORM many2many，便于查询和扩展
6. **RBAC 权限**: 55 个细粒度权限以 `resource:action` 格式编码，通过 `role_permissions` 关联表实现角色-权限绑定
7. **AI 协议双支持**: Anthropic + OpenAI-compatible (DeepSeek) 双协议，按 provider 自动切换
