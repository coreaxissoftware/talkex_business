package developers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/api-keys")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.POST("/:id/revoke", handleRevoke)
		g.DELETE("/:id", handleDelete)
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

func handleCreate(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	result, err := Create(database.DB, auth.GetUserID(c), &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func handleRevoke(c *gin.Context) {
	updated, err := Revoke(database.DB, auth.GetUserID(c), c.Param("id"))
	switch err {
	case nil:
		c.JSON(http.StatusOK, updated)
	case ErrKeyNotFound:
		c.JSON(http.StatusNotFound, gin.H{"detail": "API key not found"})
	case ErrAlreadyRevoked:
		c.JSON(http.StatusConflict, gin.H{"detail": "API key is already revoked"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
	}
}

func handleDelete(c *gin.Context) {
	err := Delete(database.DB, auth.GetUserID(c), c.Param("id"))
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	case ErrKeyNotFound:
		c.JSON(http.StatusNotFound, gin.H{"detail": "API key not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
	}
}
