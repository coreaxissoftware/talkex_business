package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthRequired is a Gin middleware that extracts and validates the Bearer
// access token. On success it sets "user_id" in the context for downstream
// handlers to use.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Missing or invalid Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := ValidateToken(tokenStr, AccessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Could not validate credentials"})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// GetUserID extracts the authenticated user's ID from the Gin context.
// Must be called after AuthRequired middleware.
func GetUserID(c *gin.Context) string {
	id, _ := c.Get("user_id")
	return id.(string)
}
