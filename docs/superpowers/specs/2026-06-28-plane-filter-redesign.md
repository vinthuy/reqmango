# 筛选系统重新设计 — Design Spec v2.0

> **Status**: v2.0 — 对标主流项目管理平台筛选规范  
> **Date**: 2026-06-28  
> **Scope**: 清空全部现有条件搜索代码，完全对标主流平台筛选规范  

---

## 1. 背景与目标

### 1.1 问题分析

| 问题 | 影响 |
|---|---|
| 多套筛选逻辑并存 | 用户体验分裂 |
| props drilling | 耦合严重 |
| RQL 与可视化筛选割裂 | 用户需在两种模式间抉择 |
| 操作符使用原始符号（`=`/`!=`） | 与主流平台语义操作符不一致 |
| 日期操作符不足 | 缺少 `before`/`after`/`between` |
| 无 State Group、标题筛选 | 功能缺失 |
| 需确认/取消按钮 | 与主流平台即时更新理念不一致 |

### 1.2 目标

完全对标主流项目管理平台筛选规范：

1. **语义操作符** — `is` / `is not` / `is any of` / `contains` / `does not contain` / `is empty` 等
2. **即时更新** — 条件变更实时筛选，无确认/取消按钮
3. **完整日期操作符** — `before` / `after` / `after or on` / `before or on` / `between`
4. **标题作为筛选字段** — Title `is` / `contains` / `does not contain`
5. **State Group** — Backlog / Started / Completed 状态组
6. **去除快速预设** — 引导用户使用 Saved Views
7. **单一 FilterBar** + Provide/Inject 数据流
8. **RQL 双向同步**
9. **完整 i18n**

### 1.3 主流平台 vs ReqMango 差距对照

| # | 差距 | 主流平台 | ReqMango 旧 | v2 修正 |
|---|---|---|---|---|
| 1 | 操作符 | `is` / `contains` / `is any of` | `=` / `~` / `in` | 改为语义标签 |
| 2 | 即时更新 | 选完即筛选 | 需点确认/取消 | 去掉确认步骤 |
| 3 | 日期操作符 | `before` / `after` / `between`... | `>=` / `<=` / `>` | 补齐 6 个操作符 |
| 4 | 标题字段 | Title 是正式筛选字段 | 独立搜索框 | 改为筛选字段 |
| 5 | State Group | Backlog/Started/Completed | 无 | 新增 |
| 6 | 预设按钮 | 用 Saved Views | 5 个快速预设 | 去除预设按钮 |

---

## 2. 架构设计

### 2.1 新架构

```
FilterBar.vue (唯一筛选入口)
  │
  ├── useFilters() composable (provide)
  │     ├── filters: FilterCondition[]     ← 所有激活的条件（即时更新）
  │     ├── rql: ComputedRef<string>       ← 自动派生
  │     └── addFilter(field, operator, value)  → 即时触发
  │
  └── UI 构成 (主流平台风格):
        [Filters icon ▾] [Title is "bug fix" ×] [State is any of [Todo,In Progress] ×] [+ Add filter] [Clear all] [RQL ▸]
```

### 2.2 数据流（即时更新）

```
FilterBar 操作
  │ addFilter() / removeFilter() — 操作即生效（无确认按钮）
  ▼
useFilters state (reactive)
  │ filters[] 变化 → rql computed 自动更新 → Project.vue watch(rql) → 重新请求 API
  ▼
视图组件 inject() → 获取 rql → 传给 API
```

### 2.3 文件变更清单

| 操作 | 文件 |
|---|---|
| 新建 | `types/filters.ts` |
| 新建 | `composables/useFilters.ts` |
| 新建 | `components/FilterBar.vue` |
| 删除 | `components/IssueFilterBar.vue` |
| 删除 | `components/QuickFilterChips.vue` |
| 删除 | `components/RQL/RQLInput.vue` |
| 删除 | `components/RQL/RQLHistory.vue` |
| 删除 | `components/RQL/index.ts` |
| 删除 | `composables/useIssueFilters.ts` |
| 删除 | `composables/useRQL.ts` |
| 删除 | `utils/rql/types.ts` |
| 修改 | `views/Project.vue` |
| 简化 | `IssueList.vue`, `IssueKanban.vue`, `IssueTreeView.vue` |
| 修改 | `locales/zh-CN.json`, `locales/en-US.json` |

