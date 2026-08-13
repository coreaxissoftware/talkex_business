// Package automation — keyword-triggered auto-replies today; will grow
// into the full no-code flow builder in later phases (per CONTEXT.md).
//
// A Rule matches inbound messages case-insensitively against any of its
// TriggerKeywords and, when matched, sends an outbound message — either
// a free-form Body or a linked TemplateID (needed when the 24h window is
// closed, though for auto-replies fired from an inbound the window is
// always open by definition).
package automation

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/datatypes"
)

type Rule struct {
	database.Base
	OwnerID         string         `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	TriggerKeywords datatypes.JSON `gorm:"type:json;default:'[]'" json:"trigger_keywords"`
	MatchType       string         `gorm:"type:varchar(20);not null;default:'contains'" json:"match_type"` // contains | exact | starts_with
	ResponseBody    string         `gorm:"type:text;not null" json:"response_body"`
	TemplateID      *string        `gorm:"type:varchar(36)" json:"template_id,omitempty"`
	Active          bool           `gorm:"not null;default:true" json:"active"`
	FireCount       int64          `gorm:"not null;default:0" json:"fire_count"`
}
