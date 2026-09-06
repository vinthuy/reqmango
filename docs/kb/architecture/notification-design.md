# Notifications（通知系统）设计

**最后更新**: 2026-08-30

## 概述

通知系统为平台提供实时用户通知能力。前端组件 `NotificationCenter.vue` 已有完整 UI，本次补齐后端 model + service + handler。

## 数据模型

```go
type Notification struct {
    BaseModel
    Title       string  // 通知标题
    Message     string  // 通知内容
    Type        string  // 类型: info | warning | error | success
    Priority    string  // 优先级: low | medium | high | urgent
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
| PATCH | `/notifications/:id/read` | 标记已读 |
| POST | `/notifications/read-all` | 全部已读 |
| DELETE | `/notifications/:id` | 删除通知 |

## 访问控制

- 用户只能看到自己（recipient_id = currentUser.ID）的通知
- 所有操作需要 JWT 认证
- 通知按 `created_at DESC` 排序（最新在前）

## 通知类型与优先级

系统支持 9 种通知类型，通过 SSE (Server-Sent Events) 实时推送到前端，辅以 REST API 轮询作为降级方案。

| Type | 用途 | 示例 |
|------|------|------|
| info | 一般信息 | "新成员加入项目" |
| warning | 警告 | "截止日期临近" |
| error | 错误 | "构建失败" |
| success | 成功 | "部署完成" |

| Priority | 含义 |
|----------|------|
| low | 低优先级，不打扰 |
| medium | 默认优先级 |
| high | 重要通知 |
| urgent | 需要立即关注 |

## 集成点

notification_service 已实现 9 种通知类型 + 提醒机制，覆盖以下自动触发场景：
- Issue 分配（assignee 变更） → 通知被分配人
- State 变更 → 通知关注者
- Comment 回复 → 通知被回复人
- Cycle 开始/结束 → 通知项目成员
- 以及更多业务事件通知

## 设计决策

1. **SSE 实时推送**：已升级为 SSE (Server-Sent Events) 实时推送。sse_hub.go 实现了连接管理，支持 9 种通知类型实时推送到前端。
2. **通知不删除而是标记已读**：保留通知历史，用户可通过"全部已读"清理未读状态
3. **Bulk Create 端点**：支持一条内容发送给多个接收者，避免 N 次单独请求
