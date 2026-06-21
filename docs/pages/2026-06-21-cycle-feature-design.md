# Cycle（周期）管理功能设计文档

**版本**: v1.0  
**创建日期**: 2026-06-21  
**状态**: 已确认  

---

## 1. 概述

参考 Plane 周期管理特性，为本项目（ReqManPy）实现完整的周期（Cycle/Sprint）管理功能。周期是项目内的迭代管理单元，用于组织、追踪和度量一段时期内的 Issue 交付进度。

### 1.1 功能范围

本次实现范围：**基础 CRUD + 状态流转 + 进度分析**

- 周期的创建、编辑、删除、列表
- 周期状态流转：upcoming → active → completed / cancelled
- Issue 与周期的关联管理
- 进度统计、状态分布分析
- 燃尽图（简化版 - 实时计算）

### 1.2 不包含

- 周期结束时自动转移未完成 Issue（自动化）
- 周期版本管理（versioning）
- 多周期并行对比分析
- 每日快照存储

---

## 2. 技术栈

| 层 | 技术 |
|----|------|
| 后端框架 | Go + Gin |
| ORM | GORM |
| 数据库 | PostgreSQL |
| 前端框架 | Vue 3 + TypeScript |
| 状态管理 | Pinia |
| 样式 | Tailwind CSS |

---

## 3. 后端 API 设计

所有接口基于 Gin 路由，挂载在 `api/v1` 下，需 JWT 认证。

### 3.1 路由清单

```
# CRUD
POST   /api/v1/projects/:projectId/cycles         创建周期
GET    /api/v1/projects/:projectId/cycles          列表（支持status筛选、分页）
GET    /api/v1/cycles/:cycleId                     详情
PUT    /api/v1/cycles/:cycleId                     更新
DELETE /api/v1/cycles/:cycleId                     软删除

# 状态流转
POST   /api/v1/cycles/:cycleId/start               upcoming → active
POST   /api/v1/cycles/:cycleId/end                 active → completed
POST   /api/v1/cycles/:cycleId/cancel              取消周期

# Issue 关联
POST   /api/v1/cycles/:cycleId/issues              ?issue_id=N 添加Issue
DELETE /api/v1/cycles/:cycleId/issues/:issueId     从周期移除Issue
GET    /api/v1/cycles/:cycleId/issues              获取周期内Issue（支持state_id/priority筛选）

# 分析统计
GET    /api/v1/cycles/:cycleId/progress            进度
GET    /api/v1/cycles/:cycleId/statistics          详细统计（进度+优先级分布+日期范围）
GET    /api/v1/cycles/:cycleId/burndown            燃尽图数据
```

### 3.2 关键设计决策

- **Cycle ID 顶层路由**: `/cycles/:cycleId` 不挂在 `projects/:projectId/` 下，由 service 层做权限校验
- **Status 运行时推断**: 不存 `status` 字段到数据库，而是根据时间戳计算：
  - `CancelledAt != nil` → cancelled
  - `CompletedAt != nil` → completed
  - `today < StartDate` → upcoming
  - `today >= StartDate && CompletedAt == nil` → active
- **软删除**: 通过 `DeletedAt` 实现，不物理删除数据

### 3.3 API 请求/响应格式

#### 创建周期

```
POST /api/v1/projects/:projectId/cycles?workspace_id=N
{
  "name": "Sprint 1",
  "description": "第一轮迭代",
  "start_date": "2026-06-22T00:00:00Z",
  "end_date": "2026-07-05T00:00:00Z",
  "timezone": "Asia/Shanghai"
}
```

#### 周期列表响应

```json
{
  "items": [{
    "id": 1,
    "name": "Sprint 1",
    "description": "...",
    "status": "active",
    "progress": 45.5,
    "total_issues": 10,
    "completed_issues": 5,
    "start_date": "2026-06-22",
    "end_date": "2026-07-05",
    "project_id": 1,
    "workspace_id": 1,
    "created_at": "...",
    "updated_at": "..."
  }],
  "total": 3,
  "limit": 20,
  "offset": 0
}
```

#### 进度统计响应

```json
{
  "cycle_id": 1,
  "cycle_name": "Sprint 1",
  "total_issues": 10,
  "completed_issues": 5,
  "progress": 50.0,
  "state_breakdown": [
    { "state": "待处理", "group": "backlog", "count": 2 },
    { "state": "进行中", "group": "started", "count": 3 },
    { "state": "已完成", "group": "done", "count": 5 }
  ]
}
```

#### 燃尽图响应

```json
{
  "cycle_id": 1,
  "cycle_name": "Sprint 1",
  "start_date": "2026-06-22",
  "end_date": "2026-07-05",
  "total_issues": 10,
  "total_days": 13,
  "days_elapsed": 5,
  "ideal_daily_burn": 0.77,
  "ideal_remaining": 6.15,
  "actual_completed": 5,
  "actual_remaining": 5,
  "is_on_track": true
}
```

