package store

import (
	"context"
	"errors"
	"fmt"

	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Store) ListSupplierReviews(ctx context.Context, supplierID uuid.UUID) ([]domain.Review, error) {
	return s.listReviews(ctx, `
		SELECT r.id, u.business_name, r.rating, r.body, r.created_at
		FROM reviews r
		JOIN users u ON u.id = r.user_id
		JOIN suppliers s ON s.id = r.supplier_id
		WHERE r.supplier_id = $1 AND s.is_verified
		ORDER BY r.created_at DESC, r.id DESC
	`, supplierID)
}

func (s *Store) ListProductReviews(ctx context.Context, productID uuid.UUID) ([]domain.Review, error) {
	return s.listReviews(ctx, `
		SELECT r.id, u.business_name, r.rating, r.body, r.created_at
		FROM reviews r
		JOIN users u ON u.id = r.user_id
		JOIN products p ON p.id = r.product_id
		JOIN suppliers s ON s.id = p.supplier_id
		WHERE r.product_id = $1 AND s.is_verified
		ORDER BY r.created_at DESC, r.id DESC
	`, productID)
}

func (s *Store) UpsertSupplierReview(ctx context.Context, supplierID, userID uuid.UUID, rating int, body string) error {
	var own bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND supplier_id = $2)`, userID, supplierID).Scan(&own)
	if err != nil {
		return fmt.Errorf("check supplier owner: %w", err)
	}
	if own {
		return domain.ErrNotParticipant
	}
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO reviews (supplier_id, user_id, rating, body)
		SELECT s.id, $2, $3, $4
		FROM suppliers s
		WHERE s.id = $1 AND s.is_verified
		ON CONFLICT (supplier_id, user_id) WHERE supplier_id IS NOT NULL
		DO UPDATE SET rating = EXCLUDED.rating, body = EXCLUDED.body, created_at = now()
	`, supplierID, userID, rating, body)
	if err != nil {
		return fmt.Errorf("save supplier review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) UpsertProductReview(ctx context.Context, productID, userID uuid.UUID, rating int, body string) error {
	var own bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users u
			JOIN products p ON p.supplier_id = u.supplier_id
			WHERE u.id = $1 AND p.id = $2
		)
	`, userID, productID).Scan(&own)
	if err != nil {
		return fmt.Errorf("check product owner: %w", err)
	}
	if own {
		return domain.ErrNotParticipant
	}
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO reviews (product_id, user_id, rating, body)
		SELECT p.id, $2, $3, $4
		FROM products p
		JOIN suppliers s ON s.id = p.supplier_id
		WHERE p.id = $1 AND s.is_verified
		ON CONFLICT (product_id, user_id) WHERE product_id IS NOT NULL
		DO UPDATE SET rating = EXCLUDED.rating, body = EXCLUDED.body, created_at = now()
	`, productID, userID, rating, body)
	if err != nil {
		return fmt.Errorf("save product review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) listReviews(ctx context.Context, query string, id uuid.UUID) ([]domain.Review, error) {
	rows, err := s.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Review, 0)
	for rows.Next() {
		var review domain.Review
		if err := rows.Scan(&review.ID, &review.Author, &review.Rating, &review.Body, &review.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		items = append(items, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	return items, nil
}

func (s *Store) UpdateOwnedProduct(ctx context.Context, supplierID uuid.UUID, productID uuid.UUID, input domain.NewProduct) (domain.Product, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Product{}, fmt.Errorf("begin product update: %w", err)
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `
		UPDATE products
		SET name = $3, description = $4, category = $5, tags = $6, unit = $7, retail_price = $8, image_url = $9
		WHERE id = $1 AND supplier_id = $2
	`, productID, supplierID, input.Name, input.Description, input.Category, input.Tags, input.Unit, input.RetailPrice, nullableText(input.ImageURL))
	if err != nil {
		return domain.Product{}, fmt.Errorf("update product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Product{}, domain.ErrNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM price_tiers WHERE product_id = $1`, productID); err != nil {
		return domain.Product{}, fmt.Errorf("clear price tiers: %w", err)
	}
	for _, tier := range input.PriceTiers {
		if _, err := tx.Exec(ctx, `
			INSERT INTO price_tiers (product_id, min_qty, unit_price) VALUES ($1, $2, $3)
		`, productID, tier.MinQty, tier.UnitPrice); err != nil {
			return domain.Product{}, fmt.Errorf("insert price tier: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Product{}, fmt.Errorf("commit product update: %w", err)
	}
	return s.FindProduct(ctx, productID, true)
}

func (s *Store) DeleteOwnedProduct(ctx context.Context, supplierID, productID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM products WHERE id = $1 AND supplier_id = $2`, productID, supplierID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return domain.ErrProductInUse
		}
		return fmt.Errorf("delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
