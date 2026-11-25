package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

type OrderStore interface {
	ListOrders(ctx context.Context, status string, viewerID *uuid.UUID) ([]domain.Order, error)
	FindOrder(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (domain.Order, error)
	CreateOrder(ctx context.Context, input domain.NewOrder) (domain.Order, error)
	JoinOrder(ctx context.Context, orderID, userID uuid.UUID, quantity int) (domain.Order, error)
	PayOrder(ctx context.Context, orderID, userID uuid.UUID, shippingAddress string) (domain.Order, error)
}

type OrderHandler struct {
	orders   OrderStore
	users    UserLookup
	tokens   *auth.Tokens
	notifier domain.Notifier
}

func (h OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	status, message := domain.ParseOrderStatus(r.URL.Query().Get("status"))
	if message != "" {
		writeError(w, http.StatusBadRequest, "Invalid query", map[string]string{"status": message})
		return
	}
	ctx := domain.WithNoticeLog(r.Context())
	orders, err := h.orders.ListOrders(ctx, status, h.viewerID(r))
	h.dispatchNotices(ctx)
	if err != nil {
		slog.Error("list orders", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	encoded := make([]domain.OrderJSON, 0, len(orders))
	for _, order := range orders {
		encoded = append(encoded, order.JSON())
	}
	writeJSON(w, http.StatusOK, encoded)
}

func (h OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	ctx := domain.WithNoticeLog(r.Context())
	order, err := h.orders.FindOrder(ctx, id, h.viewerID(r))
	h.dispatchNotices(ctx)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Order not found", nil)
		return
	}
	if err != nil {
		slog.Error("find order", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, order.JSON())
}

func (h OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	var body domain.CreateOrderBody
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	input, fieldErrors := domain.ParseCreateOrder(body, time.Now())
	if len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return
	}
	input.CreatorID = user.ID
	input.IncludeUnverified = user.Role == domain.RoleAdmin
	ctx := domain.WithNoticeLog(r.Context())
	order, err := h.orders.CreateOrder(ctx, input)
	h.dispatchNotices(ctx)
	if errors.Is(err, domain.ErrProductUnavailable) {
		writeError(w, http.StatusBadRequest, "Choose a product from the catalog", map[string]string{
			"productId": "Choose a product from the catalog",
		})
		return
	}
	if err != nil {
		slog.Error("create order", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusCreated, order.JSON())
}

func (h OrderHandler) Join(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body domain.JoinOrderBody
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	quantity, fieldErrors := domain.ParseJoinOrder(body)
	if len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return
	}
	ctx := domain.WithNoticeLog(r.Context())
	order, err := h.orders.JoinOrder(ctx, id, user.ID, quantity)
	h.dispatchNotices(ctx)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Order not found", nil)
		return
	}
	if errors.Is(err, domain.ErrOrderClosed) {
		writeError(w, http.StatusBadRequest, "This order is no longer open", nil)
		return
	}
	if errors.Is(err, domain.ErrAlreadyJoined) {
		writeError(w, http.StatusConflict, "You already joined this order", map[string]string{
			"quantity": "You already joined this order",
		})
		return
	}
	if err != nil {
		slog.Error("join order", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, order.JSON())
}

func (h OrderHandler) Pay(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body domain.PayOrderBody
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	address, fieldErrors := domain.ParsePayOrder(body)
	if len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return
	}
	ctx := domain.WithNoticeLog(r.Context())
	order, err := h.orders.PayOrder(ctx, id, user.ID, address)
	h.dispatchNotices(ctx)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Order not found", nil)
		return
	}
	if errors.Is(err, domain.ErrNotReadyForPayment) {
		writeError(w, http.StatusBadRequest, "This order is not ready for payment", nil)
		return
	}
	if errors.Is(err, domain.ErrNotParticipant) {
		writeError(w, http.StatusForbidden, "You are not part of this order", nil)
		return
	}
	if errors.Is(err, domain.ErrAlreadyPaid) {
		writeError(w, http.StatusConflict, "This order is already paid", nil)
		return
	}
	if err != nil {
		slog.Error("pay order", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, order.JSON())
}

func (h OrderHandler) dispatchNotices(ctx context.Context) {
	notices := domain.TakeNotices(ctx)
	if h.notifier == nil {
		return
	}
	for _, notice := range notices {
		if err := h.notifier.Notify(ctx, notice); err != nil {
			slog.Error("notify order", "error", err, "kind", notice.Kind, "orderId", notice.OrderID)
		}
	}
}

func (h OrderHandler) viewerID(r *http.Request) *uuid.UUID {
	if h.users == nil || h.tokens == nil {
		return nil
	}
	user, ok := optionalUser(r, h.users, h.tokens)
	if !ok {
		return nil
	}
	id := user.ID
	return &id
}
