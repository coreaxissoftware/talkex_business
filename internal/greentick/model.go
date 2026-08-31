// Package greentick — WhatsApp Official Business Account (green
// checkmark) verification workflow tracker.
//
// The application itself happens on Meta's Business Manager UI. This
// package tracks progress through the (currently 6) checklist items
// the tenant has to complete before Meta will grant the badge, exposes
// a status endpoint the dashboard renders as a progress bar, and stamps
// the final approved/rejected verdict when the tenant confirms it.
package greentick

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// Status transitions:
//
//   not_started → in_progress → submitted → approved
//                                        ↘ rejected
//
// Any state can move to "in_progress" again after a rejection so the
// tenant can iterate on the application.
const (
	StatusNotStarted = "not_started"
	StatusInProgress = "in_progress"
	StatusSubmitted  = "submitted"
	StatusApproved   = "approved"
	StatusRejected   = "rejected"
)

// Application is the tenant's row in this tracker — one per owner.
type Application struct {
	database.Base
	OwnerID string `gorm:"type:varchar(36);uniqueIndex;not null" json:"owner_id"`
	Status  string `gorm:"type:varchar(20);not null;default:'not_started'" json:"status"`

	// Checklist — each field marks a green-tick prerequisite:
	//   NotableBrand: press mentions / trademark on file
	//   OrgWebsite:   business website live and matches brand
	//   Meta200Msg:   200+ conversations in the last 90 days
	//   MetaTier2:    Meta messaging limit tier 2 or higher
	//   BusinessVerified: Meta Business Verification complete
	//   TrademarkRefs: three third-party news mentions
	NotableBrand    bool `gorm:"default:false" json:"notable_brand"`
	OrgWebsite      bool `gorm:"default:false" json:"org_website"`
	Meta200Msg      bool `gorm:"default:false" json:"meta_200_msg"`
	MetaTier2       bool `gorm:"default:false" json:"meta_tier2"`
	BusinessVerified bool `gorm:"default:false" json:"business_verified"`
	TrademarkRefs   bool `gorm:"default:false" json:"trademark_refs"`

	SubmittedAt  *time.Time `json:"submitted_at,omitempty"`
	DecidedAt    *time.Time `json:"decided_at,omitempty"`
	MetaCaseID   string     `gorm:"type:varchar(80)" json:"meta_case_id,omitempty"`
	RejectReason string     `gorm:"type:varchar(500)" json:"reject_reason,omitempty"`
}

// Progress returns the fraction of checklist items completed (0-1).
func (a *Application) Progress() float64 {
	total := 6.0
	done := 0.0
	if a.NotableBrand {
		done++
	}
	if a.OrgWebsite {
		done++
	}
	if a.Meta200Msg {
		done++
	}
	if a.MetaTier2 {
		done++
	}
	if a.BusinessVerified {
		done++
	}
	if a.TrademarkRefs {
		done++
	}
	return done / total
}
