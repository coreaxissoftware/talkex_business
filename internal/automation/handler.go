package automation

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/automation/rules")
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
	rule, err := Create(database.DB, auth.GetUserID(c), &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func ownedOr404(c *gin.Context) *Rule {
	r, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrRuleNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Rule not found"})
		return nil
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return nil
	}
	return r
}

func handleGet(c *gin.Context) {
	if r := ownedOr404(c); r != nil {
		c.JSON(http.StatusOK, r)
	}
}

func handleUpdate(c *gin.Context) {
	r := ownedOr404(c)
	if r == nil {
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
	r := ownedOr404(c)
	if r == nil {
		return
	}
	if err := Delete(database.DB, r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
