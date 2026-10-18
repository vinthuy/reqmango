-- Personal Access Tokens for CLI / MCP / CI authentication
CREATE TABLE personal_access_tokens (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by_id BIGINT,
    updated_by_id BIGINT,

    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    token_prefix VARCHAR(20) NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    scopes TEXT DEFAULT '',
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_personal_access_tokens_token_hash ON personal_access_tokens(token_hash);
CREATE INDEX idx_personal_access_tokens_user_id ON personal_access_tokens(user_id);
