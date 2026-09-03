package talkex

// keygen.go — server-side proxy for minting a TalkEx bulk API key from
// the merchant's TalkEx credentials, without their password ever
// touching our database.
//
// Flow:
//   1. Merchant clicks "Generate TalkEx key" on the Channels page.
//   2. Frontend modal asks for TalkEx username + password (+ optional
//      label). Fields never leave the modal after submit — no local
//      storage, no fetch cache.
//   3. Our backend calls TalkEx POST /auth/login with those creds:
//        - Plain success → we get a session token
//        - 2FA on the account → returns {requires_pin, pending_token};
//          we surface that to the UI so the merchant can enter the PIN
//   4. With the session token we call POST /me/api-keys{label?} →
//      raw bulk key (shown ONCE per TalkEx design)
//   5. We save the key into channels.Config for kind=talkex and return
//      it to the frontend for one-time display + clipboard copy
//
// Deliberately fresh HTTP requests per call — we never persist the
// session token, only the derived bulk key (which is scoped to
// send-only + poll-only, so leaking it is far less catastrophic than
// leaking a session token would be).

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GenerateRequest is the merchant-supplied input.
type GenerateRequest struct {
	Username string `json:"talkex_username" binding:"required"`
	Password string `json:"talkex_password" binding:"required"`
	Label    string `json:"label"`

	// PIN + PendingToken are only set on the second call when the
	// first came back with requires_pin=true.
	PIN          string `json:"pin,omitempty"`
	PendingToken string `json:"pending_token,omitempty"`

	// BaseURL — override for self-hosted / staging TalkEx. Empty →
	// defaultBaseURL.
	BaseURL string `json:"base_url,omitempty"`
}

// GenerateResponse is what the frontend gets back.
type GenerateResponse struct {
	// Populated on happy path (or on the second call after PIN)
	Key    string `json:"key,omitempty"`
	KeyID  string `json:"key_id,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Label  string `json:"label,omitempty"`

	// Populated when the account has 2FA on. Frontend re-prompts for
	// the PIN and re-submits the same call with pin + pending_token.
	RequiresPIN  bool   `json:"requires_pin,omitempty"`
	PendingToken string `json:"pending_token,omitempty"`
}

// ErrTalkExAuth wraps a 401 from the TalkEx login so the handler can
// map it to a 401 in our own API rather than a 500.
var ErrTalkExAuth = errors.New("invalid TalkEx credentials")

// Generate does the two-step (or three-step, with 2FA) flow.
func Generate(req *GenerateRequest) (*GenerateResponse, error) {
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	label := req.Label
	if label == "" {
		label = "TalkEx Business bridge"
	}

	client := &http.Client{Timeout: 12 * time.Second}

	var sessionToken string

	// Branch 1 — merchant already went through the 2FA prompt on the
	// previous call. Verify the PIN and get a session token.
	if req.PendingToken != "" && req.PIN != "" {
		tok, err := verifyPIN(client, baseURL, req.PendingToken, req.PIN)
		if err != nil {
			return nil, fmt.Errorf("verify pin: %w", err)
		}
		sessionToken = tok
	} else {
		// Branch 2 — fresh login with username + password.
		loginOut, err := doLogin(client, baseURL, req.Username, req.Password)
		if err != nil {
			return nil, err
		}
		if loginOut.RequiresPIN {
			// Bounce the flow back to the UI with the pending token so
			// the merchant can type the PIN and re-submit.
			return &GenerateResponse{
				RequiresPIN:  true,
				PendingToken: loginOut.PendingToken,
			}, nil
		}
		sessionToken = loginOut.Token
	}

	// Mint the bulk API key with the session token.
	key, err := mintKey(client, baseURL, sessionToken, label)
	if err != nil {
		return nil, fmt.Errorf("mint key: %w", err)
	}
	return key, nil
}

// ---- low-level HTTP helpers -----------------------------------------

type loginResult struct {
	Token        string `json:"token,omitempty"`
	RequiresPIN  bool   `json:"requires_pin,omitempty"`
	PendingToken string `json:"pending_token,omitempty"`
}

func doLogin(client *http.Client, baseURL, username, password string) (*loginResult, error) {
	body, _ := json.Marshal(map[string]string{
		"username":     username,
		"password":     password,
		"device_label": "TalkEx Business (key mint)",
	})
	req, err := makeReq(http.MethodPost, baseURL+"/auth/login", body)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode == http.StatusUnauthorized {
		return nil, ErrTalkExAuth
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("login: HTTP %d: %s", res.StatusCode, truncate(string(raw), 200))
	}
	var out loginResult
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("login: decode: %w", err)
	}
	if !out.RequiresPIN && out.Token == "" {
		return nil, errors.New("login: empty token")
	}
	return &out, nil
}

func verifyPIN(client *http.Client, baseURL, pending, pin string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"pending_token": pending,
		"pin":           pin,
		"device_label":  "TalkEx Business (key mint)",
	})
	req, err := makeReq(http.MethodPost, baseURL+"/auth/login/verify-pin", body)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("verify-pin: HTTP %d: %s", res.StatusCode, truncate(string(raw), 200))
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("verify-pin: decode: %w", err)
	}
	if out.Token == "" {
		return "", errors.New("verify-pin: empty token")
	}
	return out.Token, nil
}

func mintKey(client *http.Client, baseURL, sessionToken, label string) (*GenerateResponse, error) {
	body, _ := json.Marshal(map[string]string{"label": label})
	req, err := makeReq(http.MethodPost, baseURL+"/me/api-keys", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+sessionToken)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", res.StatusCode, truncate(string(raw), 200))
	}
	var out struct {
		ID     string `json:"id"`
		Key    string `json:"key"`
		Label  string `json:"label"`
		Prefix string `json:"prefix"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if out.Key == "" {
		return nil, errors.New("empty key in response")
	}
	return &GenerateResponse{
		Key: out.Key, KeyID: out.ID, Label: out.Label, Prefix: out.Prefix,
	}, nil
}

// makeReq builds a fresh JSON request. Timeout is enforced by the
// shared http.Client (12s), so we don't need a per-request context
// deadline here — one less cancel() to plumb through the call site.
func makeReq(method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "TalkExBusiness/1.0 (key-mint)")
	return req, nil
}