---

## 3. 类型设计（v2 — 对标主流平台）

### 3.1 `types/filters.ts`

```typescript
export interface FilterCondition {
  field: string           // 字段名
  operator: string        // 操作符：is, is not, is any of, is not any of, contains, does not contain, is empty, is not empty, before, after, before or on, after or on, between, not between, >, <, >=, <=
  value: any             // 原始值
  displayValue: string   // 显示值
}

export interface FilterField {
  key: string            // 后端字段名
  labelKey: string       // i18n key
  type: 'select' | 'multi' | 'date' | 'text' | 'number' | 'date_range'
  operators: string[]    // 可用操作符列表
}

export interface FilterHistoryItem {
  id: string
  timestamp: number
  filters: FilterCondition[]
  rql: string
}
```

### 3.2 字段定义（对标主流平台）

```typescript
export const FILTER_FIELDS: FilterField[] = [
  // ── Select / Multi ──
  { key: 'title',        labelKey: 'filter.fieldTitle',       type: 'text',   operators: ['is', 'is not', 'contains', 'does not contain'] },
  { key: 'state_id',     labelKey: 'filter.fieldState',       type: 'select', operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'state_group',  labelKey: 'filter.fieldStateGroup',  type: 'select', operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'priority',     labelKey: 'filter.fieldPriority',    type: 'select', operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'assignee_id',  labelKey: 'filter.fieldAssignee',    type: 'select', operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty'] },
  { key: 'label',        labelKey: 'filter.fieldLabel',       type: 'multi',  operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty'] },
  { key: 'cycle_id',     labelKey: 'filter.fieldCycle',       type: 'select', operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty'] },
  { key: 'module_id',    labelKey: 'filter.fieldModule',      type: 'select', operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty'] },
  { key: 'type_id',      labelKey: 'filter.fieldType',        type: 'select', operators: ['is', 'is any of', 'is not', 'is not any of'] },
  // ── Date ──
  { key: 'start_date',   labelKey: 'filter.fieldStartDate',   type: 'date',   operators: ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty'] },
  { key: 'target_date',  labelKey: 'filter.fieldTargetDate',  type: 'date',   operators: ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty'] },
  { key: 'created_at',   labelKey: 'filter.fieldCreatedAt',   type: 'date',   operators: ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty'] },
]
```

### 3.3 操作符 → RQL 映射

| 操作符 | RQL 语法 | 示例 |
|---|---|---|
| `is` | `field = value` | `state_id = 3` |
| `is not` | `field != value` | `state_id != 3` |
| `is any of` | `field IN [v1, v2]` | `state_id IN [3, 5]` |
| `is not any of` | `field NOT IN [v1, v2]` | `state_id NOT IN [3, 5]` |
| `contains` | `field ~ "value"` | `title ~ "bug"` |
| `does not contain` | `field !~ "value"` | `title !~ "deprecated"` |
| `is empty` | `field IS EMPTY` | `assignee_id IS EMPTY` |
| `is not empty` | `field IS NOT EMPTY` | `assignee_id IS NOT EMPTY` |
| `before` | `field < "date"` | `target_date < "2025-07-01"` |
| `after` | `field > "date"` | `target_date > "2025-06-01"` |
| `before or on` | `field <= "date"` | `target_date <= "2025-06-30"` |
| `after or on` | `field >= "date"` | `target_date >= "2025-06-01"` |
| `between` | `field >= "d1" AND field <= "d2"` | `target_date >= "2025-06-01" AND target_date <= "2025-06-30"` |
| `not between` | `field < "d1" OR field > "d2"` | `target_date < "2025-06-01" OR target_date > "2025-06-30"` |

---

## 4. Composable 设计（v2 — 即时更新）

### 4.1 `useFilters.ts`

