package reseller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/apihelpers"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/reseller")
	g.Use(auth.AuthRequired())
	{
		g.GET("/dashboard", handleDashboard)
	}
}

func handleDashboard(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	days := 30
	if v := c.Query("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 365 {
			days = n
		}
	}
	d, err := Build(database.DB, ownerID, days)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Not a reseller — no reseller-tier organization found for this account"})
		return
	}
	if err != nil {
		apihelpers.ServerError(c, err, "internal")
		return
	}
	c.JSON(http.StatusOK, d)
}
