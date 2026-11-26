package httpapi

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

func TestNotificationsBelongToTheSignedInUser(t *testing.T) {
	users := newFakeUsers()
	store := &memNotifications{byUser: map[uuid.UUID][]domain.Notification{}}
	tokens := auth.NewTokens("test-secret-value", time.Hour)
	router := NewRouter(RouterConfig{
		Users:         users,
		Pinger:        users,
		Passwords:     auth.NewTestHasher(),
		Tokens:        tokens,
		CORSOrigin:    "http://localhost:5173",
		Limiter:       NewRateLimiter(100, time.Minute),
		Notifications: store,
	})
	owner, ownerCookie := signedIn(t, users, tokens, domain.RoleBusiness, "owner@example.com")
	_, otherCookie := signedIn(t, users, tokens, domain.RoleBusiness, "other@example.com")
	store.byUser[owner.ID] = []domain.Notification{{
		ID:          uuid.New(),
		Kind:        domain.NoticeOrderConfirmed,
		OrderID:     uuid.New(),
		ProductName: "Paper cups",
		CreatedAt:   time.Now(),
	}}

	request(t, router, http.MethodGet, "/notifications", nil, nil, http.StatusUnauthorized)

	items, _ := requestList(t, router, "/notifications", ownerCookie, http.StatusOK)
	if len(items) != 1 {
		t.Fatalf("owner items = %#v", items)
	}
	item := items[0].(map[string]any)
	if !sameStrings(mapKeys(item), []string{"createdAt", "id", "kind", "orderId", "productName", "read"}) {
		t.Fatalf("keys = %#v", mapKeys(item))
	}
	if item["read"] != false || item["kind"] != domain.NoticeOrderConfirmed {
		t.Fatalf("item = %#v", item)
	}

	other, _ := requestList(t, router, "/notifications", otherCookie, http.StatusOK)
	if len(other) != 0 {
		t.Fatalf("other items = %#v", other)
	}

	request(t, router, http.MethodPost, "/notifications/read", nil, ownerCookie, http.StatusNoContent)
	items, _ = requestList(t, router, "/notifications", ownerCookie, http.StatusOK)
	if items[0].(map[string]any)["read"] != true {
		t.Fatalf("after read = %#v", items)
	}
}

type memNotifications struct {
	mu     sync.Mutex
	byUser map[uuid.UUID][]domain.Notification
}

func (m *memNotifications) ListNotifications(_ context.Context, userID uuid.UUID) ([]domain.Notification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]domain.Notification{}, m.byUser[userID]...), nil
}

func (m *memNotifications) MarkNotificationsRead(_ context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for index := range m.byUser[userID] {
		m.byUser[userID][index].Read = true
	}
	return nil
}
