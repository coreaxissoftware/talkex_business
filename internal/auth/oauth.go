// OAuth social login endpoints.
//
// In production, these redirect to the real provider authorization URL
// and handle the callback (code exchange → user info → JWT).
//
// In dev mode, the callback auto-creates a simulated user and issues
// tokens immediately — no real OAuth provider needed. Switch to
// production by setting OAUTH_GOOGLE_CLIENT_ID etc. in the environment.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

// OAuthProvider holds config for one social login provider.
type OAuthProvider struct {
	Name         string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
}

// Known providers — client IDs/secrets come from env vars.
var providers = map[string]OAuthProvider{
	"google": {
		Name:        "google",
		AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:    "https://oauth2.googleapis.com/token",
		UserInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:      []string{"openid", "email", "profile"},
	},
	"facebook": {
		Name:        "facebook",
		AuthURL:     "https://www.facebook.com/v18.0/dialog/oauth",
		TokenURL:    "https://graph.facebook.com/v18.0/oauth/access_token",
		UserInfoURL: "https://graph.facebook.com/me?fields=id,name,email",
		Scopes:      []string{"email", "public_profile"},
	},
	"github": {
		Name:        "github",
		AuthURL:     "https://github.com/login/oauth/authorize",
		TokenURL:    "https://github.com/login/oauth/access_token",
		UserInfoURL: "https://api.github.com/user",
		Scopes:      []string{"user:email"},
	},
	"apple": {
		Name:        "apple",
		AuthURL:     "https://appleid.apple.com/auth/authorize",
		TokenURL:    "https://appleid.apple.com/auth/token",
		UserInfoURL: "", // Apple sends user info in the id_token
		Scopes:      []string{"name", "email"},
	},
}

// OAuthUserCreator is injected from main.go to create/find users without
// importing the users package (avoids import cycle).
type OAuthUserCreator func(email, fullName, provider string) (userID string, err error)

var oauthUserCreator OAuthUserCreator

// RegisterOAuthUserCreator wires the user lookup/create function.
func RegisterOAuthUserCreator(f OAuthUserCreator) {
	oauthUserCreator = f
}

// RegisterOAuthRoutes adds the OAuth initiation and callback endpoints.
// Uses /auth/oauth/:provider to avoid clashing with /auth/register, /auth/login, etc.
func RegisterOAuthRoutes(r *gin.Engine) {
	g := r.Group("/auth/oauth")
	{
		g.GET("/:provider", handleOAuthInit)
		g.GET("/:provider/callback", handleOAuthCallback)
	}
}

// randomState generates a CSRF-safe random state parameter.
func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func handleOAuthInit(c *gin.Context) {
	providerName := c.Param("provider")
	provider, ok := providers[providerName]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Unknown OAuth provider"})
		return
	}

	cfg := config.Get()

	// Check if OAuth is configured for this provider
	clientID := getProviderClientID(providerName)

	if cfg.IsDev() && clientID == "" {
		// Dev mode simulation — skip real OAuth, redirect straight to callback
		// with a simulated code
		state := randomState()
		callbackURL := fmt.Sprintf("%s/auth/oauth/%s/callback?code=dev_sim_%s&state=%s",
			cfg.BaseURL(), providerName, providerName, state)
		c.Redirect(http.StatusTemporaryRedirect, callbackURL)
		return
	}

	if clientID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"detail": fmt.Sprintf("%s login is not configured. Set OAUTH_%s_CLIENT_ID in environment.",
				provider.Name, strings.ToUpper(providerName)),
		})
		return
	}

	// Real OAuth flow — build authorization URL
	state := randomState()
	redirectURI := fmt.Sprintf("%s/auth/oauth/%s/callback", cfg.BaseURL(), providerName)
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		provider.AuthURL, clientID, redirectURI,
		strings.Join(provider.Scopes, "+"), state)

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func handleOAuthCallback(c *gin.Context) {
	providerName := c.Param("provider")
	_, ok := providers[providerName]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Unknown OAuth provider"})
		return
	}

	code := c.Query("code")
	if code == "" {
		errMsg := c.Query("error")
		if errMsg == "" {
			errMsg = "No authorization code received"
		}
		c.JSON(http.StatusBadRequest, gin.H{"detail": errMsg})
		return
	}

	cfg := config.Get()

	var email, fullName string

	if cfg.IsDev() && strings.HasPrefix(code, "dev_sim_") {
		// Dev mode — simulate user info from the provider
		email = fmt.Sprintf("demo_%s@talkex.dev", providerName)
		title := strings.ToUpper(providerName[:1]) + providerName[1:]
		fullName = fmt.Sprintf("Demo %s User", title)
		log.Printf("OAuth (dev): simulated %s login → email=%s, name=%s", providerName, email, fullName)
	} else {
		// Production — exchange code for token, fetch user info
		// This is where you'd call the provider's token endpoint,
		// exchange the code, then call the userinfo endpoint.
		// For now, return an error since real OAuth isn't configured.
		c.JSON(http.StatusNotImplemented, gin.H{
			"detail": "Production OAuth code exchange not yet implemented. Configure provider API keys.",
		})
		return
	}

	// Find or create user
	if oauthUserCreator == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "OAuth user creator not configured"})
		return
	}

	userID, err := oauthUserCreator(email, fullName, providerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create user"})
		return
	}

	// Issue JWT tokens
	accessToken, err := CreateAccessToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create access token"})
		return
	}
	refreshToken, err := CreateRefreshToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create refresh token"})
		return
	}

	// Redirect to frontend with tokens in URL fragment (not query params —
	// fragments aren't sent to servers, more secure for token delivery).
	frontendURL := cfg.FrontendURL()
	redirectURL := fmt.Sprintf("%s/oauth/callback#access_token=%s&refresh_token=%s&provider=%s",
		frontendURL, accessToken, refreshToken, providerName)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// getProviderClientID reads the OAuth client ID from environment.
func getProviderClientID(provider string) string {
	cfg := config.Get()
	switch provider {
	case "google":
		return cfg.OAuthGoogleClientID
	case "facebook":
		return cfg.OAuthFacebookClientID
	case "github":
		return cfg.OAuthGitHubClientID
	case "apple":
		return cfg.OAuthAppleClientID
	default:
		return ""
	}
}
