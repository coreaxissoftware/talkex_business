package audit

import (
	"fmt"
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
		g.GET("/export-csv", handleExportCSV)
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

func handleExportCSV(c *gin.Context) {
	f := ListFilter{
		UserID: auth.GetUserID(c),
		Limit:  5000,
	}
	entries, _, err := List(database.DB, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=audit-logs.csv")
	c.Writer.WriteString("timestamp,method,path,status_code,success,latency_ms,client_ip,user_email\n")
	for _, e := range entries {
		email := ""
		if e.UserEmail != nil {
			email = *e.UserEmail
		}
		c.Writer.WriteString(fmt.Sprintf("%s,%s,%s,%d,%t,%d,%s,%s\n",
			e.CreatedAt.Format("2006-01-02T15:04:05Z"), e.Method, e.Path, e.StatusCode, e.Success, e.LatencyMs, e.ClientIP, email))
	}
}

func handleStats(c *gin.Context) {
	stats, err := GetStats(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, stats)
}
