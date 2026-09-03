// Package talkex — connector for CoreAxis's own TalkEx messenger.
//
// TalkEx is the CoreAxis-owned WhatsApp-style consumer app that lives
// at web.talkex.in, backed by a FastAPI server at
// talkex-backend.onrender.com. Businesses on TalkEx Business talk to
// their customers on TalkEx via the platform's "Bulk messaging API" —
// the WhatsApp-Business-API equivalent scoped to DMs only.
//
// The bulk API surface (as documented in the TalkEx Messenger repo):
//
//   POST /api/v1/messages
//     Header:  Authorization: Bearer <talkex_api_key>
//     Body:    {"to":"<username>","text":"...","client_msg_id":"..."}
//     Cap:     60 sends/minute per account
//     Limits:  DMs only, text only, blocked users refuse both ways
//
// A reply from the customer lands in the merchant's normal TalkEx
// inbox — not a separate business channel. To surface those replies
// inside TalkEx Business, wire an inbound poller (see fetchInbound
// below) or a WebSocket subscriber against the shared /auth/ws-ticket
// endpoint (deferred; poller is enough for a soft launch).
package talkex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
)

// defaultBaseURL is where the TalkEx FastAPI server lives. Overridable
// per-tenant via the connector's config so a self-hosted TalkEx (rare)
// or a staging URL can be pointed at instead.
const defaultBaseURL = "https://talkex-backend.onrender.com"

var httpClient = &http.Client{Timeout: 15 * time.Second}

// Connector holds the per-tenant TalkEx credentials.
type Connector struct {
	BaseURL string
	APIKey  string
}

// New builds a connector. Empty baseURL falls back to the shared TalkEx
// production URL — the common case; overriding is only for staging.
func New(baseURL, apiKey string) *Connector {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Connector{BaseURL: baseURL, APIKey: apiKey}
}

func (c *Connector) Name() string { return "talkex" }

func (c *Connector) ValidateConfig(config map[string]string) error {
	// base_url is optional — defaults to defaultBaseURL. api_key is not.
	if config["api_key"] == "" {
		return errors.New("talkex: api_key is required — generate one at " +
			"https://web.talkex.in in Settings → API keys, or via " +
			"POST /me/api-keys after logging in")
	}
	return nil
}

// Send fires one message to a single TalkEx username. TalkEx's bulk API
// addresses recipients by USERNAME (not phone, not internal id) — the
// caller must therefore pass the recipient's TalkEx handle in
// msg.ContactID. The contacts layer stores this in the CustomFields
// map under the "talkex_username" key.
func (c *Connector) Send(msg *shared.OutboundMessage) (*shared.DeliveryResult, error) {
	if msg.ContactID == "" {
		return nil, errors.New("talkex: recipient username is required in ContactID")
	}
	if msg.Body == "" {
		return nil, errors.New("talkex: text body is required (bulk API is text-only)")
	}
	if msg.Type != shared.TypeText && msg.Type != "" {
		// The bulk API is DM-text-only per spec — refuse rich media
		// loudly instead of silently dropping the attachment.
		return nil, fmt.Errorf("talkex: bulk API only supports text messages, got %q", msg.Type)
	}

	// client_msg_id gives TalkEx an idempotency key so a retry after a
	// network blip doesn't double-send. Derive it from the messaging
	// engine's own message ID when available; otherwise mint one from
	// the current nanoseconds so a fresh call is still unique.
	clientMsgID := msg.ID
	if clientMsgID == "" {
		clientMsgID = fmt.Sprintf("talkex-business-%d", time.Now().UnixNano())
	}

	payload := map[string]interface{}{
		"to":            msg.ContactID,
		"text":          msg.Body,
		"client_msg_id": clientMsgID,
	}
	body, _ := json.Marshal(payload)

	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		baseURL+"/api/v1/messages",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("talkex: request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", "TalkExBusiness/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("talkex: send failed: %w", err)
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)

	// TalkEx returns 429 with a Retry-After header when the tenant hits
	// the 60/min cap — surface it as a retryable delivery result so the
	// messaging engine backs off instead of marking failed.
	if resp.StatusCode == http.StatusTooManyRequests {
		return &shared.DeliveryResult{
			Status:    shared.StatusFailed,
			Timestamp: time.Now(),
			Error:     "talkex: rate limited (60/min cap); retry after cooldown",
		}, nil
	}
	if resp.StatusCode >= 400 {
		// Surface the server's error body for the DLQ so ops can debug
		// without opening a shell.
		return nil, fmt.Errorf("talkex: HTTP %d: %s", resp.StatusCode, truncate(string(rawBody), 200))
	}

	var result struct {
		MessageID string `json:"message_id"`
		ChatID    string `json:"chat_id"`
	}
	_ = json.Unmarshal(rawBody, &result)

	// TalkEx returns the internal chat_id + message_id — persist the
	// message_id as the external ref so read-receipt callbacks (once
	// wired) can correlate.
	externalID := result.MessageID
	if externalID == "" {
		externalID = clientMsgID
	}

	return &shared.DeliveryResult{
		ExternalID: externalID,
		Status:     shared.StatusSent,
		Timestamp:  time.Now(),
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func init() {
	// Auto-register with empty config — the real (BaseURL, APIKey) pair
	// is loaded per-tenant at send time from channels.Config so a single
	// binary can serve any number of TalkEx business accounts.
	shared.Register(&Connector{})
}
