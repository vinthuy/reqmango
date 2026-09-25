# Reqmango API Documentation

> Generated: 2026-09-25
> Base URL: `/api/v1`

---

## Authentication

All protected endpoints require a JWT token in the `Authorization` header:

```
Authorization: Bearer <token>
```

### Get Token

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}

Response: { "token": "eyJ...", "user": { ... } }
```

### Register

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password",
  "display_name": "User Name"
}
```

---

## Workspaces

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workspaces` | List all workspaces |
| POST | `/workspaces` | Create workspace |
| GET | `/workspaces/:wsParam` | Get workspace (ID or slug) |
| PATCH | `/workspaces/:wsParam` | Update workspace |
| DELETE | `/workspaces/:wsParam` | Delete workspace |

### Workspace Members

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workspaces/:wsParam/members` | List members |
| POST | `/workspaces/:wsParam/members` | Add member |
| PATCH | `/workspaces/:wsParam/members/:userId` | Update member role |
| DELETE | `/workspaces/:wsParam/members/:userId` | Remove member |

---

## Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects` | List projects |
| POST | `/projects` | Create project |
| GET | `/projects/:projectId` | Get project |
| PATCH | `/projects/:projectId` | Update project |
| DELETE | `/projects/:projectId` | Delete project |
| GET | `/projects/:projectId/members` | List project members |

---

## Issues

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:projectId/issues` | List issues (with RQL filter) |
| POST | `/projects/:projectId/issues` | Create issue |
| GET | `/issues/:issueId` | Get issue |
| PATCH | `/issues/:issueId` | Update issue |
| DELETE | `/issues/:issueId` | Delete issue (soft) |
| POST | `/issues/:issueId/archive` | Archive issue |
| POST | `/issues/:issueId/restore` | Restore issue |

### Bulk Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/issues/bulk/update` | Bulk update issues |
| POST | `/issues/bulk/delete` | Bulk delete issues |
| POST | `/issues/bulk/move` | Bulk move issues |
| POST | `/issues/bulk/copy` | Bulk copy issues |

### Issue Properties

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/issues/:issueId/assignees` | Add assignee |
| DELETE | `/issues/:issueId/assignees/:userId` | Remove assignee |
| POST | `/issues/:issueId/labels` | Add label |
| DELETE | `/issues/:issueId/labels/:labelId` | Remove label |
| POST | `/issues/:issueId/cycles/:cycleId` | Add to cycle |
| DELETE | `/issues/:issueId/cycles/:cycleId` | Remove from cycle |
| POST | `/issues/:issueId/watchers` | Watch issue |
| DELETE | `/issues/:issueId/watchers` | Unwatch issue |

### RQL Filter

Issues support RQL (Reqmango Query Language) filtering:

```
GET /projects/:projectId/issues?rql=assignee_id = 5 AND state = 'In Progress'
```

**Supported fields:** `name`, `description`, `state`, `state_id`, `priority`, `type`, `assignee`, `assignee_id`, `reporter`, `label`, `label_id`, `cycle`, `cycle_id`, `module`, `module_id`, `created_at`, `updated_at`, `start_date`, `target_date`

**Operators:** `=`, `!=`, `IN`, `NOT IN`, `LIKE`, `>`, `<`, `>=`, `<=`

---

## States

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:projectId/settings/states` | List states |
| POST | `/projects/:projectId/settings/states` | Create state |
| GET | `/projects/:projectId/settings/states/:stateId` | Get state |
| PUT | `/projects/:projectId/settings/states/:stateId` | Update state |
| DELETE | `/projects/:projectId/settings/states/:stateId` | Delete state |

### Workspace States

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workspaces/:wsParam/settings/states` | List workspace states |
| POST | `/workspaces/:wsParam/settings/states` | Create workspace state |
| PUT | `/workspaces/:wsParam/settings/states/:stateId` | Update workspace state |
| DELETE | `/workspaces/:wsParam/settings/states/:stateId` | Delete workspace state |

---

## Workflows

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:projectId/workflows` | List workflows |
| POST | `/projects/:projectId/workflows` | Create workflow |
| GET | `/projects/:projectId/workflows/:workflowId` | Get workflow |
| PUT | `/projects/:projectId/workflows/:workflowId` | Update workflow |
| DELETE | `/projects/:projectId/workflows/:workflowId` | Delete workflow |

### Workflow Nodes

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/projects/:projectId/workflows/:workflowId/nodes` | Add node |
| PUT | `/projects/:projectId/workflows/:workflowId/nodes/:nodeId` | Update node |
| DELETE | `/projects/:projectId/workflows/:workflowId/nodes/:nodeId` | Delete node |

### Workflow Edges

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/projects/:projectId/workflows/:workflowId/edges` | Add edge |
| DELETE | `/projects/:projectId/workflows/:workflowId/edges/:edgeId` | Delete edge |

