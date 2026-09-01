// Package waflows — WhatsApp Interactive Flows (Meta Flows JSON).
//
// Meta launched Flows in 2024: full in-chat multi-step forms, screens,
// and data-collection with no external webviews. Businesses author
// them as Flow JSON (routed screens + components) and reference the
// FlowID from an interactive.flow message.
//
// This module stores the tenant's Flow JSON documents (versioned),
// pushes them to Meta's Flows API on publish, and returns the FlowID
// the messaging engine embeds in outbound interactive.flow messages.
//
// Parity target: Gupshup Flows, MessageBird Forms, Yellow.ai Journey.
package waflows

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/datatypes"
)

// Status transitions:
//
//   draft → published → deprecated
//                    ↘ blocked (Meta rejected)
const (
	StatusDraft      = "draft"
	StatusPublished  = "published"
	StatusDeprecated = "deprecated"
	StatusBlocked    = "blocked"
)

// Category enum matches Meta's Flow categories exactly.
const (
	CategorySignUp        = "SIGN_UP"
	CategorySignIn        = "SIGN_IN"
	CategoryAppointment   = "APPOINTMENT_BOOKING"
	CategoryLead          = "LEAD_GENERATION"
	CategoryShopping      = "SHOPPING"
	CategoryContactSupport = "CONTACT_US"
	CategorySurvey        = "SURVEY"
	CategoryOther         = "OTHER"
)

// WAFlow is one interactive Flow the tenant owns.
type WAFlow struct {
	database.Base
	OwnerID  string `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Name     string `gorm:"type:varchar(255);not null" json:"name"`
	Category string `gorm:"type:varchar(30);not null;default:'OTHER'" json:"category"`
	Status   string `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"`
	Version  int    `gorm:"default:1;not null" json:"version"`

	// FlowJSON — the full Meta Flow JSON document (routing_model + screens).
	// Stored as-is so re-publishing is a byte-for-byte replay.
	FlowJSON datatypes.JSON `gorm:"type:json;not null" json:"flow_json"`

	// MetaFlowID — assigned by Meta on first publish; used in outbound
	// interactive.flow messages.
	MetaFlowID string `gorm:"type:varchar(64);index" json:"meta_flow_id,omitempty"`

	PublishedAt *time.Time `json:"published_at,omitempty"`

	// Endpoint — where Meta will POST data collection responses. When
	// empty, we default to /waflows/inbound/:owner_id on the tenant's
	// TalkEx origin so responses land in the conversation history.
	Endpoint string `gorm:"type:varchar(500)" json:"endpoint"`
}

// FlowResponse is one submission from an end-user completing the Flow.
// Meta POSTs these to the tenant's Endpoint; this row keeps a copy for
// audit and analytics.
type FlowResponse struct {
	database.Base
	OwnerID   string         `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	FlowID    string         `gorm:"type:varchar(36);index;not null" json:"flow_id"`
	ContactID string         `gorm:"type:varchar(36);index" json:"contact_id,omitempty"`
	ScreenID  string         `gorm:"type:varchar(80)" json:"screen_id,omitempty"`
	Data      datatypes.JSON `gorm:"type:json" json:"data"`
}
