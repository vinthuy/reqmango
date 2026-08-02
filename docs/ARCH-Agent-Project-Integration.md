# Agent-Project Integration 架构设计与详细设计

> 版本：v1.1
> 日期：2026-07-26
> 作者：架构师
> 状态：✅ 设计评审通过

---

## 一、架构设计

### 1.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            Frontend (Vue 3)                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │ AgentMember  │  │ IssueAgent   │  │ Workflow     │  │ NewUser    │ │
│  │ Manager      │  │ Panel        │  │ Designer     │  │ Guide      │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └─────┬──────┘ │
│         │                 │                 │                │          │
│  ┌──────▼─────────────────▼─────────────────▼────────────────▼──────┐  │
│  │                    API Layer (Axios)                              │  │
│  └──────────────────────────────┬───────────────────────────────────┘  │
└─────────────────────────────────┼───────────────────────────────────────┘
                                  │ HTTP
┌─────────────────────────────────▼───────────────────────────────────────┐
│                          Backend (Go + Gin)                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                     Router (router.go)                          │   │
│  │  /projects/:id/agent-members                                    │   │
│  │  /projects/:id/workflows                                        │   │
│  │  /issues/:id/assign-agent                                       │   │
│  │  /issues/:id/ai-decompose                                       │   │
│  └──────────┬──────────────────────────────────────────────────────┘   │
│             │                                                          │
│  ┌──────────▼──────────────────────────────────────────────────────┐   │
│  │                    Handler Layer                                 │   │
│  │  ┌─────────────┐ ┌──────────────┐ ┌──────────────┐             │   │
│  │  │ AgentMember │ │ IssueAgent   │ │ Workflow     │             │   │
│  │  │ Handler     │ │ Handler      │ │ Handler      │             │   │
│  │  └──────┬──────┘ └──────┬───────┘ └──────┬───────┘             │   │
│  └─────────┼───────────────┼────────────────┼──────────────────────┘   │
│            │               │                │                          │
│  ┌─────────▼───────────────▼────────────────▼──────────────────────┐   │
│  │                    Service Layer                                 │   │
│  │  ┌─────────────┐ ┌──────────────┐ ┌──────────────┐             │   │
│  │  │ AgentMember │ │ IssueAgent   │ │ Workflow     │             │   │
│  │  │ Service     │ │ Service      │ │ Service      │             │   │
│  │  └──────┬──────┘ └──────┬───────┘ └──────┬───────┘             │   │
│  │         │               │                │                      │   │
│  │  ┌──────▼──────┐ ┌──────▼───────┐ ┌──────▼───────┐             │   │
│  │  │ AgentTask   │ │ IssueService │ │ AgentService │             │   │
│  │  │ Service     │ │ (existing)   │ │ (AI module)  │             │   │
│  │  └─────────────┘ └──────────────┘ └──────────────┘             │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│             │                                                          │
│  ┌──────────▼──────────────────────────────────────────────────────┐   │
│  │                    Model Layer (GORM)                            │   │
│  │  ┌─────────────┐ ┌──────────────┐ ┌──────────────┐             │   │
│  │  │ Project     │ │ Issue        │ │ AgentWorkflow│             │   │
│  │  │ AgentMember │ │ (+AgentID)   │ │ WorkflowNode │             │   │
│  │  └─────────────┘ └──────────────┘ │ WorkflowEdge │             │   │
│  │                                    │ WorkflowRun  │             │   │
│  │  ┌─────────────┐ ┌──────────────┐ │ WorkflowNodeRun            │   │
│  │  │ Agent       │ │ AgentTask    │ └──────────────┘             │   │
│  │  │ (existing)  │ │ (existing)   │                              │   │
│  │  └─────────────┘ └──────────────┘                              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│             │                                                          │
│  ┌──────────▼──────────────────────────────────────────────────────┐   │
│  │                    PostgreSQL Database                           │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

### 1.2 模块依赖关系

```
                    ┌─────────────────┐
                    │   AgentMember   │
                    │    Service      │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   AgentService  │
                    │   (AI module)   │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼───────┐   ┌───────▼───────┐   ┌───────▼───────┐
│ IssueAgent    │   │   Workflow    │   │  AgentTask    │
│ Service       │   │   Service     │   │  Service      │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                   ┌────────▼────────┐
                   │  IssueService   │
                   │  (existing)     │
                   └─────────────────┘
```

### 1.3 数据流设计

#### 1.3.1 Issue 分配给 Agent 数据流

```
用户点击"分配给Agent"
    │
    ▼
IssueAgentHandler.AssignAgent()
    │
    ├─► 验证权限（Issue 存在、Agent 可调用）
    ├─► 更新 Issue.AgentAssigneeID
    ├─► 创建 AgentTask（关联 IssueID）
    ├─► 记录 IssueActivity（agent_assigned）
    └─► 返回 AgentTaskID
```

#### 1.3.2 Agent 执行进度回写数据流

```
Agent 执行中
    │
    ▼
AgentService.DispatchAgent()
    │
    ├─► 每 N 分钟或关键节点
    ├─► 创建 IssueActivity（agent_progress）
    ├─► 更新 AgentTask 状态
    └─► 更新 Issue 状态（如需要）
```

#### 1.3.3 Workflow 执行数据流

```
触发 Workflow（手动/事件/Cron）
    │
    ▼
WorkflowService.Execute()
    │
    ├─► 创建 WorkflowRun
    ├─► 按拓扑顺序执行节点
    │       │
    │       ▼
    │   WorkflowNodeRun
    │       │
    │       ├─► 构建 ContextPayload
    │       ├─► 调用 AgentTask
    │       ├─► 等待完成
    │       ├─► 收集 OutputContext
    │       └─► 传递给下一节点
    │
    └─► 更新 WorkflowRun 状态
```

---

## 二、数据模型设计

### 2.1 新增模型

