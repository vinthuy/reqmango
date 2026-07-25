-- Agent Templates Table
CREATE TABLE agent_templates (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by_id BIGINT,
    updated_by_id BIGINT,

    name VARCHAR(128) NOT NULL,
    description TEXT,
    is_preset BOOLEAN DEFAULT FALSE,
    icon VARCHAR(10) DEFAULT '🤖',
    system_prompt TEXT NOT NULL,
    available_skills JSONB DEFAULT '[]'::jsonb,
    available_tools JSONB DEFAULT '[]'::jsonb,
    default_config JSONB DEFAULT '{}'::jsonb,
    version VARCHAR(50) DEFAULT '1.0',
    status VARCHAR(20) DEFAULT 'active',
    workspace_id BIGINT
);

CREATE UNIQUE INDEX idx_agent_templates_name ON agent_templates(name);
CREATE INDEX idx_agent_templates_workspace_id ON agent_templates(workspace_id);
CREATE INDEX idx_agent_templates_is_preset ON agent_templates(is_preset);
CREATE INDEX idx_agent_templates_status ON agent_templates(status);
CREATE INDEX idx_agent_templates_deleted_at ON agent_templates(deleted_at);

-- Agent Configs Table
CREATE TABLE agent_configs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by_id BIGINT,
    updated_by_id BIGINT,

    name VARCHAR(128) NOT NULL,
    description TEXT,
    provider VARCHAR(30) NOT NULL,
    model VARCHAR(100) NOT NULL,
    api_key VARCHAR(500) NOT NULL,
    api_endpoint VARCHAR(255),
    inference_level VARCHAR(20) DEFAULT 'normal',
    service_level VARCHAR(20) DEFAULT 'standard',
    max_tokens INTEGER DEFAULT 4096,
    temperature DECIMAL(3,2) DEFAULT 0.7,
    top_p DECIMAL(3,2) DEFAULT 1.0,
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    workspace_id BIGINT NOT NULL
);

CREATE INDEX idx_agent_configs_workspace_id ON agent_configs(workspace_id);
CREATE INDEX idx_agent_configs_is_default ON agent_configs(is_default);
CREATE INDEX idx_agent_configs_is_active ON agent_configs(is_active);
CREATE INDEX idx_agent_configs_provider ON agent_configs(provider);
CREATE INDEX idx_agent_configs_deleted_at ON agent_configs(deleted_at);

-- Runtimes Table
CREATE TABLE runtimes (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by_id BIGINT,
    updated_by_id BIGINT,

    name VARCHAR(128) NOT NULL,
    runtime_type VARCHAR(20) NOT NULL,
    runtime_mode VARCHAR(20) DEFAULT 'pull',
    endpoint VARCHAR(255),
    status VARCHAR(20) DEFAULT 'offline',
    capacity INTEGER DEFAULT 1,
    current_load INTEGER DEFAULT 0,
    version VARCHAR(50),
    host_info JSONB,
    last_heartbeat TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb,
    workspace_id BIGINT NOT NULL
);

CREATE INDEX idx_runtimes_workspace_id ON runtimes(workspace_id);
CREATE INDEX idx_runtimes_status ON runtimes(status);
CREATE INDEX idx_runtimes_runtime_type ON runtimes(runtime_type);
CREATE INDEX idx_runtimes_deleted_at ON runtimes(deleted_at);

-- Skills Table
CREATE TABLE skills (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by_id BIGINT,
    updated_by_id BIGINT,

    name VARCHAR(128) NOT NULL,
    description TEXT,
    skill_type VARCHAR(30) DEFAULT 'custom',
    version VARCHAR(50) DEFAULT '1.0',
    status VARCHAR(20) DEFAULT 'active',
    skill_md TEXT NOT NULL,
    parameters JSONB DEFAULT '[]'::jsonb,
    tags JSONB DEFAULT '[]'::jsonb,
    usage_count INTEGER DEFAULT 0,
    is_shared BOOLEAN DEFAULT FALSE,
    workspace_id BIGINT NOT NULL
);

