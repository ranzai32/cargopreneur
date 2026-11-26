ALTER TABLE users DROP CONSTRAINT users_role_check;

ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('business', 'admin', 'supplier'));

ALTER TABLE users
    ADD COLUMN supplier_id UUID REFERENCES suppliers (id);

CREATE UNIQUE INDEX users_supplier_id_unique ON users (supplier_id) WHERE supplier_id IS NOT NULL;

CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID REFERENCES suppliers (id) ON DELETE CASCADE,
    product_id UUID REFERENCES products (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    rating INTEGER NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT reviews_rating_check CHECK (rating BETWEEN 1 AND 5),
    CONSTRAINT reviews_target_check CHECK ((supplier_id IS NOT NULL AND product_id IS NULL) OR (supplier_id IS NULL AND product_id IS NOT NULL))
);

CREATE UNIQUE INDEX reviews_supplier_user_unique ON reviews (supplier_id, user_id) WHERE supplier_id IS NOT NULL;

CREATE UNIQUE INDEX reviews_product_user_unique ON reviews (product_id, user_id) WHERE product_id IS NOT NULL;

CREATE INDEX reviews_supplier_id_idx ON reviews (supplier_id);

CREATE INDEX reviews_product_id_idx ON reviews (product_id);

INSERT INTO users (id, business_name, contact_name, email, password_hash, region, role, supplier_id) VALUES
  ('b1000000-0000-4000-8000-000000000001', 'Harbor Mill Supply', 'Harbor Desk', 'harbor@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Grand Rapids, MI', 'supplier', '11111111-1111-4111-8111-111111111111'),
  ('b1000000-0000-4000-8000-000000000002', 'Northline Paper', 'Northline Desk', 'northline@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Columbus, OH', 'supplier', '22222222-2222-4222-8222-222222222222'),
  ('b1000000-0000-4000-8000-000000000003', 'Lakeshore Janitorial', 'Lakeshore Desk', 'lakeshore@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Milwaukee, WI', 'supplier', '55555555-5555-4555-8555-555555555555')
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (id, business_name, contact_name, email, password_hash, region, role) VALUES
  ('a1000000-0000-4000-8000-000000000001', 'Maple Street Market', 'Ada Cole', 'review01@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Grand Rapids, MI', 'business'),
  ('a1000000-0000-4000-8000-000000000002', 'River Bend Hardware', 'Noah Park', 'review02@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Milwaukee, WI', 'business'),
  ('a1000000-0000-4000-8000-000000000003', 'Lantern Cafe Supply', 'Mina Cho', 'review03@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Columbus, OH', 'business'),
  ('a1000000-0000-4000-8000-000000000004', 'North Dock Grocers', 'Eli Ward', 'review04@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Grand Rapids, MI', 'business'),
  ('a1000000-0000-4000-8000-000000000005', 'Workshop Row', 'Sara Nguyen', 'review05@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Milwaukee, WI', 'business'),
  ('a1000000-0000-4000-8000-000000000006', 'Paper & Parcel', 'Jonah Reed', 'review06@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Columbus, OH', 'business'),
  ('a1000000-0000-4000-8000-000000000007', 'Clean Counter Co', 'Priya Shah', 'review07@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Newark, NJ', 'business'),
  ('a1000000-0000-4000-8000-000000000008', 'Harbor Lunchroom', 'Owen Blake', 'review08@cargoprener.local', $hash$$2a$10$g3w9HbzJSn4i8.uQRryG7ezKuazdhBwKpsKyPL9r4oj.oAdqVeJo.$hash$, 'Grand Rapids, MI', 'business')
ON CONFLICT (email) DO NOTHING;

