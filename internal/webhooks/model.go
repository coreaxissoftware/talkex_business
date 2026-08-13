// Package webhooks — outbound HMAC-signed HTTP callbacks on platform events
// (inbound.message, message.status, campaign.completed …).
//
// Endpoints store a per-endpoint Secret that signs each delivery so
// receivers can verify authenticity. Failed deliveries are logged for
// developer visibility and (later) retry.
package webhooks

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/datatypes"
)

// Event names emitted by other modules. Kept as untyped strings so new
// packages can add events without touching this file.
const (
	EventInboundMessage    = "inbound.message"
	EventMessageStatus     = "message.status"
	EventCampaignCompleted = "campaign.completed"
	EventContactCreated    = "contact.created"
)

type Endpoint struct {
	database.Base
	OwnerID   string         `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	URL       string         `gorm:"type:varchar(500);not null" json:"url"`
	Secret    string         `gorm:"type:varchar(64);not null" json:"-"`
	Events    datatypes.JSON `gorm:"type:json;default:'[]'" json:"events"`
	Active    bool           `gorm:"not null;default:true" json:"active"`
	LastFired *time.Time     `json:"last_fired_at"`
}

type Delivery struct {
	database.Base
	EndpointID   string     `gorm:"type:varchar(36);index;not null" json:"endpoint_id"`
	Event        string     `gorm:"type:varchar(50);not null" json:"event"`
	Payload      string     `gorm:"type:text;not null" json:"payload"`
	StatusCode   int        `gorm:"not null" json:"status_code"`
	Success      bool       `gorm:"not null;index" json:"success"`
	Attempts     int        `gorm:"not null;default:1" json:"attempts"`
	ErrorMessage *string    `gorm:"type:varchar(500)" json:"error_message,omitempty"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
}
