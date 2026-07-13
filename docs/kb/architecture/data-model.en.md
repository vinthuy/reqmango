# Data Model Overview

**Last Updated**: 2026-07-13

---

## Common Fields

All Go backend models embed `BaseModel`:

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

Unified `uint64` primary keys, automatic timestamps, GORM soft delete.

---

## Core Entity Relationships

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
Issue    1──1 Issue      (parent_id, subtasks)
Issue    N──M Page       (issue_pages)
Issue    N──M Release    (release_issues)
Issue    1──N Attachment
Issue    1──N Comment    (nested replies)

State    1──N StateTransition (source_state_id, target_state_id)
Module   1──1 Module     (parent_id, tree)
Page     1──1 Page       (parent_id, tree, depth≤5)

Workspace 1──1 AIConfig
Workspace 1──N Agent
Agent    1──N AgentActivity
Issue    1──N AgentActivity
Workspace 1──N Initiative
Initiative N──M Project   (initiative_projects)
Workspace 1──N Plugin
Plugin    1──N PluginEventLog
Workspace 1──N PageTemplate
Page      1──N PageVersion
Workspace 1──N MCPConfig
Workspace 1──N FieldPermission
Role      N──M FieldPermission
Project  1──N ProjectUpdate
Project  1──N ProjectPageTab
Project  1──N Webhook
Project  1──N SlackConnection
Project  1──N GitHubConnection
Project  1──N SavedDashboard
SavedDashboard 1──N DashboardWidget
Project  1──N SavedReport
Project  1──N SearchTemplate
AutomationRule 1──N AutomationExecution
User     1──N AIThread
AIThread 1──N AIMessage
User     1──N Notification
```

---

## Table List

### Users & Organization (6 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `users` | email, username, display_name, password_hash, is_active | Users |
| `workspaces` | name, slug, owner_id | Workspaces |
| `workspace_members` | workspace_id, user_id, role | Workspace members |
| `projects` | name, identifier, workspace_id, color, archived_at, template_id, project_lead_id | Projects |
| `project_members` | project_id, user_id, role | Project members |
| `project_subscribers` | project_id, user_id | Project subscribers |

### Project Configuration (5 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `states` | name, color, group, sequence, is_default, project_id | Work item states (5 fixed groups) |
| `state_transitions` | source_state_id, target_state_id, workflow_id, rule_type, approver_ids, name, description, issue_type_id, is_auto, role_allowed, project_id, workspace_id | State transition rules |
| `labels` | name, color, project_id | Labels |
| `workflows` | name, description, project_id (nullable), issue_type_id, workspace_id | Workflows (workspace-level support) |
| `automation_rules` | name, trigger_type, conditions (text), actions (text), description, is_enabled, sequence, execution_count, project_id | Automation rules |

### Work Items (7 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `issues` | name, description_html, description_json (JSONB), priority, sequence_id, sort_order, start/target_date, state_id, project_id, parent_id(depth≤5), issue_type_id, is_draft, archived_at | Core work items |
| `issue_assignees` | issue_id, user_id | Issue assignments |
| `issue_labels` | issue_id, label_id | Issue labels |
| `issue_cycles` | issue_id, cycle_id | Issue-cycle associations |
| `issue_activities` | issue_id, actor_id, verb, field, old_value, new_value | Operation history |
| `issue_pages` | issue_id, page_id | Issue-document associations |
| `release_issues` | release_id, issue_id | Release-issue associations |

### Work Item Configuration (4 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `issue_types` | name, color, icon, level, parent_type_id, project_id (nullable), workspace_id | Work item types |
| `issue_type_fields` | type_id, field_id, is_required, sequence | Type-field bindings |
| `issue_type_templates` | name, color, icon, level, parent_type_id, workspace_id | Type templates (workspace-level) |
| `issue_type_template_fields` | template_type_id, field_id, is_required | Template-field bindings |

### Custom Fields (4 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `custom_fields` | name, field_type (text/number/dropdown/boolean/date/member/url), is_required, project_id, workspace_id | Custom field definitions |
| `custom_field_options` | field_id, value, color, sequence | Dropdown options |
| `issue_custom_field_values` | issue_id, field_id, value | Issue field values |
| `project_custom_field_enrollments` | project_id, field_id, is_enabled | Project-level field enablement |

### Templates (3 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `project_templates` | name, description, workspace_id, is_default | Project templates |
| `project_template_types` | template_id, type_template_id, is_required, sequence | Template-type associations |
| `work_item_templates` | name, issue_type_id, defaults (JSONB), project_id, workspace_id | Work item creation templates |

### Cycles & Modules (3 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `cycles` | name, description, start_date, end_date, completed_at, cancelled_at, project_id | Iteration cycles |
| `modules` | name, description, parent_id, order, project_id | Functional modules (tree) |
| `module_issues` | module_id, issue_id | Module-issue associations |

### Comments & Attachments (2 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `comments` | issue_id, author_id, body, parent_id (nested), is_resolved | Comments |
| `attachments` | name, file_path, file_size, mime_type, issue_id, uploader_id | Attachments |

### Relation Types (2 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `relation_types` | name, inward_name, outward_name, workspace_id | Custom relation types |
| `issue_relations` | issue_id, related_issue_id, relation_type_id | Issue relations |

### Estimation (4 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `estimate_points` | name, value, mode, project_id | Points estimation (Fibonacci, etc.) |
| `estimate_categories` | name, mode, project_id | Categories estimation (T-shirt size, etc.) |
| `estimate_times` | name, minutes, mode, project_id | Time estimation |
| `project_estimate_settings` | project_id, mode, points/categories/time_enabled | Project estimation config |

### Releases & Saved Views (2 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `releases` | name, version, status, release_date, project_id | Releases |
| `saved_views` | name, view_type, filters/rql/sort_config/columns (JSONB), group_by, is_default, is_shared, owner_id, project_id | Saved filtered views |

### Notifications (1 table)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `notifications` | title, message, type, priority, is_read, recipient_id, sender_id, project_id, issue_id, action_url | User notifications |

### Documents (1 table)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `pages` | title, content, content_json (JSONB), parent_id (depth≤5), sequence, published, archived_at, project_id | Wiki/document pages |

### Time Tracking (1 table)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `time_tracks` | issue_id, user_id, description, started_at, ended_at, duration (seconds) | Time records |

### Recurring Work Items (1 table)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `recurrence_rules` | issue_id (unique), frequency (daily/weekly/monthly/cron), interval, cron_expr, next_run, end_date | Recurring rules |

### Conditional Fields (1 table)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `conditional_fields` | field_id, condition_type, operator, condition_values (text), workspace_id | Field visibility rules |

### AI (3 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `ai_configs` | provider, model, api_key, max_tokens, workspace_id | Workspace AI config |
| `ai_threads` | title, workspace_id, project_id, user_id | AI conversation threads |
| `ai_messages` | thread_id, role, content, tool_calls (JSONB), tool_name | AI conversation messages |

### RBAC System (3 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `roles` | name, description, scope, workspace_id, project_id, level(5/15/20), is_system, sort_order | Roles (system default + custom) |
| `permissions` | code (unique, e.g. issue:create), name, description, resource, action, scope | Permission enumeration (55 items) |
| `role_permissions` | role_id + permission_id (composite PK) | Role-permission associations |

### Agent System (2 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `agents` | name, description, agent_type, config (JSONB), is_active, workspace_id | AI Agent config |
| `agent_activities` | agent_id, issue_id, action, status, result, started_at, completed_at | Agent execution records |

### Integration (4 tables)

| Table | Key Fields | Description |
|-------|------------|-------------|
| `webhooks` | url, events, secret, is_active, project_id | Webhook config |
| `slack_connections` | channel, config, project_id | Slack integration |
| `github_connections` | repo, config, project_id | GitHub integration |
| `mcp_configs` | name, server_url, config (JSONB), workspace_id | MCP Server config |

---

## 🌐 Language

- **English** (this document)
- [中文文档](data-model.md)