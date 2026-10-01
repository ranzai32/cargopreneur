package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"cargoprener/backend/internal/domain"
)

type errorBody struct {
	Message     string            `json:"message"`
	FieldErrors map[string]string `json:"fieldErrors,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string, fieldErrors map[string]string) {
	writeJSON(w, status, errorBody{Message: message, FieldErrors: fieldErrors})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusBadRequest, "Request body is too large", nil)
			return err
		}
		if errors.Is(err, context.Canceled) {
			return err
		}
		message := "Invalid JSON"
		if errors.Is(err, io.EOF) {
			message = "Request body is required"
		}
		writeError(w, http.StatusBadRequest, message, nil)
		return err
	}
	return nil
}

func writeValidation(w http.ResponseWriter, fieldErrors map[string]string) {
	writeError(w, http.StatusUnprocessableEntity, "Check the highlighted fields", fieldErrors)
}

func userFrom(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(userKey{}).(domain.User)
	return user, ok
}

type userKey struct{}

func withUser(ctx context.Context, user domain.User) context.Context {
	user.PasswordHash = ""
	return context.WithValue(ctx, userKey{}, user)
}
