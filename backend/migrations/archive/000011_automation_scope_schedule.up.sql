-- Add project scope for workspace-level automation rules
-- 'all' = applies to all projects in the workspace
-- JSON array like '[1,2,3]' = applies to specific project IDs only
ALTER TABLE automation_rules ADD COLUMN IF NOT EXISTS scope varchar(50) DEFAULT 'all';
CREATE INDEX IF NOT EXISTS idx_automation_rules_scope ON automation_rules(scope);

-- Add scheduled trigger support
-- schedule_config stores JSON: {"frequency":"daily","time":"09:00","days":["mon","wed","fri"]}
-- or a cron expression string in value
ALTER TABLE automation_rules ADD COLUMN IF NOT EXISTS schedule_config text;
-- Track last scheduled execution to avoid duplicate triggers
ALTER TABLE automation_rules ADD COLUMN IF NOT EXISTS last_triggered_at timestamp with time zone;
