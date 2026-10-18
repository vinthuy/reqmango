-- 000015_agent_project_integration.up.sql
-- Agent-Project Integration 数据库迁移

-- 0. ProjectAgentMember 项目 Agent 成员关系
CREATE TABLE IF NOT EXISTS project_agent_members (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id),
    agent_id BIGINT NOT NULL REFERENCES agents(id),
    role VARCHAR(20) DEFAULT 'member',
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(project_id, agent_id)
);
CREATE INDEX IF NOT EXISTS idx_project_agent_members_project ON project_agent_members(project_id);

-- 1. Issue 增加 Agent 字段
ALTER TABLE issues ADD COLUMN IF NOT EXISTS agent_assignee_id BIGINT REFERENCES agents(id);
ALTER TABLE issues ADD COLUMN IF NOT EXISTS agent_task_id BIGINT REFERENCES agent_tasks(id);
CREATE INDEX IF NOT EXISTS idx_issues_agent_assignee ON issues(agent_assignee_id);

-- 2. AgentTask 增加 Workflow 字段
ALTER TABLE agent_tasks ADD COLUMN IF NOT EXISTS workflow_run_id BIGINT;
ALTER TABLE agent_tasks ADD COLUMN IF NOT EXISTS workflow_node_run_id BIGINT;
ALTER TABLE agent_tasks ADD COLUMN IF NOT EXISTS sla_breach BOOLEAN DEFAULT FALSE;

-- 3. AgentWorkflow 工作流定义
CREATE TABLE IF NOT EXISTS agent_workflows (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    project_id BIGINT NOT NULL REFERENCES projects(id),
    workspace_id BIGINT NOT NULL,
    version INT DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    trigger_type VARCHAR(20) DEFAULT 'manual',
    trigger_config JSONB,
    config JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    created_by_id BIGINT,
    updated_by_id BIGINT
);
CREATE INDEX IF NOT EXISTS idx_workflows_project ON agent_workflows(project_id);
CREATE INDEX IF NOT EXISTS idx_workflows_workspace ON agent_workflows(workspace_id);

-- 4. WorkflowNode 流程节点
CREATE TABLE IF NOT EXISTS workflow_nodes (
    id BIGSERIAL PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES agent_workflows(id) ON DELETE CASCADE,
    agent_id BIGINT NOT NULL,
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
CREATE INDEX IF NOT EXISTS idx_workflow_nodes_workflow ON workflow_nodes(workflow_id);

-- 5. WorkflowEdge 流程连接
CREATE TABLE IF NOT EXISTS workflow_edges (
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
CREATE INDEX IF NOT EXISTS idx_workflow_edges_workflow ON workflow_edges(workflow_id);

-- 6. WorkflowRun 工作流执行记录
CREATE TABLE IF NOT EXISTS workflow_runs (
    id BIGSERIAL PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES agent_workflows(id),
    issue_id BIGINT,
    status VARCHAR(20) DEFAULT 'pending',
    current_node_id BIGINT,
    context JSONB,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    total_tokens INT DEFAULT 0,
    total_cost DECIMAL(10,4) DEFAULT 0,
    error_info TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    created_by_id BIGINT,
    updated_by_id BIGINT
);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow ON workflow_runs(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_issue ON workflow_runs(issue_id);

-- 7. WorkflowNodeRun 节点执行记录
CREATE TABLE IF NOT EXISTS workflow_node_runs (
    id BIGSERIAL PRIMARY KEY,
    workflow_run_id BIGINT NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    node_id BIGINT NOT NULL REFERENCES workflow_nodes(id),
    agent_id BIGINT NOT NULL,
    agent_task_id BIGINT,
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
CREATE INDEX IF NOT EXISTS idx_workflow_node_runs_workflow_run ON workflow_node_runs(workflow_run_id);

-- 8. AgentDecision Agent 决策记录
CREATE TABLE IF NOT EXISTS agent_decisions (
    id BIGSERIAL PRIMARY KEY,
    agent_id BIGINT NOT NULL,
    issue_id BIGINT,
    agent_task_id BIGINT,
    workflow_run_id BIGINT,
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
CREATE INDEX IF NOT EXISTS idx_agent_decisions_agent ON agent_decisions(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_decisions_issue ON agent_decisions(issue_id);

-- 9. AgentCostBudget 项目级成本预算
CREATE TABLE IF NOT EXISTS agent_cost_budgets (
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
CREATE TABLE IF NOT EXISTS agent_slas (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL UNIQUE REFERENCES projects(id),
    normal_task_max INT DEFAULT 1800,
    complex_task_max INT DEFAULT 7200,
    auto_escalation BOOLEAN DEFAULT TRUE,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
