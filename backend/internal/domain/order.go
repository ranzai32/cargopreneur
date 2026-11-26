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
	OrderPaid      = "paid"
	OrderCompleted = "completed"
	OrderCancelled = "cancelled"

	PaymentUnpaid = "unpaid"
	PaymentPaid   = "paid"

	DeliveryPending   = "pending"
	DeliveryShipped   = "shipped"
	DeliveryDelivered = "delivered"
)

var (
	ErrProductUnavailable   = errors.New("product unavailable")
	ErrOrderClosed          = errors.New("order closed")
	ErrAlreadyJoined        = errors.New("already joined")
	ErrNotReadyForPayment   = errors.New("order not ready for payment")
	ErrNotParticipant       = errors.New("not a participant")
	ErrAlreadyPaid          = errors.New("already paid")
	ErrUnknownDelivery      = errors.New("unknown delivery method")
	ErrAddressLimit         = errors.New("address limit reached")
	ErrOrderNotCancellable  = errors.New("order not cancellable")
	ErrShipmentNotUpdatable = errors.New("shipment cannot be updated")
	ErrDisputeNotAllowed    = errors.New("dispute not allowed")
	ErrDisputeOpen          = errors.New("dispute already open")
	ErrDisputeClosed        = errors.New("dispute already resolved")
	ErrProductInUse         = errors.New("product in use")
)

type Order struct {
	ID                 uuid.UUID
	ProductID          uuid.UUID
	Status             string
	MOQ                int
	Deadline           time.Time
	CreatedAt          time.Time
	CollectedQty       int
	ParticipantCount   int
	YourQuantity       *int
	YourPaymentStatus  *string
	YourDeliveryStatus *string
	YourDeliveryMethod *string
	YourShippingFee    *float64
	AllPaid            bool
	Participants       []Participation
	Disputes           []Dispute
	Product            Product
}

type Participation struct {
	ID              uuid.UUID
	Quantity        int
	ShippingAddress string
	DeliveryMethod  string
	ShippingFee     float64
	PaymentStatus   string
	DeliveryStatus  string
}

type ParticipationJSON struct {
	ID              uuid.UUID `json:"id"`
	Quantity        int       `json:"quantity"`
	ShippingAddress string    `json:"shippingAddress"`
	DeliveryMethod  string    `json:"deliveryMethod"`
	ShippingFee     float64   `json:"shippingFee"`
	PaymentStatus   string    `json:"paymentStatus"`
	DeliveryStatus  string    `json:"deliveryStatus"`
}

type OrderJSON struct {
	ID                 uuid.UUID           `json:"id"`
	Status             string              `json:"status"`
	MOQ                int                 `json:"moq"`
	Deadline           time.Time           `json:"deadline"`
	CreatedAt          time.Time           `json:"createdAt"`
	CollectedQty       int                 `json:"collectedQty"`
	ParticipantCount   int                 `json:"participantCount"`
	YourQuantity       *int                `json:"yourQuantity"`
	YourPaymentStatus  *string             `json:"yourPaymentStatus"`
	YourDeliveryStatus *string             `json:"yourDeliveryStatus"`
	YourDeliveryMethod *string             `json:"yourDeliveryMethod"`
	YourShippingFee    *float64            `json:"yourShippingFee"`
	AllPaid            bool                `json:"allPaid"`
	Participants       []ParticipationJSON `json:"participants"`
	Disputes           []DisputeJSON       `json:"disputes"`
	Product            ProductJSON         `json:"product"`
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

type PayOrderBody struct {
	ShippingAddress string `json:"shippingAddress"`
	DeliveryMethod  string `json:"deliveryMethod"`
}

func (o Order) JSON() OrderJSON {
	participants := make([]ParticipationJSON, 0, len(o.Participants))
	for _, part := range o.Participants {
		participants = append(participants, part.JSON())
	}
	disputes := make([]DisputeJSON, 0, len(o.Disputes))
	for _, dispute := range o.Disputes {
		disputes = append(disputes, dispute.JSON())
	}
	return OrderJSON{
		ID:                 o.ID,
		Status:             o.Status,
		MOQ:                o.MOQ,
		Deadline:           o.Deadline.UTC(),
		CreatedAt:          o.CreatedAt.UTC(),
		CollectedQty:       o.CollectedQty,
		ParticipantCount:   o.ParticipantCount,
		YourQuantity:       o.YourQuantity,
		YourPaymentStatus:  o.YourPaymentStatus,
		YourDeliveryStatus: o.YourDeliveryStatus,
		YourDeliveryMethod: o.YourDeliveryMethod,
		YourShippingFee:    o.YourShippingFee,
		AllPaid:            o.AllPaid,
		Participants:       participants,
		Disputes:           disputes,
		Product:            o.Product.JSON(),
	}
}

func (p Participation) JSON() ParticipationJSON {
	return ParticipationJSON{
		ID:              p.ID,
		Quantity:        p.Quantity,
		ShippingAddress: p.ShippingAddress,
		DeliveryMethod:  p.DeliveryMethod,
		ShippingFee:     p.ShippingFee,
		PaymentStatus:   p.PaymentStatus,
		DeliveryStatus:  p.DeliveryStatus,
	}
}

func ShipmentUpdateAllowed(orderStatus, paymentStatus, currentDelivery, next string) bool {
	if paymentStatus != PaymentPaid {
		return false
	}
	if orderStatus != OrderPaid && orderStatus != OrderCompleted {
		return false
	}
	if currentDelivery == "" {
		currentDelivery = DeliveryPending
	}
	switch next {
	case DeliveryShipped:
		return currentDelivery == DeliveryPending
	case DeliveryDelivered:
		return currentDelivery == DeliveryShipped
	default:
		return false
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

func ParsePayOrder(body PayOrderBody) (string, string, map[string]string) {
	errs := map[string]string{}
	address := strings.TrimSpace(body.ShippingAddress)
	if address == "" {
		errs["shippingAddress"] = "Enter a shipping address"
	}
	method := strings.TrimSpace(body.DeliveryMethod)
	if _, ok := DeliveryFee(method, 1); !ok {
		errs["deliveryMethod"] = "Choose a delivery method"
	}
	if len(errs) > 0 {
		return "", "", errs
	}
	return address, method, nil
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
