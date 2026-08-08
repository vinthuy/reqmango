# Squads 多 Agent 协作增强 — 设计文档

> **日期**: 2026-08-07  
> **状态**: Draft  
> **关联**: 方向B — 在方向A（Tool Calling 引擎加固）完成后实施

---

## 1. 现状分析

### 1.1 已有基线

| 层级 | 状态 | 说明 |
|------|------|------|
| Model | ✅ 完整 | Squad / SquadMember / SquadTask / SquadExecution 4 张表，AutoMigrate 建表 |
| Service | ✅ 基本完整 | CRUD + StartExecution（4 阶段流水线：分解→执行→审核→聚合） |
| Handler | ✅ 完整 | 9 个 HTTP 端点（CRUD + 成员管理 + 执行管理） |
| Router | ✅ 完整 | 10 条路由注册 + aiAgentExecutorAdapter 适配 |
| DTO | ✅ 完整 | Request/Response 均齐全 |
| 前端 API | ✅ 完整 | 10 个函数封装（squad.ts） |
| 前端页面 | ⚠️ 骨架 | 仅列表 + 创建/编辑弹窗，缺少详情、成员管理、执行触发、执行记录 |

### 1.2 关键缺口

| 编号 | 缺口 | 影响 |
|------|------|------|
| G1 | StartExecution **同步阻塞** | HTTP 请求在 Agent 执行期间挂起，无法并发/取消 |
| G2 | SquadTask.AgentTaskID **未填充** | SquadTask 与 AgentTask 未打通，无法在 TaskBoard 看到子任务 |
| G3 | **无 SSE 进度推送** | 前端无法实时看到执行阶段和子任务状态 |
| G4 | **无取消机制** | 执行启动后无法中止 |
| G5 | **无重试机制** | 子任务失败后无自动重试 |
| G6 | **无 Workspace 权限校验** | SquadService 缺少 workspace 成员检查（方向A 已在 ToolService 建立模式） |
| G7 | **前端功能缺失** | 无详情页、成员管理 UI、执行触发按钮、执行历史/日志查看 |
| G8 | **无执行状态机** | Status 字符串自由定义，无状态转换约束 |
| G9 | **无审计日志** | 执行操作未记录到统一审计系统 |

---

## 2. 设计目标

1. **异步执行 + SSE 推送**：StartExecution 立即返回 executionID，后台 goroutine 执行，通过 SSE 实时推送进度
2. **SquadTask ↔ AgentTask 关联**：子任务执行时创建 AgentTask 记录，实现跨模块可见
3. **取消支持**：通过 context 取消信号，优雅终止执行中的子任务
4. **重试机制**：子任务失败后自动重试（可配置 max_retries）
5. **权限校验**：复用方向A 的三步权限模式
6. **前端增强**：SquadDetail 页面（成员管理 + 执行触发 + 执行日志时间线）

---

## 3. 架构设计

### 3.1 异步执行模型

```
Client                    Server                     Goroutine Pool
  │                         │                           │
  ├─ POST /executions ──────►                           │
  │◄── 202 { execution_id } ─┤                           │
  │                         │                           │
  │                         ├─ enqueue(execution) ──────►│
  │                         │                           │ phase 1: decompose
  │◄──── SSE: phase_start ──┤◄──────────────────────────┤
  │◄──── SSE: subtask_start ┤◄──────────────────────────┤ phase 2: execute
  │◄──── SSE: subtask_done ─┤◄──────────────────────────┤
  │◄──── SSE: phase_start ──┤◄──────────────────────────┤ phase 3: review
  │◄──── SSE: subtask_done ─┤◄──────────────────────────┤
  │◄──── SSE: completed ────┤◄──────────────────────────┤ phase 4: aggregate
```

### 3.2 状态机

```
                    ┌──────────┐
                    │ pending  │
                    └────┬─────┘
                         │ StartExecution
                         ▼
                    ┌──────────┐
                    │ running  │◄─────── cancel signal ──┐
                    └────┬─────┘                          │
            ┌────────────┼────────────┐                   │
            ▼            ▼            ▼                   │
     ┌──────────┐ ┌───────────┐ ┌──────────┐             │
     │completed │ │  failed   │ │ cancelled│◄─────────────┘
     └──────────┘ │(all fail) │ └──────────┘
                  └───────────┘
                  ┌───────────┐
                  │ partial_  │ (some fail, some succeed)
                  │ failed    │
                  └───────────┘
```

### 3.3 数据库变更

#### 新增字段：squad_executions 表

```sql
ALTER TABLE squad_executions ADD COLUMN cancelled_at TIMESTAMP NULL;
ALTER TABLE squad_executions ADD COLUMN cancel_reason TEXT DEFAULT '';
```

#### 新增字段：squad_tasks 表

