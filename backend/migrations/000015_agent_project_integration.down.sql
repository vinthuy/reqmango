-- 000015_agent_project_integration.down.sql
-- Agent-Project Integration 回滚迁移

-- 删除表（按依赖顺序）
DROP TABLE IF EXISTS agent_slas;
DROP TABLE IF EXISTS agent_cost_budgets;
DROP TABLE IF EXISTS agent_decisions;
DROP TABLE IF EXISTS workflow_node_runs;
DROP TABLE IF EXISTS workflow_runs;
DROP TABLE IF EXISTS workflow_edges;
DROP TABLE IF EXISTS workflow_nodes;
DROP TABLE IF EXISTS agent_workflows;

-- 删除 AgentTask 新增字段
ALTER TABLE agent_tasks DROP COLUMN IF EXISTS sla_breach;
ALTER TABLE agent_tasks DROP COLUMN IF EXISTS workflow_node_run_id;
ALTER TABLE agent_tasks DROP COLUMN IF EXISTS workflow_run_id;

-- 删除 Issue 新增字段
DROP INDEX IF EXISTS idx_issues_agent_assignee;
ALTER TABLE issues DROP COLUMN IF EXISTS agent_task_id;
ALTER TABLE issues DROP COLUMN IF EXISTS agent_assignee_id;
