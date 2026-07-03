--
-- PostgreSQL database dump
--

\restrict dQIpgiUggGs6y8yAccH6qYAWe7lVbdlon8SC5dMhWyfgQe0K6c3PVZYM3VTK2S3

-- Dumped from database version 18.4
-- Dumped by pg_dump version 18.4

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: agent_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.agent_activities (
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
    executed_at timestamp with time zone,
    agent_name character varying(128),
    task_context text
);


--
-- Name: agent_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.agent_activities_id_seq
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
-- Name: agents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.agents (
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
    system_prompt text
);


--
-- Name: agents_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.agents_id_seq
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

CREATE TABLE public.ai_configs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
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

CREATE SEQUENCE public.ai_configs_id_seq
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

CREATE TABLE public.ai_messages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    thread_id bigint NOT NULL,
    role character varying(20) NOT NULL,
    content text NOT NULL,
    tool_calls jsonb,
    tool_name character varying(50)
);


--
-- Name: ai_messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_messages_id_seq
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

CREATE TABLE public.ai_threads (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    title character varying(255),
    workspace_id bigint NOT NULL,
    project_id bigint,
    user_id bigint NOT NULL
);


--
-- Name: ai_threads_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_threads_id_seq
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
-- Name: attachments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attachments (
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

CREATE SEQUENCE public.attachments_id_seq
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
-- Name: automation_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.automation_rules (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description text,
    project_id bigint NOT NULL,
    is_enabled boolean DEFAULT true,
    sequence bigint DEFAULT 1,
    execution_count bigint DEFAULT 0,
    trigger_type character varying(50) NOT NULL,
    conditions text,
    actions text
);


--
-- Name: automation_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.automation_rules_id_seq
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
-- Name: comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.comments (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    issue_id bigint NOT NULL,
    author_id bigint,
    body text NOT NULL,
    is_resolved boolean DEFAULT false,
    parent_id bigint
);


--
-- Name: comments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.comments_id_seq
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

CREATE TABLE public.conditional_fields (
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

CREATE SEQUENCE public.conditional_fields_id_seq
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

CREATE TABLE public.custom_field_options (
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

CREATE SEQUENCE public.custom_field_options_id_seq
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

CREATE TABLE public.custom_fields (
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
    project_id bigint,
    workspace_id bigint NOT NULL
);


--
-- Name: custom_fields_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.custom_fields_id_seq
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

CREATE TABLE public.cycles (
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
    workspace_id bigint NOT NULL
);


--
-- Name: cycles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cycles_id_seq
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
-- Name: estimate_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.estimate_categories (
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

CREATE SEQUENCE public.estimate_categories_id_seq
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

CREATE TABLE public.estimate_points (
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

CREATE SEQUENCE public.estimate_points_id_seq
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

CREATE TABLE public.estimate_times (
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

CREATE SEQUENCE public.estimate_times_id_seq
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
-- Name: github_connections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.github_connections (
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

CREATE SEQUENCE public.github_connections_id_seq
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

CREATE TABLE public.initiative_projects (
    initiative_id bigint NOT NULL,
    project_id bigint NOT NULL
);


--
-- Name: initiatives; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.initiatives (
    id bigint NOT NULL,
    workspace_id bigint,
    name character varying(255) NOT NULL,
    description text,
    color character varying(20),
    status character varying(20) DEFAULT 'active'::character varying,
    target_date timestamp with time zone,
    start_date timestamp with time zone,
    sort_order bigint DEFAULT 0,
    created_by_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: initiatives_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.initiatives_id_seq
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

CREATE TABLE public.issue_activities (
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

CREATE SEQUENCE public.issue_activities_id_seq
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

CREATE TABLE public.issue_assignees (
    issue_id bigint NOT NULL,
    user_id bigint NOT NULL
);


--
-- Name: issue_custom_field_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.issue_custom_field_values (
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

CREATE SEQUENCE public.issue_custom_field_values_id_seq
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

CREATE TABLE public.issue_cycles (
    issue_id bigint NOT NULL,
    cycle_id bigint NOT NULL
);


--
-- Name: issue_labels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.issue_labels (
    issue_id bigint NOT NULL,
    label_id bigint NOT NULL
);


--
-- Name: issue_pages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.issue_pages (
    issue_id bigint NOT NULL,
    page_id bigint NOT NULL
);


--
-- Name: issue_relations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.issue_relations (
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

CREATE SEQUENCE public.issue_relations_id_seq
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

CREATE TABLE public.issue_type_fields (
    type_id bigint NOT NULL,
    field_id bigint NOT NULL,
    is_required boolean DEFAULT false,
    sequence bigint DEFAULT 1
);


--
-- Name: issue_type_template_fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.issue_type_template_fields (
    template_type_id bigint NOT NULL,
    field_id bigint NOT NULL,
    is_required boolean DEFAULT false,
    sequence bigint DEFAULT 1
);


--
-- Name: issue_type_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.issue_type_templates (
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

CREATE SEQUENCE public.issue_type_templates_id_seq
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

CREATE TABLE public.issue_types (
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

CREATE SEQUENCE public.issue_types_id_seq
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
-- Name: issues; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.issues (
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
    external_id character varying(255),
    external_source character varying(255),
    cover_image_url character varying(500),
    intake_source character varying(50),
    intake_status character varying(30)
);


--
-- Name: issues_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.issues_id_seq
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

CREATE TABLE public.labels (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    color character varying(50) DEFAULT '#6B7280'::character varying,
    description character varying(255),
    project_id bigint NOT NULL
);


--
-- Name: labels_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.labels_id_seq
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

CREATE TABLE public.mcp_configs (
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

CREATE SEQUENCE public.mcp_configs_id_seq
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
-- Name: module_issues; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.module_issues (
    module_id bigint NOT NULL,
    issue_id bigint NOT NULL
);


--
-- Name: modules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.modules (
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
    parent_id bigint,
    "order" bigint DEFAULT 0,
    archived_at timestamp with time zone,
    is_archived boolean DEFAULT false
);


--
-- Name: modules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.modules_id_seq
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

CREATE TABLE public.notifications (
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

CREATE SEQUENCE public.notifications_id_seq
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
-- Name: pages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pages (
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
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: pages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pages_id_seq
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

CREATE TABLE public.permissions (
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
    scope character varying(20) DEFAULT 'project'::character varying
);


--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permissions_id_seq
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
-- Name: project_estimate_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_estimate_settings (
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

CREATE SEQUENCE public.project_estimate_settings_id_seq
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
-- Name: project_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_members (
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

CREATE SEQUENCE public.project_members_id_seq
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

CREATE TABLE public.project_page_tabs (
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

CREATE SEQUENCE public.project_page_tabs_id_seq
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

CREATE TABLE public.project_subscribers (
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

CREATE SEQUENCE public.project_subscribers_id_seq
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

CREATE TABLE public.project_template_types (
    template_id bigint NOT NULL,
    type_template_id bigint NOT NULL,
    is_required boolean DEFAULT false,
    default_state_id bigint,
    sequence bigint DEFAULT 1
);


--
-- Name: project_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_templates (
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

CREATE SEQUENCE public.project_templates_id_seq
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

CREATE TABLE public.project_updates (
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

CREATE SEQUENCE public.project_updates_id_seq
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

CREATE TABLE public.projects (
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

CREATE SEQUENCE public.projects_id_seq
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

CREATE TABLE public.recurrence_rules (
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

CREATE SEQUENCE public.recurrence_rules_id_seq
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

CREATE TABLE public.relation_types (
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

CREATE SEQUENCE public.relation_types_id_seq
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

CREATE TABLE public.release_issues (
    release_id bigint NOT NULL,
    issue_id bigint NOT NULL
);


--
-- Name: releases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.releases (
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

CREATE SEQUENCE public.releases_id_seq
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

CREATE TABLE public.role_permissions (
    role_id bigint NOT NULL,
    permission_id bigint NOT NULL
);


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
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

CREATE SEQUENCE public.roles_id_seq
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
-- Name: saved_reports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saved_reports (
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

CREATE SEQUENCE public.saved_reports_id_seq
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

CREATE TABLE public.saved_views (
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
    sort_config jsonb,
    columns jsonb,
    group_by character varying(50),
    is_default boolean DEFAULT false,
    is_shared boolean DEFAULT false,
    owner_id bigint NOT NULL,
    project_id bigint NOT NULL,
    rql text
);


--
-- Name: saved_views_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saved_views_id_seq
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
-- Name: search_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.search_templates (
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

CREATE SEQUENCE public.search_templates_id_seq
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
-- Name: slack_connections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slack_connections (
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

CREATE SEQUENCE public.slack_connections_id_seq
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
-- Name: state_transitions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.state_transitions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(255) NOT NULL,
    description character varying(500),
    workflow_id bigint NOT NULL,
    source_state_id bigint NOT NULL,
    target_state_id bigint NOT NULL,
    issue_type_id bigint,
    is_auto boolean DEFAULT false,
    rule_type text DEFAULT 'allow'::text,
    approver_ids text,
    role_allowed character varying(50),
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: state_transitions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.state_transitions_id_seq
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

CREATE TABLE public.states (
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
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: states_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.states_id_seq
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
-- Name: time_tracks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.time_tracks (
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

CREATE SEQUENCE public.time_tracks_id_seq
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
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
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

CREATE SEQUENCE public.users_id_seq
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

CREATE TABLE public.webhooks (
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

CREATE SEQUENCE public.webhooks_id_seq
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

CREATE TABLE public.work_item_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description character varying(500),
    issue_type_id bigint,
    defaults jsonb,
    is_default boolean DEFAULT false,
    project_id bigint NOT NULL,
    workspace_id bigint NOT NULL
);


--
-- Name: work_item_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.work_item_templates_id_seq
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
-- Name: workflows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.workflows (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    updated_by_id bigint,
    name character varying(100) NOT NULL,
    description text,
    project_id bigint NOT NULL,
    issue_type_id bigint,
    is_active boolean DEFAULT true
);


--
-- Name: workflows_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.workflows_id_seq
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

CREATE TABLE public.workspace_members (
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

CREATE SEQUENCE public.workspace_members_id_seq
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

CREATE TABLE public.workspaces (
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

CREATE SEQUENCE public.workspaces_id_seq
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
-- Name: attachments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attachments ALTER COLUMN id SET DEFAULT nextval('public.attachments_id_seq'::regclass);


--
-- Name: automation_rules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.automation_rules ALTER COLUMN id SET DEFAULT nextval('public.automation_rules_id_seq'::regclass);


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
-- Name: issue_type_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_templates ALTER COLUMN id SET DEFAULT nextval('public.issue_type_templates_id_seq'::regclass);


--
-- Name: issue_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_types ALTER COLUMN id SET DEFAULT nextval('public.issue_types_id_seq'::regclass);


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
-- Name: modules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modules ALTER COLUMN id SET DEFAULT nextval('public.modules_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: pages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pages ALTER COLUMN id SET DEFAULT nextval('public.pages_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: project_estimate_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_estimate_settings ALTER COLUMN id SET DEFAULT nextval('public.project_estimate_settings_id_seq'::regclass);


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
-- Name: saved_reports id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_reports ALTER COLUMN id SET DEFAULT nextval('public.saved_reports_id_seq'::regclass);


--
-- Name: saved_views id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_views ALTER COLUMN id SET DEFAULT nextval('public.saved_views_id_seq'::regclass);


--
-- Name: search_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.search_templates ALTER COLUMN id SET DEFAULT nextval('public.search_templates_id_seq'::regclass);


--
-- Name: slack_connections id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_connections ALTER COLUMN id SET DEFAULT nextval('public.slack_connections_id_seq'::regclass);


--
-- Name: state_transitions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_transitions ALTER COLUMN id SET DEFAULT nextval('public.state_transitions_id_seq'::regclass);


--
-- Name: states id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.states ALTER COLUMN id SET DEFAULT nextval('public.states_id_seq'::regclass);


--
-- Name: time_tracks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_tracks ALTER COLUMN id SET DEFAULT nextval('public.time_tracks_id_seq'::regclass);


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

ALTER TABLE ONLY public.agent_activities
    ADD CONSTRAINT agent_activities_pkey PRIMARY KEY (id);


--
-- Name: agents agents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_pkey PRIMARY KEY (id);


--
-- Name: ai_configs ai_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_configs
    ADD CONSTRAINT ai_configs_pkey PRIMARY KEY (id);


--
-- Name: ai_messages ai_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_messages
    ADD CONSTRAINT ai_messages_pkey PRIMARY KEY (id);


--
-- Name: ai_threads ai_threads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_threads
    ADD CONSTRAINT ai_threads_pkey PRIMARY KEY (id);


--
-- Name: attachments attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_pkey PRIMARY KEY (id);


--
-- Name: automation_rules automation_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.automation_rules
    ADD CONSTRAINT automation_rules_pkey PRIMARY KEY (id);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: conditional_fields conditional_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conditional_fields
    ADD CONSTRAINT conditional_fields_pkey PRIMARY KEY (id);


--
-- Name: custom_field_options custom_field_options_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.custom_field_options
    ADD CONSTRAINT custom_field_options_pkey PRIMARY KEY (id);


--
-- Name: custom_fields custom_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.custom_fields
    ADD CONSTRAINT custom_fields_pkey PRIMARY KEY (id);


--
-- Name: cycles cycles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cycles
    ADD CONSTRAINT cycles_pkey PRIMARY KEY (id);


--
-- Name: estimate_categories estimate_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.estimate_categories
    ADD CONSTRAINT estimate_categories_pkey PRIMARY KEY (id);


--
-- Name: estimate_points estimate_points_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.estimate_points
    ADD CONSTRAINT estimate_points_pkey PRIMARY KEY (id);


--
-- Name: estimate_times estimate_times_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.estimate_times
    ADD CONSTRAINT estimate_times_pkey PRIMARY KEY (id);


--
-- Name: github_connections github_connections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_connections
    ADD CONSTRAINT github_connections_pkey PRIMARY KEY (id);


--
-- Name: initiative_projects initiative_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.initiative_projects
    ADD CONSTRAINT initiative_projects_pkey PRIMARY KEY (initiative_id, project_id);


--
-- Name: initiatives initiatives_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.initiatives
    ADD CONSTRAINT initiatives_pkey PRIMARY KEY (id);


--
-- Name: issue_activities issue_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_activities
    ADD CONSTRAINT issue_activities_pkey PRIMARY KEY (id);


--
-- Name: issue_assignees issue_assignees_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_assignees
    ADD CONSTRAINT issue_assignees_pkey PRIMARY KEY (issue_id, user_id);


--
-- Name: issue_custom_field_values issue_custom_field_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_custom_field_values
    ADD CONSTRAINT issue_custom_field_values_pkey PRIMARY KEY (id);


--
-- Name: issue_cycles issue_cycles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_cycles
    ADD CONSTRAINT issue_cycles_pkey PRIMARY KEY (issue_id, cycle_id);


--
-- Name: issue_labels issue_labels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_labels
    ADD CONSTRAINT issue_labels_pkey PRIMARY KEY (issue_id, label_id);


--
-- Name: issue_pages issue_pages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_pages
    ADD CONSTRAINT issue_pages_pkey PRIMARY KEY (issue_id, page_id);


--
-- Name: issue_relations issue_relations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_relations
    ADD CONSTRAINT issue_relations_pkey PRIMARY KEY (id);


--
-- Name: issue_type_fields issue_type_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_fields
    ADD CONSTRAINT issue_type_fields_pkey PRIMARY KEY (type_id, field_id);


--
-- Name: issue_type_template_fields issue_type_template_fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_template_fields
    ADD CONSTRAINT issue_type_template_fields_pkey PRIMARY KEY (template_type_id, field_id);


--
-- Name: issue_type_templates issue_type_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_templates
    ADD CONSTRAINT issue_type_templates_pkey PRIMARY KEY (id);


--
-- Name: issue_types issue_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_types
    ADD CONSTRAINT issue_types_pkey PRIMARY KEY (id);


--
-- Name: issues issues_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issues
    ADD CONSTRAINT issues_pkey PRIMARY KEY (id);


--
-- Name: labels labels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labels
    ADD CONSTRAINT labels_pkey PRIMARY KEY (id);


--
-- Name: mcp_configs mcp_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mcp_configs
    ADD CONSTRAINT mcp_configs_pkey PRIMARY KEY (id);


--
-- Name: module_issues module_issues_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_issues
    ADD CONSTRAINT module_issues_pkey PRIMARY KEY (module_id, issue_id);


--
-- Name: modules modules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modules
    ADD CONSTRAINT modules_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: pages pages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: project_estimate_settings project_estimate_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_estimate_settings
    ADD CONSTRAINT project_estimate_settings_pkey PRIMARY KEY (id);


--
-- Name: project_members project_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_pkey PRIMARY KEY (id);


--
-- Name: project_page_tabs project_page_tabs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_page_tabs
    ADD CONSTRAINT project_page_tabs_pkey PRIMARY KEY (id);


--
-- Name: project_subscribers project_subscribers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_subscribers
    ADD CONSTRAINT project_subscribers_pkey PRIMARY KEY (id);


--
-- Name: project_template_types project_template_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_template_types
    ADD CONSTRAINT project_template_types_pkey PRIMARY KEY (template_id, type_template_id);


--
-- Name: project_templates project_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_templates
    ADD CONSTRAINT project_templates_pkey PRIMARY KEY (id);


--
-- Name: project_updates project_updates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_updates
    ADD CONSTRAINT project_updates_pkey PRIMARY KEY (id);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: recurrence_rules recurrence_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recurrence_rules
    ADD CONSTRAINT recurrence_rules_pkey PRIMARY KEY (id);


--
-- Name: relation_types relation_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.relation_types
    ADD CONSTRAINT relation_types_pkey PRIMARY KEY (id);


--
-- Name: release_issues release_issues_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.release_issues
    ADD CONSTRAINT release_issues_pkey PRIMARY KEY (release_id, issue_id);


--
-- Name: releases releases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.releases
    ADD CONSTRAINT releases_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: saved_reports saved_reports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_reports
    ADD CONSTRAINT saved_reports_pkey PRIMARY KEY (id);


--
-- Name: saved_views saved_views_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_views
    ADD CONSTRAINT saved_views_pkey PRIMARY KEY (id);


--
-- Name: search_templates search_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.search_templates
    ADD CONSTRAINT search_templates_pkey PRIMARY KEY (id);


--
-- Name: slack_connections slack_connections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_connections
    ADD CONSTRAINT slack_connections_pkey PRIMARY KEY (id);


--
-- Name: state_transitions state_transitions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_transitions
    ADD CONSTRAINT state_transitions_pkey PRIMARY KEY (id);


--
-- Name: states states_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.states
    ADD CONSTRAINT states_pkey PRIMARY KEY (id);


--
-- Name: time_tracks time_tracks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_tracks
    ADD CONSTRAINT time_tracks_pkey PRIMARY KEY (id);


--
-- Name: project_estimate_settings uni_project_estimate_settings_project_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_estimate_settings
    ADD CONSTRAINT uni_project_estimate_settings_project_id UNIQUE (project_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: webhooks webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);


--
-- Name: work_item_templates work_item_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_item_templates
    ADD CONSTRAINT work_item_templates_pkey PRIMARY KEY (id);


--
-- Name: workflows workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT workflows_pkey PRIMARY KEY (id);


--
-- Name: workspace_members workspace_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT workspace_members_pkey PRIMARY KEY (id);


--
-- Name: workspaces workspaces_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT workspaces_pkey PRIMARY KEY (id);


--
-- Name: idx_agent_activities_agent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_agent_activities_agent_id ON public.agent_activities USING btree (agent_id);


--
-- Name: idx_agent_activities_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_agent_activities_deleted_at ON public.agent_activities USING btree (deleted_at);


--
-- Name: idx_agent_activities_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_agent_activities_issue_id ON public.agent_activities USING btree (issue_id);


--
-- Name: idx_agents_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_agents_deleted_at ON public.agents USING btree (deleted_at);


--
-- Name: idx_agents_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_agents_workspace_id ON public.agents USING btree (workspace_id);


--
-- Name: idx_ai_configs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_configs_deleted_at ON public.ai_configs USING btree (deleted_at);


--
-- Name: idx_ai_configs_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_ai_configs_workspace_id ON public.ai_configs USING btree (workspace_id);


--
-- Name: idx_ai_messages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_messages_deleted_at ON public.ai_messages USING btree (deleted_at);


--
-- Name: idx_ai_messages_thread_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_messages_thread_id ON public.ai_messages USING btree (thread_id);


--
-- Name: idx_ai_threads_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_threads_deleted_at ON public.ai_threads USING btree (deleted_at);


--
-- Name: idx_ai_threads_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_threads_project_id ON public.ai_threads USING btree (project_id);


--
-- Name: idx_ai_threads_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_threads_user_id ON public.ai_threads USING btree (user_id);


--
-- Name: idx_ai_threads_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_threads_workspace_id ON public.ai_threads USING btree (workspace_id);


--
-- Name: idx_attachments_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attachments_deleted_at ON public.attachments USING btree (deleted_at);


--
-- Name: idx_attachments_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attachments_issue_id ON public.attachments USING btree (issue_id);


--
-- Name: idx_automation_rules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_automation_rules_deleted_at ON public.automation_rules USING btree (deleted_at);


--
-- Name: idx_automation_rules_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_automation_rules_project_id ON public.automation_rules USING btree (project_id);


--
-- Name: idx_comments_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_comments_deleted_at ON public.comments USING btree (deleted_at);


--
-- Name: idx_comments_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_comments_issue_id ON public.comments USING btree (issue_id);


--
-- Name: idx_conditional_fields_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_conditional_fields_deleted_at ON public.conditional_fields USING btree (deleted_at);


--
-- Name: idx_conditional_fields_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_conditional_fields_field_id ON public.conditional_fields USING btree (field_id);


--
-- Name: idx_conditional_fields_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_conditional_fields_workspace_id ON public.conditional_fields USING btree (workspace_id);


--
-- Name: idx_custom_field_options_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_custom_field_options_deleted_at ON public.custom_field_options USING btree (deleted_at);


--
-- Name: idx_custom_field_options_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_custom_field_options_field_id ON public.custom_field_options USING btree (field_id);


--
-- Name: idx_custom_fields_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_custom_fields_deleted_at ON public.custom_fields USING btree (deleted_at);


--
-- Name: idx_custom_fields_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_custom_fields_project_id ON public.custom_fields USING btree (project_id);


--
-- Name: idx_custom_fields_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_custom_fields_workspace_id ON public.custom_fields USING btree (workspace_id);


--
-- Name: idx_cycles_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cycles_deleted_at ON public.cycles USING btree (deleted_at);


--
-- Name: idx_cycles_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cycles_project_id ON public.cycles USING btree (project_id);


--
-- Name: idx_estimate_categories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_estimate_categories_deleted_at ON public.estimate_categories USING btree (deleted_at);


--
-- Name: idx_estimate_categories_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_estimate_categories_project_id ON public.estimate_categories USING btree (project_id);


--
-- Name: idx_estimate_categories_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_estimate_categories_workspace_id ON public.estimate_categories USING btree (workspace_id);


--
-- Name: idx_estimate_points_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_estimate_points_deleted_at ON public.estimate_points USING btree (deleted_at);


--
-- Name: idx_estimate_points_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_estimate_points_project_id ON public.estimate_points USING btree (project_id);


--
-- Name: idx_estimate_points_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_estimate_points_workspace_id ON public.estimate_points USING btree (workspace_id);


--
-- Name: idx_estimate_times_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_estimate_times_deleted_at ON public.estimate_times USING btree (deleted_at);


--
-- Name: idx_estimate_times_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_estimate_times_project_id ON public.estimate_times USING btree (project_id);


--
-- Name: idx_estimate_times_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_estimate_times_workspace_id ON public.estimate_times USING btree (workspace_id);


--
-- Name: idx_github_connections_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_github_connections_deleted_at ON public.github_connections USING btree (deleted_at);


--
-- Name: idx_github_connections_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_github_connections_project_id ON public.github_connections USING btree (project_id);


--
-- Name: idx_github_connections_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_github_connections_workspace_id ON public.github_connections USING btree (workspace_id);


--
-- Name: idx_initiatives_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_initiatives_workspace_id ON public.initiatives USING btree (workspace_id);


--
-- Name: idx_issue_activities_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_activities_deleted_at ON public.issue_activities USING btree (deleted_at);


--
-- Name: idx_issue_custom_field_values_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_custom_field_values_deleted_at ON public.issue_custom_field_values USING btree (deleted_at);


--
-- Name: idx_issue_custom_field_values_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_custom_field_values_field_id ON public.issue_custom_field_values USING btree (field_id);


--
-- Name: idx_issue_custom_field_values_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_custom_field_values_issue_id ON public.issue_custom_field_values USING btree (issue_id);


--
-- Name: idx_issue_relations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_relations_deleted_at ON public.issue_relations USING btree (deleted_at);


--
-- Name: idx_issue_relations_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_relations_issue_id ON public.issue_relations USING btree (issue_id);


--
-- Name: idx_issue_relations_related_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_relations_related_issue_id ON public.issue_relations USING btree (related_issue_id);


--
-- Name: idx_issue_type_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_type_templates_deleted_at ON public.issue_type_templates USING btree (deleted_at);


--
-- Name: idx_issue_type_templates_parent_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_type_templates_parent_type_id ON public.issue_type_templates USING btree (parent_type_id);


--
-- Name: idx_issue_type_templates_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_type_templates_workspace_id ON public.issue_type_templates USING btree (workspace_id);


--
-- Name: idx_issue_types_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_types_deleted_at ON public.issue_types USING btree (deleted_at);


--
-- Name: idx_issue_types_parent_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_types_parent_type_id ON public.issue_types USING btree (parent_type_id);


--
-- Name: idx_issue_types_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_types_project_id ON public.issue_types USING btree (project_id);


--
-- Name: idx_issue_types_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issue_types_workspace_id ON public.issue_types USING btree (workspace_id);


--
-- Name: idx_issues_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issues_deleted_at ON public.issues USING btree (deleted_at);


--
-- Name: idx_issues_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issues_project_id ON public.issues USING btree (project_id);


--
-- Name: idx_issues_search; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issues_search ON public.issues USING gin (to_tsvector('english'::regconfig, (((COALESCE(name, ''::character varying))::text || ' '::text) || COALESCE(description_stripped, ''::text))));


--
-- Name: idx_issues_state_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_issues_state_id ON public.issues USING btree (state_id);


--
-- Name: idx_labels_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labels_deleted_at ON public.labels USING btree (deleted_at);


--
-- Name: idx_labels_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labels_project_id ON public.labels USING btree (project_id);


--
-- Name: idx_mcp_configs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_mcp_configs_deleted_at ON public.mcp_configs USING btree (deleted_at);


--
-- Name: idx_mcp_configs_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_mcp_configs_workspace_id ON public.mcp_configs USING btree (workspace_id);


--
-- Name: idx_modules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_modules_deleted_at ON public.modules USING btree (deleted_at);


--
-- Name: idx_modules_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_modules_parent_id ON public.modules USING btree (parent_id);


--
-- Name: idx_modules_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_modules_project_id ON public.modules USING btree (project_id);


--
-- Name: idx_modules_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_modules_workspace_id ON public.modules USING btree (workspace_id);


--
-- Name: idx_notifications_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_deleted_at ON public.notifications USING btree (deleted_at);


--
-- Name: idx_notifications_is_read; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_is_read ON public.notifications USING btree (is_read);


--
-- Name: idx_notifications_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_issue_id ON public.notifications USING btree (issue_id);


--
-- Name: idx_notifications_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_project_id ON public.notifications USING btree (project_id);


--
-- Name: idx_notifications_recipient_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_recipient_id ON public.notifications USING btree (recipient_id);


--
-- Name: idx_pages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pages_deleted_at ON public.pages USING btree (deleted_at);


--
-- Name: idx_pages_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pages_parent_id ON public.pages USING btree (parent_id);


--
-- Name: idx_pages_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pages_project_id ON public.pages USING btree (project_id);


--
-- Name: idx_permissions_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_permissions_code ON public.permissions USING btree (code);


--
-- Name: idx_permissions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_permissions_deleted_at ON public.permissions USING btree (deleted_at);


--
-- Name: idx_proj_member_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_proj_member_user ON public.project_members USING btree (project_id, user_id);


--
-- Name: idx_proj_sub_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_proj_sub_user ON public.project_subscribers USING btree (project_id, user_id);


--
-- Name: idx_project_estimate_settings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_estimate_settings_deleted_at ON public.project_estimate_settings USING btree (deleted_at);


--
-- Name: idx_project_estimate_settings_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_estimate_settings_project_id ON public.project_estimate_settings USING btree (project_id);


--
-- Name: idx_project_estimate_settings_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_estimate_settings_workspace_id ON public.project_estimate_settings USING btree (workspace_id);


--
-- Name: idx_project_members_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_members_deleted_at ON public.project_members USING btree (deleted_at);


--
-- Name: idx_project_page_tabs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_page_tabs_deleted_at ON public.project_page_tabs USING btree (deleted_at);


--
-- Name: idx_project_page_tabs_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_page_tabs_owner_id ON public.project_page_tabs USING btree (owner_id);


--
-- Name: idx_project_page_tabs_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_page_tabs_project_id ON public.project_page_tabs USING btree (project_id);


--
-- Name: idx_project_subscribers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_subscribers_deleted_at ON public.project_subscribers USING btree (deleted_at);


--
-- Name: idx_project_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_templates_deleted_at ON public.project_templates USING btree (deleted_at);


--
-- Name: idx_project_templates_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_templates_workspace_id ON public.project_templates USING btree (workspace_id);


--
-- Name: idx_project_updates_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_updates_project_id ON public.project_updates USING btree (project_id);


--
-- Name: idx_projects_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_deleted_at ON public.projects USING btree (deleted_at);


--
-- Name: idx_projects_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_workspace_id ON public.projects USING btree (workspace_id);


--
-- Name: idx_recurrence_rules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_recurrence_rules_deleted_at ON public.recurrence_rules USING btree (deleted_at);


--
-- Name: idx_recurrence_rules_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_recurrence_rules_issue_id ON public.recurrence_rules USING btree (issue_id);


--
-- Name: idx_relation_types_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_relation_types_deleted_at ON public.relation_types USING btree (deleted_at);


--
-- Name: idx_relation_types_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_relation_types_workspace_id ON public.relation_types USING btree (workspace_id);


--
-- Name: idx_releases_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_releases_deleted_at ON public.releases USING btree (deleted_at);


--
-- Name: idx_releases_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_releases_project_id ON public.releases USING btree (project_id);


--
-- Name: idx_role_permissions_permission_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_role_permissions_permission_id ON public.role_permissions USING btree (permission_id);


--
-- Name: idx_role_permissions_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_role_permissions_role_id ON public.role_permissions USING btree (role_id);


--
-- Name: idx_roles_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_roles_deleted_at ON public.roles USING btree (deleted_at);


--
-- Name: idx_roles_level; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_roles_level ON public.roles USING btree (level);


--
-- Name: idx_roles_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_roles_project_id ON public.roles USING btree (project_id);


--
-- Name: idx_roles_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_roles_workspace_id ON public.roles USING btree (workspace_id);


--
-- Name: idx_saved_reports_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saved_reports_deleted_at ON public.saved_reports USING btree (deleted_at);


--
-- Name: idx_saved_reports_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saved_reports_project_id ON public.saved_reports USING btree (project_id);


--
-- Name: idx_saved_views_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saved_views_deleted_at ON public.saved_views USING btree (deleted_at);


--
-- Name: idx_saved_views_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saved_views_owner_id ON public.saved_views USING btree (owner_id);


--
-- Name: idx_saved_views_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saved_views_project_id ON public.saved_views USING btree (project_id);


--
-- Name: idx_search_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_search_templates_deleted_at ON public.search_templates USING btree (deleted_at);


--
-- Name: idx_search_templates_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_search_templates_owner_id ON public.search_templates USING btree (owner_id);


--
-- Name: idx_search_templates_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_search_templates_project_id ON public.search_templates USING btree (project_id);


--
-- Name: idx_slack_connections_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_slack_connections_deleted_at ON public.slack_connections USING btree (deleted_at);


--
-- Name: idx_slack_connections_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_slack_connections_project_id ON public.slack_connections USING btree (project_id);


--
-- Name: idx_slack_connections_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_slack_connections_workspace_id ON public.slack_connections USING btree (workspace_id);


--
-- Name: idx_state_transitions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_state_transitions_deleted_at ON public.state_transitions USING btree (deleted_at);


--
-- Name: idx_state_transitions_workflow_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_state_transitions_workflow_id ON public.state_transitions USING btree (workflow_id);


--
-- Name: idx_states_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_states_deleted_at ON public.states USING btree (deleted_at);


--
-- Name: idx_states_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_states_project_id ON public.states USING btree (project_id);


--
-- Name: idx_time_tracks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_tracks_deleted_at ON public.time_tracks USING btree (deleted_at);


--
-- Name: idx_time_tracks_issue_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_tracks_issue_id ON public.time_tracks USING btree (issue_id);


--
-- Name: idx_time_tracks_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_time_tracks_user_id ON public.time_tracks USING btree (user_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_username ON public.users USING btree (username);


--
-- Name: idx_webhooks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_webhooks_deleted_at ON public.webhooks USING btree (deleted_at);


--
-- Name: idx_webhooks_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_webhooks_project_id ON public.webhooks USING btree (project_id);


--
-- Name: idx_work_item_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_work_item_templates_deleted_at ON public.work_item_templates USING btree (deleted_at);


--
-- Name: idx_work_item_templates_issue_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_work_item_templates_issue_type_id ON public.work_item_templates USING btree (issue_type_id);


--
-- Name: idx_work_item_templates_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_work_item_templates_project_id ON public.work_item_templates USING btree (project_id);


--
-- Name: idx_work_item_templates_workspace_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_work_item_templates_workspace_id ON public.work_item_templates USING btree (workspace_id);


--
-- Name: idx_workflows_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflows_deleted_at ON public.workflows USING btree (deleted_at);


--
-- Name: idx_workflows_issue_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflows_issue_type_id ON public.workflows USING btree (issue_type_id);


--
-- Name: idx_workflows_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workflows_project_id ON public.workflows USING btree (project_id);


--
-- Name: idx_workspace_members_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workspace_members_deleted_at ON public.workspace_members USING btree (deleted_at);


--
-- Name: idx_workspaces_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workspaces_deleted_at ON public.workspaces USING btree (deleted_at);


--
-- Name: idx_workspaces_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_workspaces_slug ON public.workspaces USING btree (slug);


--
-- Name: idx_ws_member_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_ws_member_user ON public.workspace_members USING btree (workspace_id, user_id);


--
-- Name: agent_activities fk_agent_activities_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_activities
    ADD CONSTRAINT fk_agent_activities_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);


--
-- Name: agent_activities fk_agents_activities; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_activities
    ADD CONSTRAINT fk_agents_activities FOREIGN KEY (agent_id) REFERENCES public.agents(id);


--
-- Name: agents fk_agents_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT fk_agents_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);


--
-- Name: ai_messages fk_ai_threads_messages; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_messages
    ADD CONSTRAINT fk_ai_threads_messages FOREIGN KEY (thread_id) REFERENCES public.ai_threads(id);


--
-- Name: automation_rules fk_automation_rules_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.automation_rules
    ADD CONSTRAINT fk_automation_rules_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: comments fk_comments_author; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT fk_comments_author FOREIGN KEY (author_id) REFERENCES public.users(id);


--
-- Name: comments fk_comments_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT fk_comments_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id) ON DELETE CASCADE;


--
-- Name: custom_field_options fk_custom_fields_options; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.custom_field_options
    ADD CONSTRAINT fk_custom_fields_options FOREIGN KEY (field_id) REFERENCES public.custom_fields(id);


--
-- Name: issue_type_fields fk_custom_fields_type_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_fields
    ADD CONSTRAINT fk_custom_fields_type_links FOREIGN KEY (field_id) REFERENCES public.custom_fields(id);


--
-- Name: custom_fields fk_custom_fields_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.custom_fields
    ADD CONSTRAINT fk_custom_fields_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);


--
-- Name: issue_cycles fk_cycles_issue_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_cycles
    ADD CONSTRAINT fk_cycles_issue_links FOREIGN KEY (cycle_id) REFERENCES public.cycles(id);


--
-- Name: github_connections fk_github_connections_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_connections
    ADD CONSTRAINT fk_github_connections_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: github_connections fk_github_connections_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_connections
    ADD CONSTRAINT fk_github_connections_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);


--
-- Name: issue_custom_field_values fk_issue_custom_field_values_field; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_custom_field_values
    ADD CONSTRAINT fk_issue_custom_field_values_field FOREIGN KEY (field_id) REFERENCES public.custom_fields(id) ON DELETE CASCADE;


--
-- Name: issue_custom_field_values fk_issue_custom_field_values_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_custom_field_values
    ADD CONSTRAINT fk_issue_custom_field_values_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id) ON DELETE CASCADE;


--
-- Name: issue_pages fk_issue_pages_page; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_pages
    ADD CONSTRAINT fk_issue_pages_page FOREIGN KEY (page_id) REFERENCES public.pages(id) ON DELETE CASCADE;


--
-- Name: issue_relations fk_issue_relations_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_relations
    ADD CONSTRAINT fk_issue_relations_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id) ON DELETE CASCADE;


--
-- Name: issue_relations fk_issue_relations_related_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_relations
    ADD CONSTRAINT fk_issue_relations_related_issue FOREIGN KEY (related_issue_id) REFERENCES public.issues(id) ON DELETE CASCADE;


--
-- Name: issue_relations fk_issue_relations_relation_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_relations
    ADD CONSTRAINT fk_issue_relations_relation_type FOREIGN KEY (relation_type_id) REFERENCES public.relation_types(id);


--
-- Name: issue_type_template_fields fk_issue_type_template_fields_field; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_template_fields
    ADD CONSTRAINT fk_issue_type_template_fields_field FOREIGN KEY (field_id) REFERENCES public.custom_fields(id) ON DELETE CASCADE;


--
-- Name: issue_type_templates fk_issue_type_templates_child_types; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_templates
    ADD CONSTRAINT fk_issue_type_templates_child_types FOREIGN KEY (parent_type_id) REFERENCES public.issue_type_templates(id);


--
-- Name: issue_type_template_fields fk_issue_type_templates_field_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_template_fields
    ADD CONSTRAINT fk_issue_type_templates_field_links FOREIGN KEY (template_type_id) REFERENCES public.issue_type_templates(id);


--
-- Name: issue_types fk_issue_types_child_types; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_types
    ADD CONSTRAINT fk_issue_types_child_types FOREIGN KEY (parent_type_id) REFERENCES public.issue_types(id);


--
-- Name: issue_type_fields fk_issue_types_field_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_type_fields
    ADD CONSTRAINT fk_issue_types_field_links FOREIGN KEY (type_id) REFERENCES public.issue_types(id);


--
-- Name: issue_types fk_issue_types_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_types
    ADD CONSTRAINT fk_issue_types_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);


--
-- Name: issue_activities fk_issues_activities; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_activities
    ADD CONSTRAINT fk_issues_activities FOREIGN KEY (issue_id) REFERENCES public.issues(id);


--
-- Name: issue_assignees fk_issues_assignee_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_assignees
    ADD CONSTRAINT fk_issues_assignee_links FOREIGN KEY (issue_id) REFERENCES public.issues(id);


--
-- Name: issue_cycles fk_issues_cycle_link; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_cycles
    ADD CONSTRAINT fk_issues_cycle_link FOREIGN KEY (issue_id) REFERENCES public.issues(id);


--
-- Name: issues fk_issues_issue_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issues
    ADD CONSTRAINT fk_issues_issue_type FOREIGN KEY (issue_type_id) REFERENCES public.issue_types(id);


--
-- Name: issue_labels fk_issues_label_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_labels
    ADD CONSTRAINT fk_issues_label_links FOREIGN KEY (issue_id) REFERENCES public.issues(id);


--
-- Name: issue_pages fk_issues_page_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_pages
    ADD CONSTRAINT fk_issues_page_links FOREIGN KEY (issue_id) REFERENCES public.issues(id);


--
-- Name: issues fk_issues_sub_issues; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issues
    ADD CONSTRAINT fk_issues_sub_issues FOREIGN KEY (parent_id) REFERENCES public.issues(id);


--
-- Name: issue_labels fk_labels_issue_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_labels
    ADD CONSTRAINT fk_labels_issue_links FOREIGN KEY (label_id) REFERENCES public.labels(id);


--
-- Name: mcp_configs fk_mcp_configs_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mcp_configs
    ADD CONSTRAINT fk_mcp_configs_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);


--
-- Name: module_issues fk_module_issues_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_issues
    ADD CONSTRAINT fk_module_issues_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id) ON DELETE CASCADE;


--
-- Name: module_issues fk_modules_issue_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.module_issues
    ADD CONSTRAINT fk_modules_issue_links FOREIGN KEY (module_id) REFERENCES public.modules(id);


--
-- Name: modules fk_modules_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modules
    ADD CONSTRAINT fk_modules_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: notifications fk_notifications_recipient; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notifications_recipient FOREIGN KEY (recipient_id) REFERENCES public.users(id);


--
-- Name: notifications fk_notifications_sender; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notifications_sender FOREIGN KEY (sender_id) REFERENCES public.users(id);


--
-- Name: pages fk_pages_children; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT fk_pages_children FOREIGN KEY (parent_id) REFERENCES public.pages(id);


--
-- Name: pages fk_pages_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT fk_pages_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: project_page_tabs fk_project_page_tabs_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_page_tabs
    ADD CONSTRAINT fk_project_page_tabs_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: project_subscribers fk_project_subscribers_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_subscribers
    ADD CONSTRAINT fk_project_subscribers_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: project_subscribers fk_project_subscribers_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_subscribers
    ADD CONSTRAINT fk_project_subscribers_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: project_template_types fk_project_template_types_type_template; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_template_types
    ADD CONSTRAINT fk_project_template_types_type_template FOREIGN KEY (type_template_id) REFERENCES public.issue_type_templates(id) ON DELETE CASCADE;


--
-- Name: project_template_types fk_project_templates_type_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_template_types
    ADD CONSTRAINT fk_project_templates_type_links FOREIGN KEY (template_id) REFERENCES public.project_templates(id);


--
-- Name: project_updates fk_project_updates_author; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_updates
    ADD CONSTRAINT fk_project_updates_author FOREIGN KEY (author_id) REFERENCES public.users(id);


--
-- Name: cycles fk_projects_cycles; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cycles
    ADD CONSTRAINT fk_projects_cycles FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: projects fk_projects_default_assignee; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_default_assignee FOREIGN KEY (default_assignee_id) REFERENCES public.users(id);


--
-- Name: issues fk_projects_issues; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issues
    ADD CONSTRAINT fk_projects_issues FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: labels fk_projects_labels; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labels
    ADD CONSTRAINT fk_projects_labels FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: project_members fk_projects_members; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT fk_projects_members FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: projects fk_projects_project_lead; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_project_lead FOREIGN KEY (project_lead_id) REFERENCES public.users(id);


--
-- Name: states fk_projects_states; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.states
    ADD CONSTRAINT fk_projects_states FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: recurrence_rules fk_recurrence_rules_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recurrence_rules
    ADD CONSTRAINT fk_recurrence_rules_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);


--
-- Name: saved_reports fk_saved_reports_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_reports
    ADD CONSTRAINT fk_saved_reports_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: saved_views fk_saved_views_owner; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_views
    ADD CONSTRAINT fk_saved_views_owner FOREIGN KEY (owner_id) REFERENCES public.users(id);


--
-- Name: saved_views fk_saved_views_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_views
    ADD CONSTRAINT fk_saved_views_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: search_templates fk_search_templates_owner; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.search_templates
    ADD CONSTRAINT fk_search_templates_owner FOREIGN KEY (owner_id) REFERENCES public.users(id);


--
-- Name: search_templates fk_search_templates_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.search_templates
    ADD CONSTRAINT fk_search_templates_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: slack_connections fk_slack_connections_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_connections
    ADD CONSTRAINT fk_slack_connections_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: slack_connections fk_slack_connections_workspace; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_connections
    ADD CONSTRAINT fk_slack_connections_workspace FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);


--
-- Name: issues fk_states_issues; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issues
    ADD CONSTRAINT fk_states_issues FOREIGN KEY (state_id) REFERENCES public.states(id);


--
-- Name: state_transitions fk_states_source_transitions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_transitions
    ADD CONSTRAINT fk_states_source_transitions FOREIGN KEY (source_state_id) REFERENCES public.states(id);


--
-- Name: state_transitions fk_states_target_transitions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_transitions
    ADD CONSTRAINT fk_states_target_transitions FOREIGN KEY (target_state_id) REFERENCES public.states(id);


--
-- Name: time_tracks fk_time_tracks_issue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_tracks
    ADD CONSTRAINT fk_time_tracks_issue FOREIGN KEY (issue_id) REFERENCES public.issues(id);


--
-- Name: time_tracks fk_time_tracks_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.time_tracks
    ADD CONSTRAINT fk_time_tracks_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: issue_assignees fk_users_assigned_issues; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.issue_assignees
    ADD CONSTRAINT fk_users_assigned_issues FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: project_members fk_users_projects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT fk_users_projects FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: workspace_members fk_users_workspaces; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT fk_users_workspaces FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: work_item_templates fk_work_item_templates_issue_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.work_item_templates
    ADD CONSTRAINT fk_work_item_templates_issue_type FOREIGN KEY (issue_type_id) REFERENCES public.issue_types(id);


--
-- Name: workflows fk_workflows_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workflows
    ADD CONSTRAINT fk_workflows_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: state_transitions fk_workflows_transitions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.state_transitions
    ADD CONSTRAINT fk_workflows_transitions FOREIGN KEY (workflow_id) REFERENCES public.workflows(id);


--
-- Name: workspace_members fk_workspaces_members; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspace_members
    ADD CONSTRAINT fk_workspaces_members FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);


--
-- Name: workspaces fk_workspaces_owner; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workspaces
    ADD CONSTRAINT fk_workspaces_owner FOREIGN KEY (owner_id) REFERENCES public.users(id);


--
-- Name: projects fk_workspaces_projects; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_workspaces_projects FOREIGN KEY (workspace_id) REFERENCES public.workspaces(id);


--
-- PostgreSQL database dump complete
--

\unrestrict dQIpgiUggGs6y8yAccH6qYAWe7lVbdlon8SC5dMhWyfgQe0K6c3PVZYM3VTK2S3

