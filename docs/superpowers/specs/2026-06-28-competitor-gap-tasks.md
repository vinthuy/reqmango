---
title: 竞品 Pro — 完整特性差异 GAP 分析 & 分阶段任务表
author: ReqMango
date: 2026-06-28
---

# 竞品 Pro — 完整特性差异 GAP 分析 & 分阶段开发任务

> 基准：主流项目管理平台官方文档 v.s. ReqMango 当前代码  
> 覆盖：Work Item Filters / RQL / Display Options / Views 四模块  
> 策略：四阶段推进，每阶段可独立验证

---

## 一、全量 GAP 总表 (33 项)

| # | 模块 | 特性 | 竞品 | ReqMango 现状 | 优先级 |
|---|------|------|-------|--------------|--------|
| 1 | Filters | 统一筛选入口 | Filters 按钮 → 字段下拉 | 分散在 IssueFilterBar/Kanban/List 各自实现 | 🔴 |
| 2 | Filters | 条件芯片行 | 芯片展示 + × 删除 + 点击编辑 | 部分芯片，编辑需确认 | 🔴 |
| 3 | Filters | 即时更新 | 选完值即刻筛选 | 需点击确认按钮 | 🟠 |
| 4 | Filters | 语义操作符 | `is` / `is not` / `contains` / `is any of` / `is not any of` | `=` / `!=` / `~` / `in` | 🟠 |
| 5 | Filters | 日期操作符 | `before` / `after` / `before or on` / `after or on` / `between` / `not between` / `not before` / `not after` / `not before or on` / `not after or on` | `>=` / `<=` / `>` / `<`（4种） | 🟠 |
| 6 | Filters | Title 筛选字段 | `is` / `is not` / `contains` / `does not contain` | 独立搜索框 (`name ~ "text"`) | 🟠 |
| 7 | Filters | State Group 字段 | Backlog / Unstarted / Started / Completed / Cancelled | 不存在 | 🟠 |
| 8 | Filters | Created by 字段 | `is` / `is any of` / `is not` / `is not any of` | 不存在 | 🟡 |
| 9 | Filters | Mentions 字段 | `is` / `is any of` / `is not` / `is not any of` / `is empty` | 不存在 | 🟡 |
| 10 | Filters | Milestone 字段 | `is` / `is any of` / `is not` / `is not any of` / `is empty` | Cycle 可部分替代 | 🟡 |
| 11 | Filters | Custom Properties | 文本/数字/下拉/布尔/日期/成员/URL 全类型筛选 | 不支持 | 🟢 |
| 12 | Filters | RQL 双向同步 | Filters ↔ RQL 互斥切换 | 各自独立 | 🔴 |

| 13 | PQL | 引导式编辑器 | 键入 → 字段下拉 → 操作符下拉 → 值选择 | 裸文本输入框 | 🟠 |
| 14 | PQL | BETWEEN 语法 | `field BETWEEN "d1" AND "d2"` | 不支持 | 🟠 |
| 15 | PQL | 内置函数 | `isOverdue()` / `hasNoAssignee()` / `hasNoLabel()` / `isTopLevel()` / `isSubWorkItem()` / `hasChildren()` / `hasStartAndDueDates()` | 无 | 🟡 |
| 16 | PQL | ORDER BY | 不支持（Views 管理） | 不支持 | 🟠 |
| 17 | PQL | 语义标记 → RQL 映射 | PQL 标记 (state, priority...) → 内部 RQL 转换 | 直接使用 `state_id`, `priority` | 🟡 |
| 18 | PQL | Workspace 级别搜索 | 跨项目 PQL | 仅单项目 | 🟢 |

| 19 | Display | Properties 列配置 | 13 项可选列，存 Views | IssueList 10 列，存 localStorage | 🔴 |
| 20 | Display | Group By 分组 | 11 种维度，存 Views | Kanban/Gantt 本地 `<select>`（4 种） | 🔴 |
| 21 | Display | Sub-group by 子分组 | 第二层分组（同 Group By 选项） | 不存在 | 🟡 |
| 22 | Display | Order By 排序 | 6 种排序 + Manual 拖拽 | 后端固定 `ORDER BY created_at DESC` | 🔴 |
| 23 | Display | Show empty groups | 开关空分组显示 | 不支持 | 🟡 |
| 24 | Display | Show sub-items | 展开子工作项 | 不支持 | 🟢 |