```go
// ============================================
// 1. AgentWorkflow - 工作流定义
// ============================================
type AgentWorkflow struct {
    BaseModel
    Name          string          `gorm:"size:100;not null" json:"name"`
    Description   string          `gorm:"type:text" json:"description"`
    ProjectID     uint64          `gorm:"not null;index" json:"project_id"`
    WorkspaceID   uint64          `gorm:"not null;index" json:"workspace_id"`
    Version       int             `gorm:"default:1" json:"version"`
    IsActive      bool            `gorm:"default:true" json:"is_active"`
    TriggerType   string          `gorm:"size:20;default:manual" json:"trigger_type"` // manual|event|cron|webhook
    TriggerConfig json.RawMessage `gorm:"type:jsonb" json:"trigger_config"`
    Config        json.RawMessage `gorm:"type:jsonb" json:"config"`
}

// ============================================
// 2. WorkflowNode - 流程节点
// ============================================
type WorkflowNode struct {
    BaseModel
    WorkflowID    uint64          `gorm:"not null;index" json:"workflow_id"`
    AgentID       uint64          `gorm:"not null" json:"agent_id"`
    NodeType      string          `gorm:"size:20;default:agent" json:"node_type"` // agent|condition|parallel|loop|gate
    Name          string          `gorm:"size:100;not null" json:"name"`
    Config        json.RawMessage `gorm:"type:jsonb" json:"config"`
    ContextConfig json.RawMessage `gorm:"type:jsonb" json:"context_config"`
    SortOrder     int             `gorm:"default:0" json:"sort_order"`
    Timeout       int             `gorm:"default:1800" json:"timeout"` // 秒
    RetryPolicy   string          `gorm:"size:20;default:retry" json:"retry_policy"` // retry|skip|abort
    MaxRetries    int             `gorm:"default:3" json:"max_retries"`
}

// ============================================
// 3. WorkflowEdge - 流程连接
// ============================================
type WorkflowEdge struct {
    BaseModel
    WorkflowID     uint64          `gorm:"not null;index" json:"workflow_id"`
    SourceNodeID   uint64          `gorm:"not null" json:"source_node_id"`
    TargetNodeID   uint64          `gorm:"not null" json:"target_node_id"`
    Condition      string          `gorm:"type:text" json:"condition"` // 条件表达式
    ContextMapping json.RawMessage `gorm:"type:jsonb" json:"context_mapping"`
}

// ============================================
// 4. WorkflowRun - 工作流执行记录
// ============================================
type WorkflowRun struct {
    BaseModel
    WorkflowID   uint64          `gorm:"not null;index" json:"workflow_id"`
    IssueID      *uint64         `gorm:"index" json:"issue_id"`
    Status       string          `gorm:"size:20;default:pending" json:"status"` // pending|running|completed|failed|cancelled
    CurrentNode  *uint64         `gorm:"index" json:"current_node"`
    Context      json.RawMessage `gorm:"type:jsonb" json:"context"` // 全局上下文
    StartedAt    *time.Time      `json:"started_at"`
    CompletedAt  *time.Time      `json:"completed_at"`
    TotalTokens  int             `json:"total_tokens"`
    TotalCost    float64         `json:"total_cost"`
    ErrorInfo    string          `gorm:"type:text" json:"error_info"`
}

// ============================================
// 5. WorkflowNodeRun - 节点执行记录
// ============================================
type WorkflowNodeRun struct {
    BaseModel
    WorkflowRunID uint64          `gorm:"not null;index" json:"workflow_run_id"`
    NodeID        uint64          `gorm:"not null" json:"node_id"`
    AgentID       uint64          `gorm:"not null" json:"agent_id"`
    AgentTaskID   *uint64         `gorm:"index" json:"agent_task_id"` // 关联的 AgentTask
    Status        string          `gorm:"size:20;default:pending" json:"status"` // pending|running|completed|failed|skipped
    InputContext  json.RawMessage `gorm:"type:jsonb" json:"input_context"`
    OutputContext json.RawMessage `gorm:"type:jsonb" json:"output_context"`
    StartedAt     *time.Time      `json:"started_at"`
    CompletedAt   *time.Time      `json:"completed_at"`
    TokensUsed    int             `json:"tokens_used"`
    Cost          float64         `json:"cost"`
    ErrorInfo     string          `gorm:"type:text" json:"error_info"`
    RetryCount    int             `gorm:"default:0" json:"retry_count"`
}

// ============================================
// 6. AgentDecision - Agent 决策记录（可解释性）
// ============================================
type AgentDecision struct {
    BaseModel
    AgentID       uint64  `gorm:"not null;index" json:"agent_id"`
    IssueID       *uint64 `gorm:"index" json:"issue_id"`
    AgentTaskID   *uint64 `gorm:"index" json:"agent_task_id"`
    WorkflowRunID *uint64 `gorm:"index" json:"workflow_run_id"`
    NodeType      string  `gorm:"size:50" json:"node_type"` // requirement_analysis|design|coding|testing
    Thinking      string  `gorm:"type:text" json:"thinking"` // 思考过程
    Decision      string  `gorm:"type:text" json:"decision"` // 决策结果
    Reasoning     string  `gorm:"type:text" json:"reasoning"` // 决策依据
    Alternatives  string  `gorm:"type:text" json:"alternatives"` // 备选方案（JSON）
    Confidence    float64 `json:"confidence"` // 置信度
}

// ============================================
// 7. AgentCostBudget - 项目级成本预算
// ============================================
type AgentCostBudget struct {
    BaseModel
    ProjectID     uint64   `gorm:"not null;uniqueIndex" json:"project_id"`
    MonthlyBudget float64  `json:"monthly_budget"` // 月度预算上限
    CurrentCost   float64  `json:"current_cost"`   // 当前已用成本
    AlertThreshold float64 `json:"alert_threshold"` // 告警阈值（百分比）
    AutoBlock     bool     `gorm:"default:false" json:"auto_block"` // 超预算自动阻止
    LastResetAt   *time.Time `json:"last_reset_at"` // 上次重置时间
}

// ============================================
// 8. AgentSLA - Agent 执行 SLA
// ============================================
type AgentSLA struct {
    BaseModel
    ProjectID       uint64 `gorm:"not null;uniqueIndex" json:"project_id"`
    NormalTaskMax   int    `gorm:"default:1800" json:"normal_task_max"`   // 普通任务最大时间（秒）
    ComplexTaskMax  int    `gorm:"default:7200" json:"complex_task_max"` // 复杂任务最大时间（秒）
    AutoEscalation  bool   `gorm:"default:true" json:"auto_escalation"`  // 超时自动升级
    Enabled         bool   `gorm:"default:true" json:"enabled"`
}
```

### 2.2 扩展现有模型

