package campaigns

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/campaigns")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.GET("/:id", handleGet)
		g.PATCH("/:id", handleUpdate)
		g.POST("/:id/launch", handleLaunch)
		g.POST("/:id/cancel", handleCancel)
		g.POST("/:id/approve", handleApprove)
		g.POST("/:id/reject", handleReject)
		g.POST("/:id/clone", handleClone)
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
	created, err := Create(database.DB, auth.GetUserID(c), &in)
	if err != nil {
		switch err {
		case ErrTemplateNotFound:
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
		case ErrNoRecipients:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		}
		return
	}
	c.JSON(http.StatusCreated, created)
}

func getOwnedOrAbort(c *gin.Context) *Campaign {
	camp, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrCampaignNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Campaign not found"})
		return nil
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return nil
	}
	return camp
}

func handleGet(c *gin.Context) {
	if camp := getOwnedOrAbort(c); camp != nil {
		c.JSON(http.StatusOK, camp)
	}
}

func handleUpdate(c *gin.Context) {
	camp := getOwnedOrAbort(c)
	if camp == nil {
		return
	}
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	updated, err := Update(database.DB, camp, &in)
	if err == ErrInvalidStatus {
		c.JSON(http.StatusConflict, gin.H{"detail": "Only draft/scheduled campaigns can be edited"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleLaunch(c *gin.Context) {
	camp := getOwnedOrAbort(c)
	if camp == nil {
		return
	}
	updated, err := Launch(database.DB, camp)
	if err == ErrInvalidStatus {
		c.JSON(http.StatusConflict, gin.H{"detail": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleCancel(c *gin.Context) {
	camp := getOwnedOrAbort(c)
	if camp == nil {
		return
	}
	updated, err := Cancel(database.DB, camp)
	if err == ErrInvalidStatus {
		c.JSON(http.StatusConflict, gin.H{"detail": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleApprove(c *gin.Context) {
	camp := getOwnedOrAbort(c)
	if camp == nil {
		return
	}
	updated, err := Approve(database.DB, camp, auth.GetUserID(c))
	if err == ErrInvalidStatus {
		c.JSON(http.StatusConflict, gin.H{"detail": "Campaign is not pending approval"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleReject(c *gin.Context) {
	camp := getOwnedOrAbort(c)
	if camp == nil {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	updated, err := Reject(database.DB, camp, auth.GetUserID(c), req.Reason)
	if err == ErrInvalidStatus {
		c.JSON(http.StatusConflict, gin.H{"detail": "Campaign is not pending approval"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleClone(c *gin.Context) {
	camp := getOwnedOrAbort(c)
	if camp == nil {
		return
	}
	cloned, err := Clone(database.DB, camp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, cloned)
}

func handleDelete(c *gin.Context) {
	camp := getOwnedOrAbort(c)
	if camp == nil {
		return
	}
	if err := Delete(database.DB, camp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
