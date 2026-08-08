package contacts

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/contacts")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.GET("/:id", handleGet)
		g.PATCH("/:id", handleUpdate)
		g.DELETE("/:id", handleDelete)
	}
}

func handleList(c *gin.Context) {
	contacts, err := List(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

func handleCreate(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	contact, err := Create(database.DB, auth.GetUserID(c), &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, contact)
}

func getOwnedOrAbort(c *gin.Context) *Contact {
	contact, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrContactNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Contact not found"})
		return nil
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return nil
	}
	return contact
}

func handleGet(c *gin.Context) {
	if contact := getOwnedOrAbort(c); contact != nil {
		c.JSON(http.StatusOK, contact)
	}
}

func handleUpdate(c *gin.Context) {
	contact := getOwnedOrAbort(c)
	if contact == nil {
		return
	}
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	updated, err := Update(database.DB, contact, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleDelete(c *gin.Context) {
	contact := getOwnedOrAbort(c)
	if contact == nil {
		return
	}
	if err := Delete(database.DB, contact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
