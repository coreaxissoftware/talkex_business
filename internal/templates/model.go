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
}
