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
	// Apple Sign-in needs a JWT client secret minted from an EC P-256
	// key rather than a static string. Team ID + Key ID identify the
	// key registered in the Apple Developer portal; PrivateKeyPEM is
	// the PEM-encoded contents of the downloaded .p8 file.
	OAuthAppleTeamID      string
	OAuthAppleKeyID       string
	OAuthApplePrivateKey  string

	// Razorpay payment gateway
	RazorpayKeyID         string
	RazorpaySecret        string
	RazorpayWebhookSecret string

	// Redis — when set, rate limiter / SSE hub / OTP store switch to
	// Redis so multiple API pods share state. Unset = in-memory (dev
	// or single-pod deploys).
	RedisURL string

	// Mailgun — transactional email (OTP delivery, password reset).
	// Domain must be verified in the Mailgun dashboard.
	MailgunDomain  string
	MailgunAPIKey  string
	MailgunFrom    string // e.g. "TalkEx <no-reply@mail.talkex.in>"
	MailgunBaseURL string // override for EU region

	// MSG91 — India-first SMS OTP delivery. Cheaper than Twilio on
	// Indian routes and DLT-native. TemplateID must be pre-registered
	// on TRAI DLT.
	Msg91AuthKey    string
	Msg91TemplateID string
	Msg91SenderID   string
	Msg91Route      string // default "4" (transactional)

	// Fast2SMS — India SMS. Supports two routes: the built-in "otp"
	// route works without DLT registration (fastest to go live);
	// setting a TemplateID + SenderID switches to the "dlt" route.
	Fast2SMSAPIKey     string
	Fast2SMSSenderID   string
	Fast2SMSTemplateID string
	Fast2SMSRoute      string // "dlt" | "otp" | "q"; auto-picked when empty

	// Twilio — global SMS fallback. Only used when MSG91 + Fast2SMS
	// are both unconfigured.
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string
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
			OAuthAppleTeamID:      envOr("OAUTH_APPLE_TEAM_ID", ""),
			OAuthAppleKeyID:       envOr("OAUTH_APPLE_KEY_ID", ""),
			OAuthApplePrivateKey:  envOr("OAUTH_APPLE_PRIVATE_KEY", ""),

			RazorpayKeyID:         envOr("RAZORPAY_KEY_ID", ""),
			RazorpaySecret:        envOr("RAZORPAY_SECRET", ""),
			RazorpayWebhookSecret: envOr("RAZORPAY_WEBHOOK_SECRET", ""),

			RedisURL: envOr("REDIS_URL", ""),

			MailgunDomain:  envOr("MAILGUN_DOMAIN", ""),
			MailgunAPIKey:  envOr("MAILGUN_API_KEY", ""),
			MailgunFrom:    envOr("MAILGUN_FROM", ""),
			MailgunBaseURL: envOr("MAILGUN_BASE_URL", ""),

			Msg91AuthKey:    envOr("MSG91_AUTH_KEY", ""),
			Msg91TemplateID: envOr("MSG91_TEMPLATE_ID", ""),
			Msg91SenderID:   envOr("MSG91_SENDER_ID", ""),
			Msg91Route:      envOr("MSG91_ROUTE", ""),

			Fast2SMSAPIKey:     envOr("FAST2SMS_API_KEY", ""),
			Fast2SMSSenderID:   envOr("FAST2SMS_SENDER_ID", ""),
			Fast2SMSTemplateID: envOr("FAST2SMS_TEMPLATE_ID", ""),
			Fast2SMSRoute:      envOr("FAST2SMS_ROUTE", ""),

			TwilioAccountSID: envOr("TWILIO_ACCOUNT_SID", ""),
			TwilioAuthToken:  envOr("TWILIO_AUTH_TOKEN", ""),
			TwilioFromNumber: envOr("TWILIO_FROM_NUMBER", ""),
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
