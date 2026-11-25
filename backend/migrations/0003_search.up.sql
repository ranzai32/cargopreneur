CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE FUNCTION immutable_array_to_string(items text[], separator text)
RETURNS text
LANGUAGE sql
IMMUTABLE
PARALLEL SAFE
RETURN array_to_string(items, separator);

CREATE INDEX products_search_trgm_idx
    ON products USING gin ((name || ' ' || description || ' ' || immutable_array_to_string(tags, ' ')) gin_trgm_ops);

CREATE INDEX products_category_lower_idx ON products (lower(category));

CREATE INDEX suppliers_region_lower_idx ON suppliers (lower(region));

CREATE INDEX products_created_at_idx ON products (created_at DESC);
