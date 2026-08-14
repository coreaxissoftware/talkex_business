// Package campaigns — bulk-send flows built on top of templates + contacts.
// A campaign snapshots which template to send, which contact list to send it
// to, when to send it, and rolls up the per-message delivery stats so the
// dashboard doesn't have to join across message rows every time.
package campaigns

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/datatypes"
)

// Campaign status transitions:
//   draft → scheduled → running → completed
//                              ↘ failed
//                              ↘ cancelled
const (
	StatusDraft     = "draft"
	StatusScheduled = "scheduled"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
	StatusPaused    = "paused"
)

type Campaign struct {
	database.Base
	OwnerID     string     `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	TemplateID  string     `gorm:"type:varchar(36);not null" json:"template_id"`
	Channel     string     `gorm:"type:varchar(50);not null" json:"channel"`
	Status      string     `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`

	// ContactIDs is stored as JSON — a campaign snapshots the recipient list
	// at creation time so later edits to contacts (add/remove/opt-out) don't
	// silently change what a scheduled campaign will send to.
	ContactIDs datatypes.JSON `gorm:"type:json;default:'[]'" json:"contact_ids"`

	// Roll-up counters updated as messages are processed.
	TotalCount     int `gorm:"default:0" json:"total_count"`
	SentCount      int `gorm:"default:0" json:"sent_count"`
	DeliveredCount int `gorm:"default:0" json:"delivered_count"`
	ReadCount      int `gorm:"default:0" json:"read_count"`
	FailedCount    int `gorm:"default:0" json:"failed_count"`
}
