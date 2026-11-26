ALTER TABLE order_participations
    ADD COLUMN delivery_method TEXT NOT NULL DEFAULT '',
    ADD COLUMN shipping_fee NUMERIC(12, 2) NOT NULL DEFAULT 0;

CREATE TABLE saved_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    line TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT saved_addresses_line_check CHECK (char_length(btrim(line)) >= 8)
);

CREATE INDEX saved_addresses_user_id_idx ON saved_addresses (user_id, updated_at DESC);