---

## 4. 数据模型

### 4.1 Cycle 模型

```go
type Cycle struct {
    BaseModel

    Name        string     `gorm:"size:255;not null" json:"name"`
    Description *string    `gorm:"size:1000" json:"description"`
    StartDate   time.Time  `gorm:"type:date;not null" json:"start_date"`
    EndDate     *time.Time `gorm:"type:date" json:"end_date"`
    CompletedAt *time.Time `json:"completed_at"`
    CancelledAt *time.Time `json:"cancelled_at"`
    ProjectID   uint64     `gorm:"not null;index" json:"project_id"`
    WorkspaceID uint64     `gorm:"not null" json:"workspace_id"`

    // Relationships
    Project    Project      `gorm:"foreignKey:ProjectID" json:"-"`
    IssueLinks []IssueCycle `gorm:"foreignKey:CycleID" json:"-"`
}
```

### 4.2 IssueCycle 关联表（已存在）

```go
type IssueCycle struct {
    IssueID uint64 `gorm:"primaryKey;autoIncrement:false" json:"issue_id"`
    CycleID uint64 `gorm:"primaryKey;autoIncrement:false" json:"cycle_id"`
}
```

一个 Issue 只能属于一个活跃周期（业务逻辑保证，非 DB 约束）。

---

## 5. Service 层设计

```go
type CycleService struct { db *gorm.DB }

// CRUD
func Create(workspaceID, userID uint64, req *CycleCreate) (*Cycle, error)
func GetByID(cycleID uint64) (*Cycle, error)
func ListByProject(projectID uint64, status string, limit, offset int) ([]Cycle, int64, error)
func Update(cycleID, userID uint64, req *CycleUpdate) (*Cycle, error)
func Delete(cycleID uint64) error

// Status transitions
func Start(cycleID, userID uint64) (*Cycle, error)
func End(cycleID, userID uint64) (*Cycle, error)
func Cancel(cycleID, userID uint64) (*Cycle, error)

// Issue association
func AddIssue(cycleID, issueID uint64) error
func RemoveIssue(cycleID, issueID uint64) error
func ListIssues(cycleID uint64, stateID *uint64, priority string, limit, offset int) ([]Issue, int64, error)

// Analysis
func GetProgress(cycleID uint64) (*CycleProgress, error)
func GetStatistics(cycleID uint64) (*CycleStatistics, error)
func GetBurndown(cycleID uint64) (*BurndownData, error)
```

### 5.1 关键业务逻辑

- **Start**: 校验状态不是已完成/已取消，设置 start_date 为当前日期（如果未设置），completed_at 和 cancelled_at 为 nil
- **End**: 设置 completed_at = now，触发进度快照生成
- **Cancel**: 设置 cancelled_at = now，保留 Issue 关联（不自动移除）
- **AddIssue**: 校验 Issue 和 Cycle 属于同一 project，一个 Issue 只能在一个活跃 Cycle 中
- **GetProgress**: 实时查询 State.group="done" 的 Issue 数量 / 总量
- **GetBurndown**: 基于 start_date 到 end_date 的理想线 vs 实际完成数，实时计算

---

## 6. 前端设计

### 6.1 路由新增

```
/workspaces/:workspaceId/projects/:projectId/cycles/new    CycleCreate（创建向导）
/workspaces/:workspaceId/projects/:projectId/cycles/:cycleId  CycleDetail（详情页）
```

### 6.2 组件架构

```
Project.vue
  ├─ Tab "工作项管理" → IssueList/IssueKanban + IssueDetailPanel（已有）
  ├─ Tab "周期"       → CycleList + CycleDetailPanel（新增Panel）
  └─ Tab "模块"       → ModuleList（已有）

CycleCreate.vue（独立路由，创建向导）
  ├─ 步骤1：基本信息（名称、描述、起止日期）
  ├─ 步骤2：选择 Issue（从 backlog 中挑选加入周期）
  └─ 预览 + 提交

CycleDetail.vue（独立路由，完整详情页）
  ├─ 头部：名称、状态标签、日期范围、操作按钮（开始/结束/取消/编辑/删除）
  ├─ 进度统计卡片区域：总数/完成数/进度%/状态分布
  ├─ 燃尽图区域：理想线 + 实际线
  └─ Issue 列表：支持筛选（状态/优先级）、拖入/移出
```

### 6.3 Pinia Store (stores/cycle.ts)

