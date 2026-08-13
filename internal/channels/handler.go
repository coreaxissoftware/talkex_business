package channels

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/channels")
	g.Use(auth.AuthRequired())
	{
		g.GET("/catalog", handleCatalog)
		g.GET("", handleList)
		g.PUT("/:kind", handleSetEnabled)
	}
}

func handleCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, Catalog)
}

func handleList(c *gin.Context) {
	configs, err := ListConfigs(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, configs)
}

func handleSetEnabled(c *gin.Context) {
	var in SetEnabledInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	cfg, err := SetEnabled(database.DB, auth.GetUserID(c), Kind(c.Param("kind")), &in)
	switch err {
	case nil:
		c.JSON(http.StatusOK, cfg)
	case ErrUnknownKind:
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "Unknown channel kind"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
	}
}
