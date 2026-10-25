# Notifications（通知系统）设计

**最后更新**: 2026-10-25

## 概述

通知系统把业务事件写成 `notifications` 行，并用 `GET /api/v1/sse` 推到已登录用户。`NotificationCenter.vue` 读这些记录。模型注释里的 `info | warning | error | success` 不是实际写入的类型。

## 数据模型

```go
type Notification struct {
    BaseModel
    Title       string  // 通知标题
    Message     string  // 通知内容
    Type        string  // 见下文「写入的类型」。创建接口不校验枚举
    Priority    string  // low | medium | high | urgent。自动通知里多数是 medium，分配和到期提醒是 high
    IsRead      bool    // 是否已读
    ReadAt      *string // 已读时间
    ActionURL   *string // 点击跳转链接
    RecipientID uint64  // 接收者
    SenderID    *uint64 // 发送者（可选）
    ProjectID   *uint64 // 关联项目（可选）
    IssueID     *uint64 // 关联工作项（可选）
}
```

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/notifications` | 列表（支持 unread_only, limit, offset） |
| GET | `/notifications/summary` | 未读摘要（total, unread, unread_by_type） |
| GET | `/notifications/:id` | 获取单条通知 |
| POST | `/notifications` | 创建通知 |
| POST | `/notifications/bulk` | 批量创建（多接收者） |
| POST | `/notifications/check-due-reminders` | 扫描 24 小时内到期的工作项并写提醒。要登录，并且 `IsSuperuser` 为真，否则 403 |
| PATCH | `/notifications/:id/read` | 标记已读 |
| POST | `/notifications/read-all` | 全部已读 |
| DELETE | `/notifications/:id` | 删除通知 |

## 事件和写入的类型

`TriggerNotification` 按事件名决定 `type` 和 `priority`。没有单独分支的事件落成 `info` / `medium`。

| 事件 | 谁调用 | 写入的 type | priority |
|------|--------|-------------|----------|
| `issue_assigned` | 工作项分配 | `assignment` | high |
| `issue_mentioned` | 评论里的 @ | `assignment` | high |
| `issue_state_changed` | 状态变更 | `status` | medium |
| `issue_commented` | 新评论 | `comment` | medium |
| `issue_created` | 创建工作项 | `info` | medium |
| `issue_due_soon` | 到期提醒 | `reminder` | high |
| `cycle_created` | 创建周期 | `info` | medium |
| `approval_requested` | 发起审批 | `info`（无专门分支） | medium |
| `approval_decided` | 审批决定 | `info`（无专门分支） | medium |

`issue_updated` 在 switch 里会写成 `info`，当前没有调用点。没有「周期开始/结束」事件。

到期提醒由 `CheckDueDateReminders` 执行：`target_date` 在现在到 24 小时之间、未归档，并且还没有同工作项、`type=reminder`、标题含「到期」的通知。接收人是该工作项的负责人。

SSE 连接是 `GET /api/v1/sse`。它不走 `ScopedAuthMiddleware`，handler 自己验 token，因为浏览器 `EventSource` 不能设置 `Authorization`。

## 访问控制

- 列表只返回 `recipient_id` 为当前用户的通知
- 路由都要登录
- 列表按 `created_at DESC`

## 设计决策

1. 实时通道是 SSE。REST 列表仍在，前端可以在连接断开后重新拉取。
2. 已读用 `PATCH /:id/read` 和 `POST /read-all`。`DELETE /:id` 会删掉这一条，不是只做已读。
3. `POST /notifications/bulk` 把同一内容发给多个接收者。业务侧用 `TriggerNotificationsBulk`。
