DROP INDEX IF EXISTS idx_state_transitions_workflow_kind;
ALTER TABLE state_transitions DROP COLUMN IF EXISTS workflow_kind;
