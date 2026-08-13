package notifications

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/notifications")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.GET("/unread-count", handleUnreadCount)
		g.POST("/:id/read", handleMarkRead)
		g.POST("/read-all", handleMarkAllRead)
	}
}

func handleList(c *gin.Context) {
	opts := ListOptions{
		OwnerID:    auth.GetUserID(c),
		UnreadOnly: c.Query("unread") == "true",
	}
	items, err := List(database.DB, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleUnreadCount(c *gin.Context) {
	n, err := UnreadCount(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": n})
}

func handleMarkRead(c *gin.Context) {
	err := MarkRead(database.DB, auth.GetUserID(c), c.Param("id"))
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	case ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"detail": "Notification not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
	}
}

func handleMarkAllRead(c *gin.Context) {
	if err := MarkAllRead(database.DB, auth.GetUserID(c)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
