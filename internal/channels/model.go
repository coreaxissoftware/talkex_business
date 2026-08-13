// Package channels — one row per (owner, channel-kind) tracking whether
// the owner has enabled that channel and their provider-specific config.
//
// The actual send/receive integrations live in the sub-packages
// (channels/talkex, channels/whatsapp) — this parent package is just
// the enable/config surface exposed to the dashboard.
package channels

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/datatypes"
)

// Kind is the fixed set of channels TalkEx Business supports.
type Kind string

const (
	KindTalkEx   Kind = "talkex"
	KindWhatsApp Kind = "whatsapp"
	KindTelegram Kind = "telegram"
	KindEmail    Kind = "email"
	KindSMS      Kind = "sms"
	KindRCS      Kind = "rcs"
)

// AllKinds — canonical order used by the dashboard grid. Sub-packages
// implement each connector; a Kind present here without a connector is
// simply a "Coming soon" tile.
var AllKinds = []Kind{
	KindTalkEx,
	KindWhatsApp,
	KindTelegram,
	KindEmail,
	KindSMS,
	KindRCS,
}

// Config is the enable-state + provider-config row.
type Config struct {
	database.Base
	OwnerID    string         `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Kind       Kind           `gorm:"type:varchar(20);not null" json:"kind"`
	Enabled    bool           `gorm:"not null;default:false" json:"enabled"`
	Config     datatypes.JSON `gorm:"type:json;default:'{}'" json:"config"`
	VerifiedAt *time.Time     `json:"verified_at"`
}
