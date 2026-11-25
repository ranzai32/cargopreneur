package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	OrderOpen      = "open"
	OrderConfirmed = "confirmed"
	OrderCancelled = "cancelled"
)

var (
	ErrProductUnavailable  = errors.New("product unavailable")
	ErrOrderClosed         = errors.New("order closed")
	ErrAlreadyJoined       = errors.New("already joined")
	ErrOrderNotCancellable = errors.New("order not cancellable")
)

type Order struct {
	ID               uuid.UUID
	ProductID        uuid.UUID
	Status           string
	MOQ              int
	Deadline         time.Time
	CreatedAt        time.Time
	CollectedQty     int
	ParticipantCount int
	YourQuantity     *int
	Product          Product
}

type OrderJSON struct {
	ID               uuid.UUID   `json:"id"`
	Status           string      `json:"status"`
	MOQ              int         `json:"moq"`
	Deadline         time.Time   `json:"deadline"`
	CreatedAt        time.Time   `json:"createdAt"`
	CollectedQty     int         `json:"collectedQty"`
	ParticipantCount int         `json:"participantCount"`
	YourQuantity     *int        `json:"yourQuantity"`
	Product          ProductJSON `json:"product"`
}

type NewOrder struct {
	ProductID         uuid.UUID
	CreatorID         uuid.UUID
	MOQ               int
	Deadline          time.Time
	Quantity          int
	IncludeUnverified bool
}

type CreateOrderBody struct {
	ProductID string          `json:"productId"`
	MOQ       json.RawMessage `json:"moq"`
	Deadline  string          `json:"deadline"`
	Quantity  json.RawMessage `json:"quantity"`
}

type JoinOrderBody struct {
	Quantity json.RawMessage `json:"quantity"`
}

func (o Order) JSON() OrderJSON {
	return OrderJSON{
		ID:               o.ID,
		Status:           o.Status,
		MOQ:              o.MOQ,
		Deadline:         o.Deadline.UTC(),
		CreatedAt:        o.CreatedAt.UTC(),
		CollectedQty:     o.CollectedQty,
		ParticipantCount: o.ParticipantCount,
		YourQuantity:     o.YourQuantity,
		Product:          o.Product.JSON(),
	}
}

func ParseOrderStatus(raw string) (string, string) {
	switch strings.TrimSpace(raw) {
	case "", "all":
		return "", ""
	case OrderOpen, OrderConfirmed, OrderCancelled:
		return strings.TrimSpace(raw), ""
	default:
		return "", "Use open, confirmed, cancelled, or all"
	}
}

func ParseCreateOrder(body CreateOrderBody, now time.Time) (NewOrder, map[string]string) {
	errs := map[string]string{}
	var input NewOrder
	productID, err := uuid.Parse(strings.TrimSpace(body.ProductID))
	if err != nil {
		errs["productId"] = "Must be a valid id"
	} else {
		input.ProductID = productID
	}
	moq, ok := parsePositiveInt(body.MOQ)
	if !ok {
		errs["moq"] = "Enter a minimum of at least 1"
	} else {
		input.MOQ = moq
	}
	deadline, message := parseDeadline(body.Deadline, now)
	if message != "" {
		errs["deadline"] = message
	} else {
		input.Deadline = deadline
	}
	quantity, ok := parsePositiveInt(body.Quantity)
	if !ok {
		errs["quantity"] = "Enter the quantity you need"
	} else {
		input.Quantity = quantity
	}
	if len(errs) > 0 {
		return NewOrder{}, errs
	}
	return input, nil
}

func ParseJoinOrder(body JoinOrderBody) (int, map[string]string) {
	quantity, ok := parsePositiveInt(body.Quantity)
	if !ok {
		return 0, map[string]string{"quantity": "Enter a quantity of at least 1"}
	}
	return quantity, nil
}

func NextOrderStatus(status string, deadline time.Time, collected, moq int, now time.Time) (string, bool) {
	if status != OrderOpen || !now.After(deadline) {
		return status, false
	}
	if collected >= moq {
		return OrderConfirmed, true
	}
	return OrderCancelled, true
}

func parseDeadline(raw string, now time.Time) (time.Time, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, "Choose a deadline"
	}
	if deadline, err := time.Parse(time.RFC3339, raw); err == nil {
		return acceptDeadline(deadline, now)
	}
	if len(raw) == 10 {
		if day, err := time.Parse("2006-01-02", raw); err == nil {
			deadline := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, time.UTC)
			return acceptDeadline(deadline, now)
		}
	}
	return time.Time{}, "Enter a valid date"
}

func acceptDeadline(deadline, now time.Time) (time.Time, string) {
	if deadline.Before(now) {
		return time.Time{}, "Choose a deadline that has not passed"
	}
	return deadline, ""
}

func parsePositiveInt(raw json.RawMessage) (int, bool) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return 0, false
	}
	if trimmed[0] == '-' || bytes.ContainsAny(trimmed, ".eE+") {
		return 0, false
	}
	value, err := strconv.Atoi(string(trimmed))
	if err != nil || value < 1 {
		return 0, false
	}
	return value, true
}
