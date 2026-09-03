package talkex

// Inbound poller — bridges TalkEx DM replies into the TalkEx Business
// Conversations inbox.
//
// TalkEx's Bulk API key is send-only by design (see main.py:203
// require_api_key + main.py:2033 bulk_send_message). To surface a
// customer's reply inside TalkEx Business we hit a companion read
// endpoint:
//
//   GET /api/v1/inbound?since=<unix_seconds>&limit=100
//     Header:   Authorization: Bearer <talkex_api_key>
//     Returns:  {
//       "messages": [
//         {"message_id","chat_id","from_username","text","kind",
//          "created_at","seq"},
//         ...
//       ],
//       "next_since": <float unix seconds>,
//       "count": <int>
//     }
//
// This file assumes that endpoint exists — the exact Python patch for
// TalkEx Messenger's main.py ships alongside this in docs/TALKEX_INBOUND_ENDPOINT.md.
//
// Poll loop, per merchant per registered TalkEx channel config:
//
//   1. GET /api/v1/inbound?since=<last_seen>&limit=100
//   2. For each row where from_username != our own username:
//        a. resolve Contact by (owner_id, talkex_username) — create when
//           missing so a customer messaging us for the first time doesn't
//           get dropped
//        b. RecordInbound() into the Conversations engine — this fires
//           the same downstream hooks (automation, AI auto-tag, SSE
//           notifications) as any other channel
//   3. Persist next_since as the new checkpoint
//   4. Sleep pollInterval, loop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// pollInterval — 30s is a reasonable trade-off between "customer sees
// the agent's follow-up quickly" and "we don't hammer TalkEx". The
// TalkEx endpoint's 120/min cap gives us 4x headroom at this rate.
const pollInterval = 30 * time.Second

// PollerState tracks the per-tenant checkpoint. Kept as its own table
// so a restart doesn't cause a duplicate flood.
type PollerState struct {
	OwnerID    string    `gorm:"type:varchar(36);primaryKey" json:"owner_id"`
	LastSince  float64   `gorm:"not null;default:0" json:"last_since"`
	LastRunAt  time.Time `json:"last_run_at"`
	LastError  string    `gorm:"type:varchar(500)" json:"last_error"`
}

// TableName pins to talkex_poller_state so future migrations are readable.
func (PollerState) TableName() string { return "talkex_poller_state" }

// ContactUpserter — supplied by the messaging engine so this file
// doesn't need to import contacts (which would import database in a
// way we'd have to plumb through). Returns the local contact ID for
// the given TalkEx username, creating a lightweight row when new.
type ContactUpserter func(ownerID, talkexUsername string) (contactID string, err error)

// InboundRecorder — dropped in by conversations at wire time.
type InboundRecorder func(ownerID, contactID, channel, body string) error

// ChannelLister — returns every enabled TalkEx channel config the
// engine knows about. Each element is (ownerID, baseURL, apiKey).
type ChannelLister func() []ChannelBinding

type ChannelBinding struct {
	OwnerID string
	BaseURL string
	APIKey  string
}

var (
	upserterMu sync.RWMutex
	upserter   ContactUpserter
	recorderMu sync.RWMutex
	recorder   InboundRecorder
	listerMu   sync.RWMutex
	lister     ChannelLister
)

// RegisterContactUpserter is called from cmd/server/main.go at wire time.
func RegisterContactUpserter(f ContactUpserter) {
	upserterMu.Lock()
	defer upserterMu.Unlock()
	upserter = f
}

// RegisterInboundRecorder wires the conversations-engine hook.
func RegisterInboundRecorder(f InboundRecorder) {
	recorderMu.Lock()
	defer recorderMu.Unlock()
	recorder = f
}

// RegisterChannelLister wires the channels-config enumerator.
func RegisterChannelLister(f ChannelLister) {
	listerMu.Lock()
	defer listerMu.Unlock()
	lister = f
}

// StartPoller kicks off the background loop. Safe to call once at boot.
// The loop runs forever; a nil DB or missing hooks make it a no-op so
// unit tests can import this package without spinning up state.
func StartPoller(db *gorm.DB) {
	if db == nil {
		log.Printf("talkex poller: nil db, skipping")
		return
	}
	// Auto-migrate our own checkpoint table.
	if err := db.AutoMigrate(&PollerState{}); err != nil {
		log.Printf("talkex poller: migrate state table failed: %v", err)
		return
	}
	go pollLoop(db)
	log.Printf("talkex poller: started (interval %s)", pollInterval)
}

