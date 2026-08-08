-- 000018_tool_hardening.up.sql
-- Tool Calling 加固：MCP 外键 + category 默认值 + tool_call_logs 索引

-- 1) Tool -> MCPConfig 外键
ALTER TABLE tools ADD COLUMN IF NOT EXISTS mcp_config_id BIGINT
    CONSTRAINT fk_tools_mcp_config REFERENCES mcp_configs(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_tools_mcp_config_id ON tools(mcp_config_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tools_mcp_uniq ON tools(mcp_config_id, name)
    WHERE mcp_config_id IS NOT NULL;

-- 2) category 应用层约束：不做 DB enum，保持 varchar 灵活性，只补默认值
ALTER TABLE tools ALTER COLUMN category SET DEFAULT 'general';

-- 3) tool_call_logs 索引（按调用者、时间倒序常用筛选）
CREATE INDEX IF NOT EXISTS idx_tool_call_logs_agent_id ON tool_call_logs(agent_id);
CREATE INDEX IF NOT EXISTS idx_tool_call_logs_created_at ON tool_call_logs(created_at DESC);
