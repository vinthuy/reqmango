-- Extend state_transitions
ALTER TABLE state_transitions
  ADD COLUMN approve_target_state_id BIGINT,
  ADD COLUMN reject_target_state_id BIGINT,
  ADD COLUMN approval_mode VARCHAR(20) DEFAULT 'any';

-- New approvals table
CREATE TABLE approvals (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by_id BIGINT,
    updated_by_id BIGINT,

    issue_id BIGINT NOT NULL,
    workflow_id BIGINT NOT NULL,
    transition_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    workspace_id BIGINT NOT NULL,

    requester_id BIGINT NOT NULL,
    request_note TEXT,

    source_state_id BIGINT NOT NULL,
    approve_target_state_id BIGINT NOT NULL,
    reject_target_state_id BIGINT NOT NULL,

    approver_ids JSONB NOT NULL DEFAULT '[]'::jsonb,

    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    decided_by BIGINT,
    decided_at TIMESTAMPTZ,
    decision_note TEXT
);

CREATE INDEX idx_approvals_issue_id ON approvals(issue_id);
CREATE INDEX idx_approvals_status ON approvals(status);
CREATE INDEX idx_approvals_project_id ON approvals(project_id);
CREATE INDEX idx_approvals_workspace_id ON approvals(workspace_id);
CREATE INDEX idx_approvals_requester_id ON approvals(requester_id);
CREATE INDEX idx_approvals_deleted_at ON approvals(deleted_at);

-- New approval_records table
CREATE TABLE approval_records (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    approval_id BIGINT NOT NULL,
    approver_id BIGINT NOT NULL,
    decision VARCHAR(20) NOT NULL,
    note TEXT,
    decided_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_approval_records_approval_id ON approval_records(approval_id);
CREATE INDEX idx_approval_records_approver_id ON approval_records(approver_id);

-- Extend issues table
ALTER TABLE issues
  ADD COLUMN approval_status VARCHAR(20),
  ADD COLUMN active_approval_id BIGINT;

CREATE INDEX idx_issues_approval_status ON issues(approval_status);

-- Data migration: convert existing approver_ids from comma-separated to JSON array
UPDATE state_transitions
SET approver_ids = '[' || REPLACE(approver_ids, ',', ',') || ']'
WHERE rule_type = 'approval'
  AND approver_ids IS NOT NULL
  AND approver_ids != ''
  AND approver_ids NOT LIKE '[%';
