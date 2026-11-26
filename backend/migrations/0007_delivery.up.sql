ALTER TABLE order_participations
    ADD COLUMN delivery_status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN shipped_at TIMESTAMPTZ NULL,
    ADD COLUMN delivered_at TIMESTAMPTZ NULL;

ALTER TABLE order_participations
    ADD CONSTRAINT order_participations_delivery_status_check
    CHECK (delivery_status IN ('pending', 'shipped', 'delivered'));
