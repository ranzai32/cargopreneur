INSERT INTO suppliers (id, name, description, region, is_verified) VALUES
    (
        '11111111-1111-4111-8111-111111111111',
        'Harbor Mill Supply',
        'Packaging supplier reviewed for the catalog.',
        'Grand Rapids, MI',
        true
    ),
    (
        '22222222-2222-4222-8222-222222222222',
        'Northline Paper',
        'Paper goods supplier reviewed for the catalog.',
        'Columbus, OH',
        true
    );

INSERT INTO products (id, supplier_id, name, description, category, tags, unit, retail_price) VALUES
    (
        '33333333-3333-4333-8333-333333333333',
        '11111111-1111-4111-8111-111111111111',
        'Kraft mailer boxes, 9 × 6 × 3 in',
        'Carton of 50 mailer boxes.',
        'Packaging',
        ARRAY['mailer', 'boxes'],
        'carton',
        16.50
    ),
    (
        '44444444-4444-4444-8444-444444444444',
        '22222222-2222-4222-8222-222222222222',
        'Shop towels, blue, case of 50',
        'Case of 50 blue shop towels.',
        'Workshop supplies',
        ARRAY['towels', 'workshop'],
        'case',
        28.00
    );

INSERT INTO price_tiers (product_id, min_qty, unit_price) VALUES
    ('33333333-3333-4333-8333-333333333333', 100, 11.20),
    ('33333333-3333-4333-8333-333333333333', 140, 9.60),
    ('33333333-3333-4333-8333-333333333333', 180, 8.40),
    ('44444444-4444-4444-8444-444444444444', 40, 19.50),
    ('44444444-4444-4444-8444-444444444444', 80, 16.75);
