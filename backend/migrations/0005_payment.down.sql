ALTER TABLE order_participations DROP CONSTRAINT IF EXISTS order_participations_payment_status_check;

ALTER TABLE order_participations
    DROP COLUMN IF EXISTS shipping_address,
    DROP COLUMN IF EXISTS paid_at,
    DROP COLUMN IF EXISTS payment_status;

UPDATE orders SET status = 'confirmed' WHERE status IN ('paid', 'completed');

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;

ALTER TABLE orders
    ADD CONSTRAINT orders_status_check
    CHECK (status IN ('open', 'confirmed', 'cancelled'));
