-- Baseline schema (squashed 2026-10-18).
--
-- Regenerate with:
--   pg_dump --schema-only --no-owner --no-privileges --no-comments <db> \
--     | python3 scripts/idempotent_baseline.py > backend/migrations/000001_baseline.up.sql
--
-- It is written to be re-runnable: an environment created by the server's
-- AutoMigrate already has every object, and adopting migrations there must be
-- a no-op rather than an error.

--
-- PostgreSQL database dump
--


-- Dumped from database version 18.4
-- Dumped by pg_dump version 18.4




--
-- Name: agent_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_activities (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    agent_id bigint NOT NULL,
    issue_id bigint,
    action character varying(50) NOT NULL,
    result_summary text,
    rating bigint,
    executed_at timestamp with time zone,
    agent_name character varying(128),
    task_context text,
    tool_calls jsonb
);


--
-- Name: agent_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_activities_id_seq OWNED BY public.agent_activities.id;


--
-- Name: agent_configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_configs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(128) NOT NULL,
    description text,
    provider character varying(30) NOT NULL,
    model character varying(100) NOT NULL,
    api_key character varying(500) NOT NULL,
    api_endpoint character varying(255),
    inference_level character varying(20) DEFAULT 'normal'::character varying,
    service_level character varying(20) DEFAULT 'standard'::character varying,
    max_tokens bigint DEFAULT 4096,
    temperature numeric(3,2) DEFAULT 0.7,
    top_p numeric(3,2) DEFAULT 1,
    is_default boolean DEFAULT false,
    is_active boolean DEFAULT true,
    workspace_id bigint NOT NULL
);


--
-- Name: agent_configs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_configs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_configs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_configs_id_seq OWNED BY public.agent_configs.id;


--
-- Name: agent_cost_budgets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_cost_budgets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    project_id bigint NOT NULL,
    monthly_budget numeric,
    current_cost numeric,
    alert_threshold numeric,
    auto_block boolean DEFAULT false,
    last_reset_at timestamp with time zone
);


--
-- Name: agent_cost_budgets_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_cost_budgets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_cost_budgets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_cost_budgets_id_seq OWNED BY public.agent_cost_budgets.id;


--
-- Name: agent_decisions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_decisions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    agent_id bigint NOT NULL,
    issue_id bigint,
    agent_task_id bigint,
    workflow_run_id bigint,
    node_type character varying(50),
    thinking text,
    decision text,
    reasoning text,
    alternatives jsonb,
    confidence numeric
);


--
-- Name: agent_decisions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_decisions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_decisions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_decisions_id_seq OWNED BY public.agent_decisions.id;


--
-- Name: agent_loop_iterations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_loop_iterations (
    id bigint NOT NULL,
    loop_run_id bigint NOT NULL,
    iteration_num bigint NOT NULL,
    action_taken jsonb NOT NULL,
    result_observed jsonb NOT NULL,
    reasoning text,
    decision character varying(50) NOT NULL,
    tokens_used bigint DEFAULT 0,
    duration_ms bigint,
    created_at timestamp with time zone
);


--
-- Name: agent_loop_iterations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_loop_iterations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_loop_iterations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_loop_iterations_id_seq OWNED BY public.agent_loop_iterations.id;


--
-- Name: agent_loop_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_loop_runs (
    id bigint NOT NULL,
    loop_id bigint NOT NULL,
    status character varying(50) DEFAULT 'running'::character varying,
    current_iteration bigint DEFAULT 0,
    max_iterations bigint DEFAULT 100,
    goal text NOT NULL,
    goal_metrics jsonb,
    working_memory jsonb DEFAULT '{}'::jsonb,
    tokens_used bigint DEFAULT 0,
    cost_usd numeric(10,4) DEFAULT 0,
    stopped_reason character varying(100),
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: agent_loop_runs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_loop_runs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_loop_runs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_loop_runs_id_seq OWNED BY public.agent_loop_runs.id;


--
-- Name: agent_loops; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_loops (
    id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    loop_def jsonb NOT NULL,
    version character varying(50) DEFAULT '1.0'::character varying,
    status character varying(50) DEFAULT 'active'::character varying,
    created_by_id bigint,
    updated_by_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: agent_loops_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_loops_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_loops_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_loops_id_seq OWNED BY public.agent_loops.id;


--
-- Name: agent_pipeline_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_pipeline_runs (
    id bigint NOT NULL,
    pipeline_id bigint NOT NULL,
    trigger_type character varying(50),
    trigger_context jsonb,
    status character varying(50) DEFAULT 'pending'::character varying,
    stages_result jsonb,
    tokens_used bigint DEFAULT 0,
    cost_usd numeric(10,4) DEFAULT 0,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    error_message text,
    created_at timestamp with time zone
);


--
-- Name: agent_pipeline_runs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_pipeline_runs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_pipeline_runs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_pipeline_runs_id_seq OWNED BY public.agent_pipeline_runs.id;


--
-- Name: agent_pipelines; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_pipelines (
    id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    pipeline_def jsonb NOT NULL,
    version character varying(50) DEFAULT '1.0'::character varying,
    status character varying(50) DEFAULT 'active'::character varying,
    created_by_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: agent_pipelines_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_pipelines_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_pipelines_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_pipelines_id_seq OWNED BY public.agent_pipelines.id;


--
-- Name: agent_registry; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_registry (
    id bigint NOT NULL,
    workspace_id bigint,
    name character varying(255) NOT NULL,
    display_name character varying(255),
    description text,
    capabilities jsonb DEFAULT '[]'::jsonb NOT NULL,
    agent_def jsonb NOT NULL,
    version character varying(50) DEFAULT '1.0.0'::character varying,
    author character varying(255),
    is_verified boolean DEFAULT false,
    installs bigint DEFAULT 0,
    rating numeric(3,2),
    source character varying(50) DEFAULT 'local'::character varying,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: agent_registry_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_registry_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_registry_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_registry_id_seq OWNED BY public.agent_registry.id;


--
-- Name: agent_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_sessions (
    id character varying(64) NOT NULL,
    workspace_id bigint NOT NULL,
    agent_type character varying(50) NOT NULL,
    agent_ref character varying(255),
    status character varying(50) DEFAULT 'running'::character varying,
    model_used character varying(100),
    input_summary text,
    output_summary text,
    tokens_input bigint DEFAULT 0,
    tokens_output bigint DEFAULT 0,
    cost_usd numeric(10,4) DEFAULT 0,
    tools_called jsonb DEFAULT '[]'::jsonb,
    error_message text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    updated_at timestamp with time zone,
    metadata jsonb DEFAULT '{}'::jsonb
);


--
-- Name: agent_slas; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_slas (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    project_id bigint NOT NULL,
    normal_task_max bigint DEFAULT 1800,
    complex_task_max bigint DEFAULT 7200,
    auto_escalation boolean DEFAULT true,
    enabled boolean DEFAULT true
);


--
-- Name: agent_slas_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_slas_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_slas_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_slas_id_seq OWNED BY public.agent_slas.id;


--
-- Name: agent_task_snapshots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_task_snapshots (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    agent_id bigint NOT NULL,
    task_id bigint NOT NULL,
    task_title character varying(255),
    task_status character varying(20),
    progress bigint DEFAULT 0,
    current_step text,
    step_progress bigint,
    estimated_time bigint,
    started_at timestamp with time zone
);


--
-- Name: agent_task_snapshots_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_task_snapshots_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_task_snapshots_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_task_snapshots_id_seq OWNED BY public.agent_task_snapshots.id;


--
-- Name: agent_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    title character varying(255) NOT NULL,
    description text,
    status character varying(20) DEFAULT 'enqueue'::character varying,
    priority character varying(20) DEFAULT 'normal'::character varying,
    progress bigint DEFAULT 0,
    task_type character varying(50),
    input_data jsonb,
    output_data jsonb,
    error_info text,
    logs jsonb DEFAULT '[]'::jsonb,
    failure_reason character varying(30),
    attribution jsonb,
    parent_task_id bigint,
    retry_of_task_id bigint,
    rerun_of_task_id bigint,
    agent_template_id bigint,
    agent_config_id bigint,
    runtime_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint,
    issue_id bigint,
    enqueued_at timestamp with time zone,
    claimed_at timestamp with time zone,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    cancelled_at timestamp with time zone,
    estimated_time bigint,
    actual_time bigint
);


--
-- Name: agent_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_tasks_id_seq OWNED BY public.agent_tasks.id;


--
-- Name: agent_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(128) NOT NULL,
    description text,
    is_preset boolean DEFAULT false,
    icon character varying(10) DEFAULT '🤖'::character varying,
    system_prompt text NOT NULL,
    available_skills jsonb DEFAULT '[]'::jsonb,
    available_tools jsonb DEFAULT '[]'::jsonb,
    default_config jsonb DEFAULT '{}'::jsonb,
    version character varying(50) DEFAULT '1.0'::character varying,
    status character varying(20) DEFAULT 'active'::character varying,
    workspace_id bigint,
    default_skill_id bigint,
    skill_mode character varying(20) DEFAULT 'auto'::character varying
);


--
-- Name: agent_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_templates_id_seq OWNED BY public.agent_templates.id;


--
-- Name: agent_workflows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agent_workflows (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description text,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    version bigint DEFAULT 1,
    is_active boolean DEFAULT true,
    trigger_type character varying(20) DEFAULT 'manual'::character varying,
    trigger_config jsonb,
    config jsonb
);


--
-- Name: agent_workflows_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agent_workflows_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agent_workflows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agent_workflows_id_seq OWNED BY public.agent_workflows.id;


--
-- Name: agents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.agents (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    name character varying(128) NOT NULL,
    avatar character varying(10) DEFAULT '🤖'::character varying,
    agent_type character varying(20) DEFAULT 'builtin'::character varying,
    capabilities jsonb DEFAULT '[]'::jsonb,
    status character varying(20) DEFAULT 'active'::character varying,
    model_override character varying(50),
    system_prompt text,
    permission_mode character varying(20) DEFAULT 'private'::character varying,
    invocation_targets jsonb DEFAULT '[]'::jsonb,
    visibility character varying(20) DEFAULT 'private'::character varying,
    availability character varying(20) DEFAULT 'offline'::character varying,
    workload character varying(20) DEFAULT 'idle'::character varying,
    last_active_at timestamp with time zone,
    running_task_id bigint,
    queued_task_count bigint DEFAULT 0,
    template_id bigint
);


--
-- Name: agents_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.agents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: agents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.agents_id_seq OWNED BY public.agents.id;


--
-- Name: ai_configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.ai_configs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    provider character varying(20) DEFAULT 'deepseek'::character varying,
    model character varying(50) DEFAULT 'deepseek-chat'::character varying,
    api_key character varying(500) NOT NULL,
    max_tokens bigint DEFAULT 4096,
    is_active boolean DEFAULT true,
    workspace_id bigint NOT NULL
);


--
-- Name: ai_configs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.ai_configs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_configs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_configs_id_seq OWNED BY public.ai_configs.id;


--
-- Name: ai_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.ai_messages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    thread_id bigint NOT NULL,
    role character varying(20) NOT NULL,
    content text NOT NULL,
    tool_calls jsonb,
    tool_name character varying(50)
);


--
-- Name: ai_messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.ai_messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_messages_id_seq OWNED BY public.ai_messages.id;


--
-- Name: ai_threads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.ai_threads (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title character varying(255),
    workspace_id bigint NOT NULL,
    project_id bigint,
    user_id bigint NOT NULL
);


--
-- Name: ai_threads_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.ai_threads_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_threads_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_threads_id_seq OWNED BY public.ai_threads.id;


--
-- Name: approval_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.approval_records (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    approval_id bigint NOT NULL,
    approver_id bigint NOT NULL,
    decision character varying(20) NOT NULL,
    note text,
    decided_at timestamp with time zone NOT NULL
);


--
-- Name: approval_records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.approval_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: approval_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.approval_records_id_seq OWNED BY public.approval_records.id;


--
-- Name: approvals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.approvals (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    issue_id bigint NOT NULL,
    workflow_id bigint NOT NULL,
    transition_id bigint NOT NULL,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    requester_id bigint NOT NULL,
    request_note text,
    source_state_id bigint NOT NULL,
    approve_target_state_id bigint NOT NULL,
    reject_target_state_id bigint NOT NULL,
    approver_ids jsonb DEFAULT '[]'::jsonb NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    decided_by bigint,
    decided_at timestamp with time zone,
    decision_note text
);


--
-- Name: approvals_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.approvals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: approvals_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.approvals_id_seq OWNED BY public.approvals.id;


--
-- Name: attachments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.attachments (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    file_path character varying(500) NOT NULL,
    file_size bigint,
    mime_type character varying(100),
    issue_id bigint NOT NULL,
    uploader_id bigint
);


--
-- Name: attachments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.attachments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: attachments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.attachments_id_seq OWNED BY public.attachments.id;


--
-- Name: automation_executions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.automation_executions (
    id bigint NOT NULL,
    rule_id bigint NOT NULL,
    issue_id bigint NOT NULL,
    trigger_type character varying(50) NOT NULL,
    context_json jsonb,
    actions_taken jsonb,
    status character varying(20) NOT NULL,
    error text,
    duration bigint DEFAULT 0,
    executed_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: automation_executions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.automation_executions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: automation_executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.automation_executions_id_seq OWNED BY public.automation_executions.id;


--
-- Name: automation_rule_overrides; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.automation_rule_overrides (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    rule_id bigint NOT NULL,
    project_id bigint NOT NULL,
    is_enabled boolean
);


--
-- Name: automation_rule_overrides_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.automation_rule_overrides_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: automation_rule_overrides_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.automation_rule_overrides_id_seq OWNED BY public.automation_rule_overrides.id;


--
-- Name: automation_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.automation_rules (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description text,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    is_enabled boolean DEFAULT true,
    sequence bigint DEFAULT 1,
    execution_count bigint DEFAULT 0,
    trigger_type character varying(50) NOT NULL,
    conditions text,
    actions text,
    scope character varying(20) DEFAULT 'all'::character varying,
    schedule_config text,
    last_triggered_at timestamp with time zone
);


--
-- Name: automation_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.automation_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: automation_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.automation_rules_id_seq OWNED BY public.automation_rules.id;


--
-- Name: autopilot_executions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.autopilot_executions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    task_id bigint NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    trigger_type character varying(20),
    input_data jsonb DEFAULT '{}'::jsonb,
    output_data jsonb,
    error_info text,
    logs jsonb DEFAULT '[]'::jsonb,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    failed_at timestamp with time zone,
    duration_ms bigint,
    retry_count bigint DEFAULT 0
);


--
-- Name: autopilot_executions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.autopilot_executions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: autopilot_executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.autopilot_executions_id_seq OWNED BY public.autopilot_executions.id;


--
-- Name: autopilot_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.autopilot_tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint,
    name character varying(100) NOT NULL,
    description text,
    trigger_type character varying(20) NOT NULL,
    cron_expression character varying(100),
    trigger_url character varying(255),
    task_type character varying(50) NOT NULL,
    agent_template_id bigint,
    agent_config_id bigint,
    input_data jsonb DEFAULT '{}'::jsonb,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    last_run_at timestamp with time zone,
    next_run_at timestamp with time zone,
    config jsonb DEFAULT '{}'::jsonb,
    notification_config jsonb DEFAULT '{}'::jsonb,
    timeout_seconds bigint DEFAULT 300,
    retry_count bigint DEFAULT 0,
    enabled boolean DEFAULT true
);


--
-- Name: autopilot_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.autopilot_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: autopilot_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.autopilot_tasks_id_seq OWNED BY public.autopilot_tasks.id;


--
-- Name: autopilot_webhooks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.autopilot_webhooks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    task_id bigint NOT NULL,
    endpoint character varying(255) NOT NULL,
    secret character varying(100),
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    allowed_ips jsonb DEFAULT '[]'::jsonb,
    max_requests bigint DEFAULT 100,
    request_count bigint DEFAULT 0,
    last_used_at timestamp with time zone
);


--
-- Name: autopilot_webhooks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.autopilot_webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: autopilot_webhooks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.autopilot_webhooks_id_seq OWNED BY public.autopilot_webhooks.id;


--
-- Name: build_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.build_records (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint,
    cicd_config_id bigint NOT NULL,
    trigger character varying(20) DEFAULT 'manual'::character varying NOT NULL,
    branch character varying(120),
    commit_sha character varying(64),
    issue_id bigint,
    agent_task_id bigint,
    triggered_by_id bigint,
    external_build_id character varying(128),
    build_url character varying(512),
    stages jsonb DEFAULT '[]'::jsonb,
    status character varying(20) DEFAULT 'pending'::character varying,
    progress bigint DEFAULT 0,
    current_stage character varying(100),
    error_message text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    cancelled_at timestamp with time zone,
    duration_ms bigint DEFAULT 0
);


--
-- Name: build_records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.build_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: build_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.build_records_id_seq OWNED BY public.build_records.id;


--
-- Name: chats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.chats (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint,
    issue_id bigint,
    type character varying(20) DEFAULT 'issue'::character varying,
    title character varying(255)
);


--
-- Name: chats_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.chats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: chats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chats_id_seq OWNED BY public.chats.id;


--
-- Name: cicd_configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.cicd_configs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint,
    name character varying(120) NOT NULL,
    provider character varying(30) DEFAULT 'generic'::character varying NOT NULL,
    api_endpoint character varying(512),
    project_slug character varying(255),
    default_branch character varying(120) DEFAULT 'main'::character varying,
    auth_token_ref character varying(255),
    trigger_events jsonb DEFAULT '[]'::jsonb,
    extra_config jsonb DEFAULT '{}'::jsonb,
    enabled boolean DEFAULT true
);


--
-- Name: cicd_configs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.cicd_configs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cicd_configs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cicd_configs_id_seq OWNED BY public.cicd_configs.id;


--
-- Name: comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.comments (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    issue_id bigint NOT NULL,
    author_id bigint,
    agent_id bigint,
    body text NOT NULL,
    is_resolved boolean DEFAULT false,
    parent_id bigint,
    suggestions jsonb,
    suggestion_outcomes jsonb,
    suggestions_applied_at timestamp with time zone
);


--
-- Name: comments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.comments_id_seq OWNED BY public.comments.id;


--
-- Name: conditional_fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.conditional_fields (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    field_id bigint NOT NULL,
    condition_type character varying(50) NOT NULL,
    operator character varying(50) NOT NULL,
    condition_values text,
    is_enabled boolean DEFAULT true,
    priority bigint DEFAULT 0,
    description character varying(255)
);


--
-- Name: conditional_fields_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.conditional_fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: conditional_fields_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.conditional_fields_id_seq OWNED BY public.conditional_fields.id;


