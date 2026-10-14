UPDATE products
SET created_at = '2026-07-01T15:00:00Z'
WHERE id = '33333333-3333-4333-8333-333333333333';

UPDATE products
SET created_at = '2026-07-15T15:00:00Z'
WHERE id = '44444444-4444-4444-8444-444444444444';

INSERT INTO price_tiers (product_id, min_qty, unit_price) VALUES
    ('44444444-4444-4444-8444-444444444444', 120, 14.25);

INSERT INTO suppliers (id, name, description, region, is_verified) VALUES
    (
        '55555555-5555-4555-8555-555555555555',
        'Lakeshore Janitorial',
        'Cleaning supplies reviewed for shops and workshops.',
        'Milwaukee, WI',
        true
    ),
    (
        '66666666-6666-4666-8666-666666666666',
        'QuickLot Imports',
        'Importer awaiting catalog review.',
        'Newark, NJ',
        false
    );

INSERT INTO products (id, supplier_id, name, description, category, tags, unit, retail_price, created_at) VALUES
    (
        '77777777-7777-4777-8777-777777777701',
        '11111111-1111-4111-8111-111111111111',
        'Stretch wrap film, 18 in x 1500 ft',
        'Clear pallet wrap sold by the roll.',
        'Packaging',
        ARRAY['wrap', 'pallet'],
        'roll',
        12.40,
        '2026-08-01T15:00:00Z'
    ),
    (
        '77777777-7777-4777-8777-777777777702',
        '11111111-1111-4111-8111-111111111111',
        'Corrugated boxes, 12 x 12 x 12 in',
        'Bundle of 25 shipping boxes.',
        'Packaging',
        ARRAY['boxes', 'shipping'],
        'bundle',
        22.00,
        '2026-08-10T15:00:00Z'
    ),
    (
        '77777777-7777-4777-8777-777777777703',
        '22222222-2222-4222-8222-222222222222',
        'Letter copy paper, 10-ream case',
        'Case of ten reams of letter paper.',
        'Store supplies',
        ARRAY['paper', 'office'],
        'case',
        42.00,
        '2026-08-18T15:00:00Z'
    ),
    (
        '77777777-7777-4777-8777-777777777704',
        '55555555-5555-4555-8555-555555555555',
        'Nitrile gloves, medium, box of 100',
        'Powder-free gloves for workshop and counter use.',
        'Workshop supplies',
        ARRAY['gloves', 'safety'],
        'box',
        14.80,
        '2026-08-25T15:00:00Z'
    ),
    (
        '77777777-7777-4777-8777-777777777705',
        '55555555-5555-4555-8555-555555555555',
        'Glass cleaner concentrate, 1 gallon',
        'Dilute for counters and front windows.',
        'Cleaning',
        ARRAY['cleaner', 'glass'],
        'gallon',
        18.25,
        '2026-09-02T15:00:00Z'
    ),
    (
        '77777777-7777-4777-8777-777777777706',
        '55555555-5555-4555-8555-555555555555',
        'Microfiber cloths, pack of 24',
        'Reusable cloths for glass and fixtures.',
        'Cleaning',
        ARRAY['cloths', 'cleaning'],
        'pack',
        16.00,
        '2026-09-08T15:00:00Z'
    ),
    (
        '77777777-7777-4777-8777-777777777707',
        '66666666-6666-4666-8666-666666666666',
        'LED shop bulbs, 12W, pack of 10',
        'Daylight bulbs for a sales floor.',
        'Lighting',
        ARRAY['bulbs', 'lighting'],
        'pack',
        24.00,
        '2026-09-12T15:00:00Z'
    ),
    (
        '77777777-7777-4777-8777-777777777708',
        '66666666-6666-4666-8666-666666666666',
        'Thermal receipt rolls, 50-roll case',
        'Receipt paper for a counter printer.',
        'Store supplies',
        ARRAY['receipts', 'paper'],
        'case',
        36.00,
        '2026-09-18T15:00:00Z'
    );

INSERT INTO price_tiers (product_id, min_qty, unit_price) VALUES
    ('77777777-7777-4777-8777-777777777701', 50, 8.10),
    ('77777777-7777-4777-8777-777777777701', 100, 6.90),
    ('77777777-7777-4777-8777-777777777701', 200, 5.75),
    ('77777777-7777-4777-8777-777777777702', 40, 15.40),
    ('77777777-7777-4777-8777-777777777702', 80, 13.10),
    ('77777777-7777-4777-8777-777777777702', 160, 11.25),
    ('77777777-7777-4777-8777-777777777703', 20, 31.50),
    ('77777777-7777-4777-8777-777777777703', 40, 28.00),
    ('77777777-7777-4777-8777-777777777703', 80, 24.75),
    ('77777777-7777-4777-8777-777777777704', 24, 9.90),
    ('77777777-7777-4777-8777-777777777704', 48, 8.40),
    ('77777777-7777-4777-8777-777777777704', 96, 7.15),
    ('77777777-7777-4777-8777-777777777705', 12, 12.40),
    ('77777777-7777-4777-8777-777777777705', 24, 10.80),
    ('77777777-7777-4777-8777-777777777705', 48, 9.35),
    ('77777777-7777-4777-8777-777777777706', 20, 11.20),
    ('77777777-7777-4777-8777-777777777706', 40, 9.50),
    ('77777777-7777-4777-8777-777777777706', 80, 8.10),
    ('77777777-7777-4777-8777-777777777707', 10, 15.50),
    ('77777777-7777-4777-8777-777777777707', 25, 13.20),
    ('77777777-7777-4777-8777-777777777707', 50, 11.40),
    ('77777777-7777-4777-8777-777777777708', 8, 24.80),
    ('77777777-7777-4777-8777-777777777708', 16, 21.40),
    ('77777777-7777-4777-8777-777777777708', 32, 18.60);
