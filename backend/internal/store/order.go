package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type rowQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type orderDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type orderHead struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	Status    string
	MOQ       int
	Deadline  time.Time
	CreatedAt time.Time
}

const orderSelectSQL = `
	SELECT o.id, o.product_id, o.status, o.moq, o.deadline, o.created_at,
	       COALESCE(SUM(p.quantity), 0)::int,
	       COUNT(p.id)::int,
	       MAX(CASE WHEN $2::uuid IS NOT NULL AND p.user_id = $2::uuid THEN p.quantity END),
	       MAX(CASE WHEN $2::uuid IS NOT NULL AND p.user_id = $2::uuid THEN p.payment_status END),
	       MAX(CASE WHEN $2::uuid IS NOT NULL AND p.user_id = $2::uuid THEN p.delivery_status END),
	       COALESCE(BOOL_AND(p.payment_status = 'paid'), false),
	       pr.id, pr.name, pr.description, pr.category, pr.tags, pr.unit, pr.retail_price::float8, pr.created_at, pr.image_url,
	       s.id, s.name, s.region, s.is_verified
	FROM orders o
	JOIN products pr ON pr.id = o.product_id
	JOIN suppliers s ON s.id = pr.supplier_id
	LEFT JOIN order_participations p ON p.order_id = o.id
	WHERE %s
	GROUP BY o.id, pr.id, s.id
	ORDER BY o.created_at DESC, o.id::text DESC
`

