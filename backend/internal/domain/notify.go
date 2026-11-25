package domain

import (
	"context"

	"github.com/google/uuid"
)

const (
	NoticeOrderConfirmed = "order_confirmed"
	NoticeOrderCancelled = "order_cancelled"
)

type Notice struct {
	Kind    string
	OrderID uuid.UUID
	UserIDs []uuid.UUID
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
