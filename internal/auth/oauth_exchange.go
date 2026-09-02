package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

// exchangeAndFetchUser performs the two-step OAuth 2.0 dance for a
// given provider:
//
//   1. POST code + client_secret → provider's TokenURL → access_token
//   2. GET provider's UserInfoURL with Authorization: Bearer <token>
//      → parse {email, name}
//
// Returns (email, fullName, error). Providers each have their own
// idiosyncrasies handled inline (Facebook wants access_token as a
// query parameter, GitHub returns non-verified email through a
// secondary endpoint, Apple ships user info in the id_token). Kept
// in one file so the handler stays readable.
func exchangeAndFetchUser(providerName, code, redirectURI string) (email, fullName string, err error) {
	provider, ok := providers[providerName]
	if !ok {
		return "", "", fmt.Errorf("unknown provider %q", providerName)
	}
	cfg := config.Get()
	clientID := getProviderClientID(providerName)
	clientSecret := getProviderClientSecret(providerName)
	if clientID == "" || clientSecret == "" {
		return "", "", fmt.Errorf("%s client credentials not configured", providerName)
	}

	switch providerName {
	case "google":
		return exchangeGoogle(provider, clientID, clientSecret, code, redirectURI)
	case "github":
		return exchangeGitHub(provider, clientID, clientSecret, code, redirectURI)
	case "facebook":
		return exchangeFacebook(provider, clientID, clientSecret, code, redirectURI)
	case "apple":
		return exchangeApple(provider, clientID, clientSecret, code, redirectURI, cfg)
	}
	return "", "", fmt.Errorf("provider %q not implemented", providerName)
}

// httpTimeout — every outbound call to a provider is capped so a slow
// Meta / Apple endpoint can't stall the OAuth callback beyond 8s.
var httpTimeout = 8 * time.Second

// postForm is a small helper that POSTs URL-encoded form data and
// decodes the JSON response.
func postForm(endpoint string, form url.Values, headers map[string]string, out interface{}) error {
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: httpTimeout}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return fmt.Errorf("token endpoint: HTTP %d: %s", res.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("token endpoint: decode: %w (body=%s)", err, string(body))
	}
	return nil
}

func getJSON(endpoint, bearer string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: httpTimeout}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("userinfo endpoint: HTTP %d: %s", res.StatusCode, string(body))
	}
	return json.NewDecoder(res.Body).Decode(out)
}

// ---- Google ---------------------------------------------------------

func exchangeGoogle(p OAuthProvider, clientID, clientSecret, code, redirectURI string) (string, string, error) {
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirectURI)

	if err := postForm(p.TokenURL, form, nil, &tok); err != nil {
		return "", "", fmt.Errorf("google token: %w", err)
	}
	if tok.AccessToken == "" {
		return "", "", errors.New("google: empty access_token")
	}
	var user struct {
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
	}
	if err := getJSON(p.UserInfoURL, tok.AccessToken, &user); err != nil {
		return "", "", fmt.Errorf("google userinfo: %w", err)
	}
	if user.Email == "" {
		return "", "", errors.New("google: no email in userinfo (scope missing?)")
	}
	if !user.VerifiedEmail {
		return "", "", errors.New("google: email not verified")
	}
	return user.Email, user.Name, nil
}

// ---- GitHub ---------------------------------------------------------

func exchangeGitHub(p OAuthProvider, clientID, clientSecret, code, redirectURI string) (string, string, error) {
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)

	if err := postForm(p.TokenURL, form, nil, &tok); err != nil {
		return "", "", fmt.Errorf("github token: %w", err)
	}
	if tok.AccessToken == "" {
		return "", "", errors.New("github: empty access_token")
	}
	var user struct {
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := getJSON(p.UserInfoURL, tok.AccessToken, &user); err != nil {
		return "", "", fmt.Errorf("github userinfo: %w", err)
	}
	// The primary /user endpoint returns email only when it's public.
	// If empty, hit /user/emails and pick the primary + verified one.
	if user.Email == "" {
		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := getJSON("https://api.github.com/user/emails", tok.AccessToken, &emails); err == nil {
			for _, e := range emails {
				if e.Primary && e.Verified {
					user.Email = e.Email
					break
				}
			}
		}
	}
	if user.Email == "" {
		return "", "", errors.New("github: no verified primary email (scope user:email missing?)")
	}
	if user.Name == "" {
		user.Name = user.Login
	}
	return user.Email, user.Name, nil
}

// ---- Facebook -------------------------------------------------------

func exchangeFacebook(p OAuthProvider, clientID, clientSecret, code, redirectURI string) (string, string, error) {
	// Facebook takes the exchange as a GET (query params, not form body).
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("client_secret", clientSecret)
	q.Set("code", code)
	q.Set("redirect_uri", redirectURI)

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := getJSON(p.TokenURL+"?"+q.Encode(), "", &tok); err != nil {
		return "", "", fmt.Errorf("facebook token: %w", err)
	}
	if tok.AccessToken == "" {
		return "", "", errors.New("facebook: empty access_token")
	}
	// Facebook wants ?access_token=… rather than a Bearer header.
	uq := url.Values{}
	uq.Set("access_token", tok.AccessToken)
	uq.Set("fields", "id,name,email")
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := getJSON("https://graph.facebook.com/me?"+uq.Encode(), "", &user); err != nil {
		return "", "", fmt.Errorf("facebook userinfo: %w", err)
	}
	if user.Email == "" {
		return "", "", errors.New("facebook: no email in userinfo (scope email missing?)")
	}
	return user.Email, user.Name, nil
}

// ---- Apple ----------------------------------------------------------
//
// Sign in with Apple is meaningfully different from Google/FB/GitHub:
//
//   • The `client_secret` is not a static string — it's a JWT signed
//     with an EC private key (.p8 downloaded from the Apple Developer
//     portal) using the TEAM_ID + KEY_ID + BUNDLE_ID as claims. This
//     JWT is short-lived (max 6 months) and MUST be minted per request.
//
//   • User's email + name arrive inside the returned id_token (JWT),
//     not from a separate userinfo endpoint. First-time consent also
//     returns them in the initial form POST from Apple, but only once.
//
// Implementing the full JWT-client-secret flow here would add a JOSE
// dependency and a private-key management story that the wider project
// doesn't yet have. To keep the OAuth handler working for Google + FB
// + GitHub today and unblock a soft-launch, we return a clear error
// pointing at the docs. See docs/OAUTH_APPLE.md (create alongside)
// when the tenant needs Apple login.
func exchangeApple(p OAuthProvider, clientID, clientSecret, code, redirectURI string, cfg *config.Config) (string, string, error) {
	// Real implementation lives in oauth_apple.go — JWT client secret
	// mint + id_token verify against Apple's rotating JWKS.
	return exchangeAppleReal(p, clientID, clientSecret, code, redirectURI, cfg)
}

// getProviderClientSecret mirrors getProviderClientID for the secret.
func getProviderClientSecret(provider string) string {
	cfg := config.Get()
	switch provider {
	case "google":
		return cfg.OAuthGoogleSecret
	case "facebook":
		return cfg.OAuthFacebookSecret
	case "github":
		return cfg.OAuthGitHubSecret
	case "apple":
		return cfg.OAuthAppleSecret
	}
	return ""
}