func pollLoop(db *gorm.DB) {
	// Small initial delay so a boot-time flood of migrations + first-run
	// admin tasks doesn't race with our first poll.
	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		pollOnce(db)
		<-ticker.C
	}
}

// pollOnce enumerates every enabled TalkEx channel and pulls its
// inbox. Errors from one tenant never affect another — each iteration
// captures its own recover().
func pollOnce(db *gorm.DB) {
	listerMu.RLock()
	l := lister
	listerMu.RUnlock()
	if l == nil {
		return
	}

	for _, b := range l() {
		func(binding ChannelBinding) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("talkex poller: panic recovered for owner %s: %v",
						binding.OwnerID, r)
				}
			}()
			if err := pollTenant(db, binding); err != nil {
				log.Printf("talkex poller: owner %s: %v", binding.OwnerID, err)
				_ = db.Exec(
					`UPDATE talkex_poller_state SET last_error = ?, last_run_at = ?
					 WHERE owner_id = ?`,
					err.Error(), time.Now(), binding.OwnerID,
				).Error
			}
		}(b)
	}
}

// pollTenant does one round trip for one tenant.
func pollTenant(db *gorm.DB, b ChannelBinding) error {
	if b.APIKey == "" {
		return nil // channel enabled but unconfigured; skip quietly
	}
	baseURL := b.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	// Load or create checkpoint.
	var state PollerState
	err := db.Where("owner_id = ?", b.OwnerID).First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		state = PollerState{OwnerID: b.OwnerID, LastSince: 0}
		_ = db.Create(&state).Error
	} else if err != nil {
		return fmt.Errorf("state load: %w", err)
	}

	messages, nextSince, err := fetchInbound(baseURL, b.APIKey, state.LastSince, 100)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	upserterMu.RLock()
	up := upserter
	upserterMu.RUnlock()
	recorderMu.RLock()
	rec := recorder
	recorderMu.RUnlock()
	if up == nil || rec == nil {
		return errors.New("upserter/recorder not wired")
	}

	for _, m := range messages {
		contactID, err := up(b.OwnerID, m.FromUsername)
		if err != nil {
			log.Printf("talkex poller: upsert contact %s failed: %v", m.FromUsername, err)
			continue
		}
		body := m.Text
		if body == "" && m.Kind != "text" {
			// Non-text kinds (photo/voice/etc.) — surface a placeholder
			// so the agent knows something came in.
			body = fmt.Sprintf("[%s]", m.Kind)
		}
		if err := rec(b.OwnerID, contactID, "talkex", body); err != nil {
			log.Printf("talkex poller: record inbound failed: %v", err)
			continue
		}
	}

	// Advance the checkpoint even when messages was empty — the endpoint
	// echoes back the caller's since when nothing new arrived, so we
	// only move forward on real progress.
	if nextSince > state.LastSince {
		state.LastSince = nextSince
	}
	state.LastRunAt = time.Now()
	state.LastError = ""
	return db.Save(&state).Error
}

// inboundMessage mirrors the JSON row TalkEx returns.
type inboundMessage struct {
	MessageID    string  `json:"message_id"`
	ChatID       string  `json:"chat_id"`
	FromUsername string  `json:"from_username"`
	Text         string  `json:"text"`
	Kind         string  `json:"kind"`
	CreatedAt    float64 `json:"created_at"`
	Seq          int64   `json:"seq"`
}

// fetchInbound calls GET /api/v1/inbound. Returns (rows, next_since, err).
// A network error or non-2xx is surfaced so the loop can log + backoff.
func fetchInbound(baseURL, apiKey string, since float64, limit int) ([]inboundMessage, float64, error) {
	url := fmt.Sprintf("%s/api/v1/inbound?since=%.6f&limit=%d", baseURL, since, limit)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, since, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TalkExBusinessPoller/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, since, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		// Endpoint not deployed on this TalkEx server yet — degrade
		// gracefully so the loop doesn't spam errors.
		return nil, since, errors.New("inbound endpoint not available (deploy the /api/v1/inbound patch to TalkEx)")
	}
	if resp.StatusCode >= 400 {
		return nil, since, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var out struct {
		Messages  []inboundMessage `json:"messages"`
		NextSince float64          `json:"next_since"`
		Count     int              `json:"count"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, since, fmt.Errorf("decode: %w", err)
	}
	return out.Messages, out.NextSince, nil
}

// LookupOwnerState — a small helper the /channels page can call to
// show the poller's last-run timestamp + last error in the merchant's
// dashboard.
func LookupOwnerState(ownerID string) (PollerState, error) {
	var s PollerState
	err := database.DB.Where("owner_id = ?", ownerID).First(&s).Error
	return s, err
}
