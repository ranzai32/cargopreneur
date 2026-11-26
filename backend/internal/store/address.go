package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListAddresses(ctx context.Context, userID uuid.UUID) ([]domain.SavedAddress, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, label, line, updated_at
		FROM saved_addresses
		WHERE user_id = $1
		ORDER BY updated_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	defer rows.Close()
	items := []domain.SavedAddress{}
	for rows.Next() {
		item, err := scanAddress(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	return items, nil
}

func (s *Store) CreateAddress(ctx context.Context, userID uuid.UUID, label, line string) (domain.SavedAddress, error) {
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM saved_addresses WHERE user_id = $1`, userID).Scan(&count); err != nil {
		return domain.SavedAddress{}, fmt.Errorf("count addresses: %w", err)
	}
	if count >= 12 {
		return domain.SavedAddress{}, domain.ErrAddressLimit
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO saved_addresses (user_id, label, line)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, label, line, updated_at
	`, userID, label, line)
	return scanAddress(row)
}

func (s *Store) UpdateAddress(ctx context.Context, userID, id uuid.UUID, label, line string) (domain.SavedAddress, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE saved_addresses
		SET label = $3, line = $4, updated_at = $5
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, label, line, updated_at
	`, id, userID, label, line, time.Now().UTC())
	item, err := scanAddress(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SavedAddress{}, domain.ErrNotFound
	}
	return item, err
}

func (s *Store) DeleteAddress(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM saved_addresses WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete address: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanAddress(row rowScanner) (domain.SavedAddress, error) {
	var item domain.SavedAddress
	err := row.Scan(&item.ID, &item.UserID, &item.Label, &item.Line, &item.UpdatedAt)
	if err != nil {
		return domain.SavedAddress{}, err
	}
	return item, nil
}
