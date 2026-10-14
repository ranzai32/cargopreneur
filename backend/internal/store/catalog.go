package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const productSearchExpr = `(p.name || ' ' || p.description || ' ' || array_to_string(p.tags, ' '))`

var productOrderSQL = map[string]string{
	domain.SortNewest:    "p.created_at DESC, p.id DESC",
	domain.SortName:      "p.name ASC, p.id ASC",
	domain.SortPriceAsc:  "(SELECT MIN(unit_price) FROM price_tiers pt WHERE pt.product_id = p.id) ASC NULLS LAST, p.id ASC",
	domain.SortPriceDesc: "(SELECT MIN(unit_price) FROM price_tiers pt WHERE pt.product_id = p.id) DESC NULLS LAST, p.id ASC",
}

func (s *Store) ListProducts(ctx context.Context, query domain.ProductQuery) ([]domain.Product, int, error) {
	order, ok := productOrderSQL[query.Sort]
	if !ok {
		return nil, 0, fmt.Errorf("invalid product sort")
	}
	where, args := productFilter(query)

	var total int
	countSQL := `SELECT COUNT(*) FROM products p JOIN suppliers s ON s.id = p.supplier_id WHERE ` + where
	if err := s.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	listArgs := append(append([]any{}, args...), query.PageSize, (query.Page-1)*query.PageSize)
	listSQL := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.category, p.tags, p.unit, p.retail_price::float8, p.created_at,
		       s.id, s.name, s.region, s.is_verified
		FROM products p
		JOIN suppliers s ON s.id = p.supplier_id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, where, order, len(args)+1, len(args)+2)

	rows, err := s.pool.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, product)
		ids = append(ids, product.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}

	tiers, err := s.tiersByProduct(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	for i := range products {
		if found, ok := tiers[products[i].ID]; ok {
			products[i].PriceTiers = found
		}
	}
	return products, total, nil
}

func (s *Store) FindProduct(ctx context.Context, id uuid.UUID, includeUnverified bool) (domain.Product, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT p.id, p.name, p.description, p.category, p.tags, p.unit, p.retail_price::float8, p.created_at,
		       s.id, s.name, s.region, s.is_verified
		FROM products p
		JOIN suppliers s ON s.id = p.supplier_id
		WHERE p.id = $1 AND ($2 OR s.is_verified)
	`, id, includeUnverified)
	product, err := scanProduct(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Product{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Product{}, fmt.Errorf("find product: %w", err)
	}
	tiers, err := s.tiersByProduct(ctx, []uuid.UUID{product.ID})
	if err != nil {
		return domain.Product{}, err
	}
	if found, ok := tiers[product.ID]; ok {
		product.PriceTiers = found
	}
	return product, nil
}

func (s *Store) ListSuppliers(ctx context.Context, includeUnverified bool) ([]domain.Supplier, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.id, s.name, s.description, s.region, s.is_verified, COUNT(p.id)
		FROM suppliers s
		LEFT JOIN products p ON p.supplier_id = s.id
		WHERE ($1 OR s.is_verified)
		GROUP BY s.id
		ORDER BY s.name ASC, s.id ASC
	`, includeUnverified)
	if err != nil {
		return nil, fmt.Errorf("list suppliers: %w", err)
	}
	defer rows.Close()

	suppliers := make([]domain.Supplier, 0)
	for rows.Next() {
		var supplier domain.Supplier
		if err := rows.Scan(
			&supplier.ID,
			&supplier.Name,
			&supplier.Description,
			&supplier.Region,
			&supplier.IsVerified,
			&supplier.ProductCount,
		); err != nil {
			return nil, fmt.Errorf("scan supplier: %w", err)
		}
		suppliers = append(suppliers, supplier)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list suppliers: %w", err)
	}
	return suppliers, nil
}

func (s *Store) FindSupplier(ctx context.Context, id uuid.UUID, includeUnverified bool) (domain.Supplier, error) {
	var supplier domain.Supplier
	err := s.pool.QueryRow(ctx, `
		SELECT s.id, s.name, s.description, s.region, s.is_verified,
		       (SELECT COUNT(*) FROM products p WHERE p.supplier_id = s.id)
		FROM suppliers s
		WHERE s.id = $1 AND ($2 OR s.is_verified)
	`, id, includeUnverified).Scan(
		&supplier.ID,
		&supplier.Name,
		&supplier.Description,
		&supplier.Region,
		&supplier.IsVerified,
		&supplier.ProductCount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Supplier{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Supplier{}, fmt.Errorf("find supplier: %w", err)
	}
	return supplier, nil
}

func (s *Store) ListCategories(ctx context.Context, includeUnverified bool) ([]domain.Category, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.category, COUNT(*)
		FROM products p
		JOIN suppliers s ON s.id = p.supplier_id
		WHERE ($1 OR s.is_verified)
		GROUP BY p.category
		ORDER BY lower(p.category) ASC, p.category ASC
	`, includeUnverified)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)
	for rows.Next() {
		var category domain.Category
		if err := rows.Scan(&category.Name, &category.Count); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return categories, nil
}

func (s *Store) tiersByProduct(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]domain.PriceTier, error) {
	grouped := map[uuid.UUID][]domain.PriceTier{}
	if len(ids) == 0 {
		return grouped, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT product_id, min_qty, unit_price::float8
		FROM price_tiers
		WHERE product_id = ANY($1)
		ORDER BY min_qty ASC
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("list price tiers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var productID uuid.UUID
		var tier domain.PriceTier
		if err := rows.Scan(&productID, &tier.MinQty, &tier.UnitPrice); err != nil {
			return nil, fmt.Errorf("scan price tier: %w", err)
		}
		grouped[productID] = append(grouped[productID], tier)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list price tiers: %w", err)
	}
	return grouped, nil
}

func productFilter(query domain.ProductQuery) (string, []any) {
	clauses := []string{"TRUE"}
	args := []any{}
	if !query.IncludeUnverified {
		clauses = append(clauses, "s.is_verified")
	}
	if query.Q != "" {
		args = append(args, ilikeContains(query.Q))
		clauses = append(clauses, fmt.Sprintf("%s ILIKE $%d ESCAPE '\\'", productSearchExpr, len(args)))
	}
	if query.Category != "" {
		args = append(args, escapeLike(query.Category))
		clauses = append(clauses, fmt.Sprintf("p.category ILIKE $%d ESCAPE '\\'", len(args)))
	}
	if query.SupplierID != nil {
		args = append(args, *query.SupplierID)
		clauses = append(clauses, fmt.Sprintf("p.supplier_id = $%d", len(args)))
	}
	if query.Region != "" {
		args = append(args, escapeLike(query.Region))
		clauses = append(clauses, fmt.Sprintf("s.region ILIKE $%d ESCAPE '\\'", len(args)))
	}
	return strings.Join(clauses, " AND "), args
}

func ilikeContains(value string) string {
	return "%" + escapeLike(value) + "%"
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	value = strings.ReplaceAll(value, `_`, `\_`)
	return value
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProduct(row rowScanner) (domain.Product, error) {
	var product domain.Product
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Category,
		&product.Tags,
		&product.Unit,
		&product.RetailPrice,
		&product.CreatedAt,
		&product.Supplier.ID,
		&product.Supplier.Name,
		&product.Supplier.Region,
		&product.Supplier.IsVerified,
	)
	if product.Tags == nil {
		product.Tags = []string{}
	}
	if product.PriceTiers == nil {
		product.PriceTiers = []domain.PriceTier{}
	}
	return product, err
}
