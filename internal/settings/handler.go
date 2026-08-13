package settings

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/settings")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleGet)
		g.PUT("", handleSave)
	}
}

func handleGet(c *gin.Context) {
	_, prefs, err := Get(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, prefs)
}

func handleSave(c *gin.Context) {
	var prefs PrefsData
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	_, err := Save(database.DB, auth.GetUserID(c), &prefs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, prefs)
}