INSERT INTO reviews (supplier_id, product_id, user_id, rating, body, created_at) VALUES
  ('11111111-1111-4111-8111-111111111111', NULL, 'a1000000-0000-4000-8000-000000000001', 4, 'Clear tiers and the goods matched the catalog.', '2026-11-01T15:00:00Z'),
  ('11111111-1111-4111-8111-111111111111', NULL, 'a1000000-0000-4000-8000-000000000002', 5, 'Responsive enough for a pooled order.', '2026-11-02T15:00:00Z'),
  ('11111111-1111-4111-8111-111111111111', NULL, 'a1000000-0000-4000-8000-000000000003', 5, 'We trust this supplier for a weekly restock.', '2026-11-03T15:00:00Z'),
  ('11111111-1111-4111-8111-111111111111', NULL, 'a1000000-0000-4000-8000-000000000004', 4, 'Packaging was tidy and the invoice matched.', '2026-11-04T15:00:00Z'),
  ('11111111-1111-4111-8111-111111111111', NULL, 'a1000000-0000-4000-8000-000000000005', 5, 'Good partner for shops that cannot meet a full minimum alone.', '2026-11-05T15:00:00Z'),
  ('11111111-1111-4111-8111-111111111111', NULL, 'a1000000-0000-4000-8000-000000000006', 5, 'A dependable listing, nothing missing from the carton.', '2026-11-06T15:00:00Z'),
  ('11111111-1111-4111-8111-111111111111', NULL, 'a1000000-0000-4000-8000-000000000007', 4, 'Prices drop in a way our shop can actually use.', '2026-11-07T15:00:00Z'),
  ('11111111-1111-4111-8111-111111111111', NULL, 'a1000000-0000-4000-8000-000000000008', 5, 'Would join another order from this supplier.', '2026-11-08T15:00:00Z'),
  ('22222222-2222-4222-8222-222222222222', NULL, 'a1000000-0000-4000-8000-000000000001', 4, 'Responsive enough for a pooled order.', '2026-11-09T15:00:00Z'),
  ('22222222-2222-4222-8222-222222222222', NULL, 'a1000000-0000-4000-8000-000000000002', 5, 'We trust this supplier for a weekly restock.', '2026-11-10T15:00:00Z'),
  ('22222222-2222-4222-8222-222222222222', NULL, 'a1000000-0000-4000-8000-000000000003', 3, 'Packaging was tidy and the invoice matched.', '2026-11-11T15:00:00Z'),
  ('22222222-2222-4222-8222-222222222222', NULL, 'a1000000-0000-4000-8000-000000000004', 4, 'Good partner for shops that cannot meet a full minimum alone.', '2026-11-12T15:00:00Z'),
  ('22222222-2222-4222-8222-222222222222', NULL, 'a1000000-0000-4000-8000-000000000005', 5, 'A dependable listing, nothing missing from the carton.', '2026-11-13T15:00:00Z'),
  ('22222222-2222-4222-8222-222222222222', NULL, 'a1000000-0000-4000-8000-000000000006', 3, 'Prices drop in a way our shop can actually use.', '2026-11-14T15:00:00Z'),
  ('22222222-2222-4222-8222-222222222222', NULL, 'a1000000-0000-4000-8000-000000000007', 4, 'Would join another order from this supplier.', '2026-11-15T15:00:00Z'),
  ('22222222-2222-4222-8222-222222222222', NULL, 'a1000000-0000-4000-8000-000000000008', 5, 'Clear tiers and the goods matched the catalog.', '2026-11-16T15:00:00Z'),
  ('55555555-5555-4555-8555-555555555555', NULL, 'a1000000-0000-4000-8000-000000000001', 5, 'We trust this supplier for a weekly restock.', '2026-11-17T15:00:00Z'),
  ('55555555-5555-4555-8555-555555555555', NULL, 'a1000000-0000-4000-8000-000000000002', 3, 'Packaging was tidy and the invoice matched.', '2026-11-18T15:00:00Z'),
  ('55555555-5555-4555-8555-555555555555', NULL, 'a1000000-0000-4000-8000-000000000003', 4, 'Good partner for shops that cannot meet a full minimum alone.', '2026-11-19T15:00:00Z'),
  ('55555555-5555-4555-8555-555555555555', NULL, 'a1000000-0000-4000-8000-000000000004', 5, 'A dependable listing, nothing missing from the carton.', '2026-11-20T15:00:00Z'),
  ('55555555-5555-4555-8555-555555555555', NULL, 'a1000000-0000-4000-8000-000000000005', 3, 'Prices drop in a way our shop can actually use.', '2026-11-01T15:00:00Z'),
  ('55555555-5555-4555-8555-555555555555', NULL, 'a1000000-0000-4000-8000-000000000006', 4, 'Would join another order from this supplier.', '2026-11-02T15:00:00Z'),
  ('55555555-5555-4555-8555-555555555555', NULL, 'a1000000-0000-4000-8000-000000000007', 5, 'Clear tiers and the goods matched the catalog.', '2026-11-03T15:00:00Z'),
  ('55555555-5555-4555-8555-555555555555', NULL, 'a1000000-0000-4000-8000-000000000008', 3, 'Responsive enough for a pooled order.', '2026-11-04T15:00:00Z'),
  (NULL, '33333333-3333-4333-8333-333333333333', 'a1000000-0000-4000-8000-000000000001', 5, 'Solid wholesale quality for a small shop.', '2026-10-01T15:00:00Z'),
  (NULL, '33333333-3333-4333-8333-333333333333', 'a1000000-0000-4000-8000-000000000002', 3, 'Arrived as listed and the count was exact.', '2026-10-02T15:00:00Z'),
  (NULL, '33333333-3333-4333-8333-333333333333', 'a1000000-0000-4000-8000-000000000003', 4, 'The higher tier price was worth the pooled order.', '2026-10-03T15:00:00Z'),
  (NULL, '33333333-3333-4333-8333-333333333333', 'a1000000-0000-4000-8000-000000000004', 5, 'We restocked without surprises.', '2026-10-04T15:00:00Z'),
  (NULL, '33333333-3333-4333-8333-333333333333', 'a1000000-0000-4000-8000-000000000005', 2, 'Plain goods, reliable cartons.', '2026-10-05T15:00:00Z'),
  (NULL, '33333333-3333-4333-8333-333333333333', 'a1000000-0000-4000-8000-000000000006', 5, 'Would buy this again next month.', '2026-10-06T15:00:00Z'),
  (NULL, '33333333-3333-4333-8333-333333333333', 'a1000000-0000-4000-8000-000000000007', 4, 'Better than the retail price once the minimum was met.', '2026-10-07T15:00:00Z'),
  (NULL, '33333333-3333-4333-8333-333333333333', 'a1000000-0000-4000-8000-000000000008', 5, 'A few scuffs on the outer box, contents were fine.', '2026-10-08T15:00:00Z'),
  (NULL, '44444444-4444-4444-8444-444444444444', 'a1000000-0000-4000-8000-000000000001', 2, 'Easy to store and the unit size fits our counter.', '2026-10-09T15:00:00Z'),
  (NULL, '44444444-4444-4444-8444-444444444444', 'a1000000-0000-4000-8000-000000000002', 3, 'Consistent from the last pooled order.', '2026-10-10T15:00:00Z'),
  (NULL, '44444444-4444-4444-8444-444444444444', 'a1000000-0000-4000-8000-000000000003', 5, 'Solid wholesale quality for a small shop.', '2026-10-11T15:00:00Z'),
  (NULL, '44444444-4444-4444-8444-444444444444', 'a1000000-0000-4000-8000-000000000004', 5, 'Arrived as listed and the count was exact.', '2026-10-12T15:00:00Z'),
  (NULL, '44444444-4444-4444-8444-444444444444', 'a1000000-0000-4000-8000-000000000005', 2, 'The higher tier price was worth the pooled order.', '2026-10-13T15:00:00Z'),
  (NULL, '44444444-4444-4444-8444-444444444444', 'a1000000-0000-4000-8000-000000000006', 3, 'We restocked without surprises.', '2026-10-14T15:00:00Z'),
  (NULL, '44444444-4444-4444-8444-444444444444', 'a1000000-0000-4000-8000-000000000007', 4, 'Plain goods, reliable cartons.', '2026-10-15T15:00:00Z'),
  (NULL, '44444444-4444-4444-8444-444444444444', 'a1000000-0000-4000-8000-000000000008', 5, 'Would buy this again next month.', '2026-10-16T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777701', 'a1000000-0000-4000-8000-000000000001', 2, 'Better than the retail price once the minimum was met.', '2026-10-17T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777701', 'a1000000-0000-4000-8000-000000000002', 3, 'A few scuffs on the outer box, contents were fine.', '2026-10-18T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777701', 'a1000000-0000-4000-8000-000000000003', 4, 'Easy to store and the unit size fits our counter.', '2026-10-19T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777701', 'a1000000-0000-4000-8000-000000000004', 5, 'Consistent from the last pooled order.', '2026-10-20T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777701', 'a1000000-0000-4000-8000-000000000005', 5, 'Solid wholesale quality for a small shop.', '2026-10-21T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777701', 'a1000000-0000-4000-8000-000000000006', 3, 'Arrived as listed and the count was exact.', '2026-10-22T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777701', 'a1000000-0000-4000-8000-000000000007', 4, 'The higher tier price was worth the pooled order.', '2026-10-23T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777701', 'a1000000-0000-4000-8000-000000000008', 5, 'We restocked without surprises.', '2026-10-24T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777702', 'a1000000-0000-4000-8000-000000000001', 2, 'Plain goods, reliable cartons.', '2026-10-01T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777702', 'a1000000-0000-4000-8000-000000000002', 5, 'Would buy this again next month.', '2026-10-02T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777702', 'a1000000-0000-4000-8000-000000000003', 4, 'Better than the retail price once the minimum was met.', '2026-10-03T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777702', 'a1000000-0000-4000-8000-000000000004', 5, 'A few scuffs on the outer box, contents were fine.', '2026-10-04T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777702', 'a1000000-0000-4000-8000-000000000005', 2, 'Easy to store and the unit size fits our counter.', '2026-10-05T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777702', 'a1000000-0000-4000-8000-000000000006', 3, 'Consistent from the last pooled order.', '2026-10-06T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777702', 'a1000000-0000-4000-8000-000000000007', 5, 'Solid wholesale quality for a small shop.', '2026-10-07T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777702', 'a1000000-0000-4000-8000-000000000008', 5, 'Arrived as listed and the count was exact.', '2026-10-08T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777703', 'a1000000-0000-4000-8000-000000000001', 2, 'The higher tier price was worth the pooled order.', '2026-10-09T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777703', 'a1000000-0000-4000-8000-000000000002', 3, 'We restocked without surprises.', '2026-10-10T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777703', 'a1000000-0000-4000-8000-000000000003', 4, 'Plain goods, reliable cartons.', '2026-10-11T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777703', 'a1000000-0000-4000-8000-000000000004', 5, 'Would buy this again next month.', '2026-10-12T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777703', 'a1000000-0000-4000-8000-000000000005', 2, 'Better than the retail price once the minimum was met.', '2026-10-13T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777703', 'a1000000-0000-4000-8000-000000000006', 3, 'A few scuffs on the outer box, contents were fine.', '2026-10-14T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777703', 'a1000000-0000-4000-8000-000000000007', 4, 'Easy to store and the unit size fits our counter.', '2026-10-15T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777703', 'a1000000-0000-4000-8000-000000000008', 5, 'Consistent from the last pooled order.', '2026-10-16T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777704', 'a1000000-0000-4000-8000-000000000001', 5, 'Solid wholesale quality for a small shop.', '2026-10-17T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777704', 'a1000000-0000-4000-8000-000000000002', 3, 'Arrived as listed and the count was exact.', '2026-10-18T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777704', 'a1000000-0000-4000-8000-000000000003', 4, 'The higher tier price was worth the pooled order.', '2026-10-19T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777704', 'a1000000-0000-4000-8000-000000000004', 5, 'We restocked without surprises.', '2026-10-20T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777704', 'a1000000-0000-4000-8000-000000000005', 2, 'Plain goods, reliable cartons.', '2026-10-21T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777704', 'a1000000-0000-4000-8000-000000000006', 5, 'Would buy this again next month.', '2026-10-22T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777704', 'a1000000-0000-4000-8000-000000000007', 4, 'Better than the retail price once the minimum was met.', '2026-10-23T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777704', 'a1000000-0000-4000-8000-000000000008', 5, 'A few scuffs on the outer box, contents were fine.', '2026-10-24T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777705', 'a1000000-0000-4000-8000-000000000001', 2, 'Easy to store and the unit size fits our counter.', '2026-10-01T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777705', 'a1000000-0000-4000-8000-000000000002', 3, 'Consistent from the last pooled order.', '2026-10-02T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777705', 'a1000000-0000-4000-8000-000000000003', 5, 'Solid wholesale quality for a small shop.', '2026-10-03T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777705', 'a1000000-0000-4000-8000-000000000004', 5, 'Arrived as listed and the count was exact.', '2026-10-04T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777705', 'a1000000-0000-4000-8000-000000000005', 2, 'The higher tier price was worth the pooled order.', '2026-10-05T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777705', 'a1000000-0000-4000-8000-000000000006', 3, 'We restocked without surprises.', '2026-10-06T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777705', 'a1000000-0000-4000-8000-000000000007', 4, 'Plain goods, reliable cartons.', '2026-10-07T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777705', 'a1000000-0000-4000-8000-000000000008', 5, 'Would buy this again next month.', '2026-10-08T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777706', 'a1000000-0000-4000-8000-000000000001', 2, 'Better than the retail price once the minimum was met.', '2026-10-09T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777706', 'a1000000-0000-4000-8000-000000000002', 3, 'A few scuffs on the outer box, contents were fine.', '2026-10-10T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777706', 'a1000000-0000-4000-8000-000000000003', 4, 'Easy to store and the unit size fits our counter.', '2026-10-11T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777706', 'a1000000-0000-4000-8000-000000000004', 5, 'Consistent from the last pooled order.', '2026-10-12T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777706', 'a1000000-0000-4000-8000-000000000005', 5, 'Solid wholesale quality for a small shop.', '2026-10-13T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777706', 'a1000000-0000-4000-8000-000000000006', 3, 'Arrived as listed and the count was exact.', '2026-10-14T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777706', 'a1000000-0000-4000-8000-000000000007', 4, 'The higher tier price was worth the pooled order.', '2026-10-15T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777706', 'a1000000-0000-4000-8000-000000000008', 5, 'We restocked without surprises.', '2026-10-16T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777707', 'a1000000-0000-4000-8000-000000000001', 2, 'Plain goods, reliable cartons.', '2026-10-17T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777707', 'a1000000-0000-4000-8000-000000000002', 5, 'Would buy this again next month.', '2026-10-18T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777707', 'a1000000-0000-4000-8000-000000000003', 4, 'Better than the retail price once the minimum was met.', '2026-10-19T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777707', 'a1000000-0000-4000-8000-000000000004', 5, 'A few scuffs on the outer box, contents were fine.', '2026-10-20T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777707', 'a1000000-0000-4000-8000-000000000005', 2, 'Easy to store and the unit size fits our counter.', '2026-10-21T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777707', 'a1000000-0000-4000-8000-000000000006', 3, 'Consistent from the last pooled order.', '2026-10-22T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777707', 'a1000000-0000-4000-8000-000000000007', 5, 'Solid wholesale quality for a small shop.', '2026-10-23T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777707', 'a1000000-0000-4000-8000-000000000008', 5, 'Arrived as listed and the count was exact.', '2026-10-24T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777708', 'a1000000-0000-4000-8000-000000000001', 2, 'The higher tier price was worth the pooled order.', '2026-10-01T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777708', 'a1000000-0000-4000-8000-000000000002', 3, 'We restocked without surprises.', '2026-10-02T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777708', 'a1000000-0000-4000-8000-000000000003', 4, 'Plain goods, reliable cartons.', '2026-10-03T15:00:00Z'),
  (NULL, '77777777-7777-4777-8777-777777777708', 'a1000000-0000-4000-8000-000000000004', 5, 'Would buy this again next month.', '2026-10-04T15:00:00Z');
