package conversations

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/conversations")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.GET("/search", handleSearch)
		g.POST("/bulk-assign", handleBulkAssign)
		g.POST("/bulk-read", handleBulkMarkRead)
		g.GET("/:id/messages", handleListMessages)
		g.PATCH("/:id", handleUpdate)
		g.POST("/:id/read", handleMarkRead)
		g.POST("/send", handleSend)
		g.POST("/inbound", handleInbound) // dev/simulator path — real webhook goes elsewhere
	}
}

func handleList(c *gin.Context) {
	items, err := List(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func getOwnedOrAbort(c *gin.Context) *Conversation {
	conv, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrConversationNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Conversation not found"})
		return nil
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return nil
	}
	return conv
}

func handleListMessages(c *gin.Context) {
	conv := getOwnedOrAbort(c)
	if conv == nil {
		return
	}
	msgs, err := ListMessages(database.DB, conv.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"conversation": conv,
		"window_open":  conv.IsWindowOpen(),
		"messages":     msgs,
	})
}

func handleUpdate(c *gin.Context) {
	conv := getOwnedOrAbort(c)
	if conv == nil {
		return
	}
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	updated, err := UpdateConversation(database.DB, conv, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleMarkRead(c *gin.Context) {
	conv := getOwnedOrAbort(c)
	if conv == nil {
		return
	}
	if err := MarkRead(database.DB, conv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, conv)
}

func handleSend(c *gin.Context) {
	var in SendInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	msg, conv, err := SendOutbound(database.DB, auth.GetUserID(c), &in)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, gin.H{"message": msg, "conversation": conv})
	case ErrContactNotFound:
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
	case ErrWindowClosed:
		c.JSON(http.StatusConflict, gin.H{"detail": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
	}
}

func handleSearch(c *gin.Context) {
	items, err := Search(database.DB, auth.GetUserID(c), c.Query("q"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

type bulkAssignReq struct {
	IDs         []string `json:"ids" binding:"required"`
	AgentUserID string   `json:"agent_user_id"`
	AgentName   string   `json:"agent_name"`
}

func handleBulkAssign(c *gin.Context) {
	var req bulkAssignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	affected, err := BulkAssign(database.DB, auth.GetUserID(c), req.IDs, req.AgentUserID, req.AgentName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": affected})
}

func handleBulkMarkRead(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	affected, err := BulkMarkRead(database.DB, auth.GetUserID(c), req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": affected})
}

func handleInbound(c *gin.Context) {
	var in InboundInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	msg, conv, err := RecordInbound(database.DB, auth.GetUserID(c), &in)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, gin.H{"message": msg, "conversation": conv})
	case ErrContactNotFound:
		c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
	}
}