```go
// Issue 增加字段
type Issue struct {
    // ... 现有字段 ...
    AgentAssigneeID *uint64 `gorm:"index" json:"agent_assignee_id"` // Agent 指派人
    AgentTaskID     *uint64 `gorm:"index" json:"agent_task_id"`     // 关联的 Agent 任务
}

// AgentTask 增加字段
type AgentTask struct {
    // ... 现有字段 ...
    WorkflowRunID *uint64 `gorm:"index" json:"workflow_run_id"` // 关联的工作流执行
    WorkflowNodeRunID *uint64 `gorm:"index" json:"workflow_node_run_id"` // 关联的节点执行
    SLABreach     bool    `gorm:"default:false" json:"sla_breach"` // 是否超出SLA
}
```

### 2.3 数据库迁移

```sql
-- migrations/000015_agent_project_integration.up.sql

-- 1. Issue 增加 Agent 字段
ALTER TABLE issues ADD COLUMN agent_assignee_id BIGINT REFERENCES agents(id);
ALTER TABLE issues ADD COLUMN agent_task_id BIGINT REFERENCES agent_tasks(id);
CREATE INDEX idx_issues_agent_assignee ON issues(agent_assignee_id);

-- 2. AgentTask 增加 Workflow 字段
ALTER TABLE agent_tasks ADD COLUMN workflow_run_id BIGINT REFERENCES workflow_runs(id);
ALTER TABLE agent_tasks ADD COLUMN workflow_node_run_id BIGINT REFERENCES workflow_node_runs(id);
ALTER TABLE agent_tasks ADD COLUMN sla_breach BOOLEAN DEFAULT FALSE;

-- 3. AgentWorkflow 工作流定义
CREATE TABLE agent_workflows (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    project_id BIGINT NOT NULL REFERENCES projects(id),
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    version INT DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    trigger_type VARCHAR(20) DEFAULT 'manual',
    trigger_config JSONB,
    config JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    created_by_id BIGINT REFERENCES users(id),
    updated_by_id BIGINT REFERENCES users(id)
);
CREATE INDEX idx_workflows_project ON agent_workflows(project_id);
CREATE INDEX idx_workflows_workspace ON agent_workflows(workspace_id);

-- 4. WorkflowNode 流程节点
CREATE TABLE workflow_nodes (
    id BIGSERIAL PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES agent_workflows(id) ON DELETE CASCADE,
    agent_id BIGINT NOT NULL REFERENCES agents(id),
    node_type VARCHAR(20) DEFAULT 'agent',
    name VARCHAR(100) NOT NULL,
    config JSONB,
    context_config JSONB,
    sort_order INT DEFAULT 0,
    timeout INT DEFAULT 1800,
    retry_policy VARCHAR(20) DEFAULT 'retry',
    max_retries INT DEFAULT 3,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_workflow_nodes_workflow ON workflow_nodes(workflow_id);

-- 5. WorkflowEdge 流程连接
CREATE TABLE workflow_edges (
    id BIGSERIAL PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES agent_workflows(id) ON DELETE CASCADE,
    source_node_id BIGINT NOT NULL REFERENCES workflow_nodes(id),
    target_node_id BIGINT NOT NULL REFERENCES workflow_nodes(id),
    condition TEXT,
    context_mapping JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_workflow_edges_workflow ON workflow_edges(workflow_id);

-- 6. WorkflowRun 工作流执行记录
CREATE TABLE workflow_runs (
    id BIGSERIAL PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES agent_workflows(id),
    issue_id BIGINT REFERENCES issues(id),
    status VARCHAR(20) DEFAULT 'pending',
    current_node_id BIGINT REFERENCES workflow_nodes(id),
    context JSONB,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    total_tokens INT DEFAULT 0,
    total_cost DECIMAL(10,4) DEFAULT 0,
    error_info TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    created_by_id BIGINT REFERENCES users(id),
    updated_by_id BIGINT REFERENCES users(id)
);
CREATE INDEX idx_workflow_runs_workflow ON workflow_runs(workflow_id);
CREATE INDEX idx_workflow_runs_issue ON workflow_runs(issue_id);

-- 7. WorkflowNodeRun 节点执行记录
CREATE TABLE workflow_node_runs (
    id BIGSERIAL PRIMARY KEY,
    workflow_run_id BIGINT NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    node_id BIGINT NOT NULL REFERENCES workflow_nodes(id),
    agent_id BIGINT NOT NULL REFERENCES agents(id),
    agent_task_id BIGINT REFERENCES agent_tasks(id),
    status VARCHAR(20) DEFAULT 'pending',
    input_context JSONB,
    output_context JSONB,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    tokens_used INT DEFAULT 0,
    cost DECIMAL(10,4) DEFAULT 0,
    error_info TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_workflow_node_runs_workflow_run ON workflow_node_runs(workflow_run_id);

-- 8. AgentDecision Agent 决策记录
CREATE TABLE agent_decisions (
    id BIGSERIAL PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES agents(id),
    issue_id BIGINT REFERENCES issues(id),
    agent_task_id BIGINT REFERENCES agent_tasks(id),
    workflow_run_id BIGINT REFERENCES workflow_runs(id),
    node_type VARCHAR(50),
    thinking TEXT,
    decision TEXT,
    reasoning TEXT,
    alternatives JSONB,
    confidence DECIMAL(5,4),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_agent_decisions_agent ON agent_decisions(agent_id);
CREATE INDEX idx_agent_decisions_issue ON agent_decisions(issue_id);

-- 9. AgentCostBudget 项目级成本预算
CREATE TABLE agent_cost_budgets (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL UNIQUE REFERENCES projects(id),
    monthly_budget DECIMAL(10,2) DEFAULT 0,
    current_cost DECIMAL(10,2) DEFAULT 0,
    alert_threshold DECIMAL(5,2) DEFAULT 80,
    auto_block BOOLEAN DEFAULT FALSE,
    last_reset_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 10. AgentSLA Agent 执行 SLA
CREATE TABLE agent_slas (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL UNIQUE REFERENCES projects(id),
    normal_task_max INT DEFAULT 1800,
    complex_task_max INT DEFAULT 7200,
    auto_escalation BOOLEAN DEFAULT TRUE,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

---

## 三、API 设计

### 3.1 Agent 成员管理 API

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/v1/projects/:projectId/agent-members` | ListAgentMembers | 获取项目 Agent 成员列表 |
| POST | `/api/v1/projects/:projectId/agent-members` | AddAgentMember | 添加 Agent 到项目 |
| PATCH | `/api/v1/projects/:projectId/agent-members/:agentId` | UpdateAgentMember | 更新 Agent 角色 |
| DELETE | `/api/v1/projects/:projectId/agent-members/:agentId` | RemoveAgentMember | 移除 Agent |

