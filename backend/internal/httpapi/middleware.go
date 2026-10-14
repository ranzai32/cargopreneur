package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

const sessionCookie = "session"

type UserLookup interface {
	FindUserByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

func setSessionCookie(w http.ResponseWriter, token string, secure bool, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	})
}

func clearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func optionalUser(r *http.Request, users UserLookup, tokens *auth.Tokens) (domain.User, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil || cookie.Value == "" {
		return domain.User{}, false
	}
	userID, err := tokens.Parse(cookie.Value)
	if err != nil {
		return domain.User{}, false
	}
	user, err := users.FindUserByID(r.Context(), userID)
	if err != nil {
		return domain.User{}, false
	}
	return user, true
}

func requireAuth(users UserLookup, tokens *auth.Tokens) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := authenticate(w, r, users, tokens)
			if !ok {
				return
			}
			next.ServeHTTP(w, r.WithContext(withUser(r.Context(), user)))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := userFrom(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "Authentication required", nil)
				return
			}
			if user.Role != role {
				writeError(w, http.StatusForbidden, "You do not have access to this", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func authenticate(w http.ResponseWriter, r *http.Request, users UserLookup, tokens *auth.Tokens) (domain.User, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return domain.User{}, false
	}
	userID, err := tokens.Parse(cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return domain.User{}, false
	}
	user, err := users.FindUserByID(r.Context(), userID)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return domain.User{}, false
	}
	if err != nil {
		slog.Error("load authenticated user", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return domain.User{}, false
	}
	return user, true
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(wrapped, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", wrapped.Status(), "duration", time.Since(start))
	})
}
