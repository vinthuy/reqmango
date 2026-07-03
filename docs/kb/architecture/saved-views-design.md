# Saved Views（保存视图）设计

**最后更新**: 2026-07-03

## 概述

Saved Views 允许用户将常用的筛选条件、列配置、排序方式和分组方式保存为命名视图，在项目中快速切换。灵感来自 Plane 的 Views 功能。

## 数据模型

```go
type SavedView struct {
    BaseModel
    Name        string          // 视图名称（必填）
    Description *string         // 描述（可选）
    ViewType    string          // "list" | "kanban" | "tree" | "gantt" | "calendar"
    Filters     json.RawMessage // JSONB: 筛选条件，如 {"priority":"urgent","state_id":[1,2]}
    RQL         string          // RQL 查询字符串（与 Filters 互补）
    SortConfig  json.RawMessage // JSONB: 排序配置，如 [{"field":"priority","dir":"desc"}]
    Columns     json.RawMessage // JSONB: 列可见性/顺序，如 ["name","priority","state"]
    GroupBy     *string         // 分组字段: "state_id" | "priority" | "assignee_id" | "label" | "cycle_id" | "module_id" | "type_id"
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

## 前端集成

### SavedViewSelector.vue

嵌入 `Project.vue` 的工具栏区域：
- 下拉菜单显示所有可用视图
- 选中视图后 **完整恢复** filters + viewType + columns + groupBy + sortConfig（不只是切换 view_type）
- "Save current view" 按钮将当前筛选状态持久化
- 支持设为默认、复制、删除操作

### FilterBar.vue（统一筛选栏）

2026-07 新增，替代旧的 IssueFilterBar + QuickFilterChips + RQLInput：
- Filters 按钮 → 字段下拉（12 个字段）→ 语义操作符选择 → 值输入
- 条件芯片行展示已激活筛选 + × 删除 + 点击编辑
- RQL 折叠区显示/编辑底层 RQL 字符串
- **RQL 双向同步**：FilterBar 操作 → 自动生成 RQL，手动编辑 RQL → 自动解析为条件芯片
- **即时更新**：条件变更直接触发 API 请求

### useFilters.ts（筛选状态管理）

Provide/Inject 架构，全局共享筛选状态：
```typescript
interface FiltersState {
  filters: FilterCondition[]       // 当前筛选条件
  sortBy: SortOption | null        // 当前排序
  groupBy: GroupOption | null      // 当前分组
  quickSearch: string              // 快速搜索文本
  searchHistory: string[]          // 搜索历史（localStorage, 最多 10 条）
}
```

### RQL 双向同步

- `buildRQL(filters, quickSearch)` → 自动派生 RQL 字符串
- `parseRQL(rqlStr)` → 解析 RQL → FilterCondition[] + SortOption
- 视图保存时同时存储 `filters`（JSONB）和 `rql`（字符串）

## 筛选字段（12 个）

| 字段 Key | DB 映射 | 类型 | 支持的操作符 |
|----------|---------|------|-------------|
| `sequence_id` | `sequence_id` | number | is, is not |
| `title` | `name` | text | is, is not, contains, does not contain |
| `state_id` | `state_id` | select | is, is any of, is not, is not any of |
| `state_group` | `state_group` | select | is, is any of, is not, is not any of |
| `priority` | `priority` | select | is, is any of, is not, is not any of |
| `assignee_id` | `assignee_id` | select | is, is any of, is not, is not any of, is empty, is not empty |
| `label` | `label` | multi | is, is any of, is not, is not any of, is empty, is not empty |
| `cycle_id` | `cycle_id` | select | is, is any of, is not, is not any of, is empty, is not empty |
| `module_id` | `module_id` | select | is, is any of, is not, is not any of, is empty, is not empty |
| `type_id` | `issue_type_id` | select | is, is any of, is not, is not any of |
| `start_date` | `start_date` | date | is, is not, before, after, before or on, after or on, between, not between, is empty, is not empty |
| `target_date` | `target_date` | date | is, is not, before, after, before or on, after or on, between, not between, is empty, is not empty |
| `created_at` | `created_at` | date | is, is not, before, after, before or on, after or on, between, not between, is empty, is not empty |

## 排序选项（5 种）

| Key | 方向 |
|-----|------|
| `created_at` | desc |
| `updated_at` | desc |
| `priority` | desc |
| `start_date` | asc |
| `target_date` | asc |

## 分组选项（8 种）

| Key | 说明 |
|-----|------|
| `none` | 不分组 |
| `state_id` | 按状态 |
| `priority` | 按优先级 |
| `assignee_id` | 按负责人 |
| `label` | 按标签 |
| `cycle_id` | 按周期 |
| `module_id` | 按模块 |
| `type_id` | 按类型 |

## 语义操作符

| 操作符 | RQL 映射 | 说明 |
|--------|---------|------|
| `is` | `=` | 等于 |
| `is not` | `!=` | 不等于 |
| `contains` | `LIKE "%value%"` | 包含 |
| `does not contain` | `NOT LIKE "%value%"` | 不包含 |
| `is any of` | `IN (...)` | 多选包含 |
| `is not any of` | `NOT IN (...)` | 多选排除 |
| `is empty` | `IS NULL` | 为空 |
| `is not empty` | `IS NOT NULL` | 非空 |
| `before` | `<` | 日期早于 |
| `after` | `>` | 日期晚于 |
| `before or on` | `<=` | 日期不晚于 |
| `after or on` | `>=` | 日期不早于 |
| `between` | `>= v1 AND <= v2` | 日期范围内 |
| `not between` | `< v1 OR > v2` | 日期范围外 |

## 设计决策

1. **JSONB 存储筛选条件**：不同视图的筛选维度差异大，JSONB 提供最大灵活性，无需 schema 变更
2. **按用户+项目隔离**：视图是用户级别的配置，共享功能通过 `is_shared` 标志实现
3. **默认视图排序优先**：List 查询以 `is_default DESC` 排序，确保默认视图最先展示
4. **RQL + Filters 双存储**：Filters 提供结构化数据便于 UI 恢复，RQL 提供可读字符串便于复制/分享
5. **Views 完整恢复**：选择视图时恢复 viewType + filters + columns + groupBy + sortConfig 全部状态