### 3.2 Issue Agent 分配 API

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/api/v1/issues/:issueId/assign-agent` | AssignAgent | 分配 Issue 给 Agent |
| DELETE | `/api/v1/issues/:issueId/unassign-agent` | UnassignAgent | 取消 Agent 分配 |
| GET | `/api/v1/issues/:issueId/agent-status` | GetAgentStatus | 获取 Agent 执行状态 |
| POST | `/api/v1/issues/:issueId/preview-agent` | PreviewAgentExecution | 预览 Agent 执行计划 |

### 3.3 AI 辅助分解 API

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/api/v1/issues/:issueId/ai-decompose` | AIDecompose | AI 分解工作项 |
| POST | `/api/v1/issues/:issueId/ai-generate-doc` | AIGenerateDoc | AI 生成 Wiki 文档 |

### 3.4 Workflow 编排 API

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/v1/projects/:projectId/workflows` | ListWorkflows | 获取工作流列表 |
| POST | `/api/v1/projects/:projectId/workflows` | CreateWorkflow | 创建工作流 |
| GET | `/api/v1/projects/:projectId/workflows/:workflowId` | GetWorkflow | 获取工作流详情 |
| PUT | `/api/v1/projects/:projectId/workflows/:workflowId` | UpdateWorkflow | 更新工作流 |
| DELETE | `/api/v1/projects/:projectId/workflows/:workflowId` | DeleteWorkflow | 删除工作流 |
| POST | `/api/v1/projects/:projectId/workflows/:workflowId/execute` | ExecuteWorkflow | 执行工作流 |
| GET | `/api/v1/projects/:projectId/workflows/:workflowId/runs` | ListWorkflowRuns | 获取执行历史 |
| GET | `/api/v1/projects/:projectId/workflows/:workflowId/runs/:runId` | GetWorkflowRun | 获取执行详情 |

### 3.5 Workflow Node/Edge API

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/api/v1/projects/:projectId/workflows/:workflowId/nodes` | CreateNode | 添加节点 |
| PUT | `/api/v1/projects/:projectId/workflows/:workflowId/nodes/:nodeId` | UpdateNode | 更新节点 |
| DELETE | `/api/v1/projects/:projectId/workflows/:workflowId/nodes/:nodeId` | DeleteNode | 删除节点 |
| POST | `/api/v1/projects/:projectId/workflows/:workflowId/edges` | CreateEdge | 添加连接 |
| PUT | `/api/v1/projects/:projectId/workflows/:workflowId/edges/:edgeId` | UpdateEdge | 更新连接 |
| DELETE | `/api/v1/projects/:projectId/workflows/:workflowId/edges/:edgeId` | DeleteEdge | 删除连接 |

### 3.6 成本预算和 SLA API

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/v1/projects/:projectId/ai-budget` | GetBudget | 获取成本预算 |
| PUT | `/api/v1/projects/:projectId/ai-budget` | UpdateBudget | 更新成本预算 |
| GET | `/api/v1/projects/:projectId/ai-sla` | GetSLA | 获取 SLA 配置 |
| PUT | `/api/v1/projects/:projectId/ai-sla` | UpdateSLA | 更新 SLA 配置 |

### 3.7 决策可解释性 API

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/v1/issues/:issueId/decisions` | ListDecisions | 获取 Issue 关联的决策记录 |
| GET | `/api/v1/agent-tasks/:taskId/decisions` | ListTaskDecisions | 获取任务关联的决策记录 |

### 3.8 批量操作 API

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/api/v1/issues/bulk/assign-agent` | BulkAssignAgent | 批量分配 Issue 给 Agent |

---

## 四、Service 层设计

### 4.1 AgentMemberService

```go
// internal/service/agent_member_service.go
type AgentMemberService struct {
    db *gorm.DB
}

func (s *AgentMemberService) ListByProject(projectID uint64) ([]model.ProjectAgentMember, error)
func (s *AgentMemberService) Add(projectID, agentID uint64, role string) (*model.ProjectAgentMember, error)
func (s *AgentMemberService) UpdateRole(projectID, agentID uint64, role string) error
func (s *AgentMemberService) Remove(projectID, agentID uint64) error
func (s *AgentMemberService) IsMember(projectID, agentID uint64) (bool, error)
```

### 4.2 IssueAgentService

```go
// internal/service/issue_agent_service.go
type IssueAgentService struct {
    db              *gorm.DB
    agentTaskSvc    *AgentTaskService
    agentSvc        *aiservice.AgentService  // 通过接口注入
    activitySvc     *IssueActivityService
    decisionSvc     *AgentDecisionService
    budgetSvc       *AgentCostBudgetService
    slaSvc          *AgentSLAService
}

func (s *IssueAgentService) Assign(issueID, agentID uint64, req AssignRequest) (*model.AgentTask, error)
func (s *IssueAgentService) Unassign(issueID uint64) error
func (s *IssueAgentService) GetStatus(issueID uint64) (*AgentStatusResponse, error)
func (s *IssueAgentService) PreviewExecution(issueID, agentID uint64) (*ExecutionPreview, error)
func (s *IssueAgentService) BulkAssign(issueIDs []uint64, agentID uint64) ([]*model.AgentTask, error)
```

### 4.3 WorkflowService

```go
// internal/service/workflow_service.go
type WorkflowService struct {
    db           *gorm.DB
    agentSvc     *aiservice.AgentService
    taskSvc      *AgentTaskService
    contextSvc   *ContextPayloadService
}

func (s *WorkflowService) ListByProject(projectID uint64) ([]model.AgentWorkflow, error)
func (s *WorkflowService) Create(projectID uint64, req CreateWorkflowRequest) (*model.AgentWorkflow, error)
func (s *WorkflowService) Get(workflowID uint64) (*WorkflowDetail, error)
func (s *WorkflowService) Update(workflowID uint64, req UpdateWorkflowRequest) error
func (s *WorkflowService) Delete(workflowID uint64) error

// 节点管理
func (s *WorkflowService) AddNode(workflowID uint64, req CreateNodeRequest) (*model.WorkflowNode, error)
func (s *WorkflowService) UpdateNode(nodeID uint64, req UpdateNodeRequest) error
func (s *WorkflowService) DeleteNode(nodeID uint64) error

// 连接管理
func (s *WorkflowService) AddEdge(workflowID uint64, req CreateEdgeRequest) (*model.WorkflowEdge, error)
func (s *WorkflowService) UpdateEdge(edgeID uint64, req UpdateEdgeRequest) error
func (s *WorkflowService) DeleteEdge(edgeID uint64) error

