-- Work item labels become project-only (aligned with Plane): purge workspace-level labels
DELETE FROM issue_labels WHERE label_id IN (SELECT id FROM labels WHERE project_id IS NULL);
DELETE FROM labels WHERE project_id IS NULL;
ALTER TABLE labels ALTER COLUMN project_id SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_project_name ON labels (project_id, name) WHERE deleted_at IS NULL;
