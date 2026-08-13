// Package billing — subscription plans, current subscription, and (later)
// invoices + GST. This first pass ships enough for the /billing page to
// show the current plan and let a user "upgrade" (self-serve, no real
// payment provider yet — that's a Phase 3 gap noted in CONTEXT.md).
package billing

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// PlanID is a fixed enum of the plans we sell. New tiers get added here
// and to `AllPlans` below; the frontend renders whatever this list holds.
type PlanID string

const (
	PlanStarter PlanID = "starter"
	PlanGrowth  PlanID = "growth"
	PlanScale   PlanID = "scale"
)

// Plan is the marketing description of what each PlanID includes. Kept
// in code (not a table) because the shape is small and we want a plan
// change to require a code review, not a stray SQL update.
type Plan struct {
	ID                PlanID   `json:"id"`
	Name              string   `json:"name"`
	PriceINRPerMonth  int      `json:"price_inr_per_month"`
	IncludedMessages  int      `json:"included_messages"`
	OveragePerMsgINR  float64  `json:"overage_per_msg_inr"`
	MaxContacts       int      `json:"max_contacts"`
	MaxTeamMembers    int      `json:"max_team_members"`
	Features          []string `json:"features"`
	Recommended       bool     `json:"recommended"`
}

// AllPlans is the source of truth for /billing/plans. Order matters —
// the frontend renders left-to-right in this order.
var AllPlans = []Plan{
	{
		ID:               PlanStarter,
		Name:             "Starter",
		PriceINRPerMonth: 0,
		IncludedMessages: 1000,
		OveragePerMsgINR: 0.35,
		MaxContacts:      500,
		MaxTeamMembers:   2,
		Features: []string{
			"1 channel (TalkEx)",
			"Contacts, Templates, Campaigns",
			"2-way inbox (24h window)",
			"1000 messages/month included",
			"Email support",
		},
	},
	{
		ID:               PlanGrowth,
		Name:             "Growth",
		PriceINRPerMonth: 2499,
		IncludedMessages: 15000,
		OveragePerMsgINR: 0.25,
		MaxContacts:      10000,
		MaxTeamMembers:   10,
		Recommended:      true,
		Features: []string{
			"All channels (TalkEx + WhatsApp)",
			"15,000 messages/month included",
			"Automation rules (unlimited)",
			"API keys + Webhooks",
			"Priority email + chat support",
		},
	},
	{
		ID:               PlanScale,
		Name:             "Scale",
		PriceINRPerMonth: 9999,
		IncludedMessages: 100000,
		OveragePerMsgINR: 0.15,
		MaxContacts:      0, // unlimited
		MaxTeamMembers:   0, // unlimited
		Features: []string{
			"Everything in Growth",
			"100,000 messages/month included",
			"Unlimited contacts + team members",
			"Custom template categories",
			"Dedicated account manager + SLA",
		},
	},
}

// PlanByID returns the Plan for an ID (or nil if unknown — treated as an
// unrecognised plan, caller decides what to do).
func PlanByID(id PlanID) *Plan {
	for i := range AllPlans {
		if AllPlans[i].ID == id {
			return &AllPlans[i]
		}
	}
	return nil
}

// Subscription is one row per owner — the plan they're currently on and
// when the current billing period started. Everyone starts on PlanStarter.
type Subscription struct {
	database.Base
	OwnerID          string    `gorm:"type:varchar(36);uniqueIndex;not null" json:"owner_id"`
	Plan             PlanID    `gorm:"type:varchar(20);not null;default:'starter'" json:"plan"`
	PeriodStart      time.Time `gorm:"not null" json:"period_start"`
	MessagesUsed     int       `gorm:"not null;default:0" json:"messages_used"`
	Status           string    `gorm:"type:varchar(20);not null;default:'active'" json:"status"` // active | past_due | cancelled
	CurrentInvoiceID *string   `gorm:"type:varchar(36)" json:"current_invoice_id"`
}

// Invoice — one issued at each period rollover. Simple flat schema for
// the MVP; a proper accounting layer comes later.
type Invoice struct {
	database.Base
	OwnerID       string    `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Plan          PlanID    `gorm:"type:varchar(20);not null" json:"plan"`
	PeriodStart   time.Time `gorm:"not null" json:"period_start"`
	PeriodEnd     time.Time `gorm:"not null" json:"period_end"`
	MessagesUsed  int       `gorm:"not null" json:"messages_used"`
	AmountINR     float64   `gorm:"not null" json:"amount_inr"`
	Status        string    `gorm:"type:varchar(20);not null;default:'paid'" json:"status"` // paid | pending | failed
	PaidAt        *time.Time `json:"paid_at"`
}
