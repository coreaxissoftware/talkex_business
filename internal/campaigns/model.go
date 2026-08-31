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
	StatusDraft            = "draft"
	StatusScheduled        = "scheduled"
	StatusRunning          = "running"
	StatusCompleted        = "completed"
	StatusFailed           = "failed"
	StatusCancelled        = "cancelled"
	StatusPaused           = "paused"
	StatusPendingApproval  = "pending_approval"
	StatusRejected         = "rejected"
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

	// ListID — optional reference to the contact list used to create this campaign.
	ListID *string `gorm:"type:varchar(36)" json:"list_id"`

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

	// Cost tracking
	TotalCost float64 `gorm:"default:0" json:"total_cost"`

	// Maker-checker approval fields
	ApprovalRequired bool       `gorm:"default:false;not null" json:"approval_required"`
	ApprovedBy       *string    `gorm:"type:varchar(36)" json:"approved_by"`
	ApprovedAt       *time.Time `json:"approved_at"`
	RejectedBy       *string    `gorm:"type:varchar(36)" json:"rejected_by"`
	RejectedAt       *time.Time `json:"rejected_at"`
	RejectionReason  *string    `gorm:"type:varchar(500)" json:"rejection_reason"`

	// A/B testing — when Variants is populated, the runner deterministically
	// assigns each recipient to variant i = index(contact_id) % len(variants).
	// The chosen template overrides the campaign's TemplateID for that
	// send only; per-variant counters roll up into VariantStats.
	Variants     datatypes.JSON `gorm:"type:json;default:'[]'" json:"variants"`      // []string of template IDs
	VariantStats datatypes.JSON `gorm:"type:json;default:'{}'" json:"variant_stats"` // map[string]{sent,delivered,read,clicked}
}

// VariantStat holds per-variant roll-ups the analytics page renders as
// a bar chart. Keys in Campaign.VariantStats are the template IDs from
// Campaign.Variants.
type VariantStat struct {
	Sent      int `json:"sent"`
	Delivered int `json:"delivered"`
	Read      int `json:"read"`
	Clicked   int `json:"clicked"`
}
