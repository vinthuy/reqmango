-- Field Permissions Migration
-- Adds field-level permission support

-- 1. Add field_name column to permissions table
ALTER TABLE public.permissions ADD COLUMN IF NOT EXISTS field_name character varying(100) DEFAULT '';

-- 2. Field Permissions table
CREATE TABLE IF NOT EXISTS public.field_permissions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    resource character varying(100) NOT NULL,
    field_name character varying(100) NOT NULL,
    role_id bigint NOT NULL,
    can_read boolean DEFAULT true,
    can_write boolean DEFAULT false,
    project_id bigint,
    workspace_id bigint
);

CREATE SEQUENCE IF NOT EXISTS public.field_permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.field_permissions_id_seq OWNED BY public.field_permissions.id;
ALTER TABLE ONLY public.field_permissions ALTER COLUMN id SET DEFAULT nextval('public.field_permissions_id_seq'::regclass);
ALTER TABLE ONLY public.field_permissions ADD CONSTRAINT field_permissions_pkey PRIMARY KEY (id);
CREATE INDEX idx_field_permissions_deleted_at ON public.field_permissions USING btree (deleted_at);
CREATE INDEX idx_field_permissions_resource ON public.field_permissions USING btree (resource);
CREATE INDEX idx_field_permissions_role_id ON public.field_permissions USING btree (role_id);
CREATE INDEX idx_field_permissions_project_id ON public.field_permissions USING btree (project_id);
CREATE INDEX idx_field_permissions_workspace_id ON public.field_permissions USING btree (workspace_id);
