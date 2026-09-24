-- Drop the legacy FK that pinned state_transitions.workflow_id to the old
-- `workflows` table. Agent workflows (the only workflows left) use
-- `agent_workflows`, so the constraint actively prevents creating transitions
-- for any current workflow.  The NOT NULL + index remain, enforcing an id at
-- the application level.
ALTER TABLE state_transitions DROP CONSTRAINT IF EXISTS fk_workflows_transitions;
