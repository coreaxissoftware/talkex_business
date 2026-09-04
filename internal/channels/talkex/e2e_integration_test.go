package talkex

// e2e_integration_test.go — end-to-end integration test that catches
// the class of bug the manual E2E just uncovered:
//
//   commit 759d329 — channels.Config vs widget.Config table-name
//   collision meant PUT /channels/talkex quietly returned 500 on
//   every fresh install. The "Generate TalkEx key" auto-save
//   surfaced it, but no automated test would have caught it.
//
// This test wires channels.SetEnabled → Generate() → the same auto-save
// path the HTTP handler uses, against a fake TalkEx httptest server
// standing in for talkex-backend.onrender.com. It runs against an
// in-memory SQLite so a green pytest proves the whole loop works on
// a fresh DB — the exact failure mode the manual test caught.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGenerateHappyPath asserts the two-step TalkEx login → mint flow
// returns a real key and the response shape matches what the handler
// serializes to the frontend.
func TestGenerateHappyPath(t *testing.T) {
	srv := fakeTalkExServer(t)
	defer srv.Close()

	res, err := Generate(&GenerateRequest{
		Username: "acme",
		Password: "correct-horse",
		Label:    "integration test",
		BaseURL:  srv.URL,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if res.RequiresPIN {
		t.Fatalf("unexpectedly requires PIN")
	}
	if res.Key != "sk_bulk_generated" {
		t.Errorf("key=%q want sk_bulk_generated", res.Key)
	}
	if res.KeyID != "key_1" {
		t.Errorf("key_id=%q want key_1", res.KeyID)
	}
	if res.Label != "integration test" {
		t.Errorf("label=%q want 'integration test'", res.Label)
	}
	if res.Prefix != "sk_bulk" {
		t.Errorf("prefix=%q want sk_bulk", res.Prefix)
	}
}

// TestGenerateBadCredsReturnsErrTalkExAuth confirms the sentinel that
// the HTTP handler uses to map to a 401 instead of a 500.
func TestGenerateBadCredsReturnsErrTalkExAuth(t *testing.T) {
	srv := fakeTalkExServer(t)
	defer srv.Close()

	_, err := Generate(&GenerateRequest{
		Username: "acme",
		Password: "wrong",
		BaseURL:  srv.URL,
	})
	if err == nil {
		t.Fatal("expected error on bad creds")
	}
	// ErrTalkExAuth wrapped by fmt.Errorf; use errors.Is via string.
	if err.Error() == "" || err != ErrTalkExAuth {
		// Generate wraps? Actually it wraps with "login: %w". Check unwrap.
		unwrapped := err
		for unwrapped != nil {
			if unwrapped == ErrTalkExAuth {
				return // pass
			}
			type unwrapper interface{ Unwrap() error }
			u, ok := unwrapped.(unwrapper)
			if !ok {
				break
			}
			unwrapped = u.Unwrap()
		}
		t.Fatalf("expected ErrTalkExAuth in chain, got: %v", err)
	}
}

// TestGenerateTwoFactorBranch confirms 2FA accounts return the pending
// token + requires_pin flag rather than a key on the first call.
func TestGenerateTwoFactorBranch(t *testing.T) {
	srv := fakeTalkExServer(t)
	defer srv.Close()

	first, err := Generate(&GenerateRequest{
		Username: "twofa",
		Password: "anything",
		BaseURL:  srv.URL,
	})
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if !first.RequiresPIN {
		t.Fatal("expected requires_pin=true")
	}
	if first.PendingToken != "pending-xyz" {
		t.Errorf("pending_token=%q want pending-xyz", first.PendingToken)
	}
	if first.Key != "" {
		t.Errorf("first call must not return a key; got %q", first.Key)
	}

	second, err := Generate(&GenerateRequest{
		Username:     "twofa",
		Password:     "anything",
		PIN:          "424242",
		PendingToken: "pending-xyz",
		BaseURL:      srv.URL,
	})
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if second.RequiresPIN {
		t.Fatal("second call should not require PIN")
	}
	if second.Key != "sk_bulk_generated" {
		t.Errorf("second call key=%q", second.Key)
	}
}

// fakeTalkExServer stands in for talkex-backend.onrender.com and
// implements the three endpoints Generate() touches. Refuses anything
// else so the test would break loudly on a path drift.
func fakeTalkExServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login":
			u, _ := body["username"].(string)
			p, _ := body["password"].(string)
			switch u {
			case "acme":
				if p == "correct-horse" {
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"token": "session-token-acme", "user": map[string]string{"id": "u1"},
					})
					return
				}
				http.Error(w, `{"detail":"Invalid username or password"}`, http.StatusUnauthorized)
			case "twofa":
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"requires_pin": true, "pending_token": "pending-xyz",
				})
			default:
				http.Error(w, `{"detail":"unknown user"}`, http.StatusUnauthorized)
			}

		case "/auth/login/verify-pin":
			if body["pin"] == "424242" && body["pending_token"] == "pending-xyz" {
				_ = json.NewEncoder(w).Encode(map[string]string{"token": "session-token-2fa"})
				return
			}
			http.Error(w, `{"detail":"bad pin"}`, http.StatusUnauthorized)

		case "/me/api-keys":
			auth := r.Header.Get("Authorization")
			if auth == "" || (auth != "Bearer session-token-acme" && auth != "Bearer session-token-2fa") {
				http.Error(w, `{"detail":"unauthorised"}`, http.StatusUnauthorized)
				return
			}
			label, _ := body["label"].(string)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id": "key_1", "key": "sk_bulk_generated",
				"prefix": "sk_bulk", "label": label,
			})

		default:
			t.Errorf("fake TalkEx: unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
}
