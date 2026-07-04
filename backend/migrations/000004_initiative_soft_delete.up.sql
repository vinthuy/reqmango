-- Initiative Soft Delete Migration
-- Adds deleted_at column to initiatives table for soft delete support

ALTER TABLE public.initiatives ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
CREATE INDEX IF NOT EXISTS idx_initiatives_deleted_at ON public.initiatives USING btree (deleted_at);
