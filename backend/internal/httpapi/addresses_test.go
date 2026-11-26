package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

func TestSavedAddresses(t *testing.T) {
	users := newFakeUsers()
	book := &memAddresses{}
	tokens := auth.NewTokens("test-secret-value", time.Hour)
	router := NewRouter(RouterConfig{
		Users:      users,
		Pinger:     users,
		Passwords:  auth.NewTestHasher(),
		Tokens:     tokens,
		CORSOrigin: "http://localhost:5173",
		Limiter:    NewRateLimiter(100, time.Minute),
		Addresses:  book,
	})
	_, cookie := signedIn(t, users, tokens, domain.RoleBusiness, "buyer@harbor.example")

	request(t, router, http.MethodPost, "/me/addresses", map[string]any{"line": "short"}, cookie, http.StatusUnprocessableEntity)
	created, _ := request(t, router, http.MethodPost, "/me/addresses", map[string]any{
		"label": "Shop",
		"line":  "12 Dock Street, Almaty",
	}, cookie, http.StatusCreated)
	if created["label"] != "Shop" || created["line"] != "12 Dock Street, Almaty" {
		t.Fatalf("created = %#v", created)
	}
	id, _ := created["id"].(string)

	updated, _ := request(t, router, http.MethodPost, "/me/addresses/"+id, map[string]any{
		"label": "Warehouse",
		"line":  "4 Market Street, Almaty",
	}, cookie, http.StatusOK)
	if updated["label"] != "Warehouse" || updated["line"] != "4 Market Street, Almaty" {
		t.Fatalf("updated = %#v", updated)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/me/addresses", nil)
	req.AddCookie(cookie)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d", rec.Code)
	}
	var items []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0]["label"] != "Warehouse" {
		t.Fatalf("list = %#v", items)
	}

	request(t, router, http.MethodDelete, "/me/addresses/"+id, nil, cookie, http.StatusNoContent)
	request(t, router, http.MethodDelete, "/me/addresses/"+id, nil, cookie, http.StatusNotFound)
}

type memAddresses struct {
	mu    sync.Mutex
	items []domain.SavedAddress
}

func (m *memAddresses) ListAddresses(context.Context, uuid.UUID) ([]domain.SavedAddress, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.SavedAddress, len(m.items))
	copy(out, m.items)
	return out, nil
}

func (m *memAddresses) CreateAddress(_ context.Context, userID uuid.UUID, label, line string) (domain.SavedAddress, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.items) >= 12 {
		return domain.SavedAddress{}, domain.ErrAddressLimit
	}
	item := domain.SavedAddress{ID: uuid.New(), UserID: userID, Label: label, Line: line, UpdatedAt: time.Now().UTC()}
	m.items = append(m.items, item)
	return item, nil
}

func (m *memAddresses) UpdateAddress(_ context.Context, userID, id uuid.UUID, label, line string) (domain.SavedAddress, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.items {
		if m.items[i].ID == id && m.items[i].UserID == userID {
			m.items[i].Label = label
			m.items[i].Line = line
			m.items[i].UpdatedAt = time.Now().UTC()
			return m.items[i], nil
		}
	}
	return domain.SavedAddress{}, domain.ErrNotFound
}

func (m *memAddresses) DeleteAddress(_ context.Context, userID, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.items {
		if m.items[i].ID == id && m.items[i].UserID == userID {
			m.items = append(m.items[:i], m.items[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}
