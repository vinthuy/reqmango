-- Reverts 000002: restores the discriminator and the archived transitions.
-- Development escape hatch only; the restored agent rows immediately violate
-- the model's single-parent assumption until the code is rolled back too.

ALTER TABLE state_transitions
    ADD COLUMN IF NOT EXISTS workflow_kind VARCHAR(16) NOT NULL DEFAULT 'state';

-- The foreign key cannot hold for restored agent rows, so drop it under every
-- name this column has carried.
ALTER TABLE state_transitions DROP CONSTRAINT IF EXISTS fk_workflows_transitions;
ALTER TABLE state_transitions DROP CONSTRAINT IF EXISTS fk_state_transitions_workflow;

INSERT INTO state_transitions (
    id, created_at, updated_at, deleted_at, created_by_id, updated_by_id,
    name, description, workflow_id, workflow_kind, source_state_id, target_state_id,
    issue_type_id, is_auto, rule_type, approver_ids, role_allowed,
    approve_target_state_id, reject_target_state_id, approval_mode, project_id, workspace_id
)
SELECT
    id, created_at, updated_at, deleted_at, created_by_id, updated_by_id,
    name, description, workflow_id, workflow_kind, source_state_id, target_state_id,
    issue_type_id, is_auto, rule_type, approver_ids, role_allowed,
    approve_target_state_id, reject_target_state_id, approval_mode, project_id, workspace_id
FROM state_transitions_agent_archive
ON CONFLICT (id) DO NOTHING;

DROP TABLE IF EXISTS state_transitions_agent_archive;
