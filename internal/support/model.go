// Package support — customer support surface: help articles are static
// on the frontend, but users can also file tickets that land here for a
// human to answer (this MVP: they just accumulate; a proper ops flow
// comes later).
package support

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

const (
	TicketStatusOpen       = "open"
	TicketStatusInProgress = "in_progress"
	TicketStatusResolved   = "resolved"

	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

type Ticket struct {
	database.Base
	OwnerID    string     `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Subject    string     `gorm:"type:varchar(255);not null" json:"subject"`
	Body       string     `gorm:"type:text;not null" json:"body"`
	Priority   string     `gorm:"type:varchar(10);not null;default:'normal'" json:"priority"`
	Status     string     `gorm:"type:varchar(20);not null;default:'open'" json:"status"`
	ResolvedAt *time.Time `json:"resolved_at"`
}
