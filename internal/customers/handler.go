package customers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/customers")
	g.Use(auth.AuthRequired())
	{
		g.GET("/me", handleGet)
		g.PUT("/me", handleUpsert)
	}
}

func handleGet(c *gin.Context) {
	cust, err := GetByOwner(database.DB, auth.GetUserID(c))
	if err == ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "No business profile yet"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, cust)
}

func handleUpsert(c *gin.Context) {
	var in UpsertInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	cust, err := Upsert(database.DB, auth.GetUserID(c), &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, cust)
}