--
-- Name: custom_field_options; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.custom_field_options (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    field_id bigint NOT NULL,
    value character varying(255) NOT NULL,
    color character varying(20),
    sequence bigint DEFAULT 1
);


--
-- Name: custom_field_options_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.custom_field_options_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: custom_field_options_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.custom_field_options_id_seq OWNED BY public.custom_field_options.id;


--
-- Name: custom_fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.custom_fields (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description text,
    field_type character varying(20) NOT NULL,
    is_required boolean DEFAULT false,
    default_value text,
    placeholder character varying(255),
    is_active boolean DEFAULT true,
    is_readonly boolean DEFAULT false,
    is_multi_select boolean DEFAULT false,
    number_min numeric,
    number_max numeric,
    project_id bigint,
    workspace_id bigint NOT NULL
);


--
-- Name: custom_fields_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.custom_fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: custom_fields_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.custom_fields_id_seq OWNED BY public.custom_fields.id;


--
-- Name: cycles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.cycles (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    description character varying(1000),
    start_date date NOT NULL,
    end_date date,
    completed_at timestamp with time zone,
    cancelled_at timestamp with time zone,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    auto_add_enabled boolean DEFAULT false,
    auto_add_rql character varying(1000),
    auto_close_enabled boolean DEFAULT false,
    auto_progress_enabled boolean DEFAULT false
);


--
-- Name: cycles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.cycles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cycles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cycles_id_seq OWNED BY public.cycles.id;


--
-- Name: dashboard_widgets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.dashboard_widgets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    dashboard_id bigint NOT NULL,
    widget_type character varying(30) NOT NULL,
    title character varying(100),
    description character varying(255),
    config jsonb,
    "position" jsonb,
    sort_order bigint DEFAULT 0
);


--
-- Name: dashboard_widgets_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.dashboard_widgets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: dashboard_widgets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.dashboard_widgets_id_seq OWNED BY public.dashboard_widgets.id;


--
-- Name: data_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.data_migrations (
    name character varying(128) NOT NULL,
    applied_at timestamp with time zone NOT NULL
);


--
-- Name: developer_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.developer_jobs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint,
    issue_id bigint,
    agent_task_id bigint,
    git_provider character varying(20) DEFAULT 'github'::character varying,
    git_connection_id bigint,
    title character varying(255) NOT NULL,
    requirement_text text,
    design_doc_url character varying(500),
    input_context jsonb DEFAULT '{}'::jsonb,
    branch_name character varying(255),
    base_branch character varying(255) DEFAULT 'main'::character varying,
    commit_message text,
    generated_files jsonb DEFAULT '[]'::jsonb,
    pr_number bigint,
    pr_url character varying(500),
    pr_title character varying(255),
    commit_sha character varying(64),
    status character varying(20) DEFAULT 'pending'::character varying,
    progress bigint DEFAULT 0,
    current_step character varying(100),
    error_message text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    cancelled_at timestamp with time zone
);


--
-- Name: developer_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.developer_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: developer_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.developer_jobs_id_seq OWNED BY public.developer_jobs.id;


--
-- Name: estimate_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.estimate_categories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    mode character varying(20) DEFAULT 'categories'::character varying,
    is_default boolean DEFAULT false,
    sequence bigint DEFAULT 1,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: estimate_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.estimate_categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: estimate_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.estimate_categories_id_seq OWNED BY public.estimate_categories.id;


--
-- Name: estimate_points; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.estimate_points (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    value bigint,
    mode character varying(20) DEFAULT 'points'::character varying,
    is_default boolean DEFAULT false,
    sequence bigint DEFAULT 1,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: estimate_points_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.estimate_points_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: estimate_points_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.estimate_points_id_seq OWNED BY public.estimate_points.id;


--
-- Name: estimate_times; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.estimate_times (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    minutes bigint,
    mode character varying(20) DEFAULT 'time'::character varying,
    is_default boolean DEFAULT false,
    sequence bigint DEFAULT 1,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: estimate_times_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.estimate_times_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: estimate_times_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.estimate_times_id_seq OWNED BY public.estimate_times.id;


--
-- Name: field_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.field_permissions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    resource character varying(100) NOT NULL,
    field_name character varying(100) NOT NULL,
    role_id bigint,
    can_read boolean DEFAULT true,
    can_write boolean DEFAULT false,
    project_id bigint,
    workspace_id bigint
);


--
-- Name: field_permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.field_permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: field_permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.field_permissions_id_seq OWNED BY public.field_permissions.id;


--
-- Name: git_integrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.git_integrations (
    id bigint NOT NULL,
    project_id bigint NOT NULL,
    provider character varying(20) NOT NULL,
    repo_url text,
    repo_name character varying(255),
    access_token text,
    webhook_secret character varying(255),
    active boolean DEFAULT true,
    sync_p_rs boolean DEFAULT true,
    sync_commits boolean DEFAULT true,
    sync_branches boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: git_integrations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.git_integrations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: git_integrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.git_integrations_id_seq OWNED BY public.git_integrations.id;


--
-- Name: git_issue_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.git_issue_links (
    id bigint NOT NULL,
    issue_id bigint NOT NULL,
    git_type character varying(20) NOT NULL,
    git_id character varying(255) NOT NULL,
    git_url text,
    git_title text,
    git_state character varying(20),
    git_author character varying(255),
    git_branch character varying(255),
    integration_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: git_issue_links_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.git_issue_links_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: git_issue_links_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.git_issue_links_id_seq OWNED BY public.git_issue_links.id;


--
-- Name: github_connections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.github_connections (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint NOT NULL,
    repo_owner character varying(255) NOT NULL,
    repo_name character varying(255) NOT NULL,
    access_token character varying(500),
    webhook_secret character varying(500),
    is_enabled boolean DEFAULT true,
    sync_issues boolean DEFAULT true,
    sync_p_rs boolean DEFAULT true,
    last_sync_at character varying(30),
    webhook_id bigint
);


--
-- Name: github_connections_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.github_connections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: github_connections_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.github_connections_id_seq OWNED BY public.github_connections.id;


--
-- Name: initiative_projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.initiative_projects (
    initiative_id bigint NOT NULL,
    project_id bigint NOT NULL
);


--
-- Name: initiatives; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.initiatives (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint,
    name character varying(255) NOT NULL,
    description text,
    color character varying(20),
    status character varying(20) DEFAULT 'active'::character varying,
    target_date timestamp with time zone,
    start_date timestamp with time zone,
    sort_order bigint DEFAULT 0
);


--
-- Name: initiatives_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.initiatives_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: initiatives_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.initiatives_id_seq OWNED BY public.initiatives.id;


--
-- Name: issue_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_activities (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    issue_id bigint,
    verb character varying(255) DEFAULT 'created'::character varying,
    field character varying(255),
    old_value text,
    new_value text,
    comment text,
    actor_id bigint
);


--
-- Name: issue_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.issue_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: issue_activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.issue_activities_id_seq OWNED BY public.issue_activities.id;


--
-- Name: issue_assignees; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_assignees (
    issue_id bigint NOT NULL,
    user_id bigint NOT NULL
);


--
-- Name: issue_custom_field_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_custom_field_values (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    issue_id bigint NOT NULL,
    field_id bigint NOT NULL,
    value text
);


--
-- Name: issue_custom_field_values_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.issue_custom_field_values_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: issue_custom_field_values_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.issue_custom_field_values_id_seq OWNED BY public.issue_custom_field_values.id;


--
-- Name: issue_cycles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_cycles (
    issue_id bigint NOT NULL,
    cycle_id bigint NOT NULL
);


--
-- Name: issue_labels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_labels (
    issue_id bigint NOT NULL,
    label_id bigint NOT NULL
);


--
-- Name: issue_pages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_pages (
    issue_id bigint NOT NULL,
    page_id bigint NOT NULL
);


--
-- Name: issue_relations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_relations (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    issue_id bigint NOT NULL,
    related_issue_id bigint NOT NULL,
    relation_type_id bigint NOT NULL,
    comment text
);


--
-- Name: issue_relations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.issue_relations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: issue_relations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.issue_relations_id_seq OWNED BY public.issue_relations.id;


--
-- Name: issue_type_fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_type_fields (
    type_id bigint NOT NULL,
    field_id bigint NOT NULL,
    is_required boolean DEFAULT false,
    sequence bigint DEFAULT 1
);


--
-- Name: issue_type_imports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_type_imports (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    project_id bigint NOT NULL,
    workspace_type_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: issue_type_imports_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.issue_type_imports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: issue_type_imports_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.issue_type_imports_id_seq OWNED BY public.issue_type_imports.id;


--
-- Name: issue_type_template_fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_type_template_fields (
    template_type_id bigint NOT NULL,
    field_id bigint NOT NULL,
    is_required boolean DEFAULT false,
    sequence bigint DEFAULT 1
);


--
-- Name: issue_type_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_type_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    color character varying(20) DEFAULT '#6366F1'::character varying,
    icon character varying(50) DEFAULT 'circle'::character varying,
    description text,
    level bigint DEFAULT 0,
    parent_type_id bigint,
    workspace_id bigint NOT NULL
);


--
-- Name: issue_type_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.issue_type_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: issue_type_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.issue_type_templates_id_seq OWNED BY public.issue_type_templates.id;


--
-- Name: issue_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_types (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    color character varying(20) DEFAULT '#6366F1'::character varying,
    icon character varying(50) DEFAULT 'circle'::character varying,
    description text,
    level bigint DEFAULT 0,
    parent_type_id bigint,
    allowed_child_type_ids jsonb,
    is_default boolean DEFAULT false,
    sequence bigint DEFAULT 1,
    is_active boolean DEFAULT true,
    project_id bigint,
    workspace_id bigint NOT NULL
);


--
-- Name: issue_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.issue_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: issue_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.issue_types_id_seq OWNED BY public.issue_types.id;


--
-- Name: issue_watchers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issue_watchers (
    id bigint NOT NULL,
    issue_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: issue_watchers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.issue_watchers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: issue_watchers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.issue_watchers_id_seq OWNED BY public.issue_watchers.id;


--
-- Name: issues; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.issues (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    description_html text DEFAULT '<p></p>'::text,
    description_json jsonb,
    description_stripped text,
    priority character varying(30) DEFAULT 'none'::character varying,
    sequence_id bigint DEFAULT 1,
    sort_order numeric DEFAULT 65535,
    start_date timestamp with time zone,
    target_date timestamp with time zone,
    completed_at timestamp with time zone,
    is_draft boolean DEFAULT false,
    archived_at timestamp with time zone,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    parent_id bigint,
    depth bigint DEFAULT 0,
    issue_type_id bigint,
    state_id bigint NOT NULL,
    approval_status character varying(20),
    active_approval_id bigint,
    external_id character varying(255),
    external_source character varying(255),
    cover_image_url character varying(500),
    intake_source character varying(50),
    intake_status character varying(30),
    intake_submitter character varying(255),
    intake_email character varying(255),
    intake_triaged_at timestamp with time zone,
    intake_triaged_by bigint,
    intake_snoozed_until timestamp with time zone,
    intake_duplicate_of bigint,
    intake_note text,
    agent_assignee_id bigint,
    agent_task_id bigint
);


--
-- Name: issues_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.issues_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: issues_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.issues_id_seq OWNED BY public.issues.id;


--
-- Name: labels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.labels (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    color character varying(50) DEFAULT '#6B7280'::character varying,
    description character varying(255),
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: labels_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.labels_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: labels_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.labels_id_seq OWNED BY public.labels.id;


--
-- Name: mcp_configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.mcp_configs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    description character varying(500),
    server_url character varying(500) NOT NULL,
    transport_type character varying(20) DEFAULT 'sse'::character varying,
    api_key character varying(500),
    tools_config text,
    is_enabled boolean DEFAULT true,
    last_sync_at timestamp with time zone
);


--
-- Name: mcp_configs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.mcp_configs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mcp_configs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.mcp_configs_id_seq OWNED BY public.mcp_configs.id;


--
-- Name: memory_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.memory_entries (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint,
    issue_id bigint,
    agent_id bigint,
    memory_type character varying(20) NOT NULL,
    scope character varying(20) NOT NULL,
    content text NOT NULL,
    embedding jsonb,
    metadata jsonb DEFAULT '{}'::jsonb,
    tags jsonb DEFAULT '[]'::jsonb,
    context_key character varying(255),
    context_name character varying(255),
    relevance_score numeric(5,4) DEFAULT 0,
    expires_at timestamp with time zone
);


--
-- Name: memory_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.memory_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: memory_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.memory_entries_id_seq OWNED BY public.memory_entries.id;


--
-- Name: memory_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.memory_sessions (
    id character varying(64) NOT NULL,
    workspace_id bigint NOT NULL,
    session_type character varying(50) NOT NULL,
    context_id character varying(64),
    started_at timestamp with time zone,
    updated_at timestamp with time zone,
    closed_at timestamp with time zone,
    memory_count bigint DEFAULT 0
);


--
-- Name: memory_vector_indexes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.memory_vector_indexes (
    id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    memory_entry_id bigint NOT NULL,
    vector_data jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: memory_vector_indexes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.memory_vector_indexes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: memory_vector_indexes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.memory_vector_indexes_id_seq OWNED BY public.memory_vector_indexes.id;


--
-- Name: message_reactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.message_reactions (
    id bigint NOT NULL,
    message_id bigint NOT NULL,
    user_id bigint NOT NULL,
    emoji character varying(50) NOT NULL,
    created_at timestamp with time zone
);


--
-- Name: message_reactions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.message_reactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: message_reactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.message_reactions_id_seq OWNED BY public.message_reactions.id;


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.messages (
    id bigint NOT NULL,
    chat_id bigint NOT NULL,
    sender_id bigint NOT NULL,
    sender_type character varying(20) NOT NULL,
    content text NOT NULL,
    reply_to_id bigint,
    mentions jsonb,
    edited_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone
);


--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.messages_id_seq OWNED BY public.messages.id;


--
-- Name: metric_charts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.metric_charts (
    id bigint NOT NULL,
    project_id bigint NOT NULL,
    creator_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    template_id character varying(100),
    chart_type character varying(50) NOT NULL,
    x_axis character varying(100) NOT NULL,
    y_axis character varying(100) NOT NULL,
    filters jsonb DEFAULT '{}'::jsonb,
    config jsonb DEFAULT '{}'::jsonb,
    sort_order bigint DEFAULT 0,
    is_visible boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: metric_charts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.metric_charts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: metric_charts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.metric_charts_id_seq OWNED BY public.metric_charts.id;


--
-- Name: module_inheritance_overrides; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.module_inheritance_overrides (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    project_id bigint NOT NULL,
    workspace_module_id bigint NOT NULL,
    is_excluded boolean DEFAULT false,
    override_name character varying(100),
    override_description text
);


--
-- Name: module_inheritance_overrides_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.module_inheritance_overrides_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: module_inheritance_overrides_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.module_inheritance_overrides_id_seq OWNED BY public.module_inheritance_overrides.id;


--
-- Name: module_issues; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.module_issues (
    module_id bigint NOT NULL,
    issue_id bigint NOT NULL
);


--
-- Name: modules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.modules (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description text,
    project_id bigint,
    workspace_id bigint NOT NULL,
    parent_id bigint,
    "order" bigint DEFAULT 0,
    archived_at timestamp with time zone,
    is_archived boolean DEFAULT false
);


--
-- Name: modules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.modules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: modules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.modules_id_seq OWNED BY public.modules.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.notifications (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    title character varying(255) NOT NULL,
    message text NOT NULL,
    type character varying(20) DEFAULT 'info'::character varying,
    priority character varying(20) DEFAULT 'medium'::character varying,
    is_read boolean DEFAULT false,
    read_at timestamp with time zone,
    action_url character varying(500),
    recipient_id bigint NOT NULL,
    sender_id bigint,
    project_id bigint,
    issue_id bigint
);


--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: page_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.page_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    description text DEFAULT ''::text,
    content text DEFAULT ''::text,
    content_json jsonb,
    is_default boolean DEFAULT false,
    workspace_id bigint NOT NULL,
    project_id bigint
);


--
-- Name: page_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.page_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: page_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.page_templates_id_seq OWNED BY public.page_templates.id;


--
-- Name: page_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.page_versions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    page_id bigint NOT NULL,
    title character varying(255) NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    content_json jsonb,
    version_number bigint DEFAULT 1 NOT NULL,
    change_summary character varying(500)
);


--
-- Name: page_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.page_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: page_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.page_versions_id_seq OWNED BY public.page_versions.id;


--
-- Name: pages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.pages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    title character varying(255) NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    content_json jsonb,
    published boolean DEFAULT true,
    archived_at timestamp with time zone,
    sequence bigint DEFAULT 1,
    parent_id bigint,
    depth bigint DEFAULT 0,
    locked_by_id bigint,
    locked_at timestamp with time zone,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    source_type character varying(32),
    source_id bigint
);


--
-- Name: pages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.pages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pages_id_seq OWNED BY public.pages.id;


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.permissions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    code character varying(100) NOT NULL,
    name character varying(200) NOT NULL,
    description character varying(500),
    resource character varying(100) NOT NULL,
    action character varying(50) NOT NULL,
    scope character varying(20) DEFAULT 'project'::character varying,
    field_name character varying(100) DEFAULT ''::character varying
);


--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permissions_id_seq OWNED BY public.permissions.id;


--
-- Name: personal_access_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.personal_access_tokens (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    user_id bigint NOT NULL,
    name character varying(100) NOT NULL,
    token_prefix character varying(20) NOT NULL,
    token_hash character varying(64) NOT NULL,
    scopes text DEFAULT ''::text,
    last_used_at timestamp with time zone,
    expires_at timestamp with time zone,
    revoked_at timestamp with time zone
);


--
-- Name: personal_access_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.personal_access_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: personal_access_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.personal_access_tokens_id_seq OWNED BY public.personal_access_tokens.id;


--
-- Name: plugin_event_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.plugin_event_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    plugin_id bigint NOT NULL,
    event_type character varying(50) NOT NULL,
    status character varying(20) DEFAULT 'success'::character varying,
    request_body text,
    response_body text,
    status_code bigint,
    duration_ms bigint
);


--
-- Name: plugin_event_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.plugin_event_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: plugin_event_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.plugin_event_logs_id_seq OWNED BY public.plugin_event_logs.id;


--
-- Name: plugins; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.plugins (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    slug character varying(100) NOT NULL,
    description character varying(500),
    author character varying(100),
    version character varying(20) DEFAULT '1.0.0'::character varying,
    icon_url character varying(500),
    type character varying(30) NOT NULL,
    entry_point character varying(500),
    config_schema jsonb,
    config jsonb,
    subscribed_events jsonb,
    enabled boolean DEFAULT false,
    workspace_id bigint NOT NULL,
    installed_by_id bigint NOT NULL
);


--
-- Name: plugins_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.plugins_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: plugins_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.plugins_id_seq OWNED BY public.plugins.id;


--
-- Name: project_agent_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.project_agent_members (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    project_id bigint NOT NULL,
    agent_id bigint NOT NULL,
    role character varying(20) DEFAULT 'member'::character varying,
    is_active boolean DEFAULT true
);


--
-- Name: project_agent_members_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.project_agent_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_agent_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_agent_members_id_seq OWNED BY public.project_agent_members.id;


--
-- Name: project_custom_field_enrollments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.project_custom_field_enrollments (
    project_id bigint NOT NULL,
    field_id bigint NOT NULL,
    is_enabled boolean DEFAULT true,
    sequence bigint DEFAULT 0
);


--
-- Name: project_estimate_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.project_estimate_settings (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    mode character varying(20) DEFAULT 'points'::character varying,
    points_enabled boolean DEFAULT true,
    categories_enabled boolean DEFAULT false,
    time_enabled boolean DEFAULT false
);


--
-- Name: project_estimate_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.project_estimate_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_estimate_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_estimate_settings_id_seq OWNED BY public.project_estimate_settings.id;


--
-- Name: project_intake_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.project_intake_settings (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    project_id bigint NOT NULL,
    form_enabled boolean DEFAULT true,
    webhook_enabled boolean DEFAULT true,
    email_enabled boolean DEFAULT true,
    token character varying(64),
    sla_hours bigint DEFAULT 48
);


--
-- Name: project_intake_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.project_intake_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_intake_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_intake_settings_id_seq OWNED BY public.project_intake_settings.id;


--
-- Name: project_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.project_members (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    project_id bigint NOT NULL,
    user_id bigint NOT NULL,
    role bigint DEFAULT 15,
    is_active boolean DEFAULT true
);


--
-- Name: project_members_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.project_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_members_id_seq OWNED BY public.project_members.id;


--
-- Name: project_page_tabs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.project_page_tabs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    project_id bigint NOT NULL,
    owner_id bigint NOT NULL,
    name character varying(50) NOT NULL,
    icon character varying(30),
    tab_type character varying(30) DEFAULT 'custom'::character varying NOT NULL,
    route_key character varying(50),
    target_type character varying(20),
    target_id bigint,
    target_url character varying(500),
    visible boolean DEFAULT true,
    sort_order bigint DEFAULT 0
);