// 执行
func (s *WorkflowService) Execute(workflowID uint64, issueID *uint64, ctx context.Context) (*model.WorkflowRun, error)
func (s *WorkflowService) GetRuns(workflowID uint64) ([]model.WorkflowRun, error)
func (s *WorkflowService) GetRun(runID uint64) (*WorkflowRunDetail, error)
func (s *WorkflowService) CancelRun(runID uint64) error
```

### 4.4 ContextPayloadService

```go
// internal/service/context_payload_service.go
type ContextPayloadService struct {
    db *gorm.DB
}

func (s *ContextPayloadService) BuildInitialContext(issueID uint64) (*ContextPayload, error)
func (s *ContextPayloadService) BuildNodeInput(workflowRunID, nodeID uint64, edges []model.WorkflowEdge) (*ContextPayload, error)
func (s *ContextPayloadService) MergeParallelOutputs(workflowRunID uint64, nodeIDs []uint64) (*ContextPayload, error)
func (s *ContextPayloadService) CompressContext(ctx *ContextPayload) (*ContextPayload, error)
```

**上下文压缩策略：**

| 策略 | 说明 | 触发条件 |
|------|------|----------|
| **摘要替换** | 将长文档替换为 AI 生成的摘要 | 文档超过 2000 tokens |
| **字段裁剪** | 移除非关键字段，只保留核心信息 | 上下文超过 5000 tokens |
| **分页加载** | 大文档分页加载，只传递当前页 | 文档超过 10 页 |
| **历史压缩** | 前置节点输出压缩为关键结论 | 串行节点超过 3 个 |

```go
// CompressContext 实现多级压缩
func (s *ContextPayloadService) CompressContext(ctx *ContextPayload) (*ContextPayload, error) {
    compressed := &ContextPayload{
        IssueContext: ctx.IssueContext,
        Documents:    make([]DocumentRef, 0),
        AgentOutputs: make([]AgentOutput, 0),
    }
    
    // 1. 压缩文档
    for _, doc := range ctx.Documents {
        if doc.TokenCount > 2000 {
            // 替换为摘要
            summary, err := s.summarizeDocument(doc)
            if err != nil {
                return nil, err
            }
            compressed.Documents = append(compressed.Documents, summary)
        } else {
            compressed.Documents = append(compressed.Documents, doc)
        }
    }
    
    // 2. 压缩前置 Agent 输出
    if len(ctx.AgentOutputs) > 3 {
        // 只保留最近 3 个输出，其余压缩为结论
        recent := ctx.AgentOutputs[len(ctx.AgentOutputs)-3:]
        older := ctx.AgentOutputs[:len(ctx.AgentOutputs)-3]
        
        conclusions := s.extractConclusions(older)
        compressed.AgentOutputs = append([]AgentOutput{conclusions}, recent...)
    } else {
        compressed.AgentOutputs = ctx.AgentOutputs
    }
    
    // 3. 裁剪 Issue 上下文字段
    compressed.IssueContext = s.trimIssueContext(ctx.IssueContext)
    
    return compressed, nil
}
```

### 4.5 AgentDecisionService

```go
// internal/service/agent_decision_service.go
type AgentDecisionService struct {
    db *gorm.DB
}

func (s *AgentDecisionService) Record(decision *model.AgentDecision) error
func (s *AgentDecisionService) ListByIssue(issueID uint64) ([]model.AgentDecision, error)
func (s *AgentDecisionService) ListByTask(taskID uint64) ([]model.AgentDecision, error)
```

### 4.6 AgentCostBudgetService

```go
// internal/service/agent_cost_budget_service.go
type AgentCostBudgetService struct {
    db *gorm.DB
}

func (s *AgentCostBudgetService) Get(projectID uint64) (*model.AgentCostBudget, error)
func (s *AgentCostBudgetService) Update(projectID uint64, req UpdateBudgetRequest) error
func (s *AgentCostBudgetService) CheckBudget(projectID uint64, estimatedCost float64) (bool, string, error)
func (s *AgentCostBudgetService) RecordCost(projectID uint64, cost float64) error
func (s *AgentCostBudgetService) ResetMonthly() error
```

### 4.7 AgentSLAService

```go
// internal/service/agent_sla_service.go
type AgentSLAService struct {
    db *gorm.DB
}

func (s *AgentSLAService) Get(projectID uint64) (*model.AgentSLA, error)
func (s *AgentSLAService) Update(projectID uint64, req UpdateSLARequest) error
func (s *AgentSLAService) CheckSLA(taskID uint64) (bool, error)
func (s *AgentSLAService) StartMonitoring() // 后台监控 goroutine
```

---

## 五、Handler 层设计

### 5.1 AgentMemberHandler

```go
// internal/handler/agent_member_handler.go
type AgentMemberHandler struct {
    memberSvc *service.AgentMemberService
}

// ListByProject GET /projects/:projectId/agent-members
// Add POST /projects/:projectId/agent-members
// UpdateRole PATCH /projects/:projectId/agent-members/:agentId
// Remove DELETE /projects/:projectId/agent-members/:agentId
```

### 5.2 IssueAgentHandler

```go
// internal/handler/issue_agent_handler.go
type IssueAgentHandler struct {
    issueAgentSvc *service.IssueAgentService
}

// AssignAgent POST /issues/:issueId/assign-agent
// UnassignAgent DELETE /issues/:issueId/unassign-agent
// GetAgentStatus GET /issues/:issueId/agent-status
// PreviewExecution POST /issues/:issueId/preview-agent
// BulkAssign POST /issues/bulk/assign-agent
```

### 5.3 WorkflowHandler

```go
// internal/handler/workflow_handler.go
type WorkflowHandler struct {
    workflowSvc *service.WorkflowService
}

// ListWorkflows GET /projects/:projectId/workflows
// CreateWorkflow POST /projects/:projectId/workflows
// GetWorkflow GET /projects/:projectId/workflows/:workflowId
// UpdateWorkflow PUT /projects/:projectId/workflows/:workflowId
// DeleteWorkflow DELETE /projects/:projectId/workflows/:workflowId
// ExecuteWorkflow POST /projects/:projectId/workflows/:workflowId/execute
// ListRuns GET /projects/:projectId/workflows/:workflowId/runs
// GetRun GET /projects/:projectId/workflows/:workflowId/runs/:runId
// CancelRun POST /projects/:projectId/workflows/:workflowId/runs/:runId/cancel

