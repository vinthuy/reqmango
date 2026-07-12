ALTER TABLE labels ADD COLUMN workspace_id bigint;
UPDATE labels l SET workspace_id = (SELECT p.workspace_id FROM projects p WHERE p.id = l.project_id) WHERE l.project_id IS NOT NULL;
UPDATE labels l SET workspace_id = 1 WHERE l.project_id IS NULL AND l.workspace_id IS NULL;
ALTER TABLE labels ALTER COLUMN workspace_id SET NOT NULL;
ALTER TABLE labels ADD INDEX idx_labels_workspace_id (workspace_id);

ALTER TABLE states ADD COLUMN workspace_id bigint;
UPDATE states s SET workspace_id = (SELECT p.workspace_id FROM projects p WHERE p.id = s.project_id) WHERE s.project_id IS NOT NULL;
UPDATE states s SET workspace_id = 1 WHERE s.project_id IS NULL AND s.workspace_id IS NULL;
ALTER TABLE states ALTER COLUMN workspace_id SET NOT NULL;
ALTER TABLE states ADD INDEX idx_states_workspace_id (workspace_id);

ALTER TABLE modules ADD COLUMN workspace_id bigint;
UPDATE modules m SET workspace_id = (SELECT p.workspace_id FROM projects p WHERE p.id = m.project_id) WHERE m.project_id IS NOT NULL;
UPDATE modules m SET workspace_id = 1 WHERE m.project_id IS NULL AND m.workspace_id IS NULL;
ALTER TABLE modules ALTER COLUMN workspace_id SET NOT NULL;
ALTER TABLE modules ADD INDEX idx_modules_workspace_id (workspace_id);