--
-- Name: project_page_tabs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.project_page_tabs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_page_tabs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_page_tabs_id_seq OWNED BY public.project_page_tabs.id;


--
-- Name: project_subscribers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.project_subscribers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    project_id bigint NOT NULL,
    user_id bigint NOT NULL
);


--
-- Name: project_subscribers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.project_subscribers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_subscribers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_subscribers_id_seq OWNED BY public.project_subscribers.id;


--
-- Name: project_template_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.project_template_types (
    template_id bigint NOT NULL,
    type_template_id bigint NOT NULL,
    is_required boolean DEFAULT false,
    default_state_id bigint,
    sequence bigint DEFAULT 1
);


--
-- Name: project_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.project_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description text,
    workspace_id bigint NOT NULL,
    is_default boolean DEFAULT false,
    states text,
    labels text
);


--
-- Name: project_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.project_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_templates_id_seq OWNED BY public.project_templates.id;


--
-- Name: project_updates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.project_updates (
    id bigint NOT NULL,
    project_id bigint,
    author_id bigint,
    status character varying(20) NOT NULL,
    content text,
    recap text,
    plan text,
    blockers text,
    metrics text,
    created_at timestamp with time zone
);


--
-- Name: project_updates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.project_updates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_updates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_updates_id_seq OWNED BY public.project_updates.id;


--
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.projects (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    identifier character varying(10) NOT NULL,
    description character varying(1000),
    is_public boolean DEFAULT false,
    timezone character varying(255) DEFAULT 'UTC'::character varying,
    archived_at timestamp with time zone,
    workspace_id bigint NOT NULL,
    default_assignee_id bigint,
    project_lead_id bigint,
    color character varying(20) DEFAULT '#6366F1'::character varying,
    template_id bigint
);


--
-- Name: projects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.projects_id_seq OWNED BY public.projects.id;


--
-- Name: recurrence_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.recurrence_rules (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    issue_id bigint NOT NULL,
    frequency character varying(20) NOT NULL,
    "interval" bigint DEFAULT 1,
    cron_expr character varying(100),
    next_run timestamp with time zone NOT NULL,
    end_date timestamp with time zone,
    is_active boolean DEFAULT true
);


--
-- Name: recurrence_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.recurrence_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: recurrence_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.recurrence_rules_id_seq OWNED BY public.recurrence_rules.id;


--
-- Name: relation_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.relation_types (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    inward_name character varying(100) NOT NULL,
    outward_name character varying(100) NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: relation_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.relation_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: relation_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.relation_types_id_seq OWNED BY public.relation_types.id;


--
-- Name: release_issues; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.release_issues (
    release_id bigint NOT NULL,
    issue_id bigint NOT NULL
);


--
-- Name: releases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.releases (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    version character varying(50) NOT NULL,
    description text,
    status character varying(30) DEFAULT 'planned'::character varying,
    release_date timestamp with time zone,
    project_id bigint NOT NULL
);


--
-- Name: releases_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.releases_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: releases_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.releases_id_seq OWNED BY public.releases.id;


--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.role_permissions (
    role_id bigint NOT NULL,
    permission_id bigint NOT NULL
);


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.roles (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description character varying(500),
    scope character varying(20) DEFAULT 'project'::character varying,
    workspace_id bigint,
    project_id bigint,
    is_system boolean DEFAULT false,
    sort_order bigint DEFAULT 0,
    level bigint DEFAULT 15
);


--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- Name: runtimes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.runtimes (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(128) NOT NULL,
    runtime_type character varying(20) NOT NULL,
    runtime_mode character varying(20) DEFAULT 'pull'::character varying,
    endpoint character varying(255),
    status character varying(20) DEFAULT 'offline'::character varying,
    health character varying(20) DEFAULT 'offline'::character varying,
    capacity bigint DEFAULT 1,
    current_load bigint DEFAULT 0,
    version character varying(50),
    host_info jsonb,
    last_heartbeat timestamp with time zone,
    metadata jsonb DEFAULT '{}'::jsonb,
    workspace_id bigint NOT NULL
);


--
-- Name: runtimes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.runtimes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: runtimes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.runtimes_id_seq OWNED BY public.runtimes.id;


--
-- Name: saved_dashboards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.saved_dashboards (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description character varying(500),
    is_default boolean DEFAULT false,
    is_shared boolean DEFAULT false,
    owner_id bigint NOT NULL,
    project_id bigint NOT NULL,
    date_from character varying(20),
    date_to character varying(20),
    columns bigint DEFAULT 12
);


--
-- Name: saved_dashboards_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.saved_dashboards_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saved_dashboards_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saved_dashboards_id_seq OWNED BY public.saved_dashboards.id;


--
-- Name: saved_reports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.saved_reports (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    report_type character varying(30) NOT NULL,
    group_by character varying(50),
    chart_type character varying(20) DEFAULT 'bar'::character varying,
    rql text,
    "interval" character varying(10),
    date_from character varying(20),
    date_to character varying(20),
    project_id bigint NOT NULL
);


--
-- Name: saved_reports_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.saved_reports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saved_reports_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saved_reports_id_seq OWNED BY public.saved_reports.id;


--
-- Name: saved_views; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.saved_views (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description character varying(255),
    view_type character varying(20) DEFAULT 'list'::character varying,
    filters jsonb,
    rql text,
    sort_config jsonb,
    columns jsonb,
    group_by character varying(50),
    sub_group_by character varying(50),
    is_default boolean DEFAULT false,
    is_shared boolean DEFAULT false,
    owner_id bigint NOT NULL,
    project_id bigint NOT NULL
);


--
-- Name: saved_views_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.saved_views_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saved_views_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saved_views_id_seq OWNED BY public.saved_views.id;


--
-- Name: sdlc_stages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.sdlc_stages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workflow_id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    "order" bigint NOT NULL,
    key character varying(60) NOT NULL,
    name character varying(100) NOT NULL,
    agent_role character varying(60),
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    progress bigint DEFAULT 0,
    input jsonb DEFAULT '{}'::jsonb,
    output jsonb DEFAULT '{}'::jsonb,
    logs jsonb DEFAULT '[]'::jsonb,
    error_message text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    duration_ms bigint DEFAULT 0
);


--
-- Name: sdlc_stages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.sdlc_stages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sdlc_stages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sdlc_stages_id_seq OWNED BY public.sdlc_stages.id;


--
-- Name: sdlc_workflows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.sdlc_workflows (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint,
    squad_id bigint,
    title character varying(200) NOT NULL,
    requirement text,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    progress bigint DEFAULT 0,
    current_stage character varying(100),
    config jsonb DEFAULT '{}'::jsonb,
    artifacts jsonb DEFAULT '{}'::jsonb,
    error_message text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    cancelled_at timestamp with time zone
);


--
-- Name: sdlc_workflows_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.sdlc_workflows_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sdlc_workflows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sdlc_workflows_id_seq OWNED BY public.sdlc_workflows.id;


--
-- Name: search_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.search_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description character varying(255),
    icon character varying(50),
    rql_template text NOT NULL,
    view_type character varying(20) DEFAULT 'list'::character varying,
    sort_config jsonb,
    group_by character varying(50),
    columns jsonb,
    is_built_in boolean DEFAULT false,
    is_public boolean DEFAULT true,
    owner_id bigint,
    project_id bigint
);


--
-- Name: search_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.search_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: search_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.search_templates_id_seq OWNED BY public.search_templates.id;


--
-- Name: skill_execution_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.skill_execution_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    skill_id bigint NOT NULL,
    workspace_id bigint NOT NULL,
    input_params jsonb DEFAULT '{}'::jsonb,
    output_result jsonb DEFAULT '{}'::jsonb,
    status character varying(20) DEFAULT 'running'::character varying,
    error_message text,
    tokens_used bigint DEFAULT 0,
    duration_ms bigint DEFAULT 0
);


--
-- Name: skill_execution_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.skill_execution_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: skill_execution_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.skill_execution_logs_id_seq OWNED BY public.skill_execution_logs.id;


--
-- Name: skills; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.skills (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(128) NOT NULL,
    description text,
    skill_type character varying(30) DEFAULT 'custom'::character varying,
    version character varying(50) DEFAULT '1.0'::character varying,
    status character varying(20) DEFAULT 'active'::character varying,
    skill_md text NOT NULL,
    parameters jsonb DEFAULT '[]'::jsonb,
    tags jsonb DEFAULT '[]'::jsonb,
    usage_count bigint DEFAULT 0,
    is_shared boolean DEFAULT false,
    is_preset boolean DEFAULT false,
    workspace_id bigint NOT NULL
);


--
-- Name: skills_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.skills_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: skills_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.skills_id_seq OWNED BY public.skills.id;


--
-- Name: slack_connections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.slack_connections (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint NOT NULL,
    channel_name character varying(255) NOT NULL,
    webhook_url character varying(500) NOT NULL,
    bot_token character varying(500),
    is_enabled boolean DEFAULT true,
    notify_on_create boolean DEFAULT true,
    notify_on_update boolean DEFAULT true,
    notify_on_comment boolean DEFAULT false,
    notify_on_complete boolean DEFAULT true
);


--
-- Name: slack_connections_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.slack_connections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: slack_connections_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.slack_connections_id_seq OWNED BY public.slack_connections.id;


--
-- Name: squad_executions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.squad_executions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    squad_id bigint NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    goal text,
    input_data text DEFAULT '{}'::text,
    output_data text,
    logs text DEFAULT '[]'::text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    failed_at timestamp with time zone,
    error_info text,
    cancelled_at timestamp with time zone,
    cancel_reason text DEFAULT ''::text
);


--
-- Name: squad_executions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.squad_executions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: squad_executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.squad_executions_id_seq OWNED BY public.squad_executions.id;


--
-- Name: squad_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.squad_members (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    squad_id bigint NOT NULL,
    agent_id bigint NOT NULL,
    role character varying(50) NOT NULL,
    agent_config_id bigint,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    assigned_at timestamp with time zone,
    removed_at timestamp with time zone
);


--
-- Name: squad_members_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.squad_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: squad_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.squad_members_id_seq OWNED BY public.squad_members.id;


--
-- Name: squad_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.squad_tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    squad_id bigint NOT NULL,
    agent_task_id bigint,
    member_id bigint NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    priority character varying(20),
    progress bigint DEFAULT 0,
    task_description text,
    feedback text,
    retry_count bigint DEFAULT 0,
    max_retries bigint DEFAULT 2,
    error_message text DEFAULT ''::text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone
);


--
-- Name: squad_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.squad_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: squad_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.squad_tasks_id_seq OWNED BY public.squad_tasks.id;


--
-- Name: squads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.squads (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint,
    name character varying(100) NOT NULL,
    description text,
    leader_agent_id bigint,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    config text DEFAULT '{}'::text,
    goal text
);


--
-- Name: squads_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.squads_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: squads_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.squads_id_seq OWNED BY public.squads.id;


--
-- Name: state_transitions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.state_transitions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    description character varying(500),
    workflow_id bigint NOT NULL,
    workflow_kind character varying(16) DEFAULT 'state'::character varying NOT NULL,
    source_state_id bigint NOT NULL,
    target_state_id bigint NOT NULL,
    issue_type_id bigint,
    is_auto boolean DEFAULT false,
    rule_type text DEFAULT 'allow'::text,
    approver_ids text,
    role_allowed character varying(50),
    approve_target_state_id bigint,
    reject_target_state_id bigint,
    approval_mode character varying(20) DEFAULT 'any'::character varying,
    project_id bigint,
    workspace_id bigint NOT NULL
);


--
-- Name: state_transitions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.state_transitions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: state_transitions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.state_transitions_id_seq OWNED BY public.state_transitions.id;


--
-- Name: states; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.states (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    color character varying(50) DEFAULT '#6B7280'::character varying,
    "group" character varying(50) DEFAULT 'backlog'::character varying,
    sequence bigint DEFAULT 1,
    is_default boolean DEFAULT false,
    is_active boolean DEFAULT true,
    project_id bigint,
    workspace_id bigint NOT NULL
);


--
-- Name: states_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.states_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: states_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.states_id_seq OWNED BY public.states.id;


--
-- Name: task_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.task_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    task_id bigint NOT NULL,
    level character varying(20) NOT NULL,
    message text NOT NULL,
    metadata jsonb
);


--
-- Name: task_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.task_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: task_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.task_logs_id_seq OWNED BY public.task_logs.id;


--
-- Name: tester_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.tester_jobs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    project_id bigint,
    issue_id bigint,
    agent_task_id bigint,
    title character varying(255) NOT NULL,
    requirement_text text,
    acceptance_criteria text,
    test_scope character varying(20) DEFAULT 'unit'::character varying,
    input_context jsonb DEFAULT '{}'::jsonb,
    generated_cases jsonb DEFAULT '[]'::jsonb,
    test_results jsonb DEFAULT '[]'::jsonb,
    total_cases bigint DEFAULT 0,
    pass_count bigint DEFAULT 0,
    fail_count bigint DEFAULT 0,
    skip_count bigint DEFAULT 0,
    bug_issue_ids jsonb DEFAULT '[]'::jsonb,
    status character varying(20) DEFAULT 'pending'::character varying,
    progress bigint DEFAULT 0,
    current_step character varying(100),
    error_message text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    cancelled_at timestamp with time zone
);


--
-- Name: tester_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.tester_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tester_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tester_jobs_id_seq OWNED BY public.tester_jobs.id;


--
-- Name: time_tracks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.time_tracks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    issue_id bigint NOT NULL,
    user_id bigint NOT NULL,
    description character varying(500),
    started_at timestamp with time zone NOT NULL,
    ended_at timestamp with time zone,
    duration bigint
);


--
-- Name: time_tracks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.time_tracks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: time_tracks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.time_tracks_id_seq OWNED BY public.time_tracks.id;


--
-- Name: tool_call_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.tool_call_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    agent_task_id bigint,
    tool_id bigint NOT NULL,
    agent_id bigint,
    input_params text DEFAULT '{}'::text,
    output_result text DEFAULT '{}'::text,
    status character varying(20) DEFAULT 'success'::character varying,
    error_message text,
    duration_ms bigint DEFAULT 0,
    rate_limited boolean DEFAULT false
);


--
-- Name: tool_call_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.tool_call_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tool_call_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tool_call_logs_id_seq OWNED BY public.tool_call_logs.id;


--
-- Name: tool_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.tool_permissions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    agent_template_id bigint,
    tool_id bigint NOT NULL,
    allowed boolean DEFAULT true
);


--
-- Name: tool_permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.tool_permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tool_permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tool_permissions_id_seq OWNED BY public.tool_permissions.id;


--
-- Name: tools; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.tools (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(128) NOT NULL,
    description text,
    category character varying(50) DEFAULT 'general'::character varying,
    is_builtin boolean DEFAULT false,
    status character varying(20) DEFAULT 'active'::character varying,
    tool_type character varying(30) NOT NULL,
    endpoint character varying(500),
    method character varying(10),
    auth_type character varying(20),
    auth_config text DEFAULT '{}'::text,
    params text DEFAULT '{}'::text,
    rate_limit bigint DEFAULT 0,
    timeout bigint DEFAULT 30,
    workspace_id bigint,
    mcp_config_id bigint
);


--
-- Name: tools_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.tools_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tools_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tools_id_seq OWNED BY public.tools.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    email character varying(255) NOT NULL,
    username character varying(128) NOT NULL,
    display_name character varying(255) DEFAULT ''::character varying,
    first_name character varying(255),
    last_name character varying(255),
    avatar text,
    password_hash character varying(255) NOT NULL,
    is_active boolean DEFAULT true,
    is_superuser boolean DEFAULT false,
    is_email_verified boolean DEFAULT false,
    user_timezone character varying(255) DEFAULT 'UTC'::character varying,
    last_active timestamp with time zone
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: webhooks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.webhooks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    url character varying(500) NOT NULL,
    secret character varying(255),
    events character varying(500) DEFAULT 'issue_created,issue_updated,state_changed'::character varying NOT NULL,
    is_active boolean DEFAULT true,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: webhooks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: webhooks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.webhooks_id_seq OWNED BY public.webhooks.id;