| 25 | Views | 完整恢复 | 选视图 → 恢复 filters + layout + display + sort | **仅切换 view_type** | 🔴 |
| 26 | Views | 访问级别 UI | Public / Private 切换 (Pro) | IsShared 字段在 DB 无 UI | 🟢 |
| 27 | Views | Lock 锁定 (Pro) | 锁定防编辑 | 不支持 | 🟢 |
| 28 | Views | Publish 发布 (Pro) | 生成公开 URL | 不支持 | 🟢 |
| 29 | Views | Export 导出 (Pro) | CSV / XLSX / JSON | 不支持 | 🟢 |
| 30 | Views | Favorite 收藏 | Star → 侧栏 Favorites | 不支持 | 🟡 |
| 31 | Views | Workspace Views | 跨项目视图 + 4 个内置视图 | 不支持 | 🟢 |
| 32 | Views | Teamspace Views | Teamspace 级别视图 | 不支持 | 🟢 |
| 33 | Views | Duplicate / Delete | 复制 / 删除视图 | ✅ 已实现 | ✅ |

---

## 二、分阶段任务

### Phase 1 — 基础设施重构 (Foundation)

> 目标：统一筛选入口 + Views 恢复链路打通 + Display Options 存储集成  
> 预计文件变更：~20 个  
> 可验证：FilterBar 在所有视图下一致工作，选视图恢复完整状态

| ID | 任务 | 关联 GAP | 工作量 |
|----|------|----------|--------|
| **F1.1** | 新建 `types/filters.ts` — FilterCondition / FilterField / FilterOperator / FilterPreset 类型 | #1,#4 | S |
| **F1.2** | 新建 `composables/useFilters.ts` — 筛选状态 single source of truth + RQL 自动派生 + Provide/Inject | #1,#12 | M |
| **F1.3** | 新建 `components/FilterBar.vue` — 统一筛选栏 UI（Filters 按钮 + 字段下拉 + 芯片行 + RQL 折叠区） | #1,#2 | L |
| **F1.4** | 删除旧文件：IssueFilterBar / QuickFilterChips / RQLInput / RQLHistory / useIssueFilters / useRQL / utils/rql/types | #1,#12 | S |
| **F1.5** | 重构 `Project.vue` — 替换 → FilterBar，移除所有 view props 传递 | #1 | M |
| **F1.6** | 简化 `IssueList.vue` / `IssueKanban.vue` / `IssueTreeView.vue` — 删除内置筛选代码 | #1 | M |
| **F1.7** | **修复 Views 完整恢复链路** — `handleViewSelect` → 应用 filters / viewType / columns / groupBy / sortConfig | #25 | L |
| **F1.8** | Display Properties 列配置接入 SavedView — IssueList 列从 localStorage 迁移到 `SavedView.columns` 读写 | #19 | M |
| **F1.9** | i18n 新增 `filter.*` 命名空间（28 keys）× zh-CN + en-US | #4,#6 | S |
| **F1.10** | TypeScript 检查 + Vite build 验证 | — | S |

**Phase 1 输出：FilterBar 统一可用 + 选视图完整恢复状态 + Display Properties 存储到 Views**

---

### Phase 2 — 核心体验升级 (Feature Parity)

> 目标：语义操作符 + 即时更新 + Title / State Group / Date operators 补齐  
> 预计文件变更：~10 个  
> 可验证：操作符使用语义标签，条件变更即时刷数据

| ID | 任务 | 关联 GAP | 工作量 |
|----|------|----------|--------|
| **F2.1** | FilterBar 操作符全部改为语义标签 — `is` / `is not` / `contains` / `is any of` / `is not any of` / `is empty` / `is not empty` | #4 | M |
| **F2.2** | 日期操作符补齐 — `before` / `after` / `before or on` / `after or on` / `between` / `not between` → RQL 映射 | #5 | M |
| **F2.3** | 后端 RQL executor 新增 BETWEEN / NOT BETWEEN 支持 | #5,#14 | M |
| **F2.4** | **即时更新** — 条件变更直接 addFilter()，去确认/取消步骤 | #3 | M |
| **F2.5** | Title 作为正式筛选字段 — 新增 `title` 字段 + 对应的 4 个操作符 + RQL 映射 | #6 | S |
| **F2.6** | State Group 筛选字段 — `state_group` (Backlog/Unstarted/Started/Completed/Cancelled) + RQL 映射 | #7 | M |
| **F2.7** | Group By 选择器提升至 FilterBar/Display 区域 — 替换 Kanban/Gantt 各自本地 select | #20 | M |
| **F2.8** | Order By 实现 — FilterBar 中排序下拉 (Last created / Last updated / Priority / Start date / Due date) + 后端 RQL 支持 ORDER BY | #22 | L |
| **F2.9** | i18n 新增操作符 + State Group keys | — | S |
| **F2.10** | TypeScript + Vite build 验证 | — | S |

**Phase 2 输出：筛选体验与主流平台对齐，操作符语义化、即时筛选、Date/Title/StateGroup 完整**

---

### Phase 3 — 高级特性 (Advanced)

> 目标：PQL 引导式编辑器 + 子分组 + 内置函数 + 视图收藏  
> 预计文件变更：~8 个  
> 可验证：PQL 编辑器引导式补全，Views 支持子分组和收藏

