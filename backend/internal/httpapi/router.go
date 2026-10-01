package httpapi

import (
	"log/slog"
	"net/http"

	"cargoprener/backend/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type RouterConfig struct {
	Users        UserStore
	Pinger       Pinger
	Passwords    *auth.Hasher
	Tokens       *auth.Tokens
	CookieSecure bool
	CORSOrigin   string
	Limiter      *RateLimiter
}

func NewRouter(cfg RouterConfig) http.Handler {
	authHandler := AuthHandler{
		users:     cfg.Users,
		passwords: cfg.Passwords,
		tokens:    cfg.Tokens,
		secure:    cfg.CookieSecure,
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(logRequests)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigin},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Get("/health", healthHandler(cfg.Pinger))
	router.Route("/auth", func(r chi.Router) {
		r.With(rateLimit(cfg.Limiter)).Post("/register", authHandler.Register)
		r.With(rateLimit(cfg.Limiter)).Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.With(requireAuth(cfg.Users, cfg.Tokens)).Get("/me", authHandler.Me)
	})

	return router
}

func healthHandler(pinger Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pinger.Ping(r.Context()); err != nil {
			slog.Error("database ping", "error", err)
			writeError(w, http.StatusServiceUnavailable, "Database unavailable", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