// AddNode POST /projects/:projectId/workflows/:workflowId/nodes
// UpdateNode PUT /projects/:projectId/workflows/:workflowId/nodes/:nodeId
// DeleteNode DELETE /projects/:projectId/workflows/:workflowId/nodes/:nodeId
// AddEdge POST /projects/:projectId/workflows/:workflowId/edges
// UpdateEdge PUT /projects/:projectId/workflows/:workflowId/edges/:edgeId
// DeleteEdge DELETE /projects/:projectId/workflows/:workflowId/edges/:edgeId
```

---

## 六、路由注册设计

### 6.1 新增路由组

```go
// router.go 中新增路由组

// ========== Agent 成员管理 ==========
projectAgentMembers := projects.Group("/:projectId/agent-members")
{
    projectAgentMembers.GET("", agentMemberHandler.ListByProject)
    projectAgentMembers.POST("", agentMemberHandler.Add)
    projectAgentMembers.PATCH("/:agentId", agentMemberHandler.UpdateRole)
    projectAgentMembers.DELETE("/:agentId", agentMemberHandler.Remove)
}

// ========== Issue Agent 分配 ==========
issuesGroup := issues.Group("/:issueId")
{
    issuesGroup.POST("/assign-agent", issueAgentHandler.AssignAgent)
    issuesGroup.DELETE("/unassign-agent", issueAgentHandler.UnassignAgent)
    issuesGroup.GET("/agent-status", issueAgentHandler.GetAgentStatus)
    issuesGroup.POST("/preview-agent", issueAgentHandler.PreviewExecution)
    issuesGroup.POST("/ai-decompose", issueAgentHandler.AIDecompose)
    issuesGroup.POST("/decisions", decisionHandler.ListByIssue)
}

// 批量分配
issues.POST("/bulk/assign-agent", issueAgentHandler.BulkAssign)

// ========== Workflow 编排 ==========
workflows := projects.Group("/:projectId/workflows")
{
    workflows.GET("", workflowHandler.ListWorkflows)
    workflows.POST("", workflowHandler.CreateWorkflow)
    
    workflow := workflows.Group("/:workflowId")
    {
        workflow.GET("", workflowHandler.GetWorkflow)
        workflow.PUT("", workflowHandler.UpdateWorkflow)
        workflow.DELETE("", workflowHandler.DeleteWorkflow)
        workflow.POST("/execute", workflowHandler.ExecuteWorkflow)
        workflow.GET("/runs", workflowHandler.ListRuns)
        workflow.GET("/runs/:runId", workflowHandler.GetRun)
        workflow.POST("/runs/:runId/cancel", workflowHandler.CancelRun)
        
        // 节点管理
        workflow.POST("/nodes", workflowHandler.AddNode)
        workflow.PUT("/nodes/:nodeId", workflowHandler.UpdateNode)
        workflow.DELETE("/nodes/:nodeId", workflowHandler.DeleteNode)
        
        // 连接管理
        workflow.POST("/edges", workflowHandler.AddEdge)
        workflow.PUT("/edges/:edgeId", workflowHandler.UpdateEdge)
        workflow.DELETE("/edges/:edgeId", workflowHandler.DeleteEdge)
    }
}

// ========== 成本预算和 SLA ==========
projects("/:projectId/ai-budget", budgetHandler.Get, budgetHandler.Update)
projects("/:projectId/ai-sla", slaHandler.Get, slaHandler.Update)
```

### 6.2 路由依赖注入

```go
// router.go 中服务初始化
func SetupRoutes(db *gorm.DB, ...) {
    // ... 现有服务初始化 ...
    
    // 新增服务
    agentMemberSvc := service.NewAgentMemberService(db)
    issueAgentSvc := service.NewIssueAgentService(db, agentTaskSvc, aiAgentSvc, activitySvc, decisionSvc, budgetSvc, slaSvc)
    workflowSvc := service.NewWorkflowService(db, aiAgentSvc, taskSvc, contextSvc)
    contextSvc := service.NewContextPayloadService(db)
    decisionSvc := service.NewAgentDecisionService(db)
    budgetSvc := service.NewAgentCostBudgetService(db)
    slaSvc := service.NewAgentSLAService(db)
    
    // 新增 Handler
    agentMemberHandler := handler.NewAgentMemberHandler(agentMemberSvc)
    issueAgentHandler := handler.NewIssueAgentHandler(issueAgentSvc)
    workflowHandler := handler.NewWorkflowHandler(workflowSvc)
    decisionHandler := handler.NewDecisionHandler(decisionSvc)
    budgetHandler := handler.NewBudgetHandler(budgetSvc)
    slaHandler := handler.NewSLAHandler(slaSvc)
    
    // 注册路由 ...
}
```

---

## 七、前端架构设计

### 7.1 API 层

```
frontend/src/api/
├── agent-member.ts      # Agent 成员管理
├── issue-agent.ts       # Issue Agent 分配
├── workflow.ts          # Workflow 编排
├── ai-budget.ts         # 成本预算
├── ai-sla.ts            # SLA 配置
└── agent-decision.ts    # 决策记录
```

### 7.2 组件架构

```
frontend/src/views/
├── agents/
│   ├── AgentDashboard.vue          # 已有
│   ├── AgentMemberList.vue         # 新增：项目 Agent 成员
│   ├── WorkflowList.vue            # 新增：工作流列表
│   ├── WorkflowDesigner.vue        # 新增：可视化流程设计器
│   └── WorkflowRunDetail.vue       # 新增：执行详情
│
├── issues/
│   ├── IssueDetail.vue             # 修改：增加 Agent 标签页
│   ├── IssueAgentPanel.vue         # 新增：Agent 分配面板
│   ├── IssueAgentStatus.vue        # 新增：Agent 执行状态
│   └── IssueDecisions.vue          # 新增：决策记录
│
└── components/
    ├── agents/
    │   ├── AgentMemberCard.vue     # 新增：Agent 成员卡片
    │   ├── WorkflowNodeEditor.vue  # 新增：节点编辑器
    │   ├── WorkflowEdgeEditor.vue  # 新增：连接编辑器
    │   ├── ContextPreview.vue      # 新增：上下文预览
    │   ├── ExecutionPreview.vue    # 新增：执行预览
    │   └── NewUserGuide.vue        # 新增：新用户引导
    │
    └── workflow/
        ├── WorkflowCanvas.vue      # 新增：流程图画布
        ├── WorkflowToolbar.vue     # 新增：工具栏
        ├── WorkflowProperties.vue  # 新增：属性面板
        └── WorkflowRunTimeline.vue # 新增：执行时间线
