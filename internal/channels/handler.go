package channels

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	talkexch "github.com/coreaxissoftware/talkex_business/internal/channels/talkex"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/channels")
	g.Use(auth.AuthRequired())
	{
		g.GET("/catalog", handleCatalog)
		g.GET("", handleList)
		g.PUT("/:kind", handleSetEnabled)
		// TalkEx-specific — the "Generate TalkEx key" button. Kept on
		// the channels group (not talkex's own package) so the frontend
		// only needs one auth-gated origin for the whole panel.
		g.POST("/talkex/generate-key", handleGenerateTalkExKey)
	}
}

func handleCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, Catalog)
}

func handleList(c *gin.Context) {
	configs, err := ListConfigs(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, configs)
}

func handleSetEnabled(c *gin.Context) {
	var in SetEnabledInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	cfg, err := SetEnabled(database.DB, auth.GetUserID(c), Kind(c.Param("kind")), &in)
	switch err {
	case nil:
		c.JSON(http.StatusOK, cfg)
	case ErrUnknownKind:
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "Unknown channel kind"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
	}
}

// handleGenerateTalkExKey — mints a TalkEx bulk API key from the
// merchant's TalkEx creds and (on success) auto-saves it into the
// TalkEx channel config so the merchant doesn't have to copy-paste.
// Password is never persisted.
func handleGenerateTalkExKey(c *gin.Context) {
	var in talkexch.GenerateRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	out, err := talkexch.Generate(&in)
	if err != nil {
		if errors.Is(err, talkexch.ErrTalkExAuth) {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid TalkEx username or password."})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"detail": "Could not reach TalkEx: " + err.Error()})
		return
	}
	// Two-factor branch — return without saving; frontend re-prompts.
	if out.RequiresPIN {
		c.JSON(http.StatusOK, out)
		return
	}

	// Auto-save the fresh key into the TalkEx channel config so the
	// merchant doesn't have to copy it back into the form.
	ownerID := auth.GetUserID(c)
	saveCfg := map[string]interface{}{
		"api_key":  out.Key,
		"base_url": in.BaseURL, // empty = poller uses defaultBaseURL
	}
	if _, err := SetEnabled(database.DB, ownerID, KindTalkEx, &SetEnabledInput{
		Enabled: true,
		Config:  saveCfg,
	}); err != nil {
		// Non-fatal — still return the key so the merchant can save it
		// manually if the auto-save races with another update.
		c.JSON(http.StatusOK, gin.H{
			"key":     out.Key,
			"key_id":  out.KeyID,
			"prefix":  out.Prefix,
			"label":   out.Label,
			"warning": "Key generated but could not auto-save to channel config; please save manually.",
		})
		return
	}
	c.JSON(http.StatusOK, out)
}