```sql
-- AgentTaskID 已存在但未填充，无需 DDL 变更
-- 新增 retry_count 字段
ALTER TABLE squad_tasks ADD COLUMN retry_count INT DEFAULT 0;
ALTER TABLE squad_tasks ADD COLUMN max_retries INT DEFAULT 2;
ALTER TABLE squad_tasks ADD COLUMN error_message TEXT DEFAULT '';
```

#### 新增字段：squads 表

```sql
ALTER TABLE squads ADD COLUMN config JSONB DEFAULT '{}';
-- config 扩展：{ "max_retries": 2, "timeout_seconds": 300, "concurrency": 1 }
```

### 3.4 SSE 事件扩展

| 事件名 | Payload | 说明 |
|--------|---------|------|
| `squad.execution.started` | `{ execution_id, squad_id, goal }` | 执行启动 |
| `squad.execution.phase_start` | `{ execution_id, phase }` | 阶段开始（decompose/execute/review/aggregate） |
| `squad.execution.subtask_start` | `{ execution_id, task_id, member_id, title }` | 子任务开始 |
| `squad.execution.subtask_progress` | `{ execution_id, task_id, progress }` | 子任务进度更新 |
| `squad.execution.subtask_done` | `{ execution_id, task_id, status, summary }` | 子任务完成 |
| `squad.execution.completed` | `{ execution_id, status, summary }` | 执行结束 |
| `squad.execution.cancelled` | `{ execution_id, reason }` | 执行取消 |

---

## 4. 详细设计

### 4.1 SquadService 改造

#### 4.1.1 StartExecution 异步化

```go
func (s *SquadService) StartExecution(squadID uint64, req request.SquadExecutionStart) (*response.SquadExecutionResponse, error) {
    // 1. 权限校验（复用三步模式）
    // 2. 查找 Squad + Members
    // 3. 创建 SquadExecution（status=pending）
    // 4. 启动 goroutine: go s.executeAsync(execution.ID, squad, req.UserID)
    // 5. 立即返回 execution（status=pending）
}
```

#### 4.1.2 executeAsync 核心流程

```go
func (s *SquadService) executeAsync(executionID uint64, squad model.Squad, userID uint64) {
    ctx, cancel := context.WithCancel(context.Background())
    s.cancelStore.Store(executionID, cancel) // 注册取消句柄
    defer s.cancelStore.Delete(executionID)
    defer recover() // panic 保护

    // Phase 1: Decompose
    s.broadcastPhaseStart(executionID, "decompose")
    subtasks := s.decomposeGoal(ctx, squad, ...)
    
    // Phase 2: Execute
    s.broadcastPhaseStart(executionID, "execute")
    for _, st := range subtasks {
        if ctx.Err() != nil { break } // 被取消
        s.executeSubtask(ctx, executionID, squad, st, ...)
    }
    
    // Phase 3: Review
    s.broadcastPhaseStart(executionID, "review")
    ...
    
    // Phase 4: Aggregate
    s.broadcastPhaseStart(executionID, "aggregate")
    s.aggregateResults(executionID, ...)
    
    // 广播完成事件
    s.broadcastCompleted(executionID, finalStatus)
}
```

#### 4.1.3 CancelExecution

```go
func (s *SquadService) CancelExecution(executionID uint64) error {
    cancel, ok := s.cancelStore.Load(executionID)
    if !ok { return NotFound("execution not running") }
    cancel.(context.CancelFunc)()
    // 更新 execution status = "cancelled"
    // 广播 cancelled 事件
}
```

#### 4.1.4 SquadTask ↔ AgentTask 关联

```go
func (s *SquadService) executeSubtask(ctx context.Context, executionID uint64, squad model.Squad, st subtaskSpec, ...) {
    // 1. 创建 AgentTask（复用 agent_task_service.Create）
    // 2. 创建 SquadTask（填充 AgentTaskID）
    // 3. 调用 agentSvc.DispatchAgent（通过 AgentExecutorInterface）
    // 4. 更新 SquadTask 状态 + 输出
}
```

### 4.2 重试机制

```go
func (s *SquadService) executeSubtaskWithRetry(ctx context.Context, ...) (string, error) {
    maxRetries := 2
    if squad.Config != nil {
        var cfg struct{ MaxRetries int `json:"max_retries"` }
        json.Unmarshal(squad.Config.ToRawMessage(), &cfg)
        if cfg.MaxRetries > 0 { maxRetries = cfg.MaxRetries }
    }
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        if ctx.Err() != nil { return "", ctx.Err() }
        result, err := s.runMemberTask(ctx, ...)
        if err == nil { return result, nil }
        // 更新 retry_count, error_message
        s.broadcastSubtaskProgress(executionID, taskID, map[string]int{"retry": attempt + 1})
    }
    return "", fmt.Errorf("failed after %d retries", maxRetries)
}
```

