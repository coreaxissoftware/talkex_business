package canned

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/canned-responses")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.PATCH("/:id", handleUpdate)
		g.DELETE("/:id", handleDelete)
		g.POST("/:id/use", handleBumpUsage)
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
	created, err := Create(database.DB, auth.GetUserID(c), &in)
	if err == ErrDuplicate {
		c.JSON(http.StatusConflict, gin.H{"detail": "Shortcut already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func handleUpdate(c *gin.Context) {
	r, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Not found"})
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
	updated, err := Update(database.DB, r, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleDelete(c *gin.Context) {
	r, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	if err := Delete(database.DB, r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}

func handleBumpUsage(c *gin.Context) {
	// Owner-scope check: reject foreign IDs so cross-tenant callers
	// can't skew another owner's picker ordering (usage_count DESC).
	if _, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id")); err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	BumpUsage(database.DB, c.Param("id"))
	c.Status(http.StatusNoContent)
}
