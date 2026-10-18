-- Create automation_rule_overrides table for per-project overrides of inherited workspace rules
CREATE TABLE IF NOT EXISTS automation_rule_overrides (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by_id BIGINT,
    updated_by_id BIGINT,
    rule_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    is_enabled BOOLEAN,
    CONSTRAINT idx_aro_rule_project UNIQUE (rule_id, project_id)
);

CREATE INDEX IF NOT EXISTS idx_automation_rule_overrides_deleted_at ON automation_rule_overrides (deleted_at);
CREATE INDEX IF NOT EXISTS idx_automation_rule_overrides_rule_id ON automation_rule_overrides (rule_id);
CREATE INDEX IF NOT EXISTS idx_automation_rule_overrides_project_id ON automation_rule_overrides (project_id);
