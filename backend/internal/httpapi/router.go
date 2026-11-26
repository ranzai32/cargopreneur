package httpapi

import (
	"log/slog"
	"net/http"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"
	"cargoprener/backend/internal/places"

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
	Catalog      CatalogStore
	Orders       OrderStore
	Notifier     domain.Notifier
	UploadDir    string
	Supplier     SupplierAPI
	Addresses    AddressBook
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
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Get("/health", healthHandler(cfg.Pinger))
	router.Get("/uploads/{name}", func(w http.ResponseWriter, r *http.Request) {
		serveUpload(cfg.UploadDir, w, r)
	})
	router.With(requireAuth(cfg.Users, cfg.Tokens)).Get("/addresses", addressHandler(places.OSM{}))
	if cfg.Addresses != nil {
		book := AddressHandler{book: cfg.Addresses}
		router.With(requireAuth(cfg.Users, cfg.Tokens)).Get("/me/addresses", book.List)
		router.With(requireAuth(cfg.Users, cfg.Tokens)).Post("/me/addresses", book.Create)
		router.With(requireAuth(cfg.Users, cfg.Tokens)).Post("/me/addresses/{id}", book.Update)
		router.With(requireAuth(cfg.Users, cfg.Tokens)).Delete("/me/addresses/{id}", book.Delete)
	}
	router.Route("/auth", func(r chi.Router) {
		r.With(rateLimit(cfg.Limiter)).Post("/register", authHandler.Register)
		r.With(rateLimit(cfg.Limiter)).Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.With(requireAuth(cfg.Users, cfg.Tokens)).Get("/me", authHandler.Me)
		r.With(requireAuth(cfg.Users, cfg.Tokens)).Post("/me", authHandler.UpdateMe)
	})

	if cfg.Catalog != nil {
		catalog := CatalogHandler{catalog: cfg.Catalog, users: cfg.Users, tokens: cfg.Tokens}
		router.Get("/products", catalog.ListProducts)
		router.Get("/products/{id}", catalog.GetProduct)
		router.Get("/suppliers", catalog.ListSuppliers)
		router.Get("/suppliers/{id}", catalog.GetSupplier)
		router.Get("/categories", catalog.ListCategories)
	}

	notifier := cfg.Notifier
	if notifier == nil {
		notifier = domain.NopNotifier{}
	}

	if cfg.Supplier != nil {
		supplier := SupplierHandler{api: cfg.Supplier, notifier: notifier, uploadDir: cfg.UploadDir}
		router.Get("/products/{id}/reviews", supplier.ProductReviews)
		router.Get("/suppliers/{id}/reviews", supplier.SupplierReviews)
		router.With(requireAuth(cfg.Users, cfg.Tokens)).Post("/products/{id}/reviews", supplier.WriteProductReview)
		router.With(requireAuth(cfg.Users, cfg.Tokens)).Post("/suppliers/{id}/reviews", supplier.WriteSupplierReview)
		router.Route("/supplier", func(r chi.Router) {
			r.Use(requireAuth(cfg.Users, cfg.Tokens))
			r.Use(RequireRole(domain.RoleSupplier))
			r.Get("/products", supplier.ListProducts)
			r.Post("/products", supplier.CreateProduct)
			r.Put("/products/{id}", supplier.UpdateProduct)
			r.Delete("/products/{id}", supplier.DeleteProduct)
			r.Post("/uploads", supplier.Upload)
			r.Get("/orders", supplier.ListOrders)
			r.Post("/participations/{id}/ship", supplier.Ship)
			r.Post("/participations/{id}/deliver", supplier.Deliver)
		})
	}

	if cfg.Orders != nil {
		orders := OrderHandler{orders: cfg.Orders, users: cfg.Users, tokens: cfg.Tokens, notifier: notifier}
		router.Get("/orders", orders.List)
		router.Get("/orders/{id}", orders.Get)
		router.With(requireAuth(cfg.Users, cfg.Tokens)).Post("/orders", orders.Create)
		router.With(requireAuth(cfg.Users, cfg.Tokens)).Post("/orders/{id}/join", orders.Join)
		router.With(requireAuth(cfg.Users, cfg.Tokens)).Post("/orders/{id}/pay", orders.Pay)
		router.With(requireAuth(cfg.Users, cfg.Tokens)).Post("/orders/{id}/disputes", orders.OpenDispute)
	}
	if cfg.Catalog != nil || cfg.Orders != nil {
		admin := AdminHandler{catalog: cfg.Catalog, orders: cfg.Orders, notifier: notifier}
		router.Route("/admin", func(r chi.Router) {
			r.Use(requireAuth(cfg.Users, cfg.Tokens))
			r.Use(RequireRole(domain.RoleAdmin))
			if cfg.Catalog != nil {
				r.Get("/suppliers", admin.ListSuppliers)
				r.Post("/suppliers", admin.CreateSupplier)
				r.Post("/suppliers/{id}/verify", admin.VerifySupplier)
				r.Post("/products", admin.CreateProduct)
				r.Post("/uploads", func(w http.ResponseWriter, r *http.Request) {
					url, status, message := saveUpload(cfg.UploadDir, w, r)
					if message != "" {
						writeError(w, status, message, nil)
						return
					}
					writeJSON(w, status, map[string]string{"url": url})
				})
			}
			if cfg.Orders != nil {
				r.Get("/orders", admin.ListOrders)
				r.Post("/orders/{id}/cancel", admin.CancelOrder)
				r.Post("/participations/{id}/ship", admin.ShipParticipation)
				r.Post("/participations/{id}/deliver", admin.DeliverParticipation)
				r.Get("/disputes", admin.ListDisputes)
				r.Post("/disputes/{id}/resolve", admin.ResolveDispute)
			}
		})
	}

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
