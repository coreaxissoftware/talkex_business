// Package templates — shared engine, per-channel mapping (per CONTEXT.md).
//
// Category is flagged as MVP-blocking, non-negotiable: it drives Meta's
// 2025 per-template pricing model and needs to be wired into Billing, not
// bolted on later.
package templates

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/datatypes"
)

type TemplateCategory string

const (
	CategoryMarketing      TemplateCategory = "marketing"
	CategoryUtility        TemplateCategory = "utility"
	CategoryAuthentication TemplateCategory = "authentication"
)

type TemplateStatus string

const (
	StatusDraft         TemplateStatus = "draft"
	StatusPendingReview TemplateStatus = "pending_review"
	StatusApproved      TemplateStatus = "approved"
	StatusRejected      TemplateStatus = "rejected"
)

type MessageTemplate struct {
	database.Base
	OwnerID   string           `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Name      string           `gorm:"type:varchar(255);not null" json:"name"`
	Category  TemplateCategory `gorm:"type:varchar(20);not null" json:"category"`
	Channel   string           `gorm:"type:varchar(50);not null" json:"channel"`
	Body      string           `gorm:"type:text;not null" json:"body"`
	Variables datatypes.JSON   `gorm:"type:json;default:'[]'" json:"variables"`
	Status    TemplateStatus   `gorm:"type:varchar(20);default:draft;not null" json:"status"`

	// Interactive elements — WhatsApp Cloud API supports quick-reply
	// buttons (up to 3), list-picker sections, and CTA buttons. Kept
	// as JSON so the connector can serialize per-provider without a
	// migration for each new element type.
	Buttons  datatypes.JSON `gorm:"type:json;default:'[]'" json:"buttons"`
	ListRows datatypes.JSON `gorm:"type:json;default:'[]'" json:"list_rows"`
	Header   string         `gorm:"type:varchar(60)" json:"header"`
	Footer   string         `gorm:"type:varchar(60)" json:"footer"`

	// Media attachment — a media_id from /media/upload OR an external
	// URL. MediaType is one of image | video | document | audio; kept
	// blank for text-only templates.
	MediaType string `gorm:"type:varchar(20)" json:"media_type"`
	MediaURL  string `gorm:"type:varchar(512)" json:"media_url"`

	// Meta / WhatsApp submission tracking — set once the template is
	// pushed to Meta for approval; ExternalStatus lags MetaStatus while
	// we poll for the review result.
	SubmittedAt    *int64 `json:"submitted_at,omitempty"`
	ExternalRef    string `gorm:"type:varchar(120)" json:"external_ref"`
	ExternalStatus string `gorm:"type:varchar(30)" json:"external_status"`
	RejectReason   string `gorm:"type:varchar(255)" json:"reject_reason"`
}
