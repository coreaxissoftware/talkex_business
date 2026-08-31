package conversations

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/contacts"
)

// Opt-out keyword handling — regulatory baseline every CPaaS ships.
//
// If an inbound message body matches one of these keywords the sender's
// consent is revoked immediately: Contact.OptedIn=false, OptedInAt=nil,
// and a "opted-out" label is stamped on the conversation for the agent
// to see. Broadcast/campaign runners already refuse contacts where
// OptedIn=false, so no follow-up marketing message can be sent to the
// contact once this fires.
//
// Keywords cover:
//   - WhatsApp Meta guidance: STOP, UNSUBSCRIBE
//   - SMS DLT / TRAI: STOP, START, UNSUB
//   - Global marketing: OPT OUT, CANCEL, END, QUIT, REMOVE
//   - Hindi common: BAND KARO, BAND KARE, ROKO, NAHI

// stopKeywordRe matches a whole-word occurrence of any opt-out phrase,
// case-insensitively. Anchored on word boundaries so "STOP" catches
// "please stop" but not "stopper" or "unstoppable".
var stopKeywordRe = regexp.MustCompile(
	`(?i)\b(stop|unsubscribe|unsub|opt[\s-]*out|cancel|end|quit|remove|` +
		`band[\s-]*karo|band[\s-]*kare|roko|nahi[\s-]*chahiye)\b`,
)

// OptOutLabel is the label stamped on a conversation whose contact just
// opted out. Surfaces in the inbox UI so agents know not to reply with
// marketing content.
const OptOutLabel = "opted-out"

// evaluateOptOut runs on every inbound message. If the body contains an
// opt-out keyword the contact's consent is revoked and the conversation
// is labelled. Failures log but never propagate — a broken opt-out sweep
// must not swallow the inbound message itself.
func evaluateOptOut(db *gorm.DB, ownerID string, msg *Message, conv *Conversation) {
	if msg == nil || msg.Body == "" {
		return
	}
	if !stopKeywordRe.MatchString(strings.TrimSpace(msg.Body)) {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("optout: panic recovered: %v", r)
		}
	}()

	// 1. Revoke contact consent.
	if err := db.Model(&contacts.Contact{}).
		Where("id = ? AND owner_id = ?", conv.ContactID, ownerID).
		Updates(map[string]interface{}{
			"opted_in":    false,
			"opted_in_at": nil,
		}).Error; err != nil {
		log.Printf("optout: failed to revoke consent for %s: %v", conv.ContactID, err)
		return
	}

	// 2. Stamp the label on the conversation (idempotent — dedup on merge).
	var existing []string
	if conv.Labels != "" {
		_ = json.Unmarshal([]byte(conv.Labels), &existing)
	}
	seen := make(map[string]bool, len(existing))
	for _, l := range existing {
		seen[l] = true
	}
	if !seen[OptOutLabel] {
		existing = append(existing, OptOutLabel)
		if b, err := json.Marshal(existing); err == nil {
			conv.Labels = string(b)
			_ = db.Model(conv).Update("labels", string(b)).Error
		}
	}

	log.Printf("optout: contact=%s owner=%s revoked via inbound match (%q)",
		conv.ContactID, ownerID, truncate(msg.Body, 40))

	// Lead score penalty — a hard opt-out drops the contact's score
	// sharply so they fall out of top-lead reports immediately.
	contacts.Bump(db, ownerID, conv.ContactID, contacts.EventOptOut)

	// 3. Note the moment for audit — piggyback on the message's own timestamp
	// so downstream analytics can chart opt-out velocity.
	_ = time.Now() // reserved for future audit-log integration
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