CREATE INDEX idx_skills_workspace_id ON skills(workspace_id);
CREATE INDEX idx_skills_skill_type ON skills(skill_type);
CREATE INDEX idx_skills_status ON skills(status);
CREATE INDEX idx_skills_is_shared ON skills(is_shared);
CREATE INDEX idx_skills_deleted_at ON skills(deleted_at);

-- Agent Tasks Table
CREATE TABLE agent_tasks (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by_id BIGINT,
    updated_by_id BIGINT,

    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'enqueue',
    priority VARCHAR(20) DEFAULT 'normal',
    progress INTEGER DEFAULT 0,
    task_type VARCHAR(50),
    input_data JSONB,
    output_data JSONB,
    error_info TEXT,
    logs JSONB DEFAULT '[]'::jsonb,

    agent_template_id BIGINT,
    agent_config_id BIGINT,
    runtime_id BIGINT,

    workspace_id BIGINT NOT NULL,
    project_id BIGINT,
    issue_id BIGINT,

    enqueued_at TIMESTAMPTZ DEFAULT NOW(),
    claimed_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    estimated_time INTEGER,
    actual_time INTEGER
);

CREATE INDEX idx_agent_tasks_workspace_id ON agent_tasks(workspace_id);
CREATE INDEX idx_agent_tasks_project_id ON agent_tasks(project_id);
CREATE INDEX idx_agent_tasks_issue_id ON agent_tasks(issue_id);
CREATE INDEX idx_agent_tasks_status ON agent_tasks(status);
CREATE INDEX idx_agent_tasks_priority ON agent_tasks(priority);
CREATE INDEX idx_agent_tasks_agent_template_id ON agent_tasks(agent_template_id);
CREATE INDEX idx_agent_tasks_agent_config_id ON agent_tasks(agent_config_id);
CREATE INDEX idx_agent_tasks_runtime_id ON agent_tasks(runtime_id);
CREATE INDEX idx_agent_tasks_enqueued_at ON agent_tasks(enqueued_at);
CREATE INDEX idx_agent_tasks_deleted_at ON agent_tasks(deleted_at);

-- Task Logs Table
CREATE TABLE task_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    task_id BIGINT NOT NULL,
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB
);

CREATE INDEX idx_task_logs_task_id ON task_logs(task_id);
CREATE INDEX idx_task_logs_level ON task_logs(level);
CREATE INDEX idx_task_logs_created_at ON task_logs(created_at);

-- Foreign keys
ALTER TABLE agent_templates ADD CONSTRAINT fk_agent_templates_workspace_id FOREIGN KEY (workspace_id) REFERENCES workspaces(id);
ALTER TABLE agent_configs ADD CONSTRAINT fk_agent_configs_workspace_id FOREIGN KEY (workspace_id) REFERENCES workspaces(id);
ALTER TABLE runtimes ADD CONSTRAINT fk_runtimes_workspace_id FOREIGN KEY (workspace_id) REFERENCES workspaces(id);
ALTER TABLE skills ADD CONSTRAINT fk_skills_workspace_id FOREIGN KEY (workspace_id) REFERENCES workspaces(id);
ALTER TABLE agent_tasks ADD CONSTRAINT fk_agent_tasks_workspace_id FOREIGN KEY (workspace_id) REFERENCES workspaces(id);
ALTER TABLE agent_tasks ADD CONSTRAINT fk_agent_tasks_project_id FOREIGN KEY (project_id) REFERENCES projects(id);
ALTER TABLE agent_tasks ADD CONSTRAINT fk_agent_tasks_issue_id FOREIGN KEY (issue_id) REFERENCES issues(id);
ALTER TABLE agent_tasks ADD CONSTRAINT fk_agent_tasks_agent_template_id FOREIGN KEY (agent_template_id) REFERENCES agent_templates(id);
ALTER TABLE agent_tasks ADD CONSTRAINT fk_agent_tasks_agent_config_id FOREIGN KEY (agent_config_id) REFERENCES agent_configs(id);
ALTER TABLE agent_tasks ADD CONSTRAINT fk_agent_tasks_runtime_id FOREIGN KEY (runtime_id) REFERENCES runtimes(id);
ALTER TABLE task_logs ADD CONSTRAINT fk_task_logs_task_id FOREIGN KEY (task_id) REFERENCES agent_tasks(id);

