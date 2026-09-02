package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

// Apple Sign-in — full production implementation.
//
// Flow (per https://developer.apple.com/documentation/sign_in_with_apple):
//
//   1. Mint a JWT client_secret signed with our ECDSA P-256 .p8 key,
//      claims { iss:TeamID, sub:ClientID, aud:appleid.apple.com,
//               iat:now, exp:now+6mo, kid:KeyID }.
//   2. POST code + client_secret to https://appleid.apple.com/auth/token.
//   3. Parse the returned id_token (RS256 JWT). Verify signature against
//      Apple's rotating JWKS at https://appleid.apple.com/auth/keys.
//   4. Extract email + email_verified from id_token claims. Full name
//      only comes on the very first consent, in the initial POST from
//      Apple to our callback — captured separately by the handler.

const (
	appleTokenURL  = "https://appleid.apple.com/auth/token"
	appleJWKSURL   = "https://appleid.apple.com/auth/keys"
	appleIssuer    = "https://appleid.apple.com"
	appleAudience  = "https://appleid.apple.com"
	appleSecretTTL = 15 * time.Minute // Apple caps at 6 months; short is safer.
)

// clientSecretCache serves the JWT client secret. Apple accepts the
// same JWT until it expires, so caching for ~15 min avoids re-signing
// on every callback. Safe across pods — each pod holds its own copy.
type appleSecretCache struct {
	mu       sync.Mutex
	value    string
	expires  time.Time
}

var appleCache appleSecretCache

// jwksCache holds Apple's rotating public keys so we don't refetch on
// every callback. Apple rotates every few weeks; 1-hour TTL is safe.
type appleJWKS struct {
	mu       sync.Mutex
	fetched  time.Time
	keys     map[string]*rsa.PublicKey
}

var jwksStore appleJWKS

// mintAppleClientSecret returns a fresh signed JWT usable as
// client_secret in the token exchange. Cached for appleSecretTTL.
func mintAppleClientSecret(cfg *config.Config) (string, error) {
	appleCache.mu.Lock()
	defer appleCache.mu.Unlock()
	if appleCache.value != "" && time.Now().Before(appleCache.expires) {
		return appleCache.value, nil
	}

	if cfg.OAuthAppleTeamID == "" || cfg.OAuthAppleKeyID == "" ||
		cfg.OAuthAppleClientID == "" || cfg.OAuthApplePrivateKey == "" {
		return "", errors.New("apple: missing OAUTH_APPLE_TEAM_ID / KEY_ID / CLIENT_ID / PRIVATE_KEY")
	}

	// Decode the .p8 PEM.
	// Some deployment platforms strip the newlines when the key is
	// pasted into a single-line env var. Restore them if we see the
	// classic BEGIN/END markers on one line.
	pemStr := cfg.OAuthApplePrivateKey
	if strings.Contains(pemStr, "BEGIN PRIVATE KEY") && !strings.Contains(pemStr, "\n") {
		pemStr = strings.Replace(pemStr, "-----BEGIN PRIVATE KEY-----", "-----BEGIN PRIVATE KEY-----\n", 1)
		pemStr = strings.Replace(pemStr, "-----END PRIVATE KEY-----", "\n-----END PRIVATE KEY-----", 1)
	}
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return "", errors.New("apple: private key not PEM-encoded")
	}
	key, err := jwt.ParseECPrivateKeyFromPEM(pem.EncodeToMemory(block))
	if err != nil {
		return "", fmt.Errorf("apple: parse EC key: %w", err)
	}
	if _, ok := interface{}(key).(*ecdsa.PrivateKey); !ok {
		return "", errors.New("apple: key is not an ECDSA private key")
	}

	now := time.Now()
	exp := now.Add(appleSecretTTL)
	claims := jwt.MapClaims{
		"iss": cfg.OAuthAppleTeamID,
		"iat": now.Unix(),
		"exp": exp.Unix(),
		"aud": appleAudience,
		"sub": cfg.OAuthAppleClientID,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["kid"] = cfg.OAuthAppleKeyID

	signed, err := tok.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("apple: sign client_secret: %w", err)
	}
	appleCache.value = signed
	appleCache.expires = exp.Add(-1 * time.Minute) // renew a minute before expiry
	return signed, nil
}

