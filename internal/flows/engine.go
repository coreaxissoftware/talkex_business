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

// runLocks serializes advance() for a single RunState so that a
// simultaneous inbound message + wait-timer expiry don't both walk
// past the same branch and double-fire send/assign/tag callbacks.
// One lock per RunState.ID; locks live for the process's lifetime,
// which is fine — RunState count is bounded per owner and each entry
// is a few bytes.
var (
	runLocks   = map[string]*sync.Mutex{}
	runLocksMu sync.Mutex
)

func lockRun(runID string) *sync.Mutex {
	runLocksMu.Lock()
	defer runLocksMu.Unlock()
	m, ok := runLocks[runID]
	if !ok {
		m = &sync.Mutex{}
		runLocks[runID] = m
	}
	return m
}

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
	// Serialize on this run's lock (a subsequent inbound could arrive
	// while we're still walking the initial straight-line steps).
	mu := lockRun(rs.ID)
	mu.Lock()
	defer mu.Unlock()
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
		// Serialize on this run's lock — inbound and sweeper may
		// race for the same RunState, and re-load under lock so we
		// see whatever the winner just committed.
		mu := lockRun(runs[i].ID)
		mu.Lock()
		var fresh RunState
		if err := db.Where("id = ?", runs[i].ID).First(&fresh).Error; err != nil {
			mu.Unlock()
			continue
		}
		if fresh.Status != "active" {
			mu.Unlock()
			continue
		}
		var f Flow
		if err := db.Where("id = ?", fresh.FlowID).First(&f).Error; err != nil {
			mu.Unlock()
			continue
		}
		steps, err := f.GetSteps()
		if err != nil {
			mu.Unlock()
			continue
		}
		advance(db, &f, &fresh, steps, body)
		mu.Unlock()
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
		mu := lockRun(runs[i].ID)
		mu.Lock()
		// Re-read under the lock so we see whatever HandleInbound
		// (or a prior sweep) already committed — skip if another
		// worker already advanced this run past its wait.
		var fresh RunState
		if err := db.Where("id = ? AND status = ? AND waiting_until IS NOT NULL AND waiting_until <= ?",
			runs[i].ID, "active", now).First(&fresh).Error; err != nil {
			mu.Unlock()
			continue
		}
		var f Flow
		if err := db.Where("id = ?", fresh.FlowID).First(&f).Error; err != nil {
			mu.Unlock()
			continue
		}
		steps, err := f.GetSteps()
		if err != nil {
			mu.Unlock()
			continue
		}
		fresh.WaitingUntil = nil
		advance(db, &f, &fresh, steps, "")
		mu.Unlock()
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
		case "split":
			// Journey builder — condition on the contact's own state.
			// No inbound required; evaluated immediately against the live
			// contact row. Missing contact routes to the "no" branch.
			contact, err := LoadContact(db, rs.OwnerID, rs.ContactID)
			if err != nil {
				next = step.BranchNoID
			} else if EvaluateSplit(contact, step) {
				next = step.BranchYesID
			} else {
				next = step.BranchNoID
			}
		case "webhook":
			// Fire an outbound POST; non-2xx routes to the "no" branch.
			// Timeout defaults to 5s to keep the engine responsive.
			timeout := time.Duration(step.WebhookTimeoutS) * time.Second
			if timeout <= 0 {
				timeout = 5 * time.Second
			}
			ok := postWebhook(step.WebhookURL, rs.OwnerID, rs.ContactID, step.ID, timeout)
			if ok {
				if step.BranchYesID != "" {
					next = step.BranchYesID
				}
			} else {
				if step.BranchNoID != "" {
					next = step.BranchNoID
				}
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
