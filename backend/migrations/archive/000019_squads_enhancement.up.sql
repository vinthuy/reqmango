-- Squads multi-agent enhancement: async execution + cancel + retry + permissions

-- 1) squad_executions: add cancel fields
ALTER TABLE squad_executions ADD COLUMN cancelled_at TIMESTAMP NULL;
ALTER TABLE squad_executions ADD COLUMN cancel_reason TEXT DEFAULT '';

-- 2) squad_tasks: add retry + error fields; relax agent_task_id
ALTER TABLE squad_tasks ALTER COLUMN agent_task_id DROP NOT NULL;
ALTER TABLE squad_tasks ADD COLUMN retry_count INT DEFAULT 0;
ALTER TABLE squad_tasks ADD COLUMN max_retries INT DEFAULT 2;
ALTER TABLE squad_tasks ADD COLUMN error_message TEXT DEFAULT '';

-- 3) Indexes for filtering by status
CREATE INDEX IF NOT EXISTS idx_squads_status ON squads(status);
CREATE INDEX IF NOT EXISTS idx_squad_executions_status ON squad_executions(status);
