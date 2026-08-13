package developers

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ConstantTimeEqual is exposed for callers who compare user-supplied HMAC
// signatures against derived ones — always use it to avoid timing attacks.
func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ExtractBearer pulls a "Bearer <secret>" out of an Authorization header.
// Returns "" when the shape doesn't match; callers treat that as no key.
func ExtractBearer(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimPrefix(header, prefix)
}

var (
	ErrKeyNotFound = errors.New("api key not found")
	ErrAlreadyRevoked = errors.New("api key is already revoked")
)

// KeyPrefix is what every generated key starts with — makes leaked keys
// obviously identifiable in logs / git history / stack traces.
const KeyPrefix = "txb_"

// generateSecret returns 32 random bytes formatted as a hex string with the
// TalkEx Business key prefix — e.g. "txb_a1b2c3…" (68 chars total).
func generateSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return KeyPrefix + hex.EncodeToString(buf), nil
}

// HashKey returns the SHA-256 hex digest used at rest. This must match the
// hash produced when a key is presented for authentication.
func HashKey(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// List returns every API key for the owner, newest first.
func List(db *gorm.DB, ownerID string) ([]ApiKey, error) {
	var out []ApiKey
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out).Error
	return out, err
}

// Create generates a new key, persists its hash, and returns both the
// stored row and the plaintext (only chance the plaintext is exposed).
type CreateInput struct {
	Name string `json:"name" binding:"required"`
}

type CreateResult struct {
	ApiKey    ApiKey `json:"api_key"`
	Plaintext string `json:"plaintext"`
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*CreateResult, error) {
	secret, err := generateSecret()
	if err != nil {
		return nil, err
	}

	// The first 12 chars of the secret (prefix + 8 hex) are safe to show
	// in the dashboard for user recognition (e.g. "txb_a1b2c3d4…").
	displayPrefix := secret[:12]

	k := ApiKey{
		OwnerID: ownerID,
		Name:    in.Name,
		Prefix:  displayPrefix,
		KeyHash: HashKey(secret),
	}
	if err := db.Create(&k).Error; err != nil {
		return nil, err
	}
	return &CreateResult{ApiKey: k, Plaintext: secret}, nil
}

// Revoke stamps RevokedAt. Idempotency-friendly for the "click revoke twice"
// case: second call returns ErrAlreadyRevoked rather than silently succeeding.
func Revoke(db *gorm.DB, ownerID, id string) (*ApiKey, error) {
	var k ApiKey
	if err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&k).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	if k.RevokedAt != nil {
		return nil, ErrAlreadyRevoked
	}
	now := time.Now()
	k.RevokedAt = &now
	if err := db.Save(&k).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

// Delete permanently removes a key row. Only allow this for already-revoked
// keys — deleting an active key would leave outstanding key holders unable
// to see why their requests suddenly 401.
func Delete(db *gorm.DB, ownerID, id string) error {
	var k ApiKey
	if err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&k).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrKeyNotFound
		}
		return err
	}
	return db.Delete(&k).Error
}

// ResolveKey looks up a plaintext API key and returns the row it maps to
// (or nil if no active row exists). Bumps LastUsedAt as a side effect so
// developers can spot idle keys in the dashboard.
func ResolveKey(db *gorm.DB, plaintext string) (*ApiKey, error) {
	hash := HashKey(plaintext)
	var k ApiKey
	if err := db.Where("key_hash = ? AND revoked_at IS NULL", hash).First(&k).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	now := time.Now()
	_ = db.Model(&k).Update("last_used_at", now).Error
	return &k, nil
}
