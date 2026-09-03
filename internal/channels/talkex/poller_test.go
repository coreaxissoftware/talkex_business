package talkex

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// TestFetchInboundParsesResponse hits a stand-in TalkEx server and
// asserts the poller reads the wire format correctly.
func TestFetchInboundParsesResponse(t *testing.T) {
	var gotAuth, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"messages": []map[string]any{
				{
					"message_id":    "m_1",
					"chat_id":       "dm_a_b",
					"from_username": "alice",
					"text":          "hey",
					"kind":          "text",
					"created_at":    1700000000.5,
					"seq":           42,
				},
				{
					"message_id":    "m_2",
					"chat_id":       "dm_a_b",
					"from_username": "alice",
					"text":          "you there?",
					"kind":          "text",
					"created_at":    1700000005.0,
					"seq":           43,
				},
			},
			"next_since": 1700000005.0,
			"count":      2,
		})
	}))
	defer srv.Close()

	msgs, next, err := fetchInbound(srv.URL, "sk_bulk_123", 1699999000.0, 100)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if gotAuth != "Bearer sk_bulk_123" {
		t.Errorf("auth=%q want Bearer sk_bulk_123", gotAuth)
	}
	if !strings.Contains(gotPath, "since=1699999000") || !strings.Contains(gotPath, "limit=100") {
		t.Errorf("path=%q missing since/limit", gotPath)
	}
	if next != 1700000005.0 {
		t.Errorf("next_since=%v want 1700000005", next)
	}
	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2", len(msgs))
	}
	if msgs[0].FromUsername != "alice" || msgs[0].Text != "hey" {
		t.Errorf("row 0 wrong: %+v", msgs[0])
	}
	if msgs[1].Seq != 43 {
		t.Errorf("row 1 seq=%d", msgs[1].Seq)
	}
}

// TestFetchInboundHandlesMissingEndpoint validates the graceful 404
// path — the poller must degrade to a clear error, not spam retries.
func TestFetchInboundHandlesMissingEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, since, err := fetchInbound(srv.URL, "k", 0, 100)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("error=%q missing 'not available' hint", err.Error())
	}
	if since != 0 {
		t.Errorf("since should not advance on 404, got %v", since)
	}
}

// TestFetchInboundHandlesEmpty covers the common quiet-poll case: no
// new messages, endpoint echoes the caller's since back.
func TestFetchInboundHandlesEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"messages":[],"next_since":1699999999,"count":0}`))
	}))
	defer srv.Close()

	msgs, next, err := fetchInbound(srv.URL, "k", 1699999999, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 0 {
		t.Errorf("got %d messages, want 0", len(msgs))
	}
	if next != 1699999999 {
		t.Errorf("next=%v", next)
	}
}

// TestPollerCallbackContract sanity-checks the Register* wiring so a
// future refactor that renames a hook fails fast in unit tests.
func TestPollerCallbackContract(t *testing.T) {
	var upserted, recorded sync.WaitGroup
	upserted.Add(1)
	recorded.Add(1)

	RegisterContactUpserter(func(ownerID, username string) (string, error) {
		defer upserted.Done()
		if ownerID != "owner-1" || username != "alice" {
			t.Errorf("upsert args wrong: %s %s", ownerID, username)
		}
		return "contact-42", nil
	})
	RegisterInboundRecorder(func(ownerID, contactID, channel, body string) error {
		defer recorded.Done()
		if channel != "talkex" || contactID != "contact-42" {
			t.Errorf("record args wrong: %s %s %s %s", ownerID, contactID, channel, body)
		}
		return nil
	})
	RegisterChannelLister(func() []ChannelBinding { return nil })
	// Just call the registered hooks directly — the fetch path is
	// covered by the httptest tests above.
	upserterMu.RLock()
	up := upserter
	upserterMu.RUnlock()
	recorderMu.RLock()
	rec := recorder
	recorderMu.RUnlock()

	id, err := up("owner-1", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if err := rec("owner-1", id, "talkex", "hey"); err != nil {
		t.Fatal(err)
	}
	upserted.Wait()
	recorded.Wait()
}
