package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	NoticeOrderCreated      = "order_created"
	NoticeOrderConfirmed    = "order_confirmed"
	NoticeOrderCancelled    = "order_cancelled"
	NoticePaymentReceived   = "payment_received"
	NoticeShipmentSent      = "shipment_sent"
	NoticeShipmentDelivered = "shipment_delivered"
	NoticeDisputeOpened     = "dispute_opened"
	NoticeDisputeResolved   = "dispute_resolved"
)

type Notice struct {
	Kind    string
	OrderID uuid.UUID
	UserIDs []uuid.UUID
}

type Notification struct {
	ID          uuid.UUID
	Kind        string
	OrderID     uuid.UUID
	ProductName string
	CreatedAt   time.Time
	Read        bool
}

type NotificationJSON struct {
	ID          uuid.UUID `json:"id"`
	Kind        string    `json:"kind"`
	OrderID     uuid.UUID `json:"orderId"`
	ProductName string    `json:"productName"`
	CreatedAt   time.Time `json:"createdAt"`
	Read        bool      `json:"read"`
}

func (n Notification) JSON() NotificationJSON {
	return NotificationJSON{
		ID:          n.ID,
		Kind:        n.Kind,
		OrderID:     n.OrderID,
		ProductName: n.ProductName,
		CreatedAt:   n.CreatedAt.UTC(),
		Read:        n.Read,
	}
}

type Notifier interface {
	Notify(ctx context.Context, notice Notice) error
}

type NopNotifier struct{}

func (NopNotifier) Notify(context.Context, Notice) error {
	return nil
}

type noticeLog struct {
	items []Notice
}

type noticeLogKey struct{}

func WithNoticeLog(ctx context.Context) context.Context {
	if _, ok := ctx.Value(noticeLogKey{}).(*noticeLog); ok {
		return ctx
	}
	return context.WithValue(ctx, noticeLogKey{}, &noticeLog{})
}

func AddNotice(ctx context.Context, notice Notice) {
	log, ok := ctx.Value(noticeLogKey{}).(*noticeLog)
	if !ok || log == nil {
		return
	}
	notice.UserIDs = append([]uuid.UUID(nil), notice.UserIDs...)
	log.items = append(log.items, notice)
}

func TakeNotices(ctx context.Context) []Notice {
	log, ok := ctx.Value(noticeLogKey{}).(*noticeLog)
	if !ok || log == nil || len(log.items) == 0 {
		return nil
	}
	items := log.items
	log.items = nil
	return items
}

func NoticeKindForStatus(status string) string {
	switch status {
	case OrderConfirmed:
		return NoticeOrderConfirmed
	case OrderCancelled:
		return NoticeOrderCancelled
	default:
		return ""
	}
}
