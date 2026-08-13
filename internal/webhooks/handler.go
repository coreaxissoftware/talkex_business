package webhooks

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/webhooks")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.DELETE("/:id", handleDelete)
		g.GET("/:id/deliveries", handleDeliveries)
	}
}

func handleList(c *gin.Context) {
	items, err := ListEndpoints(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleCreate(c *gin.Context) {
	var in CreateEndpointInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	result, err := CreateEndpoint(database.DB, auth.GetUserID(c), &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func handleDelete(c *gin.Context) {
	e, err := GetEndpoint(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrEndpointNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Webhook endpoint not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	if err := DeleteEndpoint(database.DB, e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}

func handleDeliveries(c *gin.Context) {
	items, err := ListDeliveries(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrEndpointNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Webhook endpoint not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}
