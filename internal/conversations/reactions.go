package conversations

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// react POST /conversations/:id/messages/:message_id/react — set or
// clear a reaction emoji on a message. Sending "" clears. Overwrites
// any prior reaction on the same message (WhatsApp semantics).
func RegisterReactionRoutes(r *gin.Engine) {
	g := r.Group("/conversations")
	g.Use(auth.AuthRequired())
	g.POST("/:id/messages/:message_id/react", handleReact)
}

func handleReact(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	convID := c.Param("id")
	msgID := c.Param("message_id")

	var body struct {
		Emoji string `json:"emoji"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	// Cap length — a UTF-8 emoji cluster is at most a few bytes.
	if len(body.Emoji) > 16 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "emoji too long"})
		return
	}

	// Verify ownership: the conversation belongs to the caller.
	var conv Conversation
	err := database.DB.Where("id = ? AND owner_id = ?", convID, ownerID).First(&conv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Conversation not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}

	// Verify the message belongs to the same conversation.
	var msg Message
	err = database.DB.Where("id = ? AND conversation_id = ?", msgID, convID).First(&msg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Message not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}

	// Update the message row's Reaction summary.
	msg.Reaction = body.Emoji
	if err := database.DB.Model(&msg).Update("reaction", body.Emoji).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}

	// Append a Reaction event for the audit trail (only if a real emoji;
	// clearing doesn't need an event row).
	if body.Emoji != "" {
		_ = database.DB.Create(&Reaction{
			OwnerID:        ownerID,
			MessageID:      msgID,
			ConversationID: convID,
			Emoji:          body.Emoji,
			Direction:      "outbound", // agent reacted
			UserID:         &ownerID,
		}).Error
	}
	c.JSON(http.StatusOK, msg)
	_ = time.Now() // reserved for future timing telemetry
}
