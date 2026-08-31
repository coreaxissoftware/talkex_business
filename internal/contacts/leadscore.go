package contacts

import (
	"log"

	"gorm.io/gorm"
)

// Lead scoring — a small, honest scoring model designed for SMB CRM
// parity with Wati / DoubleTick / Interakt. Score is 0-100, capped.
// Positive events add points; the value stored on Contact.LeadScore is
// the running total. A nightly decay job (SweepDecay) subtracts a
// small amount so cold contacts drift back down toward 0.
//
// Point weights are deliberately soft — the goal is to rank contacts
// against each other, not to be arithmetically precise. Tune per owner
// later via a settings row; the defaults ship a workable model.

// ScoreEvent enumerates the events that move the score.
type ScoreEvent string

const (
	EventInboundMessage    ScoreEvent = "inbound_message"    // +5
	EventOutboundReplied   ScoreEvent = "outbound_replied"   // +8
	EventTemplateOpened    ScoreEvent = "template_opened"    // +3
	EventLinkClicked       ScoreEvent = "link_clicked"       // +7
	EventCsatSubmitted     ScoreEvent = "csat_submitted"     // +10
	EventCampaignConverted ScoreEvent = "campaign_converted" // +20
	EventOptOut            ScoreEvent = "opt_out"            // -30 (hard drop)
)

var scoreWeights = map[ScoreEvent]int{
	EventInboundMessage:    5,
	EventOutboundReplied:   8,
	EventTemplateOpened:    3,
	EventLinkClicked:       7,
	EventCsatSubmitted:     10,
	EventCampaignConverted: 20,
	EventOptOut:            -30,
}

const (
	scoreCeiling = 100
	scoreFloor   = 0
)

// Bump adjusts a contact's LeadScore for the given event. Idempotent
// per-call (the caller decides how often to fire) and clamped to
// [0, 100]. Logs but does not fail-hard on a DB error.
func Bump(db *gorm.DB, ownerID, contactID string, event ScoreEvent) {
	weight, ok := scoreWeights[event]
	if !ok || contactID == "" {
		return
	}
	var c Contact
	if err := db.Where("id = ? AND owner_id = ?", contactID, ownerID).First(&c).Error; err != nil {
		return
	}
	next := c.LeadScore + weight
	if next > scoreCeiling {
		next = scoreCeiling
	}
	if next < scoreFloor {
		next = scoreFloor
	}
	if next == c.LeadScore {
		return
	}
	if err := db.Model(&c).Update("lead_score", next).Error; err != nil {
		log.Printf("leadscore: bump %s(%s) failed: %v", event, contactID, err)
	}
}

// SweepDecay runs the weekly cooldown — every contact's score falls by
// `decayPoints` so scored-but-cold contacts drop out of the top of the
// pack over time. Intended to run once a week from cmd/server.
func SweepDecay(db *gorm.DB, decayPoints int) (int, error) {
	if decayPoints <= 0 {
		decayPoints = 5
	}
	res := db.Exec(
		`UPDATE contacts
		 SET lead_score = CASE WHEN lead_score - ? < 0 THEN 0 ELSE lead_score - ? END
		 WHERE lead_score > 0`,
		decayPoints, decayPoints,
	)
	return int(res.RowsAffected), res.Error
}
