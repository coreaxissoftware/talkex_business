package customers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"github.com/coreaxissoftware/talkex_business/internal/apihelpers"
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
		// A brand-new tenant hasn't filled in their business profile
		// yet — that's normal, not an error. Return 200 with null so
		// the frontend can render the empty form without treating
		// this as a real 404.
		c.JSON(http.StatusOK, nil)
		return
	}
	if err != nil {
		apihelpers.ServerError(c, err, "customers.handleGet")
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
		apihelpers.ServerError(c, err, "customers.handleUpsert")
		return
	}
	c.JSON(http.StatusOK, cust)
}
