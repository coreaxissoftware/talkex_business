// Package sla runs a background sweep that finds conversations whose
// last inbound message hasn't been answered within the tenant's
// configured SLA window. On breach we emit a notification and fire a
// webhook event so ops can wire pager/slack alerts.
package sla

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// NotifyFn / WebhookFn are injected from main.go — this package can't
// import notifications or webhooks without creating cycles.
type NotifyFn func(ownerID, title, body, link string)
type WebhookFn func(ownerID, event string, payload interface{})

var (
	notifyFn  NotifyFn
	webhookFn WebhookFn
	fnMu      sync.RWMutex
)

func RegisterNotifier(f NotifyFn)   { fnMu.Lock(); notifyFn = f; fnMu.Unlock() }
func RegisterWebhook(f WebhookFn)   { fnMu.Lock(); webhookFn = f; fnMu.Unlock() }

// alerted tracks conversation IDs we've already fired for so a single
// breach doesn't emit an alert every minute. Cleared 12h after breach
// so a re-breach after a lull still alerts.
var (
	alertedMu sync.Mutex
	alerted   = map[string]time.Time{}
)

// PrefsFetcher lets the sweeper read each owner's SLA threshold without
// importing settings. Returns 0 to skip.
type PrefsFetcher func(ownerID string) (thresholdMins int)

var prefsFetcher PrefsFetcher

func RegisterPrefsFetcher(f PrefsFetcher) { prefsFetcher = f }

// row is the minimum data the sweep needs — kept as a plain struct so
// we can scan into it without importing the conversations model.
type row struct {
	ID            string
	OwnerID       string
	LastInboundAt *time.Time
	LastOutboundAt *time.Time
}

// Sweep runs one pass. Idempotent — safe to call from a ticker.
func Sweep(db *gorm.DB) {
	if prefsFetcher == nil {
		return
	}
	now := time.Now()

	// Reap old alerted entries so re-breaches after a lull re-alert.
	alertedMu.Lock()
	for id, t := range alerted {
		if now.Sub(t) > 12*time.Hour {
			delete(alerted, id)
		}
	}
	alertedMu.Unlock()

	var rows []row
	if err := db.Raw(
		`SELECT id, owner_id, last_inbound_at, last_outbound_at
		 FROM conversations
		 WHERE last_inbound_at IS NOT NULL
		   AND (last_outbound_at IS NULL OR last_outbound_at < last_inbound_at)`,
	).Scan(&rows).Error; err != nil {
		log.Printf("sla: sweep query failed: %v", err)
		return
	}

	fnMu.RLock()
	notify := notifyFn
	webhook := webhookFn
	fnMu.RUnlock()

	for _, r := range rows {
		threshold := prefsFetcher(r.OwnerID)
		if threshold <= 0 || r.LastInboundAt == nil {
			continue
		}
		elapsed := now.Sub(*r.LastInboundAt)
		if elapsed < time.Duration(threshold)*time.Minute {
			continue
		}
		alertedMu.Lock()
		_, seen := alerted[r.ID]
		if !seen {
			alerted[r.ID] = now
		}
		alertedMu.Unlock()
		if seen {
			continue
		}
		if notify != nil {
			notify(r.OwnerID, "SLA breach",
				"A conversation has been waiting for a reply beyond your SLA threshold.",
				"/conversations")
		}
		if webhook != nil {
			webhook(r.OwnerID, "sla.breached", map[string]interface{}{
				"conversation_id": r.ID,
				"elapsed_minutes": int(elapsed.Minutes()),
			})
		}
	}
}

// Start launches the periodic sweeper. Interval of 1 minute is a good
// balance between alert latency and DB load.
func Start(db *gorm.DB, interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			Sweep(db)
		}
	}()
}
