package csat

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	// Public rating submission — the customer (not authed) can rate a
	// conversation through a magic link/token in production. Kept simple
	// for now with owner-scoped submission via authed endpoints.
	g := r.Group("/csat")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.GET("/summary", handleSummary)
		g.POST("", handleSubmit)
	}
}

func handleList(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, err := List(database.DB, auth.GetUserID(c), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleSummary(c *gin.Context) {
	summary, err := GetSummary(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func handleSubmit(c *gin.Context) {
	var in SubmitInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	created, err := Submit(database.DB, auth.GetUserID(c), &in)
	if err == ErrInvalidScore {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err == ErrConversationNotOwned {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Conversation not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, created)
}
