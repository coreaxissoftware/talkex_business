package greentick

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/green-tick")
	g.Use(auth.AuthRequired())
	{
		g.GET("", handleGet)
		g.PATCH("", handleUpdate)
		g.POST("/submit", handleSubmit)
		g.POST("/decision", handleDecision)
	}
}

func getOrCreate(ownerID string) (*Application, error) {
	var a Application
	err := database.DB.Where("owner_id = ?", ownerID).First(&a).Error
	if err == nil {
		return &a, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	a = Application{OwnerID: ownerID, Status: StatusNotStarted}
	if err := database.DB.Create(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func handleGet(c *gin.Context) {
	a, err := getOrCreate(auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"application": a, "progress": a.Progress()})
}

func handleUpdate(c *gin.Context) {
	a, err := getOrCreate(auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	var in struct {
		NotableBrand     *bool `json:"notable_brand"`
		OrgWebsite       *bool `json:"org_website"`
		Meta200Msg       *bool `json:"meta_200_msg"`
		MetaTier2        *bool `json:"meta_tier2"`
		BusinessVerified *bool `json:"business_verified"`
		TrademarkRefs    *bool `json:"trademark_refs"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	if in.NotableBrand != nil {
		a.NotableBrand = *in.NotableBrand
	}
	if in.OrgWebsite != nil {
		a.OrgWebsite = *in.OrgWebsite
	}
	if in.Meta200Msg != nil {
		a.Meta200Msg = *in.Meta200Msg
	}
	if in.MetaTier2 != nil {
		a.MetaTier2 = *in.MetaTier2
	}
	if in.BusinessVerified != nil {
		a.BusinessVerified = *in.BusinessVerified
	}
	if in.TrademarkRefs != nil {
		a.TrademarkRefs = *in.TrademarkRefs
	}
	if a.Status == StatusNotStarted && a.Progress() > 0 {
		a.Status = StatusInProgress
	}
	if err := database.DB.Save(a).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"application": a, "progress": a.Progress()})
}

func handleSubmit(c *gin.Context) {
	a, err := getOrCreate(auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	if a.Progress() < 1.0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Complete every checklist item before submitting"})
		return
	}
	var body struct {
		MetaCaseID string `json:"meta_case_id"`
	}
	_ = c.ShouldBindJSON(&body)
	now := time.Now()
	a.Status = StatusSubmitted
	a.SubmittedAt = &now
	a.MetaCaseID = body.MetaCaseID
	if err := database.DB.Save(a).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func handleDecision(c *gin.Context) {
	a, err := getOrCreate(auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	var body struct {
		Approved bool   `json:"approved"`
		Reason   string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}
	now := time.Now()
	if body.Approved {
		a.Status = StatusApproved
	} else {
		a.Status = StatusRejected
		a.RejectReason = body.Reason
	}
	a.DecidedAt = &now
	if err := database.DB.Save(a).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, a)
}