### 4.3 权限校验（复用三步模式）

```go
func (s *SquadService) checkPermissions(wid, uid uint64) error {
    // Step 1: Workspace membership
    var m model.WorkspaceMember
    if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", wid, uid, true).First(&m).Error; err != nil {
        return common.Forbidden("Workspace member required")
    }
    return nil
}
```

> 注：Squad 操作相对安全（非 dangerous），只需 Step 1。Step 2/3 仅在 Tool Calling 场景需要。

### 4.4 SSE 广播

复用方向A 建立的 `SSE.BroadcastEvent` 机制：

```go
func (s *SquadService) broadcastEvent(event string, data interface{}) {
    SSE.BroadcastEvent(event, data)
}
```

事件结构统一为 `{ execution_id, squad_id, ... }` 格式。

---

## 5. 前端设计

### 5.1 页面路由

| 路径 | 组件 | 说明 |
|------|------|------|
| `/workspace/:slug/agents/squads` | SquadList.vue | 列表页（已有，增强） |
| `/workspace/:slug/agents/squads/:id` | SquadDetail.vue | 详情页（**新建**） |

### 5.2 SquadList.vue 增强

- 创建弹窗增加 `leader_agent_id` 下拉选择
- 列表增加「查看」按钮跳转详情页
- 修复 `saveSquad` 中 workspaceId 类型问题

### 5.3 SquadDetail.vue（新建）

分为 4 个 Tab：

#### Tab 1: 成员管理
- 成员列表（头像 + 角色标签 + 状态）
- 添加成员弹窗（选择 Agent + 角色 + AgentConfig）
- 移除成员（Popconfirm 确认）

#### Tab 2: 执行
- 「启动执行」按钮 + Goal 输入弹窗
- 当前执行状态卡片（如果正在运行）
- 取消执行按钮

#### Tab 3: 执行历史
- 执行记录时间线（按时间倒序）
- 每条记录：状态标签 + Goal + 时间 + 展开查看日志
- SSE 实时更新当前执行

#### Tab 4: 配置
- max_retries、timeout_seconds 等配置项

### 5.4 SSE 监听

```typescript
// useSSE.ts 新增事件
sse.on('squad.execution.started', handleExecutionStarted)
sse.on('squad.execution.phase_start', handlePhaseStart)
sse.on('squad.execution.subtask_start', handleSubtaskStart)
sse.on('squad.execution.subtask_done', handleSubtaskDone)
sse.on('squad.execution.completed', handleCompleted)
sse.on('squad.execution.cancelled', handleCancelled)
```

---

## 6. 实现任务分解

### T1: DB Migration + 模型扩展
- 创建 migration 000023
- SquadExecution 增加 cancelled_at, cancel_reason 字段
- SquadTask 增加 retry_count, max_retries, error_message 字段
- Squad.Config 类型从 json.RawMessage 改为 JSONRawMessage（SQLite 兼容）

### T2: SquadService 异步执行改造
- 添加 cancelStore (sync.Map) 存储 context.CancelFunc
- 改造 StartExecution：创建 execution + 启动 goroutine + 立即返回
- 实现 executeAsync：4 阶段流程 + context 取消检查
- 实现 CancelExecution

### T3: 重试机制 + SquadTask-AgentTask 关联
- 实现 executeSubtaskWithRetry
- executeSubtask 中创建 AgentTask 记录并填充 SquadTask.AgentTaskID
- 权限校验（checkPermissions）

### T4: SSE 广播 + 审计日志
- 7 个 SSE 事件的 broadcast 方法
- 执行操作审计日志

### T5: Handler/Router 更新
- 新增 CancelExecution handler + 路由 `DELETE /:wsParam/squads/:squadId/executions/:executionId`
- StartExecution 改为返回 202

### T6: 前端 SquadList 增强 + SquadDetail 页面
- SquadList 创建弹窗增加 leader_agent_id
- 新建 SquadDetail.vue（4 Tab）
- 路由注册
- i18n 键补充
- useSSE 新增 squad.execution.* 事件监听

### T7: 端到端验证
- go build + go vet
- 单测（CancelExecution、重试、SSE 事件）
- vue-tsc --noEmit
- 手动回归测试

---

## 7. 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| Goroutine 泄漏 | defer recover + context 超时 + cancelStore 清理 |
| 数据竞争 | SquadService 字段用 sync.Map（cancelStore），DB 操作走 GORM 事务 |
| SQLite 兼容 | JSONRawMessage 类型 + migration 使用 ALTER TABLE（SQLite 兼容语法） |
| 前端 SSE 断连 | useSSE 已有自动重连机制，SquadDetail 页面通过 SSE 自动恢复状态 |
| 取消后资源残留 | CancelExecution 将 running 的 SquadTask 标记为 cancelled |