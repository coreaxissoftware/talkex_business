package messaging

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/messaging")
	g.Use(auth.AuthRequired())
	{
		g.GET("/queue-stats", handleQueueStats)
		g.GET("/dead-letters", handleListDLQ)
		g.POST("/dead-letters/:id/retry", handleRetryDLQ)
		g.POST("/dead-letters/:id/discard", handleDiscardDLQ)
	}
}

func handleQueueStats(c *gin.Context) {
	stats := GetQueueStats(database.DB, auth.GetUserID(c))
	c.JSON(http.StatusOK, stats)
}

func handleListDLQ(c *gin.Context) {
	includeResolved := c.Query("include_resolved") == "true"
	items, err := ListDeadLetters(database.DB, auth.GetUserID(c), includeResolved)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleRetryDLQ(c *gin.Context) {
	if err := RetryDeadLetter(database.DB, auth.GetUserID(c), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not retry"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "Re-enqueued"})
}

func handleDiscardDLQ(c *gin.Context) {
	if err := DiscardDeadLetter(database.DB, auth.GetUserID(c), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Could not discard"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "Discarded"})
}
