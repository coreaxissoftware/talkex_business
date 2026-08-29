package organizations

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/organizations")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.GET("/:id", handleGet)
		g.PATCH("/:id", handleUpdate)
		g.GET("/:id/members", handleListMembers)
		g.POST("/:id/members", handleAddMember)
		g.DELETE("/:id/members/:uid", handleRemoveMember)
		g.GET("/:id/sub-orgs", handleListSubOrgs)
	}
}

func handleList(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	orgs, err := GetByOwner(database.DB, ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, orgs)
}

func handleCreate(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	var req CreateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	org, err := Create(database.DB, ownerID, &req)
	if err == ErrSlugTaken {
		c.JSON(http.StatusConflict, gin.H{"detail": err.Error()})
		return
	}
	if err == ErrMaxSubOrgs {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, org)
}

func handleGet(c *gin.Context) {
	orgID := c.Param("id")
	// Membership gate: caller must own or be a member. Anything else
	// returns 404 (not 403) so we don't leak org-exists metadata.
	if !IsMemberOrOwner(database.DB, orgID, auth.GetUserID(c)) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Organization not found"})
		return
	}
	org, err := GetByID(database.DB, orgID)
	if err == ErrOrgNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Organization not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, org)
}

func handleUpdate(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	org, err := GetByID(database.DB, c.Param("id"))
	if err == ErrOrgNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Organization not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	if org.OwnerID != ownerID {
		c.JSON(http.StatusForbidden, gin.H{"detail": ErrNotOrgOwner.Error()})
		return
	}
	var req UpdateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	updated, err := Update(database.DB, org, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleListMembers(c *gin.Context) {
	orgID := c.Param("id")
	if !IsMemberOrOwner(database.DB, orgID, auth.GetUserID(c)) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Organization not found"})
		return
	}
	members, err := ListMembers(database.DB, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, members)
}

func handleAddMember(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	orgID := c.Param("id")
	org, err := GetByID(database.DB, orgID)
	if err == ErrOrgNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Organization not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	if org.OwnerID != ownerID {
		c.JSON(http.StatusForbidden, gin.H{"detail": ErrNotOrgOwner.Error()})
		return
	}
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	role := req.Role
	if role == "" {
		role = "member"
	}
	m, err := AddMember(database.DB, orgID, req.UserID, role)
	if err == ErrMaxUsers {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func handleRemoveMember(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	orgID := c.Param("id")
	org, err := GetByID(database.DB, orgID)
	if err == ErrOrgNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Organization not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	if org.OwnerID != ownerID {
		c.JSON(http.StatusForbidden, gin.H{"detail": ErrNotOrgOwner.Error()})
		return
	}
	if err := RemoveMember(database.DB, orgID, c.Param("uid")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"detail": "Member removed"})
}

func handleListSubOrgs(c *gin.Context) {
	orgID := c.Param("id")
	// Only owners or members of the parent may enumerate its
	// downstream (reseller hierarchy is confidential).
	if !IsMemberOrOwner(database.DB, orgID, auth.GetUserID(c)) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Organization not found"})
		return
	}
	orgs, err := ListSubOrgs(database.DB, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, orgs)
}
