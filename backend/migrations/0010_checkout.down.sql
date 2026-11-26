DROP TABLE IF EXISTS saved_addresses;

ALTER TABLE order_participations
    DROP COLUMN IF EXISTS delivery_method,
    DROP COLUMN IF EXISTS shipping_fee;