```

### 7.3 Composables

```
frontend/src/composables/
├── useWorkflow.ts        # 工作流状态管理
├── useWorkflowDesigner.ts # 流程设计器逻辑
├── useAgentMember.ts     # Agent 成员管理
├── useExecutionPreview.ts # 执行预览
└── useNewUserGuide.ts    # 新用户引导
```

### 7.4 Workflow 设计器核心组件

```vue
<!-- WorkflowDesigner.vue -->
<template>
  <div class="workflow-designer">
    <!-- 工具栏 -->
    <WorkflowToolbar
      @add-node="addNode"
      @save="saveWorkflow"
      @execute="executeWorkflow"
    />
    
    <div class="designer-content">
      <!-- 左侧：节点面板 -->
      <div class="node-panel">
        <AgentNodeList
          :agents="availableAgents"
          @drag-start="onDragStart"
        />
      </div>
      
      <!-- 中间：画布 -->
      <WorkflowCanvas
        :nodes="nodes"
        :edges="edges"
        :selected-node="selectedNode"
        @select-node="selectNode"
        @add-edge="addEdge"
        @move-node="moveNode"
      />
      
      <!-- 右侧：属性面板 -->
      <WorkflowProperties
        v-if="selectedNode"
        :node="selectedNode"
        :agents="availableAgents"
        @update-node="updateNode"
        @delete-node="deleteNode"
      />
    </div>
  </div>
</template>
```

---

## 八、Workflow 执行引擎设计

### 8.1 执行流程

```go
// internal/service/workflow_executor.go
type WorkflowExecutor struct {
    workflowSvc   *WorkflowService
    agentSvc      *aiservice.AgentService
    contextSvc    *ContextPayloadService
    decisionSvc   *AgentDecisionService
    budgetSvc     *AgentCostBudgetService
    slaSvc        *AgentSLAService
    semaphore     chan struct{} // 并发控制信号量
    maxParallel   int          // 最大并行数
}

func NewWorkflowExecutor(...) *WorkflowExecutor {
    return &WorkflowExecutor{
        // ... 其他初始化 ...
        semaphore:   make(chan struct{}, 10), // 最大10个并行 Workflow
        maxParallel: 5,                       // 每个 Workflow 最大5个并行节点
    }
}

func (e *WorkflowExecutor) Execute(run *model.WorkflowRun) error {
    // 1. 加载工作流定义
    workflow, err := e.workflowSvc.Get(run.WorkflowID)
    if err != nil {
        return err
    }
    
    // 2. 获取拓扑排序后的节点列表
    topology := e.getTopology(workflow.Nodes, workflow.Edges)
    
    // 3. 构建初始上下文
    ctx, err := e.contextSvc.BuildInitialContext(run.IssueID)
    if err != nil {
        return err
    }
    
    // 4. 按拓扑顺序执行节点
    for _, layer := range topology {
        if len(layer) == 1 {
            // 串行执行
            err = e.executeNode(run, layer[0], ctx)
            if err != nil {
                return e.handleNodeFailure(run, layer[0], err)
            }
        } else {
            // 并行执行
            err = e.executeParallel(run, layer, ctx)
            if err != nil {
                return e.handleNodeFailure(run, layer[0], err)
            }
        }
        
        // 更新上下文
        ctx, err = e.contextSvc.BuildNodeInput(run.ID, layer[0].ID, workflow.Edges)
        if err != nil {
            return err
        }
    }
    
    // 5. 更新工作流状态为完成
    return e.workflowSvc.CompleteRun(run.ID)
}

func (e *WorkflowExecutor) executeNode(run *model.WorkflowRun, node *model.WorkflowNode, ctx *ContextPayload) error {
    // 1. 检查预算
    ok, msg, err := e.budgetSvc.CheckBudget(run.ProjectID, node.AgentID)
    if err != nil || !ok {
        return fmt.Errorf("budget check failed: %s", msg)
    }
    
    // 2. 构建 AgentTask 输入
    input, err := e.contextSvc.BuildNodeInput(run.ID, node.ID, nil)
    if err != nil {
        return err
    }
    
    // 3. 创建 AgentTask
    task, err := e.createAgentTask(run, node, input)
    if err != nil {
        return err
    }
    
    // 4. 记录决策
    e.decisionSvc.Record(&model.AgentDecision{
        AgentID:     node.AgentID,
        WorkflowRunID: &run.ID,
        NodeType:    node.NodeType,
        Thinking:    "开始执行节点: " + node.Name,
    })
    
    // 5. 等待任务完成（带超时）
    return e.waitForTask(task.ID, node.Timeout)
}

