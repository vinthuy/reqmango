-- Drop FK constraint on automation_rules.project_id
-- Workspace-level rules (project_id = 0) should not reference any specific project
ALTER TABLE automation_rules DROP CONSTRAINT IF EXISTS fk_automation_rules_project;
