-- Re-adding the constraint requires the `workflows` table to still exist.
-- If the legacy table has already been dropped, this migration will fail;
-- in that case the down-migration is intentionally a no-op.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'workflows') THEN
    ALTER TABLE state_transitions
      ADD CONSTRAINT fk_workflows_transitions
      FOREIGN KEY (workflow_id) REFERENCES workflows(id);
  END IF;
END $$;