| ID | 任务 | 关联 GAP | 工作量 |
|----|------|----------|--------|
| **F3.1** | PQL 引导式编辑器 — 键入时字段下拉 → 操作符下拉 → 值建议（复用 FilterBar 数据源） | #13 | L |
| **F3.2** | PQL 内置函数 — `isOverdue` / `hasNoAssignee` / `hasNoLabel` / `isTopLevel` / `isSubWorkItem` / `hasChildren` | #15 | M |
| **F3.3** | PQL/RQL 语义标记标准化 — `state` → `state_id`，`assignee` → `assignee_id`，对标主流平台字段命名 | #17 | S |
| **F3.4** | Sub-group by 子分组 — Display Options 中第二层分组下拉 + savedView.sub_group_by 存储 | #21 | M |
| **F3.5** | Show empty groups — Display Options 开关 + 各组视图适配 | #23 | S |
| **F3.6** | Created by / Mentions 筛选字段 — `created_by` + `mention` 字段 + 操作符 + RQL 映射 | #8,#9 | M |
| **F3.7** | Favorite 收藏 — View 星标 → 侧栏 Favorites 区 | #30 | S |
| **F3.8** | i18n 补充 | — | S |
| **F3.9** | 测试验证 | — | S |

**Phase 3 输出：PQL 编辑器对标主流平台 Pro，Display Options 完整（Group/Sub-group/Empty），Views 收藏**

---

### Phase 4 — Pro 特性 + 跨项目 (Enterprise)

> 目标：锁定/发布/导出视图 + Workspace Views + Custom Properties 筛选  
> 预计文件变更：~12 个  
> 可验证：视图锁定防编辑、公开链接可访问、跨项目视图工作

| ID | 任务 | 关联 GAP | 工作量 |
|----|------|----------|--------|
| **F4.1** | View 访问级别 UI — Public / Private 切换 + 图标 | #26 | S |
| **F4.2** | Lock View — 视图锁定防编辑 + 只读提示 + 解锁权限 | #27 | M |
| **F4.3** | Publish View — 生成公开 URL + Comments/Reactions/Votes 开关 + Layouts 限制 | #28 | L |
| **F4.4** | Export View — CSV / XLSX / JSON 导出 + 异步任务 + 下载 | #29 | L |
| **F4.5** | Workspace Views — 跨项目视图列表 + 4 个内置视图 (All/Assigned to Me/Created by Me/Subscribed) | #31 | L |
| **F4.6** | Milestone 筛选字段 — 对标主流平台 milestone → 映射到 ReqMango cycle（别名） | #10 | S |
| **F4.7** | Custom Properties 筛选 — 动态字段 + 类型感知操作符 | #11 | L |
| **F4.8** | Show sub-items — 列表/看板中展开子工作项 | #24 | M |
| **F4.9** | 测试验证 | — | S |

**Phase 4 输出：Pro 全特性就绪，Workspace Views 跨项目可用，Custom Properties 筛选**

---

## 三、总工时估算

| Phase | 模块覆盖 | 文件变更 | 估算 |
|-------|---------|---------|------|
| **Phase 1** Foundation | Filters + Views + Display | ~20 files | ████████ |
| **Phase 2** Core Enhancement | Filters + RQL + Display | ~12 files | ████████ |
| **Phase 3** Advanced | PQL + Display + Views | ~10 files | ██████ |
| **Phase 4** Pro/Enterprise | Views Pro + Custom | ~15 files | ████████ |
| **总计** | 覆盖 33/33 GAP | ~55 files | — |

---

## 四、每个 Phase 验收标准

### Phase 1
- [ ] FilterBar 在 List/Kanban/Tree/Calendar/Gantt 下一致显示
- [ ] 选择视图 → 恢复 filters、viewType、columns、groupBy、sortConfig
- [ ] Display 列配置存取 SavedView.columns 而非 localStorage
- [ ] 删除全部旧筛选文件，无 import 残留
- [ ] 中英文切换正常
- [ ] TypeScript + Vite build 零 error

### Phase 2
- [ ] 操作符全部语义标签 (is/contains/between)
- [ ] 日期支持 before/after/before or on/after or on/between/not between
- [ ] 添加条件即时生效，无确认步骤
- [ ] Title as filter field 工作
- [ ] State Group 工作
- [ ] Group By / Order By 选择器在 FilterBar 中
- [ ] RQL BETWEEN 后端支持

### Phase 3
- [ ] PQL 编辑器键入引导：字段下拉 → 操作符下拉 → 值建议
- [ ] `isOverdue()` 等 6 个内置函数可用
- [ ] Sub-group by 工作
- [ ] Show empty groups 工作
- [ ] Created by / Mentions 筛选工作
- [ ] Favorite 收藏视图

### Phase 4
- [ ] View Public/Private 切换
- [ ] Lock 视图只读
- [ ] Publish 生成公开 URL
- [ ] Export CSV/XLSX/JSON
- [ ] Workspace Views 列表 + 4 内置视图
- [ ] Custom Properties 动态筛选