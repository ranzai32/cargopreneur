package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/config"
	"cargoprener/backend/internal/httpapi"
	"cargoprener/backend/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := waitForDB(ctx, db); err != nil {
		slog.Error("database ping", "error", err)
		os.Exit(1)
	}
	if err := store.Migrate(cfg.DatabaseURL); err != nil {
		slog.Error("migrations", "error", err)
		os.Exit(1)
	}

	passwords := auth.NewHasher()
	if err := db.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword, passwords.Hash); err != nil {
		slog.Error("admin seed", "error", err)
		os.Exit(1)
	}

	router := httpapi.NewRouter(httpapi.RouterConfig{
		Users:        db,
		Pinger:       db,
		Passwords:    passwords,
		Tokens:       auth.NewTokens(cfg.JWTSecret, auth.TokenTTL),
		CookieSecure: cfg.CookieSecure,
		CORSOrigin:   cfg.CORSOrigin,
		Limiter:      httpapi.NewRateLimiter(10, time.Minute),
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "error", err)
		os.Exit(1)
	}
}

func waitForDB(ctx context.Context, db *store.Store) error {
	var err error
	for attempt := 0; attempt < 30; attempt++ {
		err = db.Ping(ctx)
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return err
}
