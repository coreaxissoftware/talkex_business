package contacts

import (
	"time"

	"gorm.io/gorm"
)

// LifecycleStage transitions — the same shape every serious CRM ships:
//
//   new       — created within the last 7 days, no inbound yet
//   active    — inbound within the last 30 days
//   dormant   — inbound between 30 and 90 days ago
//   churned   — no inbound for > 90 days, or explicitly opted out
//
// Kept intentionally simple: four states, one time-window rule per
// transition, evaluated in a nightly sweep. Real behavioural triggers
// (paid → churned on cancel event, dormant → active on click) can layer
// on top later without changing this shape.
const (
	StageNew     = "new"
	StageActive  = "active"
	StageDormant = "dormant"
	StageChurned = "churned"
)

// EvaluateStage returns the lifecycle stage a contact SHOULD be in
// given its current state and now. Pure — no DB, no mutation, so unit
// tests can pin every branch.
func EvaluateStage(c *Contact, now time.Time) string {
	if !c.OptedIn {
		// Opt-out is a hard drop-out; a re-opt-in flips them back to new.
		return StageChurned
	}
	if c.LastInboundAt == nil {
		if now.Sub(c.CreatedAt) <= 7*24*time.Hour {
			return StageNew
		}
		// No inbound ever + more than a week old — treat as dormant.
		return StageDormant
	}
	age := now.Sub(*c.LastInboundAt)
	switch {
	case age <= 30*24*time.Hour:
		return StageActive
	case age <= 90*24*time.Hour:
		return StageDormant
	default:
		return StageChurned
	}
}

// RefreshOne updates a single contact's LifecycleStage. Idempotent: no
// write happens if the stage hasn't changed.
func RefreshOne(db *gorm.DB, c *Contact) error {
	want := EvaluateStage(c, time.Now())
	if c.LifecycleStage == want {
		return nil
	}
	c.LifecycleStage = want
	return db.Model(c).Update("lifecycle_stage", want).Error
}

// SweepAll refreshes every contact's LifecycleStage. Intended to run
// nightly from cmd/server; safe to call on-demand for reindexing.
// Returns (updated, total).
func SweepAll(db *gorm.DB) (int, int, error) {
	const batchSize = 500
	var offset, total, updated int
	for {
		var batch []Contact
		if err := db.Order("id").Limit(batchSize).Offset(offset).Find(&batch).Error; err != nil {
			return updated, total, err
		}
		if len(batch) == 0 {
			break
		}
		total += len(batch)
		for i := range batch {
			c := &batch[i]
			want := EvaluateStage(c, time.Now())
			if c.LifecycleStage == want {
				continue
			}
			if err := db.Model(c).Update("lifecycle_stage", want).Error; err == nil {
				updated++
			}
		}
		offset += batchSize
		if len(batch) < batchSize {
			break
		}
	}
	return updated, total, nil
}
