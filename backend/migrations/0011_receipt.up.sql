ALTER TABLE order_participations
    ADD COLUMN receipt_status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN received_at TIMESTAMPTZ NULL;

ALTER TABLE order_participations
    ADD CONSTRAINT order_participations_receipt_status_check
    CHECK (receipt_status IN ('pending', 'received'));
