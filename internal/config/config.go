// Package config loads settings from environment / .env.
// Every other package reads config through Get() (singleton) rather than
// os.Getenv directly, so tests can swap it.
package config

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	JWTAccessMinutes   int
	JWTRefreshDays     int
	Port               string
	Environment        string
	CORSOrigins        []string
}

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		_ = godotenv.Load() // ignore error if .env missing — env vars may be set externally

		instance = &Config{
			DatabaseURL:      envOr("DATABASE_URL", "sqlite://talkex_business.db"),
			JWTSecret:        envOr("JWT_SECRET", "changeme-generate-a-real-secret"),
			JWTAccessMinutes: envIntOr("JWT_ACCESS_MINUTES", 15),
			JWTRefreshDays:   envIntOr("JWT_REFRESH_DAYS", 30),
			Port:             envOr("PORT", "8080"),
			Environment:      envOr("ENVIRONMENT", "development"),
			CORSOrigins:      strings.Split(envOr("CORS_ORIGINS", "http://localhost:5173"), ","),
		}
	})
	return instance
}

func (c *Config) IsDev() bool {
	return c.Environment == "development"
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