```typescript
export function useFilters(context: FilterContext) {
  const state = reactive({
    filters: [] as FilterCondition[],
  })

  // Computed
  const rql = computed(() => buildRQL(state.filters))
  const activeFilterCount = computed(() => state.filters.length)
  const isEmpty = computed(() => activeFilterCount.value === 0)

  // Filter CRUD — 即时生效，无确认步骤
  function addFilter(field: string, operator: string, value: any): void {
    const displayValue = formatDisplayValue(field, operator, value, context)
    state.filters.push({ field, operator, value, displayValue })
    // filters 变化 → rql computed 更新 → watcher 自动触发筛选
  }

  function removeFilter(index: number): void {
    state.filters.splice(index, 1)
  }

  function updateFilter(index: number, updates: Partial<FilterCondition>): void {
    const cond = state.filters[index]
    Object.assign(cond, updates)
    if (updates.value !== undefined) {
      cond.displayValue = formatDisplayValue(cond.field, cond.operator, cond.value, context)
    }
  }

  function clearAll(): void {
    state.filters = []
  }

  // Serialization
  function toRQL(): string { ... }
  function toQueryParams(): Record<string, any> { ... }

  // History (localStorage)
  function saveHistory(): void { ... }
  function getHistory(): FilterHistoryItem[] { ... }
  function clearHistory(): void { ... }

  return { state, rql, activeFilterCount, isEmpty,
    addFilter, removeFilter, updateFilter, clearAll,
    toRQL, toQueryParams, saveHistory, getHistory, clearHistory
  }
}

export function injectFilters(): ReturnType<typeof useFilters>
```

**即时更新机制**：
```
用户点击字段 → 选操作符 → 选值
  → 值选定后立即 addFilter()
  → filters[] 变化 → rql computed 重新计算
  → Project.vue watch(rql) → issueApi.listIssues({ rql })
  → 视图自动刷新
```

---

## 5. 组件设计（v2 — 对标主流平台）

### 5.1 FilterBar.vue — UI 布局（对标主流平台）

```
┌──────────────────────────────────────────────────────────────────────────┐
│ 🔽 Filters   [Title contains "bug" ×] [State is any of [Todo,In Prog ×]] │ ← Row 1: 触发按钮 + 芯片行
│              [+ Add filter ▾]                     [Clear all]             │
│                                                                          │
│ ── 点击芯片弹出内联编辑: ──                                              │
│   [Title ▼] [contains ▼] [输入值________]   ← 值变更即时生效，无确认按钮  │
│                                                                          │
│ ── RQL 高级区（可折叠）──                                                │
│   title ~ "bug" AND state_id IN [1, 3]     [Copy] [Apply to filters]    │
├──────────────────────────────────────────────────────────────────────────┤
│ [📋 List] [📌 Kanban] [🌳 Tree] [📅 Calendar] [📊 Gantt]  [💾 Save view] │
└──────────────────────────────────────────────────────────────────────────┘
```

### 5.2 关键设计变更

| 变更 | 旧设计 | v2 对标主流平台 |
|---|---|---|
| **触发方式** | 搜索框 + 预设按钮 + "+ 筛选" | Filters 按钮 → 字段下拉（与主流平台完全一致） |
| **操作确认** | 点击确认/取消 | **即时更新**：值选定即生效 |
| **标题筛选** | 独立搜索框（name ~ "text"） | **Title 字段**：is/contains/does not contain |
| **操作符** | `=` / `!=` / `~` | **语义标签**：is / is not / contains / is any of |
| **日期操作符** | `>=` / `<=` / `>` | before / after / before or on / after or on / between |
| **State Group** | 无 | Backlog / Started / Completed |
| **快速预设** | 我的/未分配/高优先级/今日/即将到期 | **已移除**，用 Saved Views 替代 |

### 5.3 交互流程

