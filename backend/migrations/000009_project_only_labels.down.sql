DROP INDEX IF EXISTS idx_labels_project_name;
ALTER TABLE labels ALTER COLUMN project_id DROP NOT NULL;
