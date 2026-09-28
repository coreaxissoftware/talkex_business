package platformenv

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/config"
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"github.com/coreaxissoftware/talkex_business/internal/users"
)

// RegisterRoutes mounts /admin/platform-settings under the auth
// middleware plus a superadmin gate. Superadmin is *not* a tenant
// role — it's a platform-level identity granted only to operator
// accounts whose email is listed in PLATFORM_SUPERADMIN_EMAILS.
func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/admin/platform-settings")
	g.Use(auth.AuthRequired(), superadminOnly())
	{
		g.GET("", handleList)
		g.PUT("/:key", handleSet)
		g.DELETE("/:key", handleClear)
	}
}

// superadminOnly rejects any request whose authenticated user's email
// isn't on the PLATFORM_SUPERADMIN_EMAILS allow-list.
//
// A comma-separated list keeps operator management out of the DB
// entirely — if the DB itself is compromised, the attacker still
// can't grant themselves platform-admin without also editing the
// systemd env file.
//
// In development (empty env) the gate is open so a local dev doesn't
// have to jump through hoops.
func superadminOnly() gin.HandlerFunc {
	allow := parseAllow(os.Getenv("PLATFORM_SUPERADMIN_EMAILS"))
	cfg := config.Get()
	return func(c *gin.Context) {
		if cfg.IsDev() && len(allow) == 0 {
			c.Next()
			return
		}
		if len(allow) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"detail": "Platform admin console is disabled: PLATFORM_SUPERADMIN_EMAILS is unset on the server.",
			})
			return
		}
		var u users.User
		if err := database.DB.Select("email").
			Where("id = ?", auth.GetUserID(c)).
			First(&u).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "Forbidden"})
			return
		}
		if !allow[strings.ToLower(strings.TrimSpace(u.Email))] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "Forbidden"})
			return
		}
		c.Next()
	}
}

func parseAllow(csv string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(csv, ",") {
		e := strings.ToLower(strings.TrimSpace(part))
		if e != "" {
			out[e] = true
		}
	}
	return out
}

func handleList(c *gin.Context) {
	items, err := List(database.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func handleSet(c *gin.Context) {
	key := c.Param("key")
	var in SetInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err := Set(database.DB, key, in.Value, auth.GetUserID(c)); err != nil {
		if err == ErrUnknownKey {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "unknown platform setting key"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	// Bust config singleton so the very next provider call reads the
	// new value without waiting for a systemctl restart.
	config.Reload()
	// Return the fresh masked item so the UI can update in place
	// without a second round-trip.
	items, _ := List(database.DB)
	for _, it := range items {
		if it.Key == strings.ToUpper(key) {
			c.JSON(http.StatusOK, it)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func handleClear(c *gin.Context) {
	key := c.Param("key")
	if err := Set(database.DB, key, "", auth.GetUserID(c)); err != nil {
		if err == ErrUnknownKey {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "unknown platform setting key"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	config.Reload()
	c.Status(http.StatusNoContent)
}
