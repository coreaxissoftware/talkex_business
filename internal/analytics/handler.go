package analytics

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/analytics")
	g.Use(auth.AuthRequired())
	{
		g.GET("/summary", handleSummary)
		g.GET("/timeseries", handleTimeseries)
	}
}

func handleSummary(c *gin.Context) {
	s, err := GetSummary(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func handleTimeseries(c *gin.Context) {
	days := 30
	if v, err := strconv.Atoi(c.Query("days")); err == nil {
		days = v
	}
	series, err := GetTimeseries(database.DB, auth.GetUserID(c), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, series)
}
