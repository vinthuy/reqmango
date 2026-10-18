-- Page Enhancements Migration
-- Adds: version history, templates, locking, and export support

-- 1. Page Versions (version history)
CREATE TABLE IF NOT EXISTS public.page_versions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    created_by_id bigint,
    page_id bigint NOT NULL,
    title character varying(255) NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    content_json jsonb,
    version_number integer NOT NULL DEFAULT 1,
    change_summary character varying(500)
);

CREATE SEQUENCE IF NOT EXISTS public.page_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.page_versions_id_seq OWNED BY public.page_versions.id;
ALTER TABLE ONLY public.page_versions ALTER COLUMN id SET DEFAULT nextval('public.page_versions_id_seq'::regclass);
ALTER TABLE ONLY public.page_versions ADD CONSTRAINT page_versions_pkey PRIMARY KEY (id);
CREATE INDEX idx_page_versions_page_id ON public.page_versions USING btree (page_id);
CREATE INDEX idx_page_versions_deleted_at ON public.page_versions USING btree (deleted_at);
ALTER TABLE ONLY public.page_versions ADD CONSTRAINT fk_page_versions_page FOREIGN KEY (page_id) REFERENCES public.pages(id) ON DELETE CASCADE;

-- 2. Page Templates
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

CREATE SEQUENCE IF NOT EXISTS public.page_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.page_templates_id_seq OWNED BY public.page_templates.id;
ALTER TABLE ONLY public.page_templates ALTER COLUMN id SET DEFAULT nextval('public.page_templates_id_seq'::regclass);
ALTER TABLE ONLY public.page_templates ADD CONSTRAINT page_templates_pkey PRIMARY KEY (id);
CREATE INDEX idx_page_templates_deleted_at ON public.page_templates USING btree (deleted_at);
CREATE INDEX idx_page_templates_workspace_id ON public.page_templates USING btree (workspace_id);
CREATE INDEX idx_page_templates_project_id ON public.page_templates USING btree (project_id);

-- 3. Page Locking (add columns to pages)
ALTER TABLE public.pages ADD COLUMN IF NOT EXISTS locked_by_id bigint;
ALTER TABLE public.pages ADD COLUMN IF NOT EXISTS locked_at timestamp with time zone;