func (e *WorkflowExecutor) executeParallel(run *model.WorkflowRun, nodes []model.WorkflowNode, ctx *ContextPayload) error {
    var wg sync.WaitGroup
    errCh := make(chan error, len(nodes))
    var mu sync.Mutex
    var errs []error
    
    for _, node := range nodes {
        wg.Add(1)
        go func(n model.WorkflowNode) {
            defer wg.Done()
            
            // 获取并发许可
            e.semaphore <- struct{}{}
            defer func() { <-e.semaphore }()
            
            if err := e.executeNode(run, &n, ctx); err != nil {
                mu.Lock()
                errs = append(errs, err)
                mu.Unlock()
                errCh <- err
            }
        }(node)
    }
    
    wg.Wait()
    close(errCh)
    
    if len(errs) > 0 {
        return fmt.Errorf("parallel execution failed: %v", errs)
    }
    return nil
}
```

### 8.2 拓扑排序

```go
func (e *WorkflowExecutor) getTopology(nodes []model.WorkflowNode, edges []model.WorkflowEdge) [][]model.WorkflowNode {
    // 构建邻接表和入度表
    inDegree := make(map[uint64]int)
    adj := make(map[uint64][]uint64)
    
    for _, edge := range edges {
        adj[edge.SourceNodeID] = append(adj[edge.SourceNodeID], edge.TargetNodeID)
        inDegree[edge.TargetNodeID]++
    }
    
    // Kahn's 拓扑排序
    var result [][]model.WorkflowNode
    var queue []uint64
    
    for _, node := range nodes {
        if inDegree[node.ID] == 0 {
            queue = append(queue, node.ID)
        }
    }
    
    for len(queue) > 0 {
        var layer []model.WorkflowNode
        var nextQueue []uint64
        
        for _, id := range queue {
            for _, node := range nodes {
                if node.ID == id {
                    layer = append(layer, node)
                    break
                }
            }
            
            for _, next := range adj[id] {
                inDegree[next]--
                if inDegree[next] == 0 {
                    nextQueue = append(nextQueue, next)
                }
            }
        }
        
        result = append(result, layer)
        queue = nextQueue
    }
    
    return result
}
```

---

## 九、前端路由设计

### 9.1 新增路由

```typescript
// router/index.ts 新增路由
{
  path: '/workspace/:slug/agents/workflows',
  name: 'WorkflowList',
  component: () => import('../views/agents/WorkflowList.vue'),
  meta: { requiresAuth: true }
},
{
  path: '/workspace/:slug/agents/workflows/:workflowId/design',
  name: 'WorkflowDesigner',
  component: () => import('../views/agents/WorkflowDesigner.vue'),
  meta: { requiresAuth: true }
},
{
  path: '/workspace/:slug/agents/workflows/runs/:runId',
  name: 'WorkflowRunDetail',
  component: () => import('../views/agents/WorkflowRunDetail.vue'),
  meta: { requiresAuth: true }
},
{
  path: '/workspace/:slug/agents/members',
  name: 'AgentMemberList',
  component: () => import('../views/agents/AgentMemberList.vue'),
  meta: { requiresAuth: true }
}
```

---

## 十、测试策略

### 10.1 单元测试

| 模块 | 测试文件 | 覆盖率目标 |
|------|----------|-----------|
| AgentMemberService | agent_member_service_test.go | > 80% |
| IssueAgentService | issue_agent_service_test.go | > 80% |
| WorkflowService | workflow_service_test.go | > 80% |
| ContextPayloadService | context_payload_service_test.go | > 80% |
| AgentDecisionService | agent_decision_service_test.go | > 80% |
| AgentCostBudgetService | agent_cost_budget_service_test.go | > 80% |
| AgentSLAService | agent_sla_service_test.go | > 80% |
| WorkflowExecutor | workflow_executor_test.go | > 80% |

### 10.2 集成测试

| 测试场景 | 测试文件 |
|----------|----------|
| Issue 分配给 Agent | issue_agent_integration_test.go |
| Workflow 完整执行 | workflow_integration_test.go |
| 并行节点执行 | workflow_parallel_test.go |
| 预算限制验证 | budget_integration_test.go |
| SLA 超时验证 | sla_integration_test.go |

### 10.3 E2E 测试

```typescript
// e2e/workflow-execution.spec.ts

test.describe('Workflow Execution', () => {
  test('should execute sequential workflow', async ({ page }) => {
    // 1. 创建工作流
    // 2. 添加节点和连接
    // 3. 执行工作流
    // 4. 验证执行结果
  });

  test('should execute parallel workflow', async ({ page }) => {
    // 1. 创建并行工作流
    // 2. 验证并行执行
    // 3. 验证上下文合并
  });

  test('should handle workflow failure', async ({ page }) => {
    // 1. 创建会失败的工作流
    // 2. 执行工作流
    // 3. 验证失败恢复机制
  });

  test('should respect budget limit', async ({ page }) => {
    // 1. 设置预算限制
    // 2. 执行超出预算的工作流
    // 3. 验证预算阻止
  });

  test('should respect SLA timeout', async ({ page }) => {
    // 1. 设置 SLA
    // 2. 执行超时的工作流
    // 3. 验证超时升级
  });
});
```

### 10.4 并发测试

| 场景 | 验证点 |
|------|--------|
| 多用户同时执行同一 Workflow | 数据一致性、无竞态条件 |
| 多 Workflow 并行执行 | 资源隔离、预算独立 |
| Agent 同时处理多个 Issue | 任务队列正确、无丢失 |

---

## 十一、实施计划

### Phase 1 后端（12项）

| 序号 | 任务 | 文件 | 预估工作量 |
|------|------|------|-----------|
| 1 | 数据库迁移 | migrations/000015 | 1天 |
| 2 | AgentMember Service/Handler | service + handler | 1天 |
| 3 | IssueAgent Service/Handler | service + handler | 2天 |
| 4 | Workflow Service/Handler | service + handler | 3天 |
| 5 | Workflow 执行引擎 | service/workflow_executor.go | 3天 |
| 6 | ContextPayload Service | service/context_payload_service.go | 2天 |
| 7 | AgentDecision Service | service/agent_decision_service.go | 1天 |
| 8 | AgentCostBudget Service | service/agent_cost_budget_service.go | 1天 |
| 9 | AgentSLA Service | service/agent_sla_service.go | 1天 |
| 10 | 路由注册 | router/router.go | 1天 |
| 11 | 单元测试 | *_test.go | 2天 |
| 12 | 集成测试 | e2e tests | 1天 |

### Phase 1 前端（12项）

| 序号 | 任务 | 文件 | 预估工作量 |
|------|------|------|-----------|
| 1 | API 层 | api/*.ts | 1天 |
| 2 | AgentMemberList.vue | views/agents/ | 1天 |
| 3 | IssueAgentPanel.vue | views/issues/ | 2天 |
| 4 | IssueAgentStatus.vue | views/issues/ | 1天 |
| 5 | WorkflowList.vue | views/agents/ | 1天 |
| 6 | WorkflowCanvas.vue | components/workflow/ | 3天 |
| 7 | WorkflowNodeEditor.vue | components/agents/ | 1天 |
| 8 | WorkflowProperties.vue | components/workflow/ | 1天 |
| 9 | ExecutionPreview.vue | components/agents/ | 1天 |
| 10 | NewUserGuide.vue | components/agents/ | 1天 |
| 11 | 路由配置 | router/index.ts | 0.5天 |
| 12 | 样式和交互调试 | 各组件 | 1天 |

---

## 十二、技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Workflow 执行引擎复杂度高 | 开发周期延长 | 先实现串行执行，再支持并行 |
| 前端流程设计器实现复杂 | 交付风险 | 使用现成的流程图库（如 vue-flow） |
| 上下文数据量过大 | Token 消耗激增 | 实现上下文压缩和摘要 |
| 并发执行的竞态条件 | 数据不一致 | 使用数据库事务 + 乐观锁 |
| Agent 执行超时处理 | 资源泄漏 | 实现超时监控和自动清理 |
