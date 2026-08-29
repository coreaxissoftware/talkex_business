package team

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/team")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.GET("/activity", handleActivity)
		g.POST("/invite", handleInvite)
		g.PATCH("/:id/role", handleUpdateRole)
		g.DELETE("/:id", handleRemove)
	}
}

func handleActivity(c *gin.Context) {
	items, err := Activity(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func handleList(c *gin.Context) {
	members, err := List(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, members)
}

func handleInvite(c *gin.Context) {
	var in InviteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	m, err := Invite(database.DB, auth.GetUserID(c), &in)
	if err == ErrAlreadyInvited {
		c.JSON(http.StatusConflict, gin.H{"detail": err.Error()})
		return
	}
	if err == ErrInvalidRole {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func getOwnedOrAbort(c *gin.Context) *Member {
	m, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrMemberNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Team member not found"})
		return nil
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return nil
	}
	return m
}

func handleUpdateRole(c *gin.Context) {
	m := getOwnedOrAbort(c)
	if m == nil {
		return
	}
	var in UpdateRoleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	updated, err := UpdateRole(database.DB, m, in.Role)
	if err == ErrInvalidRole {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleRemove(c *gin.Context) {
	m := getOwnedOrAbort(c)
	if m == nil {
		return
	}
	if err := Remove(database.DB, m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
