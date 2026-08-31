package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ApiKeyResolver — injected from main.go so this package doesn't need to
// import developers (which imports database/models and would create a
// circular reference the wrong way).
//
// Return (ownerID, true) when the raw string is a valid, non-revoked
// API key; (_, false) otherwise. Called only when the Bearer token
// fails JWT validation, so real users pay no extra cost.
type ApiKeyResolver func(raw string) (ownerID string, ok bool)

var apiKeyResolver ApiKeyResolver

func RegisterApiKeyResolver(f ApiKeyResolver) { apiKeyResolver = f }

// AuthRequired accepts either a JWT access token or a Bearer API key.
// On success it sets "user_id" in the context for downstream handlers.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		// Case-insensitive Bearer prefix — some SDKs and Postman send
		// "bearer" or "BEARER". Split on the first space so we don't
		// silently succeed on malformed inputs like "Bearerxxx".
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Missing or invalid Authorization header"})
			return
		}
		tokenStr := strings.TrimSpace(parts[1])

		// Try JWT first — the common case for browser sessions.
		if userID, err := ValidateToken(tokenStr, AccessToken); err == nil {
			c.Set("user_id", userID)
			c.Set("auth_method", "jwt")
			c.Next()
			return
		}

		// Fall back to API key resolution when a resolver is registered
		// (i.e. production; unit tests without full wiring skip this path).
		if apiKeyResolver != nil {
			if ownerID, ok := apiKeyResolver(tokenStr); ok {
				c.Set("user_id", ownerID)
				c.Set("auth_method", "api_key")
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Could not validate credentials"})
	}
}

// GetUserID extracts the authenticated user's ID from the Gin context.
// Must be called after AuthRequired middleware.
func GetUserID(c *gin.Context) string {
	id, _ := c.Get("user_id")
	return id.(string)
}
