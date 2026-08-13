// Package auth handles JWT access/refresh token issuance, verification,
// and the Gin authentication middleware.
//
// Real JWT access + refresh token rotation — this platform needs
// interoperable auth for a Developer Portal with multiple language SDKs,
// unlike the consumer app's opaque bearer tokens. See CONTEXT.md.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

type TokenType string

const (
	AccessToken        TokenType = "access"
	RefreshToken       TokenType = "refresh"
	PasswordResetToken TokenType = "password_reset"
)

// passwordResetDuration keeps the reset window short — long enough for
// the email to arrive and be clicked, short enough that a stolen link
// isn't useful the next day.
const passwordResetDuration = 30 * time.Minute

type Claims struct {
	Type TokenType `json:"type"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrWrongType    = errors.New("unexpected token type")
)

func createToken(userID string, tokenType TokenType, duration time.Duration) (string, error) {
	cfg := config.Get()
	now := time.Now()
	claims := &Claims{
		Type: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			// jti enables refresh-token revocation/rotation lists later
			// (Phase 2) without changing the token shape now.
			ID: uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func CreateAccessToken(userID string) (string, error) {
	cfg := config.Get()
	return createToken(userID, AccessToken, time.Duration(cfg.JWTAccessMinutes)*time.Minute)
}

func CreateRefreshToken(userID string) (string, error) {
	cfg := config.Get()
	return createToken(userID, RefreshToken, time.Duration(cfg.JWTRefreshDays)*24*time.Hour)
}

func CreatePasswordResetToken(userID string) (string, error) {
	return createToken(userID, PasswordResetToken, passwordResetDuration)
}

func ValidatePasswordResetToken(tokenString string) (string, error) {
	return ValidateToken(tokenString, PasswordResetToken)
}

// ValidateToken parses and validates a JWT, ensuring it matches the
// expected TokenType. Returns the subject (user ID).
func ValidateToken(tokenString string, expectedType TokenType) (string, error) {
	cfg := config.Get()
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}
	if claims.Type != expectedType {
		return "", ErrWrongType
	}
	return claims.Subject, nil
}
