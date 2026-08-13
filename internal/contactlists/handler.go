package contactlists

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/contact-lists")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleList)
		g.POST("", handleCreate)
		g.GET("/:id", handleGet)
		g.PATCH("/:id", handleUpdate)
		g.DELETE("/:id", handleDelete)
		g.GET("/:id/members", handleGetMembers)
		g.POST("/:id/members", handleAddMembers)
		g.DELETE("/:id/members", handleRemoveMembers)
	}
}

func handleList(c *gin.Context) {
	lists, err := List(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, lists)
}

func handleCreate(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	list, err := Create(database.DB, auth.GetUserID(c), &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, list)
}

func getOwnedOrAbort(c *gin.Context) *ContactList {
	list, err := GetByID(database.DB, auth.GetUserID(c), c.Param("id"))
	if err == ErrListNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Contact list not found"})
		return nil
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return nil
	}
	return list
}

func handleGet(c *gin.Context) {
	if list := getOwnedOrAbort(c); list != nil {
		c.JSON(http.StatusOK, list)
	}
}

func handleUpdate(c *gin.Context) {
	list := getOwnedOrAbort(c)
	if list == nil {
		return
	}
	var in UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	updated, err := Update(database.DB, list, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func handleDelete(c *gin.Context) {
	list := getOwnedOrAbort(c)
	if list == nil {
		return
	}
	if err := Delete(database.DB, list); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}

func handleGetMembers(c *gin.Context) {
	list := getOwnedOrAbort(c)
	if list == nil {
		return
	}
	ids, err := GetMembers(database.DB, list.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"contact_ids": ids})
}

func handleAddMembers(c *gin.Context) {
	list := getOwnedOrAbort(c)
	if list == nil {
		return
	}
	var in AddMembersInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	added, err := AddMembers(database.DB, list.ID, in.ContactIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"added": added})
}

func handleRemoveMembers(c *gin.Context) {
	list := getOwnedOrAbort(c)
	if list == nil {
		return
	}
	var in AddMembersInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if err := RemoveMembers(database.DB, list.ID, in.ContactIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"removed": len(in.ContactIDs)})
}
