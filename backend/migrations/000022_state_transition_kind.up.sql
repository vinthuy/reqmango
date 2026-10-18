-- state_transitions.workflow_id points at two independent tables whose id
-- ranges overlap: `workflows` (state machines, read by the transition rules
-- engine) and `agent_workflows` (what the workflow UI manages). Filtering on
-- workflow_id alone therefore mixed the two: a transition created for agent
-- workflow 65 also appeared under state workflow 65, and vice versa.
--
-- Add an explicit discriminator and backfill it.
ALTER TABLE state_transitions
    ADD COLUMN IF NOT EXISTS workflow_kind varchar(16) NOT NULL DEFAULT 'state';

-- Rows whose workflow only exists as an agent workflow are agent transitions.
UPDATE state_transitions t
SET workflow_kind = 'agent'
WHERE EXISTS (SELECT 1 FROM agent_workflows a WHERE a.id = t.workflow_id)
  AND NOT EXISTS (SELECT 1 FROM workflows w WHERE w.id = t.workflow_id);

-- When both tables have the id, the project decides: a row whose project only
-- matches the agent workflow was created through the agent-workflow UI.
UPDATE state_transitions t
SET workflow_kind = 'agent'
WHERE EXISTS (SELECT 1 FROM agent_workflows a
              WHERE a.id = t.workflow_id AND COALESCE(a.project_id, 0) = COALESCE(t.project_id, 0))
  AND NOT EXISTS (SELECT 1 FROM workflows w
                  WHERE w.id = t.workflow_id AND COALESCE(w.project_id, 0) = COALESCE(t.project_id, 0));

CREATE INDEX IF NOT EXISTS idx_state_transitions_workflow_kind
    ON state_transitions (workflow_id, workflow_kind);
