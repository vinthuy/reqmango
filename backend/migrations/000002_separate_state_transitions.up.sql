-- 000002: state_transitions belongs to the state machine only.
--
-- `workflows` (state machine) and `agent_workflows` (orchestration) share the
-- same id range, so `workflow_kind` was added to tell them apart. The column
-- turned out to be a stopgap: the live table keeps state-machine transitions
-- only, while orchestration keeps its own children (workflow_nodes /
-- workflow_edges / workflow_runs). Archived rows are preserved rather than
-- dropped so a rollback and any post-deploy investigation can recover them.

-- 1. Archive table: columns copied from the live table, plus an archived_at
--    stamp. The primary key is added explicitly because LIKE ... INCLUDING
--    DEFAULTS does not copy constraints, and the down migration needs it for
--    ON CONFLICT.
CREATE TABLE IF NOT EXISTS state_transitions_agent_archive (
    LIKE state_transitions INCLUDING DEFAULTS
);
ALTER TABLE state_transitions_agent_archive
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ DEFAULT now();

DO $migration$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'state_transitions_agent_archive_pkey'
    ) THEN
        ALTER TABLE state_transitions_agent_archive
            ADD CONSTRAINT state_transitions_agent_archive_pkey PRIMARY KEY (id);
    END IF;
END
$migration$;

-- 2. Move agent-kind rows out of the live table.
INSERT INTO state_transitions_agent_archive (
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
FROM state_transitions
WHERE workflow_kind = 'agent'
ON CONFLICT (id) DO NOTHING;

-- 3. A state transition whose workflow row is gone cannot satisfy the foreign
--    key restored below. Archive those too instead of deleting them, so the
--    data stays inspectable.
INSERT INTO state_transitions_agent_archive (
    id, created_at, updated_at, deleted_at, created_by_id, updated_by_id,
    name, description, workflow_id, workflow_kind, source_state_id, target_state_id,
    issue_type_id, is_auto, rule_type, approver_ids, role_allowed,
    approve_target_state_id, reject_target_state_id, approval_mode, project_id, workspace_id
)
SELECT
    t.id, t.created_at, t.updated_at, t.deleted_at, t.created_by_id, t.updated_by_id,
    t.name, t.description, t.workflow_id, t.workflow_kind, t.source_state_id, t.target_state_id,
    t.issue_type_id, t.is_auto, t.rule_type, t.approver_ids, t.role_allowed,
    t.approve_target_state_id, t.reject_target_state_id, t.approval_mode, t.project_id, t.workspace_id
FROM state_transitions t
WHERE t.workflow_kind <> 'agent'
  AND NOT EXISTS (SELECT 1 FROM workflows w WHERE w.id = t.workflow_id)
ON CONFLICT (id) DO NOTHING;

DELETE FROM state_transitions
WHERE workflow_kind = 'agent'
   OR NOT EXISTS (SELECT 1 FROM workflows w WHERE w.id = state_transitions.workflow_id);

-- 4. One parent only.
ALTER TABLE state_transitions DROP COLUMN IF EXISTS workflow_kind;

-- 5. Restore the foreign key. The constraint name matches what GORM derives
--    from the model relationship, so AutoMigrate on an already-migrated
--    database leaves it alone instead of adding a second one.
DO $migration$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint c
        WHERE c.conrelid = 'state_transitions'::regclass
          AND c.contype = 'f'
          AND c.confrelid = 'workflows'::regclass
    ) THEN
        ALTER TABLE state_transitions
            ADD CONSTRAINT fk_workflows_transitions
            FOREIGN KEY (workflow_id) REFERENCES workflows(id);
    END IF;
END
$migration$;
