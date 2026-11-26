package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

type NotificationStore interface {
	ListNotifications(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error)
	MarkNotificationsRead(ctx context.Context, userID uuid.UUID) error
}

type NotificationHandler struct {
	store NotificationStore
}

func (h NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	items, err := h.store.ListNotifications(r.Context(), user.ID)
	if err != nil {
		slog.Error("list notifications", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	encoded := make([]domain.NotificationJSON, 0, len(items))
	for _, item := range items {
		encoded = append(encoded, item.JSON())
	}
	writeJSON(w, http.StatusOK, encoded)
}

func (h NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	if err := h.store.MarkNotificationsRead(r.Context(), user.ID); err != nil {
		slog.Error("mark notifications read", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
