package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DisputeOpen     = "open"
	DisputeResolved = "resolved"

	minDisputeLength = 8
	maxDisputeLength = 500
)

type Dispute struct {
	ID           uuid.UUID
	OrderID      uuid.UUID
	OpenedBy     uuid.UUID
	BusinessName string
	Reason       string
	Status       string
	Outcome      string
	CreatedAt    time.Time
	ResolvedAt   *time.Time
}

type DisputeJSON struct {
	ID           uuid.UUID  `json:"id"`
	OrderID      uuid.UUID  `json:"orderId"`
	OpenedBy     uuid.UUID  `json:"openedBy"`
	BusinessName string     `json:"businessName"`
	Reason       string     `json:"reason"`
	Status       string     `json:"status"`
	Outcome      string     `json:"outcome"`
	CreatedAt    time.Time  `json:"createdAt"`
	ResolvedAt   *time.Time `json:"resolvedAt"`
}

type AdminDispute struct {
	Dispute
	ProductName string
	OrderStatus string
}

type AdminDisputeJSON struct {
	DisputeJSON
	ProductName string `json:"productName"`
	OrderStatus string `json:"orderStatus"`
}

type OpenDisputeBody struct {
	Reason string `json:"reason"`
}

type ResolveDisputeBody struct {
	Outcome string `json:"outcome"`
}

func (d Dispute) JSON() DisputeJSON {
	resolved := d.ResolvedAt
	if resolved != nil {
		utc := resolved.UTC()
		resolved = &utc
	}
	return DisputeJSON{
		ID:           d.ID,
		OrderID:      d.OrderID,
		OpenedBy:     d.OpenedBy,
		BusinessName: d.BusinessName,
		Reason:       d.Reason,
		Status:       d.Status,
		Outcome:      d.Outcome,
		CreatedAt:    d.CreatedAt.UTC(),
		ResolvedAt:   resolved,
	}
}

func (d AdminDispute) JSON() AdminDisputeJSON {
	return AdminDisputeJSON{
		DisputeJSON: d.Dispute.JSON(),
		ProductName: d.ProductName,
		OrderStatus: d.OrderStatus,
	}
}

func DisputeAllowed(status string) bool {
	switch status {
	case OrderConfirmed, OrderPaid, OrderCompleted:
		return true
	default:
		return false
	}
}

func ParseDisputeReason(raw string) (string, map[string]string) {
	reason := strings.TrimSpace(raw)
	if reason == "" {
		return "", map[string]string{"reason": "Describe the problem"}
	}
	if len([]rune(reason)) < minDisputeLength {
		return "", map[string]string{"reason": "Use at least 8 characters"}
	}
	if len([]rune(reason)) > maxDisputeLength {
		return "", map[string]string{"reason": "Use at most 500 characters"}
	}
	return reason, nil
}

func ParseDisputeOutcome(raw string) (string, map[string]string) {
	outcome := strings.TrimSpace(raw)
	if outcome == "" {
		return "", map[string]string{"outcome": "Record the outcome"}
	}
	if len([]rune(outcome)) > maxDisputeLength {
		return "", map[string]string{"outcome": "Use at most 500 characters"}
	}
	return outcome, nil
}