-- Insert preset agent templates
INSERT INTO agent_templates (name, description, is_preset, icon, system_prompt, available_skills, available_tools, default_config, version, status, workspace_id) VALUES
('Requirements Analyst', '专业需求分析师，擅长分析用户需求、生成需求规格说明书、提取验收标准', true, '📊', '你是一名专业的需求分析师。你的任务是分析用户需求，生成清晰、完整的需求规格说明书，包括功能需求、非功能需求、验收标准和用户故事。', '["analyze_requirement", "generate_prd", "extract_acceptance_criteria"]', '["reqmango_api"]', '{"provider": "deepseek", "model": "deepseek-chat", "temperature": 0.7}', '1.0', 'active', NULL),
('Developer', '专业开发工程师，擅长代码生成、代码审查、技术实现', true, '👨‍💻', '你是一名专业的开发工程师。你的任务是根据需求规格说明书生成高质量的代码，遵循最佳实践和编码规范。', '["generate_code", "code_review", "fix_bug", "refactor_code"]', '["github_api", "reqmango_api"]', '{"provider": "deepseek", "model": "deepseek-code", "temperature": 0.5}', '1.0', 'active', NULL),
('Tester', '专业测试工程师，擅长测试用例设计、执行测试、报告缺陷', true, '🧪', '你是一名专业的测试工程师。你的任务是设计全面的测试用例，执行测试，发现并报告缺陷。', '["generate_test_cases", "execute_test", "report_bug"]', '["reqmango_api"]', '{"provider": "deepseek", "model": "deepseek-chat", "temperature": 0.7}', '1.0', 'active', NULL),
('Code Reviewer', '专业代码审查员，擅长审查代码质量、安全性和性能', true, '🔍', '你是一名专业的代码审查员。你的任务是审查代码，确保代码质量、安全性和性能符合标准。', '["code_review", "security_review", "performance_analysis"]', '["github_api"]', '{"provider": "deepseek", "model": "deepseek-code", "temperature": 0.3}', '1.0', 'active', NULL),
('Product Manager', '专业产品经理，擅长产品规划、路线图制定、优先级排序', true, '🚀', '你是一名专业的产品经理。你的任务是规划产品路线图，制定优先级，确保产品愿景的实现。', '["product_planning", "roadmap_creation", "prioritization"]', '["reqmango_api"]', '{"provider": "deepseek", "model": "deepseek-chat", "temperature": 0.7}', '1.0', 'active', NULL),
('UX Designer', '专业用户体验设计师，擅长界面设计、交互设计、用户研究', true, '🎨', '你是一名专业的用户体验设计师。你的任务是设计用户界面，优化交互体验，进行用户研究。', '["ui_design", "interaction_design", "user_research"]', '[]', '{"provider": "deepseek", "model": "deepseek-chat", "temperature": 0.8}', '1.0', 'active', NULL),
('DevOps Engineer', '专业DevOps工程师，擅长CI/CD配置、自动化部署、基础设施管理', true, '🔧', '你是一名专业的DevOps工程师。你的任务是配置CI/CD流水线，自动化部署，管理基础设施。', '["ci_cd_config", "automated_deployment", "infrastructure_management"]', '["github_actions", "docker"]', '{"provider": "deepseek", "model": "deepseek-chat", "temperature": 0.6}', '1.0', 'active', NULL),
('Technical Writer', '专业技术文档工程师，擅长编写API文档、用户手册、技术白皮书', true, '📝', '你是一名专业的技术文档工程师。你的任务是编写清晰、准确的技术文档。', '["write_api_docs", "write_user_manual", "write_whitepaper"]', '[]', '{"provider": "deepseek", "model": "deepseek-chat", "temperature": 0.7}', '1.0', 'active', NULL),
('QA Engineer', '专业质量保证工程师，擅长质量控制、测试流程优化、自动化测试', true, '✅', '你是一名专业的质量保证工程师。你的任务是确保产品质量，优化测试流程，实现自动化测试。', '["quality_control", "test_process_optimization", "automation_test"]', '["reqmango_api"]', '{"provider": "deepseek", "model": "deepseek-chat", "temperature": 0.7}', '1.0', 'active', NULL);
