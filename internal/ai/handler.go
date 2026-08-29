package ai

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// ConversationFetcher is injected from main.go so this package doesn't
// import conversations (which imports us indirectly through hooks).
type ConversationFetcher func(ownerID, conversationID string) (contactName string, messages []Message, err error)

var fetcher ConversationFetcher

func RegisterConversationFetcher(f ConversationFetcher) { fetcher = f }

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/ai")
	g.Use(auth.AuthRequired())
	{
		g.POST("/suggest-reply", handleSuggest)
		g.POST("/summarize", handleSummarize)
		g.POST("/sentiment", handleSentiment)
		g.GET("/status", handleStatus)
	}
}

type conversationReq struct {
	ConversationID string `json:"conversation_id" binding:"required"`
}

func loadConversation(c *gin.Context) (string, []Message, bool) {
	if fetcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "AI service not wired"})
		return "", nil, false
	}
	var req conversationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return "", nil, false
	}
	_ = database.DB // parity with other handlers
	name, msgs, err := fetcher(auth.GetUserID(c), req.ConversationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Conversation not found"})
		return "", nil, false
	}
	return name, msgs, true
}

func handleSuggest(c *gin.Context) {
	name, msgs, ok := loadConversation(c)
	if !ok {
		return
	}
	reply, err := SuggestReply(c.Request.Context(), msgs, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "AI request failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"suggestion": reply, "dev_mode": devMode()})
}

func handleSummarize(c *gin.Context) {
	name, msgs, ok := loadConversation(c)
	if !ok {
		return
	}
	summary, err := Summarize(c.Request.Context(), msgs, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "AI request failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": summary, "dev_mode": devMode()})
}

func handleSentiment(c *gin.Context) {
	_, msgs, ok := loadConversation(c)
	if !ok {
		return
	}
	res, err := Sentiment(c.Request.Context(), msgs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "AI request failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"score":    res.Score,
		"reason":   res.Reason,
		"dev_mode": devMode(),
	})
}

// handleStatus reports whether real API is configured.
func handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"dev_mode": devMode(),
		"model":    string(modelID()),
		"available": true,
		"key_configured": os.Getenv("ANTHROPIC_API_KEY") != "",
	})
}
