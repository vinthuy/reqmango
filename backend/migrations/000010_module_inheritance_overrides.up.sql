CREATE TABLE IF NOT EXISTS module_inheritance_overrides (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL,
    workspace_module_id BIGINT NOT NULL,
    is_excluded BOOLEAN DEFAULT FALSE,
    override_name VARCHAR(100),
    override_description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, workspace_module_id)
);

CREATE INDEX IF NOT EXISTS idx_module_inheritance_overrides_project ON module_inheritance_overrides(project_id);
CREATE INDEX IF NOT EXISTS idx_module_inheritance_overrides_workspace_module ON module_inheritance_overrides(workspace_module_id);