func (s *Store) ListOrders(ctx context.Context, status string, viewerID *uuid.UUID) ([]domain.Order, error) {
	var orders []domain.Order
	err := s.withOrderTx(ctx, func(tx pgx.Tx) ([]domain.Notice, error) {
		notices, err := settleExpired(ctx, tx, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		loaded, err := selectOrders(ctx, tx, status, viewerID, nil)
		if err != nil {
			return nil, err
		}
		orders = loaded
		return notices, nil
	})
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *Store) FindOrder(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (domain.Order, error) {
	var order domain.Order
	err := s.withOrderTx(ctx, func(tx pgx.Tx) ([]domain.Notice, error) {
		head, err := lockOrder(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		collected, userIDs, err := participationSummary(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		notices, err := settleHead(ctx, tx, &head, collected, userIDs, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		loaded, err := selectOrders(ctx, tx, "", viewerID, &id)
		if err != nil {
			return nil, err
		}
		if len(loaded) == 0 {
			return nil, domain.ErrNotFound
		}
		order = loaded[0]
		return notices, nil
	})
	if err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

func (s *Store) CreateOrder(ctx context.Context, input domain.NewOrder) (domain.Order, error) {
	var order domain.Order
	err := s.withOrderTx(ctx, func(tx pgx.Tx) ([]domain.Notice, error) {
		if err := lockVisibleProduct(ctx, tx, input.ProductID, input.IncludeUnverified); err != nil {
			return nil, err
		}
		status := domain.OrderOpen
		if input.Quantity >= input.MOQ {
			status = domain.OrderConfirmed
		}
		var orderID uuid.UUID
		err := tx.QueryRow(ctx, `
			INSERT INTO orders (product_id, creator_id, status, moq, deadline)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, input.ProductID, input.CreatorID, status, input.MOQ, input.Deadline).Scan(&orderID)
		if err != nil {
			return nil, fmt.Errorf("insert order: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO order_participations (order_id, user_id, quantity)
			VALUES ($1, $2, $3)
		`, orderID, input.CreatorID, input.Quantity); err != nil {
			return nil, fmt.Errorf("insert creator participation: %w", err)
		}
		loaded, err := selectOrders(ctx, tx, "", &input.CreatorID, &orderID)
		if err != nil {
			return nil, err
		}
		if len(loaded) == 0 {
			return nil, domain.ErrNotFound
		}
		order = loaded[0]
		if status != domain.OrderConfirmed {
			return nil, nil
		}
		return []domain.Notice{{
			Kind:    domain.NoticeOrderConfirmed,
			OrderID: orderID,
			UserIDs: []uuid.UUID{input.CreatorID},
		}}, nil
	})
	if err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

func (s *Store) JoinOrder(ctx context.Context, orderID, userID uuid.UUID, quantity int) (domain.Order, error) {
	var order domain.Order
	var closed error
	err := s.withOrderTx(ctx, func(tx pgx.Tx) ([]domain.Notice, error) {
		head, err := lockOrder(ctx, tx, orderID)
		if err != nil {
			return nil, err
		}
		collected, userIDs, err := participationSummary(ctx, tx, orderID)
		if err != nil {
			return nil, err
		}
		notices, err := settleHead(ctx, tx, &head, collected, userIDs, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		if head.Status != domain.OrderOpen {
			closed = domain.ErrOrderClosed
			return notices, nil
		}
		for _, existing := range userIDs {
			if existing == userID {
				return nil, domain.ErrAlreadyJoined
			}
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO order_participations (order_id, user_id, quantity)
			VALUES ($1, $2, $3)
		`, orderID, userID, quantity); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return nil, domain.ErrAlreadyJoined
			}
			return nil, fmt.Errorf("insert participation: %w", err)
		}
		userIDs = append(userIDs, userID)
		collected += quantity
		if collected >= head.MOQ {
			if err := updateOrderStatus(ctx, tx, orderID, domain.OrderConfirmed); err != nil {
				return nil, err
			}
			notices = append(notices, domain.Notice{
				Kind:    domain.NoticeOrderConfirmed,
				OrderID: orderID,
				UserIDs: userIDs,
			})
		}
		loaded, err := selectOrders(ctx, tx, "", &userID, &orderID)
		if err != nil {
			return nil, err
		}
		if len(loaded) == 0 {
			return nil, domain.ErrNotFound
		}
		order = loaded[0]
		return notices, nil
	})
	if err != nil {
		return domain.Order{}, err
	}
	if closed != nil {
		return domain.Order{}, closed
	}
	return order, nil
}

func (s *Store) PayOrder(ctx context.Context, orderID, userID uuid.UUID, shippingAddress string) (domain.Order, error) {
	var order domain.Order
	var rejected error
	address := strings.TrimSpace(shippingAddress)
	err := s.withOrderTx(ctx, func(tx pgx.Tx) ([]domain.Notice, error) {
		head, err := lockOrder(ctx, tx, orderID)
		if err != nil {
			return nil, err
		}
		collected, userIDs, err := participationSummary(ctx, tx, orderID)
		if err != nil {
			return nil, err
		}
		notices, err := settleHead(ctx, tx, &head, collected, userIDs, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		if head.Status != domain.OrderConfirmed {
			rejected = domain.ErrNotReadyForPayment
			return notices, nil
		}
		var paymentStatus string
		err = tx.QueryRow(ctx, `
			SELECT payment_status
			FROM order_participations
			WHERE order_id = $1 AND user_id = $2
			FOR UPDATE
		`, orderID, userID).Scan(&paymentStatus)
		if errors.Is(err, pgx.ErrNoRows) {
			rejected = domain.ErrNotParticipant
			return notices, nil
		}
		if err != nil {
			return nil, fmt.Errorf("lock participation: %w", err)
		}
		if paymentStatus == domain.PaymentPaid {
			rejected = domain.ErrAlreadyPaid
			return notices, nil
		}
		if _, err := tx.Exec(ctx, `
			UPDATE order_participations
			SET payment_status = $3, paid_at = $4, shipping_address = $5
			WHERE order_id = $1 AND user_id = $2
		`, orderID, userID, domain.PaymentPaid, time.Now().UTC(), address); err != nil {
			return nil, fmt.Errorf("mark participation paid: %w", err)
		}
		var unpaid int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*)::int
			FROM order_participations
			WHERE order_id = $1 AND payment_status <> $2
		`, orderID, domain.PaymentPaid).Scan(&unpaid); err != nil {
			return nil, fmt.Errorf("count unpaid participations: %w", err)
		}
		if unpaid == 0 {
			if err := updateOrderStatus(ctx, tx, orderID, domain.OrderPaid); err != nil {
				return nil, err
			}
		}
		loaded, err := selectOrders(ctx, tx, "", &userID, &orderID)
		if err != nil {
			return nil, err
		}
		if len(loaded) == 0 {
			return nil, domain.ErrNotFound
		}
		order = loaded[0]
		notices = append(notices, domain.Notice{
			Kind:    domain.NoticePaymentReceived,
			OrderID: orderID,
			UserIDs: []uuid.UUID{userID},
		})
		return notices, nil
	})
	if err != nil {
		return domain.Order{}, err
	}
	if rejected != nil {
		return domain.Order{}, rejected
	}
	return order, nil
}

func (s *Store) CancelOrder(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (domain.Order, error) {
	var order domain.Order
	var blocked error
	err := s.withOrderTx(ctx, func(tx pgx.Tx) ([]domain.Notice, error) {
		head, err := lockOrder(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		collected, userIDs, err := participationSummary(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		notices, err := settleHead(ctx, tx, &head, collected, userIDs, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		if head.Status != domain.OrderOpen && head.Status != domain.OrderConfirmed {
			blocked = domain.ErrOrderNotCancellable
			return notices, nil
		}
		if err := updateOrderStatus(ctx, tx, id, domain.OrderCancelled); err != nil {
			return nil, err
		}
		notices = append(notices, domain.Notice{
			Kind:    domain.NoticeOrderCancelled,
			OrderID: id,
			UserIDs: userIDs,
		})
		loaded, err := selectOrders(ctx, tx, "", viewerID, &id)
		if err != nil {
			return nil, err
		}
		if len(loaded) == 0 {
			return nil, domain.ErrNotFound
		}
		order = loaded[0]
		return notices, nil
	})
	if err != nil {
		return domain.Order{}, err
	}
	if blocked != nil {
		return domain.Order{}, blocked
	}
	return order, nil
}

func (s *Store) ShipParticipation(ctx context.Context, participationID uuid.UUID, viewerID *uuid.UUID) (domain.Order, error) {
	return s.updateDelivery(ctx, participationID, viewerID, domain.DeliveryShipped)
}

func (s *Store) DeliverParticipation(ctx context.Context, participationID uuid.UUID, viewerID *uuid.UUID) (domain.Order, error) {
	return s.updateDelivery(ctx, participationID, viewerID, domain.DeliveryDelivered)
}

func (s *Store) updateDelivery(ctx context.Context, participationID uuid.UUID, viewerID *uuid.UUID, next string) (domain.Order, error) {
	var order domain.Order
	var rejected error
	err := s.withOrderTx(ctx, func(tx pgx.Tx) ([]domain.Notice, error) {
		var orderID uuid.UUID
		err := tx.QueryRow(ctx, `
			SELECT order_id
			FROM order_participations
			WHERE id = $1
		`, participationID).Scan(&orderID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("find participation: %w", err)
		}
		head, err := lockOrder(ctx, tx, orderID)
		if err != nil {
			return nil, err
		}
		collected, userIDs, err := participationSummary(ctx, tx, orderID)
		if err != nil {
			return nil, err
		}
		notices, err := settleHead(ctx, tx, &head, collected, userIDs, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		var userID uuid.UUID
		var paymentStatus string
		var deliveryStatus string
		err = tx.QueryRow(ctx, `
			SELECT user_id, payment_status, delivery_status
			FROM order_participations
			WHERE id = $1
			FOR UPDATE
		`, participationID).Scan(&userID, &paymentStatus, &deliveryStatus)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("lock participation: %w", err)
		}
		if !domain.ShipmentUpdateAllowed(head.Status, paymentStatus, deliveryStatus, next) {
			rejected = domain.ErrShipmentNotUpdatable
			return notices, nil
		}
		now := time.Now().UTC()
		if next == domain.DeliveryShipped {
			if _, err := tx.Exec(ctx, `
				UPDATE order_participations
				SET delivery_status = $2, shipped_at = $3
				WHERE id = $1
			`, participationID, domain.DeliveryShipped, now); err != nil {
				return nil, fmt.Errorf("mark participation shipped: %w", err)
			}
			notices = append(notices, domain.Notice{
				Kind:    domain.NoticeShipmentSent,
				OrderID: orderID,
				UserIDs: []uuid.UUID{userID},
			})
		} else {
			if _, err := tx.Exec(ctx, `
				UPDATE order_participations
				SET delivery_status = $2, shipped_at = COALESCE(shipped_at, $3), delivered_at = $3
				WHERE id = $1
			`, participationID, domain.DeliveryDelivered, now); err != nil {
				return nil, fmt.Errorf("mark participation delivered: %w", err)
			}
			var remaining int
			if err := tx.QueryRow(ctx, `
				SELECT COUNT(*)::int
				FROM order_participations
				WHERE order_id = $1 AND delivery_status <> $2
			`, orderID, domain.DeliveryDelivered).Scan(&remaining); err != nil {
				return nil, fmt.Errorf("count undelivered participations: %w", err)
			}
			if remaining == 0 {
				if err := updateOrderStatus(ctx, tx, orderID, domain.OrderCompleted); err != nil {
					return nil, err
				}
			}
			notices = append(notices, domain.Notice{
				Kind:    domain.NoticeShipmentDelivered,
				OrderID: orderID,
				UserIDs: []uuid.UUID{userID},
			})
		}
		loaded, err := selectOrders(ctx, tx, "", viewerID, &orderID)
		if err != nil {
			return nil, err
		}
		if len(loaded) == 0 {
			return nil, domain.ErrNotFound
		}
		order = loaded[0]
		return notices, nil
	})
	if err != nil {
		return domain.Order{}, err
	}
	if rejected != nil {
		return domain.Order{}, rejected
	}
	return order, nil
}

func (s *Store) withOrderTx(ctx context.Context, fn func(tx pgx.Tx) ([]domain.Notice, error)) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin order transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	notices, err := fn(tx)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit order transaction: %w", err)
	}
	for _, notice := range notices {
		domain.AddNotice(ctx, notice)
	}
	return nil
}

func lockVisibleProduct(ctx context.Context, db orderDB, productID uuid.UUID, includeUnverified bool) error {
	var id uuid.UUID
	err := db.QueryRow(ctx, `
		SELECT p.id
		FROM products p
		JOIN suppliers s ON s.id = p.supplier_id
		WHERE p.id = $1 AND ($2 OR s.is_verified)
		FOR UPDATE OF p
	`, productID, includeUnverified).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrProductUnavailable
	}
	if err != nil {
		return fmt.Errorf("lock product: %w", err)
	}
	return nil
}

func lockOrder(ctx context.Context, db orderDB, id uuid.UUID) (orderHead, error) {
	var order orderHead
	err := db.QueryRow(ctx, `
		SELECT id, product_id, status, moq, deadline, created_at
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&order.ID, &order.ProductID, &order.Status, &order.MOQ, &order.Deadline, &order.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return orderHead{}, domain.ErrNotFound
	}
	if err != nil {
		return orderHead{}, fmt.Errorf("lock order: %w", err)
	}
	return order, nil
}

func settleExpired(ctx context.Context, db orderDB, now time.Time) ([]domain.Notice, error) {
	rows, err := db.Query(ctx, `
		SELECT id, moq
		FROM orders
		WHERE status = 'open' AND deadline < $1
		ORDER BY id
		FOR UPDATE
	`, now)
	if err != nil {
		return nil, fmt.Errorf("lock expired orders: %w", err)
	}
	heads := make([]orderHead, 0)
	for rows.Next() {
		var head orderHead
		if err := rows.Scan(&head.ID, &head.MOQ); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan expired order: %w", err)
		}
		heads = append(heads, head)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("lock expired orders: %w", err)
	}
	rows.Close()

	notices := make([]domain.Notice, 0, len(heads))
	for _, head := range heads {
		collected, userIDs, err := participationSummary(ctx, db, head.ID)
		if err != nil {
			return nil, err
		}
		status := domain.OrderCancelled
		if collected >= head.MOQ {
			status = domain.OrderConfirmed
		}
		if err := updateOrderStatus(ctx, db, head.ID, status); err != nil {
			return nil, err
		}
		notices = append(notices, domain.Notice{
			Kind:    domain.NoticeKindForStatus(status),
			OrderID: head.ID,
			UserIDs: userIDs,
		})
	}
	return notices, nil
}

func settleHead(ctx context.Context, db orderDB, order *orderHead, collected int, userIDs []uuid.UUID, now time.Time) ([]domain.Notice, error) {
	next, changed := domain.NextOrderStatus(order.Status, order.Deadline, collected, order.MOQ, now)
	if !changed {
		return nil, nil
	}
	if err := updateOrderStatus(ctx, db, order.ID, next); err != nil {
		return nil, err
	}
	order.Status = next
	return []domain.Notice{{
		Kind:    domain.NoticeKindForStatus(next),
		OrderID: order.ID,
		UserIDs: userIDs,
	}}, nil
}

func updateOrderStatus(ctx context.Context, db orderDB, id uuid.UUID, status string) error {
	if _, err := db.Exec(ctx, `UPDATE orders SET status = $2 WHERE id = $1`, id, status); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

func participationSummary(ctx context.Context, db orderDB, orderID uuid.UUID) (int, []uuid.UUID, error) {
	var collected int
	var userIDs []uuid.UUID
	err := db.QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity), 0)::int,
		       COALESCE(array_agg(user_id ORDER BY created_at, id), '{}')
		FROM order_participations
		WHERE order_id = $1
	`, orderID).Scan(&collected, &userIDs)
	if err != nil {
		return 0, nil, fmt.Errorf("summarize participations: %w", err)
	}
	if userIDs == nil {
		userIDs = []uuid.UUID{}
	}
	return collected, userIDs, nil
}

func selectOrders(ctx context.Context, db orderDB, status string, viewerID *uuid.UUID, orderID *uuid.UUID) ([]domain.Order, error) {
	args := []any{status, nullableUUID(viewerID)}
	filters := []string{"($1 = '' OR o.status = $1)"}
	if orderID != nil {
		args = append(args, *orderID)
		filters = append(filters, fmt.Sprintf("o.id = $%d", len(args)))
	}
	rows, err := db.Query(ctx, fmt.Sprintf(orderSelectSQL, strings.Join(filters, " AND ")), args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	withTiers, err := attachOrderTiers(ctx, db, orders)
	if err != nil {
		return nil, err
	}
	return attachParticipants(ctx, db, withTiers)
}

func attachParticipants(ctx context.Context, db orderDB, orders []domain.Order) ([]domain.Order, error) {
	ids := make([]uuid.UUID, len(orders))
	index := make(map[uuid.UUID]int, len(orders))
	for i := range orders {
		ids[i] = orders[i].ID
		index[orders[i].ID] = i
		orders[i].Participants = []domain.Participation{}
	}
	if len(ids) == 0 {
		return orders, nil
	}
	rows, err := db.Query(ctx, `
		SELECT id, order_id, quantity, shipping_address, payment_status, delivery_status
		FROM order_participations
		WHERE order_id = ANY($1)
		ORDER BY created_at ASC, id ASC
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("list participations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var orderID uuid.UUID
		var part domain.Participation
		if err := rows.Scan(&part.ID, &orderID, &part.Quantity, &part.ShippingAddress, &part.PaymentStatus, &part.DeliveryStatus); err != nil {
			return nil, fmt.Errorf("scan participation: %w", err)
		}
		slot, ok := index[orderID]
		if !ok {
			continue
		}
		orders[slot].Participants = append(orders[slot].Participants, part)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list participations: %w", err)
	}
	return orders, nil
}

func attachOrderTiers(ctx context.Context, db orderDB, orders []domain.Order) ([]domain.Order, error) {
	ids := make([]uuid.UUID, len(orders))
	for i := range orders {
		ids[i] = orders[i].Product.ID
		if orders[i].Product.Tags == nil {
			orders[i].Product.Tags = []string{}
		}
		if orders[i].Product.PriceTiers == nil {
			orders[i].Product.PriceTiers = []domain.PriceTier{}
		}
	}
	tiers, err := loadPriceTiers(ctx, db, ids)
	if err != nil {
		return nil, err
	}
	for i := range orders {
		if found, ok := tiers[orders[i].Product.ID]; ok {
			orders[i].Product.PriceTiers = found
		}
	}
	return orders, nil
}

func scanOrder(row rowScanner) (domain.Order, error) {
	var order domain.Order
	err := row.Scan(
		&order.ID,
		&order.ProductID,
		&order.Status,
		&order.MOQ,
		&order.Deadline,
		&order.CreatedAt,
		&order.CollectedQty,
		&order.ParticipantCount,
		&order.YourQuantity,
		&order.YourPaymentStatus,
		&order.YourDeliveryStatus,
		&order.AllPaid,
		&order.Product.ID,
		&order.Product.Name,
		&order.Product.Description,
		&order.Product.Category,
		&order.Product.Tags,
		&order.Product.Unit,
		&order.Product.RetailPrice,
		&order.Product.CreatedAt,
		&order.Product.ImageURL,
		&order.Product.Supplier.ID,
		&order.Product.Supplier.Name,
		&order.Product.Supplier.Region,
		&order.Product.Supplier.IsVerified,
	)
	if order.Product.Tags == nil {
		order.Product.Tags = []string{}
	}
	if order.Product.PriceTiers == nil {
		order.Product.PriceTiers = []domain.PriceTier{}
	}
	return order, err
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}
