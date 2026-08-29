package flows

import (
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SendFn sends an outbound message. Injected by main.go so this package
// avoids importing conversations/templates (which import from us via
// automation-style hooks).
type SendFn func(ownerID, contactID, channel, body string, templateID *string) error

// AssignFn assigns a conversation to an agent. Optional.
type AssignFn func(ownerID, contactID string, agentUserID string) error

// TagFn adds a tag to a contact. Optional.
type TagFn func(ownerID, contactID, tag string) error

var (
	sendFn   SendFn
	assignFn AssignFn
	tagFn    TagFn
	fnMu     sync.RWMutex
)

func RegisterSender(f SendFn)     { fnMu.Lock(); sendFn = f; fnMu.Unlock() }
func RegisterAssigner(f AssignFn) { fnMu.Lock(); assignFn = f; fnMu.Unlock() }
func RegisterTagger(f TagFn)      { fnMu.Lock(); tagFn = f; fnMu.Unlock() }

// StartRun kicks off a fresh flow execution for a contact. Returns the
// run state so callers can persist it or debug.
func StartRun(db *gorm.DB, f *Flow, contactID, channel string) (*RunState, error) {
	steps, err := f.GetSteps()
	if err != nil {
		return nil, err
	}

	rs := &RunState{
		OwnerID:       f.OwnerID,
		FlowID:        f.ID,
		ContactID:     contactID,
		Channel:       channel,
		CurrentStepID: f.FirstStepID,
		Status:        "active",
	}
	if err := db.Create(rs).Error; err != nil {
		return nil, err
	}
	BumpRun(db, f.ID)

	// Execute inline until we hit a wait/branch or the end.
	advance(db, f, rs, steps, "")
	return rs, nil
}

// HandleInbound checks for any active RunState waiting on the contact's
// reply (branch step), and advances it based on the body.
func HandleInbound(db *gorm.DB, ownerID, contactID, body string) {
	var runs []RunState
	if err := db.Where("owner_id = ? AND contact_id = ? AND status = ?", ownerID, contactID, "active").
		Find(&runs).Error; err != nil {
		return
	}
	for i := range runs {
		var f Flow
		if err := db.Where("id = ?", runs[i].FlowID).First(&f).Error; err != nil {
			continue
		}
		steps, err := f.GetSteps()
		if err != nil {
			continue
		}
		advance(db, &f, &runs[i], steps, body)
	}
}

// SweepWaiting advances any run whose wait timer has elapsed.
func SweepWaiting(db *gorm.DB) {
	now := time.Now().Unix()
	var runs []RunState
	if err := db.Where("status = ? AND waiting_until IS NOT NULL AND waiting_until <= ?", "active", now).
		Find(&runs).Error; err != nil {
		return
	}
	for i := range runs {
		var f Flow
		if err := db.Where("id = ?", runs[i].FlowID).First(&f).Error; err != nil {
			continue
		}
		steps, err := f.GetSteps()
		if err != nil {
			continue
		}
		runs[i].WaitingUntil = nil
		advance(db, &f, &runs[i], steps, "")
	}
}

// advance runs steps forward from the current position. `inbound` is
// non-empty only when driven by an inbound message (used for branch).
func advance(db *gorm.DB, f *Flow, rs *RunState, steps map[string]Step, inbound string) {
	for i := 0; i < 32; i++ { // guard against infinite loops
		if rs.CurrentStepID == "" {
			complete(db, f, rs)
			return
		}
		step, ok := steps[rs.CurrentStepID]
		if !ok {
			complete(db, f, rs)
			return
		}

		fnMu.RLock()
		send := sendFn
		assign := assignFn
		tag := tagFn
		fnMu.RUnlock()

		next := step.NextStepID

		switch step.Type {
		case "send_message":
			if send != nil && step.OutputText != "" {
				if err := send(rs.OwnerID, rs.ContactID, rs.Channel, step.OutputText, nil); err != nil {
					log.Printf("flow %s: send failed: %v", f.ID, err)
				}
			}
		case "send_template":
			if send != nil && step.TemplateID != "" {
				tid := step.TemplateID
				_ = send(rs.OwnerID, rs.ContactID, rs.Channel, "", &tid)
			}
		case "wait":
			if step.WaitMinutes > 0 {
				// Persist wait — SweepWaiting will pick it up.
				until := time.Now().Add(time.Duration(step.WaitMinutes) * time.Minute).Unix()
				rs.WaitingUntil = &until
				rs.CurrentStepID = step.NextStepID
				db.Save(rs)
				return
			}
		case "branch":
			if inbound == "" {
				// No reply yet — pause and wait for HandleInbound.
				db.Save(rs)
				return
			}
			if step.BranchKeyword != "" &&
				strings.Contains(strings.ToLower(inbound), strings.ToLower(step.BranchKeyword)) {
				next = step.BranchYesID
			} else {
				next = step.BranchNoID
			}
			// Consume inbound so a chain of branches doesn't re-match.
			inbound = ""
		case "assign_agent":
			if assign != nil && step.AgentUserID != "" {
				_ = assign(rs.OwnerID, rs.ContactID, step.AgentUserID)
			}
		case "add_tag":
			if tag != nil && step.TagName != "" {
				_ = tag(rs.OwnerID, rs.ContactID, step.TagName)
			}
		case "end":
			complete(db, f, rs)
			return
		}

		rs.CurrentStepID = next
	}
	// Loop budget exceeded — safety stop.
	rs.Status = "failed"
	db.Save(rs)
}

func complete(db *gorm.DB, f *Flow, rs *RunState) {
	rs.Status = "completed"
	rs.CurrentStepID = ""
	rs.WaitingUntil = nil
	db.Save(rs)
	BumpComplete(db, f.ID)
}

// StartSweeper polls for waiting runs every interval.
func StartSweeper(db *gorm.DB, interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			SweepWaiting(db)
		}
	}()
}