--
-- Name: work_item_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.work_item_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    issue_type_id bigint,
    defaults jsonb,
    is_default boolean DEFAULT false,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: work_item_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.work_item_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: work_item_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.work_item_templates_id_seq OWNED BY public.work_item_templates.id;


--
-- Name: workflow_edges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workflow_edges (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workflow_id bigint NOT NULL,
    source_node_id bigint NOT NULL,
    target_node_id bigint NOT NULL,
    condition text,
    context_mapping jsonb
);


--
-- Name: workflow_edges_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.workflow_edges_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: workflow_edges_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.workflow_edges_id_seq OWNED BY public.workflow_edges.id;


--
-- Name: workflow_node_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workflow_node_runs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workflow_run_id bigint NOT NULL,
    node_id bigint NOT NULL,
    agent_id bigint NOT NULL,
    agent_task_id bigint,
    status character varying(20) DEFAULT 'pending'::character varying,
    input_context jsonb,
    output_context jsonb,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    tokens_used bigint,
    cost numeric,
    error_info text,
    retry_count bigint DEFAULT 0
);


--
-- Name: workflow_node_runs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.workflow_node_runs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: workflow_node_runs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.workflow_node_runs_id_seq OWNED BY public.workflow_node_runs.id;


--
-- Name: workflow_nodes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workflow_nodes (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workflow_id bigint NOT NULL,
    agent_id bigint NOT NULL,
    node_type character varying(20) DEFAULT 'agent'::character varying,
    name character varying(100) NOT NULL,
    config jsonb,
    context_config jsonb,
    sort_order bigint DEFAULT 0,
    timeout bigint DEFAULT 1800,
    retry_policy character varying(20) DEFAULT 'retry'::character varying,
    max_retries bigint DEFAULT 3
);


--
-- Name: workflow_nodes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.workflow_nodes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: workflow_nodes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.workflow_nodes_id_seq OWNED BY public.workflow_nodes.id;


--
-- Name: workflow_runs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workflow_runs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workflow_id bigint NOT NULL,
    issue_id bigint,
    status character varying(20) DEFAULT 'pending'::character varying,
    current_node bigint,
    context jsonb,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    total_tokens bigint,
    total_cost numeric,
    error_info text
);


--
-- Name: workflow_runs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.workflow_runs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: workflow_runs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.workflow_runs_id_seq OWNED BY public.workflow_runs.id;


--
-- Name: workflows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workflows (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description text,
    project_id bigint,
    workspace_id bigint NOT NULL,
    issue_type_id bigint,
    issue_type_ids jsonb,
    is_active boolean DEFAULT true
);


--
-- Name: workflows_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.workflows_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: workflows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.workflows_id_seq OWNED BY public.workflows.id;


--
-- Name: workspace_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workspace_members (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    workspace_id bigint NOT NULL,
    user_id bigint NOT NULL,
    role bigint DEFAULT 15,
    is_active boolean DEFAULT true
);


--
-- Name: workspace_members_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.workspace_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: workspace_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.workspace_members_id_seq OWNED BY public.workspace_members.id;


--
-- Name: workspaces; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS public.workspaces (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    slug character varying(50) NOT NULL,
    logo_url character varying(800),
    organization_size character varying(50),
    timezone character varying(255) DEFAULT 'UTC'::character varying,
    owner_id bigint NOT NULL
);


--
-- Name: workspaces_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE IF NOT EXISTS public.workspaces_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: workspaces_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.workspaces_id_seq OWNED BY public.workspaces.id;


--
-- Name: agent_activities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_activities ALTER COLUMN id SET DEFAULT nextval('public.agent_activities_id_seq'::regclass);


--
-- Name: agent_configs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_configs ALTER COLUMN id SET DEFAULT nextval('public.agent_configs_id_seq'::regclass);


--
-- Name: agent_cost_budgets id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_cost_budgets ALTER COLUMN id SET DEFAULT nextval('public.agent_cost_budgets_id_seq'::regclass);


--
-- Name: agent_decisions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_decisions ALTER COLUMN id SET DEFAULT nextval('public.agent_decisions_id_seq'::regclass);


--
-- Name: agent_loop_iterations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_loop_iterations ALTER COLUMN id SET DEFAULT nextval('public.agent_loop_iterations_id_seq'::regclass);


--
-- Name: agent_loop_runs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_loop_runs ALTER COLUMN id SET DEFAULT nextval('public.agent_loop_runs_id_seq'::regclass);


--
-- Name: agent_loops id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_loops ALTER COLUMN id SET DEFAULT nextval('public.agent_loops_id_seq'::regclass);


--
-- Name: agent_pipeline_runs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_pipeline_runs ALTER COLUMN id SET DEFAULT nextval('public.agent_pipeline_runs_id_seq'::regclass);


--
-- Name: agent_pipelines id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_pipelines ALTER COLUMN id SET DEFAULT nextval('public.agent_pipelines_id_seq'::regclass);


--
-- Name: agent_registry id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_registry ALTER COLUMN id SET DEFAULT nextval('public.agent_registry_id_seq'::regclass);


--
-- Name: agent_slas id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_slas ALTER COLUMN id SET DEFAULT nextval('public.agent_slas_id_seq'::regclass);


--
-- Name: agent_task_snapshots id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_task_snapshots ALTER COLUMN id SET DEFAULT nextval('public.agent_task_snapshots_id_seq'::regclass);


--
-- Name: agent_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_tasks ALTER COLUMN id SET DEFAULT nextval('public.agent_tasks_id_seq'::regclass);


--
-- Name: agent_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_templates ALTER COLUMN id SET DEFAULT nextval('public.agent_templates_id_seq'::regclass);


--
-- Name: agent_workflows id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_workflows ALTER COLUMN id SET DEFAULT nextval('public.agent_workflows_id_seq'::regclass);


--
-- Name: agents id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agents ALTER COLUMN id SET DEFAULT nextval('public.agents_id_seq'::regclass);


--
-- Name: ai_configs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_configs ALTER COLUMN id SET DEFAULT nextval('public.ai_configs_id_seq'::regclass);


--
-- Name: ai_messages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_messages ALTER COLUMN id SET DEFAULT nextval('public.ai_messages_id_seq'::regclass);


--
-- Name: ai_threads id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_threads ALTER COLUMN id SET DEFAULT nextval('public.ai_threads_id_seq'::regclass);


--
-- Name: approval_records id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approval_records ALTER COLUMN id SET DEFAULT nextval('public.approval_records_id_seq'::regclass);


--
-- Name: approvals id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.approvals ALTER COLUMN id SET DEFAULT nextval('public.approvals_id_seq'::regclass);


--
-- Name: attachments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attachments ALTER COLUMN id SET DEFAULT nextval('public.attachments_id_seq'::regclass);


--
-- Name: automation_executions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.automation_executions ALTER COLUMN id SET DEFAULT nextval('public.automation_executions_id_seq'::regclass);


--
-- Name: automation_rule_overrides id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.automation_rule_overrides ALTER COLUMN id SET DEFAULT nextval('public.automation_rule_overrides_id_seq'::regclass);


--
-- Name: automation_rules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.automation_rules ALTER COLUMN id SET DEFAULT nextval('public.automation_rules_id_seq'::regclass);


--
-- Name: autopilot_executions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.autopilot_executions ALTER COLUMN id SET DEFAULT nextval('public.autopilot_executions_id_seq'::regclass);


--
-- Name: autopilot_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.autopilot_tasks ALTER COLUMN id SET DEFAULT nextval('public.autopilot_tasks_id_seq'::regclass);


--
-- Name: autopilot_webhooks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.autopilot_webhooks ALTER COLUMN id SET DEFAULT nextval('public.autopilot_webhooks_id_seq'::regclass);


--
-- Name: build_records id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.build_records ALTER COLUMN id SET DEFAULT nextval('public.build_records_id_seq'::regclass);


--
-- Name: chats id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chats ALTER COLUMN id SET DEFAULT nextval('public.chats_id_seq'::regclass);


--
-- Name: cicd_configs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cicd_configs ALTER COLUMN id SET DEFAULT nextval('public.cicd_configs_id_seq'::regclass);


--
-- Name: comments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments ALTER COLUMN id SET DEFAULT nextval('public.comments_id_seq'::regclass);


--
-- Name: conditional_fields id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conditional_fields ALTER COLUMN id SET DEFAULT nextval('public.conditional_fields_id_seq'::regclass);


--
-- Name: custom_field_options id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.custom_field_options ALTER COLUMN id SET DEFAULT nextval('public.custom_field_options_id_seq'::regclass);


--
-- Name: custom_fields id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.custom_fields ALTER COLUMN id SET DEFAULT nextval('public.custom_fields_id_seq'::regclass);


--
-- Name: cycles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cycles ALTER COLUMN id SET DEFAULT nextval('public.cycles_id_seq'::regclass);


--
-- Name: dashboard_widgets id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dashboard_widgets ALTER COLUMN id SET DEFAULT nextval('public.dashboard_widgets_id_seq'::regclass);


--
-- Name: developer_jobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.developer_jobs ALTER COLUMN id SET DEFAULT nextval('public.developer_jobs_id_seq'::regclass);


--
-- Name: estimate_categories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.estimate_categories ALTER COLUMN id SET DEFAULT nextval('public.estimate_categories_id_seq'::regclass);


--
-- Name: estimate_points id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.estimate_points ALTER COLUMN id SET DEFAULT nextval('public.estimate_points_id_seq'::regclass);


--
-- Name: estimate_times id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.estimate_times ALTER COLUMN id SET DEFAULT nextval('public.estimate_times_id_seq'::regclass);


--
-- Name: field_permissions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_permissions ALTER COLUMN id SET DEFAULT nextval('public.field_permissions_id_seq'::regclass);


--
-- Name: git_integrations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.git_integrations ALTER COLUMN id SET DEFAULT nextval('public.git_integrations_id_seq'::regclass);


--
-- Name: git_issue_links id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.git_issue_links ALTER COLUMN id SET DEFAULT nextval('public.git_issue_links_id_seq'::regclass);


--
-- Name: github_connections id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_connections ALTER COLUMN id SET DEFAULT nextval('public.github_connections_id_seq'::regclass);


--
-- Name: initiatives id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.initiatives ALTER COLUMN id SET DEFAULT nextval('public.initiatives_id_seq'::regclass);


--
-- Name: issue_activities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_activities ALTER COLUMN id SET DEFAULT nextval('public.issue_activities_id_seq'::regclass);


--
-- Name: issue_custom_field_values id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_custom_field_values ALTER COLUMN id SET DEFAULT nextval('public.issue_custom_field_values_id_seq'::regclass);


--
-- Name: issue_relations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_relations ALTER COLUMN id SET DEFAULT nextval('public.issue_relations_id_seq'::regclass);


--
-- Name: issue_type_imports id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_imports ALTER COLUMN id SET DEFAULT nextval('public.issue_type_imports_id_seq'::regclass);


--
-- Name: issue_type_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_templates ALTER COLUMN id SET DEFAULT nextval('public.issue_type_templates_id_seq'::regclass);


--
-- Name: issue_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_types ALTER COLUMN id SET DEFAULT nextval('public.issue_types_id_seq'::regclass);


--
-- Name: issue_watchers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_watchers ALTER COLUMN id SET DEFAULT nextval('public.issue_watchers_id_seq'::regclass);


--
-- Name: issues id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issues ALTER COLUMN id SET DEFAULT nextval('public.issues_id_seq'::regclass);


--
-- Name: labels id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labels ALTER COLUMN id SET DEFAULT nextval('public.labels_id_seq'::regclass);


--
-- Name: mcp_configs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mcp_configs ALTER COLUMN id SET DEFAULT nextval('public.mcp_configs_id_seq'::regclass);


--
-- Name: memory_entries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.memory_entries ALTER COLUMN id SET DEFAULT nextval('public.memory_entries_id_seq'::regclass);


--
-- Name: memory_vector_indexes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.memory_vector_indexes ALTER COLUMN id SET DEFAULT nextval('public.memory_vector_indexes_id_seq'::regclass);


--
-- Name: message_reactions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_reactions ALTER COLUMN id SET DEFAULT nextval('public.message_reactions_id_seq'::regclass);


--
-- Name: messages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages ALTER COLUMN id SET DEFAULT nextval('public.messages_id_seq'::regclass);


--
-- Name: metric_charts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.metric_charts ALTER COLUMN id SET DEFAULT nextval('public.metric_charts_id_seq'::regclass);


--
-- Name: module_inheritance_overrides id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_inheritance_overrides ALTER COLUMN id SET DEFAULT nextval('public.module_inheritance_overrides_id_seq'::regclass);


