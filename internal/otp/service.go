// Package otp provides a dev-mode OTP (one-time password) service for
// phone and email verification during registration.
//
// In production, SendOTP would call Twilio/MSG91 for SMS and SendGrid/SES
// for email. In dev mode, codes are logged to the console — same pattern
// as password reset tokens (see users/handler.go handleForgotPassword).
package otp

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"
)

const (
	codeLength = 6
	codeTTL    = 5 * time.Minute
	maxAttempts = 5
)

var (
	ErrInvalidCode = errors.New("invalid or expired OTP")
	ErrTooManyAttempts = errors.New("too many verification attempts")
	ErrRateLimited = errors.New("please wait before requesting another OTP")
)

// entry holds a pending OTP.
type entry struct {
	Code      string
	ExpiresAt time.Time
	Attempts  int
	SentAt    time.Time
}

// Store is an in-memory OTP store. Production would use Redis or a DB
// table with TTL-based cleanup, but this keeps dependencies minimal.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*entry // key = "phone:<number>" or "email:<address>"
}

var defaultStore = &Store{entries: make(map[string]*entry)}

// generateCode returns a cryptographically random N-digit numeric string.
func generateCode() (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(codeLength)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", codeLength, n.Int64()), nil
}

// SendInput is the request body for POST /auth/otp/send.
type SendInput struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// VerifyInput is the request body for POST /auth/otp/verify.
type VerifyInput struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Code  string `json:"code" binding:"required"`
}

// key builds the store key from the input.
func key(phone, email string) string {
	if phone != "" {
		return "phone:" + phone
	}
	return "email:" + email
}

// Send generates and "sends" an OTP. In dev mode, it logs the code.
func Send(phone, email string) error {
	if phone == "" && email == "" {
		return errors.New("phone or email required")
	}
	k := key(phone, email)

	store := defaultStore
	store.mu.Lock()
	defer store.mu.Unlock()

	// Rate limit: 30s between sends to the same target
	if existing, ok := store.entries[k]; ok {
		if time.Since(existing.SentAt) < 30*time.Second {
			return ErrRateLimited
		}
	}

	code, err := generateCode()
	if err != nil {
		return err
	}

	store.entries[k] = &entry{
		Code:      code,
		ExpiresAt: time.Now().Add(codeTTL),
		Attempts:  0,
		SentAt:    time.Now(),
	}

	// Dev-mode: log the OTP to console (same pattern as password reset)
	target := phone
	if target == "" {
		target = email
	}
	log.Printf("OTP (dev): %s → %s", target, code)

	// Production TODO: call SMS gateway (phone) or email provider (email)
	// to actually deliver the code. The gateway selection would be based
	// on config.Get().Environment != "development".

	return nil
}

// Verify checks the submitted OTP code against the stored one.
func Verify(phone, email, code string) error {
	k := key(phone, email)

	store := defaultStore
	store.mu.Lock()
	defer store.mu.Unlock()

	e, ok := store.entries[k]
	if !ok || time.Now().After(e.ExpiresAt) {
		delete(store.entries, k)
		return ErrInvalidCode
	}

	if e.Attempts >= maxAttempts {
		delete(store.entries, k)
		return ErrTooManyAttempts
	}

	e.Attempts++

	if e.Code != code {
		return ErrInvalidCode
	}

	// OTP verified — remove it so it can't be reused
	delete(store.entries, k)
	return nil
}

// Cleanup removes expired entries. Call periodically in production.
func Cleanup() {
	store := defaultStore
	store.mu.Lock()
	defer store.mu.Unlock()

	now := time.Now()
	for k, e := range store.entries {
		if now.After(e.ExpiresAt) {
			delete(store.entries, k)
		}
	}
}
