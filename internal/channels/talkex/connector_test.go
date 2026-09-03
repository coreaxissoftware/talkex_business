package talkex

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
)

func TestSendMatchesTalkExBulkAPIShape(t *testing.T) {
	// Spin up a stand-in for talkex-backend.onrender.com so we can assert
	// the exact request the connector emits without hitting production.
	var got struct {
		Path        string
		Auth        string
		ContentType string
		Body        map[string]string
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Path = r.URL.Path
		got.Auth = r.Header.Get("Authorization")
		got.ContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		got.Body = map[string]string{}
		_ = json.Unmarshal(body, &got.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message_id":"m_42","chat_id":"c_9"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "sk_test_abc")
	msg := &shared.OutboundMessage{
		ID:        "engine-msg-777",
		ContactID: "alice", // TalkEx addresses by username, not id
		Body:      "Your order #1234 is out for delivery",
		Type:      shared.TypeText,
	}

	res, err := c.Send(msg)
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if res.Status != shared.StatusSent {
		t.Fatalf("status=%s want sent", res.Status)
	}
	if res.ExternalID != "m_42" {
		t.Fatalf("external_id=%q want m_42", res.ExternalID)
	}

	if got.Path != "/api/v1/messages" {
		t.Errorf("path=%q want /api/v1/messages", got.Path)
	}
	if got.Auth != "Bearer sk_test_abc" {
		t.Errorf("auth=%q", got.Auth)
	}
	if !strings.HasPrefix(got.ContentType, "application/json") {
		t.Errorf("content-type=%q", got.ContentType)
	}
	if got.Body["to"] != "alice" {
		t.Errorf("body.to=%q want alice", got.Body["to"])
	}
	if got.Body["text"] != "Your order #1234 is out for delivery" {
		t.Errorf("body.text=%q", got.Body["text"])
	}
	if got.Body["client_msg_id"] != "engine-msg-777" {
		t.Errorf("body.client_msg_id=%q want engine-msg-777 (idempotency)", got.Body["client_msg_id"])
	}
}

func TestSendRejectsNonTextTypes(t *testing.T) {
	// The bulk API is text-DM-only; the connector must refuse to try
	// image/video/audio sends rather than silently drop the media.
	c := New("https://ignored", "key")
	_, err := c.Send(&shared.OutboundMessage{
		ContactID: "alice",
		Body:      "caption",
		Type:      shared.TypeImage,
	})
	if err == nil || !strings.Contains(err.Error(), "text-only") && !strings.Contains(err.Error(), "text messages") {
		t.Fatalf("expected rejection of image type, got %v", err)
	}
}

func TestSendMissingRecipient(t *testing.T) {
	c := New("https://ignored", "key")
	_, err := c.Send(&shared.OutboundMessage{Body: "hi", Type: shared.TypeText})
	if err == nil {
		t.Fatal("expected error when ContactID missing")
	}
}

func TestSendMissingBody(t *testing.T) {
	c := New("https://ignored", "key")
	_, err := c.Send(&shared.OutboundMessage{ContactID: "alice", Type: shared.TypeText})
	if err == nil {
		t.Fatal("expected error when Body missing")
	}
}

func TestSendHandlesRateLimit(t *testing.T) {
	// 429 from TalkEx should surface as a failed-with-reason result,
	// not a hard connector error — lets the messaging engine retry.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "12")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := New(srv.URL, "key")
	res, err := c.Send(&shared.OutboundMessage{
		ContactID: "alice", Body: "hi", Type: shared.TypeText,
	})
	if err != nil {
		t.Fatalf("unexpected connector error: %v", err)
	}
	if res.Status != shared.StatusFailed {
		t.Fatalf("status=%s want failed", res.Status)
	}
	if !strings.Contains(res.Error, "rate limited") {
		t.Fatalf("error=%q missing rate-limit hint", res.Error)
	}
}

func TestValidateConfigRequiresAPIKey(t *testing.T) {
	c := &Connector{}
	if err := c.ValidateConfig(map[string]string{}); err == nil {
		t.Fatal("expected error when api_key missing")
	}
	if err := c.ValidateConfig(map[string]string{"api_key": "x"}); err != nil {
		t.Fatalf("unexpected error with api_key set: %v", err)
	}
}

func TestNewDefaultsToProductionURL(t *testing.T) {
	c := New("", "k")
	if c.BaseURL != defaultBaseURL {
		t.Fatalf("empty base_url should default to %q, got %q", defaultBaseURL, c.BaseURL)
	}
}