```
1. 点击 Filters 按钮（或芯片的 "+ Add filter"）
2. 下拉显示所有字段：Title / State / State Group / Priority / Assignee / Label / Cycle / Module / Type / Start Date / Target Date / Created At
3. 选择字段 → 自动展开内联选择行：
   - Select 字段：操作符下拉 + 值下拉/多选
   - Text 字段：操作符下拉 + 输入框
   - Date 字段：操作符下拉 + 日期选择器（between 时两个日期）
4. 选择值后**即时生效**，无需额外确认
5. 芯片右下角 × 移除单个条件
6. Clear all 清除全部
7. RQL 区可手动编辑 → Apply 后解析为芯片
```

### 5.4 State Group 定义

| Group | 包含的 State |
|---|---|
| Backlog | state.type === 'backlog' |
| Unstarted | state.type === 'unstarted' |
| Started | state.type === 'started' |
| Completed | state.type === 'completed' |
| Cancelled | state.type === 'cancelled' |

---

## 6. i18n Keys — `filter.*` 命名空间（v2）

| Key | zh-CN | en-US |
|---|---|---|
| `filterButton` | Filters | Filters |
| `addFilter` | Add filter | Add filter |
| `clearAll` | Clear all | Clear all |
| `rqlToggle` | RQL | RQL |
| `rqlPlaceholder` | Enter RQL query... | Enter RQL query... |
| `rqlApply` | Apply to filters | Apply to filters |
| `rqlCopy` | Copy | Copy |
| `fieldTitle` | Title | Title |
| `fieldState` | State | State |
| `fieldStateGroup` | State Group | State Group |
| `fieldPriority` | Priority | Priority |
| `fieldAssignee` | Assignee | Assignee |
| `fieldLabel` | Label | Label |
| `fieldCycle` | Cycle | Cycle |
| `fieldModule` | Module | Module |
| `fieldType` | Type | Type |
| `fieldStartDate` | Start date | Start date |
| `fieldTargetDate` | Due date | Due date |
| `fieldCreatedAt` | Created at | Created at |
| `opIs` | is | is |
| `opIsNot` | is not | is not |
| `opIsAnyOf` | is any of | is any of |
| `opIsNotAnyOf` | is not any of | is not any of |
| `opContains` | contains | contains |
| `opDoesNotContain` | does not contain | does not contain |
| `opIsEmpty` | is empty | is empty |
| `opIsNotEmpty` | is not empty | is not empty |
| `opBefore` | before | before |
| `opAfter` | after | after |
| `opBeforeOrOn` | before or on | before or on |
| `opAfterOrOn` | after or on | after or on |
| `opBetween` | between | between |
| `opNotBetween` | not between | not between |
| `selectValue` | Select value | Select value |
| `enterValue` | Enter value | Enter value |
| `saveView` | Save view | Save view |
| `noFilters` | No filters applied | No filters applied |
| `stateGroupBacklog` | Backlog | Backlog |
| `stateGroupUnstarted` | Unstarted | Unstarted |
| `stateGroupStarted` | Started | Started |
| `stateGroupCompleted` | Completed | Completed |
| `stateGroupCancelled` | Cancelled | Cancelled |

---

## 7. 实现步骤

### Phase 1: 基础设施
1. 新建 `types/filters.ts`
2. 新建 `composables/useFilters.ts`
3. 新增 i18n keys

### Phase 2: FilterBar 组件
4. 新建 `components/FilterBar.vue`

### Phase 3: 集成
5. 修改 `views/Project.vue`
6. 简化 `IssueList.vue` / `IssueKanban.vue` / `IssueTreeView.vue`

### Phase 4: 清理
7. 删除 8 个旧文件

### Phase 5: 测试
8. `npx vue-tsc --noEmit` + `npx vite build` + 功能回归

---

## 8. 验证标准

- [ ] TypeScript 零错误
- [ ] Vite 构建成功
- [ ] 操作符使用语义标签（is/contains/is any of）
- [ ] 条件变更即时生效（无确认步骤）
- [ ] Title 作为正式筛选字段
- [ ] State Group 可用
- [ ] 日期支持 before/after/after or on/before or on/between
- [ ] 无快速预设按钮残留
- [ ] RQL ↔ 芯片双向同步
- [ ] 中英文即时切换
- [ ] 旧文件全部删除