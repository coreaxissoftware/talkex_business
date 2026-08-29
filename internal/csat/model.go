// Package csat collects post-conversation Customer Satisfaction ratings.
// A rating captures a 1-5 score plus optional freeform comment, keyed to
// the conversation (and optionally the assigned agent) so support teams
// can measure quality per-agent and over time.
package csat

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type Rating struct {
	database.Base
	OwnerID        string  `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	ConversationID string  `gorm:"type:varchar(36);index;not null" json:"conversation_id"`
	ContactID      string  `gorm:"type:varchar(36);index" json:"contact_id"`
	AgentUserID    *string `gorm:"type:varchar(36);index" json:"agent_user_id,omitempty"`
	Score          int     `gorm:"not null" json:"score"` // 1-5
	Comment        string  `gorm:"type:text" json:"comment"`
	Channel        string  `gorm:"type:varchar(32);index" json:"channel"`
}
