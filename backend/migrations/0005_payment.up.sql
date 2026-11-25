ALTER TABLE orders DROP CONSTRAINT orders_status_check;

ALTER TABLE orders
    ADD CONSTRAINT orders_status_check
    CHECK (status IN ('open', 'confirmed', 'paid', 'completed', 'cancelled'));

ALTER TABLE order_participations
    ADD COLUMN payment_status TEXT NOT NULL DEFAULT 'unpaid',
    ADD COLUMN paid_at TIMESTAMPTZ NULL,
    ADD COLUMN shipping_address TEXT NOT NULL DEFAULT '';

ALTER TABLE order_participations
    ADD CONSTRAINT order_participations_payment_status_check
    CHECK (payment_status IN ('unpaid', 'paid'));
