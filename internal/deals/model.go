// Package deals — lightweight CRM pipeline. Every serious SMB CPaaS
// (Wati, DoubleTick, HubSpot Starter) ships a kanban of opportunities;
// this is our parity implementation.
//
// A Pipeline is a named set of ordered Stages ("New Lead" → "Qualified"
// → "Proposal Sent" → "Won"/"Lost"). A Deal lives in one Stage at a
// time and carries a value (in the tenant's default currency), an
// optional contact reference, and free-text notes. Deals never leave a
// pipeline — closing them just moves to the terminal Won/Lost stage.
package deals

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/datatypes"
)

// Pipeline is a named ordered set of stages the deals move through.
// Every tenant gets a default pipeline seeded on first use.
type Pipeline struct {
	database.Base
	OwnerID string `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Name    string `gorm:"type:varchar(100);not null" json:"name"`

	// Stages is a JSON []string preserving order: index 0 is the first
	// column, index len-1 the last. Two special values, "won" and
	// "lost", when present must occupy the last two slots — the runner
	// treats them as terminal.
	Stages datatypes.JSON `gorm:"type:json;not null;default:'[]'" json:"stages"`

	IsDefault bool `gorm:"default:false;index" json:"is_default"`
}

// Deal is one opportunity in a pipeline.
type Deal struct {
	database.Base
	OwnerID    string  `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	PipelineID string  `gorm:"type:varchar(36);index;not null" json:"pipeline_id"`
	Title      string  `gorm:"type:varchar(255);not null" json:"title"`
	Stage      string  `gorm:"type:varchar(50);not null;index" json:"stage"`
	Value      float64 `gorm:"not null;default:0" json:"value"`
	Currency   string  `gorm:"type:varchar(3);not null;default:'INR'" json:"currency"`

	// ContactID — the customer we're selling to. Optional; some pipelines
	// track inbound leads before a contact is created.
	ContactID *string `gorm:"type:varchar(36);index" json:"contact_id"`

	// AssignedTo — the sales rep this deal is with. Optional; drives
	// per-rep leaderboards on the dashboard.
	AssignedTo *string `gorm:"type:varchar(36);index" json:"assigned_to"`

	Notes            string     `gorm:"type:text" json:"notes"`
	ExpectedCloseAt  *time.Time `json:"expected_close_at"`
	ClosedAt         *time.Time `json:"closed_at"`
	StageChangedAt   time.Time  `gorm:"not null" json:"stage_changed_at"`
}
