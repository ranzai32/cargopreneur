ALTER TABLE disputes
    ADD COLUMN outcome TEXT NOT NULL DEFAULT '',
    ADD COLUMN resolved_at TIMESTAMPTZ;

CREATE UNIQUE INDEX disputes_one_open_per_user_idx
    ON disputes (order_id, opened_by)
    WHERE status = 'open';
