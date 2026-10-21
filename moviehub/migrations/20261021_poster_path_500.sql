-- Use only if schema.sql has already been executed before poster_path was enlarged.
-- Do not rerun schema.sql on an existing database.
BEGIN;
ALTER TABLE public.movies ALTER COLUMN poster_path TYPE VARCHAR(500);
COMMIT;