### State Transitions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:projectId/workflows/:workflowId/transitions` | List transitions |
| POST | `/projects/:projectId/workflows/:workflowId/transitions` | Create transition |
| PUT | `/projects/:projectId/workflows/:workflowId/transitions/:transitionId` | Update transition |
| DELETE | `/projects/:projectId/workflows/:workflowId/transitions/:transitionId` | Delete transition |

**Transition Request:**
```json
{
  "name": "Move to In Progress",
  "source_state_id": 1,
  "target_state_id": 2,
  "rule_type": "allow|approval",
  "approver_ids": "1,2,3",
  "role_allowed": "admin",
  "approval_mode": "any|all",
  "approve_target_state_id": 3,
  "reject_target_state_id": 4
}
```

---

## Comments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/issues/:issueId/comments` | List comments |
| POST | `/issues/:issueId/comments` | Create comment |
| PUT | `/comments/:commentId` | Update comment |
| DELETE | `/comments/:commentId` | Delete comment |

---

## Attachments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/issues/:issueId/attachments` | List attachments |
| POST | `/issues/:issueId/attachments` | Upload file (max 10MB) |
| GET | `/attachments/:attachmentId` | Get attachment info |
| GET | `/attachments/:attachmentId/download` | Download file |
| DELETE | `/attachments/:attachmentId` | Delete attachment |

**Allowed MIME types:** `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `application/pdf`, `text/plain`, `text/csv`, `application/json`, `application/octet-stream`

---

## Cycles

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:projectId/cycles` | List cycles |
| POST | `/projects/:projectId/cycles` | Create cycle |
| GET | `/cycles/:cycleId` | Get cycle |
| PATCH | `/cycles/:cycleId` | Update cycle |
| DELETE | `/cycles/:cycleId` | Delete cycle |

---

## Modules

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:projectId/modules` | List modules |
| POST | `/projects/:projectId/modules` | Create module |
| GET | `/modules/:moduleId` | Get module |
| PATCH | `/modules/:moduleId` | Update module |
| DELETE | `/modules/:moduleId` | Delete module |

---

## Automation Rules

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:projectId/automations` | List rules |
| POST | `/projects/:projectId/automations` | Create rule |
| GET | `/automations/:ruleId` | Get rule |
| PUT | `/automations/:ruleId` | Update rule |
| DELETE | `/automations/:ruleId` | Delete rule |

---

## Labels

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:projectId/labels` | List labels |
| POST | `/projects/:projectId/settings/labels` | Create label |
| PUT | `/projects/:projectId/settings/labels/:labelId` | Update label |
| DELETE | `/projects/:projectId/settings/labels/:labelId` | Delete label |

---

## Custom Fields

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:projectId/custom-fields` | List custom fields |
| POST | `/projects/:projectId/custom-fields` | Create custom field |
| PUT | `/custom-fields/:fieldId` | Update custom field |
| DELETE | `/custom-fields/:fieldId` | Delete custom field |

---

## Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/projects/:projectId/reports/run` | Run report |
| GET | `/projects/:projectId/reports/quick-charts` | Get quick charts |

---

## Notifications

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/notifications` | List notifications |
| GET | `/notifications/summary` | Get summary (unread count) |
| PUT | `/notifications/:id/read` | Mark as read |
| PUT | `/notifications/read-all` | Mark all as read |

---

## AI Agents

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workspaces/:wsParam/agents` | List agents |
| POST | `/workspaces/:wsParam/agents` | Create agent |
| GET | `/workspaces/:wsParam/agents/:id` | Get agent |
| PUT | `/workspaces/:wsParam/agents/:id` | Update agent |
| DELETE | `/workspaces/:wsParam/agents/:id` | Delete agent |
| POST | `/workspaces/:wsParam/agents/:id/dispatch` | Dispatch task |

---

## Tools

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workspaces/:wsParam/tools` | List tools |
| POST | `/workspaces/:wsParam/tools` | Create tool |
| GET | `/workspaces/:wsParam/tools/:toolId` | Get tool |
| PUT | `/workspaces/:wsParam/tools/:toolId` | Update tool |
| DELETE | `/workspaces/:wsParam/tools/:toolId` | Delete tool |
| POST | `/workspaces/:wsParam/tools/call` | Execute tool |

---

## Webhooks

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/projects/:projectId/webhooks` | List webhooks |
| POST | `/projects/:projectId/webhooks` | Create webhook |
| PUT | `/webhooks/:webhookId` | Update webhook |
| DELETE | `/webhooks/:webhookId` | Delete webhook |

---

## Error Responses

All errors follow this format:

```json
{
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

**Common HTTP Status Codes:**
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `409` - Conflict (duplicate email, etc.)
- `500` - Internal Server Error