```typescript
export const useCycleStore = defineStore('cycle', () => {
  // State
  const cycles = ref<CycleResponse[]>([])
  const currentCycle = ref<CycleResponse | null>(null)
  const progress = ref<CycleProgress | null>(null)
  const statistics = ref<CycleStatistics | null>(null)
  const burndown = ref<BurndownData | null>(null)
  const issues = ref<IssueResponse[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Actions
  async function fetchCycles(projectId, status?) { ... }
  async function fetchCycle(cycleId) { ... }
  async function createCycle(data: CycleCreate) { ... }
  async function updateCycle(cycleId, data: CycleUpdate) { ... }
  async function deleteCycle(cycleId) { ... }
  async function startCycle(cycleId) { ... }
  async function endCycle(cycleId) { ... }
  async function cancelCycle(cycleId) { ... }
  async function addIssueToCycle(cycleId, issueId) { ... }
  async function removeIssueFromCycle(cycleId, issueId) { ... }
  async function fetchCycleIssues(cycleId, filters?) { ... }
  async function fetchProgress(cycleId) { ... }
  async function fetchStatistics(cycleId) { ... }
  async function fetchBurndown(cycleId) { ... }
})
```

### 6.4 复用/调整现有组件

| 组件 | 操作 | 说明 |
|------|------|------|
| `CycleCard.vue` | 微调 | 适配新的 API 响应格式，status 改为运行时推断 |
| `CycleList.vue` | 微调 | 改用 store 驱数据，添加分页支持 |
| `types/cycle.ts` | 已完善 | 基本对齐，微调字段名 |
| `api/cycle.ts` | 需调整 | URL 路径对齐 Go 后端路由 |

### 6.5 新增文件清单

| 文件 | 说明 |
|------|------|
| `stores/cycle.ts` | Pinia store |
| `views/CycleCreate.vue` | 创建向导（两步） |
| `views/CycleDetail.vue` | 完整详情页 |
| `components/CycleDetailPanel.vue` | 侧滑面板（Project Tab 内使用） |
| `components/CycleBurndownChart.vue` | 燃尽图（Canvas/SVG） |
| `components/CycleProgressCard.vue` | 进度统计卡片组 |

---

## 7. 文件变更清单

### Go 后端

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/model/cycle.go` | 修改 | 添加 CancelledAt 字段，删除 status 字段 |
| `internal/handler/cycle_handler.go` | 重写 | 实现全部 handler 方法 |
| `internal/service/cycle_service.go` | 重写 | 实现全部业务逻辑 |
| `internal/dto/request/cycle.go` | 新建 | CycleCreateRequest, CycleUpdateRequest |
| `internal/dto/response/cycle.go` | 新建 | CycleResponse, CycleProgress, CycleStatistics, BurndownData |
| `internal/router/router.go` | 修改 | 添加 cycle 相关路由 |
| `internal/handler/issue_handler.go` | 修改 | 已有的 SetCycle/RemoveCycle 需确认一致性 |

### Vue 前端

| 文件 | 操作 | 说明 |
|------|------|------|
| `stores/cycle.ts` | 新建 | Pinia store |
| `views/CycleCreate.vue` | 新建 | 创建向导 |
| `views/CycleDetail.vue` | 新建 | 详情页 |
| `components/CycleDetailPanel.vue` | 新建 | 侧滑面板 |
| `components/CycleBurndownChart.vue` | 新建 | 燃尽图 |
| `components/CycleProgressCard.vue` | 新建 | 进度卡片 |
| `components/CycleCard.vue` | 微调 | 适配新 API |
| `components/CycleList.vue` | 微调 | store 驱动，分页 |
| `api/cycle.ts` | 修改 | 对齐后端路由 |
| `types/cycle.ts` | 微调 | 对齐后端字段 |
| `router/index.ts` | 修改 | 新增 2 条路由 |
| `views/Project.vue` | 微调 | 集成 CycleDetailPanel |

---

## 8. 错误处理

### 后端错误码

| 场景 | HTTP | 错误信息 |
|------|------|----------|
| 周期不存在 | 404 | "Cycle not found" |
| 项目不存在 | 404 | "Project not found" |
| Issue不存在 | 404 | "Issue not found" |
| 日期范围无效 | 400 | "Start date must be before end date" |
| 状态不允许操作 | 400 | "Cannot start/end/cancel a completed/cancelled cycle" |
| 重复添加 | 400 | "Issue is already in this cycle" |
| 周期不匹配 | 400 | "Issue does not belong to this cycle's project" |
| 无燃尽图数据 | 400 | "Cycle does not have start and end dates" |

### 前端错误处理

- Store 中统一 catch，设置 `error` state
- 组件中通过 `store.error` 显示 toast 或 inline 错误
- API 调用失败时不阻塞 UI，允许重试

---

**文档路径**: `docs/pages/2026-06-21-cycle-feature-design.md`
