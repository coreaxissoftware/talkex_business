package customfields

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/custom-fields")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.PATCH("/:id", handleUpdate)
		g.DELETE("/:id", handleDelete)
	}
}

func handleList(c *gin.Context) {
	fields, err := List(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, fields)
}

func handleCreate(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	f, err := Create(database.DB, auth.GetUserID(c), &in)
	if err == ErrDuplicateName || err == ErrInvalidType {
		c.JSON(http.StatusConflict, gin.H{"detail": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, f)
}

func handleUpdate(c *gin.Context) {
	f, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrFieldNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Custom field not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	updated, err := Update(database.DB, f, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleDelete(c *gin.Context) {
	f, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrFieldNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Custom field not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	if err := Delete(database.DB, f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
