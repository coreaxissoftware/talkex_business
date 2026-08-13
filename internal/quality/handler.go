package quality

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/quality")
	g.Use(auth.AuthRequired())
	{
		g.GET("/stats", handleStats)
		g.GET("/events", handleEvents)
		g.POST("/events", handleRecordEvent)
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

func handleEvents(c *gin.Context) {
	events, err := ListEvents(database.DB, auth.GetUserID(c), 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, events)
}

type recordEventInput struct {
	ContactID string    `json:"contact_id" binding:"required"`
	Channel   string    `json:"channel" binding:"required"`
	Type      EventType `json:"type" binding:"required"`
	Reason    *string   `json:"reason"`
}

func handleRecordEvent(c *gin.Context) {
	var in recordEventInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err := RecordEvent(database.DB, auth.GetUserID(c), in.ContactID, in.Channel, in.Type, in.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"detail": "Event recorded"})
}
