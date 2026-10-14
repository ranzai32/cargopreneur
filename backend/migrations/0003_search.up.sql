CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX products_search_trgm_idx
    ON products USING gin ((name || ' ' || description || ' ' || array_to_string(tags, ' ')) gin_trgm_ops);

CREATE INDEX products_category_lower_idx ON products (lower(category));

CREATE INDEX suppliers_region_lower_idx ON suppliers (lower(region));

CREATE INDEX products_created_at_idx ON products (created_at DESC);
