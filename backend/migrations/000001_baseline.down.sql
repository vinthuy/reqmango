-- Reverses the baseline: drops every object in the public schema except the
-- migration bookkeeping, so it can run on a database created by either path.
DO $$
DECLARE r record;
BEGIN
    FOR r IN SELECT tablename FROM pg_tables
             WHERE schemaname = 'public' AND tablename <> 'schema_migrations'
    LOOP
        EXECUTE format('DROP TABLE IF EXISTS public.%I CASCADE', r.tablename);
    END LOOP;

    FOR r IN SELECT sequencename FROM pg_sequences WHERE schemaname = 'public'
    LOOP
        EXECUTE format('DROP SEQUENCE IF EXISTS public.%I CASCADE', r.sequencename);
    END LOOP;
END $$;
