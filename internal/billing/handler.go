package billing

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"github.com/coreaxissoftware/talkex_business/internal/apihelpers"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/billing")
	g.Use(auth.AuthRequired())
	{
		g.GET("/plans", handleListPlans)
		g.GET("/subscription", handleGetSubscription)
		g.POST("/subscription", handleChangePlan)
		g.GET("/invoices", handleListInvoices)
	}
}

func handleListPlans(c *gin.Context) {
	c.JSON(http.StatusOK, AllPlans)
}

func handleGetSubscription(c *gin.Context) {
	sub, err := GetOrCreate(database.DB, auth.GetUserID(c))
	if err != nil {
		apihelpers.ServerError(c, err, "billing.handleGetSubscription")
		return
	}
	plan := PlanByID(sub.Plan)
	c.JSON(http.StatusOK, gin.H{"subscription": sub, "plan": plan})
}

type changePlanInput struct {
	Plan PlanID `json:"plan" binding:"required"`
}

func handleChangePlan(c *gin.Context) {
	var in changePlanInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	sub, invoice, err := ChangePlan(database.DB, auth.GetUserID(c), in.Plan)
	switch err {
	case nil:
		c.JSON(http.StatusOK, gin.H{"subscription": sub, "invoice": invoice})
	case ErrUnknownPlan:
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "unknown plan"})
	default:
		apihelpers.ServerError(c, err, "billing.handleChangePlan")
	}
}

func handleListInvoices(c *gin.Context) {
	items, err := ListInvoices(database.DB, auth.GetUserID(c))
	if err != nil {
		apihelpers.ServerError(c, err, "billing.handleListInvoices")
		return
	}
	c.JSON(http.StatusOK, items)
}
