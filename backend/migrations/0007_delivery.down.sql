ALTER TABLE order_participations DROP CONSTRAINT IF EXISTS order_participations_delivery_status_check;

ALTER TABLE order_participations
    DROP COLUMN IF EXISTS delivered_at,
    DROP COLUMN IF EXISTS shipped_at,
    DROP COLUMN IF EXISTS delivery_status;
