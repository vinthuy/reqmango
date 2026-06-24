# Saved Views（保存视图）设计

**最后更新**: 2026-06-25

## 概述

Saved Views 允许用户将常用的筛选条件、列配置、排序方式和分组方式保存为命名视图，在项目中快速切换。灵感来自 Plane 的 Views 功能。

## 数据模型

```go
type SavedView struct {
    BaseModel
    Name        string          // 视图名称（必填）
    Description *string         // 描述（可选）
    ViewType    string          // "list" | "kanban"
    Filters     json.RawMessage // JSONB: RQL 筛选条件，如 {"priority":"urgent","state_id":[1,2]}
    SortConfig  json.RawMessage // JSONB: 排序配置，如 [{"field":"priority","dir":"desc"}]
    Columns     json.RawMessage // JSONB: 列可见性/顺序，如 ["name","priority","state"]
    GroupBy     *string         // 分组字段: "state_id" | "priority" | "assignee_id"
    IsDefault   bool            // 是否为用户在此项目的默认视图
    IsShared    bool            // 是否共享给项目成员
    OwnerID     uint64          // 创建者
    ProjectID   uint64          // 所属项目
}
```

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/projects/:id/views` | 列出用户可用视图（自己的 + 共享的） |
| POST | `/projects/:id/views` | 创建新视图 |
| GET | `/projects/:id/views/:viewId` | 获取单个视图 |
| PUT | `/projects/:id/views/:viewId` | 更新视图 |
| DELETE | `/projects/:id/views/:viewId` | 删除视图（仅所有者） |
| POST | `/projects/:id/views/:viewId/set-default` | 设为默认视图 |
| POST | `/projects/:id/views/:viewId/duplicate` | 复制视图 |

## 访问控制

- 用户只能看到自己创建的视图 + 被标记为 `is_shared=true` 的视图
- 只有所有者可以更新、删除、设为默认
- 任何用户都可以复制共享视图（副本属于自己）

## 前端集成

`SavedViewSelector.vue` 组件嵌入 `Project.vue` 的工具栏区域：
- 下拉菜单显示所有可用视图
- 选中视图后 emit `select` 事件，父组件应用 filters/columns/groupBy
- "Save current view" 按钮将当前筛选状态持久化
- 支持设为默认、复制、删除操作

## 设计决策

1. **JSONB 存储筛选条件**：不同视图的筛选维度差异大，JSONB 提供最大灵活性，无需 schema 变更
2. **按用户+项目隔离**：视图是用户级别的配置，共享功能通过 `is_shared` 标志实现
3. **默认视图排序优先**：List 查询以 `is_default DESC` 排序，确保默认视图最先展示
