ALTER TABLE automation_rules DROP COLUMN IF EXISTS last_triggered_at;
ALTER TABLE automation_rules DROP COLUMN IF EXISTS schedule_config;
ALTER TABLE automation_rules DROP INDEX IF EXISTS idx_automation_rules_scope;
ALTER TABLE automation_rules DROP COLUMN IF EXISTS scope;
