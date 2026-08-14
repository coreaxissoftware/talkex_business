// Package config loads settings from environment / .env.
// Every other package reads config through Get() (singleton) rather than
// os.Getenv directly, so tests can swap it.
package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

// defaultJWTSecret is the dev-only fallback. Loading it in a non-dev
// environment is a fatal misconfiguration — any attacker who knows this
// string (i.e. anyone who reads this file) could forge tokens.
const defaultJWTSecret = "changeme-generate-a-real-secret"

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	JWTAccessMinutes   int
	JWTRefreshDays     int
	Port               string
	Environment        string
	CORSOrigins        []string

	// OAuth provider client IDs (secrets kept server-side only)
	OAuthGoogleClientID   string
	OAuthGoogleSecret     string
	OAuthFacebookClientID string
	OAuthFacebookSecret   string
	OAuthGitHubClientID   string
	OAuthGitHubSecret     string
	OAuthAppleClientID    string
	OAuthAppleSecret      string
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
			JWTSecret:        envOr("JWT_SECRET", defaultJWTSecret),
			JWTAccessMinutes: envIntOr("JWT_ACCESS_MINUTES", 15),
			JWTRefreshDays:   envIntOr("JWT_REFRESH_DAYS", 30),
			Port:             envOr("PORT", "8080"),
			Environment:      envOr("ENVIRONMENT", "development"),
			CORSOrigins:      strings.Split(envOr("CORS_ORIGINS", "http://localhost:5173"), ","),

			OAuthGoogleClientID:   envOr("OAUTH_GOOGLE_CLIENT_ID", ""),
			OAuthGoogleSecret:     envOr("OAUTH_GOOGLE_SECRET", ""),
			OAuthFacebookClientID: envOr("OAUTH_FACEBOOK_CLIENT_ID", ""),
			OAuthFacebookSecret:   envOr("OAUTH_FACEBOOK_SECRET", ""),
			OAuthGitHubClientID:   envOr("OAUTH_GITHUB_CLIENT_ID", ""),
			OAuthGitHubSecret:     envOr("OAUTH_GITHUB_SECRET", ""),
			OAuthAppleClientID:    envOr("OAUTH_APPLE_CLIENT_ID", ""),
			OAuthAppleSecret:      envOr("OAUTH_APPLE_SECRET", ""),
		}

		// Fail loud in non-dev environments if the JWT secret was left at
		// its dev default — otherwise a forgotten env var means any
		// attacker can forge tokens.
		if !instance.IsDev() {
			if instance.JWTSecret == defaultJWTSecret || len(instance.JWTSecret) < 32 {
				log.Fatalf("config: JWT_SECRET must be set to a strong value (>=32 chars) in non-dev environments")
			}
			// Wildcard + credentials is a footgun even if browsers allow
			// it silently in some CORS setups.
			for _, o := range instance.CORSOrigins {
				if strings.TrimSpace(o) == "*" {
					log.Fatalf("config: CORS_ORIGINS=* is not allowed outside development")
				}
			}
		}
	})
	return instance
}

func (c *Config) IsDev() bool {
	return c.Environment == "development"
}

// BaseURL returns the API server's public base URL (no trailing slash).
func (c *Config) BaseURL() string {
	if v := os.Getenv("BASE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://localhost:" + c.Port
}

// FrontendURL returns the frontend's public URL for OAuth redirects.
func (c *Config) FrontendURL() string {
	if v := os.Getenv("FRONTEND_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://localhost:5173"
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
