CREATE TABLE IF NOT EXISTS automation_executions (
    id bigserial PRIMARY KEY,
    rule_id bigint NOT NULL,
    issue_id bigint NOT NULL,
    trigger_type varchar(50) NOT NULL,
    context_json jsonb,
    actions_taken jsonb,
    status varchar(20) NOT NULL,
    error text,
    duration bigint NOT NULL DEFAULT 0,
    executed_at timestamp with time zone NOT NULL DEFAULT NOW(),
    created_at timestamp with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp with time zone NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_automation_executions_rule_id ON automation_executions(rule_id);
CREATE INDEX idx_automation_executions_issue_id ON automation_executions(issue_id);
CREATE INDEX idx_automation_executions_executed_at ON automation_executions(executed_at);
