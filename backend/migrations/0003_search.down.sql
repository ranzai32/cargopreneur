DROP INDEX IF EXISTS products_created_at_idx;
DROP INDEX IF EXISTS suppliers_region_lower_idx;
DROP INDEX IF EXISTS products_category_lower_idx;
DROP INDEX IF EXISTS products_search_trgm_idx;

DROP FUNCTION IF EXISTS immutable_array_to_string(text[], text);

DROP EXTENSION IF EXISTS pg_trgm;
