package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"cargoprener/backend/internal/domain"
	"cargoprener/backend/internal/places"

	"github.com/google/uuid"
)

type AddressBook interface {
	ListAddresses(ctx context.Context, userID uuid.UUID) ([]domain.SavedAddress, error)
	CreateAddress(ctx context.Context, userID uuid.UUID, label, line string) (domain.SavedAddress, error)
	UpdateAddress(ctx context.Context, userID, id uuid.UUID, label, line string) (domain.SavedAddress, error)
	DeleteAddress(ctx context.Context, userID, id uuid.UUID) error
}

type AddressHandler struct {
	book AddressBook
}

func (h AddressHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	items, err := h.book.ListAddresses(r.Context(), user.ID)
	if err != nil {
		slog.Error("list addresses", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	payload := make([]domain.SavedAddressJSON, 0, len(items))
	for _, item := range items {
		payload = append(payload, item.JSON())
	}
	writeJSON(w, http.StatusOK, payload)
}

func (h AddressHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	label, line, ok := readAddress(w, r)
	if !ok {
		return
	}
	item, err := h.book.CreateAddress(r.Context(), user.ID, label, line)
	if errors.Is(err, domain.ErrAddressLimit) {
		writeError(w, http.StatusUnprocessableEntity, "You can save up to 12 addresses", nil)
		return
	}
	if err != nil {
		slog.Error("create address", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusCreated, item.JSON())
}

func (h AddressHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	label, line, ok := readAddress(w, r)
	if !ok {
		return
	}
	item, err := h.book.UpdateAddress(r.Context(), user.ID, id, label, line)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Address not found", nil)
		return
	}
	if err != nil {
		slog.Error("update address", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, item.JSON())
}

func (h AddressHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	err := h.book.DeleteAddress(r.Context(), user.ID, id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Address not found", nil)
		return
	}
	if err != nil {
		slog.Error("delete address", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func addressHandler(source places.Suggester) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := source.Suggest(r.Context(), r.URL.Query().Get("q"))
		if err != nil {
			slog.Error("address suggest", "error", err)
			writeError(w, http.StatusBadGateway, "Address suggestions are unavailable", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

func readAddress(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	var body domain.AddressBody
	if err := decodeJSON(w, r, &body); err != nil {
		return "", "", false
	}
	label, line, fieldErrors := domain.ParseAddress(body)
	if len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return "", "", false
	}
	return label, line, true
}
