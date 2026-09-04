package paylinks

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/apihelpers"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/pay-links")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.POST("/:id/sent", handleMarkSent)
	}
}

func handleList(c *gin.Context) {
	items, err := List(database.DB, auth.GetUserID(c), c.Query("status"))
	if err != nil {
		apihelpers.ServerError(c, err, "internal")
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleCreate(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	pl, err := Create(database.DB, auth.GetUserID(c), &in)
	if errors.Is(err, ErrInvalidInput) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err != nil {
		apihelpers.ServerError(c, err, "internal")
		return
	}
	c.JSON(http.StatusCreated, pl)
}

func handleMarkSent(c *gin.Context) {
	if err := MarkSent(database.DB, c.Param("id")); err != nil {
		apihelpers.ServerError(c, err, "internal")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
