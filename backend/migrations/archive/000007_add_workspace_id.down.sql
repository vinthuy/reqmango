ALTER TABLE labels DROP INDEX IF EXISTS idx_labels_workspace_id;
ALTER TABLE labels DROP COLUMN workspace_id;

ALTER TABLE states DROP INDEX IF EXISTS idx_states_workspace_id;
ALTER TABLE states DROP COLUMN workspace_id;

ALTER TABLE modules DROP INDEX IF EXISTS idx_modules_workspace_id;
ALTER TABLE modules DROP COLUMN workspace_id;