// exchangeApple performs the real Apple token exchange and id_token
// verification. Replaces the stub in oauth_exchange.go.
func exchangeAppleReal(_ OAuthProvider, clientID, _, code, redirectURI string, cfg *config.Config) (string, string, error) {
	clientSecret, err := mintAppleClientSecret(cfg)
	if err != nil {
		return "", "", err
	}

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirectURI)

	var tok struct {
		IDToken string `json:"id_token"`
		Error   string `json:"error"`
	}
	if err := postForm(appleTokenURL, form, nil, &tok); err != nil {
		return "", "", fmt.Errorf("apple token: %w", err)
	}
	if tok.IDToken == "" {
		msg := tok.Error
		if msg == "" {
			msg = "empty id_token"
		}
		return "", "", fmt.Errorf("apple: %s", msg)
	}

	claims, err := verifyAppleIDToken(tok.IDToken, clientID)
	if err != nil {
		return "", "", fmt.Errorf("apple id_token: %w", err)
	}
	email, _ := claims["email"].(string)
	if email == "" {
		return "", "", errors.New("apple: id_token missing email (scope 'email' must be requested)")
	}
	if verified, ok := claims["email_verified"].(string); ok && verified == "false" {
		return "", "", errors.New("apple: email not verified")
	}
	// Apple never sends the full name in id_token. Handler pulls it
	// from the initial callback form (see handleOAuthCallback).
	return email, "", nil
}

// verifyAppleIDToken parses the JWT and verifies its RS256 signature
// against Apple's JWKS, checks iss/aud/exp claims. Returns the claim
// map on success.
func verifyAppleIDToken(idToken, clientID string) (jwt.MapClaims, error) {
	tok, err := jwt.Parse(idToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected alg %q", t.Header["alg"])
		}
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("id_token missing kid")
		}
		return appleJWKSKey(kid)
	}, jwt.WithIssuer(appleIssuer), jwt.WithAudience(clientID))
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

// appleJWKSKey returns the RSA public key for the given key ID.
// Fetches + caches the Apple JWKS if we haven't in the last hour.
func appleJWKSKey(kid string) (*rsa.PublicKey, error) {
	jwksStore.mu.Lock()
	defer jwksStore.mu.Unlock()

	if time.Since(jwksStore.fetched) > time.Hour || jwksStore.keys == nil {
		if err := refreshAppleJWKS(); err != nil {
			return nil, err
		}
	}
	k, ok := jwksStore.keys[kid]
	if !ok {
		// Cache miss — force refresh once, then look again.
		if err := refreshAppleJWKS(); err != nil {
			return nil, err
		}
		k, ok = jwksStore.keys[kid]
	}
	if !ok {
		return nil, fmt.Errorf("apple JWKS: unknown kid %q", kid)
	}
	return k, nil
}

// refreshAppleJWKS pulls Apple's key set and stores each RSA key by kid.
// Must be called with jwksStore.mu held.
func refreshAppleJWKS() error {
	req, err := http.NewRequest(http.MethodGet, appleJWKSURL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: httpTimeout}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("apple JWKS: HTTP %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Use string `json:"use"`
			Alg string `json:"alg"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("apple JWKS: decode: %w", err)
	}
	out := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		n := new(big.Int).SetBytes(nBytes)
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}
		out[k.Kid] = &rsa.PublicKey{N: n, E: e}
	}
	if len(out) == 0 {
		return errors.New("apple JWKS: no RSA keys parsed")
	}
	jwksStore.keys = out
	jwksStore.fetched = time.Now()
	return nil
}
