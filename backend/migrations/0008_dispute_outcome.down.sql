DROP INDEX IF EXISTS disputes_one_open_per_user_idx;

ALTER TABLE disputes
    DROP COLUMN IF EXISTS resolved_at,
    DROP COLUMN IF EXISTS outcome;