--
-- Name: modules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modules ALTER COLUMN id SET DEFAULT nextval('public.modules_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: page_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.page_templates ALTER COLUMN id SET DEFAULT nextval('public.page_templates_id_seq'::regclass);


--
-- Name: page_versions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.page_versions ALTER COLUMN id SET DEFAULT nextval('public.page_versions_id_seq'::regclass);


--
-- Name: pages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pages ALTER COLUMN id SET DEFAULT nextval('public.pages_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: personal_access_tokens id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.personal_access_tokens ALTER COLUMN id SET DEFAULT nextval('public.personal_access_tokens_id_seq'::regclass);


--
-- Name: plugin_event_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plugin_event_logs ALTER COLUMN id SET DEFAULT nextval('public.plugin_event_logs_id_seq'::regclass);


--
-- Name: plugins id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plugins ALTER COLUMN id SET DEFAULT nextval('public.plugins_id_seq'::regclass);


--
-- Name: project_agent_members id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_agent_members ALTER COLUMN id SET DEFAULT nextval('public.project_agent_members_id_seq'::regclass);


--
-- Name: project_estimate_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_estimate_settings ALTER COLUMN id SET DEFAULT nextval('public.project_estimate_settings_id_seq'::regclass);


--
-- Name: project_intake_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_intake_settings ALTER COLUMN id SET DEFAULT nextval('public.project_intake_settings_id_seq'::regclass);


--
-- Name: project_members id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members ALTER COLUMN id SET DEFAULT nextval('public.project_members_id_seq'::regclass);


--
-- Name: project_page_tabs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_page_tabs ALTER COLUMN id SET DEFAULT nextval('public.project_page_tabs_id_seq'::regclass);


--
-- Name: project_subscribers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_subscribers ALTER COLUMN id SET DEFAULT nextval('public.project_subscribers_id_seq'::regclass);


--
-- Name: project_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_templates ALTER COLUMN id SET DEFAULT nextval('public.project_templates_id_seq'::regclass);


--
-- Name: project_updates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_updates ALTER COLUMN id SET DEFAULT nextval('public.project_updates_id_seq'::regclass);


--
-- Name: projects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects ALTER COLUMN id SET DEFAULT nextval('public.projects_id_seq'::regclass);


--
-- Name: recurrence_rules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recurrence_rules ALTER COLUMN id SET DEFAULT nextval('public.recurrence_rules_id_seq'::regclass);


--
-- Name: relation_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.relation_types ALTER COLUMN id SET DEFAULT nextval('public.relation_types_id_seq'::regclass);


--
-- Name: releases id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.releases ALTER COLUMN id SET DEFAULT nextval('public.releases_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- Name: runtimes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.runtimes ALTER COLUMN id SET DEFAULT nextval('public.runtimes_id_seq'::regclass);


--
-- Name: saved_dashboards id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_dashboards ALTER COLUMN id SET DEFAULT nextval('public.saved_dashboards_id_seq'::regclass);


--
-- Name: saved_reports id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_reports ALTER COLUMN id SET DEFAULT nextval('public.saved_reports_id_seq'::regclass);


--
-- Name: saved_views id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_views ALTER COLUMN id SET DEFAULT nextval('public.saved_views_id_seq'::regclass);


--
-- Name: sdlc_stages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sdlc_stages ALTER COLUMN id SET DEFAULT nextval('public.sdlc_stages_id_seq'::regclass);


--
-- Name: sdlc_workflows id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sdlc_workflows ALTER COLUMN id SET DEFAULT nextval('public.sdlc_workflows_id_seq'::regclass);


--
-- Name: search_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.search_templates ALTER COLUMN id SET DEFAULT nextval('public.search_templates_id_seq'::regclass);


--
-- Name: skill_execution_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skill_execution_logs ALTER COLUMN id SET DEFAULT nextval('public.skill_execution_logs_id_seq'::regclass);


--
-- Name: skills id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skills ALTER COLUMN id SET DEFAULT nextval('public.skills_id_seq'::regclass);


--
-- Name: slack_connections id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_connections ALTER COLUMN id SET DEFAULT nextval('public.slack_connections_id_seq'::regclass);


--
-- Name: squad_executions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squad_executions ALTER COLUMN id SET DEFAULT nextval('public.squad_executions_id_seq'::regclass);


--
-- Name: squad_members id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squad_members ALTER COLUMN id SET DEFAULT nextval('public.squad_members_id_seq'::regclass);


--
-- Name: squad_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squad_tasks ALTER COLUMN id SET DEFAULT nextval('public.squad_tasks_id_seq'::regclass);


--
-- Name: squads id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.squads ALTER COLUMN id SET DEFAULT nextval('public.squads_id_seq'::regclass);


--
-- Name: state_transitions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_transitions ALTER COLUMN id SET DEFAULT nextval('public.state_transitions_id_seq'::regclass);


--
-- Name: states id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.states ALTER COLUMN id SET DEFAULT nextval('public.states_id_seq'::regclass);


--
-- Name: task_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_logs ALTER COLUMN id SET DEFAULT nextval('public.task_logs_id_seq'::regclass);


--
-- Name: tester_jobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tester_jobs ALTER COLUMN id SET DEFAULT nextval('public.tester_jobs_id_seq'::regclass);


--
-- Name: time_tracks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_tracks ALTER COLUMN id SET DEFAULT nextval('public.time_tracks_id_seq'::regclass);


--
-- Name: tool_call_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tool_call_logs ALTER COLUMN id SET DEFAULT nextval('public.tool_call_logs_id_seq'::regclass);


--
-- Name: tool_permissions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tool_permissions ALTER COLUMN id SET DEFAULT nextval('public.tool_permissions_id_seq'::regclass);


--
-- Name: tools id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tools ALTER COLUMN id SET DEFAULT nextval('public.tools_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: webhooks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhooks ALTER COLUMN id SET DEFAULT nextval('public.webhooks_id_seq'::regclass);


--
-- Name: work_item_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_item_templates ALTER COLUMN id SET DEFAULT nextval('public.work_item_templates_id_seq'::regclass);


--
-- Name: workflow_edges id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_edges ALTER COLUMN id SET DEFAULT nextval('public.workflow_edges_id_seq'::regclass);


--
-- Name: workflow_node_runs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_node_runs ALTER COLUMN id SET DEFAULT nextval('public.workflow_node_runs_id_seq'::regclass);


--
-- Name: workflow_nodes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_nodes ALTER COLUMN id SET DEFAULT nextval('public.workflow_nodes_id_seq'::regclass);


--
-- Name: workflow_runs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflow_runs ALTER COLUMN id SET DEFAULT nextval('public.workflow_runs_id_seq'::regclass);


--
-- Name: workflows id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflows ALTER COLUMN id SET DEFAULT nextval('public.workflows_id_seq'::regclass);


--
-- Name: workspace_members id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_members ALTER COLUMN id SET DEFAULT nextval('public.workspace_members_id_seq'::regclass);


--
-- Name: workspaces id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces ALTER COLUMN id SET DEFAULT nextval('public.workspaces_id_seq'::regclass);


--
-- Name: agent_activities agent_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_activities
    ADD CONSTRAINT agent_activities_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_configs agent_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_configs
    ADD CONSTRAINT agent_configs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_cost_budgets agent_cost_budgets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_cost_budgets
    ADD CONSTRAINT agent_cost_budgets_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_decisions agent_decisions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_decisions
    ADD CONSTRAINT agent_decisions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_loop_iterations agent_loop_iterations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_loop_iterations
    ADD CONSTRAINT agent_loop_iterations_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_loop_runs agent_loop_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_loop_runs
    ADD CONSTRAINT agent_loop_runs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_loops agent_loops_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_loops
    ADD CONSTRAINT agent_loops_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_pipeline_runs agent_pipeline_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_pipeline_runs
    ADD CONSTRAINT agent_pipeline_runs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_pipelines agent_pipelines_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_pipelines
    ADD CONSTRAINT agent_pipelines_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_registry agent_registry_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_registry
    ADD CONSTRAINT agent_registry_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_sessions agent_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_sessions
    ADD CONSTRAINT agent_sessions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_slas agent_slas_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_slas
    ADD CONSTRAINT agent_slas_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_task_snapshots agent_task_snapshots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_task_snapshots
    ADD CONSTRAINT agent_task_snapshots_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_tasks agent_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_tasks
    ADD CONSTRAINT agent_tasks_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_templates agent_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_templates
    ADD CONSTRAINT agent_templates_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_workflows agent_workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_workflows
    ADD CONSTRAINT agent_workflows_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agents agents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: ai_configs ai_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.ai_configs
    ADD CONSTRAINT ai_configs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: ai_messages ai_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.ai_messages
    ADD CONSTRAINT ai_messages_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: ai_threads ai_threads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.ai_threads
    ADD CONSTRAINT ai_threads_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: approval_records approval_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.approval_records
    ADD CONSTRAINT approval_records_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: approvals approvals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.approvals
    ADD CONSTRAINT approvals_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: attachments attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: automation_executions automation_executions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.automation_executions
    ADD CONSTRAINT automation_executions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: automation_rule_overrides automation_rule_overrides_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.automation_rule_overrides
    ADD CONSTRAINT automation_rule_overrides_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: automation_rules automation_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.automation_rules
    ADD CONSTRAINT automation_rules_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: autopilot_executions autopilot_executions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.autopilot_executions
    ADD CONSTRAINT autopilot_executions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: autopilot_tasks autopilot_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.autopilot_tasks
    ADD CONSTRAINT autopilot_tasks_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: autopilot_webhooks autopilot_webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.autopilot_webhooks
    ADD CONSTRAINT autopilot_webhooks_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: build_records build_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.build_records
    ADD CONSTRAINT build_records_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: chats chats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: cicd_configs cicd_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.cicd_configs
    ADD CONSTRAINT cicd_configs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: conditional_fields conditional_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.conditional_fields
    ADD CONSTRAINT conditional_fields_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: custom_field_options custom_field_options_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.custom_field_options
    ADD CONSTRAINT custom_field_options_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: custom_fields custom_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.custom_fields
    ADD CONSTRAINT custom_fields_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: cycles cycles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.cycles
    ADD CONSTRAINT cycles_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: dashboard_widgets dashboard_widgets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.dashboard_widgets
    ADD CONSTRAINT dashboard_widgets_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: data_migrations data_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.data_migrations
    ADD CONSTRAINT data_migrations_pkey PRIMARY KEY (name);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: developer_jobs developer_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.developer_jobs
    ADD CONSTRAINT developer_jobs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: estimate_categories estimate_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.estimate_categories
    ADD CONSTRAINT estimate_categories_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: estimate_points estimate_points_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.estimate_points
    ADD CONSTRAINT estimate_points_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: estimate_times estimate_times_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.estimate_times
    ADD CONSTRAINT estimate_times_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: field_permissions field_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.field_permissions
    ADD CONSTRAINT field_permissions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: git_integrations git_integrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.git_integrations
    ADD CONSTRAINT git_integrations_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: git_issue_links git_issue_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.git_issue_links
    ADD CONSTRAINT git_issue_links_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: github_connections github_connections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.github_connections
    ADD CONSTRAINT github_connections_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: initiative_projects initiative_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.initiative_projects
    ADD CONSTRAINT initiative_projects_pkey PRIMARY KEY (initiative_id, project_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: initiatives initiatives_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.initiatives
    ADD CONSTRAINT initiatives_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_activities issue_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_activities
    ADD CONSTRAINT issue_activities_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_assignees issue_assignees_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_assignees
    ADD CONSTRAINT issue_assignees_pkey PRIMARY KEY (issue_id, user_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_custom_field_values issue_custom_field_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_custom_field_values
    ADD CONSTRAINT issue_custom_field_values_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_cycles issue_cycles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_cycles
    ADD CONSTRAINT issue_cycles_pkey PRIMARY KEY (issue_id, cycle_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_labels issue_labels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_labels
    ADD CONSTRAINT issue_labels_pkey PRIMARY KEY (issue_id, label_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_pages issue_pages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_pages
    ADD CONSTRAINT issue_pages_pkey PRIMARY KEY (issue_id, page_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_relations issue_relations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_relations
    ADD CONSTRAINT issue_relations_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_fields issue_type_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_fields
    ADD CONSTRAINT issue_type_fields_pkey PRIMARY KEY (type_id, field_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_imports issue_type_imports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_imports
    ADD CONSTRAINT issue_type_imports_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_template_fields issue_type_template_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_template_fields
    ADD CONSTRAINT issue_type_template_fields_pkey PRIMARY KEY (template_type_id, field_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_templates issue_type_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_templates
    ADD CONSTRAINT issue_type_templates_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_types issue_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_types
    ADD CONSTRAINT issue_types_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_watchers issue_watchers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_watchers
    ADD CONSTRAINT issue_watchers_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issues issues_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issues
    ADD CONSTRAINT issues_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: labels labels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.labels
    ADD CONSTRAINT labels_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: mcp_configs mcp_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.mcp_configs
    ADD CONSTRAINT mcp_configs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: memory_entries memory_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.memory_entries
    ADD CONSTRAINT memory_entries_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: memory_sessions memory_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.memory_sessions
    ADD CONSTRAINT memory_sessions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: memory_vector_indexes memory_vector_indexes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.memory_vector_indexes
    ADD CONSTRAINT memory_vector_indexes_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: message_reactions message_reactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.message_reactions
    ADD CONSTRAINT message_reactions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: metric_charts metric_charts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.metric_charts
    ADD CONSTRAINT metric_charts_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: module_inheritance_overrides module_inheritance_overrides_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.module_inheritance_overrides
    ADD CONSTRAINT module_inheritance_overrides_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: module_issues module_issues_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.module_issues
    ADD CONSTRAINT module_issues_pkey PRIMARY KEY (module_id, issue_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: modules modules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.modules
    ADD CONSTRAINT modules_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: page_templates page_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.page_templates
    ADD CONSTRAINT page_templates_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: page_versions page_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.page_versions
    ADD CONSTRAINT page_versions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: pages pages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: personal_access_tokens personal_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.personal_access_tokens
    ADD CONSTRAINT personal_access_tokens_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: plugin_event_logs plugin_event_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.plugin_event_logs
    ADD CONSTRAINT plugin_event_logs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: plugins plugins_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.plugins
    ADD CONSTRAINT plugins_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_agent_members project_agent_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_agent_members
    ADD CONSTRAINT project_agent_members_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_custom_field_enrollments project_custom_field_enrollments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_custom_field_enrollments
    ADD CONSTRAINT project_custom_field_enrollments_pkey PRIMARY KEY (project_id, field_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_estimate_settings project_estimate_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_estimate_settings
    ADD CONSTRAINT project_estimate_settings_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_intake_settings project_intake_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_intake_settings
    ADD CONSTRAINT project_intake_settings_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_members project_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_page_tabs project_page_tabs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_page_tabs
    ADD CONSTRAINT project_page_tabs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_subscribers project_subscribers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_subscribers
    ADD CONSTRAINT project_subscribers_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_template_types project_template_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_template_types
    ADD CONSTRAINT project_template_types_pkey PRIMARY KEY (template_id, type_template_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_templates project_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_templates
    ADD CONSTRAINT project_templates_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_updates project_updates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_updates
    ADD CONSTRAINT project_updates_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: recurrence_rules recurrence_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.recurrence_rules
    ADD CONSTRAINT recurrence_rules_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: relation_types relation_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.relation_types
    ADD CONSTRAINT relation_types_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: release_issues release_issues_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.release_issues
    ADD CONSTRAINT release_issues_pkey PRIMARY KEY (release_id, issue_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: releases releases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.releases
    ADD CONSTRAINT releases_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: runtimes runtimes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.runtimes
    ADD CONSTRAINT runtimes_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: saved_dashboards saved_dashboards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.saved_dashboards
    ADD CONSTRAINT saved_dashboards_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: saved_reports saved_reports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.saved_reports
    ADD CONSTRAINT saved_reports_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: saved_views saved_views_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.saved_views
    ADD CONSTRAINT saved_views_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: sdlc_stages sdlc_stages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.sdlc_stages
    ADD CONSTRAINT sdlc_stages_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: sdlc_workflows sdlc_workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.sdlc_workflows
    ADD CONSTRAINT sdlc_workflows_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: search_templates search_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.search_templates
    ADD CONSTRAINT search_templates_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: skill_execution_logs skill_execution_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.skill_execution_logs
    ADD CONSTRAINT skill_execution_logs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: skills skills_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.skills
    ADD CONSTRAINT skills_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: slack_connections slack_connections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.slack_connections
    ADD CONSTRAINT slack_connections_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: squad_executions squad_executions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.squad_executions
    ADD CONSTRAINT squad_executions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: squad_members squad_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.squad_members
    ADD CONSTRAINT squad_members_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: squad_tasks squad_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.squad_tasks
    ADD CONSTRAINT squad_tasks_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: squads squads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.squads
    ADD CONSTRAINT squads_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: state_transitions state_transitions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.state_transitions
    ADD CONSTRAINT state_transitions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: states states_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.states
    ADD CONSTRAINT states_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: task_logs task_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.task_logs
    ADD CONSTRAINT task_logs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tester_jobs tester_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tester_jobs
    ADD CONSTRAINT tester_jobs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: time_tracks time_tracks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.time_tracks
    ADD CONSTRAINT time_tracks_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tool_call_logs tool_call_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tool_call_logs
    ADD CONSTRAINT tool_call_logs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tool_permissions tool_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tool_permissions
    ADD CONSTRAINT tool_permissions_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tools tools_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tools
    ADD CONSTRAINT tools_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_cost_budgets uni_agent_cost_budgets_project_id; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_cost_budgets
    ADD CONSTRAINT uni_agent_cost_budgets_project_id UNIQUE (project_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_slas uni_agent_slas_project_id; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_slas
    ADD CONSTRAINT uni_agent_slas_project_id UNIQUE (project_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: autopilot_webhooks uni_autopilot_webhooks_endpoint; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.autopilot_webhooks
    ADD CONSTRAINT uni_autopilot_webhooks_endpoint UNIQUE (endpoint);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_estimate_settings uni_project_estimate_settings_project_id; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_estimate_settings
    ADD CONSTRAINT uni_project_estimate_settings_project_id UNIQUE (project_id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: webhooks webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: work_item_templates work_item_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.work_item_templates
    ADD CONSTRAINT work_item_templates_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_edges workflow_edges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_edges
    ADD CONSTRAINT workflow_edges_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_node_runs workflow_node_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_node_runs
    ADD CONSTRAINT workflow_node_runs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_nodes workflow_nodes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_nodes
    ADD CONSTRAINT workflow_nodes_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_runs workflow_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_runs
    ADD CONSTRAINT workflow_runs_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflows workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT workflows_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workspace_members workspace_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT workspace_members_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workspaces workspaces_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT workspaces_pkey PRIMARY KEY (id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: idx_agent_activities_agent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_activities_agent_id ON public.agent_activities USING btree (agent_id);


--
-- Name: idx_agent_activities_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_activities_deleted_at ON public.agent_activities USING btree (deleted_at);


--
-- Name: idx_agent_activities_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_activities_issue_id ON public.agent_activities USING btree (issue_id);


--
-- Name: idx_agent_configs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_configs_deleted_at ON public.agent_configs USING btree (deleted_at);


--
-- Name: idx_agent_configs_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_configs_workspace_id ON public.agent_configs USING btree (workspace_id);


--
-- Name: idx_agent_decisions_agent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_decisions_agent_id ON public.agent_decisions USING btree (agent_id);


--
-- Name: idx_agent_decisions_agent_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_decisions_agent_task_id ON public.agent_decisions USING btree (agent_task_id);


--
-- Name: idx_agent_decisions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_decisions_deleted_at ON public.agent_decisions USING btree (deleted_at);


--
-- Name: idx_agent_decisions_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_decisions_issue_id ON public.agent_decisions USING btree (issue_id);


--
-- Name: idx_agent_decisions_workflow_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_decisions_workflow_run_id ON public.agent_decisions USING btree (workflow_run_id);


--
-- Name: idx_agent_loop_iterations_loop_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_loop_iterations_loop_run_id ON public.agent_loop_iterations USING btree (loop_run_id);


--
-- Name: idx_agent_loop_runs_loop_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_loop_runs_loop_id ON public.agent_loop_runs USING btree (loop_id);


--
-- Name: idx_agent_loops_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_loops_deleted_at ON public.agent_loops USING btree (deleted_at);


--
-- Name: idx_agent_loops_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_loops_workspace_id ON public.agent_loops USING btree (workspace_id);


--
-- Name: idx_agent_pipeline_runs_pipeline_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_pipeline_runs_pipeline_id ON public.agent_pipeline_runs USING btree (pipeline_id);


--
-- Name: idx_agent_pipelines_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_pipelines_deleted_at ON public.agent_pipelines USING btree (deleted_at);


--
-- Name: idx_agent_pipelines_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_pipelines_workspace_id ON public.agent_pipelines USING btree (workspace_id);


--
-- Name: idx_agent_registry_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_registry_deleted_at ON public.agent_registry USING btree (deleted_at);


--
-- Name: idx_agent_registry_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_registry_name ON public.agent_registry USING btree (name);


--
-- Name: idx_agent_registry_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_registry_workspace_id ON public.agent_registry USING btree (workspace_id);


--
-- Name: idx_agent_sessions_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_sessions_workspace_id ON public.agent_sessions USING btree (workspace_id);


--
-- Name: idx_agent_task_snapshots_agent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_task_snapshots_agent_id ON public.agent_task_snapshots USING btree (agent_id);


--
-- Name: idx_agent_task_snapshots_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_task_snapshots_deleted_at ON public.agent_task_snapshots USING btree (deleted_at);


--
-- Name: idx_agent_task_snapshots_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_task_snapshots_task_id ON public.agent_task_snapshots USING btree (task_id);


--
-- Name: idx_agent_tasks_agent_config_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_tasks_agent_config_id ON public.agent_tasks USING btree (agent_config_id);


--
-- Name: idx_agent_tasks_agent_template_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_tasks_agent_template_id ON public.agent_tasks USING btree (agent_template_id);


--
-- Name: idx_agent_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_tasks_deleted_at ON public.agent_tasks USING btree (deleted_at);


--
-- Name: idx_agent_tasks_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_tasks_issue_id ON public.agent_tasks USING btree (issue_id);


--
-- Name: idx_agent_tasks_parent_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_tasks_parent_task_id ON public.agent_tasks USING btree (parent_task_id);


--
-- Name: idx_agent_tasks_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_tasks_project_id ON public.agent_tasks USING btree (project_id);


--
-- Name: idx_agent_tasks_rerun_of_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_tasks_rerun_of_task_id ON public.agent_tasks USING btree (rerun_of_task_id);


--
-- Name: idx_agent_tasks_retry_of_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_tasks_retry_of_task_id ON public.agent_tasks USING btree (retry_of_task_id);


--
-- Name: idx_agent_tasks_runtime_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_tasks_runtime_id ON public.agent_tasks USING btree (runtime_id);


--
-- Name: idx_agent_tasks_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_tasks_workspace_id ON public.agent_tasks USING btree (workspace_id);


--
-- Name: idx_agent_templates_default_skill_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_templates_default_skill_id ON public.agent_templates USING btree (default_skill_id);


--
-- Name: idx_agent_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_templates_deleted_at ON public.agent_templates USING btree (deleted_at);


--
-- Name: idx_agent_templates_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_templates_name ON public.agent_templates USING btree (name);


--
-- Name: idx_agent_templates_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_templates_workspace_id ON public.agent_templates USING btree (workspace_id);


--
-- Name: idx_agent_workflows_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_workflows_deleted_at ON public.agent_workflows USING btree (deleted_at);


--
-- Name: idx_agent_workflows_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_workflows_project_id ON public.agent_workflows USING btree (project_id);


--
-- Name: idx_agent_workflows_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agent_workflows_workspace_id ON public.agent_workflows USING btree (workspace_id);


--
-- Name: idx_agents_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agents_deleted_at ON public.agents USING btree (deleted_at);


--
-- Name: idx_agents_running_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agents_running_task_id ON public.agents USING btree (running_task_id);


--
-- Name: idx_agents_template_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agents_template_id ON public.agents USING btree (template_id);


--
-- Name: idx_agents_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_agents_workspace_id ON public.agents USING btree (workspace_id);


--
-- Name: idx_ai_configs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_ai_configs_deleted_at ON public.ai_configs USING btree (deleted_at);


--
-- Name: idx_ai_configs_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_configs_workspace_id ON public.ai_configs USING btree (workspace_id);


--
-- Name: idx_ai_messages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_ai_messages_deleted_at ON public.ai_messages USING btree (deleted_at);


--
-- Name: idx_ai_messages_thread_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_ai_messages_thread_id ON public.ai_messages USING btree (thread_id);


--
-- Name: idx_ai_threads_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_ai_threads_deleted_at ON public.ai_threads USING btree (deleted_at);


--
-- Name: idx_ai_threads_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_ai_threads_project_id ON public.ai_threads USING btree (project_id);


--
-- Name: idx_ai_threads_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_ai_threads_user_id ON public.ai_threads USING btree (user_id);


--
-- Name: idx_ai_threads_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_ai_threads_workspace_id ON public.ai_threads USING btree (workspace_id);


--
-- Name: idx_approval_records_approval_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_approval_records_approval_id ON public.approval_records USING btree (approval_id);


--
-- Name: idx_approval_records_approver_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_approval_records_approver_id ON public.approval_records USING btree (approver_id);


--
-- Name: idx_approval_records_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_approval_records_deleted_at ON public.approval_records USING btree (deleted_at);


--
-- Name: idx_approvals_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_approvals_deleted_at ON public.approvals USING btree (deleted_at);


--
-- Name: idx_approvals_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_approvals_issue_id ON public.approvals USING btree (issue_id);


--
-- Name: idx_approvals_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_approvals_project_id ON public.approvals USING btree (project_id);


--
-- Name: idx_approvals_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_approvals_status ON public.approvals USING btree (status);


--
-- Name: idx_approvals_transition_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_approvals_transition_id ON public.approvals USING btree (transition_id);


--
-- Name: idx_approvals_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_approvals_workspace_id ON public.approvals USING btree (workspace_id);


--
-- Name: idx_attachments_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_attachments_deleted_at ON public.attachments USING btree (deleted_at);


--
-- Name: idx_attachments_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_attachments_issue_id ON public.attachments USING btree (issue_id);


--
-- Name: idx_automation_executions_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_automation_executions_issue_id ON public.automation_executions USING btree (issue_id);


--
-- Name: idx_automation_executions_rule_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_automation_executions_rule_id ON public.automation_executions USING btree (rule_id);


--
-- Name: idx_automation_rule_overrides_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_automation_rule_overrides_deleted_at ON public.automation_rule_overrides USING btree (deleted_at);


--
-- Name: idx_automation_rule_overrides_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_automation_rule_overrides_project_id ON public.automation_rule_overrides USING btree (project_id);


--
-- Name: idx_automation_rule_overrides_rule_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_automation_rule_overrides_rule_id ON public.automation_rule_overrides USING btree (rule_id);


--
-- Name: idx_automation_rules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_automation_rules_deleted_at ON public.automation_rules USING btree (deleted_at);


--
-- Name: idx_automation_rules_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_automation_rules_project_id ON public.automation_rules USING btree (project_id);


--
-- Name: idx_automation_rules_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_automation_rules_workspace_id ON public.automation_rules USING btree (workspace_id);


--
-- Name: idx_autopilot_executions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_autopilot_executions_deleted_at ON public.autopilot_executions USING btree (deleted_at);


--
-- Name: idx_autopilot_executions_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_autopilot_executions_task_id ON public.autopilot_executions USING btree (task_id);


--
-- Name: idx_autopilot_tasks_agent_config_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_autopilot_tasks_agent_config_id ON public.autopilot_tasks USING btree (agent_config_id);


--
-- Name: idx_autopilot_tasks_agent_template_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_autopilot_tasks_agent_template_id ON public.autopilot_tasks USING btree (agent_template_id);


--
-- Name: idx_autopilot_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_autopilot_tasks_deleted_at ON public.autopilot_tasks USING btree (deleted_at, deleted_at);


--
-- Name: idx_autopilot_tasks_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_autopilot_tasks_project_id ON public.autopilot_tasks USING btree (project_id);


--
-- Name: idx_autopilot_tasks_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_autopilot_tasks_workspace_id ON public.autopilot_tasks USING btree (workspace_id);


--
-- Name: idx_autopilot_webhooks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_autopilot_webhooks_deleted_at ON public.autopilot_webhooks USING btree (deleted_at);


--
-- Name: idx_autopilot_webhooks_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_autopilot_webhooks_task_id ON public.autopilot_webhooks USING btree (task_id);


--
-- Name: idx_autopilot_webhooks_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_autopilot_webhooks_workspace_id ON public.autopilot_webhooks USING btree (workspace_id);


--
-- Name: idx_build_records_agent_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_build_records_agent_task_id ON public.build_records USING btree (agent_task_id);


--
-- Name: idx_build_records_cicd_config_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_build_records_cicd_config_id ON public.build_records USING btree (cicd_config_id);


--
-- Name: idx_build_records_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_build_records_deleted_at ON public.build_records USING btree (deleted_at);


--
-- Name: idx_build_records_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_build_records_issue_id ON public.build_records USING btree (issue_id);


--
-- Name: idx_build_records_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_build_records_project_id ON public.build_records USING btree (project_id);


--
-- Name: idx_build_records_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_build_records_status ON public.build_records USING btree (status);


--
-- Name: idx_build_records_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_build_records_workspace_id ON public.build_records USING btree (workspace_id);


--
-- Name: idx_chats_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_chats_deleted_at ON public.chats USING btree (deleted_at);


--
-- Name: idx_chats_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_chats_issue_id ON public.chats USING btree (issue_id);


--
-- Name: idx_chats_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_chats_project_id ON public.chats USING btree (project_id);


--
-- Name: idx_chats_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_chats_workspace_id ON public.chats USING btree (workspace_id);


--
-- Name: idx_cicd_configs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_cicd_configs_deleted_at ON public.cicd_configs USING btree (deleted_at);


--
-- Name: idx_cicd_configs_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_cicd_configs_project_id ON public.cicd_configs USING btree (project_id);


--
-- Name: idx_cicd_configs_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_cicd_configs_workspace_id ON public.cicd_configs USING btree (workspace_id);


--
-- Name: idx_comments_agent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_comments_agent_id ON public.comments USING btree (agent_id);


--
-- Name: idx_comments_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON public.comments USING btree (deleted_at);


--
-- Name: idx_comments_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_comments_issue_id ON public.comments USING btree (issue_id);


--
-- Name: idx_conditional_fields_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_conditional_fields_deleted_at ON public.conditional_fields USING btree (deleted_at);


--
-- Name: idx_conditional_fields_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_conditional_fields_field_id ON public.conditional_fields USING btree (field_id);


--
-- Name: idx_conditional_fields_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_conditional_fields_workspace_id ON public.conditional_fields USING btree (workspace_id);


--
-- Name: idx_custom_field_options_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_field_options_deleted_at ON public.custom_field_options USING btree (deleted_at);


--
-- Name: idx_custom_field_options_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_field_options_field_id ON public.custom_field_options USING btree (field_id);


--
-- Name: idx_custom_fields_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_fields_deleted_at ON public.custom_fields USING btree (deleted_at);


--
-- Name: idx_custom_fields_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_fields_project_id ON public.custom_fields USING btree (project_id);


--
-- Name: idx_custom_fields_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_fields_workspace_id ON public.custom_fields USING btree (workspace_id);


--
-- Name: idx_cycles_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_cycles_deleted_at ON public.cycles USING btree (deleted_at);


--
-- Name: idx_cycles_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_cycles_project_id ON public.cycles USING btree (project_id);


--
-- Name: idx_dashboard_widgets_dashboard_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_dashboard_id ON public.dashboard_widgets USING btree (dashboard_id);


--
-- Name: idx_dashboard_widgets_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_deleted_at ON public.dashboard_widgets USING btree (deleted_at);


--
-- Name: idx_developer_jobs_agent_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_developer_jobs_agent_task_id ON public.developer_jobs USING btree (agent_task_id);


--
-- Name: idx_developer_jobs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_developer_jobs_deleted_at ON public.developer_jobs USING btree (deleted_at);


--
-- Name: idx_developer_jobs_git_connection_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_developer_jobs_git_connection_id ON public.developer_jobs USING btree (git_connection_id);


--
-- Name: idx_developer_jobs_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_developer_jobs_issue_id ON public.developer_jobs USING btree (issue_id);


--
-- Name: idx_developer_jobs_pr_number; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_developer_jobs_pr_number ON public.developer_jobs USING btree (pr_number);


--
-- Name: idx_developer_jobs_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_developer_jobs_project_id ON public.developer_jobs USING btree (project_id);


--
-- Name: idx_developer_jobs_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_developer_jobs_status ON public.developer_jobs USING btree (status);


--
-- Name: idx_developer_jobs_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_developer_jobs_workspace_id ON public.developer_jobs USING btree (workspace_id);


--
-- Name: idx_estimate_categories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_estimate_categories_deleted_at ON public.estimate_categories USING btree (deleted_at);


--
-- Name: idx_estimate_categories_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_estimate_categories_project_id ON public.estimate_categories USING btree (project_id);


--
-- Name: idx_estimate_categories_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_estimate_categories_workspace_id ON public.estimate_categories USING btree (workspace_id);


--
-- Name: idx_estimate_points_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_estimate_points_deleted_at ON public.estimate_points USING btree (deleted_at);


--
-- Name: idx_estimate_points_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_estimate_points_project_id ON public.estimate_points USING btree (project_id);


--
-- Name: idx_estimate_points_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_estimate_points_workspace_id ON public.estimate_points USING btree (workspace_id);


--
-- Name: idx_estimate_times_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_estimate_times_deleted_at ON public.estimate_times USING btree (deleted_at);


--
-- Name: idx_estimate_times_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_estimate_times_project_id ON public.estimate_times USING btree (project_id);


--
-- Name: idx_estimate_times_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_estimate_times_workspace_id ON public.estimate_times USING btree (workspace_id);


--
-- Name: idx_field_permissions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_field_permissions_deleted_at ON public.field_permissions USING btree (deleted_at);


--
-- Name: idx_field_permissions_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_field_permissions_project_id ON public.field_permissions USING btree (project_id);


--
-- Name: idx_field_permissions_resource; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_field_permissions_resource ON public.field_permissions USING btree (resource);


--
-- Name: idx_field_permissions_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_field_permissions_role_id ON public.field_permissions USING btree (role_id);


--
-- Name: idx_field_permissions_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_field_permissions_workspace_id ON public.field_permissions USING btree (workspace_id);


--
-- Name: idx_git_integrations_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_git_integrations_project_id ON public.git_integrations USING btree (project_id);


--
-- Name: idx_git_issue_links_integration_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_git_issue_links_integration_id ON public.git_issue_links USING btree (integration_id);


--
-- Name: idx_git_issue_links_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_git_issue_links_issue_id ON public.git_issue_links USING btree (issue_id);


--
-- Name: idx_github_connections_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_github_connections_deleted_at ON public.github_connections USING btree (deleted_at);


--
-- Name: idx_github_connections_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_github_connections_project_id ON public.github_connections USING btree (project_id);


--
-- Name: idx_github_connections_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_github_connections_workspace_id ON public.github_connections USING btree (workspace_id);


--
-- Name: idx_initiatives_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_initiatives_deleted_at ON public.initiatives USING btree (deleted_at);


--
-- Name: idx_initiatives_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_initiatives_workspace_id ON public.initiatives USING btree (workspace_id);


--
-- Name: idx_issue_activities_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_activities_deleted_at ON public.issue_activities USING btree (deleted_at);


--
-- Name: idx_issue_custom_field_values_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_custom_field_values_deleted_at ON public.issue_custom_field_values USING btree (deleted_at);


--
-- Name: idx_issue_custom_field_values_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_custom_field_values_field_id ON public.issue_custom_field_values USING btree (field_id);


--
-- Name: idx_issue_custom_field_values_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_custom_field_values_issue_id ON public.issue_custom_field_values USING btree (issue_id);


--
-- Name: idx_issue_relations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_relations_deleted_at ON public.issue_relations USING btree (deleted_at);


--
-- Name: idx_issue_relations_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_relations_issue_id ON public.issue_relations USING btree (issue_id);


--
-- Name: idx_issue_relations_related_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_relations_related_issue_id ON public.issue_relations USING btree (related_issue_id);


--
-- Name: idx_issue_type_imports_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_type_imports_deleted_at ON public.issue_type_imports USING btree (deleted_at);


--
-- Name: idx_issue_type_imports_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_type_imports_workspace_id ON public.issue_type_imports USING btree (workspace_id);


--
-- Name: idx_issue_type_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_type_templates_deleted_at ON public.issue_type_templates USING btree (deleted_at);


--
-- Name: idx_issue_type_templates_parent_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_type_templates_parent_type_id ON public.issue_type_templates USING btree (parent_type_id);


--
-- Name: idx_issue_type_templates_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_type_templates_workspace_id ON public.issue_type_templates USING btree (workspace_id);


--
-- Name: idx_issue_types_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_types_deleted_at ON public.issue_types USING btree (deleted_at);


--
-- Name: idx_issue_types_parent_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_types_parent_type_id ON public.issue_types USING btree (parent_type_id);


--
-- Name: idx_issue_types_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_types_project_id ON public.issue_types USING btree (project_id);


--
-- Name: idx_issue_types_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_types_workspace_id ON public.issue_types USING btree (workspace_id);


--
-- Name: idx_issue_watchers_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_watchers_issue_id ON public.issue_watchers USING btree (issue_id);


--
-- Name: idx_issue_watchers_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issue_watchers_user_id ON public.issue_watchers USING btree (user_id);


--
-- Name: idx_issues_active_approval_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issues_active_approval_id ON public.issues USING btree (active_approval_id);


--
-- Name: idx_issues_agent_assignee_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issues_agent_assignee_id ON public.issues USING btree (agent_assignee_id);


--
-- Name: idx_issues_agent_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issues_agent_task_id ON public.issues USING btree (agent_task_id);


--
-- Name: idx_issues_approval_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issues_approval_status ON public.issues USING btree (approval_status);


--
-- Name: idx_issues_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issues_deleted_at ON public.issues USING btree (deleted_at);


--
-- Name: idx_issues_intake_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issues_intake_status ON public.issues USING btree (intake_status);


--
-- Name: idx_issues_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issues_project_id ON public.issues USING btree (project_id);


--
-- Name: idx_issues_search; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issues_search ON public.issues USING gin (to_tsvector('english'::regconfig, (((COALESCE(name, ''::character varying))::text || ' '::text) || COALESCE(description_stripped, ''::text))));


--
-- Name: idx_issues_state_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_issues_state_id ON public.issues USING btree (state_id);


--
-- Name: idx_labels_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_labels_deleted_at ON public.labels USING btree (deleted_at);


--
-- Name: idx_labels_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_labels_project_id ON public.labels USING btree (project_id);


--
-- Name: idx_labels_project_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_project_name ON public.labels USING btree (project_id, name) WHERE (deleted_at IS NULL);


--
-- Name: idx_mcp_configs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_mcp_configs_deleted_at ON public.mcp_configs USING btree (deleted_at);


--
-- Name: idx_mcp_configs_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_mcp_configs_workspace_id ON public.mcp_configs USING btree (workspace_id);


--
-- Name: idx_memory_entries_agent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_memory_entries_agent_id ON public.memory_entries USING btree (agent_id);


--
-- Name: idx_memory_entries_context_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_memory_entries_context_key ON public.memory_entries USING btree (context_key);


--
-- Name: idx_memory_entries_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_memory_entries_deleted_at ON public.memory_entries USING btree (deleted_at);


--
-- Name: idx_memory_entries_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_memory_entries_issue_id ON public.memory_entries USING btree (issue_id);


--
-- Name: idx_memory_entries_memory_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_memory_entries_memory_type ON public.memory_entries USING btree (memory_type);


--
-- Name: idx_memory_entries_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_memory_entries_project_id ON public.memory_entries USING btree (project_id);


--
-- Name: idx_memory_entries_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_memory_entries_workspace_id ON public.memory_entries USING btree (workspace_id);


--
-- Name: idx_memory_sessions_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_memory_sessions_workspace_id ON public.memory_sessions USING btree (workspace_id);


--
-- Name: idx_memory_vector_indexes_memory_entry_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_memory_vector_indexes_memory_entry_id ON public.memory_vector_indexes USING btree (memory_entry_id);


--
-- Name: idx_memory_vector_indexes_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_memory_vector_indexes_workspace_id ON public.memory_vector_indexes USING btree (workspace_id);


--
-- Name: idx_message_reactions_message_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_message_reactions_message_id ON public.message_reactions USING btree (message_id);


--
-- Name: idx_messages_chat_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON public.messages USING btree (chat_id);


--
-- Name: idx_messages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_messages_deleted_at ON public.messages USING btree (deleted_at);


--
-- Name: idx_messages_reply_to_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_messages_reply_to_id ON public.messages USING btree (reply_to_id);


--
-- Name: idx_metric_charts_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_metric_charts_project_id ON public.metric_charts USING btree (project_id);


--
-- Name: idx_module_inheritance_overrides_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_module_inheritance_overrides_deleted_at ON public.module_inheritance_overrides USING btree (deleted_at);


--
-- Name: idx_module_inheritance_overrides_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_module_inheritance_overrides_project_id ON public.module_inheritance_overrides USING btree (project_id);


--
-- Name: idx_module_inheritance_overrides_workspace_module_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_module_inheritance_overrides_workspace_module_id ON public.module_inheritance_overrides USING btree (workspace_module_id);


--
-- Name: idx_modules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_modules_deleted_at ON public.modules USING btree (deleted_at);


--
-- Name: idx_modules_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_modules_parent_id ON public.modules USING btree (parent_id);


--
-- Name: idx_modules_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_modules_project_id ON public.modules USING btree (project_id);


--
-- Name: idx_modules_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_modules_workspace_id ON public.modules USING btree (workspace_id);


--
-- Name: idx_notifications_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at ON public.notifications USING btree (deleted_at);


--
-- Name: idx_notifications_is_read; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON public.notifications USING btree (is_read);


--
-- Name: idx_notifications_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_notifications_issue_id ON public.notifications USING btree (issue_id);


--
-- Name: idx_notifications_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_notifications_project_id ON public.notifications USING btree (project_id);


--
-- Name: idx_notifications_recipient_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_notifications_recipient_id ON public.notifications USING btree (recipient_id);


--
-- Name: idx_page_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_page_templates_deleted_at ON public.page_templates USING btree (deleted_at);


--
-- Name: idx_page_templates_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_page_templates_project_id ON public.page_templates USING btree (project_id);


--
-- Name: idx_page_templates_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_page_templates_workspace_id ON public.page_templates USING btree (workspace_id);


--
-- Name: idx_page_versions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_page_versions_deleted_at ON public.page_versions USING btree (deleted_at);


--
-- Name: idx_page_versions_page_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_page_versions_page_id ON public.page_versions USING btree (page_id);


--
-- Name: idx_pages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_pages_deleted_at ON public.pages USING btree (deleted_at);


--
-- Name: idx_pages_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_pages_parent_id ON public.pages USING btree (parent_id);


--
-- Name: idx_pages_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_pages_project_id ON public.pages USING btree (project_id);


--
-- Name: idx_pages_source_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_pages_source_id ON public.pages USING btree (source_id);


--
-- Name: idx_pages_source_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_pages_source_type ON public.pages USING btree (source_type);


--
-- Name: idx_permissions_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_code ON public.permissions USING btree (code);


--
-- Name: idx_permissions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_permissions_deleted_at ON public.permissions USING btree (deleted_at);


--
-- Name: idx_personal_access_tokens_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_personal_access_tokens_deleted_at ON public.personal_access_tokens USING btree (deleted_at);


--
-- Name: idx_personal_access_tokens_token_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_personal_access_tokens_token_hash ON public.personal_access_tokens USING btree (token_hash);


--
-- Name: idx_personal_access_tokens_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_personal_access_tokens_user_id ON public.personal_access_tokens USING btree (user_id);


--
-- Name: idx_plugin_event_logs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_plugin_event_logs_deleted_at ON public.plugin_event_logs USING btree (deleted_at);


--
-- Name: idx_plugin_event_logs_plugin_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_plugin_event_logs_plugin_id ON public.plugin_event_logs USING btree (plugin_id);


--
-- Name: idx_plugins_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_plugins_deleted_at ON public.plugins USING btree (deleted_at);


--
-- Name: idx_plugins_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_plugins_slug ON public.plugins USING btree (slug);


--
-- Name: idx_plugins_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_plugins_workspace_id ON public.plugins USING btree (workspace_id);


--
-- Name: idx_proj_agent_member; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_proj_agent_member ON public.project_agent_members USING btree (project_id, agent_id);


--
-- Name: idx_proj_member_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_proj_member_user ON public.project_members USING btree (project_id, user_id);


--
-- Name: idx_proj_sub_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_proj_sub_user ON public.project_subscribers USING btree (project_id, user_id);


--
-- Name: idx_project_agent_members_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_agent_members_deleted_at ON public.project_agent_members USING btree (deleted_at);


--
-- Name: idx_project_estimate_settings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_estimate_settings_deleted_at ON public.project_estimate_settings USING btree (deleted_at);


--
-- Name: idx_project_estimate_settings_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_estimate_settings_project_id ON public.project_estimate_settings USING btree (project_id);


--
-- Name: idx_project_estimate_settings_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_estimate_settings_workspace_id ON public.project_estimate_settings USING btree (workspace_id);


--
-- Name: idx_project_intake_settings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_intake_settings_deleted_at ON public.project_intake_settings USING btree (deleted_at);


--
-- Name: idx_project_intake_settings_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_project_intake_settings_project_id ON public.project_intake_settings USING btree (project_id);


--
-- Name: idx_project_intake_settings_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_project_intake_settings_token ON public.project_intake_settings USING btree (token);


--
-- Name: idx_project_members_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_members_deleted_at ON public.project_members USING btree (deleted_at);


--
-- Name: idx_project_page_tabs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_page_tabs_deleted_at ON public.project_page_tabs USING btree (deleted_at);


--
-- Name: idx_project_page_tabs_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_page_tabs_owner_id ON public.project_page_tabs USING btree (owner_id);


--
-- Name: idx_project_page_tabs_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_page_tabs_project_id ON public.project_page_tabs USING btree (project_id);


--
-- Name: idx_project_subscribers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_subscribers_deleted_at ON public.project_subscribers USING btree (deleted_at);


--
-- Name: idx_project_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_templates_deleted_at ON public.project_templates USING btree (deleted_at);


--
-- Name: idx_project_templates_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_templates_workspace_id ON public.project_templates USING btree (workspace_id);


--
-- Name: idx_project_updates_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_project_updates_project_id ON public.project_updates USING btree (project_id);


--
-- Name: idx_project_workspace_type; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_project_workspace_type ON public.issue_type_imports USING btree (project_id, workspace_type_id);


--
-- Name: idx_projects_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON public.projects USING btree (deleted_at);


--
-- Name: idx_projects_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_projects_workspace_id ON public.projects USING btree (workspace_id);


--
-- Name: idx_recurrence_rules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_recurrence_rules_deleted_at ON public.recurrence_rules USING btree (deleted_at);


--
-- Name: idx_recurrence_rules_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_recurrence_rules_issue_id ON public.recurrence_rules USING btree (issue_id);


--
-- Name: idx_relation_types_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_relation_types_deleted_at ON public.relation_types USING btree (deleted_at);


--
-- Name: idx_relation_types_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_relation_types_workspace_id ON public.relation_types USING btree (workspace_id);


--
-- Name: idx_releases_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_releases_deleted_at ON public.releases USING btree (deleted_at);


--
-- Name: idx_releases_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_releases_project_id ON public.releases USING btree (project_id);


--
-- Name: idx_role_permissions_permission_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON public.role_permissions USING btree (permission_id);


--
-- Name: idx_role_permissions_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON public.role_permissions USING btree (role_id);


--
-- Name: idx_roles_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON public.roles USING btree (deleted_at);


--
-- Name: idx_roles_level; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_roles_level ON public.roles USING btree (level);


--
-- Name: idx_roles_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_roles_project_id ON public.roles USING btree (project_id);


--
-- Name: idx_roles_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_roles_workspace_id ON public.roles USING btree (workspace_id);


--
-- Name: idx_runtimes_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_runtimes_deleted_at ON public.runtimes USING btree (deleted_at);


--
-- Name: idx_runtimes_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_runtimes_workspace_id ON public.runtimes USING btree (workspace_id);


--
-- Name: idx_saved_dashboards_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_saved_dashboards_deleted_at ON public.saved_dashboards USING btree (deleted_at);


--
-- Name: idx_saved_dashboards_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_saved_dashboards_owner_id ON public.saved_dashboards USING btree (owner_id);


--
-- Name: idx_saved_dashboards_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_saved_dashboards_project_id ON public.saved_dashboards USING btree (project_id);


--
-- Name: idx_saved_reports_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_saved_reports_deleted_at ON public.saved_reports USING btree (deleted_at);


--
-- Name: idx_saved_reports_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_saved_reports_project_id ON public.saved_reports USING btree (project_id);


--
-- Name: idx_saved_views_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_saved_views_deleted_at ON public.saved_views USING btree (deleted_at);


--
-- Name: idx_saved_views_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_saved_views_owner_id ON public.saved_views USING btree (owner_id);


--
-- Name: idx_saved_views_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_saved_views_project_id ON public.saved_views USING btree (project_id);


--
-- Name: idx_sdlc_stages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_sdlc_stages_deleted_at ON public.sdlc_stages USING btree (deleted_at);


--
-- Name: idx_sdlc_stages_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_sdlc_stages_status ON public.sdlc_stages USING btree (status);


--
-- Name: idx_sdlc_stages_workflow_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_sdlc_stages_workflow_id ON public.sdlc_stages USING btree (workflow_id);


--
-- Name: idx_sdlc_stages_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_sdlc_stages_workspace_id ON public.sdlc_stages USING btree (workspace_id);


--
-- Name: idx_sdlc_workflows_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_sdlc_workflows_deleted_at ON public.sdlc_workflows USING btree (deleted_at);


--
-- Name: idx_sdlc_workflows_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_sdlc_workflows_project_id ON public.sdlc_workflows USING btree (project_id);


--
-- Name: idx_sdlc_workflows_squad_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_sdlc_workflows_squad_id ON public.sdlc_workflows USING btree (squad_id);


--
-- Name: idx_sdlc_workflows_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_sdlc_workflows_status ON public.sdlc_workflows USING btree (status);


--
-- Name: idx_sdlc_workflows_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_sdlc_workflows_workspace_id ON public.sdlc_workflows USING btree (workspace_id);


--
-- Name: idx_search_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_search_templates_deleted_at ON public.search_templates USING btree (deleted_at);


--
-- Name: idx_search_templates_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_search_templates_owner_id ON public.search_templates USING btree (owner_id);


--
-- Name: idx_search_templates_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_search_templates_project_id ON public.search_templates USING btree (project_id);


--
-- Name: idx_skill_execution_logs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_skill_execution_logs_deleted_at ON public.skill_execution_logs USING btree (deleted_at);


--
-- Name: idx_skill_execution_logs_skill_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_skill_execution_logs_skill_id ON public.skill_execution_logs USING btree (skill_id);


--
-- Name: idx_skill_execution_logs_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_skill_execution_logs_workspace_id ON public.skill_execution_logs USING btree (workspace_id);


--
-- Name: idx_skills_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_skills_deleted_at ON public.skills USING btree (deleted_at);


--
-- Name: idx_skills_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_skills_workspace_id ON public.skills USING btree (workspace_id);


--
-- Name: idx_slack_connections_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_slack_connections_deleted_at ON public.slack_connections USING btree (deleted_at);


--
-- Name: idx_slack_connections_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_slack_connections_project_id ON public.slack_connections USING btree (project_id);


--
-- Name: idx_slack_connections_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_slack_connections_workspace_id ON public.slack_connections USING btree (workspace_id);


--
-- Name: idx_squad_executions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squad_executions_deleted_at ON public.squad_executions USING btree (deleted_at);


--
-- Name: idx_squad_executions_squad_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squad_executions_squad_id ON public.squad_executions USING btree (squad_id);


--
-- Name: idx_squad_members_agent_config_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squad_members_agent_config_id ON public.squad_members USING btree (agent_config_id);


--
-- Name: idx_squad_members_agent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squad_members_agent_id ON public.squad_members USING btree (agent_id);


--
-- Name: idx_squad_members_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squad_members_deleted_at ON public.squad_members USING btree (deleted_at);


--
-- Name: idx_squad_members_squad_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squad_members_squad_id ON public.squad_members USING btree (squad_id);


--
-- Name: idx_squad_tasks_agent_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squad_tasks_agent_task_id ON public.squad_tasks USING btree (agent_task_id);


--
-- Name: idx_squad_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squad_tasks_deleted_at ON public.squad_tasks USING btree (deleted_at);


--
-- Name: idx_squad_tasks_member_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squad_tasks_member_id ON public.squad_tasks USING btree (member_id);


--
-- Name: idx_squad_tasks_squad_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squad_tasks_squad_id ON public.squad_tasks USING btree (squad_id);


--
-- Name: idx_squads_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squads_deleted_at ON public.squads USING btree (deleted_at, deleted_at);


--
-- Name: idx_squads_leader_agent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squads_leader_agent_id ON public.squads USING btree (leader_agent_id);


--
-- Name: idx_squads_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squads_project_id ON public.squads USING btree (project_id);


--
-- Name: idx_squads_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_squads_workspace_id ON public.squads USING btree (workspace_id);


--
-- Name: idx_state_transitions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_state_transitions_deleted_at ON public.state_transitions USING btree (deleted_at);


--
-- Name: idx_state_transitions_workflow_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_state_transitions_workflow_id ON public.state_transitions USING btree (workflow_id);


--
-- Name: idx_state_transitions_workflow_kind; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_state_transitions_workflow_kind ON public.state_transitions USING btree (workflow_kind);


--
-- Name: idx_states_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_states_deleted_at ON public.states USING btree (deleted_at);


--
-- Name: idx_states_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_states_project_id ON public.states USING btree (project_id);


--
-- Name: idx_task_logs_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_task_logs_task_id ON public.task_logs USING btree (task_id);


--
-- Name: idx_tester_jobs_agent_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tester_jobs_agent_task_id ON public.tester_jobs USING btree (agent_task_id);


--
-- Name: idx_tester_jobs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tester_jobs_deleted_at ON public.tester_jobs USING btree (deleted_at);


--
-- Name: idx_tester_jobs_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tester_jobs_issue_id ON public.tester_jobs USING btree (issue_id);


--
-- Name: idx_tester_jobs_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tester_jobs_project_id ON public.tester_jobs USING btree (project_id);


--
-- Name: idx_tester_jobs_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tester_jobs_status ON public.tester_jobs USING btree (status);


--
-- Name: idx_tester_jobs_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tester_jobs_workspace_id ON public.tester_jobs USING btree (workspace_id);


--
-- Name: idx_time_tracks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_time_tracks_deleted_at ON public.time_tracks USING btree (deleted_at);


--
-- Name: idx_time_tracks_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_time_tracks_issue_id ON public.time_tracks USING btree (issue_id);


--
-- Name: idx_time_tracks_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_time_tracks_user_id ON public.time_tracks USING btree (user_id);


--
-- Name: idx_tool_call_logs_agent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tool_call_logs_agent_id ON public.tool_call_logs USING btree (agent_id);


--
-- Name: idx_tool_call_logs_agent_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tool_call_logs_agent_task_id ON public.tool_call_logs USING btree (agent_task_id);


--
-- Name: idx_tool_call_logs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tool_call_logs_deleted_at ON public.tool_call_logs USING btree (deleted_at);


--
-- Name: idx_tool_call_logs_tool_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tool_call_logs_tool_id ON public.tool_call_logs USING btree (tool_id);


--
-- Name: idx_tool_call_logs_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tool_call_logs_workspace_id ON public.tool_call_logs USING btree (workspace_id);


--
-- Name: idx_tool_permissions_agent_template_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tool_permissions_agent_template_id ON public.tool_permissions USING btree (agent_template_id);


--
-- Name: idx_tool_permissions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tool_permissions_deleted_at ON public.tool_permissions USING btree (deleted_at);


--
-- Name: idx_tool_permissions_tool_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tool_permissions_tool_id ON public.tool_permissions USING btree (tool_id);


--
-- Name: idx_tool_permissions_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tool_permissions_workspace_id ON public.tool_permissions USING btree (workspace_id);


--
-- Name: idx_tools_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tools_deleted_at ON public.tools USING btree (deleted_at);


--
-- Name: idx_tools_mcp_config_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tools_mcp_config_id ON public.tools USING btree (mcp_config_id);


--
-- Name: idx_tools_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_tools_name ON public.tools USING btree (name);


--
-- Name: idx_tools_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tools_workspace_id ON public.tools USING btree (workspace_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON public.users USING btree (username);


--
-- Name: idx_webhooks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_webhooks_deleted_at ON public.webhooks USING btree (deleted_at);


--
-- Name: idx_webhooks_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_webhooks_project_id ON public.webhooks USING btree (project_id);


--
-- Name: idx_work_item_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_work_item_templates_deleted_at ON public.work_item_templates USING btree (deleted_at);


--
-- Name: idx_work_item_templates_issue_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_work_item_templates_issue_type_id ON public.work_item_templates USING btree (issue_type_id);


--
-- Name: idx_work_item_templates_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_work_item_templates_project_id ON public.work_item_templates USING btree (project_id);


--
-- Name: idx_workflow_edges_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_edges_deleted_at ON public.workflow_edges USING btree (deleted_at);


--
-- Name: idx_workflow_edges_workflow_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_edges_workflow_id ON public.workflow_edges USING btree (workflow_id);


--
-- Name: idx_workflow_node_runs_agent_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_node_runs_agent_task_id ON public.workflow_node_runs USING btree (agent_task_id);


--
-- Name: idx_workflow_node_runs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_node_runs_deleted_at ON public.workflow_node_runs USING btree (deleted_at);


--
-- Name: idx_workflow_node_runs_workflow_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_node_runs_workflow_run_id ON public.workflow_node_runs USING btree (workflow_run_id);


--
-- Name: idx_workflow_nodes_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_nodes_deleted_at ON public.workflow_nodes USING btree (deleted_at);


--
-- Name: idx_workflow_nodes_workflow_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_nodes_workflow_id ON public.workflow_nodes USING btree (workflow_id);


--
-- Name: idx_workflow_runs_current_node; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_runs_current_node ON public.workflow_runs USING btree (current_node);


--
-- Name: idx_workflow_runs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_runs_deleted_at ON public.workflow_runs USING btree (deleted_at);


--
-- Name: idx_workflow_runs_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_runs_issue_id ON public.workflow_runs USING btree (issue_id);


--
-- Name: idx_workflow_runs_workflow_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_id ON public.workflow_runs USING btree (workflow_id);


--
-- Name: idx_workflows_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflows_deleted_at ON public.workflows USING btree (deleted_at);


--
-- Name: idx_workflows_issue_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflows_issue_type_id ON public.workflows USING btree (issue_type_id);


--
-- Name: idx_workflows_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflows_project_id ON public.workflows USING btree (project_id);


--
-- Name: idx_workflows_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workflows_workspace_id ON public.workflows USING btree (workspace_id);


--
-- Name: idx_workspace_members_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workspace_members_deleted_at ON public.workspace_members USING btree (deleted_at);


--
-- Name: idx_workspaces_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workspaces_deleted_at ON public.workspaces USING btree (deleted_at);


--
-- Name: idx_workspaces_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_workspaces_slug ON public.workspaces USING btree (slug);


--
-- Name: idx_ws_member_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_ws_member_user ON public.workspace_members USING btree (workspace_id, user_id);


--
-- Name: agent_activities fk_agent_activities_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_activities
    ADD CONSTRAINT fk_agent_activities_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_configs fk_agent_configs_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_configs
    ADD CONSTRAINT fk_agent_configs_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_decisions fk_agent_decisions_agent; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_decisions
    ADD CONSTRAINT fk_agent_decisions_agent FOREIGN KEY (agent_id) REFERENCES public.agents(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_task_snapshots fk_agent_task_snapshots_agent; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_task_snapshots
    ADD CONSTRAINT fk_agent_task_snapshots_agent FOREIGN KEY (agent_id) REFERENCES public.agents(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_task_snapshots fk_agent_task_snapshots_task; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_task_snapshots
    ADD CONSTRAINT fk_agent_task_snapshots_task FOREIGN KEY (task_id) REFERENCES public.agent_tasks(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_tasks fk_agent_tasks_agent_config; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_tasks
    ADD CONSTRAINT fk_agent_tasks_agent_config FOREIGN KEY (agent_config_id) REFERENCES public.agent_configs(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_tasks fk_agent_tasks_agent_template; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_tasks
    ADD CONSTRAINT fk_agent_tasks_agent_template FOREIGN KEY (agent_template_id) REFERENCES public.agent_templates(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_tasks fk_agent_tasks_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_tasks
    ADD CONSTRAINT fk_agent_tasks_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_tasks fk_agent_tasks_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_tasks
    ADD CONSTRAINT fk_agent_tasks_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_tasks fk_agent_tasks_runtime; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_tasks
    ADD CONSTRAINT fk_agent_tasks_runtime FOREIGN KEY (runtime_id) REFERENCES public.runtimes(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_tasks fk_agent_tasks_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_tasks
    ADD CONSTRAINT fk_agent_tasks_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_templates fk_agent_templates_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_templates
    ADD CONSTRAINT fk_agent_templates_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_edges fk_agent_workflows_edges; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_edges
    ADD CONSTRAINT fk_agent_workflows_edges FOREIGN KEY (workflow_id) REFERENCES public.agent_workflows(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_nodes fk_agent_workflows_nodes; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_nodes
    ADD CONSTRAINT fk_agent_workflows_nodes FOREIGN KEY (workflow_id) REFERENCES public.agent_workflows(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agent_activities fk_agents_activities; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agent_activities
    ADD CONSTRAINT fk_agents_activities FOREIGN KEY (agent_id) REFERENCES public.agents(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agents fk_agents_template; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agents
    ADD CONSTRAINT fk_agents_template FOREIGN KEY (template_id) REFERENCES public.agent_templates(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: agents fk_agents_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.agents
    ADD CONSTRAINT fk_agents_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: ai_messages fk_ai_threads_messages; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.ai_messages
    ADD CONSTRAINT fk_ai_threads_messages FOREIGN KEY (thread_id) REFERENCES public.ai_threads(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: build_records fk_build_records_cicd_config; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.build_records
    ADD CONSTRAINT fk_build_records_cicd_config FOREIGN KEY (cicd_config_id) REFERENCES public.cicd_configs(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: build_records fk_build_records_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.build_records
    ADD CONSTRAINT fk_build_records_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: build_records fk_build_records_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.build_records
    ADD CONSTRAINT fk_build_records_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: chats fk_chats_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.chats
    ADD CONSTRAINT fk_chats_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: messages fk_chats_messages; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.messages
    ADD CONSTRAINT fk_chats_messages FOREIGN KEY (chat_id) REFERENCES public.chats(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: cicd_configs fk_cicd_configs_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.cicd_configs
    ADD CONSTRAINT fk_cicd_configs_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: cicd_configs fk_cicd_configs_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.cicd_configs
    ADD CONSTRAINT fk_cicd_configs_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: comments fk_comments_agent; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.comments
    ADD CONSTRAINT fk_comments_agent FOREIGN KEY (agent_id) REFERENCES public.agents(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: comments fk_comments_author; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.comments
    ADD CONSTRAINT fk_comments_author FOREIGN KEY (author_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: comments fk_comments_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.comments
    ADD CONSTRAINT fk_comments_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: custom_field_options fk_custom_fields_options; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.custom_field_options
    ADD CONSTRAINT fk_custom_fields_options FOREIGN KEY (field_id) REFERENCES public.custom_fields(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_fields fk_custom_fields_type_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_fields
    ADD CONSTRAINT fk_custom_fields_type_links FOREIGN KEY (field_id) REFERENCES public.custom_fields(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: custom_fields fk_custom_fields_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.custom_fields
    ADD CONSTRAINT fk_custom_fields_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_cycles fk_cycles_issue_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_cycles
    ADD CONSTRAINT fk_cycles_issue_links FOREIGN KEY (cycle_id) REFERENCES public.cycles(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: developer_jobs fk_developer_jobs_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.developer_jobs
    ADD CONSTRAINT fk_developer_jobs_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: developer_jobs fk_developer_jobs_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.developer_jobs
    ADD CONSTRAINT fk_developer_jobs_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: developer_jobs fk_developer_jobs_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.developer_jobs
    ADD CONSTRAINT fk_developer_jobs_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: github_connections fk_github_connections_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.github_connections
    ADD CONSTRAINT fk_github_connections_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: github_connections fk_github_connections_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.github_connections
    ADD CONSTRAINT fk_github_connections_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_activities fk_issue_activities_actor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_activities
    ADD CONSTRAINT fk_issue_activities_actor FOREIGN KEY (actor_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_custom_field_values fk_issue_custom_field_values_field; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_custom_field_values
    ADD CONSTRAINT fk_issue_custom_field_values_field FOREIGN KEY (field_id) REFERENCES public.custom_fields(id) ON DELETE CASCADE;
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_custom_field_values fk_issue_custom_field_values_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_custom_field_values
    ADD CONSTRAINT fk_issue_custom_field_values_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_pages fk_issue_pages_page; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_pages
    ADD CONSTRAINT fk_issue_pages_page FOREIGN KEY (page_id) REFERENCES public.pages(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_relations fk_issue_relations_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_relations
    ADD CONSTRAINT fk_issue_relations_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_relations fk_issue_relations_related_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_relations
    ADD CONSTRAINT fk_issue_relations_related_issue FOREIGN KEY (related_issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_relations fk_issue_relations_relation_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_relations
    ADD CONSTRAINT fk_issue_relations_relation_type FOREIGN KEY (relation_type_id) REFERENCES public.relation_types(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_imports fk_issue_type_imports_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_imports
    ADD CONSTRAINT fk_issue_type_imports_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_imports fk_issue_type_imports_workspace_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_imports
    ADD CONSTRAINT fk_issue_type_imports_workspace_type FOREIGN KEY (workspace_type_id) REFERENCES public.issue_types(id) ON DELETE CASCADE;
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_template_fields fk_issue_type_template_fields_field; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_template_fields
    ADD CONSTRAINT fk_issue_type_template_fields_field FOREIGN KEY (field_id) REFERENCES public.custom_fields(id) ON DELETE CASCADE;
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_templates fk_issue_type_templates_child_types; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_templates
    ADD CONSTRAINT fk_issue_type_templates_child_types FOREIGN KEY (parent_type_id) REFERENCES public.issue_type_templates(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_template_fields fk_issue_type_templates_field_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_template_fields
    ADD CONSTRAINT fk_issue_type_templates_field_links FOREIGN KEY (template_type_id) REFERENCES public.issue_type_templates(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_types fk_issue_types_child_types; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_types
    ADD CONSTRAINT fk_issue_types_child_types FOREIGN KEY (parent_type_id) REFERENCES public.issue_types(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_type_fields fk_issue_types_field_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_type_fields
    ADD CONSTRAINT fk_issue_types_field_links FOREIGN KEY (type_id) REFERENCES public.issue_types(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_types fk_issue_types_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_types
    ADD CONSTRAINT fk_issue_types_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_activities fk_issues_activities; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_activities
    ADD CONSTRAINT fk_issues_activities FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_assignees fk_issues_assignee_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_assignees
    ADD CONSTRAINT fk_issues_assignee_links FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_cycles fk_issues_cycle_link; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_cycles
    ADD CONSTRAINT fk_issues_cycle_link FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issues fk_issues_issue_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issues
    ADD CONSTRAINT fk_issues_issue_type FOREIGN KEY (issue_type_id) REFERENCES public.issue_types(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_labels fk_issues_label_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_labels
    ADD CONSTRAINT fk_issues_label_links FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: module_issues fk_issues_module_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.module_issues
    ADD CONSTRAINT fk_issues_module_links FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_pages fk_issues_page_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_pages
    ADD CONSTRAINT fk_issues_page_links FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issues fk_issues_sub_issues; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issues
    ADD CONSTRAINT fk_issues_sub_issues FOREIGN KEY (parent_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_labels fk_labels_issue_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_labels
    ADD CONSTRAINT fk_labels_issue_links FOREIGN KEY (label_id) REFERENCES public.labels(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: mcp_configs fk_mcp_configs_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.mcp_configs
    ADD CONSTRAINT fk_mcp_configs_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: memory_entries fk_memory_entries_agent; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.memory_entries
    ADD CONSTRAINT fk_memory_entries_agent FOREIGN KEY (agent_id) REFERENCES public.agents(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: memory_entries fk_memory_entries_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.memory_entries
    ADD CONSTRAINT fk_memory_entries_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: memory_entries fk_memory_entries_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.memory_entries
    ADD CONSTRAINT fk_memory_entries_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: memory_entries fk_memory_entries_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.memory_entries
    ADD CONSTRAINT fk_memory_entries_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: messages fk_messages_chat; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.messages
    ADD CONSTRAINT fk_messages_chat FOREIGN KEY (reply_to_id) REFERENCES public.messages(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: message_reactions fk_messages_reactions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.message_reactions
    ADD CONSTRAINT fk_messages_reactions FOREIGN KEY (message_id) REFERENCES public.messages(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: module_issues fk_modules_issue_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.module_issues
    ADD CONSTRAINT fk_modules_issue_links FOREIGN KEY (module_id) REFERENCES public.modules(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: modules fk_modules_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.modules
    ADD CONSTRAINT fk_modules_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: notifications fk_notifications_recipient; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notifications_recipient FOREIGN KEY (recipient_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: notifications fk_notifications_sender; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notifications_sender FOREIGN KEY (sender_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: page_versions fk_page_versions_page; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.page_versions
    ADD CONSTRAINT fk_page_versions_page FOREIGN KEY (page_id) REFERENCES public.pages(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: pages fk_pages_children; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.pages
    ADD CONSTRAINT fk_pages_children FOREIGN KEY (parent_id) REFERENCES public.pages(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: pages fk_pages_locked_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.pages
    ADD CONSTRAINT fk_pages_locked_by FOREIGN KEY (locked_by_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: pages fk_pages_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.pages
    ADD CONSTRAINT fk_pages_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: plugin_event_logs fk_plugin_event_logs_plugin; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.plugin_event_logs
    ADD CONSTRAINT fk_plugin_event_logs_plugin FOREIGN KEY (plugin_id) REFERENCES public.plugins(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: plugins fk_plugins_installed_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.plugins
    ADD CONSTRAINT fk_plugins_installed_by FOREIGN KEY (installed_by_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: plugins fk_plugins_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.plugins
    ADD CONSTRAINT fk_plugins_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_agent_members fk_project_agent_members_agent; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_agent_members
    ADD CONSTRAINT fk_project_agent_members_agent FOREIGN KEY (agent_id) REFERENCES public.agents(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_agent_members fk_project_agent_members_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_agent_members
    ADD CONSTRAINT fk_project_agent_members_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_custom_field_enrollments fk_project_custom_field_enrollments_custom_field; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_custom_field_enrollments
    ADD CONSTRAINT fk_project_custom_field_enrollments_custom_field FOREIGN KEY (field_id) REFERENCES public.custom_fields(id) ON DELETE CASCADE;
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_custom_field_enrollments fk_project_custom_field_enrollments_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_custom_field_enrollments
    ADD CONSTRAINT fk_project_custom_field_enrollments_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_page_tabs fk_project_page_tabs_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_page_tabs
    ADD CONSTRAINT fk_project_page_tabs_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_subscribers fk_project_subscribers_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_subscribers
    ADD CONSTRAINT fk_project_subscribers_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_subscribers fk_project_subscribers_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_subscribers
    ADD CONSTRAINT fk_project_subscribers_user FOREIGN KEY (user_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_template_types fk_project_template_types_type_template; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_template_types
    ADD CONSTRAINT fk_project_template_types_type_template FOREIGN KEY (type_template_id) REFERENCES public.issue_type_templates(id) ON DELETE CASCADE;
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_template_types fk_project_templates_type_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_template_types
    ADD CONSTRAINT fk_project_templates_type_links FOREIGN KEY (template_id) REFERENCES public.project_templates(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_updates fk_project_updates_author; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_updates
    ADD CONSTRAINT fk_project_updates_author FOREIGN KEY (author_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: cycles fk_projects_cycles; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.cycles
    ADD CONSTRAINT fk_projects_cycles FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: projects fk_projects_default_assignee; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_default_assignee FOREIGN KEY (default_assignee_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issues fk_projects_issues; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issues
    ADD CONSTRAINT fk_projects_issues FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: labels fk_projects_labels; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.labels
    ADD CONSTRAINT fk_projects_labels FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_members fk_projects_members; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT fk_projects_members FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: projects fk_projects_project_lead; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_project_lead FOREIGN KEY (project_lead_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: states fk_projects_states; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.states
    ADD CONSTRAINT fk_projects_states FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: recurrence_rules fk_recurrence_rules_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.recurrence_rules
    ADD CONSTRAINT fk_recurrence_rules_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: runtimes fk_runtimes_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.runtimes
    ADD CONSTRAINT fk_runtimes_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: saved_dashboards fk_saved_dashboards_owner; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.saved_dashboards
    ADD CONSTRAINT fk_saved_dashboards_owner FOREIGN KEY (owner_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: saved_dashboards fk_saved_dashboards_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.saved_dashboards
    ADD CONSTRAINT fk_saved_dashboards_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: dashboard_widgets fk_saved_dashboards_widgets; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.dashboard_widgets
    ADD CONSTRAINT fk_saved_dashboards_widgets FOREIGN KEY (dashboard_id) REFERENCES public.saved_dashboards(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: saved_reports fk_saved_reports_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.saved_reports
    ADD CONSTRAINT fk_saved_reports_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: saved_views fk_saved_views_owner; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.saved_views
    ADD CONSTRAINT fk_saved_views_owner FOREIGN KEY (owner_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: saved_views fk_saved_views_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.saved_views
    ADD CONSTRAINT fk_saved_views_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: sdlc_workflows fk_sdlc_workflows_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.sdlc_workflows
    ADD CONSTRAINT fk_sdlc_workflows_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: sdlc_stages fk_sdlc_workflows_stages; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.sdlc_stages
    ADD CONSTRAINT fk_sdlc_workflows_stages FOREIGN KEY (workflow_id) REFERENCES public.sdlc_workflows(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: sdlc_workflows fk_sdlc_workflows_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.sdlc_workflows
    ADD CONSTRAINT fk_sdlc_workflows_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: search_templates fk_search_templates_owner; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.search_templates
    ADD CONSTRAINT fk_search_templates_owner FOREIGN KEY (owner_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: search_templates fk_search_templates_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.search_templates
    ADD CONSTRAINT fk_search_templates_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: skill_execution_logs fk_skill_execution_logs_skill; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.skill_execution_logs
    ADD CONSTRAINT fk_skill_execution_logs_skill FOREIGN KEY (skill_id) REFERENCES public.skills(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: skill_execution_logs fk_skill_execution_logs_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.skill_execution_logs
    ADD CONSTRAINT fk_skill_execution_logs_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: skills fk_skills_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.skills
    ADD CONSTRAINT fk_skills_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: slack_connections fk_slack_connections_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.slack_connections
    ADD CONSTRAINT fk_slack_connections_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: slack_connections fk_slack_connections_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.slack_connections
    ADD CONSTRAINT fk_slack_connections_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: squad_members fk_squads_members; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.squad_members
    ADD CONSTRAINT fk_squads_members FOREIGN KEY (squad_id) REFERENCES public.squads(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issues fk_states_issues; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issues
    ADD CONSTRAINT fk_states_issues FOREIGN KEY (state_id) REFERENCES public.states(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: state_transitions fk_states_source_transitions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.state_transitions
    ADD CONSTRAINT fk_states_source_transitions FOREIGN KEY (source_state_id) REFERENCES public.states(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: state_transitions fk_states_target_transitions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.state_transitions
    ADD CONSTRAINT fk_states_target_transitions FOREIGN KEY (target_state_id) REFERENCES public.states(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: task_logs fk_task_logs_task; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.task_logs
    ADD CONSTRAINT fk_task_logs_task FOREIGN KEY (task_id) REFERENCES public.agent_tasks(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tester_jobs fk_tester_jobs_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tester_jobs
    ADD CONSTRAINT fk_tester_jobs_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tester_jobs fk_tester_jobs_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tester_jobs
    ADD CONSTRAINT fk_tester_jobs_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tester_jobs fk_tester_jobs_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tester_jobs
    ADD CONSTRAINT fk_tester_jobs_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: time_tracks fk_time_tracks_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.time_tracks
    ADD CONSTRAINT fk_time_tracks_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: time_tracks fk_time_tracks_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.time_tracks
    ADD CONSTRAINT fk_time_tracks_user FOREIGN KEY (user_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tool_call_logs fk_tool_call_logs_tool; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tool_call_logs
    ADD CONSTRAINT fk_tool_call_logs_tool FOREIGN KEY (tool_id) REFERENCES public.tools(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tool_permissions fk_tool_permissions_tool; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tool_permissions
    ADD CONSTRAINT fk_tool_permissions_tool FOREIGN KEY (tool_id) REFERENCES public.tools(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tools fk_tools_mcp_config; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tools
    ADD CONSTRAINT fk_tools_mcp_config FOREIGN KEY (mcp_config_id) REFERENCES public.mcp_configs(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: tools fk_tools_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.tools
    ADD CONSTRAINT fk_tools_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: issue_assignees fk_users_assigned_issues; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.issue_assignees
    ADD CONSTRAINT fk_users_assigned_issues FOREIGN KEY (user_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: project_members fk_users_projects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT fk_users_projects FOREIGN KEY (user_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workspace_members fk_users_workspaces; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT fk_users_workspaces FOREIGN KEY (user_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_edges fk_workflow_edges_source_node; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_edges
    ADD CONSTRAINT fk_workflow_edges_source_node FOREIGN KEY (source_node_id) REFERENCES public.workflow_nodes(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_edges fk_workflow_edges_target_node; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_edges
    ADD CONSTRAINT fk_workflow_edges_target_node FOREIGN KEY (target_node_id) REFERENCES public.workflow_nodes(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_node_runs fk_workflow_node_runs_agent_task; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_node_runs
    ADD CONSTRAINT fk_workflow_node_runs_agent_task FOREIGN KEY (agent_task_id) REFERENCES public.agent_tasks(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_node_runs fk_workflow_node_runs_node; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_node_runs
    ADD CONSTRAINT fk_workflow_node_runs_node FOREIGN KEY (node_id) REFERENCES public.workflow_nodes(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_runs fk_workflow_runs_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_runs
    ADD CONSTRAINT fk_workflow_runs_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_node_runs fk_workflow_runs_node_runs; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_node_runs
    ADD CONSTRAINT fk_workflow_runs_node_runs FOREIGN KEY (workflow_run_id) REFERENCES public.workflow_runs(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflow_runs fk_workflow_runs_workflow; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflow_runs
    ADD CONSTRAINT fk_workflow_runs_workflow FOREIGN KEY (workflow_id) REFERENCES public.agent_workflows(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workflows fk_workflows_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT fk_workflows_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workspace_members fk_workspaces_members; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT fk_workspaces_members FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: workspaces fk_workspaces_owner; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT fk_workspaces_owner FOREIGN KEY (owner_id) REFERENCES public.users(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- Name: projects fk_workspaces_projects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

DO $baseline$ BEGIN
  ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_workspaces_projects FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);
EXCEPTION
  WHEN duplicate_object OR duplicate_table OR invalid_table_definition THEN NULL;
END $baseline$;


--
-- PostgreSQL database dump complete
--


