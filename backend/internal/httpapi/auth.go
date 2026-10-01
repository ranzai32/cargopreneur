package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

type UserStore interface {
	UserLookup
	CreateUser(ctx context.Context, input domain.NewUser) (domain.User, error)
	FindUserByEmail(ctx context.Context, email string) (domain.User, error)
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type AuthHandler struct {
	users     UserStore
	passwords *auth.Hasher
	tokens    *auth.Tokens
	secure    bool
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input domain.RegisterInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	input.Normalize()
	if fieldErrors := input.Validate(); len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return
	}

	hash, err := h.passwords.Hash(input.Password)
	if err != nil {
		slog.Error("hash password", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	user, err := h.users.CreateUser(r.Context(), domain.NewUser{
		BusinessName: input.BusinessName,
		ContactName:  input.ContactName,
		Email:        input.Email,
		PasswordHash: hash,
		Region:       input.Region,
		Role:         domain.RoleBusiness,
	})
	if errors.Is(err, domain.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "Email is already registered", map[string]string{
			"email": "An account with this email already exists",
		})
		return
	}
	if err != nil {
		slog.Error("create user", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	if !h.startSession(w, user.ID) {
		return
	}
	writeJSON(w, http.StatusCreated, user.JSON())
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input domain.LoginInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}
	input.Normalize()
	if fieldErrors := input.Validate(); len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return
	}

	user, err := h.users.FindUserByEmail(r.Context(), input.Email)
	if errors.Is(err, domain.ErrNotFound) {
		_ = h.passwords.Compare("", input.Password)
		writeError(w, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}
	if err != nil {
		slog.Error("find user", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	if err := h.passwords.Compare(user.PasswordHash, input.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}
	if !h.startSession(w, user.ID) {
		return
	}
	writeJSON(w, http.StatusOK, user.JSON())
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w, h.secure)
	w.WriteHeader(http.StatusNoContent)
}

func (h AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	writeJSON(w, http.StatusOK, user.JSON())
}

func (h AuthHandler) startSession(w http.ResponseWriter, userID uuid.UUID) bool {
	token, err := h.tokens.Issue(userID)
	if err != nil {
		slog.Error("issue token", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return false
	}
	setSessionCookie(w, token, h.secure, h.tokens.TTL())
	return true
}
