// Package platformenv — DB-backed platform provider secrets, live-
// swappable from a superadmin UI without editing /etc files or
// restarting the service.
//
// The platform ships with every provider readable from
// /etc/talkex-business.env on the systemd host — fine for the very
// first boot, but requires SSH + a text editor + a service restart
// every time an operator wants to wire up Mailgun, MSG91, Razorpay,
// a new OAuth provider, or rotate any of the above.
//
// This package layers a DB-backed override on top of the env:
//   * writes go into platform_settings (one row per key)
//   * reads (via Get) return the DB row when present, else fall back
//     to the process's env
//   * the change is live for the next Get call — Mailgun / MSG91 /
//     Fast2SMS / OAuth / Razorpay / Anthropic / Sentry all read
//     through Get, so a save from the UI picks up on the next request
//     without a systemctl restart
//
// Auth: only a superadmin can list / mutate these rows. Values are
// masked in the list response so a compromised support session can
// never exfiltrate the raw secrets — the API returns only the
// last-4 characters (or "••••" for shorter strings).
package platformenv

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// Setting is one row per platform-level env override.
type Setting struct {
	database.Base
	// Key is the env-var name it overrides — e.g. "MAILGUN_API_KEY".
	// Uppercased at Set time so casing typos can't create ghost
	// duplicates.
	Key string `gorm:"type:varchar(80);uniqueIndex;not null" json:"key"`

	// Value is stored in cleartext. On a Postgres deployment we're
	// behind TLS-at-rest via the disk encryption at the VPS level;
	// a proper KMS integration is a later add. Value is NEVER
	// returned to the frontend — only a mask for confirmation.
	Value string `gorm:"type:text;not null" json:"-"`

	// SetBy — user_id of the superadmin who wrote this row. Kept for
	// the audit trail without needing a separate audit lookup.
	SetBy string `gorm:"type:varchar(36)" json:"set_by,omitempty"`

	// LastRotatedAt separate from UpdatedAt because updating just to
	// re-verify a value shouldn't reset the "how old is this secret"
	// clock on the UI.
	LastRotatedAt time.Time `json:"last_rotated_at"`
}

// TableName pins to platform_settings so future migrations are grep-
// friendly.
func (Setting) TableName() string { return "platform_settings" }
