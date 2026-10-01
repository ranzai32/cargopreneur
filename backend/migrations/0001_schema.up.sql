CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_name TEXT NOT NULL,
    contact_name TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    region TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'business',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_email_lowercase CHECK (email = lower(email)),
    CONSTRAINT users_role_check CHECK (role IN ('business', 'admin'))
);

CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    region TEXT NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID NOT NULL REFERENCES suppliers (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    unit TEXT NOT NULL,
    retail_price NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT products_retail_price_check CHECK (retail_price >= 0)
);

CREATE INDEX products_supplier_id_idx ON products (supplier_id);

CREATE TABLE price_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    min_qty INTEGER NOT NULL,
    unit_price NUMERIC(12, 2) NOT NULL,
    CONSTRAINT price_tiers_product_min_qty_unique UNIQUE (product_id, min_qty),
    CONSTRAINT price_tiers_min_qty_check CHECK (min_qty > 0),
    CONSTRAINT price_tiers_unit_price_check CHECK (unit_price >= 0)
);

CREATE INDEX price_tiers_product_id_idx ON price_tiers (product_id);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    creator_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'open',
    moq INTEGER NOT NULL,
    deadline TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT orders_status_check CHECK (status IN ('open', 'confirmed', 'cancelled')),
    CONSTRAINT orders_moq_check CHECK (moq > 0)
);

CREATE INDEX orders_product_id_idx ON orders (product_id);
CREATE INDEX orders_creator_id_idx ON orders (creator_id);
CREATE INDEX orders_status_deadline_idx ON orders (status, deadline);

CREATE TABLE order_participations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT order_participations_order_user_unique UNIQUE (order_id, user_id),
    CONSTRAINT order_participations_quantity_check CHECK (quantity > 0)
);

CREATE INDEX order_participations_order_id_idx ON order_participations (order_id);
CREATE INDEX order_participations_user_id_idx ON order_participations (user_id);

CREATE TABLE disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    opened_by UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT disputes_status_check CHECK (status IN ('open', 'resolved'))
);

CREATE INDEX disputes_order_id_idx ON disputes (order_id);
CREATE INDEX disputes_opened_by_idx ON disputes (opened_by);
