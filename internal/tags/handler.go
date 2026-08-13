package tags

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/tags")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("/rename", handleRename)
		g.POST("/delete", handleDelete)
		g.POST("/bulk-apply", handleBulkApply)
	}
}

func handleList(c *gin.Context) {
	tags, err := ListAll(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, tags)
}

func handleRename(c *gin.Context) {
	var in struct {
		OldName string `json:"old_name" binding:"required"`
		NewName string `json:"new_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	count, err := Rename(database.DB, auth.GetUserID(c), in.OldName, in.NewName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": count})
}

func handleDelete(c *gin.Context) {
	var in struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	count, err := Delete(database.DB, auth.GetUserID(c), in.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": count})
}

func handleBulkApply(c *gin.Context) {
	var in struct {
		Tag        string   `json:"tag" binding:"required"`
		ContactIDs []string `json:"contact_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	count, err := BulkApply(database.DB, auth.GetUserID(c), in.Tag, in.ContactIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"applied": count})
}
