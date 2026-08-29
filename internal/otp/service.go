// Package otp provides a dev-mode OTP (one-time password) service for
// phone and email verification during registration.
//
// In production, SendOTP would call Twilio/MSG91 for SMS and SendGrid/SES
// for email. In dev mode, codes are logged to the console — same pattern
// as password reset tokens (see users/handler.go handleForgotPassword).
package otp

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/coreaxissoftware/talkex_business/internal/redisclient"
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

	code, err := generateCode()
	if err != nil {
		return err
	}

	// Redis-backed path — shared across pods.
	if rdb := redisclient.Get(); rdb != nil {
		return sendRedis(rdb, k, code, phone, email)
	}

	// In-memory fallback (dev, single pod).
	store := defaultStore
	store.mu.Lock()
	defer store.mu.Unlock()

	// Rate limit: 30s between sends to the same target
	if existing, ok := store.entries[k]; ok {
		if time.Since(existing.SentAt) < 30*time.Second {
			return ErrRateLimited
		}
	}

	store.entries[k] = &entry{
		Code:      code,
		ExpiresAt: time.Now().Add(codeTTL),
		Attempts:  0,
		SentAt:    time.Now(),
	}

	logDelivery(phone, email, code)
	return nil
}

func logDelivery(phone, email, code string) {
	target := phone
	if target == "" {
		target = email
	}
	log.Printf("OTP (dev): %s → %s", target, code)
	// Production TODO: call SMS gateway (phone) or email provider (email).
}

// sendRedis stores the code in Redis using a HASH with (code, attempts,
// sent_at) and a codeTTL expiry. The 30s send-cooldown is enforced by
// checking sent_at before overwriting.
func sendRedis(rdb *redis.Client, k, code, phone, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	rk := "otp:" + k
	sentAt, err := rdb.HGet(ctx, rk, "sent_at").Result()
	if err == nil && sentAt != "" {
		if ts, perr := time.Parse(time.RFC3339Nano, sentAt); perr == nil {
			if time.Since(ts) < 30*time.Second {
				return ErrRateLimited
			}
		}
	}

	now := time.Now()
	pipe := rdb.TxPipeline()
	pipe.HSet(ctx, rk,
		"code", code,
		"attempts", "0",
		"sent_at", now.Format(time.RFC3339Nano),
	)
	pipe.Expire(ctx, rk, codeTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	logDelivery(phone, email, code)
	return nil
}

// Verify checks the submitted OTP code against the stored one.
func Verify(phone, email, code string) error {
	k := key(phone, email)

	if rdb := redisclient.Get(); rdb != nil {
		return verifyRedis(rdb, k, code)
	}

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

	// Constant-time comparison — string equality short-circuits on
	// first differing byte and leaks the correct prefix via timing.
	if subtle.ConstantTimeCompare([]byte(e.Code), []byte(code)) != 1 {
		return ErrInvalidCode
	}

	// OTP verified — remove it so it can't be reused
	delete(store.entries, k)
	return nil
}

func verifyRedis(rdb *redis.Client, k, code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	rk := "otp:" + k
	// Atomic attempt increment; TTL survives from Send() so an expired
	// entry returns 0 and we treat it as invalid.
	attempts, err := rdb.HIncrBy(ctx, rk, "attempts", 1).Result()
	if err != nil {
		return ErrInvalidCode
	}
	if attempts == 1 {
		// This is the first attempt; keep the TTL as-is. If the key
		// didn't exist HIncrBy created it and we treat as invalid.
	}
	stored, err := rdb.HGet(ctx, rk, "code").Result()
	if err != nil || stored == "" {
		return ErrInvalidCode
	}
	if attempts > int64(maxAttempts) {
		rdb.Del(ctx, rk)
		return ErrTooManyAttempts
	}
	if subtle.ConstantTimeCompare([]byte(stored), []byte(code)) != 1 {
		return ErrInvalidCode
	}
	rdb.Del(ctx, rk)
	return nil
}

// Cleanup removes expired entries. Called periodically by StartCleanup.
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

// StartCleanup launches a background ticker that reaps expired OTPs
// so the in-memory store doesn't grow unbounded across abandoned
// signups. Call once from main.go.
func StartCleanup(interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			Cleanup()
		}
	}()
}
