// Package contacts — shared across every channel (per CONTEXT.md:
// Contacts/Templates/Campaigns/Analytics are shared; only the Channel
// Connector differs).
//
// OptedIn / OptedInAt mirror the consumer app's business_optins table —
// the consent gate required before a business can message a contact
// outside an open conversation window.
// LastInboundAt drives has_open_conversation_window() — 24h customer-
// service-window (see CONTEXT.md).
package contacts

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/datatypes"
)

type Contact struct {
	database.Base
	OwnerID       string          `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	PhoneNumber   string          `gorm:"type:varchar(20);index;not null" json:"phone_number"`
	Name          *string         `gorm:"type:varchar(255)" json:"name"`
	Email         *string         `gorm:"type:varchar(255)" json:"email"`
	Tags          datatypes.JSON  `gorm:"type:json;default:'[]'" json:"tags"`
	CustomFields  datatypes.JSON  `gorm:"type:json;default:'{}'" json:"custom_fields"`
	OptedIn         bool            `gorm:"default:false;not null" json:"opted_in"`
	OptedInAt       *time.Time      `json:"opted_in_at"`
	LastInboundAt   *time.Time      `json:"last_inbound_at"`
	FallbackChannel *string         `gorm:"type:varchar(50)" json:"fallback_channel"`
}
