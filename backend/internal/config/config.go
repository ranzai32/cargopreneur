package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	JWTSecret     string
	CORSOrigin    string
	CookieSecure  bool
	AdminEmail    string
	AdminPassword string
}

func Load() (Config, error) {
	_ = godotenv.Load(".env", "../.env")

	cfg := Config{
		Port:          env("PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		CORSOrigin:    env("CORS_ORIGIN", "http://localhost:5173"),
		CookieSecure:  envBool("COOKIE_SECURE"),
		AdminEmail:    strings.TrimSpace(os.Getenv("ADMIN_EMAIL")),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 16 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 16 characters")
	}
	if cfg.CORSOrigin == "" {
		return Config{}, fmt.Errorf("CORS_ORIGIN is required")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string) bool {
	value := strings.TrimSpace(os.Getenv(key))
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return parsed
}
