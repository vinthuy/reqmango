DROP TABLE IF EXISTS approval_records;
DROP TABLE IF EXISTS approvals;

ALTER TABLE state_transitions
  DROP COLUMN IF EXISTS approve_target_state_id,
  DROP COLUMN IF EXISTS reject_target_state_id,
  DROP COLUMN IF EXISTS approval_mode;

ALTER TABLE issues
  DROP COLUMN IF EXISTS approval_status,
  DROP COLUMN IF EXISTS active_approval_id;
