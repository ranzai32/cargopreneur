ALTER TABLE order_participations
    DROP CONSTRAINT IF EXISTS order_participations_receipt_status_check;

ALTER TABLE order_participations
    DROP COLUMN IF EXISTS received_at,
    DROP COLUMN IF EXISTS receipt_status;
