package store

import (
	"context"
	"fmt"

	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

func (s *Store) Notify(ctx context.Context, notice domain.Notice) error {
	if notice.Kind == "" {
		return nil
	}
	for _, userID := range notice.UserIDs {
		if _, err := s.pool.Exec(ctx, `
			INSERT INTO notifications (user_id, order_id, kind) VALUES ($1, $2, $3)
		`, userID, notice.OrderID, notice.Kind); err != nil {
			return fmt.Errorf("insert notification: %w", err)
		}
	}
	return nil
}

func (s *Store) ListNotifications(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT n.id, n.kind, n.order_id, p.name, n.created_at, n.read_at IS NOT NULL
		FROM notifications n
		JOIN orders o ON o.id = n.order_id
		JOIN products p ON p.id = o.product_id
		WHERE n.user_id = $1
		ORDER BY n.created_at DESC, n.id DESC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Notification, 0)
	for rows.Next() {
		var item domain.Notification
		if err := rows.Scan(&item.ID, &item.Kind, &item.OrderID, &item.ProductName, &item.CreatedAt, &item.Read); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	return items, nil
}

func (s *Store) MarkNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	if _, err := s.pool.Exec(ctx, `
		UPDATE notifications SET read_at = now()
		WHERE user_id = $1 AND read_at IS NULL
	`, userID); err != nil {
		return fmt.Errorf("mark notifications read: %w", err)
	}
	return nil
}
