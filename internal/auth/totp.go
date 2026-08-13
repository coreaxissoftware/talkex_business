package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	totpDigits = 6
	totpPeriod = 30
)

// GenerateTOTPSecret creates a random 20-byte base32-encoded secret.
func GenerateTOTPSecret() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(buf), "="), nil
}

// TOTPCode computes the current TOTP code for the given secret.
func TOTPCode(secret string, t time.Time) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", err
	}
	counter := uint64(t.Unix()) / totpPeriod
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff
	otp := int(code) % int(math.Pow10(totpDigits))

	return fmt.Sprintf("%0*d", totpDigits, otp), nil
}

// VerifyTOTP checks if the provided code matches the current or adjacent
// time step (±1 window for clock skew).
func VerifyTOTP(secret, code string) bool {
	now := time.Now()
	for _, offset := range []int{-1, 0, 1} {
		t := now.Add(time.Duration(offset*totpPeriod) * time.Second)
		expected, err := TOTPCode(secret, t)
		if err != nil {
			continue
		}
		if hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}
	return false
}

// TOTPProvisioningURI builds an otpauth:// URI for QR code generation.
func TOTPProvisioningURI(secret, email, issuer string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		issuer, email, secret, issuer, totpDigits, totpPeriod)
}
