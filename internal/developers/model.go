// Package developers — Developer Portal: API keys today, webhooks + logs
// later (per CONTEXT.md's Phase 2 developer surface).
//
// API keys are stored HASHED with only a short display prefix retained,
// so a leaked database dump never gives an attacker a working key. The
// full key is returned exactly once — at creation — and never again.
package developers

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// ApiKey rows never contain the plaintext key. `KeyHash` is a SHA-256
// hex digest of the full secret; `Prefix` is the first 8 chars of the
// plaintext for display in the dashboard.
type ApiKey struct {
	database.Base
	OwnerID    string     `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Name       string     `gorm:"type:varchar(255);not null" json:"name"`
	Prefix     string     `gorm:"type:varchar(16);not null" json:"prefix"`
	KeyHash    string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	LastUsedAt *time.Time `json:"last_used_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
}

// Active returns true if the key hasn't been revoked yet.
func (k *ApiKey) Active() bool {
	return k.RevokedAt == nil
}
