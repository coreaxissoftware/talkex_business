package audit

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/audit-logs")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.GET("/stats", handleStats)
	}
}

func handleList(c *gin.Context) {
	f := ListFilter{
		UserID:     auth.GetUserID(c),
		OnlyFailed: c.Query("failed") == "true",
		Method:     c.Query("method"),
		Search:     c.Query("search"),
	}
	if v, err := strconv.Atoi(c.Query("limit")); err == nil {
		f.Limit = v
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil {
		f.Offset = v
	}

	entries, total, err := List(database.DB, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": entries, "total": total})
}

func handleStats(c *gin.Context) {
	stats, err := GetStats(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, stats)
}
